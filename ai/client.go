package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Role constants for chat messages.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// Message represents a single chat message.
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"` // for role=tool
}

// Client is an OpenAI-compatible streaming chat client.
type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}

// NewClient creates a Client from a Config.
func NewClient(cfg *Config) *Client {
	return &Client{
		BaseURL: strings.TrimRight(cfg.BaseURL, "/"),
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		HTTP:    &http.Client{},
	}
}

// --- Internal request/response types ---

type chatStreamRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Tools    []ToolDef `json:"tools,omitempty"`
}

type streamChunk struct {
	ID      string `json:"id"`
	Choices []struct {
		Index        int     `json:"index"`
		Delta        delta   `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

type delta struct {
	Role      string          `json:"role,omitempty"`
	Content   string          `json:"content,omitempty"`
	Reasoning string          `json:"reasoning_content,omitempty"`
	ToolCalls []streamToolCal `json:"tool_calls,omitempty"`
}

type streamToolCal struct {
	Index    int               `json:"index"`
	ID       string            `json:"id,omitempty"`
	Type     string            `json:"type,omitempty"`
	Function streamToolFunc    `json:"function"`
}

type streamToolFunc struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// --- Public API ---

// ChatStream sends a streaming chat completion request.
// onDelta is called for each content chunk received.
func (c *Client) ChatStream(
	messages []Message,
	onDelta func(content string),
) error {
	cb := StreamCallbacks{OnContent: onDelta}
	return c.ChatStreamEx(messages, nil, cb)
}

// ChatStreamWithSystem prepends a system message and streams.
func (c *Client) ChatStreamWithSystem(
	systemPrompt string,
	messages []Message,
	onDelta func(content string),
) error {
	all := make([]Message, 0, len(messages)+1)
	if systemPrompt != "" {
		all = append(all, Message{Role: RoleSystem, Content: systemPrompt})
	}
	all = append(all, messages...)
	return c.ChatStream(all, onDelta)
}

// ChatStreamEx is the full-featured streaming method supporting tools and callbacks.
func (c *Client) ChatStreamEx(
	messages []Message,
	tools []ToolDef,
	callbacks StreamCallbacks,
) error {
	return c.ChatStreamExWithContext(context.Background(), messages, tools, callbacks)
}

// ChatStreamExWithContext is like ChatStreamEx but with context support for cancellation.
func (c *Client) ChatStreamExWithContext(
	ctx context.Context,
	messages []Message,
	tools []ToolDef,
	callbacks StreamCallbacks,
) error {
	reqBody := chatStreamRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   true,
		Tools:    tools,
	}

	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := c.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyJSON))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return parseSSEStreamEx(resp.Body, callbacks)
}

// parseSSEStreamEx reads SSE events and dispatches to callbacks.
// It accumulates tool_calls across chunks (streaming partial arguments).
func parseSSEStreamEx(body io.Reader, cb StreamCallbacks) error {
	scanner := newLineScanner(body)
	// Accumulate tool calls by index
	toolCalls := map[int]*ToolCall{}

	for scanner.scan() {
		line := scanner.text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		for _, choice := range chunk.Choices {
			d := choice.Delta

			// Content delta
			if d.Content != "" && cb.OnContent != nil {
				cb.OnContent(d.Content)
			}

			// Reasoning/thinking delta
			if d.Reasoning != "" && cb.OnReasoning != nil {
				cb.OnReasoning(d.Reasoning)
			}

			// Tool calls delta
			for _, stc := range d.ToolCalls {
				existing, ok := toolCalls[stc.Index]
				if !ok {
					existing = &ToolCall{
						Index: stc.Index,
						ID:    stc.ID,
						Type:  "function",
						Function: ToolFunctionCall{
							Name: stc.Function.Name,
						},
					}
					toolCalls[stc.Index] = existing
				}
				// Accumulate
				if stc.ID != "" {
					existing.ID = stc.ID
				}
				if stc.Function.Name != "" {
					existing.Function.Name = stc.Function.Name
				}
				existing.Function.Arguments += stc.Function.Arguments

				if cb.OnToolCall != nil {
					cb.OnToolCall(*existing)
				}
			}

			// Finish reason
			if choice.FinishReason != nil && cb.OnFinish != nil {
				cb.OnFinish(*choice.FinishReason)
			}
		}
	}

	return scanner.getErr()
}

// --- Minimal line scanner ---

type lineScanner struct {
	reader *bufio.Reader
	line   string
	serr   error
}

func newLineScanner(r io.Reader) *lineScanner {
	return &lineScanner{reader: bufio.NewReaderSize(r, 64*1024)}
}

func (s *lineScanner) scan() bool {
	line, err := s.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		s.serr = err
		return false
	}
	s.line = strings.TrimRight(line, "\r\n")
	if err == io.EOF && line == "" {
		return false
	}
	return true
}

func (s *lineScanner) text() string  { return s.line }
func (s *lineScanner) getErr() error { return s.serr }

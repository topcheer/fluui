// Package mock provides a mock OpenAI-compatible streaming server for testing.
package mock

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"
)

// --- OpenAI Chat Completion Types ---

// ChatRequest mirrors the OpenAI chat completion request.
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

// ChatMessage is one message in the conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// StreamChunk represents one SSE chunk in the OpenAI streaming format.
type StreamChunk struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// StreamChoice is one choice in a streaming chunk.
type StreamChoice struct {
	Index        int           `json:"index"`
	Delta        Delta         `json:"delta"`
	FinishReason *string       `json:"finish_reason"`
}

// Delta is the incremental content in a stream chunk.
type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// --- Mock Script ---

// ScriptStep defines one step in a scripted mock AI response.
type ScriptStep struct {
	// Content to stream (will be split into word-level chunks)
	Content string
	// Delay before sending this chunk
	Delay time.Duration
	// Role for this delta (first chunk usually has role)
	Role string
	// FinishReason to attach (only for final chunk)
	FinishReason string
}

// Script is a sequence of steps the mock server will replay.
type Script struct {
	Steps     []ScriptStep
	ModelName string
}

// NewTextScript creates a simple text streaming script from a string.
func NewTextScript(text, modelName string) *Script {
	words := strings.Fields(text)
	steps := make([]ScriptStep, len(words))
	for i, w := range words {
		content := w
		if i < len(words)-1 {
			content += " "
		}
		role := ""
		if i == 0 {
			role = "assistant"
		}
		var finish string
		if i == len(words)-1 {
			finish = "stop"
		}
		steps[i] = ScriptStep{
			Content:      content,
			Delay:        10 * time.Millisecond,
			Role:         role,
			FinishReason: finish,
		}
	}
	return &Script{Steps: steps, ModelName: modelName}
}

// --- Mock SSE Server ---

// Server is a mock OpenAI-compatible streaming server.
type Server struct {
	*httptest.Server
	Script      *Script
	RequestLog []ChatRequest
	chunkCount int
	mu         sync.Mutex // protects RequestLog and chunkCount
}

// NewServer creates and starts a mock streaming server.
func NewServer(script *Script) *Server {
	s := &Server{Script: script}
	s.Server = httptest.NewServer(http.HandlerFunc(s.handler))
	return s
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	body, _ := io.ReadAll(r.Body)
	var req ChatRequest
	json.Unmarshal(body, &req)
	s.mu.Lock()
	s.RequestLog = append(s.RequestLog, req)
	s.mu.Unlock()

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	chunkID := "chatcmpl-test-001"
	created := time.Now().Unix()
	modelName := s.Script.ModelName
	if modelName == "" {
		modelName = "test-model"
	}

	for _, step := range s.Script.Steps {
		time.Sleep(step.Delay)

		chunk := StreamChunk{
			ID:      chunkID,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   modelName,
			Choices: []StreamChoice{{
				Index: 0,
				Delta: Delta{
					Role:    step.Role,
					Content: step.Content,
				},
			}},
		}

		if step.FinishReason != "" {
			fr := step.FinishReason
			chunk.Choices[0].FinishReason = &fr
		}

		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
		s.mu.Lock()
		s.chunkCount++
		s.mu.Unlock()
	}

	// Final [DONE] marker
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

// --- SSE Stream Reader (Client Side) ---

// SSEEvent represents one parsed SSE event.
type SSEEvent struct {
	Data string
}

// ParseSSEStream reads an SSE stream and returns events as they arrive.
// Each call to next() returns the next event, or io.EOF when done.
type SSEStreamReader struct {
	scanner *bufio.Scanner
}

// NewSSEStreamReader creates a new SSE reader.
func NewSSEStreamReader(r io.Reader) *SSEStreamReader {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	return &SSEStreamReader{scanner: sc}
}

// Next returns the next SSE event. Returns io.EOF at end.
func (r *SSEStreamReader) Next() (*SSEEvent, error) {
	var dataLines []string

	for r.scanner.Scan() {
		line := r.scanner.Text()

		// Empty line = event boundary
		if line == "" {
			if len(dataLines) > 0 {
				return &SSEEvent{
					Data: strings.Join(dataLines, "\n"),
				}, nil
			}
			continue
		}

		// Data line
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return nil, io.EOF
			}
			dataLines = append(dataLines, data)
		}
	}

	if err := r.scanner.Err(); err != nil {
		return nil, err
	}
	if len(dataLines) > 0 {
		return &SSEEvent{Data: strings.Join(dataLines, "\n")}, nil
	}
	return nil, io.EOF
}

// StreamClient connects to the mock server and reads streaming chunks.
type StreamClient struct {
	baseURL string
	http    *http.Client
}

// NewStreamClient creates a client for the given server URL.
func NewStreamClient(baseURL string) *StreamClient {
	return &StreamClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// ChatStream sends a request and returns a reader for SSE chunks.
func (c *StreamClient) ChatStream(model string, messages []ChatMessage) (*SSEStreamReader, error) {
	reqBody, _ := json.Marshal(ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
	})

	resp, err := c.http.Post(c.baseURL+"/v1/chat/completions", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	return NewSSEStreamReader(resp.Body), nil
}

// CollectAllChunks reads all chunks from a reader and returns them.
func CollectAllChunks(r *SSEStreamReader) ([]StreamChunk, error) {
	var chunks []StreamChunk
	for {
		ev, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return chunks, err
		}
		var chunk StreamChunk
		if err := json.Unmarshal([]byte(ev.Data), &chunk); err != nil {
			return chunks, fmt.Errorf("parse chunk: %w", err)
		}
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}

// CollectAllContent reads all chunks and concatenates the content deltas.
func CollectAllContent(r *SSEStreamReader) (string, error) {
	var sb strings.Builder
	for {
		ev, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return sb.String(), err
		}
		var chunk StreamChunk
		if err := json.Unmarshal([]byte(ev.Data), &chunk); err != nil {
			return sb.String(), fmt.Errorf("parse chunk: %w", err)
		}
		if len(chunk.Choices) > 0 {
			sb.WriteString(chunk.Choices[0].Delta.Content)
		}
	}
	return sb.String(), nil
}

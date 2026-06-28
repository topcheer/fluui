package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/topcheer/fluui/ai"
	"github.com/topcheer/fluui/block"
)

// AIBridge connects the ai.Client streaming API to ChatApp content blocks.
//
// When SendUserMessage is called, the bridge:
//  1. Creates a UserMessageBlock
//  2. Sends the conversation to the AI API via streaming
//  3. Routes content/reasoning/tool_call deltas to appropriate blocks
//  4. Maintains conversation history for multi-turn chat
type AIBridge struct {
	mu     sync.Mutex
	chat   *ChatApp
	client *ai.Client

	// conversation history sent to the API
	messages []ai.Message

	// system prompt (sent as first message if non-empty)
	systemPrompt string

	// streaming state
	dispatcher *block.StreamDispatcher
	cancelFunc context.CancelFunc
	streaming  bool

	// callbacks
	onError func(err error)
}

// NewAIBridge creates a bridge between the ChatApp and an AI client.
func NewAIBridge(chat *ChatApp, client *ai.Client) *AIBridge {
	return &AIBridge{
		chat:   chat,
		client: client,
	}
}

// SetSystemPrompt sets a system prompt to prepend to every conversation.
func (b *AIBridge) SetSystemPrompt(prompt string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.systemPrompt = prompt
}

// SetOnError sets a callback invoked when the AI streaming encounters an error.
func (b *AIBridge) SetOnError(fn func(err error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onError = fn
}

// Messages returns the current conversation history.
func (b *AIBridge) Messages() []ai.Message {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]ai.Message, len(b.messages))
	copy(out, b.messages)
	return out
}

// ClearHistory resets the conversation history.
func (b *AIBridge) ClearHistory() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messages = nil
}

// IsStreaming returns true if a streaming request is in progress.
func (b *AIBridge) IsStreaming() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.streaming
}

// StopStreaming cancels any in-flight streaming request.
func (b *AIBridge) StopStreaming() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.cancelFunc != nil {
		b.cancelFunc()
		b.cancelFunc = nil
	}
}

// SendUserMessage sends a user message to the AI and streams the response
// into ChatApp blocks. This method starts a goroutine and returns immediately.
// The caller should set up onError before calling this.
func (b *AIBridge) SendUserMessage(text string) {
	b.mu.Lock()
	if b.streaming {
		b.mu.Unlock()
		if b.onError != nil {
			b.onError(fmt.Errorf("streaming already in progress"))
		}
		return
	}
	b.streaming = true

	// Add user message to conversation history
	b.messages = append(b.messages, ai.Message{
		Role:    ai.RoleUser,
		Content: text,
	})

	// Snapshot the messages for this request
	msgs := make([]ai.Message, len(b.messages))
	copy(msgs, b.messages)

	sysPrompt := b.systemPrompt
	client := b.client
	b.mu.Unlock()

	go b.streamConversation(msgs, sysPrompt, client)
}

// streamConversation runs the AI streaming request and routes deltas to blocks.
func (b *AIBridge) streamConversation(msgs []ai.Message, sysPrompt string, client *ai.Client) {
	defer func() {
		b.mu.Lock()
		b.streaming = false
		b.cancelFunc = nil
		b.mu.Unlock()
	}()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	b.mu.Lock()
	b.cancelFunc = cancel
	dispatcher := block.NewStreamDispatcher(b.chat.Container())
	b.mu.Unlock()
	defer cancel()

	// Build callbacks that route AI deltas to block dispatcher
	callbacks := ai.StreamCallbacks{
		OnContent: func(text string) {
			_ = dispatcher.Dispatch(block.StreamDelta{
				Type:    "text",
				Content: text,
			})
		},
		OnReasoning: func(text string) {
			_ = dispatcher.Dispatch(block.StreamDelta{
				Type:    "thinking",
				Content: text,
			})
		},
		OnToolCall: func(tc ai.ToolCall) {
			_ = dispatcher.Dispatch(block.StreamDelta{
				Type:     "tool_call",
				ToolName: tc.Function.Name,
				ToolArgs: tc.Function.Arguments,
			})
		},
		OnFinish: func(reason string) {
			dispatcher.Flush()
		},
	}

	// Execute the streaming request with context for cancellation
	var err error
	if sysPrompt != "" {
		allMsgs := append([]ai.Message{
			{Role: ai.RoleSystem, Content: sysPrompt},
		}, msgs...)
		err = client.ChatStreamExWithContext(ctx, allMsgs, nil, callbacks)
	} else {
		err = client.ChatStreamExWithContext(ctx, msgs, nil, callbacks)
	}

	if err != nil {
		// Complete any pending blocks
		dispatcher.Flush()

		// Notify error callback
		b.mu.Lock()
		onErr := b.onError
		b.mu.Unlock()
		if onErr != nil {
			onErr(err)
		}
		return
	}

	// Flush any remaining streaming blocks
	dispatcher.Flush()

	// Collect assistant response for conversation history
	// Find the last assistant text block — must lock chat.mu since
	// we're reading the container's block list.
	b.mu.Lock()
	b.chat.mu.Lock()
	blocks := b.chat.container.Blocks()
	for i := len(blocks) - 1; i >= 0; i-- {
		if at, ok := blocks[i].(*block.AssistantTextBlock); ok {
			b.messages = append(b.messages, ai.Message{
				Role:    ai.RoleAssistant,
				Content: at.Content(),
			})
			break
		}
	}
	b.chat.mu.Unlock()
	b.mu.Unlock()

	_ = ctx // ctx consumed by ChatStreamExWithContext above
}

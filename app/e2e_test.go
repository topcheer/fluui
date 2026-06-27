package app

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/topcheer/fluui/ai"
	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/internal/mock"
)

// newMockChatApp creates a ChatApp wired to a mock SSE server.
func newMockChatApp(script *mock.Script) (*ChatApp, *mock.Server) {
	server := serverFromScript(script)
	// The mock server's SSE endpoint is at /chat/completions
	// and it matches what ai.Client expects.
	client := &ai.Client{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "test-model",
		HTTP:    &http.Client{},
	}
	app := NewChatApp(80, 24)
	app.SetAIClient(client)
	return app, server
}

// serverFromScript is a helper that wraps mock.NewServer to let us
// reference the mock package from the app test.
func serverFromScript(script *mock.Script) *mock.Server {
	return mock.NewServer(script)
}

// waitForStreamingDone polls ChatApp.IsStreaming() until false or timeout.
func waitForStreamingDone(t *testing.T, app *ChatApp, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !app.IsStreaming() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("streaming did not complete within %v", timeout)
}

func TestE2E_BasicChat(t *testing.T) {
	script := mock.NewTextScript("Hello world", "test-model")
	app, server := newMockChatApp(script)
	defer server.Close()

	var aiErr error
	app.SetOnAIError(func(err error) { aiErr = err })

	app.SendUserMessage("hi")
	waitForStreamingDone(t, app, 5*time.Second)

	if aiErr != nil {
		t.Fatalf("unexpected AI error: %v", aiErr)
	}

	blocks := app.Container().Blocks()
	// Expect at least: UserMessage + AssistantText
	if len(blocks) < 2 {
		t.Fatalf("expected at least 2 blocks, got %d", len(blocks))
	}

	// First block should be UserMessage.
	var um *block.UserMessageBlock
	for _, b := range blocks {
		if m, ok := b.(*block.UserMessageBlock); ok {
			um = m
			break
		}
	}
	if um == nil {
		t.Fatal("expected a UserMessageBlock")
	}
	if um.Content() != "hi" {
		t.Fatalf("expected user message 'hi', got %q", um.Content())
	}

	// Should have an AssistantTextBlock with content.
	var at *block.AssistantTextBlock
	for _, b := range blocks {
		if a, ok := b.(*block.AssistantTextBlock); ok {
			at = a
			break
		}
	}
	if at == nil {
		t.Fatal("expected an AssistantTextBlock")
	}
	if at.Content() == "" {
		t.Fatal("expected non-empty assistant text")
	}
}

func TestE2E_CancelStream(t *testing.T) {
	// Use a longer script so cancellation has time to fire.
	script := &mock.Script{
		Steps:     mock.NewTextScript("one two three four five six seven eight nine ten", "test-model").Steps,
		ModelName: "test-model",
	}
	// Add delays to each step so we can cancel mid-stream.
	for i := range script.Steps {
		script.Steps[i].Delay = 50 * time.Millisecond
	}

	app, server := newMockChatApp(script)
	defer server.Close()

	app.SendUserMessage("test")

	// Wait briefly to ensure streaming started, then cancel.
	time.Sleep(100 * time.Millisecond)
	app.StopStreaming()

	// Should complete without panic.
	waitForStreamingDone(t, app, 3*time.Second)
}

func TestE2E_ErrorHandling(t *testing.T) {
	// Create a server that returns HTTP 500.
	server := mock.NewServer(&mock.Script{
		Steps:     []mock.ScriptStep{},
		ModelName: "test-model",
	})
	// Override handler to return 500
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	})

	client := &ai.Client{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "test-model",
		HTTP:    &http.Client{},
	}
	app := NewChatApp(80, 24)
	app.SetAIClient(client)
	defer server.Close()

	var aiErr error
	app.SetOnAIError(func(err error) { aiErr = err })

	app.SendUserMessage("test")
	waitForStreamingDone(t, app, 5*time.Second)

	if aiErr == nil {
		t.Fatal("expected an error from the failed stream")
	}
}

func TestE2E_MultiTurnHistory(t *testing.T) {
	script := mock.NewTextScript("response one", "test-model")
	app, server := newMockChatApp(script)
	defer server.Close()

	var aiErr error
	app.SetOnAIError(func(err error) { aiErr = err })

	// First turn
	app.SendUserMessage("first question")
	waitForStreamingDone(t, app, 5*time.Second)

	if aiErr != nil {
		t.Fatalf("unexpected error on first turn: %v", aiErr)
	}

	// Check conversation history grew
	msgs := app.aiBridge.Messages()
	if len(msgs) < 2 {
		t.Fatalf("expected at least 2 messages after first turn, got %d", len(msgs))
	}
	if msgs[0].Content != "first question" {
		t.Fatalf("expected first message 'first question', got %q", msgs[0].Content)
	}

	// Second turn
	app.SendUserMessage("second question")
	waitForStreamingDone(t, app, 5*time.Second)

	if aiErr != nil {
		t.Fatalf("unexpected error on second turn: %v", aiErr)
	}

	msgs = app.aiBridge.Messages()
	// Should have: user1, assistant1, user2, assistant2
	if len(msgs) < 4 {
		t.Fatalf("expected at least 4 messages after second turn, got %d", len(msgs))
	}
}

func TestE2E_StopStreamingMidFlight(t *testing.T) {
	// Script with delays so we can stop mid-stream
	script := mock.NewTextScript("word1 word2 word3 word4", "test-model")
	for i := range script.Steps {
		script.Steps[i].Delay = 30 * time.Millisecond
	}

	app, server := newMockChatApp(script)
	defer server.Close()

	var callCount int32
	app.SetOnAIError(func(err error) {
		atomic.AddInt32(&callCount, 1)
	})

	app.SendUserMessage("test")

	// Give it a moment to start, then stop.
	time.Sleep(50 * time.Millisecond)
	app.StopStreaming()

	// Wait for streaming to end
	waitForStreamingDone(t, app, 3*time.Second)

	// IsStreaming must be false after StopStreaming completes
	if app.IsStreaming() {
		t.Fatal("expected IsStreaming() to be false after stop")
	}
}

func TestE2E_EmptyResponse(t *testing.T) {
	// Server that returns no content chunks, just finish.
	script := &mock.Script{
		Steps: []mock.ScriptStep{
			{
				Content:      "",
				FinishReason: "stop",
				Delay:        10 * time.Millisecond,
			},
		},
		ModelName: "test-model",
	}

	app, server := newMockChatApp(script)
	defer server.Close()

	var aiErr error
	app.SetOnAIError(func(err error) { aiErr = err })

	app.SendUserMessage("test")
	waitForStreamingDone(t, app, 5*time.Second)

	// Should not error — empty response is valid.
	if aiErr != nil {
		t.Fatalf("unexpected error for empty response: %v", aiErr)
	}

	// Should have at least the UserMessageBlock.
	blocks := app.Container().Blocks()
	found := false
	for _, b := range blocks {
		if _, ok := b.(*block.UserMessageBlock); ok {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected UserMessageBlock even with empty AI response")
	}
}

func TestE2E_ConcurrentSafe(t *testing.T) {
	script := mock.NewTextScript("concurrent test response", "test-model")
	app, server := newMockChatApp(script)
	defer server.Close()

	app.SetOnAIError(func(err error) {})

	// Send a message in a goroutine and concurrently call IsStreaming().
	// IsStreaming() acquires a.mu so it's safe to call during streaming.
	done := make(chan struct{})
	go func() {
		app.SendUserMessage("concurrent test")
		close(done)
	}()

	// Concurrently poll IsStreaming — this exercises mutex coverage.
	for i := 0; i < 50; i++ {
		select {
		case <-done:
		default:
		}
		_ = app.IsStreaming()
		time.Sleep(1 * time.Millisecond)
	}

	<-done
	waitForStreamingDone(t, app, 5*time.Second)

	// After streaming completes, safe to read blocks (no concurrent writers).
	blocks := app.Container().Blocks()
	if len(blocks) < 2 {
		t.Fatalf("expected at least 2 blocks, got %d", len(blocks))
	}
}

func TestE2E_LargeResponse(t *testing.T) {
	// Generate a large response.
	words := make([]string, 100)
	for i := range words {
		words[i] = fmt.Sprintf("word%d", i)
	}
	var text string
	for i, w := range words {
		if i > 0 {
			text += " "
		}
		text += w
	}

	script := mock.NewTextScript(text, "test-model")
	app, server := newMockChatApp(script)
	defer server.Close()

	var aiErr error
	app.SetOnAIError(func(err error) { aiErr = err })

	app.SendUserMessage("give me a long response")
	waitForStreamingDone(t, app, 10*time.Second)

	if aiErr != nil {
		t.Fatalf("unexpected error: %v", aiErr)
	}

	// Find the assistant block and verify it has content.
	blocks := app.Container().Blocks()
	for _, b := range blocks {
		if at, ok := b.(*block.AssistantTextBlock); ok {
			if at.Content() == "" {
				t.Fatal("expected non-empty assistant content for large response")
			}
			return
		}
	}
	t.Fatal("expected an AssistantTextBlock in large response")
}

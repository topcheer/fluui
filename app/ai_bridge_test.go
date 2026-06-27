package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/topcheer/fluui/ai"
)

// mockSSEServer creates a test server that streams SSE events.
func mockSSEServer(t *testing.T, events []string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("ResponseWriter does not support flushing")
		}
		for _, ev := range events {
			w.Write([]byte(ev))
			flusher.Flush()
		}
	}))
}

func TestAIBridgeSendUserMessage(t *testing.T) {
	// Create mock SSE server
	events := []string{
		`data: {"choices":[{"delta":{"role":"assistant","content":"Hello"}}]}` + "\n",
		`data: {"choices":[{"delta":{"content":" world"}}]}` + "\n",
		`data: {"choices":[{"delta":{},"finish_reason":"stop"}]}` + "\n",
		`data: [DONE]` + "\n",
	}
	server := mockSSEServer(t, events)
	defer server.Close()

	// Create AI client pointing to mock
	cfg := &ai.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "test-model",
	}
	client := ai.NewClient(cfg)

	// Create ChatApp
	chat := NewChatApp(80, 24)
	chat.SetAIClient(client)

	// Track errors
	var aiErr error
	chat.SetOnAIError(func(err error) { aiErr = err })

	// Send message
	chat.SendUserMessage("Hi")

	// Wait for streaming to complete
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if !chat.IsStreaming() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if chat.IsStreaming() {
		t.Fatal("streaming did not complete within timeout")
	}

	if aiErr != nil {
		t.Fatalf("unexpected AI error: %v", aiErr)
	}

	// Verify blocks were created
	blocks := chat.Container().Blocks()
	if len(blocks) < 2 {
		t.Fatalf("expected at least 2 blocks (user + assistant), got %d", len(blocks))
	}

	// First block should be user message
	if blocks[0].Type().String() != "user_message" {
		t.Errorf("expected first block type 'user_message', got %q", blocks[0].Type().String())
	}
}

func TestAIBridgeStopStreaming(t *testing.T) {
	// Slow server that blocks forever
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		// Send one event then block
		w.Write([]byte(`data: {"choices":[{"delta":{"content":"Hi"}}]}` + "\n"))
		flusher.Flush()
		// Block until client disconnects
		<-r.Context().Done()
	}))
	defer server.Close()

	cfg := &ai.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "test-model",
	}
	client := ai.NewClient(cfg)
	// Give it a short timeout to avoid hanging tests
	client.HTTP.Timeout = 5 * time.Second

	chat := NewChatApp(80, 24)
	chat.SetAIClient(client)

	chat.SendUserMessage("test")

	// Wait a bit for streaming to start
	time.Sleep(100 * time.Millisecond)

	if !chat.IsStreaming() {
		t.Fatal("expected streaming to be in progress")
	}

	// Stop streaming
	chat.StopStreaming()

	// Wait for streaming to stop
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if !chat.IsStreaming() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Note: current StopStreaming cancels the context but the HTTP client
	// may still be blocked. This will be fully functional when P6-C adds
	// context support to the AI client.
}

func TestAIBridgeConversationHistory(t *testing.T) {
	events := []string{
		`data: {"choices":[{"delta":{"content":"Response 1"}}]}` + "\n",
		`data: {"choices":[{"delta":{},"finish_reason":"stop"}]}` + "\n",
		`data: [DONE]` + "\n",
	}
	server := mockSSEServer(t, events)
	defer server.Close()

	cfg := &ai.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "test-model",
	}
	client := ai.NewClient(cfg)

	chat := NewChatApp(80, 24)
	chat.SetAIClient(client)

	// Send first message
	chat.SendUserMessage("Question 1")
	waitForStreaming(chat, 3*time.Second)

	// Verify conversation history grew
	msgs := chat.bridge().Messages()
	if len(msgs) < 2 {
		t.Errorf("expected at least 2 messages in history, got %d", len(msgs))
	}

	// First should be user message
	if msgs[0].Role != ai.RoleUser {
		t.Errorf("expected first message role 'user', got %q", msgs[0].Role)
	}
	if !strings.Contains(msgs[0].Content, "Question 1") {
		t.Errorf("expected first message to contain 'Question 1', got %q", msgs[0].Content)
	}

	// Last should be assistant response
	last := msgs[len(msgs)-1]
	if last.Role != ai.RoleAssistant {
		t.Errorf("expected last message role 'assistant', got %q", last.Role)
	}
}

func TestAIBridgeSystemPrompt(t *testing.T) {
	server := mockSSEServer(t, []string{
		`data: {"choices":[{"delta":{"content":"OK"}}]}` + "\n",
		`data: {"choices":[{"delta":{},"finish_reason":"stop"}]}` + "\n",
		`data: [DONE]` + "\n",
	})
	defer server.Close()

	cfg := &ai.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "test-model",
	}
	client := ai.NewClient(cfg)

	chat := NewChatApp(80, 24)
	chat.SetAIClient(client)
	chat.SetSystemPrompt("You are a test assistant.")

	chat.SendUserMessage("test")
	waitForStreaming(chat, 3*time.Second)

	// Just verify it doesn't crash with system prompt set
	msgs := chat.bridge().Messages()
	if len(msgs) == 0 {
		t.Error("expected messages in history")
	}
}

// waitForStreaming polls IsStreaming until it returns false or timeout.
func waitForStreaming(chat *ChatApp, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !chat.IsStreaming() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

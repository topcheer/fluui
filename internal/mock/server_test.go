package mock

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestMockServerBasicStreaming(t *testing.T) {
	script := &Script{
		ModelName: "test-gpt",
		Steps: []ScriptStep{
			{Role: "assistant", Content: "Hello", Delay: 5 * time.Millisecond},
			{Content: " world", Delay: 5 * time.Millisecond, FinishReason: "stop"},
		},
	}

	server := NewServer(script)
	defer server.Close()

	client := NewStreamClient(server.URL)
	reader, err := client.ChatStream("test-gpt", []ChatMessage{
		{Role: "user", Content: "hi"},
	})
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	content, err := CollectAllContent(reader)
	if err != nil && err != io.EOF {
		t.Fatalf("CollectAllContent failed: %v", err)
	}

	expected := "Hello world"
	if content != expected {
		t.Errorf("content: got %q, want %q", content, expected)
	}
}

func TestMockServerWordByWord(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog"
	script := NewTextScript(text, "test-model")
	server := NewServer(script)
	defer server.Close()

	client := NewStreamClient(server.URL)
	reader, err := client.ChatStream("test-model", []ChatMessage{
		{Role: "user", Content: "test"},
	})
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	content, err := CollectAllContent(reader)
	if err != nil && err != io.EOF {
		t.Fatalf("CollectAllContent failed: %v", err)
	}

	if !strings.HasPrefix(content, "The") {
		t.Errorf("content should start with 'The', got: %q", content[:min(3, len(content))])
	}
	if !strings.HasSuffix(content, "dog") {
		t.Errorf("content should end with 'dog', got: %q", content[max(0, len(content)-3):])
	}
}

func TestMockServerChunkParsing(t *testing.T) {
	script := NewTextScript("Hello world", "test-model")
	server := NewServer(script)
	defer server.Close()

	client := NewStreamClient(server.URL)
	reader, err := client.ChatStream("test-model", []ChatMessage{
		{Role: "user", Content: "hi"},
	})
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	chunks, err := CollectAllChunks(reader)
	if err != nil && err != io.EOF {
		t.Fatalf("CollectAllChunks failed: %v", err)
	}

	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}

	// First chunk should have role
	if chunks[0].Choices[0].Delta.Role != "assistant" {
		t.Errorf("first chunk role: got %q, want assistant", chunks[0].Choices[0].Delta.Role)
	}

	// Last chunk should have finish_reason
	last := chunks[len(chunks)-1]
	if last.Choices[0].FinishReason == nil || *last.Choices[0].FinishReason != "stop" {
		t.Error("last chunk should have finish_reason=stop")
	}
}

func TestMockServerRequestLog(t *testing.T) {
	script := NewTextScript("test", "test-model")
	server := NewServer(script)
	defer server.Close()

	client := NewStreamClient(server.URL)
	_, err := client.ChatStream("test-model", []ChatMessage{
		{Role: "user", Content: "hello"},
	})
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	// Give the server time to log the request
	time.Sleep(20 * time.Millisecond)

	if len(server.RequestLog) != 1 {
		t.Fatalf("expected 1 logged request, got %d", len(server.RequestLog))
	}

	req := server.RequestLog[0]
	if req.Model != "test-model" {
		t.Errorf("model: got %q, want test-model", req.Model)
	}
	if !req.Stream {
		t.Error("stream should be true")
	}
}

func TestSSEParseEmptyData(t *testing.T) {
	// Test that [DONE] is handled correctly
	r := strings.NewReader("data: [DONE]\n\n")
	reader := NewSSEStreamReader(r)

	_, err := reader.Next()
	if err != io.EOF {
		t.Errorf("expected EOF on [DONE], got %v", err)
	}
}

func TestSSEParseMultipleEvents(t *testing.T) {
	raw := "data: {\"id\":\"1\",\"content\":\"a\"}\n\ndata: {\"id\":\"2\",\"content\":\"b\"}\n\ndata: [DONE]\n\n"
	r := strings.NewReader(raw)
	reader := NewSSEStreamReader(r)

	ev1, err := reader.Next()
	if err != nil {
		t.Fatalf("first Next: %v", err)
	}
	if !strings.Contains(ev1.Data, "\"a\"") {
		t.Errorf("first event data: %q", ev1.Data)
	}

	ev2, err := reader.Next()
	if err != nil {
		t.Fatalf("second Next: %v", err)
	}
	if !strings.Contains(ev2.Data, "\"b\"") {
		t.Errorf("second event data: %q", ev2.Data)
	}

	_, err = reader.Next()
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
}

func TestSSEParseMultiLineData(t *testing.T) {
	raw := "data: line1\ndata: line2\n\ndata: [DONE]\n\n"
	r := strings.NewReader(raw)
	reader := NewSSEStreamReader(r)

	ev, err := reader.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	expected := "line1\nline2"
	if ev.Data != expected {
		t.Errorf("data: got %q, want %q", ev.Data, expected)
	}
}

func TestSSEServerHeaders(t *testing.T) {
	script := NewTextScript("test", "test-model")
	server := NewServer(script)
	defer server.Close()

	resp, err := http.Post(server.URL+"/v1/chat/completions", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/event-stream") {
		t.Errorf("Content-Type: got %q, want text/event-stream", ct)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

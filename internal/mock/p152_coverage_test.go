package mock

import (
	"io"
	"strings"
	"testing"
)

func TestP152_Next_MultipleEvents(t *testing.T) {
	data := "data: {\"id\":1}\n\ndata: {\"id\":2}\n\ndata: [DONE]\n\n"
	r := NewSSEStreamReader(strings.NewReader(data))

	ev1, err := r.Next()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ev1.Data != "{\"id\":1}" {
		t.Errorf("expected first event data, got %q", ev1.Data)
	}

	ev2, err := r.Next()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if ev2.Data != "{\"id\":2}" {
		t.Errorf("expected second event data, got %q", ev2.Data)
	}

	_, err = r.Next()
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
}

func TestP152_Next_EmptyLinesBetween(t *testing.T) {
	data := "data: hello\n\n\ndata: world\n\ndata: [DONE]\n\n"
	r := NewSSEStreamReader(strings.NewReader(data))

	ev1, err := r.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev1.Data != "hello" {
		t.Errorf("expected 'hello', got %q", ev1.Data)
	}

	ev2, err := r.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev2.Data != "world" {
		t.Errorf("expected 'world', got %q", ev2.Data)
	}
}

func TestP152_Next_NoDataPrefix(t *testing.T) {
	data := "event: ping\ndata: {\"ok\":true}\n\ndata: [DONE]\n\n"
	r := NewSSEStreamReader(strings.NewReader(data))

	ev, err := r.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Data != "{\"ok\":true}" {
		t.Errorf("expected JSON data, got %q", ev.Data)
	}
}

func TestP152_Next_ScannerError(t *testing.T) {
	// Use a reader that returns error on read
	r := NewSSEStreamReader(&errorReader{})
	_, err := r.Next()
	if err == nil {
		t.Error("expected error from error reader")
	}
}

func TestP152_Next_TrailingDataWithoutBoundary(t *testing.T) {
	data := "data: trailing"
	r := NewSSEStreamReader(strings.NewReader(data))

	ev, err := r.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Data != "trailing" {
		t.Errorf("expected 'trailing', got %q", ev.Data)
	}
}

func TestP152_Next_EmptyInput(t *testing.T) {
	r := NewSSEStreamReader(strings.NewReader(""))
	_, err := r.Next()
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
}

func TestP152_Next_MultiLineData(t *testing.T) {
	data := "data: line1\ndata: line2\n\ndata: [DONE]\n\n"
	r := NewSSEStreamReader(strings.NewReader(data))

	ev, err := r.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Multi-line data should be joined with \n
	if ev.Data != "line1\nline2" {
		t.Errorf("expected 'line1\\nline2', got %q", ev.Data)
	}
}

func TestP152_CollectAllChunks(t *testing.T) {
	data := "data: {\"id\":\"1\",\"choices\":[{\"delta\":{\"content\":\"Hi\"}}]}\n\ndata: [DONE]\n\n"
	r := NewSSEStreamReader(strings.NewReader(data))
	chunks, err := CollectAllChunks(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 1 {
		t.Errorf("expected 1 chunk, got %d", len(chunks))
	}
}

func TestP152_CollectAllChunks_InvalidJSON(t *testing.T) {
	data := "data: {invalid json}\n\ndata: [DONE]\n\n"
	r := NewSSEStreamReader(strings.NewReader(data))
	_, err := CollectAllChunks(r)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestP152_CollectAllChunks_Empty(t *testing.T) {
	r := NewSSEStreamReader(strings.NewReader(""))
	chunks, err := CollectAllChunks(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks, got %d", len(chunks))
	}
}

func TestP152_CollectAllContent(t *testing.T) {
	data := "data: {\"choices\":[{\"delta\":{\"content\":\"Hello \"}}]}\n\ndata: {\"choices\":[{\"delta\":{\"content\":\"World\"}}]}\n\ndata: [DONE]\n\n"
	r := NewSSEStreamReader(strings.NewReader(data))
	content, err := CollectAllContent(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", content)
	}
}

func TestP152_CollectAllContent_InvalidJSON(t *testing.T) {
	data := "data: {invalid}\n\ndata: [DONE]\n\n"
	r := NewSSEStreamReader(strings.NewReader(data))
	_, err := CollectAllContent(r)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestP152_CollectAllContent_NoChoices(t *testing.T) {
	data := "data: {\"id\":\"1\"}\n\ndata: [DONE]\n\n"
	r := NewSSEStreamReader(strings.NewReader(data))
	content, err := CollectAllContent(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "" {
		t.Errorf("expected empty content, got %q", content)
	}
}

func TestP152_ChatStream_Error(t *testing.T) {
	// Connect to non-existent server
	c := NewStreamClient("http://127.0.0.1:0")
	_, err := c.ChatStream("test", []ChatMessage{{Role: "user", Content: "hi"}})
	if err == nil {
		t.Error("expected connection error")
	}
}

func TestP152_NewStreamClient(t *testing.T) {
	c := NewStreamClient("http://localhost:8080")
	if c.baseURL != "http://localhost:8080" {
		t.Errorf("expected http://localhost:8080, got %s", c.baseURL)
	}
	if c.http == nil {
		t.Error("expected non-nil http client")
	}
}

// errorReader returns an error on every Read call
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}


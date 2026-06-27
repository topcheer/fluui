package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChatStreamBasic(t *testing.T) {
	// Create a mock SSE server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("Method = %s, want POST", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Auth header = %q", r.Header.Get("Authorization"))
		}

		// Stream SSE chunks
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		chunks := []string{"Hello", " world", "!"}
		for _, c := range chunks {
			chunk := streamChunk{
				Choices: []struct {
					Index        int    `json:"index"`
					Delta        delta  `json:"delta"`
					FinishReason *string `json:"finish_reason"`
				}{{
					Index: 0,
					Delta: delta{Content: c},
				}},
			}
			data, _ := json.Marshal(chunk)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		APIKey:  "test-key",
		HTTP:    server.Client(),
	}

	var received strings.Builder
	err := client.ChatStream([]Message{{Role: RoleUser, Content: "hi"}}, func(content string) {
		received.WriteString(content)
	})
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}
	if received.String() != "Hello world!" {
		t.Errorf("Received = %q, want 'Hello world!'", received.String())
	}
}

func TestChatStreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, `{"error": "invalid api key"}`)
	}))
	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		APIKey:  "bad-key",
		HTTP:    server.Client(),
	}

	err := client.ChatStream(nil, func(string) {})
	if err == nil {
		t.Error("ChatStream should fail on 401")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("Error should mention status code, got %v", err)
	}
}

func TestChatStreamEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		APIKey:  "test",
		HTTP:    server.Client(),
	}

	count := 0
	err := client.ChatStream(nil, func(string) { count++ })
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Delta count = %d, want 0", count)
	}
}

func TestChatStreamWithSystem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		chunk := streamChunk{
			Choices: []struct {
				Index        int    `json:"index"`
				Delta        delta  `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			}{{
				Index: 0,
				Delta: delta{Content: "OK"},
			}},
		}
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		APIKey:  "test",
		HTTP:    server.Client(),
	}

	err := client.ChatStreamWithSystem("be brief", []Message{{Role: RoleUser, Content: "hi"}}, func(string) {})
	if err != nil {
		t.Fatalf("ChatStreamWithSystem failed: %v", err)
	}
}

func TestChatStreamRoleInFirstChunk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		// First chunk has role but no content
		chunk1 := streamChunk{
			Choices: []struct {
				Index        int    `json:"index"`
				Delta        delta  `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			}{{
				Index: 0,
				Delta: delta{Role: RoleAssistant},
			}},
		}
		data1, _ := json.Marshal(chunk1)
		fmt.Fprintf(w, "data: %s\n\n", data1)
		flusher.Flush()

		// Second chunk has content
		chunk2 := streamChunk{
			Choices: []struct {
				Index        int    `json:"index"`
				Delta        delta  `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			}{{
				Index: 0,
				Delta: delta{Content: "Hi!"},
			}},
		}
		data2, _ := json.Marshal(chunk2)
		fmt.Fprintf(w, "data: %s\n\n", data2)
		flusher.Flush()

		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		APIKey:  "test",
		HTTP:    server.Client(),
	}

	var received strings.Builder
	err := client.ChatStream(nil, func(content string) {
		received.WriteString(content)
	})
	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}
	// Role-only chunks should not produce content
	if received.String() != "Hi!" {
		t.Errorf("Received = %q, want 'Hi!'", received.String())
	}
}

package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// === loadDotEnv coverage tests ===

func TestLoadDotEnv_BasicFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "TEST_KEY_1=value1\nTEST_KEY_2=value2\n"
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("TEST_KEY_1")
	defer os.Unsetenv("TEST_KEY_2")

	os.Unsetenv("TEST_KEY_1")
	os.Unsetenv("TEST_KEY_2")
	loadDotEnv(envPath)

	if v := os.Getenv("TEST_KEY_1"); v != "value1" {
		t.Errorf("TEST_KEY_1 = %q, want 'value1'", v)
	}
	if v := os.Getenv("TEST_KEY_2"); v != "value2" {
		t.Errorf("TEST_KEY_2 = %q, want 'value2'", v)
	}
}

func TestLoadDotEnv_Comments(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "# This is a comment\nTEST_CMT_KEY=value\n\n# Another comment\n"
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("TEST_CMT_KEY")

	os.Unsetenv("TEST_CMT_KEY")
	loadDotEnv(envPath)

	if v := os.Getenv("TEST_CMT_KEY"); v != "value" {
		t.Errorf("TEST_CMT_KEY = %q, want 'value'", v)
	}
}

func TestLoadDotEnv_EmptyLines(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "\n\nTEST_EMPTY_KEY=value\n\n\n"
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("TEST_EMPTY_KEY")

	os.Unsetenv("TEST_EMPTY_KEY")
	loadDotEnv(envPath)

	if v := os.Getenv("TEST_EMPTY_KEY"); v != "value" {
		t.Errorf("TEST_EMPTY_KEY = %q, want 'value'", v)
	}
}

func TestLoadDotEnv_QuotedValues(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := `TEST_QUOTE_DOUBLE="hello"
TEST_QUOTE_SINGLE='world'
TEST_QUOTE_NONE=plain
`
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("TEST_QUOTE_DOUBLE")
	defer os.Unsetenv("TEST_QUOTE_SINGLE")
	defer os.Unsetenv("TEST_QUOTE_NONE")

	os.Unsetenv("TEST_QUOTE_DOUBLE")
	os.Unsetenv("TEST_QUOTE_SINGLE")
	os.Unsetenv("TEST_QUOTE_NONE")
	loadDotEnv(envPath)

	if v := os.Getenv("TEST_QUOTE_DOUBLE"); v != "hello" {
		t.Errorf("TEST_QUOTE_DOUBLE = %q, want 'hello'", v)
	}
	if v := os.Getenv("TEST_QUOTE_SINGLE"); v != "world" {
		t.Errorf("TEST_QUOTE_SINGLE = %q, want 'world'", v)
	}
	if v := os.Getenv("TEST_QUOTE_NONE"); v != "plain" {
		t.Errorf("TEST_QUOTE_NONE = %q, want 'plain'", v)
	}
}

func TestLoadDotEnv_InvalidLines(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "INVALID_LINE\nKEY=val\n=NOKEY\n"
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("KEY")

	os.Unsetenv("KEY")
	loadDotEnv(envPath)

	if v := os.Getenv("KEY"); v != "val" {
		t.Errorf("KEY = %q, want 'val'", v)
	}
}

func TestLoadDotEnv_NotOverwriteExisting(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "EXISTING_KEY=from_file\n"
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("EXISTING_KEY")

	os.Setenv("EXISTING_KEY", "from_env")
	defer os.Setenv("EXISTING_KEY", "")
	loadDotEnv(envPath)

	if v := os.Getenv("EXISTING_KEY"); v != "from_env" {
		t.Errorf("EXISTING_KEY = %q, want 'from_env'", v)
	}
}

func TestLoadDotEnv_FileNotFound(t *testing.T) {
	// Should silently return - file doesn't exist is OK
	loadDotEnv("/nonexistent/path/.env")
}

func TestLoadDotEnv_SpacesAroundValue(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "TEST_SPACE_KEY  =  spaced_value  \n"
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Unsetenv("TEST_SPACE_KEY")

	os.Unsetenv("TEST_SPACE_KEY")
	loadDotEnv(envPath)

	if v := os.Getenv("TEST_SPACE_KEY"); v != "spaced_value" {
		t.Errorf("TEST_SPACE_KEY = %q, want 'spaced_value'", v)
	}
}

// === parseSSEStreamEx coverage tests ===

func TestParseSSE_Reasoning(t *testing.T) {
	chunk := streamChunk{
		Choices: []struct {
			Index        int     `json:"index"`
			Delta        delta   `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		}{{
			Index: 0,
			Delta: delta{Reasoning: "Let me think..."},
		}},
	}
	data, _ := json.Marshal(chunk)
	input := fmt.Sprintf("data: %s\n\ndata: [DONE]\n\n", data)

	var reasoning strings.Builder
	cb := StreamCallbacks{
		OnReasoning: func(text string) {
			reasoning.WriteString(text)
		},
	}
	if err := parseSSEStreamEx(strings.NewReader(input), cb); err != nil {
		t.Fatalf("parseSSEStreamEx error: %v", err)
	}
	if reasoning.String() != "Let me think..." {
		t.Errorf("Reasoning = %q, want 'Let me think...'", reasoning.String())
	}
}

func TestParseSSE_ToolCallAccumulation(t *testing.T) {
	chunk1 := streamChunk{
		Choices: []struct {
			Index        int     `json:"index"`
			Delta        delta   `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		}{{
			Index: 0,
			Delta: delta{
				ToolCalls: []streamToolCal{{
					Index: 0,
					ID:    "call_abc",
					Function: streamToolFunc{
						Name:      "get_weather",
						Arguments: `{"city":`,
					},
				}},
			},
		}},
	}
	chunk2 := streamChunk{
		Choices: []struct {
			Index        int     `json:"index"`
			Delta        delta   `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		}{{
			Index: 0,
			Delta: delta{
				ToolCalls: []streamToolCal{{
					Index: 0,
					Function: streamToolFunc{
						Arguments: ` "SF"}`,
					},
				}},
			},
		}},
	}
	data1, _ := json.Marshal(chunk1)
	data2, _ := json.Marshal(chunk2)
	input := fmt.Sprintf("data: %s\n\ndata: %s\n\ndata: [DONE]\n\n", data1, data2)

	var lastCall ToolCall
	callCount := 0
	cb := StreamCallbacks{
		OnToolCall: func(tc ToolCall) {
			lastCall = tc
			callCount++
		},
	}
	if err := parseSSEStreamEx(strings.NewReader(input), cb); err != nil {
		t.Fatalf("parseSSEStreamEx error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("ToolCall count = %d, want 2", callCount)
	}
	if lastCall.ID != "call_abc" {
		t.Errorf("ToolCall ID = %q, want 'call_abc'", lastCall.ID)
	}
	if lastCall.Function.Name != "get_weather" {
		t.Errorf("ToolCall name = %q, want 'get_weather'", lastCall.Function.Name)
	}
	expectedArgs := `{"city": "SF"}`
	if lastCall.Function.Arguments != expectedArgs {
		t.Errorf("ToolCall args = %q, want %q", lastCall.Function.Arguments, expectedArgs)
	}
}

func TestParseSSE_FinishReason(t *testing.T) {
	reason := "stop"
	chunk := streamChunk{
		Choices: []struct {
			Index        int     `json:"index"`
			Delta        delta   `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		}{{
			Index:        0,
			Delta:        delta{Content: "final"},
			FinishReason: &reason,
		}},
	}
	data, _ := json.Marshal(chunk)
	input := fmt.Sprintf("data: %s\n\ndata: [DONE]\n\n", data)

	var finishReason string
	var content strings.Builder
	cb := StreamCallbacks{
		OnContent: func(text string) { content.WriteString(text) },
		OnFinish:  func(r string) { finishReason = r },
	}
	if err := parseSSEStreamEx(strings.NewReader(input), cb); err != nil {
		t.Fatalf("parseSSEStreamEx error: %v", err)
	}
	if content.String() != "final" {
		t.Errorf("Content = %q, want 'final'", content.String())
	}
	if finishReason != "stop" {
		t.Errorf("FinishReason = %q, want 'stop'", finishReason)
	}
}

func TestParseSSE_InvalidJSON(t *testing.T) {
	input := "data: {invalid json}\n\ndata: [DONE]\n\n"
	var content strings.Builder
	cb := StreamCallbacks{
		OnContent: func(text string) { content.WriteString(text) },
	}
	if err := parseSSEStreamEx(strings.NewReader(input), cb); err != nil {
		t.Fatalf("parseSSEStreamEx error: %v", err)
	}
	if content.String() != "" {
		t.Errorf("Content = %q, want ''", content.String())
	}
}

func TestParseSSE_NonDataLines(t *testing.T) {
	input := ": comment\n\nevent: ping\n\nid: 42\n\ndata: [DONE]\n\n"
	cb := StreamCallbacks{}
	if err := parseSSEStreamEx(strings.NewReader(input), cb); err != nil {
		t.Fatalf("parseSSEStreamEx error: %v", err)
	}
}

func TestParseSSE_ToolCallIDUpdate(t *testing.T) {
	chunk1 := streamChunk{
		Choices: []struct {
			Index        int     `json:"index"`
			Delta        delta   `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		}{{
			Index: 0,
			Delta: delta{
				ToolCalls: []streamToolCal{{
					Index: 0,
					Function: streamToolFunc{
						Name:      "search",
						Arguments: `{"q":`,
					},
				}},
			},
		}},
	}
	chunk2 := streamChunk{
		Choices: []struct {
			Index        int     `json:"index"`
			Delta        delta   `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		}{{
			Index: 0,
			Delta: delta{
				ToolCalls: []streamToolCal{{
					Index: 0,
					ID:    "call_xyz",
					Function: streamToolFunc{
						Arguments: ` "go"}`,
					},
				}},
			},
		}},
	}
	data1, _ := json.Marshal(chunk1)
	data2, _ := json.Marshal(chunk2)
	input := fmt.Sprintf("data: %s\n\ndata: %s\n\ndata: [DONE]\n\n", data1, data2)

	var lastCall ToolCall
	cb := StreamCallbacks{
		OnToolCall: func(tc ToolCall) { lastCall = tc },
	}
	if err := parseSSEStreamEx(strings.NewReader(input), cb); err != nil {
		t.Fatalf("parseSSEStreamEx error: %v", err)
	}
	if lastCall.ID != "call_xyz" {
		t.Errorf("ToolCall ID = %q, want 'call_xyz'", lastCall.ID)
	}
}

func TestParseSSE_MultipleChoices(t *testing.T) {
	chunk := streamChunk{
		Choices: []struct {
			Index        int     `json:"index"`
			Delta        delta   `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		}{
			{Index: 0, Delta: delta{Content: "a"}},
			{Index: 1, Delta: delta{Content: "b"}},
		},
	}
	data, _ := json.Marshal(chunk)
	input := fmt.Sprintf("data: %s\n\ndata: [DONE]\n\n", data)

	var received strings.Builder
	err := parseSSEStreamEx(strings.NewReader(input), StreamCallbacks{
		OnContent: func(text string) {
			received.WriteString(text)
		},
	})
	if err != nil {
		t.Fatalf("parseSSEStreamEx failed: %v", err)
	}
	if received.String() != "ab" {
		t.Errorf("Received = %q, want 'ab'", received.String())
	}
}

// === lineScanner coverage tests ===

func TestLineScanner_EmptyInput(t *testing.T) {
	s := newLineScanner(strings.NewReader(""))
	if s.scan() {
		t.Error("scan() on empty input should return false")
	}
}

func TestLineScanner_SingleLine(t *testing.T) {
	s := newLineScanner(strings.NewReader("hello\n"))
	if !s.scan() {
		t.Fatal("scan() should return true")
	}
	if s.text() != "hello" {
		t.Errorf("text() = %q, want 'hello'", s.text())
	}
	if s.scan() {
		t.Error("second scan() should return false")
	}
}

func TestLineScanner_NoTrailingNewline(t *testing.T) {
	s := newLineScanner(strings.NewReader("hello"))
	if !s.scan() {
		t.Fatal("scan() should return true for line without newline")
	}
	if s.text() != "hello" {
		t.Errorf("text() = %q, want 'hello'", s.text())
	}
	if s.scan() {
		t.Error("second scan() should return false")
	}
}

func TestLineScanner_GetErr(t *testing.T) {
	s := newLineScanner(strings.NewReader("line\n"))
	s.scan()
	if err := s.getErr(); err != nil {
		t.Errorf("getErr() = %v, want nil for normal scan", err)
	}
}

func TestLineScanner_CRLF(t *testing.T) {
	s := newLineScanner(strings.NewReader("line1\r\nline2\r\n"))
	var lines []string
	for s.scan() {
		lines = append(lines, s.text())
	}
	if len(lines) != 2 {
		t.Fatalf("Line count = %d, want 2", len(lines))
	}
	for _, l := range lines {
		if strings.Contains(l, "\r") {
			t.Errorf("Line %q contains CR", l)
		}
	}
}

// === ChatStreamExWithContext coverage ===

func TestChatStreamExWithContext_Cancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := &Client{
		BaseURL: server.URL,
		APIKey:  "test",
		HTTP:    server.Client(),
	}
	err := client.ChatStreamExWithContext(ctx, nil, nil, StreamCallbacks{})
	if err == nil {
		t.Error("ChatStreamExWithContext should fail with cancelled context")
	}
}

func TestChatStreamEx_ToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)

		chunk := streamChunk{
			Choices: []struct {
				Index        int     `json:"index"`
				Delta        delta   `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			}{{
				Index: 0,
				Delta: delta{
					ToolCalls: []streamToolCal{{
						Index: 0,
						ID:    "call_1",
						Function: streamToolFunc{
							Name:      "calc",
							Arguments: `{"x":1}`,
						},
					}},
				},
			}},
		}
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\ndata: [DONE]\n\n", data)
		flusher.Flush()
	}))
	defer server.Close()

	client := &Client{
		BaseURL: server.URL,
		APIKey:  "test",
		HTTP:    server.Client(),
	}

	var toolCalls []ToolCall
	cb := StreamCallbacks{
		OnToolCall: func(tc ToolCall) { toolCalls = append(toolCalls, tc) },
	}
	err := client.ChatStreamEx(nil, []ToolDef{{
		Type:     "function",
		Function: ToolFunction{Name: "calc", Description: "calculator"},
	}}, cb)
	if err != nil {
		t.Fatalf("ChatStreamEx failed: %v", err)
	}
	if len(toolCalls) != 1 {
		t.Fatalf("ToolCall count = %d, want 1", len(toolCalls))
	}
	if toolCalls[0].Function.Name != "calc" {
		t.Errorf("ToolCall name = %q, want 'calc'", toolCalls[0].Function.Name)
	}
}

// === Config coverage ===

func TestLoadConfig_FromEnvVars(t *testing.T) {
	os.Setenv("FLUUI_LLM_API_KEY", "sk-test-1234567890")
	defer os.Unsetenv("FLUUI_LLM_API_KEY")
	os.Setenv("FLUUI_LLM_BASE_URL", "https://api.test.com/v1")
	defer os.Unsetenv("FLUUI_LLM_BASE_URL")
	os.Setenv("FLUUI_LLM_MODEL", "gpt-4")
	defer os.Unsetenv("FLUUI_LLM_MODEL")

	cfg, err := LoadConfig("/nonexistent/.env")
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if cfg.APIKey != "sk-test-1234567890" {
		t.Errorf("APIKey = %q, want 'sk-test-1234567890'", cfg.APIKey)
	}
	if cfg.BaseURL != "https://api.test.com/v1" {
		t.Errorf("BaseURL = %q, want 'https://api.test.com/v1'", cfg.BaseURL)
	}
	if cfg.Model != "gpt-4" {
		t.Errorf("Model = %q, want 'gpt-4'", cfg.Model)
	}
}

func TestLoadConfig_DotEnvFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	content := "FLUUI_LLM_API_KEY=from-file-12345678\nFLUUI_LLM_MODEL=claude-3\n"
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	os.Unsetenv("FLUUI_LLM_API_KEY")
	defer os.Unsetenv("FLUUI_LLM_MODEL")

	cfg, err := LoadConfig(envPath)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if cfg.APIKey != "from-file-12345678" {
		t.Errorf("APIKey = %q, want 'from-file-12345678'", cfg.APIKey)
	}
	if cfg.Model != "claude-3" {
		t.Errorf("Model = %q, want 'claude-3'", cfg.Model)
	}
}

func TestNewClient_TrimsBaseURL(t *testing.T) {
	cfg := &Config{
		BaseURL: "https://api.test.com/v1/",
		APIKey:  "test",
		Model:   "gpt-4",
	}
	client := NewClient(cfg)
	if client.BaseURL != "https://api.test.com/v1" {
		t.Errorf("BaseURL = %q, want 'https://api.test.com/v1' (trailing slash trimmed)", client.BaseURL)
	}
}

func TestConfig_String(t *testing.T) {
	cfg := &Config{
		BaseURL: "https://api.test.com/v1",
		APIKey:  "sk-1234567890abcdef",
		Model:   "gpt-4",
	}
	s := cfg.String()
	if !strings.Contains(s, "gpt-4") {
		t.Errorf("String() should contain model name: %q", s)
	}
	if !strings.Contains(s, "api.test.com") {
		t.Errorf("String() should contain base URL: %q", s)
	}
	// Should NOT contain the full API key (should be masked)
	if strings.Contains(s, "sk-1234567890abcdef") {
		t.Errorf("String() should mask API key: %q", s)
	}
}

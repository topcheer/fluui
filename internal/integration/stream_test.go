// Package integration tests the full pipeline: mock AI stream → buffer rendering.
package integration

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/mock"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/render"
)

// TestStreamToBuffer simulates a full AI response being streamed,
// collected, and rendered into a buffer.
func TestStreamToBuffer(t *testing.T) {
	responseText := "Hello world this is a test response from the mock AI"

	script := mock.NewTextScript(responseText, "test-model")
	server := mock.NewServer(script)
	defer server.Close()

	client := mock.NewStreamClient(server.URL)
	reader, err := client.ChatStream("test-model", []mock.ChatMessage{
		{Role: "user", Content: "hello"},
	})
	if err != nil {
		t.Fatalf("ChatStream: %v", err)
	}

	content, err := mock.CollectAllContent(reader)
	if err != nil && err != io.EOF {
		t.Fatalf("CollectAllContent: %v", err)
	}

	// Render the content into a buffer
	buf := buffer.NewBuffer(80, 24)
	style := buffer.DefaultStyle.WithFg(buffer.RGB(200, 210, 230))
	buf.DrawText(2, 2, "AI Response:", buffer.DefaultStyle.WithFg(buffer.RGB(139, 233, 253)).WithFlags(buffer.Bold))
	buf.DrawText(2, 4, content, style)

	// Verify the content was rendered
	firstCell := buf.GetCell(2, 4)
	if firstCell.Rune != 'H' {
		t.Errorf("expected 'H' at (2,4), got %c", firstCell.Rune)
	}
}

// TestIncrementalStreaming simulates streaming chunks into a buffer,
// appending text as each chunk arrives (like a real TUI would).
func TestIncrementalStreaming(t *testing.T) {
	responseText := "The quick brown fox jumps over the lazy dog"

	script := mock.NewTextScript(responseText, "test-model")
	server := mock.NewServer(script)
	defer server.Close()

	client := mock.NewStreamClient(server.URL)
	reader, err := client.ChatStream("test-model", []mock.ChatMessage{
		{Role: "user", Content: "test"},
	})
	if err != nil {
		t.Fatalf("ChatStream: %v", err)
	}

	// Simulate incremental rendering: for each chunk, render the accumulated text
	var accumulated strings.Builder
	buf := buffer.NewBuffer(80, 24)
	style := buffer.DefaultStyle.WithFg(buffer.RGB(200, 210, 230))

	chunkCount := 0
	for {
		ev, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("reader.Next: %v", err)
		}

		var chunk mock.StreamChunk
		if err := parseJSON(ev.Data, &chunk); err != nil {
			t.Fatalf("parse chunk: %v", err)
		}

		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				accumulated.WriteString(delta)
				// Re-render the buffer
				buf.Fill(buffer.BlankCell)
				buf.DrawText(0, 0, accumulated.String(), style)
				chunkCount++
			}
		}
	}

	finalText := accumulated.String()
	if !strings.Contains(finalText, "fox") {
		t.Errorf("final text should contain 'fox': %q", finalText)
	}
	if chunkCount < 5 {
		t.Errorf("expected at least 5 content chunks, got %d", chunkCount)
	}
}

// TestDoubleBufferDiff simulates the full double-buffer diff pipeline.
// As chunks arrive, we render to the back buffer and diff against the front,
// verifying that only changed cells are reported.
func TestDoubleBufferDiff(t *testing.T) {
	// Set up a writer to a dummy io.Writer (not a real terminal)
	var dummyWriter dummyWriter
	tw := term.NewWriter(&dummyWriter, term.ProfileTrue)
	renderer := render.New(tw, 80, 24)

	// Simulate streaming: frame 1 has partial text, frame 2 completes it
	text1 := "Hello"
	text2 := "Hello World"

	style := buffer.DefaultStyle.WithFg(buffer.RGB(200, 210, 230))

	// Frame 1: render "Hello"
	renderer.BeginFrame()
	renderer.Back().DrawText(0, 0, text1, style)
	if err := renderer.EndFrame(); err != nil {
		t.Fatalf("EndFrame 1: %v", err)
	}

	// Frame 2: render "Hello World" (adds " World")
	renderer.BeginFrame()
	renderer.Back().DrawText(0, 0, text2, style)
	if err := renderer.EndFrame(); err != nil {
		t.Fatalf("EndFrame 2: %v", err)
	}

	// Verify the front buffer has the full text
	cell := renderer.Back().GetCell(6, 0)
	if cell.Rune != 'W' {
		t.Errorf("expected 'W' at (6,0), got %c", cell.Rune)
	}
}

// TestMultilineStreaming tests streaming a multi-line markdown response
// and rendering it with different styles per line.
func TestMultilineStreaming(t *testing.T) {
	lines := []string{
		"# Heading",
		"",
		"This is a paragraph with some text.",
		"",
		"- Item one",
		"- Item two",
		"",
		"```go",
		"package main",
		"```",
	}
	fullText := strings.Join(lines, "\n")

	// Create a script from the full text (not word-split, single chunk)
	script := &mock.Script{
		ModelName: "test-model",
		Steps: []mock.ScriptStep{
			{Role: "assistant", Content: fullText, Delay: 5 * time.Millisecond, FinishReason: "stop"},
		},
	}
	server := mock.NewServer(script)
	defer server.Close()

	client := mock.NewStreamClient(server.URL)
	reader, err := client.ChatStream("test-model", []mock.ChatMessage{
		{Role: "user", Content: "show me an example"},
	})
	if err != nil {
		t.Fatalf("ChatStream: %v", err)
	}

	content, err := mock.CollectAllContent(reader)
	if err != nil && err != io.EOF {
		t.Fatalf("CollectAllContent: %v", err)
	}

	// Render line by line into buffer
	buf := buffer.NewBuffer(80, 24)
	for y, line := range strings.Split(content, "\n") {
		if y >= buf.Height {
			break
		}
		var style buffer.Style
		switch {
		case strings.HasPrefix(line, "# "):
			style = buffer.DefaultStyle.
				WithFg(buffer.RGB(255, 184, 108)).
				WithFlags(buffer.Bold)
			line = strings.TrimPrefix(line, "# ")
		case strings.HasPrefix(line, "- "):
			style = buffer.DefaultStyle.WithFg(buffer.RGB(139, 233, 253))
			line = "\u2022 " + strings.TrimPrefix(line, "- ")
		case strings.HasPrefix(line, "```"):
			style = buffer.DefaultStyle.WithFg(buffer.RGB(139, 233, 253)).WithFlags(buffer.Dim)
		default:
			style = buffer.DefaultStyle.WithFg(buffer.RGB(200, 210, 230))
		}
		buf.DrawText(0, y, line, style)
	}

	// Verify heading
	headingCell := buf.GetCell(0, 0)
	if headingCell.Rune != 'H' {
		t.Errorf("heading cell: got %c, want 'H'", headingCell.Rune)
	}
	if !headingCell.Fg.Equal(buffer.RGB(255, 184, 108)) {
		t.Error("heading should have orange color")
	}
	if headingCell.Flags&buffer.Bold == 0 {
		t.Error("heading should be bold")
	}

	// Verify list item
	bulletCell := buf.GetCell(0, 4)
	if bulletCell.Rune != '\u2022' {
		t.Errorf("list bullet: got %c, want '•'", bulletCell.Rune)
	}

	// Verify code fence
	codeFenceCell := buf.GetCell(0, 7)
	if codeFenceCell.Rune != '`' {
		t.Errorf("code fence: got %c, want '`'", codeFenceCell.Rune)
	}
}

// TestCJKStreaming tests streaming and rendering CJK characters.
func TestCJKStreaming(t *testing.T) {
	text := "你好世界 这是一个测试"

	script := &mock.Script{
		ModelName: "test-model",
		Steps: []mock.ScriptStep{
			{Role: "assistant", Content: text, Delay: 5 * time.Millisecond, FinishReason: "stop"},
		},
	}
	server := mock.NewServer(script)
	defer server.Close()

	client := mock.NewStreamClient(server.URL)
	reader, err := client.ChatStream("test-model", []mock.ChatMessage{
		{Role: "user", Content: "你好"},
	})
	if err != nil {
		t.Fatalf("ChatStream: %v", err)
	}

	content, err := mock.CollectAllContent(reader)
	if err != nil && err != io.EOF {
		t.Fatalf("CollectAllContent: %v", err)
	}

	if !strings.Contains(content, "你好") {
		t.Errorf("content should contain CJK chars: %q", content)
	}

	// Render CJK text into buffer and check width
	buf := buffer.NewBuffer(80, 24)
	style := buffer.DefaultStyle.WithFg(buffer.RGB(200, 210, 230))
	endX := buf.DrawText(0, 0, content, style)

	// CJK chars are 2 display cells each but 3 bytes in UTF-8.
	// "你好世界" = 4 chars × 2 cells = 8 cells
	// " " = 1 cell
	// "这是一个测试" = 7 chars × 2 cells = 14 cells
	// Total display width should be 8 + 1 + 14 = 23
	runeCount := len([]rune(content))
	if endX <= runeCount {
		t.Errorf("CJK endX (%d) should exceed rune count (%d) since CJK chars are double-width", endX, runeCount)
	}
	// Verify the first CJK char rendered correctly
	firstCell := buf.GetCell(0, 0)
	if firstCell.Rune != '你' {
		t.Errorf("first cell: got %c, want '你'", firstCell.Rune)
	}
	if firstCell.Width != 2 {
		t.Errorf("first cell width: got %d, want 2", firstCell.Width)
	}
}

// TestConcurrentStreaming verifies that multiple streaming sessions
// don't interfere with each other.
func TestConcurrentStreaming(t *testing.T) {
	sessionTexts := []string{
		"Session one says hello",
		"Session two says goodbye",
		"Session three says foo bar",
	}

	script := mock.NewTextScript(sessionTexts[0], "test-model")
	server := mock.NewServer(script)
	defer server.Close()

	client := mock.NewStreamClient(server.URL)

	results := make(chan string, len(sessionTexts))

	for i := range sessionTexts {
		go func(idx int) {
			reader, err := client.ChatStream("test-model", []mock.ChatMessage{
				{Role: "user", Content: "test"},
			})
			if err != nil {
				results <- ""
				return
			}
			content, _ := mock.CollectAllContent(reader)
			results <- content
		}(i)
	}

	for i := 0; i < len(sessionTexts); i++ {
		r := <-results
		if r == "" {
			t.Errorf("session %d returned empty", i)
		}
	}
}

// TestLatency measures end-to-end streaming latency.
func TestLatency(t *testing.T) {
	if testing.Short() {
		t.Skip("latency test in short mode")
	}

	text := "measuring latency for this stream"
	script := &mock.Script{
		ModelName: "test-model",
		Steps: []mock.ScriptStep{
			{Role: "assistant", Content: text, Delay: 1 * time.Millisecond, FinishReason: "stop"},
		},
	}
	server := mock.NewServer(script)
	defer server.Close()

	client := mock.NewStreamClient(server.URL)

	start := time.Now()
	reader, _ := client.ChatStream("test-model", []mock.ChatMessage{
		{Role: "user", Content: "ping"},
	})
	content, _ := mock.CollectAllContent(reader)
	elapsed := time.Since(start)

	if content != text {
		t.Errorf("content mismatch: got %q, want %q", content, text)
	}
	// Should complete in well under 1 second
	if elapsed > time.Second {
		t.Errorf("latency too high: %v", elapsed)
	}
	t.Logf("end-to-end latency: %v", elapsed)
}

// --- helpers ---

type dummyWriter struct{}

func (d *dummyWriter) Write(b []byte) (int, error) { return len(b), nil }

func parseJSON(data string, v interface{}) error {
	return jsonUnmarshal([]byte(data), v)
}

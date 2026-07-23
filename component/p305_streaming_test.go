package component

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === MarkdownViewer streaming tests (Direction C) ===

func TestP305_MDViewer_SetStreaming(t *testing.T) {
	v := NewMarkdownViewer("# Hello")
	v.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	v.SetStreaming(true)
	if !v.IsStreaming() {
		t.Error("should be streaming")
	}
	v.SetStreaming(false)
	if v.IsStreaming() {
		t.Error("should not be streaming")
	}
}

func TestP305_MDViewer_AppendDelta_FirstRender(t *testing.T) {
	v := NewMarkdownViewer("")
	v.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	v.SetStreaming(true)
	v.AppendDelta("# Title")
	// First delta should trigger immediate render
	if v.StreamSource() != "# Title" {
		t.Errorf("StreamSource = %q", v.StreamSource())
	}
}

func TestP305_MDViewer_AppendDelta_Debounce(t *testing.T) {
	v := NewMarkdownViewer("")
	v.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	v.SetStreaming(true)
	// First delta renders immediately (streamDebN=1)
	v.AppendDelta("a")
	// Subsequent deltas buffer (streamDebN 2-7)
	for i := 0; i < 6; i++ {
		v.AppendDelta("x")
	}
	// source has "a" (flushed), deltaBuf has "xxxxxx" (6)
	// StreamSource returns source + deltaBuf = "axxxxxx" (7 chars)
	if v.StreamSource() != "axxxxxx" {
		t.Errorf("StreamSource = %q, want 'axxxxxx'", v.StreamSource())
	}
	// 8th call (streamDebN=8) triggers render, flushes buffer
	v.AppendDelta("y")
	// source = "a" + "xxxxxx" + "y" = "axxxxxxy" (8 chars)
	if v.StreamSource() != "axxxxxxy" {
		t.Errorf("after 8th delta StreamSource = %q, want 'axxxxxxy'", v.StreamSource())
	}
}

func TestP305_MDViewer_FlushStream(t *testing.T) {
	v := NewMarkdownViewer("")
	v.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	v.SetStreaming(true)
	v.AppendDelta("initial") // renders (first)
	v.AppendDelta(" buffered") // debounced
	v.FlushStream()
	if v.StreamSource() != "initial buffered" {
		t.Errorf("StreamSource = %q", v.StreamSource())
	}
}

func TestP305_MDViewer_SetStreaming_False_Flushes(t *testing.T) {
	v := NewMarkdownViewer("")
	v.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	v.SetStreaming(true)
	v.AppendDelta("hello") // renders
	v.AppendDelta(" world") // debuffed
	v.SetStreaming(false)   // flush
	if v.StreamSource() != "hello world" {
		t.Errorf("after flush StreamSource = %q", v.StreamSource())
	}
}

func TestP305_MDViewer_AppendDelta_NotStreaming(t *testing.T) {
	v := NewMarkdownViewer("# Base")
	v.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	// Without streaming mode, AppendDelta should render immediately
	v.AppendDelta("\nMore text")
	if v.StreamSource() != "# Base\nMore text" {
		t.Errorf("StreamSource = %q", v.StreamSource())
	}
}

func TestP305_MDViewer_StreamingMarkdown(t *testing.T) {
	v := NewMarkdownViewer("")
	v.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	v.SetStreaming(true)

	// Simulate streaming a markdown document token by token
	tokens := strings.Fields("# AI Response\n\nThis is a **bold** statement with `code`.\n\n- Item 1\n- Item 2")
	v.AppendDelta(tokens[0]) // first token, renders
	for _, tk := range tokens[1:] {
		v.AppendDelta(" " + tk)
	}
	v.FlushStream()

	src := v.StreamSource()
	if !strings.Contains(src, "bold") {
		t.Errorf("source missing 'bold': %q", src)
	}
}

func TestP305_MDViewer_Paint_Streaming(t *testing.T) {
	v := NewMarkdownViewer("# Streaming")
	v.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	v.SetStreaming(true)
	v.AppendDelta("\nContent")
	buf := buffer.NewBuffer(60, 10)
	v.Paint(buf) // should not panic during streaming
}

func TestP305_MDViewer_StreamingFullFlow(t *testing.T) {
	v := NewMarkdownViewer("")
	v.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	v.SetStreaming(true)

	// Stream a complete code block
	v.AppendDelta("```go\n")
	v.AppendDelta("func main() {\n")
	v.AppendDelta("    fmt.Println(\"Hello\")\n")
	v.AppendDelta("}\n")
	v.AppendDelta("```")
	v.SetStreaming(false)

	src := v.StreamSource()
	if !strings.Contains(src, "func main") {
		t.Errorf("source missing code: %q", src)
	}
}

// === ConversationView renderBubble optimization verification (Direction D) ===

func TestP305_ConversationView_RendersCorrectly(t *testing.T) {
	cv := NewConversationView()
	cv.AddUserMessage("Hello")
	cv.AddAssistantMessage("Hi there!", "gpt-4")
	cv.AddSystemMessage("System msg")
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	// Paint twice to verify renderBubble reuse works
	cv.Paint(buf)
	cv.Paint(buf)
	// No panic, no deadlock = success
}

func TestP305_ConversationView_ErrorMessage(t *testing.T) {
	cv := NewConversationView()
	cv.AddMessage(ConversationMessage{
		Role:    RoleAssistant,
		Content: "Something failed",
		Error:   true,
	})
	cv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	cv.Paint(buf)
}

// === stripTerminator coverage (Direction F) ===

func TestP305_StripTerminator(t *testing.T) {
	// Just call the internal function via ParseOSC133 which uses stripTerminatorChecked
	_, ok := term.ParseOSC133("\x1b]133;A\x07")
	if !ok {
		t.Error("ParseOSC133 should succeed with valid input")
	}
}

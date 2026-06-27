package app

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

func TestLastBlockText_Empty(t *testing.T) {
	app := NewChatApp(80, 24)

	text, ok := app.LastBlockText()
	if ok {
		t.Errorf("expected ok=false on empty app")
	}
	if text != "" {
		t.Errorf("expected empty string, got %q", text)
	}
}

func TestLastBlockText_AssistantText(t *testing.T) {
	app := NewChatApp(80, 24)
	at := app.AddAssistantText()
	at.AppendDelta("Hello from AI!")

	text, ok := app.LastBlockText()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if text != "Hello from AI!" {
		t.Errorf("got %q, want %q", text, "Hello from AI!")
	}

	// Verify OSC52 round-trip
	seq := term.CopyOSC52(text)
	if !strings.HasPrefix(seq, "\x1b]52;c;") {
		t.Errorf("expected OSC52 sequence, got %q", seq)
	}
	got, parsed := term.ParseOSC52Response(seq)
	if !parsed {
		t.Error("ParseOSC52Response failed")
	}
	if got != text {
		t.Errorf("OSC52 round-trip failed: got %q, want %q", got, text)
	}
}

func TestLastBlockText_UserMessage(t *testing.T) {
	app := NewChatApp(80, 24)
	app.AddUserMessage("Hello from user!")

	text, ok := app.LastBlockText()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if text != "Hello from user!" {
		t.Errorf("got %q, want %q", text, "Hello from user!")
	}
}

func TestLastBlockText_ThinkingBlock(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := app.AddThinking()
	tb.AppendDelta("Hmm, let me think...")

	text, ok := app.LastBlockText()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if text != "Hmm, let me think..." {
		t.Errorf("got %q, want %q", text, "Hmm, let me think...")
	}
}

func TestLastBlockText_ToolCallAndResult(t *testing.T) {
	app := NewChatApp(80, 24)
	app.AddToolCall("grep", "-r pattern .")
	tr := app.AddToolResult()
	tr.AppendDelta("found 3 matches")

	// ToolResultBlock is last → should copy its output
	text, ok := app.LastBlockText()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if text != "found 3 matches" {
		t.Errorf("got %q, want %q", text, "found 3 matches")
	}
}

func TestLastBlockText_ToolCallFormat(t *testing.T) {
	app := NewChatApp(80, 24)
	app.AddToolCall("grep", "-r pattern .")

	text, ok := app.LastBlockText()
	if !ok {
		t.Fatal("expected ok=true")
	}
	// ToolCallBlock extracts as "name(args)"
	expected := "grep(-r pattern .)"
	if text != expected {
		t.Errorf("got %q, want %q", text, expected)
	}
}

func TestLastBlockText_LastNonEmpty(t *testing.T) {
	app := NewChatApp(80, 24)
	at := app.AddAssistantText()
	at.AppendDelta("First response")
	// Add an empty assistant text (should be skipped)
	app.AddAssistantText()

	text, ok := app.LastBlockText()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if text != "First response" {
		t.Errorf("got %q, want %q", text, "First response")
	}
}

func TestLastBlockText_AllEmpty(t *testing.T) {
	app := NewChatApp(80, 24)
	app.AddAssistantText() // empty

	text, ok := app.LastBlockText()
	if ok {
		t.Errorf("expected ok=false when all blocks empty")
	}
	if text != "" {
		t.Errorf("expected empty string, got %q", text)
	}
}

func TestCopyLastBlockOSC52(t *testing.T) {
	app := NewChatApp(80, 24)
	at := app.AddAssistantText()
	at.AppendDelta("Copy me to clipboard!")

	var sb strings.Builder
	ok := app.CopyLastBlockOSC52(&sb)
	if !ok {
		t.Fatal("expected ok=true")
	}
	output := sb.String()
	if !strings.HasPrefix(output, "\x1b]52;c;") {
		t.Errorf("expected OSC52 sequence, got %q", output)
	}
}

func TestCopyLastBlockOSC52_EmptyContainer(t *testing.T) {
	app := NewChatApp(80, 24)

	var sb strings.Builder
	ok := app.CopyLastBlockOSC52(&sb)
	if ok {
		t.Errorf("expected ok=false on empty container")
	}
	if sb.Len() != 0 {
		t.Errorf("expected no output, got %q", sb.String())
	}
}

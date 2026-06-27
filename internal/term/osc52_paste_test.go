package term

import (
	"encoding/base64"
	"testing"
)

func TestOSC52Paste_BELTerminator(t *testing.T) {
	parser := NewParser()

	// Simulate terminal OSC52 response: ESC ] 52 ; c ; <base64> BEL
	encoded := base64.StdEncoding.EncodeToString([]byte("Hello from clipboard!"))
	seq := []byte("\x1b]52;c;" + encoded + "\x07")

	events := parser.Feed(seq)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Type != EventClipboard {
		t.Errorf("type: got %d, want EventClipboard (%d)", events[0].Type, EventClipboard)
	}

	if events[0].Clipboard != "Hello from clipboard!" {
		t.Errorf("clipboard: got %q, want %q", events[0].Clipboard, "Hello from clipboard!")
	}
}

func TestOSC52Paste_STTerminator(t *testing.T) {
	parser := NewParser()

	// Simulate terminal OSC52 response: ESC ] 52 ; c ; <base64> ESC \
	encoded := base64.StdEncoding.EncodeToString([]byte("ST terminated"))
	seq := []byte("\x1b]52;c;" + encoded + "\x1b\\")

	events := parser.Feed(seq)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Type != EventClipboard {
		t.Errorf("type: got %d, want EventClipboard", events[0].Type)
	}

	if events[0].Clipboard != "ST terminated" {
		t.Errorf("clipboard: got %q, want %q", events[0].Clipboard, "ST terminated")
	}
}

func TestOSC52Paste_SplitAcrossReads(t *testing.T) {
	parser := NewParser()

	encoded := base64.StdEncoding.EncodeToString([]byte("Split data"))

	// First read: ESC ] 52 ; c ; <partial base64>
	part1 := []byte("\x1b]52;c;" + encoded[:5])
	events := parser.Feed(part1)
	if len(events) != 0 {
		t.Fatalf("expected 0 events from partial, got %d", len(events))
	}

	// Second read: rest of base64 + BEL
	part2 := []byte(encoded[5:] + "\x07")
	events = parser.Feed(part2)
	if len(events) != 1 {
		t.Fatalf("expected 1 event from completion, got %d", len(events))
	}

	if events[0].Type != EventClipboard {
		t.Errorf("type: got %d, want EventClipboard", events[0].Type)
	}
	if events[0].Clipboard != "Split data" {
		t.Errorf("clipboard: got %q, want %q", events[0].Clipboard, "Split data")
	}
}

func TestOSC52Paste_UnicodeContent(t *testing.T) {
	parser := NewParser()

	text := "你好世界 🌍"
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	seq := []byte("\x1b]52;c;" + encoded + "\x07")

	events := parser.Feed(seq)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Clipboard != text {
		t.Errorf("clipboard: got %q, want %q", events[0].Clipboard, text)
	}
}

func TestOSC52Paste_EmptyClipboard(t *testing.T) {
	parser := NewParser()

	// Empty clipboard response: ESC ] 52 ; c ; BEL
	seq := []byte("\x1b]52;c;\x07")

	events := parser.Feed(seq)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != EventClipboard {
		t.Errorf("type: got %d, want EventClipboard", events[0].Type)
	}
	if events[0].Clipboard != "" {
		t.Errorf("clipboard: got %q, want empty", events[0].Clipboard)
	}
}

func TestOSC52Paste_PrimarySelection(t *testing.T) {
	parser := NewParser()

	encoded := base64.StdEncoding.EncodeToString([]byte("primary selection"))
	seq := []byte("\x1b]52;p;" + encoded + "\x07")

	events := parser.Feed(seq)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != EventClipboard {
		t.Errorf("type: got %d, want EventClipboard", events[0].Type)
	}
	if events[0].Clipboard != "primary selection" {
		t.Errorf("clipboard: got %q, want %q", events[0].Clipboard, "primary selection")
	}
}

func TestOSC52Paste_NotOSC52(t *testing.T) {
	parser := NewParser()

	// Non-OSC52 OSC sequence (e.g., OSC 0 = set window title)
	// Should be silently ignored (no event)
	seq := []byte("\x1b]0;Window Title\x07")

	events := parser.Feed(seq)
	if len(events) != 0 {
		t.Fatalf("expected 0 events for non-OSC52, got %d: %+v", len(events), events)
	}
}

func TestOSC52Paste_StateResetsAfter(t *testing.T) {
	parser := NewParser()

	// Send OSC52 response, then a normal key
	encoded := base64.StdEncoding.EncodeToString([]byte("test"))
	seq := []byte("\x1b]52;c;" + encoded + "\x07")

	events := parser.Feed(seq)
	if len(events) != 1 || events[0].Type != EventClipboard {
		t.Fatalf("expected 1 clipboard event, got %+v", events)
	}

	// Now send a normal key — should work normally
	events = parser.Feed([]byte{'A'})
	if len(events) != 1 || events[0].Type != EventKey {
		t.Fatalf("expected 1 key event after OSC52, got %+v", events)
	}
	if events[0].Key == nil || events[0].Key.Rune != 'A' {
		t.Errorf("expected 'A', got %+v", events[0].Key)
	}
}

func TestOSC52Paste_LongContent(t *testing.T) {
	parser := NewParser()

	// 4096 bytes of content
	var long []byte
	for i := 0; i < 4096; i++ {
		long = append(long, byte('A'+(i%26)))
	}

	encoded := base64.StdEncoding.EncodeToString(long)
	seq := []byte("\x1b]52;c;" + encoded + "\x07")

	events := parser.Feed(seq)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if len(events[0].Clipboard) != 4096 {
		t.Errorf("clipboard length: got %d, want 4096", len(events[0].Clipboard))
	}
}

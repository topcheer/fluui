package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/termcompat"
)

func TestClipboardConfig_OSC52Enabled(t *testing.T) {
	caps := termcompat.Capabilities{
		Name:    "iTerm.app",
		HasOSC52: true,
	}
	cc := NewClipboardWithCapabilities(caps)

	if !cc.CanCopy() {
		t.Fatal("expected CanCopy()=true when OSC52 enabled")
	}

	seq, err := cc.Copy("hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(seq, "\x1b]52;c;") {
		t.Fatalf("expected OSC52 sequence to contain ESC]52;c;, got %q", seq)
	}

	// Should NOT contain tmux passthrough.
	if strings.Contains(seq, "\x1bPtmux;") {
		t.Fatal("expected no tmux passthrough wrapping")
	}
}

func TestClipboardConfig_OSC52Disabled(t *testing.T) {
	caps := termcompat.Capabilities{
		Name:    "GNOME Terminal",
		HasOSC52: false,
	}
	cc := NewClipboardWithCapabilities(caps)

	if cc.CanCopy() {
		t.Fatal("expected CanCopy()=false when OSC52 disabled")
	}

	_, err := cc.Copy("hello")
	if err == nil {
		t.Fatal("expected error when OSC52 disabled")
	}
	if err != ErrClipboardNotSupported {
		t.Fatalf("expected ErrClipboardNotSupported, got %v", err)
	}
}

func TestClipboardConfig_TmuxWrapping(t *testing.T) {
	caps := termcompat.Capabilities{
		Name:      "iTerm.app",
		HasOSC52:  true,
		InsideTmux: true,
	}
	cc := NewClipboardWithCapabilities(caps)

	if !cc.CanCopy() {
		t.Fatal("expected CanCopy()=true with OSC52 + tmux")
	}

	seq, err := cc.Copy("test data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(seq, "\x1bPtmux;") {
		t.Fatal("expected tmux passthrough wrapping")
	}

	// The inner content should still contain OSC52.
	if !strings.Contains(seq, "]52;c;") {
		t.Fatal("expected OSC52 sequence inside tmux passthrough")
	}
}

func TestClipboardConfig_ScreenDisablesOSC52(t *testing.T) {
	caps := termcompat.Capabilities{
		Name:         "GNU screen",
		HasOSC52:     false,
		InsideScreen: true,
	}
	cc := NewClipboardWithCapabilities(caps)

	// ShouldUseOSC52 returns false for screen even if HasOSC52 were true.
	if cc.CanCopy() {
		t.Fatal("expected CanCopy()=false inside GNU screen")
	}

	_, err := cc.Copy("test")
	if err == nil {
		t.Fatal("expected error for screen")
	}
}

func TestClipboardConfig_SetCapabilities(t *testing.T) {
	cc := NewClipboardConfig()

	// Initially no capabilities → CanCopy is false.
	if cc.CanCopy() {
		t.Fatal("expected CanCopy()=false with empty capabilities")
	}

	// Set capabilities.
	cc.SetCapabilities(termcompat.Capabilities{
		Name:    "kitty",
		HasOSC52: true,
	})

	if !cc.CanCopy() {
		t.Fatal("expected CanCopy()=true after SetCapabilities")
	}
}

func TestClipboardConfig_CopyToWriter(t *testing.T) {
	caps := termcompat.Capabilities{
		Name:    "WezTerm",
		HasOSC52: true,
	}
	cc := NewClipboardWithCapabilities(caps)

	var buf bytes.Buffer
	err := cc.CopyToWriter(&buf, "clipboard content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "\x1b]52;c;") {
		t.Fatalf("expected OSC52 sequence in output, got %q", buf.String())
	}
}

func TestClipboardConfig_SSHStillWorks(t *testing.T) {
	// SSH session with a capable terminal should still allow OSC52.
	caps := termcompat.Capabilities{
		Name:    "iTerm.app",
		HasOSC52: true,
		IsSSH:   true,
	}
	cc := NewClipboardWithCapabilities(caps)

	// SSH doesn't disable OSC52 — the terminal decides.
	if !cc.CanCopy() {
		t.Fatal("expected CanCopy()=true for SSH with OSC52-capable terminal")
	}

	seq, err := cc.Copy("remote data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(seq, "\x1b]52;c;") {
		t.Fatalf("expected OSC52 sequence, got %q", seq)
	}
}

func TestChatApp_SetClipboardCapabilities(t *testing.T) {
	app := NewChatApp(80, 24)

	// Initially nil.
	if app.ClipboardConfig() != nil {
		t.Fatal("expected nil ClipboardConfig initially")
	}

	// Set capabilities.
	caps := termcompat.Capabilities{
		Name:    "iTerm.app",
		HasOSC52: true,
	}
	app.SetClipboardCapabilities(caps)

	cc := app.ClipboardConfig()
	if cc == nil {
		t.Fatal("expected non-nil ClipboardConfig after SetClipboardCapabilities")
	}
	if !cc.CanCopy() {
		t.Fatal("expected CanCopy()=true")
	}
}

func TestChatApp_CopyLastBlockClipboard_WithConfig(t *testing.T) {
	app := NewChatApp(80, 24)

	// Add a block with content.
	at := app.AddAssistantText()
	at.AppendDelta("hello clipboard")

	// Set capabilities with OSC52 support.
	app.SetClipboardCapabilities(termcompat.Capabilities{
		Name:    "iTerm.app",
		HasOSC52: true,
	})

	var buf bytes.Buffer
	err := app.CopyLastBlockClipboard(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "\x1b]52;c;") {
		t.Fatalf("expected OSC52 sequence in output")
	}
}

func TestChatApp_CopyLastBlockClipboard_NoConfig_Fallback(t *testing.T) {
	app := NewChatApp(80, 24)

	at := app.AddAssistantText()
	at.AppendDelta("fallback test")

	// No config set — should fall back to raw OSC52.
	var buf bytes.Buffer
	err := app.CopyLastBlockClipboard(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "\x1b]52;c;") {
		t.Fatalf("expected raw OSC52 fallback")
	}
}

func TestChatApp_CopyLastBlockClipboard_Disabled(t *testing.T) {
	app := NewChatApp(80, 24)

	at := app.AddAssistantText()
	at.AppendDelta("blocked content")

	// Set capabilities WITHOUT OSC52.
	app.SetClipboardCapabilities(termcompat.Capabilities{
		Name:    "GNOME Terminal",
		HasOSC52: false,
	})

	var buf bytes.Buffer
	err := app.CopyLastBlockClipboard(&buf)
	if err != ErrClipboardNotSupported {
		t.Fatalf("expected ErrClipboardNotSupported, got %v", err)
	}
}

func TestChatApp_CopyLastBlockClipboard_NoContent(t *testing.T) {
	app := NewChatApp(80, 24)

	app.SetClipboardCapabilities(termcompat.Capabilities{
		Name:    "iTerm.app",
		HasOSC52: true,
	})

	var buf bytes.Buffer
	err := app.CopyLastBlockClipboard(&buf)
	if err == nil {
		t.Fatal("expected error for no content")
	}
}

func TestChatApp_CopyLastBlockClipboard_TmuxWrapping(t *testing.T) {
	app := NewChatApp(80, 24)

	at := app.AddAssistantText()
	at.AppendDelta("tmux test")

	app.SetClipboardCapabilities(termcompat.Capabilities{
		Name:      "iTerm.app",
		HasOSC52:  true,
		InsideTmux: true,
	})

	var buf bytes.Buffer
	err := app.CopyLastBlockClipboard(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "\x1bPtmux;") {
		t.Fatal("expected tmux passthrough wrapping in output")
	}
}

func TestLegacy_CopyLastBlockOSC52_StillWorks(t *testing.T) {
	app := NewChatApp(80, 24)

	at := app.AddAssistantText()
	at.AppendDelta("legacy method")

	var buf bytes.Buffer
	ok := app.CopyLastBlockOSC52(&buf)
	if !ok {
		t.Fatal("expected CopyLastBlockOSC52 to return true")
	}

	if !strings.Contains(buf.String(), "\x1b]52;c;") {
		t.Fatal("expected OSC52 sequence in output")
	}
}

func TestWrapTmuxPassthrough(t *testing.T) {
	input := "\x1b]52;c;aGVsbG8=\x1b\\"
	result := wrapTmuxPassthrough(input)

	if !strings.HasPrefix(result, "\x1bPtmux;") {
		t.Fatal("expected tmux passthrough prefix")
	}
	if !strings.HasSuffix(result, "\x1b\\\x1b\\") {
		t.Fatal("expected tmux passthrough suffix")
	}

	// Verify ESC bytes are doubled inside.
	inner := result[len("\x1bPtmux;") : len(result)-len("\x1b\\\x1b\\")]
	// Original input has 2 ESC bytes. Doubled → 4 ESC bytes in inner content.
	escCount := 0
	for i := 0; i < len(inner); i++ {
		if inner[i] == 0x1b {
			escCount++
		}
	}
	if escCount != 4 {
		t.Fatalf("expected 4 ESC bytes in tmux inner content, got %d", escCount)
	}
}

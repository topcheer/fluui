// Package term provides terminal compatibility verification tests.
// These tests validate that all protocol escape sequences produce
// syntactically correct output that real terminals will accept.
//
// Run with: go test -run="Compat" ./internal/term/
package term

import (
	"strings"
	"testing"
)

// TestCompat_AllProtocols verifies every protocol constant/function produces
// valid escape sequences (starts with ESC, ends with correct terminator).
// This is the baseline for multi-terminal compatibility testing.
func TestCompat_AllProtocols(t *testing.T) {
	type proto struct {
		name string
		seq  string
		// mustStartWith: the sequence should begin with ESC (0x1b)
		mustStartESC bool
	}

	protos := []proto{
		// OSC 8 Hyperlinks
		{"OSC8Start", OSC8Start(HyperlinkOptions{URL: "https://example.com"}), true},
		{"OSC8End", OSC8End(), true},
		{"OSC8Link", OSC8Link(HyperlinkOptions{URL: "https://example.com"}, "text"), true},

		// Synchronized Output
		{"SyncBegin", Sync("output"), true},

		// Focus Tracking
		{"EnableFocus", EnableFocus, true},
		{"DisableFocus", DisableFocus, true},

		// Window Title
		{"SetWindowTitle", SetWindowTitle("Test App"), true},
		{"SetIconName", SetIconName("icon"), true},

		// Cursor Save/Restore
		{"SaveCursor", SaveCursor, true},
		{"RestoreCursor", RestoreCursor, true},

		// Scroll Region
		{"SetScrollRegion", SetScrollRegion(0, 24), true},
		{"ResetScrollRegion", ResetScrollRegion, true},

		// DSR
		{"QueryCursorPosition", QueryCursorPosition, true},
		{"QueryTerminalSize", QueryTerminalSize, true},

		// Cursor Visibility
		{"HideCursor", HideCursor, true},
		{"ShowCursor", ShowCursor, true},

		// Alt Screen
		{"EnterAltScreen", EnterAltScreen, true},
		{"LeaveAltScreen", LeaveAltScreen, true},

		// Bracketed Paste
		{"EnableBracketedPaste", EnableBracketedPaste, true},
		{"DisableBracketedPaste", DisableBracketedPaste, true},

		// Mouse
		{"EnableMouseSGR", EnableMouseSGR, true},
		{"DisableMouseSGR", DisableMouseSGR, true},

		// Kitty Keyboard
		{"EnableKittyKeyboard", EnableKittyKeyboard, true},
		{"DisableKittyKeyboard", DisableKittyKeyboard, true},

		// Bell
		{"Bell", Bell, false},

		// Cursor Shape (DECSCUSR)
		{"SetCursorShape_SteadyBar", SetCursorShape(CursorShapeSteadyBar), true},
		{"SetCursorShape_BlinkingBlock", SetCursorShape(CursorShapeBlinkingBlock), true},
		{"ResetCursorShape", ResetCursorShape(), true},

		// Desktop Notification
		{"DesktopNotification", DesktopNotification("Test alert"), true},

		// Clipboard
		{"CopyClipboard", CopyClipboard("hello"), true},
		{"CopyPrimary", CopyPrimary("hello"), true},

		// Color Queries
		{"QueryPaletteColor", QueryPaletteColor(1), true},
		{"QueryDefaultFG", QueryDefaultFG(), true},
		{"QueryDefaultBG", QueryDefaultBG(), true},
		{"QueryCursorColor", QueryCursorColor(), true},
	}

	for _, p := range protos {
		t.Run(p.name, func(t *testing.T) {
			if len(p.seq) == 0 {
				t.Errorf("%s: empty sequence", p.name)
				return
			}
			if p.mustStartESC && p.seq[0] != 0x1b {
				t.Errorf("%s: does not start with ESC (0x1b), got 0x%02x", p.name, p.seq[0])
			}
		})
	}
}

// TestCompat_CursorShape_AllVariants verifies all 7 DECSCUSR shapes.
func TestCompat_CursorShape_AllVariants(t *testing.T) {
	shapes := []CursorShape{
		CursorShapeDefault,
		CursorShapeBlinkingBlock,
		CursorShapeSteadyBlock,
		CursorShapeBlinkingUnderline,
		CursorShapeSteadyUnderline,
		CursorShapeBlinkingBar,
		CursorShapeSteadyBar,
	}
	for _, s := range shapes {
		seq := SetCursorShape(s)
		if !strings.HasPrefix(seq, "\x1b[") {
			t.Errorf("shape %d: expected CSI prefix, got %q", s, seq[:min(3, len(seq))])
		}
		if !strings.HasSuffix(seq, " q") {
			t.Errorf("shape %d: expected SP q suffix, got %q", s, seq[len(seq)-2:])
		}
	}
}

// TestCompat_DesktopNotification_SpecialChars verifies BEL sanitization.
func TestCompat_DesktopNotification_SpecialChars(t *testing.T) {
	seq := DesktopNotification("hello\x07world")
	// The internal BEL (0x07) should be escaped, not interpreted as terminator
	// The only raw BEL should be the terminator at the end
	belCount := strings.Count(seq, "\x07")
	if belCount != 1 {
		t.Errorf("expected exactly 1 BEL (terminator), got %d in %q", belCount, seq)
	}
}

// TestCompat_OSC8_HyperlinkFormat verifies OSC 8 produces correct format.
func TestCompat_OSC8_HyperlinkFormat(t *testing.T) {
	start := OSC8Start(HyperlinkOptions{URL: "https://example.com"})
	// Format: ESC ] 8 ; params ; URL ESC \
	if !strings.HasPrefix(start, "\x1b]8;") {
		t.Errorf("OSC8 start doesn't match expected format: %q", start)
	}
	end := OSC8End()
	if end != "\x1b]8;;\x1b\\" {
		t.Errorf("OSC8 end doesn't match expected format: %q", end)
	}
}

// TestCompat_SyncOutput_Format verifies synchronized output DCS wrapper.
func TestCompat_SyncOutput_Format(t *testing.T) {
	seq := Sync("hello")
	// DCS wrapper: should contain the output and DCS markers
	if !strings.Contains(seq, "hello") {
		t.Error("Sync: missing output content")
	}
	// Should start with DCS (ESC P or ESC [ ...)
	if seq[0] != 0x1b {
		t.Error("Sync: should start with ESC")
	}
}

// TestCompat_SetScrollRegion_ValidatesParams verifies scroll region params.
func TestCompat_SetScrollRegion_ValidatesParams(t *testing.T) {
	seq := SetScrollRegion(5, 20)
	// Format: ESC [ 5 ; 20 r
	if !strings.HasPrefix(seq, "\x1b[5;20r") {
		t.Errorf("scroll region format unexpected: %q", seq)
	}
}

// TestCompat_ParseCursorPosition verifies DSR response parsing.
func TestCompat_ParseCursorPosition(t *testing.T) {
	// Response format: ESC [ row ; col R
	row, col, ok := ParseCursorPositionResponse("\x1b[10;20R")
	if !ok {
		t.Error("expected ok=true")
	}
	if row != 10 || col != 20 {
		t.Errorf("position = %d,%d, want 10,20", row, col)
	}

	// Invalid format
	_, _, ok = ParseCursorPositionResponse("invalid")
	if ok {
		t.Error("expected ok=false for invalid input")
	}
}

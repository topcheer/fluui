package term

import "testing"

// P307: Cover stripTerminator (dead code from P296, replaced by stripTerminatorChecked)
// and verify all new protocol constants

func TestP307_StripTerminator(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello\x07", "hello"},           // BEL
		{"world\x1b\\", "world"},          // ST (ESC \)
		{"no-terminator", "no-terminator"}, // no terminator
		{"\x07", ""},                      // just BEL
		{"", ""},                          // empty
	}
	for _, tt := range tests {
		got := stripTerminator(tt.input)
		if got != tt.want {
			t.Errorf("stripTerminator(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestP307_ProtocolConstantsExist(t *testing.T) {
	// Verify all 21 protocol constants are non-empty
	consts := []string{
		EnableBracketedPaste, DisableBracketedPaste,
		EnableMouseSGR, DisableMouseSGR,
		EnableFocus, DisableFocus,
		EnableTrueColor,
		EnableKittyKeyboard, DisableKittyKeyboard,
		EnterAltScreen, LeaveAltScreen,
		HideCursor, ShowCursor,
		Bell,
		SaveCursor, RestoreCursor,
		SaveCursorANSI, RestoreCursorANSI,
		ResetScrollRegion,
		QueryCursorPosition, QueryTerminalSize, QueryCellSize,
		QueryDA1, QueryDA2, QueryXTVersion,
	}
	for _, c := range consts {
		if c == "" {
			t.Error("empty protocol constant found")
		}
	}
}

func TestP307_CapabilitiesStruct(t *testing.T) {
	// Verify TerminalCapabilities has all expected fields
	caps := TerminalCapabilities{
		OSC52:           true,
		Sixel:           true,
		KittyGraphics:   true,
		Iterm2Images:    true,
		OSC8:            true,
		OSC133:          true,
		TrueColor:       true,
		Color256:        true,
		KittyKeyboard:   true,
		SGRMouse:        true,
		BracketedPaste:  true,
		FocusTracking:   true,
		Sync:            true,
		WindowTitles:    true,
		CursorPosition:  true,
		TerminalSize:    true,
		TerminalName:    "test",
	}
	if !caps.HasAny() {
		t.Error("expected HasAny=true for fully-featured caps")
	}
	if caps.TerminalName != "test" {
		t.Errorf("TerminalName = %q", caps.TerminalName)
	}
}

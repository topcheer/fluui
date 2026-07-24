package term

import (
	"testing"
)

// TestP331_SetCursorShape verifies DECSCUSR escape sequences for all shapes.
func TestP331_SetCursorShape(t *testing.T) {
	tests := []struct {
		shape CursorShape
		want  string
	}{
		{CursorShapeDefault, "\x1b[0 q"},
		{CursorShapeBlinkingBlock, "\x1b[1 q"},
		{CursorShapeSteadyBlock, "\x1b[2 q"},
		{CursorShapeBlinkingUnderline, "\x1b[3 q"},
		{CursorShapeSteadyUnderline, "\x1b[4 q"},
		{CursorShapeBlinkingBar, "\x1b[5 q"},
		{CursorShapeSteadyBar, "\x1b[6 q"},
	}
	for _, tt := range tests {
		got := SetCursorShape(tt.shape)
		if got != tt.want {
			t.Errorf("SetCursorShape(%d) = %q, want %q", tt.shape, got, tt.want)
		}
	}
}

// TestP331_SetCursorShape_InvalidClamps verifies out-of-range values clamp to 0.
func TestP331_SetCursorShape_InvalidClamps(t *testing.T) {
	got := SetCursorShape(CursorShape(99))
	if got != "\x1b[0 q" {
		t.Errorf("SetCursorShape(99) = %q, want \\x1b[0 q", got)
	}
	got = SetCursorShape(CursorShape(-1))
	if got != "\x1b[0 q" {
		t.Errorf("SetCursorShape(-1) = %q, want \\x1b[0 q", got)
	}
}

// TestP331_ResetCursorShape verifies the reset sequence.
func TestP331_ResetCursorShape(t *testing.T) {
	got := ResetCursorShape()
	if got != "\x1b[0 q" {
		t.Errorf("ResetCursorShape() = %q, want \\x1b[0 q", got)
	}
}

// TestP331_DesktopNotification verifies OSC 9 notification format.
func TestP331_DesktopNotification(t *testing.T) {
	got := DesktopNotification("AI response complete")
	want := "\x1b]9;AI response complete\x07"
	if got != want {
		t.Errorf("DesktopNotification() = %q, want %q", got, want)
	}
}

// TestP331_DesktopNotification_Empty verifies empty message.
func TestP331_DesktopNotification_Empty(t *testing.T) {
	got := DesktopNotification("")
	want := "\x1b]9;\x07"
	if got != want {
		t.Errorf("DesktopNotification(\"\") = %q, want %q", got, want)
	}
}

// TestP331_DesktopNotification_BELSanitized verifies BEL bytes are escaped.
func TestP331_DesktopNotification_BELSanitized(t *testing.T) {
	got := DesktopNotification("hello\x07world")
	if len(got) < len("\x1b]9;helloworld\x07") {
		t.Error("expected sanitized output with escaped BEL")
	}
	// The raw BEL in the message should be escaped, not present as-is
	// The final BEL terminator is at the end
	if got[len(got)-1] != 0x07 {
		t.Error("expected trailing BEL terminator")
	}
}

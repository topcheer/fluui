package term

import (
	"bytes"
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP126_MoveAndStyle_Default(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileTrue)
	s := buffer.DefaultStyle
	w.MoveAndStyle(5, 3, s)
	w.Flush()
	out := buf.String()
	// Should contain cursor move: ESC[4;6H (y+1=4, x+1=6)
	if !strings.Contains(out, "\x1b[4;6H") {
		t.Errorf("expected cursor move to 4;6, got %q", out)
	}
	// Default style should emit reset
	if !strings.Contains(out, "\x1b[0m") {
		t.Errorf("expected reset for default style, got %q", out)
	}
}

func TestP126_MoveAndStyle_TrueColor(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileTrue)
	s := buffer.Style{
		Fg: buffer.RGB(255, 128, 64),
		Bg: buffer.RGB(0, 0, 0),
	}
	w.MoveAndStyle(0, 0, s)
	w.Flush()
	out := buf.String()
	if !strings.Contains(out, "\x1b[1;1H") {
		t.Errorf("expected cursor move to 1;1, got %q", out)
	}
	// Should contain 38;2;255;128;64 for FG
	if !strings.Contains(out, "38;2;255;128;64") {
		t.Errorf("expected truecolor FG, got %q", out)
	}
}

func TestP126_MoveAndStyle_StyleUnchanged(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileTrue)
	s := buffer.Style{
		Fg: buffer.RGB(255, 0, 0),
	}
	w.MoveAndStyle(0, 0, s)
	w.Flush()
	buf.Reset()
	// Same style → should NOT emit SGR again
	w.MoveAndStyle(1, 0, s)
	w.Flush()
	out := buf.String()
	if strings.Contains(out, "38;2") {
		t.Errorf("should not emit SGR for unchanged style, got %q", out)
	}
	// Should still emit cursor move
	if !strings.Contains(out, "\x1b[1;2H") {
		t.Errorf("expected cursor move, got %q", out)
	}
}

func TestP126_MoveAndStyle_StyleChanged(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileTrue)
	w.MoveAndStyle(0, 0, buffer.Style{
		Fg: buffer.RGB(255, 0, 0),
	})
	w.Flush()
	buf.Reset()
	w.MoveAndStyle(1, 0, buffer.Style{
		Fg: buffer.RGB(0, 255, 0),
	})
	w.Flush()
	out := buf.String()
	if !strings.Contains(out, "38;2;0;255;0") {
		t.Errorf("expected new SGR for changed style, got %q", out)
	}
}

func TestP126_MoveAndStyle_BoldFlag(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileTrue)
	s := buffer.Style{Flags: buffer.Bold}
	w.MoveAndStyle(2, 2, s)
	w.Flush()
	out := buf.String()
	// Bold is SGR param 1. Output: \x1b[1;39;49m
	if !strings.Contains(out, "\x1b[1;") {
		t.Errorf("expected bold flag, got %q", out)
	}
}

func TestP126_MoveAndStyle_AfterResetStyle(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileTrue)
	// Set a style first
	w.MoveAndStyle(0, 0, buffer.Style{Fg: buffer.RGB(255, 0, 0)})
	w.Flush()
	// ResetStyle clears styleSet
	w.ResetStyle()
	w.Flush()
	buf.Reset()
	// Now MoveAndStyle with default should emit reset
	w.MoveAndStyle(0, 0, buffer.DefaultStyle)
	w.Flush()
	out := buf.String()
	if !strings.Contains(out, "\x1b[0m") {
		t.Errorf("expected reset after ResetStyle, got %q", out)
	}
}

func TestP126_MoveAndStyle_256Color(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, Profile256)
	s := buffer.Style{
		Fg: buffer.Color256Val(196), // red
	}
	w.MoveAndStyle(3, 4, s)
	w.Flush()
	out := buf.String()
	// Should contain 38;5;196
	if !strings.Contains(out, "38;5;196") {
		t.Errorf("expected 256-color FG, got %q", out)
	}
}

func TestP126_MoveAndStyle_LargeCoords(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileTrue)
	w.MoveAndStyle(200, 100, buffer.DefaultStyle)
	w.Flush()
	out := buf.String()
	// Should contain ESC[101;201H
	if !strings.Contains(out, "\x1b[101;201H") {
		t.Errorf("expected cursor at 101;201, got %q", out)
	}
}

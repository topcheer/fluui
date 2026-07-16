package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P222: ensureSelectedVisibleLocked + countVisibleLinesLocked + parseLipglossColor

func TestHelpOverlay_EnsureSelectedVisible_P222(t *testing.T) {
	h := NewHelpOverlay([]HelpGroup{{Name: "test", Entries: []HelpEntry{{Keys: "ctrl+a", Description: "action a"}}}})
	h.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	// Set many items so scrolling is needed
	
	buf := buffer.NewBuffer(40, 10)
	h.Paint(buf)
}

func TestHelpOverlay_EnsureSelectedVisibleSmall_P222(t *testing.T) {
	h := NewHelpOverlay([]HelpGroup{{Name: "test", Entries: []HelpEntry{{Keys: "ctrl+a", Description: "action a"}}}})
	h.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2}) // visibleH = 2-3 = -1, should return early
	h.SetSelected(0)
	buf := buffer.NewBuffer(40, 2)
	h.Paint(buf)
}

func TestRichLog_CountVisibleLines_P222(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	rl.Write(LogInfo, "test entry")
	rl.Write(LogWarn, "warn entry")
	rl.Write(LogDebug, "debug entry")
	buf := buffer.NewBuffer(60, 10)
	rl.Paint(buf)
}

func TestRichLog_CountVisibleLinesZeroWidth_P222(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 10})
	rl.Write(LogInfo, "test")
	buf := buffer.NewBuffer(0, 10)
	rl.Paint(buf)
}

func TestParseLipglossColor_Empty_P222(t *testing.T) {
	c := parseLipglossColor("")
	if c.Type != 0 {
		t.Error("empty should return zero color")
	}
}

func TestParseLipglossColor_Hex_P222(t *testing.T) {
	c := parseLipglossColor("#ff0000")
	if c.Type == 0 {
		t.Error("hex color should return non-zero")
	}
}

func TestParseLipglossColor_Named_P222(t *testing.T) {
	for _, name := range []string{"red", "green", "blue", "yellow", "cyan", "magenta", "white"} {
		c := parseLipglossColor(name)
		if c.Type == 0 {
			t.Errorf("named color %q should return non-zero", name)
		}
	}
}

func TestParseLipglossColor_Invalid_P222(t *testing.T) {
	c := parseLipglossColor("#zzzzz")
	// Invalid hex should fall through to named color, then default
	_ = c
}

func TestParseLipglossColor_UpperCase_P222(t *testing.T) {
	c := parseLipglossColor("RED")
	if c.Type == 0 {
		t.Error("uppercase named color should work")
	}
}
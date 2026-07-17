package component

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestRichLog_Measure_Fallback_P260(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	rl.Info("test line that is quite long")
	s := rl.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W != 80 {
		t.Errorf("fallback width should be 80, got %d", s.W)
	}
	if s.H < 1 {
		t.Error("fallback height should be at least 1")
	}
}

func TestRichLog_CountVisibleLines_ZeroBounds_P260(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	rl.Info("a")
	rl.Info("b")
	rl.Info("c")
	buf := buffer.NewBuffer(80, 5)
	rl.Paint(buf)
}

func TestRichLog_AutoScrollOff_P260(t *testing.T) {
	rl := NewRichLog()
	rl.SetAutoScroll(false)
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	for i := 0; i < 20; i++ {
		rl.Info("line")
	}
	buf := buffer.NewBuffer(40, 5)
	rl.Paint(buf)
}

func TestRichLog_PageUpDown_P260(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	for i := 0; i < 20; i++ {
		rl.Info("line")
	}
	rl.HandleKey(&term.KeyEvent{Key: term.KeyPageDown})
	rl.HandleKey(&term.KeyEvent{Key: term.KeyPageUp})
}

func TestRichLog_PageUpDown_ZeroBounds_P260(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	for i := 0; i < 5; i++ {
		rl.Info("x")
	}
	rl.HandleKey(&term.KeyEvent{Key: term.KeyPageDown})
	rl.HandleKey(&term.KeyEvent{Key: term.KeyPageUp})
}

func TestHelp_Measure_SmallConstraints_P260(t *testing.T) {
	groups := []HelpGroup{{Name: "test", Entries: []HelpEntry{{Keys: "Ctrl+S", Description: "Save"}}}}
	h := NewHelpOverlay(groups)
	s := h.Measure(Constraints{MaxWidth: 10, MaxHeight: 3})
	if s.W < 20 {
		t.Error("min width should be 20")
	}
	if s.H < 5 {
		t.Error("min height should be 5")
	}
}

func TestHelp_ScrollAdjust_Both_P260(t *testing.T) {
	entries := make([]HelpEntry, 30)
	for i := range entries {
		entries[i] = HelpEntry{Keys: "Ctrl+A", Description: "action"}
	}
	groups := []HelpGroup{{Name: "g", Entries: entries}}
	h := NewHelpOverlay(groups)
	h.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 8})
	h.ScrollDown(20)
	h.ScrollUp(25)
	buf := buffer.NewBuffer(40, 8)
	h.Paint(buf)
}

func TestHelp_KeyWidthClamp_P260(t *testing.T) {
	entries := make([]HelpEntry, 5)
	for i := range entries {
		entries[i] = HelpEntry{Keys: strings.Repeat("X", 50), Description: "desc"}
	}
	groups := []HelpGroup{{Name: "g", Entries: entries}}
	h := NewHelpOverlay(groups)
	h.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	h.Paint(buf)
}

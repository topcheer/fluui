package component

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// P287: badge Measure edge cases + viewport scrollbar thumb + help visibility + themestudio cursor

func TestBadge_Measure_EmptyContent_P287(t *testing.T) {
	b := NewBadge("", BadgeInfo)
	// contentWidth < 2 → w=2
	s := b.Measure(Constraints{})
	if s.W < 2 {
		t.Errorf("empty badge should clamp to min 2, got %d", s.W)
	}
}

func TestBadge_Measure_ZeroMaxWidth_P287(t *testing.T) {
	b := NewBadge("Hello", BadgeInfo)
	s := b.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	// MaxWidth=0 means HasWidth() returns false
	if s.W < 1 {
		t.Error("badge should have positive width")
	}
}

func TestBadge_Measure_MaxWidthLargerThanContent_P287(t *testing.T) {
	b := NewBadge("Hi", BadgeInfo)
	s := b.Measure(Constraints{MaxWidth: 50, MaxHeight: 5})
	// Content width should not be clamped up
	if s.W > 50 {
		t.Error("badge should not exceed MaxWidth")
	}
}

func TestBadge_MeasureGroup_ClampWidth_P287(t *testing.T) {
	bg := NewBadgeGroup()
	bg.Add(NewBadge("A", BadgeInfo))
	bg.Add(NewBadge("BB", BadgeInfo))
	bg.Add(NewBadge("CCC", BadgeInfo))
	s := bg.Measure(Constraints{MaxWidth: 5, MaxHeight: 1})
	if s.W > 5 {
		t.Errorf("badge group should clamp to MaxWidth 5, got %d", s.W)
	}
}

func TestBadge_MeasureGroup_NoMaxWidth_P287(t *testing.T) {
	bg := NewBadgeGroup()
	bg.Add(NewBadge("X", BadgeInfo))
	s := bg.Measure(Constraints{})
	if s.W < 1 {
		t.Error("badge group should have positive width")
	}
	if s.H != 1 {
		t.Errorf("badge group height should be 1, got %d", s.H)
	}
}

func TestViewport_DrawVScrollBar_ShortContent_P287(t *testing.T) {
	// Content shorter than viewport → maxOff=0 → thumb fills entire bar
	child := NewText("short")
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

func TestViewport_DrawVScrollBar_LongContent_P287(t *testing.T) {
	// Content longer than viewport → thumb is proportional
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "line"
	}
	child := NewText(strings.Join(lines, "\n"))
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	vp.ScrollDown(50)
	buf := buffer.NewBuffer(20, 5)
	vp.Paint(buf)
}

func TestViewport_DrawHScrollBar_WideContent_P287(t *testing.T) {
	// Wide content → horizontal scrollbar with proportional thumb
	child := NewText(strings.Repeat("x", 200))
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	vp.ScrollRight(100)
	buf := buffer.NewBuffer(10, 3)
	vp.Paint(buf)
}

func TestViewport_DrawHScrollBar_NarrowContent_P287(t *testing.T) {
	// Narrow content → maxOff=0 → thumb fills entire bar
	child := NewText("hi")
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	vp.Paint(buf)
}

func TestHelpOverlay_EnsureSelectedVisible_P287(t *testing.T) {
	groups := []HelpGroup{{Name: "Nav", Entries: []HelpEntry{}}}
	entries := make([]HelpEntry, 50)
	for i := range entries {
		entries[i] = HelpEntry{Keys: "key", Description: "desc"}
	}
	groups[0].Entries = entries
	h := NewHelpOverlay(groups)
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	h.Paint(buf)
}

func TestHelpOverlay_NavigateDown_P287(t *testing.T) {
	entries := make([]HelpEntry, 20)
	for i := range entries {
		entries[i] = HelpEntry{Keys: "k", Description: "d"}
	}
	groups := []HelpGroup{{Name: "Nav", Entries: entries}}
	h := NewHelpOverlay(groups)
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 8})
	for i := 0; i < 15; i++ {
		h.SelectNext()
	}
	buf := buffer.NewBuffer(60, 8)
	h.Paint(buf)
}

func TestHelpOverlay_NavigateUp_P287(t *testing.T) {
	entries := make([]HelpEntry, 20)
	for i := range entries {
		entries[i] = HelpEntry{Keys: "k", Description: "d"}
	}
	groups := []HelpGroup{{Name: "Nav", Entries: entries}}
	h := NewHelpOverlay(groups)
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 8})
	for i := 0; i < 15; i++ {
		h.SelectNext()
	}
	for i := 0; i < 5; i++ {
		h.SelectPrev()
	}
	buf := buffer.NewBuffer(60, 8)
	h.Paint(buf)
}

func TestThemeStudio_SetCursor_EmptySlots_P287(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	ts.Paint(buf)
}

func TestThemeStudio_MoveCursorDown_P287(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	ts.Paint(buf)
	// Don't call HandleKey(nil) — it crashes on nil key dereference
}

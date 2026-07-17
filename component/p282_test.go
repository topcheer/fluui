package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestBadge_Measure_ClampToMinWidth_P282(t *testing.T) {
	b := NewBadge("", BadgeInfo)
	s := b.Measure(Constraints{})
	if s.W < 2 {
		t.Errorf("badge width should clamp to min 2, got %d", s.W)
	}
}

func TestBadge_Measure_ClampToMaxWidth_P282(t *testing.T) {
	b := NewBadge("Hello World", BadgeInfo)
	s := b.Measure(Constraints{MaxWidth: 5, MaxHeight: 10})
	if s.W > 5 {
		t.Errorf("width should clamp to MaxWidth 5, got %d", s.W)
	}
}

func TestBadge_Measure_HeightAlways1_P282(t *testing.T) {
	b := NewBadge("Test", BadgeInfo)
	s := b.Measure(Constraints{MaxWidth: 50, MaxHeight: 10})
	if s.H != 1 {
		t.Errorf("badge height should always be 1, got %d", s.H)
	}
}

func TestBadge_Measure_TightConstraints_P282(t *testing.T) {
	b := NewBadge("X", BadgeInfo)
	s := b.Measure(Constraints{MaxWidth: 1, MaxHeight: 1})
	if s.W < 1 || s.H < 1 {
		t.Error("badge should be at least 1x1")
	}
}

func TestViewport_Paint_WithScroll_P282(t *testing.T) {
	child := NewText("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	vp.Paint(buf)
}

func TestViewport_Paint_ScrollDown_P282(t *testing.T) {
	child := NewText("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	vp.ScrollDown(3)
	buf := buffer.NewBuffer(20, 5)
	vp.Paint(buf)
}

func TestRichLog_Paint_MultiLine_P282(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	rl.Info("line1")
	rl.Info("line2")
	rl.Info("line3")
	buf := buffer.NewBuffer(40, 10)
	rl.Paint(buf)
}

func TestRichLog_Paint_LongLineWrap_P282(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	rl.Info("This is a very long line that should wrap or truncate")
	buf := buffer.NewBuffer(10, 5)
	rl.Paint(buf)
}

func TestRichLog_Paint_WithLevels_P282(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowLevels(true)
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	rl.Info("info msg")
	rl.Warn("warn msg")
	rl.Error("error msg")
	rl.Debug("debug msg")
	buf := buffer.NewBuffer(40, 10)
	rl.Paint(buf)
}

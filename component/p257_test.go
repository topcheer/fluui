package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P257: Badge Measure clamp + Paint overflow + BadgeGroup layout

func TestBadge_Measure_TinyContent_P257(t *testing.T) {
	b := NewBadge("", BadgeInfo)
	s := b.Measure(Constraints{})
	if s.W < 2 {
		t.Errorf("min width should be 2, got %d", s.W)
	}
}

func TestBadge_Measure_ClampMaxWidth_P257(t *testing.T) {
	b := NewBadge("hello world test", BadgeInfo)
	s := b.Measure(Constraints{MaxWidth: 5, MaxHeight: 0})
	if s.W > 5 {
		t.Errorf("width should be clamped to 5, got %d", s.W)
	}
	if s.H != 1 {
		t.Errorf("height should be 1, got %d", s.H)
	}
}

func TestBadge_Paint_WithIcon_P257(t *testing.T) {
	b := NewBadge("test", BadgeSuccess)
	b.SetIcon("OK")
	b.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.Paint(buf)
}

func TestBadge_Paint_NeutralVariant_P257(t *testing.T) {
	b := NewBadge("info", BadgeNeutral)
	b.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.Paint(buf)
}

func TestBadge_Paint_OtherVariant_P257(t *testing.T) {
	b := NewBadge("other", BadgeVariant(999))
	b.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.Paint(buf)
}


func TestBadgeGroup_Measure_P257(t *testing.T) {
	g := NewBadgeGroup()
	g.Add(NewBadge("a", BadgeInfo))
	g.Add(NewBadge("b", BadgeSuccess))
	s := g.Measure(Constraints{})
	if s.W <= 0 {
		t.Error("group width should be > 0")
	}
}

func TestBadgeGroup_Paint_Overflow_P257(t *testing.T) {
	g := NewBadgeGroup()
	g.Add(NewBadge("first badge", BadgeInfo))
	g.Add(NewBadge("second badge", BadgeSuccess))
	g.Add(NewBadge("third badge that overflows", BadgeWarning))
	g.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 1})
	buf := buffer.NewBuffer(15, 1)
	g.Paint(buf)
}

func TestBadgeGroup_Measure_ClampWidth_P257(t *testing.T) {
	g := NewBadgeGroup()
	g.Add(NewBadge("short", BadgeInfo))
	s := g.Measure(Constraints{MaxWidth: 3})
	if s.W > 3 {
		t.Errorf("should be clamped to 3, got %d", s.W)
	}
}

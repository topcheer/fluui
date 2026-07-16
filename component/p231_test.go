package component

import (
	"testing"
)

// P231: badge.Measure constraint branches + contentWidth < 2

func TestBadge_MeasureMaxWidth_P231(t *testing.T) {
	b := NewBadge("short", BadgeInfo)
	s := b.Measure(Constraints{MinWidth: 0, MaxWidth: 3, MinHeight: 0, MaxHeight: 10})
	if s.W > 3 {
		t.Errorf("width should be clamped to 3, got %d", s.W)
	}
}

func TestBadge_MeasureMaxHeight_P231(t *testing.T) {
	b := NewBadge("hello", BadgeInfo)
	s := b.Measure(Constraints{MinWidth: 0, MaxWidth: 100, MinHeight: 0, MaxHeight: 0})
	if s.H < 1 {
		t.Error("height should be at least 1")
	}
}

func TestBadge_MeasureEmptyText_P231(t *testing.T) {
	b := NewBadge("", BadgeInfo)
	s := b.Measure(Constraints{})
	if s.W < 1 {
		t.Error("empty badge should still have width >= 1")
	}
}

func TestBadge_MeasureTinyText_P231(t *testing.T) {
	b := NewBadge("x", BadgeInfo)
	s := b.Measure(Constraints{})
	if s.W < 2 {
		t.Error("single char badge should have width >= 2 (padding)")
	}
}

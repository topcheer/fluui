package component

import (
	"testing"
)

// P217: badge.Measure edge cases

func TestBadge_MeasureMaxHeightZero_P217(t *testing.T) {
	b := NewBadge("text", BadgeInfo)
	// MaxHeight=0 with HasHeight → h>MaxHeight(0) → h clamped to 1
	s := b.Measure(Constraints{MaxHeight: 0, HasHeight: true})
	if s.H != 1 {
		t.Errorf("expected height 1, got %d", s.H)
	}
}

func TestBadge_MeasureContentWidthOne_P217(t *testing.T) {
	b := NewBadge("x", BadgeSuccess)
	// Content width 1 → should be clamped to min 2
	s := b.Measure(Constraints{})
	if s.W < 2 {
		t.Errorf("expected width >= 2, got %d", s.W)
	}
}

func TestBadge_MeasureMaxWidthOne_P217(t *testing.T) {
	b := NewBadge("long text", BadgeWarning)
	// MaxWidth=1 → w should be clamped to 1
	s := b.Measure(Constraints{MaxWidth: 1, HasWidth: true})
	if s.W != 1 {
		t.Errorf("expected width 1, got %d", s.W)
	}
}

func TestBadge_MeasureNoConstraints_P217(t *testing.T) {
	b := NewBadge("ab", BadgeError)
	s := b.Measure(Constraints{})
	if s.H != 1 {
		t.Errorf("expected height 1, got %d", s.H)
	}
	if s.W < 2 {
		t.Errorf("expected width >= 2, got %d", s.W)
	}
}
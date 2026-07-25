package component

import (
	"testing"
)

// TestP357_Timeline_Measure_Defaults covers the maxW/maxH <= 0 branches.
func TestP357_Timeline_Measure_Defaults(t *testing.T) {
	tl := NewTimeline([]TimelineEvent{{Title: "A"}})
	s := tl.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W != 50 {
		t.Errorf("default width = %d, want 50", s.W)
	}
}

// TestP357_Timeline_Measure_Empty covers zero events.
func TestP357_Timeline_Measure_Empty(t *testing.T) {
	tl := NewTimeline(nil)
	s := tl.Measure(Constraints{MaxWidth: 40, MaxHeight: 10})
	if s.H != 1 {
		t.Errorf("empty height = %d, want 1", s.H)
	}
}

// TestP357_Timeline_Measure_Clamped covers height > maxH clamping.
func TestP357_Timeline_Measure_Clamped(t *testing.T) {
	events := make([]TimelineEvent, 20)
	for i := range events {
		events[i] = TimelineEvent{Title: "E"}
	}
	tl := NewTimeline(events)
	s := tl.Measure(Constraints{MaxWidth: 50, MaxHeight: 5})
	if s.H != 5 {
		t.Errorf("clamped height = %d, want 5", s.H)
	}
}

package animation

import (
	"testing"
	"time"
)

// P197: Coverage for animation sub-80% functions

func TestSpinner_Current_P197(t *testing.T) {
	s := NewSpinner("dots")
	// Normal case
	cur := s.Current()
	if cur == "" {
		t.Error("non-empty spinner should return frame")
	}

	// Empty frames case
	s2 := Spinner{frames: []string{}}
	if s2.Current() != "" {
		t.Error("empty spinner should return empty string")
	}
}

func TestSpinner_UpdateFullCycle_P197(t *testing.T) {
	s := NewSpinner("dots")
	initial := s.Current()
	// Update multiple times to cycle through frames
	for i := 0; i < 100; i++ {
		s.Update(100 * time.Millisecond)
	}
	if s.Current() == "" {
		t.Error("should still have a frame after cycling")
	}
	_ = initial
}

func TestSpinner_UpdateLargeDelta_P197(t *testing.T) {
	s := NewSpinner("dots")
	// Very large delta to exercise multi-advance in one call
	s.Update(10 * time.Second)
}
package fuzzy

import (
	"testing"
)

func TestHighlight_NilResult_P269(t *testing.T) {
	var r *Result
	segs := r.Highlight()
	if segs != nil {
		t.Error("nil result should return nil segments")
	}
}

func TestHighlight_EmptyPositions_P269(t *testing.T) {
	r := &Result{
		Item:     "hello",
		Positions: []int{},
	}
	segs := r.Highlight()
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	if segs[0].Matched {
		t.Error("segment should be unmatched")
	}
	if segs[0].Text != "hello" {
		t.Errorf("expected 'hello', got %q", segs[0].Text)
	}
}

func TestHighlight_WithPositions_P269(t *testing.T) {
	r := &Result{
		Item:     "hello",
		Positions: []int{0, 1}, // "he" matched
	}
	segs := r.Highlight()
	if len(segs) < 2 {
		t.Fatalf("expected at least 2 segments, got %d", len(segs))
	}
	// First segment should be matched "he"
	if !segs[0].Matched {
		t.Error("first segment should be matched")
	}
	if segs[0].Text != "he" {
		t.Errorf("expected 'he', got %q", segs[0].Text)
	}
}

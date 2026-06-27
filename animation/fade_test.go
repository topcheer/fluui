package animation

import (
	"testing"
	"time"
)

func TestFadeInProgress(t *testing.T) {
	f := NewFadeIn(200 * time.Millisecond)
	if f.Progress() != 0 {
		t.Fatalf("expected initial progress 0, got %f", f.Progress())
	}

	finished := f.Update(50 * time.Millisecond)
	if finished {
		t.Fatal("should not be finished after partial update")
	}
	if f.Progress() <= 0 || f.Progress() >= 1 {
		t.Errorf("expected progress between 0 and 1, got %f", f.Progress())
	}

	// After another chunk, progress should increase.
	before := f.Progress()
	f.Update(50 * time.Millisecond)
	if f.Progress() <= before {
		t.Errorf("expected progress to increase, was %f now %f", before, f.Progress())
	}
}

func TestFadeInComplete(t *testing.T) {
	f := NewFadeIn(100 * time.Millisecond)

	// Exceed duration in one step.
	finished := f.Update(200 * time.Millisecond)
	if !finished {
		t.Fatal("should be finished after exceeding duration")
	}
	if f.Progress() != 1 {
		t.Errorf("expected progress 1, got %f", f.Progress())
	}

	// Done channel should be closed.
	select {
	case <-f.Done():
		// expected
	default:
		t.Fatal("Done channel should be closed after completion")
	}
}

func TestFadeInProgressClampedTo1(t *testing.T) {
	f := NewFadeIn(50 * time.Millisecond)
	f.Update(10 * time.Millisecond)
	if f.Progress() > 1 {
		t.Errorf("progress should not exceed 1, got %f", f.Progress())
	}
	f.Update(100 * time.Millisecond)
	if f.Progress() != 1 {
		t.Errorf("progress should be clamped to 1, got %f", f.Progress())
	}
}

func TestFadeInUpdateAfterDone(t *testing.T) {
	f := NewFadeIn(50 * time.Millisecond)
	f.Update(100 * time.Millisecond) // completes
	finished := f.Update(50 * time.Millisecond)
	if !finished {
		t.Error("Update after completion should still return true")
	}
	if f.Progress() != 1 {
		t.Errorf("progress should stay at 1, got %f", f.Progress())
	}
}

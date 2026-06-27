package animation

import (
	"testing"
	"time"
)

func TestSpinnerFrames(t *testing.T) {
	for _, style := range []string{"dots", "arc", "arrow", "bouncing"} {
		frames, ok := SpinnerFrames[style]
		if !ok {
			t.Fatalf("missing frames for style %q", style)
		}
		if len(frames) == 0 {
			t.Fatalf("style %q has empty frames", style)
		}
		s := NewSpinner(style)
		if len(s.frames) != len(frames) {
			t.Errorf("style %q: expected %d frames, got %d", style, len(frames), len(s.frames))
		}
	}

	// Unknown style falls back to dots.
	s := NewSpinner("nonexistent")
	if len(s.frames) != len(SpinnerFrames["dots"]) {
		t.Errorf("unknown style should fall back to dots")
	}
}

func TestSpinnerCurrent(t *testing.T) {
	s := NewSpinner("dots")
	cur := s.Current()
	if cur == "" {
		t.Fatal("Current() should return a non-empty string")
	}
	if cur != SpinnerFrames["dots"][0] {
		t.Errorf("expected first frame %q, got %q", SpinnerFrames["dots"][0], cur)
	}
}

func TestSpinnerUpdate(t *testing.T) {
	s := NewSpinner("dots")
	initial := s.Current()

	// Advancing by the interval should move to the next frame.
	finished := s.Update(s.interval)
	if finished {
		t.Fatal("spinner should never finish")
	}
	if s.Current() == initial {
		t.Error("expected frame to change after one interval")
	}
	if s.Current() != SpinnerFrames["dots"][1] {
		t.Errorf("expected second frame %q, got %q", SpinnerFrames["dots"][1], s.Current())
	}
}

func TestSpinnerUpdateMultipleFrames(t *testing.T) {
	s := NewSpinner("arc")
	frames := SpinnerFrames["arc"]

	// Advance several intervals at once.
	s.Update(s.interval * 3)
	if s.current != 3%len(frames) {
		t.Errorf("expected current index %d, got %d", 3%len(frames), s.current)
	}
}

func TestSpinnerWraparound(t *testing.T) {
	s := NewSpinner("arrow")
	frames := SpinnerFrames["arrow"]

	// Advance exactly len(frames) intervals — should wrap to index 0.
	s.Update(s.interval * time.Duration(len(frames)))
	if s.current != 0 {
		t.Errorf("expected wraparound to index 0, got %d", s.current)
	}
}

func TestSpinnerNeverCompletes(t *testing.T) {
	s := NewSpinner("dots")
	for i := 0; i < 1000; i++ {
		if s.Update(time.Second) {
			t.Fatal("spinner completed after many updates — should run forever")
		}
	}
	// Done channel should never be closed.
	select {
	case <-s.Done():
		t.Fatal("Done channel should never close for spinner")
	default:
	}
}

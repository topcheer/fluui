package animation

import (
	"testing"
	"time"
)

// P29b coverage tests for Lerp, EasedProgress, FadeOut, Spinner.Current, Manager lifecycle.

func TestP29b_Lerp(t *testing.T) {
	if Lerp(0, 10, 0) != 0 {
		t.Error("Lerp(0,10,0) should be 0")
	}
	if Lerp(0, 10, 1) != 10 {
		t.Error("Lerp(0,10,1) should be 10")
	}
	if Lerp(0, 10, 0.5) != 5 {
		t.Error("Lerp(0,10,0.5) should be 5")
	}
	// Negative range
	if Lerp(10, 0, 0.5) != 5 {
		t.Error("Lerp(10,0,0.5) should be 5")
	}
}

func TestP29b_EasedProgress_NilEasing(t *testing.T) {
	v := EasedProgress(0.5, nil)
	if v != 0.5 {
		t.Errorf("nil easing should use Linear, expected 0.5, got %f", v)
	}
}

func TestP29b_EasedProgress_WithEasing(t *testing.T) {
	v := EasedProgress(0.5, EaseIn)
	// EaseIn(0.5) = 0.25
	if v < 0.2 || v > 0.3 {
		t.Errorf("EaseIn at 0.5 should be ~0.25, got %f", v)
	}
}

func TestP29b_EasedProgress_ClampsInput(t *testing.T) {
	// With easing: clamps input
	v := EasedProgress(-1, EaseIn)
	if v != 0 {
		t.Errorf("negative input with easing should clamp to 0, got %f", v)
	}
	v = EasedProgress(2, EaseIn)
	if v != 1 {
		t.Errorf("input > 1 with easing should clamp to 1, got %f", v)
	}
}

// === FadeOut lifecycle ===

func TestP29b_FadeOut_Create(t *testing.T) {
	f := NewFadeOut(100 * time.Millisecond)
	if f == nil {
		t.Fatal("NewFadeOut should return non-nil")
	}
	if f.Progress() != 1 {
		t.Errorf("FadeOut should start at progress 1, got %f", f.Progress())
	}
	select {
	case <-f.Done():
		t.Error("FadeOut should not be done immediately")
	default:
	}
}

func TestP29b_FadeOut_PartialUpdate(t *testing.T) {
	f := NewFadeOut(100 * time.Millisecond)
	f.Update(50 * time.Millisecond)
	p := f.Progress()
	if p <= 0 || p >= 1 {
		t.Errorf("after 50ms of 100ms, progress should be between 0 and 1, got %f", p)
	}
}

func TestP29b_FadeOut_Complete(t *testing.T) {
	f := NewFadeOut(50 * time.Millisecond)
	done := f.Update(60 * time.Millisecond)
	if !done {
		t.Error("should be done after exceeding duration")
	}
	if f.Progress() != 0 {
		t.Errorf("progress should be 0 when done, got %f", f.Progress())
	}
	select {
	case <-f.Done():
	default:
		t.Error("Done channel should be closed")
	}
}

func TestP29b_FadeOut_UpdateAfterDone(t *testing.T) {
	f := NewFadeOut(10 * time.Millisecond)
	f.Update(20 * time.Millisecond) // completes
	done := f.Update(10 * time.Millisecond) // should return true immediately
	if !done {
		t.Error("Update after done should return true")
	}
}

func TestP29b_FadeOut_ZeroDuration(t *testing.T) {
	f := NewFadeOut(0)
	done := f.Update(1 * time.Millisecond)
	if !done {
		t.Error("zero-duration fade should complete immediately")
	}
}

// === Spinner.Current ===

func TestP29b_Spinner_Current(t *testing.T) {
	s := NewSpinner("dots")
	c := s.Current()
	if c == "" {
		t.Error("Current should return non-empty for valid style")
	}
}

func TestP29b_Spinner_CurrentAfterUpdate(t *testing.T) {
	s := NewSpinner("dots")
	first := s.Current()
	s.Update(100 * time.Millisecond) // advance one frame
	second := s.Current()
	if first == second {
		// Could be same if only 1 frame, but dots has 10
		t.Error("frame should change after Update")
	}
}

func TestP29b_Spinner_CycleFrames(t *testing.T) {
	s := NewSpinner("dots")
	frames := map[string]bool{}
	for i := 0; i < 20; i++ {
		s.Update(100 * time.Millisecond)
		frames[s.Current()] = true
	}
	if len(frames) < 3 {
		t.Errorf("expected at least 3 distinct frames, got %d", len(frames))
	}
}

// === Manager lifecycle ===

func TestP29b_Manager_DefaultFPS(t *testing.T) {
	m := NewManager(0, nil) // fps=0 → default 60
	if m == nil {
		t.Fatal("should create manager with default fps")
	}
}

func TestP29b_Manager_AddFadeIn(t *testing.T) {
	m := NewManager(60, nil)
	f := NewFadeIn(50 * time.Millisecond)
	m.Add(f)
	m.Tick()
	// Should not panic
}

func TestP29b_Manager_AddFadeOut(t *testing.T) {
	m := NewManager(60, nil)
	f := NewFadeOut(50 * time.Millisecond)
	m.Add(f)
	m.Tick()
	m.Tick()
}

func TestP29b_Manager_RemovesFinished(t *testing.T) {
	m := NewManager(60, nil)
	f := NewFadeOut(10 * time.Millisecond)
	m.Add(f)
	// Tick multiple times to finish
	for i := 0; i < 5; i++ {
		m.Tick()
	}
}

func TestP29b_Manager_OnDirty(t *testing.T) {
	dirty := false
	m := NewManager(60, func() { dirty = true })
	f := NewFadeIn(1000 * time.Millisecond)
	m.Add(f)
	m.Tick()
	if !dirty {
		t.Error("onDirty should be called when animation is still active")
	}
}

func TestP29b_Manager_StartStop(t *testing.T) {
	m := NewManager(60, func() {})
	m.Start()
	time.Sleep(20 * time.Millisecond)
	m.Stop()
	// Should not panic
}

func TestP29b_Manager_StartTwice(t *testing.T) {
	m := NewManager(60, func() {})
	m.Start()
	m.Start() // should be no-op
	m.Stop()
}

func TestP29b_Manager_StopTwice(t *testing.T) {
	m := NewManager(60, func() {})
	m.Start()
	m.Stop()
	m.Stop() // should be safe
}

func TestP29b_Manager_StopWithoutStart(t *testing.T) {
	m := NewManager(60, nil)
	m.Stop() // should be safe
}

func TestP29b_Manager_TickEmpty(t *testing.T) {
	m := NewManager(60, nil)
	m.Tick() // should be safe with no animations
}

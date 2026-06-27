package animation

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestManagerAdd(t *testing.T) {
	m := NewManager(60, nil)
	f := NewFadeIn(100 * time.Millisecond)
	m.Add(f)

	if len(m.animations) != 1 {
		t.Fatalf("expected 1 animation, got %d", len(m.animations))
	}

	// Add a spinner (never completes) too.
	s := NewSpinner("dots")
	m.Add(s)
	if len(m.animations) != 2 {
		t.Fatalf("expected 2 animations, got %d", len(m.animations))
	}
}

func TestManagerTick(t *testing.T) {
	m := NewManager(60, nil)
	f := NewFadeIn(m.interval * 5) // completes after 5 ticks
	m.Add(f)

	// Tick a few times — not yet complete.
	for i := 0; i < 4; i++ {
		m.Tick()
	}
	select {
	case <-f.Done():
		t.Fatal("fade should not be done after 4 ticks")
	default:
	}

	// 5th tick completes it.
	m.Tick()
	select {
	case <-f.Done():
		// expected
	default:
		t.Fatal("fade should be done after 5 ticks")
	}

	// Finished animations get removed.
	if len(m.animations) != 0 {
		t.Errorf("expected 0 animations after completion, got %d", len(m.animations))
	}
}

func TestManagerTickRemovesFinishedOnly(t *testing.T) {
	m := NewManager(60, nil)
	fade := NewFadeIn(m.interval) // completes after 1 tick
	spinner := NewSpinner("dots")  // never completes
	m.Add(fade)
	m.Add(spinner)

	m.Tick()

	// Spinner should remain.
	if len(m.animations) != 1 {
		t.Fatalf("expected 1 animation remaining, got %d", len(m.animations))
	}
	if m.animations[0] != spinner {
		t.Error("remaining animation should be the spinner")
	}
}

func TestManagerOnDirty(t *testing.T) {
	var dirty atomic.Int32
	m := NewManager(60, func() {
		dirty.Add(1)
	})

	// Spinner never completes → always triggers onDirty.
	m.Add(NewSpinner("dots"))
	m.Tick()
	m.Tick()
	m.Tick()

	if dirty.Load() != 3 {
		t.Errorf("expected onDirty called 3 times, got %d", dirty.Load())
	}
}

func TestManagerStartStop(t *testing.T) {
	var dirty atomic.Int32
	m := NewManager(120, func() {
		dirty.Add(1)
	})

	m.Add(NewSpinner("dots"))
	m.Start()

	// Let the background ticker run for a bit.
	time.Sleep(80 * time.Millisecond)
	m.Stop()

	count := dirty.Load()
	if count == 0 {
		t.Fatal("background ticker should have called onDirty at least once")
	}

	// After Stop, dirty should not keep increasing.
	time.Sleep(50 * time.Millisecond)
	if dirty.Load() != count {
		t.Error("onDirty should not fire after Stop")
	}
}

func TestManagerStopIdempotent(t *testing.T) {
	m := NewManager(60, nil)
	m.Start()
	m.Stop()
	m.Stop() // should not panic
}

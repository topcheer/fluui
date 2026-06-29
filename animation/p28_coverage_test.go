package animation

import (
	"math"
	"testing"
	"time"
)

const eps = 1e-9

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < eps
}

// --- Easing function tests ---

func TestP28_Linear(t *testing.T) {
	if !approxEqual(Linear(0), 0) {
		t.Error("Linear(0) should be 0")
	}
	if !approxEqual(Linear(1), 1) {
		t.Error("Linear(1) should be 1")
	}
	if !approxEqual(Linear(0.5), 0.5) {
		t.Error("Linear(0.5) should be 0.5")
	}
}

func TestP28_EaseIn(t *testing.T) {
	if !approxEqual(EaseIn(0), 0) {
		t.Error("EaseIn(0) should be 0")
	}
	if !approxEqual(EaseIn(1), 1) {
		t.Error("EaseIn(1) should be 1")
	}
	v := EaseIn(0.5)
	if v <= 0 || v >= 1 {
		t.Errorf("EaseIn(0.5) should be between 0 and 1, got %f", v)
	}
}

func TestP28_EaseOut(t *testing.T) {
	if !approxEqual(EaseOut(0), 0) {
		t.Error("EaseOut(0) should be 0")
	}
	if !approxEqual(EaseOut(1), 1) {
		t.Error("EaseOut(1) should be 1")
	}
	v := EaseOut(0.5)
	if v <= 0 || v >= 1 {
		t.Errorf("EaseOut(0.5) should be between 0 and 1, got %f", v)
	}
}

func TestP28_EaseInOut(t *testing.T) {
	if !approxEqual(EaseInOut(0), 0) {
		t.Error("EaseInOut(0) should be 0")
	}
	if !approxEqual(EaseInOut(1), 1) {
		t.Error("EaseInOut(1) should be 1")
	}
	if !approxEqual(EaseInOut(0.5), 0.5) {
		t.Error("EaseInOut(0.5) should be 0.5")
	}
}

func TestP28_EaseInCubic(t *testing.T) {
	if !approxEqual(EaseInCubic(0), 0) {
		t.Error("EaseInCubic(0) should be 0")
	}
	if !approxEqual(EaseInCubic(1), 1) {
		t.Error("EaseInCubic(1) should be 1")
	}
}

func TestP28_EaseOutCubic(t *testing.T) {
	if !approxEqual(EaseOutCubic(0), 0) {
		t.Error("EaseOutCubic(0) should be 0")
	}
	if !approxEqual(EaseOutCubic(1), 1) {
		t.Error("EaseOutCubic(1) should be 1")
	}
}

func TestP28_EaseInOutCubic(t *testing.T) {
	if !approxEqual(EaseInOutCubic(0), 0) {
		t.Error("EaseInOutCubic(0) should be 0")
	}
	if !approxEqual(EaseInOutCubic(1), 1) {
		t.Error("EaseInOutCubic(1) should be 1")
	}
	if !approxEqual(EaseInOutCubic(0.5), 0.5) {
		t.Error("EaseInOutCubic(0.5) should be 0.5")
	}
}

func TestP28_EaseOutBack(t *testing.T) {
	if !approxEqual(EaseOutBack(0), 0) {
		t.Errorf("EaseOutBack(0) should be ~0, got %f", EaseOutBack(0))
	}
	if !approxEqual(EaseOutBack(1), 1) {
		t.Errorf("EaseOutBack(1) should be ~1, got %f", EaseOutBack(1))
	}
	v := EaseOutBack(0.8)
	if v < 0 {
		t.Errorf("EaseOutBack(0.8) should be >= 0, got %f", v)
	}
}

func TestP28_EaseInBack(t *testing.T) {
	if !approxEqual(EaseInBack(0), 0) {
		t.Errorf("EaseInBack(0) should be ~0, got %f", EaseInBack(0))
	}
	if !approxEqual(EaseInBack(1), 1) {
		t.Errorf("EaseInBack(1) should be ~1, got %f", EaseInBack(1))
	}
	v := EaseInBack(0.2)
	if v > 1 {
		t.Errorf("EaseInBack(0.2) should be <= 1, got %f", v)
	}
}

func TestP28_EaseOutElastic(t *testing.T) {
	if !approxEqual(EaseOutElastic(0), 0) {
		t.Error("EaseOutElastic(0) should be 0")
	}
	if !approxEqual(EaseOutElastic(1), 1) {
		t.Error("EaseOutElastic(1) should be 1")
	}
	v := EaseOutElastic(0.3)
	if math.IsNaN(v) {
		t.Error("EaseOutElastic(0.3) should not be NaN")
	}
}

func TestP28_EaseOutBounce(t *testing.T) {
	if !approxEqual(EaseOutBounce(0), 0) {
		t.Error("EaseOutBounce(0) should be 0")
	}
	if !approxEqual(EaseOutBounce(1), 1) {
		t.Error("EaseOutBounce(1) should be 1")
	}
	v := EaseOutBounce(0.5)
	if v < 0 || v > 1 {
		t.Errorf("EaseOutBounce(0.5) should be 0..1, got %f", v)
	}
}

func TestP28_EaseInBounce(t *testing.T) {
	if !approxEqual(EaseInBounce(0), 0) {
		t.Error("EaseInBounce(0) should be 0")
	}
	if !approxEqual(EaseInBounce(1), 1) {
		t.Error("EaseInBounce(1) should be 1")
	}
	v := EaseInBounce(0.5)
	if v < 0 || v > 1 {
		t.Errorf("EaseInBounce(0.5) should be 0..1, got %f", v)
	}
}

func TestP28_Clamp(t *testing.T) {
	if !approxEqual(Clamp(-1), 0) {
		t.Error("Clamp(-1) should be 0")
	}
	if !approxEqual(Clamp(2), 1) {
		t.Error("Clamp(2) should be 1")
	}
	if !approxEqual(Clamp(0.5), 0.5) {
		t.Error("Clamp(0.5) should be 0.5")
	}
}

// --- Monotonicity checks ---

func TestP28_EasingMonotonic(t *testing.T) {
	easings := []struct {
		name string
		fn   Easing
	}{
		{"Linear", Linear},
		{"EaseIn", EaseIn},
		{"EaseOut", EaseOut},
		{"EaseInCubic", EaseInCubic},
		{"EaseOutCubic", EaseOutCubic},
	}
	for _, e := range easings {
		t.Run(e.name, func(t *testing.T) {
			prev := e.fn(0)
			for i := 1; i <= 100; i++ {
				curr := e.fn(float64(i) / 100)
				if curr < prev-0.001 {
					t.Errorf("%s not monotonic: %f > %f at step %d", e.name, prev, curr, i)
				}
				prev = curr
			}
		})
	}
}

// --- Animation lifecycle ---

func TestP28_ManagerAddRemove(t *testing.T) {
	m := NewManager(60, func() {})
	if m == nil {
		t.Fatal("NewManager should return non-nil")
	}
	m.Tick()
}

func TestP28_FadeInDone(t *testing.T) {
	f := NewFadeIn(100 * time.Millisecond)
	if f == nil {
		t.Fatal("NewFadeIn should return non-nil")
	}
	select {
	case <-f.Done():
		t.Error("FadeIn should not be done immediately")
	default:
	}
	for i := 0; i < 200; i++ {
		f.Update(time.Millisecond)
	}
	select {
	case <-f.Done():
	default:
		t.Error("FadeIn should be done after 200ms")
	}
}

func TestP28_SpinnerFrames(t *testing.T) {
	s := NewSpinner("dots")
	if s == nil {
		t.Fatal("NewSpinner should return non-nil")
	}
	select {
	case <-s.Done():
		t.Error("Spinner should never be done")
	default:
	}
	s.Update(100 * time.Millisecond)
	s.Update(100 * time.Millisecond)
}

package animation

import (
	"math"
	"testing"
)

// P237: cover easing custom-param branches + bounce segments

func TestEaseOutBack_CustomBack_P237(t *testing.T) {
	// Custom back param → c1 = 2.0
	r := EaseOutBack(0.5, 2.0)
	if math.IsNaN(r) {
		t.Error("EaseOutBack with custom back should not be NaN")
	}
	// Verify custom param is actually used (different from default)
	rDefault := EaseOutBack(0.5)
	if r == rDefault {
		t.Error("custom back param should produce different result")
	}
}

func TestEaseInBack_CustomBack_P237(t *testing.T) {
	r := EaseInBack(0.5, 3.0)
	if math.IsNaN(r) {
		t.Error("EaseInBack with custom back should not be NaN")
	}
	rDefault := EaseInBack(0.5)
	if r == rDefault {
		t.Error("custom back param should produce different result")
	}
}

func TestEaseOutBounce_AllSegments_P237(t *testing.T) {
	// Test all 4 bounce segments with specific t values
	// Segment 1: t < 1/2.75 ≈ 0.364
	r1 := EaseOutBounce(0.2)
	if r1 <= 0 || r1 > 1 {
		t.Errorf("segment 1: EaseOutBounce(0.2) = %f, want (0,1]", r1)
	}
	// Segment 2: 1/2.75 <= t < 2/2.75 ≈ 0.727
	r2 := EaseOutBounce(0.5)
	if r2 < 0 || r2 > 1 {
		t.Errorf("segment 2: EaseOutBounce(0.5) = %f, want [0,1]", r2)
	}
	// Segment 3: 2/2.75 <= t < 2.5/2.75 ≈ 0.909
	r3 := EaseOutBounce(0.85)
	if r3 < 0 || r3 > 1 {
		t.Errorf("segment 3: EaseOutBounce(0.85) = %f, want [0,1]", r3)
	}
	// Segment 4: t >= 2.5/2.75
	r4 := EaseOutBounce(0.95)
	if r4 < 0 || r4 > 1 {
		t.Errorf("segment 4: EaseOutBounce(0.95) = %f, want [0,1]", r4)
	}
	// At t=1, should be exactly 1
	if r := EaseOutBounce(1.0); math.Abs(r-1.0) > 1e-10 {
		t.Errorf("EaseOutBounce(1.0) = %f, want 1.0", r)
	}
}

func TestEaseInBounce_P237(t *testing.T) {
	r := EaseInBounce(0.5)
	if math.IsNaN(r) {
		t.Error("EaseInBounce should not be NaN")
	}
}

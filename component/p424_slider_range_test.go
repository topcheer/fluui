package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestSliderRange_New_P424(t *testing.T) {
	sr := NewSliderRange()
	if sr.Low() != 0 {
		t.Errorf("Low = %v, want 0", sr.Low())
	}
	if sr.High() != 100 {
		t.Errorf("High = %v, want 100", sr.High())
	}
	if sr.Min() != 0 {
		t.Errorf("Min = %v, want 0", sr.Min())
	}
	if sr.Max() != 100 {
		t.Errorf("Max = %v, want 100", sr.Max())
	}
	if sr.Step() != 1 {
		t.Errorf("Step = %v, want 1", sr.Step())
	}
}

func TestSliderRange_WithBounds_P424(t *testing.T) {
	sr := NewSliderRangeWithBounds(0, 1000, 200, 800, 50)
	if sr.Low() != 200 {
		t.Errorf("Low = %v, want 200", sr.Low())
	}
	if sr.High() != 800 {
		t.Errorf("High = %v, want 800", sr.High())
	}
	if sr.Step() != 50 {
		t.Errorf("Step = %v, want 50", sr.Step())
	}
}

func TestSliderRange_SetLow_Clamped_P424(t *testing.T) {
	sr := NewSliderRange()
	sr.SetLow(50)
	if sr.Low() != 50 {
		t.Fatalf("Low = %v, want 50", sr.Low())
	}
	// Low cannot exceed high (high=100)
	sr.SetLow(200)
	if sr.Low() != 100 {
		t.Errorf("Low = %v, want 100 (clamped to high)", sr.Low())
	}
	// Low cannot be below min
	sr.SetLow(-100)
	if sr.Low() != 0 {
		t.Errorf("Low = %v, want 0 (clamped to min)", sr.Low())
	}
}

func TestSliderRange_SetHigh_Clamped_P424(t *testing.T) {
	sr := NewSliderRange()
	sr.SetHigh(80)
	if sr.High() != 80 {
		t.Fatalf("High = %v, want 80", sr.High())
	}
	// High cannot exceed max
	sr.SetHigh(500)
	if sr.High() != 100 {
		t.Errorf("High = %v, want 100 (clamped to max)", sr.High())
	}
	// High cannot go below low
	sr.SetLow(40)
	sr.SetHigh(10)
	if sr.High() != 40 {
		t.Errorf("High = %v, want 40 (clamped to low)", sr.High())
	}
}

func TestSliderRange_OnChange_P424(t *testing.T) {
	sr := NewSliderRange()
	var gotLow, gotHigh float64
	sr.SetOnChange(func(lo, hi float64) {
		gotLow = lo
		gotHigh = hi
	})
	sr.SetLow(30)
	if gotLow != 30 || gotHigh != 100 {
		t.Errorf("OnChange got (%v, %v), want (30, 100)", gotLow, gotHigh)
	}
	sr.SetHigh(70)
	if gotLow != 30 || gotHigh != 70 {
		t.Errorf("OnChange got (%v, %v), want (30, 70)", gotLow, gotHigh)
	}
}

func TestSliderRange_ActiveThumb_P424(t *testing.T) {
	sr := NewSliderRange()
	if sr.ActiveThumb() != ThumbLow {
		t.Error("default active thumb should be ThumbLow")
	}
	sr.SetActiveThumb(ThumbHigh)
	if sr.ActiveThumb() != ThumbHigh {
		t.Error("active thumb should be ThumbHigh")
	}
}

func TestSliderRange_SetRange_P424(t *testing.T) {
	sr := NewSliderRangeWithBounds(0, 100, 20, 80, 5)
	sr.SetRange(10, 50)
	// low was 20, within [10,50] → stays 20
	if sr.Low() != 20 {
		t.Errorf("Low = %v, want 20", sr.Low())
	}
	// high was 80, exceeds 50 → clamped to 50
	if sr.High() != 50 {
		t.Errorf("High = %v, want 50", sr.High())
	}
}

func TestSliderRange_IncrementDecrement_P424(t *testing.T) {
	sr := NewSliderRangeWithBounds(0, 100, 30, 70, 5)
	sr.SetActiveThumb(ThumbLow)
	sr.Increment() // 30 → 35
	if sr.Low() != 35 {
		t.Errorf("Low = %v, want 35", sr.Low())
	}
	sr.Decrement() // 35 → 30
	if sr.Low() != 30 {
		t.Errorf("Low = %v, want 30", sr.Low())
	}
	sr.SetActiveThumb(ThumbHigh)
	sr.Decrement() // 70 → 65
	if sr.High() != 65 {
		t.Errorf("High = %v, want 65", sr.High())
	}
}

func TestSliderRange_HandleKey_P424(t *testing.T) {
	sr := NewSliderRangeWithBounds(0, 100, 30, 70, 5)
	sr.SetActiveThumb(ThumbLow)

	// Right arrow increments
	handled := sr.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if !handled {
		t.Error("KeyRight should be handled")
	}
	if sr.Low() != 35 {
		t.Errorf("Low = %v, want 35 after right", sr.Low())
	}

	// Left arrow decrements
	sr.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if sr.Low() != 30 {
		t.Errorf("Low = %v, want 30 after left", sr.Low())
	}

	// Tab switches active thumb
	sr.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if sr.ActiveThumb() != ThumbHigh {
		t.Error("Tab should switch to ThumbHigh")
	}

	// Home sets active (high) to low value
	sr.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	if sr.High() != 30 {
		t.Errorf("High = %v, want 30 (collapsed to low)", sr.High())
	}

	// End sets active (high) to max
	sr.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	if sr.High() != 100 {
		t.Errorf("High = %v, want 100 (max)", sr.High())
	}
}

func TestSliderRange_HandleKey_Vim_P424(t *testing.T) {
	sr := NewSliderRangeWithBounds(0, 100, 30, 70, 5)
	sr.SetActiveThumb(ThumbLow)
	// 'l' increments (vim right)
	sr.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: 'l'})
	if sr.Low() != 35 {
		t.Errorf("Low = %v, want 35 after 'l'", sr.Low())
	}
	// 'h' decrements (vim left)
	sr.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: 'h'})
	if sr.Low() != 30 {
		t.Errorf("Low = %v, want 30 after 'h'", sr.Low())
	}
}

func TestSliderRange_HandleKey_NilKey_P424(t *testing.T) {
	sr := NewSliderRange()
	if sr.HandleKey(nil) {
		t.Error("nil key should not be handled")
	}
}

func TestSliderRange_Measure_P424(t *testing.T) {
	sr := NewSliderRange()
	sz := sr.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	// With showValues=true, height should be 2
	if sz.H != 2 {
		t.Errorf("H = %v, want 2", sz.H)
	}
	if sz.W != 30 {
		t.Errorf("W = %v, want 30", sz.W)
	}
	// With max width constraint
	sz = sr.Measure(Constraints{MaxWidth: 10, MaxHeight: 1})
	if sz.W != 10 {
		t.Errorf("W = %v, want 10", sz.W)
	}
	if sz.H != 1 {
		t.Errorf("H = %v, want 1", sz.H)
	}
}

func TestSliderRange_Paint_NoPanic_P424(t *testing.T) {
	sr := NewSliderRangeWithBounds(0, 100, 30, 70, 5)
	sr.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2})
	buf := buffer.NewBuffer(40, 2)
	sr.Paint(buf) // should not panic
}

func TestSliderRange_Paint_ZeroBounds_P424(t *testing.T) {
	sr := NewSliderRange()
	sr.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	sr.Paint(buf) // should not panic
}

func TestSliderRange_String_P424(t *testing.T) {
	sr := NewSliderRangeWithBounds(0, 100, 30, 70, 5)
	str := sr.String()
	if str == "" {
		t.Error("String should not be empty")
	}
}

func TestSliderRange_SetStep_Invalid_P424(t *testing.T) {
	sr := NewSliderRange()
	sr.SetStep(-1) // should be ignored
	if sr.Step() != 1 {
		t.Errorf("Step = %v, want 1 (negative ignored)", sr.Step())
	}
	sr.SetStep(0) // should be ignored
	if sr.Step() != 1 {
		t.Errorf("Step = %v, want 1 (zero ignored)", sr.Step())
	}
}

func TestSliderRange_FluentChain_P424(t *testing.T) {
	sr := NewSliderRange().
		SetLow(20).
		SetHigh(80).
		SetStep(5).
		SetLabel("Price").
		SetShowValues(false).
		SetActiveThumb(ThumbHigh)

	if sr.Low() != 20 || sr.High() != 80 || sr.Step() != 5 {
		t.Error("fluent setters failed")
	}
	if sr.Label() != "Price" {
		t.Error("label not set")
	}
	if sr.ShowValues() {
		t.Error("showValues should be false")
	}
	if sr.ActiveThumb() != ThumbHigh {
		t.Error("active thumb should be ThumbHigh")
	}
}

func TestSliderRange_Style_P424(t *testing.T) {
	sr := NewSliderRange()
	st := DefaultSliderRangeStyle()
	sr.SetStyle(st)
	got := sr.Style()
	if got.Filled.Fg != st.Filled.Fg {
		t.Error("style not set correctly")
	}
}

// --- Benchmarks ---

func BenchmarkSliderRange_Paint_P424(b *testing.B) {
	sr := NewSliderRangeWithBounds(0, 100, 30, 70, 5)
	sr.SetShowValues(false)
	sr.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sr.Paint(buf)
	}
}

func BenchmarkSliderRange_Measure_P424(b *testing.B) {
	sr := NewSliderRange()
	cs := Constraints{MaxWidth: 80, MaxHeight: 3}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sr.Measure(cs)
	}
}

package component

import (
	"math"
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── Construction ──────────────────────────────────────────────

func TestNewSlider(t *testing.T) {
	s := NewSlider()
	if s == nil {
		t.Fatal("expected non-nil Slider")
	}
	if s.Min() != 0 {
		t.Errorf("Min = %v, want 0", s.Min())
	}
	if s.Max() != 100 {
		t.Errorf("Max = %v, want 100", s.Max())
	}
	if s.Value() != 0 {
		t.Errorf("Value = %v, want 0", s.Value())
	}
	if s.Step() != 1 {
		t.Errorf("Step = %v, want 1", s.Step())
	}
}

func TestNewSlider_UniqueID(t *testing.T) {
	s1 := NewSlider()
	s2 := NewSlider()
	if s1.ID() == s2.ID() {
		t.Error("expected unique IDs")
	}
}

func TestSlider_ImplementsComponent(t *testing.T) {
	var _ Component = NewSlider()
}

// ─── Value & Range ─────────────────────────────────────────────

func TestSlider_SetValue(t *testing.T) {
	s := NewSlider()
	s.SetValue(50)
	if s.Value() != 50 {
		t.Errorf("Value = %v, want 50", s.Value())
	}
}

func TestSlider_SetValue_ClampHigh(t *testing.T) {
	s := NewSlider()
	s.SetValue(200)
	if s.Value() != 100 {
		t.Errorf("Value = %v, want 100 (clamped)", s.Value())
	}
}

func TestSlider_SetValue_ClampLow(t *testing.T) {
	s := NewSlider()
	s.SetValue(-50)
	if s.Value() != 0 {
		t.Errorf("Value = %v, want 0 (clamped)", s.Value())
	}
}

func TestSlider_SetRange(t *testing.T) {
	s := NewSlider()
	s.SetRange(10, 20)
	if s.Min() != 10 || s.Max() != 20 {
		t.Errorf("range = [%v, %v], want [10, 20]", s.Min(), s.Max())
	}
}

func TestSlider_SetRange_ClampsValue(t *testing.T) {
	s := NewSlider()
	s.SetValue(50)
	s.SetRange(0, 10)
	if s.Value() != 10 {
		t.Errorf("Value = %v, want 10 (clamped to new max)", s.Value())
	}
}

func TestSlider_SetStep(t *testing.T) {
	s := NewSlider()
	s.SetStep(5)
	if s.Step() != 5 {
		t.Errorf("Step = %v, want 5", s.Step())
	}
}

func TestSlider_FloatStep(t *testing.T) {
	s := NewSlider()
	s.SetStep(0.5)
	s.SetValue(0)
	s.Increment()
	if s.Value() != 0.5 {
		t.Errorf("Value = %v, want 0.5", s.Value())
	}
}

func TestSlider_Increment(t *testing.T) {
	s := NewSlider()
	s.SetStep(1)
	s.Increment()
	if s.Value() != 1 {
		t.Errorf("Value = %v, want 1", s.Value())
	}
}

func TestSlider_Decrement(t *testing.T) {
	s := NewSlider()
	s.SetValue(10)
	s.SetStep(1)
	s.Decrement()
	if s.Value() != 9 {
		t.Errorf("Value = %v, want 9", s.Value())
	}
}

func TestSlider_Increment_ClampAtMax(t *testing.T) {
	s := NewSlider()
	s.SetValue(99)
	s.SetStep(5)
	s.Increment()
	if s.Value() != 100 {
		t.Errorf("Value = %v, want 100 (clamped)", s.Value())
	}
}

func TestSlider_Decrement_ClampAtMin(t *testing.T) {
	s := NewSlider()
	s.SetStep(5)
	s.Decrement()
	if s.Value() != 0 {
		t.Errorf("Value = %v, want 0 (clamped)", s.Value())
	}
}

func TestSlider_IncrementBy(t *testing.T) {
	s := NewSlider()
	s.SetStep(5)
	s.IncrementBy(4)
	if s.Value() != 20 {
		t.Errorf("Value = %v, want 20", s.Value())
	}
}

func TestSlider_Ratio(t *testing.T) {
	s := NewSlider()
	s.SetValue(50)
	r := s.Ratio()
	if math.Abs(r-0.5) > 0.001 {
		t.Errorf("Ratio = %v, want ~0.5", r)
	}
}

func TestSlider_Ratio_AtMin(t *testing.T) {
	s := NewSlider()
	s.SetValue(0)
	if s.Ratio() != 0 {
		t.Errorf("Ratio = %v, want 0", s.Ratio())
	}
}

func TestSlider_Ratio_AtMax(t *testing.T) {
	s := NewSlider()
	s.SetValue(100)
	if s.Ratio() != 1 {
		t.Errorf("Ratio = %v, want 1", s.Ratio())
	}
}

func TestSlider_Ratio_ZeroRange(t *testing.T) {
	s := NewSlider()
	s.SetRange(5, 5)
	if s.Ratio() != 0 {
		t.Errorf("Ratio = %v, want 0 for zero range", s.Ratio())
	}
}

func TestSlider_SetFromRatio(t *testing.T) {
	s := NewSlider()
	s.SetFromRatio(0.5)
	if s.Value() != 50 {
		t.Errorf("Value = %v, want 50", s.Value())
	}
}

func TestSlider_SetFromRatio_Clamped(t *testing.T) {
	s := NewSlider()
	s.SetFromRatio(2.0)
	if s.Value() != 100 {
		t.Errorf("Value = %v, want 100 (clamped)", s.Value())
	}
}

// ─── Orientation ───────────────────────────────────────────────

func TestSlider_Orientation(t *testing.T) {
	s := NewSlider()
	if s.Orientation() != SliderHorizontal {
		t.Error("expected horizontal by default")
	}
	s.SetOrientation(SliderVertical)
	if s.Orientation() != SliderVertical {
		t.Error("expected vertical after SetOrientation")
	}
}

// ─── Style & Config ────────────────────────────────────────────

func TestSlider_SetStyle(t *testing.T) {
	s := NewSlider()
	style := SliderStyle{
		Track:     buffer.Style{Fg: buffer.Red},
		Filled:    buffer.Style{Fg: buffer.Blue},
		Handle:    buffer.Style{Fg: buffer.Yellow},
		Label:     buffer.Style{Fg: buffer.White},
		ValueText: buffer.Style{Fg: buffer.Cyan},
	}
	s.SetStyle(style)
	if s.Style().Track.Fg != buffer.Red {
		t.Error("style not set")
	}
}

func TestSlider_SetLabel(t *testing.T) {
	s := NewSlider()
	s.SetLabel("Volume")
	if s.Label() != "Volume" {
		t.Errorf("Label = %q, want 'Volume'", s.Label())
	}
}

func TestSlider_SetShowValue(t *testing.T) {
	s := NewSlider()
	s.SetShowValue(false)
	if s.ShowValue() {
		t.Error("ShowValue should be false")
	}
	s.SetShowValue(true)
	if !s.ShowValue() {
		t.Error("ShowValue should be true")
	}
}

func TestSlider_OnChange(t *testing.T) {
	s := NewSlider()
	called := false
	s.SetOnChange(func(v float64) {
		called = true
	})
	s.SetValue(50)
	if !called {
		t.Error("OnChange should have fired")
	}
}

func TestSlider_OnChange_NotFiredWhenUnchanged(t *testing.T) {
	s := NewSlider()
	s.SetValue(50)
	called := false
	s.SetOnChange(func(v float64) {
		called = true
	})
	s.SetValue(50) // same value
	if called {
		t.Error("OnChange should not fire for unchanged value")
	}
}

// ─── Keyboard ──────────────────────────────────────────────────

func TestSlider_HandleKey_LeftRight(t *testing.T) {
	s := NewSlider()
	s.SetValue(50)
	s.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if s.Value() != 49 {
		t.Errorf("Value = %v, want 49", s.Value())
	}
	s.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if s.Value() != 50 {
		t.Errorf("Value = %v, want 50", s.Value())
	}
}

func TestSlider_HandleKey_UpDown(t *testing.T) {
	s := NewSlider()
	s.SetValue(50)
	s.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if s.Value() != 49 {
		t.Errorf("Value = %v, want 49", s.Value())
	}
	s.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if s.Value() != 50 {
		t.Errorf("Value = %v, want 50", s.Value())
	}
}

func TestSlider_HandleKey_Home(t *testing.T) {
	s := NewSlider()
	s.SetValue(50)
	s.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	if s.Value() != 0 {
		t.Errorf("Value = %v, want 0", s.Value())
	}
}

func TestSlider_HandleKey_End(t *testing.T) {
	s := NewSlider()
	s.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	if s.Value() != 100 {
		t.Errorf("Value = %v, want 100", s.Value())
	}
}

func TestSlider_HandleKey_HL(t *testing.T) {
	s := NewSlider()
	s.SetValue(50)
	s.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: 'h'})
	if s.Value() != 49 {
		t.Errorf("Value = %v, want 49 (h=decrement)", s.Value())
	}
	s.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: 'l'})
	if s.Value() != 50 {
		t.Errorf("Value = %v, want 50 (l=increment)", s.Value())
	}
}

func TestSlider_HandleKey_Nil(t *testing.T) {
	s := NewSlider()
	if s.HandleKey(nil) {
		t.Error("nil key should not be consumed")
	}
}

func TestSlider_HandleKey_Unhandled(t *testing.T) {
	s := NewSlider()
	if s.HandleKey(&term.KeyEvent{Key: term.KeyEscape}) {
		t.Error("Escape should not be consumed")
	}
}

// ─── Measure ───────────────────────────────────────────────────

func TestSlider_Measure(t *testing.T) {
	s := NewSlider()
	size := s.Measure(Constraints{})
	if size.W < 10 {
		t.Errorf("W = %d, should be >= 10", size.W)
	}
	if size.H < 1 {
		t.Errorf("H = %d, should be >= 1", size.H)
	}
}

func TestSlider_Measure_Clamped(t *testing.T) {
	s := NewSlider()
	size := s.Measure(Constraints{MaxWidth: 5, MaxHeight: 1})
	if size.W > 5 {
		t.Errorf("W = %d, should be clamped to 5", size.W)
	}
	if size.H > 1 {
		t.Errorf("H = %d, should be clamped to 1", size.H)
	}
}

// ─── Paint ─────────────────────────────────────────────────────

func TestSlider_Paint_NoPanic(t *testing.T) {
	s := NewSlider()
	s.SetValue(50)
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	s.Paint(buf)
}

func TestSlider_Paint_ZeroBounds(t *testing.T) {
	s := NewSlider()
	buf := buffer.NewBuffer(10, 5)
	s.Paint(buf)
}

func TestSlider_Paint_Vertical(t *testing.T) {
	s := NewSlider()
	s.SetOrientation(SliderVertical)
	s.SetValue(50)
	s.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 10})
	buf := buffer.NewBuffer(3, 10)
	s.Paint(buf)
}

func TestSlider_Paint_WithLabel(t *testing.T) {
	s := NewSlider()
	s.SetLabel("Brightness")
	s.SetValue(75)
	s.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	buf := buffer.NewBuffer(30, 3)
	s.Paint(buf)
}

// ─── Misc ──────────────────────────────────────────────────────

func TestSlider_String(t *testing.T) {
	s := NewSlider()
	if s.String() == "" {
		t.Error("String should not be empty")
	}
}

func TestSlider_Children(t *testing.T) {
	s := NewSlider()
	if s.Children() != nil {
		t.Error("Children should return nil")
	}
}

func TestSlider_clampFloat(t *testing.T) {
	if clampFloat(5, 0, 10) != 5 {
		t.Error("clampFloat(5,0,10) should be 5")
	}
	if clampFloat(-1, 0, 10) != 0 {
		t.Error("clampFloat(-1,0,10) should be 0")
	}
	if clampFloat(15, 0, 10) != 10 {
		t.Error("clampFloat(15,0,10) should be 10")
	}
}

// ─── Concurrency ───────────────────────────────────────────────

func TestSlider_ConcurrentAccess(t *testing.T) {
	s := NewSlider()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			s.Value()
			s.Ratio()
		}()
		go func() {
			defer wg.Done()
			s.Increment()
			s.Decrement()
		}()
		go func() {
			defer wg.Done()
			s.SetValue(50)
		}()
	}
	wg.Wait()
}

func TestSlider_ConcurrentPaint(t *testing.T) {
	s := NewSlider()
	s.SetValue(50)
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			buf := buffer.NewBuffer(20, 3)
			s.Paint(buf)
		}()
		go func() {
			defer wg.Done()
			s.Increment()
		}()
	}
	wg.Wait()
}

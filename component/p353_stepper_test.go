package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP353_Stepper_Create(t *testing.T) {
	s := NewStepper([]StepperStep{
		{Title: "Input"},
		{Title: "Process"},
		{Title: "Output"},
	})
	if s.StepCount() != 3 {
		t.Errorf("count = %d", s.StepCount())
	}
	if s.CurrentStep() != 0 {
		t.Errorf("current = %d", s.CurrentStep())
	}
}

func TestP353_Stepper_Next(t *testing.T) {
	s := NewStepper([]StepperStep{{Title: "A"}, {Title: "B"}, {Title: "C"}})
	if !s.Next() {
		t.Error("Next should succeed")
	}
	if s.CurrentStep() != 1 {
		t.Errorf("current = %d", s.CurrentStep())
	}
	s.Next()
	if s.Next() {
		t.Error("Next at last step should return false")
	}
}

func TestP353_Stepper_Prev(t *testing.T) {
	s := NewStepper([]StepperStep{{Title: "A"}, {Title: "B"}})
	s.Next()
	if !s.Prev() {
		t.Error("Prev should succeed")
	}
	if s.CurrentStep() != 0 {
		t.Errorf("current = %d", s.CurrentStep())
	}
	if s.Prev() {
		t.Error("Prev at first step should return false")
	}
}

func TestP353_Stepper_SetCurrent(t *testing.T) {
	s := NewStepper([]StepperStep{{Title: "A"}, {Title: "B"}, {Title: "C"}})
	s.SetCurrent(2)
	if s.CurrentStep() != 2 {
		t.Errorf("current = %d", s.CurrentStep())
	}
	s.SetCurrent(-1)
	if s.CurrentStep() != 0 {
		t.Errorf("negative should clamp: %d", s.CurrentStep())
	}
	s.SetCurrent(99)
	if s.CurrentStep() != 2 {
		t.Errorf("overflow should clamp: %d", s.CurrentStep())
	}
}

func TestP353_Stepper_SetSteps(t *testing.T) {
	s := NewStepper([]StepperStep{{Title: "A"}, {Title: "B"}, {Title: "C"}})
	s.SetCurrent(2)
	s.SetSteps([]StepperStep{{Title: "X"}})
	if s.CurrentStep() != 0 {
		t.Errorf("current should clamp: %d", s.CurrentStep())
	}
}

func TestP353_Stepper_IsComplete(t *testing.T) {
	s := NewStepper([]StepperStep{{Title: "A"}, {Title: "B"}})
	if s.IsComplete() {
		t.Error("should not be complete at step 0")
	}
	s.Next()
	if !s.IsComplete() {
		t.Error("should be complete at last step")
	}
}

func TestP353_Stepper_Vertical(t *testing.T) {
	s := NewStepper([]StepperStep{{Title: "A"}})
	if s.IsVertical() {
		t.Error("should default to horizontal")
	}
	s.SetVertical(true)
	if !s.IsVertical() {
		t.Error("should be vertical")
	}
}

func TestP353_Stepper_Measure_Horizontal(t *testing.T) {
	s := NewStepper([]StepperStep{{Title: "AB"}, {Title: "CD"}})
	sz := s.Measure(Constraints{MaxWidth: 80, MaxHeight: 1})
	if sz.H != 1 {
		t.Errorf("horizontal height = %d, want 1", sz.H)
	}
}

func TestP353_Stepper_Measure_Vertical(t *testing.T) {
	s := NewStepper([]StepperStep{{Title: "A"}, {Title: "B"}, {Title: "C"}})
	s.SetVertical(true)
	sz := s.Measure(Constraints{MaxWidth: 40, MaxHeight: 10})
	if sz.H < 5 {
		t.Errorf("vertical height = %d, expected at least 5", sz.H)
	}
}

func TestP353_Stepper_Paint_Horizontal(t *testing.T) {
	s := NewStepper([]StepperStep{
		{Title: "Input"},
		{Title: "Process"},
		{Title: "Output"},
	})
	s.Next()
	s.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 1})
	buf := buffer.NewBuffer(50, 1)
	s.Paint(buf)
	if buf.GetCell(0, 0).Rune == 0 {
		t.Error("expected non-empty cell")
	}
}

func TestP353_Stepper_Paint_Vertical(t *testing.T) {
	s := NewStepper([]StepperStep{
		{Title: "Step 1"},
		{Title: "Step 2"},
	})
	s.SetVertical(true)
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 6})
	buf := buffer.NewBuffer(20, 6)
	s.Paint(buf)
}

func TestP353_Stepper_Paint_Empty(t *testing.T) {
	s := NewStepper(nil)
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	s.Paint(buf)
}

func TestP353_Stepper_Paint_ZeroBounds(t *testing.T) {
	s := NewStepper([]StepperStep{{Title: "A"}})
	s.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(20, 1)
	s.Paint(buf)
}

func TestP353_Stepper_Paint_NarrowWidth(t *testing.T) {
	s := NewStepper([]StepperStep{
		{Title: "VeryLongTitle"},
		{Title: "AnotherLongOne"},
	})
	s.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	s.Paint(buf)
}

func BenchmarkStepper_Paint(b *testing.B) {
	s := NewStepper([]StepperStep{
		{Title: "Parse"},
		{Title: "Analyze"},
		{Title: "Generate"},
		{Title: "Review"},
		{Title: "Output"},
	})
	s.SetCurrent(2)
	s.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Paint(buf)
	}
}

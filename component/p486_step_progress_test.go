package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestStepProgressBasic(t *testing.T) {
	sp := NewStepProgress()
	sp.AddStep("Account")
	sp.AddStep("Profile")
	sp.AddStep("Confirm")

	if sp.StepCount() != 3 {
		t.Errorf("StepCount = %d, want 3", sp.StepCount())
	}
	if sp.CurrentStep() != 0 {
		t.Errorf("CurrentStep = %d, want 0", sp.CurrentStep())
	}
}

func TestStepProgressSetCurrent(t *testing.T) {
	sp := NewStepProgress()
	sp.AddStep("A")
	sp.AddStep("B")
	sp.AddStep("C")

	sp.SetCurrentStep(2)
	if sp.CurrentStep() != 2 {
		t.Errorf("CurrentStep = %d, want 2", sp.CurrentStep())
	}
}

func TestStepProgressClamp(t *testing.T) {
	sp := NewStepProgress()
	sp.AddStep("A")
	sp.AddStep("B")

	sp.SetCurrentStep(-5)
	if sp.CurrentStep() != 0 {
		t.Errorf("Clamped negative: CurrentStep = %d, want 0", sp.CurrentStep())
	}
	sp.SetCurrentStep(100)
	if sp.CurrentStep() != 1 {
		t.Errorf("Clamped overflow: CurrentStep = %d, want 1", sp.CurrentStep())
	}
}

func TestStepProgressComplete(t *testing.T) {
	sp := NewStepProgress()
	sp.AddStep("A")
	sp.AddStep("B")
	sp.SetCurrentStep(0)

	sp.Complete()
	// All steps should be completed state
	if sp.stepStateLocked(0) != StepCompleted {
		t.Error("step 0 should be completed after Complete()")
	}
	if sp.stepStateLocked(1) != StepCompleted {
		t.Error("step 1 should be completed after Complete()")
	}
}

func TestStepProgressReset(t *testing.T) {
	sp := NewStepProgress()
	sp.AddStep("A")
	sp.AddStep("B")
	sp.SetCurrentStep(1)
	sp.Complete()

	sp.Reset()
	if sp.CurrentStep() != 0 {
		t.Errorf("After reset: CurrentStep = %d, want 0", sp.CurrentStep())
	}
	if sp.stepStateLocked(0) != StepCurrent {
		t.Error("step 0 should be current after reset")
	}
}

func TestStepProgressStateLogic(t *testing.T) {
	sp := NewStepProgress()
	sp.AddStep("A")
	sp.AddStep("B")
	sp.AddStep("C")
	sp.SetCurrentStep(1)

	if sp.stepStateLocked(0) != StepCompleted {
		t.Error("step 0 should be completed")
	}
	if sp.stepStateLocked(1) != StepCurrent {
		t.Error("step 1 should be current")
	}
	if sp.stepStateLocked(2) != StepUpcoming {
		t.Error("step 2 should be upcoming")
	}
}

func TestStepProgressMeasure(t *testing.T) {
	sp := NewStepProgress()
	sp.AddStep("A")
	sp.AddStep("B")
	sp.AddStep("C")
	s := sp.Measure(Constraints{})
	if s.W < 10 {
		t.Errorf("W = %d, want >= 10", s.W)
	}
	if s.H < 1 {
		t.Errorf("H = %d, want >= 1", s.H)
	}
}

func TestStepProgressPaint(t *testing.T) {
	sp := NewStepProgress()
	sp.AddStep("Start")
	sp.AddStep("Middle")
	sp.AddStep("End")
	sp.SetCurrentStep(1)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 2})

	buf := buffer.NewBuffer(60, 2)
	sp.Paint(buf)

	// Find check mark for completed step 0
	foundCheck := false
	for x := 0; x < 60; x++ {
		if buf.GetCell(x, 0).Rune == '✓' {
			foundCheck = true
			break
		}
	}
	if !foundCheck {
		t.Error("completed check ✓ not found")
	}

	// Find current dot for step 1
	foundDot := false
	for x := 0; x < 60; x++ {
		if buf.GetCell(x, 0).Rune == '●' {
			foundDot = true
			break
		}
	}
	if !foundDot {
		t.Error("current dot ● not found")
	}
}

func TestStepProgressPaintCompleted(t *testing.T) {
	sp := NewStepProgress()
	sp.AddStep("A")
	sp.AddStep("B")
	sp.AddStep("C")
	sp.Complete()
	sp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 2})

	buf := buffer.NewBuffer(60, 2)
	sp.Paint(buf)

	// All should be ✓
	checkCount := 0
	for x := 0; x < 60; x++ {
		if buf.GetCell(x, 0).Rune == '✓' {
			checkCount++
		}
	}
	if checkCount != 3 {
		t.Errorf("check count = %d, want 3 (all completed)", checkCount)
	}
}

func TestStepProgressPaintEmpty(t *testing.T) {
	sp := NewStepProgress()
	sp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 2})
	buf := buffer.NewBuffer(60, 2)
	sp.Paint(buf) // should not panic
}

func TestStepProgressChildren(t *testing.T) {
	sp := NewStepProgress()
	if sp.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestStepProgressStyle(t *testing.T) {
	sp := NewStepProgress()
	sp.SetStyle(StepProgressStyle{
		Completed: buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Current:   buffer.Style{Fg: buffer.RGB(0, 0, 255), Flags: buffer.Bold},
		Upcoming:  buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Connector: buffer.Style{Fg: buffer.RGB(150, 150, 150)},
		Label:     buffer.Style{Fg: buffer.RGB(200, 200, 200)},
	})
	sp.AddStep("A")
	sp.AddStep("B")
	sp.SetCurrentStep(0)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2})
	buf := buffer.NewBuffer(40, 2)
	sp.Paint(buf)
}

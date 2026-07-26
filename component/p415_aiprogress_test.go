package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP415_NewAIProgress(t *testing.T) {
	a := NewAIProgress()
	if a.Phase() != AIPhaseIdle { t.Errorf("Phase = %v", a.Phase()) }
	if a.PhaseLabel() != "Idle" { t.Errorf("Label = %q", a.PhaseLabel()) }
	if !a.ShowLabel() { t.Error("should show label") }
	if !a.ShowSpin() { t.Error("should show spin") }
	if a.ID() == "" { t.Error("ID empty") }
}

func TestP415_SetPhase(t *testing.T) {
	a := NewAIProgress()
	a.SetPhase(AIPhaseGenerating)
	if a.Phase() != AIPhaseGenerating { t.Errorf("Phase = %v", a.Phase()) }
	if a.PhaseLabel() != "Generating..." { t.Errorf("Label = %q", a.PhaseLabel()) }
}

func TestP415_IsBusy(t *testing.T) {
	a := NewAIProgress()
	if a.IsBusy() { t.Error("Idle should not be busy") }
	a.SetPhase(AIPhaseThinking)
	if !a.IsBusy() { t.Error("Thinking should be busy") }
	a.SetPhase(AIPhaseComplete)
	if a.IsBusy() { t.Error("Complete should not be busy") }
	a.SetPhase(AIPhaseError)
	if a.IsBusy() { t.Error("Error should not be busy") }
}

func TestP415_Tick(t *testing.T) {
	a := NewAIProgress()
	a.SetPhase(AIPhaseThinking)
	for i := 0; i < 20; i++ { a.Tick() } // should cycle through frames
}

func TestP415_Tick_Complete(t *testing.T) {
	a := NewAIProgress()
	a.SetPhase(AIPhaseComplete)
	a.Tick() // no spin frames, should not panic
}

func TestP415_SetProgress(t *testing.T) {
	a := NewAIProgress()
	a.SetProgress(0.5)
	if a.Progress() != 0.5 { t.Errorf("Progress = %v", a.Progress()) }
	a.SetProgress(-1) // hide
	if a.Progress() != -1 { t.Errorf("Progress = %v", a.Progress()) }
}

func TestP415_Measure(t *testing.T) {
	a := NewAIProgress()
	a.SetPhase(AIPhaseGenerating)
	s := a.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if s.H != 1 { t.Errorf("H = %d", s.H) }
	if s.W < 15 { t.Errorf("W = %d, too small", s.W) }
}

func TestP415_Measure_WithProgress(t *testing.T) {
	a := NewAIProgress()
	a.SetProgress(0.75)
	s := a.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	// "○ Idle 75%" = 2 + 4 + 1 + 3 = 10
	if s.W < 10 { t.Errorf("W = %d", s.W) }
}

func TestP415_Paint_Thinking(t *testing.T) {
	a := NewAIProgress()
	a.SetPhase(AIPhaseThinking)
	a.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	a.Paint(buf)
	// Should have spinner at 0 and label text after
	c := buf.GetCell(0, 0)
	if c.Rune == 0 { t.Log("spinner cell might be braille (wide)") }
}

func TestP415_Paint_Generating(t *testing.T) {
	a := NewAIProgress()
	a.SetPhase(AIPhaseGenerating)
	a.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	a.Paint(buf)
}

func TestP415_Paint_Complete(t *testing.T) {
	a := NewAIProgress()
	a.SetPhase(AIPhaseComplete)
	a.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	a.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != '\u2714' { t.Errorf("complete icon = %q, want ✔", string(c.Rune)) }
}

func TestP415_Paint_Error(t *testing.T) {
	a := NewAIProgress()
	a.SetPhase(AIPhaseError)
	a.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	a.Paint(buf)
}

func TestP415_Paint_WithProgress(t *testing.T) {
	a := NewAIProgress()
	a.SetPhase(AIPhaseGenerating)
	a.SetProgress(0.42)
	a.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	a.Paint(buf)
}

func TestP415_Paint_NoSpin(t *testing.T) {
	a := NewAIProgress()
	a.SetShowSpin(false)
	a.SetPhase(AIPhaseThinking)
	a.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	a.Paint(buf)
	// First cell should be label text, not spinner
	c := buf.GetCell(0, 0)
	if c.Rune != 'T' { t.Errorf("cell = %q, want 'T' (Thinking)", string(c.Rune)) }
}

func TestP415_Paint_NoLabel(t *testing.T) {
	a := NewAIProgress()
	a.SetShowLabel(false)
	a.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	a.Paint(buf)
}

func TestP415_Paint_ZeroBounds(t *testing.T) {
	a := NewAIProgress()
	a.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	a.Paint(buf)
}

func TestP415_Paint_NonZeroOffset(t *testing.T) {
	a := NewAIProgress()
	a.SetPhase(AIPhaseComplete)
	a.SetBounds(Rect{X: 10, Y: 5, W: 20, H: 1})
	buf := buffer.NewBuffer(40, 10)
	a.Paint(buf)
	c := buf.GetCell(10, 5)
	if c.Rune != '\u2714' { t.Errorf("offset cell = %q", string(c.Rune)) }
}

func TestP415_Concurrent(t *testing.T) {
	a := NewAIProgress()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ { a.Tick(); a.SetPhase(AIPhaseGenerating) }
		close(done)
	}()
	for i := 0; i < 500; i++ { _ = a.Phase() }
	<-done
}

func TestP415_SatisfiesComponent(t *testing.T) {
	var _ Component = (*AIProgress)(nil)
}

func BenchmarkP415_AIProgress_Paint(b *testing.B) {
	a := NewAIProgress()
	a.SetPhase(AIPhaseGenerating)
	a.SetProgress(0.65)
	a.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { a.Paint(buf) }
}

package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── CodeBlockDiff Tests ───

func TestCodeBlockDiffBasic(t *testing.T) {
	cd := NewCodeBlockDiff()
	cd.SetDiff([]string{"a", "b"}, []string{"a", "c"})
	if n := cd.LineCount(); n != 2 {
		t.Errorf("LineCount = %d, want 2", n)
	}
}

func TestCodeBlockDiffAllSame(t *testing.T) {
	cd := NewCodeBlockDiff()
	cd.SetDiff([]string{"same"}, []string{"same"})
	if n := cd.LineCount(); n != 1 {
		t.Errorf("LineCount = %d, want 1", n)
	}
}

func TestCodeBlockDiffAdded(t *testing.T) {
	cd := NewCodeBlockDiff()
	cd.SetDiff([]string{}, []string{"new line"})
	if n := cd.LineCount(); n != 1 {
		t.Errorf("LineCount = %d, want 1", n)
	}
}

func TestCodeBlockDiffRemoved(t *testing.T) {
	cd := NewCodeBlockDiff()
	cd.SetDiff([]string{"old"}, []string{})
	if n := cd.LineCount(); n != 1 {
		t.Errorf("LineCount = %d, want 1", n)
	}
}

func TestCodeBlockDiffPaint(t *testing.T) {
	cd := NewCodeBlockDiff()
	cd.SetDiff([]string{"old"}, []string{"new"})
	cd.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	cd.Paint(buf)
	// Row 1 should have '-' prefix
	hasRemoved := false
	for i := 0; i < 40; i++ {
		if buf.GetCell(i, 1).Rune == '-' {
			hasRemoved = true
			break
		}
	}
	if !hasRemoved {
		t.Error("Paint should show removed line prefix")
	}
}

func TestCodeBlockDiffChildren(t *testing.T) {
	cd := NewCodeBlockDiff()
	if c := cd.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestCodeBlockDiffStyle(t *testing.T) {
	cd := NewCodeBlockDiff()
	cd.SetStyle(CodeBlockDiffStyle{
		Added:     buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Removed:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Context:   buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		LineNum:   buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Separator: buffer.Style{Fg: buffer.RGB(32, 32, 32)},
	})
	cd.SetDiff([]string{"a"}, []string{"b"})
	buf := buffer.NewBuffer(40, 3)
	cd.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	cd.Paint(buf)
}

// ─── AIReasoningChain Tests ───

func TestAIReasoningChainBasic(t *testing.T) {
	rc := NewAIReasoningChain()
	rc.AddStep("premise", "conclusion")
	if n := rc.StepCount(); n != 1 {
		t.Errorf("StepCount = %d, want 1", n)
	}
}

func TestAIReasoningChainOverflow(t *testing.T) {
	rc := NewAIReasoningChain()
	for i := 0; i < reasoningChainMaxSteps+5; i++ {
		rc.AddStep("p", "c")
	}
	if n := rc.StepCount(); n != reasoningChainMaxSteps {
		t.Errorf("StepCount = %d, want %d (capped)", n, reasoningChainMaxSteps)
	}
}

func TestAIReasoningChainClear(t *testing.T) {
	rc := NewAIReasoningChain()
	rc.AddStep("p", "c")
	rc.SetConclusion("done")
	rc.Clear()
	if n := rc.StepCount(); n != 0 {
		t.Errorf("StepCount after Clear = %d, want 0", n)
	}
	if rc.finalAnswer != "" {
		t.Errorf("finalAnswer = %q, want ''", rc.finalAnswer)
	}
}

func TestAIReasoningChainPaint(t *testing.T) {
	rc := NewAIReasoningChain()
	rc.AddStep("query", "search")
	rc.SetConclusion("answer found")
	rc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	rc.Paint(buf)
	// Should have Step label
	hasStep := false
	for i := 0; i < 40; i++ {
		if buf.GetCell(i, 0).Rune == 'S' {
			hasStep = true
			break
		}
	}
	if !hasStep {
		t.Error("Paint should show 'Step' label")
	}
}

func TestAIReasoningChainChildren(t *testing.T) {
	rc := NewAIReasoningChain()
	if c := rc.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestAIReasoningChainStyle(t *testing.T) {
	rc := NewAIReasoningChain()
	rc.SetStyle(AIReasoningChainStyle{
		Premise:     buffer.Style{Fg: buffer.RGB(147, 197, 253)},
		Arrow:       buffer.Style{Fg: buffer.RGB(251, 191, 36)},
		Conclusion:  buffer.Style{Fg: buffer.RGB(203, 213, 225)},
		FinalAnswer: buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		StepNum:     buffer.Style{Fg: buffer.RGB(148, 163, 184)},
	})
	rc.AddStep("a", "b").SetConclusion("done")
	buf := buffer.NewBuffer(40, 4)
	rc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	rc.Paint(buf)
}

// ─── StatusDot Tests ───

func TestStatusDotBasic(t *testing.T) {
	sd := NewStatusDot()
	sd.SetState(StatusActive)
	if s := sd.State(); s != StatusActive {
		t.Errorf("State = %d, want StatusActive", s)
	}
}

func TestStatusDotAllStates(t *testing.T) {
	states := []StatusDotState{StatusIdle, StatusActive, StatusSuccess, StatusWarning, StatusError}
	for _, s := range states {
		sd := NewStatusDot()
		sd.SetState(s)
		if sd.State() != s {
			t.Errorf("State = %d, want %d", sd.State(), s)
		}
	}
}

func TestStatusDotInvalid(t *testing.T) {
	sd := NewStatusDot()
	sd.SetState(StatusDotState(99))
	if s := sd.State(); s != StatusIdle {
		t.Errorf("State = %d, want StatusIdle (clamped)", s)
	}
}

func TestStatusDotCustomLabel(t *testing.T) {
	sd := NewStatusDot()
	sd.SetLabel("Custom")
	if sd.label != "Custom" {
		t.Errorf("label = %q, want 'Custom'", sd.label)
	}
}

func TestStatusDotHideDefault(t *testing.T) {
	sd := NewStatusDot()
	sd.SetShowDefaultLabel(false)
	sd.SetLabel("")
	sd.SetBounds(Rect{X: 0, Y: 0, W: 14, H: 1})
	buf := buffer.NewBuffer(14, 1)
	sd.Paint(buf)
	// Should still show dot icon
	if r := buf.GetCell(0, 0).Rune; r != statusDotIcons[StatusIdle] {
		t.Errorf("First rune = %q, want %q", r, statusDotIcons[StatusIdle])
	}
}

func TestStatusDotPaint(t *testing.T) {
	sd := NewStatusDot()
	sd.SetState(StatusError)
	sd.SetLabel("Failed")
	sd.SetBounds(Rect{X: 0, Y: 0, W: 14, H: 1})
	buf := buffer.NewBuffer(14, 1)
	sd.Paint(buf)
	if r := buf.GetCell(0, 0).Rune; r != '✗' {
		t.Errorf("First rune = %q, want '✗'", r)
	}
}

func TestStatusDotChildren(t *testing.T) {
	sd := NewStatusDot()
	if c := sd.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestStatusDotStyle(t *testing.T) {
	sd := NewStatusDot()
	sd.SetStyle(StatusDotStyle{
		Idle:    buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Active:  buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Success: buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Warning: buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Error:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Label:   buffer.Style{Fg: buffer.RGB(255, 255, 255)},
	})
	sd.SetState(StatusSuccess)
	buf := buffer.NewBuffer(14, 1)
	sd.SetBounds(Rect{X: 0, Y: 0, W: 14, H: 1})
	sd.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintCodeBlockDiff(b *testing.B) {
	cd := NewCodeBlockDiff()
	cd.SetDiff([]string{"func main() {", "    x := 1", "    return x", "}"}, []string{"func main() {", "    x := 2", "    return x + 1", "}"})
	cd.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cd.Paint(buf)
	}
}

func BenchmarkPaintAIReasoningChain(b *testing.B) {
	rc := NewAIReasoningChain()
	rc.AddStep("Analyze query", "Parse intent")
	rc.AddStep("Search context", "Found 3 results")
	rc.AddStep("Generate answer", "Synthesize response")
	rc.SetConclusion("Final answer")
	rc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 7})
	buf := buffer.NewBuffer(40, 7)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rc.Paint(buf)
	}
}

func BenchmarkPaintStatusDot(b *testing.B) {
	sd := NewStatusDot()
	sd.SetState(StatusActive)
	sd.SetLabel("Running")
	sd.SetBounds(Rect{X: 0, Y: 0, W: 14, H: 1})
	buf := buffer.NewBuffer(14, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sd.Paint(buf)
	}
}

package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── PipelineFlow Tests ───

func TestPipelineFlowBasic(t *testing.T) {
	pf := NewPipelineFlow()
	pf.AddStage("Input", StageDone)
	pf.AddStage("Tokenize", StageActive)
	if count := pf.Count(); count != 2 {
		t.Errorf("Count = %d, want 2", count)
	}
}

func TestPipelineFlowMultiple(t *testing.T) {
	pf := NewPipelineFlow()
	pf.AddStage("A", StageDone)
	pf.AddStage("B", StageDone)
	pf.AddStage("C", StageActive)
	pf.AddStage("D", StagePending)
	if count := pf.Count(); count != 4 {
		t.Errorf("Count = %d, want 4", count)
	}
}

func TestPipelineFlowOverflow(t *testing.T) {
	pf := NewPipelineFlow()
	for i := 0; i < pipelineMaxStages+5; i++ {
		pf.AddStage("S", StagePending)
	}
	if count := pf.Count(); count != pipelineMaxStages {
		t.Errorf("Count = %d, want %d (capped)", count, pipelineMaxStages)
	}
}

func TestPipelineFlowClear(t *testing.T) {
	pf := NewPipelineFlow()
	pf.AddStage("A", StageDone)
	pf.Clear()
	if count := pf.Count(); count != 0 {
		t.Errorf("Count after Clear = %d, want 0", count)
	}
}

func TestPipelineFlowSetStatus(t *testing.T) {
	pf := NewPipelineFlow()
	pf.AddStage("A", StagePending)
	pf.AddStage("B", StagePending)
	pf.SetStageStatus(0, StageDone)
	if pf.stages[0].status != StageDone {
		t.Errorf("Stage 0 status = %d, want StageDone(%d)", pf.stages[0].status, StageDone)
	}
}

func TestPipelineFlowSetStatusInvalid(t *testing.T) {
	pf := NewPipelineFlow()
	pf.AddStage("A", StagePending)
	// Should not panic with invalid index
	pf.SetStageStatus(-1, StageDone)
	pf.SetStageStatus(99, StageDone)
}

func TestPipelineFlowEmpty(t *testing.T) {
	pf := NewPipelineFlow()
	if count := pf.Count(); count != 0 {
		t.Errorf("Count = %d, want 0", count)
	}
}

func TestPipelineFlowPaint(t *testing.T) {
	pf := NewPipelineFlow()
	pf.AddStage("Input", StageDone)
	pf.AddStage("Process", StageActive)
	pf.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	pf.Paint(buf)
	// First should be ✓ icon (StageDone)
	if r := buf.GetCell(0, 0).Rune; r != '✓' {
		t.Errorf("First rune = %q, want '✓'", r)
	}
}

func TestPipelineFlowChildren(t *testing.T) {
	pf := NewPipelineFlow()
	if children := pf.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestPipelineFlowStyle(t *testing.T) {
	pf := NewPipelineFlow()
	custom := PipelineFlowStyle{
		Done:    buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Active:  buffer.Style{Fg: buffer.RGB(255, 255, 0)},
		Pending: buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Error:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Arrow:   buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	}
	pf.SetStyle(custom)
	pf.AddStage("A", StageError)
	pf.AddStage("B", StageActive)
	buf := buffer.NewBuffer(40, 1)
	pf.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	pf.Paint(buf)
}

// ─── ThinkingBudget Tests ───

func TestThinkingBudgetBasic(t *testing.T) {
	tb := NewThinkingBudget()
	tb.SetBudget(500, 1000)
	if pct := tb.Percent(); pct != 50 {
		t.Errorf("Percent = %d, want 50", pct)
	}
}

func TestThinkingBudgetFull(t *testing.T) {
	tb := NewThinkingBudget()
	tb.SetBudget(1000, 1000)
	if pct := tb.Percent(); pct != 100 {
		t.Errorf("Percent = %d, want 100", pct)
	}
}

func TestThinkingBudgetZero(t *testing.T) {
	tb := NewThinkingBudget()
	tb.SetBudget(0, 1000)
	if pct := tb.Percent(); pct != 0 {
		t.Errorf("Percent = %d, want 0", pct)
	}
}

func TestThinkingBudgetNegative(t *testing.T) {
	tb := NewThinkingBudget()
	tb.SetBudget(-10, -5)
	if pct := tb.Percent(); pct != 0 {
		t.Errorf("Percent = %d, want 0", pct)
	}
}

func TestThinkingBudgetOverflow(t *testing.T) {
	tb := NewThinkingBudget()
	tb.SetBudget(2000, 1000)
	if pct := tb.Percent(); pct != 100 {
		t.Errorf("Percent = %d, want 100 (capped)", pct)
	}
}

func TestThinkingBudgetColorLevels(t *testing.T) {
	tb := NewThinkingBudget()
	// Normal: < 70%
	tb.SetBudget(50, 100)
	if tb.curStyle.Fg != tb.style.Normal.Fg {
		t.Error("Expected Normal style for 50%")
	}
	// High: 70-89%
	tb.SetBudget(75, 100)
	if tb.curStyle.Fg != tb.style.High.Fg {
		t.Error("Expected High style for 75%")
	}
	// Critical: >= 90%
	tb.SetBudget(95, 100)
	if tb.curStyle.Fg != tb.style.Critical.Fg {
		t.Error("Expected Critical style for 95%")
	}
}

func TestThinkingBudgetWidth(t *testing.T) {
	tb := NewThinkingBudget()
	tb.SetWidth(40)
	if tb.width != 40 {
		t.Errorf("width = %d, want 40", tb.width)
	}
	tb.SetWidth(5)
	if tb.width != 15 {
		t.Errorf("width = %d, want 15 (clamped)", tb.width)
	}
}

func TestThinkingBudgetPaint(t *testing.T) {
	tb := NewThinkingBudget()
	tb.SetBudget(500, 1000)
	tb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2})
	buf := buffer.NewBuffer(40, 2)
	tb.Paint(buf)
	// Should start with 'T' from "Think"
	if r := buf.GetCell(0, 0).Rune; r != 'T' {
		t.Errorf("First rune = %q, want 'T'", r)
	}
}

func TestThinkingBudgetChildren(t *testing.T) {
	tb := NewThinkingBudget()
	if children := tb.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestThinkingBudgetStyle(t *testing.T) {
	tb := NewThinkingBudget()
	custom := ThinkingBudgetStyle{
		Normal:   buffer.Style{Fg: buffer.RGB(128, 0, 128)},
		High:     buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Critical: buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Label:    buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Value:    buffer.Style{Fg: buffer.RGB(255, 255, 255)},
	}
	tb.SetStyle(custom)
	tb.SetBudget(300, 1000)
	buf := buffer.NewBuffer(40, 2)
	tb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2})
	tb.Paint(buf)
}

// ─── BufferHealthBar Tests ───

func TestBufferHealthBasic(t *testing.T) {
	bh := NewBufferHealthBar()
	bh.SetUtilization(65, 100)
	if pct := bh.Percent(); pct != 65 {
		t.Errorf("Percent = %d, want 65", pct)
	}
}

func TestBufferHealthFull(t *testing.T) {
	bh := NewBufferHealthBar()
	bh.SetUtilization(100, 100)
	if pct := bh.Percent(); pct != 100 {
		t.Errorf("Percent = %d, want 100", pct)
	}
}

func TestBufferHealthZero(t *testing.T) {
	bh := NewBufferHealthBar()
	bh.SetUtilization(0, 100)
	if pct := bh.Percent(); pct != 0 {
		t.Errorf("Percent = %d, want 0", pct)
	}
}

func TestBufferHealthNegative(t *testing.T) {
	bh := NewBufferHealthBar()
	bh.SetUtilization(-10, -5)
	if pct := bh.Percent(); pct != 0 {
		t.Errorf("Percent = %d, want 0", pct)
	}
}

func TestBufferHealthOverflow(t *testing.T) {
	bh := NewBufferHealthBar()
	bh.SetUtilization(200, 100)
	if pct := bh.Percent(); pct != 100 {
		t.Errorf("Percent = %d, want 100 (capped)", pct)
	}
}

func TestBufferHealthColorLevels(t *testing.T) {
	bh := NewBufferHealthBar()
	// Normal: < 60%
	bh.SetUtilization(40, 100)
	if bh.curStyle.Fg != bh.style.Normal.Fg {
		t.Error("Expected Normal style for 40%")
	}
	// High: 60-84%
	bh.SetUtilization(70, 100)
	if bh.curStyle.Fg != bh.style.High.Fg {
		t.Error("Expected High style for 70%")
	}
	// Critical: >= 85%
	bh.SetUtilization(90, 100)
	if bh.curStyle.Fg != bh.style.Critical.Fg {
		t.Error("Expected Critical style for 90%")
	}
}

func TestBufferHealthWidth(t *testing.T) {
	bh := NewBufferHealthBar()
	bh.SetWidth(30)
	if bh.width != 30 {
		t.Errorf("width = %d, want 30", bh.width)
	}
	bh.SetWidth(5)
	if bh.width != 10 {
		t.Errorf("width = %d, want 10 (clamped)", bh.width)
	}
}

func TestBufferHealthPaint(t *testing.T) {
	bh := NewBufferHealthBar()
	bh.SetUtilization(50, 100)
	bh.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	bh.Paint(buf)
	// Should start with 'B' from "Buf"
	if r := buf.GetCell(0, 0).Rune; r != 'B' {
		t.Errorf("First rune = %q, want 'B'", r)
	}
}

func TestBufferHealthChildren(t *testing.T) {
	bh := NewBufferHealthBar()
	if children := bh.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestBufferHealthStyle(t *testing.T) {
	bh := NewBufferHealthBar()
	custom := BufferHealthStyle{
		Normal:   buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		High:     buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Critical: buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Label:    buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Pct:      buffer.Style{Fg: buffer.RGB(255, 255, 255)},
	}
	bh.SetStyle(custom)
	bh.SetUtilization(75, 100)
	buf := buffer.NewBuffer(30, 1)
	bh.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	bh.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintPipelineFlow(b *testing.B) {
	pf := NewPipelineFlow()
	pf.AddStage("Input", StageDone)
	pf.AddStage("Tokenize", StageDone)
	pf.AddStage("Embed", StageActive)
	pf.AddStage("Output", StagePending)
	pf.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 1})
	buf := buffer.NewBuffer(50, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pf.Paint(buf)
	}
}

func BenchmarkPaintThinkingBudget(b *testing.B) {
	tb := NewThinkingBudget()
	tb.SetBudget(500, 1000)
	tb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2})
	buf := buffer.NewBuffer(40, 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.Paint(buf)
	}
}

func BenchmarkPaintBufferHealthBar(b *testing.B) {
	bh := NewBufferHealthBar()
	bh.SetUtilization(65, 100)
	bh.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bh.Paint(buf)
	}
}

package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── TokenUsageGauge tests ───

func TestTokenUsageGaugeBasic(t *testing.T) {
	tug := NewTokenUsageGauge()
	tug.SetBudget(128000)
	tug.SetUsed(45000)
	tug.SetRate(85.5)
	if tug.Measure(Constraints{}).H < 3 { t.Error("H too small") }
}

func TestTokenUsageGaugeSetUsed(t *testing.T) {
	tug := NewTokenUsageGauge()
	tug.SetBudget(100)
	tug.SetUsed(50)
}

func TestTokenUsageGaugeOverflow(t *testing.T) {
	tug := NewTokenUsageGauge()
	tug.SetBudget(100)
	tug.SetUsed(500) // 500% should clamp to 100%
	tug.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	tug.Paint(buf)
}

func TestTokenUsageGaugePaint(t *testing.T) {
	tug := NewTokenUsageGauge()
	tug.SetBudget(1000)
	tug.SetUsed(750)
	tug.SetRate(50)
	tug.AdvancePulse()
	tug.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	tug.Paint(buf)
	// Check gauge brackets exist
	foundBracket := false
	for x := 0; x < 30; x++ {
		if buf.GetCell(x, 1).Rune == '[' { foundBracket = true; break }
	}
	if !foundBracket { t.Error("gauge bracket not found") }
}

func TestTokenUsageGaugeEmpty(t *testing.T) {
	tug := NewTokenUsageGauge()
	tug.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	tug.Paint(buf)
}

func TestTokenUsageGaugeChildren(t *testing.T) {
	tug := NewTokenUsageGauge()
	if tug.Children() != nil { t.Error("Children should be nil") }
}

func TestTokenUsageGaugeStyle(t *testing.T) {
	tug := NewTokenUsageGauge()
	tug.SetStyle(TokenUsageGaugeStyle{Normal: buffer.Style{Fg: buffer.RGB(0,255,0)}, Warning: buffer.Style{Fg: buffer.RGB(255,255,0)}, Critical: buffer.Style{Fg: buffer.RGB(255,0,0)}, Label: buffer.Style{Fg: buffer.RGB(150,150,150)}, Value: buffer.Style{Fg: buffer.RGB(255,255,255)}, Pulse: buffer.Style{Fg: buffer.RGB(0,0,255)}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	tug.SetUsed(50)
	tug.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	tug.Paint(buf)
}

// ─── CacheHitRatioBar tests ───

func TestCacheHitRatioBasic(t *testing.T) {
	ch := NewCacheHitRatioBar()
	ch.SetHits(850)
	ch.SetMisses(150)
	if ch.HitPercent() != 85 { t.Errorf("HitPercent = %d, want 85", ch.HitPercent()) }
}

func TestCacheHitRatioAllHits(t *testing.T) {
	ch := NewCacheHitRatioBar()
	ch.SetHits(100)
	ch.SetMisses(0)
	if ch.HitPercent() != 100 { t.Errorf("HitPercent = %d, want 100", ch.HitPercent()) }
}

func TestCacheHitRatioAllMisses(t *testing.T) {
	ch := NewCacheHitRatioBar()
	ch.SetHits(0)
	ch.SetMisses(100)
	if ch.HitPercent() != 0 { t.Errorf("HitPercent = %d, want 0", ch.HitPercent()) }
}

func TestCacheHitRatioEmpty(t *testing.T) {
	ch := NewCacheHitRatioBar()
	ch.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	buf := buffer.NewBuffer(50, 3)
	ch.Paint(buf)
}

func TestCacheHitRatioPaint(t *testing.T) {
	ch := NewCacheHitRatioBar()
	ch.SetHits(80)
	ch.SetMisses(20)
	ch.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	buf := buffer.NewBuffer(50, 3)
	ch.Paint(buf)
	foundHit := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 0).Rune == '█' { foundHit = true; break }
	}
	if !foundHit { t.Error("hit bar not found") }
}

func TestCacheHitRatioChildren(t *testing.T) {
	ch := NewCacheHitRatioBar()
	if ch.Children() != nil { t.Error("Children should be nil") }
}

func TestCacheHitRatioStyle(t *testing.T) {
	ch := NewCacheHitRatioBar()
	ch.SetStyle(CacheHitRatioStyle{Hit: buffer.Style{Fg: buffer.RGB(0,255,0)}, Miss: buffer.Style{Fg: buffer.RGB(255,0,0)}, Label: buffer.Style{Fg: buffer.RGB(150,150,150)}, Percent: buffer.Style{Fg: buffer.RGB(255,255,255), Flags: buffer.Bold}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	ch.SetHits(50)
	ch.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	buf := buffer.NewBuffer(50, 3)
	ch.Paint(buf)
}

// ─── PromptTemplateTree tests ───

func TestPromptTemplateTreeBasic(t *testing.T) {
	ptt := NewPromptTemplateTree()
	ptt.AddNode(0, "System", "You are {{role}}.", true)
	if ptt.NodeCount() != 1 { t.Errorf("NodeCount = %d, want 1", ptt.NodeCount()) }
}

func TestPromptTemplateTreeNested(t *testing.T) {
	ptt := NewPromptTemplateTree()
	ptt.AddNode(0, "Root", "", true)
	ptt.AddNode(1, "Child", "Hello {{name}}", true)
	ptt.AddNode(2, "Grandchild", "Score: {{score}}", true)
	if ptt.NodeCount() != 3 { t.Errorf("NodeCount = %d, want 3", ptt.NodeCount()) }
}

func TestPromptTemplateTreeToggle(t *testing.T) {
	ptt := NewPromptTemplateTree()
	ptt.AddNode(0, "X", "", false)
	ptt.ToggleExpand(0)
	ptt.mu.Lock()
	if !ptt.nodes[0].Expanded { t.Error("should be expanded after toggle") }
	ptt.mu.Unlock()
}

func TestPromptTemplateTreeClear(t *testing.T) {
	ptt := NewPromptTemplateTree()
	ptt.AddNode(0, "X", "", true)
	ptt.Clear()
	if ptt.NodeCount() != 0 { t.Errorf("NodeCount = %d, want 0", ptt.NodeCount()) }
}

func TestPromptTemplateTreeEmpty(t *testing.T) {
	ptt := NewPromptTemplateTree()
	if ptt.NodeCount() != 0 { t.Errorf("NodeCount = %d, want 0", ptt.NodeCount()) }
}

func TestPromptTemplateTreeMeasure(t *testing.T) {
	ptt := NewPromptTemplateTree()
	ptt.AddNode(0, "X", "y", true)
	s := ptt.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestPromptTemplateTreePaint(t *testing.T) {
	ptt := NewPromptTemplateTree()
	ptt.AddNode(0, "System", "You are {{role}}.", true)
	ptt.AddNode(1, "Greeting", "Hello {{user}}!", true)
	ptt.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 5})
	buf := buffer.NewBuffer(50, 5)
	ptt.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	// Check tree guide for nested node
	foundGuide := false
	for x := 0; x < 50; x++ {
		r := buf.GetCell(x, 2).Rune
		if r == '│' || r == '├' { foundGuide = true; break }
	}
	if !foundGuide { t.Error("tree guide not found") }
}

func TestPromptTemplateTreePaintEmpty(t *testing.T) {
	ptt := NewPromptTemplateTree()
	ptt.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	buf := buffer.NewBuffer(50, 3)
	ptt.Paint(buf)
}

func TestPromptTemplateTreeChildren(t *testing.T) {
	ptt := NewPromptTemplateTree()
	if ptt.Children() != nil { t.Error("Children should be nil") }
}

func TestPromptTemplateTreeStyle(t *testing.T) {
	ptt := NewPromptTemplateTree()
	ptt.SetStyle(PromptTemplateStyle{Label: buffer.Style{Fg: buffer.RGB(255,0,255)}, Var: buffer.Style{Fg: buffer.RGB(255,165,0)}, Text: buffer.Style{Fg: buffer.RGB(200,200,200)}, Guide: buffer.Style{Fg: buffer.RGB(80,80,80)}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	ptt.AddNode(0, "X", "{{v}}", true)
	ptt.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	ptt.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintTokenUsageGauge(b *testing.B) {
	tug := NewTokenUsageGauge()
	tug.SetBudget(128000)
	tug.SetUsed(45230)
	tug.SetRate(85.5)
	tug.AdvancePulse()
	tug.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tug.Paint(buf)
	}
}

func BenchmarkPaintCacheHitRatioBar(b *testing.B) {
	ch := NewCacheHitRatioBar()
	ch.SetHits(8543)
	ch.SetMisses(1457)
	ch.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 3})
	buf := buffer.NewBuffer(60, 3)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ch.Paint(buf)
	}
}

func BenchmarkPaintPromptTemplateTree(b *testing.B) {
	ptt := NewPromptTemplateTree()
	ptt.AddNode(0, "System", "You are a {{role}} assistant.", true)
	ptt.AddNode(1, "Context", "Given {{context}}, answer the user.", true)
	ptt.AddNode(2, "Format", "Respond in {{format}} style.", true)
	ptt.AddNode(0, "User", "{{user_input}}", true)
	ptt.AddNode(1, "History", "Previous: {{conversation}}", true)
	ptt.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 8})
	buf := buffer.NewBuffer(60, 8)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ptt.Paint(buf)
	}
}

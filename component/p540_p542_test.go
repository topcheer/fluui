package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── TokenEfficiencyBar Tests ───

func TestTokenEfficiencyBasic(t *testing.T) {
	bar := NewTokenEfficiencyBar()
	bar.SetTokens(500, 1000, 500)
	// 500/2000 = 25%
	if pct := bar.EfficiencyPercent(); pct != 25 {
		t.Errorf("EfficiencyPercent = %d, want 25", pct)
	}
}

func TestTokenEfficiencyZero(t *testing.T) {
	bar := NewTokenEfficiencyBar()
	bar.SetTokens(0, 0, 0)
	if pct := bar.EfficiencyPercent(); pct != 0 {
		t.Errorf("EfficiencyPercent = %d, want 0", pct)
	}
}

func TestTokenEfficiencyAllUseful(t *testing.T) {
	bar := NewTokenEfficiencyBar()
	bar.SetTokens(1000, 0, 0)
	if pct := bar.EfficiencyPercent(); pct != 100 {
		t.Errorf("EfficiencyPercent = %d, want 100", pct)
	}
}

func TestTokenEfficiencyNegative(t *testing.T) {
	bar := NewTokenEfficiencyBar()
	bar.SetTokens(-10, -5, -3)
	if pct := bar.EfficiencyPercent(); pct != 0 {
		t.Errorf("EfficiencyPercent = %d, want 0", pct)
	}
}

func TestTokenEfficiencyWidth(t *testing.T) {
	bar := NewTokenEfficiencyBar()
	bar.SetWidth(50)
	if bar.width != 50 {
		t.Errorf("width = %d, want 50", bar.width)
	}
	// Test min width clamp
	bar.SetWidth(5)
	if bar.width != 20 {
		t.Errorf("width = %d, want 20 (clamped)", bar.width)
	}
}

func TestTokenEfficiencyPaint(t *testing.T) {
	bar := NewTokenEfficiencyBar()
	bar.SetTokens(300, 700, 200)
	bar.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	buf := buffer.NewBuffer(50, 3)
	bar.Paint(buf)
	// Should have non-space cells
	hasContent := false
	for i := 0; i < 50; i++ {
		if buf.GetCell(i, 0).Rune != ' ' && buf.GetCell(i, 0).Rune != 0 {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Paint produced no visible bar content on row 0")
	}
}

func TestTokenEfficiencyChildren(t *testing.T) {
	bar := NewTokenEfficiencyBar()
	if children := bar.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestTokenEfficiencyStyle(t *testing.T) {
	bar := NewTokenEfficiencyBar()
	custom := TokenEfficiencyStyle{
		Useful:   buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Overhead: buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Label:    buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Score:    buffer.Style{Fg: buffer.RGB(255, 255, 255)},
	}
	bar.SetStyle(custom)
	// Just verify it doesn't panic
	bar.SetTokens(100, 200, 100)
	buf := buffer.NewBuffer(50, 3)
	bar.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	bar.Paint(buf)
}

// ─── ToolCallBadge Tests ───

func TestToolCallBadgeBasic(t *testing.T) {
	badge := NewToolCallBadge()
	badge.SetCalls(10, 2)
	if total := badge.TotalCalls(); total != 12 {
		t.Errorf("TotalCalls = %d, want 12", total)
	}
}

func TestToolCallBadgeAllSuccess(t *testing.T) {
	badge := NewToolCallBadge()
	badge.SetCalls(15, 0)
	if rate := badge.SuccessRate(); rate != 100 {
		t.Errorf("SuccessRate = %d, want 100", rate)
	}
}

func TestToolCallBadgeAllFailure(t *testing.T) {
	badge := NewToolCallBadge()
	badge.SetCalls(0, 8)
	if rate := badge.SuccessRate(); rate != 0 {
		t.Errorf("SuccessRate = %d, want 0", rate)
	}
}

func TestToolCallBadgeEmpty(t *testing.T) {
	badge := NewToolCallBadge()
	if total := badge.TotalCalls(); total != 0 {
		t.Errorf("TotalCalls = %d, want 0", total)
	}
	if rate := badge.SuccessRate(); rate != 0 {
		t.Errorf("SuccessRate = %d, want 0", rate)
	}
}

func TestToolCallBadgeNegative(t *testing.T) {
	badge := NewToolCallBadge()
	badge.SetCalls(-5, -3)
	if total := badge.TotalCalls(); total != 0 {
		t.Errorf("TotalCalls = %d, want 0 (clamped)", total)
	}
}

func TestToolCallBadgePaint(t *testing.T) {
	badge := NewToolCallBadge()
	badge.SetCalls(12, 3)
	badge.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	badge.Paint(buf)
	// First char should be '['
	c := buf.GetCell(0, 0)
	if c.Rune != '[' {
		t.Errorf("First cell rune = %q, want '['", c.Rune)
	}
}

func TestToolCallBadgeChildren(t *testing.T) {
	badge := NewToolCallBadge()
	if children := badge.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestToolCallBadgeStyle(t *testing.T) {
	badge := NewToolCallBadge()
	custom := ToolCallBadgeStyle{
		Success: buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Failure: buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Label:   buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Count:   buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Bracket: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	}
	badge.SetStyle(custom)
	badge.SetCalls(5, 1)
	buf := buffer.NewBuffer(30, 1)
	badge.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	badge.Paint(buf)
}

// ─── ModelLatencyGraph Tests ───

func TestModelLatencyGraphBasic(t *testing.T) {
	g := NewModelLatencyGraph()
	g.AddLatency(100)
	g.AddLatency(200)
	if cur := g.CurrentLatency(); cur != 200 {
		t.Errorf("CurrentLatency = %d, want 200", cur)
	}
}

func TestModelLatencyGraphAverage(t *testing.T) {
	g := NewModelLatencyGraph()
	g.AddLatency(100)
	g.AddLatency(200)
	g.AddLatency(300)
	if avg := g.AverageLatency(); avg != 200 {
		t.Errorf("AverageLatency = %d, want 200", avg)
	}
}

func TestModelLatencyGraphEmpty(t *testing.T) {
	g := NewModelLatencyGraph()
	if cur := g.CurrentLatency(); cur != 0 {
		t.Errorf("CurrentLatency = %d, want 0", cur)
	}
	if avg := g.AverageLatency(); avg != 0 {
		t.Errorf("AverageLatency = %d, want 0", avg)
	}
}

func TestModelLatencyGraphNegative(t *testing.T) {
	g := NewModelLatencyGraph()
	g.AddLatency(-50)
	if cur := g.CurrentLatency(); cur != 0 {
		t.Errorf("CurrentLatency = %d, want 0 (clamped)", cur)
	}
}

func TestModelLatencyGraphOverflow(t *testing.T) {
	g := NewModelLatencyGraph()
	for i := 1; i <= latencyMaxPoints+10; i++ {
		g.AddLatency(i * 10)
	}
	// Should still work, most recent = (latencyMaxPoints+10)*10
	expected := (latencyMaxPoints + 10) * 10
	if cur := g.CurrentLatency(); cur != expected {
		t.Errorf("CurrentLatency = %d, want %d", cur, expected)
	}
}

func TestModelLatencyGraphClear(t *testing.T) {
	g := NewModelLatencyGraph()
	g.AddLatency(100)
	g.AddLatency(200)
	g.Clear()
	if cur := g.CurrentLatency(); cur != 0 {
		t.Errorf("CurrentLatency after Clear = %d, want 0", cur)
	}
}

func TestModelLatencyGraphPaint(t *testing.T) {
	g := NewModelLatencyGraph()
	g.AddLatency(50)
	g.AddLatency(100)
	g.AddLatency(150)
	g.AddLatency(200)
	g.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 6})
	buf := buffer.NewBuffer(30, 6)
	g.Paint(buf)
	// Should have some bar content
	hasContent := false
	for row := 0; row < 5; row++ {
		for col := 0; col < 30; col++ {
			r := buf.GetCell(col, row).Rune
			if r == '▄' || r == '█' {
				hasContent = true
				break
			}
		}
		if hasContent { break }
	}
	if !hasContent {
		t.Error("Paint produced no bar content")
	}
}

func TestModelLatencyGraphPaintEmpty(t *testing.T) {
	g := NewModelLatencyGraph()
	g.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 6})
	buf := buffer.NewBuffer(30, 6)
	g.Paint(buf)
	// Should show "No latency data"
	c := buf.GetCell(0, 0)
	if c.Rune != 'N' {
		t.Errorf("First cell rune = %q, want 'N'", c.Rune)
	}
}

func TestModelLatencyGraphChildren(t *testing.T) {
	g := NewModelLatencyGraph()
	if children := g.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestModelLatencyGraphStyle(t *testing.T) {
	g := NewModelLatencyGraph()
	custom := LatencyGraphStyle{
		Line:    buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Fill:    buffer.Style{Fg: buffer.RGB(0, 100, 0)},
		Label:   buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Peak:    buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Current: buffer.Style{Fg: buffer.RGB(255, 255, 255)},
	}
	g.SetStyle(custom)
	g.AddLatency(100)
	buf := buffer.NewBuffer(30, 6)
	g.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 6})
	g.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintTokenEfficiencyBar(b *testing.B) {
	bar := NewTokenEfficiencyBar()
	bar.SetTokens(500, 1000, 500)
	bar.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	buf := buffer.NewBuffer(50, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bar.Paint(buf)
	}
}

func BenchmarkPaintToolCallBadge(b *testing.B) {
	badge := NewToolCallBadge()
	badge.SetCalls(12, 3)
	badge.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		badge.Paint(buf)
	}
}

func BenchmarkPaintModelLatencyGraph(b *testing.B) {
	g := NewModelLatencyGraph()
	for i := 1; i <= 20; i++ {
		g.AddLatency(i * 15)
	}
	g.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 6})
	buf := buffer.NewBuffer(40, 6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Paint(buf)
	}
}

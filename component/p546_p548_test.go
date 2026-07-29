package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── PromptVariant Tests ───

func TestPromptVariantBasic(t *testing.T) {
	pv := NewPromptVariant()
	pv.SetVariantA("A", 85, 1200, false)
	pv.SetVariantB("B", 92, 980, true)
	// Just verify it doesn't panic
}

func TestPromptVariantNoWinner(t *testing.T) {
	pv := NewPromptVariant()
	pv.SetVariantA("A", 85, 1200, false)
	pv.SetVariantB("B", 85, 1200, false)
	if pv.aLabelStr != "A" {
		t.Errorf("aLabelStr = %q, want 'A'", pv.aLabelStr)
	}
}

func TestPromptVariantWinnerA(t *testing.T) {
	pv := NewPromptVariant()
	pv.SetVariantA("A", 95, 1000, true)
	pv.SetVariantB("B", 80, 1200, false)
	if pv.aLabelStr != "A ✓" {
		t.Errorf("aLabelStr = %q, want 'A ✓'", pv.aLabelStr)
	}
}

func TestPromptVariantZero(t *testing.T) {
	pv := NewPromptVariant()
	pv.SetVariantA("", 0, 0, false)
	pv.SetVariantB("", 0, 0, false)
	// Should not panic
}

func TestPromptVariantWidth(t *testing.T) {
	pv := NewPromptVariant()
	pv.SetWidth(50)
	if pv.width != 50 {
		t.Errorf("width = %d, want 50", pv.width)
	}
	pv.SetWidth(5)
	if pv.width != 20 {
		t.Errorf("width = %d, want 20 (clamped)", pv.width)
	}
}

func TestPromptVariantPaint(t *testing.T) {
	pv := NewPromptVariant()
	pv.SetVariantA("A", 85, 1200, false)
	pv.SetVariantB("B", 92, 980, true)
	pv.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	pv.Paint(buf)
	// Check divider exists
	hasDivider := false
	for i := 0; i < 40; i++ {
		if buf.GetCell(i, 0).Rune == '│' {
			hasDivider = true
			break
		}
	}
	if !hasDivider {
		t.Error("Paint should have a divider character")
	}
}

func TestPromptVariantChildren(t *testing.T) {
	pv := NewPromptVariant()
	if children := pv.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestPromptVariantStyle(t *testing.T) {
	pv := NewPromptVariant()
	custom := PromptVariantStyle{
		Label:   buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Value:   buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Winner:  buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Loser:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Divider: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	}
	pv.SetStyle(custom)
	pv.SetVariantA("X", 70, 500, true)
	pv.SetVariantB("Y", 65, 450, false)
	buf := buffer.NewBuffer(40, 3)
	pv.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	pv.Paint(buf)
}

// ─── TokenCostChart Tests ───

func TestTokenCostChartBasic(t *testing.T) {
	c := NewTokenCostChart()
	c.AddCost(10)
	c.AddCost(20)
	c.AddCost(5)
	total := c.TotalCost()
	if total != 35 {
		t.Errorf("TotalCost = %d, want 35", total)
	}
}

func TestTokenCostChartEmpty(t *testing.T) {
	c := NewTokenCostChart()
	if total := c.TotalCost(); total != 0 {
		t.Errorf("TotalCost = %d, want 0", total)
	}
}

func TestTokenCostChartNegative(t *testing.T) {
	c := NewTokenCostChart()
	c.AddCost(-5)
	if total := c.TotalCost(); total != 0 {
		t.Errorf("TotalCost = %d, want 0 (clamped)", total)
	}
}

func TestTokenCostChartOverflow(t *testing.T) {
	c := NewTokenCostChart()
	for i := 1; i <= costChartMaxPoints+10; i++ {
		c.AddCost(i)
	}
	// Should not panic, circular buffer handles overflow
}

func TestTokenCostChartClear(t *testing.T) {
	c := NewTokenCostChart()
	c.AddCost(10)
	c.AddCost(20)
	c.Clear()
	if total := c.TotalCost(); total != 0 {
		t.Errorf("TotalCost after Clear = %d, want 0", total)
	}
}

func TestTokenCostChartPaint(t *testing.T) {
	c := NewTokenCostChart()
	c.AddCost(5)
	c.AddCost(10)
	c.AddCost(15)
	c.AddCost(3)
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 6})
	buf := buffer.NewBuffer(30, 6)
	c.Paint(buf)
	// Should have bar content
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

func TestTokenCostChartPaintEmpty(t *testing.T) {
	c := NewTokenCostChart()
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 6})
	buf := buffer.NewBuffer(30, 6)
	c.Paint(buf)
	if r := buf.GetCell(0, 0).Rune; r != 'N' {
		t.Errorf("First rune = %q, want 'N'", r)
	}
}

func TestTokenCostChartChildren(t *testing.T) {
	c := NewTokenCostChart()
	if children := c.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestTokenCostChartStyle(t *testing.T) {
	c := NewTokenCostChart()
	custom := TokenCostChartStyle{
		Bar:   buffer.Style{Fg: buffer.RGB(255, 215, 0)},
		Peak:  buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Label: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Total: buffer.Style{Fg: buffer.RGB(255, 255, 255)},
	}
	c.SetStyle(custom)
	c.AddCost(100)
	buf := buffer.NewBuffer(30, 6)
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 6})
	c.Paint(buf)
}

// ─── AISafetyBadge Tests ───

func TestAISafetyBadgeBasic(t *testing.T) {
	b := NewAISafetyBadge()
	if level := b.Level(); level != SafetyUnknown {
		t.Errorf("Level = %d, want SafetyUnknown(%d)", level, SafetyUnknown)
	}
}

func TestAISafetyBadgeSafe(t *testing.T) {
	b := NewAISafetyBadge()
	b.SetClassification(SafetySafe, "clean")
	if level := b.Level(); level != SafetySafe {
		t.Errorf("Level = %d, want SafetySafe(%d)", level, SafetySafe)
	}
}

func TestAISafetyBadgeWarning(t *testing.T) {
	b := NewAISafetyBadge()
	b.SetClassification(SafetyWarning, "sensitive")
	if level := b.Level(); level != SafetyWarning {
		t.Errorf("Level = %d, want SafetyWarning(%d)", level, SafetyWarning)
	}
}

func TestAISafetyBadgeBlocked(t *testing.T) {
	b := NewAISafetyBadge()
	b.SetClassification(SafetyBlocked, "harmful")
	if level := b.Level(); level != SafetyBlocked {
		t.Errorf("Level = %d, want SafetyBlocked(%d)", level, SafetyBlocked)
	}
}

func TestAISafetyBadgeInvalidLevel(t *testing.T) {
	b := NewAISafetyBadge()
	b.SetClassification(SafetyLevel(99), "test")
	if level := b.Level(); level != SafetyUnknown {
		t.Errorf("Level = %d, want SafetyUnknown (clamped)", level)
	}
}

func TestAISafetyBadgeNoCategory(t *testing.T) {
	b := NewAISafetyBadge()
	b.SetClassification(SafetySafe, "")
	// Label should just be "SAFE" without category
	if b.labelStr != "SAFE" {
		t.Errorf("labelStr = %q, want 'SAFE'", b.labelStr)
	}
}

func TestAISafetyBadgePaint(t *testing.T) {
	b := NewAISafetyBadge()
	b.SetClassification(SafetySafe, "clean")
	b.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.Paint(buf)
	if r := buf.GetCell(0, 0).Rune; r != '[' {
		t.Errorf("First rune = %q, want '['", r)
	}
	// Second should be ✓ icon
	if r := buf.GetCell(1, 0).Rune; r != '✓' {
		t.Errorf("Second rune = %q, want '✓'", r)
	}
}

func TestAISafetyBadgeChildren(t *testing.T) {
	b := NewAISafetyBadge()
	if children := b.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestAISafetyBadgeStyle(t *testing.T) {
	b := NewAISafetyBadge()
	custom := AISafetyBadgeStyle{
		Safe:    buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Warning: buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Blocked: buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Unknown: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Label:   buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Bracket: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	}
	b.SetStyle(custom)
	b.SetClassification(SafetyBlocked, "spam")
	buf := buffer.NewBuffer(20, 1)
	b.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	b.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintPromptVariant(b *testing.B) {
	pv := NewPromptVariant()
	pv.SetVariantA("A", 85, 1200, false)
	pv.SetVariantB("B", 92, 980, true)
	pv.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pv.Paint(buf)
	}
}

func BenchmarkPaintTokenCostChart(b *testing.B) {
	c := NewTokenCostChart()
	for i := 1; i <= 15; i++ {
		c.AddCost(i * 3)
	}
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 6})
	buf := buffer.NewBuffer(30, 6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkPaintAISafetyBadge(b *testing.B) {
	badge := NewAISafetyBadge()
	badge.SetClassification(SafetySafe, "clean")
	badge.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		badge.Paint(buf)
	}
}

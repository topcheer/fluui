package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestCostTracker_New_P464(t *testing.T) {
	ct := NewCostTracker()
	if ct.Cost() != 0 {
		t.Errorf("Cost = %v, want 0", ct.Cost())
	}
	if ct.TotalTokens() != 0 {
		t.Errorf("TotalTokens = %d", ct.TotalTokens())
	}
}

func TestCostTracker_SetPricing_P464(t *testing.T) {
	ct := NewCostTracker()
	ct.SetPricing(15, 60)
	ct.SetTokens(1000000, 500000)
	cost := ct.Cost()
	// 1M * $15 + 0.5M * $60 = 15 + 30 = 45
	if cost < 44.9 || cost > 45.1 {
		t.Errorf("Cost = %v, want ~45", cost)
	}
}

func TestCostTracker_AddTokens_P464(t *testing.T) {
	ct := NewCostTracker()
	ct.AddTokens(100, 200)
	ct.AddTokens(300, 400)
	if ct.InputTokens() != 400 {
		t.Errorf("Input = %d", ct.InputTokens())
	}
	if ct.OutputTokens() != 600 {
		t.Errorf("Output = %d", ct.OutputTokens())
	}
}

func TestCostTracker_TotalTokens_P464(t *testing.T) {
	ct := NewCostTracker()
	ct.SetTokens(1000, 2000)
	if ct.TotalTokens() != 3000 {
		t.Errorf("Total = %d", ct.TotalTokens())
	}
}

func TestCostTracker_Budget_P464(t *testing.T) {
	ct := NewCostTracker()
	ct.SetPricing(10, 30)
	ct.SetBudget(5.0)
	ct.SetTokens(100000, 100000) // cost = 1 + 3 = 4, under budget
	if ct.IsOverBudget() {
		t.Error("should not be over budget")
	}
	ct.SetTokens(200000, 200000) // cost = 2 + 6 = 8, over budget
	if !ct.IsOverBudget() {
		t.Error("should be over budget")
	}
}

func TestCostTracker_BudgetPercent_P464(t *testing.T) {
	ct := NewCostTracker()
	ct.SetPricing(10, 0)
	ct.SetBudget(10.0)
	ct.SetTokens(500000, 0) // cost = 5, 50% of 10
	pct := ct.BudgetPercent()
	if pct < 49.9 || pct > 50.1 {
		t.Errorf("BudgetPercent = %v, want ~50", pct)
	}
}

func TestCostTracker_NoBudget_P464(t *testing.T) {
	ct := NewCostTracker()
	ct.SetTokens(1000, 1000)
	if ct.IsOverBudget() {
		t.Error("no budget = never over")
	}
	if ct.BudgetPercent() != 0 {
		t.Errorf("BudgetPercent = %v, want 0", ct.BudgetPercent())
	}
}

func TestCostTracker_Style_P464(t *testing.T) {
	ct := NewCostTracker()
	ct.SetStyle(DefaultCostTrackerStyle())
}

func TestCostTracker_Measure_P464(t *testing.T) {
	ct := NewCostTracker()
	sz := ct.Measure(Constraints{})
	if sz.H != 1 {
		t.Errorf("H = %d", sz.H)
	}
}

func TestCostTracker_Paint_NoPanic_P464(t *testing.T) {
	ct := NewCostTracker()
	ct.SetPricing(15, 60)
	ct.SetTokens(50000, 12000)
	ct.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 1})
	buf := buffer.NewBuffer(50, 1)
	ct.Paint(buf)
}

func TestCostTracker_Paint_OverBudget_P464(t *testing.T) {
	ct := NewCostTracker()
	ct.SetPricing(100, 300)
	ct.SetBudget(1.0)
	ct.SetTokens(100000, 100000)
	ct.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	ct.Paint(buf)
}

func TestCostTracker_Paint_ZeroBounds_P464(t *testing.T) {
	ct := NewCostTracker()
	ct.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	ct.Paint(buf)
}

func TestCostTracker_Children_P464(t *testing.T) {
	if NewCostTracker().Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestFormatCost_P464(t *testing.T) {
	if formatCost(0) != "0.000000" {
		t.Errorf("formatCost(0) = %q", formatCost(0))
	}
	if formatCost(1.5) != "1.50" {
		t.Errorf("formatCost(1.5) = %q", formatCost(1.5))
	}
}

func BenchmarkCostTracker_Paint_P464(b *testing.B) {
	ct := NewCostTracker()
	ct.SetPricing(15, 60)
	ct.SetTokens(50000, 12000)
	ct.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 1})
	buf := buffer.NewBuffer(50, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ct.Paint(buf)
	}
}

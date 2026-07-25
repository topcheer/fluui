package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P296: TokenUsageWidget tests

func TestP296_NewTokenUsageWidget(t *testing.T) {
	w := NewTokenUsageWidget("gpt-4")
	if w.Model() != "gpt-4" {
		t.Errorf("Model = %q", w.Model())
	}
	if w.ID() == "" {
		t.Error("ID should not be empty")
	}
}

func TestP296_SetModel(t *testing.T) {
	w := NewTokenUsageWidget("claude-3")
	w.SetModel("claude-3.5-sonnet")
	if w.Model() != "claude-3.5-sonnet" {
		t.Errorf("Model = %q", w.Model())
	}
}

func TestP296_AddTokens(t *testing.T) {
	w := NewTokenUsageWidget("gpt-4")
	w.AddTokens(100, 50)
	w.AddTokens(200, 100)
	if w.InputTokens() != 300 {
		t.Errorf("InputTokens = %d, want 300", w.InputTokens())
	}
	if w.OutputTokens() != 150 {
		t.Errorf("OutputTokens = %d, want 150", w.OutputTokens())
	}
	if w.TotalTokens() != 450 {
		t.Errorf("TotalTokens = %d, want 450", w.TotalTokens())
	}
}

func TestP296_SetTokens(t *testing.T) {
	w := NewTokenUsageWidget("gpt-4")
	w.AddTokens(1000, 500)
	w.SetTokens(100, 50)
	if w.InputTokens() != 100 {
		t.Errorf("InputTokens = %d, want 100", w.InputTokens())
	}
	if w.OutputTokens() != 50 {
		t.Errorf("OutputTokens = %d, want 50", w.OutputTokens())
	}
}

func TestP296_SetContextUsage(t *testing.T) {
	w := NewTokenUsageWidget("gpt-4")
	w.SetContextUsage(5000, 128000)
	if w.ContextUsed() != 5000 {
		t.Errorf("ContextUsed = %d", w.ContextUsed())
	}
	if w.ContextTotal() != 128000 {
		t.Errorf("ContextTotal = %d", w.ContextTotal())
	}
}

func TestP296_ContextPercent(t *testing.T) {
	w := NewTokenUsageWidget("gpt-4")
	w.SetContextUsage(64000, 128000)
	if pct := w.ContextPercent(); pct != 50 {
		t.Errorf("ContextPercent = %.1f, want 50", pct)
	}
	// zero total
	w.SetContextUsage(100, 0)
	if pct := w.ContextPercent(); pct != 0 {
		t.Errorf("ContextPercent = %.1f, want 0", pct)
	}
	// over 100%
	w.SetContextUsage(200000, 100000)
	if pct := w.ContextPercent(); pct != 100 {
		t.Errorf("ContextPercent = %.1f, want 100 (clamped)", pct)
	}
}

func TestP296_SetPricing(t *testing.T) {
	w := NewTokenUsageWidget("gpt-4")
	w.SetPricing(TokenPricing{InputPerMillion: 10, OutputPerMillion: 30})
	w.SetTokens(1_000_000, 1_000_000)
	cost := w.EstimatedCost()
	if cost != 40 {
		t.Errorf("Cost = %.2f, want 40.00", cost)
	}
}

func TestP296_EstimatedCost_DefaultPricing(t *testing.T) {
	w := NewTokenUsageWidget("gpt-4")
	w.SetTokens(100_000, 50_000)
	// Default: $3/M input, $15/M output
	// input: 100000/1M * 3 = 0.30
	// output: 50000/1M * 15 = 0.75
	// total: 1.05
	cost := w.EstimatedCost()
	if cost < 1.04 || cost > 1.06 {
		t.Errorf("Cost = %.4f, want ~1.05", cost)
	}
}

func TestP296_Reset(t *testing.T) {
	w := NewTokenUsageWidget("gpt-4")
	w.AddTokens(1000, 500)
	w.SetContextUsage(5000, 128000)
	w.Reset()
	if w.InputTokens() != 0 || w.OutputTokens() != 0 {
		t.Error("tokens should be zero after reset")
	}
	if w.ContextUsed() != 0 {
		t.Error("context should be zero after reset")
	}
	if w.ContextTotal() != 128000 {
		t.Error("context total should NOT be reset")
	}
}

func TestP296_Measure(t *testing.T) {
	w := NewTokenUsageWidget("gpt-4")
	s := w.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if s.H != 1 {
		t.Errorf("H = %d, want 1", s.H)
	}
	if s.W != 80 {
		t.Errorf("W = %d, want 80", s.W)
	}
	// default width
	s2 := w.Measure(Constraints{})
	if s2.W != 80 {
		t.Errorf("default W = %d, want 80", s2.W)
	}
}

func TestP296_Paint(t *testing.T) {
	w := NewTokenUsageWidget("gpt-4")
	w.AddTokens(1500, 800)
	w.SetContextUsage(64000, 128000)
	w.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	w.Paint(buf)
}

func TestP296_Paint_ZeroBounds(t *testing.T) {
	w := NewTokenUsageWidget("gpt-4")
	w.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	w.Paint(buf) // should not panic
}

func TestP296_Paint_NarrowWidth(t *testing.T) {
	w := NewTokenUsageWidget("very-long-model-name-here")
	w.AddTokens(100000, 50000)
	w.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})
	buf := buffer.NewBuffer(5, 1)
	w.Paint(buf) // should not panic, should truncate
}

func TestP296_Paint_NoContext(t *testing.T) {
	w := NewTokenUsageWidget("gpt-4")
	w.AddTokens(500, 200)
	// No context set — should still render without ctx bar
	w.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	w.Paint(buf)
}

func TestP296_Paint_NonZeroOffset(t *testing.T) {
	w := NewTokenUsageWidget("gpt-4")
	w.AddTokens(500, 200)
	w.SetContextUsage(1000, 8000)
	w.SetBounds(Rect{X: 10, Y: 5, W: 50, H: 1})
	buf := buffer.NewBuffer(70, 10)
	w.Paint(buf)
}

func TestP296_Concurrent(t *testing.T) {
	w := NewTokenUsageWidget("gpt-4")
	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			w.AddTokens(1, 1)
		}
		close(done)
	}()
	for i := 0; i < 500; i++ {
		_ = w.TotalTokens()
		_ = w.ContextPercent()
	}
	<-done
}

func TestP296_SatisfiesComponent(t *testing.T) {
	var _ Component = (*TokenUsageWidget)(nil)
}

// P370: Comprehensive Paint coverage — early-return paths, context bar, edges

func TestP370_Paint_EmptyModel(t *testing.T) {
	w := NewTokenUsageWidget("")
	w.AddTokens(100, 50)
	w.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	w.Paint(buf) // should show "unknown" without panic
}

func TestP370_Paint_NarrowWidths(t *testing.T) {
	// Test various widths to exercise all early-return paths in Paint
	for width := 1; width <= 35; width++ {
		w := NewTokenUsageWidget("gpt-4")
		w.AddTokens(1500, 800)
		w.SetContextUsage(64000, 128000)
		w.SetBounds(Rect{X: 0, Y: 0, W: width, H: 1})
		buf := buffer.NewBuffer(width, 1)
		w.Paint(buf) // must not panic at any width
	}
}

func TestP370_Paint_LargeTokens(t *testing.T) {
	w := NewTokenUsageWidget("gpt-4")
	w.AddTokens(2_500_000, 1_800_000) // M-range tokens
	w.SetContextUsage(100000, 128000)
	w.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	buf := buffer.NewBuffer(80, 1)
	w.Paint(buf)
}

func TestP370_Paint_ContextBarFull(t *testing.T) {
	w := NewTokenUsageWidget("claude-3")
	w.AddTokens(500, 200)
	w.SetContextUsage(128000, 128000) // 100%
	w.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	buf := buffer.NewBuffer(80, 1)
	w.Paint(buf)
}

func TestP370_ctxPercentLocked_Negative(t *testing.T) {
	w := NewTokenUsageWidget("gpt-4")
	// negative used → pct clamped to 0 via pct<0 branch
	w.SetContextUsage(-1000, 10000)
	if pct := w.ContextPercent(); pct != 0 {
		t.Errorf("ContextPercent with negative used = %.1f, want 0", pct)
	}
}

package component

import (
	"strconv"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── CostTracker: Real-time AI API Cost Estimator ───
//
// CostTracker renders a compact display of accumulated API costs based on
// token usage and per-million pricing. Common in AI dashboards and budget
// monitoring tools.
//
// Usage:
//
//	ct := NewCostTracker()
//	ct.SetPricing(15.0, 60.0) // $15/M input, $60/M output
//	ct.AddTokens(50000, 12000) // 50K input, 12K output
//	ct.Paint(buf) // renders "$0.18 (50K in + 12K out)"

// CostTrackerStyle holds visual styles.
type CostTrackerStyle struct {
	Cost     buffer.Style
	Label    buffer.Style
	Tokens   buffer.Style
	Warning  buffer.Style
}

// DefaultCostTrackerStyle returns sensible defaults.
func DefaultCostTrackerStyle() CostTrackerStyle {
	return CostTrackerStyle{
		Cost:    buffer.Style{Fg: buffer.RGB(255, 215, 0), Flags: buffer.Bold},  // gold
		Label:   buffer.Style{Fg: buffer.RGB(150, 150, 150)},
		Tokens:  buffer.Style{Fg: buffer.RGB(100, 200, 100)},
		Warning: buffer.Style{Fg: buffer.RGB(220, 80, 80), Flags: buffer.Bold},
	}
}

// CostTracker renders accumulated AI API costs.
type CostTracker struct {
	BaseComponent
	mu          sync.RWMutex
	inputTokens int
	outputTokens int
	priceIn     float64 // per million tokens
	priceOut    float64 // per million tokens
	budget      float64 // optional budget limit
	style       CostTrackerStyle
}

// NewCostTracker creates a cost tracker with default pricing ($0).
func NewCostTracker() *CostTracker {
	ct := &CostTracker{
		style: DefaultCostTrackerStyle(),
	}
	ct.SetID(GenerateID("cost"))
	return ct
}

// SetPricing sets input/output price per million tokens.
func (ct *CostTracker) SetPricing(inputPerM, outputPerM float64) *CostTracker {
	ct.mu.Lock()
	ct.priceIn = inputPerM
	ct.priceOut = outputPerM
	ct.mu.Unlock()
	return ct
}

// AddTokens accumulates token usage.
func (ct *CostTracker) AddTokens(input, output int) *CostTracker {
	ct.mu.Lock()
	ct.inputTokens += input
	ct.outputTokens += output
	ct.mu.Unlock()
	return ct
}

// SetTokens sets exact token counts.
func (ct *CostTracker) SetTokens(input, output int) *CostTracker {
	ct.mu.Lock()
	ct.inputTokens = input
	ct.outputTokens = output
	ct.mu.Unlock()
	return ct
}

// InputTokens returns input token count.
func (ct *CostTracker) InputTokens() int {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.inputTokens
}

// OutputTokens returns output token count.
func (ct *CostTracker) OutputTokens() int {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.outputTokens
}

// TotalTokens returns input + output.
func (ct *CostTracker) TotalTokens() int {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.inputTokens + ct.outputTokens
}

// Cost returns the estimated cost in USD.
func (ct *CostTracker) Cost() float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	in := float64(ct.inputTokens) / 1e6 * ct.priceIn
	out := float64(ct.outputTokens) / 1e6 * ct.priceOut
	return in + out
}

// SetBudget sets an optional budget limit for warning display.
func (ct *CostTracker) SetBudget(b float64) *CostTracker {
	ct.mu.Lock()
	ct.budget = b
	ct.mu.Unlock()
	return ct
}

// Budget returns the budget limit (0 if not set).
func (ct *CostTracker) Budget() float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.budget
}

// IsOverBudget returns true if cost exceeds budget.
func (ct *CostTracker) IsOverBudget() bool {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	if ct.budget <= 0 {
		return false
	}
	cost := float64(ct.inputTokens)/1e6*ct.priceIn + float64(ct.outputTokens)/1e6*ct.priceOut
	return cost > ct.budget
}

// BudgetPercent returns cost/budget * 100 (0 if no budget).
func (ct *CostTracker) BudgetPercent() float64 {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	if ct.budget <= 0 {
		return 0
	}
	cost := float64(ct.inputTokens)/1e6*ct.priceIn + float64(ct.outputTokens)/1e6*ct.priceOut
	return cost / ct.budget * 100
}

// SetStyle sets the visual style.
func (ct *CostTracker) SetStyle(s CostTrackerStyle) *CostTracker {
	ct.mu.Lock()
	ct.style = s
	ct.mu.Unlock()
	return ct
}

// Measure computes the desired size.
func (ct *CostTracker) Measure(cs Constraints) Size {
	return Size{W: 40, H: 1}
}

// Paint renders the cost tracker.
func (ct *CostTracker) Paint(buf *buffer.Buffer) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	b := ct.bounds
	if b.W < 4 || b.H < 1 {
		return
	}

	cost := float64(ct.inputTokens)/1e6*ct.priceIn + float64(ct.outputTokens)/1e6*ct.priceOut

	var costStyle buffer.Style
	if ct.budget > 0 && cost > ct.budget {
		costStyle = ct.style.Warning
	} else {
		costStyle = ct.style.Cost
	}

	x := b.X
	// Cost
	costStr := "$" + formatCost(cost)
	for _, r := range costStr {
		if x >= b.X+b.W {
			break
		}
		buf.SetCell(x, b.Y, buffer.Cell{Rune: r, Fg: costStyle.Fg, Bg: costStyle.Bg, Flags: costStyle.Flags, Width: 1})
		x++
	}

	// Token breakdown
	if x < b.X+b.W {
		detail := " (" + formatTokenK(ct.inputTokens) + " in +" + formatTokenK(ct.outputTokens) + " out)"
		for _, r := range detail {
			if x >= b.X+b.W {
				break
			}
			buf.SetCell(x, b.Y, buffer.Cell{Rune: r, Fg: ct.style.Tokens.Fg, Bg: ct.style.Tokens.Bg, Width: 1})
			x++
		}
	}

	// Budget warning
	if ct.budget > 0 && cost > ct.budget && x < b.X+b.W {
		pct := cost / ct.budget * 100
		warn := " ⚠" + strconv.FormatFloat(pct, 'f', 0, 64) + "%"
		for _, r := range warn {
			if x >= b.X+b.W {
				break
			}
			buf.SetCell(x, b.Y, buffer.Cell{Rune: r, Fg: ct.style.Warning.Fg, Bg: ct.style.Warning.Bg, Flags: ct.style.Warning.Flags, Width: 1})
			x++
		}
	}
}

// formatCost formats USD compactly.
func formatCost(v float64) string {
	if v >= 1 {
		return strconv.FormatFloat(v, 'f', 2, 64)
	}
	if v >= 0.01 {
		return strconv.FormatFloat(v, 'f', 4, 64)
	}
	return strconv.FormatFloat(v, 'f', 6, 64)
}

// Children returns nil.
func (ct *CostTracker) Children() []Component { return nil }

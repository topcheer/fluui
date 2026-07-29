package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ThinkingBudget: AI Reasoning Token Budget Indicator ───
//
// ThinkingBudget renders a budget bar showing how many reasoning/thinking
// tokens have been used versus the allocated budget. Displays a percentage
// and color-codes based on utilization.
//
// Usage:
//
//	tb := NewThinkingBudget()
//	tb.SetBudget(500, 1000) // 500 used out of 1000 budget
//	tb.Paint(buf)

// ThinkingBudgetStyle holds styling.
type ThinkingBudgetStyle struct {
	Normal  buffer.Style
	High    buffer.Style
	Critical buffer.Style
	Label   buffer.Style
	Value   buffer.Style
}

// DefaultThinkingBudgetStyle returns defaults.
func DefaultThinkingBudgetStyle() ThinkingBudgetStyle {
	return ThinkingBudgetStyle{
		Normal:   buffer.Style{Fg: buffer.RGB(168, 85, 247)},
		High:     buffer.Style{Fg: buffer.RGB(245, 158, 11)},
		Critical: buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Label:    buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Value:    buffer.Style{Fg: buffer.RGB(226, 232, 240)},
	}
}

// ThinkingBudget renders a reasoning token budget indicator.
type ThinkingBudget struct {
	BaseComponent
	mu sync.Mutex

	used   int
	budget int
	width  int
	style  ThinkingBudgetStyle
	// cached
	usedStr   string
	budgetStr string
	pctStr    string
	detailStr string
	barFill   int // 0-20 segments
	curStyle  buffer.Style
}

// NewThinkingBudget creates a ThinkingBudget.
func NewThinkingBudget() *ThinkingBudget {
	tb := &ThinkingBudget{width: 28, budget: 1000, style: DefaultThinkingBudgetStyle()}
	tb.SetID(GenerateID("thinkbudget"))
	tb.recomputeLocked()
	return tb
}

// SetBudget sets used tokens and total budget.
func (tb *ThinkingBudget) SetBudget(used, budget int) *ThinkingBudget {
	tb.mu.Lock()
	if used < 0 { used = 0 }
	if budget < 1 { budget = 1 }
	if used > budget { used = budget }
	tb.used = used
	tb.budget = budget
	tb.recomputeLocked()
	tb.mu.Unlock()
	return tb
}

func (tb *ThinkingBudget) recomputeLocked() {
	pct := tb.used * 100 / tb.budget
	tb.usedStr = itoa(tb.used)
	tb.budgetStr = itoa(tb.budget)
	tb.pctStr = itoa(pct) + "%"
	tb.detailStr = tb.usedStr + "/" + tb.budgetStr + " (" + tb.pctStr + ")"

	const barMax = 20
	tb.barFill = tb.used * barMax / tb.budget

	if pct >= 90 {
		tb.curStyle = tb.style.Critical
	} else if pct >= 70 {
		tb.curStyle = tb.style.High
	} else {
		tb.curStyle = tb.style.Normal
	}
}

// Percent returns the budget utilization percentage.
func (tb *ThinkingBudget) Percent() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.used * 100 / tb.budget
}

// SetWidth sets the bar width.
func (tb *ThinkingBudget) SetWidth(w int) *ThinkingBudget {
	tb.mu.Lock()
	if w < 15 { w = 15 }
	tb.width = w
	tb.mu.Unlock()
	return tb
}

// SetStyle sets custom style.
func (tb *ThinkingBudget) SetStyle(s ThinkingBudgetStyle) *ThinkingBudget {
	tb.mu.Lock()
	tb.style = s
	tb.recomputeLocked()
	tb.mu.Unlock()
	return tb
}

// Measure returns preferred size.
func (tb *ThinkingBudget) Measure(cs Constraints) Size {
	w := tb.width + 14
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 2}
}

// Paint renders the thinking budget indicator.
func (tb *ThinkingBudget) Paint(buf *buffer.Buffer) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	b := tb.Bounds()
	x, y := b.X, b.Y
	w := b.W
	if w < 15 { w = 40 }

	labelStyle := tb.style.Label
	valueStyle := tb.style.Value
	barStyle := tb.curStyle

	// Row 0: label + bar
	col := x
	for _, r := range "Think " {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for i := 0; i < tb.barFill; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '▰', Fg: barStyle.Fg, Bg: barStyle.Bg, Flags: barStyle.Flags, Width: 1})
		col++
	}
	for i := tb.barFill; i < 20; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '▱', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Row 1: values
	detail := tb.detailStr
	col = x
	for _, r := range detail {
		if col >= buf.Width { break }
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (tb *ThinkingBudget) Children() []Component { return nil }

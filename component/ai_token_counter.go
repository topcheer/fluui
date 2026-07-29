package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AITokenCounter: Real-time Token Count Display ───
//
// AITokenCounter renders a compact live-updating token counter showing
// prompt tokens, completion tokens, and total with a per-token cost estimate.
// Designed for real-time display during AI streaming responses.
//
// Usage:
//
//	tc := NewAITokenCounter()
//	tc.SetCounts(500, 350) // 500 prompt, 350 completion
//	tc.SetRate(0.002)      // $0.002 per 1K tokens
//	tc.Paint(buf)

// AITokenCounterStyle holds styling.
type AITokenCounterStyle struct {
	Prompt     buffer.Style
	Completion buffer.Style
	Total      buffer.Style
	Cost       buffer.Style
	Separator  buffer.Style
	Label      buffer.Style
}

// DefaultAITokenCounterStyle returns defaults.
func DefaultAITokenCounterStyle() AITokenCounterStyle {
	return AITokenCounterStyle{
		Prompt:     buffer.Style{Fg: buffer.RGB(96, 165, 250)},
		Completion: buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Total:      buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
		Cost:       buffer.Style{Fg: buffer.RGB(234, 179, 8)},
		Separator:  buffer.Style{Fg: buffer.RGB(71, 85, 105)},
		Label:      buffer.Style{Fg: buffer.RGB(148, 163, 184)},
	}
}

// AITokenCounter renders a live token count display.
type AITokenCounter struct {
	BaseComponent
	mu sync.Mutex

	promptTokens     int
	completionTokens int
	ratePerMillion   int // cost in cents per 1M tokens
	style            AITokenCounterStyle
	// cached
	promptStr     string
	completionStr string
	totalStr      string
	costStr       string
	detailStr     string
}

// NewAITokenCounter creates an AITokenCounter.
func NewAITokenCounter() *AITokenCounter {
	tc := &AITokenCounter{ratePerMillion: 200, style: DefaultAITokenCounterStyle()} // $2/M
	tc.SetID(GenerateID("tokcount"))
	tc.recomputeLocked()
	return tc
}

// SetCounts sets prompt and completion token counts.
func (tc *AITokenCounter) SetCounts(prompt, completion int) *AITokenCounter {
	tc.mu.Lock()
	if prompt < 0 {
		prompt = 0
	}
	if completion < 0 {
		completion = 0
	}
	tc.promptTokens = prompt
	tc.completionTokens = completion
	tc.recomputeLocked()
	tc.mu.Unlock()
	return tc
}

// SetRate sets cost rate in cents per 1 million tokens.
func (tc *AITokenCounter) SetRate(centsPerMillion int) *AITokenCounter {
	tc.mu.Lock()
	if centsPerMillion < 0 {
		centsPerMillion = 0
	}
	tc.ratePerMillion = centsPerMillion
	tc.recomputeLocked()
	tc.mu.Unlock()
	return tc
}

func (tc *AITokenCounter) recomputeLocked() {
	total := tc.promptTokens + tc.completionTokens
	tc.promptStr = itoa(tc.promptTokens)
	tc.completionStr = itoa(tc.completionTokens)
	tc.totalStr = itoa(total)

	// Cost in cents = total * ratePerMillion / 1M
	costCents := total * tc.ratePerMillion / 1000000
	if costCents == 0 && total > 0 {
		costCents = 1 // minimum 1 cent if any tokens
	}
	dollars := costCents / 100
	remainder := costCents % 100
	tc.costStr = "$" + itoa(dollars) + "." + formatCents(remainder)

	tc.detailStr = tc.promptStr + "+" + tc.completionStr + "=" + tc.totalStr
}

// TotalTokens returns total token count.
func (tc *AITokenCounter) TotalTokens() int {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.promptTokens + tc.completionTokens
}

// SetStyle sets custom style.
func (tc *AITokenCounter) SetStyle(s AITokenCounterStyle) *AITokenCounter {
	tc.mu.Lock()
	tc.style = s
	tc.mu.Unlock()
	return tc
}

// Measure returns preferred size.
func (tc *AITokenCounter) Measure(cs Constraints) Size {
	w := 32
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the token counter.
func (tc *AITokenCounter) Paint(buf *buffer.Buffer) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	b := tc.Bounds()
	x, y := b.X, b.Y

	costStyle := tc.style.Cost
	sepStyle := tc.style.Separator
	labelStyle := tc.style.Label

	col := x

	// Detail: prompt+completion=total
	for _, r := range tc.detailStr {
		if col >= buf.Width {
			break
		}
		var st buffer.Style
		// Color the total portion bold
		st = labelStyle
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		col++
	}

	// Separator
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
		col++
	}

	// Cost label
	for _, r := range "Cost:" {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Cost value
	for _, r := range tc.costStr {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: costStyle.Fg, Bg: costStyle.Bg, Flags: costStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (tc *AITokenCounter) Children() []Component { return nil }

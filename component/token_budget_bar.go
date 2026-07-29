package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── TokenBudgetBar: Multi-zone Token Budget Allocation Bar ───
//
// TokenBudgetBar renders a horizontal bar showing how token budget is
// allocated across different categories (system prompt, conversation,
// tools, output reserve). Each zone has a distinct color.
//
// Usage:
//
//	tb := NewTokenBudgetBar()
//	tb.SetZones(
//	    TokenZone{Name: "sys", Tokens: 2000, Color: buffer.RGB(168, 85, 247)},
//	    TokenZone{Name: "conv", Tokens: 6000, Color: buffer.RGB(59, 130, 246)},
//	    TokenZone{Name: "out", Tokens: 2000, Color: buffer.RGB(34, 197, 94)},
//	)
//	tb.Paint(buf)

// TokenZone represents a named budget zone.
type TokenZone struct {
	Name   string
	Tokens int
	Color  buffer.Color
}

// TokenBudgetBarStyle holds styling.
type TokenBudgetBarStyle struct {
	Empty buffer.Style
	Label buffer.Style
}

// DefaultTokenBudgetBarStyle returns defaults.
func DefaultTokenBudgetBarStyle() TokenBudgetBarStyle {
	return TokenBudgetBarStyle{
		Empty: buffer.Style{Fg: buffer.RGB(30, 41, 59)},
		Label: buffer.Style{Fg: buffer.RGB(148, 163, 184)},
	}
}

const tokenBudgetMaxZones = 8

// TokenBudgetBar renders a multi-zone token allocation bar.
type TokenBudgetBar struct {
	BaseComponent
	mu sync.Mutex

	zones [tokenBudgetMaxZones]TokenZone
	count int
	total int
	width int
	style TokenBudgetBarStyle
	// cached
	barWidths [tokenBudgetMaxZones]int
	totalStr  string
}

// NewTokenBudgetBar creates a TokenBudgetBar.
func NewTokenBudgetBar() *TokenBudgetBar {
	tb := &TokenBudgetBar{width: 30, style: DefaultTokenBudgetBarStyle()}
	tb.SetID(GenerateID("tokbudget"))
	tb.recomputeLocked()
	return tb
}

// SetZones sets the budget allocation zones.
func (tb *TokenBudgetBar) SetZones(zones ...TokenZone) *TokenBudgetBar {
	tb.mu.Lock()
	tb.count = 0
	tb.total = 0
	for _, z := range zones {
		if tb.count >= tokenBudgetMaxZones {
			break
		}
		if z.Tokens < 0 {
			z.Tokens = 0
		}
		tb.zones[tb.count] = z
		tb.total += z.Tokens
		tb.count++
	}
	tb.recomputeLocked()
	tb.mu.Unlock()
	return tb
}

func (tb *TokenBudgetBar) recomputeLocked() {
	tb.totalStr = itoa(tb.total)
	const barW = 20
	if tb.total == 0 {
		for i := range tb.barWidths {
			tb.barWidths[i] = 0
		}
		return
	}
	for i := 0; i < tb.count; i++ {
		tb.barWidths[i] = tb.zones[i].Tokens * barW / tb.total
	}
}

// TotalTokens returns the total allocated tokens.
func (tb *TokenBudgetBar) TotalTokens() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.total
}

// ZoneCount returns the number of zones.
func (tb *TokenBudgetBar) ZoneCount() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.count
}

// SetWidth sets the bar width.
func (tb *TokenBudgetBar) SetWidth(w int) *TokenBudgetBar {
	tb.mu.Lock()
	if w < 10 {
		w = 10
	}
	tb.width = w
	tb.mu.Unlock()
	return tb
}

// SetStyle sets custom style.
func (tb *TokenBudgetBar) SetStyle(s TokenBudgetBarStyle) *TokenBudgetBar {
	tb.mu.Lock()
	tb.style = s
	tb.mu.Unlock()
	return tb
}

// Measure returns preferred size.
func (tb *TokenBudgetBar) Measure(cs Constraints) Size {
	w := tb.width + 10
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 2}
}

// Paint renders the token budget bar.
func (tb *TokenBudgetBar) Paint(buf *buffer.Buffer) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	b := tb.Bounds()
	x, y := b.X, b.Y

	emptyStyle := tb.style.Empty
	labelStyle := tb.style.Label

	// Row 0: zone bars
	col := x
	for i := 0; i < tb.count; i++ {
		zoneStyle := buffer.Style{Fg: tb.zones[i].Color}
		for j := 0; j < tb.barWidths[i]; j++ {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, y, buffer.Cell{Rune: '█', Fg: zoneStyle.Fg, Bg: zoneStyle.Bg, Flags: zoneStyle.Flags, Width: 1})
			col++
		}
	}
	// Fill remaining with empty
	for col < buf.Width && col < x+20 {
		buf.SetCell(col, y, buffer.Cell{Rune: '░', Fg: emptyStyle.Fg, Bg: emptyStyle.Bg, Flags: emptyStyle.Flags, Width: 1})
		col++
	}

	// Row 1: zone labels
	col = x
	for i := 0; i < tb.count; i++ {
		zoneStyle := buffer.Style{Fg: tb.zones[i].Color}
		label := tb.zones[i].Name + ":" + itoa(tb.zones[i].Tokens)
		for _, r := range label {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: zoneStyle.Fg, Bg: zoneStyle.Bg, Flags: zoneStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, y+1, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (tb *TokenBudgetBar) Children() []Component { return nil }

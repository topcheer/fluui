package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── CompactStatCard: Mini KPI Stat Card ───
//
// CompactStatCard renders a small 2-row KPI card with a label, value,
// and optional trend indicator (up/down/flat). Designed for dense
// dashboard layouts where multiple stats need to fit in limited space.
//
// Usage:
//
//	sc := NewCompactStatCard()
//	sc.SetLabel("QPS")
//	sc.SetValue("1,234")
//	sc.SetTrend(TrendUp, 12) // +12%
//	sc.Paint(buf)

// TrendDirection represents the trend direction.
type TrendDirection int

const (
	TrendFlat TrendDirection = 0
	TrendUp   TrendDirection = 1
	TrendDown TrendDirection = 2
)

var trendIcons = [...]rune{'→', '↑', '↓'}

// CompactStatCardStyle holds styling.
type CompactStatCardStyle struct {
	Label  buffer.Style
	Value  buffer.Style
	Up     buffer.Style
	Down   buffer.Style
	Flat   buffer.Style
	Border buffer.Style
}

// DefaultCompactStatCardStyle returns defaults.
func DefaultCompactStatCardStyle() CompactStatCardStyle {
	return CompactStatCardStyle{
		Label:  buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Value:  buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
		Up:     buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Down:   buffer.Style{Fg: buffer.RGB(239, 68, 68)},
		Flat:   buffer.Style{Fg: buffer.RGB(100, 116, 139)},
		Border: buffer.Style{Fg: buffer.RGB(51, 65, 85)},
	}
}

// CompactStatCard renders a mini KPI stat card.
type CompactStatCard struct {
	BaseComponent
	mu sync.Mutex

	label    string
	value    string
	trend    TrendDirection
	trendPct int
	style    CompactStatCardStyle
	// cached
	trendStr   string
	trendStyle buffer.Style
}

// NewCompactStatCard creates a CompactStatCard.
func NewCompactStatCard() *CompactStatCard {
	sc := &CompactStatCard{style: DefaultCompactStatCardStyle()}
	sc.SetID(GenerateID("statcard2"))
	sc.recomputeLocked()
	return sc
}

// SetLabel sets the stat label.
func (sc *CompactStatCard) SetLabel(l string) *CompactStatCard {
	sc.mu.Lock()
	sc.label = l
	sc.mu.Unlock()
	return sc
}

// SetValue sets the stat value string.
func (sc *CompactStatCard) SetValue(v string) *CompactStatCard {
	sc.mu.Lock()
	sc.value = v
	sc.mu.Unlock()
	return sc
}

// SetTrend sets the trend direction and percentage.
func (sc *CompactStatCard) SetTrend(dir TrendDirection, pct int) *CompactStatCard {
	sc.mu.Lock()
	sc.trend = dir
	sc.trendPct = pct
	sc.recomputeLocked()
	sc.mu.Unlock()
	return sc
}

func (sc *CompactStatCard) recomputeLocked() {
	pctStr := itoa(sc.trendPct) + "%"
	sc.trendStr = pctStr

	switch sc.trend {
	case TrendUp:
		sc.trendStyle = sc.style.Up
	case TrendDown:
		sc.trendStyle = sc.style.Down
	default:
		sc.trendStyle = sc.style.Flat
	}
}

// Label returns the current label.
func (sc *CompactStatCard) Label() string {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.label
}

// SetStyle sets custom style.
func (sc *CompactStatCard) SetStyle(s CompactStatCardStyle) *CompactStatCard {
	sc.mu.Lock()
	sc.style = s
	sc.recomputeLocked()
	sc.mu.Unlock()
	return sc
}

// Measure returns preferred size.
func (sc *CompactStatCard) Measure(cs Constraints) Size {
	w := 14
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 3}
}

// Paint renders the compact stat card.
func (sc *CompactStatCard) Paint(buf *buffer.Buffer) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	b := sc.Bounds()
	x, y := b.X, b.Y
	w := b.W
	if w < 8 {
		w = 14
	}

	labelStyle := sc.style.Label
	valueStyle := sc.style.Value
	borderStyle := sc.style.Border
	trendStyle := sc.trendStyle

	// Top border
	col := x
	for i := 0; i < w; i++ {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: '─', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
		col++
	}

	// Row 1: label
	col = x + 1
	for _, r := range sc.label {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Row 2: value + trend
	col = x + 1
	for _, r := range sc.value {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y+2, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}
	// Trend indicator
	if col < buf.Width {
		buf.SetCell(col, y+2, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y+2, buffer.Cell{Rune: trendIcons[sc.trend], Fg: trendStyle.Fg, Bg: trendStyle.Bg, Flags: trendStyle.Flags, Width: 1})
		col++
	}
	for _, r := range sc.trendStr {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y+2, buffer.Cell{Rune: r, Fg: trendStyle.Fg, Bg: trendStyle.Bg, Flags: trendStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (sc *CompactStatCard) Children() []Component { return nil }

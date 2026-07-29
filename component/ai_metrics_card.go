package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AIMetricsCard: AI Performance Metrics Display Card ───
//
// AIMetricsCard renders a compact card showing key AI performance metrics:
// tokens/sec, latency, cost per request, and success rate.
//
// Usage:
//
//	mc := NewAIMetricsCard()
//	mc.SetTokensPerSec(85.5)
//	mc.SetLatency(450) // ms
//	mc.SetCostPerReq(0.0023)
//	mc.SetSuccessRate(99.2)
//	mc.Paint(buf)

// AIMetricsStyle holds styling.
type AIMetricsStyle struct {
	Label  buffer.Style
	Value  buffer.Style
	Unit   buffer.Style
	Border buffer.Style
}

// DefaultAIMetricsStyle returns defaults.
func DefaultAIMetricsStyle() AIMetricsStyle {
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	value := buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold}
	unit := buffer.Style{Fg: buffer.RGB(100, 116, 139)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return AIMetricsStyle{Label: label, Value: value, Unit: unit, Border: border}
}

// AIMetricsCard renders AI performance metrics in a compact card.
type AIMetricsCard struct {
	BaseComponent
	mu sync.Mutex

	tps     float64
	latency int // ms
	cost    float64 // dollars
	success float64 // percentage
	style   AIMetricsStyle
	// cached
	tpsStr  string
	latStr  string
	costStr string
	sucStr  string
}

// NewAIMetricsCard creates an AIMetricsCard.
func NewAIMetricsCard() *AIMetricsCard {
	mc := &AIMetricsCard{style: DefaultAIMetricsStyle()}
	mc.SetID(GenerateID("aimetrics"))
	mc.cacheStringsLocked()
	return mc
}

func (mc *AIMetricsCard) cacheStringsLocked() {
	mc.tpsStr = formatFloatMC(mc.tps) + " tok/s"
	mc.latStr = itoa(mc.latency) + "ms"
	mc.costStr = "$" + formatFloatMC(mc.cost)
	mc.sucStr = formatFloatMC(mc.success) + "%"
}

// formatFloat formats a float to 1 decimal place without strconv import.
func formatFloatMC(f float64) string {
	whole := int(f)
	frac := int((f - float64(whole)) * 10)
	if frac < 0 { frac = -frac }
	if frac > 9 { frac = 9 }
	return itoa(whole) + "." + itoa(frac)
}

// SetTokensPerSec sets tokens per second.
func (mc *AIMetricsCard) SetTokensPerSec(v float64) *AIMetricsCard {
	mc.mu.Lock()
	mc.tps = v
	mc.cacheStringsLocked()
	mc.mu.Unlock()
	return mc
}

// SetLatency sets latency in milliseconds.
func (mc *AIMetricsCard) SetLatency(ms int) *AIMetricsCard {
	mc.mu.Lock()
	mc.latency = ms
	mc.cacheStringsLocked()
	mc.mu.Unlock()
	return mc
}

// SetCostPerReq sets cost per request in dollars.
func (mc *AIMetricsCard) SetCostPerReq(c float64) *AIMetricsCard {
	mc.mu.Lock()
	mc.cost = c
	mc.cacheStringsLocked()
	mc.mu.Unlock()
	return mc
}

// SetSuccessRate sets success rate percentage.
func (mc *AIMetricsCard) SetSuccessRate(pct float64) *AIMetricsCard {
	mc.mu.Lock()
	mc.success = pct
	mc.cacheStringsLocked()
	mc.mu.Unlock()
	return mc
}

// SetStyle sets custom style.
func (mc *AIMetricsCard) SetStyle(s AIMetricsStyle) *AIMetricsCard {
	mc.mu.Lock()
	mc.style = s
	mc.mu.Unlock()
	return mc
}

// Measure returns the preferred size.
func (mc *AIMetricsCard) Measure(cs Constraints) Size {
	w := 28
	h := 6 // border + 4 metric rows + border
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the metrics card into the buffer.
func (mc *AIMetricsCard) Paint(buf *buffer.Buffer) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	b := mc.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 15 { w = 28 }
	if h < 6 { h = 6 }

	bs := mc.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	labelStyle := mc.style.Label
	valueStyle := mc.style.Value

	// 4 metric rows
	type metric struct {
		label string
		value string
	}
	metrics := [4]metric{
		{"Tokens/s", mc.tpsStr},
		{"Latency", mc.latStr},
		{"Cost/req", mc.costStr},
		{"Success", mc.sucStr},
	}

	for idx, m := range metrics {
		rowY := y + 1 + idx
		if rowY >= y+h-1 || rowY >= buf.Height { break }

		col := x + 1
		// Label
		for _, r := range m.label {
			if col >= x+w-1 || col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}
		// Padding
		for col < x+w-2-len([]rune(m.value)) && col < buf.Width {
			buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}
		// Value (right-aligned)
		for _, r := range m.value {
			if col >= x+w-1 || col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (mc *AIMetricsCard) Children() []Component { return nil }

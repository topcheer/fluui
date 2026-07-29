package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── TokenCostChart: Token Cost Over Time Mini Chart ───
//
// TokenCostChart renders a compact bar chart showing cumulative token cost
// (in cents or custom units) over recent API calls. Helps visualize spending
// trends and identify cost spikes.
//
// Usage:
//
//	c := NewTokenCostChart()
//	c.AddCost(5)   // 5 cents
//	c.AddCost(12)  // 12 cents
//	c.Paint(buf)

// TokenCostChartStyle holds styling.
type TokenCostChartStyle struct {
	Bar     buffer.Style
	Peak    buffer.Style
	Label   buffer.Style
	Total   buffer.Style
}

// DefaultTokenCostChartStyle returns defaults.
func DefaultTokenCostChartStyle() TokenCostChartStyle {
	return TokenCostChartStyle{
		Bar:   buffer.Style{Fg: buffer.RGB(234, 179, 8)},
		Peak:  buffer.Style{Fg: buffer.RGB(239, 68, 68)},
		Label: buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Total: buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
	}
}

const costChartMaxPoints = 24

// TokenCostChart renders a token cost mini chart.
type TokenCostChart struct {
	BaseComponent
	mu sync.Mutex

	points [costChartMaxPoints]int // cost per call, circular buffer
	count  int
	head   int
	width  int
	height int
	style  TokenCostChartStyle
	// cached
	heights  [costChartMaxPoints]int
	totalStr  string
	peakStr   string
	avgStr    string
	labelStr  string
}

// NewTokenCostChart creates a TokenCostChart.
func NewTokenCostChart() *TokenCostChart {
	c := &TokenCostChart{width: 30, height: 5, style: DefaultTokenCostChartStyle()}
	c.SetID(GenerateID("costchart"))
	c.recomputeLocked()
	return c
}

// AddCost adds a cost sample (in cents or custom units).
func (c *TokenCostChart) AddCost(cost int) *TokenCostChart {
	c.mu.Lock()
	if cost < 0 { cost = 0 }
	c.points[c.head] = cost
	c.head = (c.head + 1) % costChartMaxPoints
	if c.count < costChartMaxPoints { c.count++ }
	c.recomputeLocked()
	c.mu.Unlock()
	return c
}

// Clear resets all cost data.
func (c *TokenCostChart) Clear() *TokenCostChart {
	c.mu.Lock()
	c.count = 0
	c.head = 0
	c.recomputeLocked()
	c.mu.Unlock()
	return c
}

func (c *TokenCostChart) recomputeLocked() {
	if c.count == 0 {
		c.totalStr = "$0.00"
		c.peakStr = "0c"
		c.avgStr = "0c"
		c.labelStr = "Total:$0.00 avg:0c peak:0c"
		return
	}

	total := 0
	peak := 0
	for i := 0; i < c.count; i++ {
		v := c.points[i]
		total += v
		if v > peak { peak = v }
	}
	avg := total / c.count

	// Format total as dollars
	dollars := total / 100
	cents := total % 100
	c.totalStr = "$" + itoa(dollars) + "." + formatCents(cents)
	c.peakStr = itoa(peak) + "c"
	c.avgStr = itoa(avg) + "c"
	c.labelStr = "Total:" + c.totalStr + " avg:" + c.avgStr + " peak:" + c.peakStr

	// Compute bar heights
	if peak == 0 { peak = 1 }
	maxH := c.height - 1
	if maxH < 1 { maxH = 1 }
	for i := 0; i < c.count; i++ {
		h := c.points[i] * maxH / peak
		if h < 1 { h = 1 }
		c.heights[i] = h
	}
}

// formatCents formats a 0-99 value as two-digit string.
func formatCents(c int) string {
	if c < 10 {
		return "0" + itoa(c)
	}
	return itoa(c)
}

// TotalCost returns total cost in cents.
func (c *TokenCostChart) TotalCost() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	total := 0
	for i := 0; i < c.count; i++ { total += c.points[i] }
	return total
}

// SetSize sets chart dimensions.
func (c *TokenCostChart) SetSize(w, h int) *TokenCostChart {
	c.mu.Lock()
	if w < 10 { w = 10 }
	if h < 3 { h = 3 }
	c.width = w
	c.height = h
	c.recomputeLocked()
	c.mu.Unlock()
	return c
}

// SetStyle sets custom style.
func (c *TokenCostChart) SetStyle(s TokenCostChartStyle) *TokenCostChart {
	c.mu.Lock()
	c.style = s
	c.mu.Unlock()
	return c
}

// Measure returns preferred size.
func (c *TokenCostChart) Measure(cs Constraints) Size {
	w := c.width
	h := c.height + 1
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the cost chart.
func (c *TokenCostChart) Paint(buf *buffer.Buffer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	b := c.Bounds()
	x, y := b.X, b.Y
	w := b.W
	h := b.H
	if w < 10 { w = 30 }
	if h < 3 { h = 5 }

	barStyle := c.style.Bar
	peakStyle := c.style.Peak
	labelStyle := c.style.Label
	totalStyle := c.style.Total

	graphH := h - 1
	if graphH < 1 { graphH = 1 }

	if c.count == 0 {
		for i, r := range "No cost data" {
			if x+i >= buf.Width { break }
			buf.SetCell(x+i, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		}
		return
	}

	// Determine peak value for highlighting
	peakVal := 0
	for i := 0; i < c.count; i++ {
		if c.points[i] > peakVal { peakVal = c.points[i] }
	}

	nDisplay := c.count
	if nDisplay > w { nDisplay = w }
	startIdx := c.count - nDisplay

	for col := 0; col < nDisplay; col++ {
		barH := c.heights[startIdx+col]
		cx := x + col
		if cx >= buf.Width { break }
		isPeak := c.points[startIdx+col] == peakVal && peakVal > 0
		var st buffer.Style
		if isPeak {
			st = peakStyle
		} else {
			st = barStyle
		}
		for row := 0; row < barH; row++ {
			yy := y + graphH - 1 - row
			if yy < 0 || yy >= buf.Height { continue }
			var rune_ rune
			if row == barH-1 {
				rune_ = '▄'
			} else {
				rune_ = '█'
			}
			buf.SetCell(cx, yy, buffer.Cell{Rune: rune_, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		}
	}

	// Label row
	labelY := y + graphH
	if labelY < buf.Height {
		label := c.labelStr
		for i, r := range label {
			cx := x + i
			if cx >= buf.Width { break }
			var st buffer.Style
			if i < len("Total:")+len(c.totalStr) && i >= len("Total:") {
				st = totalStyle
			} else {
				st = labelStyle
			}
			buf.SetCell(cx, labelY, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		}
	}
}

// Children returns nil.
func (c *TokenCostChart) Children() []Component { return nil }

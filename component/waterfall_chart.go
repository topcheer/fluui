package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── WaterfallChart: Financial Waterfall / Bridge Chart ───
//
// WaterfallChart renders a sequence of positive (green) and negative (red)
// changes that build from a starting value to an ending total. Common in
// financial analysis, budget reconciliation, and revenue bridges.
//
// Usage:
//
//	wc := NewWaterfallChart()
//	wc.AddBar("Start", 100, WaterfallStart)
//	wc.AddBar("+Sales", 40, WaterfallPositive)
//	wc.AddBar("-Costs", -25, WaterfallNegative)
//	wc.AddBar("End", 0, WaterfallEnd) // auto-computed = 115
//	wc.SetBounds(Rect{X:0, Y:0, W:60, H:15})
//	wc.Paint(buf)

// WaterfallBarType classifies each bar.
type WaterfallBarType int

const (
	WaterfallStart     WaterfallBarType = iota // initial baseline
	WaterfallPositive                          // increase (green, from bottom up)
	WaterfallNegative                          // decrease (red, from top down)
	WaterfallEnd                               // final total (auto-computed)
)

// WaterfallBar represents a single bar in the chart.
type WaterfallBar struct {
	Label string
	Value float64
	Type  WaterfallBarType
}

// WaterfallChartStyle holds visual styles.
type WaterfallChartStyle struct {
	Positive buffer.Style
	Negative buffer.Style
	Total    buffer.Style
	Connector buffer.Style
	Label    buffer.Style
	Axis     buffer.Style
}

// DefaultWaterfallChartStyle returns sensible defaults.
func DefaultWaterfallChartStyle() WaterfallChartStyle {
	return WaterfallChartStyle{
		Positive:  buffer.Style{Fg: buffer.RGB(16, 163, 127)},
		Negative:  buffer.Style{Fg: buffer.RGB(220, 80, 80)},
		Total:     buffer.Style{Fg: buffer.RGB(100, 149, 237)},
		Connector: buffer.Style{Fg: buffer.RGB(80, 80, 80)},
		Label:     buffer.Style{Fg: buffer.White},
		Axis:      buffer.Style{Fg: buffer.RGB(100, 100, 100)},
	}
}

// WaterfallChart renders a waterfall/bridge chart.
type WaterfallChart struct {
	BaseComponent
	mu    sync.RWMutex
	bars  []WaterfallBar
	style WaterfallChartStyle
}

// NewWaterfallChart creates an empty waterfall chart.
func NewWaterfallChart() *WaterfallChart {
	wc := &WaterfallChart{
		style: DefaultWaterfallChartStyle(),
	}
	wc.SetID(GenerateID("waterfall"))
	return wc
}

// AddBar adds a bar to the chart.
func (wc *WaterfallChart) AddBar(bar WaterfallBar) *WaterfallChart {
	wc.mu.Lock()
	wc.bars = append(wc.bars, bar)
	wc.mu.Unlock()
	return wc
}

// SetBars replaces all bars.
func (wc *WaterfallChart) SetBars(bars []WaterfallBar) *WaterfallChart {
	wc.mu.Lock()
	wc.bars = bars
	wc.mu.Unlock()
	return wc
}

// Bars returns the current bars.
func (wc *WaterfallChart) Bars() []WaterfallBar {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	return wc.bars
}

// BarCount returns the number of bars.
func (wc *WaterfallChart) BarCount() int {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	return len(wc.bars)
}

// Clear removes all bars.
func (wc *WaterfallChart) Clear() *WaterfallChart {
	wc.mu.Lock()
	wc.bars = wc.bars[:0]
	wc.mu.Unlock()
	return wc
}

// SetStyle sets the visual style.
func (wc *WaterfallChart) SetStyle(s WaterfallChartStyle) *WaterfallChart {
	wc.mu.Lock()
	wc.style = s
	wc.mu.Unlock()
	return wc
}

// Style returns the current style.
func (wc *WaterfallChart) Style() WaterfallChartStyle {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	return wc.style
}

// computeRunningTotals returns per-bar start/end values and maxAbs.
func (wc *WaterfallChart) computeRunningTotals() ([]float64, []float64, float64) {
	wc.mu.Lock()
	defer wc.mu.Unlock()
	return wc.computeRunningTotalsLocked()
}

// Measure computes the desired size.
func (wc *WaterfallChart) Measure(cs Constraints) Size {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	w := len(wc.bars)*4 + 4
	if w < 20 {
		w = 20
	}
	h := 15
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the waterfall chart.
func (wc *WaterfallChart) Paint(buf *buffer.Buffer) {
	wc.mu.Lock()
	defer wc.mu.Unlock()

	b := wc.bounds
	if b.W < 6 || b.H < 4 || len(wc.bars) == 0 {
		return
	}

	n := len(wc.bars)
	chartH := b.H - 2 // reserve 2 rows: 1 for labels, 1 for axis
	if chartH < 2 {
		chartH = 2
	}

	// Use locked compute
	starts, running, maxAbs := wc.computeRunningTotalsLocked()

	barW := b.W / n
	if barW < 2 {
		barW = 2
	}
	gap := 1
	if barW > 4 {
		gap = barW / 4
	}
	drawW := barW - gap

	for i, bar := range wc.bars {
		startVal := starts[i]
		endVal := running[i]
		lo := startVal
		hi := endVal
		if lo > hi {
			lo, hi = hi, lo
		}

		// Convert values to y positions (0 at bottom = maxAbs, chartH-1 at top)
		loY := b.Y + chartH - int((lo+maxAbs)/(2*maxAbs)*float64(chartH))
		hiY := b.Y + chartH - int((hi+maxAbs)/(2*maxAbs)*float64(chartH))
		if hiY < b.Y {
			hiY = b.Y
		}
		if loY >= b.Y+chartH {
			loY = b.Y + chartH - 1
		}

		// Determine style
		var style buffer.Style
		switch bar.Type {
		case WaterfallPositive:
			style = wc.style.Positive
		case WaterfallNegative:
			style = wc.style.Negative
		default:
			style = wc.style.Total
		}

		// Draw bar
		barX := b.X + i*barW
		for y := hiY; y <= loY; y++ {
			for x := 0; x < drawW; x++ {
				ax := barX + x
				if ax >= b.X+b.W {
					break
				}
				buf.SetCell(ax, y, buffer.Cell{Rune: '█', Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
			}
		}

		// Draw connector line to next bar
		if i < n-1 {
			connY := b.Y + chartH - int((endVal+maxAbs)/(2*maxAbs)*float64(chartH))
			if connY < b.Y {
				connY = b.Y
			}
			if connY >= b.Y+chartH {
				connY = b.Y + chartH - 1
			}
			for x := barX + drawW; x < barX + barW; x++ {
				if x >= b.X+b.W {
					break
				}
				buf.SetCell(x, connY, buffer.Cell{Rune: '─', Fg: wc.style.Connector.Fg, Bg: wc.style.Connector.Bg, Width: 1})
			}
		}

		// Draw label
		labelY := b.Y + b.H - 1
		labelRunes := []rune(bar.Label)
		for x, r := range labelRunes {
			ax := barX + x
			if ax >= b.X+b.W || x >= drawW+gap {
				break
			}
			buf.SetCell(ax, labelY, buffer.Cell{Rune: r, Fg: wc.style.Label.Fg, Bg: wc.style.Label.Bg, Flags: wc.style.Label.Flags, Width: 1})
		}
	}

	// Draw axis line
	axisY := b.Y + chartH
	for x := 0; x < b.W; x++ {
		buf.SetCell(b.X+x, axisY, buffer.Cell{Rune: '─', Fg: wc.style.Axis.Fg, Bg: wc.style.Axis.Bg, Width: 1})
	}
}

// computeRunningTotalsLocked is the lock-holding version (caller must hold wc.mu).
func (wc *WaterfallChart) computeRunningTotalsLocked() ([]float64, []float64, float64) {
	n := len(wc.bars)
	running := make([]float64, n)
	starts := make([]float64, n)
	cumulative := 0.0
	maxAbs := 0.0

	for i, bar := range wc.bars {
		switch bar.Type {
		case WaterfallStart:
			cumulative = bar.Value
			starts[i] = 0
			running[i] = bar.Value
		case WaterfallPositive:
			starts[i] = cumulative
			cumulative += bar.Value
			running[i] = cumulative
		case WaterfallNegative:
			cumulative += bar.Value // value is negative, decreases total
			starts[i] = cumulative
			running[i] = cumulative - bar.Value // old cumulative
		case WaterfallEnd:
			starts[i] = 0
			running[i] = cumulative
		}

		for _, v := range []float64{running[i], starts[i]} {
			abs := v
			if abs < 0 {
				abs = -abs
			}
			if abs > maxAbs {
				maxAbs = abs
			}
		}
	}

	if maxAbs == 0 {
		maxAbs = 1
	}
	return starts, running, maxAbs
}

// Children returns nil.
func (wc *WaterfallChart) Children() []Component { return nil }

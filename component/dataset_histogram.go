package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── DatasetDistributionHistogram: Dataset Value Distribution ───
//
// DatasetDistributionHistogram renders a histogram of value distributions
// with configurable bin count, auto-scaled bars, and axis labels.
// Useful for ML data exploration and statistical analysis.
//
// Usage:
//
//	h := NewDatasetDistributionHistogram()
//	h.SetLabel("Token Lengths")
//	h.AddBin("0-10", 45)
//	h.AddBin("10-20", 120)
//	h.AddBin("20-30", 80)
//	h.Paint(buf)

// HistogramBin represents a single histogram bar.
type HistogramBin struct {
	Label string
	Count int
	// cached
	CountStr string
	BarH     int
}

// HistogramStyle holds styling.
type HistogramStyle struct {
	Bar     buffer.Style
	Label   buffer.Style
	Value   buffer.Style
	Border  buffer.Style
}

// DefaultHistogramStyle returns defaults.
func DefaultHistogramStyle() HistogramStyle {
	bar := buffer.Style{Fg: buffer.RGB(96, 165, 250)}
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	value := buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return HistogramStyle{Bar: bar, Label: label, Value: value, Border: border}
}

// DatasetDistributionHistogram renders a value distribution histogram.
type DatasetDistributionHistogram struct {
	BaseComponent
	mu sync.Mutex

	label   string
	bins    []HistogramBin
	maxH    int
	style   HistogramStyle
}

// NewDatasetDistributionHistogram creates a DatasetDistributionHistogram.
func NewDatasetDistributionHistogram() *DatasetDistributionHistogram {
	h := &DatasetDistributionHistogram{maxH: 8, style: DefaultHistogramStyle()}
	h.SetID(GenerateID("histogram"))
	return h
}

// SetLabel sets the chart title.
func (h *DatasetDistributionHistogram) SetLabel(l string) *DatasetDistributionHistogram {
	h.mu.Lock()
	h.label = l
	h.mu.Unlock()
	return h
}

// Label returns the title.
func (h *DatasetDistributionHistogram) Label() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.label
}

// AddBin adds a histogram bin (caches count string and bar height).
func (h *DatasetDistributionHistogram) AddBin(label string, count int) *DatasetDistributionHistogram {
	h.mu.Lock()
	bin := HistogramBin{Label: label, Count: count, CountStr: itoa(count)}
	h.bins = append(h.bins, bin)
	h.recomputeBarsLocked()
	h.mu.Unlock()
	return h
}

// recomputeBarsLocked recalculates bar heights from counts.
func (h *DatasetDistributionHistogram) recomputeBarsLocked() {
	maxCount := 1
	for _, b := range h.bins {
		if b.Count > maxCount { maxCount = b.Count }
	}
	for i := range h.bins {
		ratio := float64(h.bins[i].Count) / float64(maxCount)
		h.bins[i].BarH = int(ratio * float64(h.maxH))
		if h.bins[i].BarH < 1 && h.bins[i].Count > 0 { h.bins[i].BarH = 1 }
	}
}

// BinCount returns the number of bins.
func (h *DatasetDistributionHistogram) BinCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.bins)
}

// SetMaxBarHeight sets the maximum bar height in rows.
func (h *DatasetDistributionHistogram) SetMaxBarHeight(n int) *DatasetDistributionHistogram {
	h.mu.Lock()
	if n < 1 { n = 1 }
	h.maxH = n
	h.recomputeBarsLocked()
	h.mu.Unlock()
	return h
}

// Clear removes all bins.
func (h *DatasetDistributionHistogram) Clear() *DatasetDistributionHistogram {
	h.mu.Lock()
	h.bins = h.bins[:0]
	h.mu.Unlock()
	return h
}

// SetStyleS sets custom style.
func (h *DatasetDistributionHistogram) SetStyleS(s HistogramStyle) *DatasetDistributionHistogram {
	h.mu.Lock()
	h.style = s
	h.mu.Unlock()
	return h
}

// Measure returns preferred size.
func (h *DatasetDistributionHistogram) Measure(cs Constraints) Size {
	h.mu.Lock()
	binCount := len(h.bins)
	mh := h.maxH
	h.mu.Unlock()
	w := binCount*5 + 4
	if w < 20 { w = 20 }
	hh := mh + 4
	if hh < 5 { hh = 5 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && hh > cs.MaxHeight { hh = cs.MaxHeight }
	return Size{W: w, H: hh}
}

// Paint renders the histogram into the buffer.
func (h *DatasetDistributionHistogram) Paint(buf *buffer.Buffer) {
	h.mu.Lock()
	defer h.mu.Unlock()

	b := h.Bounds()
	x, y := b.X, b.Y
	w, hh := b.W, b.H
	if w < 20 { w = 30 }
	if hh < 5 { hh = 8 }

	bs := h.style.Border
	for row := 0; row < hh && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == hh-1 && col == 0 { ch = '└' } else if row == hh-1 && col == w-1 { ch = '┘' } else if row == 0 || row == hh-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	// Title
	labelStyle := h.style.Label
	titleCol := x + 1
	for _, r := range h.label {
		if titleCol >= x+w-1 || titleCol >= buf.Width { break }
		buf.SetCell(titleCol, y+1, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		titleCol++
	}

	barStyle := h.style.Bar
	valueStyle := h.style.Value

	// Draw bars from bottom up
	barAreaTop := y + 2
	barAreaBottom := y + hh - 2

	for bIdx, bin := range h.bins {
		barCol := x + 2 + bIdx*5
		if barCol >= x+w-1 || barCol >= buf.Width { break }

		// Draw bar from bottom up
		for i := 0; i < bin.BarH; i++ {
			barY := barAreaBottom - i
			if barY < barAreaTop || barY >= buf.Height { break }
			if barCol < buf.Width {
				buf.SetCell(barCol, barY, buffer.Cell{Rune: '█', Fg: barStyle.Fg, Bg: barStyle.Bg, Flags: barStyle.Flags, Width: 1})
			}
			if barCol+1 < buf.Width {
				buf.SetCell(barCol+1, barY, buffer.Cell{Rune: '█', Fg: barStyle.Fg, Bg: barStyle.Bg, Flags: barStyle.Flags, Width: 1})
			}
		}

		// Count value above bar
		valY := barAreaBottom - bin.BarH
		if valY >= barAreaTop && valY < buf.Height && bin.CountStr != "" {
			for i, r := range bin.CountStr {
				cx := barCol + i
				if cx >= x+w-1 || cx >= buf.Width { break }
				buf.SetCell(cx, valY, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
			}
		}

		// Bin label below bar
		labelY := barAreaBottom + 1
		if labelY < buf.Height {
			for i, r := range bin.Label {
				cx := barCol + i
				if cx >= x+w-1 || cx >= buf.Width { break }
				buf.SetCell(cx, labelY, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			}
		}
	}
}

// Children returns nil.
func (h *DatasetDistributionHistogram) Children() []Component { return nil }

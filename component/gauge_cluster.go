package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── GaugeCluster: Multi-Gauge Grid Display ───
//
// GaugeCluster renders multiple mini gauges in a compact grid layout,
// each with label, value bar, and threshold colors.
//
// Usage:
//
//	gc := NewGaugeCluster()
//	gc.SetColumns(2)
//	gc.AddGauge("CPU", 75, 100)
//	gc.AddGauge("Memory", 50, 100)
//	gc.AddGauge("Disk", 90, 100)
//	gc.Paint(buf)

// GaugeEntry represents a single gauge in the cluster.
type GaugeEntry struct {
	Label string
	Value float64
	Max   float64
	// cached fill count
	Filled int
	Width  int
}

// GaugeClusterStyle holds styling.
type GaugeClusterStyle struct {
	Normal   buffer.Style
	Warning  buffer.Style
	Critical buffer.Style
	Label    buffer.Style
	Value    buffer.Style
	Border   buffer.Style
}

// DefaultGaugeClusterStyle returns defaults.
func DefaultGaugeClusterStyle() GaugeClusterStyle {
	normal := buffer.Style{Fg: buffer.RGB(34, 197, 94)}
	warn := buffer.Style{Fg: buffer.RGB(234, 179, 8)}
	crit := buffer.Style{Fg: buffer.RGB(239, 68, 68)}
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	val := buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return GaugeClusterStyle{Normal: normal, Warning: warn, Critical: crit, Label: label, Value: val, Border: border}
}

// GaugeCluster renders multiple gauges in a grid.
type GaugeCluster struct {
	BaseComponent
	mu sync.Mutex

	entries []GaugeEntry
	columns int
	style   GaugeClusterStyle
	barWidth int
}

// NewGaugeCluster creates a GaugeCluster.
func NewGaugeCluster() *GaugeCluster {
	gc := &GaugeCluster{
		columns:  2,
		barWidth: 10,
		style:    DefaultGaugeClusterStyle(),
	}
	gc.SetID(GenerateID("gaugecluster"))
	return gc
}

// AddGauge adds a gauge to the cluster.
func (gc *GaugeCluster) AddGauge(label string, value, max float64) *GaugeCluster {
	gc.mu.Lock()
	entry := GaugeEntry{Label: label, Value: value, Max: max, Width: gc.barWidth}
	ratio := 0.0
	if max > 0 {
		ratio = value / max
		if ratio > 1 { ratio = 1 }
		if ratio < 0 { ratio = 0 }
	}
	entry.Filled = int(ratio * float64(gc.barWidth))
	gc.entries = append(gc.entries, entry)
	gc.mu.Unlock()
	return gc
}

// GaugeCount returns the number of gauges.
func (gc *GaugeCluster) GaugeCount() int {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	return len(gc.entries)
}

// SetColumns sets the grid column count.
func (gc *GaugeCluster) SetColumns(c int) *GaugeCluster {
	gc.mu.Lock()
	if c < 1 { c = 1 }
	gc.columns = c
	gc.mu.Unlock()
	return gc
}

// SetBarWidth sets the gauge bar width.
func (gc *GaugeCluster) SetBarWidth(w int) *GaugeCluster {
	gc.mu.Lock()
	gc.barWidth = w
	gc.mu.Unlock()
	return gc
}

// Clear removes all gauges.
func (gc *GaugeCluster) Clear() *GaugeCluster {
	gc.mu.Lock()
	gc.entries = gc.entries[:0]
	gc.mu.Unlock()
	return gc
}

// SetStyle sets custom style.
func (gc *GaugeCluster) SetStyle(s GaugeClusterStyle) *GaugeCluster {
	gc.mu.Lock()
	gc.style = s
	gc.mu.Unlock()
	return gc
}

// Measure returns the preferred size.
func (gc *GaugeCluster) Measure(cs Constraints) Size {
	gc.mu.Lock()
	count := len(gc.entries)
	cols := gc.columns
	bw := gc.barWidth
	gc.mu.Unlock()
	if cols < 1 { cols = 1 }
	rows := (count + cols - 1) / cols
	if rows < 1 { rows = 1 }
	w := cols * (bw + 15)
	if w < 20 { w = 20 }
	h := rows + 2
	if h < 3 { h = 3 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the gauge cluster into the buffer.
func (gc *GaugeCluster) Paint(buf *buffer.Buffer) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	b := gc.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 20 { w = 40 }
	if h < 3 { h = 5 }

	bs := gc.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	cols := gc.columns
	if cols < 1 { cols = 1 }
	colWidth := w / cols
	if colWidth < 5 { colWidth = 5 }
	labelStyle := gc.style.Label

	for idx, entry := range gc.entries {
		colIdx := idx % cols
		rowIdx := idx / cols
		cellX := x + 1 + colIdx*colWidth
		cellY := y + 1 + rowIdx
		if cellY >= y+h-1 || cellY >= buf.Height { break }

		// Determine threshold style
		ratio := 0.0
		if entry.Max > 0 {
			ratio = entry.Value / entry.Max
			if ratio > 1 { ratio = 1 }
		}
		var fillStyle buffer.Style
		switch {
		case ratio >= 0.85:
			fillStyle = gc.style.Critical
		case ratio >= 0.6:
			fillStyle = gc.style.Warning
		default:
			fillStyle = gc.style.Normal
		}

		// Label
		col := cellX
		for _, r := range entry.Label {
			if col >= cellX+colWidth-1 || col >= buf.Width { break }
			buf.SetCell(col, cellY, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}

		// Bar after label
		barStart := cellX + 8
		if barStart < col { barStart = col + 1 }
		bw := entry.Width
		if barStart+bw > cellX+colWidth-1 { bw = cellX + colWidth - 1 - barStart }
		if bw < 1 { bw = 1 }

		for i := 0; i < bw; i++ {
			cx := barStart + i
			if cx >= buf.Width { break }
			var ch rune
			var style buffer.Style
			if i < entry.Filled {
				ch = '▰'
				style = fillStyle
			} else {
				ch = '▱'
				style = labelStyle
			}
			buf.SetCell(cx, cellY, buffer.Cell{Rune: ch, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
		}
	}
}

// Children returns nil.
func (gc *GaugeCluster) Children() []Component { return nil }

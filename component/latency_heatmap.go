package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── LatencyHeatmap: Request Latency Heatmap Grid ───
//
// LatencyHeatmap renders a grid of latency values with color mapping
// (green/yellow/red based on thresholds). Rows = time periods, cols = endpoints.
//
// Usage:
//
//	lh := NewLatencyHeatmap()
//	lh.SetColumns([]string{"/api", "/auth", "/db"})
//	lh.SetThresholds(100, 500) // warn at 100ms, crit at 500ms
//	lh.AddRow("09:00", []int{50, 200, 800})
//	lh.Paint(buf)

// LatencyHeatmapStyle holds styling.
type LatencyHeatmapStyle struct {
	Normal   buffer.Style
	Warning  buffer.Style
	Critical buffer.Style
	Label    buffer.Style
	Border   buffer.Style
}

// DefaultLatencyHeatmapStyle returns defaults.
func DefaultLatencyHeatmapStyle() LatencyHeatmapStyle {
	normal := buffer.Style{Fg: buffer.RGB(34, 197, 94)}
	warn := buffer.Style{Fg: buffer.RGB(234, 179, 8)}
	crit := buffer.Style{Fg: buffer.RGB(239, 68, 68)}
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return LatencyHeatmapStyle{Normal: normal, Warning: warn, Critical: crit, Label: label, Border: border}
}

// HeatmapRow represents a single time-period row.
type HeatmapRow struct {
	Label   string
	Values  []int
	// cached display strings
	Strs    []string
}

// LatencyHeatmap renders a latency heatmap grid.
type LatencyHeatmap struct {
	BaseComponent
	mu sync.Mutex

	columns   []string
	rows      []HeatmapRow
	warnMs    int
	critMs    int
	cellWidth int
	style     LatencyHeatmapStyle
}

// NewLatencyHeatmap creates a LatencyHeatmap.
func NewLatencyHeatmap() *LatencyHeatmap {
	lh := &LatencyHeatmap{
		warnMs:    100,
		critMs:    500,
		cellWidth: 8,
		style:     DefaultLatencyHeatmapStyle(),
	}
	lh.SetID(GenerateID("latency"))
	return lh
}

// SetColumns sets endpoint column headers.
func (lh *LatencyHeatmap) SetColumns(cols []string) *LatencyHeatmap {
	lh.mu.Lock()
	lh.columns = cols
	lh.mu.Unlock()
	return lh
}

// SetThresholds sets warn/critical latency in ms.
func (lh *LatencyHeatmap) SetThresholds(warn, crit int) *LatencyHeatmap {
	lh.mu.Lock()
	lh.warnMs = warn
	lh.critMs = crit
	lh.mu.Unlock()
	return lh
}

// SetCellWidth sets the width of each cell in characters.
func (lh *LatencyHeatmap) SetCellWidth(w int) *LatencyHeatmap {
	lh.mu.Lock()
	if w < 4 { w = 4 }
	lh.cellWidth = w
	lh.mu.Unlock()
	return lh
}

// AddRow adds a time-period row with latency values (caches display strings).
func (lh *LatencyHeatmap) AddRow(label string, values []int) *LatencyHeatmap {
	lh.mu.Lock()
	row := HeatmapRow{Label: label, Values: values}
	row.Strs = make([]string, len(values))
	for i, v := range values {
		row.Strs[i] = itoa(v)
	}
	lh.rows = append(lh.rows, row)
	lh.mu.Unlock()
	return lh
}

// RowCount returns the number of data rows.
func (lh *LatencyHeatmap) RowCount() int {
	lh.mu.Lock()
	defer lh.mu.Unlock()
	return len(lh.rows)
}

// Clear removes all rows.
func (lh *LatencyHeatmap) Clear() *LatencyHeatmap {
	lh.mu.Lock()
	lh.rows = lh.rows[:0]
	lh.mu.Unlock()
	return lh
}

// SetStyle sets custom style.
func (lh *LatencyHeatmap) SetStyle(s LatencyHeatmapStyle) *LatencyHeatmap {
	lh.mu.Lock()
	lh.style = s
	lh.mu.Unlock()
	return lh
}

// latencyStyleLocked returns style for a latency value.
func (lh *LatencyHeatmap) latencyStyleLocked(v int) buffer.Style {
	if v >= lh.critMs { return lh.style.Critical }
	if v >= lh.warnMs { return lh.style.Warning }
	return lh.style.Normal
}

// Measure returns preferred size.
func (lh *LatencyHeatmap) Measure(cs Constraints) Size {
	lh.mu.Lock()
	colCount := len(lh.columns)
	rowCount := len(lh.rows)
	cw := lh.cellWidth
	lh.mu.Unlock()
	w := 12 + colCount*cw
	if w < 20 { w = 20 }
	h := rowCount + 3 // border + header + rows + border
	if h < 4 { h = 4 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the latency heatmap into the buffer.
func (lh *LatencyHeatmap) Paint(buf *buffer.Buffer) {
	lh.mu.Lock()
	defer lh.mu.Unlock()

	b := lh.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 20 { w = 40 }
	if h < 4 { h = 6 }

	bs := lh.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	labelStyle := lh.style.Label
	labelColW := 10

	// Column headers
	headerY := y + 1
	col := x + 1 + labelColW
	for _, colName := range lh.columns {
		// Center column name in cell
		nameLen := len(colName)
		offset := (lh.cellWidth - nameLen) / 2
		if offset < 0 { offset = 0 }
		for i := 0; i < lh.cellWidth && col < x+w-1 && col < buf.Width; i++ {
			var ch rune = ' '
			if i >= offset && i < offset+nameLen {
				ch = rune(colName[i-offset])
			}
			buf.SetCell(col, headerY, buffer.Cell{Rune: ch, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}
	}

	// Data rows
	for rowIdx, row := range lh.rows {
		rowY := y + 2 + rowIdx
		if rowY >= y+h-1 || rowY >= buf.Height { break }

		// Row label
		col := x + 1
		for _, r := range row.Label {
			if col >= x+labelColW || col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}
		// Pad to label width
		for col < x+1+labelColW && col < buf.Width {
			buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}

		// Cell values
		for _, valStr := range row.Strs {
			for i := 0; i < lh.cellWidth && col < x+w-1 && col < buf.Width; i++ {
				buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
				col++
			}
			// Overwrite with value right-aligned in cell
			valLen := len(valStr)
			valStart := col - lh.cellWidth + (lh.cellWidth - valLen) / 2
			if valStart < 0 { valStart = 0 }
			_ = valLen
			_ = valStart
		}
	}

	// Second pass: draw values with correct styling
	for rowIdx, row := range lh.rows {
		rowY := y + 2 + rowIdx
		if rowY >= y+h-1 || rowY >= buf.Height { break }

		col := x + 1 + labelColW
		for valIdx, valStr := range row.Strs {
			style := lh.latencyStyleLocked(row.Values[valIdx])
			valLen := len(valStr)
			valStart := col + (lh.cellWidth - valLen) / 2
			for i, r := range valStr {
				cx := valStart + i
				if cx >= x+w-1 || cx >= buf.Width { break }
				buf.SetCell(cx, rowY, buffer.Cell{Rune: r, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
			}
			col += lh.cellWidth
		}
	}
}

// Children returns nil.
func (lh *LatencyHeatmap) Children() []Component { return nil }

package component

import (
	"strconv"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── DataLabel: Styled Data Value Label ───
//
// DataLabel renders a labeled data value with optional unit, trend indicator
// (up/down/flat arrow), and color-coded sentiment. Useful in dashboards,
// metrics displays, and KPI cards.
//
// Usage:
//
//	dl := NewDataLabel()
//	dl.SetLabel("Revenue")
//	dl.SetValue(42.5)
//	dl.SetUnit("K")
//	dl.SetTrend(DataTrendUp)
//	dl.Paint(buf)

// DataTrend describes the value trend direction.
type DataTrend int

const (
	DataTrendFlat DataTrend = iota
	DataTrendUp
	DataTrendDown
)

// DataLabelStyle holds styling for DataLabel.
type DataLabelStyle struct {
	Label   buffer.Style
	Value   buffer.Style
	Unit    buffer.Style
	Up      buffer.Style
	Down    buffer.Style
	Flat    buffer.Style
	Border  buffer.Style
}

// DefaultDataLabelStyle returns defaults.
func DefaultDataLabelStyle() DataLabelStyle {
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)}    // slate-400
	value := buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold} // slate-200 bold
	unit := buffer.Style{Fg: buffer.RGB(100, 116, 139)}    // slate-500
	up := buffer.Style{Fg: buffer.RGB(34, 197, 94)}        // green-500
	down := buffer.Style{Fg: buffer.RGB(239, 68, 68)}      // red-500
	flat := buffer.Style{Fg: buffer.RGB(148, 163, 184)}    // slate-400
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}    // slate-600
	return DataLabelStyle{Label: label, Value: value, Unit: unit, Up: up, Down: down, Flat: flat, Border: border}
}

// DataLabel renders a labeled data value with trend indicator.
type DataLabel struct {
	BaseComponent
	mu sync.Mutex

	label    string
	value    float64
	unit     string
	trend    DataTrend
	style    DataLabelStyle
	// cached formatted value string
	valStr string
}

// NewDataLabel creates a DataLabel with defaults.
func NewDataLabel() *DataLabel {
	dl := &DataLabel{style: DefaultDataLabelStyle()}
	dl.SetID(GenerateID("datalabel"))
	return dl
}

// SetLabel sets the display label.
func (dl *DataLabel) SetLabel(l string) *DataLabel {
	dl.mu.Lock()
	dl.label = l
	dl.mu.Unlock()
	return dl
}

// Label returns the label.
func (dl *DataLabel) Label() string {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	return dl.label
}

// SetValue sets the numeric value.
func (dl *DataLabel) SetValue(v float64) *DataLabel {
	dl.mu.Lock()
	dl.value = v
	dl.valStr = strconv.FormatFloat(v, 'f', 1, 64)
	dl.mu.Unlock()
	return dl
}

// Value returns the current value.
func (dl *DataLabel) Value() float64 {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	return dl.value
}

// SetUnit sets the unit suffix.
func (dl *DataLabel) SetUnit(u string) *DataLabel {
	dl.mu.Lock()
	dl.unit = u
	dl.mu.Unlock()
	return dl
}

// Unit returns the unit.
func (dl *DataLabel) Unit() string {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	return dl.unit
}

// SetTrend sets the trend direction.
func (dl *DataLabel) SetTrend(t DataTrend) *DataLabel {
	dl.mu.Lock()
	dl.trend = t
	dl.mu.Unlock()
	return dl
}

// Trend returns the trend.
func (dl *DataLabel) Trend() DataTrend {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	return dl.trend
}

// SetStyle sets the custom style.
func (dl *DataLabel) SetStyle(s DataLabelStyle) *DataLabel {
	dl.mu.Lock()
	dl.style = s
	dl.mu.Unlock()
	return dl
}

// Measure returns the preferred size.
func (dl *DataLabel) Measure(cs Constraints) Size {
	w := 20
	h := 3 // label + value + border
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the data label into the buffer.
func (dl *DataLabel) Paint(buf *buffer.Buffer) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	b := dl.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 { w = 20 }
	if h < 3 { h = 3 }

	// Draw border
	bs := dl.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	// Label on first row
	labelStyle := dl.style.Label
	col := x + 1
	for _, r := range dl.label {
		if col >= x+w-1 || col >= buf.Width { break }
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Value + unit + trend on second row
	valStyle := dl.style.Value
	unitStyle := dl.style.Unit
	col = x + 1

	// Trend arrow
	var trendChar rune
	var trendStyle buffer.Style
	switch dl.trend {
	case DataTrendUp:
		trendChar = '↑'
		trendStyle = dl.style.Up
	case DataTrendDown:
		trendChar = '↓'
		trendStyle = dl.style.Down
	default:
		trendChar = '→'
		trendStyle = dl.style.Flat
	}
	if col < x+w-1 && col < buf.Width {
		buf.SetCell(col, y+2, buffer.Cell{Rune: trendChar, Fg: trendStyle.Fg, Bg: trendStyle.Bg, Flags: trendStyle.Flags, Width: 1})
	}
	col++
	if col < x+w-1 && col < buf.Width {
		buf.SetCell(col, y+2, buffer.Cell{Rune: ' ', Fg: valStyle.Fg, Bg: valStyle.Bg, Flags: valStyle.Flags, Width: 1})
	}
	col++

	// Value
	for _, r := range dl.valStr {
		if col >= x+w-1 || col >= buf.Width { break }
		buf.SetCell(col, y+2, buffer.Cell{Rune: r, Fg: valStyle.Fg, Bg: valStyle.Bg, Flags: valStyle.Flags, Width: 1})
		col++
	}

	// Unit
	for _, r := range dl.unit {
		if col >= x+w-1 || col >= buf.Width { break }
		buf.SetCell(col, y+2, buffer.Cell{Rune: r, Fg: unitStyle.Fg, Bg: unitStyle.Bg, Flags: unitStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (dl *DataLabel) Children() []Component { return nil }

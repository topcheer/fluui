package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Legend: Chart Legend Display ───
//
// Legend renders a chart legend with colored markers and labels, showing
// data series names with their corresponding colors. Useful alongside
// charts (pie, bar, line) for identifying data series.
//
// Usage:
//
//	l := NewLegend()
//	l.AddEntry("Revenue", buffer.RGB(34, 197, 94))
//	l.AddEntry("Costs", buffer.RGB(239, 68, 68))
//	l.AddEntry("Profit", buffer.RGB(96, 165, 250))
//	l.Paint(buf)

// LegendEntry represents a single legend item.
type LegendEntry struct {
	Label string
	Color buffer.Color
}

// LegendStyle holds styling for Legend.
type LegendStyle struct {
	Label  buffer.Style
	Border buffer.Style
}

// DefaultLegendStyle returns defaults.
func DefaultLegendStyle() LegendStyle {
	label := buffer.Style{Fg: buffer.RGB(226, 232, 240)} // slate-200
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}  // slate-600
	return LegendStyle{Label: label, Border: border}
}

// Legend displays chart data series with colored markers.
type Legend struct {
	BaseComponent
	mu sync.Mutex

	entries []LegendEntry
	style   LegendStyle
}

// NewLegend creates a Legend with defaults.
func NewLegend() *Legend {
	l := &Legend{style: DefaultLegendStyle()}
	l.SetID(GenerateID("legend"))
	return l
}

// AddEntry adds a legend item with label and color.
func (l *Legend) AddEntry(label string, color buffer.Color) *Legend {
	l.mu.Lock()
	l.entries = append(l.entries, LegendEntry{Label: label, Color: color})
	l.mu.Unlock()
	return l
}

// SetEntries replaces all entries.
func (l *Legend) SetEntries(entries []LegendEntry) *Legend {
	l.mu.Lock()
	l.entries = entries
	l.mu.Unlock()
	return l
}

// EntryCount returns the number of legend entries.
func (l *Legend) EntryCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries)
}

// Clear removes all entries.
func (l *Legend) Clear() *Legend {
	l.mu.Lock()
	l.entries = l.entries[:0]
	l.mu.Unlock()
	return l
}

// SetStyle sets the custom style.
func (l *Legend) SetStyle(s LegendStyle) *Legend {
	l.mu.Lock()
	l.style = s
	l.mu.Unlock()
	return l
}

// Measure returns the preferred size.
func (l *Legend) Measure(cs Constraints) Size {
	l.mu.Lock()
	count := len(l.entries)
	l.mu.Unlock()
	w := 25
	h := count + 2
	if h < 3 { h = 3 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the legend into the buffer.
func (l *Legend) Paint(buf *buffer.Buffer) {
	l.mu.Lock()
	defer l.mu.Unlock()

	b := l.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 { w = 25 }
	if h < 3 { h = 3 }

	bs := l.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	labelStyle := l.style.Label

	for idx, entry := range l.entries {
		rowY := y + 1 + idx
		if rowY >= y+h-1 || rowY >= buf.Height { break }

		col := x + 2
		// Color marker (█ in entry color)
		if col < x+w-1 && col < buf.Width {
			buf.SetCell(col, rowY, buffer.Cell{Rune: '█', Fg: entry.Color, Bg: buffer.NoColor(), Width: 1})
		}
		col++
		// Space
		if col < x+w-1 && col < buf.Width {
			buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		}
		col++
		// Label
		for _, r := range entry.Label {
			if col >= x+w-1 || col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (l *Legend) Children() []Component { return nil }

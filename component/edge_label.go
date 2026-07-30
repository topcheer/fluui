package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── EdgeLabel: Labeled Edge/Connector Between Components ───
//
// EdgeLabel renders a labeled connector line (horizontal or vertical)
// with an optional label in the middle. Useful for flow diagrams,
// org charts, and data pipeline visualizations.
//
// Usage:
//
//	el := NewEdgeLabel()
//	el.SetLabel("data")
//	el.SetLength(20)
//	el.SetVertical(false)
//	el.Paint(buf)

// EdgeLabelStyle holds styling.
type EdgeLabelStyle struct {
	Line  buffer.Style
	Label buffer.Style
}

// DefaultEdgeLabelStyle returns defaults.
func DefaultEdgeLabelStyle() EdgeLabelStyle {
	return EdgeLabelStyle{
		Line:  buffer.Style{Fg: buffer.RGB(71, 85, 105)},
		Label: buffer.Style{Fg: buffer.RGB(148, 163, 184), Flags: buffer.Italic},
	}
}

// EdgeLabel renders a labeled connector line.
type EdgeLabel struct {
	BaseComponent
	mu sync.Mutex

	label    string
	length   int
	vertical bool
	style    EdgeLabelStyle
	// cached
	labelRunes []rune
}

// NewEdgeLabel creates an EdgeLabel.
func NewEdgeLabel() *EdgeLabel {
	el := &EdgeLabel{length: 16, style: DefaultEdgeLabelStyle()}
	el.SetID(GenerateID("edgelabel"))
	el.recomputeLocked()
	return el
}

// SetLabel sets the edge label text.
func (el *EdgeLabel) SetLabel(s string) *EdgeLabel {
	el.mu.Lock()
	el.label = s
	el.recomputeLocked()
	el.mu.Unlock()
	return el
}

// SetLength sets the edge length.
func (el *EdgeLabel) SetLength(n int) *EdgeLabel {
	el.mu.Lock()
	if n < 3 {
		n = 3
	}
	el.length = n
	el.mu.Unlock()
	return el
}

// SetVertical toggles vertical orientation.
func (el *EdgeLabel) SetVertical(v bool) *EdgeLabel {
	el.mu.Lock()
	el.vertical = v
	el.mu.Unlock()
	return el
}

func (el *EdgeLabel) recomputeLocked() {
	el.labelRunes = []rune(el.label)
}

// Label returns the current label.
func (el *EdgeLabel) Label() string {
	el.mu.Lock()
	defer el.mu.Unlock()
	return el.label
}

// SetStyle sets custom style.
func (el *EdgeLabel) SetStyle(s EdgeLabelStyle) *EdgeLabel {
	el.mu.Lock()
	el.style = s
	el.mu.Unlock()
	return el
}

// Measure returns preferred size.
func (el *EdgeLabel) Measure(cs Constraints) Size {
	el.mu.Lock()
	defer el.mu.Unlock()
	if el.vertical {
		w := 3
		if len(el.labelRunes) > w {
			w = len(el.labelRunes)
		}
		return Size{W: w, H: el.length}
	}
	return Size{W: el.length, H: 1}
}

// Paint renders the edge label.
func (el *EdgeLabel) Paint(buf *buffer.Buffer) {
	el.mu.Lock()
	defer el.mu.Unlock()

	b := el.Bounds()
	x, y := b.X, b.Y

	lineStyle := el.style.Line
	labelStyle := el.style.Label

	if el.vertical {
		midY := y + el.length/2
		for row := 0; row < el.length; row++ {
			yy := y + row
			if yy >= buf.Height {
				break
			}
			if yy == midY {
				col := x
				for _, r := range el.labelRunes {
					if col >= buf.Width {
						break
					}
					buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
					col++
				}
			} else {
				if x < buf.Width {
					buf.SetCell(x, yy, buffer.Cell{Rune: '│', Fg: lineStyle.Fg, Bg: lineStyle.Bg, Flags: lineStyle.Flags, Width: 1})
				}
			}
		}
	} else {
		mid := el.length / 2
		labelLen := len(el.labelRunes)
		labelStart := mid - labelLen/2

		col := x
		for i := 0; i < el.length; i++ {
			if col >= buf.Width {
				break
			}
			if i >= labelStart && i < labelStart+labelLen {
				idx := i - labelStart
				buf.SetCell(col, y, buffer.Cell{Rune: el.labelRunes[idx], Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			} else {
				buf.SetCell(col, y, buffer.Cell{Rune: '─', Fg: lineStyle.Fg, Bg: lineStyle.Bg, Flags: lineStyle.Flags, Width: 1})
			}
			col++
		}
	}
}

// Children returns nil.
func (el *EdgeLabel) Children() []Component { return nil }

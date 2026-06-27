package component

import (
	"github.com/topcheer/fluui/internal/buffer"
)

// Border is a decorative container that draws a box around its child.
type Border struct {
	BaseComponent
	Child Component
	Title string
	Style buffer.Style
}

// NewBorder creates a Border wrapping the given child.
func NewBorder(child Component) *Border {
	return &Border{
		Child: child,
		Style: buffer.Style{},
	}
}

// Measure returns the child's desired size plus 2 columns (left/right borders)
// and 2 rows (top/bottom borders).
func (b *Border) Measure(cs Constraints) Size {
	// Inset constraints for the child: subtract border padding
	inner := Constraints{
		MinWidth:  max0(cs.MinWidth - 2),
		MaxWidth:  max0(cs.MaxWidth - 2),
		MinHeight: max0(cs.MinHeight - 2),
		MaxHeight: max0(cs.MaxHeight - 2),
	}

	childSize := b.Child.Measure(inner)
	return Size{
		W: childSize.W + 2,
		H: childSize.H + 2,
	}
}

// Paint draws the border frame, title, and then the child inside.
func (b *Border) Paint(buf *buffer.Buffer) {
	x, y := b.bounds.X, b.bounds.Y
	w, h := b.bounds.W, b.bounds.H
	if w < 2 || h < 2 {
		return
	}

	s := b.Style

	// Corners
	buf.SetCell(x, y, buffer.Cell{Rune: '\u250c', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags}) // top-left
	buf.SetCell(x+w-1, y, buffer.Cell{Rune: '\u2510', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags}) // top-right
	buf.SetCell(x, y+h-1, buffer.Cell{Rune: '\u2514', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags}) // bottom-left
	buf.SetCell(x+w-1, y+h-1, buffer.Cell{Rune: '\u2518', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags}) // bottom-right

	// Top and bottom edges
	for i := 1; i < w-1; i++ {
		buf.SetCell(x+i, y, buffer.Cell{Rune: '\u2500', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
		buf.SetCell(x+i, y+h-1, buffer.Cell{Rune: '\u2500', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
	}

	// Left and right edges
	for i := 1; i < h-1; i++ {
		buf.SetCell(x, y+i, buffer.Cell{Rune: '\u2502', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
		buf.SetCell(x+w-1, y+i, buffer.Cell{Rune: '\u2502', Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
	}

	// Title centered on top edge
	if b.Title != "" {
		titleW := buffer.StringWidth(b.Title)
		startX := x + (w-titleW)/2
		if startX < x+1 {
			startX = x + 1
		}
		buf.DrawText(startX, y, b.Title, s)
	}

	// Layout child in the inner area
	innerRect := Rect{
		X: x + 1,
		Y: y + 1,
		W: w - 2,
		H: h - 2,
	}
	b.Child.SetBounds(innerRect)
	b.Child.Paint(buf)
}

func max0(v int) int {
	if v < 0 {
		return 0
	}
	return v
}

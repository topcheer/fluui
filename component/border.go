package component

import (
	"github.com/topcheer/fluui/internal/buffer"
)

// BorderType specifies the visual style of border characters.
type BorderType uint8

const (
	// BorderSingle uses light single-line box drawing (default).
	BorderSingle BorderType = iota
	// BorderRounded uses single-line with rounded corners.
	BorderRounded
	// BorderDouble uses double-line box drawing.
	BorderDouble
	// BorderHeavy uses heavy/bold single-line box drawing.
	BorderHeavy
	// BorderNone draws no visible border. The child still gets 1-cell padding
	// for the border area (rendered as blank), preserving layout measurements.
	BorderNone
)

// borderRunes holds the 8 rune values for each border style.
// Order: topLeft, topRight, bottomLeft, bottomRight, horizontal, vertical,
// topLeftJoin, topRightJoin (joins connect title to border).
type borderRunes struct {
	tl, tr, bl, br rune
	h, v           rune
}

var borderRuneSets = map[BorderType]borderRunes{
	BorderSingle:  {'\u250c', '\u2510', '\u2514', '\u2518', '\u2500', '\u2502'},
	BorderRounded: {'\u256d', '\u256e', '\u2570', '\u256f', '\u2500', '\u2502'},

	BorderDouble:  {'\u2554', '\u2557', '\u255a', '\u255d', '\u2550', '\u2551'},
	BorderHeavy:   {'\u250f', '\u2513', '\u2517', '\u251b', '\u2501', '\u2503'},
	BorderNone:    {' ', ' ', ' ', ' ', ' ', ' '},
}

// TitleAlign specifies horizontal title alignment on the top border.
type TitleAlign uint8

const (
	// TitleCenter centers the title (default, backward compatible).
	TitleCenter TitleAlign = iota
	// TitleLeft left-aligns the title.
	TitleLeft
	// TitleRight right-aligns the title.
	TitleRight
)

// BorderSide is a bitmask specifying which sides of the border to draw.
type BorderSide uint8

const (
	BorderTop    BorderSide = 1 << iota // 1
	BorderBottom                        // 2
	BorderLeft                          // 4
	BorderRight                         // 8
	// BorderAll draws all four sides.
	BorderAll BorderSide = BorderTop | BorderBottom | BorderLeft | BorderRight
)

// Padding specifies inner spacing between the border and the child component.
type Padding struct {
	Top, Bottom, Left, Right int
}

// NoPadding is a zero-value padding shortcut.
func NoPadding() Padding { return Padding{} }

// UniformPadding returns padding with the same value on all sides.
func UniformPadding(n int) Padding {
	return Padding{Top: n, Bottom: n, Left: n, Right: n}
}

// Horizontal returns left + right padding.
func (p Padding) Horizontal() int { return p.Left + p.Right }

// Vertical returns top + bottom padding.
func (p Padding) Vertical() int { return p.Top + p.Bottom }

// Border is a decorative container that draws a configurable box around its child.
//
// By default, the border uses BorderSingle style with centered title and all
// sides visible, matching the pre-P35 behavior.
type Border struct {
	BaseComponent
	Child  Component
	Title  string
	Style  buffer.Style
	// Type sets the border character style. Default: BorderSingle.
	Type BorderType
	// TitleAlign sets the horizontal title position. Default: TitleCenter.
	TitleAlign TitleAlign
	// Sides controls which sides are drawn. Default: BorderAll.
	Sides BorderSide
	// InnerPadding sets spacing between the border and the child.
	InnerPadding Padding
}

// NewBorder creates a Border wrapping the given child with default settings.
func NewBorder(child Component) *Border {
	return &Border{
		Child:      child,
		Style:      buffer.Style{},
		Type:       BorderSingle,
		TitleAlign: TitleCenter,
		Sides:      BorderAll,
	}
}

// totalHorizontalBorder returns how many columns the border + padding consume.
func (b *Border) totalHorizontalBorder() int {
	pad := b.InnerPadding.Horizontal()
	if b.Sides&BorderLeft != 0 {
		pad++
	}
	if b.Sides&BorderRight != 0 {
		pad++
	}
	return pad
}

func (b *Border) totalVerticalBorder() int {
	pad := b.InnerPadding.Vertical()
	if b.Sides&BorderTop != 0 {
		pad++
	}
	if b.Sides&BorderBottom != 0 {
		pad++
	}
	return pad
}

// Measure returns the child's desired size plus border and padding.
func (b *Border) Measure(cs Constraints) Size {
	dw := b.totalHorizontalBorder()
	dh := b.totalVerticalBorder()

	inner := Constraints{
		MinWidth:  max0(cs.MinWidth - dw),
		MaxWidth:  max0(cs.MaxWidth - dw),
		MinHeight: max0(cs.MinHeight - dh),
		MaxHeight: max0(cs.MaxHeight - dh),
	}

	childSize := b.Child.Measure(inner)
	return Size{
		W: childSize.W + dw,
		H: childSize.H + dh,
	}
}

// Paint draws the border frame, title, and then the child inside.
func (b *Border) Paint(buf *buffer.Buffer) {
	x, y := b.bounds.X, b.bounds.Y
	w, h := b.bounds.W, b.bounds.H

	dw := b.totalHorizontalBorder()
	dh := b.totalVerticalBorder()

	if w < dw || h < dh {
		return
	}

	s := b.Style
	rs := borderRuneSets[b.Type]

	hasTop := b.Sides&BorderTop != 0
	hasBottom := b.Sides&BorderBottom != 0
	hasLeft := b.Sides&BorderLeft != 0
	hasRight := b.Sides&BorderRight != 0

	// Draw top border line
	if hasTop {
		for i := 0; i < w; i++ {
			cx := x + i
			cell := buf.GetCell(cx, y)
			if i == 0 && hasLeft {
				buf.SetCell(cx, y, buffer.Cell{Rune: rs.tl, Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
			} else if i == w-1 && hasRight {
				buf.SetCell(cx, y, buffer.Cell{Rune: rs.tr, Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
			} else if i == 0 && !hasLeft {
				// corner but no left side — draw horizontal edge
				buf.SetCell(cx, y, buffer.Cell{Rune: rs.h, Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
			} else if i == w-1 && !hasRight {
				buf.SetCell(cx, y, buffer.Cell{Rune: rs.h, Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
			} else {
				// Only draw horizontal if cell is currently blank or matches edge style
				if cell.Rune == ' ' || cell.Rune == 0 {
					buf.SetCell(cx, y, buffer.Cell{Rune: rs.h, Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
				}
			}
		}
	}

	// Draw bottom border line
	if hasBottom {
		for i := 0; i < w; i++ {
			cx := x + i
			if i == 0 && hasLeft {
				buf.SetCell(cx, y+h-1, buffer.Cell{Rune: rs.bl, Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
			} else if i == w-1 && hasRight {
				buf.SetCell(cx, y+h-1, buffer.Cell{Rune: rs.br, Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
			} else if i == 0 && !hasLeft {
				buf.SetCell(cx, y+h-1, buffer.Cell{Rune: rs.h, Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
			} else if i == w-1 && !hasRight {
				buf.SetCell(cx, y+h-1, buffer.Cell{Rune: rs.h, Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
			} else {
				buf.SetCell(cx, y+h-1, buffer.Cell{Rune: rs.h, Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
			}
		}
	}

	// Draw left and right vertical edges
	for i := 1; i < h-1; i++ {
		if hasLeft {
			buf.SetCell(x, y+i, buffer.Cell{Rune: rs.v, Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
		}
		if hasRight {
			buf.SetCell(x+w-1, y+i, buffer.Cell{Rune: rs.v, Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
		}
	}

	// Draw title on top border
	if b.Title != "" && hasTop {
		titleW := buffer.StringWidth(b.Title)
		var startX int
		innerW := w
		switch b.TitleAlign {
		case TitleLeft:
			startX = x + 1 + b.InnerPadding.Left
		case TitleRight:
			startX = x + w - 1 - b.InnerPadding.Right - titleW
			if startX < x+1 {
				startX = x + 1
			}
		default: // TitleCenter
			startX = x + (innerW-titleW)/2
			if startX < x+1 {
				startX = x + 1
			}
		}
		// Clamp title within bounds
		maxEnd := x + w - 1
		for i, r := range b.Title {
			tx := startX + i
			if tx >= maxEnd {
				break
			}
			if tx >= x {
				buf.SetCell(tx, y, buffer.Cell{Rune: r, Width: 1, Fg: s.Fg, Bg: s.Bg, Flags: s.Flags})
			}
		}
	}

	// Layout child in the inner area
	innerX := x
	if hasLeft {
		innerX++
	}
	innerY := y
	if hasTop {
		innerY++
	}
	innerX += b.InnerPadding.Left
	innerY += b.InnerPadding.Top

	innerW := w - dw
	innerH := h - dh
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}

	b.Child.SetBounds(Rect{
		X: innerX,
		Y: innerY,
		W: innerW,
		H: innerH,
	})
	b.Child.Paint(buf)
}

func max0(v int) int {
	if v < 0 {
		return 0
	}
	return v
}

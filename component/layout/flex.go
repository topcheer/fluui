// Package layout implements layout containers that position child components.
//
// Flex is the primary layout primitive: it arranges children in a row or column
// with configurable gaps, similar to CSS flexbox in its simplest form.
package layout

import (
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// Direction determines the main axis of a Flex container.
type Direction uint8

const (
	// FlexRow arranges children left-to-right (horizontal).
	FlexRow Direction = iota
	// FlexColumn arranges children top-to-bottom (vertical).
	FlexColumn
)

// Flex arranges child components along a single axis (row or column).
// Gap defines spacing between adjacent children.
type Flex struct {
	component.BaseComponent

	Direction Direction
	Gap       int
	children  []component.Component
}

// NewFlex creates a Flex with the given direction and no gap.
func NewFlex(dir Direction) *Flex {
	return &Flex{
		Direction: dir,
	}
}

// NewFlexGap creates a Flex with the given direction and gap.
func NewFlexGap(dir Direction, gap int) *Flex {
	return &Flex{
		Direction: dir,
		Gap:       gap,
	}
}

// AddChild appends a child component.
func (f *Flex) AddChild(c component.Component) {
	f.children = append(f.children, c)
}

// Children returns the child components.
func (f *Flex) Children() []component.Component {
	return f.children
}

// Measure computes the desired size of the Flex based on its children.
//
// Row mode:  width = sum(child widths) + gaps, height = max(child heights)
// Column mode: height = sum(child heights) + gaps, width = max(child widths)
func (f *Flex) Measure(cs component.Constraints) component.Size {
	if len(f.children) == 0 {
		return component.Size{}
	}

	totalMain := 0
	maxCross := 0

	for i, child := range f.children {
		size := child.Measure(component.Unbounded())

		if f.Direction == FlexRow {
			totalMain += size.W
			if size.H > maxCross {
				maxCross = size.H
			}
		} else {
			totalMain += size.H
			if size.W > maxCross {
				maxCross = size.W
			}
		}

		// Add gap between children (not after the last one)
		if i < len(f.children)-1 {
			totalMain += f.Gap
		}
	}

	w, h := totalMain, maxCross
	if f.Direction == FlexColumn {
		w, h = maxCross, totalMain
	}

	// Respect max constraints if specified
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}

	return component.Size{W: w, H: h}
}

// SetBounds assigns positions to all children within the given bounds.
//
// Each child gets its measured size along the main axis and is placed
// sequentially with gaps. The cross axis is set to the child's measured size
// (left/top aligned).
func (f *Flex) SetBounds(r component.Rect) {
	f.BaseComponent.SetBounds(r)

	if len(f.children) == 0 {
		return
	}

	offset := 0

	for _, child := range f.children {
		size := child.Measure(component.Unbounded())

		if f.Direction == FlexRow {
			child.SetBounds(component.Rect{
				X: r.X + offset,
				Y: r.Y,
				W: size.W,
				H: size.H,
			})
			offset += size.W + f.Gap
		} else {
			child.SetBounds(component.Rect{
				X: r.X,
				Y: r.Y + offset,
				W: size.W,
				H: size.H,
			})
			offset += size.H + f.Gap
		}
	}
}

// Paint renders all children into the buffer.
func (f *Flex) Paint(buf *buffer.Buffer) {
	for _, child := range f.children {
		child.Paint(buf)
	}
}

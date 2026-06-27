package layout

import (
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// Padding adds uniform or non-uniform inner padding around a single child.
// The child is positioned inside the padding area.
type Padding struct {
	component.BaseComponent
	top, right, bottom, left int
	child                    component.Component
}

// NewPadding creates a Padding component with the given insets.
func NewPadding(top, right, bottom, left int, child component.Component) *Padding {
	return &Padding{
		top:    top,
		right:  right,
		bottom: bottom,
		left:   left,
		child:  child,
	}
}

// Measure returns the child's size plus the padding on each axis.
func (p *Padding) Measure(cs component.Constraints) component.Size {
	// Reduce available space by padding
	innerW := cs.MaxWidth
	innerH := cs.MaxHeight
	if innerW > 0 {
		innerW -= p.left + p.right
		if innerW < 0 {
			innerW = 0
		}
	}
	if innerH > 0 {
		innerH -= p.top + p.bottom
		if innerH < 0 {
			innerH = 0
		}
	}

	childSize := p.child.Measure(component.Bounded(innerW, innerH))
	return component.Size{
		W: childSize.W + p.left + p.right,
		H: childSize.H + p.top + p.bottom,
	}
}

// SetBounds positions the child inside the padding area of the given bounds.
func (p *Padding) SetBounds(r component.Rect) {
	p.BaseComponent.SetBounds(r)

	innerX := r.X + p.left
	innerY := r.Y + p.top
	innerW := r.W - p.left - p.right
	innerH := r.H - p.top - p.bottom
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}

	p.child.SetBounds(component.Rect{
		X: innerX,
		Y: innerY,
		W: innerW,
		H: innerH,
	})
}

// Paint renders the padded child.
func (p *Padding) Paint(buf *buffer.Buffer) {
	p.child.Paint(buf)
}

// Children returns the padded child.
func (p *Padding) Children() []component.Component {
	return []component.Component{p.child}
}

package layout

import (
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// Center centers its single child both horizontally and vertically
// within the allocated bounds.
type Center struct {
	component.BaseComponent
	child component.Component
}

// NewCenter creates a Center layout wrapping the given child.
func NewCenter(child component.Component) *Center {
	return &Center{
		child: child,
	}
}

// Measure returns the child's natural size.
func (c *Center) Measure(cs component.Constraints) component.Size {
	return c.child.Measure(cs)
}

// SetBounds positions the child centered within the given bounds.
// If the child is larger than the bounds, it is top-left aligned.
func (c *Center) SetBounds(r component.Rect) {
	c.BaseComponent.SetBounds(r)

	childSize := c.child.Measure(component.Bounded(r.W, r.H))
	cw := childSize.W
	ch := childSize.H
	if cw > r.W {
		cw = r.W
	}
	if ch > r.H {
		ch = r.H
	}

	offsetX := (r.W - cw) / 2
	offsetY := (r.H - ch) / 2
	if offsetX < 0 {
		offsetX = 0
	}
	if offsetY < 0 {
		offsetY = 0
	}

	c.child.SetBounds(component.Rect{
		X: r.X + offsetX,
		Y: r.Y + offsetY,
		W: cw,
		H: ch,
	})
}

// Paint renders the centered child.
func (c *Center) Paint(buf *buffer.Buffer) {
	c.child.Paint(buf)
}

// Children returns the centered child.
func (c *Center) Children() []component.Component {
	return []component.Component{c.child}
}

package layout

import (
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// Stack arranges children by overlapping them: all children receive the
// same bounds. Children are painted in order, so later children appear on top.
// This is useful for overlays, badges, or layered compositions.
type Stack struct {
	component.BaseComponent
	children []component.Component
}

// NewStack creates a Stack with the given children. Later children
// are painted on top of earlier ones.
func NewStack(children ...component.Component) *Stack {
	return &Stack{
		children: children,
	}
}

// Measure returns the maximum width and height across all children.
func (s *Stack) Measure(cs component.Constraints) component.Size {
	maxW := 0
	maxH := 0
	for _, child := range s.children {
		sz := child.Measure(cs)
		if sz.W > maxW {
			maxW = sz.W
		}
		if sz.H > maxH {
			maxH = sz.H
		}
	}
	if cs.MaxWidth > 0 && maxW > cs.MaxWidth {
		maxW = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && maxH > cs.MaxHeight {
		maxH = cs.MaxHeight
	}
	return component.Size{W: maxW, H: maxH}
}

// SetBounds sets the bounds for the stack and all children.
// Every child receives the exact same bounds as the stack.
func (s *Stack) SetBounds(r component.Rect) {
	s.BaseComponent.SetBounds(r)
	for _, child := range s.children {
		child.SetBounds(r)
	}
}

// Paint renders all children in order. Later children paint on top.
func (s *Stack) Paint(buf *buffer.Buffer) {
	for _, child := range s.children {
		child.Paint(buf)
	}
}

// Children returns the stack's child components.
func (s *Stack) Children() []component.Component {
	return s.children
}

// AddChild appends a child to the top of the stack.
func (s *Stack) AddChild(child component.Component) {
	s.children = append(s.children, child)
}

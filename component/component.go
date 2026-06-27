// Package component defines the core component system for the TUI.
//
// A Component is a self-contained UI element that can measure itself,
// be laid out within bounds, and paint into a buffer.
package component

import (
	"fmt"
	"sync/atomic"

	"github.com/topcheer/fluui/internal/buffer"
)

// Rect describes a rectangular area.
type Rect struct {
	X, Y, W, H int
}

// Size describes a 2D size.
type Size struct {
	W, H int
}

// Constraints describe the layout bounds a component must fit within.
// A value of 0 means "no constraint" in that dimension.
type Constraints struct {
	MaxWidth  int
	MaxHeight int
	MinWidth  int
	MinHeight int
	Has       bool // marks this constraint as explicitly set (used by some components)
}

// HasWidth returns true if MaxWidth is specified (> 0).
func (c Constraints) HasWidth() bool { return c.MaxWidth > 0 }

// HasHeight returns true if MaxHeight is specified (> 0).
func (c Constraints) HasHeight() bool { return c.MaxHeight > 0 }

// Clamp ensures Min ≤ Max; if a dimension is invalid, it zeroes it.
func (c *Constraints) Clamp() {
	if c.MinWidth > c.MaxWidth && c.MaxWidth > 0 {
		c.MinWidth = c.MaxWidth
	}
	if c.MinHeight > c.MaxHeight && c.MaxHeight > 0 {
		c.MinHeight = c.MaxHeight
	}
}

// Unbounded returns constraints with no limits.
func Unbounded() Constraints {
	return Constraints{}
}

// Fixed returns constraints that require an exact size.
func Fixed(w, h int) Constraints {
	return Constraints{MinWidth: w, MaxWidth: w, MinHeight: h, MaxHeight: h}
}

// ClampWidth clamps the given width to the constraints.
func (c Constraints) ClampWidth(w int) int {
	if c.MaxWidth > 0 && w > c.MaxWidth {
		w = c.MaxWidth
	}
	if c.MinWidth > 0 && w < c.MinWidth {
		w = c.MinWidth
	}
	return w
}

// ClampHeight clamps the given height to the constraints.
func (c Constraints) ClampHeight(h int) int {
	if c.MaxHeight > 0 && h > c.MaxHeight {
		h = c.MaxHeight
	}
	if c.MinHeight > 0 && h < c.MinHeight {
		h = c.MinHeight
	}
	return h
}

// Bounded returns constraints with only maximum limits.
func Bounded(maxW, maxH int) Constraints {
	return Constraints{MaxWidth: maxW, MaxHeight: maxH}
}

// Component is the core interface every UI element implements.
type Component interface {
	// ID returns a unique identifier for this component.
	ID() string

	// Measure computes the component's desired size given layout constraints.
	// The returned Size must fit within the constraints (if non-zero).
	Measure(cs Constraints) Size

	// SetBounds sets the final position and size assigned by the parent's layout.
	SetBounds(r Rect)

	// Bounds returns the current position and size.
	Bounds() Rect

	// Paint renders the component into the given buffer.
	// The buffer is guaranteed to be at least Bounds().W × Bounds().H.
	Paint(buf *buffer.Buffer)

	// Children returns child components, if any.
	// Leaf components return nil.
	Children() []Component
}

// BaseComponent provides default implementations for common Component methods.
// Embed this in concrete components to reduce boilerplate.
type BaseComponent struct {
	id    string
	bounds Rect
}

// SetID sets the component identifier.
func (b *BaseComponent) SetID(id string) { b.id = id }

// ID returns the component identifier.
func (b *BaseComponent) ID() string { return b.id }

// SetBounds sets the component's position and size.
func (b *BaseComponent) SetBounds(r Rect) { b.bounds = r }

// Bounds returns the component's current bounds.
func (b *BaseComponent) Bounds() Rect { return b.bounds }

// Measure returns a zero size by default. Concrete components override this.
func (b *BaseComponent) Measure(cs Constraints) Size { return Size{} }

// Paint does nothing by default. Concrete components override this.
func (b *BaseComponent) Paint(buf *buffer.Buffer) {}

// Children returns nil by default. Leaf components inherit this.
func (b *BaseComponent) Children() []Component { return nil }

// idCounter is used by GenerateID to produce unique component IDs.
var idCounter uint64

// GenerateID returns a unique identifier string with the given prefix.
// Example: GenerateID("padding") → "padding-1"
func GenerateID(prefix string) string {
	n := atomic.AddUint64(&idCounter, 1)
	return fmt.Sprintf("%s-%d", prefix, n)
}

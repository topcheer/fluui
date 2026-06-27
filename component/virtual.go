package component

import (
	"github.com/topcheer/fluui/internal/buffer"
)

// VisiblePainter is an optional interface for containers that can optimize
// rendering by painting only children visible within a Y range.
//
// ScrollView detects this interface and uses PaintVisible instead of Paint
// when the child supports it, enabling virtual scrolling for large lists.
//
// The visibleY0 and visibleY1 parameters are in the child's local coordinate
// space (relative to the child's top edge, not the screen).
type VisiblePainter interface {
	// PaintVisible paints only children whose Y range intersects
	// [visibleY0, visibleY1) into the buffer.
	PaintVisible(buf *buffer.Buffer, visibleY0, visibleY1 int)
}

// IsVisiblePainter reports whether a component implements VisiblePainter.
func IsVisiblePainter(c Component) bool {
	_, ok := c.(VisiblePainter)
	return ok
}

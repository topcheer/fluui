package component

import (
	"github.com/topcheer/fluui/internal/buffer"
)

// Text is a leaf component that renders a single line of styled text.
type Text struct {
	BaseComponent
	Content string
	Style   buffer.Style
}

// NewText creates a Text component with the given content and default style.
func NewText(content string) *Text {
	return &Text{
		Content: content,
		Style:   buffer.Style{},
	}
}

// Measure returns the desired size: width = string display width, height = 1.
func (t *Text) Measure(cs Constraints) Size {
	w := buffer.StringWidth(t.Content)
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	h := 1
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint draws the text into the buffer at the component's bounds origin.
func (t *Text) Paint(buf *buffer.Buffer) {
	buf.DrawText(t.bounds.X, t.bounds.Y, t.Content, t.Style)
}

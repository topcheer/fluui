package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── StreamingCursor: Animated Streaming Indicator ───
//
// StreamingCursor renders an animated cursor indicator that shows
// whether an AI response is actively streaming. Uses spinner-style
// animation characters cycling through a fixed sequence.
//
// Usage:
//
//	c := NewStreamingCursor()
//	c.SetActive(true)
//	c.SetFrame(2) // advance animation frame
//	c.Paint(buf)

// StreamingCursorStyle holds styling.
type StreamingCursorStyle struct {
	Active   buffer.Style
	Idle     buffer.Style
	BracketL buffer.Style
	BracketR buffer.Style
}

// DefaultStreamingCursorStyle returns defaults.
func DefaultStreamingCursorStyle() StreamingCursorStyle {
	return StreamingCursorStyle{
		Active:   buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Idle:     buffer.Style{Fg: buffer.RGB(100, 116, 139)},
		BracketL: buffer.Style{Fg: buffer.RGB(71, 85, 105)},
		BracketR: buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

var streamingCursorFrames = [...]rune{'▋', '▌', '▍', '▎', '▏', '▎', '▍', '▌'}

// StreamingCursor renders an animated streaming cursor.
type StreamingCursor struct {
	BaseComponent
	mu sync.Mutex

	active bool
	frame  int
	label  string
	style  StreamingCursorStyle
}

// NewStreamingCursor creates a StreamingCursor.
func NewStreamingCursor() *StreamingCursor {
	c := &StreamingCursor{label: "streaming", style: DefaultStreamingCursorStyle()}
	c.SetID(GenerateID("streamcur"))
	return c
}

// SetActive toggles streaming state.
func (c *StreamingCursor) SetActive(active bool) *StreamingCursor {
	c.mu.Lock()
	c.active = active
	c.mu.Unlock()
	return c
}

// SetFrame advances the animation frame.
func (c *StreamingCursor) SetFrame(n int) *StreamingCursor {
	c.mu.Lock()
	c.frame = n
	c.mu.Unlock()
	return c
}

// SetLabel sets the text label shown after the cursor.
func (c *StreamingCursor) SetLabel(s string) *StreamingCursor {
	c.mu.Lock()
	c.label = s
	c.mu.Unlock()
	return c
}

// SetStyle sets custom style.
func (c *StreamingCursor) SetStyle(s StreamingCursorStyle) *StreamingCursor {
	c.mu.Lock()
	c.style = s
	c.mu.Unlock()
	return c
}

// IsActive returns whether the cursor is in active (streaming) state.
func (c *StreamingCursor) IsActive() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.active
}

// Measure returns preferred size.
func (c *StreamingCursor) Measure(cs Constraints) Size {
	w := len(c.label) + 4
	if w < 10 { w = 10 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// Paint renders the streaming cursor.
func (c *StreamingCursor) Paint(buf *buffer.Buffer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	b := c.Bounds()
	x, y := b.X, b.Y

	bracketLStyle := c.style.BracketL
	bracketRStyle := c.style.BracketR

	col := x

	// [ prefix
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: '[', Fg: bracketLStyle.Fg, Bg: bracketLStyle.Bg, Flags: bracketLStyle.Flags, Width: 1})
		col++
	}

	if c.active {
		activeStyle := c.style.Active
		// Animated cursor block
		frameIdx := c.frame % len(streamingCursorFrames)
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: streamingCursorFrames[frameIdx], Fg: activeStyle.Fg, Bg: activeStyle.Bg, Flags: activeStyle.Flags, Width: 1})
			col++
		}
		// space
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: activeStyle.Fg, Bg: activeStyle.Bg, Flags: activeStyle.Flags, Width: 1})
			col++
		}
		// Label
		for _, r := range c.label {
			if col >= buf.Width { break }
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: activeStyle.Fg, Bg: activeStyle.Bg, Flags: activeStyle.Flags, Width: 1})
			col++
		}
	} else {
		idleStyle := c.style.Idle
		// Idle: show ● then label
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: '●', Fg: idleStyle.Fg, Bg: idleStyle.Bg, Flags: idleStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: idleStyle.Fg, Bg: idleStyle.Bg, Flags: idleStyle.Flags, Width: 1})
			col++
		}
		idleLabel := "idle"
		for _, r := range idleLabel {
			if col >= buf.Width { break }
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: idleStyle.Fg, Bg: idleStyle.Bg, Flags: idleStyle.Flags, Width: 1})
			col++
		}
	}

	// ] suffix
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ']', Fg: bracketRStyle.Fg, Bg: bracketRStyle.Bg, Flags: bracketRStyle.Flags, Width: 1})
	}
}

// Children returns nil.
func (c *StreamingCursor) Children() []Component { return nil }

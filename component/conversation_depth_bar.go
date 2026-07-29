package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ConversationDepthBar: Conversation Turn Depth Indicator ───
//
// ConversationDepthBar renders a horizontal bar showing how deep the
// current conversation is, with each turn represented as a segment.
// Helps visualize conversation length and approaching context limits.
//
// Usage:
//
//	c := NewConversationDepthBar()
//	c.SetDepth(8, 20) // 8 turns used out of 20 max
//	c.Paint(buf)

// ConversationDepthStyle holds styling.
type ConversationDepthStyle struct {
	Filled  buffer.Style
	Empty   buffer.Style
	Label   buffer.Style
	Counter buffer.Style
}

// DefaultConversationDepthStyle returns defaults.
func DefaultConversationDepthStyle() ConversationDepthStyle {
	return ConversationDepthStyle{
		Filled:  buffer.Style{Fg: buffer.RGB(59, 130, 246)},
		Empty:   buffer.Style{Fg: buffer.RGB(30, 41, 59)},
		Label:   buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Counter: buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
	}
}

// ConversationDepthBar renders a conversation depth indicator.
type ConversationDepthBar struct {
	BaseComponent
	mu sync.Mutex

	depth   int
	maxDepth int
	width   int
	style   ConversationDepthStyle
	// cached
	depthStr  string
	maxStr    string
	labelStr  string
	counterStr string
}

// NewConversationDepthBar creates a ConversationDepthBar.
func NewConversationDepthBar() *ConversationDepthBar {
	c := &ConversationDepthBar{width: 28, maxDepth: 20, style: DefaultConversationDepthStyle()}
	c.SetID(GenerateID("convdepth"))
	c.recomputeLocked()
	return c
}

// SetDepth sets current turn depth and maximum turns.
func (c *ConversationDepthBar) SetDepth(depth, maxDepth int) *ConversationDepthBar {
	c.mu.Lock()
	if depth < 0 { depth = 0 }
	if maxDepth < 1 { maxDepth = 1 }
	if depth > maxDepth { depth = maxDepth }
	c.depth = depth
	c.maxDepth = maxDepth
	c.recomputeLocked()
	c.mu.Unlock()
	return c
}

func (c *ConversationDepthBar) recomputeLocked() {
	c.depthStr = itoa(c.depth)
	c.maxStr = itoa(c.maxDepth)
	c.labelStr = "Turns "
	c.counterStr = c.depthStr + "/" + c.maxStr
}

// Depth returns current depth.
func (c *ConversationDepthBar) Depth() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.depth
}

// SetWidth sets the bar width.
func (c *ConversationDepthBar) SetWidth(w int) *ConversationDepthBar {
	c.mu.Lock()
	if w < 10 { w = 10 }
	c.width = w
	c.mu.Unlock()
	return c
}

// SetStyle sets custom style.
func (c *ConversationDepthBar) SetStyle(s ConversationDepthStyle) *ConversationDepthBar {
	c.mu.Lock()
	c.style = s
	c.mu.Unlock()
	return c
}

// Measure returns preferred size.
func (c *ConversationDepthBar) Measure(cs Constraints) Size {
	w := c.width + 16
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// Paint renders the conversation depth bar.
func (c *ConversationDepthBar) Paint(buf *buffer.Buffer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	b := c.Bounds()
	x, y := b.X, b.Y
	w := b.W
	if w < 10 { w = 40 }

	filledStyle := c.style.Filled
	emptyStyle := c.style.Empty
	labelStyle := c.style.Label
	counterStyle := c.style.Counter

	col := x

	// Label
	for _, r := range c.labelStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Bar segments
	barW := w - len(c.labelStr) - len(c.counterStr) - 2
	if barW < c.maxDepth { barW = c.maxDepth }
	segW := barW / c.maxDepth
	if segW < 1 { segW = 1 }

	for i := 0; i < c.maxDepth; i++ {
		var style_ buffer.Style
		var rune_ rune
		if i < c.depth {
			style_ = filledStyle
			rune_ = '▸'
		} else {
			style_ = emptyStyle
			rune_ = '·'
		}
		for j := 0; j < segW; j++ {
			if col >= buf.Width { break }
			buf.SetCell(col, y, buffer.Cell{Rune: rune_, Fg: style_.Fg, Bg: style_.Bg, Flags: style_.Flags, Width: 1})
			col++
		}
	}

	// Counter
	for _, r := range c.counterStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: counterStyle.Fg, Bg: counterStyle.Bg, Flags: counterStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (c *ConversationDepthBar) Children() []Component { return nil }

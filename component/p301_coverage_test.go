package component

import (
	"github.com/topcheer/fluui/internal/buffer"
	"testing"
)

// P301: Coverage for ChatComposer paintActive branches

// inputH < 1: bounds.H = 0 (but Paint returns early), so test H=1 (inputH=0→clamped to 1)
func TestP301_Paint_TinyHeight(t *testing.T) {
	c := NewChatComposer()
	c.SetText("hello")
	c.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	c.Paint(buf) // inputH clamped to 1
}

// contentW < 1: very narrow width
func TestP301_Paint_VeryNarrow(t *testing.T) {
	c := NewChatComposer()
	c.SetText("hello")
	c.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 4})
	buf := buffer.NewBuffer(1, 4)
	c.Paint(buf) // contentW clamped to 1
}

// Scrolled lines: more lines than inputH
func TestP301_Paint_ScrolledLines(t *testing.T) {
	c := NewChatComposer()
	c.SetText("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8")
	c.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 4}) // inputH=3, 8 lines → scroll
	buf := buffer.NewBuffer(60, 4)
	c.Paint(buf)
}

// Placeholder truncation: long placeholder in narrow width
func TestP301_Paint_PlaceholderTruncated(t *testing.T) {
	c := NewChatComposer()
	c.SetPlaceholder("This is a very long placeholder that should get truncated")
	c.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 4})
	buf := buffer.NewBuffer(10, 4)
	c.Paint(buf)
}

// Line truncation: long line content in narrow width
func TestP301_Paint_LineTruncated(t *testing.T) {
	c := NewChatComposer()
	c.SetText("very long line content that exceeds the available width")
	c.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 4})
	buf := buffer.NewBuffer(10, 4)
	c.Paint(buf)
}

// Token display: tokens present, fits on border
func TestP301_Paint_TokenDisplay(t *testing.T) {
	c := NewChatComposer()
	c.SetText("hi")
	c.SetTokenCount(500, 200)
	c.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 4})
	buf := buffer.NewBuffer(60, 4)
	c.Paint(buf)
}

// Token display too wide: token string longer than width-2
func TestP301_Paint_TokenTooWide(t *testing.T) {
	c := NewChatComposer()
	c.SetTokenCount(999999, 999999) // "↑1000.0k ↓1000.0k" is long
	c.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 4})
	buf := buffer.NewBuffer(5, 4)
	c.Paint(buf) // token string won't fit, should skip
}

// Slash mode with border painting
func TestP301_Paint_SlashModeWithTokens(t *testing.T) {
	c := NewChatComposer()
	c.SetText("/search query here")
	c.SetTokenCount(100, 50)
	c.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 4})
	buf := buffer.NewBuffer(60, 4)
	c.Paint(buf)
}

// Hint text truncated
func TestP301_Paint_HintTruncated(t *testing.T) {
	c := NewChatComposer()
	c.SetHint("This is a very long hint that should get truncated to fit the width")
	c.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 4})
	buf := buffer.NewBuffer(10, 4)
	c.Paint(buf)
}

// botY >= bounds.Y+bounds.H: no room for bottom border
func TestP301_Paint_NoBottomBorder(t *testing.T) {
	c := NewChatComposer()
	c.SetText("hello")
	c.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 2}) // inputH=1, botY=Y+1 < Y+2-1
	buf := buffer.NewBuffer(60, 2)
	c.Paint(buf)
}

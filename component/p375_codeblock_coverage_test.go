package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P375: Coverage for codeblock paintStreamingCursorLocked remaining paths

func TestP375_StreamingCursor_EmptyLines_NoTitle(t *testing.T) {
	// Empty lines without title → cursor at top-left, showTitle=false path
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetShowTitle(false)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf) // should not panic
}

func TestP375_StreamingCursor_WithTitleAndContent(t *testing.T) {
	// Content + title + streaming → hits showTitle y++ path at line 589
	cb := NewCodeBlock("go", "func main() {\n\tprintln(\"hi\")\n}")
	cb.SetStreaming(true)
	cb.SetShowTitle(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)
}

func TestP375_StreamingCursor_LongLine_XClamp(t *testing.T) {
	// Long line exceeds width → x clamped to bounds.X+bounds.W-1
	cb := NewCodeBlock("go", "very long line of code that exceeds the bounds width")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	cb.Paint(buf)
}

func TestP375_StreamingCursor_BelowVisibleArea(t *testing.T) {
	// Cursor Y position exceeds bounds → y >= bounds.Y+bounds.H return
	cb := NewCodeBlock("go", "short")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2})
	cb.ScrollTo(100) // scroll way past content
	buf := buffer.NewBuffer(40, 2)
	cb.Paint(buf) // should not panic, should early-return
}

func TestP375_StreamingCursor_SingleLine(t *testing.T) {
	// Single line, no scroll → basic cursor placement
	cb := NewCodeBlock("py", "print('hello')")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	cb.Paint(buf)
}

func TestP375_StreamingCursor_ManyLinesSmallHeight(t *testing.T) {
	// Many lines with small height → lastIdx clamped to len-1
	cb := NewCodeBlock("js", "a\nb\nc\nd\ne\nf\ng")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	cb.Paint(buf)
}

func TestP375_StreamingCursor_NilLines(t *testing.T) {
	// Force cb.lines to nil → hits the len(cb.lines)==0 guard.
	// Normally unreachable via public API (strings.Split always returns ≥1 element),
	// but tests the defensive guard path.
	cb := NewCodeBlock("go", "x")
	cb.SetStreaming(true)
	cb.mu.Lock()
	cb.lines = nil
	cb.mu.Unlock()
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)
}

func TestP375_StreamingCursor_NilLines_WithTitle(t *testing.T) {
	// Same but with title → hits showTitle y++ inside empty-lines block
	cb := NewCodeBlock("go", "x")
	cb.SetStreaming(true)
	cb.SetShowTitle(true)
	cb.mu.Lock()
	cb.lines = nil
	cb.mu.Unlock()
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)
}

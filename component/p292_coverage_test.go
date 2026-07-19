package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P292: Cover remaining viewport scrollbar branches
// drawVScrollBar 73.7% → 90%+, drawHScrollBar 73.7% → 90%+

// barH <= 0 early return
func TestP292_Viewport_VBarZeroHeight(t *testing.T) {
	vp := NewViewport(NewParagraph("line1\nline2\nline3"))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	vp.Paint(buffer.NewBuffer(10, 1))
}

// thumbH < 1 clamp (content very tall relative to bar)
func TestP292_Viewport_VBarTinyThumb(t *testing.T) {
	lines := ""
	for i := 0; i < 200; i++ {
		lines += "line\n"
	}
	vp := NewViewport(NewParagraph(lines))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 4})
	vp.ScrollDown(100)
	vp.Paint(buffer.NewBuffer(10, 4))
}

// thumbY+thumbH clamping (scrolled to bottom)
func TestP292_Viewport_VBarThumbClamp(t *testing.T) {
	vp := NewViewport(NewParagraph("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8"))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 4})
	// scroll way past end to trigger clamp
	vp.ScrollDown(100)
	vp.Paint(buffer.NewBuffer(20, 4))
}

// maxOff == 0 thumb = full bar
func TestP292_Viewport_VBarFullThumb(t *testing.T) {
	vp := NewViewport(NewParagraph("short"))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	vp.Paint(buffer.NewBuffer(20, 5))
}

// H scrollbar equivalents
// barW <= 0 early return
func TestP292_Viewport_HBarZeroWidth(t *testing.T) {
	vp := NewViewport(NewParagraph("some text"))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 5})
	vp.Paint(buffer.NewBuffer(1, 5))
}

// thumbW < 1 clamp (very wide content)
func TestP292_Viewport_HBarTinyThumb(t *testing.T) {
	wide := ""
	for i := 0; i < 200; i++ {
		wide += "x"
	}
	vp := NewViewport(NewParagraph(wide))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 3})
	vp.ScrollRight(100)
	vp.Paint(buffer.NewBuffer(8, 3))
}

// thumbX+thumbW clamping (scrolled to far right)
func TestP292_Viewport_HBarThumbClamp(t *testing.T) {
	vp := NewViewport(NewParagraph("very long text that overflows the viewport width for sure"))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 4})
	vp.ScrollRight(200)
	vp.Paint(buffer.NewBuffer(10, 4))
}

// maxOff == 0 for horizontal (content fits)
func TestP292_Viewport_HBarFullThumb(t *testing.T) {
	vp := NewViewport(NewParagraph("short"))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 4})
	vp.Paint(buffer.NewBuffer(50, 4))
}

// Both scrollbars visible simultaneously with offset
func TestP292_Viewport_BothBarsWithOffset(t *testing.T) {
	lines := ""
	for i := 0; i < 50; i++ {
		lines += "this is a wide line of text content\n"
	}
	vp := NewViewport(NewParagraph(lines))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 5})
	vp.ScrollDown(20)
	vp.ScrollRight(30)
	vp.Paint(buffer.NewBuffer(15, 5))
}

// Non-zero bounds offset for scrollbar position
func TestP292_Viewport_ScrollbarNonZeroOffset(t *testing.T) {
	vp := NewViewport(NewParagraph("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8"))
	vp.SetBounds(Rect{X: 5, Y: 3, W: 20, H: 5})
	vp.ScrollDown(3)
	vp.Paint(buffer.NewBuffer(30, 10))
}

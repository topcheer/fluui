package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P267: viewport SetBounds offset clamp (enlarge viewport) + barH/barW<=0 + thumb clamp

// overflowChild overflows both dimensions
type overflowChild struct{ BaseComponent }

func (c *overflowChild) Measure(cs Constraints) Size { return Size{W: 200, H: 50} }
func (c *overflowChild) Paint(buf *buffer.Buffer) {
	for y := 0; y < 50; y++ {
		buf.DrawText(0, y, "x", buffer.Style{})
	}
}

func TestViewport_SetBounds_ClampOffsetY_OnEnlarge_P267(t *testing.T) {
	child := &tallChild{} // H=50
	v := NewViewport(child)
	// Small viewport first → large maxOffsetY
	v.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	v.ScrollToBottom()
	// Now enlarge viewport → maxOffsetY shrinks → offsetY clamped
	v.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 30})
	buf := buffer.NewBuffer(20, 30)
	v.Paint(buf)
}

func TestViewport_SetBounds_ClampOffsetX_OnEnlarge_P267(t *testing.T) {
	child := &wideChild{} // W=200
	v := NewViewport(child)
	v.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 10})
	v.ScrollToRight()
	// Now enlarge viewport → maxOffsetX shrinks → offsetX clamped
	v.SetBounds(Rect{X: 0, Y: 0, W: 100, H: 10})
	buf := buffer.NewBuffer(100, 10)
	v.Paint(buf)
}

func TestViewport_DrawVScrollBar_BarHZero_P267(t *testing.T) {
	// H=1 → barH = 1 - 1(hbar) = 0 → early return
	child := &overflowChild{} // W=200, H=50 → both scrollbars visible
	v := NewViewport(child)
	v.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})
	buf := buffer.NewBuffer(5, 1)
	v.Paint(buf)
}

func TestViewport_DrawHScrollBar_BarWZero_P267(t *testing.T) {
	// W=1 → barW = 1 - 1(vbar) = 0 → early return
	child := &overflowChild{}
	v := NewViewport(child)
	v.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 5})
	buf := buffer.NewBuffer(1, 5)
	v.Paint(buf)
}

func TestViewport_DrawVScrollBar_ThumbClamp_P267(t *testing.T) {
	// Content with specific offset that causes thumbY+thumbH > bounds.Y+barH
	child := &overflowChild{} // H=50
	v := NewViewport(child)
	v.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	// Scroll to a position where the thumb would overflow
	v.ScrollToBottom()
	buf := buffer.NewBuffer(30, 10)
	v.Paint(buf)
}

func TestViewport_DrawHScrollBar_ThumbClamp_P267(t *testing.T) {
	child := &overflowChild{} // W=200
	v := NewViewport(child)
	v.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 10})
	v.ScrollToRight()
	buf := buffer.NewBuffer(10, 10)
	v.Paint(buf)
}

func TestBadge_Paint_IconFillsWidth_P267(t *testing.T) {
	// Icon exactly fills available width → space-after-icon branch
	b := NewBadge("", BadgeNeutral)
	b.SetIcon("AB")
	b.SetBounds(Rect{X: 0, Y: 0, W: 2, H: 1})
	buf := buffer.NewBuffer(2, 1)
	b.Paint(buf)
}

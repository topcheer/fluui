package component

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P290: component viewport scrollbar edge cases — barH<=0, maxOff==0, thumbClamp

func TestViewport_DrawVBar_Height1_P290(t *testing.T) {
	// H=1 → barH = H - hBarHeight() = 1 - 1 = 0 → early return
	child := NewText("overflow content that is wider than viewport width here")
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	vp.Paint(buf)
}

func TestViewport_DrawHBar_Width1_P290(t *testing.T) {
	// W=1 → barW = W - vBarWidth() = 1 - 1 = 0 → early return
	child := NewText("overflow\nline2\nline3\nline4\nline5")
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 3})
	buf := buffer.NewBuffer(1, 3)
	vp.Paint(buf)
}

func TestViewport_DrawVBar_ThumbClamp_P290(t *testing.T) {
	// Large content → thumbH = barH*barH/contentH, clamp to 1
	child := NewText(strings.Repeat("line\n", 200))
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 4})
	vp.ScrollDown(50)
	buf := buffer.NewBuffer(10, 4)
	vp.Paint(buf)
}

func TestViewport_DrawHBar_ThumbClamp_P290(t *testing.T) {
	// Wide content → thumbW = barW*barW/contentW, clamp to 1
	child := NewText(strings.Repeat("x", 300))
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	vp.ScrollRight(100)
	buf := buffer.NewBuffer(5, 3)
	vp.Paint(buf)
}

func TestViewport_DrawVBar_ThumbPositionClamp_P290(t *testing.T) {
	// Scroll to near end → thumbY+thumbH > bounds → clamp
	child := NewText(strings.Repeat("line\n", 100))
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	vp.ScrollDown(95)
	buf := buffer.NewBuffer(10, 5)
	vp.Paint(buf)
}

func TestViewport_DrawHBar_ThumbPositionClamp_P290(t *testing.T) {
	// Scroll to near right edge
	child := NewText(strings.Repeat("x", 200))
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	vp.ScrollRight(190)
	buf := buffer.NewBuffer(10, 3)
	vp.Paint(buf)
}

func TestViewport_DrawVBar_BothScroll_P290(t *testing.T) {
	// Both scrollbars active: content wider AND taller
	child := NewText(strings.Repeat("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n", 50))
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	vp.ScrollDown(10)
	vp.ScrollRight(10)
	buf := buffer.NewBuffer(5, 3)
	vp.Paint(buf)
}

package component

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P255: Viewport drawVScrollBar + drawHScrollBar edge cases

type tallChild struct{ BaseComponent }

func (c *tallChild) Measure(cs Constraints) Size { return Size{W: 100, H: 50} }
func (c *tallChild) Paint(buf *buffer.Buffer) {
	for i := 0; i < 50; i++ {
		buf.DrawText(0, i, "line"+string(rune('A'+i%26)), buffer.Style{})
	}
}
func (c *tallChild) HandleKey(k interface{}) bool { return false }

type wideChild struct{ BaseComponent }

func (c *wideChild) Measure(cs Constraints) Size { return Size{W: 200, H: 5} }
func (c *wideChild) Paint(buf *buffer.Buffer) {
	buf.DrawText(0, 0, strings.Repeat("X", 200), buffer.Style{})
}

func TestViewport_DrawVScrollBar_Overflow_P255(t *testing.T) {
	v := NewViewport(&tallChild{})
	v.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	v.ScrollDown(5)
	buf := buffer.NewBuffer(20, 10)
	v.Paint(buf)
}

func TestViewport_DrawVScrollBar_Fits_P255(t *testing.T) {
	child := NewViewport(&tallChild{})
	child.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	child.Paint(buf)
}

func TestViewport_DrawHScrollBar_Overflow_P255(t *testing.T) {
	v := NewViewport(&wideChild{})
	v.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	v.ScrollRight(5)
	buf := buffer.NewBuffer(10, 5)
	v.Paint(buf)
}

func TestViewport_DrawHScrollBar_Fits_P255(t *testing.T) {
	v := NewViewport(&wideChild{})
	v.SetBounds(Rect{X: 0, Y: 0, W: 200, H: 5})
	buf := buffer.NewBuffer(200, 5)
	v.Paint(buf)
}

func TestViewport_DrawVScrollBar_Tiny_P255(t *testing.T) {
	v := NewViewport(&tallChild{})
	v.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})
	buf := buffer.NewBuffer(5, 1)
	v.Paint(buf)
}

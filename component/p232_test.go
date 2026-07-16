package component

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// tallComponent returns a fixed Measure size for scrollbar testing
type tallComponent struct {
	BaseComponent
	w, h int
}

func (t *tallComponent) Measure(cs Constraints) Size { return Size{W: t.w, H: t.h} }
func (t *tallComponent) Paint(buf *buffer.Buffer)    {}

// P232: viewport drawVScrollBar + drawHScrollBar branch coverage

func TestViewport_VScrollBarOverflow_P232(t *testing.T) {
	v := NewViewport(&tallComponent{w: 80, h: 50})
	v.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	v.ScrollDown(3)
	buf := buffer.NewBuffer(20, 5)
	v.Paint(buf)
}

func TestViewport_VScrollBarFits_P232(t *testing.T) {
	v := NewViewport(&tallComponent{w: 10, h: 3})
	v.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	v.Paint(buf)
}

func TestViewport_HScrollBarOverflow_P232(t *testing.T) {
	v := NewViewport(&tallComponent{w: 200, h: 3})
	v.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	v.ScrollRight(50)
	buf := buffer.NewBuffer(10, 5)
	v.Paint(buf)
}

func TestViewport_BothScrollbarsOverflow_P232(t *testing.T) {
	v := NewViewport(&tallComponent{w: 200, h: 50})
	v.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	v.ScrollDown(10)
	v.ScrollRight(50)
	buf := buffer.NewBuffer(10, 5)
	v.Paint(buf)
}

func TestViewport_ScrollToEnd_P232(t *testing.T) {
	v := NewViewport(&tallComponent{w: 200, h: 50})
	v.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	v.ScrollToBottom()
	v.ScrollToRight()
	buf := buffer.NewBuffer(20, 10)
	v.Paint(buf)
}

func TestViewport_ParagraphOverflow_P232(t *testing.T) {
	v := NewViewport(NewParagraph(strings.Repeat("line\n", 30)))
	v.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	v.Paint(buf)
}

func TestLoadingIndicator_StartStop_P232(t *testing.T) {
	l := NewLoadingIndicator("Loading")
	l.Start()
	l.Start() // double-start
	l.Stop()
}

func TestLoadingIndicator_AnimationFrames_P232(t *testing.T) {
	l := NewLoadingIndicator("Loading")
	l.Start()
	buf := buffer.NewBuffer(10, 1)
	for i := 0; i < 5; i++ {
		l.Paint(buf)
	}
	l.Stop()
}

func TestViewport_VScrollBarMinHeight_P232(t *testing.T) {
	// Force barH = bounds.H - hBarHeight() = 1 - 1 = 0 → early return
	v := NewViewport(&tallComponent{w: 200, h: 50})
	v.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	v.Paint(buf)
}

func TestViewport_HScrollBarMinWidth_P232(t *testing.T) {
	// Force barW = bounds.W - vBarWidth() = 1 - 1 = 0 → early return  
	v := NewViewport(&tallComponent{w: 200, h: 50})
	v.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 10})
	buf := buffer.NewBuffer(1, 10)
	v.Paint(buf)
}

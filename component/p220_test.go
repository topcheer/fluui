package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P220: loading_indicator.Start double-call + viewport.drawVScrollBar edge cases

func TestLoadingIndicator_StartDoubleCall_P220(t *testing.T) {
	li := NewLoadingIndicator("test")
	li.Start()
	li.Start() // second call should return early (already running)
	li.Stop()
}

func TestLoadingIndicator_StartStopLoop_P220(t *testing.T) {
	li := NewLoadingIndicator("test")
	li.Start()
	li.Stop()
	li.Start() // restart after stop — exercises new ticker creation path
	li.Stop()
}

func TestViewport_DrawVScrollBarFits_P220(t *testing.T) {
	vp := NewViewport(NewFill(' ', buffer.Style{}))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	// Content smaller than viewport — maxOff=0, thumb = full bar
	vp.SetContent(NewFill('x', buffer.Style{}))
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

func TestViewport_DrawVScrollBarOverflow_P220(t *testing.T) {
	vp := NewViewport(NewFill(' ', buffer.Style{}))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	// Content taller than viewport
	content := NewFill('y', buffer.Style{})
	content.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 50})
	vp.SetContent(content)
	vp.ScrollToBottom()
	buf := buffer.NewBuffer(20, 5)
	vp.Paint(buf)
}

func TestViewport_DrawVScrollBarZeroHeight_P220(t *testing.T) {
	vp := NewViewport(NewFill(' ', buffer.Style{}))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	vp.SetContent(NewFill('z', buffer.Style{}))
	buf := buffer.NewBuffer(10, 1)
	vp.Paint(buf)
	// barH = H - hBarHeight, if <=0 should return early
}
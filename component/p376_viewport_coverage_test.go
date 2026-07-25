package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P376: Viewport scrollbar coverage — defensive branches and edge cases

func TestP376_VScrollBar_ContentFits(t *testing.T) {
	// Force showVBar=true but content fits (maxOff==0 path)
	// by directly calling drawVScrollBar with manipulated state
	v := NewViewport(NewText("short"))
	v.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	v.Measure(Constraints{MaxWidth: 20, MaxHeight: 10})
	// After Measure, contentH=1 (one line), bounds.H=10 → showVBar=false
	// Force the scrollbar to draw by setting showVBar and calling directly
	v.mu.Lock()
	v.showVBar = true
	v.contentH = 10 // same as bounds.H → maxOff=0
	v.mu.Unlock()
	buf := buffer.NewBuffer(20, 10)
	v.Paint(buf)
}

func TestP376_VScrollBar_BarHZero(t *testing.T) {
	// Viewport with height=1 and both scrollbars → barH = bounds.H - hBarHeight() = 0
	v := NewViewport(NewText("x"))
	v.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})
	v.mu.Lock()
	v.showVBar = true
	v.showHBar = true
	v.contentH = 10
	v.contentW = 10
	v.mu.Unlock()
	buf := buffer.NewBuffer(5, 1)
	v.Paint(buf) // should not panic, barH=0 early return
}

func TestP376_VScrollBar_ThumbClamp(t *testing.T) {
	// Large content with scroll near bottom → thumb clamped to bottom
	v := NewViewport(NewText("line"))
	v.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	v.mu.Lock()
	v.showVBar = true
	v.contentH = 100
	v.offsetY = 95 // scrolled near bottom
	v.mu.Unlock()
	buf := buffer.NewBuffer(20, 5)
	v.Paint(buf) // thumb should be clamped
}

func TestP376_HScrollBar_ContentFits(t *testing.T) {
	// Force showHBar=true but content fits (maxOff==0 path)
	v := NewViewport(NewText("short"))
	v.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	v.mu.Lock()
	v.showHBar = true
	v.contentW = 20 // same as bounds.W → maxOff=0
	v.mu.Unlock()
	buf := buffer.NewBuffer(20, 10)
	v.Paint(buf)
}

func TestP376_HScrollBar_BarWZero(t *testing.T) {
	// Width=1 with both scrollbars → barW = bounds.W - vBarWidth() = 0
	v := NewViewport(NewText("x"))
	v.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 5})
	v.mu.Lock()
	v.showVBar = true
	v.showHBar = true
	v.contentH = 10
	v.contentW = 10
	v.mu.Unlock()
	buf := buffer.NewBuffer(1, 5)
	v.Paint(buf) // should not panic, barW=0 early return
}

func TestP376_HScrollBar_ThumbClamp(t *testing.T) {
	// Wide content scrolled right → thumb clamped to right edge
	v := NewViewport(NewText("wide"))
	v.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	v.mu.Lock()
	v.showHBar = true
	v.contentW = 100
	v.offsetX = 95 // scrolled far right
	v.mu.Unlock()
	buf := buffer.NewBuffer(10, 5)
	v.Paint(buf) // thumb should be clamped
}

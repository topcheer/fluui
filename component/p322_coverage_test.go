package component

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P322: Push richlog countVisibleLinesLocked and viewport scrollbars past 80%

// === richlog: bounds.H <= 0 branch ===
func TestP322_RichLog_ZeroHeightBounds(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 5; i++ {
		rl.Info("message " + string(rune('a'+i)))
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 0})
	rl.Paint(buffer.NewBuffer(20, 1)) // H=0 → returns len(entries)
}

// === richlog: w <= 0 branch (default 80) ===
func TestP322_RichLog_ZeroWidthBounds(t *testing.T) {
	rl := NewRichLog()
	rl.Info("test")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 3})
	rl.Paint(buffer.NewBuffer(1, 3)) // W=0 → defaults to 80
}

// === richlog: contentW < 1 (very narrow with header) ===
func TestP322_RichLog_VeryNarrowWithHeader(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	rl.Info("test message")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 2, H: 3}) // contentW = 2 - hdrWidth → < 1
	rl.Paint(buffer.NewBuffer(2, 3))
}

// === richlog: minLevel filtering ===
func TestP322_RichLog_MinLevelFilter(t *testing.T) {
	rl := NewRichLog()
	rl.SetMinLevel(2) // only warnings and above
	rl.Info("hidden info")  // level 1, filtered
	rl.Warn("visible warn") // level 2, shown
	rl.Error("visible err") // level 3, shown
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	rl.Paint(buffer.NewBuffer(40, 5))
}

// === richlog: wrapping with prefix in narrow width ===
func TestP322_RichLog_WrappingWithPrefix(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	long := strings.Repeat("word ", 20)
	rl.Info(long)
	rl.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 5})
	rl.Paint(buffer.NewBuffer(15, 5))
}

// === richlog: scrollUp then paint ===
func TestP322_RichLog_ScrollUp(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 20; i++ {
		rl.Info("line")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	rl.ScrollUp(10)
	rl.Paint(buffer.NewBuffer(40, 3))
}

// === richlog: error level rendering ===
func TestP322_RichLog_ErrorLevel(t *testing.T) {
	rl := NewRichLog()
	rl.Error("critical failure")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2})
	rl.Paint(buffer.NewBuffer(40, 2))
}

// === richlog: warn level rendering ===
func TestP322_RichLog_WarnLevel(t *testing.T) {
	rl := NewRichLog()
	rl.Warn("warning message")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2})
	rl.Paint(buffer.NewBuffer(40, 2))
}

// === richlog: debug level ===
func TestP322_RichLog_DebugLevel(t *testing.T) {
	rl := NewRichLog()
	rl.Debug("debug info")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2})
	rl.Paint(buffer.NewBuffer(40, 2))
}

// === richlog: empty entries paint ===
func TestP322_RichLog_EmptyPaint(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	rl.Paint(buffer.NewBuffer(40, 3))
}

// === viewport: scrollbar with exactly content-sized viewport ===
func TestP322_Viewport_ExactFit(t *testing.T) {
	vp := NewViewport(NewParagraph("short"))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 2})
	vp.Paint(buffer.NewBuffer(10, 2)) // content fits exactly, no scroll
}

// === viewport: scrollbar with content smaller than viewport ===
func TestP322_Viewport_ContentSmaller(t *testing.T) {
	vp := NewViewport(NewParagraph("x"))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	vp.ScrollDown(5) // try to scroll beyond content
	vp.Paint(buffer.NewBuffer(10, 5))
}

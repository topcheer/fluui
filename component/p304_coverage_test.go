package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P304: Cover remaining sub-80% functions with actionable branches

// === Badge.Measure 73.3% → 85%+ ===
// Missing: w<1 after MaxWidth=0, h<1 after MaxHeight=0, w<2 (small badge)

func TestP304_Badge_MeasureZeroMaxWidth(t *testing.T) {
	b := NewBadgeWithSize("X", BadgeInfo, BadgeSizeSmall)
	// BadgeSizeSmall: padding=0, contentWidth=1 → w<2 → w=2
	// MaxWidth=0 → HasWidth()=false → no clamp
	s := b.Measure(Constraints{})
	if s.W != 2 {
		t.Errorf("w=%d, want 2 (clamped from <2)", s.W)
	}
	if s.H != 1 {
		t.Errorf("h=%d, want 1", s.H)
	}
}

func TestP304_Badge_MeasureMaxWidthZero(t *testing.T) {
	b := NewBadge("hello", BadgeSuccess)
	// contentWidth = 5 + 2*padding(Normal=1) = 7
	// MaxWidth=0 → HasWidth=false → w stays 7
	// w>=2 → no clamp, w>=1 → no clamp
	s := b.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W < 2 {
		t.Errorf("w=%d, should be >= 2", s.W)
	}
}

func TestP304_Badge_MeasureHeightClamp(t *testing.T) {
	b := NewBadge("AB", BadgeWarning)
	// MaxHeight=0 → HasHeight=false → h stays 1
	s := b.Measure(Constraints{MaxHeight: 0})
	if s.H != 1 {
		t.Errorf("h=%d, want 1", s.H)
	}
}

// === CodeBlock.paintStreamingCursorLocked 74.2% → 85%+ ===
// Missing: empty lines with title, lastIdx clamping, y out of bounds

func TestP304_CodeBlock_StreamingCursor_EmptyWithTitle(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetTitle("test.go")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	cb.Paint(buffer.NewBuffer(20, 5))
}

func TestP304_CodeBlock_StreamingCursor_LongContent(t *testing.T) {
	cb := NewCodeBlock("go", "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8")
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	cb.Paint(buffer.NewBuffer(30, 3))
}

func TestP304_CodeBlock_StreamingCursor_PlainFallback(t *testing.T) {
	cb := NewCodeBlock("unknown_lang", "code line here")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	cb.Paint(buffer.NewBuffer(30, 3))
}

func TestP304_CodeBlock_StreamingCursor_XClamp(t *testing.T) {
	// Very long line that exceeds bounds width
	cb := NewCodeBlock("go", "this is a very long line that exceeds the viewport width")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	cb.Paint(buffer.NewBuffer(5, 3))
}

func TestP304_CodeBlock_StreamingCursor_YOutOfBounds(t *testing.T) {
	cb := NewCodeBlock("go", "line1\nline2\nline3")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	cb.Paint(buffer.NewBuffer(20, 1)) // y will be out of bounds
}

// === RichLog.countVisibleLinesLocked 78.6% → 85%+ ===
// Missing: zero-height viewport, showTime+showLevels combined

func TestP304_RichLog_CountVisible_ZeroHeight(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	for i := 0; i < 10; i++ {
		rl.Info("test message")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 0})
	rl.Paint(buffer.NewBuffer(20, 1)) // H=0 shouldn't panic
}

func TestP304_RichLog_CountVisible_NarrowWidth(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	rl.Info("very long message that needs wrapping across multiple lines")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 2})
	rl.Paint(buffer.NewBuffer(10, 2))
}

func TestP304_RichLog_CountVisible_Scrolled(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	for i := 0; i < 30; i++ {
		rl.Info("message")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 2})
	rl.ScrollUp(15)
	rl.Paint(buffer.NewBuffer(20, 2))
}

// === ThemeStudio.setCursorLocked 75% → 85%+ ===
// Missing: empty slots, wrap-around negative

func TestP304_ThemeStudio_EmptySlots(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.mu.Lock()
	ts.slots = nil // force empty
	ts.setCursorLocked(5)
	if ts.cursor != 0 {
		t.Errorf("cursor=%d, want 0 for empty slots", ts.cursor)
	}
	ts.mu.Unlock()
}

// === Viewport.drawVScrollBar 73.7% → 85%+ ===
// Missing: barH<=0 early return (only when showVBar=false)
// and thumbH<1 clamp, thumbY+thumbH clamping

func TestP304_Viewport_VBarTinyThumb(t *testing.T) {
	// Content >> viewport → thumbH = barH²/contentH, which can be <1
	lines := ""
	for i := 0; i < 500; i++ {
		lines += "x\n"
	}
	vp := NewViewport(NewParagraph(lines))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	vp.ScrollDown(200)
	vp.Paint(buffer.NewBuffer(10, 3))
}

func TestP304_Viewport_HBarTinyThumb(t *testing.T) {
	wide := ""
	for i := 0; i < 500; i++ {
		wide += "x"
	}
	vp := NewViewport(NewParagraph(wide))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	vp.ScrollRight(200)
	vp.Paint(buffer.NewBuffer(5, 3))
}

func TestP304_Viewport_BothBarsThumbClamp(t *testing.T) {
	lines := ""
	for i := 0; i < 200; i++ {
		lines += "very wide line content that overflows\n"
	}
	vp := NewViewport(NewParagraph(lines))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 3})
	vp.ScrollDown(100)
	vp.ScrollRight(100)
	vp.Paint(buffer.NewBuffer(8, 3))
}

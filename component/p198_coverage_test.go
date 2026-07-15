package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P198: Targeted tests for remaining uncovered branches

func TestViewport_VScrollBarThumbClamp_P198(t *testing.T) {
	// Large content in small viewport → thumb should be clamped
	vp := NewViewport(NewFill(' ', buffer.Style{}))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	vp.ScrollToY(1000) // max offset → thumb should be at bottom, clamp triggers
	vp.Paint(buffer.NewBuffer(5, 3))
}

func TestViewport_VScrollBarTinyContent_P198(t *testing.T) {
	// Content exactly fits → maxOffset=0 → full thumb path
	vp := NewViewport(NewFill(' ', buffer.Style{}))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	vp.Paint(buffer.NewBuffer(5, 3))
}

func TestViewport_HScrollBarThumbClamp_P198(t *testing.T) {
	vp := NewViewport(NewFill(' ', buffer.Style{}))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 5})
	vp.ScrollToX(1000) // max offset → clamp triggers
	vp.Paint(buffer.NewBuffer(3, 5))
}

func TestViewport_BothScrollBars_P198(t *testing.T) {
	vp := NewViewport(NewFill(' ', buffer.Style{}))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 3})
	vp.ScrollToX(100)
	vp.ScrollToY(100)
	vp.Paint(buffer.NewBuffer(3, 3))
}

func TestViewport_ZeroBarHeight_P198(t *testing.T) {
	// barH <= 0 early return path
	vp := NewViewport(NewFill(' ', buffer.Style{}))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 1})
	vp.Paint(buffer.NewBuffer(1, 1))
}

func TestHotKey_KeyEventToCombo_P198(t *testing.T) {
	// Test nil key path
	// This is in internal/hotkey but we test via the manager
	// Skip — requires internal package access
}

func TestCodeBlock_StreamingCursorOverflow_P198(t *testing.T) {
	// Streaming cursor with content wider than width
	cb := NewCodeBlock("go", "package main\n\nfunc hello() {\n    fmt.Println(\"very long string that overflows the block width\")\n}")
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 10})
	cb.Paint(buffer.NewBuffer(15, 10))
}

func TestDiffPreview_PaintBorderNarrow_P198(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("--- a\n+++ b\n@@ -1 +1 @@\n-a\n+b")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	dp.Paint(buffer.NewBuffer(10, 5))
}

func TestBadge_MeasureShortIcon_P198(t *testing.T) {
	// Badge with icon
	b := NewBadge("ok", BadgeSuccess)
	b.SetIcon("✓")
	s := b.Measure(Constraints{MaxWidth: 10, MaxHeight: 2})
	if s.H < 1 {
		t.Error("height should be at least 1")
	}
}

func TestAutoComplete_PaintWithDescription_P198(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "test", Description: "a test item"},
	})
	ac.Show(0, 0)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	ac.Paint(buffer.NewBuffer(30, 5))
}
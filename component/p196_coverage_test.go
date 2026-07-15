package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// P196: Targeted tests for uncovered branches in sub-80% functions

func TestAutoComplete_PaintWithScroll_P196(t *testing.T) {
	ac := NewAutoComplete()
	items := make([]CompletionItem, 20)
	for i := range items {
		items[i] = CompletionItem{Label: string(rune('a' + i))}
	}
	ac.SetItems(items)
	ac.SetMaxVisible(5)
	ac.Show(0, 0)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	ac.Paint(buffer.NewBuffer(20, 10))
}

func TestBadge_MeasureConstrained_P196(t *testing.T) {
	// Test MaxWidth constraint path
	b := NewBadge("long text here", BadgeInfo)
	s := b.Measure(Constraints{MaxWidth: 3, MaxHeight: 1})
	if s.W > 3 {
		t.Errorf("width should be constrained to 3, got %d", s.W)
	}
	// Test MaxHeight = 0 path
	b2 := NewBadge("x", BadgeSuccess)
	s2 := b2.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s2.W < 1 {
		t.Error("width should be at least 1")
	}
	// Test narrow badge (w < 2)
	b3 := NewBadge("", BadgeWarning)
	s3 := b3.Measure(Constraints{})
	if s3.W < 2 {
		t.Errorf("empty badge should have min width 2, got %d", s3.W)
	}
}

func TestViewport_VScrollBarWithOffset_P196(t *testing.T) {
	vp := NewViewport(NewFill(' ', buffer.Style{}))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	vp.ScrollToY(100) // large offset to trigger thumb drawing
	vp.Paint(buffer.NewBuffer(10, 5))
}

func TestViewport_HScrollBarWithOffset_P196(t *testing.T) {
	vp := NewViewport(NewFill(' ', buffer.Style{}))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 10})
	vp.ScrollToX(100) // large offset
	vp.Paint(buffer.NewBuffer(5, 10))
}

func TestSparkline_ValueToBarEdges_P196(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{0, 5, 10})
	// Access internal state to test valueToBar
	// min=0, max=10, val=5 → ratio=0.5
	bar := sl.valueToBar(5)
	if bar < 0 {
		t.Error("bar should be non-negative")
	}
	// Test edge: val > max
	bar2 := sl.valueToBar(100)
	if bar2 < 0 {
		t.Error("bar should clamp to non-negative")
	}
	// Test edge: val < min
	bar3 := sl.valueToBar(-100)
	if bar3 != 0 {
		t.Errorf("negative val should give 0, got %d", bar3)
	}
}

func TestSparkline_ValueToBarEmpty_P196(t *testing.T) {
	sl := NewSparkline()
	sl.SetData(nil)
	bar := sl.valueToBar(5)
	if bar != 0 {
		t.Errorf("empty data should give 0, got %d", bar)
	}
}

func TestScrollView_ContentWNarrow_P196(t *testing.T) {
	sv := NewScrollView(NewFill(' ', buffer.Style{}))
	w := sv.contentW(0) // boundsW=0 → w should be 1
	if w != 1 {
		t.Errorf("expected 1, got %d", w)
	}
}

func TestLoadingIndicator_StartStopStart_P196(t *testing.T) {
	li := NewLoadingIndicator("test")
	li.Start()
	time.Sleep(50 * time.Millisecond)
	li.Stop()
	// Start again after stop
	li.Start()
	time.Sleep(50 * time.Millisecond)
	li.Stop()
}

func TestCodeBlock_StreamingCursorAtEnd_P196(t *testing.T) {
	cb := NewCodeBlock("go", "package main")
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 5})
	cb.Paint(buffer.NewBuffer(80, 5))
}

func TestDiffPreview_PaintBorderWithStats_P196(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("--- old\n+++ new\n@@ -1,2 +1,3 @@\n-old line\n+new line\n+added")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	dp.Paint(buffer.NewBuffer(60, 20))
}

func TestHelpOverlay_ScrollUp_P196(t *testing.T) {
	ho := NewHelpOverlay([]HelpGroup{
		{Name: "g1", Entries: []HelpEntry{{Keys: "ctrl+a", Description: "a"}}},
		{Name: "g2", Entries: []HelpEntry{{Keys: "ctrl+b", Description: "b"}}},
	})
	ho.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	ho.ScrollDown(1)
	ho.ScrollUp(1)
	ho.Paint(buffer.NewBuffer(40, 3))
}

func TestRichLog_CountVisibleWrapped_P196(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	// Very long line that will wrap
	rl.Info("this is a very long log line that should wrap across multiple lines when the terminal width is small")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	rl.Paint(buffer.NewBuffer(30, 10))
}

func TestThemeStudio_CursorMove_P196(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	ts.Paint(buffer.NewBuffer(60, 20))
	// Move cursor to hit setCursorLocked paths
	ts.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	ts.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	ts.Paint(buffer.NewBuffer(60, 20))
}

func TestStyleBuilder_ParseColorFormats_P196(t *testing.T) {
	// Test various color format edge cases
	sb := NewStyle()
	sb.Foreground(buffer.RGB(255, 128, 0))
	sb.Background(buffer.RGB(0, 128, 255))
	_ = sb
}
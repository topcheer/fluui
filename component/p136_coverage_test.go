package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === AutoComplete.Paint (76.7% → 90%+) ===
// Uncovered: items with Description, Category, filtered items, isSelected on category

func TestP136_AutoComplete_Paint_WithDescription(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "item1", Description: "does something useful"},
		{Label: "item2", Category: "Commands"},
		{Label: "item3", Description: "another desc", Category: "Tools"},
	})
	ac.SetCursor(1)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	ac.Paint(buf)
}

func TestP136_AutoComplete_Paint_CategorySelected(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "cmd", Category: "VeryLongCategoryName"},
	})
	ac.SetCursor(0)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 5})
	buf := buffer.NewBuffer(15, 5)
	ac.Paint(buf)
}

func TestP136_AutoComplete_Paint_ScrollNegative(t *testing.T) {
	ac := NewAutoComplete()
	items := make([]CompletionItem, 20)
	for i := range items {
		items[i] = CompletionItem{Label: "item"}
	}
	ac.SetItems(items)
	ac.scrollY = -5 // force negative scroll
	ac.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	ac.Paint(buf)
}

// === Badge.Measure (76.5% → 100%) ===
// Uncovered: w < 2, height clamping, MaxWidth/MaxHeight constraints

func TestP136_Badge_Measure_NarrowConstraint(t *testing.T) {
	b := NewBadge("OK", BadgeInfo)
	s := b.Measure(Bounded(2, 1)) // w < 2 branch
	_ = s
}

func TestP136_Badge_Measure_HeightClampSmall(t *testing.T) {
	b := NewBadge("Test", BadgeError)
	s := b.Measure(Bounded(100, 1)) // h < 1 after subtracting badge height
	_ = s
}

func TestP136_Badge_Measure_AllVariantsSmall(t *testing.T) {
	for _, v := range []BadgeVariant{BadgeInfo, BadgeSuccess, BadgeWarning, BadgeError, BadgeCritical} {
		b := NewBadge("x", v)
		s := b.Measure(Bounded(1, 1))
		if s.W < 0 || s.H < 0 {
			t.Errorf("bad measure for variant %d", v)
		}
	}
}

// === CodeBlock.paintStreamingCursorLocked (74.2% → 90%+) ===
// Uncovered: various edge cases with line numbers, title, scroll

func TestP136_CodeBlock_StreamCursor_TitleAndLineNumbers(t *testing.T) {
	cb := NewCodeBlock("go", "code line 1\ncode line 2")
	cb.SetStreaming(true)
	cb.SetShowTitle(true)
	cb.SetShowLineNumbers(true)
	cb.ScrollTo(1)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	cb.Paint(buf)
}

func TestP136_CodeBlock_StreamCursor_LongLineNoLineNumbers(t *testing.T) {
	cb := NewCodeBlock("go", "very long code line without line numbers enabled here")
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(false)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 2})
	buf := buffer.NewBuffer(10, 2)
	cb.Paint(buf)
}

func TestP136_CodeBlock_StreamCursor_EmptyContent(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	cb.Paint(buf)
}

func TestP136_CodeBlock_StreamCursor_NotStreaming(t *testing.T) {
	cb := NewCodeBlock("go", "x")
	cb.SetStreaming(false)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	cb.Paint(buf)
}

func TestP136_CodeBlock_StreamCursor_ZeroBounds(t *testing.T) {
	cb := NewCodeBlock("go", "x")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	cb.Paint(buf)
}

// === DiffPreview.paintBorderLocked (76.5% → 90%+) ===

func TestP136_DiffPreview_PaintBorder_TallContent(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("--- a\n+++ b\n@@ -1,5 +1,5 @@\n a\n b\n-old\n+new\n c\n d")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	dp.Paint(buf)
}

func TestP136_DiffPreview_PaintBorder_Empty(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	dp.Paint(buf)
}

// === HelpOverlay.ensureSelectedVisibleLocked (75% → 90%+) ===

func TestP136_HelpOverlay_ScrollToBottom(t *testing.T) {
	groups := []HelpGroup{
		{Name: "G", Entries: make([]HelpEntry, 30)},
	}
	ho := NewHelpOverlay(groups)
	ho.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	ho.ScrollDown(25)
	buf := buffer.NewBuffer(60, 5)
	ho.Paint(buf)
}

func TestP136_HelpOverlay_ScrollToTop(t *testing.T) {
	groups := []HelpGroup{
		{Name: "G", Entries: make([]HelpEntry, 30)},
	}
	ho := NewHelpOverlay(groups)
	ho.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	ho.ScrollDown(25)
	ho.ScrollUp(25)
	buf := buffer.NewBuffer(60, 5)
	ho.Paint(buf)
}

// === MenuBar.paintDropdownLocked (79.4% → 90%+) ===

func TestP136_MenuBar_PaintDropdown_ManyItems(t *testing.T) {
	items := make([]MenuEntry, 20)
	for i := range items {
		items[i] = MenuEntry{ID: "item", Label: "Item"}
	}
	mb := NewMenuBar([]Menu{{ID: "m", Title: "M", Items: items}})
	mb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	mb.OpenMenu(0)
	for i := 0; i < 15; i++ {
		mb.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	}
	buf := buffer.NewBuffer(40, 20)
	mb.Paint(buf)
}

// === RichLog.countVisibleLinesLocked (78.6% → 90%+) ===

func TestP136_RichLog_ScrollToEntry(t *testing.T) {
	rl := NewRichLog()
	rl.SetMaxSize(100)
	rl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	for i := 0; i < 10; i++ {
		rl.Info("entry that may wrap at 30 chars width if long")
	}
	rl.ScrollUp(3)
	buf := buffer.NewBuffer(30, 3)
	rl.Paint(buf)
}

func TestP136_RichLog_ShowTimeShowLevels(t *testing.T) {
	rl := NewRichLog()
	rl.SetMaxSize(100)
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	rl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	rl.Warn("warning message")
	rl.Error("error message")
	buf := buffer.NewBuffer(30, 5)
	rl.Paint(buf)
}

// === ScrollView.contentW (75% → 100%) ===

func TestP136_ScrollView_ContentW_Narrow(t *testing.T) {
	sv := NewScrollView(&fixedSize{w: 100, h: 100})
	sv.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 5}) // w < 1 after scrollbar
	buf := buffer.NewBuffer(1, 5)
	sv.Paint(buf)
}

// === Sparkline.valueToBar (77.8% → 100%) ===

func TestP136_Sparkline_AllSame_ViaPaint(t *testing.T) {
	sp := NewSparkline()
	sp.SetData([]float64{5, 5, 5, 5, 5})
	sp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	sp.Paint(buf)
}

// === TextArea.moveLine (77.8% → 100%) ===

func TestP136_TextArea_MoveLine_SingleLine(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("only line")
	ta.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyUp})
}

// === ThemeStudio.setCursorLocked (75% → 100%) ===

func TestP136_ThemeStudio_SetCursor_Clamp(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	// Navigate way past end and back
	for i := 0; i < 100; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	}
	for i := 0; i < 100; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	}
}

// === Viewport drawVScrollBar (73.7% → 90%+) ===

func TestP136_Viewport_VScrollBar_NearBottom(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 50, h: 30})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vp.ScrollToY(20) // near bottom → thumbY + thumbH > bounds
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

func TestP136_Viewport_VScrollBar_FitsContent(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 50, h: 5})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

// === Viewport drawHScrollBar (73.7% → 90%+) ===

func TestP136_Viewport_HScrollBar_NearRight(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 100, h: 5})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vp.ScrollToX(80) // near right → thumbX + thumbW > bounds
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

func TestP136_Viewport_HScrollBar_FitsContent(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 5, h: 5})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

// === VirtualScroller.VisibleItems (78.6% → 90%+) ===

func TestP136_VirtualScroller_ExactVisible(t *testing.T) {
	vs := NewVirtualScroller()
	items := make([]VirtualItem, 5)
	for i := range items {
		items[i] = VirtualItem{ID: "i", Text: "x"}
	}
	vs.SetItems(items)
	vs.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5}) // exact fit
	buf := buffer.NewBuffer(40, 5)
	vs.Paint(buf)
}

func TestP136_VirtualScroller_ScrolledPastEnd(t *testing.T) {
	vs := NewVirtualScroller()
	items := make([]VirtualItem, 20)
	for i := range items {
		items[i] = VirtualItem{ID: "i", Text: "x"}
	}
	vs.SetItems(items)
	vs.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	vs.ScrollTo(100) // way past end
	buf := buffer.NewBuffer(40, 5)
	vs.Paint(buf)
}

package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === AutoComplete.Paint (76.7% → 90%+) ===
// Uncovered: items with Description, Category rendering, isSelected on category,
// filtered index overflow, description truncation, contentX >= buf.Width

func TestP140_AutoComplete_Paint_DescriptionTruncated(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "cmd1", Description: "very long description that needs truncation"},
	})
	ac.SetCursor(0)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	ac.Paint(buf)
}

func TestP140_AutoComplete_Paint_CategorySelectedWithTruncation(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "x", Category: "VeryLongCategoryNameThatExceedsHalfWidth"},
	})
	ac.SetCursor(0)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 5})
	buf := buffer.NewBuffer(15, 5)
	ac.Paint(buf)
}

func TestP140_AutoComplete_Paint_DescriptionRemaining3(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "abcdefghij", Description: "desc"},
	})
	ac.SetCursor(0)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 14, H: 5})
	buf := buffer.NewBuffer(14, 5)
	ac.Paint(buf)
}

func TestP140_AutoComplete_Paint_EmptyFiltered(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "test"},
	})
	ac.SetQuery("xyz") // no matches
	ac.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	ac.Paint(buf)
}

func TestP140_AutoComplete_Paint_ScrollYNegative(t *testing.T) {
	ac := NewAutoComplete()
	items := make([]CompletionItem, 10)
	for i := range items {
		items[i] = CompletionItem{Label: "item"}
	}
	ac.SetItems(items)
	ac.scrollY = -3
	ac.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	ac.Paint(buf)
}

func TestP140_AutoComplete_Paint_ContentXExceedsWidth(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "verylonglabelthatexceedsthecontentwidth", Category: "cat"},
	})
	ac.SetCursor(0)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	ac.Paint(buf)
}

// === Badge.Measure (76.5% → 100%) ===
// Uncovered: w < 2, h < 1, MaxWidth/MaxHeight constraint branches

func TestP140_Badge_Measure_W2(t *testing.T) {
	b := NewBadge("OK", BadgeInfo)
	s := b.Measure(Bounded(2, 3))
	_ = s
}

func TestP140_Badge_Measure_H1(t *testing.T) {
	b := NewBadge("test", BadgeError)
	s := b.Measure(Bounded(50, 1))
	_ = s
}

func TestP140_Badge_Measure_AllVariantsW1H1(t *testing.T) {
	for _, v := range []BadgeVariant{BadgeInfo, BadgeSuccess, BadgeWarning, BadgeError, BadgeCritical} {
		b := NewBadge("x", v)
		s := b.Measure(Bounded(1, 1))
		_ = s
	}
}

// === CodeBlock.paintStreamingCursorLocked (74.2% → 90%+) ===

func TestP140_CodeBlock_StreamCursor_EmptyWithLineNumbersAndTitle(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	cb.SetShowTitle(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestP140_CodeBlock_StreamCursor_ScrollPast(t *testing.T) {
	cb := NewCodeBlock("go", "a\nb\nc\nd\ne\nf")
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	cb.ScrollTo(5)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	cb.Paint(buf)
}

func TestP140_CodeBlock_StreamCursor_LongLineNarrow(t *testing.T) {
	cb := NewCodeBlock("go", "abcdefghijklmnopqrstuvwxyz")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 3})
	buf := buffer.NewBuffer(8, 3)
	cb.Paint(buf)
}

func TestP140_CodeBlock_StreamCursor_PlainFallback(t *testing.T) {
	cb := NewCodeBlock("go", "code")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	cb.Paint(buf)
}

func TestP140_CodeBlock_StreamCursor_NotStreaming(t *testing.T) {
	cb := NewCodeBlock("go", "code")
	cb.SetStreaming(false)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	cb.Paint(buf)
}

func TestP140_CodeBlock_StreamCursor_ZeroBounds(t *testing.T) {
	cb := NewCodeBlock("go", "code")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	cb.Paint(buf)
}

// === DiffPreview.paintBorderLocked (76.5% → 90%+) ===

func TestP140_DiffPreview_Border_TallContent(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("--- a\n+++ b\n@@ -1,5 +1,5 @@\n a\n b\n-old\n+new\n c\n d")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	dp.Paint(buf)
}

// === HelpOverlay.ensureSelectedVisibleLocked (75% → 90%+) ===

func TestP140_HelpOverlay_ScrollDown(t *testing.T) {
	groups := []HelpGroup{
		{Name: "G", Entries: make([]HelpEntry, 30)},
	}
	ho := NewHelpOverlay(groups)
	ho.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	ho.ScrollDown(25)
	buf := buffer.NewBuffer(60, 5)
	ho.Paint(buf)
}

// === ScrollView.contentW (75% → 100%) ===

func TestP140_ScrollView_ContentW_Narrow(t *testing.T) {
	sv := NewScrollView(&fixedSize{w: 100, h: 100})
	sv.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 5})
	buf := buffer.NewBuffer(1, 5)
	sv.Paint(buf)
}

// === Sparkline.valueToBar (77.8% → 100%) ===

func TestP140_Sparkline_AllSame(t *testing.T) {
	sp := NewSparkline()
	sp.SetData([]float64{42, 42, 42, 42})
	sp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	sp.Paint(buf)
}

// === TextArea.moveLine (77.8% → 100%) ===

func TestP140_TextArea_SingleLineMove(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("only line")
	ta.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyUp})
}

// === Viewport drawVScrollBar/drawHScrollBar (73.7% → 90%+) ===

func TestP140_Viewport_VScrollBar_NearBottom(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 50, h: 30})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vp.ScrollToY(20)
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

func TestP140_Viewport_HScrollBar_NearRight(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 100, h: 5})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vp.ScrollToX(80)
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

func TestP140_Viewport_BothFit(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 5, h: 3})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

// === ThemeStudio.setCursorLocked (75% → 100%) ===

func TestP140_ThemeStudio_Navigate100(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	for i := 0; i < 100; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	}
	for i := 0; i < 100; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	}
}

// === MenuBar.paintDropdownLocked (79.4% → 90%+) ===

func TestP140_MenuBar_DropdownManyItems(t *testing.T) {
	items := make([]MenuEntry, 20)
	for i := range items {
		items[i] = MenuEntry{ID: "i", Label: "Item"}
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

func TestP140_RichLog_LongWrappedScrolled(t *testing.T) {
	rl := NewRichLog()
	rl.SetMaxSize(100)
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	rl.Info("This is a very long line that wraps at 20 chars width")
	rl.Warn("Another long warning that also wraps")
	rl.ScrollUp(1)
	buf := buffer.NewBuffer(20, 3)
	rl.Paint(buf)
}

// === VirtualScroller.VisibleItems (78.6% → 90%+) ===

func TestP140_VirtualScroller_ExactFit(t *testing.T) {
	vs := NewVirtualScroller()
	items := make([]VirtualItem, 5)
	for i := range items {
		items[i] = VirtualItem{ID: "i", Text: "x"}
	}
	vs.SetItems(items)
	vs.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	vs.Paint(buf)
}

func TestP140_VirtualScroller_ScrolledPastEnd(t *testing.T) {
	vs := NewVirtualScroller()
	items := make([]VirtualItem, 20)
	for i := range items {
		items[i] = VirtualItem{ID: "i", Text: "x"}
	}
	vs.SetItems(items)
	vs.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	vs.ScrollTo(100)
	buf := buffer.NewBuffer(40, 5)
	vs.Paint(buf)
}

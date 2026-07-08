package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === CodeBlock paintStreamingCursorLocked (74.2% → 85%+) ===

func TestP123_CodeBlock_StreamingCursor_EmptyWithTitle(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetTitle("test.go")
	cb.SetShowTitle(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	cb.Paint(buf)
}

func TestP123_CodeBlock_StreamingCursor_NotStreaming(t *testing.T) {
	cb := NewCodeBlock("go", "line1\nline2\nline3")
	cb.SetStreaming(false)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	cb.Paint(buf)
}

func TestP123_CodeBlock_StreamingCursor_LineNumbers(t *testing.T) {
	cb := NewCodeBlock("go", "a\nb\nc\nd\ne")
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	cb.Paint(buf)
}

func TestP123_CodeBlock_StreamingCursor_ScrollOffset(t *testing.T) {
	cb := NewCodeBlock("go", "a\nb\nc\nd\ne\nf\ng")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	cb.ScrollTo(5)
	buf := buffer.NewBuffer(20, 3)
	cb.Paint(buf)
}

func TestP123_CodeBlock_StreamingCursor_Narrow(t *testing.T) {
	cb := NewCodeBlock("go", "package main\nfunc main()")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	buf := buffer.NewBuffer(5, 3)
	cb.Paint(buf)
}

func TestP123_CodeBlock_StreamingCursor_ZeroBounds(t *testing.T) {
	cb := NewCodeBlock("go", "test")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf)
}

// === ContextMenu setCursorLocked (73.3% → 90%+) ===

func TestP123_ContextMenu_SetCursorNegative(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "A"))
	cm.AddItem(NewMenuItem("b", "B"))
	cm.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	cm.SetCursor(-5) // should clamp
}

func TestP123_ContextMenu_SetCursorOverflow(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "A"))
	cm.AddItem(NewMenuItem("b", "B"))
	cm.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	cm.SetCursor(100)
}

func TestP123_ContextMenu_SetCursorEmpty(t *testing.T) {
	cm := NewContextMenu()
	cm.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	cm.SetCursor(0)
}

// === Badge Measure (76.5% → 90%+) ===

func TestP123_Badge_MeasureShortText(t *testing.T) {
	b := NewBadge("X", BadgeSuccess)
	s := b.Measure(Bounded(50, 5))
	if s.W <= 0 || s.H <= 0 {
		t.Error("expected positive size")
	}
}

func TestP123_Badge_MeasureWithIcon(t *testing.T) {
	b := NewBadge("Test", BadgeError)
	b.SetIcon("*")
	s := b.Measure(Bounded(50, 5))
	if s.W <= 0 {
		t.Error("expected positive width")
	}
}

func TestP123_Badge_MeasureNarrowClamp(t *testing.T) {
	b := NewBadge("VeryLongBadgeText", BadgeWarning)
	s := b.Measure(Bounded(3, 1))
	if s.W > 3 {
		t.Errorf("expected width <= 3, got %d", s.W)
	}
}

// === ProgressBar formatPercent (77.8% → 90%+) ===

func TestP123_ProgressBar_FormatPercent_100(t *testing.T) {
	p := NewProgressBar()
	p.SetProgress(1.0)
	p.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	p.Paint(buf)
}

func TestP123_ProgressBar_FormatPercent_50(t *testing.T) {
	p := NewProgressBar()
	p.SetProgress(0.5)
	p.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	p.Paint(buf)
}

func TestP123_ProgressBar_FormatPercent_0(t *testing.T) {
	p := NewProgressBar()
	p.SetProgress(0)
	p.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	p.Paint(buf)
}

func TestP123_ProgressBar_FormatPercent_Over100(t *testing.T) {
	p := NewProgressBar()
	p.SetProgress(1.5) // >100%
	p.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	p.Paint(buf)
}

// === DiffPreview SetShowLineNumbers/SetShowStats (0% → 100%) ===

func TestP123_DiffPreview_SetShowLineNumbers(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+added\n-removed")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	dp.Paint(buf) // before
	dp.SetShowLineNumbers(true)
	dp.Paint(buf) // after
}

func TestP123_DiffPreview_SetShowStats(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+added\n-removed\n context")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	dp.Paint(buf)
	dp.SetShowStats(true)
	dp.Paint(buf)
}

// === Viewport drawVScrollBar (73.7% → 90%+) ===

func TestP123_Viewport_DrawVScrollBar(t *testing.T) {
	vp := NewViewport(newFixedChild(10, 30))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	vp.Paint(buf) // content taller than viewport → scrollbar appears
}

func TestP123_Viewport_DrawVScrollBar_ScrollToBottom(t *testing.T) {
	vp := NewViewport(newFixedChild(10, 30))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	vp.ScrollToBottom()
	buf := buffer.NewBuffer(20, 5)
	vp.Paint(buf)
}

// === Viewport drawHScrollBar (73.7% → 90%+) ===

func TestP123_Viewport_DrawHScrollBar(t *testing.T) {
	vp := NewViewport(newFixedChild(50, 5))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf) // content wider than viewport → h-scrollbar appears
}

func TestP123_Viewport_DrawHScrollBar_ScrollRight(t *testing.T) {
	vp := NewViewport(newFixedChild(50, 5))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vp.ScrollToX(30)
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

// === AutoComplete Paint (76.7% → 85%+) ===

func TestP123_AutoComplete_Paint_EmptyItems(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	ac.Paint(buf)
}

func TestP123_AutoComplete_Paint_WithSelection(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "item1", Value: "1"},
		{Label: "item2", Value: "2"},
		{Label: "item3", Value: "3"},
	})
	ac.SetCursor(1)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	ac.Paint(buf)
}

// === BaseComponent Paint (0% → 100%) ===

func TestP123_BaseComponent_Paint(t *testing.T) {
	bc := BaseComponent{}
	bc.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	bc.Paint(buf) // should be no-op, not panic
}

func TestP123_BaseComponent_Measure(t *testing.T) {
	bc := BaseComponent{}
	s := bc.Measure(Bounded(100, 50))
	if s.W != 0 || s.H != 0 {
		t.Error("expected zero size for BaseComponent")
	}
}

// === Table drawCellLocked (75% → 85%+) ===

func TestP123_Table_DrawCellLongContent(t *testing.T) {
	tbl := NewTable([]string{"A", "B"})
	tbl.AddRow([]string{"very long content that exceeds cell width", "B"})
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 5})
	buf := buffer.NewBuffer(15, 5)
	tbl.Paint(buf)
}

func TestP123_Table_DrawCellUnicode(t *testing.T) {
	tbl := NewTable([]string{"Name", "Type"})
	tbl.AddRow([]string{"日本語テキスト", "unicode"})
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	tbl.Paint(buf)
}

// === ScrollView contentW tested via Paint with overflow ===

func TestP123_ScrollView_PaintWithOverflow(t *testing.T) {
	sv := NewScrollView(newFixedChild(10, 5))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	sv.Paint(buf) // exercises contentW internally
}

// === Sparkline recomputeRange (77.8% → 90%+) ===

func TestP123_Sparkline_RecomputeRange(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{1.0, 5.0, 3.0, 8.0, 2.0})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	sl.Paint(buf)
}

func TestP123_Sparkline_RecomputeRange_AllSame(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{5.0, 5.0, 5.0, 5.0})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	sl.Paint(buf)
}

// === TextArea moveLine (77.8% → 90%+) ===

func TestP123_TextArea_MoveLineUp(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("line1\nline2\nline3")
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyUp}) // move back up
}

func TestP123_TextArea_MoveLineEmpty(t *testing.T) {
	ta := NewTextArea()
	ta.HandleKey(&term.KeyEvent{Key: term.KeyUp})   // moveLine on empty
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown}) // moveLine on empty
}

// === RichLog countVisibleLinesLocked (78.6% → 90%+) ===

func TestP123_RichLog_MultiLineWrap(t *testing.T) {
	rl := NewRichLog()
	rl.SetMaxSize(100)
	rl.Write(LogInfo, "short")
	rl.Write(LogInfo, "a very long line that should wrap when the viewport is narrow")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	rl.Paint(buf)
}

func TestP123_RichLog_NoEntries(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	rl.Paint(buf)
}

// === ThemeStudio setCursorLocked (75% → 85%+) ===

func TestP123_ThemeStudio_SetCursorNegative(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	ts.HandleKey(&term.KeyEvent{Key: term.KeyUp}) // cursor above first
}

func TestP123_ThemeStudio_SetCursorOverflow(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	// Navigate down many times to overflow
	for i := 0; i < 50; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	}
}

// === VirtualScroller VisibleItems (78.6% → 85%+) ===

func TestP123_VirtualScroller_VisibleItems(t *testing.T) {
	vs := NewVirtualScroller()
	items := make([]VirtualItem, 50)
	for i := range items {
		items[i] = VirtualItem{ID: "i", Text: "item", Data: nil}
	}
	vs.SetItems(items)
	vs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	vs.Paint(buf)
}

func TestP123_VirtualScroller_VisibleItemsScrolled(t *testing.T) {
	vs := NewVirtualScroller()
	items := make([]VirtualItem, 50)
	for i := range items {
		items[i] = VirtualItem{ID: "i", Text: "item", Data: nil}
	}
	vs.SetItems(items)
	vs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	vs.ScrollTo(20)
	buf := buffer.NewBuffer(20, 5)
	vs.Paint(buf)
}

// === Form HandleKey (78.3% → 85%+) ===

func TestP123_Form_HandleKeyTab(t *testing.T) {
	f := NewForm()
	f.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	f.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	f.HandleKey(&term.KeyEvent{Key: term.KeyTab, Modifiers: term.ModShift})
}

// === SelectField Value (66.7% → 90%+) ===

func TestP123_SelectField_ValueEmpty(t *testing.T) {
	sf := NewSelectField("Label", "key", []string{})
	v := sf.Value()
	if v != "" {
		t.Errorf("expected empty, got %s", v)
	}
}

func TestP123_SelectField_ValueNegativeIndex(t *testing.T) {
	sf := NewSelectField("Label", "key", []string{"A", "B", "C"})
	sf.SetSelectedIndex(-5)
	v := sf.Value()
	_ = v // should not panic
}

// === Gauge paintVertical (71.1% → 90%+) ===

func TestP123_Gauge_Vertical(t *testing.T) {
	g := NewGauge()
	g.SetOrientation(GaugeVertical)
	g.SetValue(0.5)
	g.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 10})
	buf := buffer.NewBuffer(5, 10)
	g.Paint(buf)
}

func TestP123_Gauge_VerticalFull(t *testing.T) {
	g := NewGauge()
	g.SetOrientation(GaugeVertical)
	g.SetValue(1.0)
	g.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 10})
	buf := buffer.NewBuffer(5, 10)
	g.Paint(buf)
}

func TestP123_Gauge_VerticalEmpty(t *testing.T) {
	g := NewGauge()
	g.SetOrientation(GaugeVertical)
	g.SetValue(0)
	g.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 10})
	buf := buffer.NewBuffer(5, 10)
	g.Paint(buf)
}

// === HelpOverlay ensureSelectedVisibleLocked (75% → 85%+) ===

func TestP123_HelpOverlay_ScrollDown(t *testing.T) {
	groups := []HelpGroup{
		{Name: "G1", Entries: []HelpEntry{
			{Keys: "a", Description: "desc a"},
			{Keys: "b", Description: "desc b"},
			{Keys: "c", Description: "desc c"},
			{Keys: "d", Description: "desc d"},
			{Keys: "e", Description: "desc e"},
		}},
	}
	ho := NewHelpOverlay(groups)
	ho.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 4})
	ho.ScrollDown(3)
	buf := buffer.NewBuffer(50, 4)
	ho.Paint(buf)
}

// === Markdown: NewHighlighterWithStyle nil fallback ===
// (tested in markdown package separately)

// helper: fixed-size child for viewport tests
type fixedSizeChild struct {
	BaseComponent
	w, h int
}

func (f *fixedSizeChild) Measure(cs Constraints) Size {
	return Size{W: f.w, H: f.h}
}

func (f *fixedSizeChild) Paint(buf *buffer.Buffer) {}

func newFixedChild(w, h int) *fixedSizeChild {
	return &fixedSizeChild{w: w, h: h}
}

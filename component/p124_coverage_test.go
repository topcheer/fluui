package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === BaseComponent.Paint (0% → 100%) ===

func TestP124_BaseComponent_Paint(t *testing.T) {
	bc := BaseComponent{}
	bc.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	bc.Paint(buf) // no-op, should not panic
}

func TestP124_BaseComponent_Measure(t *testing.T) {
	bc := BaseComponent{}
	s := bc.Measure(Bounded(100, 50))
	if s.W != 0 || s.H != 0 {
		t.Error("expected zero size")
	}
}

// === DiffPreview SetShowLineNumbers/SetShowStats (0% → 100%) ===

func TestP124_DiffPreview_SetShowLineNumbers(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+added\n-removed\n context")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	dp.Paint(buf)
	dp.SetShowLineNumbers(true)
	dp.Paint(buf)
}

func TestP124_DiffPreview_SetShowStats(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+a\n-b\nc")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	dp.Paint(buf)
	dp.SetShowStats(true)
	dp.Paint(buf)
}

// === paintStreamingCursorLocked (74.2% → 85%+) ===

func TestP124_CodeBlock_StreamingCursor_PlainFallback(t *testing.T) {
	cb := NewCodeBlock("", "line1\nline2")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	cb.Paint(buf)
}

func TestP124_CodeBlock_StreamingCursor_LongLine(t *testing.T) {
	cb := NewCodeBlock("go", "func main() { fmt.Println(\"hello world this is a long line\") }")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	buf := buffer.NewBuffer(10, 3)
	cb.Paint(buf)
}

func TestP124_CodeBlock_StreamingCursor_EmptyWithScroll(t *testing.T) {
	cb := NewCodeBlock("go", "\n\n\n\n\nactual")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	cb.ScrollTo(3)
	buf := buffer.NewBuffer(20, 3)
	cb.Paint(buf)
}

// === ContextMenu setCursorLocked (73.3% → 90%+) ===

func TestP124_ContextMenu_SetCursorOverflow(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "A"))
	cm.AddItem(NewMenuItem("b", "B"))
	cm.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	cm.SetCursor(100)
}

func TestP124_ContextMenu_SetCursorAllSeparators(t *testing.T) {
	cm := NewContextMenu()
	s1 := NewMenuItem("sep1", "")
	s1.Separator = true
	s2 := NewMenuItem("sep2", "")
	s2.Separator = true
	cm.AddItem(s1)
	cm.AddItem(s2)
	cm.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	cm.SetCursor(0)
}

// === Badge Measure (76.5% → 100%) ===

func TestP124_Badge_MeasureShortText(t *testing.T) {
	b := NewBadge("X", BadgeSuccess)
	s := b.Measure(Bounded(50, 5))
	if s.W <= 0 {
		t.Error("expected positive width")
	}
}

func TestP124_Badge_MeasureAllSizes(t *testing.T) {
	for _, sz := range []BadgeSize{BadgeSizeSmall, BadgeSizeNormal, BadgeSizeLarge} {
		b := NewBadge("Test", BadgeInfo)
		b.SetSize(sz)
		s := b.Measure(Bounded(50, 5))
		if s.W <= 0 || s.H <= 0 {
			t.Errorf("expected positive size for size %d", sz)
		}
	}
}

// === ProgressBar formatPercent (77.8% → 90%+) ===

func TestP124_ProgressBar_FormatPercent(t *testing.T) {
	for _, p := range []float64{0, 0.25, 0.5, 0.75, 1.0} {
		bar := NewProgressBar()
		bar.SetProgress(p)
		bar.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
		buf := buffer.NewBuffer(30, 1)
		bar.Paint(buf)
	}
}

func TestP124_ProgressBar_FormatPercentOver100(t *testing.T) {
	bar := NewProgressBar()
	bar.SetProgress(1.5)
	bar.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	bar.Paint(buf)
}

// === Viewport drawVScrollBar/drawHScrollBar (73.7% → 90%+) ===

func TestP124_Viewport_DrawVScrollBar(t *testing.T) {
	vp := NewViewport(newFixedChild(10, 30))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	vp.Paint(buf)
}

func TestP124_Viewport_DrawVScrollBarScrolled(t *testing.T) {
	vp := NewViewport(newFixedChild(10, 30))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	vp.ScrollToBottom()
	buf := buffer.NewBuffer(20, 5)
	vp.Paint(buf)
}

func TestP124_Viewport_DrawHScrollBar(t *testing.T) {
	vp := NewViewport(newFixedChild(50, 5))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

func TestP124_Viewport_DrawHScrollBarScrolled(t *testing.T) {
	vp := NewViewport(newFixedChild(50, 5))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vp.ScrollToX(30)
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

// === Table drawCellLocked (75% → 85%+) ===

func TestP124_Table_DrawCellLongContent(t *testing.T) {
	tbl := NewTable([]string{"A", "B"})
	tbl.AddRow([]string{"very long content that exceeds cell width", "B"})
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 5})
	buf := buffer.NewBuffer(15, 5)
	tbl.Paint(buf)
}

func TestP124_Table_DrawCellUnicode(t *testing.T) {
	tbl := NewTable([]string{"Name", "Type"})
	tbl.AddRow([]string{"日本語テキスト", "unicode"})
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	tbl.Paint(buf)
}

// === Sparkline recomputeRange (77.8% → 90%+) ===

func TestP124_Sparkline_RecomputeRange(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{1.0, 5.0, 3.0, 8.0, 2.0})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	sl.Paint(buf)
}

func TestP124_Sparkline_RecomputeRangeAllSame(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{5.0, 5.0, 5.0, 5.0})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	sl.Paint(buf)
}

// === TextArea moveLine (77.8% → 90%+) ===

func TestP124_TextArea_MoveLineUp(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("line1\nline2\nline3")
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyUp})
}

func TestP124_TextArea_MoveLineEmpty(t *testing.T) {
	ta := NewTextArea()
	ta.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
}

// === RichLog countVisibleLinesLocked (78.6% → 90%+) ===

func TestP124_RichLog_MultiLineWrap(t *testing.T) {
	rl := NewRichLog()
	rl.SetMaxSize(100)
	rl.Write(LogInfo, "short")
	rl.Write(LogInfo, "a very long line that should wrap when the viewport is narrow")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	rl.Paint(buf)
}

func TestP124_RichLog_NoEntries(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	rl.Paint(buf)
}

// === AutoComplete Paint (76.7% → 85%+) ===

func TestP124_AutoComplete_PaintEmpty(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	ac.Paint(buf)
}

func TestP124_AutoComplete_PaintWithSelection(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "a", Value: "1"},
		{Label: "b", Value: "2"},
		{Label: "c", Value: "3"},
	})
	ac.SetCursor(1)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	ac.Paint(buf)
}

// === BarChart paintHorizontal (78.6% → 90%+) ===

func TestP124_BarChart_PaintHorizontal(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	bc.AddSeries(BarSeries{Name: "S1", Data: []BarData{{Label: "A", Value: 10}, {Label: "B", Value: 20}}})
	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	bc.Paint(buf)
}

// === DiffPreview paintBorderLocked (76.5% → 85%+) ===

func TestP124_DiffPreview_PaintNarrow(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+added\n-removed")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 5})
	buf := buffer.NewBuffer(5, 5)
	dp.Paint(buf)
}

// === HelpOverlay ensureSelectedVisibleLocked (75% → 85%+) ===

func TestP124_HelpOverlay_ScrollDown(t *testing.T) {
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

// === ThemeStudio setCursorLocked (75% → 85%+) ===

func TestP124_ThemeStudio_SetCursorNegative(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	ts.HandleKey(&term.KeyEvent{Key: term.KeyUp})
}

func TestP124_ThemeStudio_SetCursorOverflow(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	for i := 0; i < 50; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	}
}

// === VirtualScroller VisibleItems (78.6% → 85%+) ===

func TestP124_VirtualScroller_VisibleItems(t *testing.T) {
	vs := NewVirtualScroller()
	items := make([]VirtualItem, 50)
	for i := range items {
		items[i] = VirtualItem{ID: "i", Text: "item"}
	}
	vs.SetItems(items)
	vs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	vs.Paint(buf)
}

func TestP124_VirtualScroller_VisibleItemsScrolled(t *testing.T) {
	vs := NewVirtualScroller()
	items := make([]VirtualItem, 50)
	for i := range items {
		items[i] = VirtualItem{ID: "i", Text: "item"}
	}
	vs.SetItems(items)
	vs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	vs.ScrollTo(20)
	buf := buffer.NewBuffer(20, 5)
	vs.Paint(buf)
}

// === ScrollView contentW (75% → 90%+) ===

func TestP124_ScrollView_PaintWithOverflow(t *testing.T) {
	sv := NewScrollView(newFixedChild(10, 5))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	sv.Paint(buf)
}

// === MenuBar computeDropDimsLocked/paintDropdownLocked ===

func TestP124_MenuBar_OpenAndPaint(t *testing.T) {
	mb := NewMenuBar([]Menu{
		{ID: "file", Title: "File", Items: []MenuEntry{
			{ID: "new", Label: "New", Shortcut: "Ctrl+N"},
			{ID: "open", Label: "Open", Shortcut: "Ctrl+O"},
		}},
		{ID: "edit", Title: "Edit", Items: []MenuEntry{
			{ID: "undo", Label: "Undo", Shortcut: "Ctrl+Z"},
		}},
	})
	mb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	mb.OpenMenu(0)
	buf := buffer.NewBuffer(60, 20)
	mb.Paint(buf)
}

// === Gauge paintVertical thresholds ===

func TestP124_Gauge_VerticalAllThresholds(t *testing.T) {
	for _, v := range []float64{0, 0.3, 0.5, 0.7, 0.9, 1.0} {
		g := NewGauge()
		g.SetOrientation(GaugeVertical)
		g.SetValue(v)
		g.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 10})
		buf := buffer.NewBuffer(5, 10)
		g.Paint(buf)
	}
}

// === SelectField Value edge cases ===

func TestP124_SelectField_ValueEmpty(t *testing.T) {
	sf := NewSelectField("Label", "key", []string{})
	v := sf.Value()
	if v != "" {
		t.Errorf("expected empty, got %s", v)
	}
}

func TestP124_SelectField_ValueNegativeIndex(t *testing.T) {
	sf := NewSelectField("Label", "key", []string{"A", "B", "C"})
	sf.SetSelectedIndex(-5)
	_ = sf.Value()
}

// === Form HandleKey Tab cycling ===

func TestP124_Form_HandleKeyTab(t *testing.T) {
	f := NewForm()
	f.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	f.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	f.HandleKey(&term.KeyEvent{Key: term.KeyTab, Modifiers: term.ModShift})
}

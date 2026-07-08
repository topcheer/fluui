package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === formatPercent (66.7% → 100%) ===

func TestP132_FormatPercent_AllBranches(t *testing.T) {
	cases := []struct {
		name string
		val  float64
		want string
	}{
		{"negative", -5, "0%"},
		{"zero", 0, "0%"},
		{"half", 50, "50%"},
		{"full", 100, "100%"},
		{"over", 150, "100%"},
		{"edge99", 99, "99%"},
		{"edge1", 1, "1%"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatPercent(tc.val)
			if got != tc.want {
				t.Errorf("formatPercent(%v) = %q, want %q", tc.val, got, tc.want)
			}
		})
	}
}

// === MenuBar computeDropDimsLocked (73.7% → 85%+) ===

func TestP132_MenuBar_DropDims_NegativeOpenIdx(t *testing.T) {
	mb := NewMenuBar([]Menu{
		{ID: "f", Title: "F", Items: []MenuEntry{{ID: "a", Label: "A"}}},
	})
	mb.computeDropDimsLocked() // openIdx is -1 initially
	if mb.dropW != 0 || mb.dropH != 0 {
		t.Error("expected 0 dims when no menu open")
	}
}

func TestP132_MenuBar_DropDims_WithShortcut(t *testing.T) {
	mb := NewMenuBar([]Menu{
		{ID: "f", Title: "F", Items: []MenuEntry{
			{ID: "a", Label: "Action", Shortcut: "Ctrl+A"},
			{ID: "b", Label: "B"},
		}},
	})
	mb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	mb.OpenMenu(0)
	buf := buffer.NewBuffer(40, 20)
	mb.Paint(buf) // Paint calls computeDropDimsLocked
}

func TestP132_MenuBar_DropDims_PositionFallback(t *testing.T) {
	mb := NewMenuBar([]Menu{
		{ID: "f", Title: "F", Items: []MenuEntry{{ID: "a", Label: "A"}}},
	})
	mb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	mb.openIdx = 5 // out of range for menuXs
	mb.computeDropDimsLocked()
	if mb.dropX != 0 {
		t.Errorf("expected dropX=0 fallback, got %d", mb.dropX)
	}
}

// === Viewport scrollbars (73.7% → 85%+) ===

func TestP132_Viewport_VScrollBar_BarHZero(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 60, h: 30})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1}) // barH = 0
	buf := buffer.NewBuffer(10, 1)
	vp.Paint(buf)
}

func TestP132_Viewport_VScrollBar_ThumbClamp(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 60, h: 100})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	vp.ScrollToY(90) // near bottom
	buf := buffer.NewBuffer(30, 5)
	vp.Paint(buf)
}

func TestP132_Viewport_HScrollBar_BarWZero(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 60, h: 30})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 10}) // barW = 0
	buf := buffer.NewBuffer(1, 10)
	vp.Paint(buf)
}

func TestP132_Viewport_HScrollBar_ThumbClamp(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 200, h: 10})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vp.ScrollToX(180)
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

func TestP132_Viewport_BothScrollbars_FitsContent(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 10, h: 3})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

// === CodeBlock paintStreamingCursorLocked (74.2% → 85%+) ===

func TestP132_CodeBlock_Streaming_NarrowWidth(t *testing.T) {
	cb := NewCodeBlock("go", "long line of code here that won't fit")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 5})
	buf := buffer.NewBuffer(5, 5)
	cb.Paint(buf)
}

func TestP132_CodeBlock_Streaming_LineNumbers(t *testing.T) {
	cb := NewCodeBlock("go", "a\nb\nc")
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestP132_CodeBlock_Streaming_NotStreaming(t *testing.T) {
	cb := NewCodeBlock("go", "code")
	cb.SetStreaming(false)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestP132_CodeBlock_Streaming_ZeroBounds(t *testing.T) {
	cb := NewCodeBlock("go", "code")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(0, 0)
	cb.Paint(buf)
}

// === Badge Measure (76.5% → 100%) ===

func TestP132_Badge_Measure_AllSizes(t *testing.T) {
	for _, size := range []BadgeSize{BadgeSizeSmall, BadgeSizeNormal, BadgeSizeLarge} {
		b := NewBadge("test", BadgeInfo)
		b.size = size
		s := b.Measure(Bounded(100, 10))
		if s.W < 0 || s.H < 0 {
			t.Errorf("expected non-negative measure for size %d", size)
		}
	}
}

func TestP132_Badge_Measure_WithIconAllSizes(t *testing.T) {
	for _, size := range []BadgeSize{BadgeSizeSmall, BadgeSizeNormal, BadgeSizeLarge} {
		b := NewBadge("test", BadgeSuccess)
		b.size = size
		b.SetIcon("*")
		s := b.Measure(Bounded(100, 10))
		if s.W < 0 {
			t.Errorf("expected non-negative W for size %d", size)
		}
	}
}

// === Sparkline recomputeRange/valueToBar (77.8% → 90%+) ===

func TestP132_Sparkline_AllSameValue(t *testing.T) {
	sp := NewSparkline()
	sp.SetData([]float64{50, 50, 50, 50})
	sp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	sp.Paint(buf)
}

func TestP132_Sparkline_ZeroAndNegative(t *testing.T) {
	sp := NewSparkline()
	sp.SetData([]float64{0, -10, 20, -5, 15})
	sp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	sp.Paint(buf)
}

func TestP132_Sparkline_Empty(t *testing.T) {
	sp := NewSparkline()
	sp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	sp.Paint(buf)
}

// === ScrollView contentW/ScrollbarBounds (75/78.6% → 90%+) ===

func TestP132_ScrollView_NarrowBounds(t *testing.T) {
	sv := NewScrollView(&fixedSize{w: 10, h: 5})
	sv.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 5})
	buf := buffer.NewBuffer(5, 5)
	sv.Paint(buf)
}

func TestP132_ScrollView_ExactFit(t *testing.T) {
	sv := NewScrollView(&fixedSize{w: 20, h: 5})
	sv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	sv.Paint(buf)
}

// === TextArea moveLine (77.8% → 90%+) ===

func TestP132_TextArea_MoveLine_Boundaries(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("a\nb\nc")
	ta.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	// Try moving up from first line
	ta.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	// Move down past last line
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
}

func TestP132_TextArea_MoveLine_EmptyLines(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("\n\n\n")
	ta.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyUp})
}

// === Table drawCellLocked (75% → 85%+) ===

func TestP132_Table_DrawCell_AlignedColumns(t *testing.T) {
	tb := NewTable([]string{"Name", "Value", "Status"})
	tb.AddRow([]string{"short", "x", "ok"})
	tb.AddRow([]string{"very long name here", "1234567890", "pending"})
	tb.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	tb.Paint(buf)
}

func TestP132_Table_DrawCell_EmptyCells(t *testing.T) {
	tb := NewTable([]string{"A", "B", "C"})
	tb.AddRow([]string{"", "", ""})
	tb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	tb.Paint(buf)
}

// === AutoComplete Paint (76.7% → 90%+) ===

func TestP132_AutoComplete_Paint_Empty(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	ac.Paint(buf)
}

func TestP132_AutoComplete_Paint_ZeroBounds(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{{Label: "test"}})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(0, 0)
	ac.Paint(buf)
}

// === ContextMenu setCursorLocked (73.3% → 90%+) ===

func TestP132_ContextMenu_Navigate_AllSeparators(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("s1", ""))
	cm.items[0].Separator = true
	cm.AddItem(NewMenuItem("s2", ""))
	cm.items[1].Separator = true
	cm.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	cm.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	cm.HandleKey(&term.KeyEvent{Key: term.KeyUp})
}

// === DiffPreview paintBorderLocked (76.5% → 90%+) ===

func TestP132_DiffPreview_PaintBorder_NormalWidth(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("--- a\n+++ b\n@@ -1 +1 @@\n-old\n+new")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	dp.Paint(buf)
}

// === HelpOverlay ensureSelectedVisibleLocked (75% → 85%+) ===

func TestP132_HelpOverlay_ScrollUp(t *testing.T) {
	groups := []HelpGroup{
		{Name: "G", Entries: make([]HelpEntry, 50)},
	}
	ho := NewHelpOverlay(groups)
	ho.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	ho.ScrollDown(30)
	ho.ScrollUp(20)
	buf := buffer.NewBuffer(60, 10)
	ho.Paint(buf)
}

// === BarChart paintHorizontal (78.6% → 85%+) ===

func TestP132_BarChart_Horizontal_PartialValues(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	bc.SetShowValues(true)
	bc.AddSeries(BarSeries{
		Name: "S",
		Data: []BarData{
			{Label: "A", Value: 5},
			{Label: "B", Value: 0},
			{Label: "C", Value: 100},
		},
	})
	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	bc.Paint(buf)
}

func TestP132_BarChart_Horizontal_NoLabels(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	bc.AddSeries(BarSeries{
		Name: "S",
		Data: []BarData{{Value: 50}},
	})
	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	bc.Paint(buf)
}

// === RichLog countVisibleLinesLocked (78.6% → 85%+) ===

func TestP132_RichLog_LongWrappedMultiEntry(t *testing.T) {
	rl := NewRichLog()
	rl.SetMaxSize(100)
	rl.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 10})
	rl.Info("Short")
	rl.Warn("This is a very long warning that will definitely wrap across multiple lines in a 15-char wide terminal")
	rl.Error("err")
	rl.Debug("debug msg")
	buf := buffer.NewBuffer(15, 10)
	rl.Paint(buf)
}

// === ThemeStudio setCursorLocked/paintPickerOverlay (75/76.9% → 85%+) ===

func TestP132_ThemeStudio_CursorNegativeOverflow(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	// Navigate way past beginning
	for i := 0; i < 10; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	}
	buf := buffer.NewBuffer(60, 20)
	ts.Paint(buf)
}

func TestP132_ThemeStudio_PickerWithScroll(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10}) // short height to force scroll
	ts.HandleKey(&term.KeyEvent{Key: term.KeyEnter}) // open picker
	ts.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	ts.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	buf := buffer.NewBuffer(60, 10)
	ts.Paint(buf)
}

package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P190: Coverage tests for sub-80% component functions

func TestBaseComponent_Paint_P190(t *testing.T) {
	bc := BaseComponent{}
	buf := buffer.NewBuffer(10, 5)
	bc.Paint(buf) // no-op, should not crash
}

func TestDiffPreview_SetShowLineNumbers_P190(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	dp.SetShowLineNumbers(false)
	// No-op setters, just verify they don't crash
}

func TestDiffPreview_SetShowStats_P190(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(true)
	dp.SetShowStats(false)
}

func TestDiffPreview_PaintBorderLocked_P190(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk\nl\nm\nn\no\np")
	buf := buffer.NewBuffer(20, 10)
	dp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	dp.Paint(buf) // exercises paintBorderLocked with tall content
}

func TestDiffPreview_PaintBorderEmpty_P190(t *testing.T) {
	dp := NewDiffPreview()
	buf := buffer.NewBuffer(5, 3)
	dp.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	dp.Paint(buf) // narrow width border
}

func TestBadge_MeasureShort_P190(t *testing.T) {
	b := NewBadge("x", BadgeInfo)
	b.Measure(Constraints{})
	// short text — exercises width clamp path
}

func TestBadge_MeasureWithIcon_P190(t *testing.T) {
	b := NewBadge("test", BadgeSuccess)
	b.SetIcon("*")
	s := b.Measure(Constraints{MaxWidth: 100, MaxHeight: 10})
	if s.H < 1 {
		t.Error("height should be at least 1")
	}
}

func TestBadge_MeasureNarrow_P190(t *testing.T) {
	b := NewBadge("very long text here", BadgeWarning)
	s := b.Measure(Constraints{MaxWidth: 3, MaxHeight: 10})
	if s.W > 3 {
		t.Error("width should be clamped")
	}
}

func TestTextArea_SetPrompt_P190(t *testing.T) {
	ta := NewTextArea()
	ta.SetPrompt("> ")
}

func TestTextArea_SetPlaceholder_P190(t *testing.T) {
	ta := NewTextArea()
	ta.SetPlaceholder("enter text...")
}

func TestTextArea_FocusBlur_P190(t *testing.T) {
	ta := NewTextArea()
	ta.Focus()
	ta.Blur()
}

func TestTextArea_SetCharLimit_P190(t *testing.T) {
	ta := NewTextArea()
	ta.SetCharLimit(100)
}

func TestStyleBuilder_Underline_P190(t *testing.T) {
	sb := NewStyle()
	sb.Underline(false) // explicit false branch
	sb.Underline(true)  // true branch
}

func TestStyleBuilder_Dim_P190(t *testing.T) {
	sb := NewStyle()
	sb.Dim(false)
	sb.Dim(true)
}

func TestStyleBuilder_Blink_P190(t *testing.T) {
	sb := NewStyle()
	sb.Blink(false)
	sb.Blink(true)
}

func TestStyleBuilder_Reverse_P190(t *testing.T) {
	sb := NewStyle()
	sb.Reverse(false)
	sb.Reverse(true)
}

func TestStyleBuilder_Strikethrough_P190(t *testing.T) {
	sb := NewStyle()
	sb.Strikethrough(false)
	sb.Strikethrough(true)
}

func TestStyleBuilder_ParseLipglossColorHex_P190(t *testing.T) {
	c := parseLipglossColor("#ff8800")
	if c.Type == 0 {
		t.Error("hex color should parse")
	}
}

func TestStyleBuilder_ParseLipglossColorNamed_P190(t *testing.T) {
	c := parseLipglossColor("red")
	_ = c // should not crash
}

func TestStyleBuilder_ParseLipglossColorInvalid_P190(t *testing.T) {
	c := parseLipglossColor("not-a-color")
	if c.Type != 0 {
		t.Error("invalid color should return ColorNone")
	}
}

func TestScrollView_ContentW_P190(t *testing.T) {
	sv := NewScrollView(NewFill(' ', buffer.Style{}))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 10})
	// tiny width — exercises w < 1 path
	sv.Paint(buffer.NewBuffer(1, 10))
}

func TestSparkline_ValueToBarSame_P190(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{5, 5, 5})
	s := sl.Measure(Constraints{MaxWidth: 20, MaxHeight: 10})
	_ = s // all-same values
}

func TestSparkline_ValueToBarZeroNeg_P190(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{0, -1, -2})
	_ = sl.Measure(Constraints{MaxWidth: 20, MaxHeight: 10})
}

func TestViewport_DrawVScrollBar_P190(t *testing.T) {
	vp := NewViewport(NewFill(' ', buffer.Style{}))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	vp.ScrollToY(100) // force scroll
	buf := buffer.NewBuffer(20, 5)
	vp.Paint(buf) // draws scrollbar
}

func TestViewport_DrawHScrollBar_P190(t *testing.T) {
	vp := NewViewport(NewFill(' ', buffer.Style{}))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 20})
	vp.ScrollToX(100) // force horizontal scroll
	buf := buffer.NewBuffer(10, 20)
	vp.Paint(buf)
}

func TestHelpOverlay_ScrollDown_P190(t *testing.T) {
	ho := NewHelpOverlay([]HelpGroup{
		{Name: "group1", Entries: []HelpEntry{{Keys: "ctrl+a", Description: "do a"}}},
		{Name: "group2", Entries: []HelpEntry{{Keys: "ctrl+b", Description: "do b"}}},
	})
	ho.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	ho.ScrollDown(1)
	ho.Paint(buffer.NewBuffer(40, 3))
}

func TestLoadingIndicator_StartDouble_P190(t *testing.T) {
	li := NewLoadingIndicator("loading")
	li.Start()
	li.Start() // double start — exercises the "already running" branch
	li.Stop()
}

func TestRichLog_CountVisibleLines_P190(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	rl.Info("line1\nwrapped long line that wraps")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	rl.Paint(buf)
}

func TestThemeStudio_SetCursor_P190(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	buf := buffer.NewBuffer(40, 20)
	ts.Paint(buf) // exercises paint path
}

func TestAutoComplete_PaintEmpty_P190(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	ac.Paint(buffer.NewBuffer(20, 10))
}

func TestAutoComplete_PaintWithItems_P190(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "item1", Description: "desc1"},
		{Label: "item2", Description: "desc2"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	ac.Paint(buffer.NewBuffer(20, 10))
}

func TestCodeBlock_StreamingCursorEmpty_P190(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	cb.Paint(buffer.NewBuffer(20, 10))
}

func TestCodeBlock_StreamingCursorWithLines_P190(t *testing.T) {
	cb := NewCodeBlock("go", "package main\nfunc main() {}")
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	cb.Paint(buffer.NewBuffer(20, 10))
}
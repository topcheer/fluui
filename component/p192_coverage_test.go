package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P192: Coverage tests for sub-80% component functions

func TestBaseComponent_Paint_P192(t *testing.T) {
	bc := BaseComponent{}
	bc.Paint(buffer.NewBuffer(10, 5))
}

func TestDiffPreview_Setters_P192(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	dp.SetShowStats(true)
}

func TestDiffPreview_PaintBorderNarrow_P192(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("a\nb\nc")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 5})
	dp.Paint(buffer.NewBuffer(3, 5))
}

func TestDiffPreview_PaintBorderTall_P192(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk\nl\nm\nn\no")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	dp.Paint(buffer.NewBuffer(40, 3))
}

func TestBadge_MeasureAllVariants_P192(t *testing.T) {
	for _, v := range []BadgeVariant{BadgeInfo, BadgeSuccess, BadgeWarning, BadgeError} {
		b := NewBadge("x", v)
		b.Measure(Constraints{})
		b2 := NewBadge("text", v)
		b2.SetIcon("*")
		b2.Measure(Constraints{MaxWidth: 2, MaxHeight: 1})
	}
}

func TestTextArea_BubblesAPI_P192(t *testing.T) {
	ta := NewTextArea()
	ta.SetPrompt("> ")
	ta.SetPlaceholder("enter...")
	ta.Focus()
	ta.Blur()
	ta.SetCharLimit(100)
}

func TestCodeBlock_StreamingCursor_P192(t *testing.T) {
	cb := NewCodeBlock("go", "package main\nfunc main() {}")
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	cb.Paint(buffer.NewBuffer(20, 5))
}

func TestCodeBlock_StreamingCursorNarrow_P192(t *testing.T) {
	cb := NewCodeBlock("go", "long line of code here that exceeds width")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 5})
	cb.Paint(buffer.NewBuffer(5, 5))
}

func TestViewport_VScrollBar_P192(t *testing.T) {
	vp := NewViewport(NewFill(' ', buffer.Style{}))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	vp.ScrollToY(50)
	vp.Paint(buffer.NewBuffer(20, 3))
}

func TestViewport_HScrollBar_P192(t *testing.T) {
	vp := NewViewport(NewFill(' ', buffer.Style{}))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 20})
	vp.ScrollToX(50)
	vp.Paint(buffer.NewBuffer(3, 20))
}

func TestAutoComplete_PaintScroll_P192(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "a", Description: "desc a"},
		{Label: "b", Description: "desc b"},
		{Label: "c"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	ac.Paint(buffer.NewBuffer(20, 3))
}

func TestSparkline_AllSame_P192(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{5, 5, 5, 5})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	sl.Paint(buffer.NewBuffer(20, 5))
}

func TestSparkline_ZeroNeg_P192(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{0, -1, -2, -3})
	_ = sl.Measure(Constraints{MaxWidth: 20, MaxHeight: 5})
}

func TestScrollView_Narrow_P192(t *testing.T) {
	sv := NewScrollView(NewFill(' ', buffer.Style{}))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 5})
	sv.Paint(buffer.NewBuffer(1, 5))
}

func TestStyleBuilder_ParseLipglossColor_P192(t *testing.T) {
	parseLipglossColor("#ff8800")
	parseLipglossColor("red")
	parseLipglossColor("invalid")
	parseLipglossColor("")
}

func TestStyleBuilder_Inherit_P192(t *testing.T) {
	parent := NewStyle().Bold(true).Foreground(buffer.RGB(255, 0, 0))
	child := NewStyle().Italic(true)
	child.Inherit(parent)
}

func TestHelpOverlay_ScrollDown_P192(t *testing.T) {
	ho := NewHelpOverlay([]HelpGroup{
		{Name: "g1", Entries: []HelpEntry{{Keys: "ctrl+a", Description: "a"}}},
		{Name: "g2", Entries: []HelpEntry{{Keys: "ctrl+b", Description: "b"}}},
	})
	ho.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	ho.ScrollDown(1)
	ho.Paint(buffer.NewBuffer(40, 3))
}

func TestLoadingIndicator_DoubleStart_P192(t *testing.T) {
	li := NewLoadingIndicator("loading")
	li.Start()
	li.Start()
	li.Stop()
}

func TestRichLog_WrappedWithTime_P192(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	rl.Info("very long wrapped line that exceeds width and wraps to multiple lines")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	rl.Paint(buffer.NewBuffer(20, 10))
}

func TestThemeStudio_CursorCycle_P192(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	ts.Paint(buffer.NewBuffer(40, 20))
}
package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// P188: Target exact uncovered branches in sub-80% functions

// === Badge.Measure 73.3% → 85%+ (missing: w<2 clamp, h<1 clamp) ===
func TestP188_Badge_MeasureNarrowClamp(t *testing.T) {
	b := NewBadge("X", BadgeInfo)
	s := b.Measure(Constraints{MaxWidth: 1, MaxHeight: 1})
	if s.W > 1 {
		t.Errorf("expected w<=1, got %d", s.W)
	}
}

// === AutoComplete.Paint 76.7% → 85%+ (missing: category selected+truncated, desc remaining>3, scrollY<0) ===
func TestP188_AutoComplete_PaintCategorySelected(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "cmd1", Description: "desc1", Category: "cat1"},
		{Label: "cmd2", Description: "desc2", Category: "cat2"},
		{Label: "verylonglabelthatistruncated", Description: "verylongdescription", Category: "cat1"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	ac.SetQuery("cmd")
	ac.SetCursor(2)
	ac.Paint(buffer.NewBuffer(20, 5))
}

func TestP188_AutoComplete_PaintNegativeScroll(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "item1", Description: "d1"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	ac.SetQuery("item")
	ac.Paint(buffer.NewBuffer(20, 3))
}

// === CodeBlock.paintStreamingCursorLocked 74.2% → 85%+ ===
func TestP188_CodeBlock_StreamingCursorAllCases(t *testing.T) {
	cb := NewCodeBlock("go", "line1\nline2\nline3")
	cb.SetStreaming(true)
	cb.SetTitle("test.go")
	cb.SetShowLineNumbers(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	cb.Paint(buffer.NewBuffer(10, 3))
	// not streaming
	cb2 := NewCodeBlock("go", "x:=1")
	cb2.SetStreaming(false)
	cb2.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	cb2.Paint(buffer.NewBuffer(10, 3))
	// empty
	cb3 := NewCodeBlock("go", "")
	cb3.SetStreaming(true)
	cb3.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	cb3.Paint(buffer.NewBuffer(10, 3))
	// zero bounds
	cb4 := NewCodeBlock("go", "x:=1")
	cb4.SetStreaming(true)
	cb4.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	cb4.Paint(buffer.NewBuffer(1, 1))
}

// === DiffPreview.paintBorderLocked 76.5% → 85%+ (missing: empty diff, tall diff) ===
func TestP188_DiffPreview_PaintBorderEdgeCases(t *testing.T) {
	dp := NewDiffPreview()
	// empty
	dp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	dp.Paint(buffer.NewBuffer(20, 3))
	// tall content beyond bounds
	dp2 := NewDiffPreview()
	dp2.SetDiff("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
	dp2.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	dp2.Paint(buffer.NewBuffer(20, 3))
}

// === HelpOverlay.ensureSelectedVisibleLocked 75% → 85%+ ===
func TestP188_HelpOverlay_ScrollUpDown(t *testing.T) {
	groups := []HelpGroup{
		{Name: "Global", Entries: []HelpEntry{
			{Keys: "Ctrl+Q", Description: "Quit"},
			{Keys: "Ctrl+S", Description: "Save"},
			{Keys: "Ctrl+O", Description: "Open"},
			{Keys: "Ctrl+F", Description: "Find"},
		}},
	}
	h := NewHelpOverlay(groups)
	h.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2})
	h.ScrollDown(3)
	h.Paint(buffer.NewBuffer(40, 2))
	h.ScrollUp(1)
	h.Paint(buffer.NewBuffer(40, 2))
}

// === RichLog.countVisibleLinesLocked 78.6% → 85%+ ===
func TestP188_RichLog_CountVisibleWrapped(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	for i := 0; i < 30; i++ {
		rl.Info("very long message that wraps across narrow viewport width for testing countVisibleLines")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 3})
	rl.ScrollUp(10)
	rl.Paint(buffer.NewBuffer(15, 3))
}

// === ScrollView.contentW 75% → 85%+ (missing: w<1 after scrollbar) ===
func TestP188_ScrollView_NarrowWithScrollbar(t *testing.T) {
	sv := NewScrollView(NewParagraph("test\ntext\nthat\noverflows\nthe\nheight"))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 3})
	sv.Paint(buffer.NewBuffer(1, 3))
}

// === Sparkline.valueToBar 77.8% → 85%+ (missing: all same, zero+negative) ===
func TestP188_Sparkline_ValueToBarEdges(t *testing.T) {
	sl := NewSparkline()
	// all same non-zero
	sl.SetData([]float64{5, 5, 5, 5})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	sl.Paint(buffer.NewBuffer(10, 3))
	// zero and negative
	sl2 := NewSparkline()
	sl2.SetData([]float64{0, -5, 10, -3})
	sl2.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	sl2.Paint(buffer.NewBuffer(10, 3))
}

// === StyleBuilder.parseLipglossColor 75% → 85%+ ===
func TestP188_StyleBuilder_ParseColorAllCases(t *testing.T) {
	sb := NewStyle()
	sb.ForegroundHex("#ff8800")
	sb.ForegroundHex("#abc")
	sb.ForegroundHex("#XYZ")
	sb.ForegroundHex("")
	sb.ForegroundColor("red")
	sb.ForegroundColor("brightblue")
	sb.ForegroundColor("nonexistent")
	sb.ForegroundColor("")
}

// === StyleBuilder.Inherit 71.4% → 85%+ ===
func TestP188_StyleBuilder_InheritAll(t *testing.T) {
	parent := NewStyle().Bold().Italic().Underline().Dim().Blink().Reverse().Strikethrough()
	child := NewStyle().Inherit(parent)
	if child.Style().Flags == 0 {
		t.Error("expected inherited flags")
	}
	emptyChild := NewStyle().Inherit(NewStyle())
	if emptyChild.Style().Flags != 0 {
		t.Error("expected zero flags from empty parent")
	}
}

// === LoadingIndicator.Start 75% → 85%+ (missing: double start/stop) ===
func TestP188_LoadingIndicator_DoubleStartStop(t *testing.T) {
	l := NewLoadingIndicator("loading...")
	l.Start()
	l.Start()
	l.Stop()
	l.Stop()
	l.Start()
	l.Stop()
}

// === Viewport scrollbars 73.7% → 85%+ ===
func TestP188_Viewport_ScrollbarEdges(t *testing.T) {
	// V scrollbar near bottom
	vp := NewViewport(NewParagraph("l1\nl2\nl3\nl4\nl5\nl6\nl7\nl8"))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	vp.ScrollDown(10)
	vp.Paint(buffer.NewBuffer(20, 3))
	// H scrollbar near right
	vp2 := NewViewport(NewParagraph("very long text that overflows viewport width for horizontal scroll"))
	vp2.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	vp2.ScrollRight(100)
	vp2.Paint(buffer.NewBuffer(10, 5))
	// both overflow
	vp3 := NewViewport(NewParagraph("very long line overflow\nl2\nl3\nl4\nl5\nl6\nl7"))
	vp3.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	vp3.ScrollDown(5)
	vp3.ScrollRight(20)
	vp3.Paint(buffer.NewBuffer(10, 3))
}

// === ContextMenu.setCursorLocked 73.3% → 85%+ ===
func TestP188_ContextMenu_CursorOverflowUnderflow(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("item1", "Item 1"))
	cm.AddItem(NewMenuItem("item2", "Item 2"))
	cm.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	cm.SetCursor(10)
	cm.SetCursor(-1)
	cm.Paint(buffer.NewBuffer(20, 10))
}

// === ThemeStudio.setCursorLocked 75% → 85%+ ===
func TestP188_ThemeStudio_CursorCycle(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	for i := 0; i < 50; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	}
	for i := 0; i < 50; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	}
}

// === SessionSidebar.Measure 75% → 85%+ ===
func TestP188_SessionSidebar_MeasureItems(t *testing.T) {
	sb := NewSessionSidebar()
	sb.SetItems([]SessionItem{
		{ID: "1", Title: "Session 1", Workspace: "G1"},
		{ID: "2", Title: "Session 2", Workspace: "G1"},
	})
	s := sb.Measure(Constraints{MaxWidth: 30, MaxHeight: 20})
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected non-zero, got %dx%d", s.W, s.H)
	}
}

// === TextArea.moveLine 77.8% → 85%+ ===
func TestP188_TextArea_MoveLineEmpty(t *testing.T) {
	ta := NewTextArea()
	ta.moveLine(1)
	ta.moveLine(-1)
}

func TestP188_TextArea_MoveLineSingleLine(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("only one line")
	ta.moveLine(1)
	ta.moveLine(-1)
}

// === BaseComponent.Paint 0% → 100% ===
func TestP188_BaseComponent_Paint(t *testing.T) {
	bc := BaseComponent{}
	bc.Paint(buffer.NewBuffer(10, 5))
}

func TestP188_BaseComponent_Measure(t *testing.T) {
	bc := BaseComponent{}
	s := bc.Measure(Constraints{})
	if s.W != 0 || s.H != 0 {
		t.Errorf("expected zero size, got %dx%d", s.W, s.H)
	}
}

func TestP188_BaseComponent_Children(t *testing.T) {
	bc := BaseComponent{}
	if bc.Children() != nil {
		t.Error("expected nil children")
	}
}

// === DiffPreview setters 0% → 100% ===
func TestP188_DiffPreview_SetShowLineNumbers(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	if !dp.ShowLineNumbers() {
		t.Error("expected line numbers shown")
	}
	dp.SetShowLineNumbers(false)
}

func TestP188_DiffPreview_SetShowStats(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(true)
	dp.SetShowStats(false)
}

// === TextArea bubbles API 0% → 100% ===
func TestP188_TextArea_SetPrompt(t *testing.T) {
	ta := NewTextArea()
	ta.SetPrompt("> ")
}

func TestP188_TextArea_SetPlaceholder(t *testing.T) {
	ta := NewTextArea()
	ta.SetPlaceholder("type here...")
}

func TestP188_TextArea_Focus(t *testing.T) {
	ta := NewTextArea()
	ta.Focus()
}

func TestP188_TextArea_Blur(t *testing.T) {
	ta := NewTextArea()
	ta.Blur()
}

func TestP188_TextArea_SetCharLimit(t *testing.T) {
	ta := NewTextArea()
	ta.SetCharLimit(100)
}

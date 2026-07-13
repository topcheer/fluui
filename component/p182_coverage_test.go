package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === Viewport scrollbar edge cases (73.7% → 90%+) ===

func TestP182_Viewport_VScrollBarNearBottom(t *testing.T) {
	content := NewParagraph("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8")
	vp := NewViewport(content)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	vp.ScrollDown(10)
	vp.Paint(buffer.NewBuffer(20, 3))
}

func TestP182_Viewport_HScrollBarNearRight(t *testing.T) {
	content := NewParagraph("very long text that extends far beyond the viewport width for horizontal scroll testing")
	vp := NewViewport(content)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	vp.ScrollRight(100)
	vp.Paint(buffer.NewBuffer(10, 5))
}

func TestP182_Viewport_BothScrollbarsOverflow(t *testing.T) {
	content := NewParagraph("very long line that overflows width\nline2\nline3\nline4\nline5\nline6\nline7")
	vp := NewViewport(content)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	vp.ScrollDown(5)
	vp.ScrollRight(20)
	vp.Paint(buffer.NewBuffer(10, 3))
}

// === Badge Measure (73.3% → 100%) ===

func TestP182_Badge_MeasureAllVariants(t *testing.T) {
	variants := []BadgeVariant{BadgeInfo, BadgeSuccess, BadgeWarning, BadgeError}
	for _, v := range variants {
		b := NewBadge("Test", v)
		s := b.Measure(Constraints{MaxWidth: 50, MaxHeight: 50})
		if s.W <= 0 || s.H <= 0 {
			t.Errorf("variant %d: expected non-zero, got %dx%d", v, s.W, s.H)
		}
	}
}

func TestP182_Badge_MeasureWithIconNarrow(t *testing.T) {
	b := NewBadge("Long Text", BadgeInfo)
	b.SetIcon("*")
	s := b.Measure(Constraints{MaxWidth: 2, MaxHeight: 50})
	if s.W > 2 {
		t.Errorf("expected width <= 2, got %d", s.W)
	}
}

// === CodeBlock paintStreamingCursor (74.2% → 90%+) ===

func TestP182_CodeBlock_StreamingCursorTitleLineNumbers(t *testing.T) {
	cb := NewCodeBlock("go", "x := 1\ny := 2")
	cb.SetTitle("test.go")
	cb.SetShowLineNumbers(true)
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	cb.Paint(buffer.NewBuffer(30, 5))
}

func TestP182_CodeBlock_StreamingCursorNotStreaming(t *testing.T) {
	cb := NewCodeBlock("go", "x := 1")
	cb.SetStreaming(false)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	cb.Paint(buffer.NewBuffer(20, 5))
}

func TestP182_CodeBlock_StreamingCursorEmpty(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	cb.Paint(buffer.NewBuffer(20, 5))
}

func TestP182_CodeBlock_StreamingCursorLongNarrow(t *testing.T) {
	cb := NewCodeBlock("go", "x := someVeryLongFunctionName(argument1, argument2)")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 3})
	cb.Paint(buffer.NewBuffer(15, 3))
}

func TestP182_CodeBlock_StreamingCursorZeroBounds(t *testing.T) {
	cb := NewCodeBlock("go", "x := 1")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	cb.Paint(buffer.NewBuffer(1, 1))
}

// === Sparkline valueToBar (77.8% → 100%) ===

func TestP182_Sparkline_AllSameNonZero(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{5, 5, 5, 5, 5})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	sl.Paint(buffer.NewBuffer(20, 3))
}

func TestP182_Sparkline_ZeroAndNegative(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{-10, 0, 10, -5, 5})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	sl.Paint(buffer.NewBuffer(20, 3))
}

// === ContextMenu setCursorLocked (73.3% → 100%) ===

func TestP182_ContextMenu_SetCursorOverflow(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("item1", "Item 1"))
	cm.AddItem(NewMenuItem("item2", "Item 2"))
	cm.AddItem(NewMenuItem("item3", "Item 3"))
	cm.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	cm.SetCursor(10) // overflow
	cm.SetCursor(-1) // underflow
	cm.Paint(buffer.NewBuffer(20, 10))
}

// === ScrollView contentW (75% → 100%) ===

func TestP182_ScrollView_TinyWidth(t *testing.T) {
	sv := NewScrollView(NewParagraph("test content"))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 3})
	sv.Paint(buffer.NewBuffer(1, 3))
}

// === DiffPreview paintBorder (76.5% → 90%+) ===

func TestP182_DiffPreview_PaintBorderTall(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	dp.Paint(buffer.NewBuffer(30, 3))
}

func TestP182_DiffPreview_PaintBorderEmpty(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	dp.Paint(buffer.NewBuffer(30, 3))
}

// === ThemeStudio setCursorLocked (75% → 100%) ===

func TestP182_ThemeStudio_CursorCycle(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	for i := 0; i < 50; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	}
	for i := 0; i < 50; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	}
}

// === RichLog countVisibleLinesLocked (78.6% → 90%+) ===

func TestP182_RichLog_ScrolledWrappedTimeLevels(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	for i := 0; i < 20; i++ {
		rl.Info("a long message that wraps across narrow viewport width for sure")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	rl.ScrollUp(5)
	rl.Paint(buffer.NewBuffer(20, 3))
}

// === AutoComplete Paint (76.7% → 90%+) ===

func TestP182_AutoComplete_PaintEmptyItems(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	ac.Paint(buffer.NewBuffer(30, 10))
}

func TestP182_AutoComplete_PaintWithDescription(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "test1", Description: "desc1"},
		{Label: "test2", Description: "desc2"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	ac.SetQuery("test")
	ac.Paint(buffer.NewBuffer(30, 10))
}

func TestP182_AutoComplete_PaintScrollAll(t *testing.T) {
	ac := NewAutoComplete()
	items := make([]CompletionItem, 20)
	for i := range items {
		items[i] = CompletionItem{Label: "item" + itoa(i)}
	}
	ac.SetItems(items)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	ac.SetQuery("item")
	ac.Paint(buffer.NewBuffer(30, 5))
}

// === StyleBuilder Inherit (71.4% → 100%) ===

func TestP182_StyleBuilder_InheritAllFlags(t *testing.T) {
	parent := NewStyle().Bold().Italic().Underline().Dim().Blink().Reverse().Strikethrough()
	child := NewStyle().Inherit(parent)
	style := child.Style()
	if style.Flags == 0 {
		t.Error("expected inherited flags")
	}
}

func TestP182_StyleBuilder_InheritEmpty(t *testing.T) {
	parent := NewStyle()
	child := NewStyle().Inherit(parent)
	if child.Style().Flags != 0 {
		t.Error("expected zero flags from empty parent")
	}
}

// === StyleBuilder parseLipglossColor (75% → 100%) ===

func TestP182_StyleBuilder_ParseColorHex(t *testing.T) {
	sb := NewStyle()
	sb.ForegroundHex("#ff8800")
	sb.ForegroundHex("#abc")
	sb.ForegroundHex("#XYZ")
	sb.ForegroundHex("")
}

func TestP182_StyleBuilder_ParseColorNamed(t *testing.T) {
	sb := NewStyle()
	sb.ForegroundColor("red")
	sb.ForegroundColor("blue")
	sb.ForegroundColor("nonexistent")
	sb.ForegroundColor("")
}

// === HelpOverlay ensureSelectedVisibleLocked (75% → 90%+) ===

func TestP182_HelpOverlay_ScrollDown(t *testing.T) {
	groups := []HelpGroup{
		{Name: "Global", Entries: []HelpEntry{
			{Keys: "Ctrl+Q", Description: "Quit"},
			{Keys: "Ctrl+S", Description: "Save"},
			{Keys: "Ctrl+O", Description: "Open"},
		}},
	}
	h := NewHelpOverlay(groups)
	h.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2})
	h.ScrollDown(2)
	h.Paint(buffer.NewBuffer(40, 2))
}

// === SessionSidebar Measure (75% → 100%) ===

func TestP182_SessionSidebar_MeasureWithItems(t *testing.T) {
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

// === TextArea moveLine (77.8% → 100%) ===

func TestP182_TextArea_MoveLineEmpty(t *testing.T) {
	ta := NewTextArea()
	ta.moveLine(1)
	ta.moveLine(-1)
}

func TestP182_TextArea_MoveLineSingleLine(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("only one line")
	ta.moveLine(1)
	ta.moveLine(-1)
}

// === LoadingIndicator Start (75% → 100%) ===

func TestP182_LoadingIndicator_DoubleStart(t *testing.T) {
	l := NewLoadingIndicator("test")
	l.Start()
	l.Start()
	l.Stop()
	l.Stop()
}

// === BaseComponent Paint/Measure (0% → 100%) ===

func TestP182_BaseComponent_Paint(t *testing.T) {
	bc := BaseComponent{}
	bc.Paint(buffer.NewBuffer(10, 5))
}

func TestP182_BaseComponent_Measure(t *testing.T) {
	bc := BaseComponent{}
	s := bc.Measure(Constraints{})
	if s.W != 0 || s.H != 0 {
		t.Errorf("expected zero size, got %dx%d", s.W, s.H)
	}
}

func TestP182_BaseComponent_Children(t *testing.T) {
	bc := BaseComponent{}
	if bc.Children() != nil {
		t.Error("expected nil children")
	}
}

// === DiffPreview SetShowLineNumbers/SetShowStats (0% → 100%) ===

func TestP182_DiffPreview_SetShowLineNumbers(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	if !dp.ShowLineNumbers() {
		t.Error("expected line numbers shown")
	}
	dp.SetShowLineNumbers(false)
}

func TestP182_DiffPreview_SetShowStats(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(true)
	dp.SetShowStats(false)
}

// === TextArea bubbles API (0% → 100%) ===

func TestP182_TextArea_SetPrompt(t *testing.T) {
	ta := NewTextArea()
	ta.SetPrompt("> ")
}

func TestP182_TextArea_SetPlaceholder(t *testing.T) {
	ta := NewTextArea()
	ta.SetPlaceholder("type here...")
	if ta.Placeholder() != "" {
		t.Error("expected empty placeholder")
	}
}

func TestP182_TextArea_Focus(t *testing.T) {
	ta := NewTextArea()
	ta.Focus()
}

func TestP182_TextArea_Blur(t *testing.T) {
	ta := NewTextArea()
	ta.Blur()
}

func TestP182_TextArea_SetCharLimit(t *testing.T) {
	ta := NewTextArea()
	ta.SetCharLimit(100)
}

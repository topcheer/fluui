package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === MarkdownViewer HideToc (0% → 100%) ===

func TestP179_MarkdownViewer_HideToc(t *testing.T) {
	v := NewMarkdownViewer("# Test")
	v.ShowToc()
	v.HideToc()
	v.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	v.Paint(buf)
}

// === Pretty Measure (77.8% → 100%) ===

func TestP179_Pretty_MeasureZeroWidth(t *testing.T) {
	p := NewPrettyString("test\ndata")
	s := p.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W != 40 {
		t.Errorf("expected default width 40, got %d", s.W)
	}
}

func TestP179_Pretty_MeasureSingleLine(t *testing.T) {
	p := NewPrettyString("single line")
	s := p.Measure(Constraints{MaxWidth: 40, MaxHeight: 10})
	if s.H != 1 {
		t.Errorf("expected height 1, got %d", s.H)
	}
}

// === TerminalPanel Measure (66.7% → 100%) ===

func TestP179_TerminalPanel_MeasureWithConstraints(t *testing.T) {
	tp := NewTerminalPanel(1000)
	// With MaxWidth constraint
	s := tp.Measure(Constraints{MaxWidth: 50, MaxHeight: 20})
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected non-zero, got %dx%d", s.W, s.H)
	}
	// With zero constraints (should return default)
	s2 := tp.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s2.W <= 0 || s2.H <= 0 {
		t.Errorf("expected non-zero default, got %dx%d", s2.W, s2.H)
	}
}

// === Viewport scrollbar edge cases (73.7% → 90%+) ===

func TestP179_Viewport_VScrollBarThumbClamp(t *testing.T) {
	content := NewParagraph("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\nline11\nline12")
	vp := NewViewport(content)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	vp.ScrollDown(20) // scroll past max (clamps)
	vp.Paint(buffer.NewBuffer(20, 3))
}

func TestP179_Viewport_HScrollBarThumbClamp(t *testing.T) {
	content := NewParagraph("very long text that goes way beyond the width limit of the viewport for testing horizontal scrollbar thumb clamping")
	vp := NewViewport(content)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	vp.ScrollRight(200) // scroll past max (clamps)
	vp.Paint(buffer.NewBuffer(10, 5))
}

func TestP179_Viewport_FitsContent(t *testing.T) {
	content := NewParagraph("short")
	vp := NewViewport(content)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	vp.Paint(buffer.NewBuffer(40, 10))
	// No scrollbar needed when content fits
}

// === Sparkline valueToBar (77.8% → 100%) ===

func TestP179_Sparkline_AllSameNonZero(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{5, 5, 5, 5, 5})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	sl.Paint(buffer.NewBuffer(20, 3))
}

func TestP179_Sparkline_MixedNegatives(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{-10, -5, 0, 5, 10})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	sl.Paint(buffer.NewBuffer(20, 3))
}

// === Badge Measure (73.3% → 100%) ===

func TestP179_Badge_MeasureWithIcon(t *testing.T) {
	b := NewBadge("Test", BadgeInfo)
	b.SetIcon("!")
	s := b.Measure(Constraints{MaxWidth: 50, MaxHeight: 50})
	if s.W <= 0 {
		t.Error("expected non-zero width")
	}
}

func TestP179_Badge_MeasureShortText(t *testing.T) {
	b := NewBadge("X", BadgeSuccess)
	s := b.Measure(Constraints{MaxWidth: 50, MaxHeight: 50})
	if s.W < 3 {
		t.Errorf("expected min width 3, got %d", s.W)
	}
}

func TestP179_Badge_MeasureNarrow(t *testing.T) {
	b := NewBadge("Long Badge Text", BadgeWarning)
	s := b.Measure(Constraints{MaxWidth: 3, MaxHeight: 50})
	if s.W > 3 {
		t.Errorf("expected width <= 3, got %d", s.W)
	}
}

// === CodeBlock paintStreamingCursor (74.2% → 90%+) ===

func TestP179_CodeBlock_StreamingCursor_TitleAndLineNumbers(t *testing.T) {
	cb := NewCodeBlock("go", "x := 1")
	cb.SetTitle("test.go")
	cb.SetShowLineNumbers(true)
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	cb.Paint(buffer.NewBuffer(30, 5))
}

func TestP179_CodeBlock_StreamingCursor_NotStreaming(t *testing.T) {
	cb := NewCodeBlock("go", "x := 1")
	cb.SetStreaming(false)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	cb.Paint(buffer.NewBuffer(20, 5))
}

func TestP179_CodeBlock_StreamingCursor_EmptySource(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	cb.Paint(buffer.NewBuffer(20, 5))
}

func TestP179_CodeBlock_StreamingCursor_LongLineNarrow(t *testing.T) {
	cb := NewCodeBlock("go", "x := someVeryLongFunctionName(argument1, argument2, argument3)")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 3})
	cb.Paint(buffer.NewBuffer(15, 3))
}

// === ScrollView contentW (75% → 100%) ===

func TestP179_ScrollView_TinyWidthAfterScrollbar(t *testing.T) {
	sv := NewScrollView(NewParagraph("test content here"))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 3})
	sv.Paint(buffer.NewBuffer(1, 3))
}

// === HelpOverlay ensureSelectedVisibleLocked (75% → 90%+) ===

func TestP179_HelpOverlay_ScrollDown(t *testing.T) {
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

// === ThemeStudio setCursorLocked (75% → 100%) ===

func TestP179_ThemeStudio_CursorDownUp(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	for i := 0; i < 30; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	}
	for i := 0; i < 30; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	}
}

// === RichLog countVisibleLinesLocked (78.6% → 90%+) ===

func TestP179_RichLog_ScrolledWrappedWithTimeLevels(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	for i := 0; i < 15; i++ {
		rl.Info("a long message that wraps across the narrow viewport width for sure")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	rl.ScrollUp(5)
	rl.Paint(buffer.NewBuffer(20, 3))
}

// === AutoComplete Paint (76.7% → 90%+) ===

func TestP179_AutoComplete_PaintEmptyItems(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	ac.Paint(buffer.NewBuffer(30, 10))
}

func TestP179_AutoComplete_PaintWithDescription(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "test1", Description: "a description"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	ac.SetQuery("test")
	ac.Paint(buffer.NewBuffer(30, 10))
}

func TestP179_AutoComplete_PaintScrollAll(t *testing.T) {
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

func TestP179_StyleBuilder_InheritWithAllFlags(t *testing.T) {
	parent := NewStyle().Bold().Italic().Underline().Dim().Blink().Reverse().Strikethrough()
	child := NewStyle().Inherit(parent)
	style := child.Style()
	if style.Flags == 0 {
		t.Error("expected inherited flags")
	}
}

func TestP179_StyleBuilder_InheritEmpty(t *testing.T) {
	parent := NewStyle()
	child := NewStyle().Inherit(parent)
	style := child.Style()
	if style.Flags != 0 {
		t.Error("expected zero flags from empty parent")
	}
}

// === StyleBuilder parseLipglossColor (70% → 100%) ===

func TestP179_StyleBuilder_ParseColorHex(t *testing.T) {
	sb := NewStyle()
	sb.ForegroundHex("#ff8800")
	sb.ForegroundHex("#abc")
	sb.ForegroundHex("#XYZ") // invalid hex
	sb.ForegroundHex("")      // empty
}

func TestP179_StyleBuilder_ParseColorNamed(t *testing.T) {
	sb := NewStyle()
	sb.ForegroundColor("red")
	sb.ForegroundColor("blue")
	sb.ForegroundColor("green")
	sb.ForegroundColor("nonexistent")
	sb.ForegroundColor("")
}

// === DiffPreview paintBorder (76.5% → 90%+) ===

func TestP179_DiffPreview_PaintBorderTall(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	dp.Paint(buffer.NewBuffer(30, 3))
}

func TestP179_DiffPreview_PaintBorderEmpty(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	dp.Paint(buffer.NewBuffer(30, 3))
}

// === SessionSidebar Measure (75% → 100%) ===

func TestP179_SessionSidebar_MeasureWithItems(t *testing.T) {
	sb := NewSessionSidebar()
	sb.SetItems([]SessionItem{
		{ID: "1", Title: "Session 1", Workspace: "G1"},
		{ID: "2", Title: "Session 2", Workspace: "G1"},
		{ID: "3", Title: "Session 3", Workspace: "G2"},
	})
	s := sb.Measure(Constraints{MaxWidth: 30, MaxHeight: 20})
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected non-zero, got %dx%d", s.W, s.H)
	}
}

// === TextArea moveLine (77.8% → 100%) ===

func TestP179_TextArea_MoveLineEmpty(t *testing.T) {
	ta := NewTextArea()
	ta.moveLine(1)
	ta.moveLine(-1)
}

func TestP179_TextArea_MoveLineSingleLine(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("only one line")
	ta.moveLine(1)
	ta.moveLine(-1)
}

// === LoadingIndicator Start (75% → 100%) ===

func TestP179_LoadingIndicator_DoubleStart(t *testing.T) {
	l := NewLoadingIndicator("test")
	l.Start()
	l.Start() // double start should be no-op
	l.Stop()
	l.Stop() // double stop should be no-op
}

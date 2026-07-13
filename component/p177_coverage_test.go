package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === TerminalPanel SGR parsing ===

func TestP177_TerminalPanel_SGR_Reset(t *testing.T) {
	tp := NewTerminalPanel(1000)
	tp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	// Set some attributes then reset
	tp.HandleKey(&term.KeyEvent{Rune: 'x'})
	// Write SGR reset via Write method (if available) or test parseSGRLocked directly
	tp.mu.Lock()
	tp.parseSGRLocked([]byte("0"))
	tp.mu.Unlock()
	if tp.curFlags != 0 {
		t.Error("expected flags reset to 0")
	}
}

func TestP177_TerminalPanel_SGR_Bold(t *testing.T) {
	tp := NewTerminalPanel(1000)
	tp.mu.Lock()
	tp.parseSGRLocked([]byte("1"))
	if tp.curFlags&buffer.Bold == 0 {
		t.Error("expected Bold flag set")
	}
	tp.mu.Unlock()
}

func TestP177_TerminalPanel_SGR_AllAttributes(t *testing.T) {
	tp := NewTerminalPanel(1000)
	tp.mu.Lock()
	// Set all attributes
	tp.parseSGRLocked([]byte("1;2;3;4;5;7;9"))
	flags := tp.curFlags
	if flags&buffer.Bold == 0 {
		t.Error("missing Bold")
	}
	if flags&buffer.Dim == 0 {
		t.Error("missing Dim")
	}
	if flags&buffer.Italic == 0 {
		t.Error("missing Italic")
	}
	if flags&buffer.Underline == 0 {
		t.Error("missing Underline")
	}
	if flags&buffer.Blink == 0 {
		t.Error("missing Blink")
	}
	if flags&buffer.Reverse == 0 {
		t.Error("missing Reverse")
	}
	if flags&buffer.Strikethrough == 0 {
		t.Error("missing Strikethrough")
	}
	tp.mu.Unlock()
}

func TestP177_TerminalPanel_SGR_UnsetAttributes(t *testing.T) {
	tp := NewTerminalPanel(1000)
	tp.mu.Lock()
	tp.parseSGRLocked([]byte("1;2;3;4;5;7;9"))
	// Unset all
	tp.parseSGRLocked([]byte("22;23;24;25;27;29"))
	if tp.curFlags != 0 {
		t.Error("expected all flags unset")
	}
	tp.mu.Unlock()
}

func TestP177_TerminalPanel_SGR_ForegroundColors(t *testing.T) {
	tp := NewTerminalPanel(1000)
	tp.mu.Lock()
	// Named colors 30-37
	for i := 30; i <= 37; i++ {
		tp.parseSGRLocked([]byte(itoa(i)))
	}
	// Bright named 90-97
	for i := 90; i <= 97; i++ {
		tp.parseSGRLocked([]byte(itoa(i)))
	}
	// Default
	tp.parseSGRLocked([]byte("39"))
	tp.mu.Unlock()
}

func TestP177_TerminalPanel_SGR_BackgroundColors(t *testing.T) {
	tp := NewTerminalPanel(1000)
	tp.mu.Lock()
	// Named colors 40-47
	for i := 40; i <= 47; i++ {
		tp.parseSGRLocked([]byte(itoa(i)))
	}
	// Bright named 100-107
	for i := 100; i <= 107; i++ {
		tp.parseSGRLocked([]byte(itoa(i)))
	}
	// Default
	tp.parseSGRLocked([]byte("49"))
	tp.mu.Unlock()
}

func TestP177_TerminalPanel_SGR_256Color(t *testing.T) {
	tp := NewTerminalPanel(1000)
	tp.mu.Lock()
	tp.parseSGRLocked([]byte("38;5;196"))
	tp.mu.Unlock()
}

func TestP177_TerminalPanel_SGR_TrueColor(t *testing.T) {
	tp := NewTerminalPanel(1000)
	tp.mu.Lock()
	tp.parseSGRLocked([]byte("38;2;255;128;64"))
	tp.parseSGRLocked([]byte("48;2;0;0;255"))
	tp.mu.Unlock()
}

func TestP177_TerminalPanel_SGR_ExtendedColorDefault(t *testing.T) {
	tp := NewTerminalPanel(1000)
	tp.mu.Lock()
	// Extended color with invalid sub-type
	consumed, c := tp.parseExtendedColorLocked([]int{99})
	if consumed != 1 {
		t.Errorf("expected consumed 1, got %d", consumed)
	}
	_ = c
	tp.mu.Unlock()
}

func TestP177_TerminalPanel_SGR_256ColorShort(t *testing.T) {
	tp := NewTerminalPanel(1000)
	tp.mu.Lock()
	// 256-color with missing value
	consumed, _ := tp.parseExtendedColorLocked([]int{5})
	if consumed != 1 {
		t.Errorf("expected consumed 1, got %d", consumed)
	}
	tp.mu.Unlock()
}

func TestP177_TerminalPanel_SGR_TrueColorShort(t *testing.T) {
	tp := NewTerminalPanel(1000)
	tp.mu.Lock()
	// Truecolor with missing values
	consumed, _ := tp.parseExtendedColorLocked([]int{2, 255})
	if consumed != 1 {
		t.Errorf("expected consumed 1, got %d", consumed)
	}
	tp.mu.Unlock()
}

func TestP177_TerminalPanel_SGR_EmptyParams(t *testing.T) {
	tp := NewTerminalPanel(1000)
	tp.mu.Lock()
	tp.parseSGRLocked([]byte(""))
	tp.mu.Unlock()
}

func TestP177_TerminalPanel_SGR_InvalidParam(t *testing.T) {
	tp := NewTerminalPanel(1000)
	tp.mu.Lock()
	tp.parseSGRLocked([]byte("abc;1"))
	if tp.curFlags&buffer.Bold == 0 {
		t.Error("expected Bold after invalid+valid params")
	}
	tp.mu.Unlock()
}

func TestP177_TerminalPanel_parseEscapeLocked(t *testing.T) {
	tp := NewTerminalPanel(1000)
	tp.mu.Lock()
	// Short data
	n := tp.parseEscapeLocked([]byte{0x1b})
	if n != 1 {
		t.Errorf("expected 1, got %d", n)
	}
	// Unknown escape (2 bytes consumed)
	n = tp.parseEscapeLocked([]byte{0x1b, 'x'})
	if n != 2 {
		t.Errorf("expected 2, got %d", n)
	}
	// OSC sequence
	n = tp.parseEscapeLocked([]byte{0x1b, ']', '0', ';', 't', 0x07})
	_ = n
	tp.mu.Unlock()
}

func TestP177_TerminalPanel_MeasureZero(t *testing.T) {
	tp := NewTerminalPanel(1000)
	s := tp.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected non-zero, got %dx%d", s.W, s.H)
	}
}

func TestP177_TerminalPanel_HandleKeySpecial(t *testing.T) {
	tp := NewTerminalPanel(1000)
	tp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	// PageUp/PageDown
	tp.HandleKey(&term.KeyEvent{Key: term.KeyPageUp})
	tp.HandleKey(&term.KeyEvent{Key: term.KeyPageDown})
	// Home/End
	tp.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	tp.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	// Escape
	tp.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
}

// === Viewport scrollbar edge cases ===

func TestP177_Viewport_VScrollBarNearBottom(t *testing.T) {
	vp := NewViewport(NewParagraph("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10"))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	vp.ScrollDown(8) // near bottom
	vp.Paint(buffer.NewBuffer(20, 5))
}

func TestP177_Viewport_HScrollBarNearRight(t *testing.T) {
	vp := NewViewport(NewParagraph("a very long line that exceeds width and needs horizontal scrolling"))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	vp.ScrollRight(50) // near right
	vp.Paint(buffer.NewBuffer(10, 5))
}

func TestP177_Viewport_BothScrollbars(t *testing.T) {
	content := NewParagraph("long line 1 that wraps\nline 2\nline 3\nline 4\nline 5\nline 6\nline 7\nline 8\nline 9\nline 10")
	vp := NewViewport(content)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	vp.ScrollDown(2)
	vp.ScrollRight(3)
	vp.Paint(buffer.NewBuffer(10, 3))
}

// === Sparkline all-same ===

func TestP177_Sparkline_AllSameZero(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{0, 0, 0, 0, 0})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	sl.Paint(buffer.NewBuffer(20, 3))
}

func TestP177_Sparkline_NegativeValues(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{-5, -3, 0, 3, 5})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	sl.Paint(buffer.NewBuffer(20, 3))
}

// === RichLog wrapped+scrolled ===

func TestP177_RichLog_ScrolledWrapped(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	for i := 0; i < 20; i++ {
		rl.Info("message that is long enough to wrap across the narrow width")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	rl.ScrollUp(10)
	rl.Paint(buffer.NewBuffer(20, 5))
}

// === CodeBlock streaming cursor edge cases ===

func TestP177_CodeBlock_StreamingCursorEmpty(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	cb.SetTitle("empty.go")
	cb.Paint(buffer.NewBuffer(20, 5))
}

func TestP177_CodeBlock_StreamingCursorLongLine(t *testing.T) {
	cb := NewCodeBlock("go", "x := someVeryLongFunctionCall(that, takes, many, arguments, here)")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 3})
	cb.SetStreaming(true)
	cb.Paint(buffer.NewBuffer(15, 3))
}

// === AutoComplete Paint edge cases ===

func TestP177_AutoComplete_PaintScrollDown(t *testing.T) {
	ac := NewAutoComplete()
	items := make([]CompletionItem, 15)
	for i := range items {
		items[i] = CompletionItem{Label: "item" + itoa(i), Description: "desc"}
	}
	ac.SetItems(items)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	ac.SetQuery("item")
	ac.Paint(buffer.NewBuffer(30, 5))
}

func TestP177_AutoComplete_PaintWithCategory(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "test1", Description: "desc1", Category: "Functions"},
		{Label: "test2", Description: "desc2", Category: "Variables"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	ac.SetQuery("test")
	ac.Paint(buffer.NewBuffer(30, 10))
}

// === Badge Measure edge cases ===

func TestP177_Badge_MeasureAllVariants(t *testing.T) {
	variants := []BadgeVariant{BadgeInfo, BadgeSuccess, BadgeWarning, BadgeError}
	for _, v := range variants {
		b := NewBadge("Test", v)
		b.SetIcon("!")
		s := b.Measure(Constraints{MaxWidth: 50, MaxHeight: 50})
		if s.W <= 0 || s.H <= 0 {
			t.Errorf("variant %d: expected non-zero size", v)
		}
		// Narrow
		s2 := b.Measure(Constraints{MaxWidth: 2, MaxHeight: 50})
		if s2.W > 2 {
			t.Errorf("variant %d: expected width <= 2", v)
		}
	}
}

// === DiffPreview paintBorder ===

func TestP177_DiffPreview_PaintBorderEmpty(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	dp.Paint(buffer.NewBuffer(30, 3))
}

func TestP177_DiffPreview_PaintBorderTall(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	dp.Paint(buffer.NewBuffer(30, 3))
}

// === ScrollView narrow ===

func TestP177_ScrollView_TinyWidth(t *testing.T) {
	sv := NewScrollView(NewParagraph("test"))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 3})
	sv.Paint(buffer.NewBuffer(1, 3))
}

// === HelpOverlay scroll ===

func TestP177_HelpOverlay_Scroll(t *testing.T) {
	groups := []HelpGroup{
		{Name: "Global", Entries: []HelpEntry{
			{Keys: "Ctrl+Q", Description: "Quit"},
			{Keys: "Ctrl+S", Description: "Save"},
			{Keys: "Ctrl+O", Description: "Open"},
			{Keys: "Ctrl+X", Description: "Cut"},
			{Keys: "Ctrl+V", Description: "Paste"},
		}},
	}
	h := NewHelpOverlay(groups)
	h.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	h.ScrollDown(2)
	h.Paint(buffer.NewBuffer(40, 3))
}

// === ThemeStudio cursor navigation ===

func TestP177_ThemeStudio_CursorCycle(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	for i := 0; i < 50; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	}
	for i := 0; i < 50; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	}
}

// === StyleBuilder Inherit ===

func TestP177_StyleBuilder_InheritFlags(t *testing.T) {
	parent := NewStyle().Bold().Italic()
	child := NewStyle().Inherit(parent)
	style := child.Style()
	if style.Flags&buffer.Bold == 0 {
		t.Error("expected inherited Bold")
	}
	if style.Flags&buffer.Italic == 0 {
		t.Error("expected inherited Italic")
	}
}

func TestP177_StyleBuilder_parseLipglossColor(t *testing.T) {
	// Test hex colors
	sb := NewStyle()
	sb.ForegroundHex("#ff8800")
	sb.ForegroundHex("#abc") // short hex
	sb.ForegroundHex("#f00") // short hex
	sb.ForegroundHex("invalid") // invalid
	sb.ForegroundColor("blue") // named
	sb.ForegroundColor("red")
	sb.ForegroundColor("nonexistent") // unknown named
}

// === TextArea moveLine edge cases ===

func TestP177_TextArea_MoveLineSingleLine(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("single line")
	ta.moveLine(1) // down — should be no-op
	ta.moveLine(-1) // up — should be no-op
}

func TestP177_TextArea_MoveLineEmpty(t *testing.T) {
	ta := NewTextArea()
	ta.moveLine(1)
	ta.moveLine(-1)
}

// === SessionSidebar Measure ===

func TestP177_SessionSidebar_Measure(t *testing.T) {
	sb := NewSessionSidebar()
	sb.SetItems([]SessionItem{
		{ID: "1", Title: "S1", Workspace: "G1"},
	})
	s := sb.Measure(Constraints{MaxWidth: 30, MaxHeight: 20})
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected non-zero, got %dx%d", s.W, s.H)
	}
}

// === LoadingIndicator Start coverage ===

func TestP177_LoadingIndicator_StartStopCycle(t *testing.T) {
	l := NewLoadingIndicator("test")
	l.Start()
	l.Start() // double start = no-op
	l.Stop()
	l.Stop() // double stop = no-op
}

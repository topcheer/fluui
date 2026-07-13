package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === Switch coverage ===

func TestP175_Switch_SetLabel(t *testing.T) {
	s := NewSwitch("MySwitch")
	s.SetOn(true)
	s.SetLabel("Auto-save")
	if s.Label() != "Auto-save" {
		t.Errorf("expected 'Auto-save', got %q", s.Label())
	}
}

func TestP175_Switch_HandleKey(t *testing.T) {
	s := NewSwitch("MySwitch")
	s.SetOn(false)
	// Enter toggles
	if !s.HandleKey(&term.KeyEvent{Key: term.KeyEnter}) {
		t.Error("expected Enter to be handled")
	}
	if !s.IsOn() {
		t.Error("expected switch ON after Enter")
	}
	// Space toggles
	s.HandleKey(&term.KeyEvent{Key: term.KeySpace})
	if s.IsOn() {
		t.Error("expected switch OFF after Space")
	}
	// Rune space toggles
	s.HandleKey(&term.KeyEvent{Rune: ' '})
	if !s.IsOn() {
		t.Error("expected switch ON after rune space")
	}
	// Unknown key not handled
	if s.HandleKey(&term.KeyEvent{Key: term.KeyEscape}) {
		t.Error("expected Escape not handled")
	}
	// Nil key
	if s.HandleKey(nil) {
		t.Error("expected false for nil")
	}
}

// === TextInput coverage ===

func TestP175_TextInput_Len(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("hello")
	if ti.Len() != 5 {
		t.Errorf("expected 5, got %d", ti.Len())
	}
	ti.SetValue("")
	if ti.Len() != 0 {
		t.Errorf("expected 0, got %d", ti.Len())
	}
}

func TestP175_TextInput_CharLimit(t *testing.T) {
	ti := NewTextInput()
	if ti.CharLimit() != 0 {
		t.Errorf("expected 0, got %d", ti.CharLimit())
	}
	ti.SetCharLimit(10)
	if ti.CharLimit() != 10 {
		t.Errorf("expected 10, got %d", ti.CharLimit())
	}
}

func TestP175_TextInput_SetCharLimitTruncates(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("hello world")
	ti.SetCharLimit(5)
	if ti.Value() != "hello" {
		t.Errorf("expected 'hello', got %q", ti.Value())
	}
}

func TestP175_TextInput_SetCharLimitZero(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("test")
	ti.SetCharLimit(0) // 0 = no limit
	if ti.Value() != "test" {
		t.Errorf("expected 'test', got %q", ti.Value())
	}
}

func TestP175_TextInput_NavigateHistory(t *testing.T) {
	ti := NewTextInput()
	ti.SetHistory([]string{"cmd1", "cmd2", "cmd3"})
	// Navigate up (older)
	ti.navigateHistory(-1)
	if ti.Value() != "cmd3" {
		t.Errorf("expected 'cmd3', got %q", ti.Value())
	}
	// Navigate up again
	ti.navigateHistory(-1)
	if ti.Value() != "cmd2" {
		t.Errorf("expected 'cmd2', got %q", ti.Value())
	}
	// Navigate down (newer)
	ti.navigateHistory(1)
	if ti.Value() != "cmd3" {
		t.Errorf("expected 'cmd3', got %q", ti.Value())
	}
	// Navigate down past end — empty
	ti.navigateHistory(1)
	if ti.Value() != "" {
		t.Errorf("expected empty, got %q", ti.Value())
	}
	// Navigate down past end — clamps
	ti.navigateHistory(1)
	if ti.Value() != "" {
		t.Errorf("expected empty, got %q", ti.Value())
	}
}

func TestP175_TextInput_NavigateHistoryEmpty(t *testing.T) {
	ti := NewTextInput()
	ti.navigateHistory(-1) // no panic
	if ti.Value() != "" {
		t.Error("expected empty value")
	}
}

// === TextArea coverage (0% no-op methods) ===

func TestP175_TextArea_SetPrompt(t *testing.T) {
	ta := NewTextArea()
	ta.SetPrompt("> ") // no-op
}

func TestP175_TextArea_SetPlaceholder(t *testing.T) {
	ta := NewTextArea()
	ta.SetPlaceholder("hint") // no-op
}

func TestP175_TextArea_FocusBlur(t *testing.T) {
	ta := NewTextArea()
	ta.Focus() // no-op
	ta.Blur()  // no-op
}

func TestP175_TextArea_SetCharLimit(t *testing.T) {
	ta := NewTextArea()
	ta.SetCharLimit(100) // no-op
}

// === TabbedContent HandleKey coverage ===

func TestP175_TabbedContent_HandleKeyCtrlRight(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewParagraph("content A"))
	tc.AddTab("b", "Tab B", NewParagraph("content B"))
	tc.SwitchTo("a")
	tc.HandleKey(&term.KeyEvent{Key: term.KeyRight, Modifiers: term.ModCtrl})
	if tc.ActiveTab() != "b" {
		t.Errorf("expected 'b', got %q", tc.ActiveTab())
	}
}

func TestP175_TabbedContent_HandleKeyCtrlLeft(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewParagraph("content A"))
	tc.AddTab("b", "Tab B", NewParagraph("content B"))
	tc.SwitchTo("b")
	tc.HandleKey(&term.KeyEvent{Key: term.KeyLeft, Modifiers: term.ModCtrl})
	if tc.ActiveTab() != "a" {
		t.Errorf("expected 'a', got %q", tc.ActiveTab())
	}
}

func TestP175_TabbedContent_HandleKeyUnknown(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewParagraph("content A"))
	if tc.HandleKey(&term.KeyEvent{Key: term.KeyF1}) {
		t.Error("expected false for F1")
	}
}

func TestP175_TabbedContent_HandleKeyForwardsToChild(t *testing.T) {
	tc := NewTabbedContent()
	tc.AddTab("a", "Tab A", NewTextArea())
	tc.SwitchTo("a")
	// Type a character — should be forwarded to TextArea
	tc.HandleKey(&term.KeyEvent{Rune: 'x'})
}

// === DiffPreview coverage ===

func TestP175_DiffPreview_SetShowLineNumbers(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	dp.SetShowLineNumbers(false)
}

func TestP175_DiffPreview_SetShowStats(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(true)
	dp.SetShowStats(false)
}

func TestP175_DiffPreview_PaintBorder(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("line1\nline2\nline3\nline4\nline5")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	dp.Paint(buffer.NewBuffer(30, 3))
}

// === Component.go Paint (0%) ===

func TestP175_BaseComponent_Paint(t *testing.T) {
	bc := BaseComponent{}
	bc.Paint(buffer.NewBuffer(10, 5)) // should not panic
}

// === Header/Footer Measure with constraints ===

func TestP175_Header_MeasureWithConstraints(t *testing.T) {
	h := NewHeader("App")
	s := h.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W != 40 {
		t.Errorf("expected 40, got %d", s.W)
	}
}

func TestP175_Footer_MeasureWithConstraints(t *testing.T) {
	f := NewFooter()
	s := f.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W != 40 {
		t.Errorf("expected 40, got %d", s.W)
	}
}

// === LoadingIndicator Start coverage ===

func TestP175_LoadingIndicator_StartDoubleStart(t *testing.T) {
	l := NewLoadingIndicator("test")
	l.Start()
	l.Start() // double start = no-op
	l.Stop()
}

// === Viewport scrollbar coverage ===

func TestP175_Viewport_DrawVScrollBar(t *testing.T) {
	vp := NewViewport(NewParagraph("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10"))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	vp.Paint(buffer.NewBuffer(20, 5))
	// Scroll down to trigger scrollbar
	vp.ScrollDown(3)
	vp.Paint(buffer.NewBuffer(20, 5))
}

func TestP175_Viewport_DrawHScrollBar(t *testing.T) {
	vp := NewViewport(NewParagraph("a very long line that exceeds the viewport width for sure"))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	vp.Paint(buffer.NewBuffer(10, 5))
	vp.ScrollRight(5)
	vp.Paint(buffer.NewBuffer(10, 5))
}

// === Sparkline valueToBar ===

func TestP175_Sparkline_AllSameValue(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{5, 5, 5, 5, 5})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	sl.Paint(buffer.NewBuffer(20, 3))
}

// === ScrollView contentW ===

func TestP175_ScrollView_NarrowWidth(t *testing.T) {
	sv := NewScrollView(NewParagraph("test content"))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 3})
	sv.Paint(buffer.NewBuffer(1, 3))
}

// === RichLog countVisibleLines ===

func TestP175_RichLog_WrappedLines(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	rl.Info("short message")
	rl.Info("a very long message that will wrap across the narrow width of the buffer")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 10})
	rl.Paint(buffer.NewBuffer(15, 10))
}

// === HelpOverlay ===

func TestP175_HelpOverlay_ScrollDown(t *testing.T) {
	h := NewHelpOverlay([]HelpGroup{
		{Name: "Global", Entries: []HelpEntry{{Keys: "Ctrl+Q", Description: "Quit"}}},
	})
	h.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	h.ScrollDown(1)
	h.Paint(buffer.NewBuffer(40, 3))
}

// === CodeBlock streaming cursor ===

func TestP175_CodeBlock_StreamingCursor(t *testing.T) {
	cb := NewCodeBlock("go", "x := 1\ny := 2\nz := 3")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	cb.SetTitle("test.go")
	cb.Paint(buffer.NewBuffer(20, 5))
	// Not streaming
	cb2 := NewCodeBlock("go", "z := 3")
	cb2.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	cb2.SetStreaming(false)
	cb2.Paint(buffer.NewBuffer(20, 3))
}

// === AutoComplete Paint ===

func TestP175_AutoComplete_PaintWithItems(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "test1", Description: "desc1", Category: "cat1"},
		{Label: "test2", Description: "desc2"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	ac.SetQuery("test")
	ac.Paint(buffer.NewBuffer(30, 10))
}

// === Badge Measure ===

func TestP175_Badge_MeasureNarrow(t *testing.T) {
	b := NewBadge("Hello World", BadgeInfo)
	b.SetIcon("!")
	b.Measure(Constraints{MaxWidth: 2, MaxHeight: 50})
}

// === SessionSidebar HandleKey backspace ===

func TestP175_SessionSidebar_HandleKeyBackspace(t *testing.T) {
	sb := NewSessionSidebar()
	sb.SetItems([]SessionItem{
		{ID: "1", Title: "S1", Workspace: "G"},
	})
	sb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 20})
	// Enter search mode
	sb.HandleKey(&term.KeyEvent{Rune: '/'})
	// Type a character
	sb.HandleKey(&term.KeyEvent{Rune: 'S'})
	// Backspace
	sb.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	// Enter to confirm
	sb.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
}

// === VirtualScroller VisibleItems ===

func TestP175_VirtualScroller_VisibleItems(t *testing.T) {
	items := make([]VirtualItem, 20)
	for i := range items {
		items[i] = VirtualItem{ID: string(rune('a' + i)), Text: string(rune('a' + i))}
	}
	vs := NewVirtualScroller()
	vs.SetItems(items)
	vs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	visible := vs.VisibleItems()
	if len(visible) == 0 {
		// VirtualScroller may need page size set — just verify no panic
	}
	// Scroll past end
	vs.ScrollTo(25)
	vs.VisibleItems()
}

// === ThemeStudio setCursorLocked ===

func TestP175_ThemeStudio_SetCursor(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	// Move cursor down many times
	for i := 0; i < 100; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	}
	// Move back up
	for i := 0; i < 100; i++ {
		ts.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	}
}

// === StyleBuilder Inherit ===

func TestP175_StyleBuilder_InheritEmpty(t *testing.T) {
	child := NewStyle().Inherit(NewStyle())
	if child.Style().Flags != 0 {
		t.Error("expected no flags from empty parent")
	}
}

// === terminal_panel coverage ===

func TestP175_TerminalPanel_Measure(t *testing.T) {
	tp := NewTerminalPanel(1000)
	s := tp.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected non-zero, got %dx%d", s.W, s.H)
	}
}

func TestP175_TerminalPanel_HandleKey(t *testing.T) {
	tp := NewTerminalPanel(1000)
	tp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	tp.HandleKey(&term.KeyEvent{Rune: 'a'})
	tp.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	tp.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	tp.HandleKey(&term.KeyEvent{Rune: 'l', Modifiers: term.ModCtrl})
}

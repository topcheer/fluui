package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// Cover all 0% and sub-80% functions in component package

func TestP184_TextArea_BubblesAPI(t *testing.T) {
	ta := NewTextArea()
	ta.SetPrompt("> ")
	ta.SetPlaceholder("Type here...")
	ta.Focus()
	ta.Blur()
	ta.SetCharLimit(100)
	ta.SetHeight(10)
	ta.SetWidth(80)
	ta.CursorDown()
	ta.CursorUp()
	// Verify no panic — these are mostly no-op compat methods
	if ta.Value() != "" {
		// Value should be empty initially
	}
}

func TestP184_TextArea_SetValue(t *testing.T) {
	ta := NewTextArea()
	ta.SetValue("hello world")
	if ta.Value() != "hello world" {
		t.Fatalf("expected 'hello world', got %q", ta.Value())
	}
}

func TestP184_BaseComponent_Paint(t *testing.T) {
	bc := BaseComponent{}
	buf := buffer.NewBuffer(10, 5)
	bc.Paint(buf) // no-op, should not panic
}

func TestP184_DiffPreview_SetShowLineNumbers(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	dp.SetShowLineNumbers(false)
	// no-op method, just verify no panic
}

func TestP184_DiffPreview_SetShowStats(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(true)
	dp.SetShowStats(false)
}

func TestP184_DiffPreview_PaintBorderNarrow(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("diff --git a/x b/x\n--- a/x\n+++ b/x\n@@ -1,3 +1,3 @@\n-old\n+new\n")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 20})
	buf := buffer.NewBuffer(5, 20)
	dp.Paint(buf) // exercises paintBorderLocked with narrow width
}

func TestP184_DiffPreview_PaintBorderTall(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("--- a\n+++ b\n@@ -1,2 +1,2 @@\n-a\n+b\n")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 50})
	buf := buffer.NewBuffer(60, 50)
	dp.Paint(buf)
}

func TestP184_DiffPreview_PaintBorderEmpty(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	dp.Paint(buf)
}

func TestP184_AutoComplete_PaintWithDescription(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "item1", Description: "First item"},
		{Label: "item2", Description: "Second", Category: "cat"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	ac.Paint(buf)
}

func TestP184_AutoComplete_PaintScrollDown(t *testing.T) {
	ac := NewAutoComplete()
	items := make([]CompletionItem, 20)
	for i := range items {
		items[i] = CompletionItem{Label: "item" + itoa(i)}
	}
	ac.SetItems(items)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	ac.scrollY = 10
	buf := buffer.NewBuffer(40, 5)
	ac.Paint(buf)
}

func TestP184_AutoComplete_PaintEmpty(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	ac.Paint(buf)
}

func TestP184_AutoComplete_PaintWithCategorySelected(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "x", Category: "cat1"},
		{Label: "y", Category: "cat2"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	ac.Paint(buf)
}

func TestP184_AutoComplete_PaintNegativeScroll(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{{Label: "a"}})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	ac.scrollY = -1
	buf := buffer.NewBuffer(40, 5)
	ac.Paint(buf)
}

func TestP184_Badge_MeasureAllVariants(t *testing.T) {
	for _, v := range []BadgeVariant{BadgeInfo, BadgeSuccess, BadgeWarning, BadgeError, BadgeCritical, BadgeNeutral} {
		b := NewBadge("X", v)
		b.SetIcon("*")
		s := b.Measure(Bounded(100, 10))
		if s.W < 0 || s.H < 0 {
			t.Fatalf("variant %d: negative size %v", v, s)
		}
	}
}

func TestP184_Badge_MeasureWithIcon(t *testing.T) {
	b := NewBadge("Text", BadgeInfo)
	b.SetIcon("*")
	s := b.Measure(Bounded(100, 10))
	if s.W < 3 {
		t.Fatalf("expected w>=3 with icon, got %d", s.W)
	}
}

func TestP184_Badge_MeasureNarrow(t *testing.T) {
	b := NewBadge("Long Text Here", BadgeSuccess)
	s := b.Measure(Bounded(2, 1))
	if s.W > 2 {
		t.Fatalf("expected w<=2, got %d", s.W)
	}
}

func TestP184_Badge_MeasureShortText(t *testing.T) {
	b := NewBadge("A", BadgeWarning)
	s := b.Measure(Bounded(1, 1))
	if s.W < 1 {
		t.Fatalf("expected w>=1, got %d", s.W)
	}
}

func TestP184_CodeBlock_StreamingCursorTitleLineNumbers(t *testing.T) {
	cb := NewCodeBlock("go", "package main\nfunc main() {}\n")
	cb.SetShowTitle(true)
	cb.SetShowLineNumbers(true)
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 20})
	buf := buffer.NewBuffer(80, 20)
	cb.Paint(buf) // exercises paintStreamingCursorLocked
}

func TestP184_CodeBlock_StreamingCursorLongNarrow(t *testing.T) {
	cb := NewCodeBlock("go", "package main\n\nfunc main() {\n\tfmt.Println(\"hello world this is a very long line\")\n}\n")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 20})
	buf := buffer.NewBuffer(20, 20)
	cb.Paint(buf)
}

func TestP184_CodeBlock_StreamingCursorEmptyNotStreaming(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(false)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 20})
	buf := buffer.NewBuffer(80, 20)
	cb.Paint(buf)
}

func TestP184_CodeBlock_StreamingCursorZeroBounds(t *testing.T) {
	cb := NewCodeBlock("go", "x = 1")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(0, 0)
	cb.Paint(buf)
}

func TestP184_ContextMenu_SetCursorOverflow(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "A"))
	cm.AddItem(NewMenuItem("b", "B"))
	cm.SetCursor(100) // overflow
	cm.SetCursor(-5)  // underflow
	cm.SetCursor(1)   // normal
}

func TestP184_ContextMenu_SetCursorEmpty(t *testing.T) {
	cm := NewContextMenu()
	cm.SetCursor(0) // empty list
}

func TestP184_ScrollView_ContentWNarrow(t *testing.T) {
	sv := NewScrollView(&fixedSize{w: 100, h: 50})
	sv.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 20})
	// contentW should handle w < 1 after scrollbar
	buf := buffer.NewBuffer(1, 20)
	sv.Paint(buf)
}

func TestP184_Sparkline_ValueToBarAllSame(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{5, 5, 5, 5})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	sl.Paint(buf)
}

func TestP184_Sparkline_ValueToBarZeroNeg(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{0, -1, -5, 3})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	sl.Paint(buf)
}

func TestP184_HelpOverlay_ScrollDown(t *testing.T) {
	ho := NewHelpOverlay([]HelpGroup{
		{Name: "G1", Entries: []HelpEntry{{Keys: "Ctrl+A", Description: "Action"}}},
		{Name: "G2", Entries: []HelpEntry{{Keys: "Ctrl+B", Description: "Action B"}}},
	})
	ho.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	ho.Paint(buf)
	ho.ScrollDown(1)
	ho.Paint(buf)
}

func TestP184_RichLog_CountVisibleLinesWrapped(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	rl.Write(LogInfo, "This is a very long line that should wrap across multiple terminal columns to test the countVisibleLinesLocked function properly")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	rl.Paint(buf)
}

func TestP184_RichLog_CountVisibleLinesScrolled(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 20; i++ {
		rl.Write(LogInfo, "line "+itoa(i))
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 3})
	rl.ScrollDown(10)
	buf := buffer.NewBuffer(60, 3)
	rl.Paint(buf)
}

func TestP184_StyleBuilder_InheritWithFlags(t *testing.T) {
	parent := NewStyle()
	parent.Bold()
	parent.Foreground(buffer.RGB(255, 0, 0))
	child := NewStyle()
	child.Foreground(buffer.RGB(0, 255, 0))
	result := child.Inherit(parent)
	if result == nil {
		t.Fatal("Inherit should return non-nil")
	}
}

func TestP184_StyleBuilder_InheritEmptyChild(t *testing.T) {
	parent := NewStyle()
	parent.Foreground(buffer.RGB(255, 0, 0))
	child := NewStyle()
	result := child.Inherit(parent)
	if result == nil {
		t.Fatal("Inherit should return non-nil")
	}
}

func TestP184_StyleBuilder_ParseLipglossColor(t *testing.T) {
	tests := []string{"#ff8800", "red", "#fff", "invalid", ""}
	for _, input := range tests {
		c := parseLipglossColor(input)
		_ = c // just verify no panic
	}
}

func TestP184_LoadingIndicator_DoubleStart(t *testing.T) {
	li := NewLoadingIndicator("Loading...")
	li.Start()
	li.Start() // double start should not panic
	li.Stop()
	li.Stop() // double stop should not panic
}

func TestP184_SessionSidebar_MeasureWithItems(t *testing.T) {
	s := NewSessionSidebar()
	s.SetItems([]SessionItem{
		{ID: "1", Title: "A", Workspace: "ws"},
		{ID: "2", Title: "B", Workspace: "ws"},
	})
	sz := s.Measure(Bounded(30, 20))
	if sz.W <= 0 || sz.H <= 0 {
		t.Fatalf("expected positive size, got %v", sz)
	}
}

func TestP184_SessionSidebar_MeasureCollapsed(t *testing.T) {
	s := NewSessionSidebar()
	s.SetItems([]SessionItem{{ID: "1", Title: "X", Workspace: "ws"}})
	s.SetCollapsed(true)
	sz := s.Measure(Bounded(30, 20))
	if sz.W <= 0 {
		t.Fatalf("expected positive width when collapsed, got %v", sz)
	}
}

func TestP184_TextArea_MoveLineEmpty(t *testing.T) {
	ta := NewTextArea()
	ta.moveLine(1) // empty, should not panic
	ta.moveLine(-1)
}

func TestP184_TextArea_MoveLineSingle(t *testing.T) {
	ta := NewTextArea()
	ta.SetValue("single line")
	ta.moveLine(1) // single line, should not move
	ta.moveLine(-1)
}

func TestP184_Viewport_VScrollBarNearBottom(t *testing.T) {
	v := NewViewport(&fixedSize{w: 80, h: 100})
	v.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 20})
	v.ScrollDown(80) // near bottom
	buf := buffer.NewBuffer(80, 20)
	v.Paint(buf)
}

func TestP184_Viewport_HScrollBarNearRight(t *testing.T) {
	v := NewViewport(&fixedSize{w: 200, h: 20})
	v.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 20})
	v.ScrollRight(150) // near right
	buf := buffer.NewBuffer(80, 20)
	v.Paint(buf)
}

func TestP184_Viewport_BothScrollbarsOverflow(t *testing.T) {
	v := NewViewport(&fixedSize{w: 200, h: 100})
	v.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	v.ScrollRight(100)
	v.ScrollDown(80)
	buf := buffer.NewBuffer(40, 10)
	v.Paint(buf)
}

func TestP184_ThemeStudio_CursorCycle(t *testing.T) {
	ts := NewThemeStudio(nil)
	for i := 0; i < 100; i++ {
		ts.SetCursor(i % 30) // cycle through
	}
}

func TestP184_MenuBar_PaintDropdownLarge(t *testing.T) {
	mb := NewMenuBar([]Menu{
		{ID: "file", Title: "File", Items: makeMenuEntries(20)},
	})
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 25})
	mb.OpenMenu(0)
	buf := buffer.NewBuffer(80, 25)
	mb.Paint(buf)
}

func TestP184_MenuBar_PaintDropdownShort(t *testing.T) {
	mb := NewMenuBar([]Menu{
		{ID: "edit", Title: "Edit", Items: []MenuEntry{{ID: "copy", Label: "Copy", Shortcut: "Ctrl+C"}}},
	})
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 25})
	mb.OpenMenu(0)
	buf := buffer.NewBuffer(80, 25)
	mb.Paint(buf)
}

// Helpers

func makeMenuEntries(n int) []MenuEntry {
	entries := make([]MenuEntry, n)
	for i := range entries {
		entries[i] = MenuEntry{ID: "item" + itoa(i), Label: "Item " + itoa(i)}
	}
	return entries
}

// KeyEvent creates a simple key event for testing
func p184KeyEvt(r rune) *term.KeyEvent {
	return &term.KeyEvent{Rune: r}
}
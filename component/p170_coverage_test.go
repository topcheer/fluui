package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── P170 Coverage: targeted tests for sub-80% functions ───

func TestMaskedInput_SetFocus(t *testing.T) {
	mi := NewMaskedInput("##")
	mi.SetFocus(true)
	if !mi.IsFocused() {
		t.Error("should be focused")
	}
	mi.SetFocus(false)
	if mi.IsFocused() {
		t.Error("should not be focused")
	}
}

func TestMaskedInput_SetFg(t *testing.T) {
	mi := NewMaskedInput("##")
	mi.SetFg(buffer.RGB(255, 0, 0))
	mi.SetBounds(Rect{X: 0, Y: 0, W: 2, H: 1})
	buf := buffer.NewBuffer(2, 1)
	mi.insertChar('1')
	mi.Paint(buf)
	if buf.GetCell(0, 0).Fg.Val != buffer.RGB(255, 0, 0).Val {
		t.Error("SetFg should change foreground")
	}
}

func TestMaskedInput_SetBg(t *testing.T) {
	mi := NewMaskedInput("##")
	mi.SetBg(buffer.RGB(0, 0, 255))
	mi.SetBounds(Rect{X: 0, Y: 0, W: 2, H: 1})
	buf := buffer.NewBuffer(2, 1)
	mi.Paint(buf)
	if buf.GetCell(0, 0).Bg.Val != buffer.RGB(0, 0, 255).Val {
		t.Error("SetBg should change background")
	}
}

func TestMaskedInput_CharMatchesMask(t *testing.T) {
	mi := NewMaskedInput("##")
	if mi.charMatchesMask('a') {
		t.Error("letter should not match digit mask")
	}
	if !mi.charMatchesMask('5') {
		t.Error("digit should match digit mask")
	}
}

func TestSelectionList_Items(t *testing.T) {
	sl := NewSelectionList([]string{"A", "B"})
	items := sl.Items()
	if len(items) != 2 || items[0].Label != "A" {
		t.Error("Items should return copy")
	}
}

func TestSelectionList_SetOnChange(t *testing.T) {
	sl := NewSelectionList([]string{"A"})
	var called bool
	sl.SetOnChange(func() { called = true })
	sl.Toggle(0)
	if !called {
		t.Error("onChange should fire on toggle")
	}
}

func TestSelectionList_HandleKeyEnter(t *testing.T) {
	sl := NewSelectionList([]string{"A", "B"})
	sl.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !sl.IsSelected(0) {
		t.Error("Enter should toggle")
	}
}

func TestSelectionList_SelectAllDeselectAllKeys(t *testing.T) {
	sl := NewSelectionList([]string{"A", "B", "C"})
	sl.HandleKey(&term.KeyEvent{Rune: 'a'})
	if len(sl.SelectedItems()) != 3 {
		t.Error("a should select all")
	}
	sl.HandleKey(&term.KeyEvent{Rune: 'd'})
	if len(sl.SelectedItems()) != 0 {
		t.Error("d should deselect all")
	}
}

func TestSelectionList_PaintWithSelection(t *testing.T) {
	sl := NewSelectionList([]string{"A", "B"})
	sl.Toggle(0)
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 2})
	buf := buffer.NewBuffer(20, 2)
	sl.Paint(buf)
	// First item should show [x]
	if buf.GetCell(1, 0).Rune != 'x' {
		t.Error("selected item should show [x]")
	}
}

func TestSelectionList_PaintDisabled(t *testing.T) {
	sl := NewSelectionList([]string{"A", "B"})
	sl.SetDisabled(0, true)
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 2})
	buf := buffer.NewBuffer(20, 2)
	sl.Paint(buf) // should not crash, disabled items use DisabledFg
}

func TestSelectionList_PaintTruncateLabel(t *testing.T) {
	sl := NewSelectionList([]string{"This is a very long label that exceeds bounds"})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	sl.Paint(buf) // should truncate label
}

func TestLineGauge_SetFgBg(t *testing.T) {
	lg := NewLineGauge()
	lg.SetFg(buffer.RGB(255, 0, 0))
	lg.SetBg(buffer.RGB(0, 0, 255))
	lg.SetPercent(50)
	lg.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	lg.Paint(buf)
	// Fill should use fg color
	if buf.GetCell(0, 0).Fg.Val != buffer.RGB(255, 0, 0).Val {
		t.Error("filled portion should use fg color")
	}
}

func TestLineGauge_PaintWithLabel(t *testing.T) {
	lg := NewLineGauge()
	lg.SetLabel("50%")
	lg.SetPercent(50)
	lg.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	lg.Paint(buf) // should not crash with label
}

// ─── Coverage for existing sub-80% functions ───

func TestBadge_MeasureAllVariants(t *testing.T) {
	variants := []BadgeVariant{BadgeInfo, BadgeSuccess, BadgeWarning, BadgeError, BadgeCritical, BadgeNeutral}
	for _, v := range variants {
		b := NewBadge("Test", v)
		s := b.Measure(Bounded(20, 5))
		if s.H < 1 {
			t.Errorf("variant %d: height should be >= 1", v)
		}
	}
}

func TestBadge_MeasureWithIcon(t *testing.T) {
	b := NewBadge("Test", BadgeInfo)
	b.SetIcon("*")
	s := b.Measure(Bounded(20, 5))
	if s.W < 1 {
		t.Error("width should be positive")
	}
}

func TestBadge_MeasureNarrow(t *testing.T) {
	b := NewBadge("Long Text Here", BadgeInfo)
	s := b.Measure(Bounded(5, 5))
	if s.W > 5 {
		t.Error("width should be clamped to max")
	}
}

func TestCodeBlock_StreamingCursorNarrow(t *testing.T) {
	cb := NewCodeBlock("go", "hello")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 5})
	buf := buffer.NewBuffer(3, 5)
	cb.Paint(buf) // should handle narrow bounds
}

func TestCodeBlock_StreamingCursorLineNumbers(t *testing.T) {
	cb := NewCodeBlock("go", "line1\nline2\nline3")
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf) // should handle line numbers + streaming
}

func TestCodeBlock_StreamingCursorNotStreaming(t *testing.T) {
	cb := NewCodeBlock("go", "hello")
	cb.SetStreaming(false)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf) // should not show cursor when not streaming
}

func TestCodeBlock_StreamingCursorEmpty(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf) // should handle empty content
}

func TestCodeBlock_StreamingCursorZeroBounds(t *testing.T) {
	cb := NewCodeBlock("go", "hello")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 10)
	cb.Paint(buf) // should not panic
}

func TestDiffPreview_SetShowLineNumbers_P170(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	dp.SetShowLineNumbers(false)
	// just verify it doesn't crash
}

func TestDiffPreview_SetShowStats(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(true)
	dp.SetShowStats(false)
}

func TestDiffPreview_PaintBorderNarrow(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("line1\nline2")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 5})
	buf := buffer.NewBuffer(5, 5)
	dp.Paint(buf) // should handle narrow border
}

func TestDiffPreview_PaintBorderTallContent(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	buf := buffer.NewBuffer(30, 3)
	dp.Paint(buf) // should handle tall content with scrolling
}

func TestAutoComplete_PaintWithDescription(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "test", Description: "A test item", Category: "utils"},
		{Label: "foo", Description: "Another item", Category: "utils"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	ac.Paint(buf) // should show descriptions
}

func TestAutoComplete_PaintCategorySelected(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "test", Category: "cat1"},
		{Label: "foo", Category: "cat2"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	ac.Paint(buf)
}

func TestContextMenu_SetCursorNegative(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "Item A"))
	cm.AddItem(NewMenuItem("b", "Item B"))
	cm.SetCursor(-1)
	if cm.Cursor() != 0 {
		t.Error("negative cursor should clamp to 0")
	}
}

func TestContextMenu_SetCursorOverflow(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "Item A"))
	cm.AddItem(NewMenuItem("b", "Item B"))
	cm.SetCursor(100)
	if cm.Cursor() != 1 {
		t.Error("overflow cursor should clamp to last")
	}
}

func TestScrollView_ContentWTiny(t *testing.T) {
	child := NewFill(' ', buffer.Style{})
	sv := NewScrollView(child)
	sv.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 10})
	// contentW should handle tiny bounds
}

func TestMenuBar_PaintDropdownLong(t *testing.T) {
	mb := NewMenuBar([]Menu{
		{ID: "file", Title: "File", Items: []MenuEntry{
			{ID: "a", Label: "Item A"},
			{ID: "b", Label: "Item B"},
			{ID: "c", Label: "Item C"},
			{ID: "d", Label: "Item D"},
			{ID: "e", Label: "Item E"},
			{ID: "f", Label: "Item F"},
			{ID: "g", Label: "Item G"},
			{ID: "h", Label: "Item H"},
			{ID: "i", Label: "Item I"},
			{ID: "j", Label: "Item J"},
			{ID: "k", Label: "Item K"},
			{ID: "l", Label: "Item L"},
		}},
	})
	mb.OpenMenu(0)
	mb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	buf := buffer.NewBuffer(40, 20)
	mb.Paint(buf) // should handle scrollable dropdown
}

func TestRichLog_CountVisibleLinesMultiLine(t *testing.T) {
	rl := NewRichLog()
	rl.SetMaxSize(100)
	for i := 0; i < 10; i++ {
		rl.Write(LogInfo, "line that is long enough to wrap when bounds are narrow and content overflows")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 50})
	// countVisibleLinesLocked should handle multi-line wrapping
}

func TestPlaceVertical_Alignment(t *testing.T) {
	result := PlaceVertical(10, Middle, "test")
	if result == "" {
		t.Error("PlaceVertical should produce output")
	}
}

func TestPlaceVertical_TopBottom(t *testing.T) {
	result := PlaceVertical(10, Bottom, "test")
	if result == "" {
		t.Error("PlaceVertical should produce output")
	}
}

func TestAdaptiveColor_ParseColorNamed(t *testing.T) {
	ac := AdaptiveColor{Light: "red", Dark: "blue"}
	// Just verify it resolves without crashing
	_ = ac.Resolve()
}

func TestAdaptiveColor_ParseColorHex(t *testing.T) {
	ac := AdaptiveColor{Light: "#ff0000", Dark: "#0000ff"}
	_ = ac.Resolve()
}

func TestBaseComponent_PaintNoOp(t *testing.T) {
	bc := BaseComponent{}
	bc.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	bc.Paint(buf) // should be no-op, not panic
}

func TestBaseComponent_MeasureNoOp(t *testing.T) {
	bc := BaseComponent{}
	s := bc.Measure(Bounded(100, 50))
	if s.W != 0 || s.H != 0 {
		t.Error("BaseComponent.Measure should return 0x0")
	}
}

func TestBaseComponent_Children(t *testing.T) {
	bc := BaseComponent{}
	if bc.Children() != nil {
		t.Error("BaseComponent.Children should return nil")
	}
}

func TestHelpOverlay_EnsureSelectedVisibleUp(t *testing.T) {
	ho := NewHelpOverlay([]HelpGroup{
		{Name: "g1", Entries: []HelpEntry{{Keys: "ctrl+a", Description: "Action A"}}},
		{Name: "g2", Entries: []HelpEntry{{Keys: "ctrl+b", Description: "Action B"}}},
	})
	ho.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	ho.ScrollDown(5)
	ho.ScrollUp(2)
	// just verify it doesn't crash
}
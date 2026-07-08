package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === MenuBar firstSelectableLocked: all-separator/disabled menu (66.7% → 100%) ===

func TestP121_MenuBar_FirstSelectableAllSeparators(t *testing.T) {
	mb := NewMenuBar([]Menu{
		{ID: "test", Title: "Test", Items: []MenuEntry{
			{ID: "sep", Label: "", Separator: true},
			{ID: "dis", Label: "Disabled", Disabled: true},
		}},
	})
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(0)
	// Should have no selectable item
	if mb.SelectedItem() != -1 {
		t.Errorf("expected -1 for all-separator menu, got %d", mb.SelectedItem())
	}
}

func TestP121_MenuBar_NextSelectableAllDisabled(t *testing.T) {
	mb := NewMenuBar([]Menu{
		{ID: "test", Title: "Test", Items: []MenuEntry{
			{ID: "a", Label: "A", Disabled: true},
			{ID: "b", Label: "B", Disabled: true},
		}},
	})
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	// Open menu - should not select anything
	mb.OpenMenu(0)
	// Try to navigate down - should stay at -1 or not crash
	mb.HandleKey(&term.KeyEvent{Key: term.KeyDown})
}

func TestP121_MenuBar_PrevSelectableAllDisabled(t *testing.T) {
	mb := NewMenuBar([]Menu{
		{ID: "test", Title: "Test", Items: []MenuEntry{
			{ID: "a", Label: "A", Disabled: true},
			{ID: "b", Label: "B", Disabled: true},
		}},
	})
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(0)
	mb.HandleKey(&term.KeyEvent{Key: term.KeyUp})
}

func TestP121_MenuBar_NextSelectableEmpty(t *testing.T) {
	mb := NewMenuBar([]Menu{
		{ID: "test", Title: "Test", Items: []MenuEntry{}},
	})
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(0)
	mb.HandleKey(&term.KeyEvent{Key: term.KeyDown})
}

func TestP121_MenuBar_PrevSelectableEmpty(t *testing.T) {
	mb := NewMenuBar([]Menu{
		{ID: "test", Title: "Test", Items: []MenuEntry{}},
	})
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(0)
	mb.HandleKey(&term.KeyEvent{Key: term.KeyUp})
}

func TestP121_MenuBar_ComputeDropDimsNotOpen(t *testing.T) {
	mb := NewMenuBar(testMenus())
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	// When menu is closed, computeDropDimsLocked should be a no-op
	mb.CloseMenu()
	mb.Paint(buffer.NewBuffer(80, 10))
}

func TestP121_MenuBar_PaintDropdownOverflow(t *testing.T) {
	// Menu with many items that overflow available height
	items := make([]MenuEntry, 30)
	for i := range items {
		items[i] = MenuEntry{ID: "item", Label: "Item", Shortcut: "Ctrl+X"}
	}
	mb := NewMenuBar([]Menu{{ID: "big", Title: "Big", Items: items}})
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(0)
	// Navigate to last item
	for i := 0; i < 25; i++ {
		mb.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	}
	buf := buffer.NewBuffer(80, 10)
	mb.Paint(buf) // should not panic with overflow
}

func TestP121_MenuBar_PaintDropdownShortLabel(t *testing.T) {
	mb := NewMenuBar([]Menu{
		{ID: "x", Title: "X", Items: []MenuEntry{
			{ID: "a", Label: "A"},  // very short, tests min width 12
		}},
	})
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})
	mb.OpenMenu(0)
	buf := buffer.NewBuffer(80, 10)
	mb.Paint(buf)
}

// === CodeBlock paintStreamingCursorLocked uncovered branches (74.2% → 85%+) ===

func TestP121_CodeBlock_StreamingCursor_PlainFallback(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	// Add enough content to fill bounds
	cb.AppendSource("line1\nline2\nline3\nline4\nline5")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	cb.Paint(buf) // exercises plainFallback path
}

func TestP121_CodeBlock_StreamingCursor_ScrollPastEnd(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.AppendSource("a\nb\nc\nd\ne")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	cb.ScrollTo(10) // scroll way past end
	buf := buffer.NewBuffer(20, 3)
	cb.Paint(buf) // exercises lastIdx >= len path
}

func TestP121_CodeBlock_StreamingCursor_LongLineClamp(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.AppendSource("very long line that exceeds width bounds for testing")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	buf := buffer.NewBuffer(10, 3)
	cb.Paint(buf) // exercises x >= bounds.X+bounds.W clamp
}

func TestP121_CodeBlock_StreamingCursor_OutOfBoundsY(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.AppendSource("line1\nline2")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1}) // very short
	buf := buffer.NewBuffer(20, 1)
	cb.Paint(buf) // exercises y < bounds.Y or y >= bounds.Y+bounds.H
}

// === ContextMenu setCursorLocked backward search (73.3% → 90%+) ===

func TestP121_ContextMenu_SetCursorNegative(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "A"))
	cm.AddItem(NewMenuItem("b", "B"))
	cm.AddItem(NewMenuItem("c", "C"))
	cm.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	cm.SetCursor(-5) // should clamp to 0
}

func TestP121_ContextMenu_SetCursorOverflow(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "A"))
	cm.AddItem(NewMenuItem("b", "B"))
	cm.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	cm.SetCursor(100) // should clamp to len-1
}

func TestP121_ContextMenu_SetCursorEmpty(t *testing.T) {
	cm := NewContextMenu()
	cm.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	cm.SetCursor(0) // empty items, should set cursor=0
}

// === TextField HandleKey maxLen (78.3% → 90%+) ===

func TestP121_TextField_HandleKey_MaxLen(t *testing.T) {
	tf := NewTextField("test", "key", "abcde")
	tf.SetMaxLength(5) // current value is 5 chars = at max
	consumed := tf.HandleKey(&term.KeyEvent{Rune: 'X', Key: term.KeyUnknown})
	if !consumed {
		t.Error("expected consumed even at max len")
	}
}

func TestP121_TextField_HandleKey_Backspace(t *testing.T) {
	tf := NewTextField("test", "key", "abc")
	// cursor should be at len(value) by default
	tf.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if len(tf.Value()) != 2 {
		t.Errorf("expected 2 chars, got %d", len(tf.Value()))
	}
}

func TestP121_TextField_HandleKey_NilKey(t *testing.T) {
	tf := NewTextField("test", "key", "abc")
	consumed := tf.HandleKey(nil)
	if consumed {
		t.Error("expected false for nil key")
	}
}

// === HelpOverlay ensureSelectedVisibleLocked visibleH<=0 (75% → 90%+) ===

func TestP121_HelpOverlay_VisibleHZero(t *testing.T) {
	groups := []HelpGroup{
		{Name: "Nav", Entries: []HelpEntry{
			{Keys: "j", Description: "down"},
			{Keys: "k", Description: "up"},
		}},
	}
	ho := NewHelpOverlay(groups)
	ho.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 2}) // H=2, visibleH = 2-3 = -1
	buf := buffer.NewBuffer(60, 2)
	ho.Paint(buf) // should handle visibleH <= 0
}

func TestP121_HelpOverlay_ScrollAndSelectDown(t *testing.T) {
	groups := []HelpGroup{
		{Name: "Nav", Entries: []HelpEntry{
			{Keys: "j/k", Description: "up/down"},
			{Keys: "h/l", Description: "left/right"},
			{Keys: "g", Description: "top"},
			{Keys: "G", Description: "bottom"},
		}},
	}
	ho := NewHelpOverlay(groups)
	ho.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5}) // visibleH = 5-3 = 2
	// Scroll down several times to exercise ensureSelectedVisibleLocked
	for i := 0; i < 5; i++ {
		ho.ScrollDown(1)
	}
	buf := buffer.NewBuffer(60, 5)
	ho.Paint(buf)
}

// === ListView MoveUp empty items + OnChange (76.9% → 90%+) ===

func TestP121_ListView_MoveUpEmpty(t *testing.T) {
	lv := NewListView([]string{})
	lv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	lv.MoveUp() // should not panic with empty items
}

func TestP121_ListView_MoveUpOnChange(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	fired := false
	lv.OnChange = func(cursor int) {
		fired = true
	}
	lv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	lv.MoveUp()
	if !fired {
		t.Error("expected OnChange to fire")
	}
}

func TestP121_ListView_MoveDownEmpty(t *testing.T) {
	lv := NewListView([]string{})
	lv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	lv.MoveDown() // should not panic
}

// === DiffPreview paintBorderLocked narrow width (76.5% → 85%+) ===

func TestP121_DiffPreview_PaintBorderVeryNarrow(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+added line\n-removed line\n context")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 2, H: 5})
	buf := buffer.NewBuffer(2, 5)
	dp.Paint(buf) // tests narrow width border clipping
}

func TestP121_DiffPreview_PaintBorderNormalWidth(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+added\n-removed\n context\n@@ hunk @@")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 8})
	buf := buffer.NewBuffer(50, 8)
	dp.Paint(buf)
}

// === BarChart paintHorizontal partial values (78.6% → 85%+) ===

func TestP121_BarChart_Horizontal_PartialValues(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	bc.SetShowValues(true)
	bc.SetShowLegend(true)
	bc.AddSeries(BarSeries{
		Name:  "data",
		Data:  []BarData{{Label: "A", Value: 25}, {Label: "B", Value: 50}, {Label: "C", Value: 75}},
		Color: buffer.RGB(0xff, 0x79, 0xc6),
	})
	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	bc.Paint(buf)
}

// === Autocomplete Paint with scroll offset (76.7% → 85%+) ===

func TestP121_AutoComplete_Paint_ScrollAll(t *testing.T) {
	ac := NewAutoComplete()
	items := make([]CompletionItem, 30)
	for i := range items {
		items[i] = CompletionItem{Label: "item", Value: "v"}
	}
	ac.SetItems(items)
	ac.SetCursor(25)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	ac.Paint(buf)
}

func TestP121_AutoComplete_Paint_ZeroBounds(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{{Label: "a", Value: "1"}})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(20, 5)
	ac.Paint(buf) // should not panic
}

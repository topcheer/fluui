package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── ContextMenu HandleKey coverage ───

func TestP76_ContextMenu_HandleKey_Hidden(t *testing.T) {
	cm := NewContextMenu()
	// Hidden menu should not consume keys
	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if consumed {
		t.Error("hidden menu should not consume keys")
	}
}

func TestP76_ContextMenu_HandleKey_RightNoSubmenu(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("item1", "Item 1")
	cm.AddLabel("item2", "Item 2")
	cm.Show(0, 0)
	cm.SetCursor(0)

	// Right arrow on item without submenu should not consume
	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if consumed {
		t.Error("right arrow on item without submenu should not consume")
	}
}

func TestP76_ContextMenu_HandleKey_RightWithSubmenu(t *testing.T) {
	cm := NewContextMenu()
	sub := NewContextMenu()
	sub.AddLabel("sub1", "Sub Item 1")
	parent := NewMenuItem("parent", "Parent")
	parent.Submenu = sub
	cm.AddItem(parent)
	cm.Show(0, 0)
	cm.SetCursor(0)

	// Right arrow should activate submenu
	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if !consumed {
		t.Error("right arrow with submenu should consume")
	}
}

func TestP76_ContextMenu_HandleKey_LeftClosesSubmenu(t *testing.T) {
	cm := NewContextMenu()
	sub := NewContextMenu()
	sub.AddLabel("sub1", "Sub Item 1")
	parent := NewMenuItem("parent", "Parent")
	parent.Submenu = sub
	cm.AddItem(parent)
	cm.Show(0, 0)
	cm.SetCursor(0)

	// First activate submenu
	cm.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	// Then press Left to close it
	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if !consumed {
		t.Error("left arrow in submenu should consume and close")
	}
}

func TestP76_ContextMenu_HandleKey_DefaultNoConsume(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("item1", "Item 1")
	cm.Show(0, 0)

	// Random key should not be consumed
	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if consumed {
		t.Error("unknown key should not consume")
	}
}

func TestP76_ContextMenu_HandleKey_EnterActivate(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("item1", "Item 1")
	cm.Show(0, 0)
	cm.SetCursor(0)

	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !consumed {
		t.Error("enter should consume")
	}
}

func TestP76_ContextMenu_HandleKey_EscapeHide(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("item1", "Item 1")
	cm.Show(0, 0)

	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Error("escape should consume")
	}
	if cm.Visible() {
		t.Error("escape should hide menu")
	}
}

// ─── CodeBlock paintStreamingCursor coverage ───

func TestP76_CodeBlock_StreamingCursor_Empty(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	buf := buffer.NewBuffer(40, 10)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	cb.Paint(buf)
}

func TestP76_CodeBlock_StreamingCursor_WithTitle(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetShowTitle(true)
	cb.SetTitle("test.go")
	buf := buffer.NewBuffer(40, 10)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	cb.Paint(buf)
}

func TestP76_CodeBlock_StreamingCursor_WithLineNumbers(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	buf := buffer.NewBuffer(40, 10)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	cb.Paint(buf)
}

func TestP76_CodeBlock_StreamingCursor_WithContent(t *testing.T) {
	cb := NewCodeBlock("go", "package main")
	cb.SetStreaming(true)
	buf := buffer.NewBuffer(40, 10)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	cb.Paint(buf)
}

func TestP76_CodeBlock_StreamingCursor_ContentExceedsView(t *testing.T) {
	cb := NewCodeBlock("go", "line1\nline2\nline3\nline4\nline5")
	cb.SetStreaming(true)
	buf := buffer.NewBuffer(40, 3)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	cb.Paint(buf)
}

func TestP76_CodeBlock_StreamingCursor_LineExceedsWidth(t *testing.T) {
	cb := NewCodeBlock("go", "very long line that exceeds the width of the code block area")
	cb.SetStreaming(true)
	buf := buffer.NewBuffer(20, 5)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	cb.Paint(buf)
}

// ─── Badge Measure coverage ───

func TestP76_Badge_Measure_Small(t *testing.T) {
	b := NewBadge("X", BadgeSuccess)
	s := b.Measure(Unbounded())
	if s.H < 1 {
		t.Errorf("badge measure H = %d, want >= 1", s.H)
	}
}

func TestP76_Badge_Measure_LongText(t *testing.T) {
	b := NewBadge("Very Long Badge Text Here", BadgeWarning)
	s := b.Measure(Unbounded())
	if s.W < 10 {
		t.Errorf("badge measure W = %d, want >= 10", s.W)
	}
}

func TestP76_Badge_Measure_Empty(t *testing.T) {
	b := NewBadge("", BadgeInfo)
	s := b.Measure(Unbounded())
	if s.H < 1 {
		t.Errorf("empty badge H = %d, want >= 1", s.H)
	}
}

func TestP76_Badge_Measure_Constrained(t *testing.T) {
	b := NewBadge("Hello World", BadgeError)
	s := b.Measure(Bounded(10, 5))
	if s.W > 10 {
		t.Errorf("constrained badge W = %d, want <= 10", s.W)
	}
}

// ─── Canvas SetCell/SetCellBG coverage ───

func TestP76_Canvas_SetCell_OutOfBounds(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	// Negative coordinates
	c.SetCell(-1, -1, '#', buffer.NamedColor(buffer.NamedWhite))
	// Beyond bounds
	c.SetCell(100, 100, '#', buffer.NamedColor(buffer.NamedWhite))
}

func TestP76_Canvas_SetCellBG_OutOfBounds(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	// Negative
	c.SetCellBG(-1, -1, '#', buffer.NamedColor(buffer.NamedWhite), buffer.NamedColor(buffer.NamedBlue))
	// Beyond bounds
	c.SetCellBG(50, 50, '#', buffer.NamedColor(buffer.NamedWhite), buffer.NamedColor(buffer.NamedBlue))
}

// ─── Dialog PressButton coverage ───

func TestP76_Dialog_PressButton_EmptyButtons(t *testing.T) {
	d := NewDialog(DialogInfo, "Test", "Message")
	// PressButton with no buttons should not panic
	d.PressButton()
}

func TestP76_Dialog_PressButton_WithButtons(t *testing.T) {
	d := NewDialog(DialogConfirm, "Confirm", "Are you sure?")
	d.SetButtons([]DialogButton{
		{Label: "OK", Result: DialogResultOK},
		{Label: "Cancel", Result: DialogResultCancel},
	})
	d.SetCursor(0)
	// Should activate first button
	d.PressButton()
}

func TestP76_Dialog_PressButton_SecondButton(t *testing.T) {
	d := NewDialog(DialogConfirm, "Confirm", "Are you sure?")
	d.SetButtons([]DialogButton{
		{Label: "OK", Result: DialogResultOK},
		{Label: "Cancel", Result: DialogResultCancel},
	})
	d.SetCursor(1)
	d.PressButton()
}

// ─── Checkbox setNavigableCursor coverage ───

func TestP76_Checkbox_NavigateDisabledItems(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C", "D"})
	cb.SetItems([]CheckboxItem{
		{Label: "A"},
		{Label: "B", Disabled: true},
		{Label: "C", Disabled: true},
		{Label: "D"},
	})
	cb.MoveDown() // should skip B, C and land on D
	idx := cb.Cursor()
	if idx != 3 {
		t.Errorf("after MoveDown with disabled items, cursor = %d, want 3", idx)
	}
}

func TestP76_Checkbox_AllDisabled(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B"})
	cb.SetItems([]CheckboxItem{
		{Label: "A", Disabled: true},
		{Label: "B", Disabled: true},
	})
	cb.SetCursor(0)
	cb.MoveDown() // should stay or wrap
	// Should not panic
}

// ─── BarChart paintHorizontal coverage ───

func TestP76_BarChart_PaintHorizontal(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	bc.AddSeries(BarSeries{
		Name: "Test",
		Data: []BarData{
			{Label: "A", Value: 10},
			{Label: "B", Value: 20},
			{Label: "C", Value: 30},
		},
	})
	bc.SetShowGrid(true)
	bc.SetShowAxes(true)
	bc.SetShowLegend(true)
	bc.SetShowValues(true)
	buf := buffer.NewBuffer(60, 10)
	bc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	bc.Paint(buf)
}

func TestP76_BarChart_PaintHorizontal_SmallBounds(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	bc.AddSeries(BarSeries{
		Name: "Test",
		Data: []BarData{{Label: "A", Value: 5}},
	})
	buf := buffer.NewBuffer(15, 4)
	bc.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 4})
	bc.Paint(buf)
}

// ─── Collapsible Measure coverage ───

func TestP76_Collapsible_Measure_Collapsed(t *testing.T) {
	child := NewTooltip("child content")
	c := NewCollapsible("Title", child)
	c.Collapse()
	s := c.Measure(Unbounded())
	if s.H < 1 {
		t.Errorf("collapsed measure H = %d, want >= 1", s.H)
	}
}

func TestP76_Collapsible_Measure_Expanded(t *testing.T) {
	child := NewTooltip("child text")
	c := NewCollapsible("Title", child)
	c.Expand()
	s := c.Measure(Unbounded())
	if s.H < 2 {
		t.Errorf("expanded measure H = %d, want >= 2", s.H)
	}
}

func TestP76_Collapsible_SetBounds_Collapsed(t *testing.T) {
	child := NewTooltip("child")
	c := NewCollapsible("Title", child)
	c.Collapse()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	c.Paint(buf)
}

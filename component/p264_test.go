package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestCanvas_SetCell_NilActive_P264(t *testing.T) {
	c := NewCanvas()
	c.SetCell(0, 0, 'X', buffer.RGB(255, 0, 0))
}

func TestCanvas_SetCellBG_NilActive_P264(t *testing.T) {
	c := NewCanvas()
	c.SetCellBG(0, 0, 'X', buffer.RGB(255, 0, 0), buffer.RGB(0, 0, 255))
}

func TestCanvas_SetCell_WithLayer_P264(t *testing.T) {
	c := NewCanvas()
	c.AddLayer("main")
	c.SetCell(5, 5, 'A', buffer.RGB(255, 0, 0))
	c.SetCellBG(6, 6, 'B', buffer.RGB(0, 255, 0), buffer.RGB(0, 0, 0))
}

func TestCheckbox_SetItems_CursorReset_P264(t *testing.T) {
	cb := NewCheckbox([]string{"a", "b", "c", "d", "e"})
	cb.SetCursor(4)
	cb.SetItems([]CheckboxItem{{Label: "x"}, {Label: "y"}})
	if cb.Cursor() >= 2 {
		t.Errorf("cursor should reset, got %d", cb.Cursor())
	}
}

func TestCheckbox_SetItems_NoReset_P264(t *testing.T) {
	cb := NewCheckbox([]string{"a", "b", "c"})
	cb.SetCursor(1)
	cb.SetItems([]CheckboxItem{{Label: "x"}, {Label: "y"}, {Label: "z"}})
	if cb.Cursor() != 1 {
		t.Errorf("cursor should stay at 1, got %d", cb.Cursor())
	}
}

func TestAutocomplete_ClampScroll_Negative_P264(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetPosition(0, 0)
	ac.SetMaxVisible(5)
	ac.SetItems([]CompletionItem{
		{Label: "item1"}, {Label: "item2"}, {Label: "item3"},
	})
	ac.SetQuery("item")
	ac.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	ac.scrollY = -5
	ac.clampScrollLocked()
}

func TestAutocomplete_Paint_WithDescCat_P264(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetPosition(0, 0)
	ac.SetItems([]CompletionItem{
		{Label: "func", Description: "function", Category: "builtin"},
		{Label: "var", Description: "variable", Category: "builtin"},
		{Label: "type", Description: "type def", Category: "user"},
	})
	ac.SetQuery("")
	ac.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	ac.Paint(buf)
}

func TestAutocomplete_Paint_TruncateItems_P264(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetPosition(0, 0)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	items := make([]CompletionItem, 20)
	for i := range items {
		items[i] = CompletionItem{Label: "item" + string(rune('A'+i%26))}
	}
	ac.SetItems(items)
	ac.SetQuery("")
	buf := buffer.NewBuffer(30, 3)
	ac.Paint(buf)
}

func TestCollapsible_Measure_ClampWidth_P264(t *testing.T) {
	child := &fixedSize{w: 50, h: 10}
	col := NewCollapsible("Title", child)
	s := col.Measure(Constraints{MaxWidth: 20, MaxHeight: 3})
	if s.W > 20 {
		t.Errorf("width should be clamped to 20, got %d", s.W)
	}
	if s.H < 1 {
		t.Error("height should be at least 1")
	}
}

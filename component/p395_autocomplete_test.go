package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P395: Autocomplete Paint coverage — category/description/scroll paths

func TestP395_AutoComplete_Paint_WithCategoryAndDesc(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "fmt", Category: "stdlib", Description: "format package"},
		{Label: "os", Category: "stdlib", Description: "OS functions"},
	})
	ac.SetQuery("f")
	ac.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	ac.Paint(buf)
}

func TestP395_AutoComplete_Paint_SelectedWithCategory(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "apple", Category: "fruit", Description: "red fruit"},
		{Label: "banana", Category: "fruit", Description: "yellow fruit"},
		{Label: "carrot", Category: "veg", Description: "orange veg"},
	})
	ac.SetQuery("")
	ac.SetCursor(1) // select "banana"
	ac.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	ac.Paint(buf)
}

func TestP395_AutoComplete_Paint_Scrolled(t *testing.T) {
	ac := NewAutoComplete()
	items := make([]CompletionItem, 20)
	for i := range items {
		items[i] = CompletionItem{Label: "item" + string(rune('a'+i))}
	}
	ac.SetItems(items)
	ac.SetQuery("")
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	// Scroll past first page
	ac.Measure(Constraints{MaxWidth: 30, MaxHeight: 5})
	ac.SetCursor(10)
	buf := buffer.NewBuffer(30, 5)
	ac.Paint(buf)
}

func TestP395_AutoComplete_Paint_LongCategory(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "x", Category: "very_long_category_name_that_exceeds_limits", Description: "d"},
	})
	ac.SetQuery("x")
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	ac.Paint(buf) // category should be truncated
}

func TestP395_AutoComplete_Measure_WithDescAndCategory(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "test", Category: "cat", Description: "desc"},
	})
	ac.SetQuery("t")
	s := ac.Measure(Constraints{MaxWidth: 80, MaxHeight: 20})
	if s.W < 15 {
		t.Errorf("W = %d, too small", s.W)
	}
}

func TestP395_AutoComplete_Paint_SelectedScroll(t *testing.T) {
	ac := NewAutoComplete()
	items := make([]CompletionItem, 10)
	for i := range items {
		items[i] = CompletionItem{Label: "item" + string(rune('a'+i))}
	}
	ac.SetItems(items)
	ac.SetQuery("")
	ac.SetCursor(8) // select item near bottom
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	ac.Measure(Constraints{MaxWidth: 30, MaxHeight: 5})
	buf := buffer.NewBuffer(30, 5)
	ac.Paint(buf) // should auto-scroll to show selected
}

func TestP395_AutoComplete_Paint_NonZeroOffset(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "hello"},
	})
	ac.SetQuery("h")
	ac.SetBounds(Rect{X: 5, Y: 3, W: 20, H: 5})
	buf := buffer.NewBuffer(30, 10)
	ac.Paint(buf)
	// Top border should be at y=3 — just verify paint doesn't panic
	_ = buf.GetCell(5, 3)
}

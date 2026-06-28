package component

import (
	"fmt"
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P21-A: Additional VirtualScroller tests covering gaps in p17 tests.
// Existing p17_virtualscroller_test.go covers: New, SetItems, AddItem,
// AddItems, Items, ItemAt, Clear, SetCursor, MoveDown, MoveUp, MoveToStart,
// MoveToEnd, CurrentItem, ScrollTo, VisibleRange, VisibleItems, Filter,
// SetStyle, SetHeader, SetShowScrollbar, Measure, Paint, String.
//
// These tests cover: ItemCount, MovePageDown/Up, ShowBorder, Header getter,
// Style getter, Paint with header, Filter edge cases, concurrent filter.

func TestVirtualScroller_ItemCount_P21(t *testing.T) {
	vs := NewVirtualScroller()
	if vs.ItemCount() != 0 {
		t.Errorf("ItemCount() = %d, want 0 for empty", vs.ItemCount())
	}
	vs.AddItems([]VirtualItem{
		{ID: "a", Text: "Apple"},
		{ID: "b", Text: "Banana"},
		{ID: "c", Text: "Cherry"},
	})
	if vs.ItemCount() != 3 {
		t.Errorf("ItemCount() = %d, want 3", vs.ItemCount())
	}
}

func TestVirtualScroller_ItemCount_AfterClear_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.AddItems([]VirtualItem{
		{ID: "1", Text: "one"},
		{ID: "2", Text: "two"},
	})
	vs.Clear()
	if vs.ItemCount() != 0 {
		t.Errorf("ItemCount() = %d, want 0 after Clear", vs.ItemCount())
	}
}

func TestVirtualScroller_ItemAt_Negative_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.AddItem(VirtualItem{ID: "1", Text: "one"})
	if vs.ItemAt(-1) != nil {
		t.Error("ItemAt(-1) should return nil")
	}
}

func TestVirtualScroller_ItemAt_BeyondEnd_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.AddItem(VirtualItem{ID: "1", Text: "one"})
	if vs.ItemAt(5) != nil {
		t.Error("ItemAt(5) should return nil for single item")
	}
}

func TestVirtualScroller_CurrentItem_Empty_P21(t *testing.T) {
	vs := NewVirtualScroller()
	if vs.CurrentItem() != nil {
		t.Error("CurrentItem() should return nil for empty scroller")
	}
}

func TestVirtualScroller_MovePageDown_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vs.SetItems(makeP21Items(100))
	vs.Paint(buffer.NewBuffer(20, 10))
	initialCursor := vs.Cursor()
	vs.MovePageDown()
	if vs.Cursor() <= initialCursor {
		t.Errorf("MovePageDown: cursor %d should be > initial %d", vs.Cursor(), initialCursor)
	}
}

func TestVirtualScroller_MovePageUp_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vs.SetItems(makeP21Items(100))
	vs.Paint(buffer.NewBuffer(20, 10))
	vs.MoveToEnd()
	highCursor := vs.Cursor()
	vs.MovePageUp()
	if vs.Cursor() >= highCursor {
		t.Errorf("MovePageUp: cursor %d should be < %d", vs.Cursor(), highCursor)
	}
}

func TestVirtualScroller_MovePageDown_Clamp_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vs.SetItems(makeP21Items(5))
	vs.Paint(buffer.NewBuffer(20, 10))
	vs.MovePageDown()
	if vs.Cursor() != 4 {
		t.Errorf("MovePageDown with 5 items: cursor %d, want 4 (clamped)", vs.Cursor())
	}
}

func TestVirtualScroller_MovePageUp_Clamp_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vs.SetItems(makeP21Items(5))
	vs.Paint(buffer.NewBuffer(20, 10))
	vs.MovePageUp()
	if vs.Cursor() != 0 {
		t.Errorf("MovePageUp with 5 items: cursor %d, want 0 (clamped)", vs.Cursor())
	}
}

func TestVirtualScroller_SetShowBorder_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetShowBorder(false)
	vs.SetShowBorder(true)
}

func TestVirtualScroller_Header_Getter_P21(t *testing.T) {
	vs := NewVirtualScroller()
	if vs.Header() != "" {
		t.Errorf("Header() = %q, want empty", vs.Header())
	}
	vs.SetHeader("Files")
	if vs.Header() != "Files" {
		t.Errorf("Header() = %q, want 'Files'", vs.Header())
	}
}

func TestVirtualScroller_Style_Getter_P21(t *testing.T) {
	vs := NewVirtualScroller()
	s := vs.Style()
	if s.Selected.Flags&buffer.Bold == 0 {
		t.Error("default style should have Bold on Selected")
	}
}

func TestVirtualScroller_Filter_NoMatch_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetItems([]VirtualItem{
		{ID: "1", Text: "Apple"},
		{ID: "2", Text: "Banana"},
	})
	result := vs.Filter("zzz")
	if len(result) != 0 {
		t.Errorf("Filter('zzz') returned %d results, want 0", len(result))
	}
}

func TestVirtualScroller_Items_ModifySafe_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetItems([]VirtualItem{{ID: "1", Text: "original"}})
	items := vs.Items()
	if len(items) > 0 {
		items[0].Text = "modified"
	}
	original := vs.Items()
	if len(original) > 0 && original[0].Text == "modified" {
		t.Error("modifying returned slice should not affect internal state")
	}
}

func TestVirtualScroller_VisibleItems_PartialRange_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 6})
	vs.SetItems(makeP21Items(50))
	buf := buffer.NewBuffer(20, 6)
	vs.Paint(buf)
	items := vs.VisibleItems()
	if len(items) == 0 {
		t.Error("VisibleItems() should return items for valid viewport")
	}
	if len(items) > 50 {
		t.Errorf("VisibleItems() returned %d, should not exceed total", len(items))
	}
}

func TestVirtualScroller_Paint_WithHeader_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vs.SetItems(makeP21Items(20))
	vs.SetHeader("My List")
	buf := buffer.NewBuffer(20, 10)
	vs.Paint(buf)
	found := false
	for x := 0; x < 20; x++ {
		if buf.GetCell(x, 1).Rune != 0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Paint with header should render header content")
	}
}

func TestVirtualScroller_Paint_WithoutBorder_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vs.SetItems(makeP21Items(20))
	vs.SetShowBorder(false)
	buf := buffer.NewBuffer(20, 10)
	vs.Paint(buf)
}

func TestVirtualScroller_Paint_ScrollbarLargeDataset_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	vs.SetItems(makeP21Items(10000))
	vs.ScrollTo(5000)
	buf := buffer.NewBuffer(30, 10)
	vs.Paint(buf)
}

func TestVirtualScroller_SetCursor_Empty_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetCursor(10)
	if vs.Cursor() != 0 {
		t.Errorf("Cursor() = %d, want 0 for empty list", vs.Cursor())
	}
}

func TestVirtualScroller_MoveToEnd_Empty_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.MoveToEnd()
	if vs.Cursor() != 0 {
		t.Errorf("Cursor() = %d, want 0 for empty list", vs.Cursor())
	}
}

func TestVirtualScroller_String_Format_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetItems(makeP21Items(10))
	vs.SetCursor(5)
	s := vs.String()
	if s == "" {
		t.Error("String() should not be empty")
	}
}

func TestVirtualScroller_Concurrent_Filter_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetItems(makeP21Items(100))
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = vs.Filter("item")
		}()
	}
	wg.Wait()
}

func TestVirtualScroller_ScrollY_Accessor_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vs.SetItems(makeP21Items(100))
	vs.ScrollTo(5)
	if vs.ScrollY() != 5 {
		t.Errorf("ScrollY() = %d, want 5", vs.ScrollY())
	}
}

func TestVirtualScroller_ScrollTo_Negative_P21(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vs.SetItems(makeP21Items(100))
	vs.ScrollTo(-10)
	if vs.ScrollY() != 0 {
		t.Errorf("ScrollY() = %d, want 0 (clamped from negative)", vs.ScrollY())
	}
}

// makeP21Items creates n VirtualItems for testing.
func makeP21Items(n int) []VirtualItem {
	items := make([]VirtualItem, n)
	for i := range items {
		items[i] = VirtualItem{
			ID:   fmt.Sprintf("item-%d", i),
			Text: fmt.Sprintf("item %d", i),
		}
	}
	return items
}

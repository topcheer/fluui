package component

import (
	"fmt"
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ===== VirtualScroller Tests =====

func TestNewVirtualScroller(t *testing.T) {
	vs := NewVirtualScroller()
	if vs.ItemCount() != 0 {
		t.Errorf("expected 0 items, got %d", vs.ItemCount())
	}
	if vs.Cursor() != 0 {
		t.Errorf("expected cursor 0")
	}
	if vs.ScrollY() != 0 {
		t.Errorf("expected scrollY 0")
	}
}

func TestVirtualScroller_SetItems(t *testing.T) {
	vs := NewVirtualScroller()
	items := make([]VirtualItem, 100)
	for i := range items {
		items[i] = VirtualItem{ID: fmt.Sprintf("item-%d", i), Text: fmt.Sprintf("Item %d", i)}
	}
	vs.SetItems(items)
	if vs.ItemCount() != 100 {
		t.Errorf("expected 100 items, got %d", vs.ItemCount())
	}
}

func TestVirtualScroller_AddItem(t *testing.T) {
	vs := NewVirtualScroller()
	vs.AddItem(VirtualItem{ID: "a", Text: "A"})
	vs.AddItem(VirtualItem{ID: "b", Text: "B"})
	if vs.ItemCount() != 2 {
		t.Errorf("expected 2, got %d", vs.ItemCount())
	}
}

func TestVirtualScroller_AddItems(t *testing.T) {
	vs := NewVirtualScroller()
	vs.AddItems([]VirtualItem{
		{ID: "a", Text: "A"},
		{ID: "b", Text: "B"},
		{ID: "c", Text: "C"},
	})
	if vs.ItemCount() != 3 {
		t.Errorf("expected 3, got %d", vs.ItemCount())
	}
}

func TestVirtualScroller_Items_Copy(t *testing.T) {
	vs := NewVirtualScroller()
	vs.AddItem(VirtualItem{ID: "a", Text: "A"})
	items := vs.Items()
	items[0].Text = "modified"
	if vs.ItemAt(0).Text != "A" {
		t.Error("Items() should return a copy")
	}
}

func TestVirtualScroller_ItemAt(t *testing.T) {
	vs := NewVirtualScroller()
	vs.AddItem(VirtualItem{ID: "a", Text: "A"})
	if vs.ItemAt(0).ID != "a" {
		t.Error("expected item a")
	}
	if vs.ItemAt(-1) != nil {
		t.Error("expected nil for negative")
	}
	if vs.ItemAt(99) != nil {
		t.Error("expected nil for OOB")
	}
}

func TestVirtualScroller_Clear(t *testing.T) {
	vs := NewVirtualScroller()
	vs.AddItem(VirtualItem{ID: "a", Text: "A"})
	vs.Clear()
	if vs.ItemCount() != 0 {
		t.Error("expected 0 after clear")
	}
	if vs.Cursor() != 0 || vs.ScrollY() != 0 {
		t.Error("cursor and scroll should reset")
	}
}

func TestVirtualScroller_SetCursor(t *testing.T) {
	vs := NewVirtualScroller()
	for i := 0; i < 50; i++ {
		vs.AddItem(VirtualItem{ID: fmt.Sprintf("i%d", i), Text: fmt.Sprintf("Item %d", i)})
	}
	vs.SetCursor(10)
	if vs.Cursor() != 10 {
		t.Errorf("expected cursor 10, got %d", vs.Cursor())
	}
}

func TestVirtualScroller_SetCursor_Clamp(t *testing.T) {
	vs := NewVirtualScroller()
	vs.AddItem(VirtualItem{ID: "a", Text: "A"})
	vs.SetCursor(100)
	if vs.Cursor() != 0 {
		t.Errorf("expected 0, got %d", vs.Cursor())
	}
	vs.SetCursor(-5)
	if vs.Cursor() != 0 {
		t.Errorf("expected 0, got %d", vs.Cursor())
	}
}

func TestVirtualScroller_MoveDown(t *testing.T) {
	vs := NewVirtualScroller()
	for i := 0; i < 20; i++ {
		vs.AddItem(VirtualItem{ID: fmt.Sprintf("i%d", i), Text: fmt.Sprintf("Item %d", i)})
	}
	vs.MoveDown(5)
	if vs.Cursor() != 5 {
		t.Errorf("expected 5, got %d", vs.Cursor())
	}
}

func TestVirtualScroller_MoveUp(t *testing.T) {
	vs := NewVirtualScroller()
	for i := 0; i < 20; i++ {
		vs.AddItem(VirtualItem{ID: fmt.Sprintf("i%d", i), Text: fmt.Sprintf("Item %d", i)})
	}
	vs.SetCursor(10)
	vs.MoveUp(3)
	if vs.Cursor() != 7 {
		t.Errorf("expected 7, got %d", vs.Cursor())
	}
}

func TestVirtualScroller_MoveDown_Clamp(t *testing.T) {
	vs := NewVirtualScroller()
	for i := 0; i < 5; i++ {
		vs.AddItem(VirtualItem{ID: fmt.Sprintf("i%d", i), Text: fmt.Sprintf("Item %d", i)})
	}
	vs.MoveDown(100)
	if vs.Cursor() != 4 {
		t.Errorf("expected 4, got %d", vs.Cursor())
	}
}

func TestVirtualScroller_MoveToStart(t *testing.T) {
	vs := NewVirtualScroller()
	for i := 0; i < 20; i++ {
		vs.AddItem(VirtualItem{ID: fmt.Sprintf("i%d", i), Text: fmt.Sprintf("Item %d", i)})
	}
	vs.SetCursor(10)
	vs.MoveToStart()
	if vs.Cursor() != 0 || vs.ScrollY() != 0 {
		t.Error("should be at start")
	}
}

func TestVirtualScroller_MoveToEnd(t *testing.T) {
	vs := NewVirtualScroller()
	for i := 0; i < 100; i++ {
		vs.AddItem(VirtualItem{ID: fmt.Sprintf("i%d", i), Text: fmt.Sprintf("Item %d", i)})
	}
	vs.MoveToEnd()
	if vs.Cursor() != 99 {
		t.Errorf("expected 99, got %d", vs.Cursor())
	}
}

func TestVirtualScroller_CurrentItem(t *testing.T) {
	vs := NewVirtualScroller()
	vs.AddItem(VirtualItem{ID: "a", Text: "A"})
	if vs.CurrentItem().ID != "a" {
		t.Error("expected a")
	}
	vs.Clear()
	if vs.CurrentItem() != nil {
		t.Error("expected nil when empty")
	}
}

func TestVirtualScroller_ScrollTo(t *testing.T) {
	vs := NewVirtualScroller()
	for i := 0; i < 100; i++ {
		vs.AddItem(VirtualItem{ID: fmt.Sprintf("i%d", i), Text: fmt.Sprintf("Item %d", i)})
	}
	vs.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	vs.ScrollTo(50)
	// scrollY should be clamped to maxScroll
	maxS := 100 - 10 + 2 // viewportHeight = 10-2=8 (border), maxScroll=100-8=92
	if vs.ScrollY() < 0 || vs.ScrollY() > maxS {
		t.Errorf("scrollY %d out of range [0, %d]", vs.ScrollY(), maxS)
	}
}

func TestVirtualScroller_VisibleRange(t *testing.T) {
	vs := NewVirtualScroller()
	for i := 0; i < 100; i++ {
		vs.AddItem(VirtualItem{ID: fmt.Sprintf("i%d", i), Text: fmt.Sprintf("Item %d", i)})
	}
	vs.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	vs.ScrollTo(20)
	buf := buffer.NewBuffer(40, 10)
	vs.Paint(buf) // triggers visibleRange computation
	start, end := vs.VisibleRange()
	if start != 20 {
		t.Errorf("expected start 20, got %d", start)
	}
	if end <= start {
		t.Errorf("end should be > start, got start=%d end=%d", start, end)
	}
}

func TestVirtualScroller_VisibleItems(t *testing.T) {
	vs := NewVirtualScroller()
	for i := 0; i < 100; i++ {
		vs.AddItem(VirtualItem{ID: fmt.Sprintf("i%d", i), Text: fmt.Sprintf("Item %d", i)})
	}
	vs.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	vs.Paint(buf)
	vis := vs.VisibleItems()
	if len(vis) == 0 {
		t.Error("expected visible items")
	}
	if len(vis) > 10 {
		t.Errorf("too many visible items: %d", len(vis))
	}
}

func TestVirtualScroller_VisibleItems_Empty(t *testing.T) {
	vs := NewVirtualScroller()
	if vs.VisibleItems() != nil {
		t.Error("expected nil for empty")
	}
}

func TestVirtualScroller_Filter(t *testing.T) {
	vs := NewVirtualScroller()
	vs.AddItems([]VirtualItem{
		{ID: "apple", Text: "Apple"},
		{ID: "banana", Text: "Banana"},
		{ID: "apricot", Text: "Apricot"},
		{ID: "cherry", Text: "Cherry"},
	})
	indices := vs.Filter("ap")
	if len(indices) != 2 {
		t.Errorf("expected 2 matches, got %d", len(indices))
	}
}

func TestVirtualScroller_Filter_EmptyQuery(t *testing.T) {
	vs := NewVirtualScroller()
	vs.AddItems([]VirtualItem{
		{ID: "a", Text: "A"},
		{ID: "b", Text: "B"},
	})
	indices := vs.Filter("")
	if len(indices) != 2 {
		t.Errorf("expected 2 for empty query, got %d", len(indices))
	}
}

func TestVirtualScroller_SetStyle(t *testing.T) {
	vs := NewVirtualScroller()
	s := DefaultVirtualScrollerStyle()
	s.Normal = buffer.Style{Fg: buffer.RGB(255, 0, 0)}
	vs.SetStyle(s)
	if vs.Style().Normal.Fg != s.Normal.Fg {
		t.Error("style mismatch")
	}
}

func TestVirtualScroller_SetHeader(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetHeader("My List")
	if vs.Header() != "My List" {
		t.Error("header mismatch")
	}
}

func TestVirtualScroller_SetShowScrollbar(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetShowScrollbar(false)
	// Just verify no panic
}

func TestVirtualScroller_Measure(t *testing.T) {
	vs := NewVirtualScroller()
	sz := vs.Measure(Bounded(50, 20))
	if sz.W != 50 || sz.H != 20 {
		t.Errorf("expected 50x20, got %dx%d", sz.W, sz.H)
	}
}

func TestVirtualScroller_Measure_Unbounded(t *testing.T) {
	vs := NewVirtualScroller()
	sz := vs.Measure(Unbounded())
	if sz.W != 40 || sz.H != 10 {
		t.Errorf("expected default 40x10, got %dx%d", sz.W, sz.H)
	}
}

func TestVirtualScroller_Paint_NoPanic(t *testing.T) {
	vs := NewVirtualScroller()
	for i := 0; i < 50; i++ {
		vs.AddItem(VirtualItem{ID: fmt.Sprintf("i%d", i), Text: fmt.Sprintf("Item %d", i)})
	}
	vs.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	vs.Paint(buf)
}

func TestVirtualScroller_Paint_ZeroBounds(t *testing.T) {
	vs := NewVirtualScroller()
	vs.AddItem(VirtualItem{ID: "a", Text: "A"})
	vs.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 10)
	vs.Paint(buf)
}

func TestVirtualScroller_Paint_Empty(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	vs.Paint(buf)
}

func TestVirtualScroller_Paint_LargeDataset(t *testing.T) {
	vs := NewVirtualScroller()
	for i := 0; i < 10000; i++ {
		vs.AddItem(VirtualItem{ID: fmt.Sprintf("i%d", i), Text: fmt.Sprintf("Item %d", i)})
	}
	vs.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20}) // bounds first
	vs.SetCursor(5000) // then move cursor - scroll will follow
	buf := buffer.NewBuffer(60, 20)
	vs.Paint(buf)
	start, end := vs.VisibleRange()
	if start > 5000 || end < 5000 {
		t.Errorf("cursor 5000 not in visible range [%d, %d)", start, end)
	}
}

func TestVirtualScroller_Children(t *testing.T) {
	vs := NewVirtualScroller()
	if vs.Children() != nil {
		t.Error("should have no children")
	}
}

func TestVirtualScroller_String(t *testing.T) {
	vs := NewVirtualScroller()
	vs.AddItem(VirtualItem{ID: "a", Text: "A"})
	s := vs.String()
	if s == "" {
		t.Error("String() should not be empty")
	}
}

func TestVirtualScroller_EnsureCursorVisible(t *testing.T) {
	vs := NewVirtualScroller()
	for i := 0; i < 100; i++ {
		vs.AddItem(VirtualItem{ID: fmt.Sprintf("i%d", i), Text: fmt.Sprintf("Item %d", i)})
	}
	vs.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	// Move cursor far down — scroll should follow
	vs.SetCursor(50)
	buf := buffer.NewBuffer(40, 10)
	vs.Paint(buf)
	start, _ := vs.VisibleRange()
	if start > 50 {
		t.Errorf("scroll should include cursor 50, visibleStart=%d", start)
	}
}

func TestVirtualScroller_Concurrent(t *testing.T) {
	vs := NewVirtualScroller()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				vs.AddItem(VirtualItem{ID: fmt.Sprintf("g%d-i%d", n, j), Text: "X"})
				vs.MoveDown(1)
				_ = vs.Items()
				_ = vs.Cursor()
				_ = vs.ItemCount()
			}
		}(i)
	}
	wg.Wait()
}

func TestVirtualScroller_ConcurrentPaint(t *testing.T) {
	vs := NewVirtualScroller()
	for i := 0; i < 100; i++ {
		vs.AddItem(VirtualItem{ID: fmt.Sprintf("i%d", i), Text: fmt.Sprintf("Item %d", i)})
	}
	vs.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				buf := buffer.NewBuffer(40, 10)
				vs.Paint(buf)
				vs.MoveDown(1)
			}
		}()
	}
	wg.Wait()
}

// ===== Pagination Tests =====

func TestNewPagination(t *testing.T) {
	p := NewPagination()
	if p.TotalPages() != 0 {
		t.Error("expected 0 pages")
	}
	if p.CurrentPage() != 0 {
		t.Error("expected page 0")
	}
	if p.ItemsPerPage() != 20 {
		t.Error("expected 20 per page default")
	}
}

func TestPagination_SetTotalItems(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	if p.TotalPages() != 5 {
		t.Errorf("expected 5 pages, got %d", p.TotalPages())
	}
}

func TestPagination_SetItemsPerPage(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(25)
	if p.TotalPages() != 4 {
		t.Errorf("expected 4 pages, got %d", p.TotalPages())
	}
}

func TestPagination_SetItemsPerPage_Zero(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(0) // should be ignored
	if p.ItemsPerPage() != 20 {
		t.Error("0 perPage should be ignored")
	}
}

func TestPagination_SetTotalPages(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(10)
	if p.TotalPages() != 10 {
		t.Errorf("expected 10, got %d", p.TotalPages())
	}
}

func TestPagination_SetPage(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(10)
	p.SetPage(5)
	if p.CurrentPage() != 5 {
		t.Errorf("expected 5, got %d", p.CurrentPage())
	}
}

func TestPagination_SetPage_Clamp(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(5)
	p.SetPage(100)
	if p.CurrentPage() != 4 {
		t.Errorf("expected 4, got %d", p.CurrentPage())
	}
	p.SetPage(-5)
	if p.CurrentPage() != 0 {
		t.Errorf("expected 0, got %d", p.CurrentPage())
	}
}

func TestPagination_NextPage(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(5)
	if !p.NextPage() {
		t.Error("expected true")
	}
	if p.CurrentPage() != 1 {
		t.Errorf("expected 1, got %d", p.CurrentPage())
	}
}

func TestPagination_NextPage_Last(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(3)
	p.SetPage(2)
	if p.NextPage() {
		t.Error("expected false on last page")
	}
}

func TestPagination_PrevPage(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(5)
	p.SetPage(3)
	if !p.PrevPage() {
		t.Error("expected true")
	}
	if p.CurrentPage() != 2 {
		t.Errorf("expected 2, got %d", p.CurrentPage())
	}
}

func TestPagination_PrevPage_First(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(3)
	if p.PrevPage() {
		t.Error("expected false on first page")
	}
}

func TestPagination_FirstPage(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(5)
	p.SetPage(3)
	p.FirstPage()
	if p.CurrentPage() != 0 {
		t.Error("expected 0")
	}
}

func TestPagination_LastPage(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(5)
	p.LastPage()
	if p.CurrentPage() != 4 {
		t.Error("expected 4")
	}
}

func TestPagination_HasNext(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(3)
	if !p.HasNext() {
		t.Error("should have next")
	}
	p.SetPage(2)
	if p.HasNext() {
		t.Error("should not have next on last")
	}
}

func TestPagination_HasPrev(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(3)
	if p.HasPrev() {
		t.Error("should not have prev on first")
	}
	p.SetPage(1)
	if !p.HasPrev() {
		t.Error("should have prev")
	}
}

func TestPagination_IsEmpty(t *testing.T) {
	p := NewPagination()
	if !p.IsEmpty() {
		t.Error("should be empty")
	}
	p.SetTotalPages(1)
	if p.IsEmpty() {
		t.Error("should not be empty")
	}
}

func TestPagination_PageStartIndex(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(20)
	p.SetPage(2)
	if p.PageStartIndex() != 40 {
		t.Errorf("expected 40, got %d", p.PageStartIndex())
	}
}

func TestPagination_PageEndIndex(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(95)
	p.SetItemsPerPage(20)
	p.SetPage(4) // last page (0-indexed)
	if p.PageEndIndex() != 95 {
		t.Errorf("expected 95, got %d", p.PageEndIndex())
	}
}

func TestPagination_CurrentPageItems(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(20)
	p.SetPage(1)
	items := p.CurrentPageItems()
	if len(items) != 20 {
		t.Errorf("expected 20, got %d", len(items))
	}
	if items[0] != 20 {
		t.Errorf("expected first index 20, got %d", items[0])
	}
}

func TestPagination_OnPageChange(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(5)
	changed := -1
	p.OnPageChange = func(page int) { changed = page }
	p.NextPage()
	if changed != 1 {
		t.Errorf("expected callback with 1, got %d", changed)
	}
}

func TestPagination_SetPageRange(t *testing.T) {
	p := NewPagination()
	p.SetPageRange(5)
	if p.PageRange() != 5 {
		t.Error("expected 5")
	}
	p.SetPageRange(-1)
	if p.PageRange() != 5 {
		t.Error("negative should be ignored")
	}
}

func TestPagination_SetStyle(t *testing.T) {
	p := NewPagination()
	s := DefaultPaginationStyle()
	s.Selected = buffer.Style{Fg: buffer.RGB(255, 0, 0)}
	p.SetStyle(s)
	if p.Style().Selected.Fg != s.Selected.Fg {
		t.Error("style mismatch")
	}
}

func TestPagination_Measure(t *testing.T) {
	p := NewPagination()
	sz := p.Measure(Bounded(60, 5))
	if sz.W != 60 || sz.H != 1 {
		t.Errorf("expected 60x1, got %dx%d", sz.W, sz.H)
	}
}

func TestPagination_Paint_NoPanic(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(10)
	p.SetPage(5)
	p.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	p.Paint(buf)
}

func TestPagination_Paint_Empty(t *testing.T) {
	p := NewPagination()
	p.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	p.Paint(buf)
}

func TestPagination_Paint_ZeroBounds(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(5)
	p.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 10)
	p.Paint(buf)
}

func TestPagination_Children(t *testing.T) {
	p := NewPagination()
	if p.Children() != nil {
		t.Error("should have no children")
	}
}

func TestPagination_String(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(5)
	p.SetPage(2)
	s := p.String()
	if s == "" {
		t.Error("String() should not be empty")
	}
}

func TestPagination_Concurrent(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(20)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				p.NextPage()
				p.PrevPage()
				_ = p.CurrentPage()
				_ = p.TotalPages()
			}
		}()
	}
	wg.Wait()
}

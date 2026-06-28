package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P21-A: Comprehensive Pagination tests covering all public API gaps.
// Existing tests in p17_virtualscroller_test.go cover basic navigation;
// these tests focus on getters, edge cases, paint output, and concurrency.
// Pages are 0-indexed internally.

// ── Getters ──

func TestPagination_ItemsPerPage_Getter(t *testing.T) {
	p := NewPagination()
	p.SetItemsPerPage(25)
	if got := p.ItemsPerPage(); got != 25 {
		t.Errorf("ItemsPerPage() = %d, want 25", got)
	}
}

func TestPagination_TotalItems_Getter(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(500)
	if got := p.TotalItems(); got != 500 {
		t.Errorf("TotalItems() = %d, want 500", got)
	}
}

func TestPagination_PageRange_Getter(t *testing.T) {
	p := NewPagination()
	p.SetPageRange(7)
	if got := p.PageRange(); got != 7 {
		t.Errorf("PageRange() = %d, want 7", got)
	}
}

func TestPagination_Style_Getter(t *testing.T) {
	p := NewPagination()
	s := DefaultPaginationStyle()
	s.Selected.Flags |= buffer.Bold
	p.SetStyle(s)
	got := p.Style()
	if got.Selected.Flags&buffer.Bold == 0 {
		t.Error("Style().Selected should have Bold flag set")
	}
}

// ── Edge Cases ──

func TestPagination_SetTotalItems_Zero(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)
	p.SetPage(3)
	p.SetTotalItems(0) // Reset to 0
	if !p.IsEmpty() {
		t.Error("IsEmpty() should be true after SetTotalItems(0)")
	}
	if p.TotalPages() != 0 {
		t.Errorf("TotalPages() = %d, want 0", p.TotalPages())
	}
}

func TestPagination_SetItemsPerPage_Negative(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)
	p.SetItemsPerPage(-5) // Should be ignored
	perPage := p.ItemsPerPage()
	if perPage < 1 {
		t.Errorf("ItemsPerPage() = %d, should be at least 1", perPage)
	}
}

func TestPagination_SetPage_Negative(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)
	p.SetPage(-5)
	// 0-indexed: negative clamps to page 0
	if p.CurrentPage() != 0 {
		t.Errorf("CurrentPage() = %d, want 0 (negative clamped)", p.CurrentPage())
	}
}

func TestPagination_SetPageRange_Negative(t *testing.T) {
	p := NewPagination()
	p.SetPageRange(-3)
	if pr := p.PageRange(); pr < 0 {
		t.Errorf("PageRange() = %d, should be non-negative", pr)
	}
}

func TestPagination_SetTotalPages_Overrides(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)
	if p.TotalPages() != 10 {
		t.Errorf("TotalPages() = %d, want 10", p.TotalPages())
	}
	p.SetTotalPages(5)
	if p.TotalPages() != 5 {
		t.Errorf("TotalPages() = %d, want 5 after override", p.TotalPages())
	}
}

func TestPagination_PageStartIndex_MiddlePage(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)
	p.SetPage(3)
	// 0-indexed: page 3 starts at index 30 (3*10)
	if got := p.PageStartIndex(); got != 30 {
		t.Errorf("PageStartIndex() = %d, want 30", got)
	}
}

func TestPagination_PageEndIndex_MiddlePage(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)
	p.SetPage(3)
	// 0-indexed: page 3 ends at index 40 (exclusive, clamped to totalItems=100)
	if got := p.PageEndIndex(); got != 40 {
		t.Errorf("PageEndIndex() = %d, want 40", got)
	}
}

func TestPagination_PageEndIndex_LastPartialPage(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(95)
	p.SetItemsPerPage(10)
	p.LastPage() // page 9 (0-indexed), items 90-94
	end := p.PageEndIndex()
	// Clamped to totalItems=95
	if end != 95 {
		t.Errorf("PageEndIndex() = %d, want 95 (clamped to totalItems)", end)
	}
}

func TestPagination_CurrentPageItems_LastPage(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(25)
	p.SetItemsPerPage(10)
	p.LastPage() // page 2 (0-indexed), items 20-24
	items := p.CurrentPageItems()
	if len(items) != 5 {
		t.Errorf("CurrentPageItems() len = %d, want 5", len(items))
	}
}

func TestPagination_CurrentPageItems_Empty(t *testing.T) {
	p := NewPagination()
	items := p.CurrentPageItems()
	if len(items) != 0 {
		t.Errorf("CurrentPageItems() on empty = %d items, want 0", len(items))
	}
}

// ── Navigation Callbacks ──

func TestPagination_NextPage_Callback(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)

	var calledPage int
	p.OnPageChange = func(page int) {
		calledPage = page
	}
	p.NextPage() // 0-indexed: page 0 → page 1
	if calledPage != 1 {
		t.Errorf("OnPageChange called with %d, want 1", calledPage)
	}
}

func TestPagination_PrevPage_Callback(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)
	p.SetPage(5)

	var calledPage int
	p.OnPageChange = func(page int) {
		calledPage = page
	}
	p.PrevPage() // page 5 → page 4
	if calledPage != 4 {
		t.Errorf("OnPageChange called with %d, want 4", calledPage)
	}
}

func TestPagination_FirstPage_Callback(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)
	p.SetPage(5)

	var called bool
	p.OnPageChange = func(page int) {
		called = true
	}
	p.FirstPage()
	if !called {
		t.Error("OnPageChange not called from FirstPage")
	}
}

func TestPagination_LastPage_Callback(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)

	var calledPage int
	p.OnPageChange = func(page int) {
		calledPage = page
	}
	p.LastPage() // 0-indexed: last page = 9
	if calledPage != 9 {
		t.Errorf("OnPageChange called with %d, want 9", calledPage)
	}
}

func TestPagination_MultipleNextPage(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)

	// 0-indexed: start at 0, 5 NextPage calls → page 5
	for i := 0; i < 5; i++ {
		p.NextPage()
	}
	if p.CurrentPage() != 5 {
		t.Errorf("CurrentPage() after 5 NextPage = %d, want 5", p.CurrentPage())
	}
}

func TestPagination_NextPage_AtLast(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)
	p.LastPage()
	if p.NextPage() {
		t.Error("NextPage() at last page should return false")
	}
}

func TestPagination_PrevPage_AtFirst(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)
	p.FirstPage()
	if p.PrevPage() {
		t.Error("PrevPage() at first page should return false")
	}
}

// ── Navigation on Empty ──

func TestPagination_NextPage_Empty(t *testing.T) {
	p := NewPagination()
	if p.NextPage() {
		t.Error("NextPage() on empty should return false")
	}
}

func TestPagination_PrevPage_Empty(t *testing.T) {
	p := NewPagination()
	if p.PrevPage() {
		t.Error("PrevPage() on empty should return false")
	}
}

func TestPagination_HasNext_Empty(t *testing.T) {
	p := NewPagination()
	if p.HasNext() {
		t.Error("HasNext() on empty should be false")
	}
}

func TestPagination_HasPrev_Empty(t *testing.T) {
	p := NewPagination()
	if p.HasPrev() {
		t.Error("HasPrev() on empty should be false")
	}
}

// ── Paint Rendering ──

func TestPagination_Paint_WithContent(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)
	p.SetPage(5)
	p.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	p.Paint(buf)
	hasContent := false
	for x := 0; x < 60; x++ {
		if buf.GetCell(x, 0).Rune != 0 {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Paint should render some content")
	}
}

func TestPagination_Paint_WithEllipsis(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(1000)
	p.SetItemsPerPage(10)
	p.SetPage(50)
	p.SetPageRange(3)
	p.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	buf := buffer.NewBuffer(80, 1)
	p.Paint(buf)
	hasContent := false
	for x := 0; x < 80; x++ {
		if buf.GetCell(x, 0).Rune != 0 {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Paint with many pages should render ellipsis content")
	}
}

func TestPagination_Paint_SinglePage(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(5)
	p.SetItemsPerPage(10) // Only 1 page
	p.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	p.Paint(buf) // should not panic
}

func TestPagination_Paint_WithCallback(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)
	p.SetPage(5)
	p.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})

	called := false
	p.OnPageChange = func(page int) { called = true }

	buf := buffer.NewBuffer(60, 1)
	p.Paint(buf)
	if called {
		t.Error("Paint should not trigger OnPageChange")
	}
}

// ── Concurrent Access ──

func TestPagination_Concurrent_Nav(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(1000)
	p.SetItemsPerPage(10)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.NextPage()
			p.CurrentPage()
			p.HasNext()
			p.HasPrev()
			p.PageStartIndex()
			p.PageEndIndex()
		}()
	}
	wg.Wait()
}

// ── Measure ──

func TestPagination_Measure_Bounded(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)
	p.SetPage(5)

	s := p.Measure(Constraints{MaxWidth: 80, MaxHeight: 3})
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("Measure returned zero size: %+v", s)
	}
}

func TestPagination_Measure_Unbounded(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)

	s := p.Measure(Constraints{})
	if s.W <= 0 {
		t.Errorf("Measure unbounded returned zero width: %+v", s)
	}
}

func TestPagination_RecomputePages(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(20)
	if p.TotalPages() != 5 {
		t.Errorf("TotalPages() = %d, want 5", p.TotalPages())
	}
	p.SetItemsPerPage(50)
	if p.TotalPages() != 2 {
		t.Errorf("TotalPages() after resize = %d, want 2", p.TotalPages())
	}
}

func TestPagination_RecomputePages_PageClamped(t *testing.T) {
	p := NewPagination()
	p.SetTotalPages(10)
	p.SetPage(8)
	p.SetTotalPages(3)
	if p.CurrentPage() > 2 {
		t.Errorf("CurrentPage() = %d, should be clamped to <= 2", p.CurrentPage())
	}
}

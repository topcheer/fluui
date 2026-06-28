package component

import (
	"fmt"
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ============================================================
// P18-B: AutoComplete Component Tests
// ============================================================

func newACItems(n int) []CompletionItem {
	items := make([]CompletionItem, n)
	for i := 0; i < n; i++ {
		items[i] = CompletionItem{
			Label: fmt.Sprintf("Item %d", i),
			Value: fmt.Sprintf("item-%d", i),
		}
	}
	return items
}

// ─── Construction ────────────────────────────────────────────────

func TestAutoComplete_New(t *testing.T) {
	ac := NewAutoComplete()
	if ac == nil {
		t.Fatal("NewAutoComplete returned nil")
	}
	if ac.ID() == "" {
		t.Error("ID should not be empty")
	}
	if ac.Visible() {
		t.Error("should start invisible")
	}
	if ac.ItemCount() != 0 {
		t.Error("should start with no items")
	}
	if ac.FilteredCount() != 0 {
		t.Error("should start with no filtered results")
	}
	if ac.MaxVisible() != 10 {
		t.Errorf("MaxVisible = %d, want 10", ac.MaxVisible())
	}
}

func TestAutoComplete_UniqueID(t *testing.T) {
	ac1 := NewAutoComplete()
	ac2 := NewAutoComplete()
	if ac1.ID() == ac2.ID() {
		t.Error("IDs should be unique")
	}
}

func TestAutoComplete_ImplementsComponent(t *testing.T) {
	var _ Component = NewAutoComplete()
}

// ─── Item management ─────────────────────────────────────────────

func TestAutoComplete_SetItems(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(5))
	if ac.ItemCount() != 5 {
		t.Errorf("ItemCount = %d, want 5", ac.ItemCount())
	}
}

func TestAutoComplete_SetItems_EmptyQueryShowsAll(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(5))
	if ac.FilteredCount() != 5 {
		t.Errorf("FilteredCount = %d, want 5 for empty query", ac.FilteredCount())
	}
}

func TestAutoComplete_AddItem(t *testing.T) {
	ac := NewAutoComplete()
	ac.AddItem(CompletionItem{Label: "hello", Value: "hello"})
	ac.AddItem(CompletionItem{Label: "world", Value: "world"})
	if ac.ItemCount() != 2 {
		t.Errorf("ItemCount = %d, want 2", ac.ItemCount())
	}
}

func TestAutoComplete_ItemsReturnsCopy(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(3))
	items := ac.Items()
	items[0].Label = "mutated"
	if ac.Items()[0].Label == "mutated" {
		t.Error("Items() should return a copy")
	}
}

func TestAutoComplete_Clear(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(5))
	ac.Clear()
	if ac.ItemCount() != 0 {
		t.Errorf("ItemCount = %d after clear", ac.ItemCount())
	}
	if ac.FilteredCount() != 0 {
		t.Errorf("FilteredCount = %d after clear", ac.FilteredCount())
	}
}

// ─── Query & filtering ───────────────────────────────────────────

func TestAutoComplete_SetQuery_FuzzyMatch(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "golang", Value: "golang"},
		{Label: "python", Value: "python"},
		{Label: "groovy", Value: "groovy"},
		{Label: "ruby", Value: "ruby"},
	})
	ac.SetQuery("go")
	if ac.FilteredCount() < 2 {
		t.Errorf("expected >= 2 results for 'go', got %d", ac.FilteredCount())
	}
}

func TestAutoComplete_SetQuery_NoMatch(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "apple"},
		{Label: "banana"},
	})
	ac.SetQuery("xyz")
	if ac.HasResults() {
		t.Error("expected no results for 'xyz'")
	}
}

func TestAutoComplete_SetQuery_EmptyMatchesAll(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(10))
	ac.SetQuery("")
	if ac.FilteredCount() != 10 {
		t.Errorf("FilteredCount = %d for empty query, want 10", ac.FilteredCount())
	}
}

// ─── Cursor navigation ───────────────────────────────────────────

func TestAutoComplete_MoveDown(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(5))
	if ac.Cursor() != 0 {
		t.Errorf("initial cursor = %d, want 0", ac.Cursor())
	}
	ac.MoveDown()
	if ac.Cursor() != 1 {
		t.Errorf("after MoveDown cursor = %d, want 1", ac.Cursor())
	}
}

func TestAutoComplete_MoveDown_WrapAround(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(3))
	ac.SetCursor(2)
	ac.MoveDown()
	if ac.Cursor() != 0 {
		t.Errorf("after wrap cursor = %d, want 0", ac.Cursor())
	}
}

func TestAutoComplete_MoveUp(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(5))
	ac.SetCursor(2)
	ac.MoveUp()
	if ac.Cursor() != 1 {
		t.Errorf("after MoveUp cursor = %d, want 1", ac.Cursor())
	}
}

func TestAutoComplete_MoveUp_WrapAround(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(3))
	ac.MoveUp()
	if ac.Cursor() != 2 {
		t.Errorf("after wrap cursor = %d, want 2", ac.Cursor())
	}
}

func TestAutoComplete_SetCursor_Clamp(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(3))
	ac.SetCursor(-5)
	if ac.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 (clamped)", ac.Cursor())
	}
	ac.SetCursor(100)
	if ac.Cursor() != 2 {
		t.Errorf("cursor = %d, want 2 (clamped)", ac.Cursor())
	}
}

func TestAutoComplete_MoveDown_EmptyResults(t *testing.T) {
	ac := NewAutoComplete()
	ac.MoveDown()
	if ac.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 for empty", ac.Cursor())
	}
}

func TestAutoComplete_CurrentItem(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "first", Value: "first"},
		{Label: "second", Value: "second"},
	})
	item := ac.CurrentItem()
	if item == nil {
		t.Fatal("CurrentItem should not be nil")
	}
	if item.Value != "first" {
		t.Errorf("CurrentItem.Value = %q, want 'first'", item.Value)
	}
	ac.MoveDown()
	item = ac.CurrentItem()
	if item == nil || item.Value != "second" {
		t.Errorf("CurrentItem after MoveDown = %v", item)
	}
}

func TestAutoComplete_CurrentItem_Empty(t *testing.T) {
	ac := NewAutoComplete()
	if ac.CurrentItem() != nil {
		t.Error("CurrentItem should be nil when empty")
	}
}

// ─── Visibility & position ───────────────────────────────────────

func TestAutoComplete_ShowHide(t *testing.T) {
	ac := NewAutoComplete()
	if ac.Visible() {
		t.Error("should start hidden")
	}
	ac.Show(10, 20)
	if !ac.Visible() {
		t.Error("should be visible after Show")
	}
	x, y := ac.Position()
	if x != 10 || y != 20 {
		t.Errorf("Position = (%d,%d), want (10,20)", x, y)
	}
	ac.Hide()
	if ac.Visible() {
		t.Error("should be hidden after Hide")
	}
}

func TestAutoComplete_SetPosition(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetPosition(5, 15)
	x, y := ac.Position()
	if x != 5 || y != 15 {
		t.Errorf("Position = (%d,%d), want (5,15)", x, y)
	}
}

// ─── Configuration ───────────────────────────────────────────────

func TestAutoComplete_SetMaxVisible(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetMaxVisible(5)
	if ac.MaxVisible() != 5 {
		t.Errorf("MaxVisible = %d, want 5", ac.MaxVisible())
	}
}

func TestAutoComplete_SetMaxVisible_ClampToMin1(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetMaxVisible(0)
	if ac.MaxVisible() != 1 {
		t.Errorf("MaxVisible = %d, want 1 (clamped)", ac.MaxVisible())
	}
}

func TestAutoComplete_SetStyle(t *testing.T) {
	ac := NewAutoComplete()
	custom := AutoCompleteStyle{
		Normal:   buffer.Style{Flags: buffer.Bold},
		Selected: buffer.Style{Flags: buffer.Reverse | buffer.Bold},
	}
	ac.SetStyle(custom)
	if ac.Style().Normal.Flags != buffer.Bold {
		t.Error("style not set correctly")
	}
}

func TestAutoComplete_DefaultAutoCompleteStyle(t *testing.T) {
	s := DefaultAutoCompleteStyle()
	// Just verify it doesn't panic and returns something
	_ = s
}

func TestAutoComplete_SetCaseSensitive(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "Apple"},
		{Label: "apple"},
		{Label: "BANANA"},
	})

	// Case insensitive (default): "a" matches both Apple and apple
	ac.SetQuery("a")
	if ac.FilteredCount() < 2 {
		t.Errorf("case-insensitive: expected >= 2, got %d", ac.FilteredCount())
	}

	// Case sensitive: "A" only matches Apple
	ac.SetCaseSensitive(true)
	ac.SetQuery("A")
	items := ac.FilteredItems()
	foundLower := false
	for _, it := range items {
		if it.Label == "apple" {
			foundLower = true
		}
	}
	if foundLower {
		t.Error("should not match 'apple' in case-sensitive mode for 'A'")
	}
}

// ─── Selection ───────────────────────────────────────────────────

func TestAutoComplete_Select(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "first", Value: "first"},
		{Label: "second", Value: "second"},
	})

	var selected CompletionItem
	ac.SetOnSelect(func(item CompletionItem) {
		selected = item
	})

	ac.Show(0, 0)
	ac.Select()
	if selected.Value != "first" {
		t.Errorf("selected = %q, want 'first'", selected.Value)
	}
	if ac.Visible() {
		t.Error("should be hidden after Select")
	}
}

func TestAutoComplete_Select_Empty(t *testing.T) {
	ac := NewAutoComplete()
	ac.Select() // should not panic
}

func TestAutoComplete_Dismiss(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(3))
	ac.Show(0, 0)

	var dismissed bool
	ac.SetOnDismiss(func() {
		dismissed = true
	})

	ac.Hide()
	if !dismissed {
		t.Error("OnDismiss should have been called")
	}
	if ac.Visible() {
		t.Error("should be hidden after Hide")
	}
}

// ─── Key handling ────────────────────────────────────────────────

func TestAutoComplete_HandleKey_Navigation(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(5))

	if !ac.HandleKey(&term.KeyEvent{Key: term.KeyDown}) {
		t.Error("HandleKey Down should return true")
	}
	if ac.Cursor() != 1 {
		t.Errorf("cursor = %d after Down, want 1", ac.Cursor())
	}
	if !ac.HandleKey(&term.KeyEvent{Key: term.KeyUp}) {
		t.Error("HandleKey Up should return true")
	}
	if ac.Cursor() != 0 {
		t.Errorf("cursor = %d after Up, want 0", ac.Cursor())
	}
}

func TestAutoComplete_HandleKey_Tab(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{{Label: "hello", Value: "hello"}})
	ac.Show(0, 0)

	var selected CompletionItem
	ac.SetOnSelect(func(item CompletionItem) {
		selected = item
	})

	if !ac.HandleKey(&term.KeyEvent{Key: term.KeyTab}) {
		t.Error("HandleKey Tab should return true")
	}
	if selected.Value != "hello" {
		t.Errorf("selected = %q, want 'hello'", selected.Value)
	}
}

func TestAutoComplete_HandleKey_Enter(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{{Label: "world", Value: "world"}})

	if !ac.HandleKey(&term.KeyEvent{Key: term.KeyEnter}) {
		t.Error("HandleKey Enter should return true")
	}
}

func TestAutoComplete_HandleKey_Escape(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(3))
	ac.Show(0, 0)

	if !ac.HandleKey(&term.KeyEvent{Key: term.KeyEscape}) {
		t.Error("HandleKey Escape should return true")
	}
	if ac.Visible() {
		t.Error("should be hidden after Escape")
	}
}

func TestAutoComplete_HandleKey_Unhandled(t *testing.T) {
	ac := NewAutoComplete()
	if ac.HandleKey(&term.KeyEvent{Key: term.KeyLeft}) {
		t.Error("HandleKey should return false for unhandled key")
	}
}

func TestAutoComplete_HandleKey_Nil(t *testing.T) {
	ac := NewAutoComplete()
	if ac.HandleKey(nil) {
		t.Error("HandleKey(nil) should return false")
	}
}

// ─── Measure ─────────────────────────────────────────────────────

func TestAutoComplete_Measure_Basic(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "short"},
		{Label: "longer item name"},
	})
	size := ac.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.W < 20 {
		t.Errorf("width = %d, want >= 20", size.W)
	}
	if size.H < 4 {
		t.Errorf("height = %d, want >= 4", size.H)
	}
}

func TestAutoComplete_Measure_ClampedToMaxVisible(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(50))
	ac.SetMaxVisible(5)
	size := ac.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.H > 7 {
		t.Errorf("height = %d, want <= 7", size.H)
	}
}

func TestAutoComplete_Measure_Empty(t *testing.T) {
	ac := NewAutoComplete()
	size := ac.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.W < 15 {
		t.Errorf("width = %d for empty, want >= 15", size.W)
	}
	if size.H != 3 {
		t.Errorf("height = %d for empty, want 3", size.H)
	}
}

func TestAutoComplete_Measure_ClampedToConstraints(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(20))
	size := ac.Measure(Constraints{MaxWidth: 10, MaxHeight: 5})
	if size.W > 10 {
		t.Errorf("width = %d, want <= 10", size.W)
	}
	if size.H > 5 {
		t.Errorf("height = %d, want <= 5", size.H)
	}
}

// ─── Paint ───────────────────────────────────────────────────────

func TestAutoComplete_Paint_NoPanic(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(5))
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	ac.Show(0, 0)
	ac.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	buf := buffer.NewBuffer(30, 10)
	ac.Paint(buf) // should not panic
}

func TestAutoComplete_Paint_ZeroBounds(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(5))
	buf := buffer.NewBuffer(30, 10)
	ac.Paint(buf) // should not panic
}

func TestAutoComplete_Paint_Invisible(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(5))
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	ac.Paint(buf) // should not panic, should not draw
}

func TestAutoComplete_Paint_RendersBorder(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "alpha"},
		{Label: "beta"},
		{Label: "gamma"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 8})
	ac.Show(0, 0)
	ac.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	buf := buffer.NewBuffer(20, 8)
	ac.Paint(buf)
	// Check top-left corner was drawn
	if buf.GetCell(0, 0).Rune != '\u250c' {
		t.Errorf("top-left corner = %q, want '\\u250c'", string(buf.GetCell(0, 0).Rune))
	}
}

// ─── Scrolling ───────────────────────────────────────────────────

func TestAutoComplete_ScrollY(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(30))
	ac.SetMaxVisible(5)
	if ac.ScrollY() != 0 {
		t.Errorf("initial ScrollY = %d, want 0", ac.ScrollY())
	}
	ac.SetCursor(10)
	if ac.ScrollY() != 6 {
		t.Errorf("ScrollY = %d, want 6", ac.ScrollY())
	}
}

func TestAutoComplete_ScrollY_Up(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(30))
	ac.SetMaxVisible(5)
	ac.SetCursor(10)
	ac.SetCursor(2)
	if ac.ScrollY() != 2 {
		t.Errorf("ScrollY = %d, want 2", ac.ScrollY())
	}
}

// ─── Misc ────────────────────────────────────────────────────────

func TestAutoComplete_Children(t *testing.T) {
	ac := NewAutoComplete()
	if ac.Children() != nil {
		t.Error("Children should return nil")
	}
}

func TestAutoComplete_String(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(3))
	if ac.String() == "" {
		t.Error("String should not be empty")
	}
}

func TestAutoComplete_FilteredItems(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "apple"},
		{Label: "apricot"},
		{Label: "banana"},
	})
	ac.SetQuery("ap")
	items := ac.FilteredItems()
	if len(items) < 2 {
		t.Errorf("expected >= 2 filtered items, got %d", len(items))
	}
}

// ─── Concurrency ─────────────────────────────────────────────────

func TestAutoComplete_ConcurrentAccess(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(100))

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				ac.SetQuery(fmt.Sprintf("Item %d", (n*50+j)%100))
				ac.MoveDown()
				ac.Cursor()
				ac.HasResults()
				ac.CurrentItem()
			}
		}(i)
	}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				ac.FilteredItems()
				ac.FilteredCount()
				ac.Items()
			}
		}()
	}
	wg.Wait()
}

func TestAutoComplete_ConcurrentPaint(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems(newACItems(50))
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 12})
	ac.Show(0, 0)
	ac.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			buf := buffer.NewBuffer(30, 12)
			ac.Paint(buf)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			ac.SetQuery(fmt.Sprintf("Item %d", i%50))
			ac.MoveDown()
		}
	}()
	wg.Wait()
}

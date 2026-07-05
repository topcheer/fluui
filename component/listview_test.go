package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestListView_New(t *testing.T) {
	lv := NewListView([]string{"Apple", "Banana", "Cherry"})
	if lv == nil {
		t.Fatal("NewListView returned nil")
	}
	if lv.ItemCount() != 3 {
		t.Errorf("ItemCount = %d, want 3", lv.ItemCount())
	}
	if lv.Cursor() != 0 {
		t.Errorf("Cursor = %d, want 0", lv.Cursor())
	}
	if lv.ID() == "" {
		t.Error("ID should not be empty")
	}
}

func TestListView_Items(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	items := lv.Items()
	if len(items) != 3 {
		t.Fatalf("len(Items) = %d, want 3", len(items))
	}
	if items[0].Label != "A" {
		t.Errorf("items[0].Label = %q, want %q", items[0].Label, "A")
	}

	// Verify it's a copy
	items[0].Label = "X"
	items2 := lv.Items()
	if items2[0].Label != "A" {
		t.Error("Items() did not return a copy")
	}
}

func TestListView_SetItems(t *testing.T) {
	lv := NewListView([]string{"A", "B"})
	lv.SetCursor(1)
	lv.SetItems([]ListItem{
		{Label: "X"},
		{Label: "Y"},
		{Label: "Z"},
	})
	if lv.ItemCount() != 3 {
		t.Errorf("ItemCount = %d, want 3", lv.ItemCount())
	}
	if lv.Cursor() != 1 {
		t.Errorf("Cursor = %d, want 1 (clamped)", lv.Cursor())
	}
}

func TestListView_SetItems_Shrink(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C", "D", "E"})
	lv.SetCursor(4)
	lv.SetItems([]ListItem{{Label: "X"}})
	if lv.Cursor() != 0 {
		t.Errorf("Cursor = %d, want 0 after shrink", lv.Cursor())
	}
}

func TestListView_AddItem(t *testing.T) {
	lv := NewListView([]string{"A"})
	lv.AddItem("B", 2)
	if lv.ItemCount() != 2 {
		t.Errorf("ItemCount = %d, want 2", lv.ItemCount())
	}
	items := lv.Items()
	if items[1].Label != "B" {
		t.Errorf("items[1].Label = %q, want %q", items[1].Label, "B")
	}
}

func TestListView_AddItemWithIcon(t *testing.T) {
	lv := NewListView([]string{"A"})
	lv.AddItemWithIcon("Folder", "/path", '*')
	items := lv.Items()
	if items[1].Icon != '*' {
		t.Errorf("Icon = %v, want %q", items[1].Icon, '*')
	}
}

func TestListView_RemoveItem(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	lv.RemoveItem(1)
	if lv.ItemCount() != 2 {
		t.Errorf("ItemCount = %d, want 2", lv.ItemCount())
	}
	if lv.Labels()[1] != "C" {
		t.Errorf("Labels[1] = %q, want %q", lv.Labels()[1], "C")
	}
}

func TestListView_RemoveItem_InvalidIdx(t *testing.T) {
	lv := NewListView([]string{"A"})
	lv.RemoveItem(-1)
	lv.RemoveItem(5)
	if lv.ItemCount() != 1 {
		t.Errorf("ItemCount = %d, want 1", lv.ItemCount())
	}
}

func TestListView_Labels(t *testing.T) {
	lv := NewListView([]string{"X", "Y", "Z"})
	labels := lv.Labels()
	if len(labels) != 3 {
		t.Fatalf("len(Labels) = %d, want 3", len(labels))
	}
	if labels[0] != "X" || labels[1] != "Y" || labels[2] != "Z" {
		t.Errorf("Labels = %v, want [X Y Z]", labels)
	}
}

func TestListView_Cursor(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	if lv.Cursor() != 0 {
		t.Errorf("Cursor = %d, want 0", lv.Cursor())
	}
	lv.SetCursor(2)
	if lv.Cursor() != 2 {
		t.Errorf("Cursor = %d, want 2", lv.Cursor())
	}
}

func TestListView_SetCursor_Wrap(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	lv.SetCursor(5) // overflow wraps
	if lv.Cursor() != 0 {
		t.Errorf("Cursor = %d, want 0 (wrap)", lv.Cursor())
	}
	lv.SetCursor(-1) // negative wraps to last
	if lv.Cursor() != 2 {
		t.Errorf("Cursor = %d, want 2 (wrap)", lv.Cursor())
	}
}

func TestListView_MoveDown(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	lv.MoveDown()
	if lv.Cursor() != 1 {
		t.Errorf("Cursor = %d, want 1", lv.Cursor())
	}
	lv.MoveDown()
	if lv.Cursor() != 2 {
		t.Errorf("Cursor = %d, want 2", lv.Cursor())
	}
	lv.MoveDown() // wraps to 0
	if lv.Cursor() != 0 {
		t.Errorf("Cursor = %d, want 0 (wrap)", lv.Cursor())
	}
}

func TestListView_MoveUp(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	lv.MoveUp() // wraps to last
	if lv.Cursor() != 2 {
		t.Errorf("Cursor = %d, want 2 (wrap)", lv.Cursor())
	}
	lv.MoveUp()
	if lv.Cursor() != 1 {
		t.Errorf("Cursor = %d, want 1", lv.Cursor())
	}
}

func TestListView_MoveTop(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	lv.SetCursor(2)
	lv.MoveTop()
	if lv.Cursor() != 0 {
		t.Errorf("Cursor = %d, want 0", lv.Cursor())
	}
}

func TestListView_MoveBottom(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	lv.MoveBottom()
	if lv.Cursor() != 2 {
		t.Errorf("Cursor = %d, want 2", lv.Cursor())
	}
}

func TestListView_SelectedItem(t *testing.T) {
	lv := NewListView([]string{"A", "B"})
	lv.SetCursor(1)
	item, ok := lv.SelectedItem()
	if !ok {
		t.Fatal("SelectedItem returned false")
	}
	if item.Label != "B" {
		t.Errorf("Label = %q, want %q", item.Label, "B")
	}
}

func TestListView_SelectedItem_Empty(t *testing.T) {
	lv := NewListView(nil)
	_, ok := lv.SelectedItem()
	if ok {
		t.Error("SelectedItem should return false for empty list")
	}
}

func TestListView_Select(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	lv.SetCursor(1)
	called := false
	lv.OnSelect = func(item ListItem, idx int) {
		called = true
		if item.Label != "B" {
			t.Errorf("Label = %q, want %q", item.Label, "B")
		}
		if idx != 1 {
			t.Errorf("idx = %d, want 1", idx)
		}
	}
	lv.Select()
	if !called {
		t.Error("OnSelect was not called")
	}
}

func TestListView_Select_Empty(t *testing.T) {
	lv := NewListView(nil)
	called := false
	lv.OnSelect = func(item ListItem, idx int) {
		called = true
	}
	lv.Select()
	if called {
		t.Error("OnSelect should not be called for empty list")
	}
}

func TestListView_OnChange(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	moved := -1
	lv.OnChange = func(cursor int) {
		moved = cursor
	}
	lv.MoveDown()
	if moved != 1 {
		t.Errorf("OnChange cursor = %d, want 1", moved)
	}
}

func TestListView_HandleKey_Down(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	key := &term.KeyEvent{Key: term.KeyDown}
	if !lv.HandleKey(key) {
		t.Error("HandleKey returned false for KeyDown")
	}
	if lv.Cursor() != 1 {
		t.Errorf("Cursor = %d, want 1", lv.Cursor())
	}
}

func TestListView_HandleKey_Up(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	lv.SetCursor(1)
	key := &term.KeyEvent{Key: term.KeyUp}
	if !lv.HandleKey(key) {
		t.Error("HandleKey returned false for KeyUp")
	}
	if lv.Cursor() != 0 {
		t.Errorf("Cursor = %d, want 0", lv.Cursor())
	}
}

func TestListView_HandleKey_JK(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	// j = down
	key := &term.KeyEvent{Key: term.KeyUnknown, Rune: 'j'}
	if !lv.HandleKey(key) {
		t.Error("HandleKey returned false for 'j'")
	}
	if lv.Cursor() != 1 {
		t.Errorf("Cursor = %d, want 1", lv.Cursor())
	}
	// k = up
	key = &term.KeyEvent{Key: term.KeyUnknown, Rune: 'k'}
	if !lv.HandleKey(key) {
		t.Error("HandleKey returned false for 'k'")
	}
	if lv.Cursor() != 0 {
		t.Errorf("Cursor = %d, want 0", lv.Cursor())
	}
}

func TestListView_HandleKey_Home(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	lv.SetCursor(2)
	key := &term.KeyEvent{Key: term.KeyHome}
	if !lv.HandleKey(key) {
		t.Error("HandleKey returned false for KeyHome")
	}
	if lv.Cursor() != 0 {
		t.Errorf("Cursor = %d, want 0", lv.Cursor())
	}
}

func TestListView_HandleKey_End(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	key := &term.KeyEvent{Key: term.KeyEnd}
	if !lv.HandleKey(key) {
		t.Error("HandleKey returned false for KeyEnd")
	}
	if lv.Cursor() != 2 {
		t.Errorf("Cursor = %d, want 2", lv.Cursor())
	}
}

func TestListView_HandleKey_Enter(t *testing.T) {
	lv := NewListView([]string{"A", "B"})
	lv.SetCursor(1)
	called := false
	lv.OnSelect = func(item ListItem, idx int) {
		called = true
	}
	key := &term.KeyEvent{Key: term.KeyEnter}
	if !lv.HandleKey(key) {
		t.Error("HandleKey returned false for Enter")
	}
	if !called {
		t.Error("OnSelect not called on Enter")
	}
}

func TestListView_HandleKey_Nil(t *testing.T) {
	lv := NewListView([]string{"A"})
	if lv.HandleKey(nil) {
		t.Error("HandleKey(nil) should return false")
	}
}

func TestListView_HandleKey_Unhandled(t *testing.T) {
	lv := NewListView([]string{"A"})
	key := &term.KeyEvent{Key: term.KeyUnknown, Rune: 'x'}
	if lv.HandleKey(key) {
		t.Error("HandleKey should return false for unhandled key")
	}
}

func TestListView_HandleKey_CustomHandler(t *testing.T) {
	lv := NewListView([]string{"A", "B"})
	customCalled := false
	lv.OnKey = func(key *term.KeyEvent) bool {
		if key.Rune == 'x' {
			customCalled = true
			return true
		}
		return false
	}
	key := &term.KeyEvent{Key: term.KeyUnknown, Rune: 'x'}
	if !lv.HandleKey(key) {
		t.Error("HandleKey should return true for custom handler")
	}
	if !customCalled {
		t.Error("Custom handler not called")
	}
}

func TestListView_DisabledItemSkip(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C", "D"})
	lv.SetItems([]ListItem{
		{Label: "A"},
		{Label: "B", Disabled: true},
		{Label: "C"},
		{Label: "D"},
	})
	lv.SetCursor(0)
	lv.MoveDown()
	if lv.Cursor() != 2 {
		t.Errorf("Cursor = %d, want 2 (skip disabled B)", lv.Cursor())
	}
}

func TestListView_AllDisabled(t *testing.T) {
	lv := NewListView([]string{"A", "B"})
	lv.SetItems([]ListItem{
		{Label: "A", Disabled: true},
		{Label: "B", Disabled: true},
	})
	lv.SetCursor(0) // should stay at 0, all disabled
	if lv.Cursor() != 0 {
		t.Errorf("Cursor = %d, want 0 (all disabled)", lv.Cursor())
	}
}

func TestListView_PageDown(t *testing.T) {
	lv := NewListView([]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"})
	lv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	lv.PageDown()
	if lv.Cursor() < 2 {
		t.Errorf("Cursor = %d, should be >= 2 after PageDown with h=3", lv.Cursor())
	}
}

func TestListView_PageUp(t *testing.T) {
	lv := NewListView([]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"})
	lv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	lv.SetCursor(9)
	lv.PageUp()
	if lv.Cursor() > 7 {
		t.Errorf("Cursor = %d, should be <= 7 after PageUp with h=3", lv.Cursor())
	}
}

func TestListView_ScrollOffset(t *testing.T) {
	lv := NewListView([]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"})
	lv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	if lv.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset = %d, want 0", lv.ScrollOffset())
	}
	lv.SetCursor(5)
	if lv.ScrollOffset() < 3 {
		t.Errorf("ScrollOffset = %d, should be >= 3 for cursor=5 h=3", lv.ScrollOffset())
	}
}

func TestListView_Measure(t *testing.T) {
	lv := NewListView([]string{"Short", "Very Long Label Here", "Mid"})
	s := lv.Measure(Constraints{})
	if s.H != 3 {
		t.Errorf("H = %d, want 3", s.H)
	}
	if s.W < 10 {
		t.Errorf("W = %d, should be >= 10", s.W)
	}
}

func TestListView_Measure_Empty(t *testing.T) {
	lv := NewListView(nil)
	s := lv.Measure(Constraints{})
	if s.H != 1 {
		t.Errorf("H = %d, want 1 for empty list", s.H)
	}
}

func TestListView_Measure_Bounded(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	s := lv.Measure(Constraints{MaxWidth: 2, MaxHeight: 1})
	if s.W > 2 {
		t.Errorf("W = %d, should be <= 2", s.W)
	}
	if s.H != 1 {
		t.Errorf("H = %d, want 1", s.H)
	}
}

func TestListView_Paint(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	lv.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	buf := buffer.NewBuffer(10, 3)
	lv.Paint(buf)
	// Verify content was written
	cell := buf.GetCell(0, 0) // cursor indicator at col 0
	if cell.Rune != '>' {
		t.Errorf("Cell(0,0).Rune = %q, want '>'", cell.Rune)
	}
	cell = buf.GetCell(2, 0) // first char of label "A" (col 0=indicator, 1=space)
	if cell.Rune != 'A' {
		t.Errorf("Cell(2,0).Rune = %q, want 'A'", cell.Rune)
	}
}

func TestListView_Paint_Empty(t *testing.T) {
	lv := NewListView(nil)
	lv.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	buf := buffer.NewBuffer(10, 3)
	// Should not panic
	lv.Paint(buf)
}

func TestListView_Paint_ZeroBounds(t *testing.T) {
	lv := NewListView([]string{"A"})
	buf := buffer.NewBuffer(10, 3)
	// Should not panic with zero bounds
	lv.Paint(buf)
}

func TestListView_Paint_WithIcon(t *testing.T) {
	lv := NewListView(nil)
	lv.SetItems([]ListItem{
		{Label: "Folder", Icon: '*'},
	})
	lv.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 1})
	buf := buffer.NewBuffer(15, 1)
	lv.Paint(buf)
	cell := buf.GetCell(1, 0) // icon at col 1 (col 0=cursor indicator)
	if cell.Rune != '*' {
		t.Errorf("Cell(1,0).Rune = %q, want '*'", cell.Rune)
	}
}

func TestListView_Paint_Scrollbar(t *testing.T) {
	lv := NewListView([]string{"0", "1", "2", "3", "4", "5", "6", "7"})
	lv.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	lv.SetCursor(5)
	buf := buffer.NewBuffer(10, 3)
	lv.Paint(buf)
	// Scrollbar should be at last column
	scrollCol := 9
	cell := buf.GetCell(scrollCol, 0)
	if cell.Rune != '|' && cell.Rune != '*' {
		t.Errorf("Scrollbar char at col=%d = %q, want '|' or '*'", scrollCol, cell.Rune)
	}
}

func TestListView_Filter(t *testing.T) {
	lv := NewListView([]string{"Apple", "Banana", "Apricot", "Cherry"})
	indices := lv.Filter("ap")
	if len(indices) != 2 {
		t.Fatalf("len(Filter) = %d, want 2", len(indices))
	}
	if indices[0] != 0 || indices[1] != 2 {
		t.Errorf("Filter indices = %v, want [0 2]", indices)
	}
}

func TestListView_Filter_Empty(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	indices := lv.Filter("")
	if len(indices) != 3 {
		t.Errorf("len(Filter('')) = %d, want 3", len(indices))
	}
}

func TestListView_Filter_NoMatch(t *testing.T) {
	lv := NewListView([]string{"A", "B"})
	indices := lv.Filter("xyz")
	if len(indices) != 0 {
		t.Errorf("len(Filter('xyz')) = %d, want 0", len(indices))
	}
}

func TestListView_SetFilter(t *testing.T) {
	lv := NewListView([]string{"Apple", "Banana", "Apricot"})
	count := lv.SetFilter("ap")
	if count != 2 {
		t.Errorf("SetFilter count = %d, want 2", count)
	}
	if lv.Cursor() != 0 {
		t.Errorf("Cursor = %d, want 0 (first match)", lv.Cursor())
	}
}

func TestListView_Style(t *testing.T) {
	lv := NewListView([]string{"A"})
	style := DefaultListViewStyle()
	lv.SetStyle(style)
	got := lv.Style()
	if got.Selected.Flags != buffer.Reverse {
		t.Errorf("Selected.Flags = %v, want Reverse", got.Selected.Flags)
	}
}

func TestListView_Concurrent(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C", "D", "E"})
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				lv.MoveDown()
				lv.Items()
				lv.Cursor()
				_, _ = lv.SelectedItem()
			}
		}()
	}
	wg.Wait()
}

func TestListView_Concurrent_Paint(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C", "D", "E"})
	lv.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				lv.MoveDown()
				buf := buffer.NewBuffer(10, 3)
				lv.Paint(buf)
			}
		}(i)
	}
	wg.Wait()
}

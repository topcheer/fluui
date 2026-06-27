package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── MenuItem tests ─────────────────────────────────────────

func TestMenuItem_New(t *testing.T) {
	mi := NewMenuItem("cut", "Cut")
	if mi.ID != "cut" {
		t.Errorf("ID: got %q, want %q", mi.ID, "cut")
	}
	if mi.Label != "Cut" {
		t.Errorf("Label: got %q, want %q", mi.Label, "Cut")
	}
	if !mi.Enabled {
		t.Error("should be enabled by default")
	}
	if mi.Separator {
		t.Error("should not be separator")
	}
	if mi.HasSubmenu() {
		t.Error("should not have submenu")
	}
}

func TestMenuItem_NewSeparator(t *testing.T) {
	mi := NewSeparator()
	if !mi.Separator {
		t.Error("should be separator")
	}
}

func TestMenuItem_Chaining(t *testing.T) {
	called := false
	mi := NewMenuItem("save", "Save").
		SetShortcut("Ctrl+S").
		SetIcon("💾").
		SetEnabled(true).
		SetAction(func() { called = true })

	if mi.Shortcut != "Ctrl+S" {
		t.Errorf("Shortcut: got %q", mi.Shortcut)
	}
	if mi.Icon != "💾" {
		t.Errorf("Icon: got %q", mi.Icon)
	}
	mi.Action()
	if !called {
		t.Error("Action not called")
	}
}

func TestMenuItem_SetEnabled(t *testing.T) {
	mi := NewMenuItem("test", "Test").SetEnabled(false)
	if mi.Enabled {
		t.Error("should be disabled")
	}
	mi.SetEnabled(true)
	if !mi.Enabled {
		t.Error("should be enabled")
	}
}

func TestMenuItem_SetSubmenu(t *testing.T) {
	sub := NewContextMenu()
	mi := NewMenuItem("submenu", "Submenu").SetSubmenu(sub)
	if !mi.HasSubmenu() {
		t.Error("should have submenu")
	}
	if mi.Submenu != sub {
		t.Error("submenu pointer mismatch")
	}
}

// ─── ContextMenu construction ───────────────────────────────

func TestContextMenu_New(t *testing.T) {
	cm := NewContextMenu()
	if cm.ID() == "" {
		t.Error("ID should not be empty")
	}
	if cm.ItemCount() != 0 {
		t.Errorf("ItemCount: got %d, want 0", cm.ItemCount())
	}
	if cm.Visible() {
		t.Error("should not be visible")
	}
	if cm.Cursor() != 0 {
		t.Errorf("Cursor: got %d, want 0", cm.Cursor())
	}
}

func TestContextMenu_ID(t *testing.T) {
	cm := NewContextMenu()
	if cm.ID() == "" {
		t.Error("ID should not be empty")
	}
	cm2 := NewContextMenu()
	if cm.ID() == cm2.ID() {
		t.Error("IDs should be unique")
	}
}

// ─── AddItem / Remove / Clear ───────────────────────────────

func TestContextMenu_AddItem(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddLabel("b", "Beta")
	if cm.ItemCount() != 2 {
		t.Errorf("ItemCount: got %d, want 2", cm.ItemCount())
	}
}

func TestContextMenu_AddSeparator(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddSeparator()
	cm.AddLabel("b", "Beta")
	if cm.ItemCount() != 3 {
		t.Errorf("ItemCount: got %d, want 3", cm.ItemCount())
	}
}

func TestContextMenu_Remove(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddLabel("b", "Beta")
	cm.AddLabel("c", "Gamma")

	if !cm.Remove("b") {
		t.Error("Remove should find 'b'")
	}
	if cm.ItemCount() != 2 {
		t.Errorf("ItemCount after remove: got %d, want 2", cm.ItemCount())
	}
	if cm.Find("b") != nil {
		t.Error("should not find removed item")
	}
	if cm.Find("c") == nil {
		t.Error("should still find 'c'")
	}
}

func TestContextMenu_RemoveNotFound(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	if cm.Remove("nonexistent") {
		t.Error("should return false for missing ID")
	}
}

func TestContextMenu_Clear(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddLabel("b", "Beta")
	cm.Clear()
	if cm.ItemCount() != 0 {
		t.Errorf("ItemCount: got %d, want 0", cm.ItemCount())
	}
	if cm.Cursor() != 0 {
		t.Errorf("Cursor: got %d, want 0", cm.Cursor())
	}
}

func TestContextMenu_ItemsCopy(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	items := cm.Items()
	items[0] = NewMenuItem("modified", "Modified")
	// Original should not change because Items returns a copy of the slice
	if cm.Items()[0].ID != "a" {
		t.Error("Items() slice modification should not affect original")
	}
}

func TestContextMenu_ItemAt(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddSeparator()
	cm.AddLabel("b", "Beta")

	if cm.ItemAt(0).Label != "Alpha" {
		t.Error("ItemAt(0) should be Alpha")
	}
	if cm.ItemAt(1) == nil || !cm.ItemAt(1).Separator {
		t.Error("ItemAt(1) should be separator")
	}
	if cm.ItemAt(2).Label != "Beta" {
		t.Error("ItemAt(2) should be Beta")
	}
	if cm.ItemAt(3) != nil {
		t.Error("ItemAt(3) should be nil")
	}
}

func TestContextMenu_Find(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddLabel("b", "Beta")
	if cm.Find("a").Label != "Alpha" {
		t.Error("Find should return Alpha")
	}
	if cm.Find("nonexistent") != nil {
		t.Error("Find should return nil for missing")
	}
}

// ─── Cursor navigation ──────────────────────────────────────

func TestContextMenu_MoveDown(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddLabel("b", "Beta")
	cm.AddLabel("c", "Gamma")
	cm.Show(0, 0)

	cm.MoveDown()
	if cm.Cursor() != 1 {
		t.Errorf("Cursor: got %d, want 1", cm.Cursor())
	}
	cm.MoveDown()
	if cm.Cursor() != 2 {
		t.Errorf("Cursor: got %d, want 2", cm.Cursor())
	}
	cm.MoveDown() // wrap
	if cm.Cursor() != 0 {
		t.Errorf("Cursor after wrap: got %d, want 0", cm.Cursor())
	}
}

func TestContextMenu_MoveUp(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddLabel("b", "Beta")
	cm.Show(0, 0)

	cm.MoveUp() // wrap to last
	if cm.Cursor() != 1 {
		t.Errorf("Cursor after wrap up: got %d, want 1", cm.Cursor())
	}
	cm.MoveUp()
	if cm.Cursor() != 0 {
		t.Errorf("Cursor: got %d, want 0", cm.Cursor())
	}
}

func TestContextMenu_MoveSkipsSeparator(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha") // 0
	cm.AddSeparator()         // 1
	cm.AddLabel("b", "Beta")  // 2
	cm.Show(0, 0)

	cm.MoveDown() // should skip separator, go to 2
	if cm.Cursor() != 2 {
		t.Errorf("Cursor: got %d, want 2 (skip separator)", cm.Cursor())
	}
}

func TestContextMenu_MoveSkipsDisabled(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")                        // 0
	cm.AddItem(NewMenuItem("b", "Beta").SetEnabled(false)) // 1 disabled
	cm.AddLabel("c", "Gamma")                        // 2
	cm.Show(0, 0)

	cm.MoveDown() // should skip disabled, go to 2
	if cm.Cursor() != 2 {
		t.Errorf("Cursor: got %d, want 2 (skip disabled)", cm.Cursor())
	}
}

func TestContextMenu_SetCursor(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddLabel("b", "Beta")
	cm.AddLabel("c", "Gamma")
	cm.Show(0, 0)

	cm.SetCursor(2)
	if cm.Cursor() != 2 {
		t.Errorf("Cursor: got %d, want 2", cm.Cursor())
	}
}

func TestContextMenu_SetCursorSkipsSeparator(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddSeparator()
	cm.AddLabel("b", "Beta")
	cm.Show(0, 0)

	cm.SetCursor(1) // separator
	if cm.Cursor() == 1 {
		t.Error("Cursor should not land on separator")
	}
}

func TestContextMenu_ShowResetsCursor(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddLabel("b", "Beta")
	cm.SetCursor(1)
	cm.Show(5, 10)
	if cm.Cursor() != 0 {
		t.Errorf("Cursor after Show: got %d, want 0", cm.Cursor())
	}
}

// ─── Show / Hide ────────────────────────────────────────────

func TestContextMenu_Show(t *testing.T) {
	cm := NewContextMenu()
	cm.Show(5, 10)
	if !cm.Visible() {
		t.Error("should be visible")
	}
	x, y := cm.Position()
	if x != 5 || y != 10 {
		t.Errorf("Position: got (%d,%d), want (5,10)", x, y)
	}
}

func TestContextMenu_Hide(t *testing.T) {
	cm := NewContextMenu()
	cm.Show(0, 0)
	cm.Hide()
	if cm.Visible() {
		t.Error("should not be visible after Hide")
	}
}

func TestContextMenu_OnClose(t *testing.T) {
	called := false
	cm := NewContextMenu()
	cm.OnClose = func() { called = true }
	cm.Show(0, 0)
	cm.Hide()
	if !called {
		t.Error("OnClose should be called on Hide")
	}
}

func TestContextMenu_SetPosition(t *testing.T) {
	cm := NewContextMenu()
	cm.SetPosition(3, 7)
	x, y := cm.Position()
	if x != 3 || y != 7 {
		t.Errorf("Position: got (%d,%d), want (3,7)", x, y)
	}
}

// ─── Activate ───────────────────────────────────────────────

func TestContextMenu_Activate(t *testing.T) {
	called := false
	cm := NewContextMenu()
	cm.AddLabel("save", "Save").SetAction(func() { called = true })
	cm.Show(0, 0)

	item := cm.Activate()
	if item == nil {
		t.Fatal("Activate should return item")
	}
	if !called {
		t.Error("action should be called")
	}
}

func TestContextMenu_ActivateDisabled(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("test", "Test").SetEnabled(false))
	cm.Show(0, 0)

	item := cm.Activate()
	if item != nil {
		t.Error("Activate on disabled should return nil")
	}
}

func TestContextMenu_ActivateSeparator(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddSeparator()
	cm.AddLabel("b", "Beta")
	cm.Show(0, 0)

	// Force cursor onto the separator by manipulating internal state
	cm.mu.Lock()
	cm.cursor = 1 // separator index
	cm.mu.Unlock()

	item := cm.Activate()
	if item != nil {
		t.Error("Activate on separator should return nil")
	}
}

func TestContextMenu_OnSelect(t *testing.T) {
	var selected *MenuItem
	cm := NewContextMenu()
	cm.OnSelect = func(it *MenuItem) { selected = it }
	cm.AddLabel("a", "Alpha")
	cm.Show(0, 0)
	cm.Activate()
	if selected == nil || selected.Label != "Alpha" {
		t.Error("OnSelect should receive Alpha")
	}
}

func TestContextMenu_ActivateSubmenu(t *testing.T) {
	sub := NewContextMenu()
	sub.AddLabel("sub1", "Sub 1")
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddItem(NewMenuItem("more", "More").SetSubmenu(sub))
	cm.Show(0, 0)
	cm.SetCursor(1)

	item := cm.Activate()
	if item == nil || !item.HasSubmenu() {
		t.Fatal("should return submenu item")
	}
	if !sub.Visible() {
		t.Error("submenu should be visible after activate")
	}
}

// ─── Style ──────────────────────────────────────────────────

func TestContextMenu_SetStyle(t *testing.T) {
	cm := NewContextMenu()
	st := DefaultContextMenuStyle()
	st.Selected = buffer.Style{Flags: buffer.Bold | buffer.Reverse}
	cm.SetStyle(st)
	if cm.Style().Selected.Flags != buffer.Bold|buffer.Reverse {
		t.Error("style not set")
	}
}

func TestDefaultContextMenuStyle(t *testing.T) {
	st := DefaultContextMenuStyle()
	if st.Selected.Flags&buffer.Reverse == 0 {
		t.Error("Selected should have Reverse flag")
	}
	if st.Disabled.Flags&buffer.Dim == 0 {
		t.Error("Disabled should have Dim flag")
	}
}

// ─── Measure ────────────────────────────────────────────────

func TestContextMenu_MeasureEmpty(t *testing.T) {
	cm := NewContextMenu()
	sz := cm.Measure(Unbounded())
	if sz.H < 1 {
		t.Error("H should be at least 1")
	}
	if sz.W < 1 {
		t.Error("W should be at least 1")
	}
}

func TestContextMenu_MeasureWithItems(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddLabel("b", "BetaBetaBeta")
	cm.AddSeparator()
	sz := cm.Measure(Unbounded())
	if sz.H != 5 { // 3 items + 2 borders
		t.Errorf("H: got %d, want 5", sz.H)
	}
	if sz.W < 10 {
		t.Errorf("W should be at least 10, got %d", sz.W)
	}
}

func TestContextMenu_MeasureWithShortcut(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("save", "Save").SetShortcut("Ctrl+S")
	sz := cm.Measure(Unbounded())
	// "Save" (4) + "Ctrl+S" (6) + 2 gap + 4 padding = 16
	if sz.W < 12 {
		t.Errorf("W should include shortcut, got %d", sz.W)
	}
}

func TestContextMenu_MeasureClamped(t *testing.T) {
	cm := NewContextMenu()
	for i := 0; i < 20; i++ {
		cm.AddLabel("item", "ItemItemItemItemItem")
	}
	sz := cm.Measure(Bounded(15, 5))
	if sz.W > 15 {
		t.Errorf("W: got %d, should be clamped to 15", sz.W)
	}
	if sz.H > 5 {
		t.Errorf("H: got %d, should be clamped to 5", sz.H)
	}
}

// ─── Paint ──────────────────────────────────────────────────

func TestContextMenu_PaintHidden(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	buf := buffer.NewBuffer(30, 10)
	cm.Paint(buf) // should be a no-op (not visible)
	// Buffer should not have border characters for this menu
	if buf.GetCell(0, 0).Rune == '┌' {
		t.Error("hidden menu should not paint border characters")
	}
}

func TestContextMenu_PaintBasic(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddLabel("b", "Beta")
	cm.Show(5, 5)
	cm.Measure(Unbounded()) // compute width

	buf := buffer.NewBuffer(30, 10)
	cm.Paint(buf)

	// Top-left corner
	if buf.GetCell(5, 5).Rune != '┌' {
		t.Errorf("Expected ┌ at (5,5), got %q", string(buf.GetCell(5, 5).Rune))
	}
	// Bottom-left corner
	botY := 5 + 2 + 1 // y + items + 1
	if buf.GetCell(5, botY).Rune != '└' {
		t.Errorf("Expected └ at (5,%d), got %q", botY, string(buf.GetCell(5, botY).Rune))
	}
	// First item content
	if buf.GetCell(7, 6).Rune != 'A' {
		t.Errorf("Expected 'A' at (7,6), got %q", string(buf.GetCell(7, 6).Rune))
	}
}

func TestContextMenu_PaintSeparator(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddSeparator()
	cm.AddLabel("b", "Beta")
	cm.Show(0, 0)
	cm.Measure(Unbounded())

	buf := buffer.NewBuffer(30, 10)
	cm.Paint(buf)

	// Row 2 (y=2) should be a separator line of '─'
	for i := 1; i < cm.width-1; i++ {
		if buf.GetCell(i, 2).Rune != '─' {
			t.Errorf("Expected ─ at (%d,2), got %q", i, string(buf.GetCell(i, 2).Rune))
			break
		}
	}
}

func TestContextMenu_PaintWithShortcut(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("save", "Save").SetShortcut("^S")
	cm.Show(0, 0)
	cm.Measure(Unbounded())

	buf := buffer.NewBuffer(30, 5)
	cm.Paint(buf)

	// Shortcut should appear at the right side
	found := false
	for i := cm.width - 3; i >= 0; i-- {
		if buf.GetCell(i, 1).Rune == '^' {
			found = true
			break
		}
	}
	if !found {
		t.Error("Shortcut ^S not found in paint output")
	}
}

func TestContextMenu_PaintDisabled(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("dis", "Disabled").SetEnabled(false))
	cm.Show(0, 0)
	cm.Measure(Unbounded())

	buf := buffer.NewBuffer(30, 5)
	cm.Paint(buf)

	cell := buf.GetCell(2, 1)
	// Disabled should have Dim flag
	if cell.Flags&buffer.Dim == 0 {
		t.Error("Disabled item should have Dim flag")
	}
}

// ─── HandleKey ──────────────────────────────────────────────

func TestContextMenu_HandleKeyDown(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddLabel("b", "Beta")
	cm.Show(0, 0)

	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if !consumed {
		t.Error("Key should be consumed")
	}
	if cm.Cursor() != 1 {
		t.Errorf("Cursor: got %d, want 1", cm.Cursor())
	}
}

func TestContextMenu_HandleKeyUp(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddLabel("b", "Beta")
	cm.Show(0, 0)
	cm.SetCursor(1)

	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if !consumed {
		t.Error("Key should be consumed")
	}
	if cm.Cursor() != 0 {
		t.Errorf("Cursor: got %d, want 0", cm.Cursor())
	}
}

func TestContextMenu_HandleKeyEnter(t *testing.T) {
	called := false
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha").SetAction(func() { called = true })
	cm.Show(0, 0)

	cm.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !called {
		t.Error("Action should be called on Enter")
	}
}

func TestContextMenu_HandleKeyEscape(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.Show(0, 0)

	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Error("Esc should be consumed")
	}
	if cm.Visible() {
		t.Error("menu should be hidden after Esc")
	}
}

func TestContextMenu_HandleKeyHidden(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	// Not shown
	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if consumed {
		t.Error("should not consume keys when hidden")
	}
}

func TestContextMenu_HandleKeyUnknown(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.Show(0, 0)

	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if consumed {
		t.Error("Tab should not be consumed")
	}
}

// ─── Mouse / HitTest ────────────────────────────────────────

func TestContextMenu_HitTestInside(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha") // row 0 → screen y=1
	cm.AddLabel("b", "Beta")  // row 1 → screen y=2
	cm.Show(0, 0)
	cm.Measure(Unbounded())

	idx := cm.HitTest(2, 1) // first item
	if idx != 0 {
		t.Errorf("HitTest(2,1): got %d, want 0", idx)
	}
	idx = cm.HitTest(5, 2) // second item
	if idx != 1 {
		t.Errorf("HitTest(5,2): got %d, want 1", idx)
	}
}

func TestContextMenu_HitTestOutside(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.Show(5, 5)
	cm.Measure(Unbounded())

	if cm.HitTest(0, 0) != -1 {
		t.Error("should return -1 for outside click")
	}
}

func TestContextMenu_HitTestHidden(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	// Not shown
	if cm.HitTest(0, 0) != -1 {
		t.Error("hidden menu should return -1")
	}
}

func TestContextMenu_ClickAt(t *testing.T) {
	called := false
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddLabel("b", "Beta").SetAction(func() { called = true })
	cm.Show(0, 0)
	cm.Measure(Unbounded())

	ok := cm.ClickAt(2, 2) // click on second item
	if !ok {
		t.Error("ClickAt should return true inside menu")
	}
	if !called {
		t.Error("action should be called")
	}
	if cm.Cursor() != 1 {
		t.Errorf("Cursor: got %d, want 1", cm.Cursor())
	}
}

func TestContextMenu_ClickAtOutside(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.Show(0, 0)

	ok := cm.ClickAt(100, 100)
	if ok {
		t.Error("ClickAt outside should return false")
	}
}

func TestContextMenu_ClickAtSeparator(t *testing.T) {
	called := false
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha").SetAction(func() { called = true })
	cm.AddSeparator()
	cm.Show(0, 0)
	cm.Measure(Unbounded())

	ok := cm.ClickAt(2, 2) // click on separator
	if !ok {
		t.Error("ClickAt on separator should still return true")
	}
	if called {
		t.Error("action should NOT be called for separator click")
	}
}

func TestContextMenu_ClickAtDisabled(t *testing.T) {
	called := false
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("dis", "Disabled").SetAction(func() { called = true }).SetEnabled(false))
	cm.Show(0, 0)
	cm.Measure(Unbounded())

	ok := cm.ClickAt(2, 1) // click on disabled
	if !ok {
		t.Error("ClickAt on disabled should return true")
	}
	if called {
		t.Error("action should NOT be called for disabled click")
	}
}

// ─── Children ───────────────────────────────────────────────

func TestContextMenu_ChildrenNoSubmenu(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.Show(0, 0)
	if children := cm.Children(); len(children) != 0 {
		t.Errorf("Children: got %d, want 0", len(children))
	}
}

func TestContextMenu_ChildrenWithSubmenu(t *testing.T) {
	sub := NewContextMenu()
	sub.AddLabel("sub1", "Sub 1")
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("more", "More").SetSubmenu(sub))
	cm.Show(0, 0)
	cm.SetCursor(0)
	cm.Activate() // opens submenu

	children := cm.Children()
	if len(children) != 1 {
		t.Fatalf("Children: got %d, want 1", len(children))
	}
}

// ─── String ─────────────────────────────────────────────────

func TestContextMenu_String(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddSeparator()
	cm.AddLabel("b", "Beta").SetShortcut("Ctrl+B")
	s := cm.String()
	if !contains(s, "Alpha") {
		t.Errorf("String should contain Alpha: %q", s)
	}
	if !contains(s, "Ctrl+B") {
		t.Errorf("String should contain Ctrl+B: %q", s)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ─── Concurrency ────────────────────────────────────────────

func TestContextMenu_ConcurrentAccess(t *testing.T) {
	cm := NewContextMenu()
	for i := 0; i < 10; i++ {
		cm.AddLabel("item", "Item")
	}
	cm.Show(0, 0)

	var wg sync.WaitGroup
	wg.Add(3)

	// Writer: add and remove items
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			cm.AddLabel("tmp", "Temp")
			cm.Remove("tmp")
		}
	}()

	// Reader: navigate
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			cm.MoveDown()
			cm.MoveUp()
			_ = cm.Cursor()
			_ = cm.Items()
		}
	}()

	// Reader: measure + paint
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			cm.Measure(Unbounded())
		}
	}()

	wg.Wait()
}

func TestContextMenu_ConcurrentPaint(t *testing.T) {
	cm := NewContextMenu()
	cm.AddLabel("a", "Alpha")
	cm.AddLabel("b", "Beta")
	cm.Show(0, 0)
	cm.Measure(Unbounded())

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			buf := buffer.NewBuffer(30, 10)
			cm.Paint(buf)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			cm.MoveDown()
			cm.SetStyle(DefaultContextMenuStyle())
		}
	}()

	wg.Wait()
}

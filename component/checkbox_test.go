package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── Construction ──────────────────────────────────────────────

func TestNewCheckbox(t *testing.T) {
	cb := NewCheckbox([]string{"Apple", "Banana", "Cherry"})
	if cb == nil {
		t.Fatal("expected non-nil checkbox")
	}
	if cb.ItemCount() != 3 {
		t.Errorf("ItemCount = %d, want 3", cb.ItemCount())
	}
}

func TestNewCheckbox_Empty(t *testing.T) {
	cb := NewCheckbox(nil)
	if cb.ItemCount() != 0 {
		t.Errorf("ItemCount = %d, want 0", cb.ItemCount())
	}
}

func TestNewCheckbox_UniqueID(t *testing.T) {
	cb1 := NewCheckbox([]string{"A"})
	cb2 := NewCheckbox([]string{"B"})
	if cb1.ID() == cb2.ID() {
		t.Error("expected unique IDs")
	}
}

func TestCheckbox_ImplementsComponent(t *testing.T) {
	var _ Component = NewCheckbox([]string{"X"})
}

// ─── Items ─────────────────────────────────────────────────────

func TestCheckbox_Items(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B"})
	items := cb.Items()
	if len(items) != 2 {
		t.Fatalf("len = %d, want 2", len(items))
	}
	if items[0].Label != "A" || items[1].Label != "B" {
		t.Error("labels mismatch")
	}
}

func TestCheckbox_ItemsReturnsCopy(t *testing.T) {
	cb := NewCheckbox([]string{"A"})
	items := cb.Items()
	items[0].Checked = true
	if cb.IsChecked(0) {
		t.Error("modifying returned slice should not affect original")
	}
}

func TestCheckbox_SetItems(t *testing.T) {
	cb := NewCheckbox([]string{"A"})
	cb.SetItems([]CheckboxItem{
		{Label: "X", Checked: true},
		{Label: "Y"},
	})
	if cb.ItemCount() != 2 {
		t.Errorf("ItemCount = %d, want 2", cb.ItemCount())
	}
	if !cb.IsChecked(0) {
		t.Error("item 0 should be checked")
	}
}

func TestCheckbox_AddItem(t *testing.T) {
	cb := NewCheckbox([]string{"A"})
	cb.AddItem("B")
	if cb.ItemCount() != 2 {
		t.Errorf("ItemCount = %d, want 2", cb.ItemCount())
	}
}

func TestCheckbox_CheckedItems(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C"})
	cb.SetChecked(0, true)
	cb.SetChecked(2, true)
	checked := cb.CheckedItems()
	if len(checked) != 2 {
		t.Fatalf("checked count = %d, want 2", len(checked))
	}
	if checked[0].Label != "A" || checked[1].Label != "C" {
		t.Error("checked labels mismatch")
	}
}

func TestCheckbox_CheckedLabels(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B"})
	cb.SetChecked(1, true)
	labels := cb.CheckedLabels()
	if len(labels) != 1 || labels[0] != "B" {
		t.Errorf("labels = %v, want [B]", labels)
	}
}

// ─── Cursor navigation ─────────────────────────────────────────

func TestCheckbox_Cursor(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C"})
	if cb.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", cb.Cursor())
	}
}

func TestCheckbox_MoveDown(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C"})
	cb.MoveDown()
	if cb.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1", cb.Cursor())
	}
}

func TestCheckbox_MoveDownWrap(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B"})
	cb.MoveDown()
	cb.MoveDown()
	if cb.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 (wrap)", cb.Cursor())
	}
}

func TestCheckbox_MoveUp(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C"})
	cb.MoveDown()
	cb.MoveUp()
	if cb.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", cb.Cursor())
	}
}

func TestCheckbox_MoveUpWrap(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B"})
	cb.MoveUp()
	if cb.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1 (wrap)", cb.Cursor())
	}
}

func TestCheckbox_SetCursor(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C"})
	cb.SetCursor(2)
	if cb.Cursor() != 2 {
		t.Errorf("cursor = %d, want 2", cb.Cursor())
	}
}

func TestCheckbox_MoveDown_SkipDisabled(t *testing.T) {
	items := []CheckboxItem{
		{Label: "A"},
		{Label: "B", Disabled: true},
		{Label: "C"},
	}
	cb := NewCheckbox(nil)
	cb.SetItems(items)
	cb.MoveDown()
	if cb.Cursor() != 2 {
		t.Errorf("cursor = %d, want 2 (skip disabled)", cb.Cursor())
	}
}

// ─── Toggle ────────────────────────────────────────────────────

func TestCheckbox_Toggle(t *testing.T) {
	cb := NewCheckbox([]string{"A"})
	cb.Toggle()
	if !cb.IsChecked(0) {
		t.Error("expected checked after toggle")
	}
	cb.Toggle()
	if cb.IsChecked(0) {
		t.Error("expected unchecked after second toggle")
	}
}

func TestCheckbox_ToggleDisabled(t *testing.T) {
	items := []CheckboxItem{{Label: "A", Disabled: true}}
	cb := NewCheckbox(nil)
	cb.SetItems(items)
	cb.Toggle()
	if cb.IsChecked(0) {
		t.Error("disabled item should not be toggled")
	}
}

func TestCheckbox_SetChecked(t *testing.T) {
	cb := NewCheckbox([]string{"A"})
	cb.SetChecked(0, true)
	if !cb.IsChecked(0) {
		t.Error("expected checked")
	}
}

func TestCheckbox_SetChecked_Disabled(t *testing.T) {
	items := []CheckboxItem{{Label: "A", Disabled: true}}
	cb := NewCheckbox(nil)
	cb.SetItems(items)
	cb.SetChecked(0, true)
	if cb.IsChecked(0) {
		t.Error("disabled item should not be set")
	}
}

func TestCheckbox_CheckAll(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C"})
	cb.CheckAll()
	if len(cb.CheckedItems()) != 3 {
		t.Errorf("checked = %d, want 3", len(cb.CheckedItems()))
	}
}

func TestCheckbox_UncheckAll(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B"})
	cb.CheckAll()
	cb.UncheckAll()
	if len(cb.CheckedItems()) != 0 {
		t.Errorf("checked = %d, want 0", len(cb.CheckedItems()))
	}
}

func TestCheckbox_CheckAll_SkipDisabled(t *testing.T) {
	items := []CheckboxItem{
		{Label: "A"},
		{Label: "B", Disabled: true},
	}
	cb := NewCheckbox(nil)
	cb.SetItems(items)
	cb.CheckAll()
	if cb.IsChecked(0) != true {
		t.Error("A should be checked")
	}
	if cb.IsChecked(1) != false {
		t.Error("B (disabled) should remain unchecked")
	}
}

func TestCheckbox_OnChange(t *testing.T) {
	cb := NewCheckbox([]string{"A"})
	changed := false
	cb.OnChange = func(items []CheckboxItem) {
		changed = true
	}
	cb.Toggle()
	if !changed {
		t.Error("OnChange should have fired")
	}
}

// ─── Keyboard ──────────────────────────────────────────────────

func TestCheckbox_HandleKey_UpDown(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C"})
	cb.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if cb.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1", cb.Cursor())
	}
	cb.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if cb.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", cb.Cursor())
	}
}

func TestCheckbox_HandleKey_Space(t *testing.T) {
	cb := NewCheckbox([]string{"A"})
	consumed := cb.HandleKey(&term.KeyEvent{Key: term.KeySpace})
	if !consumed {
		t.Error("Space should be consumed")
	}
	if !cb.IsChecked(0) {
		t.Error("item should be toggled")
	}
}

func TestCheckbox_HandleKey_JK(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B"})
	cb.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: 'j'})
	if cb.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1 after j", cb.Cursor())
	}
	cb.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: 'k'})
	if cb.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 after k", cb.Cursor())
	}
}

func TestCheckbox_HandleKey_CtrlA(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B"})
	consumed := cb.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: 'a', Modifiers: term.ModCtrl})
	if !consumed {
		t.Error("Ctrl+A should be consumed")
	}
	if len(cb.CheckedItems()) != 2 {
		t.Errorf("checked = %d, want 2", len(cb.CheckedItems()))
	}
}

func TestCheckbox_HandleKey_CtrlD(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B"})
	cb.CheckAll()
	cb.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: 'd', Modifiers: term.ModCtrl})
	if len(cb.CheckedItems()) != 0 {
		t.Errorf("checked = %d, want 0", len(cb.CheckedItems()))
	}
}

func TestCheckbox_HandleKey_Nil(t *testing.T) {
	cb := NewCheckbox([]string{"A"})
	if cb.HandleKey(nil) {
		t.Error("nil key should not be consumed")
	}
}

func TestCheckbox_HandleKey_Unhandled(t *testing.T) {
	cb := NewCheckbox([]string{"A"})
	if cb.HandleKey(&term.KeyEvent{Key: term.KeyEscape}) {
		t.Error("Escape should not be consumed")
	}
}

// ─── Style ─────────────────────────────────────────────────────

func TestCheckbox_SetStyle(t *testing.T) {
	cb := NewCheckbox([]string{"A"})
	style := DefaultCheckboxStyle()
	style.Normal = buffer.Style{Fg: buffer.Red}
	cb.SetStyle(style)
	if cb.Style().Normal.Fg != buffer.Red {
		t.Error("style not set")
	}
}

// ─── Measure ───────────────────────────────────────────────────

func TestCheckbox_Measure(t *testing.T) {
	cb := NewCheckbox([]string{"Short", "VeryLongLabelHere"})
	size := cb.Measure(Constraints{})
	if size.H != 2 {
		t.Errorf("H = %d, want 2", size.H)
	}
	if size.W < 10 {
		t.Errorf("W = %d, should be >= 10", size.W)
	}
}

func TestCheckbox_Measure_Empty(t *testing.T) {
	cb := NewCheckbox(nil)
	size := cb.Measure(Constraints{})
	if size.H != 1 {
		t.Errorf("H = %d, want 1", size.H)
	}
}

func TestCheckbox_Measure_Clamped(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C", "D", "E"})
	size := cb.Measure(Constraints{MaxWidth: 5, MaxHeight: 3})
	if size.H > 3 {
		t.Errorf("H = %d, should be clamped to 3", size.H)
	}
}

// ─── Paint ─────────────────────────────────────────────────────

func TestCheckbox_Paint_NoPanic(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B"})
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	cb.Paint(buf)
}

func TestCheckbox_Paint_ZeroBounds(t *testing.T) {
	cb := NewCheckbox([]string{"A"})
	buf := buffer.NewBuffer(10, 5)
	cb.Paint(buf) // should not panic
}

func TestCheckbox_Paint_ShowChecked(t *testing.T) {
	cb := NewCheckbox([]string{"A"})
	cb.SetChecked(0, true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	cb.Paint(buf)
	cell := buf.GetCell(1, 0)
	if cell.Rune != 'x' {
		t.Errorf("cell[1] rune = %q, want 'x' (checked)", cell.Rune)
	}
}

// ─── Misc ──────────────────────────────────────────────────────

func TestCheckbox_String(t *testing.T) {
	cb := NewCheckbox([]string{"A"})
	if cb.String() != "Checkbox" {
		t.Errorf("String = %q, want 'Checkbox'", cb.String())
	}
}

func TestCheckbox_Children(t *testing.T) {
	cb := NewCheckbox([]string{"A"})
	if cb.Children() != nil {
		t.Error("Children should return nil")
	}
}

// ─── Concurrency ───────────────────────────────────────────────

func TestCheckbox_ConcurrentAccess(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C"})
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			cb.Items()
			cb.Cursor()
			cb.IsChecked(0)
		}()
		go func() {
			defer wg.Done()
			cb.MoveDown()
			cb.Toggle()
		}()
		go func() {
			defer wg.Done()
			cb.SetChecked(0, true)
			cb.CheckedItems()
		}()
	}
	wg.Wait()
}

func TestCheckbox_ConcurrentPaint(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C", "D", "E"})
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			buf := buffer.NewBuffer(20, 5)
			cb.Paint(buf)
		}()
		go func() {
			defer wg.Done()
			cb.MoveDown()
			cb.Toggle()
		}()
	}
	wg.Wait()
}

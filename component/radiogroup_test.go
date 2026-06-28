package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── Construction ──────────────────────────────────────────────

func TestNewRadioGroup(t *testing.T) {
	rg := NewRadioGroup([]string{"Red", "Green", "Blue"})
	if rg == nil {
		t.Fatal("expected non-nil RadioGroup")
	}
	if rg.ItemCount() != 3 {
		t.Errorf("ItemCount = %d, want 3", rg.ItemCount())
	}
}

func TestNewRadioGroup_Empty(t *testing.T) {
	rg := NewRadioGroup(nil)
	if rg.ItemCount() != 0 {
		t.Errorf("ItemCount = %d, want 0", rg.ItemCount())
	}
}

func TestNewRadioGroup_UniqueID(t *testing.T) {
	rg1 := NewRadioGroup([]string{"A"})
	rg2 := NewRadioGroup([]string{"B"})
	if rg1.ID() == rg2.ID() {
		t.Error("expected unique IDs")
	}
}

func TestRadioGroup_ImplementsComponent(t *testing.T) {
	var _ Component = NewRadioGroup([]string{"X"})
}

// ─── Labels ────────────────────────────────────────────────────

func TestRadioGroup_Labels(t *testing.T) {
	rg := NewRadioGroup([]string{"Red", "Blue"})
	labels := rg.Labels()
	if len(labels) != 2 {
		t.Fatalf("len = %d, want 2", len(labels))
	}
	if labels[0] != "Red" || labels[1] != "Blue" {
		t.Error("labels mismatch")
	}
}

func TestRadioGroup_LabelsReturnsCopy(t *testing.T) {
	rg := NewRadioGroup([]string{"A"})
	labels := rg.Labels()
	labels[0] = "Modified"
	if rg.Labels()[0] == "Modified" {
		t.Error("modifying returned slice should not affect original")
	}
}

// ─── Selection ─────────────────────────────────────────────────

func TestRadioGroup_NoSelectionByDefault(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B"})
	if rg.SelectedIndex() != -1 {
		t.Errorf("SelectedIndex = %d, want -1", rg.SelectedIndex())
	}
}

func TestRadioGroup_SelectedLabel_Empty(t *testing.T) {
	rg := NewRadioGroup([]string{"A"})
	if rg.SelectedLabel() != "" {
		t.Errorf("SelectedLabel = %q, want empty", rg.SelectedLabel())
	}
}

func TestRadioGroup_Select(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C"})
	rg.SetCursor(1)
	rg.Select()
	if rg.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", rg.SelectedIndex())
	}
	if rg.SelectedLabel() != "B" {
		t.Errorf("SelectedLabel = %q, want B", rg.SelectedLabel())
	}
}

func TestRadioGroup_SetSelected(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C"})
	rg.SetSelected(2)
	if rg.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex = %d, want 2", rg.SelectedIndex())
	}
}

func TestRadioGroup_MutuallyExclusive(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C"})
	rg.SetSelected(0)
	rg.SetSelected(2)
	if rg.SelectedIndex() != 2 {
		t.Errorf("SelectedIndex = %d, want 2", rg.SelectedIndex())
	}
}

func TestRadioGroup_SetSelected_Disabled(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B"})
	rg.SetDisabled(0, true)
	rg.SetSelected(0)
	if rg.SelectedIndex() == 0 {
		t.Error("should not select disabled option")
	}
}

func TestRadioGroup_SetDisabledClearsActive(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B"})
	rg.SetSelected(0)
	rg.SetDisabled(0, true)
	if rg.SelectedIndex() != -1 {
		t.Error("disabling active option should clear selection")
	}
}

func TestRadioGroup_IsDisabled(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B"})
	rg.SetDisabled(1, true)
	if !rg.IsDisabled(1) {
		t.Error("expected disabled")
	}
	if rg.IsDisabled(0) {
		t.Error("expected not disabled")
	}
}

// ─── Cursor ────────────────────────────────────────────────────

func TestRadioGroup_Cursor(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B"})
	if rg.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", rg.Cursor())
	}
}

func TestRadioGroup_MoveDown(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C"})
	rg.MoveDown()
	if rg.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1", rg.Cursor())
	}
}

func TestRadioGroup_MoveDownWrap(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B"})
	rg.MoveDown()
	rg.MoveDown()
	if rg.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 (wrap)", rg.Cursor())
	}
}

func TestRadioGroup_MoveUp(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C"})
	rg.MoveDown()
	rg.MoveUp()
	if rg.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", rg.Cursor())
	}
}

func TestRadioGroup_MoveUpWrap(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B"})
	rg.MoveUp()
	if rg.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1 (wrap)", rg.Cursor())
	}
}

func TestRadioGroup_MoveDown_SkipDisabled(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C"})
	rg.SetDisabled(1, true)
	rg.MoveDown()
	if rg.Cursor() != 2 {
		t.Errorf("cursor = %d, want 2 (skip disabled)", rg.Cursor())
	}
}

// ─── Keyboard ──────────────────────────────────────────────────

func TestRadioGroup_HandleKey_UpDown(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C"})
	rg.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if rg.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1", rg.Cursor())
	}
	rg.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if rg.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0", rg.Cursor())
	}
}

func TestRadioGroup_HandleKey_Enter(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B"})
	rg.SetCursor(1)
	consumed := rg.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !consumed {
		t.Error("Enter should be consumed")
	}
	if rg.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex = %d, want 1", rg.SelectedIndex())
	}
}

func TestRadioGroup_HandleKey_Space(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B"})
	consumed := rg.HandleKey(&term.KeyEvent{Key: term.KeySpace})
	if !consumed {
		t.Error("Space should be consumed")
	}
	if rg.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex = %d, want 0", rg.SelectedIndex())
	}
}

func TestRadioGroup_HandleKey_JK(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B"})
	rg.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: 'j'})
	if rg.Cursor() != 1 {
		t.Errorf("cursor = %d, want 1 after j", rg.Cursor())
	}
	rg.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: 'k'})
	if rg.Cursor() != 0 {
		t.Errorf("cursor = %d, want 0 after k", rg.Cursor())
	}
}

func TestRadioGroup_HandleKey_Nil(t *testing.T) {
	rg := NewRadioGroup([]string{"A"})
	if rg.HandleKey(nil) {
		t.Error("nil key should not be consumed")
	}
}

func TestRadioGroup_HandleKey_Unhandled(t *testing.T) {
	rg := NewRadioGroup([]string{"A"})
	if rg.HandleKey(&term.KeyEvent{Key: term.KeyEscape}) {
		t.Error("Escape should not be consumed")
	}
}

// ─── Style ─────────────────────────────────────────────────────

func TestRadioGroup_SetStyle(t *testing.T) {
	rg := NewRadioGroup([]string{"A"})
	style := DefaultRadioGroupStyle()
	style.Normal = buffer.Style{Fg: buffer.Red}
	rg.SetStyle(style)
	if rg.Style().Normal.Fg != buffer.Red {
		t.Error("style not set")
	}
}

// ─── Measure ───────────────────────────────────────────────────

func TestRadioGroup_Measure(t *testing.T) {
	rg := NewRadioGroup([]string{"Short", "VeryLongLabelHere"})
	size := rg.Measure(Constraints{})
	if size.H != 2 {
		t.Errorf("H = %d, want 2", size.H)
	}
	if size.W < 10 {
		t.Errorf("W = %d, should be >= 10", size.W)
	}
}

func TestRadioGroup_Measure_Empty(t *testing.T) {
	rg := NewRadioGroup(nil)
	size := rg.Measure(Constraints{})
	if size.H != 1 {
		t.Errorf("H = %d, want 1", size.H)
	}
}

func TestRadioGroup_Measure_Clamped(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C", "D", "E"})
	size := rg.Measure(Constraints{MaxHeight: 3})
	if size.H > 3 {
		t.Errorf("H = %d, should be clamped to 3", size.H)
	}
}

// ─── Paint ─────────────────────────────────────────────────────

func TestRadioGroup_Paint_NoPanic(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B"})
	rg.SetSelected(0)
	rg.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	rg.Paint(buf)
}

func TestRadioGroup_Paint_ZeroBounds(t *testing.T) {
	rg := NewRadioGroup([]string{"A"})
	buf := buffer.NewBuffer(10, 5)
	rg.Paint(buf)
}

// ─── Misc ──────────────────────────────────────────────────────

func TestRadioGroup_OnChange(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B"})
	changed := false
	rg.OnChange = func(label string, idx int) {
		changed = true
	}
	rg.SetSelected(1)
	if !changed {
		t.Error("OnChange should have fired")
	}
}

func TestRadioGroup_String(t *testing.T) {
	rg := NewRadioGroup([]string{"A"})
	if rg.String() != "RadioGroup" {
		t.Errorf("String = %q, want 'RadioGroup'", rg.String())
	}
}

func TestRadioGroup_Children(t *testing.T) {
	rg := NewRadioGroup([]string{"A"})
	if rg.Children() != nil {
		t.Error("Children should return nil")
	}
}

// ─── Concurrency ───────────────────────────────────────────────

func TestRadioGroup_ConcurrentAccess(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C"})
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			rg.Labels()
			rg.Cursor()
			rg.SelectedIndex()
		}()
		go func() {
			defer wg.Done()
			rg.MoveDown()
			rg.Select()
		}()
		go func() {
			defer wg.Done()
			rg.SetSelected(0)
			rg.SetDisabled(1, false)
		}()
	}
	wg.Wait()
}

func TestRadioGroup_ConcurrentPaint(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C"})
	rg.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			buf := buffer.NewBuffer(20, 3)
			rg.Paint(buf)
		}()
		go func() {
			defer wg.Done()
			rg.MoveDown()
			rg.Select()
		}()
	}
	wg.Wait()
}

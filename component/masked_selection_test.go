package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── MaskedInput Tests ───

func TestMaskedInput_Basic(t *testing.T) {
	mi := NewMaskedInput("##/##/####")
	if mi.Mask() != "##/##/####" {
		t.Error("mask mismatch")
	}
}

func TestMaskedInput_DigitMask(t *testing.T) {
	mi := NewMaskedInput("###-####")
	mi.insertChar('1')
	mi.insertChar('2')
	mi.insertChar('3')
	if mi.Value() != "123-####" {
		t.Errorf("expected '123-####', got '%s'", mi.Value())
	}
	if mi.RawValue() != "123" {
		t.Errorf("expected raw '123', got '%s'", mi.RawValue())
	}
}

func TestMaskedInput_LetterMask(t *testing.T) {
	mi := NewMaskedInput("AA-##")
	mi.insertChar('a')
	mi.insertChar('B')
	mi.insertChar('5')
	if mi.Value() != "aB-5#" {
		t.Errorf("expected 'aB-5#', got '%s'", mi.Value())
	}
}

func TestMaskedInput_UpperCaseMask(t *testing.T) {
	mi := NewMaskedInput("LL-##")
	mi.insertChar('a')
	mi.insertChar('b')
	if mi.Value() != "AB-##" {
		t.Errorf("expected 'AB-##', got '%s'", mi.Value())
	}
}

func TestMaskedInput_LowerCaseMask(t *testing.T) {
	mi := NewMaskedInput("ll-##")
	mi.insertChar('A')
	mi.insertChar('B')
	if mi.Value() != "ab-##" {
		t.Errorf("expected 'ab-##', got '%s'", mi.Value())
	}
}

func TestMaskedInput_AnyCharMask(t *testing.T) {
	mi := NewMaskedInput("**")
	mi.insertChar('x')
	mi.insertChar('1')
	if mi.Value() != "x1" {
		t.Errorf("expected 'x1', got '%s'", mi.Value())
	}
}

func TestMaskedInput_RejectInvalid(t *testing.T) {
	mi := NewMaskedInput("###")
	mi.insertChar('a') // letter, should be rejected for # mask
	if mi.RawValue() != "" {
		t.Error("letter should be rejected for digit mask")
	}
}

func TestMaskedInput_IsComplete(t *testing.T) {
	mi := NewMaskedInput("##")
	if mi.IsComplete() {
		t.Error("should not be complete")
	}
	mi.insertChar('1')
	mi.insertChar('2')
	if !mi.IsComplete() {
		t.Error("should be complete")
	}
}

func TestMaskedInput_Backspace(t *testing.T) {
	mi := NewMaskedInput("####")
	mi.insertChar('1')
	mi.insertChar('2')
	mi.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if mi.RawValue() != "1" {
		t.Errorf("expected '1', got '%s'", mi.RawValue())
	}
}

func TestMaskedInput_SetRawValue(t *testing.T) {
	mi := NewMaskedInput("##/##/####")
	mi.SetRawValue("12312025")
	if mi.Value() != "12/31/2025" {
		t.Errorf("expected '12/31/2025', got '%s'", mi.Value())
	}
}

func TestMaskedInput_HandleKeyNil(t *testing.T) {
	mi := NewMaskedInput("##")
	if mi.HandleKey(nil) {
		t.Error("nil key should not be consumed")
	}
}

func TestMaskedInput_OnChange(t *testing.T) {
	mi := NewMaskedInput("##")
	var changed string
	mi.SetOnChange(func(val string) { changed = val })
	mi.insertChar('1')
	if changed != "1#" {
		t.Errorf("expected '1#', got '%s'", changed)
	}
}

func TestMaskedInput_Measure(t *testing.T) {
	mi := NewMaskedInput("##/##/####")
	s := mi.Measure(Bounded(30, 10))
	if s.W != 10 || s.H != 1 {
		t.Errorf("expected 10x1, got %dx%d", s.W, s.H)
	}
}

func TestMaskedInput_Paint(t *testing.T) {
	mi := NewMaskedInput("##")
	mi.SetBounds(Rect{X: 0, Y: 0, W: 2, H: 1})
	buf := buffer.NewBuffer(2, 1)
	mi.Paint(buf) // should not panic
}

func TestMaskedInput_CursorNavigation(t *testing.T) {
	mi := NewMaskedInput("####")
	mi.insertChar('1')
	mi.insertChar('2')
	mi.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if mi.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", mi.Cursor())
	}
	mi.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if mi.Cursor() != 2 {
		t.Errorf("expected cursor 2, got %d", mi.Cursor())
	}
	mi.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	if mi.Cursor() != 0 {
		t.Errorf("expected cursor 0, got %d", mi.Cursor())
	}
	mi.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	if mi.Cursor() != 2 {
		t.Errorf("expected cursor 2, got %d", mi.Cursor())
	}
}

// ─── SelectionList Tests ───

func TestSelectionList_Basic(t *testing.T) {
	sl := NewSelectionList([]string{"Apple", "Banana", "Cherry"})
	if sl.ItemCount() != 3 {
		t.Error("should have 3 items")
	}
}

func TestSelectionList_Toggle(t *testing.T) {
	sl := NewSelectionList([]string{"A", "B", "C"})
	sl.Toggle(1)
	if !sl.IsSelected(1) {
		t.Error("B should be selected")
	}
	sl.Toggle(1)
	if sl.IsSelected(1) {
		t.Error("B should be deselected")
	}
}

func TestSelectionList_SelectedItems(t *testing.T) {
	sl := NewSelectionList([]string{"A", "B", "C"})
	sl.Toggle(0)
	sl.Toggle(2)
	selected := sl.SelectedItems()
	if len(selected) != 2 || selected[0] != 0 || selected[1] != 2 {
		t.Errorf("expected [0 2], got %v", selected)
	}
}

func TestSelectionList_SelectedLabels(t *testing.T) {
	sl := NewSelectionList([]string{"Apple", "Banana"})
	sl.Toggle(0)
	labels := sl.SelectedLabels()
	if len(labels) != 1 || labels[0] != "Apple" {
		t.Errorf("expected [Apple], got %v", labels)
	}
}

func TestSelectionList_SelectAll(t *testing.T) {
	sl := NewSelectionList([]string{"A", "B", "C"})
	sl.SelectAll()
	if len(sl.SelectedItems()) != 3 {
		t.Error("all should be selected")
	}
}

func TestSelectionList_DeselectAll(t *testing.T) {
	sl := NewSelectionList([]string{"A", "B"})
	sl.SelectAll()
	sl.DeselectAll()
	if len(sl.SelectedItems()) != 0 {
		t.Error("none should be selected")
	}
}

func TestSelectionList_Cursor(t *testing.T) {
	sl := NewSelectionList([]string{"A", "B", "C"})
	sl.SetCursor(2)
	if sl.Cursor() != 2 {
		t.Error("cursor should be 2")
	}
	sl.SetCursor(-1)
	if sl.Cursor() != 0 {
		t.Error("cursor should clamp to 0")
	}
	sl.SetCursor(100)
	if sl.Cursor() != 2 {
		t.Error("cursor should clamp to last")
	}
}

func TestSelectionList_HandleKey(t *testing.T) {
	sl := NewSelectionList([]string{"A", "B"})
	sl.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if sl.Cursor() != 1 {
		t.Error("down should move cursor")
	}
	sl.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if sl.Cursor() != 0 {
		t.Error("up should move cursor")
	}
	sl.HandleKey(&term.KeyEvent{Key: term.KeySpace})
	if !sl.IsSelected(0) {
		t.Error("space should toggle")
	}
}

func TestSelectionList_HandleKeyNil(t *testing.T) {
	sl := NewSelectionList([]string{"A"})
	if sl.HandleKey(nil) {
		t.Error("nil key should not be consumed")
	}
}

func TestSelectionList_VimKeys(t *testing.T) {
	sl := NewSelectionList([]string{"A", "B", "C"})
	sl.HandleKey(&term.KeyEvent{Rune: 'j'})
	if sl.Cursor() != 1 {
		t.Error("j should move down")
	}
	sl.HandleKey(&term.KeyEvent{Rune: 'k'})
	if sl.Cursor() != 0 {
		t.Error("k should move up")
	}
}

func TestSelectionList_SetItems(t *testing.T) {
	sl := NewSelectionList([]string{"A"})
	sl.SetItems([]SelectionItem{{Label: "X", Selected: true}})
	if sl.ItemCount() != 1 {
		t.Error("should have 1 item")
	}
	if !sl.IsSelected(0) {
		t.Error("X should be selected")
	}
}

func TestSelectionList_Disabled(t *testing.T) {
	sl := NewSelectionList([]string{"A", "B"})
	sl.SetDisabled(1, true)
	sl.Toggle(1)
	if sl.IsSelected(1) {
		t.Error("disabled item should not toggle")
	}
}

func TestSelectionList_Measure(t *testing.T) {
	sl := NewSelectionList([]string{"Apple", "Banana"})
	s := sl.Measure(Bounded(50, 10))
	if s.H != 2 {
		t.Errorf("expected height 2, got %d", s.H)
	}
}

func TestSelectionList_Paint(t *testing.T) {
	sl := NewSelectionList([]string{"A", "B"})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 2})
	buf := buffer.NewBuffer(20, 2)
	sl.Paint(buf) // should not panic
}

func TestSelectionList_PaintZeroBounds(t *testing.T) {
	sl := NewSelectionList([]string{"A"})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 10)
	sl.Paint(buf) // should not panic
}

// ─── LineGauge Tests ───

func TestLineGauge_Basic(t *testing.T) {
	lg := NewLineGauge()
	if lg.Percent() != 0 {
		t.Error("should start at 0")
	}
}

func TestLineGauge_SetPercent(t *testing.T) {
	lg := NewLineGauge()
	lg.SetPercent(65)
	if lg.Percent() != 65 {
		t.Error("should be 65")
	}
}

func TestLineGauge_ClampPercent(t *testing.T) {
	lg := NewLineGauge()
	lg.SetPercent(-10)
	if lg.Percent() != 0 {
		t.Error("should clamp to 0")
	}
	lg.SetPercent(200)
	if lg.Percent() != 100 {
		t.Error("should clamp to 100")
	}
}

func TestLineGauge_SetLabel(t *testing.T) {
	lg := NewLineGauge()
	lg.SetLabel("Uploading...")
	// just verify it doesn't crash
}

func TestLineGauge_Measure(t *testing.T) {
	lg := NewLineGauge()
	s := lg.Measure(Bounded(50, 10))
	if s.H != 1 {
		t.Error("should be 1 line high")
	}
}

func TestLineGauge_Paint(t *testing.T) {
	lg := NewLineGauge()
	lg.SetPercent(50)
	lg.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	lg.Paint(buf) // should not panic
}

func TestLineGauge_PaintZeroBounds(t *testing.T) {
	lg := NewLineGauge()
	lg.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 10)
	lg.Paint(buf) // should not panic
}

func TestLineGauge_PaintFull(t *testing.T) {
	lg := NewLineGauge()
	lg.SetPercent(100)
	lg.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	lg.Paint(buf)
	// All cells should be fill char
	if buf.GetCell(0, 0).Rune != '█' {
		t.Error("should be filled at 100%")
	}
}

func TestLineGauge_PaintEmpty(t *testing.T) {
	lg := NewLineGauge()
	lg.SetPercent(0)
	lg.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	lg.Paint(buf)
	if buf.GetCell(0, 0).Rune != '░' {
		t.Error("should be empty at 0%")
	}
}
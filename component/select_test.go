package component

import (
	"sync/atomic"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestSelect_Basic(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "Apple", Value: "apple"},
		{Label: "Banana", Value: "banana"},
		{Label: "Cherry", Value: "cherry"},
	})
	if s.Value() != "" {
		t.Fatal("no selection initially")
	}
	if s.SelectedIndex() != -1 {
		t.Fatal("should be -1 initially")
	}
}

func TestSelect_SetSelectedIndex(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
	})
	s.SetSelectedIndex(1)
	if s.Value() != "b" {
		t.Fatalf("expected 'b', got %q", s.Value())
	}
	if s.Label() != "B" {
		t.Fatalf("expected 'B', got %q", s.Label())
	}
	s.SetSelectedIndex(-5) // out of range
	if s.Value() != "" {
		t.Fatal("should reset to none")
	}
}

func TestSelect_SetOptions(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "Old", Value: "old"},
	})
	s.SetSelectedIndex(0)
	s.SetOptions([]SelectOption{
		{Label: "New1", Value: "n1"},
		{Label: "New2", Value: "n2"},
	})
	if s.SelectedIndex() != -1 {
		t.Fatal("selection should reset when options change")
	}
	opts := s.Options()
	if len(opts) != 2 {
		t.Fatalf("expected 2 options, got %d", len(opts))
	}
}

func TestSelect_OpenClose(t *testing.T) {
	s := NewSelect([]SelectOption{{Label: "X", Value: "x"}})
	if s.IsOpen() {
		t.Fatal("should start closed")
	}
	s.Open()
	if !s.IsOpen() {
		t.Fatal("should be open")
	}
	s.Close()
	if s.IsOpen() {
		t.Fatal("should be closed")
	}
	s.Toggle()
	if !s.IsOpen() {
		t.Fatal("should toggle open")
	}
	s.Toggle()
	if s.IsOpen() {
		t.Fatal("should toggle closed")
	}
}

func TestSelect_OnChange(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
	})
	var changed atomic.Bool
	s.SetOnChange(func(val string, idx int) {
		if val == "b" && idx == 1 {
			changed.Store(true)
		}
	})
	s.SetSelectedIndex(1)
	if !changed.Load() {
		t.Fatal("onChange should have fired")
	}
}

func TestSelect_HandleKey_OpenClose(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
	})
	// Enter opens
	k := &term.KeyEvent{Key: term.KeyEnter}
	if !s.HandleKey(k) {
		t.Fatal("Enter should open Select")
	}
	if !s.IsOpen() {
		t.Fatal("should be open after Enter")
	}
	// Escape closes
	s.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if s.IsOpen() {
		t.Fatal("should be closed after Escape")
	}
}

func TestSelect_HandleKey_NavigateAndSelect(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
		{Label: "C", Value: "c"},
	})
	s.Open()
	// Down to B
	s.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	// Down to C
	s.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	// Enter to select C
	s.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if s.Value() != "c" {
		t.Fatalf("expected 'c', got %q", s.Value())
	}
	if s.IsOpen() {
		t.Fatal("should close after Enter selection")
	}
}

func TestSelect_HandleKey_Vim(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
		{Label: "C", Value: "c"},
	})
	s.Open()
	s.HandleKey(&term.KeyEvent{Rune: 'j'}) // down
	s.HandleKey(&term.KeyEvent{Rune: 'j'}) // down to C
	s.HandleKey(&term.KeyEvent{Rune: 'k'}) // up to B
	s.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if s.Value() != "b" {
		t.Fatalf("expected 'b', got %q", s.Value())
	}
}

func TestSelect_HandleKey_HomeEnd(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
		{Label: "C", Value: "c"},
	})
	s.Open()
	s.HandleKey(&term.KeyEvent{Key: term.KeyEnd}) // go to last
	s.HandleKey(&term.KeyEvent{Key: term.KeyHome}) // go to first
	s.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if s.Value() != "a" {
		t.Fatalf("expected 'a', got %q", s.Value())
	}
}

func TestSelect_HandleKey_ScrollClamp(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
	})
	s.Open()
	// Try to go above first
	s.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	s.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	// Try to go below last
	s.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	s.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	s.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	// Should not panic, cursor clamped
}

func TestSelect_HandleKey_ClosedReturnsFalse(t *testing.T) {
	s := NewSelect([]SelectOption{{Label: "X", Value: "x"}})
	// When closed, non-Enter keys should return false
	k := &term.KeyEvent{Rune: 'j'}
	if s.HandleKey(k) {
		t.Fatal("closed Select should not consume non-Enter keys")
	}
}

func TestSelect_Measure(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "Short", Value: "s"},
		{Label: "A Very Long Option Label", Value: "long"},
	})
	sz := s.Measure(Constraints{MaxWidth: 100})
	if sz.H != 1 {
		t.Fatalf("expected height 1, got %d", sz.H)
	}
	if sz.W < 10 {
		t.Fatalf("expected width >= 10, got %d", sz.W)
	}
}

func TestSelect_MeasureMaxWidth(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "Long Label Here", Value: "x"},
	})
	sz := s.Measure(Constraints{MaxWidth: 5})
	if sz.W > 5 {
		t.Fatalf("expected width <= 5, got %d", sz.W)
	}
}

func TestSelect_PaintCollapsed(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "Hello", Value: "hello"},
	})
	s.SetSelectedIndex(0)
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	s.Paint(buf)
	// Check "Hello" appears
	cell := buf.GetCell(0, 0)
	if cell.Rune != 'H' {
		t.Fatalf("expected 'H', got %q", cell.Rune)
	}
}

func TestSelect_PaintOpen(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "Option1", Value: "o1"},
		{Label: "Option2", Value: "o2"},
		{Label: "Option3", Value: "o3"},
	})
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	s.Open()
	buf := buffer.NewBuffer(20, 4)
	s.Paint(buf)
	// Popup should have options at y=1,2,3
	cell := buf.GetCell(0, 1)
	if cell.Rune != 'O' {
		t.Fatalf("expected popup row at y=1, got %q", cell.Rune)
	}
}

func TestSelect_PaintNoSelection(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "A", Value: "a"},
	})
	s.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 1})
	buf := buffer.NewBuffer(15, 1)
	s.Paint(buf)
	// Should show "(none)"
	cell := buf.GetCell(0, 0)
	if cell.Rune != '(' {
		t.Fatalf("expected '(' for no selection, got %q", cell.Rune)
	}
}

func TestSelect_PaintZeroBounds(t *testing.T) {
	s := NewSelect([]SelectOption{{Label: "X", Value: "x"}})
	s.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(0, 0)
	s.Paint(buf) // should not panic
}

func TestSelect_HandleMouse(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
		{Label: "C", Value: "c"},
	})
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	// Click opens
	s.HandleMouse(5, 0, "down")
	if !s.IsOpen() {
		t.Fatal("click should open Select")
	}
	// Click on option B (y=2)
	s.HandleMouse(5, 2, "down")
	if s.Value() != "b" {
		t.Fatalf("expected 'b' from click, got %q", s.Value())
	}
	if s.IsOpen() {
		t.Fatal("should close after selecting")
	}
}

func TestSelect_HandleMouseMove(t *testing.T) {
	s := NewSelect([]SelectOption{{Label: "X", Value: "x"}})
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	// Mouse move should not toggle
	if s.HandleMouse(5, 0, "move") {
		t.Fatal("mouse move should not consume")
	}
}

func TestSelect_HandleMouseOutsideCloses(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
	})
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	s.Open()
	// Click far away
	s.HandleMouse(50, 20, "down")
	if s.IsOpen() {
		t.Fatal("should close when clicking outside")
	}
}

func TestSelect_SetPopupHeight(t *testing.T) {
	s := NewSelect([]SelectOption{{Label: "X", Value: "x"}})
	s.SetPopupHeight(5)
	// Just verify no panic
}

func TestSelect_SetWidth(t *testing.T) {
	s := NewSelect([]SelectOption{{Label: "X", Value: "x"}})
	s.SetWidth(30)
	sz := s.Measure(Unbounded())
	if sz.W < 30 {
		t.Fatalf("expected width >= 30, got %d", sz.W)
	}
}

func TestSelect_SetStyle(t *testing.T) {
	s := NewSelect([]SelectOption{{Label: "X", Value: "x"}})
	s.SetStyle(SelectStyle{})
	s.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	s.Paint(buf) // should not panic with empty style
}

func TestSelect_EmptyOptions(t *testing.T) {
	s := NewSelect(nil)
	s.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	s.Open()
	buf := buffer.NewBuffer(10, 1)
	s.Paint(buf) // should not panic
	s.HandleKey(&term.KeyEvent{Key: term.KeyDown}) // no panic on empty
}

func TestSelect_Concurrent(t *testing.T) {
	s := NewSelect([]SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
	})
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			s.SetSelectedIndex(i % 2)
		}
		close(done)
	}()
	for i := 0; i < 100; i++ {
		_ = s.Value()
		_ = s.IsOpen()
	}
	<-done
}
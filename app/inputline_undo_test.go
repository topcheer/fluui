package app

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

// --- Undo/Redo Stack unit tests ---

func TestUndoStack_New(t *testing.T) {
	s := newUndoStack()
	if s == nil {
		t.Fatal("newUndoStack returned nil")
	}
	if s.canUndo() {
		t.Error("new stack should not be able to undo")
	}
	if s.canRedo() {
		t.Error("new stack should not be able to redo")
	}
}

func TestUndoStack_SaveAndUndo(t *testing.T) {
	il := NewInputLine("> ")
	il.buf = []rune("hello")
	il.cursor = 5

	il.saveUndo()
	il.buf = []rune("hello!")
	il.cursor = 6

	if !il.CanUndo() {
		t.Fatal("CanUndo should be true after save + edit")
	}
	if !il.Undo() {
		t.Fatal("Undo should succeed")
	}
	if il.Text() != "hello" {
		t.Errorf("after undo text = %q, want %q", il.Text(), "hello")
	}
	if il.Cursor() != 5 {
		t.Errorf("after undo cursor = %d, want 5", il.Cursor())
	}
}

func TestUndoStack_RedoAfterUndo(t *testing.T) {
	il := NewInputLine("> ")
	il.saveUndo()
	il.buf = []rune("world")
	il.cursor = 5

	il.Undo()

	if !il.CanRedo() {
		t.Fatal("CanRedo should be true after undo")
	}
	if !il.Redo() {
		t.Fatal("Redo should succeed")
	}
	if il.Text() != "world" {
		t.Errorf("after redo text = %q, want %q", il.Text(), "world")
	}
	if il.Cursor() != 5 {
		t.Errorf("after redo cursor = %d, want 5", il.Cursor())
	}
}

func TestUndoStack_RedoClearedOnNewEdit(t *testing.T) {
	il := NewInputLine("> ")
	il.saveUndo()
	il.buf = []rune("a")
	il.cursor = 1

	il.Undo() // undo back to ""

	// New edit should clear redo
	il.saveUndo()
	il.buf = []rune("b")
	il.cursor = 1

	if il.CanRedo() {
		t.Error("redo should be cleared after new edit")
	}
	if il.Text() != "b" {
		t.Errorf("text = %q, want %q", il.Text(), "b")
	}
}

func TestUndoStack_MultipleUndos(t *testing.T) {
	il := NewInputLine("> ")

	// Save 3 states: "", "a", "ab"
	il.saveUndo()
	il.buf = []rune("a")
	il.cursor = 1

	il.saveUndo()
	il.buf = []rune("ab")
	il.cursor = 2

	il.saveUndo()
	il.buf = []rune("abc")
	il.cursor = 3

	if il.UndoCount() != 3 {
		t.Fatalf("UndoCount = %d, want 3", il.UndoCount())
	}

	// Undo back to "ab"
	il.Undo()
	if il.Text() != "ab" {
		t.Errorf("undo 1: text = %q, want %q", il.Text(), "ab")
	}

	// Undo back to "a"
	il.Undo()
	if il.Text() != "a" {
		t.Errorf("undo 2: text = %q, want %q", il.Text(), "a")
	}

	// Undo back to ""
	il.Undo()
	if il.Text() != "" {
		t.Errorf("undo 3: text = %q, want %q", il.Text(), "")
	}

	// No more undos
	if il.Undo() {
		t.Error("undo should fail when stack is empty")
	}
}

func TestUndoStack_UndoNoOp(t *testing.T) {
	il := NewInputLine("> ")
	il.buf = []rune("test")

	if il.Undo() {
		t.Error("Undo should return false on empty stack")
	}
	if il.Redo() {
		t.Error("Redo should return false on empty stack")
	}
}

func TestUndoStack_ClearHistory(t *testing.T) {
	il := NewInputLine("> ")
	il.saveUndo()
	il.buf = []rune("x")
	il.saveUndo()
	il.buf = []rune("xy")

	if il.UndoCount() != 2 {
		t.Fatalf("UndoCount = %d, want 2", il.UndoCount())
	}

	il.ClearUndoHistory()
	if il.CanUndo() {
		t.Error("CanUndo should be false after ClearUndoHistory")
	}
	if il.CanRedo() {
		t.Error("CanRedo should be false after ClearUndoHistory")
	}
}

func TestUndoStack_MaxLimit(t *testing.T) {
	il := NewInputLine("> ")

	// Push more than maxUndoStates
	for i := 0; i < maxUndoStates+50; i++ {
		il.saveUndo()
		il.buf = append(il.buf, 'a')
		il.cursor++
	}

	if il.UndoCount() > maxUndoStates {
		t.Errorf("UndoCount = %d, should be <= %d", il.UndoCount(), maxUndoStates)
	}
}

// --- InputLine HandleKey integration tests ---

func TestInputLine_CtrlZ_Undo(t *testing.T) {
	il := NewInputLine("> ")

	// Type "hi"
	il.HandleKey(&term.KeyEvent{Rune: 'h', Key: term.KeyUnknown})
	il.HandleKey(&term.KeyEvent{Rune: 'i', Key: term.KeyUnknown})

	if il.Text() != "hi" {
		t.Fatalf("text = %q, want %q", il.Text(), "hi")
	}

	// Ctrl+Z should undo the 'i'
	il.HandleKey(&term.KeyEvent{Rune: 'z', Modifiers: term.ModCtrl})

	if il.Text() != "h" {
		t.Errorf("after Ctrl+Z: text = %q, want %q", il.Text(), "h")
	}
}

func TestInputLine_CtrlZ_CtrlShiftZ(t *testing.T) {
	il := NewInputLine("> ")

	// Type "hello"
	for _, r := range "hello" {
		il.HandleKey(&term.KeyEvent{Rune: r, Key: term.KeyUnknown})
	}

	// Undo twice
	il.HandleKey(&term.KeyEvent{Rune: 'z', Modifiers: term.ModCtrl})
	il.HandleKey(&term.KeyEvent{Rune: 'z', Modifiers: term.ModCtrl})

	if il.Text() != "hel" {
		t.Errorf("after 2 undos: text = %q, want %q", il.Text(), "hel")
	}

	// Ctrl+Shift+Z (redo)
	il.HandleKey(&term.KeyEvent{Rune: 'Z', Modifiers: term.ModCtrl | term.ModShift})

	if il.Text() != "hell" {
		t.Errorf("after redo: text = %q, want %q", il.Text(), "hell")
	}
}

func TestInputLine_CtrlY_Redo(t *testing.T) {
	il := NewInputLine("> ")

	il.HandleKey(&term.KeyEvent{Rune: 'a', Key: term.KeyUnknown})
	il.HandleKey(&term.KeyEvent{Rune: 'b', Key: term.KeyUnknown})

	// Undo
	il.HandleKey(&term.KeyEvent{Rune: 'z', Modifiers: term.ModCtrl})
	if il.Text() != "a" {
		t.Fatalf("after undo: text = %q, want %q", il.Text(), "a")
	}

	// Ctrl+Y = redo
	il.HandleKey(&term.KeyEvent{Rune: 'y', Modifiers: term.ModCtrl})
	if il.Text() != "ab" {
		t.Errorf("after Ctrl+Y redo: text = %q, want %q", il.Text(), "ab")
	}
}

func TestInputLine_BackspaceUndo(t *testing.T) {
	il := NewInputLine("> ")

	// Type "hello"
	for _, r := range "hello" {
		il.HandleKey(&term.KeyEvent{Rune: r, Key: term.KeyUnknown})
	}

	// Backspace twice
	il.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	il.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})

	if il.Text() != "hel" {
		t.Fatalf("after 2 backspaces: text = %q, want %q", il.Text(), "hel")
	}

	// Undo should restore 'o'
	il.Undo()
	if il.Text() != "hell" {
		t.Errorf("after undo backspace: text = %q, want %q", il.Text(), "hell")
	}
}

func TestInputLine_CtrlU_Undo(t *testing.T) {
	il := NewInputLine("> ")

	for _, r := range "hello" {
		il.HandleKey(&term.KeyEvent{Rune: r, Key: term.KeyUnknown})
	}

	// Ctrl+U clears all
	il.HandleKey(&term.KeyEvent{Rune: 'u', Modifiers: term.ModCtrl})
	if il.Text() != "" {
		t.Fatalf("after Ctrl+U: text = %q, want empty", il.Text())
	}

	// Undo should restore "hello"
	il.Undo()
	if il.Text() != "hello" {
		t.Errorf("after undo Ctrl+U: text = %q, want %q", il.Text(), "hello")
	}
}

func TestInputLine_CtrlW_Undo(t *testing.T) {
	il := NewInputLine("> ")

	for _, r := range "hello world" {
		il.HandleKey(&term.KeyEvent{Rune: r, Key: term.KeyUnknown})
	}

	// Ctrl+W deletes "world"
	il.HandleKey(&term.KeyEvent{Rune: 'w', Modifiers: term.ModCtrl})
	if il.Text() != "hello " {
		t.Fatalf("after Ctrl+W: text = %q, want %q", il.Text(), "hello ")
	}

	// Undo should restore "hello world"
	il.Undo()
	if il.Text() != "hello world" {
		t.Errorf("after undo Ctrl+W: text = %q, want %q", il.Text(), "hello world")
	}
}

func TestInputLine_UndoDoesNotAffectSubmit(t *testing.T) {
	var submitted string
	il := NewInputLineWithHandler("> ", func(s string) { submitted = s })

	il.HandleKey(&term.KeyEvent{Rune: 'h', Key: term.KeyUnknown})
	il.HandleKey(&term.KeyEvent{Rune: 'i', Key: term.KeyUnknown})

	// Submit
	il.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if submitted != "hi" {
		t.Fatalf("submitted = %q, want %q", submitted, "hi")
	}
	// Input cleared after submit
	if il.Text() != "" {
		t.Errorf("after submit text = %q, want empty", il.Text())
	}
}

func TestInputLine_UndoCount(t *testing.T) {
	il := NewInputLine("> ")

	if il.UndoCount() != 0 {
		t.Fatalf("initial UndoCount = %d, want 0", il.UndoCount())
	}

	il.HandleKey(&term.KeyEvent{Rune: 'a', Key: term.KeyUnknown})
	if il.UndoCount() != 1 {
		t.Errorf("after 1 char: UndoCount = %d, want 1", il.UndoCount())
	}

	il.HandleKey(&term.KeyEvent{Rune: 'b', Key: term.KeyUnknown})
	if il.UndoCount() != 2 {
		t.Errorf("after 2 chars: UndoCount = %d, want 2", il.UndoCount())
	}
}

func TestInputLine_RedoCount(t *testing.T) {
	il := NewInputLine("> ")

	il.HandleKey(&term.KeyEvent{Rune: 'a', Key: term.KeyUnknown})
	il.HandleKey(&term.KeyEvent{Rune: 'b', Key: term.KeyUnknown})

	if il.RedoCount() != 0 {
		t.Fatalf("initial RedoCount = %d, want 0", il.RedoCount())
	}

	il.Undo()
	il.Undo()

	if il.RedoCount() != 2 {
		t.Errorf("after 2 undos: RedoCount = %d, want 2", il.RedoCount())
	}
}

func TestInputLine_ConcurrentUndo(t *testing.T) {
	il := NewInputLine("> ")

	// Seed with some text so undo has data to work with
	for i := 0; i < 20; i++ {
		il.HandleKey(&term.KeyEvent{Rune: 'x', Key: term.KeyUnknown})
	}

	var wg sync.WaitGroup

	// Undo/Redo operations are serialized by undoStack's internal mutex.
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 30; i++ {
			il.Undo()
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 30; i++ {
			il.Redo()
		}
	}()

	wg.Wait()
}

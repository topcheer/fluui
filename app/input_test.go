package app

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestInputLineCreation(t *testing.T) {
	i := NewInputLine("> ")
	if i == nil {
		t.Fatal("expected non-nil InputLine")
	}
	if i.prompt != "> " {
		t.Fatalf("expected prompt '> ', got %q", i.prompt)
	}
	if !i.Empty() {
		t.Fatal("expected empty on creation")
	}
	if i.Cursor() != 0 {
		t.Fatalf("expected cursor 0, got %d", i.Cursor())
	}
	if i.Len() != 0 {
		t.Fatalf("expected len 0, got %d", i.Len())
	}
	if i.Text() != "" {
		t.Fatalf("expected empty text, got %q", i.Text())
	}
}

func TestInputLineTypeChars(t *testing.T) {
	i := NewInputLine("> ")

	for _, r := range "hello" {
		i.HandleKey(&term.KeyEvent{Rune: r})
	}

	if i.Text() != "hello" {
		t.Fatalf("expected 'hello', got %q", i.Text())
	}
	if i.Len() != 5 {
		t.Fatalf("expected len 5, got %d", i.Len())
	}
	if i.Cursor() != 5 {
		t.Fatalf("expected cursor at 5, got %d", i.Cursor())
	}
}

func TestInputLineBackspace(t *testing.T) {
	i := NewInputLine("> ")
	for _, r := range "hello" {
		i.HandleKey(&term.KeyEvent{Rune: r})
	}

	i.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if i.Text() != "hell" {
		t.Fatalf("expected 'hell', got %q", i.Text())
	}
	if i.Cursor() != 4 {
		t.Fatalf("expected cursor at 4, got %d", i.Cursor())
	}

	// Backspace on empty buffer should not panic.
	i2 := NewInputLine("> ")
	i2.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if !i2.Empty() {
		t.Fatal("expected still empty")
	}
}

func TestInputLineEnter(t *testing.T) {
	var submitted string
	i := NewInputLineWithHandler("> ", func(text string) {
		submitted = text
	})

	for _, r := range "test" {
		i.HandleKey(&term.KeyEvent{Rune: r})
	}

	i.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	if submitted != "test" {
		t.Fatalf("expected onSubmit to receive 'test', got %q", submitted)
	}
	if !i.Empty() {
		t.Fatal("expected buffer cleared after submit")
	}
	if i.Cursor() != 0 {
		t.Fatalf("expected cursor 0 after clear, got %d", i.Cursor())
	}
}

func TestInputLineClear(t *testing.T) {
	i := NewInputLine("> ")
	for _, r := range "hello world" {
		i.HandleKey(&term.KeyEvent{Rune: r})
	}

	// Ctrl+U clears everything.
	i.HandleKey(&term.KeyEvent{Rune: 'u', Modifiers: term.ModCtrl})

	if !i.Empty() {
		t.Fatal("expected empty after Ctrl+U")
	}
	if i.Cursor() != 0 {
		t.Fatalf("expected cursor 0, got %d", i.Cursor())
	}
}

func TestInputLineCursor(t *testing.T) {
	i := NewInputLine("> ")
	for _, r := range "hello" {
		i.HandleKey(&term.KeyEvent{Rune: r})
	}
	// cursor is at 5 (end)

	// Left → cursor 4
	i.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if i.Cursor() != 4 {
		t.Fatalf("expected cursor 4, got %d", i.Cursor())
	}

	// Left x3 → cursor 1
	i.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	i.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	i.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if i.Cursor() != 1 {
		t.Fatalf("expected cursor 1, got %d", i.Cursor())
	}

	// Home → cursor 0
	i.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	if i.Cursor() != 0 {
		t.Fatalf("expected cursor 0, got %d", i.Cursor())
	}

	// Left at 0 stays at 0
	i.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if i.Cursor() != 0 {
		t.Fatalf("expected cursor 0, got %d", i.Cursor())
	}

	// Right → cursor 1
	i.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if i.Cursor() != 1 {
		t.Fatalf("expected cursor 1, got %d", i.Cursor())
	}

	// End → cursor 5
	i.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	if i.Cursor() != 5 {
		t.Fatalf("expected cursor 5, got %d", i.Cursor())
	}

	// Right at end stays at end
	i.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if i.Cursor() != 5 {
		t.Fatalf("expected cursor 5, got %d", i.Cursor())
	}
}

func TestInputLineInsertMiddle(t *testing.T) {
	i := NewInputLine("> ")
	for _, r := range "helo" {
		i.HandleKey(&term.KeyEvent{Rune: r})
	}
	// cursor at 4, text = "helo"

	// Move cursor to position 2 (between 'e' and 'l').
	i.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	i.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if i.Cursor() != 2 {
		t.Fatalf("expected cursor 2, got %d", i.Cursor())
	}

	// Insert 'l' at cursor.
	i.HandleKey(&term.KeyEvent{Rune: 'l'})
	if i.Text() != "hello" {
		t.Fatalf("expected 'hello', got %q", i.Text())
	}
	if i.Cursor() != 3 {
		t.Fatalf("expected cursor at 3, got %d", i.Cursor())
	}
}

func TestInputLinePaint(t *testing.T) {
	i := NewInputLine("> ")
	i.SetText("hi")

	// Set bounds and paint into a small buffer.
	i.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	i.Paint(buf)

	// Check prompt characters are present.
	cell0 := buf.Cells[0]
	if cell0.Rune != '>' {
		t.Fatalf("expected '>' at x=0, got %c", cell0.Rune)
	}
	cell1 := buf.Cells[1]
	if cell1.Rune != ' ' {
		t.Fatalf("expected ' ' at x=1, got %c", cell1.Rune)
	}

	// Check text 'h' at x=2.
	cell2 := buf.Cells[2]
	if cell2.Rune != 'h' {
		t.Fatalf("expected 'h' at x=2, got %c", cell2.Rune)
	}

	// Check text 'i' at x=3.
	cell3 := buf.Cells[3]
	if cell3.Rune != 'i' {
		t.Fatalf("expected 'i' at x=3, got %c", cell3.Rune)
	}
}

func TestInputLineEmpty(t *testing.T) {
	i := NewInputLine("> ")
	if !i.Empty() {
		t.Fatal("expected empty on creation")
	}

	i.HandleKey(&term.KeyEvent{Rune: 'x'})
	if i.Empty() {
		t.Fatal("expected not empty after typing")
	}

	i.Clear()
	if !i.Empty() {
		t.Fatal("expected empty after Clear")
	}
}

func TestInputLineWordDelete(t *testing.T) {
	i := NewInputLine("> ")
	for _, r := range "hello world" {
		i.HandleKey(&term.KeyEvent{Rune: r})
	}
	// cursor at 11, text = "hello world"

	// Ctrl+W deletes "world".
	i.HandleKey(&term.KeyEvent{Rune: 'w', Modifiers: term.ModCtrl})
	if i.Text() != "hello " {
		t.Fatalf("expected 'hello ', got %q", i.Text())
	}
	if i.Cursor() != 6 {
		t.Fatalf("expected cursor at 6, got %d", i.Cursor())
	}

	// Ctrl+W again deletes the space and "hello".
	i.HandleKey(&term.KeyEvent{Rune: 'w', Modifiers: term.ModCtrl})
	if i.Text() != "" {
		t.Fatalf("expected empty, got %q", i.Text())
	}
	if i.Cursor() != 0 {
		t.Fatalf("expected cursor 0, got %d", i.Cursor())
	}
}

func TestInputLineCtrlA(t *testing.T) {
	i := NewInputLine("> ")
	for _, r := range "abc" {
		i.HandleKey(&term.KeyEvent{Rune: r})
	}
	// cursor at 3

	i.HandleKey(&term.KeyEvent{Rune: 'a', Modifiers: term.ModCtrl})
	if i.Cursor() != 0 {
		t.Fatalf("expected cursor 0 after Ctrl+A, got %d", i.Cursor())
	}
}

func TestInputLineCtrlE(t *testing.T) {
	i := NewInputLine("> ")
	for _, r := range "abc" {
		i.HandleKey(&term.KeyEvent{Rune: r})
	}
	// cursor at 3, move to start
	i.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	if i.Cursor() != 0 {
		t.Fatal("expected cursor 0")
	}

	i.HandleKey(&term.KeyEvent{Rune: 'e', Modifiers: term.ModCtrl})
	if i.Cursor() != 3 {
		t.Fatalf("expected cursor 3 after Ctrl+E, got %d", i.Cursor())
	}
}

func TestInputLineSetText(t *testing.T) {
	i := NewInputLine("> ")
	i.SetText("preset")
	if i.Text() != "preset" {
		t.Fatalf("expected 'preset', got %q", i.Text())
	}
	if i.Cursor() != 6 {
		t.Fatalf("expected cursor at end (6), got %d", i.Cursor())
	}
}


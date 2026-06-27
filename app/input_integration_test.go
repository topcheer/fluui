package app

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestInputLineInChatApp(t *testing.T) {
	app := NewChatApp(80, 24)
	il := NewInputLine("> ")
	app.SetInputLine(il)

	if app.InputLine() != il {
		t.Fatal("expected InputLine() to return the attached line")
	}

	// Type characters through ChatApp.HandleKey
	for _, r := range "hello" {
		app.HandleKey(&term.KeyEvent{Rune: r})
	}

	if il.Text() != "hello" {
		t.Fatalf("expected InputLine text 'hello', got %q", il.Text())
	}
	if il.Len() != 5 {
		t.Fatalf("expected len 5, got %d", il.Len())
	}
}

func TestInputLineSubmit(t *testing.T) {
	app := NewChatApp(80, 24)

	var submitted string
	submittedCalled := false
	app.OnSubmit(func(text string) {
		submitted = text
		submittedCalled = true
	})

	// Type characters
	for _, r := range "test message" {
		app.HandleKey(&term.KeyEvent{Rune: r})
	}

	// Press Enter
	app.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	if !submittedCalled {
		t.Fatal("expected onSubmit callback to be called on Enter")
	}
	if submitted != "test message" {
		t.Fatalf("expected submitted text 'test message', got %q", submitted)
	}

	// InputLine should be cleared after submit
	if !app.InputLine().Empty() {
		t.Fatal("expected InputLine to be empty after submit")
	}
}

func TestInputLineKeyRouting(t *testing.T) {
	app := NewChatApp(80, 24)
	il := NewInputLine("> ")
	app.SetInputLine(il)

	// Custom key handler to detect if scroll view received the key.
	scrollKeyReceived := false
	app.OnKey(func(key *term.KeyEvent) {
		// onKey is only reached if HandleKey falls through past the InputLine.
		scrollKeyReceived = true
	})

	// Type a printable character — InputLine should consume it.
	scrollKeyReceived = false
	app.HandleKey(&term.KeyEvent{Rune: 'a'})
	if scrollKeyReceived {
		t.Fatal("printable char should be consumed by InputLine, not reach scroll handler")
	}
	if il.Text() != "a" {
		t.Fatalf("expected InputLine text 'a', got %q", il.Text())
	}

	// Backspace — InputLine should consume it.
	scrollKeyReceived = false
	app.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if scrollKeyReceived {
		t.Fatal("backspace should be consumed by InputLine")
	}
	if !il.Empty() {
		t.Fatal("expected InputLine empty after backspace")
	}

	// Left arrow — InputLine consumes it.
	scrollKeyReceived = false
	app.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if scrollKeyReceived {
		t.Fatal("Left arrow should be consumed by InputLine")
	}

	// Enter — InputLine consumes it.
	scrollKeyReceived = false
	app.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if scrollKeyReceived {
		t.Fatal("Enter should be consumed by InputLine")
	}

	// Ctrl+U — InputLine consumes it.
	scrollKeyReceived = false
	app.HandleKey(&term.KeyEvent{Rune: 'u', Modifiers: term.ModCtrl})
	if scrollKeyReceived {
		t.Fatal("Ctrl+U should be consumed by InputLine")
	}
}

func TestInputLineRenderWithComponent(t *testing.T) {
	app := NewChatApp(80, 24)
	il := NewInputLine("> ")
	il.SetText("test input")
	app.SetInputLine(il)

	buf := buffer.NewBuffer(80, 24)
	app.Render(buf)

	// The InputLine is rendered on the last row (y=23).
	// The prompt ">" should appear at x=1, y=23.
	cell := buf.GetCell(1, 23)
	if cell.Rune != '>' {
		t.Fatalf("expected '>' at (1,23), got %c", cell.Rune)
	}

	// The text 't' should appear at x=3, y=23 (after "> ").
	cell = buf.GetCell(3, 23)
	if cell.Rune != 't' {
		t.Fatalf("expected 't' at (3,23), got %c", cell.Rune)
	}

	// Separator line at y=22.
	cell = buf.GetCell(0, 22)
	if cell.Rune != '─' {
		t.Fatalf("expected separator '─' at (0,22), got %c", cell.Rune)
	}
}

func TestInputLineRenderFallbackNoComponent(t *testing.T) {
	app := NewChatApp(80, 24)
	app.SetInputHeight(2)

	buf := buffer.NewBuffer(80, 24)
	app.Render(buf)

	// Without InputLine component, static prompt "▶ " at (1, 23).
	cell := buf.GetCell(1, 23)
	if cell.Rune != '▶' {
		t.Fatalf("expected '▶' at (1,23) fallback prompt, got %c", cell.Rune)
	}
}

func TestInputLineOnSubmitCreatesDefault(t *testing.T) {
	app := NewChatApp(80, 24)

	// OnSubmit should create a default InputLine if none is set.
	called := false
	app.OnSubmit(func(text string) {
		called = true
	})

	if app.InputLine() == nil {
		t.Fatal("expected OnSubmit to create a default InputLine")
	}

	// Type and submit.
	for _, r := range "hi" {
		app.HandleKey(&term.KeyEvent{Rune: r})
	}
	app.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	if !called {
		t.Fatal("expected submit callback to fire")
	}
}

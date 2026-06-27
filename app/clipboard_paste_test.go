package app

import (
	"testing"
)

func TestHandleClipboard_NoCallback(t *testing.T) {
	app := NewChatApp(80, 24)

	// No callback, no InputLine → should return false
	consumed := app.HandleClipboard("pasted text")
	if consumed {
		t.Error("expected false with no callback and no InputLine")
	}
}

func TestHandleClipboard_WithCallback(t *testing.T) {
	app := NewChatApp(80, 24)

	var got string
	app.OnClipboard(func(text string) {
		got = text
	})

	consumed := app.HandleClipboard("clipboard content")
	if !consumed {
		t.Error("expected true with callback")
	}
	if got != "clipboard content" {
		t.Errorf("got %q, want %q", got, "clipboard content")
	}
}

func TestHandleClipboard_WithInputLine(t *testing.T) {
	app := NewChatApp(80, 24)
	app.OnSubmit(func(text string) {}) // attaches InputLine

	// No clipboard callback, but InputLine exists → insert text
	consumed := app.HandleClipboard("pasted!")
	if !consumed {
		t.Error("expected true with InputLine")
	}

	// Verify text was inserted into InputLine
	text := app.inputLine.Text()
	if text != "pasted!" {
		t.Errorf("inputLine text: got %q, want %q", text, "pasted!")
	}
}

func TestHandleClipboard_CallbackOverridesInputLine(t *testing.T) {
	app := NewChatApp(80, 24)
	app.OnSubmit(func(text string) {}) // attaches InputLine

	var got string
	app.OnClipboard(func(text string) {
		got = text
	})

	consumed := app.HandleClipboard("from clipboard")
	if !consumed {
		t.Error("expected true")
	}
	if got != "from clipboard" {
		t.Errorf("callback: got %q, want %q", got, "from clipboard")
	}

	// Callback takes priority — InputLine should NOT receive the text
	if app.inputLine.Text() != "" {
		t.Errorf("InputLine should be empty, got %q", app.inputLine.Text())
	}
}

func TestInsertText(t *testing.T) {
	il := NewInputLine("> ")

	// Insert at beginning (cursor at 0)
	il.InsertText("Hello")
	if il.Text() != "Hello" {
		t.Errorf("got %q, want %q", il.Text(), "Hello")
	}
	if il.cursor != 5 {
		t.Errorf("cursor: got %d, want 5", il.cursor)
	}

	// Move cursor to middle and insert
	il.cursor = 2
	il.InsertText("XX")
	if il.Text() != "HeXXllo" {
		t.Errorf("got %q, want %q", il.Text(), "HeXXllo")
	}
	if il.cursor != 4 {
		t.Errorf("cursor: got %d, want 4", il.cursor)
	}
}

func TestInsertText_EmptyString(t *testing.T) {
	il := NewInputLine("> ")
	il.InsertText("")
	if il.Text() != "" {
		t.Errorf("got %q, want empty", il.Text())
	}
	if il.cursor != 0 {
		t.Errorf("cursor: got %d, want 0", il.cursor)
	}
}

func TestInsertText_Unicode(t *testing.T) {
	il := NewInputLine("> ")
	il.InsertText("你好世界")
	if il.Text() != "你好世界" {
		t.Errorf("got %q, want %q", il.Text(), "你好世界")
	}
	if il.cursor != 4 {
		t.Errorf("cursor: got %d, want 4", il.cursor)
	}
}

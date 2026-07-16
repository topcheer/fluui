package textinput

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P234: cover View(), InsertRune(), SetStyle(), SetCursorMode() — all at 0%

func TestTextInput_View_P234(t *testing.T) {
	m := New()
	m.SetValue("hello")
	if m.View() != "hello" {
		t.Errorf("View = %q, want 'hello'", m.View())
	}
}

func TestTextInput_InsertRune_P234(t *testing.T) {
	m := New()
	m.SetValue("abc")
	m.SetCursor(1) // cursor between a and b
	m.InsertRune('X')
	if m.Value() != "aXbc" {
		t.Errorf("Value = %q, want 'aXbc'", m.Value())
	}
}

func TestTextInput_SetStyle_P234(t *testing.T) {
	m := New()
	style := buffer.Style{Fg: buffer.RGB(255, 0, 0)}
	m.SetStyle(style)
	// No crash — style is applied to underlying TextInput
}

func TestTextInput_SetCursorMode_P234(t *testing.T) {
	m := New()
	m.SetCursorMode(CursorBlink)
	m.SetCursorMode(CursorStatic)
	m.SetCursorMode(CursorHide)
}

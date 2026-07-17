package textinput

import "testing"

func TestModel_SetHeight_P283(t *testing.T) {
	m := New()
	m.SetHeight(5) // no-op, shouldn't panic
}

func TestModel_Close_P283(t *testing.T) {
	m := New()
	m.Close() // no-op
}

func TestModel_SetCursorMode_P283(t *testing.T) {
	m := New()
	m.SetCursorMode(CursorStatic)
	m.SetCursorMode(CursorHide)
	m.SetCursorMode(CursorBlink)
}

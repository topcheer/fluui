package event

import "testing"

func TestDispatch_MouseHandler_P249(t *testing.T) {
	d := NewDispatcher()
	called := false
	d.OnMouse(func(e Event) bool { called = true; return true })
	d.Dispatch(Event{Type: TypeMouse})
	if !called { t.Error("mouse handler should be called") }
}

func TestDispatch_ResizeHandler_P249(t *testing.T) {
	d := NewDispatcher()
	called := false
	d.OnResize(func(e Event) bool { called = true; return true })
	d.Dispatch(Event{Type: TypeResize, Width: 80, Height: 24})
	if !called { t.Error("resize handler should be called") }
}

func TestDispatch_PasteHandler_P249(t *testing.T) {
	d := NewDispatcher()
	called := false
	d.OnPaste(func(e Event) bool { called = true; return true })
	d.Dispatch(Event{Type: TypePaste, Paste: "hello"})
	if !called { t.Error("paste handler should be called") }
}

func TestDispatch_PasteNoHandler_P249(t *testing.T) {
	d := NewDispatcher()
	// No handler registered → should not panic
	d.Dispatch(Event{Type: TypePaste, Paste: "text"})
}

func TestDispatch_ResizeNoHandler_P249(t *testing.T) {
	d := NewDispatcher()
	d.Dispatch(Event{Type: TypeResize, Width: 80, Height: 24})
}

func TestReadRaw_NilTerminal_P249(t *testing.T) {
	l := &Loop{terminal: nil}
	l.readRaw()
}

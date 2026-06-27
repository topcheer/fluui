package event

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

func TestNewDispatcher(t *testing.T) {
	d := NewDispatcher()
	if d == nil {
		t.Fatal("NewDispatcher() returned nil")
	}
}

func TestDispatch_UnhandledReturnsFalse(t *testing.T) {
	d := NewDispatcher()
	// No handlers registered — should return false
	ev := Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'a'}}
	if d.Dispatch(ev) {
		t.Error("Dispatch with no handlers should return false")
	}
}

func TestDispatch_KeyBinding(t *testing.T) {
	d := NewDispatcher()
	called := false
	d.BindKey(KeyShortcut{Rune: 'q'}, func(e Event) bool {
		called = true
		return true
	})

	ev := Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'q'}}
	result := d.Dispatch(ev)
	if !result {
		t.Error("Dispatch should return true for consumed event")
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestDispatch_DefaultKey(t *testing.T) {
	d := NewDispatcher()
	called := false
	d.OnKey(func(e Event) bool {
		called = true
		return false // not consumed
	})

	// Key not matching any binding → falls through to default
	ev := Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'x'}}
	result := d.Dispatch(ev)
	if result {
		t.Error("Dispatch should return false (handler returned false)")
	}
	if !called {
		t.Error("default handler was not called")
	}
}

func TestDispatch_BindingBeforeDefault(t *testing.T) {
	d := NewDispatcher()
	order := []string{}

	d.BindKey(KeyShortcut{Rune: 'a'}, func(e Event) bool {
		order = append(order, "binding")
		return true // consumed
	})
	d.OnKey(func(e Event) bool {
		order = append(order, "default")
		return true
	})

	ev := Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'a'}}
	d.Dispatch(ev)

	if len(order) != 1 || order[0] != "binding" {
		t.Errorf("expected [binding], got %v", order)
	}
}

func TestDispatch_Mouse(t *testing.T) {
	d := NewDispatcher()
	called := false
	d.OnMouse(func(e Event) bool {
		called = true
		return true
	})

	ev := Event{Type: TypeMouse, Mouse: &term.MouseEvent{X: 5, Y: 10}}
	result := d.Dispatch(ev)
	if !result {
		t.Error("Dispatch should return true for consumed mouse event")
	}
	if !called {
		t.Error("mouse handler was not called")
	}
}

func TestDispatch_Resize(t *testing.T) {
	d := NewDispatcher()
	var gotW, gotH int
	d.OnResize(func(e Event) bool {
		gotW = e.Width
		gotH = e.Height
		return true
	})

	ev := Event{Type: TypeResize, Width: 100, Height: 30}
	d.Dispatch(ev)
	if gotW != 100 || gotH != 30 {
		t.Errorf("resize handler got W=%d H=%d, want 100x30", gotW, gotH)
	}
}

func TestDispatch_Paste(t *testing.T) {
	d := NewDispatcher()
	var gotText string
	d.OnPaste(func(e Event) bool {
		gotText = e.Paste
		return true
	})

	ev := Event{Type: TypePaste, Paste: "pasted text"}
	d.Dispatch(ev)
	if gotText != "pasted text" {
		t.Errorf("paste handler got %q, want 'pasted text'", gotText)
	}
}

func TestDispatch_MouseNoHandler(t *testing.T) {
	d := NewDispatcher()
	ev := Event{Type: TypeMouse, Mouse: &term.MouseEvent{}}
	if d.Dispatch(ev) {
		t.Error("Dispatch should return false when no mouse handler")
	}
}

func TestDispatch_ResizeNoHandler(t *testing.T) {
	d := NewDispatcher()
	ev := Event{Type: TypeResize}
	if d.Dispatch(ev) {
		t.Error("Dispatch should return false when no resize handler")
	}
}

func TestDispatch_PasteNoHandler(t *testing.T) {
	d := NewDispatcher()
	ev := Event{Type: TypePaste}
	if d.Dispatch(ev) {
		t.Error("Dispatch should return false when no paste handler")
	}
}

func TestDispatch_KeyNilKeyEvent(t *testing.T) {
	d := NewDispatcher()
	// TypeKey with nil Key pointer → should fall through to default key
	called := false
	d.OnKey(func(e Event) bool {
		called = true
		return false
	})
	ev := Event{Type: TypeKey, Key: nil}
	result := d.Dispatch(ev)
	// No bindings match nil key, default handler is called
	if !called {
		t.Error("default key handler should have been called")
	}
	if result {
		t.Error("should return false")
	}
}

// --- Shortcut helpers ---

func TestCtrl(t *testing.T) {
	s := Ctrl(term.KeyEnter)
	if s.Key != term.KeyEnter {
		t.Errorf("Key = %v, want KeyEnter", s.Key)
	}
	if s.Modifiers != term.ModCtrl {
		t.Errorf("Modifiers = %v, want ModCtrl", s.Modifiers)
	}
}

func TestCtrlRune(t *testing.T) {
	s := CtrlRune('c')
	if s.Rune != 'c' {
		t.Errorf("Rune = %q, want 'c'", s.Rune)
	}
	if s.Modifiers != term.ModCtrl {
		t.Errorf("Modifiers = %v, want ModCtrl", s.Modifiers)
	}
}

func TestAlt(t *testing.T) {
	s := Alt(term.KeyEnter)
	if s.Key != term.KeyEnter {
		t.Errorf("Key = %v", s.Key)
	}
	if s.Modifiers != term.ModAlt {
		t.Errorf("Modifiers = %v, want ModAlt", s.Modifiers)
	}
}

func TestPlain(t *testing.T) {
	s := Plain(term.KeyEnter)
	if s.Key != term.KeyEnter {
		t.Errorf("Key = %v", s.Key)
	}
	if s.Modifiers != 0 {
		t.Errorf("Modifiers = %v, want 0", s.Modifiers)
	}
}

func TestPlainRune(t *testing.T) {
	s := PlainRune('x')
	if s.Rune != 'x' {
		t.Errorf("Rune = %q", s.Rune)
	}
	if s.Modifiers != 0 {
		t.Errorf("Modifiers = %v, want 0", s.Modifiers)
	}
}

func TestDispatch_MultipleBindings(t *testing.T) {
	d := NewDispatcher()
	aCalled, bCalled := false, false

	d.BindKey(KeyShortcut{Rune: 'a'}, func(e Event) bool {
		aCalled = true
		return true
	})
	d.BindKey(KeyShortcut{Rune: 'b'}, func(e Event) bool {
		bCalled = true
		return true
	})

	// Dispatch 'a' → only a handler called
	d.Dispatch(Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'a'}})
	if !aCalled {
		t.Error("handler 'a' was not called")
	}
	if bCalled {
		t.Error("handler 'b' should not be called for 'a'")
	}
}

func TestDispatch_BindingConsumesEvent(t *testing.T) {
	d := NewDispatcher()
	bindingCalled, defaultCalled := false, false

	d.BindKey(KeyShortcut{Rune: 'q'}, func(e Event) bool {
		bindingCalled = true
		return true // consumed
	})
	d.OnKey(func(e Event) bool {
		defaultCalled = true
		return true
	})

	d.Dispatch(Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'q'}})

	if !bindingCalled {
		t.Error("binding handler was not called")
	}
	if defaultCalled {
		t.Error("default handler should not be called when binding consumed")
	}
}

// TestDispatch_BindingReturnsFalse verifies that when a binding handler
// returns false, the event is reported as unconsumed. The dispatcher does
// NOT fall through to the default handler — the binding consumes the dispatch
// slot regardless of its return value.
func TestDispatch_BindingReturnsFalse(t *testing.T) {
	d := NewDispatcher()
	bindingCalled, defaultCalled := false, false

	d.BindKey(KeyShortcut{Rune: 'q'}, func(e Event) bool {
		bindingCalled = true
		return false // NOT consumed
	})
	d.OnKey(func(e Event) bool {
		defaultCalled = true
		return false
	})

	result := d.Dispatch(Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'q'}})

	if !bindingCalled {
		t.Error("binding handler was not called")
	}
	if defaultCalled {
		t.Error("default handler should not be called — binding consumes dispatch slot")
	}
	if result {
		t.Error("Dispatch should return false (binding returned false)")
	}
}

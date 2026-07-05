package event

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

func TestP46_TypeFocusValue(t *testing.T) {
	// TypeFocus should be 5 (after TypeKey=0, TypeMouse=1, TypePaste=2, TypeResize=3, TypeQuit=4)
	if TypeFocus != 5 {
		t.Errorf("TypeFocus = %d, want 5", TypeFocus)
	}
}

func TestP46_FromTermEvent_FocusIn(t *testing.T) {
	te := term.Event{Type: term.EventFocus, Focused: true}
	e := FromTermEvent(te)
	if e.Type != TypeFocus {
		t.Errorf("expected TypeFocus, got %d", e.Type)
	}
	if !e.Focused {
		t.Error("expected Focused=true")
	}
}

func TestP46_FromTermEvent_FocusOut(t *testing.T) {
	te := term.Event{Type: term.EventFocus, Focused: false}
	e := FromTermEvent(te)
	if e.Type != TypeFocus {
		t.Errorf("expected TypeFocus, got %d", e.Type)
	}
	if e.Focused {
		t.Error("expected Focused=false")
	}
}

func TestP46_Dispatcher_OnFocus(t *testing.T) {
	d := NewDispatcher()
	called := false
	d.OnFocus(func(e Event) bool {
		called = true
		if !e.Focused {
			t.Error("expected Focused=true")
		}
		return true
	})

	result := d.Dispatch(Event{Type: TypeFocus, Focused: true})
	if !called {
		t.Error("expected focus handler to be called")
	}
	if !result {
		t.Error("expected true return")
	}
}

func TestP46_Dispatcher_OnFocus_NoHandler(t *testing.T) {
	d := NewDispatcher()
	// No focus handler registered
	result := d.Dispatch(Event{Type: TypeFocus, Focused: true})
	if result {
		t.Error("expected false when no handler registered")
	}
}

func TestP46_Dispatcher_OnFocus_FocusOut(t *testing.T) {
	d := NewDispatcher()
	var receivedFocus bool
	d.OnFocus(func(e Event) bool {
		receivedFocus = e.Focused
		return true
	})

	d.Dispatch(Event{Type: TypeFocus, Focused: false})
	if receivedFocus {
		t.Error("expected Focused=false in handler")
	}
}

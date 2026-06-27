package event

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

func TestEventType_Constants(t *testing.T) {
	// Verify event type ordering and values
	if TypeKey != 0 {
		t.Errorf("TypeKey = %d, want 0", TypeKey)
	}
	if TypeMouse != 1 {
		t.Errorf("TypeMouse = %d, want 1", TypeMouse)
	}
	if TypePaste != 2 {
		t.Errorf("TypePaste = %d, want 2", TypePaste)
	}
	if TypeResize != 3 {
		t.Errorf("TypeResize = %d, want 3", TypeResize)
	}
	if TypeQuit != 4 {
		t.Errorf("TypeQuit = %d, want 4", TypeQuit)
	}
}

func TestFromTermEvent_Key(t *testing.T) {
	ke := term.Event{
		Type: term.EventKey,
		Key:  &term.KeyEvent{Rune: 'a'},
	}
	ev := FromTermEvent(ke)
	if ev.Type != TypeKey {
		t.Errorf("Type = %v, want TypeKey", ev.Type)
	}
	if ev.Key == nil || ev.Key.Rune != 'a' {
		t.Errorf("Key.Rune = %q, want 'a'", ev.Key)
	}
}

func TestFromTermEvent_Mouse(t *testing.T) {
	me := term.Event{
		Type: term.EventMouse,
		Mouse: &term.MouseEvent{X: 10, Y: 5},
	}
	ev := FromTermEvent(me)
	if ev.Type != TypeMouse {
		t.Errorf("Type = %v, want TypeMouse", ev.Type)
	}
	if ev.Mouse == nil || ev.Mouse.X != 10 || ev.Mouse.Y != 5 {
		t.Errorf("Mouse = %+v, want X=10 Y=5", ev.Mouse)
	}
}

func TestFromTermEvent_Paste(t *testing.T) {
	pe := term.Event{
		Type:  term.EventPaste,
		Paste: "hello world",
	}
	ev := FromTermEvent(pe)
	if ev.Type != TypePaste {
		t.Errorf("Type = %v, want TypePaste", ev.Type)
	}
	if ev.Paste != "hello world" {
		t.Errorf("Paste = %q, want 'hello world'", ev.Paste)
	}
}

func TestFromTermEvent_Resize(t *testing.T) {
	re := term.Event{
		Type:   term.EventResize,
		Width:  120,
		Height: 40,
	}
	ev := FromTermEvent(re)
	if ev.Type != TypeResize {
		t.Errorf("Type = %v, want TypeResize", ev.Type)
	}
	if ev.Width != 120 || ev.Height != 40 {
		t.Errorf("W=%d H=%d, want 120x40", ev.Width, ev.Height)
	}
}

func TestFromTermEvent_Unknown(t *testing.T) {
	// An event type not handled should return zero-value Event
	ev := FromTermEvent(term.Event{Type: 99})
	if ev.Type != 0 {
		t.Errorf("Unknown event: Type = %d, want 0", ev.Type)
	}
}

// --- KeyShortcut ---

func TestKeyShortcut_MatchByKey(t *testing.T) {
	s := KeyShortcut{Key: term.KeyEnter}
	if !s.Match(&term.KeyEvent{Key: term.KeyEnter}) {
		t.Error("Match(KeyEnter) should be true")
	}
	if s.Match(&term.KeyEvent{Key: term.KeyEscape}) {
		t.Error("Match(KeyEscape) should be false")
	}
}

func TestKeyShortcut_MatchByRune(t *testing.T) {
	s := KeyShortcut{Rune: 'q'}
	if !s.Match(&term.KeyEvent{Rune: 'q'}) {
		t.Error("Match('q') should be true")
	}
	if s.Match(&term.KeyEvent{Rune: 'Q'}) {
		t.Error("Match('Q') should be false (case-sensitive)")
	}
}

func TestKeyShortcut_MatchWithModifiers(t *testing.T) {
	s := KeyShortcut{Key: term.KeyEnter, Modifiers: term.ModCtrl}
	if !s.Match(&term.KeyEvent{Key: term.KeyEnter, Modifiers: term.ModCtrl}) {
		t.Error("Match(Ctrl+Enter) should be true")
	}
	if s.Match(&term.KeyEvent{Key: term.KeyEnter}) {
		t.Error("Match(Enter without Ctrl) should be false")
	}
}

func TestKeyShortcut_MatchNilKey(t *testing.T) {
	s := KeyShortcut{Key: term.KeyEnter}
	if s.Match(nil) {
		t.Error("Match(nil) should be false")
	}
}

func TestKeyShortcut_MatchKeyUnknown(t *testing.T) {
	// When shortcut.Key is KeyUnknown and Rune is set, match on rune only
	s := KeyShortcut{Rune: 'x'}
	if !s.Match(&term.KeyEvent{Rune: 'x'}) {
		t.Error("Match(rune 'x') should be true")
	}
}

func TestKeyShortcut_NoMatch(t *testing.T) {
	s := KeyShortcut{Key: term.KeyEnter, Modifiers: term.ModCtrl}
	// Different key, same modifier
	if s.Match(&term.KeyEvent{Key: term.KeyTab, Modifiers: term.ModCtrl}) {
		t.Error("Should not match different key")
	}
}

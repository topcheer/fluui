package event

import (
	"github.com/topcheer/fluui/internal/term"
)

// EventType identifies the kind of event.
type EventType uint8

const (
	// Terminal events
	TypeKey    EventType = iota
	TypeMouse
	TypePaste
	TypeResize
	TypeQuit
	TypeFocus
)

// Event wraps a terminal event for the internal event system.
type Event struct {
	Type    EventType
	Key     *term.KeyEvent
	Mouse   *term.MouseEvent
	Paste   string
	Width   int
	Height  int
	Focused bool // for TypeFocus: true=gained, false=lost
}

// Handler processes an event and returns whether it was consumed.
type Handler func(Event) bool

// FromTermEvent converts a term.Event to an internal Event.
func FromTermEvent(te term.Event) Event {
	switch te.Type {
	case term.EventKey:
		return Event{Type: TypeKey, Key: te.Key}
	case term.EventMouse:
		return Event{Type: TypeMouse, Mouse: te.Mouse}
	case term.EventPaste:
		return Event{Type: TypePaste, Paste: te.Paste}
	case term.EventResize:
		return Event{Type: TypeResize, Width: te.Width, Height: te.Height}
	case term.EventFocus:
		return Event{Type: TypeFocus, Focused: te.Focused}
	}
	return Event{}
}

// KeyShortcut creates a quick key check.
type KeyShortcut struct {
	Key       term.KeyCode
	Modifiers term.ModMask
	Rune      rune
}

// Match checks if a key event matches this shortcut.
func (s KeyShortcut) Match(k *term.KeyEvent) bool {
	if k == nil {
		return false
	}
	if s.Key != term.KeyUnknown && k.Key == s.Key {
		return k.Modifiers == s.Modifiers
	}
	if s.Rune != 0 && k.Rune == s.Rune {
		return k.Modifiers == s.Modifiers
	}
	return false
}

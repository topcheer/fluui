package event

import (
	"github.com/topcheer/fluui/internal/term"
)

// QuitHandler is set by the App to signal shutdown.
type QuitHandler func()

// KeyBinding binds a key combination to a handler.
type KeyBinding struct {
	Shortcut KeyShortcut
	Handler  Handler
}

// Dispatcher manages event routing and key bindings.
type Dispatcher struct {
	bindings   []KeyBinding
	mouseHandler Handler
	resizeHandler Handler
	pasteHandler  Handler
	focusHandler  Handler
	defaultKey  Handler
}

// NewDispatcher creates a new event dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

// BindKey registers a keyboard shortcut.
func (d *Dispatcher) BindKey(shortcut KeyShortcut, handler Handler) {
	d.bindings = append(d.bindings, KeyBinding{Shortcut: shortcut, Handler: handler})
}

// OnMouse sets the mouse event handler.
func (d *Dispatcher) OnMouse(h Handler) { d.mouseHandler = h }

// OnResize sets the resize event handler.
func (d *Dispatcher) OnResize(h Handler) { d.resizeHandler = h }

// OnPaste sets the paste event handler.
func (d *Dispatcher) OnPaste(h Handler) { d.pasteHandler = h }

// OnKey sets a default key handler (for unbound keys).
func (d *Dispatcher) OnKey(h Handler) { d.defaultKey = h }

// OnFocus sets the focus event handler.
// The handler is called when the terminal gains or loses focus
// (requires focus tracking support, enabled automatically by term.Open).
func (d *Dispatcher) OnFocus(h Handler) { d.focusHandler = h }

// Dispatch routes an event to the appropriate handler.
func (d *Dispatcher) Dispatch(e Event) bool {
	switch e.Type {
	case TypeKey:
		if e.Key != nil {
			// Check bindings first
			for _, b := range d.bindings {
				if b.Shortcut.Match(e.Key) {
					return b.Handler(e)
				}
			}
		}
		if d.defaultKey != nil {
			return d.defaultKey(e)
		}

	case TypeMouse:
		if d.mouseHandler != nil {
			return d.mouseHandler(e)
		}

	case TypeResize:
		if d.resizeHandler != nil {
			return d.resizeHandler(e)
		}

	case TypePaste:
		if d.pasteHandler != nil {
			return d.pasteHandler(e)
		}

	case TypeFocus:
		if d.focusHandler != nil {
			return d.focusHandler(e)
		}
	}
	return false
}

// Key helper constructors.
func Ctrl(key term.KeyCode) KeyShortcut {
	return KeyShortcut{Key: key, Modifiers: term.ModCtrl}
}

// CtrlRune creates a KeyShortcut for Ctrl+<rune>.
func CtrlRune(r rune) KeyShortcut {
	return KeyShortcut{Rune: r, Modifiers: term.ModCtrl}
}

// Alt creates a KeyShortcut for Alt+<key>.
func Alt(key term.KeyCode) KeyShortcut {
	return KeyShortcut{Key: key, Modifiers: term.ModAlt}
}

// Plain creates a KeyShortcut for a plain key (no modifiers).
func Plain(key term.KeyCode) KeyShortcut {
	return KeyShortcut{Key: key}
}

// PlainRune creates a KeyShortcut for a plain rune (no modifiers).
func PlainRune(r rune) KeyShortcut {
	return KeyShortcut{Rune: r}
}

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

// FocusableComponent is a component that can receive keyboard focus and
// handle events directly. When focused, it gets first chance to consume
// key/mouse events before global bindings.
type FocusableComponent interface {
	// HandleEvent processes an event. Returns true if consumed (stops bubbling),
	// false if the event should bubble up to global handlers.
	HandleEvent(e Event) bool
}

// Dispatcher manages event routing and key bindings.
// Dispatch order for key events: focused component → key bindings → default handler.
type Dispatcher struct {
	bindings      []KeyBinding
	mouseHandler  Handler
	resizeHandler Handler
	pasteHandler  Handler
	focusHandler  Handler
	customHandler Handler
	defaultKey    Handler

	// Focus-aware event bubbling
	focused    FocusableComponent // currently focused component (nil = no focus)
	focusChain []FocusableComponent // ordered list: [0] = topmost focus target
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

// OnCustom sets the handler for Cmd/Msg results (TypeCustom events).
func (d *Dispatcher) OnCustom(h Handler) { d.customHandler = h }

// Dispatch routes an event to the appropriate handler.
// For key events, the dispatch order is:
// 1. Focused component (if set, via focus chain top-to-bottom)
// 2. Key bindings
// 3. Default key handler
func (d *Dispatcher) Dispatch(e Event) bool {
	switch e.Type {
	case TypeKey:
		if e.Key != nil {
			// 1. Try focused component(s) first
			for _, fc := range d.focusChain {
				if fc != nil && fc.HandleEvent(e) {
					return true // consumed by focused component
				}
			}
			if d.focused != nil && d.focused.HandleEvent(e) {
				return true
			}
			// 2. Check global key bindings
			for _, b := range d.bindings {
				if b.Shortcut.Match(e.Key) {
					return b.Handler(e)
				}
			}
		}
		// 3. Default key handler
		if d.defaultKey != nil {
			return d.defaultKey(e)
		}

	case TypeMouse:
		// Mouse events also go through focus chain first
		for _, fc := range d.focusChain {
			if fc != nil && fc.HandleEvent(e) {
				return true
			}
		}
		if d.focused != nil && d.focused.HandleEvent(e) {
			return true
		}
		if d.mouseHandler != nil {
			return d.mouseHandler(e)
		}

	case TypeResize:
		if d.resizeHandler != nil {
			return d.resizeHandler(e)
		}

	case TypePaste:
		// Paste events go to focused component first
		for _, fc := range d.focusChain {
			if fc != nil && fc.HandleEvent(e) {
				return true
			}
		}
		if d.focused != nil && d.focused.HandleEvent(e) {
			return true
		}
		if d.pasteHandler != nil {
			return d.pasteHandler(e)
		}

	case TypeFocus:
		if d.focusHandler != nil {
			return d.focusHandler(e)
		}

	case TypeCustom:
		if d.customHandler != nil {
			return d.customHandler(e)
		}
	}
	return false
}

// --- Focus Management ---

// SetFocused sets the currently focused component.
// The focused component gets first chance to consume key, mouse, and paste events.
// Set to nil to clear focus.
func (d *Dispatcher) SetFocused(fc FocusableComponent) {
	d.focused = fc
}

// Focused returns the currently focused component, or nil.
func (d *Dispatcher) Focused() FocusableComponent {
	return d.focused
}

// PushFocus adds a component to the top of the focus chain.
// Components higher in the chain get events first.
// This enables nested focus (e.g., modal dialog over input field).
func (d *Dispatcher) PushFocus(fc FocusableComponent) {
	d.focusChain = append(d.focusChain, fc)
}

// PopFocus removes and returns the topmost component from the focus chain.
// Returns nil if the chain is empty.
func (d *Dispatcher) PopFocus() FocusableComponent {
	n := len(d.focusChain)
	if n == 0 {
		return nil
	}
	fc := d.focusChain[n-1]
	d.focusChain = d.focusChain[:n-1]
	return fc
}

// FocusChain returns the current focus chain (read-only copy).
func (d *Dispatcher) FocusChain() []FocusableComponent {
	result := make([]FocusableComponent, len(d.focusChain))
	copy(result, d.focusChain)
	return result
}

// ClearFocus removes all focused components.
func (d *Dispatcher) ClearFocus() {
	d.focused = nil
	d.focusChain = d.focusChain[:0]
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

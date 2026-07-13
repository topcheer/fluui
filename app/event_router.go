package app

import (
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/term"
)

// ─── EventRouter: Declarative Key Routing ───
//
// EventRouter integrates KeybindingManager + PanelManager into a single
// OnKey handler that replaces the monolithic handleKeyPress function.
//
// Dispatch order (matches ggcode's desired priority):
//
//	1. Active Panel (PanelManager.HandleKey)     — panel gets first chance
//	2. Context Keybindings (KeybindingManager)   — declarative shortcuts
//	3. Global Keybindings (KeybindingManager)    — app-wide shortcuts
//	4. Fallback handler                          — last resort
//
// This replaces the 50+ if-branch pattern:
//
//	// Before:
//	if m.fileBrowser != nil { return m.handleFileBrowserKey(msg) }
//	if m.previewPanel != nil { return m.handlePreviewKey(msg) }
//	if msg.String() == "ctrl+r" { m.sidebarVisible = !m.sidebarVisible }
//	if msg.String() == "ctrl+l" { m.handleClearChat() }
//	// ... 46 more branches
//
//	// After:
//	router.RegisterGlobal("toggle-sidebar", "ctrl+r", "Toggle sidebar", func() bool { ... })
//	router.RegisterGlobal("clear-chat", "ctrl+l", "Clear chat", func() bool { ... })
//	router.RegisterContext("editor", "save", "ctrl+s", "Save", func() bool { ... })
//	app.OnKey(router.HandleKey)  // that's it
type EventRouter struct {
	panels *PanelManager
	keys   *component.KeybindingManager

	// Fallback handler called when no panel or keybinding matches.
	// If nil, unmatched keys are silently consumed.
	fallback func(ev *term.KeyEvent) bool
}

// NewEventRouter creates a router connecting a PanelManager and
// KeybindingManager. Both must be non-nil.
func NewEventRouter(panels *PanelManager, keys *component.KeybindingManager) *EventRouter {
	return &EventRouter{
		panels: panels,
		keys:   keys,
	}
}

// SetFallback sets a handler for keys that no panel or keybinding consumes.
// If not set, unmatched keys are consumed silently (return true).
func (r *EventRouter) SetFallback(fn func(ev *term.KeyEvent) bool) {
	r.fallback = fn
}

// RegisterGlobal adds an app-wide keybinding.
// These are always active regardless of which panel is open.
func (r *EventRouter) RegisterGlobal(command, keys, help string, handler func() bool) error {
	return r.keys.RegisterIn("global", command, keys, help, handler)
}

// RegisterContext adds a context-scoped keybinding.
// Active only when the context is pushed (PushContext/PopContext).
func (r *EventRouter) RegisterContext(context, command, keys, help string, handler func() bool) error {
	return r.keys.RegisterIn(context, command, keys, help, handler)
}

// PushContext activates a context scope (e.g., "editor", "search", "modal").
// Context bindings take priority over global bindings.
func (r *EventRouter) PushContext(ctx string) {
	r.keys.PushContext(ctx)
}

// PopContext deactivates the current context scope.
func (r *EventRouter) PopContext() string {
	return r.keys.PopContext()
}

// ActiveContext returns the current context name.
func (r *EventRouter) ActiveContext() string {
	return r.keys.ActiveContext()
}

// HandleKey is the single OnKey handler that replaces handleKeyPress.
// Wire it with: app.OnKey(router.HandleKey)
func (r *EventRouter) HandleKey(ev *term.KeyEvent) bool {
	// 1. Active panel gets first chance
	if r.panels.HandleKey(ev) {
		return true
	}

	// 2. Check keybindings (context first, then global — handled internally)
	if r.keys.HandleKey(ev) {
		return true
	}

	// 3. Fallback handler
	if r.fallback != nil {
		return r.fallback(ev)
	}

	// 4. Consume silently — prevents terminal from echoing
	return true
}

// HandleMouse routes mouse events to the active panel.
// Wire it with: app.OnMouse(router.HandleMouse)
func (r *EventRouter) HandleMouse(x, y int, action string) bool {
	return r.panels.HandleMouse(x, y, action)
}

// HelpText returns formatted help for all active keybindings.
func (r *EventRouter) HelpText() string {
	return r.keys.HelpText()
}

// PushPanel opens a new panel and returns.
// Convenience for: panels.Push(p)
func (r *EventRouter) PushPanel(p Panel) {
	r.panels.Push(p)
}

// PopPanel closes the current panel.
// Convenience for: panels.Pop()
func (r *EventRouter) PopPanel() Panel {
	return r.panels.Pop()
}
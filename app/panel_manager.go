package app

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── PanelManager: Panel Stack Architecture ───
//
// PanelManager replaces the monolithic handleKeyPress pattern:
//
//	// Before: 50+ if-branches
//	func (m Model) handleKeyPress(msg tea.KeyPressMsg) {
//	    if m.fileBrowser != nil { return m.handleFileBrowserKey(msg) }
//	    if m.previewPanel != nil { return m.handlePreviewKey(msg) }
//	    if m.modelPanel != nil { return m.handleModelPanelKey(msg) }
//	    if m.providerPanel != nil { return m.handleProviderPanelKey(msg) }
//	    if m.qqPanel != nil { return m.handleQQPanelKey(msg) }
//	    if m.tgPanel != nil { return m.handleTGPanelKey(msg) }
//	    // ... 44 more branches
//	}
//
//	// After: one-line routing
//	panelMgr.HandleKey(ev)  // routes to active panel automatically
//
// Each Panel owns its state. No shared 236-field Model struct.
// Esc closes the topmost panel (standard modal behavior).

// Panel is the interface for a screen-level UI component that owns
// its own state and event handling. This is the unit of composition
// in fluui's component-declarative architecture.
//
// A Panel is NOT a Component — it's higher level. A Panel manages
// a full screen area and handles all input while active. Internally
// it may compose multiple Components.
type Panel interface {
	// ID returns a unique identifier for this panel type.
	ID() string

	// Title returns the panel title (shown in sidebar/breadcrumb).
	Title() string

	// HandleKey processes a key event. Returns true if consumed.
	// The active panel gets first chance at all key events.
	HandleKey(ev *term.KeyEvent) bool

	// HandleMouse processes a mouse event. Returns true if consumed.
	HandleMouse(x, y int, action string) bool

	// Paint renders the panel content into the buffer.
	Paint(buf *buffer.Buffer, w, h int)

	// OnShow is called when this panel becomes the active (topmost) panel.
	OnShow()

	// OnHide is called when this panel is no longer the active panel.
	OnHide()
}

// BasePanel provides default no-op implementations for optional Panel methods.
// Embed this to avoid implementing all 7 methods.
type BasePanel struct{}

func (BasePanel) HandleMouse(x, y int, action string) bool { return false }
func (BasePanel) OnShow()                                  {}
func (BasePanel) OnHide()                                  {}

// PanelManager manages a stack of Panels. Only the topmost (active)
// panel receives keyboard and mouse events. Esc closes the topmost
// panel (unless it's the root panel).
//
// Thread-safe.
type PanelManager struct {
	mu     sync.RWMutex
	panels []Panel
	root   Panel // bottom panel — never popped

	onChange func() // callback when panel stack changes (for MarkDirty)
}

// NewPanelManager creates a PanelManager with a root panel.
// The root panel is always at the bottom and cannot be popped.
func NewPanelManager(root Panel) *PanelManager {
	return &PanelManager{
		panels: []Panel{root},
		root:   root,
	}
}

// SetOnChange sets a callback invoked when the panel stack changes
// (push/pop/replace). Use this to trigger a re-render.
func (pm *PanelManager) SetOnChange(fn func()) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.onChange = fn
}

func (pm *PanelManager) notifyChange() {
	if pm.onChange != nil {
		pm.onChange()
	}
}

// Push opens a new panel on top of the stack. The new panel becomes
// active and receives all keyboard/mouse events until popped.
// OnShow is called on the new panel; OnHide on the previously active panel.
func (pm *PanelManager) Push(p Panel) {
	pm.mu.Lock()
	if len(pm.panels) > 0 {
		pm.panels[len(pm.panels)-1].OnHide()
	}
	pm.panels = append(pm.panels, p)
	pm.mu.Unlock()
	p.OnShow()
	pm.notifyChange()
}

// Pop closes the topmost panel. Returns the popped panel, or nil
// if only the root panel remains (root cannot be popped).
// OnHide is called on the popped panel; OnShow on the new topmost.
func (pm *PanelManager) Pop() Panel {
	pm.mu.Lock()
	if len(pm.panels) <= 1 {
		pm.mu.Unlock()
		return nil
	}
	popped := pm.panels[len(pm.panels)-1]
	pm.panels = pm.panels[:len(pm.panels)-1]
	newTop := pm.panels[len(pm.panels)-1]
	pm.mu.Unlock()

	popped.OnHide()
	newTop.OnShow()
	pm.notifyChange()
	return popped
}

// Replace swaps the topmost panel with a new one. If only the root
// remains, the new panel is pushed on top. Returns the old topmost panel.
func (pm *PanelManager) Replace(p Panel) Panel {
	pm.mu.Lock()
	if len(pm.panels) <= 1 {
		pm.panels = append(pm.panels, p)
		pm.mu.Unlock()
		p.OnShow()
		pm.notifyChange()
		return nil
	}
	old := pm.panels[len(pm.panels)-1]
	pm.panels[len(pm.panels)-1] = p
	pm.mu.Unlock()

	old.OnHide()
	p.OnShow()
	pm.notifyChange()
	return old
}

// Active returns the topmost (active) panel, or nil if empty.
func (pm *PanelManager) Active() Panel {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	if len(pm.panels) == 0 {
		return nil
	}
	return pm.panels[len(pm.panels)-1]
}

// Root returns the root (bottom) panel.
func (pm *PanelManager) Root() Panel {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.root
}

// Depth returns the number of panels in the stack (including root).
func (pm *PanelManager) Depth() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.panels)
}

// IsRoot returns true if only the root panel is in the stack.
func (pm *PanelManager) IsRoot() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.panels) == 1
}

// FindByID returns the first panel with the given ID (searching top-down).
func (pm *PanelManager) FindByID(id string) Panel {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	for i := len(pm.panels) - 1; i >= 0; i-- {
		if pm.panels[i].ID() == id {
			return pm.panels[i]
		}
	}
	return nil
}

// HandleKey routes a key event to the active panel.
// If the active panel doesn't consume it and it's Escape, the panel is popped.
// Returns true if the event was consumed.
func (pm *PanelManager) HandleKey(ev *term.KeyEvent) bool {
	pm.mu.RLock()
	active := pm.panels[len(pm.panels)-1]
	canPop := len(pm.panels) > 1
	pm.mu.RUnlock()

	// Active panel gets first chance
	if active.HandleKey(ev) {
		return true
	}

	// Esc closes the topmost panel (if not root)
	if canPop && ev.Key == term.KeyEscape {
		pm.Pop()
		return true
	}

	return false
}

// HandleMouse routes a mouse event to the active panel.
func (pm *PanelManager) HandleMouse(x, y int, action string) bool {
	pm.mu.RLock()
	active := pm.panels[len(pm.panels)-1]
	pm.mu.RUnlock()
	return active.HandleMouse(x, y, action)
}

// Paint renders the active panel into the buffer.
func (pm *PanelManager) Paint(buf *buffer.Buffer, w, h int) {
	pm.mu.RLock()
	active := pm.panels[len(pm.panels)-1]
	pm.mu.RUnlock()
	active.Paint(buf, w, h)
}

// CloseAll pops all panels except the root.
func (pm *PanelManager) CloseAll() {
	pm.mu.Lock()
	if len(pm.panels) <= 1 {
		pm.mu.Unlock()
		return
	}
	// Notify all non-root panels
	for i := len(pm.panels) - 1; i >= 1; i-- {
		pm.panels[i].OnHide()
	}
	pm.panels = pm.panels[:1]
	pm.mu.Unlock()

	pm.root.OnShow()
	pm.notifyChange()
}

// Panels returns a copy of the panel stack (bottom-to-top).
func (pm *PanelManager) Panels() []Panel {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	result := make([]Panel, len(pm.panels))
	copy(result, pm.panels)
	return result
}
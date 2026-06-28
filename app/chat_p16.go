package app

import (
	"fmt"
	"time"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// --- StatusBar Integration ---

// SetStatusBar attaches a StatusBar component to this ChatApp.
// When set, the status bar is painted at the very bottom of the screen,
// below the input line area.
func (a *ChatApp) SetStatusBar(sb *component.StatusBar) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.statusBar = sb
	if sb != nil {
		a.statusBarHeight = 1
	} else {
		a.statusBarHeight = 0
	}
}

// StatusBar returns the attached StatusBar, or nil if none.
func (a *ChatApp) StatusBar() *component.StatusBar {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.statusBar
}

// SetModel sets the AI model name in the status bar (if attached).
func (a *ChatApp) SetModel(model string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.statusBar != nil {
		a.statusBar.SetModel(model)
	}
}

// SetTokenRate sets the token generation rate in the status bar.
func (a *ChatApp) SetTokenRate(rate int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.statusBar != nil {
		a.statusBar.SetTokenRate(rate)
	}
}

// SetContextWindow sets the context window usage in the status bar.
func (a *ChatApp) SetContextWindow(used, total int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.statusBar != nil {
		a.statusBar.SetContextWindow(used, total)
	}
}

// UpdateClock refreshes the clock display in the status bar.
func (a *ChatApp) UpdateClock() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.statusBar != nil {
		a.statusBar.SetClock(time.Now())
	}
}

// --- TabBar Integration ---

// SetTabBar attaches a TabBar component to this ChatApp.
// When set, the tab bar is painted at the top of the screen,
// above the scroll view content.
func (a *ChatApp) SetTabBar(tb *component.TabBar) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tabBar = tb
	if tb != nil {
		a.tabBarHeight = 1
	} else {
		a.tabBarHeight = 0
	}
}

// TabBar returns the attached TabBar, or nil if none.
func (a *ChatApp) TabBar() *component.TabBar {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.tabBar
}

// AddSession creates a new chat session tab.
// Returns the session index, or -1 if no tab bar is attached.
func (a *ChatApp) AddSession(name string) int {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.tabBar == nil {
		return -1
	}
	return a.tabBar.AddTab(fmt.Sprintf("session-%d", a.tabBar.TabCount()), name)
}

// ActiveSession returns the index of the active session tab.
func (a *ChatApp) ActiveSession() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.tabBar == nil {
		return 0
	}
	return a.tabBar.ActiveIndex()
}

// SwitchSession switches to the session at the given index.
func (a *ChatApp) SwitchSession(idx int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.tabBar != nil {
		a.tabBar.SetActive(idx)
	}
}

// NextSession switches to the next session tab (wraps around).
func (a *ChatApp) NextSession() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.tabBar != nil {
		a.tabBar.NextTab()
	}
}

// PrevSession switches to the previous session tab (wraps around).
func (a *ChatApp) PrevSession() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.tabBar != nil {
		a.tabBar.PrevTab()
	}
}

// CloseSession closes the active session tab.
func (a *ChatApp) CloseSession() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.tabBar != nil {
		a.tabBar.CloseActive()
	}
}

// SessionCount returns the number of open session tabs.
func (a *ChatApp) SessionCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.tabBar == nil {
		return 0
	}
	return a.tabBar.TabCount()
}

// --- SelectionManager Integration ---

// SetSelectionManager attaches a SelectionManager to this ChatApp.
// When set, mouse drag events are routed to the selection manager
// for text selection and copy operations.
func (a *ChatApp) SetSelectionManager(sm *SelectionManager) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.selectionMgr = sm
}

// SelectionManager returns the attached SelectionManager, or nil if none.
func (a *ChatApp) SelectionManager() *SelectionManager {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.selectionMgr
}

// HasSelection returns true if there is an active text selection.
func (a *ChatApp) HasSelection() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.selectionMgr == nil {
		return false
	}
	return a.selectionMgr.HasSelection()
}

// ClearSelection clears the current text selection.
func (a *ChatApp) ClearSelection() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.selectionMgr != nil {
		a.selectionMgr.Clear()
	}
}

// --- Enhanced HandleMouse with P15 components ---

// HandleMouseP16 processes mouse events with P15 component support.
// It routes events to: overlays → tab bar → selection → scroll → custom.
func (a *ChatApp) HandleMouseP16(mouse *term.MouseEvent) bool {
	if mouse == nil {
		return false
	}
	a.mu.Lock()
	overlays := a.overlays
	tabBar := a.tabBar
	onMouse := a.onMouse
	a.mu.Unlock()

	// Route to overlays first
	if overlays.HandleMouse(mouse.X, mouse.Y) {
		return true
	}

	// Route to tab bar (click on tabs)
	if tabBar != nil {
		idx := tabBar.HitTest(mouse.X, mouse.Y)
		if idx >= 0 {
			if mouse.Action == term.MouseDown && mouse.Button == term.MouseLeft {
				a.mu.Lock()
				tabBar.SetActive(idx)
				a.mu.Unlock()
				return true
			}
		}
		// Check close button click
		if mouse.Action == term.MouseDown && mouse.Button == term.MouseLeft {
			tabIdx, ok := tabBar.IsCloseButton(mouse.X, mouse.Y)
			if ok {
				a.mu.Lock()
				tabBar.RemoveTab(fmt.Sprintf("session-%d", tabIdx))
				a.mu.Unlock()
				return true
			}
		}
	}

	// Selection manager (drag selection) — delegate to its own HandleMouse
	a.mu.Lock()
	selMgr := a.selectionMgr
	a.mu.Unlock()
	if selMgr != nil && selMgr.HandleMouse(mouse) {
		return true
	}

	// Scroll wheel
	if mouse.Action == term.MouseWheel {
		switch mouse.Button {
		case term.MouseWheelUp:
			a.ScrollUp()
			return true
		case term.MouseWheelDown:
			a.ScrollDown()
			return true
		}
	}

	// Custom handler
	if onMouse != nil {
		onMouse(mouse)
	}
	return true
}

// --- Enhanced Render with P15 components ---

// renderP16 renders the P15 components (tab bar at top, status bar at bottom).
// This is called from within Render after the main content but before overlays.
func (a *ChatApp) renderP16(buf *buffer.Buffer, w, h int) {
	// Tab bar at top
	if a.tabBar != nil && a.tabBarHeight > 0 {
		a.tabBar.SetBounds(component.Rect{X: 0, Y: 0, W: w, H: 1})
		a.tabBar.Paint(buf)
	}

	// Status bar at bottom
	if a.statusBar != nil && a.statusBarHeight > 0 {
		statusY := h - a.statusBarHeight
		a.statusBar.SetBounds(component.Rect{X: 0, Y: statusY, W: w, H: 1})
		a.statusBar.Paint(buf)
	}

	// Apply selection highlights
	if a.selectionMgr != nil && a.selectionMgr.HasSelection() {
		a.selectionMgr.ApplyHighlight(buf)
	}
}

// --- Session Management Helpers ---

// SessionInfo holds metadata about a chat session.
type SessionInfo struct {
	ID    string
	Name  string
	Index int
}

// Sessions returns a list of all session tabs.
func (a *ChatApp) Sessions() []SessionInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.tabBar == nil {
		return nil
	}
	tabs := a.tabBar.Tabs()
	result := make([]SessionInfo, len(tabs))
	for i, tab := range tabs {
		result[i] = SessionInfo{
			ID:    tab.ID,
			Name:  tab.Title,
			Index: i,
		}
	}
	return result
}

// ActiveSessionName returns the title of the active session.
func (a *ChatApp) ActiveSessionName() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.tabBar == nil {
		return ""
	}
	tab := a.tabBar.TabAt(a.tabBar.ActiveIndex())
	if tab == nil {
		return ""
	}
	return tab.Title
}

// --- Enhanced Key Handling for P15 ---

// handleP16Keys processes P15-specific key events.
// Returns true if the event was consumed.
func (a *ChatApp) handleP16Keys(key *term.KeyEvent) bool {
	a.mu.Lock()
	tabBar := a.tabBar
	a.mu.Unlock()

	if tabBar == nil {
		return false
	}

	// Alt+[/]: prev/next session
	// Alt+W: close session
	// Alt+1-9: switch to session N
	if key.Modifiers&term.ModAlt != 0 {
		switch key.Rune {
		case ']':
			a.NextSession()
			return true
		case '[':
			a.PrevSession()
			return true
		case 'w':
			a.CloseSession()
			return true
		}
		if key.Rune >= '1' && key.Rune <= '9' {
			idx := int(key.Rune - '1')
			a.SwitchSession(idx)
			return true
		}
	}

	return false
}

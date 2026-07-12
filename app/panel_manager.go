package app

import (
	"sync"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// Panel is a component that can be registered with the PanelManager.
// Each panel occupies the main content area and is switched via tab-like navigation.
type Panel interface {
	component.Component
	ID() string
	Title() string
	Icon() string
	OnActivate()
	OnDeactivate()
}

// PanelManager manages multiple panels with a tab bar at the bottom.
// Only the active panel is rendered; others are hidden.
// Thread-safe via sync.RWMutex.
type PanelManager struct {
	component.BaseComponent

	mu        sync.RWMutex
	panels    map[string]Panel
	order     []string // panel IDs in registration order
	activeID  string
	visible   []string // visible tab IDs (for scroll arrows)
	scrollIdx int      // first visible tab index
	unread    map[string]int
	current   *theme.Theme
}

// NewPanelManager creates an empty panel manager.
func NewPanelManager() *PanelManager {
	return &PanelManager{
		panels: make(map[string]Panel),
		unread: make(map[string]int),
	}
}

// RegisterPanel registers a panel at the given index (appends if index < 0 or > len).
func (pm *PanelManager) RegisterPanel(p Panel, index int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	id := p.ID()
	if _, exists := pm.panels[id]; exists {
		return // already registered
	}

	if index < 0 || index >= len(pm.order) {
		pm.order = append(pm.order, id)
	} else {
		pm.order = append(pm.order[:index], append([]string{id}, pm.order[index:]...)...)
	}
	pm.panels[id] = p

	// First panel becomes active
	if pm.activeID == "" {
		pm.activeID = id
		p.OnActivate()
	}
}

// UnregisterPanel removes a panel by ID.
func (pm *PanelManager) UnregisterPanel(id string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.panels[id]; !exists {
		return
	}

	wasActive := pm.activeID == id
	delete(pm.panels, id)
	delete(pm.unread, id)

	// Remove from order
	for i, oid := range pm.order {
		if oid == id {
			pm.order = append(pm.order[:i], pm.order[i+1:]...)
			break
		}
	}

	// If we removed the active panel, switch to the first remaining
	if wasActive && len(pm.order) > 0 {
		pm.activeID = pm.order[0]
		if p, ok := pm.panels[pm.activeID]; ok {
			pm.mu.Unlock()
			p.OnActivate()
			pm.mu.Lock()
		}
	} else if wasActive {
		pm.activeID = ""
	}
}

// SwitchTo activates the panel with the given ID. Returns false if not found.
func (pm *PanelManager) SwitchTo(id string) bool {
	pm.mu.Lock()

	newPanel, ok := pm.panels[id]
	if !ok {
		pm.mu.Unlock()
		return false
	}

	oldID := pm.activeID
	if oldID == id {
		pm.mu.Unlock()
		return true
	}

	var oldPanel Panel
	if oldID != "" {
		oldPanel = pm.panels[oldID]
	}
	pm.activeID = id
	pm.unread[id] = 0 // clear unread on view

	pm.mu.Unlock()

	// Callbacks outside lock to avoid deadlock
	if oldPanel != nil {
		oldPanel.OnDeactivate()
	}
	newPanel.OnActivate()
	return true
}

// Next switches to the next panel (wraps around).
func (pm *PanelManager) Next() {
	pm.mu.RLock()
	if len(pm.order) == 0 {
		pm.mu.RUnlock()
		return
	}
	currentIdx := 0
	for i, id := range pm.order {
		if id == pm.activeID {
			currentIdx = i
			break
		}
	}
	nextID := pm.order[(currentIdx+1)%len(pm.order)]
	pm.mu.RUnlock()
	pm.SwitchTo(nextID)
}

// Prev switches to the previous panel (wraps around).
func (pm *PanelManager) Prev() {
	pm.mu.RLock()
	if len(pm.order) == 0 {
		pm.mu.RUnlock()
		return
	}
	currentIdx := 0
	for i, id := range pm.order {
		if id == pm.activeID {
			currentIdx = i
			break
		}
	}
	prevIdx := currentIdx - 1
	if prevIdx < 0 {
		prevIdx = len(pm.order) - 1
	}
	prevID := pm.order[prevIdx]
	pm.mu.RUnlock()
	pm.SwitchTo(prevID)
}

// SwitchByIndex switches to the panel at 1-based index (1-9).
func (pm *PanelManager) SwitchByIndex(n int) bool {
	pm.mu.RLock()
	if n < 1 || n > len(pm.order) {
		pm.mu.RUnlock()
		return false
	}
	id := pm.order[n-1]
	pm.mu.RUnlock()
	return pm.SwitchTo(id)
}

// ActivePanel returns the currently active panel, or nil if none.
func (pm *PanelManager) ActivePanel() Panel {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.panels[pm.activeID]
}

// ActiveID returns the ID of the currently active panel.
func (pm *PanelManager) ActiveID() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.activeID
}

// PanelCount returns the total number of registered panels.
func (pm *PanelManager) PanelCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.order)
}

// HasPanel checks if a panel with the given ID is registered.
func (pm *PanelManager) HasPanel(id string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	_, ok := pm.panels[id]
	return ok
}

// SetUnread sets the unread message count for a panel (shown as badge in tab bar).
func (pm *PanelManager) SetUnread(id string, count int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.unread[id] = count
}

// SetTheme updates the theme used for rendering.
func (pm *PanelManager) SetTheme(t *theme.Theme) {
	pm.mu.Lock()
	pm.current = t
	pm.mu.Unlock()
}

// Measure returns the panel manager's desired size.
func (pm *PanelManager) Measure(cs component.Constraints) component.Size {
	w := cs.MaxWidth
	h := cs.MaxHeight
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	return component.Size{W: w, H: h}
}

// SetBounds sets the panel manager's bounds and propagates to the active panel.
func (pm *PanelManager) SetBounds(r component.Rect) {
	pm.BaseComponent.SetBounds(r)
	pm.mu.RLock()
	active := pm.panels[pm.activeID]
	pm.mu.RUnlock()
	if active != nil {
		// Content area excludes the tab bar at the bottom (1 line)
		content := component.Rect{X: r.X, Y: r.Y, W: r.W, H: r.H - 1}
		if content.H < 0 {
			content.H = 0
		}
		active.SetBounds(content)
	}
}

// Paint renders the active panel content and the tab bar.
func (pm *PanelManager) Paint(buf *buffer.Buffer) {
	pm.mu.RLock()
	active := pm.panels[pm.activeID]
	bounds := pm.BaseComponent.Bounds()
	t := pm.current
	pm.mu.RUnlock()

	// Paint active panel
	if active != nil {
		active.Paint(buf)
	}

	// Paint tab bar at the bottom
	tabY := bounds.Y + bounds.H - 1
	if tabY < 0 || tabY >= buf.Height {
		return
	}

	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Determine colors
	var activeFg, activeBg, inactiveFg, inactiveBg buffer.Color
	if t != nil {
		activeFg = t.Fg
		activeBg = t.Accent
		inactiveFg = t.Muted
		inactiveBg = t.Bg
	} else {
		activeFg = buffer.RGB(255, 255, 255)
		activeBg = buffer.RGB(98, 114, 164) // dracula comment
		inactiveFg = buffer.RGB(139, 143, 159)
		inactiveBg = buffer.RGB(40, 42, 54)
	}

	x := bounds.X
	maxX := bounds.X + bounds.W

	for i, id := range pm.order {
		p := pm.panels[id]
		if p == nil {
			continue
		}

		// Build label: "[N] Icon Title"
		label := "[" + itoa(i+1) + "]"
		icon := p.Icon()
		if icon != "" {
			label += " " + icon
		}
		label += " " + p.Title()

		// Add unread badge
		unread := pm.unread[id]
		if unread > 0 {
			label += "(" + itoa(unread) + ")"
		}

		// Check if it fits
		labelLen := visibleLen(label)
		if x+labelLen > maxX {
			// Draw "›" overflow indicator
			if x < maxX {
				buf.SetCell(x, tabY, buffer.Cell{Rune: ' ', Width: 1})
			}
			break
		}

		// Determine colors
		isActive := id == pm.activeID
		fg, bg := inactiveFg, inactiveBg
		if isActive {
			fg, bg = activeFg, activeBg
		}

		// Draw label
		for _, r := range label {
			if x >= maxX {
				break
			}
			buf.SetCell(x, tabY, buffer.Cell{Rune: r, Width: 1, Fg: fg, Bg: bg})
			x++
		}
		// Separator space
		if x < maxX {
			buf.SetCell(x, tabY, buffer.Cell{Rune: ' ', Width: 1, Bg: inactiveBg})
			x++
		}
	}
}

// HandleKey processes panel switching keys first, then forwards to the active panel.
func (pm *PanelManager) HandleKey(k *term.KeyEvent) bool {
	// Alt+1..9 → switch panel
	if k.Modifiers&term.ModAlt != 0 {
		if n := keyToDigit(k); n > 0 {
			if pm.SwitchByIndex(n) {
				return true
			}
		}
	}

	// Ctrl+Tab → next, Ctrl+Shift+Tab → prev
	if k.Key == term.KeyTab {
		if k.Modifiers&term.ModCtrl != 0 {
			if k.Modifiers&term.ModShift != 0 {
				pm.Prev()
			} else {
				pm.Next()
			}
			return true
		}
	}

	// Forward to active panel
	pm.mu.RLock()
	active := pm.panels[pm.activeID]
	pm.mu.RUnlock()

	if active != nil {
		type keyHandler interface {
			HandleKey(*term.KeyEvent) bool
		}
		if kh, ok := active.(keyHandler); ok {
			return kh.HandleKey(k)
		}
	}
	return false
}

// --- helpers ---

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func keyToDigit(k *term.KeyEvent) int {
	if k.Rune >= '1' && k.Rune <= '9' {
		return int(k.Rune - '0')
	}
	return 0
}

func visibleLen(s string) int {
	count := 0
	for range s {
		count++
	}
	return count
}

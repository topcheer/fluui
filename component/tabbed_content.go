package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// TabbedContent combines a TabBar with a ContentSwitcher, providing
// a complete tabbed interface where selecting a tab shows its content.
type TabbedContent struct {
	BaseComponent

	tabs     *TabBar
	switcher *ContentSwitcher

	mu sync.RWMutex
}

// NewTabbedContent creates a tabbed content container.
func NewTabbedContent() *TabbedContent {
	return &TabbedContent{
		tabs:     NewTabBar(),
		switcher: NewContentSwitcher(),
	}
}

// AddTab adds a tab with the given ID, label, and content component.
func (tc *TabbedContent) AddTab(id, label string, content Component) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.tabs.AddTab(id, label)
	tc.switcher.Add(id, content)
	// Activate first tab
	if tc.tabs.TabCount() == 1 {
		tc.switcher.SetCurrent(id)
	}
}

// RemoveTab removes a tab by ID.
func (tc *TabbedContent) RemoveTab(id string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.tabs.RemoveTab(id)
	tc.switcher.Remove(id)
	// Sync switcher to active tab
	tc.syncSwitcherLocked()
}

// SwitchTo activates the tab with the given ID.
func (tc *TabbedContent) SwitchTo(id string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.tabs.SetActiveByID(id)
	tc.syncSwitcherLocked()
}

// NextTab switches to the next tab.
func (tc *TabbedContent) NextTab() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.tabs.NextTab()
	tc.syncSwitcherLocked()
}

// PrevTab switches to the previous tab.
func (tc *TabbedContent) PrevTab() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.tabs.PrevTab()
	tc.syncSwitcherLocked()
}

// ActiveTab returns the active tab ID.
func (tc *TabbedContent) ActiveTab() string {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	idx := tc.tabs.ActiveIndex()
	tab := tc.tabs.TabAt(idx)
	if tab != nil {
		return tab.ID
	}
	return ""
}

// TabCount returns the number of tabs.
func (tc *TabbedContent) TabCount() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.tabs.TabCount()
}

// SetStyle sets the tab bar style.
func (tc *TabbedContent) SetStyle(s TabBarStyle) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.tabs.SetStyle(s)
}

// syncSwitcherLocked updates the switcher to match the active tab.
func (tc *TabbedContent) syncSwitcherLocked() {
	idx := tc.tabs.ActiveIndex()
	tab := tc.tabs.TabAt(idx)
	if tab != nil {
		tc.switcher.SetCurrent(tab.ID)
	}
}

// Measure returns the desired size.
func (tc *TabbedContent) Measure(cs Constraints) Size {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	h := 1 // tab bar
	contentSize := tc.switcher.Measure(cs)
	h += contentSize.H
	if h < 3 {
		h = 3
	}
	w := cs.MaxWidth
	if w <= 0 {
		w = 40
	}
	return Size{W: w, H: h}
}

// SetBounds sets the layout bounds.
func (tc *TabbedContent) SetBounds(r Rect) {
	tc.mu.Lock()
	tc.BaseComponent.SetBounds(r)
	tc.tabs.SetBounds(Rect{X: r.X, Y: r.Y, W: r.W, H: 1})
	tc.switcher.SetBounds(Rect{X: r.X, Y: r.Y + 1, W: r.W, H: r.H - 1})
	tc.mu.Unlock()
}

// Paint renders the tab bar and active content.
func (tc *TabbedContent) Paint(buf *buffer.Buffer) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	if buf == nil {
		return
	}
	tc.tabs.Paint(buf)
	tc.switcher.Paint(buf)
}

// HandleKey processes keyboard navigation (Tab/Ctrl+Left/Ctrl+Right).
func (tc *TabbedContent) HandleKey(k *term.KeyEvent) bool {
	if k == nil {
		return false
	}
	tc.mu.Lock()
	defer tc.mu.Unlock()

	switch k.Key {
	case term.KeyRight:
		if k.Modifiers&(term.ModAlt|term.ModCtrl) != 0 {
			tc.tabs.NextTab()
			tc.syncSwitcherLocked()
			return true
		}
	case term.KeyLeft:
		if k.Modifiers&(term.ModAlt|term.ModCtrl) != 0 {
			tc.tabs.PrevTab()
			tc.syncSwitcherLocked()
			return true
		}
	case term.KeyTab:
		tc.tabs.NextTab()
		tc.syncSwitcherLocked()
		return true
	}

	// Forward to active content
	if child := tc.switcher.CurrentComponent(); child != nil {
		if hk, ok := child.(interface{ HandleKey(*term.KeyEvent) bool }); ok {
			return hk.HandleKey(k)
		}
	}
	return false
}

// HandleMouse handles mouse clicks on the tab bar.
func (tc *TabbedContent) HandleMouse(x, y int, action term.MouseAction) bool {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	bounds := tc.Bounds()
	if y == bounds.Y {
		idx := tc.tabs.HitTest(x, y)
		if idx >= 0 {
			tc.tabs.SetActive(idx)
			tc.syncSwitcherLocked()
			return true
		}
	}
	return false
}

// Children returns the tabs and switcher.
func (tc *TabbedContent) Children() []Component {
	return []Component{tc.tabs, tc.switcher}
}

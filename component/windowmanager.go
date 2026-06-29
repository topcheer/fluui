package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// WindowManager manages a tree of split panes with focus tracking.
// It allows adding/removing panes, resizing splits, and cycling focus
// between panes. The root is always a SplitPane or a single Component.
type WindowManager struct {
	mu sync.RWMutex

	root    Component         // root layout (SplitPane or single component)
	panes   []*ManagedPane    // all registered panes
	focused int               // index of focused pane
}

// ManagedPane wraps a Component with metadata for window management.
type ManagedPane struct {
	Component                       // the actual component
	ID        string                // unique identifier
	Label     string                // display label for status
	closable  bool                  // whether this pane can be closed
}

// NewWindowManager creates a WindowManager with a single initial pane.
func NewWindowManager(initial Component) *WindowManager {
	wm := &WindowManager{
		panes: make([]*ManagedPane, 0),
	}
	wm.addPane(initial, "main", true)
	wm.root = initial
	return wm
}

// addPane registers a new pane. Caller must hold lock.
func (wm *WindowManager) addPane(c Component, label string, closable bool) {
	mp := &ManagedPane{
		Component: c,
		ID:        GenerateID("pane"),
		Label:     label,
		closable:  closable,
	}
	wm.panes = append(wm.panes, mp)
}

// Root returns the root layout component.
func (wm *WindowManager) Root() Component {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.root
}

// Panes returns all managed panes.
func (wm *WindowManager) Panes() []*ManagedPane {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	result := make([]*ManagedPane, len(wm.panes))
	copy(result, wm.panes)
	return result
}

// PaneCount returns the number of panes.
func (wm *WindowManager) PaneCount() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return len(wm.panes)
}

// Focused returns the currently focused pane, or nil if none.
func (wm *WindowManager) Focused() *ManagedPane {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	if wm.focused < 0 || wm.focused >= len(wm.panes) {
		return nil
	}
	return wm.panes[wm.focused]
}

// FocusedIndex returns the index of the focused pane.
func (wm *WindowManager) FocusedIndex() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.focused
}

// FocusNext moves focus to the next pane (wraps around).
func (wm *WindowManager) FocusNext() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if len(wm.panes) <= 1 {
		return
	}
	wm.focused = (wm.focused + 1) % len(wm.panes)
}

// FocusPrev moves focus to the previous pane (wraps around).
func (wm *WindowManager) FocusPrev() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if len(wm.panes) <= 1 {
		return
	}
	wm.focused = (wm.focused - 1 + len(wm.panes)) % len(wm.panes)
}

// FocusIndex sets focus to the pane at the given index.
func (wm *WindowManager) FocusIndex(idx int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if idx < 0 || idx >= len(wm.panes) {
		return
	}
	wm.focused = idx
}

// SplitRight adds a new pane by splitting the focused pane horizontally.
// The new pane is placed on the right.
func (wm *WindowManager) SplitRight(newPane Component, label string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.addPane(newPane, label, true)
	wm.rebuild()
}

// SplitDown adds a new pane by splitting the focused pane vertically.
// The new pane is placed on the bottom.
func (wm *WindowManager) SplitDown(newPane Component, label string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.addPane(newPane, label, true)
	wm.rebuild()
}

// ClosePane removes the focused pane and rebuilds the layout.
// Returns true if a pane was closed.
func (wm *WindowManager) ClosePane() bool {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if len(wm.panes) <= 1 {
		return false // can't close the last pane
	}
	if !wm.panes[wm.focused].closable {
		return false
	}
	// Remove focused pane
	wm.panes = append(wm.panes[:wm.focused], wm.panes[wm.focused+1:]...)
	if wm.focused >= len(wm.panes) {
		wm.focused = len(wm.panes) - 1
	}
	wm.rebuild()
	return true
}

// rebuild reconstructs the layout tree from the pane list.
// Caller must hold lock.
func (wm *WindowManager) rebuild() {
	if len(wm.panes) == 0 {
		wm.root = nil
		return
	}
	if len(wm.panes) == 1 {
		wm.root = wm.panes[0].Component
		return
	}

	// Build a balanced split tree
	comps := make([]Component, len(wm.panes))
	for i, p := range wm.panes {
		comps[i] = p.Component
	}

	// For 2 panes: simple split
	if len(comps) == 2 {
		sp := NewSplitPane(comps[0], comps[1])
		sp.SetDirection(SplitHorizontal)
		wm.root = sp
		return
	}

	// For 3+ panes: nested splits
	// First pane on left, rest split recursively on right
	left := comps[0]
	right := wm.buildTree(comps[1:], SplitHorizontal)
	sp := NewSplitPane(left, right)
	sp.SetDirection(SplitHorizontal)
	wm.root = sp
}

// buildTree recursively builds a split tree from a list of components.
func (wm *WindowManager) buildTree(comps []Component, dir SplitDirection) Component {
	if len(comps) == 1 {
		return comps[0]
	}
	if len(comps) == 2 {
		sp := NewSplitPane(comps[0], comps[1])
		sp.SetDirection(dir)
		return sp
	}
	// Split first from rest, alternating direction
	nextDir := SplitHorizontal
	if dir == SplitHorizontal {
		nextDir = SplitVertical
	}
	left := comps[0]
	right := wm.buildTree(comps[1:], nextDir)
	sp := NewSplitPane(left, right)
	sp.SetDirection(dir)
	return sp
}

// Measure delegates to the root component.
func (wm *WindowManager) Measure(cs Constraints) Size {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	if wm.root == nil {
		return Size{}
	}
	return wm.root.Measure(cs)
}

// Paint renders the root component and highlights the focused pane border.
func (wm *WindowManager) Paint(buf *buffer.Buffer) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	if wm.root == nil {
		return
	}
	wm.root.Paint(buf)

	// Highlight focused pane bounds with a distinct border
	if wm.focused >= 0 && wm.focused < len(wm.panes) {
		b := wm.panes[wm.focused].Component.Bounds()
		wm.highlightFocus(buf, b)
	}
}

// highlightFocus draws corner markers around the focused pane.
func (wm *WindowManager) highlightFocus(buf *buffer.Buffer, b Rect) {
	if b.W <= 0 || b.H <= 0 {
		return
	}
	highlightColor := buffer.RGB(100, 200, 255)
	corners := [][2]int{
		{b.X, b.Y},               // top-left
		{b.X + b.W - 1, b.Y},     // top-right
		{b.X, b.Y + b.H - 1},     // bottom-left
		{b.X + b.W - 1, b.Y + b.H - 1}, // bottom-right
	}
	for _, c := range corners {
		buf.SetCell(c[0], c[1], buffer.Cell{
			Rune:  '◆',
			Width: 1,
			Fg:    highlightColor,
		})
	}
}

// SetBounds sets bounds on the root component.
func (wm *WindowManager) SetBounds(b Rect) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if wm.root != nil {
		wm.root.SetBounds(b)
	}
}

// Bounds returns the root component's bounds.
func (wm *WindowManager) Bounds() Rect {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	if wm.root == nil {
		return Rect{}
	}
	return wm.root.Bounds()
}

// SetShowHandle toggles the divider grip on all SplitPanes in the tree.
func (wm *WindowManager) SetShowHandle(show bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	wm.applyShowHandle(wm.root, show)
}

func (wm *WindowManager) applyShowHandle(c Component, show bool) {
	if c == nil {
		return
	}
	if sp, ok := c.(*SplitPane); ok {
		sp.SetShowHandle(show)
		for _, child := range sp.Children() {
			wm.applyShowHandle(child, show)
		}
	}
}

// Equalize sets all splits to 50/50 ratio.
func (wm *WindowManager) Equalize() {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	wm.applyEqualize(wm.root)
}

func (wm *WindowManager) applyEqualize(c Component) {
	if c == nil {
		return
	}
	if sp, ok := c.(*SplitPane); ok {
		sp.SetRatio(0.5)
		for _, child := range sp.Children() {
			wm.applyEqualize(child)
		}
	}
}

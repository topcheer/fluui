package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ContentSwitcher: Swap Between Child Components ───
//
// ContentSwitcher is a container that swaps between multiple child
// components by ID. Inspired by Textual's ContentSwitcher.
//
// Usage:
//	cs := NewContentSwitcher()
//	cs.Add("home", homePanel)
//	cs.Add("settings", settingsPanel)
//	cs.SetCurrent("settings")  // now shows settingsPanel
//	cs.Current()               // "settings"

// ContentSwitcher shows one child at a time, swapping by ID.
type ContentSwitcher struct {
	mu       sync.RWMutex
	BaseComponent
	children map[string]Component
	order    []string // track insertion order
	current  string
}

// NewContentSwitcher creates an empty switcher.
func NewContentSwitcher() *ContentSwitcher {
	return &ContentSwitcher{
		children: make(map[string]Component),
	}
}

// Add registers a child component with an ID.
func (cs *ContentSwitcher) Add(id string, child Component) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if _, exists := cs.children[id]; !exists {
		cs.order = append(cs.order, id)
	}
	cs.children[id] = child
	if cs.current == "" {
		cs.current = id
	}
}

// Remove unregisters a child.
func (cs *ContentSwitcher) Remove(id string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.children, id)
	for i, oid := range cs.order {
		if oid == id {
			cs.order = append(cs.order[:i], cs.order[i+1:]...)
			break
		}
	}
	if cs.current == id {
		if len(cs.order) > 0 {
			cs.current = cs.order[0]
		} else {
			cs.current = ""
		}
	}
}

// SetCurrent switches the visible child by ID.
// Returns false if the ID doesn't exist.
func (cs *ContentSwitcher) SetCurrent(id string) bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if _, exists := cs.children[id]; !exists {
		return false
	}
	cs.current = id
	return true
}

// Current returns the active child ID.
func (cs *ContentSwitcher) Current() string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.current
}

// CurrentComponent returns the active child component (or nil).
func (cs *ContentSwitcher) CurrentComponent() Component {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.children[cs.current]
}

// IDs returns all registered IDs in insertion order.
func (cs *ContentSwitcher) IDs() []string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	result := make([]string, len(cs.order))
	copy(result, cs.order)
	return result
}

// Next switches to the next child (wraps around).
func (cs *ContentSwitcher) Next() {
	cs.mu.Lock()
	if len(cs.order) == 0 {
		cs.mu.Unlock()
		return
	}
	for i, id := range cs.order {
		if id == cs.current {
			cs.current = cs.order[(i+1)%len(cs.order)]
			cs.mu.Unlock()
			return
		}
	}
	cs.mu.Unlock()
}

// Prev switches to the previous child (wraps around).
func (cs *ContentSwitcher) Prev() {
	cs.mu.Lock()
	if len(cs.order) == 0 {
		cs.mu.Unlock()
		return
	}
	for i, id := range cs.order {
		if id == cs.current {
			cs.current = cs.order[(i-1+len(cs.order))%len(cs.order)]
			cs.mu.Unlock()
			return
		}
	}
	cs.mu.Unlock()
}

// Count returns the number of registered children.
func (cs *ContentSwitcher) Count() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return len(cs.children)
}

func (cs *ContentSwitcher) Measure(constraints Constraints) Size {
	cs.mu.RLock()
	current := cs.children[cs.current]
	cs.mu.RUnlock()

	if current == nil {
		return Size{W: 0, H: 0}
	}
	return current.Measure(constraints)
}

func (cs *ContentSwitcher) Paint(buf *buffer.Buffer) {
	cs.mu.RLock()
	current := cs.children[cs.current]
	cs.mu.RUnlock()

	if current == nil {
		return
	}

	bounds := cs.Bounds()
	current.SetBounds(bounds)
	current.Paint(buf)
}

func (cs *ContentSwitcher) Children() []Component {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	result := make([]Component, 0, len(cs.order))
	for _, id := range cs.order {
		if c := cs.children[id]; c != nil {
			result = append(result, c)
		}
	}
	return result
}
// Package focus implements focus management for focusable UI components.
//
// A Focusable component can receive and lose keyboard focus. The FocusManager
// maintains an ordered ring of Focusable items and supports cyclic Tab / Shift-Tab
// traversal, direct focus setting, and safe removal.
package focus

// Focusable represents a component that can receive keyboard focus.
type Focusable interface {
	// Focus marks this component as focused.
	Focus()
	// Blur removes focus from this component.
	Blur()
	// Focused reports whether this component is currently focused.
	Focused() bool
}

// FocusManager manages an ordered ring of Focusable items.
// At most one item is focused at a time. When current is -1, no item is focused.
type FocusManager struct {
	items   []Focusable
	current int // index of the focused item, -1 = none
}

// NewFocusManager creates an empty FocusManager.
func NewFocusManager() *FocusManager {
	return &FocusManager{
		current: -1,
	}
}

// Add appends a Focusable to the ring. If this is the first item, it does NOT
// auto-focus it — call Set or Next explicitly to begin focus traversal.
func (m *FocusManager) Add(f Focusable) {
	m.items = append(m.items, f)
}

// Remove deletes f from the ring. If f was focused, focus moves to the item
// now occupying the same index (or wraps to the last item if f was last).
// After removal the ring remains consistent and no item is left focused spuriously.
func (m *FocusManager) Remove(f Focusable) {
	for i, item := range m.items {
		if item == f {
			// Blur the item being removed.
			f.Blur()
			m.items = append(m.items[:i], m.items[i+1:]...)

			if len(m.items) == 0 {
				m.current = -1
				return
			}

			// Adjust current index.
			switch {
			case i < m.current:
				// Removed item was before current; shift index down.
				m.current--
			case i == m.current:
				// Removed item was current; clamp index and focus replacement.
				if m.current >= len(m.items) {
					m.current = len(m.items) - 1
				}
				m.items[m.current].Focus()
			case i > m.current:
				// Removed item was after current; index unchanged.
			}
			return
		}
	}
}

// Next moves focus to the next item in the ring (cyclic).
// Calling Next on an empty manager is a no-op.
func (m *FocusManager) Next() {
	n := len(m.items)
	if n == 0 {
		return
	}
	m.blurCurrent()
	m.current = (m.current + 1) % n
	m.focusCurrent()
}

// Prev moves focus to the previous item in the ring (cyclic).
// Calling Prev on an empty manager is a no-op.
// When no item is currently focused, the first Prev focuses the first item
// (same as Next), then subsequent calls traverse backwards.
func (m *FocusManager) Prev() {
	n := len(m.items)
	if n == 0 {
		return
	}
	m.blurCurrent()
	if m.current < 0 {
		// Nothing focused yet — start at the first item.
		m.current = 0
	} else {
		m.current = (m.current - 1 + n) % n
	}
	m.focusCurrent()
}

// Set focuses the given Focusable. If f is not in the ring, focus is unchanged.
func (m *FocusManager) Set(f Focusable) {
	for i, item := range m.items {
		if item == f {
			m.blurCurrent()
			m.current = i
			m.focusCurrent()
			return
		}
	}
}

// Clear removes focus from all items. No item will be focused afterwards.
func (m *FocusManager) Clear() {
	m.blurCurrent()
	m.current = -1
}

// Current returns the currently focused Focusable, or nil if none.
func (m *FocusManager) Current() Focusable {
	if m.current < 0 || m.current >= len(m.items) {
		return nil
	}
	return m.items[m.current]
}

// Len returns the number of items in the ring.
func (m *FocusManager) Len() int {
	return len(m.items)
}

// blurCurrent calls Blur() on the currently focused item, if any.
func (m *FocusManager) blurCurrent() {
	if m.current >= 0 && m.current < len(m.items) {
		m.items[m.current].Blur()
	}
}

// focusCurrent calls Focus() on the item at the current index, if valid.
func (m *FocusManager) focusCurrent() {
	if m.current >= 0 && m.current < len(m.items) {
		m.items[m.current].Focus()
	}
}

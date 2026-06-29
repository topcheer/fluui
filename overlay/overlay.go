// Package overlay implements the overlay/modal layer system.
//
// Overlays sit above the main content area and support z-index stacking.
// Modal overlays block input to layers below them. Typical uses include
// modal dialogs, full-screen code viewers, and tooltip popups.
package overlay

import (
	"sort"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// Overlay is the interface for any overlay layer.
type Overlay interface {
	component.Component

	// Identity
	ID() string

	// Stacking
	Z() int    // z-index (higher = on top)
	Visible() bool
	SetVisible(bool)

	// Behavior
	Modal() bool // modal overlays block events to lower layers

	// Input
	HandleKey(key interface{}) bool   // returns true if consumed
	HandleMouse(x, y int) bool        // returns true if consumed
}

// BaseOverlay provides common overlay fields and default methods.
type BaseOverlay struct {
	component.BaseComponent
	id     string
	z      int
	visible bool
	modal  bool
}

// NewBaseOverlay creates a BaseOverlay.
func NewBaseOverlay(id string, z int, modal bool) BaseOverlay {
	return BaseOverlay{
		id:      id,
		z:       z,
		modal:   modal,
		visible: true,
	}
}

// ID returns the overlay's unique identifier.
func (o *BaseOverlay) ID() string         { return o.id }

// Z returns the overlay's z-index for stacking order.
func (o *BaseOverlay) Z() int              { return o.z }

// Visible returns whether the overlay is currently visible.
func (o *BaseOverlay) Visible() bool       { return o.visible }

// SetVisible sets the overlay's visibility.
func (o *BaseOverlay) SetVisible(v bool)   { o.visible = v }

// Modal returns whether the overlay blocks input to underlying layers.
func (o *BaseOverlay) Modal() bool         { return o.modal }

// Default handlers do nothing (not consumed).
func (o *BaseOverlay) HandleKey(_ interface{}) bool  { return false }
func (o *BaseOverlay) HandleMouse(_, _ int) bool     { return false }

// OverlayManager manages a stack of overlays, sorted by z-index.
type OverlayManager struct {
	overlays []Overlay
}

// NewOverlayManager creates an empty manager.
func NewOverlayManager() *OverlayManager {
	return &OverlayManager{}
}

// Add inserts an overlay, maintaining z-index order (ascending).
func (m *OverlayManager) Add(o Overlay) {
	// Avoid duplicate IDs
	for _, existing := range m.overlays {
		if existing.ID() == o.ID() {
			return
		}
	}
	m.overlays = append(m.overlays, o)
	m.sort()
}

// Remove deletes an overlay by ID.
func (m *OverlayManager) Remove(id string) bool {
	for i, o := range m.overlays {
		if o.ID() == id {
			m.overlays = append(m.overlays[:i], m.overlays[i+1:]...)
			return true
		}
	}
	return false
}

// Show adds an overlay and makes it visible. Convenience over Add.
func (m *OverlayManager) Show(o Overlay) {
	o.SetVisible(true)
	m.Add(o)
}

// HideAll hides all overlays.
func (m *OverlayManager) HideAll() {
	for _, o := range m.overlays {
		o.SetVisible(false)
	}
}

// Measure measures all visible overlays and sets their bounds within the screen.
// Must be called before Paint so overlays know their position.
func (m *OverlayManager) Measure(w, h int) {
	cs := component.Constraints{MaxWidth: w, MaxHeight: h}
	screenRect := component.Rect{X: 0, Y: 0, W: w, H: h}
	for _, o := range m.overlays {
		if o.Visible() {
			o.Measure(cs)
			o.SetBounds(screenRect)
		}
	}
}

// Top returns the topmost visible overlay, or nil if none.
func (m *OverlayManager) Top() Overlay {
	for i := len(m.overlays) - 1; i >= 0; i-- {
		if m.overlays[i].Visible() {
			return m.overlays[i]
		}
	}
	return nil
}

// Visible returns all visible overlays in ascending z-index order.
func (m *OverlayManager) Visible() []Overlay {
	var result []Overlay
	for _, o := range m.overlays {
		if o.Visible() {
			result = append(result, o)
		}
	}
	return result
}

// Len returns the total number of overlays (visible + hidden).
func (m *OverlayManager) Len() int {
	return len(m.overlays)
}

// Get retrieves an overlay by ID.
func (m *OverlayManager) Get(id string) Overlay {
	for _, o := range m.overlays {
		if o.ID() == id {
			return o
		}
	}
	return nil
}

// Paint renders all visible overlays into the buffer, in z-index order
// (lowest first, highest last so it appears on top).
func (m *OverlayManager) Paint(buf *buffer.Buffer) {
	for _, o := range m.overlays {
		if o.Visible() {
			o.Paint(buf)
		}
	}
}

// HandleKey routes keyboard input to the topmost visible modal overlay.
// Returns true if any overlay consumed the event.
func (m *OverlayManager) HandleKey(key interface{}) bool {
	for i := len(m.overlays) - 1; i >= 0; i-- {
		o := m.overlays[i]
		if !o.Visible() {
			continue
		}
		if o.HandleKey(key) {
			return true
		}
		if o.Modal() {
			// Modal blocks events from reaching lower layers
			return true
		}
	}
	return false
}

// HandleMouse routes mouse input to the topmost visible overlay at the position.
// Returns true if any overlay consumed the event.
func (m *OverlayManager) HandleMouse(x, y int) bool {
	for i := len(m.overlays) - 1; i >= 0; i-- {
		o := m.overlays[i]
		if !o.Visible() {
			continue
		}
		if o.HandleMouse(x, y) {
			return true
		}
		// Check if the click is within the overlay bounds
		bounds := o.Bounds()
		if x >= bounds.X && x < bounds.X+bounds.W &&
			y >= bounds.Y && y < bounds.Y+bounds.H {
			// Click hit this overlay, consume it
			if o.Modal() {
				return true
			}
		}
	}
	return false
}

// HasModal returns true if any visible modal overlay exists.
func (m *OverlayManager) HasModal() bool {
	for _, o := range m.overlays {
		if o.Visible() && o.Modal() {
			return true
		}
	}
	return false
}

// sort re-sorts overlays by z-index ascending.
func (m *OverlayManager) sort() {
	sort.Slice(m.overlays, func(i, j int) bool {
		return m.overlays[i].Z() < m.overlays[j].Z()
	})
}

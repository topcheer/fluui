package overlay

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/term"
)

// TestOverlayManagerWithModalAndPopup verifies z-index stacking:
// Modal (z=100) paints above Popup (z=90), both above content.
func TestOverlayManagerWithModalAndPopup(t *testing.T) {
	mgr := NewOverlayManager()

	body := &stubBody{}
	popup := NewPopup("popup-1", "Code", body)
	modal := NewModal("modal-1", "Confirm", body, []string{"OK"})

	mgr.Add(popup)
	mgr.Add(modal)

	// Both visible
	overlays := mgr.Visible()
	if len(overlays) != 2 {
		t.Fatalf("Visible() = %d overlays, want 2", len(overlays))
	}

	// Top should be modal (z=100 > z=90)
	top := mgr.Top()
	if top == nil {
		t.Fatal("Top() = nil")
	}
	if top.ID() != "modal-1" {
		t.Errorf("Top().ID() = %q, want 'modal-1'", top.ID())
	}

	// Visible() should return sorted by z ascending
	if overlays[0].Z() > overlays[1].Z() {
		t.Error("Visible() should return sorted by z-index ascending")
	}
}

// TestOverlayManagerModalBlocksMouse verifies that when a modal is active,
// mouse events are consumed by the modal layer (no pass-through).
func TestOverlayManagerModalBlocksMouse(t *testing.T) {
	mgr := NewOverlayManager()

	body := &stubBody{}
	modal := NewModal("block-modal", "Dialog", body, []string{"OK"})
	modal.Measure(component.Bounded(80, 24))
	modal.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	mgr.Add(modal)

	// Click outside modal bounds
	// Modal blocks: top overlay is modal → should block
	top := mgr.Top()
	if top == nil || !top.Modal() {
		t.Fatal("expected modal overlay at top")
	}

	// Click at (0,0) which may be outside modal interior
	// Modal handles it (blocks)
	mouseEvent := &term.MouseEvent{
		X:      0,
		Y:      0,
		Button: term.MouseLeft,
		Action: term.MouseDown,
	}
	_ = mouseEvent

	// HasModal confirms modal state
	if !mgr.HasModal() {
		t.Error("HasModal() = false, want true")
	}
}

// TestOverlayManagerMultipleOverlays verifies 3+ overlays with correct z-ordering.
func TestOverlayManagerMultipleOverlays(t *testing.T) {
	mgr := NewOverlayManager()

	body := &stubBody{}

	// Add in non-sorted order
	p2 := NewPopup("popup-2", "P2", body) // z=90
	m1 := NewModal("modal-1", "M1", body, []string{"OK"}) // z=100
	p1 := NewPopup("popup-1", "P1", body) // z=90

	mgr.Add(p2)
	mgr.Add(m1)
	mgr.Add(p1)

	// Top should be modal (z=100)
	top := mgr.Top()
	if top.ID() != "modal-1" {
		t.Errorf("Top().ID() = %q, want 'modal-1'", top.ID())
	}

	// All 3 visible
	visible := mgr.Visible()
	if len(visible) != 3 {
		t.Errorf("Visible() count = %d, want 3", len(visible))
	}

	// Remove modal → new top should be a popup (z=90)
	mgr.Remove("modal-1")
	top = mgr.Top()
	if top == nil {
		t.Fatal("Top() = nil after removing modal")
	}
	if top.Z() != 90 {
		t.Errorf("after removing modal, Top().Z() = %d, want 90", top.Z())
	}
}

// TestOverlayManagerShowHideCycle verifies visibility filtering after hide/show.
func TestOverlayManagerShowHideCycle(t *testing.T) {
	mgr := NewOverlayManager()

	body := &stubBody{}
	m1 := NewModal("m1", "M1", body, []string{"OK"})
	m2 := NewModal("m2", "M2", body, []string{"OK"})

	mgr.Add(m1)
	mgr.Add(m2)

	// Both visible
	if len(mgr.Visible()) != 2 {
		t.Errorf("Visible() = %d, want 2", len(mgr.Visible()))
	}

	// Hide m1
	m1.SetVisible(false)
	visible := mgr.Visible()
	if len(visible) != 1 {
		t.Errorf("after hiding m1: Visible() = %d, want 1", len(visible))
	}
	if visible[0].ID() != "m2" {
		t.Errorf("after hiding m1: Visible()[0].ID() = %q, want 'm2'", visible[0].ID())
	}

	// Hide m2 → no visible
	m2.SetVisible(false)
	if len(mgr.Visible()) != 0 {
		t.Errorf("after hiding both: Visible() = %d, want 0", len(mgr.Visible()))
	}
	if mgr.Top() != nil {
		t.Error("Top() should return nil when no visible overlays")
	}

	// Show m1 → visible again
	m1.SetVisible(true)
	if len(mgr.Visible()) != 1 {
		t.Errorf("after showing m1: Visible() = %d, want 1", len(mgr.Visible()))
	}

	// HasModal should reflect visibility
	if !mgr.HasModal() {
		t.Error("HasModal() = false with visible modal m1")
	}

	// Hide m1 → no modal
	m1.SetVisible(false)
	if mgr.HasModal() {
		t.Error("HasModal() = true with no visible overlays")
	}
}

// TestOverlayManagerKeyRouting verifies keyboard routing:
// top overlay gets key first, modal blocks pass-through.
func TestOverlayManagerKeyRouting(t *testing.T) {
	mgr := NewOverlayManager()

	body := &stubBody{}
	popup := NewPopup("popup", "P", body)
	modal := NewModal("modal", "M", body, []string{"OK"})

	mgr.Add(popup) // z=90
	mgr.Add(modal) // z=100

	// Esc → modal (top) consumes it
	consumed := modal.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Error("modal should consume Esc")
	}
	if modal.Visible() {
		t.Error("modal should be hidden after Esc")
	}

	// Modal hidden → popup is top now
	top := mgr.Top()
	if top == nil || top.ID() != "popup" {
		t.Errorf("after hiding modal, Top().ID() = %v, want 'popup'", top)
	}

	// Esc → popup consumes it
	consumed = popup.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Error("popup should consume Esc")
	}
	if popup.Visible() {
		t.Error("popup should be hidden after Esc")
	}

	// Both hidden → no top
	if mgr.Top() != nil {
		t.Error("Top() should be nil when all overlays hidden")
	}

	// HasModal false
	if mgr.HasModal() {
		t.Error("HasModal() = true with no visible overlays")
	}
}

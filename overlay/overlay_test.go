package overlay

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// mockOverlay is a minimal Overlay implementation for testing.
type mockOverlay struct {
	BaseOverlay
	content string
	w, h    int
	keyHit  bool
	mouseHit bool
}

func newMockOverlay(id string, z int, modal bool, w, h int) *mockOverlay {
	return &mockOverlay{
		BaseOverlay: NewBaseOverlay(id, z, modal),
		w:           w,
		h:           h,
	}
}

func (m *mockOverlay) Measure(cs component.Constraints) component.Size {
	return component.Size{W: m.w, H: m.h}
}

func (m *mockOverlay) SetBounds(r component.Rect) {
	m.BaseComponent.SetBounds(r)
}

func (m *mockOverlay) Paint(buf *buffer.Buffer) {
	// no-op
}

func (m *mockOverlay) HandleKey(_ interface{}) bool { return m.keyHit }
func (m *mockOverlay) HandleMouse(_, _ int) bool     { return m.mouseHit }

func TestOverlayManagerEmpty(t *testing.T) {
	m := NewOverlayManager()
	if m.Len() != 0 {
		t.Errorf("Len() = %d, want 0", m.Len())
	}
	if m.Top() != nil {
		t.Error("Top() on empty manager should be nil")
	}
	if m.HasModal() {
		t.Error("HasModal() on empty manager should be false")
	}
}

func TestOverlayManagerAdd(t *testing.T) {
	m := NewOverlayManager()
	m.Add(newMockOverlay("a", 10, false, 5, 3))
	m.Add(newMockOverlay("b", 20, false, 5, 3))

	if m.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", m.Len())
	}
}

func TestOverlayManagerZOrder(t *testing.T) {
	m := NewOverlayManager()
	// Add in reverse z-order
	m.Add(newMockOverlay("low", 5, false, 5, 3))
	m.Add(newMockOverlay("high", 50, false, 5, 3))
	m.Add(newMockOverlay("mid", 10, false, 5, 3))

	// Top should be the highest z-index
	top := m.Top()
	if top.ID() != "high" {
		t.Errorf("Top().ID() = %q, want 'high'", top.ID())
	}
}

func TestOverlayManagerRemove(t *testing.T) {
	m := NewOverlayManager()
	m.Add(newMockOverlay("a", 10, false, 5, 3))
	m.Add(newMockOverlay("b", 20, false, 5, 3))

	if !m.Remove("a") {
		t.Error("Remove('a') should return true")
	}
	if m.Len() != 1 {
		t.Errorf("Len() = %d, want 1", m.Len())
	}
	if m.Get("a") != nil {
		t.Error("Get('a') after remove should be nil")
	}
	if m.Remove("nonexistent") {
		t.Error("Remove('nonexistent') should return false")
	}
}

func TestOverlayManagerDuplicateID(t *testing.T) {
	m := NewOverlayManager()
	m.Add(newMockOverlay("a", 10, false, 5, 3))
	m.Add(newMockOverlay("a", 20, false, 5, 3)) // duplicate

	if m.Len() != 1 {
		t.Errorf("Len() = %d, want 1 (duplicate ignored)", m.Len())
	}
}

func TestOverlayManagerVisible(t *testing.T) {
	m := NewOverlayManager()
	o1 := newMockOverlay("a", 10, false, 5, 3)
	o2 := newMockOverlay("b", 20, false, 5, 3)
	o2.SetVisible(false)
	m.Add(o1)
	m.Add(o2)

	visible := m.Visible()
	if len(visible) != 1 {
		t.Fatalf("len(Visible()) = %d, want 1", len(visible))
	}
	if visible[0].ID() != "a" {
		t.Errorf("Visible()[0].ID() = %q, want 'a'", visible[0].ID())
	}
}

func TestOverlayManagerTop(t *testing.T) {
	m := NewOverlayManager()
	o1 := newMockOverlay("a", 10, false, 5, 3)
	o2 := newMockOverlay("b", 20, false, 5, 3)
	m.Add(o1)
	m.Add(o2)

	// Top should be b (higher z)
	if m.Top().ID() != "b" {
		t.Errorf("Top().ID() = %q, want 'b'", m.Top().ID())
	}

	// Hide b → top should be a
	o2.SetVisible(false)
	if m.Top().ID() != "a" {
		t.Errorf("Top().ID() = %q, want 'a'", m.Top().ID())
	}

	// Hide all → top nil
	o1.SetVisible(false)
	if m.Top() != nil {
		t.Error("Top() when all hidden should be nil")
	}
}

func TestOverlayManagerHasModal(t *testing.T) {
	m := NewOverlayManager()
	m.Add(newMockOverlay("a", 10, false, 5, 3))
	if m.HasModal() {
		t.Error("HasModal() should be false")
	}

	m.Add(newMockOverlay("b", 20, true, 5, 3)) // modal
	if !m.HasModal() {
		t.Error("HasModal() should be true after adding modal")
	}
}

func TestOverlayManagerHandleKey(t *testing.T) {
	m := NewOverlayManager()
	o1 := newMockOverlay("a", 10, false, 5, 3)
	o2 := newMockOverlay("b", 20, true, 5, 3) // modal
	o2.keyHit = true
	m.Add(o1)
	m.Add(o2)

	// Modal overlay consumes key
	if !m.HandleKey(nil) {
		t.Error("HandleKey should return true (modal consumed)")
	}

	// Remove modal → no overlay consumes
	m.Remove("b")
	if m.HandleKey(nil) {
		t.Error("HandleKey should return false (no consumer)")
	}
}

func TestOverlayManagerHandleKeyModalBlocks(t *testing.T) {
	m := NewOverlayManager()
	o1 := newMockOverlay("a", 10, false, 5, 3)
	o1.keyHit = true // would consume, but modal blocks
	o2 := newMockOverlay("b", 20, true, 5, 3) // modal, doesn't consume
	m.Add(o1)
	m.Add(o2)

	// Modal blocks event from reaching lower layer
	if !m.HandleKey(nil) {
		t.Error("HandleKey should return true (modal blocks)")
	}
}

func TestOverlayManagerPaint(t *testing.T) {
	m := NewOverlayManager()
	o1 := newMockOverlay("a", 10, false, 5, 3)
	o2 := newMockOverlay("b", 20, false, 5, 3)
	o2.SetVisible(false)
	m.Add(o1)
	m.Add(o2)

	// Paint should not panic
	buf := buffer.NewBuffer(80, 24)
	o1.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 3})
	o2.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 3})
	m.Paint(buf)
}

func TestBaseOverlayDefaults(t *testing.T) {
	o := NewBaseOverlay("test", 5, false)
	if o.ID() != "test" {
		t.Errorf("ID() = %q", o.ID())
	}
	if o.Z() != 5 {
		t.Errorf("Z() = %d, want 5", o.Z())
	}
	if o.Modal() {
		t.Error("Modal() should be false")
	}
	if !o.Visible() {
		t.Error("Visible() should be true by default")
	}
}

func TestOverlayManagerGet(t *testing.T) {
	m := NewOverlayManager()
	m.Add(newMockOverlay("find-me", 10, false, 5, 3))

	o := m.Get("find-me")
	if o == nil {
		t.Fatal("Get('find-me') should not be nil")
	}
	if o.ID() != "find-me" {
		t.Errorf("Get().ID() = %q", o.ID())
	}
	if m.Get("missing") != nil {
		t.Error("Get('missing') should be nil")
	}
}

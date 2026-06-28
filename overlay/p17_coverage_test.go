package overlay

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// === OverlayManager.Show coverage ===

func TestOverlayManager_Show(t *testing.T) {
	m := NewOverlayManager()
	o := newMockOverlay("test", 50, false, 5, 3)
	o.SetVisible(false)

	m.Show(o)
	if m.Len() != 1 {
		t.Errorf("Len = %d, want 1 after Show", m.Len())
	}
	if !o.Visible() {
		t.Error("Show should set Visible=true")
	}
}

// === OverlayManager.HideAll coverage ===

func TestOverlayManager_HideAll(t *testing.T) {
	m := NewOverlayManager()
	o1 := newMockOverlay("a", 10, false, 5, 3)
	o2 := newMockOverlay("b", 20, true, 5, 3)
	m.Add(o1)
	m.Add(o2)

	m.HideAll()

	for _, o := range m.overlays {
		if o.Visible() {
			t.Error("HideAll should set Visible=false on all overlays")
		}
	}
}

// === BaseOverlay default HandleKey/HandleMouse coverage ===

func TestBaseOverlay_DefaultHandleKey(t *testing.T) {
	o := NewBaseOverlay("test", 5, false)
	if o.HandleKey(nil) {
		t.Error("Default HandleKey should return false")
	}
}

func TestBaseOverlay_DefaultHandleMouse(t *testing.T) {
	o := NewBaseOverlay("test", 5, false)
	if o.HandleMouse(0, 0) {
		t.Error("Default HandleMouse should return false")
	}
}

// === HandleMouse bounds-checking coverage ===

func TestOverlayManager_HandleMouse_WithinModalBounds(t *testing.T) {
	m := NewOverlayManager()
	o := newMockOverlay("modal", 50, true, 80, 24)
	o.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	o.mouseHit = false // doesn't consume directly, but modal blocks
	m.Add(o)

	if !m.HandleMouse(40, 12) {
		t.Error("HandleMouse should return true for click within modal bounds")
	}
}

func TestOverlayManager_HandleMouse_OutsideModalBounds(t *testing.T) {
	m := NewOverlayManager()
	o := newMockOverlay("modal", 50, true, 10, 5)
	o.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 5})
	m.Add(o)

	if m.HandleMouse(50, 20) {
		t.Error("HandleMouse should return false for click outside modal bounds")
	}
}

func TestOverlayManager_HandleMouse_HiddenOverlay(t *testing.T) {
	m := NewOverlayManager()
	o := newMockOverlay("hidden", 50, true, 80, 24)
	o.SetVisible(false)
	o.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	m.Add(o)

	if m.HandleMouse(40, 12) {
		t.Error("HandleMouse should skip hidden overlays")
	}
}

// === Modal.SetStyle coverage ===

func TestModal_SetStyle(t *testing.T) {
	modal := NewModal("test", "Title", nil, []string{"OK"})
	customStyle := ModalStyle{
		Border: buffer.Style{Fg: buffer.Red, Bg: buffer.Black},
	}
	modal.SetStyle(customStyle)
	// Verify via Paint (style field is unexported)
	buf := buffer.NewBuffer(80, 24)
	modal.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	modal.Paint(buf)
}

// === Measure coverage ===

func TestOverlayManager_Measure(t *testing.T) {
	m := NewOverlayManager()
	o := newMockOverlay("test", 50, false, 20, 10)
	m.Add(o)

	m.Measure(80, 24)

	bounds := o.Bounds()
	if bounds.W != 80 || bounds.H != 24 {
		t.Errorf("Bounds = %+v, want W=80 H=24", bounds)
	}
}

func TestOverlayManager_Measure_HiddenOverlay(t *testing.T) {
	m := NewOverlayManager()
	o := newMockOverlay("hidden", 50, false, 20, 10)
	o.SetVisible(false)
	m.Add(o)

	// Should not measure hidden overlays (no panic = pass)
	m.Measure(80, 24)
}

// === Paint with visible overlays ===

func TestOverlayManager_Paint_VisibleOverlay(t *testing.T) {
	m := NewOverlayManager()
	o := newMockOverlay("test", 50, false, 5, 3)
	o.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 3})
	m.Add(o)

	buf := buffer.NewBuffer(80, 24)
	m.Paint(buf) // should not panic
}

// === HasModal with hidden modal ===

func TestOverlayManager_HasModal_HiddenNotCounted(t *testing.T) {
	m := NewOverlayManager()
	o := newMockOverlay("modal", 50, true, 5, 3)
	o.SetVisible(false)
	m.Add(o)

	if m.HasModal() {
		t.Error("HasModal should return false for hidden modal overlay")
	}
}

// === HandleKey skips hidden overlays ===

func TestOverlayManager_HandleKey_SkipHidden(t *testing.T) {
	m := NewOverlayManager()
	o := newMockOverlay("hidden", 50, true, 5, 3)
	o.SetVisible(false)
	o.keyHit = true
	m.Add(o)

	if m.HandleKey(nil) {
		t.Error("HandleKey should skip hidden overlays")
	}
}

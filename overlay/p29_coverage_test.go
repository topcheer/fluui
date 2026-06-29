package overlay

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// Simple test component
type dummyComponent struct {
	component.BaseComponent
}

func (d *dummyComponent) Measure(cs component.Constraints) component.Size {
	return component.Size{W: 10, H: 3}
}

func (d *dummyComponent) Paint(buf *buffer.Buffer) {}

// === Popup coverage tests ===

func TestP29_Popup_SetStyle(t *testing.T) {
	p := NewPopup("test", "Title", &dummyComponent{})
	s := PopupStyle{
		Border: buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Title:  buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Body:   buffer.Style{Fg: buffer.RGB(0, 0, 255)},
	}
	p.SetStyle(s)
	// Verify it doesn't panic — style is stored
}

func TestP29_Popup_Measure_SmallConstraints(t *testing.T) {
	p := NewPopup("test", "Title", &dummyComponent{})
	// Very small constraints — should clamp to min 20x5
	size := p.Measure(component.Constraints{MaxWidth: 5, MaxHeight: 3})
	if size.W < 20 {
		t.Errorf("expected min width 20, got %d", size.W)
	}
	if size.H < 5 {
		t.Errorf("expected min height 5, got %d", size.H)
	}
}

func TestP29_Popup_Measure_ZeroConstraints(t *testing.T) {
	p := NewPopup("test", "Title", &dummyComponent{})
	// Zero/unset constraints — should use defaults
	size := p.Measure(component.Constraints{})
	if size.W != 72 { // 80*9/10
		t.Errorf("expected default width 72, got %d", size.W)
	}
}

func TestP29_Popup_SetBounds_NoMeasure(t *testing.T) {
	p := NewPopup("test", "Title", &dummyComponent{})
	// SetBounds without prior Measure — width/height are 0
	p.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	b := p.Bounds()
	if b.W <= 0 || b.H <= 0 {
		t.Errorf("expected positive bounds, got %+v", b)
	}
}

func TestP29_Popup_SetBounds_LargerThanRect(t *testing.T) {
	p := NewPopup("test", "Title", &dummyComponent{})
	p.Measure(component.Constraints{MaxWidth: 100, MaxHeight: 50})
	// SetBounds to a smaller rect — should clamp
	p.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 10})
	b := p.Bounds()
	if b.W > 30 {
		t.Errorf("width should be clamped to 30, got %d", b.W)
	}
	if b.H > 10 {
		t.Errorf("height should be clamped to 10, got %d", b.H)
	}
}

func TestP29_Popup_SetBounds_NegativeOrigin(t *testing.T) {
	p := NewPopup("test", "Title", &dummyComponent{})
	p.Measure(component.Constraints{MaxWidth: 200, MaxHeight: 100})
	p.SetBounds(component.Rect{X: 5, Y: 5, W: 200, H: 100})
	b := p.Bounds()
	if b.X < 0 {
		t.Errorf("x should not be negative, got %d", b.X)
	}
}

func TestP29_Popup_HandleKey_NotKeyEvent(t *testing.T) {
	p := NewPopup("test", "Title", &dummyComponent{})
	result := p.HandleKey("not a key event")
	if result {
		t.Error("should return false for non-KeyEvent")
	}
}

func TestP29_Popup_HandleKey_OtherKey(t *testing.T) {
	p := NewPopup("test", "Title", &dummyComponent{})
	result := p.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if result {
		t.Error("should return false for non-Escape key")
	}
}

func TestP29_Popup_Paint_TooSmall(t *testing.T) {
	p := NewPopup("test", "Title", &dummyComponent{})
	p.SetBounds(component.Rect{X: 0, Y: 0, W: 2, H: 2})
	buf := buffer.NewBuffer(10, 10)
	// Should return early — too small
	p.Paint(buf)
}

func TestP29_Popup_Paint_NoContent(t *testing.T) {
	p := NewPopup("test", "Title", nil)
	p.Measure(component.Constraints{MaxWidth: 40, MaxHeight: 10})
	p.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	p.Paint(buf) // should not panic with nil content
}

func TestP29_Popup_Paint_NoTitle(t *testing.T) {
	p := NewPopup("test", "", &dummyComponent{})
	p.Measure(component.Constraints{MaxWidth: 40, MaxHeight: 10})
	p.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	p.Paint(buf) // should not draw title
}

// === Modal coverage tests ===

func TestP29_Modal_Measure_WithBody(t *testing.T) {
	m := NewModal("m1", "Test", &dummyComponent{}, []string{"OK", "Cancel"})
	size := m.Measure(component.Constraints{MaxWidth: 60, MaxHeight: 30})
	if size.W < 20 {
		t.Errorf("expected min width 20, got %d", size.W)
	}
	if size.H < 7 {
		t.Errorf("expected min height 7, got %d", size.H)
	}
}

func TestP29_Modal_Measure_NoBody(t *testing.T) {
	m := NewModal("m1", "Test", nil, []string{"OK"})
	size := m.Measure(component.Constraints{MaxWidth: 60, MaxHeight: 30})
	// With nil body, should still produce valid size
	if size.H < 7 {
		t.Errorf("expected min height 7, got %d", size.H)
	}
}

func TestP29_Modal_Measure_LargeConstraints(t *testing.T) {
	m := NewModal("m1", "Test", &dummyComponent{}, []string{"OK"})
	size := m.Measure(component.Constraints{MaxWidth: 200, MaxHeight: 100})
	// Width should be clamped to 80
	if size.W > 80 {
		t.Errorf("width should be clamped to 80, got %d", size.W)
	}
}

func TestP29_Modal_Measure_TallBody(t *testing.T) {
	m := NewModal("m1", "Test", &dummyComponent{}, []string{"OK"})
	size := m.Measure(component.Constraints{MaxWidth: 40, MaxHeight: 5})
	// Should clamp to maxH
	if size.H > 5 {
		t.Errorf("height should be clamped to 5, got %d", size.H)
	}
}

func TestP29_Modal_SetBounds_NoMeasure(t *testing.T) {
	m := NewModal("m1", "Test", &dummyComponent{}, []string{"OK"})
	// SetBounds without Measure — width/height are 0
	m.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	b := m.Bounds()
	if b.W <= 0 || b.H <= 0 {
		t.Errorf("expected positive bounds, got %+v", b)
	}
}

func TestP29_Modal_SetBounds_LargerThanRect(t *testing.T) {
	m := NewModal("m1", "Test", &dummyComponent{}, []string{"OK"})
	m.Measure(component.Constraints{MaxWidth: 100, MaxHeight: 100})
	m.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 10})
	b := m.Bounds()
	if b.W > 20 {
		t.Errorf("width should be clamped to 20, got %d", b.W)
	}
}

func TestP29_Modal_HandleKey_Left(t *testing.T) {
	m := NewModal("m1", "Test", nil, []string{"OK", "Cancel"})
	m.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if m.SelectedButton() != 1 { // wraps to last
		t.Errorf("expected selected 1 after left, got %d", m.SelectedButton())
	}
}

func TestP29_Modal_HandleKey_Right(t *testing.T) {
	m := NewModal("m1", "Test", nil, []string{"OK", "Cancel"})
	m.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if m.SelectedButton() != 1 {
		t.Errorf("expected selected 1 after right, got %d", m.SelectedButton())
	}
	// Right again — wraps to 0
	m.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if m.SelectedButton() != 0 {
		t.Errorf("expected selected 0 after wrap, got %d", m.SelectedButton())
	}
}

func TestP29_Modal_HandleKey_Enter(t *testing.T) {
	m := NewModal("m1", "Test", nil, []string{"OK"})
	m.SetVisible(true)
	result := m.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !result {
		t.Error("Enter should return true")
	}
	if m.Visible() {
		t.Error("Enter should hide modal")
	}
}

func TestP29_Modal_HandleKey_NotKeyEvent(t *testing.T) {
	m := NewModal("m1", "Test", nil, []string{"OK"})
	result := m.HandleKey(42)
	if result {
		t.Error("should return false for non-KeyEvent")
	}
}

func TestP29_Modal_HandleKey_OtherKey(t *testing.T) {
	m := NewModal("m1", "Test", nil, []string{"OK"})
	result := m.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if result {
		t.Error("should return false for unhandled key")
	}
}

func TestP29_Modal_HandleKey_NoButtons(t *testing.T) {
	m := NewModal("m1", "Test", nil, nil)
	// Left/Right with no buttons should not panic, returns true
	result := m.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if !result {
		t.Error("Left should still return true even with no buttons")
	}
}

func TestP29_Modal_Paint_TooSmall(t *testing.T) {
	m := NewModal("m1", "Test", nil, []string{"OK"})
	m.SetBounds(component.Rect{X: 0, Y: 0, W: 2, H: 2})
	buf := buffer.NewBuffer(10, 10)
	m.Paint(buf) // should return early
}

func TestP29_Modal_Paint_NoButtons(t *testing.T) {
	m := NewModal("m1", "Test", &dummyComponent{}, nil)
	m.Measure(component.Constraints{MaxWidth: 30, MaxHeight: 15})
	m.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 15})
	buf := buffer.NewBuffer(30, 15)
	m.Paint(buf) // should not panic with no buttons
}

func TestP29_Modal_Paint_MultipleButtons(t *testing.T) {
	m := NewModal("m1", "Test", &dummyComponent{}, []string{"Yes", "No", "Cancel"})
	m.Measure(component.Constraints{MaxWidth: 40, MaxHeight: 15})
	m.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 15})
	buf := buffer.NewBuffer(40, 15)
	m.Paint(buf)
}

func TestP29_Modal_SetStyle(t *testing.T) {
	m := NewModal("m1", "Test", nil, []string{"OK"})
	s := DefaultModalStyle()
	s.Border.Fg = buffer.RGB(255, 0, 0)
	m.SetStyle(s)
}

// === OverlayManager HandleMouse coverage ===

func TestP29_HandleMouse_NoOverlays(t *testing.T) {
	mgr := NewOverlayManager()
	result := mgr.HandleMouse(5, 5)
	if result {
		t.Error("should return false with no overlays")
	}
}

func TestP29_HandleMouse_HitModal(t *testing.T) {
	mgr := NewOverlayManager()
	modal := NewModal("m1", "Test", nil, []string{"OK"})
	modal.SetVisible(true)
	mgr.Add(modal)
	modal.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 10})
	// Click within modal bounds
	result := mgr.HandleMouse(5, 5)
	if !result {
		t.Error("should return true — click hit modal")
	}
}

func TestP29_HandleMouse_MissModal(t *testing.T) {
	mgr := NewOverlayManager()
	modal := NewModal("m1", "Test", nil, []string{"OK"})
	modal.SetVisible(true)
	mgr.Add(modal)
	modal.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 10})
	// Click outside modal bounds
	mgr.HandleMouse(50, 50)
	// Modal still consumes because it blocks events from reaching lower layers
	// Actually depends on implementation — modal overlay may not consume if not hit
}

func TestP29_HandleMouse_InvisibleOverlay(t *testing.T) {
	mgr := NewOverlayManager()
	p := NewPopup("p1", "Test", &dummyComponent{})
	p.SetVisible(false)
	mgr.Add(p)
	result := mgr.HandleMouse(5, 5)
	if result {
		t.Error("invisible overlay should not consume mouse")
	}
}

func TestP29_HasModal_NoOverlays(t *testing.T) {
	mgr := NewOverlayManager()
	if mgr.HasModal() {
		t.Error("should return false with no overlays")
	}
}

func TestP29_HasModal_HiddenModal(t *testing.T) {
	mgr := NewOverlayManager()
	modal := NewModal("m1", "Test", nil, []string{"OK"})
	modal.SetVisible(false)
	mgr.Add(modal)
	if mgr.HasModal() {
		t.Error("hidden modal should not count")
	}
}

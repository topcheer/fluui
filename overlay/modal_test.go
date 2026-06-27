package overlay

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// stubBody is a minimal Component for testing Modal/Popup bodies.
type stubBody struct {
	text string
}

func (s *stubBody) ID() string                                  { return "stub" }
func (s *stubBody) Measure(cs component.Constraints) component.Size { return component.Size{W: 10, H: 3} }
func (s *stubBody) SetBounds(r component.Rect)                 {}
func (s *stubBody) Bounds() component.Rect                     { return component.Rect{} }
func (s *stubBody) Paint(buf *buffer.Buffer)                   {}
func (s *stubBody) Children() []component.Component            { return nil }

func TestModalCreation(t *testing.T) {
	body := &stubBody{text: "hello"}
	m := NewModal("test-modal", "Confirm", body, []string{"OK", "Cancel"})

	if m.ID() != "test-modal" {
		t.Errorf("ID() = %q, want 'test-modal'", m.ID())
	}
	if m.title != "Confirm" {
		t.Errorf("title = %q, want 'Confirm'", m.title)
	}
	if len(m.buttons) != 2 {
		t.Fatalf("len(buttons) = %d, want 2", len(m.buttons))
	}
	if m.buttons[0] != "OK" || m.buttons[1] != "Cancel" {
		t.Errorf("buttons = %v, want [OK, Cancel]", m.buttons)
	}
	if !m.Modal() {
		t.Error("Modal() = false, want true")
	}
	if !m.Visible() {
		t.Error("Visible() = false, want true (newly created)")
	}
	if m.Z() != 100 {
		t.Errorf("Z() = %d, want 100", m.Z())
	}
}

func TestModalMeasure(t *testing.T) {
	body := &stubBody{}
	m := NewModal("m1", "Title", body, []string{"OK"})

	sz := m.Measure(component.Bounded(80, 24))
	// Width: 80/2 = 40, within [20, 80]
	if sz.W != 40 {
		t.Errorf("width = %d, want 40", sz.W)
	}
	// Height: bodyH(3) + 3 = 6, min 7
	if sz.H != 7 {
		t.Errorf("height = %d, want 7 (body 3 + 3 border/buttons, min 7)", sz.H)
	}

	// Test clamping with small screen
	m2 := NewModal("m2", "T", body, []string{"OK"})
	sz2 := m2.Measure(component.Bounded(20, 10))
	if sz2.W != 20 {
		t.Errorf("small screen width = %d, want 20 (clamped)", sz2.W)
	}
	if sz2.H < 7 {
		t.Errorf("small screen height = %d, want >= 7", sz2.H)
	}
}

func TestModalPaint(t *testing.T) {
	body := &stubBody{}
	m := NewModal("paint-test", "My Modal", body, []string{"OK", "Cancel"})
	m.Measure(component.Bounded(80, 24))
	m.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	buf := buffer.NewBuffer(80, 24)
	m.Paint(buf)

	bounds := m.Bounds()

	// Check top-left corner ╭
	if c := buf.GetCell(bounds.X, bounds.Y); c.Rune != '\u256d' {
		t.Errorf("top-left corner: got %q, want ╭", c.Rune)
	}
	// Check top-right corner ╮
	if c := buf.GetCell(bounds.X+bounds.W-1, bounds.Y); c.Rune != '\u256e' {
		t.Errorf("top-right corner: got %q, want ╮", c.Rune)
	}
	// Check bottom-left corner ╰
	if c := buf.GetCell(bounds.X, bounds.Y+bounds.H-1); c.Rune != '\u2570' {
		t.Errorf("bottom-left corner: got %q, want ╰", c.Rune)
	}
	// Check bottom-right corner ╯
	if c := buf.GetCell(bounds.X+bounds.W-1, bounds.Y+bounds.H-1); c.Rune != '\u256f' {
		t.Errorf("bottom-right corner: got %q, want ╯", c.Rune)
	}

	// Check mask outside modal area (top-left corner of screen)
	if c := buf.GetCell(0, 0); c.Rune != ' ' && c.Bg.Equal(m.style.Mask.Bg) {
		// 0,0 is inside the modal if modal starts at 0,0 — check a point clearly outside
	}
	// Check a point definitely outside the modal (if modal doesn't fill screen)
	outsideX := bounds.X + bounds.W
	outsideY := bounds.Y + bounds.H
	if outsideX < 80 && outsideY < 24 {
		if c := buf.GetCell(outsideX, outsideY); c.Bg.Equal(m.style.Mask.Bg) {
			// mask is drawn — good
		} else {
			t.Error("expected mask cell outside modal bounds")
		}
	}

	// Check button bar — "OK" button should be rendered
	btnY := bounds.Y + bounds.H - 2
	foundOK := false
	foundCancel := false
	for x := bounds.X; x < bounds.X+bounds.W; x++ {
		c := buf.GetCell(x, btnY)
		// Look for button text characters
		s := string(c.Rune)
		if s == "O" {
			foundOK = true
		}
		if s == "C" {
			foundCancel = true
		}
	}
	if !foundOK {
		t.Error("OK button not found in button bar")
	}
	if !foundCancel {
		t.Error("Cancel button not found in button bar")
	}
}

func TestModalEscClose(t *testing.T) {
	m := NewModal("esc-test", "T", &stubBody{}, []string{"OK", "Cancel"})
	if !m.Visible() {
		t.Fatal("modal should start visible")
	}

	consumed := m.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Error("HandleKey(Escape) should return true")
	}
	if m.Visible() {
		t.Error("modal should be hidden after Esc")
	}
}

func TestModalEnterClose(t *testing.T) {
	m := NewModal("enter-test", "T", &stubBody{}, []string{"OK", "Cancel"})
	if !m.Visible() {
		t.Fatal("modal should start visible")
	}

	consumed := m.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !consumed {
		t.Error("HandleKey(Enter) should return true")
	}
	if m.Visible() {
		t.Error("modal should be hidden after Enter")
	}
}

func TestModalButtonNavigation(t *testing.T) {
	m := NewModal("nav-test", "T", &stubBody{}, []string{"OK", "Cancel", "Help"})
	if m.SelectedButton() != 0 {
		t.Errorf("initial selected = %d, want 0", m.SelectedButton())
	}

	// Right → select Cancel (index 1)
	m.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if m.SelectedButton() != 1 {
		t.Errorf("after Right: selected = %d, want 1", m.SelectedButton())
	}

	// Right → select Help (index 2)
	m.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if m.SelectedButton() != 2 {
		t.Errorf("after Right x2: selected = %d, want 2", m.SelectedButton())
	}

	// Right → wrap to OK (index 0)
	m.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if m.SelectedButton() != 0 {
		t.Errorf("after Right x3: selected = %d, want 0 (wrap)", m.SelectedButton())
	}

	// Left → wrap to Help (index 2)
	m.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if m.SelectedButton() != 2 {
		t.Errorf("after Left from 0: selected = %d, want 2 (wrap)", m.SelectedButton())
	}
}

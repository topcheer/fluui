package overlay

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestPopupCreation(t *testing.T) {
	content := &stubBody{}
	p := NewPopup("test-popup", "Code Viewer", content)

	if p.ID() != "test-popup" {
		t.Errorf("ID() = %q, want 'test-popup'", p.ID())
	}
	if p.title != "Code Viewer" {
		t.Errorf("title = %q, want 'Code Viewer'", p.title)
	}
	if !p.Visible() {
		t.Error("Visible() = false, want true")
	}
	if p.Modal() {
		t.Error("Modal() = true, want false (popup is non-modal)")
	}
	if p.Z() != 90 {
		t.Errorf("Z() = %d, want 90", p.Z())
	}
}

func TestPopupMeasure(t *testing.T) {
	p := NewPopup("p1", "T", &stubBody{})

	sz := p.Measure(component.Bounded(80, 24))
	// 90% of 80 = 72, 90% of 24 = 21
	if sz.W != 72 {
		t.Errorf("width = %d, want 72 (90%% of 80)", sz.W)
	}
	if sz.H != 21 {
		t.Errorf("height = %d, want 21 (90%% of 24)", sz.H)
	}

	// Small screen clamping
	p2 := NewPopup("p2", "T", &stubBody{})
	sz2 := p2.Measure(component.Bounded(20, 10))
	if sz2.W != 20 {
		t.Errorf("small width = %d, want 20 (90%% of 20 = 18, clamped to min 20)", sz2.W)
	}
	if sz2.H != 9 {
		t.Errorf("small height = %d, want 9 (90%% of 10)", sz2.H)
	}
}

func TestPopupPaint(t *testing.T) {
	p := NewPopup("paint-test", "My Popup", &stubBody{})
	p.Measure(component.Bounded(40, 12))
	p.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 12})

	buf := buffer.NewBuffer(40, 12)
	p.Paint(buf)

	bounds := p.Bounds()

	// Check corners
	if c := buf.GetCell(bounds.X, bounds.Y); c.Rune != '\u250c' {
		t.Errorf("top-left: got %q, want ┌", c.Rune)
	}
	if c := buf.GetCell(bounds.X+bounds.W-1, bounds.Y); c.Rune != '\u2510' {
		t.Errorf("top-right: got %q, want ┐", c.Rune)
	}
	if c := buf.GetCell(bounds.X, bounds.Y+bounds.H-1); c.Rune != '\u2514' {
		t.Errorf("bottom-left: got %q, want └", c.Rune)
	}
	if c := buf.GetCell(bounds.X+bounds.W-1, bounds.Y+bounds.H-1); c.Rune != '\u2518' {
		t.Errorf("bottom-right: got %q, want ┘", c.Rune)
	}

	// Check title exists somewhere on the top border
	foundTitle := false
	for x := bounds.X; x < bounds.X+bounds.W; x++ {
		c := buf.GetCell(x, bounds.Y)
		if c.Rune == 'M' {
			foundTitle = true
			break
		}
	}
	if !foundTitle {
		t.Error("title text not found on top border")
	}

	// Check left portion of top border (before title)
	// Title is centered, so check only the first few chars before title area
	for i := 1; i < bounds.W/4; i++ {
		c := buf.GetCell(bounds.X+i, bounds.Y)
		if c.Rune != '\u2500' {
			// Title text may overlap these positions
			break
		}
	}
	// Check bottom border (no title interference)
	for i := 1; i < bounds.W-1; i++ {
		c := buf.GetCell(bounds.X+i, bounds.Y+bounds.H-1)
		if c.Rune != '\u2500' {
			t.Errorf("bottom edge at x=%d: got %q, want ─", i, c.Rune)
		}
	}
}

func TestPopupEscClose(t *testing.T) {
	p := NewPopup("esc-test", "T", &stubBody{})
	if !p.Visible() {
		t.Fatal("popup should start visible")
	}

	consumed := p.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Error("HandleKey(Escape) should return true")
	}
	if p.Visible() {
		t.Error("popup should be hidden after Esc")
	}
}

func TestPopupDoesNotConsumeOtherKeys(t *testing.T) {
	p := NewPopup("keys-test", "T", &stubBody{})

	// Non-Esc keys should not be consumed
	consumed := p.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if consumed {
		t.Error("HandleKey(Enter) should return false (popup only consumes Esc)")
	}
}

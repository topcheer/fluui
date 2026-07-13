package component

import (
	"sync/atomic"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestButton_New(t *testing.T) {
	b := NewButton("OK")
	if b.Label() != "OK" {
		t.Errorf("expected 'OK', got %q", b.Label())
	}
	if !b.Enabled() {
		t.Error("expected enabled by default")
	}
	if b.IsActive() {
		t.Error("expected inactive by default")
	}
}

func TestButton_NewWithVariant(t *testing.T) {
	b := NewButtonWithVariant("Submit", ButtonPrimary)
	if b.Variant() != ButtonPrimary {
		t.Errorf("expected primary, got %v", b.Variant())
	}
}

func TestButton_SetLabel(t *testing.T) {
	b := NewButton("Old")
	b.SetLabel("New")
	if b.Label() != "New" {
		t.Errorf("expected 'New', got %q", b.Label())
	}
}

func TestButton_SetVariant(t *testing.T) {
	b := NewButton("Test")
	b.SetVariant(ButtonDanger)
	if b.Variant() != ButtonDanger {
		t.Errorf("expected danger, got %v", b.Variant())
	}
}

func TestButton_SetActive(t *testing.T) {
	b := NewButton("Test")
	b.SetActive(true)
	if !b.IsActive() {
		t.Error("expected active")
	}
	b.SetActive(false)
	if b.IsActive() {
		t.Error("expected inactive")
	}
}

func TestButton_SetEnabled(t *testing.T) {
	b := NewButton("Test")
	b.SetEnabled(false)
	if b.Enabled() {
		t.Error("expected disabled")
	}
	b.SetEnabled(true)
	if !b.Enabled() {
		t.Error("expected enabled")
	}
}

func TestButton_SetOnClick(t *testing.T) {
	b := NewButton("Test")
	called := false
	b.SetOnClick(func() { called = true })
	b.Click()
	if !called {
		t.Error("expected onClick to be called")
	}
}

func TestButton_ClickDisabled(t *testing.T) {
	b := NewButton("Test")
	called := false
	b.SetOnClick(func() { called = true })
	b.SetEnabled(false)
	b.Click()
	if called {
		t.Error("expected onClick NOT to be called when disabled")
	}
}

func TestButton_ClickNoHandler(t *testing.T) {
	b := NewButton("Test")
	b.Click() // should not panic
}

func TestButton_HandleKeyEnter(t *testing.T) {
	b := NewButton("Test")
	called := false
	b.SetOnClick(func() { called = true })
	handled := b.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !handled {
		t.Error("expected Enter to be handled")
	}
	if !called {
		t.Error("expected onClick called")
	}
}

func TestButton_HandleKeySpace(t *testing.T) {
	b := NewButton("Test")
	called := false
	b.SetOnClick(func() { called = true })
	handled := b.HandleKey(&term.KeyEvent{Rune: ' '})
	if !handled {
		t.Error("expected Space to be handled")
	}
	if !called {
		t.Error("expected onClick called")
	}
}

func TestButton_HandleKeyDisabled(t *testing.T) {
	b := NewButton("Test")
	b.SetEnabled(false)
	handled := b.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if handled {
		t.Error("expected false when disabled")
	}
}

func TestButton_HandleKeyUnknown(t *testing.T) {
	b := NewButton("Test")
	handled := b.HandleKey(&term.KeyEvent{Rune: 'x'})
	if handled {
		t.Error("expected false for unknown key")
	}
}

func TestButton_HandleKeyNil(t *testing.T) {
	b := NewButton("Test")
	if b.HandleKey(nil) {
		t.Error("expected false for nil key")
	}
}

func TestButton_HandleMouseClick(t *testing.T) {
	b := NewButton("Test")
	b.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	called := false
	b.SetOnClick(func() { called = true })
	handled := b.HandleMouse(5, 0, "click")
	if !handled {
		t.Error("expected mouse click to be handled")
	}
	if !called {
		t.Error("expected onClick called")
	}
}

func TestButton_HandleMouseOutside(t *testing.T) {
	b := NewButton("Test")
	b.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	handled := b.HandleMouse(20, 0, "click")
	if handled {
		t.Error("expected false for click outside bounds")
	}
}

func TestButton_HandleMouseDisabled(t *testing.T) {
	b := NewButton("Test")
	b.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	b.SetEnabled(false)
	handled := b.HandleMouse(5, 0, "click")
	if handled {
		t.Error("expected false when disabled")
	}
}

func TestButton_Measure(t *testing.T) {
	b := NewButton("OK")
	s := b.Measure(Constraints{MaxWidth: 50, MaxHeight: 10})
	if s.H != 1 {
		t.Errorf("expected height 1, got %d", s.H)
	}
	// "OK" = 2 chars + 4 padding = 6
	if s.W != 6 {
		t.Errorf("expected width 6, got %d", s.W)
	}
}

func TestButton_MeasureShortLabel(t *testing.T) {
	b := NewButton("X")
	s := b.Measure(Constraints{MaxWidth: 50, MaxHeight: 10})
	// min width is 6
	if s.W < 6 {
		t.Errorf("expected min width 6, got %d", s.W)
	}
}

func TestButton_MeasureClamped(t *testing.T) {
	b := NewButton("Very Long Button Label")
	s := b.Measure(Constraints{MaxWidth: 10, MaxHeight: 10})
	if s.W > 10 {
		t.Errorf("expected width <= 10, got %d", s.W)
	}
}

func TestButton_Paint(t *testing.T) {
	b := NewButton("OK")
	b.SetBounds(Rect{X: 0, Y: 0, W: 6, H: 1})
	buf := buffer.NewBuffer(6, 1)
	b.Paint(buf)
	// Content is centered: " OK " starts at position 1
	if buf.GetCell(0, 0).Rune != ' ' {
		t.Error("expected padding at start")
	}
	if buf.GetCell(2, 0).Rune != 'O' {
		t.Error("expected 'O' at position 2 (centered)")
	}
	if buf.GetCell(3, 0).Rune != 'K' {
		t.Error("expected 'K' at position 3 (centered)")
	}
}

func TestButton_PaintActive(t *testing.T) {
	b := NewButton("Test")
	b.SetActive(true)
	b.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	b.Paint(buf) // should not panic
}

func TestButton_PaintDisabled(t *testing.T) {
	b := NewButton("Test")
	b.SetEnabled(false)
	b.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	b.Paint(buf)
}

func TestButton_PaintNilBuffer(t *testing.T) {
	b := NewButton("Test")
	b.Paint(nil)
}

func TestButton_PaintZeroWidth(t *testing.T) {
	b := NewButton("Test")
	b.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 1})
	b.Paint(buffer.NewBuffer(1, 1))
}

func TestButton_PaintLongLabel(t *testing.T) {
	b := NewButton("Very Long Label That Exceeds Width")
	b.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})
	b.Paint(buffer.NewBuffer(5, 1)) // should truncate
}

func TestButton_PaintWithVariant(t *testing.T) {
	variants := []ButtonVariant{ButtonDefault, ButtonPrimary, ButtonSuccess, ButtonWarning, ButtonDanger}
	for _, v := range variants {
		b := NewButtonWithVariant("Test", v)
		b.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
		b.Paint(buffer.NewBuffer(10, 1))
	}
}

func TestButton_Children(t *testing.T) {
	b := NewButton("Test")
	if b.Children() != nil {
		t.Error("expected nil children")
	}
}

func TestButton_VariantString(t *testing.T) {
	tests := []struct {
		v    ButtonVariant
		want string
	}{
		{ButtonDefault, "default"},
		{ButtonPrimary, "primary"},
		{ButtonSuccess, "success"},
		{ButtonWarning, "warning"},
		{ButtonDanger, "danger"},
	}
	for _, tt := range tests {
		if got := tt.v.String(); got != tt.want {
			t.Errorf("variant %d: expected %q, got %q", tt.v, tt.want, got)
		}
	}
}

func TestButton_SetStyle(t *testing.T) {
	b := NewButton("Test")
	b.SetStyle(DefaultButtonStyle())
	b.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	b.Paint(buffer.NewBuffer(10, 1))
}

func TestButton_Concurrent(t *testing.T) {
	b := NewButton("Test")
	var called atomic.Int32
	b.SetOnClick(func() { called.Add(1) })
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			b.SetLabel("Test")
			b.Click()
		}
		close(done)
	}()
	for i := 0; i < 50; i++ {
		b.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
		b.Paint(buffer.NewBuffer(10, 1))
	}
	<-done
	if called.Load() != 50 {
		t.Errorf("expected 50 clicks, got %d", called.Load())
	}
}

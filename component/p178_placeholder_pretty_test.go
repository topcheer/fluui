package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === Placeholder tests ===

func TestPlaceholder_New(t *testing.T) {
	p := NewPlaceholder("Main Panel")
	if p.Label() != "Main Panel" {
		t.Errorf("expected 'Main Panel', got %q", p.Label())
	}
}

func TestPlaceholder_SetLabel(t *testing.T) {
	p := NewPlaceholder("")
	p.SetLabel("New Label")
	if p.Label() != "New Label" {
		t.Errorf("expected 'New Label', got %q", p.Label())
	}
}

func TestPlaceholder_Measure(t *testing.T) {
	p := NewPlaceholder("Test")
	s := p.Measure(Constraints{MaxWidth: 80, MaxHeight: 10})
	if s.H != 3 {
		t.Errorf("expected height 3, got %d", s.H)
	}
	if s.W < 10 {
		t.Errorf("expected width >= 10, got %d", s.W)
	}
}

func TestPlaceholder_MeasureShortLabel(t *testing.T) {
	p := NewPlaceholder("X")
	s := p.Measure(Constraints{MaxWidth: 80, MaxHeight: 10})
	if s.W < 10 {
		t.Errorf("expected min width 10, got %d", s.W)
	}
}

func TestPlaceholder_Paint(t *testing.T) {
	p := NewPlaceholder("Content")
	p.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	p.Paint(buf)
	// Top-left corner
	if buf.GetCell(0, 0).Rune != '┌' {
		t.Errorf("expected corner, got %c", buf.GetCell(0, 0).Rune)
	}
	// Bottom-right corner
	if buf.GetCell(19, 2).Rune != '┘' {
		t.Errorf("expected corner, got %c", buf.GetCell(19, 2).Rune)
	}
}

func TestPlaceholder_PaintNilBuffer(t *testing.T) {
	p := NewPlaceholder("Test")
	p.Paint(nil)
}

func TestPlaceholder_PaintZeroBounds(t *testing.T) {
	p := NewPlaceholder("Test")
	p.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	p.Paint(buffer.NewBuffer(1, 1))
}

func TestPlaceholder_PaintLongLabel(t *testing.T) {
	p := NewPlaceholder("This is a very long label that exceeds the width")
	p.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	buf := buffer.NewBuffer(10, 3)
	p.Paint(buf) // should truncate
}

func TestPlaceholder_HandleKey(t *testing.T) {
	p := NewPlaceholder("Test")
	if p.HandleKey(&term.KeyEvent{Key: term.KeyEnter}) {
		t.Error("expected false")
	}
}

func TestPlaceholder_Children(t *testing.T) {
	p := NewPlaceholder("Test")
	if p.Children() != nil {
		t.Error("expected nil children")
	}
}

func TestPlaceholder_SetStyle(t *testing.T) {
	p := NewPlaceholder("Test")
	p.SetStyle(DefaultPlaceholderStyle())
	p.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	p.Paint(buffer.NewBuffer(20, 3))
}

func TestPlaceholder_Concurrent(t *testing.T) {
	p := NewPlaceholder("Test")
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			p.SetLabel("Test")
		}
		close(done)
	}()
	for i := 0; i < 50; i++ {
		p.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
		p.Paint(buffer.NewBuffer(20, 3))
	}
	<-done
}

// === Pretty tests ===

func TestPretty_New(t *testing.T) {
	p := NewPretty(map[string]any{"key": "value"})
	if p.data == "" {
		t.Error("expected non-empty data")
	}
}

func TestPretty_NewString(t *testing.T) {
	p := NewPrettyString(`{"name": "test", "count": 42}`)
	if p.data == "" {
		t.Error("expected non-empty data")
	}
}

func TestPretty_NewNil(t *testing.T) {
	p := NewPretty(nil)
	if p.data != "nil" {
		t.Errorf("expected 'nil', got %q", p.data)
	}
}

func TestPretty_Measure(t *testing.T) {
	p := NewPrettyString("line1\nline2\nline3")
	s := p.Measure(Constraints{MaxWidth: 40, MaxHeight: 10})
	if s.H != 3 {
		t.Errorf("expected height 3, got %d", s.H)
	}
}

func TestPretty_Paint(t *testing.T) {
	p := NewPrettyString(`{"key": "value", "num": 42}`)
	p.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	p.Paint(buf)
	// Should have non-space at 0,0
	if buf.GetCell(0, 0).Rune == ' ' {
		t.Error("expected non-space at 0,0")
	}
}

func TestPretty_PaintNilBuffer(t *testing.T) {
	p := NewPrettyString("test")
	p.Paint(nil)
}

func TestPretty_PaintZeroWidth(t *testing.T) {
	p := NewPrettyString("test")
	p.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 5})
	p.Paint(buffer.NewBuffer(1, 5))
}

func TestPretty_PaintMultiline(t *testing.T) {
	p := NewPrettyString("line1\nline2\nline3")
	p.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	p.Paint(buf)
	// Each line should have content
	if buf.GetCell(0, 0).Rune != 'l' {
		t.Error("expected 'l' at 0,0")
	}
	if buf.GetCell(0, 1).Rune != 'l' {
		t.Error("expected 'l' at 0,1")
	}
}

func TestPretty_HandleKey(t *testing.T) {
	p := NewPrettyString("test")
	if p.HandleKey(&term.KeyEvent{Key: term.KeyEnter}) {
		t.Error("expected false")
	}
}

func TestPretty_Children(t *testing.T) {
	p := NewPrettyString("test")
	if p.Children() != nil {
		t.Error("expected nil children")
	}
}

func TestPretty_SetStyle(t *testing.T) {
	p := NewPrettyString(`{"key": "value"}`)
	p.SetStyle(DefaultPrettyStyle())
	p.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	p.Paint(buffer.NewBuffer(40, 5))
}

func TestPretty_Concurrent(t *testing.T) {
	p := NewPrettyString("test\ndata")
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			p.SetStyle(DefaultPrettyStyle())
		}
		close(done)
	}()
	for i := 0; i < 50; i++ {
		p.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
		p.Paint(buffer.NewBuffer(40, 5))
	}
	<-done
}

func TestPretty_WithStruct(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}
	p := NewPretty(Person{Name: "Alice", Age: 30})
	if p.data == "" {
		t.Error("expected non-empty data")
	}
}

func TestPretty_WithSlice(t *testing.T) {
	p := NewPretty([]int{1, 2, 3, 4, 5})
	if p.data == "" {
		t.Error("expected non-empty data")
	}
}

func TestPretty_WithString(t *testing.T) {
	p := NewPretty("hello world")
	if p.data == "" {
		t.Error("expected non-empty data")
	}
}

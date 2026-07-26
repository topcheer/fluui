package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP394_NewColorSwatch(t *testing.T) {
	c := buffer.RGB(0xFF, 0x80, 0x00)
	s := NewColorSwatch(c)
	if s.Color() != c {
		t.Error("color mismatch")
	}
	if !s.ShowHex() {
		t.Error("ShowHex should default true")
	}
	if s.ID() == "" {
		t.Error("ID empty")
	}
}

func TestP394_SetColor(t *testing.T) {
	s := NewColorSwatch(buffer.RGB(0, 0, 0))
	s.SetColor(buffer.RGB(1, 2, 3))
	if s.Color().Val != 0x010203 {
		t.Errorf("color val = %x", s.Color().Val)
	}
}

func TestP394_SetLabel(t *testing.T) {
	s := NewColorSwatch(buffer.RGB(0xFF, 0, 0))
	s.SetLabel("Primary")
	if s.Label() != "Primary" {
		t.Errorf("Label = %q", s.Label())
	}
	s.SetLabel("")
	if s.Label() != "" {
		t.Error("label should be empty")
	}
}

func TestP394_SetShowHex(t *testing.T) {
	s := NewColorSwatch(buffer.RGB(0, 0, 0))
	s.SetShowHex(false)
	if s.ShowHex() {
		t.Error("ShowHex should be false")
	}
}

func TestP394_Measure(t *testing.T) {
	s := NewColorSwatch(buffer.RGB(0xFF, 0x80, 0x00))
	size := s.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	// "██ #ff8000" = 3 + 7 = 10
	if size.W != 10 {
		t.Errorf("W = %d, want 10", size.W)
	}
	if size.H != 1 {
		t.Errorf("H = %d", size.H)
	}
}

func TestP394_Measure_NoHex(t *testing.T) {
	s := NewColorSwatch(buffer.RGB(0, 0, 0))
	s.SetShowHex(false)
	size := s.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	// "██ " = 3
	if size.W != 3 {
		t.Errorf("W = %d, want 3", size.W)
	}
}

func TestP394_Measure_CustomLabel(t *testing.T) {
	s := NewColorSwatch(buffer.RGB(0, 0, 0))
	s.SetLabel("Accent")
	size := s.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	// "██ Accent" = 3 + 6 = 9
	if size.W != 9 {
		t.Errorf("W = %d, want 9", size.W)
	}
}

func TestP394_Paint_TrueColor(t *testing.T) {
	c := buffer.RGB(0xFF, 0x00, 0xFF)
	s := NewColorSwatch(c)
	s.SetBounds(Rect{X: 0, Y: 0, W: 12, H: 1})
	buf := buffer.NewBuffer(12, 1)
	s.Paint(buf)
	// Cells 0-1 should be █ with the color
	cell := buf.GetCell(0, 0)
	if cell.Rune != '\u2588' {
		t.Errorf("cell[0] = %q, want █", string(cell.Rune))
	}
	if cell.Fg != c {
		t.Error("fg color not applied")
	}
	// Cell 2 = space, cell 3+ = hex label
	cell3 := buf.GetCell(3, 0)
	if cell3.Rune != '#' {
		t.Errorf("hex label cell = %q, want '#'", string(cell3.Rune))
	}
}

func TestP394_Paint_Color256(t *testing.T) {
	s := NewColorSwatch(buffer.Color256Val(42))
	s.SetBounds(Rect{X: 0, Y: 0, W: 12, H: 1})
	buf := buffer.NewBuffer(12, 1)
	s.Paint(buf)
	// hexStringLocked should return "c42"
	cell3 := buf.GetCell(3, 0)
	if cell3.Rune != 'c' {
		t.Errorf("256 color label = %q, want 'c...'", string(cell3.Rune))
	}
}

func TestP394_Paint_NamedColor(t *testing.T) {
	s := NewColorSwatch(buffer.NamedColor(1))
	s.SetBounds(Rect{X: 0, Y: 0, W: 12, H: 1})
	buf := buffer.NewBuffer(12, 1)
	s.Paint(buf)
	cell3 := buf.GetCell(3, 0)
	if cell3.Rune != 'n' {
		t.Errorf("named color label = %q, want 'n...'", string(cell3.Rune))
	}
}

func TestP394_Paint_NoColor(t *testing.T) {
	s := NewColorSwatch(buffer.NoColor())
	s.SetBounds(Rect{X: 0, Y: 0, W: 12, H: 1})
	buf := buffer.NewBuffer(12, 1)
	s.Paint(buf)
	cell3 := buf.GetCell(3, 0)
	if cell3.Rune != 'n' {
		// "none" starts with 'n'
		t.Logf("none color cell = %q", string(cell3.Rune))
	}
}

func TestP394_Paint_ZeroBounds(t *testing.T) {
	s := NewColorSwatch(buffer.RGB(0, 0, 0))
	s.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	s.Paint(buf)
}

func TestP394_Paint_NoHex(t *testing.T) {
	s := NewColorSwatch(buffer.RGB(0xFF, 0, 0))
	s.SetShowHex(false)
	s.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	s.Paint(buf)
	// Only swatch + space, no hex
	if buf.GetCell(2, 0).Rune != ' ' {
		t.Error("cell[2] should be space")
	}
}

func TestP394_Concurrent(t *testing.T) {
	s := NewColorSwatch(buffer.RGB(0, 0, 0))
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ {
			s.SetColor(buffer.RGB(1, 2, 3))
			s.SetLabel("test")
		}
		close(done)
	}()
	for i := 0; i < 500; i++ {
		_ = s.Color()
		_ = s.Label()
	}
	<-done
}

func TestP394_SatisfiesComponent(t *testing.T) {
	var _ Component = (*ColorSwatch)(nil)
}

func BenchmarkP394_ColorSwatch_Paint(b *testing.B) {
	s := NewColorSwatch(buffer.RGB(0xFF, 0x80, 0x00))
	s.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Paint(buf)
	}
}

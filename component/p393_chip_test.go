package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P393: Chip component tests

func TestP393_NewChip(t *testing.T) {
	c := NewChip("gpt-4")
	if c.Text() != "gpt-4" {
		t.Errorf("Text = %q", c.Text())
	}
	if c.ID() == "" {
		t.Error("ID should not be empty")
	}
	if c.Variant() != ChipFilled {
		t.Errorf("default variant = %v, want ChipFilled", c.Variant())
	}
}

func TestP393_SetText(t *testing.T) {
	c := NewChip("old")
	c.SetText("new")
	if c.Text() != "new" {
		t.Errorf("Text = %q", c.Text())
	}
}

func TestP393_SetIcon(t *testing.T) {
	c := NewChip("python")
	c.SetIcon("🐍")
	if c.Icon() != "🐍" {
		t.Errorf("Icon = %q", c.Icon())
	}
	c.SetIcon("")
	if c.Icon() != "" {
		t.Error("Icon should be empty")
	}
}

func TestP393_SetVariant(t *testing.T) {
	c := NewChip("test")
	c.SetVariant(ChipOutlined)
	if c.Variant() != ChipOutlined {
		t.Errorf("Variant = %v", c.Variant())
	}
	c.SetVariant(ChipSubtle)
	if c.Variant() != ChipSubtle {
		t.Errorf("Variant = %v", c.Variant())
	}
}

func TestP393_SetColors(t *testing.T) {
	c := NewChip("custom")
	fg := buffer.RGB(0xFF, 0x00, 0x00)
	bg := buffer.RGB(0x00, 0xFF, 0x00)
	c.SetColors(fg, bg)
	c.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	c.Paint(buf)
	cell := buf.GetCell(1, 0)
	if cell.Fg != fg {
		t.Error("custom fg not applied")
	}
	if cell.Bg != bg {
		t.Error("custom bg not applied")
	}
}

func TestP393_Measure(t *testing.T) {
	c := NewChip("hello")
	s := c.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	// " hello " = 7 (1 pad + 5 text + 1 pad)
	if s.W != 7 {
		t.Errorf("W = %d, want 7", s.W)
	}
	if s.H != 1 {
		t.Errorf("H = %d, want 1", s.H)
	}
}

func TestP393_Measure_WithIcon(t *testing.T) {
	c := NewChip("py")
	c.SetIcon("🐍")
	s := c.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	// " 🐍 py " = 7 (1 pad + 1 icon + 1 space + 2 text + 1 pad)
	if s.W != 7 {
		t.Errorf("WithIcon W = %d, want 7", s.W)
	}
}

func TestP393_Measure_Clamp(t *testing.T) {
	c := NewChip("hello")
	s := c.Measure(Constraints{MaxWidth: 3, MaxHeight: 1})
	if s.W != 3 {
		t.Errorf("Clamped W = %d, want 3", s.W)
	}
}

func TestP393_Paint_Filled(t *testing.T) {
	c := NewChip("AI")
	c.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})
	buf := buffer.NewBuffer(5, 1)
	c.Paint(buf)
	// Cell 0 = padding space, Cell 1 = 'A', Cell 2 = 'I', Cell 3 = padding
	c0 := buf.GetCell(0, 0)
	if c0.Rune != ' ' {
		t.Errorf("cell[0] = %q, want ' '", string(c0.Rune))
	}
	c1 := buf.GetCell(1, 0)
	if c1.Rune != 'A' {
		t.Errorf("cell[1] = %q, want 'A'", string(c1.Rune))
	}
	c2 := buf.GetCell(2, 0)
	if c2.Rune != 'I' {
		t.Errorf("cell[2] = %q, want 'I'", string(c2.Rune))
	}
	c3 := buf.GetCell(3, 0)
	if c3.Rune != ' ' {
		t.Errorf("cell[3] = %q, want ' '", string(c3.Rune))
	}
}

func TestP393_Paint_Outlined(t *testing.T) {
	c := NewChip("tag")
	c.SetVariant(ChipOutlined)
	c.SetBounds(Rect{X: 0, Y: 0, W: 6, H: 1})
	buf := buffer.NewBuffer(6, 1)
	c.Paint(buf)
	// First cell should be '['
	c0 := buf.GetCell(0, 0)
	if c0.Rune != '[' {
		t.Errorf("outlined cell[0] = %q, want '['", string(c0.Rune))
	}
	// Last cell should be ']'
	c5 := buf.GetCell(5, 0)
	if c5.Rune != ']' {
		t.Errorf("outlined cell[5] = %q, want ']'", string(c5.Rune))
	}
}

func TestP393_Paint_Subtle(t *testing.T) {
	c := NewChip("muted")
	c.SetVariant(ChipSubtle)
	c.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 1})
	buf := buffer.NewBuffer(8, 1)
	c.Paint(buf)
	// Should render with muted bg
	c0 := buf.GetCell(0, 0)
	if c0.Rune != ' ' {
		t.Errorf("subtle cell[0] = %q, want ' '", string(c0.Rune))
	}
}

func TestP393_Paint_WithIcon(t *testing.T) {
	c := NewChip("py")
	c.SetIcon("★")
	c.SetBounds(Rect{X: 0, Y: 0, W: 7, H: 1})
	buf := buffer.NewBuffer(7, 1)
	c.Paint(buf)
	c1 := buf.GetCell(1, 0)
	if c1.Rune != '★' {
		t.Errorf("icon cell[1] = %q, want '★'", string(c1.Rune))
	}
}

func TestP393_Paint_ZeroBounds(t *testing.T) {
	c := NewChip("test")
	c.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	c.Paint(buf) // should not panic
}

func TestP393_Paint_NarrowWidth(t *testing.T) {
	c := NewChip("long text here")
	c.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	c.Paint(buf) // should clip without panic
}

func TestP393_Paint_NonZeroOffset(t *testing.T) {
	c := NewChip("AI")
	c.SetBounds(Rect{X: 5, Y: 2, W: 5, H: 1})
	buf := buffer.NewBuffer(15, 5)
	c.Paint(buf)
	cell := buf.GetCell(6, 2)
	if cell.Rune != 'A' {
		t.Errorf("offset cell = %q, want 'A'", string(cell.Rune))
	}
}

func TestP393_ChipCountLabel(t *testing.T) {
	tests := []struct {
		label string
		count int
		want  string
	}{
		{"python", 12, "python (12)"},
		{"go", 0, "go"},
		{"rust", 1, "rust (1)"},
		{"", 5, " (5)"},
	}
	for _, tt := range tests {
		got := chipCountLabel(tt.label, tt.count)
		if got != tt.want {
			t.Errorf("chipCountLabel(%q, %d) = %q, want %q", tt.label, tt.count, got, tt.want)
		}
	}
}

func TestP393_Concurrent(t *testing.T) {
	c := NewChip("test")
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ {
			c.SetText("concurrent")
			c.SetVariant(ChipOutlined)
		}
		close(done)
	}()
	for i := 0; i < 500; i++ {
		_ = c.Text()
		_ = c.Variant()
	}
	<-done
}

func TestP393_SatisfiesComponent(t *testing.T) {
	var _ Component = (*Chip)(nil)
}

// Benchmarks
func BenchmarkP393_Chip_Paint_Filled(b *testing.B) {
	c := NewChip("gpt-4")
	c.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 1})
	buf := buffer.NewBuffer(8, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkP393_Chip_Paint_WithIcon(b *testing.B) {
	c := NewChip("python")
	c.SetIcon("🐍")
	c.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

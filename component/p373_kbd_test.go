package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P373: KBD component tests

func TestP373_NewKBD(t *testing.T) {
	k := NewKBD("Ctrl+C")
	if k.Text() != "Ctrl+C" {
		t.Errorf("Text = %q", k.Text())
	}
	if k.ID() == "" {
		t.Error("ID should not be empty")
	}
	if k.Variant() != KBDInverse {
		t.Errorf("default variant = %v, want KBDInverse", k.Variant())
	}
}

func TestP373_SetText(t *testing.T) {
	k := NewKBD("Enter")
	k.SetText("Escape")
	if k.Text() != "Escape" {
		t.Errorf("Text = %q", k.Text())
	}
}

func TestP373_SetVariant(t *testing.T) {
	k := NewKBD("Tab")
	k.SetVariant(KBDBracket)
	if k.Variant() != KBDBracket {
		t.Errorf("Variant = %v", k.Variant())
	}
	k.SetVariant(KBDBordered)
	if k.Variant() != KBDBordered {
		t.Errorf("Variant = %v", k.Variant())
	}
}

func TestP373_Measure(t *testing.T) {
	// Inverse: 1 row, text+2 wide
	k := NewKBD("OK")
	s := k.Measure(Constraints{MaxWidth: 20, MaxHeight: 5})
	if s.W != 4 || s.H != 1 {
		t.Errorf("Inverse Measure = %v, want {4,1}", s)
	}

	// Bracket: 1 row, text+2 wide
	k.SetVariant(KBDBracket)
	s = k.Measure(Constraints{MaxWidth: 20, MaxHeight: 5})
	if s.W != 4 || s.H != 1 {
		t.Errorf("Bracket Measure = %v, want {4,1}", s)
	}

	// Bordered: 3 rows, text+2 wide
	k.SetVariant(KBDBordered)
	s = k.Measure(Constraints{MaxWidth: 20, MaxHeight: 5})
	if s.W != 4 || s.H != 3 {
		t.Errorf("Bordered Measure = %v, want {4,3}", s)
	}

	// Clamped
	k.SetVariant(KBDBordered)
	s = k.Measure(Constraints{MaxWidth: 3, MaxHeight: 2})
	if s.W != 3 || s.H != 2 {
		t.Errorf("Clamped Measure = %v, want {3,2}", s)
	}
}

func TestP373_Paint_Inverse(t *testing.T) {
	k := NewKBD("A")
	k.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	k.Paint(buf)
	// Cell 0 = padding space, Cell 1 = 'A', Cell 2 = padding space
	c0 := buf.GetCell(0, 0)
	if c0.Rune != ' ' {
		t.Errorf("cell[0] = %q, want ' '", string(c0.Rune))
	}
	c1 := buf.GetCell(1, 0)
	if c1.Rune != 'A' {
		t.Errorf("cell[1] = %q, want 'A'", string(c1.Rune))
	}
	c2 := buf.GetCell(2, 0)
	if c2.Rune != ' ' {
		t.Errorf("cell[2] = %q, want ' '", string(c2.Rune))
	}
}

func TestP373_Paint_Bracket(t *testing.T) {
	k := NewKBD("OK")
	k.SetVariant(KBDBracket)
	k.SetBounds(Rect{X: 0, Y: 0, W: 4, H: 1})
	buf := buffer.NewBuffer(4, 1)
	k.Paint(buf)
	c0 := buf.GetCell(0, 0)
	if c0.Rune != '[' {
		t.Errorf("cell[0] = %q, want '['", string(c0.Rune))
	}
	c3 := buf.GetCell(3, 0)
	if c3.Rune != ']' {
		t.Errorf("cell[3] = %q, want ']'", string(c3.Rune))
	}
}

func TestP373_Paint_Bordered(t *testing.T) {
	k := NewKBD("K")
	k.SetVariant(KBDBordered)
	k.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 3})
	buf := buffer.NewBuffer(3, 3)
	k.Paint(buf)
	// Top: ┌─┐
	c := buf.GetCell(0, 0)
	if c.Rune != kbdTL {
		t.Errorf("top-left = %q, want %q", string(c.Rune), string(kbdTL))
	}
	c = buf.GetCell(1, 0)
	if c.Rune != kbdH {
		t.Errorf("top-mid = %q, want %q", string(c.Rune), string(kbdH))
	}
	c = buf.GetCell(2, 0)
	if c.Rune != kbdTR {
		t.Errorf("top-right = %q, want %q", string(c.Rune), string(kbdTR))
	}
	// Middle: │K│
	c = buf.GetCell(0, 1)
	if c.Rune != kbdV {
		t.Errorf("mid-left = %q, want %q", string(c.Rune), string(kbdV))
	}
	c = buf.GetCell(1, 1)
	if c.Rune != 'K' {
		t.Errorf("mid-text = %q, want 'K'", string(c.Rune))
	}
	c = buf.GetCell(2, 1)
	if c.Rune != kbdV {
		t.Errorf("mid-right = %q, want %q", string(c.Rune), string(kbdV))
	}
	// Bottom: └─┘
	c = buf.GetCell(0, 2)
	if c.Rune != kbdBL {
		t.Errorf("bot-left = %q, want %q", string(c.Rune), string(kbdBL))
	}
	c = buf.GetCell(2, 2)
	if c.Rune != kbdBR {
		t.Errorf("bot-right = %q, want %q", string(c.Rune), string(kbdBR))
	}
}

func TestP373_Paint_ZeroBounds(t *testing.T) {
	k := NewKBD("Ctrl")
	k.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	k.Paint(buf) // should not panic
}

func TestP373_Paint_NarrowWidth(t *testing.T) {
	k := NewKBD("Ctrl")
	k.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 1})
	buf := buffer.NewBuffer(1, 1)
	k.Paint(buf) // should not panic, should clip
}

func TestP373_Paint_NonZeroOffset(t *testing.T) {
	k := NewKBD("A")
	k.SetBounds(Rect{X: 5, Y: 2, W: 3, H: 1})
	buf := buffer.NewBuffer(10, 5)
	k.Paint(buf)
	c := buf.GetCell(6, 2)
	if c.Rune != 'A' {
		t.Errorf("offset cell = %q, want 'A'", string(c.Rune))
	}
}

func TestP373_Paint_Bordered_2RowHeight(t *testing.T) {
	// Bordered with only 2 rows height — bottom border should be skipped
	k := NewKBD("X")
	k.SetVariant(KBDBordered)
	k.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 2})
	buf := buffer.NewBuffer(3, 2)
	k.Paint(buf) // should not panic
}

func TestP373_Concurrent(t *testing.T) {
	k := NewKBD("Ctrl")
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ {
			k.SetText("Alt")
			k.SetVariant(KBDBracket)
		}
		close(done)
	}()
	for i := 0; i < 500; i++ {
		_ = k.Text()
		_ = k.Variant()
	}
	<-done
}

func TestP373_SatisfiesComponent(t *testing.T) {
	var _ Component = (*KBD)(nil)
}

// P373: KBD benchmark — zero alloc
func BenchmarkP373_KBD_Paint_Inverse(b *testing.B) {
	k := NewKBD("Ctrl+C")
	k.SetBounds(Rect{X: 0, Y: 0, W: 7, H: 1})
	buf := buffer.NewBuffer(7, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k.Paint(buf)
	}
}

func BenchmarkP373_KBD_Paint_Bordered(b *testing.B) {
	k := NewKBD("K")
	k.SetVariant(KBDBordered)
	k.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 3})
	buf := buffer.NewBuffer(3, 3)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		k.Paint(buf)
	}
}

package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P385: DiffStatBar component tests

func TestP385_NewDiffStatBar(t *testing.T) {
	d := NewDiffStatBar(12, 5)
	if d.Additions() != 12 {
		t.Errorf("Additions = %d, want 12", d.Additions())
	}
	if d.Deletions() != 5 {
		t.Errorf("Deletions = %d, want 5", d.Deletions())
	}
	if d.ID() == "" {
		t.Error("ID should not be empty")
	}
	if d.Style() != DiffStatStyleFull {
		t.Errorf("default style = %v, want Full", d.Style())
	}
	if d.BarWidth() != 10 {
		t.Errorf("default barWidth = %d, want 10", d.BarWidth())
	}
}

func TestP385_SetStats(t *testing.T) {
	d := NewDiffStatBar(0, 0)
	d.SetStats(100, 50, 3)
	if d.Additions() != 100 {
		t.Errorf("Additions = %d", d.Additions())
	}
	if d.Deletions() != 50 {
		t.Errorf("Deletions = %d", d.Deletions())
	}
	if d.Files() != 3 {
		t.Errorf("Files = %d", d.Files())
	}
}

func TestP385_SetBarWidth(t *testing.T) {
	d := NewDiffStatBar(10, 5)
	d.SetBarWidth(20)
	if d.BarWidth() != 20 {
		t.Errorf("BarWidth = %d, want 20", d.BarWidth())
	}
	// Zero clamps to 1
	d.SetBarWidth(0)
	if d.BarWidth() != 1 {
		t.Errorf("BarWidth = %d, want 1 (clamped)", d.BarWidth())
	}
}

func TestP385_SetStyle(t *testing.T) {
	d := NewDiffStatBar(10, 5)
	d.SetStyle(DiffStatStyleText)
	if d.Style() != DiffStatStyleText {
		t.Errorf("Style = %v", d.Style())
	}
	d.SetStyle(DiffStatStyleBar)
	if d.Style() != DiffStatStyleBar {
		t.Errorf("Style = %v", d.Style())
	}
}

func TestP385_SetColors(t *testing.T) {
	d := NewDiffStatBar(10, 5)
	add := buffer.RGB(0x00, 0xFF, 0x00)
	del := buffer.RGB(0xFF, 0x00, 0x00)
	d.SetColors(add, del)
	// Just verify it doesn't panic and renders
	d.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	d.Paint(buf)
}

func TestP385_Measure(t *testing.T) {
	d := NewDiffStatBar(12, 5)
	s := d.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if s.H != 1 {
		t.Errorf("H = %d, want 1", s.H)
	}
	if s.W < 5 {
		t.Errorf("W = %d, too small", s.W)
	}

	// Text style
	d.SetStyle(DiffStatStyleText)
	s = d.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if s.H != 1 {
		t.Errorf("Text H = %d", s.H)
	}

	// Clamped
	d.SetStyle(DiffStatStyleFull)
	s = d.Measure(Constraints{MaxWidth: 3, MaxHeight: 1})
	if s.W != 3 {
		t.Errorf("Clamped W = %d, want 3", s.W)
	}
}

func TestP385_Paint_Full(t *testing.T) {
	d := NewDiffStatBar(12, 5)
	d.SetStats(12, 5, 3)
	d.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	d.Paint(buf)
	// Verify bar cells exist (first 10 cells should be ▓)
	c := buf.GetCell(0, 0)
	if c.Rune != '\u2593' {
		t.Errorf("bar cell[0] = %q, want ▓", string(c.Rune))
	}
}

func TestP385_Paint_Text(t *testing.T) {
	d := NewDiffStatBar(12, 5)
	d.SetStyle(DiffStatStyleText)
	d.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	d.Paint(buf)
	// First char should be '+'
	c := buf.GetCell(0, 0)
	if c.Rune != '+' {
		t.Errorf("text cell[0] = %q, want '+'", string(c.Rune))
	}
}

func TestP385_Paint_Bar(t *testing.T) {
	d := NewDiffStatBar(20, 10)
	d.SetStyle(DiffStatStyleBar)
	d.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	d.Paint(buf)
	// Bar cells should be ▓
	c := buf.GetCell(0, 0)
	if c.Rune != '\u2593' {
		t.Errorf("bar cell[0] = %q, want ▓", string(c.Rune))
	}
}

func TestP385_Paint_ZeroChanges(t *testing.T) {
	d := NewDiffStatBar(0, 0)
	d.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	d.Paint(buf) // should not panic, all neutral
}

func TestP385_Paint_OnlyAdditions(t *testing.T) {
	d := NewDiffStatBar(50, 0)
	d.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	d.Paint(buf)
	// All bar cells should be green
	c := buf.GetCell(0, 0)
	if c.Rune != '\u2593' {
		t.Errorf("cell[0] = %q", string(c.Rune))
	}
}

func TestP385_Paint_OnlyDeletions(t *testing.T) {
	d := NewDiffStatBar(0, 50)
	d.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	d.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != '\u2593' {
		t.Errorf("cell[0] = %q", string(c.Rune))
	}
}

func TestP385_Paint_SingleFile(t *testing.T) {
	d := NewDiffStatBar(10, 3)
	d.SetStats(10, 3, 1) // single file
	d.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	d.Paint(buf) // should show "1 file)" not "1 files)"
}

func TestP385_Paint_ZeroBounds(t *testing.T) {
	d := NewDiffStatBar(10, 5)
	d.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	d.Paint(buf) // should not panic
}

func TestP385_Paint_NarrowWidth(t *testing.T) {
	d := NewDiffStatBar(100, 50)
	d.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	d.Paint(buf) // should clip without panic
}

func TestP385_Paint_NonZeroOffset(t *testing.T) {
	d := NewDiffStatBar(10, 5)
	d.SetBounds(Rect{X: 5, Y: 2, W: 40, H: 1})
	buf := buffer.NewBuffer(50, 5)
	d.Paint(buf)
	c := buf.GetCell(5, 2)
	if c.Rune != '\u2593' {
		t.Errorf("offset cell = %q, want ▓", string(c.Rune))
	}
}

func TestP385_Concurrent(t *testing.T) {
	d := NewDiffStatBar(10, 5)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ {
			d.SetStats(i, i/2, i/10)
			d.SetStyle(DiffStatStyleText)
		}
		close(done)
	}()
	for i := 0; i < 500; i++ {
		_ = d.Additions()
		_ = d.Style()
	}
	<-done
}

func TestP385_SatisfiesComponent(t *testing.T) {
	var _ Component = (*DiffStatBar)(nil)
}

func TestP385_NumDigits(t *testing.T) {
	tests := []struct {
		n    int
		want int
	}{
		{0, 1},
		{5, 1},
		{9, 1},
		{10, 2},
		{99, 2},
		{100, 3},
		{999, 3},
		{1000, 4},
		{-42, 2},
	}
	for _, tt := range tests {
		got := numDigits(tt.n)
		if got != tt.want {
			t.Errorf("numDigits(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}

// Benchmark
func BenchmarkP385_DiffStatBar_Paint_Full(b *testing.B) {
	d := NewDiffStatBar(120, 50)
	d.SetStats(120, 50, 3)
	d.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Paint(buf)
	}
}

func BenchmarkP385_DiffStatBar_Paint_Text(b *testing.B) {
	d := NewDiffStatBar(120, 50)
	d.SetStyle(DiffStatStyleText)
	d.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Paint(buf)
	}
}

package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// === Divider tests ===

func TestP404_NewDivider(t *testing.T) {
	d := NewDivider("Section")
	if d.Label() != "Section" { t.Errorf("Label = %q", d.Label()) }
	if d.Orientation() != DividerHorizontal { t.Error("should be horizontal") }
	if d.Char() != '\u2500' { t.Errorf("Char = %q", string(d.Char())) }
	if d.ID() == "" { t.Error("ID empty") }
}

func TestP404_Divider_SetLabel(t *testing.T) {
	d := NewDivider("old")
	d.SetLabel("new")
	if d.Label() != "new" { t.Errorf("Label = %q", d.Label()) }
}

func TestP404_Divider_SetOrientation(t *testing.T) {
	d := NewDivider("")
	d.SetOrientation(DividerVertical)
	if d.Orientation() != DividerVertical { t.Error("should be vertical") }
}

func TestP404_Divider_SetChar(t *testing.T) {
	d := NewDivider("")
	d.SetChar('*')
	if d.Char() != '*' { t.Errorf("Char = %q", string(d.Char())) }
}

func TestP404_Divider_SetColor(t *testing.T) {
	d := NewDivider("x")
	d.SetColor(buffer.RGB(1, 2, 3))
	d.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	d.Paint(buf)
}

func TestP404_Divider_Measure_Horizontal(t *testing.T) {
	d := NewDivider("")
	s := d.Measure(Constraints{MaxWidth: 30, MaxHeight: 5})
	if s.H != 1 { t.Errorf("H = %d", s.H) }
	if s.W != 30 { t.Errorf("W = %d", s.W) }
}

func TestP404_Divider_Measure_Vertical(t *testing.T) {
	d := NewDivider("")
	d.SetOrientation(DividerVertical)
	s := d.Measure(Constraints{MaxWidth: 5, MaxHeight: 20})
	if s.W != 1 { t.Errorf("W = %d", s.W) }
	if s.H != 20 { t.Errorf("H = %d", s.H) }
}

func TestP404_Divider_Paint_NoLabel(t *testing.T) {
	d := NewDivider("")
	d.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	d.Paint(buf)
	for i := 0; i < 10; i++ {
		c := buf.GetCell(i, 0)
		if c.Rune != '\u2500' { t.Errorf("cell[%d] = %q, want ─", i, string(c.Rune)) }
	}
}

func TestP404_Divider_Paint_WithLabel(t *testing.T) {
	d := NewDivider("Title")
	d.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	d.Paint(buf)
	// Should have ─ chars and " Title " in middle
	found := false
	for i := 0; i < 20; i++ {
		if buf.GetCell(i, 0).Rune == 'T' { found = true }
	}
	if !found { t.Error("label text not found in output") }
}

func TestP404_Divider_Paint_Vertical(t *testing.T) {
	d := NewDivider("")
	d.SetOrientation(DividerVertical)
	d.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 5})
	buf := buffer.NewBuffer(1, 5)
	d.Paint(buf)
	for i := 0; i < 5; i++ {
		c := buf.GetCell(0, i)
		if c.Rune != '\u2502' { t.Errorf("cell[%d] = %q, want │", i, string(c.Rune)) }
	}
}

func TestP404_Divider_Paint_LongLabel(t *testing.T) {
	d := NewDivider("very long label that exceeds width")
	d.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	d.Paint(buf) // should just draw label
}

func TestP404_Divider_Paint_ZeroBounds(t *testing.T) {
	d := NewDivider("x")
	d.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	d.Paint(buf)
}

func TestP404_Divider_Concurrent(t *testing.T) {
	d := NewDivider("x")
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ { d.SetLabel("concurrent") }
		close(done)
	}()
	for i := 0; i < 500; i++ { _ = d.Label() }
	<-done
}

func TestP404_Divider_SatisfiesComponent(t *testing.T) {
	var _ Component = (*Divider)(nil)
}

func BenchmarkP404_Divider_Paint_NoLabel(b *testing.B) {
	d := NewDivider("")
	d.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { d.Paint(buf) }
}

func BenchmarkP404_Divider_Paint_WithLabel(b *testing.B) {
	d := NewDivider("Settings")
	d.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { d.Paint(buf) }
}

// === Coverage: statcard Measure 70.6% → target 90%+ ===
// === Coverage: toast Measure 78.6% → target 90%+ ===

func TestP404_StatCard_Measure_EdgeCases(t *testing.T) {
	sc := NewStatCard("Latency", "42ms")
	sc.SetDelta("+5%", true)
	// Exercise delta > label/value width path
	s := sc.Measure(Constraints{MaxWidth: 100, MaxHeight: 10})
	if s.H != 3 { t.Errorf("H = %d, want 3", s.H) }

	// No delta path
	sc.ClearDelta()
	s = sc.Measure(Constraints{MaxWidth: 100, MaxHeight: 10})
	if s.H != 3 { t.Errorf("H = %d, want 3 (still 3 for layout)", s.H) }

	// Tiny constraints — exercises clamping
	s = sc.Measure(Constraints{MaxWidth: 1, MaxHeight: 1})
	
	

	// Empty constraints
	s = sc.Measure(Constraints{})
	if s.W < 1 { t.Error("W should be >= 1") }
}

func TestP404_Toast_Measure_EdgeCases(t *testing.T) {
	// Short message exercises minimum width clamp
	tt := NewToast("x", ToastInfo)
	s := tt.Measure(Constraints{MaxWidth: 100, MaxHeight: 10})
	if s.W != 6 { t.Errorf("W = %d, want 6 (minimum)", s.W) }

	// Long message
	tt2 := NewToast("This is a longer message", ToastSuccess)
	s2 := tt2.Measure(Constraints{MaxWidth: 100, MaxHeight: 10})
	if s2.W < 10 { t.Errorf("W = %d, too small", s2.W) }

	// Empty message
	tt3 := NewToast("", ToastInfo)
	s3 := tt3.Measure(Constraints{MaxWidth: 100, MaxHeight: 10})
	if s3.W < 6 { t.Errorf("W = %d, want >= 6", s3.W) }

	// Zero constraints
	s4 := tt.Measure(Constraints{})
	if s4.W < 1 { t.Error("W should be >= 1") }
}

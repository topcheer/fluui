package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P389: ConfidenceMeter component tests

func TestP389_NewConfidenceMeter(t *testing.T) {
	c := NewConfidenceMeter(0.85)
	if c.Value() != 0.85 {
		t.Errorf("Value = %v, want 0.85", c.Value())
	}
	if c.ID() == "" {
		t.Error("ID should not be empty")
	}
	if !c.ShowPct() {
		t.Error("ShowPct should default true")
	}
	if !c.ShowLabel() {
		t.Error("ShowLabel should default true")
	}
	if c.BarWidth() != 12 {
		t.Errorf("BarWidth = %d, want 12", c.BarWidth())
	}
}

func TestP389_SetValue(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	c.SetValue(0.9)
	if c.Value() != 0.9 {
		t.Errorf("Value = %v", c.Value())
	}
	// Clamp
	c.SetValue(-0.5)
	if c.Value() != 0 {
		t.Errorf("Value = %v, want 0 (clamped)", c.Value())
	}
	c.SetValue(1.5)
	if c.Value() != 1 {
		t.Errorf("Value = %v, want 1 (clamped)", c.Value())
	}
}

func TestP389_SetLabel(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	c.SetLabel("Certainty")
	if c.Label() != "Certainty" {
		t.Errorf("Label = %q", c.Label())
	}
}

func TestP389_SetBarWidth(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	c.SetBarWidth(20)
	if c.BarWidth() != 20 {
		t.Errorf("BarWidth = %d", c.BarWidth())
	}
	c.SetBarWidth(0)
	if c.BarWidth() != 1 {
		t.Errorf("BarWidth = %d, want 1 (clamped)", c.BarWidth())
	}
}

func TestP389_SetShowPct(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	c.SetShowPct(false)
	if c.ShowPct() {
		t.Error("ShowPct should be false")
	}
}

func TestP389_SetShowLabel(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	c.SetShowLabel(false)
	if c.ShowLabel() {
		t.Error("ShowLabel should be false")
	}
}

func TestP389_SetColor(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	custom := buffer.RGB(0xFF, 0x00, 0xFF)
	c.SetColor(custom)
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	c.Paint(buf) // should not panic
}

func TestP389_Measure(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	c.SetLabel("Conf")
	s := c.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if s.H != 1 {
		t.Errorf("H = %d, want 1", s.H)
	}
	// "Conf " (5) + bar (12) + " NN%" (4) = 21
	if s.W != 21 {
		t.Errorf("W = %d, want 21", s.W)
	}
}

func TestP389_Measure_NoLabel(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	c.SetShowLabel(false)
	s := c.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	// bar (12) + " NN%" (4) = 16
	if s.W != 16 {
		t.Errorf("W = %d, want 16", s.W)
	}
}

func TestP389_Measure_NoPct(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	c.SetShowPct(false)
	c.SetLabel("Conf")
	s := c.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	// "Conf " (5) + bar (12) = 17
	if s.W != 17 {
		t.Errorf("W = %d, want 17", s.W)
	}
}

func TestP389_Measure_Clamp(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	s := c.Measure(Constraints{MaxWidth: 3, MaxHeight: 1})
	if s.W != 3 {
		t.Errorf("Clamped W = %d, want 3", s.W)
	}
}

func TestP389_Paint_HighConfidence(t *testing.T) {
	c := NewConfidenceMeter(0.9)
	c.SetLabel("Conf")
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	c.Paint(buf)
	// First cells should be label, then filled bar (█)
	// Find first █ cell
	foundFilled := false
	for i := 0; i < 40; i++ {
		cell := buf.GetCell(i, 0)
		if cell.Rune == '\u2588' {
			foundFilled = true
			break
		}
	}
	if !foundFilled {
		t.Error("should have filled bar cells (█)")
	}
}

func TestP389_Paint_LowConfidence(t *testing.T) {
	c := NewConfidenceMeter(0.2)
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	c.Paint(buf) // should render with few filled cells
}

func TestP389_Paint_ZeroValue(t *testing.T) {
	c := NewConfidenceMeter(0)
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	c.Paint(buf) // all empty cells
}

func TestP389_Paint_FullValue(t *testing.T) {
	c := NewConfidenceMeter(1.0)
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	c.Paint(buf) // all filled cells
}

func TestP389_Paint_NoLabel(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	c.SetShowLabel(false)
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	c.Paint(buf)
	// First cell should be bar (no label prefix)
	cell := buf.GetCell(0, 0)
	if cell.Rune != '\u2588' && cell.Rune != '\u2591' {
		t.Errorf("cell[0] = %q, want bar char", string(cell.Rune))
	}
}

func TestP389_Paint_NoPct(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	c.SetShowPct(false)
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	c.Paint(buf)
}

func TestP389_Paint_ZeroBounds(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	c.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	c.Paint(buf) // should not panic
}

func TestP389_Paint_NarrowWidth(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	c.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	c.Paint(buf) // should clip without panic
}

func TestP389_Paint_NonZeroOffset(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	c.SetShowLabel(false)
	c.SetBounds(Rect{X: 5, Y: 2, W: 30, H: 1})
	buf := buffer.NewBuffer(40, 5)
	c.Paint(buf)
	cell := buf.GetCell(5, 2)
	if cell.Rune != '\u2588' && cell.Rune != '\u2591' {
		t.Errorf("offset cell = %q", string(cell.Rune))
	}
}

func TestP389_Concurrent(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ {
			c.SetValue(0.3)
			c.SetLabel("test")
		}
		close(done)
	}()
	for i := 0; i < 500; i++ {
		_ = c.Value()
		_ = c.Label()
	}
	<-done
}

func TestP389_SatisfiesComponent(t *testing.T) {
	var _ Component = (*ConfidenceMeter)(nil)
}

// Benchmarks
func BenchmarkP389_ConfidenceMeter_Paint(b *testing.B) {
	c := NewConfidenceMeter(0.75)
	c.SetLabel("Confidence")
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkP389_ConfidenceMeter_Paint_NoLabel(b *testing.B) {
	c := NewConfidenceMeter(0.75)
	c.SetShowLabel(false)
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

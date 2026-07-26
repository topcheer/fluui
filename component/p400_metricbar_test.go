package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP400_NewMetricBar(t *testing.T) {
	m := NewMetricBar("CPU", 42.5, 0, 100)
	if m.Label() != "CPU" { t.Errorf("Label = %q", m.Label()) }
	if m.Value() != 42.5 { t.Errorf("Value = %v", m.Value()) }
	if m.Min() != 0 { t.Errorf("Min = %v", m.Min()) }
	if m.Max() != 100 { t.Errorf("Max = %v", m.Max()) }
	if m.BarWidth() != 10 { t.Errorf("BarWidth = %d", m.BarWidth()) }
	if m.ID() == "" { t.Error("ID empty") }
}

func TestP400_SetLabel(t *testing.T) {
	m := NewMetricBar("old", 0, 0, 1)
	m.SetLabel("new")
	if m.Label() != "new" { t.Errorf("Label = %q", m.Label()) }
}

func TestP400_SetValue(t *testing.T) {
	m := NewMetricBar("x", 0, 0, 1)
	m.SetValue(0.5)
	if m.Value() != 0.5 { t.Errorf("Value = %v", m.Value()) }
}

func TestP400_SetRange(t *testing.T) {
	m := NewMetricBar("x", 5, 0, 10)
	m.SetRange(0, 100)
	if m.Max() != 100 { t.Errorf("Max = %v", m.Max()) }
}

func TestP400_SetUnit(t *testing.T) {
	m := NewMetricBar("x", 0, 0, 1)
	m.SetUnit("ms")
	if m.Unit() != "ms" { t.Errorf("Unit = %q", m.Unit()) }
}

func TestP400_SetBarWidth(t *testing.T) {
	m := NewMetricBar("x", 0, 0, 1)
	m.SetBarWidth(20)
	if m.BarWidth() != 20 { t.Errorf("BarWidth = %d", m.BarWidth()) }
	m.SetBarWidth(0)
	if m.BarWidth() != 1 { t.Errorf("BarWidth = %d, want 1", m.BarWidth()) }
}

func TestP400_SetColor(t *testing.T) {
	m := NewMetricBar("x", 0.5, 0, 1)
	m.SetColor(buffer.RGB(1, 2, 3))
	m.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	m.Paint(buf)
}

func TestP400_Measure(t *testing.T) {
	m := NewMetricBar("CPU", 50, 0, 100)
	m.SetUnit("%")
	s := m.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if s.H != 1 { t.Errorf("H = %d", s.H) }
	if s.W < 15 { t.Errorf("W = %d, too small", s.W) }
}

func TestP400_Paint(t *testing.T) {
	m := NewMetricBar("CPU", 75, 0, 100)
	m.SetUnit("%")
	m.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	m.Paint(buf)
	// Should have label text, then bar cells
	c := buf.GetCell(0, 0)
	if c.Rune != 'C' { t.Errorf("cell[0] = %q, want 'C'", string(c.Rune)) }
}

func TestP400_Paint_FullBar(t *testing.T) {
	m := NewMetricBar("Mem", 100, 0, 100)
	m.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	m.Paint(buf)
}

func TestP400_Paint_ZeroValue(t *testing.T) {
	m := NewMetricBar("Disk", 0, 0, 100)
	m.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	m.Paint(buf)
}

func TestP400_Paint_ColorThresholds(t *testing.T) {
	for _, v := range []float64{0.1, 0.5, 0.7, 0.9} {
		m := NewMetricBar("x", v*100, 0, 100)
		m.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
		buf := buffer.NewBuffer(30, 1)
		m.Paint(buf)
	}
}

func TestP400_Paint_ZeroRange(t *testing.T) {
	m := NewMetricBar("x", 5, 5, 5) // min == max → pct = 0
	m.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	m.Paint(buf)
}

func TestP400_Paint_ZeroBounds(t *testing.T) {
	m := NewMetricBar("x", 1, 0, 2)
	m.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	m.Paint(buf)
}

func TestP400_Paint_NonZeroOffset(t *testing.T) {
	m := NewMetricBar("CPU", 50, 0, 100)
	m.SetBounds(Rect{X: 10, Y: 5, W: 30, H: 1})
	buf := buffer.NewBuffer(50, 10)
	m.Paint(buf)
	c := buf.GetCell(10, 5)
	if c.Rune != 'C' { t.Errorf("offset cell = %q", string(c.Rune)) }
}

func TestP400_PctClamped(t *testing.T) {
	m := NewMetricBar("x", 150, 0, 100) // value > max
	m.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	m.Paint(buf) // should clamp to 100%

	m2 := NewMetricBar("x", -10, 0, 100) // value < min
	m2.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf2 := buffer.NewBuffer(30, 1)
	m2.Paint(buf2) // should clamp to 0%
}

func TestP400_Concurrent(t *testing.T) {
	m := NewMetricBar("x", 50, 0, 100)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ { m.SetValue(75); m.SetLabel("concurrent") }
		close(done)
	}()
	for i := 0; i < 500; i++ { _ = m.Value(); _ = m.Label() }
	<-done
}

func TestP400_SatisfiesComponent(t *testing.T) {
	var _ Component = (*MetricBar)(nil)
}

func BenchmarkP400_MetricBar_Paint(b *testing.B) {
	m := NewMetricBar("CPU Usage", 73.5, 0, 100)
	m.SetUnit("%")
	m.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { m.Paint(buf) }
}

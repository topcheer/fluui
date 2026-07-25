package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP361_Funnel_Create(t *testing.T) {
	f := NewFunnelChart([]FunnelSlice{
		{Label: "Visits", Value: 1000},
		{Label: "Sign-ups", Value: 500},
		{Label: "Active", Value: 200},
		{Label: "Paid", Value: 50},
	})
	if f.SliceCount() != 4 {
		t.Errorf("count = %d", f.SliceCount())
	}
	if f.TotalValue() != 1000 {
		t.Errorf("total = %f, want 1000", f.TotalValue())
	}
}

func TestP361_Funnel_SetSlices(t *testing.T) {
	f := NewFunnelChart(nil)
	f.SetSlices([]FunnelSlice{{Label: "A", Value: 10}})
	if f.SliceCount() != 1 {
		t.Errorf("count = %d", f.SliceCount())
	}
}

func TestP361_Funnel_Empty(t *testing.T) {
	f := NewFunnelChart(nil)
	if f.TotalValue() != 0 {
		t.Error("empty total should be 0")
	}
}

func TestP361_Funnel_Measure(t *testing.T) {
	f := NewFunnelChart([]FunnelSlice{{Label: "A", Value: 10}, {Label: "B", Value: 5}})
	s := f.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W != 40 {
		t.Errorf("default width = %d, want 40", s.W)
	}
	if s.H != 2 {
		t.Errorf("height = %d, want 2", s.H)
	}
}

func TestP361_Funnel_Paint(t *testing.T) {
	f := NewFunnelChart([]FunnelSlice{
		{Label: "Visits", Value: 1000},
		{Label: "Paid", Value: 50},
	})
	f.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	f.Paint(buf)

	if buf.GetCell(0, 0).Bg.Type == 0 && buf.GetCell(0, 0).Bg.Val == 0 {
		t.Error("expected colored cells")
	}
}

func TestP361_Funnel_Paint_Empty(t *testing.T) {
	f := NewFunnelChart(nil)
	f.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	f.Paint(buf)
}

func TestP361_Funnel_Paint_ZeroValue(t *testing.T) {
	f := NewFunnelChart([]FunnelSlice{{Label: "A", Value: 0}})
	f.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	f.Paint(buf)
}

func TestP361_Funnel_Paint_ZeroBounds(t *testing.T) {
	f := NewFunnelChart([]FunnelSlice{{Label: "A", Value: 10}})
	f.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(40, 4)
	f.Paint(buf)
}

func TestP361_Funnel_Paint_LongLabel(t *testing.T) {
	f := NewFunnelChart([]FunnelSlice{
		{Label: "This is a very long label that won't fit", Value: 100},
	})
	f.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	f.Paint(buf)
}

func TestP361_Funnel_Paint_ManySlices(t *testing.T) {
	slices := make([]FunnelSlice, 10)
	for i := range slices {
		slices[i] = FunnelSlice{Label: "S", Value: float64(100 - i*10)}
	}
	f := NewFunnelChart(slices)
	f.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	f.Paint(buf) // should clip to height 5
}

func BenchmarkFunnelChart_Paint(b *testing.B) {
	f := NewFunnelChart([]FunnelSlice{
		{Label: "Input", Value: 1500},
		{Label: "Cached", Value: 800},
		{Label: "Output", Value: 500},
		{Label: "Error", Value: 50},
	})
	f.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		f.Paint(buf)
	}
}

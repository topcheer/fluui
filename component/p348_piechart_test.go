package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

func TestP348_PieChart_Create(t *testing.T) {
	p := NewPieChart([]PieSlice{
		{Label: "Input", Value: 100},
		{Label: "Output", Value: 50},
		{Label: "Cache", Value: 30},
	})
	if p.SliceCount() != 3 {
		t.Errorf("count = %d, want 3", p.SliceCount())
	}
	if p.TotalValue() != 180 {
		t.Errorf("total = %f, want 180", p.TotalValue())
	}
}

func TestP348_PieChart_SetSlices(t *testing.T) {
	p := NewPieChart(nil)
	p.SetSlices([]PieSlice{{Label: "A", Value: 10}})
	if p.SliceCount() != 1 {
		t.Errorf("count = %d", p.SliceCount())
	}
}

func TestP348_PieChart_Donut(t *testing.T) {
	p := NewPieChart([]PieSlice{{Label: "A", Value: 10}})
	if p.IsDonut() {
		t.Error("should start as pie")
	}
	p.SetDonut(true)
	if !p.IsDonut() {
		t.Error("should be donut after SetDonut(true)")
	}
}

func TestP348_PieChart_Radius(t *testing.T) {
	p := NewPieChart([]PieSlice{{Label: "A", Value: 10}})
	p.SetRadius(5)
}

func TestP348_PieChart_TotalZero(t *testing.T) {
	p := NewPieChart([]PieSlice{
		{Label: "A", Value: 0},
		{Label: "B", Value: 0},
	})
	if p.TotalValue() != 0 {
		t.Errorf("total = %f, want 0", p.TotalValue())
	}
}

func TestP348_PieChart_Measure(t *testing.T) {
	p := NewPieChart([]PieSlice{{Label: "A", Value: 10}})
	s := p.Measure(Constraints{MaxWidth: 40, MaxHeight: 15})
	if s.W < 1 || s.H < 1 {
		t.Error("expected non-zero size")
	}
}

func TestP348_PieChart_Paint(t *testing.T) {
	p := NewPieChart([]PieSlice{
		{Label: "Input", Value: 100},
		{Label: "Output", Value: 50},
	})
	p.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	p.Paint(buf)

	// Should have drawn something (non-empty cells)
	filled := 0
	for y := 0; y < 10; y++ {
		for x := 0; x < 40; x++ {
			c := buf.GetCell(x, y)
			if c.Rune != 0 && c.Rune != ' ' {
				filled++
			}
		}
	}
	if filled == 0 {
		t.Error("expected non-empty cells after Paint")
	}
}

func TestP348_PieChart_Paint_Donut(t *testing.T) {
	p := NewPieChart([]PieSlice{
		{Label: "A", Value: 40},
		{Label: "B", Value: 30},
		{Label: "C", Value: 30},
	})
	p.SetDonut(true)
	p.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	p.Paint(buf) // should render donut without panic
}

func TestP348_PieChart_Paint_Empty(t *testing.T) {
	p := NewPieChart(nil)
	p.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	p.Paint(buf) // should not panic
}

func TestP348_PieChart_Paint_ZeroTotal(t *testing.T) {
	p := NewPieChart([]PieSlice{{Label: "A", Value: 0}})
	p.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	p.Paint(buf) // should not panic or divide by zero
}

func TestP348_PieChart_Paint_ZeroBounds(t *testing.T) {
	p := NewPieChart([]PieSlice{{Label: "A", Value: 10}})
	p.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(30, 10)
	p.Paint(buf) // should not panic
}

func TestP348_PieChart_Paint_LongLabels(t *testing.T) {
	p := NewPieChart([]PieSlice{
		{Label: "VeryLongModelNameHere", Value: 50},
		{Label: "AnotherLongLabel", Value: 30},
	})
	p.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 8})
	buf := buffer.NewBuffer(20, 8)
	p.Paint(buf) // should truncate labels without panic
}

func TestP348_PieChart_Paint_SingleSlice(t *testing.T) {
	p := NewPieChart([]PieSlice{{Label: "Only", Value: 100}})
	p.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 8})
	buf := buffer.NewBuffer(20, 8)
	p.Paint(buf) // single slice = full circle
}

func TestP348_PieChart_Paint_CustomRadius(t *testing.T) {
	p := NewPieChart([]PieSlice{
		{Label: "A", Value: 50},
		{Label: "B", Value: 50},
	})
	p.SetRadius(3)
	p.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	p.Paint(buf)
}

func TestP348_FormatFloat(t *testing.T) {
	tests := []struct {
		v    float64
		prec int
		want string
	}{
		{0.0, 0, "0"},
		{42.0, 0, "42"},
		{3.14, 2, "3.14"},
		{50.0, 0, "50"},
		{99.9, 1, "99.9"},
		{-5.0, 0, "-5"},
	}
	for _, tt := range tests {
		got := pieFormatFloat(tt.v, tt.prec)
		if got != tt.want {
			t.Errorf("pieFormatFloat(%f, %d) = %q, want %q", tt.v, tt.prec, got, tt.want)
		}
	}
}

func TestP348_SliceColor(t *testing.T) {
	th := theme.Get()
	// Just ensure no panic and returns valid colors
	_ = sliceColor(0, th)
	_ = sliceColor(1, th)
	// Test wrap-around
	sliceColor(10, th)
}

func BenchmarkPieChart_Paint(b *testing.B) {
	p := NewPieChart([]PieSlice{
		{Label: "Input tokens", Value: 1500},
		{Label: "Output tokens", Value: 800},
		{Label: "Cache read", Value: 300},
		{Label: "Reasoning", Value: 200},
	})
	p.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 12})
	buf := buffer.NewBuffer(50, 12)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p.Paint(buf)
	}
}

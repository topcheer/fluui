package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP362_Radar_Create(t *testing.T) {
	r := NewRadarChart([]RadarAxis{
		{Label: "Speed", Value: 80, Max: 100},
		{Label: "Quality", Value: 90, Max: 100},
		{Label: "Cost", Value: 40, Max: 100},
		{Label: "Safety", Value: 95, Max: 100},
		{Label: "Scale", Value: 60, Max: 100},
	})
	if r.AxisCount() != 5 {
		t.Errorf("count = %d", r.AxisCount())
	}
}

func TestP362_Radar_SetAxes(t *testing.T) {
	r := NewRadarChart(nil)
	r.SetAxes([]RadarAxis{{Label: "A", Value: 50, Max: 100}})
	if r.AxisCount() != 1 {
		t.Errorf("count = %d", r.AxisCount())
	}
}

func TestP362_Radar_TooFewAxes(t *testing.T) {
	r := NewRadarChart([]RadarAxis{{Label: "A", Value: 50, Max: 100}})
	r.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	r.Paint(buf) // < 3 axes → should not render, not panic
}

func TestP362_Radar_Measure(t *testing.T) {
	r := NewRadarChart([]RadarAxis{
		{Label: "A", Value: 50, Max: 100},
		{Label: "B", Value: 50, Max: 100},
		{Label: "C", Value: 50, Max: 100},
	})
	s := r.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W != 30 || s.H != 15 {
		t.Errorf("defaults = %dx%d, want 30x15", s.W, s.H)
	}
}

func TestP362_Radar_Paint(t *testing.T) {
	r := NewRadarChart([]RadarAxis{
		{Label: "Speed", Value: 80, Max: 100},
		{Label: "Quality", Value: 90, Max: 100},
		{Label: "Cost", Value: 40, Max: 100},
		{Label: "Safety", Value: 95, Max: 100},
	})
	r.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 12})
	buf := buffer.NewBuffer(30, 12)
	r.Paint(buf)

	filled := 0
	for y := 0; y < 12; y++ {
		for x := 0; x < 30; x++ {
			if buf.GetCell(x, y).Rune != 0 {
				filled++
			}
		}
	}
	if filled < 5 {
		t.Errorf("expected at least 5 filled cells, got %d", filled)
	}
}

func TestP362_Radar_Paint_ZeroBounds(t *testing.T) {
	r := NewRadarChart([]RadarAxis{
		{Label: "A", Value: 50, Max: 100},
		{Label: "B", Value: 50, Max: 100},
		{Label: "C", Value: 50, Max: 100},
	})
	r.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(30, 12)
	r.Paint(buf)
}

func TestP362_Radar_Paint_MaxZero(t *testing.T) {
	r := NewRadarChart([]RadarAxis{
		{Label: "A", Value: 50, Max: 0},
		{Label: "B", Value: 50, Max: 0},
		{Label: "C", Value: 50, Max: 0},
	})
	r.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 12})
	buf := buffer.NewBuffer(30, 12)
	r.Paint(buf) // Max=0 should default to 1, not panic
}

func TestP362_Radar_Paint_OverMax(t *testing.T) {
	r := NewRadarChart([]RadarAxis{
		{Label: "A", Value: 150, Max: 100},
		{Label: "B", Value: 50, Max: 100},
		{Label: "C", Value: 200, Max: 100},
	})
	r.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 12})
	buf := buffer.NewBuffer(30, 12)
	r.Paint(buf) // values > Max should clamp to 1.0
}

func TestP362_DrawLine(t *testing.T) {
	buf := buffer.NewBuffer(20, 10)
	drawLine(buf, 0, 0, 10, 5, buffer.Style{}, Rect{X: 0, Y: 0, W: 20, H: 10})
	// Should have drawn some cells
	if buf.GetCell(0, 0).Rune == 0 {
		t.Error("expected non-empty at start")
	}
}

func BenchmarkRadarChart_Paint(b *testing.B) {
	r := NewRadarChart([]RadarAxis{
		{Label: "Speed", Value: 80, Max: 100},
		{Label: "Quality", Value: 90, Max: 100},
		{Label: "Cost", Value: 40, Max: 100},
		{Label: "Safety", Value: 95, Max: 100},
		{Label: "Scale", Value: 60, Max: 100},
		{Label: "UX", Value: 70, Max: 100},
	})
	r.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 12})
	buf := buffer.NewBuffer(30, 12)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.Paint(buf)
	}
}

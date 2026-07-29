package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestScatterPlotBasic(t *testing.T) {
	sp := NewScatterPlot()
	sp.AddPoint(10, 20)
	sp.AddPoint(50, 80)
	sp.AddPoint(30, 60)

	if sp.PointCount() != 3 {
		t.Errorf("PointCount = %d, want 3", sp.PointCount())
	}
	pts := sp.Points()
	if len(pts) != 3 {
		t.Fatalf("Points len = %d, want 3", len(pts))
	}
	if pts[0].X != 10 || pts[0].Y != 20 {
		t.Errorf("pts[0] = %+v, want {10,20}", pts[0])
	}
}

func TestScatterPlotSetPoints(t *testing.T) {
	sp := NewScatterPlot()
	sp.SetPoints([]ScatterPoint{
		{X: 1, Y: 2},
		{X: 3, Y: 4},
	})
	if sp.PointCount() != 2 {
		t.Errorf("PointCount = %d, want 2", sp.PointCount())
	}
}

func TestScatterPlotManualRange(t *testing.T) {
	sp := NewScatterPlot()
	sp.SetXRange(0, 100)
	sp.SetYRange(0, 50)
	sp.AddPoint(50, 25)

	xMin, xMax := sp.XRange()
	if xMin != 0 || xMax != 100 {
		t.Errorf("XRange = (%f,%f), want (0,100)", xMin, xMax)
	}
	yMin, yMax := sp.YRange()
	if yMin != 0 || yMax != 50 {
		t.Errorf("YRange = (%f,%f), want (0,50)", yMin, yMax)
	}
}

func TestScatterPlotAutoRange(t *testing.T) {
	sp := NewScatterPlot()
	sp.SetAutoScale(true)
	sp.AddPoint(10, 20)
	sp.AddPoint(30, 60)
	sp.AddPoint(50, 80)

	xMin, xMax := sp.XRange()
	if xMin != 10 {
		t.Errorf("autoXMin = %f, want 10", xMin)
	}
	if xMax != 50 {
		t.Errorf("autoXMax = %f, want 50", xMax)
	}

	yMin, yMax := sp.YRange()
	if yMin != 20 {
		t.Errorf("autoYMin = %f, want 20", yMin)
	}
	if yMax != 80 {
		t.Errorf("autoYMax = %f, want 80", yMax)
	}
}

func TestScatterPlotEmptyAutoRange(t *testing.T) {
	sp := NewScatterPlot()
	sp.SetAutoScale(true)
	// No points — should return 0,1
	xMin, xMax := sp.XRange()
	if xMin != 0 || xMax != 1 {
		t.Errorf("empty XRange = (%f,%f), want (0,1)", xMin, xMax)
	}
}

func TestScatterPlotMeasure(t *testing.T) {
	sp := NewScatterPlot()
	s := sp.Measure(Constraints{})
	if s.W < 10 {
		t.Errorf("W = %d, want >= 10", s.W)
	}
	if s.H < 5 {
		t.Errorf("H = %d, want >= 5", s.H)
	}
}

func TestScatterPlotPaint(t *testing.T) {
	sp := NewScatterPlot()
	sp.SetXRange(0, 100)
	sp.SetYRange(0, 100)
	sp.AddPoint(50, 50)
	sp.AddPoint(10, 90)
	sp.AddPoint(90, 10)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 15})

	buf := buffer.NewBuffer(30, 15)
	sp.Paint(buf)

	// Check border drawn
	if buf.GetCell(0, 0).Rune != '┌' {
		t.Error("top-left corner missing")
	}
	if buf.GetCell(29, 14).Rune != '┘' {
		t.Error("bottom-right corner missing")
	}

	// Count point characters in plot area
	pointCount := 0
	for y := 1; y < 14; y++ {
		for x := 2; x < 29; x++ {
			if buf.GetCell(x, y).Rune == '·' {
				pointCount++
			}
		}
	}
	if pointCount == 0 {
		t.Error("no data points rendered")
	}
}

func TestScatterPlotPaintEmpty(t *testing.T) {
	sp := NewScatterPlot()
	sp.SetAutoScale(true)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 15})
	buf := buffer.NewBuffer(30, 15)
	sp.Paint(buf) // should not panic with no points
}

func TestScatterPlotChildren(t *testing.T) {
	sp := NewScatterPlot()
	if sp.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestScatterPlotStyle(t *testing.T) {
	sp := NewScatterPlot()
	sp.SetStyle(ScatterPlotStyle{
		Point:     buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Axis:      buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Grid:      buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Border:    buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		PointChar: '*',
	})
	sp.SetXRange(0, 10)
	sp.SetYRange(0, 10)
	sp.AddPoint(5, 5)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	sp.Paint(buf)

	// Check star chars were drawn
	starFound := false
	for y := 1; y < 9; y++ {
		for x := 2; x < 19; x++ {
			if buf.GetCell(x, y).Rune == '*' {
				starFound = true
			}
		}
	}
	if !starFound {
		t.Error("no '*' point chars found")
	}
}

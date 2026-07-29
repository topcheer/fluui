package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestSankeyChartBasic(t *testing.T) {
	sc := NewSankeyChart()
	sc.AddFlow("Revenue", "Marketing", 500)
	sc.AddFlow("Revenue", "Engineering", 800)

	flows := sc.Flows()
	if len(flows) != 2 {
		t.Fatalf("Flows = %d, want 2", len(flows))
	}

	sources := sc.Sources()
	if len(sources) != 1 || sources[0] != "Revenue" {
		t.Errorf("Sources = %v, want [Revenue]", sources)
	}

	targets := sc.Targets()
	if len(targets) != 2 {
		t.Errorf("Targets = %d, want 2", len(targets))
	}
}

func TestSankeyChartSetFlows(t *testing.T) {
	sc := NewSankeyChart()
	flows := []SankeyFlow{
		{Source: "A", Target: "B", Value: 100},
		{Source: "A", Target: "C", Value: 200},
		{Source: "B", Target: "D", Value: 50},
	}
	sc.SetFlows(flows)
	if len(sc.Flows()) != 3 {
		t.Errorf("Flows = %d, want 3", len(sc.Flows()))
	}
}

func TestSankeyChartMeasure(t *testing.T) {
	sc := NewSankeyChart()
	s := sc.Measure(Constraints{})
	if s.W < 30 {
		t.Errorf("W = %d, want >= 30", s.W)
	}
	if s.H < 10 {
		t.Errorf("H = %d, want >= 10", s.H)
	}
}

func TestSankeyChartPaint(t *testing.T) {
	sc := NewSankeyChart()
	sc.AddFlow("Revenue", "Marketing", 500)
	sc.AddFlow("Revenue", "Engineering", 800)
	sc.AddFlow("Marketing", "Ads", 300)
	sc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 15})

	buf := buffer.NewBuffer(50, 15)
	sc.Paint(buf)

	// Check that some cells were drawn
	drawnCount := 0
	for y := 0; y < 15; y++ {
		for x := 0; x < 50; x++ {
			if buf.GetCell(x, y).Rune != 0 {
				drawnCount++
			}
		}
	}
	if drawnCount == 0 {
		t.Error("no cells were drawn in Paint")
	}
}

func TestSankeyChartPaintEmpty(t *testing.T) {
	sc := NewSankeyChart()
	sc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 15})

	buf := buffer.NewBuffer(50, 15)
	sc.Paint(buf) // should not panic with empty flows
}

func TestSankeyChartChildren(t *testing.T) {
	sc := NewSankeyChart()
	if sc.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestSankeyChartStyle(t *testing.T) {
	sc := NewSankeyChart()
	custom := SankeyChartStyle{
		Node:      buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Flow:      buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Label:     buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		NodeWidth: 5,
	}
	sc.SetStyle(custom)
	sc.AddFlow("A", "B", 100)
	sc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 15})
	buf := buffer.NewBuffer(50, 15)
	sc.Paint(buf) // should not panic
}

func TestSankeyFlowStruct(t *testing.T) {
	f := SankeyFlow{Source: "src", Target: "tgt", Value: 42}
	if f.Source != "src" || f.Target != "tgt" || f.Value != 42 {
		t.Errorf("SankeyFlow = %+v", f)
	}
}

package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestTreemapChart_New_P454(t *testing.T) {
	tc := NewTreemapChart()
	if tc.NodeCount() != 0 {
		t.Errorf("NodeCount = %d, want 0", tc.NodeCount())
	}
}

func TestTreemapChart_AddNode_P454(t *testing.T) {
	tc := NewTreemapChart()
	tc.AddNode(TreemapNode{Label: "A", Value: 40})
	tc.AddNode(TreemapNode{Label: "B", Value: 30})
	if tc.NodeCount() != 2 {
		t.Errorf("NodeCount = %d, want 2", tc.NodeCount())
	}
	if tc.TotalValue() != 70 {
		t.Errorf("TotalValue = %v, want 70", tc.TotalValue())
	}
}

func TestTreemapChart_SetNodes_P454(t *testing.T) {
	tc := NewTreemapChart()
	tc.SetNodes([]TreemapNode{{Label: "X", Value: 10}, {Label: "Y", Value: 20}})
	if tc.NodeCount() != 2 {
		t.Errorf("NodeCount = %d, want 2", tc.NodeCount())
	}
}

func TestTreemapChart_Nodes_P454(t *testing.T) {
	tc := NewTreemapChart()
	tc.AddNode(TreemapNode{Label: "X", Value: 10})
	nodes := tc.Nodes()
	if len(nodes) != 1 || nodes[0].Label != "X" {
		t.Errorf("Nodes mismatch: %v", nodes)
	}
}

func TestTreemapChart_AutoColor_P454(t *testing.T) {
	tc := NewTreemapChart()
	tc.AddNode(TreemapNode{Label: "X", Value: 10})
	nodes := tc.Nodes()
	if nodes[0].Color.Type == 0 {
		t.Error("color should be auto-assigned")
	}
}

func TestTreemapChart_Clear_P454(t *testing.T) {
	tc := NewTreemapChart()
	tc.AddNode(TreemapNode{Label: "X", Value: 10})
	tc.Clear()
	if tc.NodeCount() != 0 {
		t.Error("should have 0 nodes after Clear")
	}
}

func TestTreemapChart_Style_P454(t *testing.T) {
	tc := NewTreemapChart()
	st := DefaultTreemapChartStyle()
	tc.SetStyle(st)
}

func TestTreemapChart_Measure_P454(t *testing.T) {
	tc := NewTreemapChart()
	sz := tc.Measure(Constraints{})
	if sz.W < 10 || sz.H < 10 {
		t.Errorf("size too small: %v", sz)
	}
}

func TestTreemapChart_Paint_NoPanic_P454(t *testing.T) {
	tc := NewTreemapChart()
	tc.AddNode(TreemapNode{Label: "Documents", Value: 40})
	tc.AddNode(TreemapNode{Label: "Photos", Value: 30})
	tc.AddNode(TreemapNode{Label: "Videos", Value: 20})
	tc.AddNode(TreemapNode{Label: "Music", Value: 10})
	tc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 15})
	buf := buffer.NewBuffer(50, 15)
	tc.Paint(buf)
}

func TestTreemapChart_Paint_ZeroBounds_P454(t *testing.T) {
	tc := NewTreemapChart()
	tc.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	tc.Paint(buf)
}

func TestTreemapChart_Paint_Empty_P454(t *testing.T) {
	tc := NewTreemapChart()
	tc.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	tc.Paint(buf)
}

func TestTreemapChart_Layout_Sorting_P454(t *testing.T) {
	tc := NewTreemapChart()
	tc.AddNode(TreemapNode{Label: "Small", Value: 5})
	tc.AddNode(TreemapNode{Label: "Big", Value: 50})
	tc.AddNode(TreemapNode{Label: "Medium", Value: 20})
	tc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 15})
	buf := buffer.NewBuffer(40, 15)
	tc.Paint(buf) // should sort by value desc internally
}

func TestTreemapChart_Children_P454(t *testing.T) {
	if NewTreemapChart().Children() != nil {
		t.Error("Children should be nil")
	}
}

func BenchmarkTreemapChart_Paint_P454(b *testing.B) {
	tc := NewTreemapChart()
	for i, v := range []int{40, 30, 20, 15, 10, 5} {
		tc.AddNode(TreemapNode{Label: "node", Value: float64(v)})
		_ = i
	}
	tc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 15})
	buf := buffer.NewBuffer(50, 15)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tc.Paint(buf)
	}
}

package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestNetworkGraph_New_P444(t *testing.T) {
	ng := NewNetworkGraph()
	if ng.NodeCount() != 0 || ng.EdgeCount() != 0 {
		t.Errorf("counts = %d/%d, want 0/0", ng.NodeCount(), ng.EdgeCount())
	}
}

func TestNetworkGraph_AddNode_P444(t *testing.T) {
	ng := NewNetworkGraph()
	ng.AddNode(GraphNode{ID: "a", Label: "Server"})
	ng.AddNode(GraphNode{ID: "b", Label: "DB"})
	if ng.NodeCount() != 2 {
		t.Errorf("NodeCount = %d, want 2", ng.NodeCount())
	}
}

func TestNetworkGraph_AddEdge_P444(t *testing.T) {
	ng := NewNetworkGraph()
	ng.AddNode(GraphNode{ID: "a", Label: "A"})
	ng.AddNode(GraphNode{ID: "b", Label: "B"})
	ng.AddEdge(GraphEdge{From: "a", To: "b"})
	if ng.EdgeCount() != 1 {
		t.Errorf("EdgeCount = %d, want 1", ng.EdgeCount())
	}
}

func TestNetworkGraph_SetNodes_P444(t *testing.T) {
	ng := NewNetworkGraph()
	ng.SetNodes([]GraphNode{{ID: "x"}, {ID: "y"}, {ID: "z"}})
	if ng.NodeCount() != 3 {
		t.Errorf("NodeCount = %d, want 3", ng.NodeCount())
	}
}

func TestNetworkGraph_SetEdges_P444(t *testing.T) {
	ng := NewNetworkGraph()
	ng.SetEdges([]GraphEdge{{From: "a", To: "b"}, {From: "b", To: "c"}})
	if ng.EdgeCount() != 2 {
		t.Errorf("EdgeCount = %d, want 2", ng.EdgeCount())
	}
}

func TestNetworkGraph_Nodes_P444(t *testing.T) {
	ng := NewNetworkGraph()
	ng.AddNode(GraphNode{ID: "x", Label: "X"})
	nodes := ng.Nodes()
	if len(nodes) != 1 || nodes[0].ID != "x" {
		t.Errorf("Nodes mismatch: %v", nodes)
	}
}

func TestNetworkGraph_Edges_P444(t *testing.T) {
	ng := NewNetworkGraph()
	ng.AddEdge(GraphEdge{From: "a", To: "b"})
	edges := ng.Edges()
	if len(edges) != 1 || edges[0].From != "a" {
		t.Errorf("Edges mismatch: %v", edges)
	}
}

func TestNetworkGraph_Clear_P444(t *testing.T) {
	ng := NewNetworkGraph()
	ng.AddNode(GraphNode{ID: "x"})
	ng.AddEdge(GraphEdge{From: "x", To: "y"})
	ng.Clear()
	if ng.NodeCount() != 0 || ng.EdgeCount() != 0 {
		t.Error("should have 0 nodes/edges after Clear")
	}
}

func TestNetworkGraph_Style_P444(t *testing.T) {
	ng := NewNetworkGraph()
	st := DefaultNetworkGraphStyle()
	ng.SetStyle(st)
	if ng.Style().Node.Fg != st.Node.Fg {
		t.Error("style mismatch")
	}
}

func TestNetworkGraph_Measure_P444(t *testing.T) {
	ng := NewNetworkGraph()
	sz := ng.Measure(Constraints{})
	if sz.W < 10 || sz.H < 10 {
		t.Errorf("size too small: %v", sz)
	}
}

func TestNetworkGraph_Paint_NoPanic_P444(t *testing.T) {
	ng := NewNetworkGraph()
	ng.AddNode(GraphNode{ID: "srv", Label: "Server"})
	ng.AddNode(GraphNode{ID: "db", Label: "Database"})
	ng.AddNode(GraphNode{ID: "cache", Label: "Cache"})
	ng.AddEdge(GraphEdge{From: "srv", To: "db"})
	ng.AddEdge(GraphEdge{From: "srv", To: "cache"})
	ng.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 15})
	buf := buffer.NewBuffer(40, 15)
	ng.Paint(buf)
}

func TestNetworkGraph_Paint_ZeroBounds_P444(t *testing.T) {
	ng := NewNetworkGraph()
	ng.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	ng.Paint(buf)
}

func TestNetworkGraph_Paint_Empty_P444(t *testing.T) {
	ng := NewNetworkGraph()
	ng.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	ng.Paint(buf)
}

func TestNetworkGraph_Children_P444(t *testing.T) {
	if NewNetworkGraph().Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestNetworkGraph_AutoColor_P444(t *testing.T) {
	ng := NewNetworkGraph()
	ng.AddNode(GraphNode{ID: "x", Label: "X"})
	nodes := ng.Nodes()
	if nodes[0].Color.Type == 0 {
		t.Error("color should be auto-assigned")
	}
}

func BenchmarkNetworkGraph_Paint_P444(b *testing.B) {
	ng := NewNetworkGraph()
	for i := 0; i < 8; i++ {
		ng.AddNode(GraphNode{ID: "n", Label: "node"})
	}
	ng.AddEdge(GraphEdge{From: "n", To: "n"})
	ng.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 15})
	buf := buffer.NewBuffer(40, 15)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ng.Paint(buf)
	}
}

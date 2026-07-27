package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestOrgChart_New_P448(t *testing.T) {
	oc := NewOrgChart()
	if oc.NodeCount() != 0 {
		t.Errorf("NodeCount = %d, want 0", oc.NodeCount())
	}
	if oc.HasRoot() {
		t.Error("should not have root")
	}
	if oc.Depth() != 0 {
		t.Errorf("Depth = %d, want 0", oc.Depth())
	}
}

func TestOrgChart_SetRoot_P448(t *testing.T) {
	oc := NewOrgChart()
	oc.SetRoot(OrgNode{ID: "ceo", Label: "CEO"})
	if !oc.HasRoot() {
		t.Error("should have root")
	}
	if oc.NodeCount() != 1 {
		t.Errorf("NodeCount = %d, want 1", oc.NodeCount())
	}
	if oc.Depth() != 1 {
		t.Errorf("Depth = %d, want 1", oc.Depth())
	}
}

func TestOrgChart_AddChild_P448(t *testing.T) {
	oc := NewOrgChart()
	oc.SetRoot(OrgNode{ID: "ceo", Label: "CEO"})
	oc.AddChild("ceo", OrgNode{ID: "cto", Label: "CTO"})
	oc.AddChild("ceo", OrgNode{ID: "cfo", Label: "CFO"})
	oc.AddChild("cto", OrgNode{ID: "dev", Label: "Dev"})
	if oc.NodeCount() != 4 {
		t.Errorf("NodeCount = %d, want 4", oc.NodeCount())
	}
	if oc.Depth() != 3 {
		t.Errorf("Depth = %d, want 3", oc.Depth())
	}
}

func TestOrgChart_AddChild_InvalidParent_P448(t *testing.T) {
	oc := NewOrgChart()
	oc.SetRoot(OrgNode{ID: "root", Label: "Root"})
	oc.AddChild("nonexistent", OrgNode{ID: "x", Label: "X"})
	if oc.NodeCount() != 1 {
		t.Errorf("NodeCount = %d, want 1 (invalid parent ignored)", oc.NodeCount())
	}
}

func TestOrgChart_AutoColor_P448(t *testing.T) {
	oc := NewOrgChart()
	oc.SetRoot(OrgNode{ID: "root", Label: "Root"})
	// Color auto-assigned (checking via Paint, not direct field access)
	oc.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	oc.Paint(buf) // should not panic
}

func TestOrgChart_Style_P448(t *testing.T) {
	oc := NewOrgChart()
	st := DefaultOrgChartStyle()
	oc.SetStyle(st)
	if oc.Style().Node.Fg != st.Node.Fg {
		t.Error("style mismatch")
	}
}

func TestOrgChart_Measure_P448(t *testing.T) {
	oc := NewOrgChart()
	oc.SetRoot(OrgNode{ID: "r", Label: "R"})
	oc.AddChild("r", OrgNode{ID: "a", Label: "A"})
	oc.AddChild("r", OrgNode{ID: "b", Label: "B"})
	sz := oc.Measure(Constraints{})
	if sz.H < 5 {
		t.Errorf("H = %d, want >= 5", sz.H)
	}
}

func TestOrgChart_Paint_NoPanic_P448(t *testing.T) {
	oc := NewOrgChart()
	oc.SetRoot(OrgNode{ID: "ceo", Label: "CEO"})
	oc.AddChild("ceo", OrgNode{ID: "cto", Label: "CTO"})
	oc.AddChild("ceo", OrgNode{ID: "cfo", Label: "CFO"})
	oc.AddChild("cto", OrgNode{ID: "dev", Label: "Dev"})
	oc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	oc.Paint(buf)
}

func TestOrgChart_Paint_NoRoot_P448(t *testing.T) {
	oc := NewOrgChart()
	oc.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	oc.Paint(buf) // no root, no-op
}

func TestOrgChart_Paint_ZeroBounds_P448(t *testing.T) {
	oc := NewOrgChart()
	oc.SetRoot(OrgNode{ID: "r", Label: "R"})
	oc.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	oc.Paint(buf)
}

func TestOrgChart_Children_P448(t *testing.T) {
	if NewOrgChart().Children() != nil {
		t.Error("Children should be nil")
	}
}

func BenchmarkOrgChart_Paint_P448(b *testing.B) {
	oc := NewOrgChart()
	oc.SetRoot(OrgNode{ID: "ceo", Label: "CEO"})
	oc.AddChild("ceo", OrgNode{ID: "cto", Label: "CTO"})
	oc.AddChild("ceo", OrgNode{ID: "cfo", Label: "CFO"})
	oc.AddChild("cto", OrgNode{ID: "dev", Label: "Dev"})
	oc.AddChild("cto", OrgNode{ID: "ops", Label: "Ops"})
	oc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 12})
	buf := buffer.NewBuffer(60, 12)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		oc.Paint(buf)
	}
}

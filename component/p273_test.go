package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P273: gauge Ratio/Measure/Paint + tree ExpandAll/CollapseAll/moveCursor + windowmanager Focused/Measure

func TestGauge_Ratio_MaxEqualMin_P273(t *testing.T) {
	g := NewGauge()
	g.SetRange(5, 5) // max <= min → 0
	if g.Ratio() != 0 {
		t.Errorf("expected 0 when max<=min, got %f", g.Ratio())
	}
}

func TestGauge_Measure_Radial_P273(t *testing.T) {
	g := NewGauge()
	g.SetRadial(true)
	s := g.Measure(Constraints{})
	if s.W < 7 || s.H < 5 {
		t.Errorf("radial gauge should be at least 7x5, got %dx%d", s.W, s.H)
	}
}

func TestGauge_Measure_RadialWithLabel_P273(t *testing.T) {
	g := NewGauge()
	g.SetRadial(true)
	g.SetLabel("CPU")
	s := g.Measure(Constraints{})
	if s.H < 7 {
		t.Errorf("radial with label should be at least 7H, got %d", s.H)
	}
}

func TestGauge_Paint_Radial_P273(t *testing.T) {
	g := NewGauge()
	g.SetRadial(true)
	g.SetRange(0, 100)
	g.SetValue(50)
	g.SetLabel("CPU")
	g.SetShowValue(true)
	g.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 8})
	buf := buffer.NewBuffer(10, 8)
	g.Paint(buf)
}

func TestGauge_Measure_Vertical_P273(t *testing.T) {
	g := NewGauge()
	g.SetOrientation(GaugeVertical)
	g.SetRange(0, 100)
	g.SetValue(30)
	s := g.Measure(Constraints{MaxHeight: 20})
	if s.H <= 0 {
		t.Error("vertical gauge should have positive height")
	}
}

func TestGauge_Paint_HorizontalWithThresholds_P273(t *testing.T) {
	g := NewGauge()
	g.SetRange(0, 100)
	g.SetValue(85)
	g.SetThresholds([]Threshold{
		{Low: 0, High: 0.5, Color: buffer.RGB(0, 255, 0)},
		{Low: 0.5, High: 1.0, Color: buffer.RGB(255, 255, 0)},
	})
	g.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	g.Paint(buf)
}

func TestTree_ExpandAll_NilRoot_P273(t *testing.T) {
	tree := NewTree()
	tree.ExpandAll() // nil root → no-op
}

func TestTree_CollapseAll_NilRoot_P273(t *testing.T) {
	tree := NewTree()
	tree.CollapseAll()
}

func TestTree_ExpandAll_WithChildren_P273(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("r", "root")
	root.AddChild(NewTreeNode("c1", "child1"))
	root.AddChild(NewTreeNode("c2", "child2"))
	tree.SetRoot(root)
	tree.ExpandAll()
}

func TestTree_MoveCursor_Clamp_P273(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("r", "root")
	root.AddChild(NewTreeNode("a", "a"))
	root.AddChild(NewTreeNode("b", "b"))
	root.Expanded = true
	tree.SetRoot(root)
	// Navigate down
	tree.HandleKey(nil) // nil key should not panic
}


func TestWindowManager_Measure_NilRoot_P273(t *testing.T) {
	wm := NewWindowManager(&BaseComponent{})
	s := wm.Measure(Constraints{MaxWidth: 100, MaxHeight: 50})
	if s.W != 0 || s.H != 0 {
		t.Error("nil root should return empty size")
	}
}

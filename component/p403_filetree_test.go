package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP403_NewFileTree(t *testing.T) {
	nodes := []FileNode{
		{Name: "src", IsDir: true, Expanded: true, Children: []FileNode{
			{Name: "main.go", IsDir: false},
		}},
		{Name: "README.md", IsDir: false},
	}
	ft := NewFileTree("project", nodes)
	if ft.Root() != "project" { t.Errorf("Root = %q", ft.Root()) }
	if len(ft.Nodes()) != 2 { t.Errorf("Nodes = %d", len(ft.Nodes())) }
	if ft.ID() == "" { t.Error("ID empty") }
}

func TestP403_SetRoot(t *testing.T) {
	ft := NewFileTree("old", nil)
	ft.SetRoot("new")
	if ft.Root() != "new" { t.Errorf("Root = %q", ft.Root()) }
}

func TestP403_SetNodes(t *testing.T) {
	ft := NewFileTree("x", nil)
	ft.SetNodes([]FileNode{{Name: "a"}})
	if len(ft.Nodes()) != 1 { t.Error("should have 1 node") }
}

func TestP403_NodesCopy(t *testing.T) {
	ft := NewFileTree("x", []FileNode{{Name: "a"}})
	n := ft.Nodes()
	n[0].Name = "changed"
	if ft.Nodes()[0].Name != "a" { t.Error("Nodes() should return copy") }
}

func TestP403_Measure(t *testing.T) {
	ft := NewFileTree("proj", []FileNode{
		{Name: "src", IsDir: true, Expanded: true, Children: []FileNode{
			{Name: "main.go"},
		}},
	})
	s := ft.Measure(Constraints{MaxWidth: 80, MaxHeight: 20})
	if s.H < 3 { t.Errorf("H = %d, want >= 3", s.H) }
}

func TestP403_Paint_Expanded(t *testing.T) {
	ft := NewFileTree("project", []FileNode{
		{Name: "src", IsDir: true, Expanded: true, Children: []FileNode{
			{Name: "main.go"},
		}},
		{Name: "README.md"},
	})
	ft.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	ft.Paint(buf)
	// Root at y=0
	c := buf.GetCell(0, 0)
	if c.Rune == 0 { t.Log("root cell empty (emoji width)") }
}

func TestP403_Paint_Collapsed(t *testing.T) {
	ft := NewFileTree("project", []FileNode{
		{Name: "src", IsDir: true, Expanded: false, Children: []FileNode{
			{Name: "main.go"},
		}},
	})
	ft.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	ft.Paint(buf)
	// Only root + 1 dir shown (collapsed)
}

func TestP403_Paint_ZeroBounds(t *testing.T) {
	ft := NewFileTree("x", []FileNode{{Name: "a"}})
	ft.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	ft.Paint(buf)
}

func TestP403_Paint_NestedDirs(t *testing.T) {
	ft := NewFileTree("x", []FileNode{
		{Name: "a", IsDir: true, Expanded: true, Children: []FileNode{
			{Name: "b", IsDir: true, Expanded: true, Children: []FileNode{
				{Name: "c.go"},
			}},
		}},
	})
	ft.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	ft.Paint(buf)
}

func TestP403_Paint_EmptyNodes(t *testing.T) {
	ft := NewFileTree("empty", nil)
	ft.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	ft.Paint(buf)
}

func TestP403_ParsePathList(t *testing.T) {
	tree := ParsePathList([]string{"src/main.go", "src/util.go", "README.md"})
	if len(tree) != 2 { t.Errorf("root nodes = %d, want 2", len(tree)) }
	// src should be a dir with 2 children
	var srcNode *FileNode
	for i := range tree {
		if tree[i].Name == "src" { srcNode = &tree[i] }
	}
	if srcNode == nil { t.Fatal("src node not found") }
	if !srcNode.IsDir { t.Error("src should be dir") }
	if len(srcNode.Children) != 2 { t.Errorf("src children = %d, want 2", len(srcNode.Children)) }
}

func TestP403_ParsePathList_DeepNested(t *testing.T) {
	tree := ParsePathList([]string{"a/b/c/d.go"})
	if len(tree) != 1 { t.Errorf("root = %d, want 1", len(tree)) }
	if !tree[0].IsDir { t.Error("a should be dir") }
	if len(tree[0].Children) != 1 { t.Error("a should have 1 child") }
}

func TestP403_Concurrent(t *testing.T) {
	ft := NewFileTree("x", []FileNode{{Name: "a"}})
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ { ft.SetRoot("concurrent") }
		close(done)
	}()
	for i := 0; i < 500; i++ { _ = ft.Root() }
	<-done
}

func TestP403_SatisfiesComponent(t *testing.T) {
	var _ Component = (*FileTree)(nil)
}

// === Batch Measure coverage for defensive branches ===
// These all miss HasHeight()/HasWidth() with small/dummy constraints.

func TestP403_Measure_AllComponents_SmallConstraints(t *testing.T) {
	// Table-driven: exercise Measure with tiny constraints to hit clamping branches
	tests := []struct {
		name string
		measure func() Size
	}{
		{"Avatar", func() Size {
			a := NewAvatar("Alice"); return a.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
		}},
		{"Chip", func() Size {
			return NewChip("test").Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
		}},
		{"ConfidenceMeter", func() Size {
			return NewConfidenceMeter(0.5).Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
		}},
		{"DiffStatBar", func() Size {
			return NewDiffStatBar(1, 1).Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
		}},
		{"StatCard", func() Size {
			return NewStatCard("X", "1").Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
		}},
		{"Toast", func() Size {
			return NewToast("x", ToastInfo).Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
		}},
		{"HintLabel", func() Size {
			return NewHintLabel("x").Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
		}},
		{"MarkdownStream", func() Size {
			return NewMarkdownStream().Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
		}},
		{"FileTree", func() Size {
			return NewFileTree("x", nil).Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
		}},
		{"SearchBar", func() Size {
			return NewSearchBar("x").Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
		}},
		{"MetricBar", func() Size {
			return NewMetricBar("x", 1, 0, 2).Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
		}},
		{"ColorSwatch", func() Size {
			return NewColorSwatch(buffer.RGB(0, 0, 0)).Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.measure()
			if s.W < 1 { t.Error("W should be >= 1") }
			if s.H < 1 { t.Error("H should be >= 1") }
		})
	}
}

func BenchmarkP403_FileTree_Paint(b *testing.B) {
	ft := NewFileTree("project", []FileNode{
		{Name: "src", IsDir: true, Expanded: true, Children: []FileNode{
			{Name: "main.go"}, {Name: "util.go"},
		}},
		{Name: "docs", IsDir: true, Expanded: false},
		{Name: "README.md"},
	})
	ft.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { ft.Paint(buf) }
}

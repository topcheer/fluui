package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// --- Test helpers ---

func keyTree(k term.KeyCode) *term.KeyEvent {
	return &term.KeyEvent{Key: k}
}

// buildTestTree creates a tree for testing:
// root/
//   ├── src/
//   │   ├── main.go
//   │   └── utils.go
//   ├── docs/
//   │   └── readme.md
//   └── go.mod
func buildTestTree() *TreeNode {
	root := NewTreeNode("root", "root")
	root.Expanded = true

	src := NewTreeNode("src", "src")
	src.Expanded = true
	src.AddChild(NewTreeNode("main", "main.go"))
	src.AddChild(NewTreeNode("utils", "utils.go"))

	docs := NewTreeNode("docs", "docs")
	docs.AddChild(NewTreeNode("readme", "readme.md"))

	goMod := NewTreeNode("gomod", "go.mod")

	root.AddChild(src)
	root.AddChild(docs)
	root.AddChild(goMod)
	return root
}

// --- Tests ---

func TestTree_New(t *testing.T) {
	tree := NewTree()
	if tree.VisibleCount() != 0 {
		t.Errorf("VisibleCount: got %d, want 0", tree.VisibleCount())
	}
	if tree.CurrentNode() != nil {
		t.Error("CurrentNode: expected nil for empty tree")
	}
}

func TestTree_SetRoot(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("root", "Root")
	tree.SetRoot(root)
	if tree.VisibleCount() != 1 {
		t.Errorf("VisibleCount: got %d, want 1 (only root, children collapsed)", tree.VisibleCount())
	}
	if tree.CurrentNode() == nil {
		t.Error("CurrentNode should not be nil after SetRoot")
	}
}

func TestTree_SetRootExpanded(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	// root(expanded) + src(expanded) + main + utils + docs + gomod = 6 visible
	if tree.VisibleCount() != 6 {
		t.Errorf("VisibleCount: got %d, want 6", tree.VisibleCount())
	}
}

func TestTreeNode_AddChild(t *testing.T) {
	parent := NewTreeNode("p", "parent")
	child := NewTreeNode("c", "child")
	parent.AddChild(child)
	if len(parent.Children) != 1 {
		t.Fatalf("Children: got %d, want 1", len(parent.Children))
	}
	if child.parent != parent {
		t.Error("child.parent should be set")
	}
}

func TestTreeNode_IsLeaf(t *testing.T) {
	leaf := NewTreeNode("l", "leaf")
	if !leaf.IsLeaf() {
		t.Error("leaf should be IsLeaf")
	}
	parent := NewTreeNode("p", "parent")
	parent.AddChild(NewTreeNode("c", "child"))
	if parent.IsLeaf() {
		t.Error("parent should not be IsLeaf")
	}
}

func TestTreeNode_HasChildren(t *testing.T) {
	n := NewTreeNode("n", "n")
	if n.HasChildren() {
		t.Error("empty node should not have children")
	}
	n.AddChild(NewTreeNode("c", "c"))
	if !n.HasChildren() {
		t.Error("node with child should have children")
	}
}

func TestTree_NavDown(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	// Start at root (index 0)
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	tree.HandleKey(keyTree(term.KeyDown))
	node := tree.CurrentNode()
	if node == nil || node.ID != "src" {
		t.Errorf("After Down: ID=%v, want src", node)
	}
}

func TestTree_NavUp(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// Move down first
	tree.HandleKey(keyTree(term.KeyDown))
	tree.HandleKey(keyTree(term.KeyDown))
	// Now at main.go
	// Move up
	tree.HandleKey(keyTree(term.KeyUp))
	node := tree.CurrentNode()
	if node == nil || node.ID != "src" {
		t.Errorf("After Up: ID=%v, want src", node)
	}
}

func TestTree_NavUpAtTop(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	tree.HandleKey(keyTree(term.KeyUp)) // already at top
	node := tree.CurrentNode()
	if node == nil || node.ID != "root" {
		t.Errorf("Up at top: ID=%v, want root", node)
	}
}

func TestTree_NavDownAtBottom(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	tree.SetCurrent(tree.VisibleCount() - 1)
	tree.HandleKey(keyTree(term.KeyDown)) // at bottom
	idx := tree.CurrentIndex()
	if idx != tree.VisibleCount()-1 {
		t.Errorf("Down at bottom: idx=%d, want %d", idx, tree.VisibleCount()-1)
	}
}

func TestTree_ExpandCollapsed(t *testing.T) {
	tree := NewTree()
	// Build a tree where root is NOT expanded
	root := NewTreeNode("root", "root")
	root.AddChild(NewTreeNode("c1", "child1"))
	root.AddChild(NewTreeNode("c2", "child2"))
	tree.SetRoot(root)
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	if tree.VisibleCount() != 1 {
		t.Fatalf("before expand: VisibleCount=%d, want 1", tree.VisibleCount())
	}

	// Right to expand
	tree.HandleKey(keyTree(term.KeyRight))
	if tree.VisibleCount() != 3 {
		t.Errorf("after expand: VisibleCount=%d, want 3", tree.VisibleCount())
	}
}

func TestTree_CollapseExpanded(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// root is expanded with 6 visible
	tree.HandleKey(keyTree(term.KeyLeft)) // collapse root
	if tree.VisibleCount() != 1 {
		t.Errorf("after collapse: VisibleCount=%d, want 1", tree.VisibleCount())
	}
}

func TestTree_RightOnExpandedMovesToChild(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// root is already expanded → Right should descend to first child
	tree.HandleKey(keyTree(term.KeyRight))
	node := tree.CurrentNode()
	if node == nil || node.ID != "src" {
		t.Errorf("Right on expanded: ID=%v, want src", node)
	}
}

func TestTree_LeftOnCollapsedMovesToParent(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// Navigate to a leaf (main.go)
	tree.SetCurrent(2) // main.go
	tree.HandleKey(keyTree(term.KeyLeft)) // leaf → move to parent
	node := tree.CurrentNode()
	if node == nil || node.ID != "src" {
		t.Errorf("Left on leaf: ID=%v, want src (parent)", node)
	}
}

func TestTree_LeftOnExpandedCollapses(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	tree.SetCurrent(1) // src (expanded)
	tree.HandleKey(keyTree(term.KeyLeft))
	node := tree.CurrentNode()
	if node == nil || node.ID != "src" {
		t.Errorf("After collapse: ID=%v, want src (stayed)", node)
	}
	// src should now be collapsed
	if node.Expanded {
		t.Error("src should be collapsed")
	}
}

func TestTree_SpaceSelectsNode(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	var selected *TreeNode
	tree.OnSelect(func(n *TreeNode) {
		selected = n
	})

	tree.HandleKey(keyTree(term.KeySpace))
	if selected == nil || selected.ID != "root" {
		t.Errorf("OnSelect: ID=%v, want root", selected)
	}
	if !selected.Selected {
		t.Error("root should be Selected=true")
	}
}

func TestTree_EnterSelectsNode(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	var selected *TreeNode
	tree.OnSelect(func(n *TreeNode) {
		selected = n
	})

	// Navigate to main.go and Enter
	tree.SetCurrent(2)
	tree.HandleKey(keyTree(term.KeyEnter))
	if selected == nil || selected.ID != "main" {
		t.Errorf("Enter select: ID=%v, want main", selected)
	}
}

func TestTree_OnToggleCallback(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("root", "root")
	root.AddChild(NewTreeNode("c1", "child1"))
	tree.SetRoot(root)
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	var toggled *TreeNode
	tree.OnToggle(func(n *TreeNode) {
		toggled = n
	})

	tree.HandleKey(keyTree(term.KeyRight))
	if toggled == nil || toggled.ID != "root" {
		t.Errorf("OnToggle: ID=%v, want root", toggled)
	}
	if !toggled.Expanded {
		t.Error("root should be expanded after toggle")
	}
}

func TestTree_ToggleExpandMethod(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("root", "root")
	root.Expanded = true
	root.AddChild(NewTreeNode("c1", "child1"))
	tree.SetRoot(root)
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	tree.ToggleExpand() // collapse root
	if root.Expanded {
		t.Error("root should be collapsed after ToggleExpand")
	}
	if tree.VisibleCount() != 1 {
		t.Errorf("VisibleCount: got %d, want 1", tree.VisibleCount())
	}
}

func TestTree_ToggleExpandOnLeaf(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// Navigate to leaf (gomod at index 5)
	tree.SetCurrent(5)
	before := tree.VisibleCount()
	tree.ToggleExpand()
	if tree.VisibleCount() != before {
		t.Errorf("ToggleExpand on leaf should not change visible count: before=%d after=%d", before, tree.VisibleCount())
	}
}

func TestTree_ExpandAll(t *testing.T) {
	tree := NewTree()
	// All collapsed initially
	root := NewTreeNode("root", "root")
	src := NewTreeNode("src", "src")
	src.AddChild(NewTreeNode("main", "main.go"))
	root.AddChild(src)
	tree.SetRoot(root)
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	if tree.VisibleCount() != 1 {
		t.Fatalf("before: VisibleCount=%d, want 1", tree.VisibleCount())
	}
	tree.ExpandAll()
	if tree.VisibleCount() != 3 {
		t.Errorf("after ExpandAll: VisibleCount=%d, want 3", tree.VisibleCount())
	}
}

func TestTree_CollapseAll(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	if tree.VisibleCount() != 6 {
		t.Fatalf("before: VisibleCount=%d, want 6", tree.VisibleCount())
	}
	tree.CollapseAll()
	if tree.VisibleCount() != 1 {
		t.Errorf("after CollapseAll: VisibleCount=%d, want 1", tree.VisibleCount())
	}
	if tree.CurrentIndex() != 0 {
		t.Errorf("CurrentIndex after CollapseAll: %d, want 0", tree.CurrentIndex())
	}
}

func TestTree_HomeKey(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	tree.SetCurrent(4)
	tree.HandleKey(keyTree(term.KeyHome))
	node := tree.CurrentNode()
	if node == nil || node.ID != "root" {
		t.Errorf("Home: ID=%v, want root", node)
	}
}

func TestTree_EndKey(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	tree.HandleKey(keyTree(term.KeyEnd))
	idx := tree.CurrentIndex()
	if idx != tree.VisibleCount()-1 {
		t.Errorf("End: idx=%d, want %d", idx, tree.VisibleCount()-1)
	}
}

func TestTree_PageDown(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})

	tree.HandleKey(keyTree(term.KeyPageDown))
	idx := tree.CurrentIndex()
	if idx != 3 {
		t.Errorf("PageDown: idx=%d, want 3 (0+3)", idx)
	}
}

func TestTree_PageUp(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})

	tree.SetCurrent(5)
	tree.HandleKey(keyTree(term.KeyPageUp))
	idx := tree.CurrentIndex()
	if idx != 2 {
		t.Errorf("PageUp: idx=%d, want 2 (5-3)", idx)
	}
}

func TestTree_Path(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// Navigate to main.go (index 2)
	tree.SetCurrent(2)
	path := tree.Path()
	if len(path) != 3 {
		t.Fatalf("Path len: got %d, want 3", len(path))
	}
	wantPath := []string{"root", "src", "main"}
	for i, p := range path {
		if p != wantPath[i] {
			t.Errorf("Path[%d]: got %q, want %q", i, p, wantPath[i])
		}
	}
}

func TestTree_PathLabels(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	tree.SetCurrent(5) // go.mod
	labels := tree.PathLabels()
	if len(labels) != 2 {
		t.Fatalf("PathLabels len: got %d, want 2", len(labels))
	}
	want := []string{"root", "go.mod"}
	for i, l := range labels {
		if l != want[i] {
			t.Errorf("PathLabels[%d]: got %q, want %q", i, l, want[i])
		}
	}
}

func TestTree_PathRootNode(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// At root
	path := tree.Path()
	if len(path) != 1 || path[0] != "root" {
		t.Errorf("Path at root: got %v, want [root]", path)
	}
}

func TestTree_PathEmpty(t *testing.T) {
	tree := NewTree()
	path := tree.Path()
	if path != nil {
		t.Errorf("Path on empty tree: got %v, want nil", path)
	}
}

func TestTree_PaintBasic(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	buf := buffer.NewBuffer(40, 10)
	tree.Paint(buf)

	// Root line should contain "root"
	cell := buf.GetCell(2, 0) // after "▼ " prefix
	if cell.Rune != 'r' {
		t.Errorf("Expected 'r' at (2,0), got %q", string(cell.Rune))
	}
}

func TestTree_PaintCursorHighlight(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	buf := buffer.NewBuffer(20, 10)
	tree.Paint(buf)

	// Cursor is at row 0 — should have reverse flag
	cell := buf.GetCell(0, 0)
	if cell.Flags&buffer.Reverse == 0 {
		t.Error("Cursor row should have Reverse flag")
	}
}

func TestTree_PaintEmpty(t *testing.T) {
	tree := NewTree()
	tree.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})

	buf := buffer.NewBuffer(20, 5)
	// Should not panic
	tree.Paint(buf)
}

func TestTree_PaintTooSmall(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})

	buf := buffer.NewBuffer(1, 1)
	// Should not panic
	tree.Paint(buf)
}

func TestTree_PaintIcons(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("root", "root")
	root.Icon = "📁"
	root.Expanded = true
	file := NewTreeNode("f", "file.txt")
	file.Icon = "📄"
	root.AddChild(file)
	tree.SetRoot(root)
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})

	buf := buffer.NewBuffer(40, 5)
	tree.Paint(buf)

	// Root line: "▼ 📁 root"
	// Check for 📁 after the ▼ and space
	foundIcon := false
	for x := 0; x < 10; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Rune == '📁' || cell.Rune == '📄' {
			foundIcon = true
			break
		}
	}
	if !foundIcon {
		t.Error("Expected to find icon rune in painted output")
	}
}

func TestTree_ScrollBehavior(t *testing.T) {
	tree := NewTree()
	// Build a tall tree
	root := NewTreeNode("root", "root")
	root.Expanded = true
	for i := 0; i < 20; i++ {
		root.AddChild(NewTreeNode("c", "child"))
	}
	tree.SetRoot(root)
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})

	// Navigate to last node
	tree.SetCurrent(20) // last child
	scrollY := tree.ScrollY()
	if scrollY < 0 {
		t.Errorf("scrollY should be >= 0, got %d", scrollY)
	}
	// scrollY should ensure cursor is visible (scrollY <= 20 < scrollY+5)
	if scrollY > 20 || scrollY+5 <= 20 {
		t.Errorf("scrollY=%d doesn't make cursor 20 visible in H=5", scrollY)
	}
}

func TestTree_HandleKeyNil(t *testing.T) {
	tree := NewTree()
	if tree.HandleKey(nil) {
		t.Error("HandleKey(nil) should return false")
	}
}

func TestTree_SetCurrent(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	tree.SetCurrent(3)
	node := tree.CurrentNode()
	if node == nil || node.ID != "utils" {
		t.Errorf("SetCurrent(3): ID=%v, want utils", node)
	}
}

func TestTree_SetCurrentOutOfRange(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	tree.SetCurrent(999)
	idx := tree.CurrentIndex()
	if idx != tree.VisibleCount()-1 {
		t.Errorf("SetCurrent(999): idx=%d, want %d (clamped)", idx, tree.VisibleCount()-1)
	}

	tree.SetCurrent(-5)
	idx = tree.CurrentIndex()
	if idx != 0 {
		t.Errorf("SetCurrent(-5): idx=%d, want 0 (clamped)", idx)
	}
}

func TestTree_SelectedNodes(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// Select root and main
	tree.SetCurrent(0)
	tree.HandleKey(keyTree(term.KeySpace))
	tree.SetCurrent(2)
	tree.HandleKey(keyTree(term.KeySpace))

	selected := tree.SelectedNodes()
	if len(selected) != 2 {
		t.Errorf("SelectedNodes: got %d, want 2", len(selected))
	}
}

func TestTree_DeepNesting(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("r0", "level0")
	root.Expanded = true
	current := root
	for i := 1; i <= 5; i++ {
		child := NewTreeNode("r"+string(rune('0'+i)), "level"+string(rune('0'+i)))
		child.Expanded = true
		current.AddChild(child)
		current = child
	}
	tree.SetRoot(root)
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	if tree.VisibleCount() != 6 {
		t.Errorf("VisibleCount: got %d, want 6", tree.VisibleCount())
	}

	// Navigate to deepest
	tree.SetCurrent(5)
	path := tree.Path()
	if len(path) != 6 {
		t.Errorf("Path len: got %d, want 6", len(path))
	}
}

func TestTree_ExpandedIndicatorInPaint(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("root", "root")
	root.Expanded = true
	root.AddChild(NewTreeNode("c", "child"))
	tree.SetRoot(root)
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})

	buf := buffer.NewBuffer(40, 5)
	tree.Paint(buf)

	// First char should be ▼ (expanded)
	cell := buf.GetCell(0, 0)
	if cell.Rune != '▼' {
		t.Errorf("Expected ▼ at (0,0), got %q", string(cell.Rune))
	}
}

func TestTree_CollapsedIndicatorInPaint(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("root", "root")
	root.AddChild(NewTreeNode("c", "child"))
	tree.SetRoot(root)
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})

	buf := buffer.NewBuffer(40, 5)
	tree.Paint(buf)

	// First char should be ▶ (collapsed)
	cell := buf.GetCell(0, 0)
	if cell.Rune != '▶' {
		t.Errorf("Expected ▶ at (0,0), got %q", string(cell.Rune))
	}
}

func TestTree_Measure(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("root", "root")
	root.Expanded = true
	root.AddChild(NewTreeNode("c", "child_with_long_name"))
	tree.SetRoot(root)

	size := tree.Measure(Unbounded())
	if size.H != 2 {
		t.Errorf("H: got %d, want 2", size.H)
	}
	if size.W < 1 {
		t.Errorf("W should be > 0, got %d", size.W)
	}
}

func TestTree_ConcurrentAccess(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			tree.HandleKey(keyTree(term.KeyDown))
			tree.HandleKey(keyTree(term.KeyUp))
			tree.ToggleExpand()
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_ = tree.CurrentNode()
			_ = tree.VisibleCount()
			_ = tree.Path()
		}
		done <- true
	}()

	<-done
	<-done
	// If we get here without panic or race detector firing, we're good
}

func TestTree_SpaceTogglesSelectionOff(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// Select then deselect
	tree.HandleKey(keyTree(term.KeySpace))
	node := tree.CurrentNode()
	if !node.Selected {
		t.Fatal("should be selected after first Space")
	}
	tree.HandleKey(keyTree(term.KeySpace))
	if node.Selected {
		t.Error("should be deselected after second Space")
	}
}

func TestTree_LeafRightKey(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// Navigate to leaf (gomod)
	tree.SetCurrent(5)
	beforeIdx := tree.CurrentIndex()
	tree.HandleKey(keyTree(term.KeyRight))
	// Right on a leaf should not crash, might stay or move
	afterNode := tree.CurrentNode()
	if afterNode == nil {
		t.Error("CurrentNode should not be nil after Right on leaf")
	}
	_ = beforeIdx
}

func TestTree_SetStyle(t *testing.T) {
	tree := NewTree()
	style := buffer.Style{Fg: buffer.NamedColor(buffer.NamedRed)}
	tree.SetStyle(style)
	// No error means it worked
	tree.SetSelectedStyle(buffer.Style{Flags: buffer.Bold})
}

func TestTree_SelectedCheckmarkInPaint(t *testing.T) {
	tree := NewTree()
	tree.SetRoot(buildTestTree())
	tree.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// Select root
	tree.HandleKey(keyTree(term.KeySpace))

	buf := buffer.NewBuffer(40, 10)
	tree.Paint(buf)

	// Last column should have checkmark
	cell := buf.GetCell(39, 0)
	if cell.Rune != '✓' {
		t.Errorf("Expected ✓ at (39,0), got %q", string(cell.Rune))
	}
}

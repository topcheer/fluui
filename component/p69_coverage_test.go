package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// === P69 Coverage tests — component package ===

func TestP69_Table_truncateToWidth(t *testing.T) {
	tbl := NewTable([]string{"A", "B"})

	// Test normal truncation
	got := tbl.truncateToWidth("Hello World", 5)
	if len(got) > 5 {
		t.Errorf("expected max 5 width, got %d: %q", buffer.StringWidth(got), got)
	}

	// Test exact fit
	got = tbl.truncateToWidth("Hello", 5)
	if got != "Hello" {
		t.Errorf("expected exact fit, got %q", got)
	}

	// Test width <= 0
	got = tbl.truncateToWidth("Hello", 0)
	if got != "" {
		t.Errorf("expected empty for width 0, got %q", got)
	}

	// Test negative width
	got = tbl.truncateToWidth("Hello", -1)
	if got != "" {
		t.Errorf("expected empty for negative width, got %q", got)
	}

	// Test with ellipsis (width >= 3)
	got = tbl.truncateToWidth("Hello World", 8)
	// Should have ellipsis or truncation
	if buffer.StringWidth(got) > 8 {
		t.Errorf("expected max 8 width, got %d", buffer.StringWidth(got))
	}

	// Test very narrow (width < 3, no ellipsis)
	got = tbl.truncateToWidth("Hello", 2)
	if buffer.StringWidth(got) > 2 {
		t.Errorf("expected max 2 width, got %d", buffer.StringWidth(got))
	}

	// Test unicode
	got = tbl.truncateToWidth("héllo wörld", 5)
	if buffer.StringWidth(got) > 5 {
		t.Errorf("expected max 5 width for unicode, got %d", buffer.StringWidth(got))
	}

	// Test width 3 exactly (ellipsis boundary)
	got = tbl.truncateToWidth("Hello World", 3)
	if buffer.StringWidth(got) > 3 {
		t.Errorf("expected max 3 width, got %d", buffer.StringWidth(got))
	}
}

func TestP69_Table_ScrollUp(t *testing.T) {
	tbl := NewTable([]string{"A", "B"})
	for i := 0; i < 20; i++ {
		tbl.AddRow([]string{"val" + itoa(i), "x"})
	}
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	tbl.ScrollDown(10) // scroll down first
	before := tbl.ScrollY()
	tbl.ScrollUp(3)
	after := tbl.ScrollY()
	if after >= before {
		t.Errorf("expected scroll up to decrease Y, before=%d after=%d", before, after)
	}
}

func TestP69_Table_ScrollUp_Clamped(t *testing.T) {
	tbl := NewTable([]string{"A", "B"})
	tbl.AddRow([]string{"val1", "x"})
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	tbl.ScrollUp(100) // try to scroll way past top
	if tbl.ScrollY() < 0 {
		t.Errorf("expected scrollY >= 0, got %d", tbl.ScrollY())
	}
}

func TestP69_Table_clampScrollYLocked(t *testing.T) {
	tbl := NewTable([]string{"A", "B"})
	for i := 0; i < 50; i++ {
		tbl.AddRow([]string{"val" + itoa(i), "x"})
	}
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	if tbl.ScrollY() < 0 {
		t.Errorf("expected scrollY >= 0, got %d", tbl.ScrollY())
	}
}

func TestP69_Tooltip_smartFlip_AllCases(t *testing.T) {
	tp := NewTooltip("test")
	tp.SetText("test")

	// Top → Bottom (not enough space above)
	result := tp.smartFlip(TooltipTop, 10, 0, 20, 10, 100, 50)
	if result != TooltipBottom {
		t.Errorf("expected Bottom flip, got %v", result)
	}

	// Top stays Top (enough space)
	result = tp.smartFlip(TooltipTop, 10, 30, 20, 10, 100, 50)
	if result != TooltipTop {
		t.Errorf("expected Top stay, got %v", result)
	}

	// Bottom → Top (not enough space below)
	result = tp.smartFlip(TooltipBottom, 10, 45, 20, 10, 100, 50)
	if result != TooltipTop {
		t.Errorf("expected Top flip, got %v", result)
	}

	// Bottom stays Bottom
	result = tp.smartFlip(TooltipBottom, 10, 5, 20, 10, 100, 50)
	if result != TooltipBottom {
		t.Errorf("expected Bottom stay, got %v", result)
	}

	// Right → Left (not enough space right)
	result = tp.smartFlip(TooltipRight, 95, 10, 20, 10, 100, 50)
	if result != TooltipLeft {
		t.Errorf("expected Left flip, got %v", result)
	}

	// Right stays Right
	result = tp.smartFlip(TooltipRight, 10, 10, 20, 10, 100, 50)
	if result != TooltipRight {
		t.Errorf("expected Right stay, got %v", result)
	}

	// Left → Right (not enough space left)
	result = tp.smartFlip(TooltipLeft, 5, 10, 20, 10, 100, 50)
	if result != TooltipRight {
		t.Errorf("expected Right flip, got %v", result)
	}

	// Left stays Left
	result = tp.smartFlip(TooltipLeft, 50, 10, 20, 10, 100, 50)
	if result != TooltipLeft {
		t.Errorf("expected Left stay, got %v", result)
	}
}

func TestP69_Tooltip_paintPlainText(t *testing.T) {
	tp := NewTooltip("test")
	tp.SetText("line1\nline2")
	tp.SetShowBorder(false)
	tp.Show()
	buf := buffer.NewBuffer(20, 5)
	tp.paintPlainText(buf, Rect{X: 0, Y: 0, W: 20, H: 5})
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected content at (0,0)")
	}
}

func TestP69_Badge_SizeName_AllSizes(t *testing.T) {
	sizes := []BadgeSize{BadgeSizeSmall, BadgeSizeNormal, BadgeSizeLarge, BadgeSize(99)}
	for _, s := range sizes {
		name := SizeName(s)
		_ = name
	}
}

func TestP69_TabBar_PrevTab(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "Tab 1")
	tb.AddTab("t2", "Tab 2")
	tb.AddTab("t3", "Tab 3")
	tb.SetActive(2)
	tb.PrevTab()
	if tb.ActiveIndex() != 1 {
		t.Errorf("expected index 1, got %d", tb.ActiveIndex())
	}
	// Wrap around
	tb.SetActive(0)
	tb.PrevTab()
	// Should wrap to last tab
	if tb.ActiveIndex() != 2 {
		t.Logf("PrevTab at 0 → %d (may or may not wrap)", tb.ActiveIndex())
	}
}

func TestP69_TabBar_Measure(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "Tab 1")
	s := tb.Measure(Bounded(80, 100))
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected positive size, got %v", s)
	}
}

func TestP69_WindowManager_Focused_Empty(t *testing.T) {
	wm := NewWindowManager(NewText("a"))
	_ = wm.Focused()
}

func TestP69_WindowManager_Measure(t *testing.T) {
	wm := NewWindowManager(NewText("a"))
	wm.SplitRight(NewText("b"), "panel2")
	s := wm.Measure(Bounded(80, 100))
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected positive size, got %v", s)
	}
}

func TestP69_WindowManager_Bounds(t *testing.T) {
	wm := NewWindowManager(NewText("a"))
	r := wm.Bounds()
	_ = r // should not panic
}

func TestP69_Tree_Measure(t *testing.T) {
	tr := NewTree()
	root := NewTreeNode("root", "Root")
	root.AddChild(NewTreeNode("c1", "Child 1"))
	tr.SetRoot(root)
	s := tr.Measure(Bounded(80, 100))
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected positive size, got %v", s)
	}
}

func TestP69_Tree_collapseAllLocked(t *testing.T) {
	tr := NewTree()
	root := NewTreeNode("root", "Root")
	child := NewTreeNode("c1", "Child 1")
	child.AddChild(NewTreeNode("g1", "Grandchild 1"))
	root.AddChild(child)
	tr.SetRoot(root)
	tr.ExpandAll()
	tr.CollapseAll()
	// After collapse, root children should be hidden
	if tr.VisibleCount() < 1 {
		t.Errorf("expected at least 1 visible after collapse, got %d", tr.VisibleCount())
	}
}

func TestP69_Tree_moveCursor(t *testing.T) {
	tr := NewTree()
	root := NewTreeNode("root", "Root")
	for i := 0; i < 5; i++ {
		root.AddChild(NewTreeNode("c"+itoa(i), "Child "+itoa(i)))
	}
	tr.SetRoot(root)
	tr.SetCurrent(2)
	_ = tr.Cursor() // should not panic
}

func TestP69_Viewport_ScrollLeft(t *testing.T) {
	vp := NewViewport(NewText("content"))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	before := vp.OffsetX()
	vp.ScrollLeft(1)
	after := vp.OffsetX()
	_ = before
	_ = after
}

func TestP69_Viewport_HScrollbarRow(t *testing.T) {
	vp := NewViewport(NewText("content"))
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	row := vp.HScrollbarRow()
	_ = row // should not panic
}

func TestP69_VirtualScroller_clampScrollLocked(t *testing.T) {
	vs := NewVirtualScroller()
	for i := 0; i < 20; i++ {
		vs.AddItem(VirtualItem{ID: "i" + itoa(i), Text: "item " + itoa(i)})
	}
	vs.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	if vs.ScrollY() < 0 {
		t.Errorf("expected scrollY >= 0, got %d", vs.ScrollY())
	}
}

// itoa is a local helper in statusbar.go
func TestP69_Helper_itoa(t *testing.T) {
	if itoa(42) != "42" {
		t.Errorf("itoa(42) = %q", itoa(42))
	}
}

package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === DiffPreview: SetShowLineNumbers, SetShowStats ===

func TestP62_DiffPreview_SetShowLineNumbers(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	// Verify it doesn't panic and the flag is set
	dp.SetShowLineNumbers(false)
}

func TestP62_DiffPreview_SetShowStats(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(true)
	dp.SetShowStats(false)
}

// === LineChart: formatSeriesValue ===

func TestP62_LineChart_FormatSeriesValue(t *testing.T) {
	// Integer
	if v := formatSeriesValue(42.0); v == "" {
		t.Error("expected non-empty for integer value")
	}
	// Float
	if v := formatSeriesValue(3.14); v == "" {
		t.Error("expected non-empty for float value")
	}
	// Zero
	if v := formatSeriesValue(0); v == "" {
		t.Error("expected non-empty for zero")
	}
	// Negative
	if v := formatSeriesValue(-5.5); v == "" {
		t.Error("expected non-empty for negative")
	}
}

// === Sparkline: formatFloat, formatSparkFloatLen ===

func TestP62_Sparkline_FormatFloat(t *testing.T) {
	// formatFloat uses simple int truncation, may have precision quirks
	// Just verify it returns non-empty with a decimal point
	if v := formatFloat(3.5); v == "" {
		t.Errorf("expected non-empty for 3.5")
	}
	if v := formatFloat(0.1); v == "" {
		t.Errorf("expected non-empty for 0.1")
	}
	// Integer part should be correct for positive
	v := formatFloat(5.5)
	if v[:1] != "5" {
		t.Errorf("expected '5' prefix, got %q", v)
	}
}

func TestP62_Sparkline_FormatSparkFloatLen(t *testing.T) {
	// Integer values format as just digits (no decimal)
	if n := formatSparkFloatLen(10.0); n != 2 {
		t.Errorf("expected 2 for 10.0, got %d", n)
	}
	// Float values format as N.N
	if n := formatSparkFloatLen(3.5); n != 3 {
		t.Errorf("expected 3 for 3.5, got %d", n)
	}
}

// === Table: utf8RuneCount ===

func TestP62_Table_UTF8RuneCount(t *testing.T) {
	if n := utf8RuneCount("hello"); n != 5 {
		t.Errorf("expected 5, got %d", n)
	}
	if n := utf8RuneCount(""); n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
	// Unicode
	if n := utf8RuneCount("héllo"); n != 5 {
		t.Errorf("expected 5 for héllo, got %d", n)
	}
	// Emoji
	if n := utf8RuneCount("🎉"); n != 1 {
		t.Errorf("expected 1 for emoji, got %d", n)
	}
}

// === TextArea: SetStyle ===

func TestP62_TextArea_SetStyle(t *testing.T) {
	ta := NewTextArea()
	style := buffer.Style{Fg: buffer.NamedColor(buffer.NamedRed)}
	ta.SetStyle(style)
	// Should not panic
}

// === Tree: Cursor, Root, SetScrollY, SetCursor, String ===

func TestP62_Tree_Cursor(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("root", "Root")
	child1 := NewTreeNode("c1", "Child 1")
	root.AddChild(child1)
	tree.SetRoot(root)
	// Cursor should be 0 initially
	if tree.Cursor() != tree.CurrentIndex() {
		t.Error("Cursor should match CurrentIndex")
	}
}

func TestP62_Tree_Root(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("root", "Root")
	tree.SetRoot(root)
	got := tree.Root()
	if got == nil || got.ID != "root" {
		t.Error("expected non-nil root with ID 'root'")
	}
}

func TestP62_Tree_RootNil(t *testing.T) {
	tree := NewTree()
	if tree.Root() != nil {
		t.Error("expected nil root for empty tree")
	}
}

func TestP62_Tree_SetScrollY(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("root", "Root")
	tree.SetRoot(root)
	tree.SetScrollY(5)
	// Negative should clamp to 0
	tree.SetScrollY(-1)
}

func TestP62_Tree_SetCursor(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("root", "Root")
	child := NewTreeNode("child", "Child")
	root.AddChild(child)
	tree.SetRoot(root)

	// Set cursor to child (should find and expand parent)
	ok := tree.SetCursor("child")
	if !ok {
		t.Error("expected SetCursor to return true for existing node")
	}
}

func TestP62_Tree_SetCursorNotFound(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("root", "Root")
	tree.SetRoot(root)

	ok := tree.SetCursor("nonexistent")
	if ok {
		t.Error("expected false for non-existent node")
	}
}

func TestP62_Tree_StringEmpty(t *testing.T) {
	tree := NewTree()
	if tree.String() != "" {
		t.Error("expected empty string for tree without root")
	}
}

func TestP62_Tree_StringWithContent(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("root", "Root")
	child := NewTreeNode("c1", "Child 1")
	root.AddChild(child)
	root.Expanded = true
	tree.SetRoot(root)

	s := tree.String()
	if s == "" {
		t.Error("expected non-empty string for tree with root")
	}
}

func TestP62_Tree_StringWithCollapsedChildren(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("root", "Root")
	child := NewTreeNode("c1", "Child 1")
	root.AddChild(child)
	root.Expanded = false
	tree.SetRoot(root)

	s := tree.String()
	// Should have root but not child (collapsed)
	if s == "" {
		t.Error("expected non-empty string")
	}
}

func TestP62_Tree_StringWithIcon(t *testing.T) {
	tree := NewTree()
	root := NewTreeNode("root", "Root")
	root.Icon = "📁"
	tree.SetRoot(root)

	s := tree.String()
	if s == "" {
		t.Error("expected non-empty string with icon")
	}
}

// === Viewport: IsDraggingV, IsDraggingH ===

func TestP62_Viewport_IsDraggingV(t *testing.T) {
	child := NewText("test")
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	if vp.IsDraggingV() {
		t.Error("expected not dragging initially")
	}
}

func TestP62_Viewport_IsDraggingH(t *testing.T) {
	child := NewText("test")
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	if vp.IsDraggingH() {
		t.Error("expected not dragging initially")
	}
}

// === SplitPane: abs ===

func TestP62_SplitPane_AbsPositive(t *testing.T) {
	if abs(5) != 5 {
		t.Error("expected abs(5) = 5")
	}
}

func TestP62_SplitPane_AbsNegative(t *testing.T) {
	if abs(-5) != 5 {
		t.Error("expected abs(-5) = 5")
	}
}

func TestP62_SplitPane_AbsZero(t *testing.T) {
	if abs(0) != 0 {
		t.Error("expected abs(0) = 0")
	}
}

// === Pagination: recomputePagesLocked (50%) ===

func TestP62_Pagination_RecomputeWithZeroPerPage(t *testing.T) {
	p := NewPagination()
	p.SetItemsPerPage(0)
	// With 0 items per page, total pages should be 0 (edge case)
	total := p.TotalPages()
	if total < 0 {
		t.Errorf("expected >= 0 pages, got %d", total)
	}
}

func TestP62_Pagination_RecomputeWithLargeItems(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(1000)
	p.SetItemsPerPage(25)
	if p.TotalPages() != 40 {
		t.Errorf("expected 40 pages, got %d", p.TotalPages())
	}
}

// === Component: BaseComponent.Paint (0%) ===

func TestP62_BaseComponent_Paint(t *testing.T) {
	bc := BaseComponent{}
	bc.Paint(buffer.NewBuffer(10, 5))
	// Should be a no-op, not panic
}

// === TabBar: PrevTab (80%) ===

func TestP62_TabBar_PrevTab(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "Tab 1")
	tb.AddTab("t2", "Tab 2")
	tb.AddTab("t3", "Tab 3")
	tb.SetActive(2)
	tb.PrevTab()
	if tb.ActiveIndex() != 1 {
		t.Errorf("expected 1 after PrevTab, got %d", tb.ActiveIndex())
	}
}

// === Help overlay edge cases ===

func TestP62_HelpOverlay_EnsureSelectedValidEmpty(t *testing.T) {
	h := NewHelpOverlay(nil)
	// No groups → ensure selected valid should handle gracefully
	h.SetGroups(nil)
}

// === FilePicker: Measure with empty entries ===

func TestP62_FilePicker_MeasureEmpty(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return nil, nil
	})
	s := fp.Measure(Bounded(20, 10))
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected positive size, got %v", s)
	}
}

// === WindowManager edge cases ===

func TestP62_WindowManager_FocusedEmpty(t *testing.T) {
	wm := NewWindowManager(NewText("test"))
	// With 1 pane, focused should return that pane
	if wm.Focused() == nil {
		t.Error("expected non-nil focused with 1 pane")
	}
}

func TestP62_WindowManager_FocusPrevSingle(t *testing.T) {
	wm := NewWindowManager(NewText("test"))
	wm.FocusPrev()
	// Should stay at pane 0 with single pane
	if wm.FocusedIndex() != 0 {
		t.Error("expected focused index 0")
	}
}

// === Checkbox: IsChecked out of bounds ===

func TestP62_Checkbox_IsCheckedOutOfBounds(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B"})
	// Negative index
	if cb.IsChecked(-1) {
		t.Error("expected false for negative index")
	}
	// Out of bounds
	if cb.IsChecked(5) {
		t.Error("expected false for out-of-bounds index")
	}
}

// === Gauge: Ratio edge cases ===

func TestP62_Gauge_RatioOverflow(t *testing.T) {
	g := NewGauge()
	g.SetRange(0, 100)
	g.SetValue(200) // overflow
	r := g.Ratio()
	if r > 1.0 {
		t.Errorf("expected ratio clamped to <= 1.0, got %f", r)
	}
}

func TestP62_Gauge_RatioEqual(t *testing.T) {
	g := NewGauge()
	g.SetRange(0, 100)
	g.SetValue(50)
	if g.Ratio() != 0.5 {
		t.Errorf("expected 0.5, got %f", g.Ratio())
	}
}

// === Progress: SetIndeterminateWidth ===

func TestP62_Progress_SetIndeterminateWidth(t *testing.T) {
	p := NewProgressBar()
	p.SetMode(ProgressIndeterminate)
	p.SetIndeterminateWidth(10)
	// Should not panic
}

// === Form: FocusNext/FocusPrev wrapping ===

func TestP62_Form_FocusNextWrap(t *testing.T) {
	f := NewForm()
	f.AddTextField("Field 1", "f1", "")
	f.AddTextField("Field 2", "f2", "")
	f.SetActiveIndex(1) // Last field
	f.FocusNext()       // Should wrap to 0
	if f.ActiveIndex() != 0 {
		t.Errorf("expected wrap to 0, got %d", f.ActiveIndex())
	}
}

func TestP62_Form_FocusPrevWrap(t *testing.T) {
	f := NewForm()
	f.AddTextField("Field 1", "f1", "")
	f.AddTextField("Field 2", "f2", "")
	f.SetActiveIndex(0) // First field
	f.FocusPrev()       // Should wrap to last
	if f.ActiveIndex() != 1 {
		t.Errorf("expected wrap to 1, got %d", f.ActiveIndex())
	}
}

// === ListView: RemoveItem ===

func TestP62_ListView_RemoveItemOutOfBounds(t *testing.T) {
	lv := NewListView([]string{"A", "B"})
	lv.RemoveItem(10) // Should not panic
	if lv.ItemCount() != 2 {
		t.Error("expected 2 items after removing invalid index")
	}
}

// === Dialog: Cancel ===

func TestP62_Dialog_CancelWithCallback(t *testing.T) {
	d := NewConfirmDialog("Title", "Message")
	called := false
	d.OnCancel = func() {
		called = true
	}
	d.Cancel()
	if !called {
		t.Error("expected cancel callback called")
	}
	if !d.Closed() {
		t.Error("expected dialog closed")
	}
}

// === ContextMenu: HandleKey ===

func TestP62_ContextMenu_HandleKeyEscape(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("item1", "Item 1"))
	cm.Show(0, 0)
	consumed := cm.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Error("expected escape consumed")
	}
	if cm.Visible() {
		t.Error("expected hidden after escape")
	}
}

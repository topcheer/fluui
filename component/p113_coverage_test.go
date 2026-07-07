package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── 0% → 100% functions ───

func TestP113_BaseComponent_Paint(t *testing.T) {
	bc := BaseComponent{}
	buf := buffer.NewBuffer(5, 3)
	bc.Paint(buf) // no-op, should not panic
}

func TestP113_BaseComponent_Measure(t *testing.T) {
	bc := BaseComponent{}
	s := bc.Measure(Bounded(10, 5))
	if s.W != 0 || s.H != 0 {
		t.Errorf("expected zero size, got %v", s)
	}
}

func TestP113_DiffPreview_SetShowLineNumbers(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(false)
	if !dp.ShowLineNumbers() {
		t.Error("ShowLineNumbers should always return true")
	}
	dp.SetShowLineNumbers(true)
	if !dp.ShowLineNumbers() {
		t.Error("ShowLineNumbers should always return true")
	}
}

func TestP113_DiffPreview_SetShowStats(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(false)
	dp.SetShowStats(true)
}

// ─── Viewport drawVScrollBar (73.7%) ───

func TestP113_Viewport_DrawVScrollBar(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 60, h: 50})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
	// Scroll down to trigger scrollbar thumb rendering
	vp.ScrollDown(20)
	vp.Paint(buf)
}

func TestP113_Viewport_DrawHScrollBar(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 100, h: 5})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	vp.ScrollRight(30)
	vp.Paint(buf)
}

func TestP113_Viewport_ScrollToX(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 100, h: 5})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vp.ScrollToX(50)
	if vp.OffsetX() != 50 {
		t.Errorf("expected OffsetX=50, got %d", vp.OffsetX())
	}
	// Test clamping
	vp.ScrollToX(999)
	if vp.OffsetX() != vp.MaxOffsetX() {
		t.Errorf("expected clamped OffsetX=%d, got %d", vp.MaxOffsetX(), vp.OffsetX())
	}
}

// ─── Badge Measure (76.5%) ───

func TestP113_Badge_Measure_AllSizes(t *testing.T) {
	for _, size := range []BadgeSize{BadgeSizeSmall, BadgeSizeNormal, BadgeSizeLarge} {
		b := NewBadge("text", BadgeInfo)
		b.SetSize(size)
		s := b.Measure(Bounded(80, 24))
		if s.W <= 0 || s.H <= 0 {
			t.Errorf("size %d: expected non-zero measure, got %v", size, s)
		}
	}
}

func TestP113_Badge_Measure_WithIcon(t *testing.T) {
	b := NewBadge("text", BadgeSuccess)
	b.SetIcon("✓")
	s := b.Measure(Bounded(80, 24))
	if s.W <= 0 {
		t.Errorf("expected non-zero width with icon, got %v", s)
	}
}

func TestP113_Badge_Measure_NarrowClamp(t *testing.T) {
	b := NewBadge("long text here", BadgeWarning)
	s := b.Measure(Bounded(3, 5))
	if s.W > 3 {
		t.Errorf("expected width clamped to 3, got %d", s.W)
	}
}

func TestP113_Badge_Measure_ShortText(t *testing.T) {
	b := NewBadge("hi", BadgeError)
	s := b.Measure(Bounded(80, 24))
	if s.W < 2 {
		t.Errorf("expected width >= 2, got %d", s.W)
	}
}

func TestP113_Badge_SizeName(t *testing.T) {
	for _, tc := range []struct {
		size BadgeSize
		want string
	}{
		{BadgeSizeSmall, "small"},
		{BadgeSizeNormal, "normal"},
		{BadgeSizeLarge, "large"},
		{BadgeSize(99), "unknown"},
	} {
		got := SizeName(tc.size)
		if got != tc.want {
			t.Errorf("SizeName(%d) = %q, want %q", tc.size, got, tc.want)
		}
	}
}

// ─── ProgressBar formatPercent (77.8%) ───

func TestP113_ProgressBar_FormatPercent(t *testing.T) {
	pb := NewProgressBar()
	pb.SetProgress(0.5)
	// Just paint to exercise formatPercent
	buf := buffer.NewBuffer(30, 1)
	pb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	pb.Paint(buf)
}

func TestP113_ProgressBar_FormatPercent_EdgeCases(t *testing.T) {
	for _, v := range []float64{0.0, 0.333, 0.5, 0.75, 1.0} {
		pb := NewProgressBar()
		pb.SetProgress(v)
		buf := buffer.NewBuffer(30, 1)
		pb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
		pb.Paint(buf)
	}
}

// ─── Table drawCellLocked (75.0%) ───

func TestP113_Table_DrawCellLocked(t *testing.T) {
	tbl := NewTable([]string{"A", "B"})
	tbl.SetRows([][]string{{"x", "y"}})
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	tbl.Paint(buf)
}

func TestP113_Table_TruncateToWidth(t *testing.T) {
	tbl := NewTable([]string{"Col"})
	// Exercise via Paint with narrow bounds
	tbl.SetRows([][]string{{"very long content that needs truncation"}})
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	buf := buffer.NewBuffer(5, 3)
	tbl.Paint(buf)
}

// ─── ScrollView contentW (75.0%) ───

func TestP113_ScrollView_ContentW(t *testing.T) {
	sv := NewScrollView(&fixedSize{w: 40, h: 30})
	sv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	sv.Paint(buf)
}

// ─── ContextMenu setCursorLocked (73.3%) ───

func TestP113_ContextMenu_SetCursorLocked(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "A"))
	cm.AddItem(NewMenuItem("b", "B"))
	cm.AddItem(NewMenuItem("c", "C"))
	cm.SetCursor(-1) // negative: should clamp to 0
	cm.SetCursor(99) // overflow: should clamp to last
	cm.SetCursor(1)  // valid
}

// ─── Sparkline recomputeRange/valueToBar (77.8%) ───

func TestP113_Sparkline_RecomputeRange(t *testing.T) {
	sp := NewSparkline()
	sp.SetData([]float64{1.0, 5.0, 3.0, 8.0, 2.0})
	sp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	sp.Paint(buf)
}

func TestP113_Sparkline_ValueToBar(t *testing.T) {
	sp := NewSparkline()
	sp.SetData([]float64{0.0, 0.5, 1.0})
	sp.SetAutoScale(true)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	sp.Paint(buf)
}

// ─── BarChart paintHorizontal (78.6%) ───

func TestP113_BarChart_PaintHorizontal(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	bc.AddSeries(BarSeries{
		Name: "test",
		Data: []BarData{{Label: "A", Value: 10}, {Label: "B", Value: 20}},
	})
	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	bc.Paint(buf)
}

func TestP113_BarChart_PaintHorizontal_NoLabels(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	bc.AddSeries(BarSeries{
		Name: "test",
		Data: []BarData{{Value: 5}, {Value: 15}},
	})
	bc.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 8})
	buf := buffer.NewBuffer(30, 8)
	bc.Paint(buf)
}

// ─── AutoComplete Paint (76.7%) ───

func TestP113_AutoComplete_Paint(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{{Label: "alpha"}, {Label: "beta"}})
	ac.SetCursor(0)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	ac.Paint(buf)
}

// ─── ListView MoveUp (76.9%) ───

func TestP113_ListView_MoveUp(t *testing.T) {
	lv := NewListView([]string{"a", "b", "c"})
	lv.SetCursor(2)
	lv.MoveUp()
	if lv.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", lv.Cursor())
	}
	// Wrap to bottom
	lv.SetCursor(0)
	lv.MoveUp()
	if lv.Cursor() != 2 {
		t.Errorf("expected wrap to 2, got %d", lv.Cursor())
	}
}

// ─── TextArea moveLine (77.8%) ───

func TestP113_TextArea_MoveLine(t *testing.T) {
	ta := NewTextArea()
	// Move down via key handling
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
}

// ─── Form HandleKey (78.3%) ───

func TestP113_Form_HandleKey(t *testing.T) {
	f := NewForm()
	f.AddField(NewTextField("name", "name", ""))
	f.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	// Tab between fields
	f.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	f.HandleKey(&term.KeyEvent{Key: term.KeyTab})
}

// ─── CodeBlock paintStreamingCursorLocked (74.2%) ───

func TestP113_CodeBlock_PaintStreamingCursor(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.AppendSource("func main() {")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestP113_CodeBlock_PaintStreamingCursor_LongLine(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.AppendSource("func main() { fmt.Println(\"hello world this is a very long line\") }")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf)
}

// ─── HelpOverlay ensureSelectedVisibleLocked (75.0%) ───

func TestP113_HelpOverlay_EnsureSelected(t *testing.T) {
	groups := []HelpGroup{
		{Name: "g1", Entries: []HelpEntry{{Keys: "a", Description: "aaa"}}},
		{Name: "g2", Entries: []HelpEntry{{Keys: "b", Description: "bbb"}}},
	}
	ho := NewHelpOverlay(groups)
	ho.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3}) // small to force scrolling
	buf := buffer.NewBuffer(40, 3)
	ho.Paint(buf)
	// Scroll down and paint again
	ho.ScrollDown(1)
	ho.Paint(buf)
}

// ─── DiffPreview paintBorderLocked (76.5%) ───

func TestP113_DiffPreview_PaintBorder(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+ added line\n- removed line\n  context")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	dp.Paint(buf)
}

func TestP113_DiffPreview_PaintBorder_Narrow(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+ added line\n- removed line")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 5})
	buf := buffer.NewBuffer(5, 5)
	dp.Paint(buf)
}

// ─── ThemeStudio setCursorLocked (75.0%) ───

func TestP113_ThemeStudio_SetCursor(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	ts.Paint(buf)
}

// ─── VirtualScroller VisibleItems (78.6%) ───

func TestP113_VirtualScroller_VisibleItems(t *testing.T) {
	items := make([]VirtualItem, 50)
	for i := range items {
		items[i] = VirtualItem{ID: string(rune('a' + i%26)), Text: "item"}
	}
	vs := NewVirtualScroller()
	vs.SetItems(items)
	vs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	vs.Paint(buf)
}

// ─── SelectField Value edge cases ───

func TestP113_SelectField_Value(t *testing.T) {
	sf := NewSelectField("label", "key", []string{"a", "b", "c"})
	sf.SetSelectedIndex(1)
	if sf.Value() != "b" {
		t.Errorf("expected 'b', got %q", sf.Value())
	}
	sf.SetSelectedIndex(-1)
	// negative index should return empty or first
	_ = sf.Value()
	sf.SetSelectedIndex(99)
	_ = sf.Value()
}

// ─── DebugInspector paintEventsLocked (77.4%) ───

func TestP113_DebugInspector_PaintEvents(t *testing.T) {
	di := NewDebugInspector()
	di.SetVisible(true)
	di.SetMode(InspectEvents)
	di.RecordKey(&term.KeyEvent{Key: term.KeyEnter})
	di.RecordKey(&term.KeyEvent{Key: term.KeySpace})
	di.RecordMouse(&term.MouseEvent{X: 1, Y: 2, Button: 1})
	di.RecordResize(80, 24)
	di.RecordCustom("test event")
	di.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	di.Paint(buf)
}

// ─── Gauge paintVertical (71%) ───

func TestP113_Gauge_PaintVertical(t *testing.T) {
	g := NewGauge()
	g.SetValue(0.7)
	g.SetOrientation(GaugeVertical)
	g.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 10})
	buf := buffer.NewBuffer(5, 10)
	g.Paint(buf)
}

func TestP113_Gauge_PaintVertical_Thresholds(t *testing.T) {
	for _, v := range []float64{0.1, 0.4, 0.6, 0.8, 1.0} {
		g := NewGauge()
		g.SetValue(v)
		g.SetOrientation(GaugeVertical)
		g.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 10})
		buf := buffer.NewBuffer(5, 10)
		g.Paint(buf)
	}
}

// ─── Tree rebuildCollapsed ───

func TestP113_Tree_CollapsedRebuild(t *testing.T) {
	root := NewTreeNode("root", "Root")
	child1 := NewTreeNode("c1", "Child1")
	child2 := NewTreeNode("c2", "Child2")
	child1.AddChild(NewTreeNode("gc1", "GrandChild1"))
	root.AddChild(child1)
	root.AddChild(child2)

	tr := NewTree()
	tr.SetRoot(root)
	tr.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	tr.Paint(buf)
	// Collapse all then paint
	tr.CollapseAll()
	tr.Paint(buf)
}

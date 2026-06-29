package component_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/component/layout"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ============================================================
// P26-B: Component → Layout → Buffer Integration Tests
// ============================================================
// These tests verify the full chain: component creation →
// layout container (Flex/Stack/Padding/Center) → Measure →
// SetBounds → Paint → buffer output.
// ============================================================

// helper to count non-blank cells in a buffer
func bufNonBlank(buf *buffer.Buffer) int {
	count := 0
	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune != 0 && cell.Rune != ' ' {
				count++
			}
		}
	}
	return count
}

// --- Flex horizontal layout pipeline ---

func TestP26B_FlexHorizontalMeasurePaint(t *testing.T) {
	flex := layout.NewFlex(layout.FlexRow)
	flex.AddChild(component.NewText("Left"))
	flex.AddChild(component.NewText("Right"))

	flex.Measure(component.Unbounded())
	flex.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})

	buf := buffer.NewBuffer(80, 1)
	flex.Paint(buf)

	if bufNonBlank(buf) == 0 {
		t.Error("flex horizontal paint produced no content")
	}
}

func TestP26B_FlexVerticalStacks(t *testing.T) {
	flex := layout.NewFlex(layout.FlexColumn)
	flex.AddChild(component.NewText("Top"))
	flex.AddChild(component.NewText("Middle"))
	flex.AddChild(component.NewText("Bottom"))

	sz := flex.Measure(component.Unbounded())
	flex.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	if sz.H < 3 {
		t.Errorf("3 text blocks should measure height >= 3, got %d", sz.H)
	}

	buf := buffer.NewBuffer(80, 24)
	flex.Paint(buf)

	if bufNonBlank(buf) == 0 {
		t.Error("flex vertical paint produced no content")
	}
}

// --- Stack overlap pipeline ---

func TestP26B_StackOverlap(t *testing.T) {
	stack := layout.NewStack()
	stack.AddChild(component.NewText("Background"))
	stack.AddChild(component.NewText("Overlay"))

	stack.Measure(component.Unbounded())
	stack.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})

	buf := buffer.NewBuffer(80, 1)
	stack.Paint(buf) // should not panic

	if bufNonBlank(buf) == 0 {
		t.Error("stack paint produced no content")
	}
}

// --- Padding wraps child ---

func TestP26B_PaddingWrapsChild(t *testing.T) {
	inner := component.NewText("Padded")
	pad := layout.NewPadding(2, 2, 2, 2, inner)

	sz := pad.Measure(component.Unbounded())
	pad.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})

	// Inner child should be smaller than outer bounds
	r := inner.Bounds()
	if r.X < 2 {
		t.Errorf("inner X should be >= left padding (2), got %d", r.X)
	}
	if r.Y < 2 {
		t.Errorf("inner Y should be >= top padding (2), got %d", r.Y)
	}

	buf := buffer.NewBuffer(80, 10)
	pad.Paint(buf) // should not panic

	if sz.W <= 0 || sz.H <= 0 {
		t.Errorf("padding measure should be positive: %dx%d", sz.W, sz.H)
	}
}

// --- Center centers child ---

func TestP26B_CenterCentersChild(t *testing.T) {
	child := component.NewText("Center")
	center := layout.NewCenter(child)

	center.Measure(component.Unbounded())
	center.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	r := child.Bounds()
	if r.W >= 80 {
		t.Errorf("child width should be less than container, got %d", r.W)
	}

	buf := buffer.NewBuffer(80, 24)
	center.Paint(buf)
}

// --- Nested layout: Flex inside Stack ---

func TestP26B_NestedFlexInStack(t *testing.T) {
	inner := layout.NewFlex(layout.FlexRow)
	inner.AddChild(component.NewText("A"))
	inner.AddChild(component.NewText("B"))

	outer := layout.NewStack()
	outer.AddChild(inner)

	outer.Measure(component.Unbounded())
	outer.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})

	buf := buffer.NewBuffer(80, 1)
	outer.Paint(buf) // should not panic
}

// --- Nested: column of rows ---

func TestP26B_NestedColumnOfRows(t *testing.T) {
	col := layout.NewFlex(layout.FlexColumn)

	for i := 0; i < 3; i++ {
		row := layout.NewFlex(layout.FlexRow)
		row.AddChild(component.NewText(fmt.Sprintf("R%dA", i)))
		row.AddChild(component.NewText(fmt.Sprintf("R%dB", i)))
		col.AddChild(row)
	}

	col.Measure(component.Unbounded())
	col.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	buf := buffer.NewBuffer(80, 24)
	col.Paint(buf) // should not panic

	if bufNonBlank(buf) == 0 {
		t.Error("nested column of rows produced no content")
	}
}

// --- Deep nesting (5 levels) ---

func TestP26B_DeepNesting(t *testing.T) {
	current := layout.NewFlex(layout.FlexRow)
	current.AddChild(component.NewText("inner"))

	for i := 0; i < 5; i++ {
		pad := layout.NewPadding(1, 1, 1, 1, current)
		wrap := layout.NewStack(pad)
		current = layout.NewFlex(layout.FlexRow)
		current.AddChild(wrap)
	}

	current.Measure(component.Unbounded())
	current.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	buf := buffer.NewBuffer(80, 24)
	current.Paint(buf) // should not panic at any nesting level
}

// --- Table → HandleKey → re-render ---

func TestP26B_TableHandleKeyReRender(t *testing.T) {
	tbl := component.NewTable(
		[]string{"ID", "Name"},
		[]string{"1", "Alice"},
		[]string{"2", "Bob"},
		[]string{"3", "Charlie"},
	)
	tbl.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	tbl.Measure(component.Constraints{})

	buf1 := buffer.NewBuffer(80, 10)
	tbl.Paint(buf1)
	beforeCount := bufNonBlank(buf1)

	// Navigate down
	tbl.HandleKey(&term.KeyEvent{Key: term.KeyDown})

	// Re-paint after navigation
	buf2 := buffer.NewBuffer(80, 10)
	tbl.Paint(buf2)
	afterCount := bufNonBlank(buf2)

	if beforeCount == 0 || afterCount == 0 {
		t.Error("table should produce visible content")
	}
}

// --- TabBar → HandleKey → re-render ---

func TestP26B_TabBarHandleKeyReRender(t *testing.T) {
	tb := component.NewTabBar()
	tb.AddTab("t1", "Tab 1")
	tb.AddTab("t2", "Tab 2")
	tb.AddTab("t3", "Tab 3")
	tb.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})
	tb.Measure(component.Constraints{})

	initial := tb.ActiveIndex()

	// Try switching (TabBar uses NextTab/PrevTab, not HandleKey)
	tb.NextTab()

	buf := buffer.NewBuffer(80, 1)
	tb.Paint(buf)

	_ = initial
}

// --- FilePicker → HandleKey → re-render ---

func TestP26B_FilePickerHandleKey(t *testing.T) {
	fp := component.NewFilePicker(".")
	fp.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 20})
	fp.Measure(component.Constraints{})

	// Navigate
	fp.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	fp.HandleKey(&term.KeyEvent{Key: term.KeyUp})

	buf := buffer.NewBuffer(80, 20)
	fp.Paint(buf) // should not panic
}

// --- Resize → relayout cycle ---

func TestP26B_ResizeRelayoutCycle(t *testing.T) {
	flex := layout.NewFlex(layout.FlexColumn)
	flex.AddChild(component.NewText("Header"))
	flex.AddChild(component.NewTable([]string{"A"}, []string{"1"}))
	flex.AddChild(component.NewStatusBar())

	sizes := []component.Rect{
		{X: 0, Y: 0, W: 80, H: 24},
		{X: 0, Y: 0, W: 120, H: 40},
		{X: 0, Y: 0, W: 40, H: 10},
		{X: 0, Y: 0, W: 200, H: 60},
		{X: 0, Y: 0, W: 20, H: 5},
	}

	for _, r := range sizes {
		flex.SetBounds(r)
		flex.Measure(component.Constraints{})
		buf := buffer.NewBuffer(r.W, r.H)
		flex.Paint(buf) // should not panic at any size
	}
}

// --- Complex dashboard-like layout ---

func TestP26B_DashboardLayout(t *testing.T) {
	root := layout.NewFlex(layout.FlexColumn)

	// Header
	header := component.NewTabBar()
	header.AddTab("dashboard", "Dashboard")
	header.AddTab("settings", "Settings")
	root.AddChild(header)

	// Content row
	content := layout.NewFlex(layout.FlexRow)
	leftPanel := component.NewGauge()
	leftPanel.SetValue(45)
	rightTable := component.NewTable(
		[]string{"Metric", "Value"},
		[]string{"CPU", "45%"},
		[]string{"MEM", "2.1GB"},
	)
	content.AddChild(leftPanel)
	content.AddChild(rightTable)
	root.AddChild(content)

	// Footer
	footer := component.NewStatusBar()
	root.AddChild(footer)

	root.Measure(component.Unbounded())
	root.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	buf := buffer.NewBuffer(80, 24)
	root.Paint(buf)

	if bufNonBlank(buf) == 0 {
		t.Error("dashboard layout produced no content")
	}
}

// --- Empty containers don't panic ---

func TestP26B_EmptyFlexPaint(t *testing.T) {
	flex := layout.NewFlex(layout.FlexRow)
	flex.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	flex.Measure(component.Constraints{})
	flex.Paint(buffer.NewBuffer(80, 24)) // should not panic
}

func TestP26B_EmptyStackPaint(t *testing.T) {
	stack := layout.NewStack()
	stack.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	stack.Measure(component.Constraints{})
	stack.Paint(buffer.NewBuffer(80, 24)) // should not panic
}

// --- Many children in Flex ---

func TestP26B_FlexManyChildren(t *testing.T) {
	flex := layout.NewFlex(layout.FlexColumn)
	for i := 0; i < 20; i++ {
		flex.AddChild(component.NewText(fmt.Sprintf("Row %d", i)))
	}

	flex.Measure(component.Unbounded())
	flex.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	buf := buffer.NewBuffer(80, 24)
	flex.Paint(buf) // should render visible window without panic
}

// --- Concurrent paint stress ---

func TestP26B_ConcurrentLayoutPaint(t *testing.T) {
	flex := layout.NewFlex(layout.FlexColumn)
	flex.AddChild(component.NewStatusBar())
	flex.AddChild(component.NewTabBar())
	flex.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	flex.Measure(component.Constraints{})

	var wg sync.WaitGroup
	const goroutines = 20
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := buffer.NewBuffer(80, 24)
			flex.Paint(buf)
		}()
	}
	wg.Wait()
}

// --- Large table in Flex ---

func TestP26B_LargeTableInFlex(t *testing.T) {
	headers := []string{"ID", "Name", "Status"}
	var rows [][]string
	for i := 0; i < 1000; i++ {
		rows = append(rows, []string{fmt.Sprintf("%d", i), fmt.Sprintf("item-%d", i), "active"})
	}

	tbl := component.NewTable(headers, rows...)
	flex := layout.NewFlex(layout.FlexColumn)
	flex.AddChild(tbl)

	flex.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	flex.Measure(component.Constraints{})

	buf := buffer.NewBuffer(80, 24)
	flex.Paint(buf) // should not panic
}

// --- Long text in table ---

func TestP26B_LongTextTableRendering(t *testing.T) {
	longText := strings.Repeat("x", 200)
	tbl := component.NewTable([]string{"Description"}, []string{longText})
	tbl.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	tbl.Measure(component.Constraints{})

	buf := buffer.NewBuffer(80, 10)
	tbl.Paint(buf) // should clip without panic
}

// --- Component measure consistency ---

func TestP26B_ComponentMeasureConsistency(t *testing.T) {
	tbl := component.NewTable(
		[]string{"A", "B"},
		[]string{"1", "2"},
		[]string{"3", "4"},
	)
	sz1 := tbl.Measure(component.Constraints{})
	sz2 := tbl.Measure(component.Constraints{})
	if sz1 != sz2 {
		t.Errorf("measure inconsistent: %v vs %v", sz1, sz2)
	}
}

// --- Tree expand/collapse lifecycle ---

func TestP26B_TreeExpandCollapseLifecycle(t *testing.T) {
	root := component.NewTreeNode("root", "Root")
	child1 := component.NewTreeNode("c1", "Child 1")
	child2 := component.NewTreeNode("c2", "Child 2")
	child1.AddChild(component.NewTreeNode("gc1", "Grandchild"))
	root.AddChild(child1)
	root.AddChild(child2)

	tree := component.NewTree()
	tree.SetData(root)
	tree.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	tree.Measure(component.Constraints{})

	// Expand
	tree.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	buf1 := buffer.NewBuffer(80, 24)
	tree.Paint(buf1)

	// Navigate down
	tree.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	buf2 := buffer.NewBuffer(80, 24)
	tree.Paint(buf2)

	// Collapse
	tree.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	buf3 := buffer.NewBuffer(80, 24)
	tree.Paint(buf3)

	// All paints should succeed without panic
	_ = buf1
	_ = buf2
	_ = buf3
}

// --- CommandPalette search → render ---

func TestP26B_CommandPaletteSearchRender(t *testing.T) {
	cp := component.NewCommandPalette()
	for i := 0; i < 50; i++ {
		cp.AddCommand(component.Command{
			ID:       fmt.Sprintf("cmd-%d", i),
			Label:    fmt.Sprintf("Command %d", i),
			Category: "test",
		})
	}
	cp.Show(0, 0) // must show before SetQuery (Show resets query)
	cp.SetQuery("command 4")
	cp.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	cp.Measure(component.Constraints{})

	buf := buffer.NewBuffer(80, 24)
	cp.Paint(buf)

	if bufNonBlank(buf) == 0 {
		t.Error("command palette search should produce results")
	}
}

// --- Sparkline data → render ---

func TestP26B_SparklineDataRender(t *testing.T) {
	sl := component.NewSparkline()
	var data []float64
	for i := 0; i < 100; i++ {
		data = append(data, float64(i)*0.5)
	}
	sl.SetData(data)
	sl.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 5})
	sl.Measure(component.Constraints{})

	buf := buffer.NewBuffer(80, 5)
	sl.Paint(buf)

	if bufNonBlank(buf) == 0 {
		t.Error("sparkline should produce visible content")
	}
}

// --- RadioGroup selection → render ---

func TestP26B_RadioGroupSelectRender(t *testing.T) {
	rg := component.NewRadioGroup([]string{"Red", "Green", "Blue"})
	rg.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 5})
	rg.Measure(component.Constraints{})

	// Navigate
	rg.HandleKey(&term.KeyEvent{Key: term.KeyDown})

	buf := buffer.NewBuffer(80, 5)
	rg.Paint(buf)

	if bufNonBlank(buf) == 0 {
		t.Error("radio group should produce visible content")
	}
}

// --- Slider value → render ---

func TestP26B_SliderValueRender(t *testing.T) {
	s := component.NewSliderWithRange(0, 100, 50, 1)
	s.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})
	s.Measure(component.Constraints{})

	s.HandleKey(&term.KeyEvent{Key: term.KeyRight})

	buf := buffer.NewBuffer(80, 1)
	s.Paint(buf)

	if bufNonBlank(buf) == 0 {
		t.Error("slider should produce visible content")
	}
}

// --- Full screen layout with all panels ---

func TestP26B_FullScreenLayout(t *testing.T) {
	root := layout.NewFlex(layout.FlexColumn)

	// Top: tab bar
	tb := component.NewTabBar()
	tb.AddTab("tab1", "Dashboard")
	tb.AddTab("tab2", "Settings")
	root.AddChild(tb)

	// Middle: content (horizontal split)
	content := layout.NewFlex(layout.FlexRow)

	leftCol := layout.NewFlex(layout.FlexColumn)
	leftCol.AddChild(component.NewGauge())
	leftCol.AddChild(component.NewSparkline())
	content.AddChild(leftCol)

	content.AddChild(component.NewTable(
		[]string{"Name", "Value"},
		[]string{"CPU", "45%"},
		[]string{"MEM", "2.1GB"},
	))

	root.AddChild(content)

	// Bottom: status bar
	root.AddChild(component.NewStatusBar())

	root.Measure(component.Unbounded())
	root.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	buf := buffer.NewBuffer(80, 24)
	root.Paint(buf) // should not panic

	if bufNonBlank(buf) == 0 {
		t.Error("full screen layout should produce visible content")
	}
}

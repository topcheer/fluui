package component

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ============================================================
// P25-B: Component Edge Case & Stress Tests
// ============================================================

// --- Table ---

func TestP25B_TableLargeDataset(t *testing.T) {
	headers := []string{"ID", "Name", "Value"}
	var rows [][]string
	for i := 0; i < 10000; i++ {
		rows = append(rows, []string{fmt.Sprintf("%d", i), fmt.Sprintf("Item-%d", i), "active"})
	}
	table := NewTable(headers, rows...)
	table.SetBounds(Rect{0, 0, 80, 24})
	table.Measure(Constraints{})
	table.Paint(buffer.NewBuffer(80, 24))
}

func TestP25B_TableEmptyRows(t *testing.T) {
	table := NewTable([]string{"A", "B"})
	table.SetBounds(Rect{0, 0, 80, 10})
	table.Measure(Constraints{})
	table.Paint(buffer.NewBuffer(80, 10))
}

func TestP25B_TableSingleRow(t *testing.T) {
	table := NewTable([]string{"Col"}, []string{"Value"})
	table.SetBounds(Rect{0, 0, 80, 10})
	table.Measure(Constraints{})
	table.Paint(buffer.NewBuffer(80, 10))
}

func TestP25B_TableWideColumns(t *testing.T) {
	headers := []string{"Very Long Header Name"}
	rows := [][]string{{strings.Repeat("X", 500)}}
	table := NewTable(headers, rows...)
	table.SetBounds(Rect{0, 0, 80, 10})
	table.Measure(Constraints{})
	table.Paint(buffer.NewBuffer(80, 10))
}

// --- Checkbox ---

func TestP25B_CheckboxEmpty(t *testing.T) {
	cb := NewCheckbox(nil)
	cb.SetBounds(Rect{0, 0, 80, 24})
	cb.Measure(Constraints{})
	cb.Paint(buffer.NewBuffer(80, 24))
}

func TestP25B_CheckboxSingleItem(t *testing.T) {
	cb := NewCheckbox([]string{"Only one"})
	cb.SetBounds(Rect{0, 0, 80, 10})
	cb.Measure(Constraints{})
	cb.Paint(buffer.NewBuffer(80, 10))
}

func TestP25B_CheckboxManyItems(t *testing.T) {
	var labels []string
	for i := 0; i < 100; i++ {
		labels = append(labels, fmt.Sprintf("Item %d", i))
	}
	cb := NewCheckbox(labels)
	cb.SetBounds(Rect{0, 0, 80, 24})
	cb.Measure(Constraints{})
	cb.Paint(buffer.NewBuffer(80, 24))
}

// --- RadioGroup ---

func TestP25B_RadioGroupEmpty(t *testing.T) {
	rg := NewRadioGroup(nil)
	rg.SetBounds(Rect{0, 0, 80, 24})
	rg.Measure(Constraints{})
	rg.Paint(buffer.NewBuffer(80, 24))
}

func TestP25B_RadioGroupManyOptions(t *testing.T) {
	var labels []string
	for i := 0; i < 100; i++ {
		labels = append(labels, fmt.Sprintf("Option %d", i))
	}
	rg := NewRadioGroup(labels)
	rg.SetBounds(Rect{0, 0, 80, 24})
	rg.Measure(Constraints{})
	rg.Paint(buffer.NewBuffer(80, 24))
}

// --- Slider ---

func TestP25B_SliderBoundaries(t *testing.T) {
	s := NewSliderWithRange(0, 100, 50, 1)
	s.SetBounds(Rect{0, 0, 80, 1})
	s.Measure(Constraints{})
	s.Paint(buffer.NewBuffer(80, 1))

	s.SetValue(-100)
	if v := s.Value(); v < 0 {
		t.Errorf("value below min not clamped: got %v", v)
	}
	s.SetValue(200)
	if v := s.Value(); v > 100 {
		t.Errorf("value above max not clamped: got %v", v)
	}
}

func TestP25B_SliderZeroRange(t *testing.T) {
	s := NewSliderWithRange(50, 50, 50, 1)
	s.SetBounds(Rect{0, 0, 80, 1})
	s.Measure(Constraints{})
	s.Paint(buffer.NewBuffer(80, 1))
}

// --- Sparkline ---

func TestP25B_SparklineEmpty(t *testing.T) {
	sl := NewSparkline()
	sl.SetBounds(Rect{0, 0, 80, 5})
	sl.Measure(Constraints{})
	sl.Paint(buffer.NewBuffer(80, 5))
}

func TestP25B_SparklineSingleValue(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{42.0})
	sl.SetBounds(Rect{0, 0, 80, 5})
	sl.Measure(Constraints{})
	sl.Paint(buffer.NewBuffer(80, 5))
}

func TestP25B_SparklineLargeDataset(t *testing.T) {
	sl := NewSparkline()
	var data []float64
	for i := 0; i < 10000; i++ {
		data = append(data, float64(i)*1.5)
	}
	sl.SetData(data)
	sl.SetBounds(Rect{0, 0, 80, 5})
	sl.Measure(Constraints{})
	sl.Paint(buffer.NewBuffer(80, 5))
}

// --- Gauge ---

func TestP25B_GaugeBoundaries(t *testing.T) {
	g := NewGauge()
	g.SetValue(50)
	g.SetBounds(Rect{0, 0, 80, 1})
	g.Measure(Constraints{})
	g.Paint(buffer.NewBuffer(80, 1))

	g.SetValue(-10)
	g.SetValue(200)
	g.SetValue(100)
	g.Paint(buffer.NewBuffer(80, 1))
}

func TestP25B_GaugeZeroWidth(t *testing.T) {
	g := NewGauge()
	g.SetValue(50)
	g.SetBounds(Rect{0, 0, 0, 1})
	g.Measure(Constraints{})
	g.Paint(buffer.NewBuffer(0, 1))
}

// --- TabBar ---

func TestP25B_TabBarEmpty(t *testing.T) {
	tb := NewTabBar()
	tb.SetBounds(Rect{0, 0, 80, 1})
	tb.Measure(Constraints{})
	tb.Paint(buffer.NewBuffer(80, 1))
}

func TestP25B_TabBarManyTabs(t *testing.T) {
	tb := NewTabBar()
	for i := 0; i < 100; i++ {
		tb.AddTab(fmt.Sprintf("tab-%d", i), fmt.Sprintf("Tab %d", i))
	}
	tb.SetBounds(Rect{0, 0, 80, 1})
	tb.Measure(Constraints{})
	tb.Paint(buffer.NewBuffer(80, 1))
}

// --- Tree ---

func TestP25B_TreeEmpty(t *testing.T) {
	tree := NewTree()
	tree.SetBounds(Rect{0, 0, 80, 24})
	tree.Measure(Constraints{})
	tree.Paint(buffer.NewBuffer(80, 24))
}

func TestP25B_TreeDeepNesting(t *testing.T) {
	root := NewTreeNode("root", "Root")
	current := root
	for i := 0; i < 50; i++ {
		child := NewTreeNode(fmt.Sprintf("node-%d", i), fmt.Sprintf("Node %d", i))
		current.AddChild(child)
		current = child
	}
	tree := NewTree()
	tree.SetData(root)
	tree.SetBounds(Rect{0, 0, 80, 24})
	tree.Measure(Constraints{})
	tree.Paint(buffer.NewBuffer(80, 24))
}

func TestP25B_TreeWideChildren(t *testing.T) {
	root := NewTreeNode("root", "Root")
	for i := 0; i < 1000; i++ {
		root.AddChild(NewTreeNode(fmt.Sprintf("n%d", i), strings.Repeat("X", 200)))
	}
	tree := NewTree()
	tree.SetData(root)
	tree.SetBounds(Rect{0, 0, 80, 24})
	tree.Measure(Constraints{})
	tree.Paint(buffer.NewBuffer(80, 24))
}

// --- StatusBar ---

func TestP25B_StatusBarEmpty(t *testing.T) {
	sb := NewStatusBar()
	sb.SetBounds(Rect{0, 0, 80, 1})
	sb.Measure(Constraints{})
	sb.Paint(buffer.NewBuffer(80, 1))
}

// --- ProgressBar ---

func TestP25B_ProgressBarBoundaries(t *testing.T) {
	pb := NewProgressBar()
	pb.SetProgress(50)
	pb.SetBounds(Rect{0, 0, 80, 1})
	pb.Measure(Constraints{})
	pb.Paint(buffer.NewBuffer(80, 1))

	pb.SetProgress(-10)
	pb.SetProgress(200)
	pb.SetProgress(100)
	pb.Paint(buffer.NewBuffer(80, 1))
}

// --- CommandPalette ---

func TestP25B_CommandPaletteEmpty(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetBounds(Rect{0, 0, 80, 24})
	cp.Measure(Constraints{})
	cp.Paint(buffer.NewBuffer(80, 24))
}

func TestP25B_CommandPaletteManyCommands(t *testing.T) {
	cp := NewCommandPalette()
	for i := 0; i < 500; i++ {
		cp.AddCommand(Command{ID: fmt.Sprintf("cmd-%d", i), Label: fmt.Sprintf("Command %d", i), Category: "test"})
	}
	cp.SetBounds(Rect{0, 0, 80, 24})
	cp.Measure(Constraints{})
	cp.Paint(buffer.NewBuffer(80, 24))
}

func TestP25B_CommandPaletteSearchLongQuery(t *testing.T) {
	cp := NewCommandPalette()
	for i := 0; i < 100; i++ {
		cp.AddCommand(Command{ID: fmt.Sprintf("cmd-%d", i), Label: fmt.Sprintf("Command %d", i), Category: "test"})
	}
	cp.SetQuery(strings.Repeat("a", 1000))
	cp.SetBounds(Rect{0, 0, 80, 24})
	cp.Measure(Constraints{})
	cp.Paint(buffer.NewBuffer(80, 24))
}

// --- Spinner ---

func TestP25B_SpinnerRapidStartStop(t *testing.T) {
	s := NewSpinner("Loading")
	s.SetBounds(Rect{0, 0, 80, 1})
	s.Measure(Constraints{})
	for i := 0; i < 100; i++ {
		s.Start()
		s.Stop()
	}
	s.Paint(buffer.NewBuffer(80, 1))
}

func TestP25B_SpinnerEmptyLabel(t *testing.T) {
	s := NewSpinner("")
	s.SetBounds(Rect{0, 0, 80, 1})
	s.Measure(Constraints{})
	s.Start()
	s.Paint(buffer.NewBuffer(80, 1))
	s.Stop()
}

// --- AutoComplete ---

func TestP25B_AutoCompleteEmpty(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetBounds(Rect{0, 0, 80, 24})
	ac.Measure(Constraints{})
	ac.Paint(buffer.NewBuffer(80, 24))
}

func TestP25B_AutoCompleteManyItems(t *testing.T) {
	ac := NewAutoComplete()
	var items []CompletionItem
	for i := 0; i < 1000; i++ {
		items = append(items, CompletionItem{Label: fmt.Sprintf("suggestion-%d", i)})
	}
	ac.SetItems(items)
	ac.SetBounds(Rect{0, 0, 80, 24})
	ac.Measure(Constraints{})
	ac.Paint(buffer.NewBuffer(80, 24))
}

// --- FilePicker ---

func TestP25B_FilePickerEmptyDir(t *testing.T) {
	fp := NewFilePicker(".")
	fp.SetBounds(Rect{0, 0, 80, 24})
	fp.Measure(Constraints{})
	fp.Paint(buffer.NewBuffer(80, 24))
}

// --- Pagination ---

func TestP25B_PaginationZeroItems(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(0)
	p.SetItemsPerPage(10)
	if p.TotalPages() != 0 {
		t.Errorf("0 items should have 0 pages, got %d", p.TotalPages())
	}
}

func TestP25B_PaginationExactMultiple(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)
	if p.TotalPages() != 10 {
		t.Errorf("100/10 = 10 pages, got %d", p.TotalPages())
	}
}

func TestP25B_PaginationRemainder(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(105)
	p.SetItemsPerPage(10)
	if p.TotalPages() != 11 {
		t.Errorf("105/10 = 11 pages, got %d", p.TotalPages())
	}
}

func TestP25B_PaginationNavigate(t *testing.T) {
	p := NewPagination()
	p.SetTotalItems(100)
	p.SetItemsPerPage(10)
	for i := 0; i < 9; i++ {
		p.NextPage()
	}
	if p.HasNext() {
		t.Error("should be on last page")
	}
	for i := 0; i < 9; i++ {
		p.PrevPage()
	}
	if p.HasPrev() {
		t.Error("should be on first page")
	}
}

// --- VirtualScroller ---

func TestP25B_VirtualScrollerHugeDataset(t *testing.T) {
	vs := NewVirtualScroller()
	var items []VirtualItem
	for i := 0; i < 100000; i++ {
		items = append(items, VirtualItem{ID: fmt.Sprintf("item-%d", i), Text: fmt.Sprintf("Item %d", i)})
	}
	vs.SetItems(items)
	vs.SetBounds(Rect{0, 0, 80, 24})
	vs.Measure(Constraints{})
	vs.Paint(buffer.NewBuffer(80, 24))
}

func TestP25B_VirtualScrollerEmpty(t *testing.T) {
	vs := NewVirtualScroller()
	vs.SetBounds(Rect{0, 0, 80, 24})
	vs.Measure(Constraints{})
	vs.Paint(buffer.NewBuffer(80, 24))
}

func TestP25B_VirtualScrollerConcurrentAccess(t *testing.T) {
	vs := NewVirtualScroller()
	var wg sync.WaitGroup
	const goroutines = 10
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			var items []VirtualItem
			for j := 0; j < 100; j++ {
				items = append(items, VirtualItem{ID: fmt.Sprintf("g%d-%d", id, j), Text: fmt.Sprintf("Item %d", j)})
			}
			vs.SetItems(items)
		}(i)
	}
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			vs.Cursor()
			vs.Items()
		}()
	}
	wg.Wait()
}

// --- Concurrent paint ---

func TestP25B_ConcurrentPaint(t *testing.T) {
	table := NewTable([]string{"A", "B"}, []string{"1", "2"}, []string{"3", "4"})
	table.SetBounds(Rect{0, 0, 80, 24})
	table.Measure(Constraints{})

	var wg sync.WaitGroup
	const goroutines = 20
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			table.Paint(buffer.NewBuffer(80, 24))
		}()
	}
	wg.Wait()
}

func TestP25B_RapidResizeCycle(t *testing.T) {
	table := NewTable([]string{"Col1", "Col2", "Col3"},
		[]string{"a", "b", "c"},
		[]string{"d", "e", "f"},
	)
	sizes := []Rect{
		{0, 0, 1, 1},
		{0, 0, 200, 50},
		{0, 0, 80, 24},
		{0, 0, 10, 3},
		{0, 0, 500, 100},
	}
	for _, s := range sizes {
		table.SetBounds(s)
		table.Measure(Constraints{})
		table.Paint(buffer.NewBuffer(s.W, s.H))
	}
}

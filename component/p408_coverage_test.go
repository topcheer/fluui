package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P408: Coverage for circular_progress Paint, divider Measure,
// breadcrumb Paint, filetree Measure, markdown_stream Paint,
// heatmap Measure, metricbar Measure

// === CircularProgress Paint branches ===

func TestP408_CircularProgress_Paint_AllStyles(t *testing.T) {
	for _, style := range []CircularProgressStyle{ProgressStyleRing, ProgressStyleDots, ProgressStyleBlock} {
		for _, v := range []float64{0, 0.1, 0.25, 0.4, 0.5, 0.7, 0.8, 0.95, 1.0} {
			c := NewCircularProgress(v)
			c.SetStyle(style)
			c.SetLabel("Test")
			c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
			buf := buffer.NewBuffer(20, 1)
			c.Paint(buf)
		}
	}
}

func TestP408_CircularProgress_Paint_CustomColor(t *testing.T) {
	c := NewCircularProgress(0.5)
	c.SetColor(buffer.RGB(1, 2, 3))
	c.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 1})
	buf := buffer.NewBuffer(15, 1)
	c.Paint(buf)
}

func TestP408_CircularProgress_Paint_NonZeroOffset(t *testing.T) {
	c := NewCircularProgress(0.75)
	c.SetLabel("Load")
	c.SetBounds(Rect{X: 10, Y: 5, W: 20, H: 1})
	buf := buffer.NewBuffer(40, 10)
	c.Paint(buf)
}

// === Divider Measure edge cases ===

func TestP408_Divider_Measure_VerticalZero(t *testing.T) {
	d := NewDivider("")
	d.SetOrientation(DividerVertical)
	s := d.Measure(Constraints{}) // no MaxHeight → default 1
	if s.W != 1 { t.Errorf("W = %d, want 1", s.W) }
	if s.H != 1 { t.Errorf("H = %d, want 1", s.H) }
}

func TestP408_Divider_Measure_HorizontalZero(t *testing.T) {
	d := NewDivider("Section")
	s := d.Measure(Constraints{}) // no MaxWidth → default 20
	if s.H != 1 { t.Errorf("H = %d", s.H) }
}

func TestP408_Divider_Measure_VerticalClamped(t *testing.T) {
	d := NewDivider("")
	d.SetOrientation(DividerVertical)
	s := d.Measure(Constraints{MaxWidth: 5, MaxHeight: 3})
	if s.W != 1 { t.Errorf("W = %d, want 1", s.W) }
	if s.H != 3 { t.Errorf("H = %d, want 3", s.H) }
}

// === Breadcrumb Paint: multi-item + delimiter + truncation ===

func TestP408_Breadcrumb_Paint_MultiItem(t *testing.T) {
	b := NewBreadcrumb([]string{"Home", "Settings", "Audio"})
	b.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != 'H' { t.Errorf("cell[0] = %q", string(c.Rune)) }
}

func TestP408_Breadcrumb_Paint_Truncate(t *testing.T) {
	b := NewBreadcrumb([]string{"VeryLongItem", "Another", "Third"})
	b.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})
	buf := buffer.NewBuffer(5, 1)
	b.Paint(buf) // should truncate with ellipsis
}

// === FileTree Measure: collapsed + nested ===

func TestP408_FileTree_Measure_Collapsed(t *testing.T) {
	ft := NewFileTree("proj", []FileNode{
		{Name: "src", IsDir: true, Expanded: false, Children: []FileNode{
			{Name: "main.go"},
		}},
	})
	s := ft.Measure(Constraints{MaxWidth: 40, MaxHeight: 20})
	if s.H != 2 { t.Errorf("H = %d, want 2 (root + collapsed dir)", s.H) }
}

func TestP408_FileTree_Measure_Empty(t *testing.T) {
	ft := NewFileTree("empty", nil)
	s := ft.Measure(Constraints{MaxWidth: 40, MaxHeight: 20})
	if s.H != 1 { t.Errorf("H = %d, want 1", s.H) }
}

// === MarkdownStream Paint: streaming cursor with multiline ===

func TestP408_MDStream_Paint_LongSource(t *testing.T) {
	m := NewMarkdownStream()
	m.SetSource("Line 1\nLine 2\nLine 3\nLine 4\nLine 5")
	m.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	m.Paint(buf)
}

func TestP408_MDStream_Paint_CursorOff(t *testing.T) {
	m := NewMarkdownStream()
	m.SetSource("Hello")
	m.SetCursorOn(false)
	m.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	m.Paint(buf) // no cursor drawn
}

// === Heatmap Measure edge cases ===

func TestP408_Heatmap_Measure_EdgeCases(t *testing.T) {
	h := NewHeatmap(5, 3)
	s := h.Measure(Constraints{})
	if s.W < 1 || s.H < 1 { t.Error("should be >= 1") }
}

// === MetricBar Measure: with unit + no constraints ===

func TestP408_MetricBar_Measure_EdgeCases(t *testing.T) {
	m := NewMetricBar("CPU", 50, 0, 100)
	m.SetUnit("%")
	s := m.Measure(Constraints{})
	if s.W < 1 { t.Error("W should be >= 1") }
}

// === DiffViewer Measure edge cases ===


// === FilePicker Paint edge ===

func TestP408_FilePicker_Measure_EdgeCases(t *testing.T) {
	fp := NewFilePicker(".")
	s := fp.Measure(Constraints{})
	if s.W < 1 { t.Error("W should be >= 1") }
}

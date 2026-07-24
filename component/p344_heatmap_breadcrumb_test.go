package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Heatmap Tests ────────────────────────────────────────

func TestP344_Heatmap_Create(t *testing.T) {
	h := NewHeatmap(7, 24) // 7 days × 24 hours
	rows, cols := h.Dimensions()
	if rows != 7 || cols != 24 {
		t.Errorf("dims = %dx%d, want 7x24", rows, cols)
	}
}

func TestP344_Heatmap_SetCell(t *testing.T) {
	h := NewHeatmap(3, 5)
	h.SetCell(1, 2, 50)
	h.SetCell(2, 4, 100)

	data := h.Data()
	if data[1][2].Value != 50 {
		t.Errorf("cell(1,2) = %d, want 50", data[1][2].Value)
	}
	if data[2][4].Value != 100 {
		t.Errorf("cell(2,4) = %d, want 100", data[2][4].Value)
	}
	if h.MaxValue() != 100 {
		t.Errorf("max = %d, want 100", h.MaxValue())
	}
}

func TestP344_Heatmap_SetCell_Invalid(t *testing.T) {
	h := NewHeatmap(2, 3)
	h.SetCell(-1, 0, 10) // no panic
	h.SetCell(0, -1, 10)
	h.SetCell(99, 99, 10)
}

func TestP344_Heatmap_SetData(t *testing.T) {
	h := NewHeatmap(2, 2)
	data := [][]HeatmapCell{
		{{Value: 10}, {Value: 20}},
		{{Value: 30}, {Value: 40}},
	}
	h.SetData(data)
	rows, cols := h.Dimensions()
	if rows != 2 || cols != 2 {
		t.Errorf("dims = %dx%d", rows, cols)
	}
	if h.MaxValue() != 40 {
		t.Errorf("max = %d, want 40", h.MaxValue())
	}
}

func TestP344_Heatmap_ColLabels(t *testing.T) {
	h := NewHeatmap(3, 7)
	h.SetColLabels([]string{"M", "T", "W", "T", "F", "S", "S"})
	m := h.Measure(Constraints{MaxWidth: 50, MaxHeight: 10})
	if m.H < 4 { // 3 rows + 1 label row
		t.Errorf("height = %d, expected at least 4", m.H)
	}
}

func TestP344_Heatmap_Measure(t *testing.T) {
	h := NewHeatmap(5, 10)
	s := h.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if s.H < 5 {
		t.Errorf("height = %d, expected at least 5", s.H)
	}
	if s.W < 10 {
		t.Errorf("width = %d, expected at least 10", s.W)
	}
}

func TestP344_Heatmap_Paint(t *testing.T) {
	h := NewHeatmap(3, 5)
	h.SetCell(0, 0, 10)
	h.SetCell(1, 1, 50)
	h.SetCell(2, 2, 100)
	h.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	h.Paint(buf)

	// Cells should have non-default Bg where values > 0
	cell := buf.GetCell(0, 0)
	if cell.Bg.Type == 0 && cell.Bg.Val == 0 {
		t.Error("expected non-default background at (0,0)")
	}
}

func TestP344_Heatmap_Paint_Empty(t *testing.T) {
	h := NewHeatmap(0, 0)
	h.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	h.Paint(buf) // should not panic
}

func TestP344_Heatmap_IntensityLevels(t *testing.T) {
	if intensityLevel(0, 100) != 0 {
		t.Error("0/max should be level 0")
	}
	if intensityLevel(10, 100) != 1 {
		t.Error("10% should be level 1")
	}
	if intensityLevel(30, 100) != 2 {
		t.Error("30% should be level 2")
	}
	if intensityLevel(60, 100) != 3 {
		t.Error("60% should be level 3")
	}
	if intensityLevel(80, 100) != 4 {
		t.Error("80% should be level 4")
	}
	if intensityLevel(50, 0) != 0 {
		t.Error("max=0 should give level 0")
	}
}

// ─── Breadcrumb Tests ─────────────────────────────────────

func TestP344_Breadcrumb_Create(t *testing.T) {
	b := NewBreadcrumb([]string{"Home", "Settings", "AI"})
	if b.ItemCount() != 3 {
		t.Errorf("count = %d, want 3", b.ItemCount())
	}
}

func TestP344_Breadcrumb_SetItems(t *testing.T) {
	b := NewBreadcrumb(nil)
	b.SetItems([]string{"A", "B", "C", "D"})
	if b.ItemCount() != 4 {
		t.Errorf("count = %d", b.ItemCount())
	}
}

func TestP344_Breadcrumb_PushPop(t *testing.T) {
	b := NewBreadcrumb([]string{"Home"})
	b.Push("Files")
	b.Push("Documents")
	if b.ItemCount() != 3 {
		t.Errorf("count = %d after push", b.ItemCount())
	}
	last := b.Pop()
	if last != "Documents" {
		t.Errorf("pop = %q, want Documents", last)
	}
	if b.ItemCount() != 2 {
		t.Errorf("count = %d after pop", b.ItemCount())
	}
}

func TestP344_Breadcrumb_Pop_Empty(t *testing.T) {
	b := NewBreadcrumb(nil)
	if b.Pop() != "" {
		t.Error("pop on empty should return empty string")
	}
}

func TestP344_Breadcrumb_SetActive(t *testing.T) {
	b := NewBreadcrumb([]string{"A", "B", "C"})
	b.SetActive(1)
	if b.ActiveIndex() != 1 {
		t.Errorf("active = %d, want 1", b.ActiveIndex())
	}
}

func TestP344_Breadcrumb_SetDelimiter(t *testing.T) {
	b := NewBreadcrumb([]string{"A", "B"})
	b.SetDelimiter(" / ")
	if b.String() != "A / B" {
		t.Errorf("string = %q", b.String())
	}
}

func TestP344_Breadcrumb_String(t *testing.T) {
	b := NewBreadcrumb([]string{"Home", "Settings", "AI"})
	s := b.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}

func TestP344_Breadcrumb_Measure(t *testing.T) {
	b := NewBreadcrumb([]string{"Home", "Settings"})
	s := b.Measure(Constraints{MaxWidth: 80, MaxHeight: 1})
	if s.H != 1 {
		t.Errorf("height = %d, want 1", s.H)
	}
	if s.W < 5 {
		t.Errorf("width = %d, too small", s.W)
	}
}

func TestP344_Breadcrumb_Paint(t *testing.T) {
	b := NewBreadcrumb([]string{"Home", "Settings", "AI"})
	b.SetActive(1)
	b.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.Paint(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected non-empty cell")
	}
}

func TestP344_Breadcrumb_Paint_Empty(t *testing.T) {
	b := NewBreadcrumb(nil)
	b.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.Paint(buf) // should not panic
}

func TestP344_Breadcrumb_Paint_Truncation(t *testing.T) {
	b := NewBreadcrumb([]string{"VeryLongItem1", "VeryLongItem2", "VeryLongItem3"})
	b.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	b.Paint(buf) // should not panic, should truncate
}

func BenchmarkHeatmap_Paint(b *testing.B) {
	h := NewHeatmap(7, 24)
	for r := 0; r < 7; r++ {
		for c := 0; c < 24; c++ {
			h.SetCell(r, c, r*c)
		}
	}
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h.Paint(buf)
	}
}

func BenchmarkBreadcrumb_Paint(b *testing.B) {
	bc := NewBreadcrumb([]string{"Home", "Settings", "AI", "Models", "GPT-4"})
	bc.SetActive(2)
	bc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bc.Paint(buf)
	}
}

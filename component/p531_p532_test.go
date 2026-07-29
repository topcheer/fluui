package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestTokenStreamVisualizerBasic(t *testing.T) {
	tsv := NewTokenStreamVisualizer()
	tsv.SetText("Hello AI")
	tsv.SetReceived(10)
	tsv.SetTotal(50)
	tsv.SetStatus(TokenStreamActive)
	if tsv.Text() != "Hello AI" { t.Errorf("Text = %q", tsv.Text()) }
	if tsv.Status() != TokenStreamActive { t.Errorf("Status = %d", tsv.Status()) }
}

func TestTokenStreamVisualizerStatusIcons(t *testing.T) {
	if tsStatusIcon(TokenStreamIdle) != '○' { t.Error("idle icon") }
	if tsStatusIcon(TokenStreamActive) != '▶' { t.Error("active icon") }
	if tsStatusIcon(TokenStreamDone) != '✓' { t.Error("done icon") }
	if tsStatusIcon(TokenStreamError) != '✗' { t.Error("error icon") }
}

func TestTokenStreamVisualizerMeasure(t *testing.T) {
	tsv := NewTokenStreamVisualizer()
	s := tsv.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestTokenStreamVisualizerPaint(t *testing.T) {
	tsv := NewTokenStreamVisualizer()
	tsv.SetText("Streaming text")
	tsv.SetReceived(5)
	tsv.SetTotal(20)
	tsv.SetStatus(TokenStreamActive)
	tsv.SetCursor(true)
	tsv.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 5})
	buf := buffer.NewBuffer(50, 5)
	tsv.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	foundIcon := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == '▶' { foundIcon = true; break }
	}
	if !foundIcon { t.Error("status icon not found") }
}

func TestTokenStreamVisualizerPaintCursor(t *testing.T) {
	tsv := NewTokenStreamVisualizer()
	tsv.SetText("ABC")
	tsv.SetStatus(TokenStreamActive)
	tsv.SetCursor(true)
	tsv.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 5})
	buf := buffer.NewBuffer(50, 5)
	tsv.Paint(buf)
	foundCursor := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 2).Rune == '▋' { foundCursor = true; break }
	}
	if !foundCursor { t.Error("cursor not found") }
}

func TestTokenStreamVisualizerChildren(t *testing.T) {
	tsv := NewTokenStreamVisualizer()
	if tsv.Children() != nil { t.Error("Children should be nil") }
}

func TestTokenStreamVisualizerStyle(t *testing.T) {
	tsv := NewTokenStreamVisualizer()
	tsv.SetStyle(TokenStreamStyle{Text: buffer.Style{Fg: buffer.RGB(200,200,200)}, Cursor: buffer.Style{Fg: buffer.RGB(0,255,0)}, Progress: buffer.Style{Fg: buffer.RGB(255,0,0)}, Status: [4]buffer.Style{{}, {}, {}, {}}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	tsv.SetText("x")
	tsv.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	tsv.Paint(buf)
}

// ─── LatencyHeatmap tests ───

func TestLatencyHeatmapBasic(t *testing.T) {
	lh := NewLatencyHeatmap()
	lh.SetColumns([]string{"/api", "/db"})
	lh.AddRow("09:00", []int{50, 200})
	if lh.RowCount() != 1 { t.Errorf("RowCount = %d, want 1", lh.RowCount()) }
}

func TestLatencyHeatmapMultiple(t *testing.T) {
	lh := NewLatencyHeatmap()
	lh.SetColumns([]string{"A", "B", "C"})
	lh.AddRow("08:00", []int{10, 20, 30})
	lh.AddRow("09:00", []int{50, 60, 70})
	lh.AddRow("10:00", []int{100, 500, 800})
	if lh.RowCount() != 3 { t.Errorf("RowCount = %d, want 3", lh.RowCount()) }
}

func TestLatencyHeatmapClear(t *testing.T) {
	lh := NewLatencyHeatmap()
	lh.AddRow("x", []int{1})
	lh.Clear()
	if lh.RowCount() != 0 { t.Errorf("RowCount = %d, want 0", lh.RowCount()) }
}

func TestLatencyHeatmapEmpty(t *testing.T) {
	lh := NewLatencyHeatmap()
	if lh.RowCount() != 0 { t.Errorf("RowCount = %d, want 0", lh.RowCount()) }
}

func TestLatencyHeatmapMeasure(t *testing.T) {
	lh := NewLatencyHeatmap()
	lh.SetColumns([]string{"A", "B"})
	lh.AddRow("x", []int{1, 2})
	s := lh.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 4 { t.Errorf("H = %d", s.H) }
}

func TestLatencyHeatmapPaint(t *testing.T) {
	lh := NewLatencyHeatmap()
	lh.SetColumns([]string{"/api", "/db"})
	lh.SetThresholds(100, 500)
	lh.AddRow("09:00", []int{50, 600})
	lh.AddRow("10:00", []int{200, 800})
	lh.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 8})
	buf := buffer.NewBuffer(50, 8)
	lh.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
}

func TestLatencyHeatmapPaintEmpty(t *testing.T) {
	lh := NewLatencyHeatmap()
	lh.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 6})
	buf := buffer.NewBuffer(40, 6)
	lh.Paint(buf)
}

func TestLatencyHeatmapChildren(t *testing.T) {
	lh := NewLatencyHeatmap()
	if lh.Children() != nil { t.Error("Children should be nil") }
}

func TestLatencyHeatmapStyle(t *testing.T) {
	lh := NewLatencyHeatmap()
	lh.SetStyle(LatencyHeatmapStyle{Normal: buffer.Style{Fg: buffer.RGB(0,255,0)}, Warning: buffer.Style{Fg: buffer.RGB(255,255,0)}, Critical: buffer.Style{Fg: buffer.RGB(255,0,0)}, Label: buffer.Style{Fg: buffer.RGB(150,150,150)}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	lh.SetColumns([]string{"X"})
	lh.AddRow("r", []int{999})
	lh.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 6})
	buf := buffer.NewBuffer(40, 6)
	lh.Paint(buf)
}

func BenchmarkPaintTokenStreamVisualizer(b *testing.B) {
	tsv := NewTokenStreamVisualizer()
	tsv.SetText("The quick brown fox jumps over the lazy dog in the terminal UI rendering system")
	tsv.SetReceived(45)
	tsv.SetTotal(100)
	tsv.SetStatus(TokenStreamActive)
	tsv.SetCursor(true)
	tsv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tsv.Paint(buf)
	}
}

func BenchmarkPaintLatencyHeatmap(b *testing.B) {
	lh := NewLatencyHeatmap()
	lh.SetColumns([]string{"/api/v1", "/auth", "/db/query", "/cache"})
	lh.SetThresholds(100, 500)
	lh.AddRow("08:00", []int{45, 80, 120, 30})
	lh.AddRow("09:00", []int{60, 150, 200, 45})
	lh.AddRow("10:00", []int{85, 300, 550, 60})
	lh.AddRow("11:00", []int{120, 600, 800, 90})
	lh.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 8})
	buf := buffer.NewBuffer(60, 8)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		lh.Paint(buf)
	}
}

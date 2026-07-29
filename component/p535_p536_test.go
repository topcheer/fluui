package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ConfidenceIntervalChart tests ───

func TestConfidenceIntervalBasic(t *testing.T) {
	cic := NewConfidenceIntervalChart()
	cic.SetLabel("Accuracy")
	cic.SetBounds(0.72, 0.85, 0.92)
	if cic.Label() != "Accuracy" { t.Errorf("Label = %q", cic.Label()) }
}

func TestConfidenceIntervalRange(t *testing.T) {
	cic := NewConfidenceIntervalChart()
	cic.SetRange(0, 100)
	cic.SetBounds(40, 55, 70)
}

func TestConfidenceIntervalMeasure(t *testing.T) {
	cic := NewConfidenceIntervalChart()
	s := cic.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestConfidenceIntervalPaint(t *testing.T) {
	cic := NewConfidenceIntervalChart()
	cic.SetLabel("Accuracy")
	cic.SetBounds(0.72, 0.85, 0.92)
	cic.SetBounds_(Rect{X: 0, Y: 0, W: 50, H: 4})
	buf := buffer.NewBuffer(50, 4)
	cic.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
}

func TestConfidenceIntervalPaintEmpty(t *testing.T) {
	cic := NewConfidenceIntervalChart()
	cic.SetBounds_(Rect{X: 0, Y: 0, W: 50, H: 4})
	buf := buffer.NewBuffer(50, 4)
	cic.Paint(buf)
}

func TestConfidenceIntervalChildren(t *testing.T) {
	cic := NewConfidenceIntervalChart()
	if cic.Children() != nil { t.Error("Children should be nil") }
}

func TestConfidenceIntervalStyle(t *testing.T) {
	cic := NewConfidenceIntervalChart()
	cic.SetStyleS(ConfidenceIntervalStyle{Label: buffer.Style{Fg: buffer.RGB(150,150,150)}, Mean: buffer.Style{Fg: buffer.RGB(0,255,0)}, Range: buffer.Style{Fg: buffer.RGB(100,100,255)}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	cic.SetBounds(0.3, 0.5, 0.7)
	cic.SetBounds_(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	cic.Paint(buf)
}

// ─── DatasetDistributionHistogram tests ───

func TestHistogramBasic(t *testing.T) {
	h := NewDatasetDistributionHistogram()
	h.AddBin("0-10", 45)
	if h.BinCount() != 1 { t.Errorf("BinCount = %d, want 1", h.BinCount()) }
}

func TestHistogramMultiple(t *testing.T) {
	h := NewDatasetDistributionHistogram()
	h.AddBin("A", 10)
	h.AddBin("B", 20)
	h.AddBin("C", 30)
	if h.BinCount() != 3 { t.Errorf("BinCount = %d, want 3", h.BinCount()) }
}

func TestHistogramClear(t *testing.T) {
	h := NewDatasetDistributionHistogram()
	h.AddBin("A", 10)
	h.Clear()
	if h.BinCount() != 0 { t.Errorf("BinCount = %d, want 0", h.BinCount()) }
}

func TestHistogramEmpty(t *testing.T) {
	h := NewDatasetDistributionHistogram()
	if h.BinCount() != 0 { t.Errorf("BinCount = %d, want 0", h.BinCount()) }
}

func TestHistogramMaxBarHeight(t *testing.T) {
	h := NewDatasetDistributionHistogram()
	h.SetMaxBarHeight(5)
	h.AddBin("A", 100)
	h.AddBin("B", 50)
	h.mu.Lock()
	if h.bins[0].BarH != 5 { t.Errorf("max bin height = %d, want 5", h.bins[0].BarH) }
	if h.bins[1].BarH != 2 { t.Errorf("half bin height = %d, want 2", h.bins[1].BarH) }
	h.mu.Unlock()
}

func TestHistogramMeasure(t *testing.T) {
	h := NewDatasetDistributionHistogram()
	h.AddBin("X", 10)
	s := h.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 5 { t.Errorf("H = %d", s.H) }
}

func TestHistogramPaint(t *testing.T) {
	h := NewDatasetDistributionHistogram()
	h.SetLabel("Token Lengths")
	h.AddBin("0-10", 45)
	h.AddBin("10-20", 120)
	h.AddBin("20-30", 80)
	h.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	h.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	foundBar := false
	for y2 := 0; y2 < 10; y2++ {
		for x2 := 0; x2 < 40; x2++ {
			if buf.GetCell(x2, y2).Rune == '█' { foundBar = true; break }
		}
	}
	if !foundBar { t.Error("bar not found") }
}

func TestHistogramPaintEmpty(t *testing.T) {
	h := NewDatasetDistributionHistogram()
	h.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 8})
	buf := buffer.NewBuffer(40, 8)
	h.Paint(buf)
}

func TestHistogramChildren(t *testing.T) {
	h := NewDatasetDistributionHistogram()
	if h.Children() != nil { t.Error("Children should be nil") }
}

func TestHistogramStyle(t *testing.T) {
	h := NewDatasetDistributionHistogram()
	h.SetStyleS(HistogramStyle{Bar: buffer.Style{Fg: buffer.RGB(0,255,0)}, Label: buffer.Style{Fg: buffer.RGB(150,150,150)}, Value: buffer.Style{Fg: buffer.RGB(255,255,255)}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	h.AddBin("X", 50)
	h.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 8})
	buf := buffer.NewBuffer(40, 8)
	h.Paint(buf)
}

func TestHistogramLabel(t *testing.T) {
	h := NewDatasetDistributionHistogram()
	h.SetLabel("Distribution")
	if h.Label() != "Distribution" { t.Errorf("Label = %q", h.Label()) }
}

// Benchmarks

func BenchmarkPaintConfidenceIntervalChart(b *testing.B) {
	cic := NewConfidenceIntervalChart()
	cic.SetLabel("Model Accuracy Estimate")
	cic.SetBounds(0.72, 0.85, 0.92)
	cic.SetBounds_(Rect{X: 0, Y: 0, W: 60, H: 4})
	buf := buffer.NewBuffer(60, 4)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cic.Paint(buf)
	}
}

func BenchmarkPaintDatasetHistogram(b *testing.B) {
	h := NewDatasetDistributionHistogram()
	h.SetLabel("Token Length Distribution")
	h.AddBin("0-5", 15)
	h.AddBin("5-10", 45)
	h.AddBin("10-15", 120)
	h.AddBin("15-20", 80)
	h.AddBin("20-25", 35)
	h.AddBin("25-30", 10)
	h.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 12})
	buf := buffer.NewBuffer(50, 12)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		h.Paint(buf)
	}
}

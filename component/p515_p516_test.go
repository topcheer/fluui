package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestMarkdownSuperscriptBasic(t *testing.T) {
	ms := NewMarkdownSuperscript()
	ms.SetMarkdown("E = mc^2^")
	if ms.SuperscriptCount() != 1 { t.Errorf("SuperscriptCount = %d, want 1", ms.SuperscriptCount()) }
}

func TestMarkdownSuperscriptMultiple(t *testing.T) {
	ms := NewMarkdownSuperscript()
	ms.SetMarkdown("x^n^ + y^n^ = z^n^")
	if ms.SuperscriptCount() != 3 { t.Errorf("SuperscriptCount = %d, want 3", ms.SuperscriptCount()) }
}

func TestMarkdownSuperscriptNoMarker(t *testing.T) {
	ms := NewMarkdownSuperscript()
	ms.SetMarkdown("Just regular text")
	if ms.SuperscriptCount() != 0 { t.Errorf("SuperscriptCount = %d, want 0", ms.SuperscriptCount()) }
}

func TestMarkdownSuperscriptEmpty(t *testing.T) {
	ms := NewMarkdownSuperscript()
	ms.SetMarkdown("")
	if ms.SuperscriptCount() != 0 { t.Errorf("SuperscriptCount = %d, want 0", ms.SuperscriptCount()) }
}

func TestMarkdownSuperscriptUnclosed(t *testing.T) {
	ms := NewMarkdownSuperscript()
	ms.SetMarkdown("Unclosed ^caret")
	if ms.SuperscriptCount() != 0 { t.Errorf("SuperscriptCount = %d, want 0 (unclosed)", ms.SuperscriptCount()) }
}

func TestMarkdownSuperscriptMeasure(t *testing.T) {
	ms := NewMarkdownSuperscript()
	ms.SetMarkdown("x^2^")
	s := ms.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestMarkdownSuperscriptPaint(t *testing.T) {
	ms := NewMarkdownSuperscript()
	ms.SetMarkdown("E = mc^2^ energy")
	ms.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 4})
	buf := buffer.NewBuffer(50, 4)
	ms.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	foundText := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == 'E' { foundText = true; break }
	}
	if !foundText { t.Error("text not found") }
}

func TestMarkdownSuperscriptChildren(t *testing.T) {
	ms := NewMarkdownSuperscript()
	if ms.Children() != nil { t.Error("Children should be nil") }
}

func TestMarkdownSuperscriptStyle(t *testing.T) {
	ms := NewMarkdownSuperscript()
	ms.SetStyle(SuperscriptStyle{Text: buffer.Style{Fg: buffer.RGB(200,200,200)}, Superscript: buffer.Style{Fg: buffer.RGB(255,165,0)}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	ms.SetMarkdown("x^2^")
	ms.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	ms.Paint(buf)
}

// ─── GaugeCluster tests ───

func TestGaugeClusterBasic(t *testing.T) {
	gc := NewGaugeCluster()
	gc.AddGauge("CPU", 75, 100)
	if gc.GaugeCount() != 1 { t.Errorf("GaugeCount = %d, want 1", gc.GaugeCount()) }
}

func TestGaugeClusterMultiple(t *testing.T) {
	gc := NewGaugeCluster()
	gc.AddGauge("CPU", 75, 100)
	gc.AddGauge("Mem", 50, 100)
	gc.AddGauge("Disk", 90, 100)
	if gc.GaugeCount() != 3 { t.Errorf("GaugeCount = %d, want 3", gc.GaugeCount()) }
}

func TestGaugeClusterColumns(t *testing.T) {
	gc := NewGaugeCluster()
	gc.SetColumns(3)
	gc.AddGauge("A", 50, 100)
	gc.AddGauge("B", 60, 100)
	gc.AddGauge("C", 70, 100)
	gc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 4})
	buf := buffer.NewBuffer(60, 4)
	gc.Paint(buf)
}

func TestGaugeClusterClear(t *testing.T) {
	gc := NewGaugeCluster()
	gc.AddGauge("A", 50, 100)
	gc.Clear()
	if gc.GaugeCount() != 0 { t.Errorf("GaugeCount = %d, want 0", gc.GaugeCount()) }
}

func TestGaugeClusterEmpty(t *testing.T) {
	gc := NewGaugeCluster()
	if gc.GaugeCount() != 0 { t.Errorf("GaugeCount = %d, want 0", gc.GaugeCount()) }
}

func TestGaugeClusterMeasure(t *testing.T) {
	gc := NewGaugeCluster()
	gc.AddGauge("A", 50, 100)
	s := gc.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestGaugeClusterPaint(t *testing.T) {
	gc := NewGaugeCluster()
	gc.SetColumns(2)
	gc.AddGauge("CPU", 75, 100)
	gc.AddGauge("Mem", 50, 100)
	gc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 5})
	buf := buffer.NewBuffer(50, 5)
	gc.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	foundLabel := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == 'C' { foundLabel = true; break }
	}
	if !foundLabel { t.Error("gauge label not found") }
}

func TestGaugeClusterPaintEmpty(t *testing.T) {
	gc := NewGaugeCluster()
	gc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 5})
	buf := buffer.NewBuffer(50, 5)
	gc.Paint(buf)
}

func TestGaugeClusterChildren(t *testing.T) {
	gc := NewGaugeCluster()
	if gc.Children() != nil { t.Error("Children should be nil") }
}

func TestGaugeClusterStyle(t *testing.T) {
	gc := NewGaugeCluster()
	gc.SetStyle(GaugeClusterStyle{Normal: buffer.Style{Fg: buffer.RGB(0,255,0)}, Warning: buffer.Style{Fg: buffer.RGB(255,255,0)}, Critical: buffer.Style{Fg: buffer.RGB(255,0,0)}, Label: buffer.Style{Fg: buffer.RGB(150,150,150)}, Value: buffer.Style{Fg: buffer.RGB(255,255,255)}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	gc.AddGauge("X", 90, 100)
	gc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	gc.Paint(buf)
}

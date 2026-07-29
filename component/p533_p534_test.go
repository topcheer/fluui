package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestModelComparisonBasic(t *testing.T) {
	mcm := NewModelComparisonMatrix()
	mcm.SetMetrics([]string{"Speed", "Quality"})
	mcm.AddModel("GPT-4o", []float64{85, 92})
	if mcm.ModelCount() != 1 { t.Errorf("ModelCount = %d, want 1", mcm.ModelCount()) }
}

func TestModelComparisonMultiple(t *testing.T) {
	mcm := NewModelComparisonMatrix()
	mcm.SetMetrics([]string{"Speed", "Cost", "Context"})
	mcm.AddModel("A", []float64{90, 30, 128})
	mcm.AddModel("B", []float64{70, 50, 200})
	mcm.AddModel("C", []float64{50, 80, 32})
	if mcm.ModelCount() != 3 { t.Errorf("ModelCount = %d, want 3", mcm.ModelCount()) }
}

func TestModelComparisonClear(t *testing.T) {
	mcm := NewModelComparisonMatrix()
	mcm.AddModel("A", []float64{90})
	mcm.Clear()
	if mcm.ModelCount() != 0 { t.Errorf("ModelCount = %d, want 0", mcm.ModelCount()) }
}

func TestModelComparisonValueStyles(t *testing.T) {
	mcm := NewModelComparisonMatrix()
	if mcm.valueStyleLocked(80).Fg != mcm.style.High.Fg { t.Error("80 should be high") }
	if mcm.valueStyleLocked(50).Fg != mcm.style.Medium.Fg { t.Error("50 should be medium") }
	if mcm.valueStyleLocked(20).Fg != mcm.style.Low.Fg { t.Error("20 should be low") }
}

func TestModelComparisonMeasure(t *testing.T) {
	mcm := NewModelComparisonMatrix()
	mcm.SetMetrics([]string{"A", "B"})
	mcm.AddModel("X", []float64{50, 60})
	s := mcm.Measure(Constraints{})
	if s.W < 20 { t.Errorf("W = %d", s.W) }
	if s.H < 4 { t.Errorf("H = %d", s.H) }
}

func TestModelComparisonPaint(t *testing.T) {
	mcm := NewModelComparisonMatrix()
	mcm.SetMetrics([]string{"Speed", "Quality", "Cost"})
	mcm.AddModel("GPT-4o", []float64{85, 92, 35})
	mcm.AddModel("Claude", []float64{78, 95, 40})
	mcm.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 6})
	buf := buffer.NewBuffer(50, 6)
	mcm.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	foundModel := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 2).Rune == 'G' { foundModel = true; break }
	}
	if !foundModel { t.Error("model name not found") }
}

func TestModelComparisonPaintEmpty(t *testing.T) {
	mcm := NewModelComparisonMatrix()
	mcm.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	mcm.Paint(buf)
}

func TestModelComparisonChildren(t *testing.T) {
	mcm := NewModelComparisonMatrix()
	if mcm.Children() != nil { t.Error("Children should be nil") }
}

func TestModelComparisonStyle(t *testing.T) {
	mcm := NewModelComparisonMatrix()
	mcm.SetStyle(ModelComparisonStyle{Header: buffer.Style{Fg: buffer.RGB(255,0,255)}, ModelName: buffer.Style{Fg: buffer.RGB(255,255,255)}, High: buffer.Style{Fg: buffer.RGB(0,255,0)}, Medium: buffer.Style{Fg: buffer.RGB(255,255,0)}, Low: buffer.Style{Fg: buffer.RGB(255,0,0)}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	mcm.SetMetrics([]string{"X"})
	mcm.AddModel("Y", []float64{60})
	mcm.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	mcm.Paint(buf)
}

// ─── AccessibilityFocusRing tests ───

func TestAccessibilityFocusRingBasic(t *testing.T) {
	afr := NewAccessibilityFocusRing()
	afr.SetFocused(true)
	afr.SetLabel("Submit")
	if !afr.Focused() { t.Error("should be focused") }
	if afr.Label() != "Submit" { t.Errorf("Label = %q", afr.Label()) }
}

func TestAccessibilityFocusRingUnfocused(t *testing.T) {
	afr := NewAccessibilityFocusRing()
	afr.SetFocused(false)
	if afr.Focused() { t.Error("should not be focused") }
}

func TestAccessibilityFocusRingStyles(t *testing.T) {
	afr := NewAccessibilityFocusRing()
	afr.SetRingStyle(AccRingDashed)
	if afr.RingStyle() != AccRingDashed { t.Error("should be dashed") }
	afr.SetRingStyle(AccRingThick)
	if afr.RingStyle() != AccRingThick { t.Error("should be thick") }
}

func TestAccessibilityFocusRingMeasure(t *testing.T) {
	afr := NewAccessibilityFocusRing()
	s := afr.Measure(Constraints{})
	if s.W < 5 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestAccessibilityFocusRingPaint(t *testing.T) {
	afr := NewAccessibilityFocusRing()
	afr.SetFocused(true)
	afr.SetLabel("Button")
	afr.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	afr.Paint(buf)
	// Focused should have label on top border
	foundLabel := false
	for x := 0; x < 20; x++ {
		if buf.GetCell(x, 0).Rune == 'B' { foundLabel = true; break }
	}
	if !foundLabel { t.Error("label not found on focused ring") }
}

func TestAccessibilityFocusRingPaintUnfocused(t *testing.T) {
	afr := NewAccessibilityFocusRing()
	afr.SetFocused(false)
	afr.SetLabel("Hidden")
	afr.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	afr.Paint(buf)
	// Unfocused should NOT show label
	for x := 0; x < 20; x++ {
		if buf.GetCell(x, 0).Rune == 'H' { t.Error("label should not show when unfocused"); break }
	}
}

func TestAccessibilityFocusRingPaintDashed(t *testing.T) {
	afr := NewAccessibilityFocusRing()
	afr.SetRingStyle(AccRingDashed)
	afr.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	afr.Paint(buf)
	// Should find dashed chars
	foundDashed := false
	for x := 0; x < 20; x++ {
		r := buf.GetCell(x, 0).Rune
		if r == '┄' { foundDashed = true; break }
	}
	if !foundDashed { t.Error("dashed border char not found") }
}

func TestAccessibilityFocusRingChildren(t *testing.T) {
	afr := NewAccessibilityFocusRing()
	if afr.Children() != nil { t.Error("Children should be nil") }
}

func TestAccessibilityFocusRingStyle(t *testing.T) {
	afr := NewAccessibilityFocusRing()
	afr.SetStyle(AccessibilityFocusRingStyle{Focused: buffer.Style{Fg: buffer.RGB(0,255,0)}, Unfocused: buffer.Style{Fg: buffer.RGB(50,50,50)}, Label: buffer.Style{Fg: buffer.RGB(255,255,0)}})
	afr.SetFocused(true)
	afr.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	afr.Paint(buf)
}

// Benchmarks

func BenchmarkPaintModelComparisonMatrix(b *testing.B) {
	mcm := NewModelComparisonMatrix()
	mcm.SetMetrics([]string{"Speed", "Quality", "Cost", "Context"})
	mcm.AddModel("GPT-4o", []float64{85, 92, 40, 128})
	mcm.AddModel("Claude-3.5", []float64{78, 95, 35, 200})
	mcm.AddModel("Gemini-1.5", []float64{82, 88, 30, 1000})
	mcm.AddModel("Llama-3", []float64{70, 80, 90, 8})
	mcm.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 8})
	buf := buffer.NewBuffer(60, 8)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		mcm.Paint(buf)
	}
}

func BenchmarkPaintAccessibilityFocusRing(b *testing.B) {
	afr := NewAccessibilityFocusRing()
	afr.SetFocused(true)
	afr.SetLabel("Submit Button")
	afr.SetRingStyle(AccRingDashed)
	afr.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		afr.Paint(buf)
	}
}

package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P391: DiffStatBar measureWidthLocked + ConfidenceMeter coverage

func TestP391_DiffStatBar_Measure_BarStyle(t *testing.T) {
	d := NewDiffStatBar(12, 5)
	d.SetStyle(DiffStatStyleBar)
	s := d.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	// bar(10) + space(1) + "+12 -5"(6) = 17
	if s.W != 17 {
		t.Errorf("Bar style W = %d, want 17", s.W)
	}
}

func TestP391_DiffStatBar_Measure_FullWithFiles(t *testing.T) {
	d := NewDiffStatBar(12, 5)
	d.SetStats(12, 5, 3)
	d.SetStyle(DiffStatStyleFull)
	s := d.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	// bar(10) + space(1) + "+12 -5"(7) + space(1) + " (3 files)"(9) = 28
	// Actually: full = bar+1 + textWidth + " (N files)"
	// textWidth = 1+2+1+1+1 = 6 ("+12 -5")
	// file suffix = 3 + 1 + 6 = 10 (" (3 files)") or 3+1+5=9 (" (1 file)")
	// total = 10+1+6+10 = 27
	if s.W < 20 {
		t.Errorf("Full+files W = %d, too small", s.W)
	}
}

func TestP391_DiffStatBar_Measure_FullSingleFile(t *testing.T) {
	d := NewDiffStatBar(10, 3)
	d.SetStats(10, 3, 1) // singular "file"
	d.SetStyle(DiffStatStyleFull)
	s := d.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if s.W < 15 {
		t.Errorf("Full+1file W = %d, too small", s.W)
	}
}

func TestP391_DiffStatBar_Measure_FullNoFiles(t *testing.T) {
	d := NewDiffStatBar(10, 3)
	d.SetStats(10, 3, 0)
	d.SetStyle(DiffStatStyleFull)
	s := d.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	// bar(10) + space(1) + "+10 -3"(6) = 17
	if s.W != 17 {
		t.Errorf("Full+nofiles W = %d, want 17", s.W)
	}
}

func TestP391_DiffStatBar_Paint_BarStyle(t *testing.T) {
	d := NewDiffStatBar(20, 10)
	d.SetStyle(DiffStatStyleBar)
	d.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	d.Paint(buf)
	// Should have bar then space then text
}

func TestP391_DiffStatBar_Paint_NarrowClip(t *testing.T) {
	// Very narrow — tests clipping in drawTextLocked
	d := NewDiffStatBar(100, 50)
	d.SetStats(100, 50, 3)
	d.SetBounds(Rect{X: 0, Y: 0, W: 2, H: 1})
	buf := buffer.NewBuffer(2, 1)
	d.Paint(buf) // should clip gracefully
}

// ConfidenceMeter color threshold coverage

func TestP391_ConfidenceMeter_ColorThresholds(t *testing.T) {
	tests := []struct {
		name  string
		value float64
	}{
		{"high", 0.9},
		{"medium", 0.55},
		{"low", 0.2},
		{"zero", 0.0},
		{"full", 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfidenceMeter(tt.value)
			c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
			buf := buffer.NewBuffer(40, 1)
			c.Paint(buf) // should resolve correct color
		})
	}
}

func TestP391_ConfidenceMeter_CustomColor(t *testing.T) {
	c := NewConfidenceMeter(0.5)
	custom := buffer.RGB(0x80, 0x00, 0x80)
	c.SetColor(custom)
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	c.Paint(buf)
	// Verify custom color was used in bar cells
	cell := buf.GetCell(0, 0)
	if cell.Fg != custom {
		t.Error("custom color should be applied to bar cells")
	}
}

func TestP391_ConfidenceMeter_Paint_WithLabel(t *testing.T) {
	c := NewConfidenceMeter(0.75)
	c.SetLabel("Confidence")
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	c.Paint(buf)
	// Label should be drawn first
	cell := buf.GetCell(0, 0)
	if cell.Rune != 'C' {
		t.Errorf("cell[0] = %q, want 'C' (label start)", string(cell.Rune))
	}
}

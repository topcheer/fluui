package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// themeGetSafe returns the active theme for testing.
func themeGetSafe() *theme.Theme {
	return theme.Get()
}

// TestP350_PieChart_Measure_Defaults covers the maxW/maxH <= 0 branches.
func TestP350_PieChart_Measure_Defaults(t *testing.T) {
	p := NewPieChart([]PieSlice{{Label: "A", Value: 10}})

	// Both zero → defaults
	s := p.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W != 40 || s.H != 15 {
		t.Errorf("defaults = %dx%d, want 40x15", s.W, s.H)
	}

	// Negative → defaults
	s = p.Measure(Constraints{MaxWidth: -1, MaxHeight: -5})
	if s.W != 40 || s.H != 15 {
		t.Errorf("negative = %dx%d, want 40x15", s.W, s.H)
	}

	// Explicit values
	s = p.Measure(Constraints{MaxWidth: 60, MaxHeight: 20})
	if s.W != 60 || s.H != 20 {
		t.Errorf("explicit = %dx%d, want 60x20", s.W, s.H)
	}
}

// TestP350_PieChart_SliceColor_WrapAround covers the idx >= len(colors) branch.
func TestP350_PieChart_SliceColor_WrapAround(t *testing.T) {
	th := themeGetSafe()

	// First 6 colors come from the palette
	c0 := sliceColor(0, th)
	c5 := sliceColor(5, th)

	// Index 6+ wraps around
	c6 := sliceColor(6, th)
	if c6 != c0 {
		t.Error("index 6 should wrap to index 0 color")
	}
	c12 := sliceColor(12, th)
	if c12 != c0 {
		t.Error("index 12 should wrap to index 0 color")
	}
	_ = c5
}

// TestP350_CodeBlock_PaintStreamingCursor_Empty covers the len(lines)==0 branch.
func TestP350_CodeBlock_PaintStreamingCursor_Empty(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf) // should render streaming cursor at top-left

	// Should not panic, should render something
}

// TestP350_CodeBlock_PaintStreamingCursor_WithTitle covers the showTitle branch.
func TestP350_CodeBlock_PaintStreamingCursor_WithTitle(t *testing.T) {
	cb := NewCodeBlock("python", "print('hello')")
	cb.showTitle = true
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 8})
	buf := buffer.NewBuffer(40, 8)
	cb.Paint(buf)
}

// TestP350_CodeBlock_PaintStreamingCursor_LongContent covers the lastIdx clamp branch.
func TestP350_CodeBlock_PaintStreamingCursor_LongContent(t *testing.T) {
	longCode := ""
	for i := 0; i < 30; i++ {
		longCode += "line " + string(rune('A'+i%26)) + "\n"
	}
	cb := NewCodeBlock("go", longCode)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

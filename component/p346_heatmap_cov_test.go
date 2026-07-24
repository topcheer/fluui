package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// TestP346_Heatmap_Paint_WithLabels covers the label row + truncation branch.
func TestP346_Heatmap_Paint_WithLabels(t *testing.T) {
	h := NewHeatmap(3, 5)
	h.SetColLabels([]string{"Monday", "Tue", "W", "Th", "Friday"})
	h.SetCell(0, 0, 10)
	h.SetCell(1, 1, 50)
	h.SetCell(2, 2, 100)
	h.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	h.Paint(buf)

	// Label row at y=0, should have truncated "Monday" → "M"
	cell := buf.GetCell(0, 0)
	if cell.Rune != 'M' {
		t.Errorf("truncated label rune = %q, want 'M'", string(cell.Rune))
	}
}

// TestP346_Heatmap_Paint_ZeroBounds covers the early return for zero bounds.
func TestP346_Heatmap_Paint_ZeroBounds(t *testing.T) {
	h := NewHeatmap(2, 3)
	h.SetCell(0, 0, 50)
	h.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 5)
	h.Paint(buf) // should not panic, should do nothing
}

// TestP346_Heatmap_Paint_NarrowWidth covers the x+1 >= bounds.W break.
func TestP346_Heatmap_Paint_NarrowWidth(t *testing.T) {
	h := NewHeatmap(2, 10)
	for r := 0; r < 2; r++ {
		for c := 0; c < 10; c++ {
			h.SetCell(r, c, 50)
		}
	}
	h.SetBounds(Rect{X: 0, Y: 0, W: 4, H: 5}) // very narrow
	buf := buffer.NewBuffer(4, 5)
	h.Paint(buf) // should clip without panic
}

// TestP346_Heatmap_Paint_TooManyLabels covers labels > cols.
func TestP346_Heatmap_Paint_TooManyLabels(t *testing.T) {
	h := NewHeatmap(1, 2)
	h.SetColLabels([]string{"A", "B", "C", "D"}) // more labels than cols
	h.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	h.Paint(buf) // should not panic
}

// TestP346_Heatmap_Paint_LimitedHeight covers y >= bounds.H break.
func TestP346_Heatmap_Paint_LimitedHeight(t *testing.T) {
	h := NewHeatmap(10, 3)
	for r := 0; r < 10; r++ {
		for c := 0; c < 3; c++ {
			h.SetCell(r, c, 50)
		}
	}
	h.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 2}) // very short
	buf := buffer.NewBuffer(10, 2)
	h.Paint(buf) // should clip rows without panic
}

// TestP346_Heatmap_Paint_ZeroLevel covers the level == 0 (empty) branch.
func TestP346_Heatmap_Paint_ZeroLevel(t *testing.T) {
	h := NewHeatmap(1, 3)
	// All cells default to 0 (empty level)
	h.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	buf := buffer.NewBuffer(10, 3)
	h.Paint(buf)

	// Level 0 cells should use th.Border (non-zero Bg)
	cell := buf.GetCell(0, 0)
	if cell.Rune != ' ' {
		t.Errorf("empty cell rune = %q, want ' '", string(cell.Rune))
	}
}

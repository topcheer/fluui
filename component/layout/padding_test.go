package layout

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestPaddingMeasure(t *testing.T) {
	child := component.NewText("Hello") // 5x1
	padding := NewPadding(1, 2, 1, 2, child) // top,right,bottom,left

	sz := padding.Measure(component.Unbounded())
	// width = 5 + 2 + 2 = 9, height = 1 + 1 + 1 = 3
	if sz.W != 9 {
		t.Errorf("width = %d, want 9 (5 + 2 + 2)", sz.W)
	}
	if sz.H != 3 {
		t.Errorf("height = %d, want 3 (1 + 1 + 1)", sz.H)
	}
}

func TestPaddingSetBounds(t *testing.T) {
	child := component.NewText("Hi")
	padding := NewPadding(1, 1, 1, 1, child)

	padding.SetBounds(component.Rect{X: 10, Y: 10, W: 20, H: 10})

	cb := child.Bounds()
	expected := component.Rect{X: 11, Y: 11, W: 18, H: 8}
	if cb != expected {
		t.Errorf("child bounds = %+v, want %+v", cb, expected)
	}
}

func TestPaddingPaint(t *testing.T) {
	child := component.NewText("AB")
	padding := NewPadding(0, 0, 0, 2, child) // left padding = 2

	padding.SetBounds(component.Rect{X: 0, Y: 0, W: 4, H: 1})

	buf := buffer.NewBuffer(4, 1)
	padding.Paint(buf)

	// Cell(0,0) is padding area — should not have child text
	if c := buf.GetCell(0, 0); c.Rune == 'A' || c.Rune == 'B' {
		t.Errorf("cell(0,0): got %q, want padding (not child text)", c.Rune)
	}
	if c := buf.GetCell(2, 0); c.Rune != 'A' {
		t.Errorf("cell(2,0): got %q, want 'A'", c.Rune)
	}
	if c := buf.GetCell(3, 0); c.Rune != 'B' {
		t.Errorf("cell(3,0): got %q, want 'B'", c.Rune)
	}
}

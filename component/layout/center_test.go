package layout

import (
	"testing"

	"github.com/topcheer/fluui/component"
		"github.com/topcheer/fluui/internal/buffer"
)

func TestCenterMeasure(t *testing.T) {
	child := component.NewText("Hello") // 5x1
	center := NewCenter(child)

	sz := center.Measure(component.Unbounded())
	if sz.W != 5 {
		t.Errorf("width = %d, want 5", sz.W)
	}
	if sz.H != 1 {
		t.Errorf("height = %d, want 1", sz.H)
	}
}

func TestCenterSetBounds(t *testing.T) {
	child := component.NewText("Hi") // 2x1
	center := NewCenter(child)

	// Center in 80x24
	center.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	cb := child.Bounds()
	expectedX := 0 + (80-2)/2 // = 39
	expectedY := 0 + (24-1)/2  // = 11

	if cb.X != expectedX {
		t.Errorf("child X = %d, want %d", cb.X, expectedX)
	}
	if cb.Y != expectedY {
		t.Errorf("child Y = %d, want %d", cb.Y, expectedY)
	}
	if cb.W != 2 {
		t.Errorf("child W = %d, want 2", cb.W)
	}
	if cb.H != 1 {
		t.Errorf("child H = %d, want 1", cb.H)
	}
}

func TestCenterPaint(t *testing.T) {
	child := component.NewText("X")
	center := NewCenter(child)
	center.SetBounds(component.Rect{X: 0, Y: 0, W: 3, H: 3})

	buf := buffer.NewBuffer(3, 3)
	center.Paint(buf)

	// Child 'X' (1x1) centered in 3x3 → at (1,1)
	if c := buf.GetCell(1, 1); c.Rune != 'X' {
		t.Errorf("center cell: got %q, want 'X'", c.Rune)
	}
}

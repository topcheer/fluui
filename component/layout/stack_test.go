package layout

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestStackMeasure(t *testing.T) {
	child1 := component.NewText("Hi")   // 2x1
	child2 := component.NewText("Hello") // 5x1

	stack := NewStack(child1, child2)
	sz := stack.Measure(component.Unbounded())

	if sz.W != 5 {
		t.Errorf("width = %d, want 5 (max child width)", sz.W)
	}
	if sz.H != 1 {
		t.Errorf("height = %d, want 1 (max child height)", sz.H)
	}
}

func TestStackSetBounds(t *testing.T) {
	child1 := component.NewText("A")
	child2 := component.NewText("B")

	stack := NewStack(child1, child2)
	stack.SetBounds(component.Rect{X: 5, Y: 3, W: 10, H: 5})

	// Both children should have the same bounds
	b1 := child1.Bounds()
	b2 := child2.Bounds()
	if b1 != b2 {
		t.Errorf("children have different bounds: %v vs %v", b1, b2)
	}
	if b1.X != 5 || b1.Y != 3 || b1.W != 10 || b1.H != 5 {
		t.Errorf("child bounds = %+v, want {5,3,10,5}", b1)
	}
}

func TestStackPaint(t *testing.T) {
	child1 := component.NewText("AAA")
	child2 := component.NewText("B")

	stack := NewStack(child1, child2)
	stack.SetBounds(component.Rect{X: 0, Y: 0, W: 3, H: 1})

	buf := buffer.NewBuffer(3, 1)
	stack.Paint(buf)

	// child2 paints last, so 'B' overwrites 'A' at position 0
	if c := buf.GetCell(0, 0); c.Rune != 'B' {
		t.Errorf("cell(0,0): got %q, want 'B' (last child on top)", c.Rune)
	}
	// Positions 1-2 should still be 'A' from child1
	if c := buf.GetCell(1, 0); c.Rune != 'A' {
		t.Errorf("cell(1,0): got %q, want 'A'", c.Rune)
	}
	if c := buf.GetCell(2, 0); c.Rune != 'A' {
		t.Errorf("cell(2,0): got %q, want 'A'", c.Rune)
	}
}

func TestStackChildren(t *testing.T) {
	child1 := component.NewText("A")
	child2 := component.NewText("B")
	stack := NewStack(child1, child2)

	children := stack.Children()
	if len(children) != 2 {
		t.Fatalf("len(Children()) = %d, want 2", len(children))
	}
}

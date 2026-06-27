package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// Note: TestBorderMeasure, TestBorderPaint already exist in widgets_test.go.
// Our tests use distinct names.

func TestNewBorder_Defaults(t *testing.T) {
	child := NewText("inside")
	border := NewBorder(child)
	if border.Child == nil {
		t.Fatal("Child: expected non-nil")
	}
	if text, ok := border.Child.(*Text); !ok || text.Content != "inside" {
		t.Errorf("Child: expected Text with 'inside'")
	}
	if border.Title != "" {
		t.Errorf("Title: got %q, want empty", border.Title)
	}
}

func TestBorderMeasurePadding(t *testing.T) {
	child := NewText("Hello") // width=5
	border := NewBorder(child)
	size := border.Measure(Unbounded())
	// child width + 2 (left/right border), child height + 2 (top/bottom border)
	if size.W != 7 {
		t.Errorf("W: got %d, want 7", size.W)
	}
	if size.H != 3 {
		t.Errorf("H: got %d, want 3", size.H)
	}
}

func TestBorderMeasureClampedConstraints(t *testing.T) {
	child := NewText("Hello World") // width=11
	border := NewBorder(child)
	// MaxWidth=5 -> inner MaxWidth=3 -> child clamped to 3 -> border = 3+2=5
	size := border.Measure(Constraints{MaxWidth: 5})
	if size.W != 5 {
		t.Errorf("W: got %d, want 5", size.W)
	}
	if size.H != 3 {
		t.Errorf("H: got %d, want 3", size.H)
	}
}

func TestBorderPaintCorners(t *testing.T) {
	child := NewText("Hi")
	border := NewBorder(child)
	border.SetBounds(Rect{X: 0, Y: 0, W: 6, H: 4})

	buf := buffer.NewBuffer(6, 4)
	border.Paint(buf)

	// Corners
	if tl := buf.GetCell(0, 0); tl.Rune != '\u250c' {
		t.Errorf("top-left: got %q, want \u250c", string(tl.Rune))
	}
	if tr := buf.GetCell(5, 0); tr.Rune != '\u2510' {
		t.Errorf("top-right: got %q, want \u2510", string(tr.Rune))
	}
	if bl := buf.GetCell(0, 3); bl.Rune != '\u2514' {
		t.Errorf("bottom-left: got %q, want \u2514", string(bl.Rune))
	}
	if br := buf.GetCell(5, 3); br.Rune != '\u2518' {
		t.Errorf("bottom-right: got %q, want \u2518", string(br.Rune))
	}

	// Edges (horizontal)
	for i := 1; i < 5; i++ {
		if cell := buf.GetCell(i, 0); cell.Rune != '\u2500' {
			t.Errorf("top edge [%d,0]: got %q, want \u2500", i, string(cell.Rune))
		}
	}

	// Edges (vertical)
	for i := 1; i < 3; i++ {
		if cell := buf.GetCell(0, i); cell.Rune != '\u2502' {
			t.Errorf("left edge [0,%d]: got %q, want \u2502", i, string(cell.Rune))
		}
	}

	// Child content inside
	if inner := buf.GetCell(1, 1); inner.Rune != 'H' {
		t.Errorf("inner [1,1]: got %q, want 'H'", string(inner.Rune))
	}
}

func TestBorderPaintWithCustomTitle(t *testing.T) {
	child := NewText("X")
	border := NewBorder(child)
	border.Title = "MyBox"
	border.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})

	buf := buffer.NewBuffer(10, 3)
	border.Paint(buf)

	// Title should be centered on top edge
	for i, r := range "MyBox" {
		cell := buf.GetCell(2+i, 0)
		if cell.Rune != r {
			t.Errorf("title char [%d,0]: got %q, want %q", 2+i, string(cell.Rune), string(r))
		}
	}
}

func TestBorderPaintTooSmall(t *testing.T) {
	child := NewText("X")
	border := NewBorder(child)

	// Width too small — should not panic
	border.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 4})
	buf := buffer.NewBuffer(1, 4)
	border.Paint(buf)

	// Height too small
	border.SetBounds(Rect{X: 0, Y: 0, W: 4, H: 1})
	buf2 := buffer.NewBuffer(4, 1)
	border.Paint(buf2)

	// Both zero
	border.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf3 := buffer.NewBuffer(1, 1)
	border.Paint(buf3)
}

func TestBorderChildrenReturnsChild(t *testing.T) {
	child := NewText("inside")
	border := NewBorder(child)
	// Border stores child in .Child field; Children() inherits from BaseComponent (nil).
	if border.Child == nil {
		t.Error("Border.Child should be non-nil")
	}
	children := border.Children()
	// BaseComponent.Children() returns nil; Border does not override it.
	if children != nil {
		t.Errorf("Children: got %v, want nil (Border does not override Children)", children)
	}
}

func TestBorderSetCustomStyle(t *testing.T) {
	child := NewText("X")
	border := NewBorder(child)
	border.Style = buffer.Style{
		Fg:    buffer.NamedColor(buffer.NamedGreen),
		Flags: buffer.Underline,
	}
	border.SetBounds(Rect{X: 0, Y: 0, W: 4, H: 3})

	buf := buffer.NewBuffer(4, 3)
	border.Paint(buf)

	tl := buf.GetCell(0, 0)
	if tl.Fg.Type != buffer.ColorNamed || tl.Fg.Val != buffer.NamedGreen {
		t.Errorf("Fg: got %+v, want NamedGreen", tl.Fg)
	}
	if tl.Flags != buffer.Underline {
		t.Errorf("Flags: got %d, want %d", tl.Flags, buffer.Underline)
	}
}

func TestBorderPaintAtOffset(t *testing.T) {
	child := NewText("AB")
	border := NewBorder(child)
	border.SetBounds(Rect{X: 2, Y: 1, W: 6, H: 4})

	buf := buffer.NewBuffer(10, 6)
	border.Paint(buf)

	if tl := buf.GetCell(2, 1); tl.Rune != '\u250c' {
		t.Errorf("top-left at offset: got %q, want \u250c", string(tl.Rune))
	}
	if br := buf.GetCell(7, 4); br.Rune != '\u2518' {
		t.Errorf("bottom-right at offset: got %q, want \u2518", string(br.Rune))
	}
}

func TestBorderMeasureEmptyChild(t *testing.T) {
	child := NewText("") // width=0
	border := NewBorder(child)
	size := border.Measure(Unbounded())
	if size.W != 2 {
		t.Errorf("W: got %d, want 2", size.W)
	}
	if size.H != 3 {
		t.Errorf("H: got %d, want 3", size.H)
	}
}

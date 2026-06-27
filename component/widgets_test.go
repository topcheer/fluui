package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestTextMeasure(t *testing.T) {
	txt := NewText("Hello")
	sz := txt.Measure(Unbounded())
	if sz.W != 5 {
		t.Errorf("width = %d, want 5", sz.W)
	}
	if sz.H != 1 {
		t.Errorf("height = %d, want 1", sz.H)
	}
}

func TestTextMeasureWide(t *testing.T) {
	// Chinese characters are double-width
	txt := NewText("你好")
	sz := txt.Measure(Unbounded())
	if sz.W != 4 {
		t.Errorf("width = %d, want 4 (two double-width runes)", sz.W)
	}
}

func TestTextPaint(t *testing.T) {
	txt := NewText("Hello")
	txt.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})
	buf := buffer.NewBuffer(5, 1)
	txt.Paint(buf)

	for i, want := range "Hello" {
		cell := buf.GetCell(i, 0)
		if cell.Rune != want {
			t.Errorf("cell(%d,0).Rune = %q, want %q", i, cell.Rune, want)
		}
	}
}

func TestBorderMeasure(t *testing.T) {
	child := NewText("0123456789") // 10 wide, 1 tall
	child.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})

	border := NewBorder(child)
	sz := border.Measure(Constraints{
		MinWidth:  0,
		MaxWidth:  100,
		MinHeight: 0,
		MaxHeight: 100,
	})

	if sz.W != 12 {
		t.Errorf("width = %d, want 12", sz.W)
	}
	if sz.H != 3 {
		t.Errorf("height = %d, want 3", sz.H)
	}
}

func TestBorderPaint(t *testing.T) {
	child := NewText("Hi")
	border := NewBorder(child)
	border.SetBounds(Rect{X: 0, Y: 0, W: 6, H: 4})
	buf := buffer.NewBuffer(6, 4)
	border.Paint(buf)

	// Corners
	tests := []struct {
		x, y int
		want rune
		name string
	}{
		{0, 0, '\u250c', "top-left ┌"},
		{5, 0, '\u2510', "top-right ┐"},
		{0, 3, '\u2514', "bottom-left └"},
		{5, 3, '\u2518', "bottom-right ┘"},
	}
	for _, tc := range tests {
		cell := buf.GetCell(tc.x, tc.y)
		if cell.Rune != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, cell.Rune, tc.want)
		}
	}

	// Top edge
	if cell := buf.GetCell(2, 0); cell.Rune != '\u2500' {
		t.Errorf("top edge: got %q, want ─", cell.Rune)
	}

	// Left edge
	if cell := buf.GetCell(0, 1); cell.Rune != '\u2502' {
		t.Errorf("left edge: got %q, want │", cell.Rune)
	}

	// Child content inside border (offset x=1, y=1)
	if cell := buf.GetCell(1, 1); cell.Rune != 'H' {
		t.Errorf("child text[0]: got %q, want 'H'", cell.Rune)
	}
	if cell := buf.GetCell(2, 1); cell.Rune != 'i' {
		t.Errorf("child text[1]: got %q, want 'i'", cell.Rune)
	}
}

func TestBorderTitle(t *testing.T) {
	child := NewText("test")
	border := NewBorder(child)
	border.Title = "My"
	border.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 4})
	buf := buffer.NewBuffer(8, 4)
	border.Paint(buf)

	// Title centered on top: x = (8-2)/2 = 3
	if cell := buf.GetCell(3, 0); cell.Rune != 'M' {
		t.Errorf("title[0]: got %q, want 'M'", cell.Rune)
	}
	if cell := buf.GetCell(4, 0); cell.Rune != 'y' {
		t.Errorf("title[1]: got %q, want 'y'", cell.Rune)
	}
}

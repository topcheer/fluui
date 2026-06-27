package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// Note: TestTextMeasure, TestTextPaint, TestBorderMeasure, TestBorderPaint
// already exist in widgets_test.go. Our tests use distinct names.

func TestNewText_Defaults(t *testing.T) {
	text := NewText("Hello")
	if text.Content != "Hello" {
		t.Errorf("Content: got %q, want %q", text.Content, "Hello")
	}
	// Default style should be zero value
	if text.Style.Fg.Type != buffer.ColorNone {
		t.Errorf("Style.Fg: expected ColorNone, got %+v", text.Style.Fg)
	}
}

func TestTextMeasureWidth(t *testing.T) {
	text := NewText("Hello World")
	size := text.Measure(Unbounded())
	expectedW := buffer.StringWidth("Hello World")
	if size.W != expectedW {
		t.Errorf("W: got %d, want %d", size.W, expectedW)
	}
	if size.H != 1 {
		t.Errorf("H: got %d, want 1", size.H)
	}
}

func TestTextMeasureClampedWidth(t *testing.T) {
	text := NewText("Hello World") // width = 11
	size := text.Measure(Constraints{MaxWidth: 5})
	if size.W != 5 {
		t.Errorf("clamped W: got %d, want 5", size.W)
	}
	if size.H != 1 {
		t.Errorf("H: got %d, want 1", size.H)
	}
}

func TestTextPaintStyled(t *testing.T) {
	text := NewText("X")
	text.Style = buffer.Style{
		Fg:    buffer.NamedColor(buffer.NamedRed),
		Flags: buffer.Bold,
	}
	text.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 1})

	buf := buffer.NewBuffer(1, 1)
	text.Paint(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != 'X' {
		t.Errorf("rune: got %q, want 'X'", string(cell.Rune))
	}
	if cell.Fg.Type != buffer.ColorNamed || cell.Fg.Val != buffer.NamedRed {
		t.Errorf("Fg: got %+v, want NamedRed", cell.Fg)
	}
	if cell.Flags != buffer.Bold {
		t.Errorf("Flags: got %d, want %d", cell.Flags, buffer.Bold)
	}
}

func TestTextPaintAtOffset(t *testing.T) {
	text := NewText("AB")
	text.SetBounds(Rect{X: 5, Y: 3, W: 2, H: 1})

	buf := buffer.NewBuffer(10, 5)
	text.Paint(buf)

	if cell := buf.GetCell(5, 3); cell.Rune != 'A' {
		t.Errorf("cell[5,3]: got %q, want 'A'", string(cell.Rune))
	}
	if cell := buf.GetCell(6, 3); cell.Rune != 'B' {
		t.Errorf("cell[6,3]: got %q, want 'B'", string(cell.Rune))
	}
}

func TestTextEmptyContentMeasure(t *testing.T) {
	text := NewText("")
	size := text.Measure(Unbounded())
	if size.W != 0 {
		t.Errorf("W: got %d, want 0", size.W)
	}
	if size.H != 1 {
		t.Errorf("H: got %d, want 1", size.H)
	}

	// Paint should not panic
	text.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})
	buf := buffer.NewBuffer(5, 1)
	text.Paint(buf)
}

func TestTextChildrenNil(t *testing.T) {
	text := NewText("Hello")
	children := text.Children()
	if children != nil {
		t.Errorf("Children: got %v, want nil (leaf component)", children)
	}
}

func TestTextMeasureUnicodeWidth(t *testing.T) {
	text := NewText("你好")
	size := text.Measure(Unbounded())
	expectedW := buffer.StringWidth("你好") // each CJK char is width 2
	if size.W != expectedW {
		t.Errorf("W: got %d, want %d", size.W, expectedW)
	}
}

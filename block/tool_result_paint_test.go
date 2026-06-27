package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestToolResultPaintBorder(t *testing.T) {
	b := NewToolResultBlock("trp-1")
	b.AppendDelta("hello")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 4})

	buf := buffer.NewBuffer(30, 4)
	b.Paint(buf)

	// Top-left corner
	if buf.GetCell(0, 0).Rune != '╭' {
		t.Errorf("top-left = %q, want '╭'", buf.GetCell(0, 0).Rune)
	}

	// Bottom-left corner
	if buf.GetCell(0, 3).Rune != '╰' {
		t.Errorf("bottom-left = %q, want '╰'", buf.GetCell(0, 3).Rune)
	}

	// Top border should contain "Result" label
	foundResult := false
	for x := 0; x < 30; x++ {
		rest := readRunes(buf, x, 0, 6)
		if rest == "Result" {
			foundResult = true
			break
		}
	}
	if !foundResult {
		t.Error("expected 'Result' label in top border")
	}
}

func TestToolResultPaintContent(t *testing.T) {
	b := NewToolResultBlock("trp-2")
	b.AppendDelta("hello world\nsecond line")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 4})

	buf := buffer.NewBuffer(30, 4)
	b.Paint(buf)

	// Row 1 should start with │ (content border)
	if buf.GetCell(0, 1).Rune != '│' {
		t.Errorf("content row 1 border = %q, want '│'", buf.GetCell(0, 1).Rune)
	}

	// Row 1 should contain "hello" after the │ prefix
	rest := readRunes(buf, 2, 1, 5)
	if rest != "hello" {
		t.Errorf("content row 1 = %q, want 'hello'", rest)
	}

	// Row 2 should contain "second"
	rest = readRunes(buf, 2, 2, 6)
	if rest != "second" {
		t.Errorf("content row 2 = %q, want 'second'", rest)
	}
}

func TestToolResultPaintCollapsed(t *testing.T) {
	b := NewToolResultBlock("trp-3")
	// Add 8 lines so it auto-collapses on Complete
	for i := 0; i < 8; i++ {
		b.AppendDelta("line " + string(rune('A'+i)) + "\n")
	}
	b.Complete()

	if !b.Collapsed() {
		t.Fatal("precondition: should be collapsed")
	}

	// Measure gives maxPreview+2 = 7
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 7})

	buf := buffer.NewBuffer(30, 7)
	b.Paint(buf)

	// Rows 1-5 should have content (maxPreview=5 lines)
	// Row 6 is bottom border
	// Verify that "line A" through "line E" are present (first 5 lines)
	expected := []string{"line A", "line B", "line C", "line D", "line E"}
	for i, want := range expected {
		rest := readRunes(buf, 2, 1+i, len(want))
		if rest != want {
			t.Errorf("collapsed row %d = %q, want %q", i, rest, want)
		}
	}

	// Bottom border at row 6
	if buf.GetCell(0, 6).Rune != '╰' {
		t.Errorf("bottom border row 6 = %q, want '╰'", buf.GetCell(0, 6).Rune)
	}
}

func TestToolResultPaintExpanded(t *testing.T) {
	b := NewToolResultBlock("trp-4")
	// Add 4 lines — no auto-collapse
	b.AppendDelta("alpha\nbeta\ngamma\ndelta")
	b.Complete()

	if b.Collapsed() {
		t.Fatal("precondition: should NOT be collapsed with only 4 lines")
	}

	// Measure gives 4+2 = 6
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 6})

	buf := buffer.NewBuffer(30, 6)
	b.Paint(buf)

	// All 4 content lines should be visible
	expected := []string{"alpha", "beta", "gamma", "delta"}
	for i, want := range expected {
		rest := readRunes(buf, 2, 1+i, len(want))
		if rest != want {
			t.Errorf("expanded row %d = %q, want %q", i, rest, want)
		}
	}

	// Bottom border at row 5
	if buf.GetCell(0, 5).Rune != '╰' {
		t.Errorf("bottom border row 5 = %q, want '╰'", buf.GetCell(0, 5).Rune)
	}
}

func TestToolResultPaintBorderColor(t *testing.T) {
	b := NewToolResultBlock("trp-5")
	b.AppendDelta("test")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 4})

	buf := buffer.NewBuffer(30, 4)
	b.Paint(buf)

	// Border color should be RGB(0x62, 0x72, 0xA4)
	cell := buf.GetCell(0, 0) // ╭ corner
	want := buffer.RGB(0x62, 0x72, 0xA4)
	if !cell.Fg.Equal(want) {
		t.Errorf("border Fg = %s, want %s", cell.Fg, want)
	}
}

func TestToolResultPaintContentColor(t *testing.T) {
	b := NewToolResultBlock("trp-6")
	b.AppendDelta("colored text")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 4})

	buf := buffer.NewBuffer(30, 4)
	b.Paint(buf)

	// Content color should be RGB(0xBD, 0x93, 0xF9) purple
	// Row 1 (first content line), col 1 (after │ ) has a space prefix, col 2 has content
	cell := buf.GetCell(2, 1) // first content char after "│ "
	want := buffer.RGB(0xBD, 0x93, 0xF9)
	if !cell.Fg.Equal(want) {
		t.Errorf("content Fg = %s, want %s", cell.Fg, cell.Fg)
	}
}

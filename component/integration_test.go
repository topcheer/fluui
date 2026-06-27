package component_test

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/component/layout"
	"github.com/topcheer/fluui/internal/buffer"
)

// --- helpers ---

// makeText creates a Text with a fixed width for predictable layout.
func makeText(s string) *component.Text { return component.NewText(s) }

// measureChild is a shortcut to measure a child with unbounded constraints.
func measureChild(c component.Component) component.Size {
	return c.Measure(component.Unbounded())
}

// --- Integration tests ---

// TestIntegrationBorderWithText: Border wraps Text.
func TestIntegrationBorderWithText(t *testing.T) {
	text := makeText("Hello") // 5x1
	border := component.NewBorder(text)

	// Measure: 5+2=7 wide, 1+2=3 tall
	sz := border.Measure(component.Bounded(100, 100))
	if sz.W != 7 {
		t.Errorf("border width = %d, want 7", sz.W)
	}
	if sz.H != 3 {
		t.Errorf("border height = %d, want 3", sz.H)
	}

	// Paint into a 7x3 buffer
	border.SetBounds(component.Rect{X: 0, Y: 0, W: 7, H: 3})
	buf := buffer.NewBuffer(7, 3)
	border.Paint(buf)

	// Verify corners
	if c := buf.GetCell(0, 0); c.Rune != '\u250c' {
		t.Errorf("top-left: got %q, want \u250c", c.Rune)
	}
	if c := buf.GetCell(6, 0); c.Rune != '\u2510' {
		t.Errorf("top-right: got %q, want \u2510", c.Rune)
	}
	if c := buf.GetCell(0, 2); c.Rune != '\u2514' {
		t.Errorf("bottom-left: got %q, want \u2514", c.Rune)
	}
	if c := buf.GetCell(6, 2); c.Rune != '\u2518' {
		t.Errorf("bottom-right: got %q, want \u2518", c.Rune)
	}

	// Verify text content at (1,1) inside the border
	for i, r := range "Hello" {
		c := buf.GetCell(1+i, 1)
		if c.Rune != r {
			t.Errorf("text[%d]: got %q, want %q", i, c.Rune, r)
		}
	}
}

// TestIntegrationFlexRowWithBorders: Flex(Row) contains 3 Border(Text) children.
func TestIntegrationFlexRowWithBorders(t *testing.T) {
	b1 := component.NewBorder(makeText("AB"))
	b2 := component.NewBorder(makeText("CD"))
	b3 := component.NewBorder(makeText("EF"))

	flex := layout.NewFlexGap(layout.FlexRow, 1)
	flex.AddChild(b1)
	flex.AddChild(b2)
	flex.AddChild(b3)

	// Measure: each border is 4x3, total = 4+1+4+1+4 = 14, height = 3
	sz := flex.Measure(component.Bounded(100, 100))
	if sz.W != 14 {
		t.Errorf("flex width = %d, want 14", sz.W)
	}
	if sz.H != 3 {
		t.Errorf("flex height = %d, want 3", sz.H)
	}

	flex.SetBounds(component.Rect{X: 0, Y: 0, W: 14, H: 3})
	buf := buffer.NewBuffer(14, 3)
	flex.Paint(buf)

	// Verify each border's top-left corner
	checks := []struct {
		x    int
		name string
	}{
		{0, "b1"},
		{5, "b2"},
		{10, "b3"},
	}
	for _, tc := range checks {
		if c := buf.GetCell(tc.x, 0); c.Rune != '\u250c' {
			t.Errorf("%s top-left @ x=%d: got %q, want \u250c", tc.name, tc.x, c.Rune)
		}
	}

	// Verify child text inside each border
	textChecks := []struct {
		x   int
		msg string
	}{
		{1, "AB"},
		{6, "CD"},
		{11, "EF"},
	}
	for _, tc := range textChecks {
		for i, r := range tc.msg {
			c := buf.GetCell(tc.x+i, 1)
			if c.Rune != r {
				t.Errorf("text @(%d,1)[%d]: got %q, want %q", tc.x, i, c.Rune, r)
			}
		}
	}
}

// TestIntegrationScrollViewWithLongText: ScrollView wraps a tall content.
func TestIntegrationScrollViewWithLongText(t *testing.T) {
	col := layout.NewFlex(layout.FlexColumn)
	for _, s := range []string{"L0", "L1", "L2", "L3", "L4"} {
		col.AddChild(makeText(s))
	}

	sv := component.NewScrollView(col)
	sz := sv.Measure(component.Bounded(3, 2))
	if sz.H != 2 {
		t.Errorf("viewport height = %d, want 2", sz.H)
	}
	if sv.MaxOffset() != 3 {
		t.Errorf("maxOffset = %d, want 3 (content=5 - viewport=2)", sv.MaxOffset())
	}

	// Use 3-wide viewport: cols 0-1 for content, col 2 for scrollbar
	sv.SetBounds(component.Rect{X: 0, Y: 0, W: 3, H: 2})

	// At offset 0: should see L0, L1
	buf0 := buffer.NewBuffer(3, 2)
	sv.Paint(buf0)
	if c := buf0.GetCell(0, 0); c.Rune != 'L' {
		t.Errorf("offset=0 row0: got %q, want 'L'", c.Rune)
	}
	if c := buf0.GetCell(0, 1); c.Rune != 'L' {
		t.Errorf("offset=0 row1: got %q, want 'L'", c.Rune)
	}

	// Scroll to offset 2: should see L2, L3
	sv.ScrollTo(2)
	buf2 := buffer.NewBuffer(3, 2)
	sv.Paint(buf2)
	if c := buf2.GetCell(0, 0); c.Rune != 'L' {
		t.Errorf("offset=2 row0: got %q, want 'L'", c.Rune)
	}
	if c := buf2.GetCell(1, 0); c.Rune != '2' {
		t.Errorf("offset=2 row0 col1: got %q, want '2'", c.Rune)
	}
	if c := buf2.GetCell(1, 1); c.Rune != '3' {
		t.Errorf("offset=2 row1 col1: got %q, want '3'", c.Rune)
	}
}

// TestIntegrationNestedFlex: Flex(Column) containing Flex(Row) children.
func TestIntegrationNestedFlex(t *testing.T) {
	row1 := layout.NewFlexGap(layout.FlexRow, 1)
	row1.AddChild(makeText("A"))
	row1.AddChild(makeText("B"))

	row2 := layout.NewFlex(layout.FlexRow)
	row2.AddChild(makeText("CD"))

	col := layout.NewFlex(layout.FlexColumn)
	col.AddChild(row1)
	col.AddChild(row2)

	sz := col.Measure(component.Bounded(100, 100))
	if sz.W != 3 {
		t.Errorf("column width = %d, want 3 (max of rows)", sz.W)
	}
	if sz.H != 2 {
		t.Errorf("column height = %d, want 2", sz.H)
	}

	col.SetBounds(component.Rect{X: 0, Y: 0, W: 3, H: 2})
	buf := buffer.NewBuffer(3, 2)
	col.Paint(buf)

	if c := buf.GetCell(0, 0); c.Rune != 'A' {
		t.Errorf("(0,0): got %q, want 'A'", c.Rune)
	}
	if c := buf.GetCell(2, 0); c.Rune != 'B' {
		t.Errorf("(2,0): got %q, want 'B'", c.Rune)
	}
	if c := buf.GetCell(0, 1); c.Rune != 'C' {
		t.Errorf("(0,1): got %q, want 'C'", c.Rune)
	}
	if c := buf.GetCell(1, 1); c.Rune != 'D' {
		t.Errorf("(1,1): got %q, want 'D'", c.Rune)
	}
}

// TestIntegrationBorderWithFlex: Border wraps a Flex(Column).
func TestIntegrationBorderWithFlex(t *testing.T) {
	col := layout.NewFlex(layout.FlexColumn)
	col.AddChild(makeText("Hi"))
	col.AddChild(makeText("Yo"))

	border := component.NewBorder(col)
	sz := border.Measure(component.Bounded(100, 100))
	if sz.W != 4 {
		t.Errorf("border width = %d, want 4", sz.W)
	}
	if sz.H != 4 {
		t.Errorf("border height = %d, want 4", sz.H)
	}

	border.SetBounds(component.Rect{X: 0, Y: 0, W: 4, H: 4})
	buf := buffer.NewBuffer(4, 4)
	border.Paint(buf)

	if c := buf.GetCell(0, 0); c.Rune != '\u250c' {
		t.Errorf("top-left: got %q, want \u250c", c.Rune)
	}
	if c := buf.GetCell(3, 3); c.Rune != '\u2518' {
		t.Errorf("bottom-right: got %q, want \u2518", c.Rune)
	}
	if c := buf.GetCell(1, 1); c.Rune != 'H' {
		t.Errorf("(1,1): got %q, want 'H'", c.Rune)
	}
	if c := buf.GetCell(2, 1); c.Rune != 'i' {
		t.Errorf("(2,1): got %q, want 'i'", c.Rune)
	}
	if c := buf.GetCell(1, 2); c.Rune != 'Y' {
		t.Errorf("(1,2): got %q, want 'Y'", c.Rune)
	}
	if c := buf.GetCell(2, 2); c.Rune != 'o' {
		t.Errorf("(2,2): got %q, want 'o'", c.Rune)
	}
}

// TestIntegrationBorderWithFlexRowMultiple: Border wraps a Flex(Row) with multiple children.
func TestIntegrationBorderWithFlexRowMultiple(t *testing.T) {
	row := layout.NewFlexGap(layout.FlexRow, 1)
	row.AddChild(makeText("AB"))
	row.AddChild(makeText("CD"))

	border := component.NewBorder(row)
	sz := border.Measure(component.Bounded(100, 100))
	if sz.W != 7 {
		t.Errorf("border width = %d, want 7", sz.W)
	}
	if sz.H != 3 {
		t.Errorf("border height = %d, want 3", sz.H)
	}

	border.SetBounds(component.Rect{X: 0, Y: 0, W: 7, H: 3})
	buf := buffer.NewBuffer(7, 3)
	border.Paint(buf)

	if c := buf.GetCell(1, 1); c.Rune != 'A' {
		t.Errorf("(1,1): got %q, want 'A'", c.Rune)
	}
	if c := buf.GetCell(2, 1); c.Rune != 'B' {
		t.Errorf("(2,1): got %q, want 'B'", c.Rune)
	}
	if c := buf.GetCell(4, 1); c.Rune != 'C' {
		t.Errorf("(4,1): got %q, want 'C'", c.Rune)
	}
	if c := buf.GetCell(5, 1); c.Rune != 'D' {
		t.Errorf("(5,1): got %q, want 'D'", c.Rune)
	}
}

// TestIntegrationDeepNesting: Border(Border(Text)) — double border.
func TestIntegrationDeepNesting(t *testing.T) {
	inner := component.NewBorder(makeText("X"))
	outer := component.NewBorder(inner)

	sz := outer.Measure(component.Bounded(100, 100))
	if sz.W != 5 {
		t.Errorf("outer width = %d, want 5", sz.W)
	}
	if sz.H != 5 {
		t.Errorf("outer height = %d, want 5", sz.H)
	}

	outer.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 5})
	buf := buffer.NewBuffer(5, 5)
	outer.Paint(buf)

	if c := buf.GetCell(0, 0); c.Rune != '\u250c' {
		t.Errorf("outer top-left: got %q, want \u250c", c.Rune)
	}
	if c := buf.GetCell(1, 1); c.Rune != '\u250c' {
		t.Errorf("inner top-left: got %q, want \u250c", c.Rune)
	}
	if c := buf.GetCell(2, 2); c.Rune != 'X' {
		t.Errorf("center text: got %q, want 'X'", c.Rune)
	}
}

// TestIntegrationWideCharBorder: Border wrapping wide (CJK) text.
func TestIntegrationWideCharBorder(t *testing.T) {
	text := makeText("\u4f60\u597d") // 你好, 4 wide cells
	border := component.NewBorder(text)

	sz := border.Measure(component.Bounded(100, 100))
	if sz.W != 6 {
		t.Errorf("border width = %d, want 6 (4+2)", sz.W)
	}
	if sz.H != 3 {
		t.Errorf("border height = %d, want 3", sz.H)
	}

	border.SetBounds(component.Rect{X: 0, Y: 0, W: 6, H: 3})
	buf := buffer.NewBuffer(6, 3)
	border.Paint(buf)

	if c := buf.GetCell(1, 1); c.Rune != '\u4f60' {
		t.Errorf("(1,1): got %q, want '\u4f60'", c.Rune)
	}
	if c := buf.GetCell(3, 1); c.Rune != '\u597d' {
		t.Errorf("(3,1): got %q, want '\u597d'", c.Rune)
	}
}

// TestIntegrationSuite runs all integration tests and produces a summary.
func TestIntegrationSuite(t *testing.T) {
	tests := []string{
		"TestIntegrationBorderWithText",
		"TestIntegrationFlexRowWithBorders",
		"TestIntegrationScrollViewWithLongText",
		"TestIntegrationNestedFlex",
		"TestIntegrationBorderWithFlex",
		"TestIntegrationBorderWithFlexRowMultiple",
		"TestIntegrationDeepNesting",
		"TestIntegrationWideCharBorder",
	}
	t.Logf("Integration suite: %d scenarios\n%s", len(tests), strings.Join(tests, "\n"))
}

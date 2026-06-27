package buffer

import (
	"testing"
)

// ============================================================
// Zero-size and negative-size buffer tests
// ============================================================

func TestBufferZeroSize(t *testing.T) {
	b := NewBuffer(0, 0)
	if b.Width != 0 || b.Height != 0 {
		t.Errorf("expected 0x0, got %dx%d", b.Width, b.Height)
	}
	if len(b.Cells) != 0 {
		t.Errorf("expected 0 cells, got %d", len(b.Cells))
	}

	// Operations should not panic on 0x0 buffer.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("0x0 buffer panicked: %v", r)
		}
	}()

	b.SetCell(0, 0, Cell{Rune: 'x', Width: 1})
	b.SetCell(-1, -1, Cell{Rune: 'x', Width: 1})

	cell := b.GetCell(0, 0)
	if cell.Rune != ' ' {
		t.Errorf("GetCell on 0x0 buffer should return BlankCell, got %c", cell.Rune)
	}

	b.Fill(Cell{Rune: '#', Width: 1})
	b.FillRect(Rect{X: 0, Y: 0, W: 5, H: 5}, Cell{Rune: '#', Width: 1})

	endX := b.DrawText(0, 0, "hello", DefaultStyle)
	if endX != 0 {
		t.Errorf("DrawText on 0x0 buffer should return start x, got %d", endX)
	}
}

func TestBufferNegativeSize(t *testing.T) {
	// make(-1*-1) would panic, but NewBuffer should handle gracefully.
	// Actually NewBuffer doesn't guard against negative — it calls make(w*h)
	// which with two negatives gives a positive. Let's test defensively.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("negative size buffer panicked: %v", r)
		}
	}()

	// w=-2, h=-3 → w*h = 6. This won't panic in make but is semantically wrong.
	// We just ensure it doesn't crash.
	b := NewBuffer(-2, -3)
	// idx checks bounds so all operations should be no-ops.
	b.SetCell(0, 0, Cell{Rune: 'x', Width: 1})
	cell := b.GetCell(0, 0)
	_ = cell // just ensure no panic
}

// ============================================================
// SetCell / GetCell out-of-range tests
// ============================================================

func TestSetCellOutOfRange(t *testing.T) {
	b := NewBuffer(3, 3)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetCell out of range panicked: %v", r)
		}
	}()

	// Beyond right edge
	b.SetCell(3, 0, Cell{Rune: 'X', Width: 1})
	// Beyond bottom edge
	b.SetCell(0, 3, Cell{Rune: 'Y', Width: 1})
	// Negative coordinates
	b.SetCell(-1, 0, Cell{Rune: 'Z', Width: 1})
	b.SetCell(0, -1, Cell{Rune: 'W', Width: 1})
	// Way out of range
	b.SetCell(100, 100, Cell{Rune: '!', Width: 1})

	// Verify nothing changed
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			cell := b.GetCell(x, y)
			if cell.Rune != ' ' {
				t.Errorf("cell(%d,%d) was modified: %c", x, y, cell.Rune)
			}
		}
	}
}

func TestGetCellOutOfRange(t *testing.T) {
	b := NewBuffer(2, 2)
	b.SetCell(0, 0, Cell{Rune: 'A', Width: 1})

	tests := []struct {
		name string
		x, y int
	}{
		{"right edge+1", 2, 0},
		{"bottom edge+1", 0, 2},
		{"negative x", -1, 0},
		{"negative y", 0, -1},
		{"both negative", -1, -1},
		{"far away", 1000, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cell := b.GetCell(tt.x, tt.y)
			if cell.Rune != BlankCell.Rune {
				t.Errorf("GetCell(%d,%d): got %c, want blank", tt.x, tt.y, cell.Rune)
			}
			if cell.Width != BlankCell.Width {
				t.Errorf("GetCell(%d,%d) width: got %d, want %d", tt.x, tt.y, cell.Width, BlankCell.Width)
			}
		})
	}
}

// ============================================================
// DrawText overflow tests
// ============================================================

func TestDrawTextOverflow(t *testing.T) {
	b := NewBuffer(5, 1)
	style := DefaultStyle.WithFg(RGB(255, 0, 0))

	endX := b.DrawText(0, 0, "Hello World", style)

	// Buffer is only 5 wide; text should be truncated.
	if endX != 5 {
		t.Errorf("DrawText returned %d, expected 5 (buffer width)", endX)
	}

	// Check that first 5 chars were written
	expected := "Hello"
	for i, c := range expected {
		cell := b.GetCell(i, 0)
		if cell.Rune != c {
			t.Errorf("cell %d: got %c, want %c", i, cell.Rune, c)
		}
	}
}

func TestDrawTextOverflowStartAtEdge(t *testing.T) {
	b := NewBuffer(3, 1)

	// Start drawing at x=2, text "AB" needs 2 cells but only 1 fits
	endX := b.DrawText(2, 0, "AB", DefaultStyle)
	if endX != 3 {
		t.Errorf("expected endX=3, got %d", endX)
	}
	// Only 'A' should be written
	cell0 := b.GetCell(2, 0)
	if cell0.Rune != 'A' {
		t.Errorf("cell(2,0): got %c, want A", cell0.Rune)
	}
}

func TestDrawTextOverflowStartBeyondBuffer(t *testing.T) {
	b := NewBuffer(5, 1)

	// Start beyond buffer width
	endX := b.DrawText(10, 0, "hello", DefaultStyle)
	if endX != 10 {
		t.Errorf("expected endX=10, got %d", endX)
	}
}

func TestDrawTextMultiWidth(t *testing.T) {
	b := NewBuffer(5, 1)
	endX := b.DrawText(0, 0, "你好", DefaultStyle)

	// 你 = width 2, 好 = width 2, total = 4
	if endX != 4 {
		t.Errorf("expected endX=4, got %d", endX)
	}

	// 你 at cell (0,0) with padding cell at (1,0)
	c0 := b.GetCell(0, 0)
	if c0.Rune != '你' || c0.Width != 2 {
		t.Errorf("cell(0,0): got rune=%c width=%d, want 你 width=2", c0.Rune, c0.Width)
	}
	c1 := b.GetCell(1, 0)
	if c1.Rune != 0 || c1.Width != 0 {
		t.Errorf("padding cell(1,0): got rune=%c width=%d, want rune=0 width=0", c1.Rune, c1.Width)
	}

	// 好 at cell (2,0) with padding cell at (3,0)
	c2 := b.GetCell(2, 0)
	if c2.Rune != '好' || c2.Width != 2 {
		t.Errorf("cell(2,0): got rune=%c width=%d, want 好 width=2", c2.Rune, c2.Width)
	}
}

func TestDrawTextMultiWidthAtEdge(t *testing.T) {
	// Buffer width 3, draw CJK char at x=2 — only 1 cell left, can't fit width-2 char
	b := NewBuffer(3, 1)
	endX := b.DrawText(2, 0, "你", DefaultStyle)

	// 你 is width 2, starts at x=2 but buffer only has 1 cell left
	// DrawText checks x < width (2 < 3 = true), writes the char
	// Then padding at x+1=3 but 3 >= width so no padding
	if endX != 4 {
		t.Errorf("expected endX=4, got %d", endX)
	}

	// The CJK char was written at position 2
	cell := b.GetCell(2, 0)
	if cell.Rune != '你' {
		t.Errorf("expected 你 at (2,0), got %c", cell.Rune)
	}
}

func TestDrawTextClampedOverflow(t *testing.T) {
	b := NewBuffer(5, 1)
	endX := b.DrawTextClamped(0, 0, "Hello World", DefaultStyle)

	// "Hello" = 5 chars, fits exactly. " World" doesn't fit.
	if endX != 5 {
		t.Errorf("expected endX=5, got %d", endX)
	}
	for i, c := range "Hello" {
		cell := b.GetCell(i, 0)
		if cell.Rune != c {
			t.Errorf("cell %d: got %c, want %c", i, cell.Rune, c)
		}
	}
}

func TestDrawTextClampedCJK(t *testing.T) {
	b := NewBuffer(5, 1)
	// "你好世" = 6 cells, buffer is 5 → only "你好" fits (4 cells)
	endX := b.DrawTextClamped(0, 0, "你好世", DefaultStyle)
	if endX != 4 {
		t.Errorf("expected endX=4, got %d", endX)
	}
}

// ============================================================
// Fill / FillRect tests
// ============================================================

func TestBufferFillAll(t *testing.T) {
	b := NewBuffer(3, 2)
	fillCell := Cell{Rune: '#', Width: 1, Fg: RGB(255, 0, 0)}
	b.Fill(fillCell)

	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			cell := b.GetCell(x, y)
			if cell.Rune != '#' {
				t.Errorf("cell(%d,%d): got %c, want #", x, y, cell.Rune)
			}
			if !cell.Fg.Equal(RGB(255, 0, 0)) {
				t.Errorf("cell(%d,%d) fg not set", x, y)
			}
		}
	}
}

func TestBufferFillRect(t *testing.T) {
	b := NewBuffer(5, 5)
	b.FillRect(Rect{X: 1, Y: 1, W: 3, H: 2}, Cell{Rune: '#', Width: 1})

	// Check inside rect
	for y := 1; y < 3; y++ {
		for x := 1; x < 4; x++ {
			cell := b.GetCell(x, y)
			if cell.Rune != '#' {
				t.Errorf("rect cell(%d,%d): got %c, want #", x, y, cell.Rune)
			}
		}
	}

	// Check outside rect is still blank
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			if x >= 1 && x < 4 && y >= 1 && y < 3 {
				continue // inside rect, already checked
			}
			cell := b.GetCell(x, y)
			if cell.Rune != ' ' {
				t.Errorf("outside rect(%d,%d): got %c, want space", x, y, cell.Rune)
			}
		}
	}
}

func TestBufferFillRectOverflow(t *testing.T) {
	b := NewBuffer(3, 3)
	// Rect extends beyond buffer — FillRect should clamp internally.
	b.FillRect(Rect{X: 2, Y: 2, W: 10, H: 10}, Cell{Rune: 'X', Width: 1})

	// Only (2,2) should be filled
	cell := b.GetCell(2, 2)
	if cell.Rune != 'X' {
		t.Errorf("cell(2,2): got %c, want X", cell.Rune)
	}

	// Everything else should be blank
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			if x == 2 && y == 2 {
				continue
			}
			cell := b.GetCell(x, y)
			if cell.Rune != ' ' {
				t.Errorf("cell(%d,%d): got %c, want space", x, y, cell.Rune)
			}
		}
	}
}

// ============================================================
// NewCell wide char tests
// ============================================================

func TestNewCellWideChar(t *testing.T) {
	// CJK character should get Width=2
	style := DefaultStyle.WithFg(RGB(0, 255, 0))
	cell := NewCell('你', style)

	if cell.Rune != '你' {
		t.Errorf("rune: got %c, want 你", cell.Rune)
	}
	if cell.Width != 2 {
		t.Errorf("width: got %d, want 2", cell.Width)
	}
	if !cell.Fg.Equal(RGB(0, 255, 0)) {
		t.Error("fg not set correctly")
	}
}

func TestNewCellASCII(t *testing.T) {
	cell := NewCell('a', DefaultStyle)
	if cell.Width != 1 {
		t.Errorf("ASCII width: got %d, want 1", cell.Width)
	}
}

func TestNewCellZeroWidth(t *testing.T) {
	cell := NewCell('\u0300', DefaultStyle) // combining mark
	if cell.Width != 0 {
		t.Errorf("combining width: got %d, want 0", cell.Width)
	}
}

func TestNewCellEmoji(t *testing.T) {
	cell := NewCell('😀', DefaultStyle)
	if cell.Width != 2 {
		t.Errorf("emoji width: got %d, want 2", cell.Width)
	}
}

// ============================================================
// SetRow tests
// ============================================================

func TestSetRowBasic(t *testing.T) {
	b := NewBuffer(5, 1)
	cells := []Cell{
		{Rune: 'H', Width: 1},
		{Rune: 'i', Width: 1},
		{Rune: '!', Width: 1},
	}
	b.SetRow(0, cells, 1) // start at x=1

	if b.GetCell(0, 0).Rune != ' ' {
		t.Error("cell(0,0) should be blank")
	}
	if b.GetCell(1, 0).Rune != 'H' {
		t.Errorf("cell(1,0): got %c, want H", b.GetCell(1, 0).Rune)
	}
	if b.GetCell(2, 0).Rune != 'i' {
		t.Errorf("cell(2,0): got %c, want i", b.GetCell(2, 0).Rune)
	}
}

func TestSetRowOverflow(t *testing.T) {
	b := NewBuffer(3, 1)
	cells := []Cell{
		{Rune: 'A', Width: 1},
		{Rune: 'B', Width: 1},
		{Rune: 'C', Width: 1},
		{Rune: 'D', Width: 1},
		{Rune: 'E', Width: 1},
	}
	b.SetRow(0, cells, 0)

	// Only 3 cells should fit
	for i, want := range []rune{'A', 'B', 'C'} {
		cell := b.GetCell(i, 0)
		if cell.Rune != want {
			t.Errorf("cell(%d,0): got %c, want %c", i, cell.Rune, want)
		}
	}
}

func TestSetRowWithCJK(t *testing.T) {
	b := NewBuffer(5, 1)
	cells := []Cell{
		{Rune: '你', Width: 2},
		{Rune: '好', Width: 2},
	}
	b.SetRow(0, cells, 0)

	// 你 at 0, padding at 1
	if b.GetCell(0, 0).Rune != '你' {
		t.Errorf("cell(0,0): got %c, want 你", b.GetCell(0, 0).Rune)
	}
	if b.GetCell(1, 0).Rune != 0 {
		t.Errorf("padding cell(1,0): got %c, want 0", b.GetCell(1, 0).Rune)
	}
	// 好 at 2, padding at 3
	if b.GetCell(2, 0).Rune != '好' {
		t.Errorf("cell(2,0): got %c, want 好", b.GetCell(2, 0).Rune)
	}
}

// ============================================================
// Blit tests
// ============================================================

func TestBlitBasic(t *testing.T) {
	src := NewBuffer(3, 3)
	dst := NewBuffer(5, 5)

	src.SetCell(0, 0, Cell{Rune: 'A', Width: 1})
	src.SetCell(1, 0, Cell{Rune: 'B', Width: 1})
	src.SetCell(0, 1, Cell{Rune: 'C', Width: 1})

	dst.Blit(src, 0, 0, 1, 1, 3, 3)

	if dst.GetCell(1, 1).Rune != 'A' {
		t.Errorf("blit(1,1): got %c, want A", dst.GetCell(1, 1).Rune)
	}
	if dst.GetCell(2, 1).Rune != 'B' {
		t.Errorf("blit(2,1): got %c, want B", dst.GetCell(2, 1).Rune)
	}
	if dst.GetCell(1, 2).Rune != 'C' {
		t.Errorf("blit(1,2): got %c, want C", dst.GetCell(1, 2).Rune)
	}
}

func TestBlitOutOfRange(t *testing.T) {
	src := NewBuffer(2, 2)
	dst := NewBuffer(2, 2)

	src.SetCell(0, 0, Cell{Rune: 'X', Width: 1})

	// Blit with dst offset beyond buffer
	dst.Blit(src, 0, 0, 10, 10, 2, 2)

	// Nothing should change in dst
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			if dst.GetCell(x, y).Rune != ' ' {
				t.Errorf("cell(%d,%d) modified by out-of-range blit", x, y)
			}
		}
	}
}

// ============================================================
// Buffer Diff edge cases
// ============================================================

func TestDiffSizeChanged(t *testing.T) {
	front := NewBuffer(2, 2)
	back := NewBuffer(3, 2)

	ops := Diff(front, back)
	// Size changed → should return all cells from back
	if len(ops) != 6 {
		t.Errorf("expected 6 diff ops for size change, got %d", len(ops))
	}
}

func TestDiffIdentical(t *testing.T) {
	b := NewBuffer(3, 3)
	b.Fill(Cell{Rune: '#', Width: 1})

	ops := Diff(b, b)
	if len(ops) != 0 {
		t.Errorf("expected 0 diff ops for identical buffers, got %d", len(ops))
	}
}

// ============================================================
// Cell helpers edge cases
// ============================================================

func TestCellEqualWithLink(t *testing.T) {
	linkA := &Link{URL: "https://example.com", Text: "example"}
	linkB := &Link{URL: "https://example.com", Text: "different text"}
	linkC := &Link{URL: "https://other.com", Text: "other"}

	c1 := Cell{Rune: 'a', Width: 1, Link: linkA}
	c2 := Cell{Rune: 'a', Width: 1, Link: linkB} // same URL
	c3 := Cell{Rune: 'a', Width: 1, Link: linkC} // different URL
	c4 := Cell{Rune: 'a', Width: 1, Link: nil}   // no link

	if !c1.Equal(c2) {
		t.Error("cells with same URL link should be equal")
	}
	if c1.Equal(c3) {
		t.Error("cells with different URL links should not be equal")
	}
	if c1.Equal(c4) {
		t.Error("cell with link should not equal cell without link")
	}
}

func TestStyledCell(t *testing.T) {
	fg := RGB(255, 0, 0)
	bg := RGB(0, 0, 255)
	cell := StyledCell('X', 1, fg, bg, Bold)

	if cell.Rune != 'X' {
		t.Errorf("rune: got %c, want X", cell.Rune)
	}
	if cell.Width != 1 {
		t.Errorf("width: got %d, want 1", cell.Width)
	}
	if !cell.Fg.Equal(fg) {
		t.Error("fg mismatch")
	}
	if !cell.Bg.Equal(bg) {
		t.Error("bg mismatch")
	}
	if cell.Flags != Bold {
		t.Errorf("flags: got %d, want %d", cell.Flags, Bold)
	}
}

func TestCellAddFlags(t *testing.T) {
	cell := Cell{Rune: 'a', Width: 1}
	cell = cell.AddFlags(Bold | Italic)

	if cell.Flags&Bold == 0 {
		t.Error("Bold flag not set")
	}
	if cell.Flags&Italic == 0 {
		t.Error("Italic flag not set")
	}
}

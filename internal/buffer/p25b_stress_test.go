package buffer

import (
	"strings"
	"sync"
	"testing"
)

// ============================================================
// P25-B: Buffer Edge Case & Stress Tests
// ============================================================

// --- Extreme sizes ---

func TestP25B_LargeBuffer(t *testing.T) {
	b := NewBuffer(1000, 1000)
	if b.Width != 1000 || b.Height != 1000 {
		t.Fatalf("expected 1000x1000, got %dx%d", b.Width, b.Height)
	}
	if len(b.Cells) != 1000*1000 {
		t.Fatalf("expected 1M cells, got %d", len(b.Cells))
	}
	// Fill all cells
	for y := 0; y < 1000; y++ {
		for x := 0; x < 1000; x++ {
			b.SetCell(x, y, Cell{Rune: 'x', Width: 1})
		}
	}
	cell := b.GetCell(999, 999)
	if cell.Rune != 'x' {
		t.Error("expected 'x' at (999,999)")
	}
}

func TestP25B_ZeroSizeBuffer(t *testing.T) {
	b := NewBuffer(0, 0)
	if b.Width != 0 || b.Height != 0 {
		t.Fatalf("expected 0x0")
	}
	if len(b.Cells) != 0 {
		t.Fatalf("expected 0 cells")
	}
	// Operations on zero-size buffer should not panic
	b.SetCell(0, 0, Cell{Rune: 'X', Width: 1})
	_ = b.GetCell(0, 0)
	b.Fill(Cell{Rune: 'X', Width: 1})
	b.DrawText(0, 0, "hello", DefaultStyle)
}

func TestP25B_OneByOne(t *testing.T) {
	b := NewBuffer(1, 1)
	b.SetCell(0, 0, Cell{Rune: 'X', Width: 1})
	cell := b.GetCell(0, 0)
	if cell.Rune != 'X' {
		t.Errorf("expected 'X', got %q", cell.Rune)
	}
}

// --- Out-of-bounds safety ---

func TestP25B_SetCellNegativeCoords(t *testing.T) {
	b := NewBuffer(10, 10)
	b.SetCell(-1, -1, Cell{Rune: 'x', Width: 1})
	b.SetCell(-1, 0, Cell{Rune: 'x', Width: 1})
	b.SetCell(0, -1, Cell{Rune: 'x', Width: 1})
	cell := b.GetCell(0, 0)
	if cell.Rune != ' ' {
		t.Error("negative coords should not modify buffer")
	}
}

func TestP25B_SetCellOverflowCoords(t *testing.T) {
	b := NewBuffer(10, 10)
	b.SetCell(100, 100, Cell{Rune: 'x', Width: 1})
	b.SetCell(10, 10, Cell{Rune: 'x', Width: 1})
	cell := b.GetCell(0, 0)
	if cell.Rune != ' ' {
		t.Error("overflow coords should not modify buffer")
	}
}

func TestP25B_GetCellNegative(t *testing.T) {
	b := NewBuffer(10, 10)
	cell := b.GetCell(-1, -1)
	if cell.Rune != ' ' && cell.Rune != 0 {
		t.Errorf("expected blank for negative coords, got %q", string(cell.Rune))
	}
}

func TestP25B_GetCellOverflow(t *testing.T) {
	b := NewBuffer(10, 10)
	cell := b.GetCell(100, 100)
	if cell.Rune != ' ' && cell.Rune != 0 {
		t.Errorf("expected blank for overflow coords, got %q", string(cell.Rune))
	}
}

// --- DrawText edge cases ---

func TestP25B_DrawTextEmpty(t *testing.T) {
	b := NewBuffer(10, 5)
	b.DrawText(0, 0, "", DefaultStyle)
	cell := b.GetCell(0, 0)
	if cell.Rune != ' ' {
		t.Error("empty string should not modify buffer")
	}
}

func TestP25B_DrawTextOverflow(t *testing.T) {
	b := NewBuffer(5, 3)
	b.DrawText(0, 0, strings.Repeat("A", 100), DefaultStyle)
	for x := 0; x < 5; x++ {
		cell := b.GetCell(x, 0)
		if cell.Rune != 'A' {
			t.Errorf("expected 'A' at x=%d, got %q", x, cell.Rune)
		}
	}
}

func TestP25B_DrawTextNegativePosition(t *testing.T) {
	b := NewBuffer(10, 5)
	b.DrawText(-1, -1, "test", DefaultStyle)
	// Should not crash
}

func TestP25B_DrawTextWideCharAtEdge(t *testing.T) {
	b := NewBuffer(3, 1)
	// Wide char at position 2 — only 1 cell left
	b.DrawText(2, 0, "\u4e2d", DefaultStyle)
	// Should not panic
	_ = b.GetCell(2, 0)
}

func TestP25B_DrawTextUnicode(t *testing.T) {
	b := NewBuffer(80, 1)
	text := "\u4e16\u754c\u4f60\u597d" // CJK
	b.DrawText(0, 0, text, DefaultStyle)
	cell := b.GetCell(0, 0)
	if cell.Rune != '\u4e16' {
		t.Errorf("expected CJK char, got %q", string(cell.Rune))
	}
	if cell.Width != 2 {
		t.Errorf("expected width 2, got %d", cell.Width)
	}
}

func TestP25B_DrawTextWithStyle(t *testing.T) {
	b := NewBuffer(10, 1)
	style := Style{Fg: Red, Bg: Blue, Flags: Bold | Underline}
	b.DrawText(0, 0, "ABC", style)
	cell := b.GetCell(0, 0)
	if cell.Flags&Bold == 0 {
		t.Error("Bold not applied")
	}
	if cell.Flags&Underline == 0 {
		t.Error("Underline not applied")
	}
	if !cell.Fg.Equal(Red) {
		t.Error("Fg Red not applied")
	}
	if !cell.Bg.Equal(Blue) {
		t.Error("Bg Blue not applied")
	}
}

// --- Fill edge cases ---

func TestP25B_FillEntireBuffer(t *testing.T) {
	b := NewBuffer(10, 10)
	c := Cell{Rune: '#', Width: 1, Fg: Red}
	b.Fill(c)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			cell := b.GetCell(x, y)
			if cell.Rune != '#' {
				t.Errorf("expected '#' at (%d,%d), got %q", x, y, cell.Rune)
			}
		}
	}
}

func TestP25B_FillRect(t *testing.T) {
	b := NewBuffer(80, 24)
	b.FillRect(Rect{X: 10, Y: 5, W: 20, H: 10}, Cell{Rune: '#', Width: 1, Fg: Green})

	// Inside rect
	for _, tc := range []struct{ x, y int }{{10, 5}, {29, 14}, {15, 10}} {
		cell := b.GetCell(tc.x, tc.y)
		if cell.Rune != '#' {
			t.Errorf("expected '#' at (%d,%d), got %q", tc.x, tc.y, cell.Rune)
		}
	}
	// Outside rect
	for _, tc := range []struct{ x, y int }{{0, 0}, {50, 20}, {9, 5}, {30, 14}} {
		cell := b.GetCell(tc.x, tc.y)
		if cell.Rune == '#' {
			t.Errorf("should not have '#' at (%d,%d)", tc.x, tc.y)
		}
	}
}

// --- Blit edge cases ---

func TestP25B_BlitBasic(t *testing.T) {
	src := NewBuffer(5, 5)
	for i := 0; i < 5; i++ {
		src.SetCell(i, i, Cell{Rune: 'X', Width: 1})
	}
	dst := NewBuffer(10, 10)
	dst.Blit(src, 0, 0, 0, 0, 5, 5)
	for i := 0; i < 5; i++ {
		cell := dst.GetCell(i, i)
		if cell.Rune != 'X' {
			t.Errorf("expected 'X' at (%d,%d), got %q", i, i, cell.Rune)
		}
	}
}

func TestP25B_BlitPartialOverlap(t *testing.T) {
	src := NewBuffer(5, 5)
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			src.SetCell(x, y, Cell{Rune: 'S', Width: 1})
		}
	}
	dst := NewBuffer(10, 10)
	dst.Blit(src, 0, 0, 8, 8, 5, 5)
	// Only (8,8) and (8,9) and (9,8) and (9,9) fit
	cell := dst.GetCell(8, 8)
	if cell.Rune != 'S' {
		t.Error("expected 'S' at (8,8)")
	}
	cell2 := dst.GetCell(7, 7)
	if cell2.Rune == 'S' {
		t.Error("should not have 'S' at (7,7)")
	}
}

// --- SetRow ---

func TestP25B_SetRowBasic(t *testing.T) {
	b := NewBuffer(10, 3)
	cells := []Cell{
		{Rune: 'A', Width: 1},
		{Rune: 'B', Width: 1},
		{Rune: 'C', Width: 1},
	}
	b.SetRow(1, cells, 2)
	if b.GetCell(2, 1).Rune != 'A' {
		t.Error("SetRow failed")
	}
	if b.GetCell(3, 1).Rune != 'B' {
		t.Error("SetRow failed")
	}
}

func TestP25B_SetRowOverflow(t *testing.T) {
	b := NewBuffer(3, 3)
	cells := []Cell{{Rune: 'X', Width: 1}, {Rune: 'X', Width: 1}, {Rune: 'X', Width: 1}, {Rune: 'X', Width: 1}, {Rune: 'X', Width: 1}}
	b.SetRow(0, cells, 0)
	// Should only write 3 cells (buffer width)
	for x := 0; x < 3; x++ {
		if b.GetCell(x, 0).Rune != 'X' {
			t.Error("SetRow should fill to width")
		}
	}
}

// --- Diff ---

func TestP25B_DiffIdentical(t *testing.T) {
	b1 := NewBuffer(10, 10)
	b2 := NewBuffer(10, 10)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			cell := Cell{Rune: 'X', Width: 1}
			b1.SetCell(x, y, cell)
			b2.SetCell(x, y, cell)
		}
	}
	ops := Diff(b1, b2)
	if len(ops) != 0 {
		t.Errorf("identical buffers should have 0 diff ops, got %d", len(ops))
	}
}

func TestP25B_DiffCompletelyDifferent(t *testing.T) {
	b1 := NewBuffer(10, 10)
	b2 := NewBuffer(10, 10)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			b1.SetCell(x, y, Cell{Rune: 'A', Width: 1})
			b2.SetCell(x, y, Cell{Rune: 'B', Width: 1})
		}
	}
	ops := Diff(b1, b2)
	if len(ops) != 100 {
		t.Errorf("expected 100 diff ops, got %d", len(ops))
	}
}

func TestP25B_DiffPartialRow(t *testing.T) {
	b1 := NewBuffer(80, 24)
	b2 := NewBuffer(80, 24)
	// Make only middle row different
	for x := 0; x < 80; x++ {
		b1.SetCell(x, 12, Cell{Rune: 'A', Width: 1})
		b2.SetCell(x, 12, Cell{Rune: 'B', Width: 1})
	}
	ops := Diff(b1, b2)
	if len(ops) != 80 {
		t.Errorf("partial diff should have 80 ops, got %d", len(ops))
	}
}

func TestP25B_DiffDifferentSizes(t *testing.T) {
	b1 := NewBuffer(5, 5)
	b2 := NewBuffer(10, 10)
	ops := Diff(b1, b2)
	// Should return all cells of back buffer
	if len(ops) != 100 {
		t.Errorf("different sizes should return all back cells, got %d", len(ops))
	}
}

func TestP25B_DiffIntoReuse(t *testing.T) {
	b1 := NewBuffer(80, 24)
	b2 := NewBuffer(80, 24)
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			b1.SetCell(x, y, Cell{Rune: 'A', Width: 1})
			b2.SetCell(x, y, Cell{Rune: 'B', Width: 1})
		}
	}
	base := make([]DiffOp, 0, 100)
	ops := DiffInto(b1, b2, base)
	if len(ops) != 1920 {
		t.Errorf("expected 1920 ops, got %d", len(ops))
	}
}

// --- RepeatString ---

func TestP25B_RepeatStringZero(t *testing.T) {
	if s := RepeatString("hello", 0); s != "" {
		t.Errorf("expected empty, got %q", s)
	}
}

func TestP25B_RepeatStringNegative(t *testing.T) {
	if s := RepeatString("hello", -5); s != "" {
		t.Errorf("expected empty for negative, got %q", s)
	}
}

func TestP25B_RepeatStringLarge(t *testing.T) {
	s := RepeatString("ab", 10000)
	if len(s) != 20000 {
		t.Errorf("expected 20000 chars, got %d", len(s))
	}
}

// --- NewCell & Cell methods ---

func TestP25B_NewCellBasic(t *testing.T) {
	c := NewCell('A', DefaultStyle)
	if c.Rune != 'A' || c.Width != 1 {
		t.Errorf("NewCell('A'): got rune=%q width=%d", c.Rune, c.Width)
	}
}

func TestP25B_NewCellWideRune(t *testing.T) {
	c := NewCell('\u4e2d', DefaultStyle)
	if c.Width != 2 {
		t.Errorf("expected width 2 for wide rune, got %d", c.Width)
	}
}

func TestP25B_NewCellWithStyle(t *testing.T) {
	style := Style{Fg: Red, Bg: Blue, Flags: Bold | Italic}
	c := NewCell('X', style)
	if !c.Fg.Equal(Red) {
		t.Error("Fg not set")
	}
	if !c.Bg.Equal(Blue) {
		t.Error("Bg not set")
	}
	if c.Flags&(Bold|Italic) != Bold|Italic {
		t.Error("Flags not set")
	}
}

func TestP25B_CellEqual(t *testing.T) {
	c1 := NewCell('X', Style{Fg: Red, Flags: Bold})
	c2 := NewCell('X', Style{Fg: Red, Flags: Bold})
	c3 := NewCell('Y', DefaultStyle)
	if !c1.Equal(c2) {
		t.Error("identical cells should be equal")
	}
	if c1.Equal(c3) {
		t.Error("different cells should not be equal")
	}
}

func TestP25B_CellAddFlags(t *testing.T) {
	c := NewCell('X', DefaultStyle)
	c = c.AddFlags(Bold | Underline)
	if c.Flags&Bold == 0 || c.Flags&Underline == 0 {
		t.Error("AddFlags failed")
	}
}

func TestP25B_CellWithFg(t *testing.T) {
	c := NewCell('X', DefaultStyle)
	c = c.WithFg(Green)
	if !c.Fg.Equal(Green) {
		t.Error("WithFg failed")
	}
}

func TestP25B_CellFastEqualNil(t *testing.T) {
	c1 := NewCell('X', DefaultStyle)
	c2 := NewCell('X', DefaultStyle)
	if !cellFastEqual(c1, c2) {
		t.Error("nil-link cells should be fast-equal")
	}
	c3 := NewCell('Y', DefaultStyle)
	if cellFastEqual(c1, c3) {
		t.Error("different cells should not be fast-equal")
	}
}

func TestP25B_CellFastEqualSameLink(t *testing.T) {
	link := &Link{URL: "https://example.com"}
	c1 := Cell{Rune: 'L', Width: 1, Link: link}
	c2 := Cell{Rune: 'L', Width: 1, Link: link}
	if !cellFastEqual(c1, c2) {
		t.Error("same-link cells should be fast-equal")
	}
}

func TestP25B_CellFastEqualDifferentLink(t *testing.T) {
	c1 := Cell{Rune: 'L', Width: 1, Link: &Link{URL: "https://a.com"}}
	c2 := Cell{Rune: 'L', Width: 1, Link: &Link{URL: "https://b.com"}}
	if cellFastEqual(c1, c2) {
		t.Error("different-link cells should not be fast-equal")
	}
}

// --- Links ---

func TestP25B_BufferSetGetLink(t *testing.T) {
	b := NewBuffer(80, 24)
	link := &Link{URL: "https://example.com", Text: "example"}
	b.SetCell(0, 0, Cell{Rune: 'L', Width: 1, Link: link})
	cell := b.GetCell(0, 0)
	if cell.Link != link {
		t.Error("link pointer not preserved")
	}
}

func TestP25B_BufferMultipleLinks(t *testing.T) {
	b := NewBuffer(80, 1)
	link1 := &Link{URL: "https://a.com"}
	link2 := &Link{URL: "https://b.com"}
	for x := 0; x < 10; x++ {
		b.SetCell(x, 0, Cell{Rune: 'A', Width: 1, Link: link1})
	}
	for x := 10; x < 20; x++ {
		b.SetCell(x, 0, Cell{Rune: 'B', Width: 1, Link: link2})
	}
	if b.GetCell(0, 0).Link != link1 {
		t.Error("link1 not preserved")
	}
	if b.GetCell(10, 0).Link != link2 {
		t.Error("link2 not preserved")
	}
}

// --- Style tests ---

func TestP25B_StyleAllFlags(t *testing.T) {
	allFlags := Bold | Italic | Underline | Strikethrough | Reverse | Dim | Blink
	s := Style{Flags: allFlags}
	c := NewCell('X', s)
	if c.Flags != allFlags {
		t.Error("all style flags not preserved")
	}
}

func TestP25B_StyleAddFlags(t *testing.T) {
	s := DefaultStyle.AddFlags(Bold)
	if !s.HasFlag(Bold) {
		t.Error("AddFlags Bold failed")
	}
	s = s.AddFlags(Italic)
	if !s.HasFlag(Bold) || !s.HasFlag(Italic) {
		t.Error("cumulative AddFlags failed")
	}
}

func TestP25B_StyleWithFg(t *testing.T) {
	s := DefaultStyle.WithFg(Red)
	if !s.Fg.Equal(Red) {
		t.Error("WithFg failed")
	}
}

func TestP25B_StyleEqual(t *testing.T) {
	s1 := Style{Fg: Red, Bg: Blue, Flags: Bold}
	s2 := Style{Fg: Red, Bg: Blue, Flags: Bold}
	s3 := Style{Fg: Red, Bg: Blue, Flags: Italic}
	if !s1.Equal(s2) {
		t.Error("identical styles should be equal")
	}
	if s1.Equal(s3) {
		t.Error("different styles should not be equal")
	}
}

// --- Color tests ---

func TestP25B_RGBColor(t *testing.T) {
	c := RGB(255, 128, 0)
	if c.Type != ColorTrue {
		t.Error("expected ColorTrue")
	}
	if c.R() != 255 || c.G() != 128 || c.B() != 0 {
		t.Errorf("RGB components: R=%d G=%d B=%d", c.R(), c.G(), c.B())
	}
}

func TestP25B_NoColor(t *testing.T) {
	c := NoColor()
	if !c.IsDefault() {
		t.Error("NoColor should be default")
	}
}

func TestP25B_NamedColor(t *testing.T) {
	c := NamedColor(NamedRed)
	if c.Type != ColorNamed {
		t.Error("expected ColorNamed")
	}
	if c.Val != NamedRed {
		t.Error("wrong named color value")
	}
}

func TestP25B_HexColor(t *testing.T) {
	c := Hex("#ff8000")
	if c.Type != ColorTrue {
		t.Error("expected ColorTrue for hex")
	}
	if c.R() != 255 || c.G() != 128 || c.B() != 0 {
		t.Errorf("hex color components wrong")
	}
}

func TestP25B_HexColorInvalid(t *testing.T) {
	c := Hex("xyz")
	if !c.IsDefault() {
		t.Error("invalid hex should return default color")
	}
}

func TestP25B_ColorEqual(t *testing.T) {
	c1 := RGB(1, 2, 3)
	c2 := RGB(1, 2, 3)
	c3 := RGB(4, 5, 6)
	if !c1.Equal(c2) {
		t.Error("identical colors should be equal")
	}
	if c1.Equal(c3) {
		t.Error("different colors should not be equal")
	}
}

// --- Concurrent stress ---

func TestP25B_ConcurrentGet(t *testing.T) {
	b := NewBuffer(80, 24)
	// Pre-fill with known content
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			b.SetCell(x, y, Cell{Rune: 'X', Width: 1})
		}
	}

	var wg sync.WaitGroup
	const goroutines = 20
	// Concurrent reads are safe (no mutex needed for read-only access)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				cell := b.GetCell(j%80, (id*3+j)%24)
				if cell.Rune != 'X' {
					t.Error("concurrent read got unexpected cell")
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestP25B_ConcurrentDiff(t *testing.T) {
	b1 := NewBuffer(80, 24)
	b2 := NewBuffer(80, 24)
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			b1.SetCell(x, y, Cell{Rune: 'A', Width: 1})
			b2.SetCell(x, y, Cell{Rune: 'B', Width: 1})
		}
	}
	var wg sync.WaitGroup
	const goroutines = 20
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ops := Diff(b1, b2)
			if len(ops) == 0 {
				t.Error("concurrent diff should find differences")
			}
		}()
	}
	wg.Wait()
}

func TestP25B_RapidFillDiff(t *testing.T) {
	b1 := NewBuffer(80, 24)
	b2 := NewBuffer(80, 24)
	for iter := 0; iter < 100; iter++ {
		for y := 0; y < 24; y++ {
			for x := 0; x < 80; x++ {
				r := rune('A' + (x+y+iter)%26)
				b1.SetCell(x, y, Cell{Rune: r, Width: 1})
				b2.SetCell(x, y, Cell{Rune: r + 1, Width: 1})
			}
		}
		_ = Diff(b1, b2)
	}
}

func TestP25B_ConcurrentDrawTextSafe(t *testing.T) {
	// DrawText on separate buffers is safe
	var wg sync.WaitGroup
	const goroutines = 20
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			b := NewBuffer(80, 1)
			for j := 0; j < 50; j++ {
				b.DrawText(0, 0, "hello world", DefaultStyle)
			}
			_ = id
		}(i)
	}
	wg.Wait()
}

// --- RuneWidth ---

func TestP25B_RuneWidthASCII(t *testing.T) {
	if RuneWidth('A') != 1 {
		t.Error("ASCII should be width 1")
	}
	if RuneWidth(' ') != 1 {
		t.Error("space should be width 1")
	}
}

func TestP25B_RuneWidthCJK(t *testing.T) {
	if RuneWidth('\u4e2d') != 2 {
		t.Error("CJK should be width 2")
	}
	if RuneWidth('\u4e16') != 2 {
		t.Error("CJK should be width 2")
	}
}

func TestP25B_RuneWidthEmoji(t *testing.T) {
	if RuneWidth('\U0001f600') != 2 {
		t.Error("emoji should be width 2")
	}
}

func TestP25B_RuneWidthCombining(t *testing.T) {
	if RuneWidth('\u0301') != 0 {
		t.Error("combining char should be width 0")
	}
	if RuneWidth('\u200d') != 0 {
		t.Error("zero-width joiner should be width 0")
	}
}

func TestP25B_RuneWidthControl(t *testing.T) {
	// Control characters
	if RuneWidth(0) != 1 {
		t.Logf("null rune width: %d", RuneWidth(0))
	}
	if RuneWidth('\x1b') != 1 {
		t.Logf("escape rune width: %d", RuneWidth('\x1b'))
	}
}

// --- SGR Sequence ---

func TestP25B_SGRSequence(t *testing.T) {
	s := Style{Fg: RGB(255, 0, 0), Bg: NoColor(), Flags: Bold}
	seq := s.SGRSequence()
	if !strings.Contains(seq, "1") {
		t.Error("SGR should contain Bold code")
	}
	if !strings.Contains(seq, "38;2;255;0;0") {
		t.Error("SGR should contain true color fg")
	}
}

func TestP25B_SGRSequenceEmpty(t *testing.T) {
	s := DefaultStyle
	seq := s.SGRSequence()
	// Should just be default fg/bg
	_ = seq
}

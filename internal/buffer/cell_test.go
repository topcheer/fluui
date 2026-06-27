package buffer

import (
	"testing"
)

func TestNewCell(t *testing.T) {
	style := Style{Fg: RGB(255, 0, 0), Bg: RGB(0, 0, 255), Flags: Bold}
	c := NewCell('X', style)
	if c.Rune != 'X' {
		t.Errorf("Rune = %q, want 'X'", c.Rune)
	}
	if c.Width != 1 {
		t.Errorf("Width = %d, want 1", c.Width)
	}
	if !c.Fg.Equal(RGB(255, 0, 0)) {
		t.Errorf("Fg = %v, want RGB(255,0,0)", c.Fg)
	}
	if !c.Bg.Equal(RGB(0, 0, 255)) {
		t.Errorf("Bg = %v, want RGB(0,0,255)", c.Bg)
	}
	if c.Flags != Bold {
		t.Errorf("Flags = %d, want Bold(%d)", c.Flags, Bold)
	}
	if c.Link != nil {
		t.Error("Link should be nil")
	}
}

func TestNewCell_CJK(t *testing.T) {
	c := NewCell('你', DefaultStyle)
	if c.Rune != '你' {
		t.Errorf("Rune = %q", c.Rune)
	}
	if c.Width != 2 {
		t.Errorf("Width = %d, want 2 for CJK", c.Width)
	}
}

func TestStyledCell_Basic(t *testing.T) {
	c := StyledCell('A', 1, RGB(1, 2, 3), RGB(4, 5, 6), Italic)
	if c.Rune != 'A' || c.Width != 1 {
		t.Errorf("Rune=%q Width=%d", c.Rune, c.Width)
	}
	if !c.Fg.Equal(RGB(1, 2, 3)) {
		t.Errorf("Fg = %v", c.Fg)
	}
	if !c.Bg.Equal(RGB(4, 5, 6)) {
		t.Errorf("Bg = %v", c.Bg)
	}
	if c.Flags != Italic {
		t.Errorf("Flags = %d", c.Flags)
	}
}

func TestCellEqual_Comprehensive(t *testing.T) {
	a := StyledCell('A', 1, RGB(1, 1, 1), RGB(2, 2, 2), Bold)
	b := StyledCell('A', 1, RGB(1, 1, 1), RGB(2, 2, 2), Bold)
	if !a.Equal(b) {
		t.Error("identical cells should be equal")
	}

	// Different rune
	c := StyledCell('B', 1, RGB(1, 1, 1), RGB(2, 2, 2), Bold)
	if a.Equal(c) {
		t.Error("different rune should not be equal")
	}

	// Different width
	d := StyledCell('A', 2, RGB(1, 1, 1), RGB(2, 2, 2), Bold)
	if a.Equal(d) {
		t.Error("different width should not be equal")
	}

	// Different flags
	e := StyledCell('A', 1, RGB(1, 1, 1), RGB(2, 2, 2), Italic)
	if a.Equal(e) {
		t.Error("different flags should not be equal")
	}
}

func TestCellEqual_Links(t *testing.T) {
	without := StyledCell('A', 1, NoColor(), NoColor(), 0)
	with := without
	with.Link = &Link{URL: "https://example.com", Text: "link"}

	// nil vs nil
	if !without.Equal(without) {
		t.Error("same cell should be equal")
	}

	// nil vs non-nil
	if without.Equal(with) {
		t.Error("nil link vs non-nil link should not be equal")
	}

	// Same URL
	sameLink := without
	sameLink.Link = &Link{URL: "https://example.com", Text: "different text"}
	if !with.Equal(sameLink) {
		t.Error("same URL links should be equal regardless of text")
	}

	// Different URL
	diffLink := without
	diffLink.Link = &Link{URL: "https://other.com", Text: "link"}
	if with.Equal(diffLink) {
		t.Error("different URL links should not be equal")
	}
}

func TestCellWithStyle(t *testing.T) {
	c := StyledCell('A', 1, RGB(0, 0, 0), RGB(0, 0, 0), 0)
	newStyle := Style{Fg: RGB(1, 2, 3), Bg: RGB(4, 5, 6), Flags: Bold | Underline}
	c2 := c.WithStyle(newStyle)
	if !c2.Fg.Equal(RGB(1, 2, 3)) {
		t.Errorf("Fg = %v", c2.Fg)
	}
	if !c2.Bg.Equal(RGB(4, 5, 6)) {
		t.Errorf("Bg = %v", c2.Bg)
	}
	if c2.Flags != Bold|Underline {
		t.Errorf("Flags = %d", c2.Flags)
	}
	// Original rune/width preserved
	if c2.Rune != c.Rune || c2.Width != c.Width {
		t.Error("WithStyle should preserve rune and width")
	}
}

func TestCellWithFg(t *testing.T) {
	c := StyledCell('A', 1, NoColor(), NoColor(), 0)
	c2 := c.WithFg(RGB(255, 0, 0))
	if !c2.Fg.Equal(RGB(255, 0, 0)) {
		t.Errorf("Fg = %v", c2.Fg)
	}
	// Other fields unchanged
	if !c2.Bg.Equal(NoColor()) {
		t.Error("Bg should be unchanged")
	}
}

func TestCellWithBg(t *testing.T) {
	c := StyledCell('A', 1, NoColor(), NoColor(), 0)
	c2 := c.WithBg(RGB(0, 255, 0))
	if !c2.Bg.Equal(RGB(0, 255, 0)) {
		t.Errorf("Bg = %v", c2.Bg)
	}
	if !c2.Fg.Equal(NoColor()) {
		t.Error("Fg should be unchanged")
	}
}

func TestCellAddFlags_Comprehensive(t *testing.T) {
	c := StyledCell('A', 1, NoColor(), NoColor(), 0)
	c2 := c.AddFlags(Bold)
	if c2.Flags != Bold {
		t.Errorf("Flags = %d, want Bold", c2.Flags)
	}
	c3 := c2.AddFlags(Italic)
	if c3.Flags != Bold|Italic {
		t.Errorf("Flags = %d, want Bold|Italic", c3.Flags)
	}
	// Original unchanged
	if c.Flags != 0 {
		t.Error("Original cell was mutated")
	}
}

func TestBlankCell(t *testing.T) {
	if BlankCell.Rune != ' ' {
		t.Errorf("BlankCell.Rune = %q, want space", BlankCell.Rune)
	}
	if BlankCell.Width != 1 {
		t.Errorf("BlankCell.Width = %d, want 1", BlankCell.Width)
	}
}

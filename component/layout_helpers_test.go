package component

import (
	"strings"
	"testing"
)

func TestJoinHorizontal_Top(t *testing.T) {
	left := "AB\nCD"
	right := "EF\nGH"
	got := JoinHorizontal(Top, left, right)
	// Each block is 2 wide, result should be 4 wide, 2 tall
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "ABEF" {
		t.Errorf("line 0: expected 'ABEF', got %q", lines[0])
	}
	if lines[1] != "CDGH" {
		t.Errorf("line 1: expected 'CDGH', got %q", lines[1])
	}
}

func TestJoinHorizontal_Middle(t *testing.T) {
	tall := "A\nB\nC"
	short := "X"
	got := JoinHorizontal(Middle, tall, short)
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	// X should be in the middle (row 1)
	if lines[1][1] != 'X' {
		t.Errorf("expected X at row 1 col 1, got %q", lines[1])
	}
}

func TestJoinHorizontal_Bottom(t *testing.T) {
	tall := "A\nB\nC"
	short := "X"
	got := JoinHorizontal(Bottom, tall, short)
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	// X should be at bottom (row 2)
	if lines[2][1] != 'X' {
		t.Errorf("expected X at row 2 col 1, got %q", lines[2])
	}
}

func TestJoinHorizontal_Empty(t *testing.T) {
	got := JoinHorizontal(Top)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestJoinHorizontal_Single(t *testing.T) {
	got := JoinHorizontal(Top, "AB")
	if got != "AB" {
		t.Errorf("expected 'AB', got %q", got)
	}
}

func TestJoinHorizontal_DifferentWidths(t *testing.T) {
	left := "A\nB"
	right := "CDE\nFGH"
	got := JoinHorizontal(Top, left, right)
	lines := strings.Split(got, "\n")
	// left is 1 wide, right is 3 wide → total 4
	if lines[0] != "ACDE" {
		t.Errorf("line 0: expected 'ACDE', got %q", lines[0])
	}
	if lines[1] != "BFGH" {
		t.Errorf("line 1: expected 'BFGH', got %q", lines[1])
	}
}

func TestJoinVertical_Left(t *testing.T) {
	top := "AB\nCD"
	bottom := "EF\nGH"
	got := JoinVertical(Left, top, bottom)
	lines := strings.Split(got, "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}
	if lines[0] != "AB" {
		t.Errorf("line 0: expected 'AB', got %q", lines[0])
	}
	if lines[1] != "CD" {
		t.Errorf("line 1: expected 'CD', got %q", lines[1])
	}
	if lines[2] != "EF" {
		t.Errorf("line 2: expected 'EF', got %q", lines[2])
	}
	if lines[3] != "GH" {
		t.Errorf("line 3: expected 'GH', got %q", lines[3])
	}
}

func TestJoinVertical_Center(t *testing.T) {
	top := "AB"
	bottom := "CDEFG"
	got := JoinVertical(Center, top, bottom)
	lines := strings.Split(got, "\n")
	// top has width 2, bottom has width 5
	// center: (5-2)/2 = 1 left pad, right pad = 2
	if lines[0] != " AB  " {
		t.Errorf("line 0: expected ' AB ', got %q", lines[0])
	}
	if lines[1] != "CDEFG" {
		t.Errorf("line 1: expected 'CDEFG', got %q", lines[1])
	}
}

func TestJoinVertical_Right(t *testing.T) {
	top := "AB"
	bottom := "CDEFG"
	got := JoinVertical(Right, top, bottom)
	lines := strings.Split(got, "\n")
	// right: 3 spaces pad before AB
	if lines[0] != "   AB" {
		t.Errorf("line 0: expected '   AB', got %q", lines[0])
	}
	if lines[1] != "CDEFG" {
		t.Errorf("line 1: expected 'CDEFG', got %q", lines[1])
	}
}

func TestJoinVertical_Empty(t *testing.T) {
	got := JoinVertical(Left)
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestPlace_Center(t *testing.T) {
	content := "Hi"
	got := Place(10, 5, Center, Middle, content)
	lines := strings.Split(got, "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
	// Content at row 2 (middle of 5), col 4 (center of 10, width 2)
	if !strings.Contains(lines[2], "Hi") {
		t.Errorf("expected 'Hi' in middle row, got %q", lines[2])
	}
}

func TestPlace_TopLeft(t *testing.T) {
	content := "Hi"
	got := Place(10, 5, Left, Top, content)
	lines := strings.Split(got, "\n")
	if lines[0] != "Hi        " {
		t.Errorf("expected 'Hi' padded to 10, got %q", lines[0])
	}
}

func TestPlace_BottomRight(t *testing.T) {
	content := "Hi"
	got := Place(10, 5, Right, Bottom, content)
	lines := strings.Split(got, "\n")
	// Content at last row, right-aligned
	if !strings.HasSuffix(lines[4], "Hi") {
		t.Errorf("expected 'Hi' at end of last row, got %q", lines[4])
	}
}

func TestPlaceHorizontal_Center(t *testing.T) {
	got := PlaceHorizontal(10, Center, "Hi")
	// (10-2)/2 = 4 left pad
	if got != "    Hi    " {
		t.Errorf("expected '    Hi    ', got %q", got)
	}
}

func TestPlaceHorizontal_Left(t *testing.T) {
	got := PlaceHorizontal(10, Left, "Hi")
	if got != "Hi        " {
		t.Errorf("expected 'Hi        ', got %q", got)
	}
}

func TestPlaceHorizontal_Right(t *testing.T) {
	got := PlaceHorizontal(10, Right, "Hi")
	if got != "        Hi" {
		t.Errorf("expected '        Hi', got %q", got)
	}
}

func TestPlaceHorizontal_TooWide(t *testing.T) {
	got := PlaceHorizontal(2, Left, "Hello")
	if got != "Hello" {
		t.Errorf("expected 'Hello' (no truncation), got %q", got)
	}
}

func TestPlaceVertical_Top(t *testing.T) {
	got := PlaceVertical(5, Top, "Hi")
	lines := strings.Split(got, "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
	if lines[0] != "Hi" {
		t.Errorf("expected 'Hi' at top, got %q", lines[0])
	}
}

func TestPlaceVertical_Bottom(t *testing.T) {
	got := PlaceVertical(5, Bottom, "Hi")
	lines := strings.Split(got, "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines, got %d", len(lines))
	}
	if lines[4] != "Hi" {
		t.Errorf("expected 'Hi' at bottom, got %q", lines[4])
	}
}

func TestWidth(t *testing.T) {
	if Width("Hello") != 5 {
		t.Errorf("expected 5, got %d", Width("Hello"))
	}
	if Width("") != 0 {
		t.Errorf("expected 0, got %d", Width(""))
	}
}

func TestMaxWidth(t *testing.T) {
	if MaxWidth("AB\nCDEF\nG") != 4 {
		t.Errorf("expected 4, got %d", MaxWidth("AB\nCDEF\nG"))
	}
}

func TestHeight(t *testing.T) {
	if Height("A\nB\nC") != 3 {
		t.Errorf("expected 3, got %d", Height("A\nB\nC"))
	}
	if Height("") != 0 {
		t.Errorf("expected 0, got %d", Height(""))
	}
	if Height("Single") != 1 {
		t.Errorf("expected 1, got %d", Height("Single"))
	}
}

func TestJoinHorizontal_ThreeBlocks(t *testing.T) {
	a := "X\nY"
	b := "Z"
	c := "W\nV"
	got := JoinHorizontal(Top, a, b, c)
	lines := strings.Split(got, "\n")
	if lines[0] != "XZW" {
		t.Errorf("expected 'XZW', got %q", lines[0])
	}
	// Row 1: Y   V → "Y V" (b is padded with space since b only has 1 line)
	if lines[1] != "Y V" {
		t.Errorf("expected 'YV', got %q", lines[1])
	}
}

func TestJoinVertical_ThreeBlocks(t *testing.T) {
	got := JoinVertical(Left, "A", "B", "C")
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "A" || lines[1] != "B" || lines[2] != "C" {
		t.Errorf("expected A,B,C; got %v", lines)
	}
}
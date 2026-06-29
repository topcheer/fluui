package buffer

import "testing"

func TestHighlightMatches_Basic(t *testing.T) {
	buf := NewBuffer(20, 3)
	buf.DrawText(0, 0, "hello world", DefaultStyle)
	buf.DrawText(0, 1, "goodbye world", DefaultStyle)

	style := Style{Flags: Reverse}
	matches := []TextMatch{
		{X: 0, Y: 0, Length: 5},  // "hello"
		{X: 0, Y: 1, Length: 7},  // "goodbye"
	}
	HighlightMatches(buf, matches, style)

	// "hello" cells should have Reverse
	for x := 0; x < 5; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Flags&Reverse == 0 {
			t.Errorf("cell[%d,0] should have Reverse flag", x)
		}
	}
	// "goodbye" cells should have Reverse
	for x := 0; x < 7; x++ {
		cell := buf.GetCell(x, 1)
		if cell.Flags&Reverse == 0 {
			t.Errorf("cell[%d,1] should have Reverse flag", x)
		}
	}
}

func TestHighlightMatches_OutOfBounds(t *testing.T) {
	buf := NewBuffer(5, 1)
	matches := []TextMatch{
		{X: 3, Y: 0, Length: 10}, // extends past width
	}
	HighlightMatches(buf, matches, Style{Flags: Bold})
	// Should not panic. Cells 3-4 modified, cells 5+ ignored.
}

func TestHighlightMatches_Empty(t *testing.T) {
	buf := NewBuffer(10, 1)
	HighlightMatches(buf, nil, DefaultStyle)
	// Should not panic
}

func TestHighlightCurrentMatch(t *testing.T) {
	buf := NewBuffer(20, 1)
	buf.DrawText(0, 0, "foo bar foo", DefaultStyle)

	matches := []TextMatch{
		{X: 0, Y: 0, Length: 3},
		{X: 8, Y: 0, Length: 3},
	}
	normalStyle := Style{Flags: Bold}
	currentStyle := Style{Flags: Reverse}

	HighlightCurrentMatch(buf, matches, 1, normalStyle, currentStyle)

	// First match should have Bold (normal)
	cell := buf.GetCell(0, 0)
	if cell.Flags&Bold == 0 {
		t.Error("first match should have Bold (normal style)")
	}
	if cell.Flags&Reverse != 0 {
		t.Error("first match should NOT have Reverse")
	}

	// Second match (current) should have Reverse
	cell = buf.GetCell(8, 0)
	if cell.Flags&Reverse == 0 {
		t.Error("current match should have Reverse")
	}
}

func TestFindTextInRow_Basic(t *testing.T) {
	buf := NewBuffer(20, 1)
	buf.DrawText(0, 0, "hello world hello", DefaultStyle)

	matches := FindTextInRow(buf, 0, "hello")
	if len(matches) != 2 {
		t.Fatalf("len(matches) = %d, want 2", len(matches))
	}
	if matches[0].X != 0 {
		t.Errorf("matches[0].X = %d, want 0", matches[0].X)
	}
	if matches[1].X != 12 {
		t.Errorf("matches[1].X = %d, want 12", matches[1].X)
	}
}

func TestFindTextInRow_NoMatch(t *testing.T) {
	buf := NewBuffer(20, 1)
	buf.DrawText(0, 0, "hello world", DefaultStyle)

	matches := FindTextInRow(buf, 0, "xyz")
	if len(matches) != 0 {
		t.Errorf("len(matches) = %d, want 0", len(matches))
	}
}

func TestFindTextInRow_EmptyQuery(t *testing.T) {
	buf := NewBuffer(20, 1)
	buf.DrawText(0, 0, "hello", DefaultStyle)
	matches := FindTextInRow(buf, 0, "")
	if len(matches) != 0 {
		t.Errorf("len(matches) = %d, want 0 for empty query", len(matches))
	}
}

func TestFindTextInRow_OutOfBounds(t *testing.T) {
	buf := NewBuffer(10, 2)
	matches := FindTextInRow(buf, -1, "test")
	if matches != nil {
		t.Error("should return nil for negative y")
	}
	matches = FindTextInRow(buf, 5, "test")
	if matches != nil {
		t.Error("should return nil for out-of-bounds y")
	}
}

func TestFindTextInBuffer_Basic(t *testing.T) {
	buf := NewBuffer(20, 3)
	buf.DrawText(0, 0, "hello world", DefaultStyle)
	buf.DrawText(0, 1, "goodbye world", DefaultStyle)
	buf.DrawText(0, 2, "hello again", DefaultStyle)

	matches := FindTextInBuffer(buf, "hello")
	if len(matches) != 2 {
		t.Fatalf("len(matches) = %d, want 2", len(matches))
	}
	if matches[0].Y != 0 || matches[1].Y != 2 {
		t.Errorf("match rows = %d, %d, want 0, 2", matches[0].Y, matches[1].Y)
	}
}

func TestFindTextInBuffer_EmptyQuery(t *testing.T) {
	buf := NewBuffer(10, 1)
	matches := FindTextInBuffer(buf, "")
	if matches != nil {
		t.Error("should return nil for empty query")
	}
}

func TestFindTextInBuffer_OverlappingQuery(t *testing.T) {
	buf := NewBuffer(20, 1)
	buf.DrawText(0, 0, "aaaa", DefaultStyle)
	// "aa" matches at positions 0, 2 (non-overlapping)
	matches := FindTextInBuffer(buf, "aa")
	if len(matches) != 2 {
		t.Errorf("len(matches) = %d, want 2", len(matches))
	}
}

func TestIndexFromString_Basic(t *testing.T) {
	s := "hello world"
	idx := indexFromString(s, "world", 0)
	if idx != 6 {
		t.Errorf("indexFromString = %d, want 6", idx)
	}
}

func TestIndexFromString_WithStart(t *testing.T) {
	s := "foo bar foo"
	idx := indexFromString(s, "foo", 4)
	if idx != 8 {
		t.Errorf("indexFromString = %d, want 8", idx)
	}
}

func TestIndexFromString_NotFound(t *testing.T) {
	idx := indexFromString("hello", "xyz", 0)
	if idx != -1 {
		t.Errorf("indexFromString = %d, want -1", idx)
	}
}

func TestIndexFromString_EmptySub(t *testing.T) {
	idx := indexFromString("hello", "", 0)
	if idx != 0 {
		t.Errorf("indexFromString = %d, want 0", idx)
	}
}

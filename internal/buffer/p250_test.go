package buffer

import "testing"

func TestHighlightCurrentMatch_BeyondBuffer_P250(t *testing.T) {
	buf := NewBuffer(5, 2)
	buf.DrawText(0, 0, "hello", Style{})
	// Match at x=3, length=10 → idx will go out of bounds
	matches := []TextMatch{{X: 3, Y: 0, Length: 10}}
	style := Style{Fg: RGB(255, 0, 0)}
	HighlightCurrentMatch(buf, matches, 0, Style{}, style)
	// Should not panic — break when idx < 0
}

func TestFindTextInRow_OutOfBounds_P250(t *testing.T) {
	buf := NewBuffer(5, 2)
	buf.DrawText(0, 0, "hello", Style{})
	// y < 0
	if m := FindTextInRow(buf, -1, "ll"); m != nil {
		t.Error("y<0 should return nil")
	}
	// y >= Height
	if m := FindTextInRow(buf, 10, "ll"); m != nil {
		t.Error("y>=Height should return nil")
	}
}

func TestFindTextInRow_EmptyQuery_P250(t *testing.T) {
	buf := NewBuffer(5, 1)
	buf.DrawText(0, 0, "hello", Style{})
	if m := FindTextInRow(buf, 0, ""); m != nil {
		t.Error("empty query should return nil")
	}
}

func TestFindTextInRow_MultipleMatches_P250(t *testing.T) {
	buf := NewBuffer(10, 1)
	buf.DrawText(0, 0, "ababab", Style{})
	matches := FindTextInRow(buf, 0, "ab")
	if len(matches) != 3 {
		t.Errorf("expected 3 matches, got %d", len(matches))
	}
}

func TestFindTextInRow_MatchAtEnd_P250(t *testing.T) {
	buf := NewBuffer(6, 1)
	buf.DrawText(0, 0, "xxabab", Style{})
	matches := FindTextInRow(buf, 0, "ab")
	// Last "ab" at index 4 → searchStart=6 >= len("xxabab")=6 → break
	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}
}

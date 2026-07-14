package buffer

import (
	"testing"
)

// === appendFG (92.9% → 100%) ===

func TestP186_AppendFG_AllTypes(t *testing.T) {
	tests := []struct {
		name string
		c    Color
		want string
	}{
		{"None", Color{}, "39"},
		{"Named0", NamedColor(NamedBlack), "30"},
		{"Named7", NamedColor(NamedWhite), "37"},
		{"Named8", NamedColor(NamedBrightBlack), "90"},
		{"Named15", NamedColor(NamedBrightWhite), "97"},
		{"256", Color256Val(42), "38;5;42"},
		{"True", RGB(255, 128, 0), "38;2;255;128;0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(tt.c.appendFG(nil))
			if result != tt.want {
				t.Errorf("appendFG(%s): expected %q, got %q", tt.name, tt.want, result)
			}
		})
	}
}

func TestP186_AppendFG_UnknownType(t *testing.T) {
	c := Color{Type: 99, Val: 0}
	result := string(c.appendFG(nil))
	if result != "39" {
		t.Errorf("expected '39' for unknown type, got %q", result)
	}
}

// === appendBG (92.9% → 100%) ===

func TestP186_AppendBG_AllTypes(t *testing.T) {
	tests := []struct {
		name string
		c    Color
		want string
	}{
		{"None", Color{}, "49"},
		{"Named0", NamedColor(NamedBlack), "40"},
		{"Named7", NamedColor(NamedWhite), "47"},
		{"Named8", NamedColor(NamedBrightBlack), "100"},
		{"Named15", NamedColor(NamedBrightWhite), "107"},
		{"256", Color256Val(42), "48;5;42"},
		{"True", RGB(255, 128, 0), "48;2;255;128;0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(tt.c.appendBG(nil))
			if result != tt.want {
				t.Errorf("appendBG(%s): expected %q, got %q", tt.name, tt.want, result)
			}
		})
	}
}

func TestP186_AppendBG_UnknownType(t *testing.T) {
	c := Color{Type: 99, Val: 0}
	result := string(c.appendBG(nil))
	if result != "49" {
		t.Errorf("expected '49' for unknown type, got %q", result)
	}
}

// === Color.String (83% → 100%) ===

func TestP186_ColorString_AllTypes(t *testing.T) {
	tests := []struct {
		name string
		c    Color
	}{
		{"None", Color{}},
		{"Named", NamedColor(NamedRed)},
		{"NamedBright", NamedColor(NamedBrightRed)},
		{"256", Color256Val(42)},
		{"True", RGB(255, 128, 0)},
		{"Unknown", Color{Type: 99}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.c.String()
			_ = s // just ensure no panic
		})
	}
}

// === HighlightCurrentMatch (92.3% → 100%) ===

func TestP186_HighlightCurrentMatch_NilMatches(t *testing.T) {
	buf := NewBuffer(10, 3)
	HighlightCurrentMatch(buf, nil, 0, Style{}, Style{})
}

func TestP186_HighlightCurrentMatch_NegativeCurrent(t *testing.T) {
	buf := NewBuffer(10, 3)
	buf.DrawText(0, 0, "hello world", Style{})
	matches := []TextMatch{{X: 0, Y: 0, Length: 5}}
	HighlightCurrentMatch(buf, matches, -1, Style{Fg: NamedColor(NamedRed)}, Style{Fg: NamedColor(NamedYellow)})
}

func TestP186_HighlightCurrentMatch_OutOfBounds(t *testing.T) {
	buf := NewBuffer(10, 3)
	buf.DrawText(0, 0, "hello", Style{})
	matches := []TextMatch{{X: 0, Y: 0, Length: 5}}
	// Match extends beyond buffer width
	HighlightCurrentMatch(buf, matches, 0, Style{Fg: NamedColor(NamedRed)}, Style{Fg: NamedColor(NamedYellow)})
}

// === FindTextInRow (90.0% → 100%) ===

func TestP186_FindTextInRow_EmptyQuery(t *testing.T) {
	buf := NewBuffer(10, 3)
	buf.DrawText(0, 0, "hello", Style{})
	matches := FindTextInRow(buf, 0, "")
	if matches != nil {
		t.Error("expected nil for empty query")
	}
}

func TestP186_FindTextInRow_NegativeY(t *testing.T) {
	buf := NewBuffer(10, 3)
	matches := FindTextInRow(buf, -1, "test")
	if matches != nil {
		t.Error("expected nil for negative y")
	}
}

func TestP186_FindTextInRow_YOutOfBounds(t *testing.T) {
	buf := NewBuffer(10, 3)
	matches := FindTextInRow(buf, 10, "test")
	if matches != nil {
		t.Error("expected nil for y out of bounds")
	}
}

func TestP186_FindTextInRow_Match(t *testing.T) {
	buf := NewBuffer(20, 3)
	buf.DrawText(0, 0, "hello world hello", Style{})
	matches := FindTextInRow(buf, 0, "hello")
	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}
}

func TestP186_FindTextInRow_NoMatch(t *testing.T) {
	buf := NewBuffer(20, 3)
	buf.DrawText(0, 0, "hello world", Style{})
	matches := FindTextInRow(buf, 0, "xyz")
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

// === indexFromString (94.7% → 100%) ===

func TestP186_IndexFromString_Found(t *testing.T) {
	idx := indexFromString("hello world", "world", 0)
	if idx != 6 {
		t.Errorf("expected 6, got %d", idx)
	}
}

func TestP186_IndexFromString_NotFound(t *testing.T) {
	idx := indexFromString("hello world", "xyz", 0)
	if idx != -1 {
		t.Errorf("expected -1, got %d", idx)
	}
}

func TestP186_IndexFromString_EmptySub(t *testing.T) {
	idx := indexFromString("hello", "", 3)
	if idx != 3 {
		t.Errorf("expected 3 (start), got %d", idx)
	}
}

func TestP186_IndexFromString_StartBeyond(t *testing.T) {
	idx := indexFromString("hello", "hello", 10)
	if idx != -1 {
		t.Errorf("expected -1 for start beyond string, got %d", idx)
	}
}

func TestP186_IndexFromString_NegativeStart(t *testing.T) {
	idx := indexFromString("hello", "hello", -5)
	if idx != 0 {
		t.Errorf("expected 0 for negative start, got %d", idx)
	}
}

// === DrawTextClamped (93% → 100%) ===

func TestP186_DrawTextClamped_ExactWidth(t *testing.T) {
	buf := NewBuffer(5, 1)
	buf.DrawTextClamped(0, 0, "hello", Style{})
	if buf.GetCell(4, 0).Rune != 'o' {
		t.Error("expected 'o' at position 4")
	}
}

func TestP186_DrawTextClamped_Overflow(t *testing.T) {
	buf := NewBuffer(3, 1)
	buf.DrawTextClamped(0, 0, "hello", Style{})
	if buf.GetCell(2, 0).Rune != 'l' {
		t.Error("expected 'l' at position 2")
	}
}

func TestP186_DrawTextClamped_Empty(t *testing.T) {
	buf := NewBuffer(5, 1)
	buf.DrawTextClamped(0, 0, "", Style{})
}

func TestP186_DrawTextClamped_PastEdge(t *testing.T) {
	buf := NewBuffer(5, 1)
	buf.DrawTextClamped(6, 0, "hello", Style{})
}

func TestP186_DrawTextClamped_PartialFit(t *testing.T) {
	buf := NewBuffer(5, 1)
	buf.DrawTextClamped(3, 0, "hello", Style{})
	if buf.GetCell(3, 0).Rune != 'h' {
		t.Error("expected 'h' at position 3")
	}
	if buf.GetCell(4, 0).Rune != 'e' {
		t.Error("expected 'e' at position 4")
	}
}

// === Cell.Equal (91% → 100%) ===

func TestP186_CellEqual_AllFields(t *testing.T) {
	c := Cell{Rune: 'A', Width: 1, Fg: NamedColor(NamedRed), Bg: NamedColor(NamedBlue), Flags: Bold}
	// Test each field difference
	if c.Equal(Cell{Rune: 'B', Width: 1, Fg: NamedColor(NamedRed), Bg: NamedColor(NamedBlue), Flags: Bold}) {
		t.Error("expected false for different Rune")
	}
	if c.Equal(Cell{Rune: 'A', Width: 2, Fg: NamedColor(NamedRed), Bg: NamedColor(NamedBlue), Flags: Bold}) {
		t.Error("expected false for different Width")
	}
	if c.Equal(Cell{Rune: 'A', Width: 1, Fg: NamedColor(NamedBlue), Bg: NamedColor(NamedBlue), Flags: Bold}) {
		t.Error("expected false for different Fg")
	}
	if c.Equal(Cell{Rune: 'A', Width: 1, Fg: NamedColor(NamedRed), Bg: NamedColor(NamedRed), Flags: Bold}) {
		t.Error("expected false for different Bg")
	}
	if c.Equal(Cell{Rune: 'A', Width: 1, Fg: NamedColor(NamedRed), Bg: NamedColor(NamedBlue), Flags: Italic}) {
		t.Error("expected false for different Flags")
	}
	if !c.Equal(c) {
		t.Error("expected true for identical cell")
	}
}

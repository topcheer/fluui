package buffer

import (
	"testing"
)

// === Color.String (83.3% → 100%) — unknown type branch ===

func TestP141_ColorString_UnknownType(t *testing.T) {
	c := Color{Type: 99} // unknown type
	got := c.String()
	if got != "?" {
		t.Errorf("expected '?' for unknown type, got %q", got)
	}
}

// === appendFG/appendBG (92.9% → 100%) ===

func TestP141_AppendFG_ColorNone(t *testing.T) {
	c := Color{Type: ColorNone}
	b := c.appendFG(nil)
	if len(b) == 0 {
		t.Error("expected non-empty for ColorNone FG")
	}
}

func TestP141_AppendBG_ColorNone(t *testing.T) {
	c := Color{Type: ColorNone}
	b := c.appendBG(nil)
	if len(b) == 0 {
		t.Error("expected non-empty for ColorNone BG")
	}
}

func TestP141_AppendFG_ColorNamed(t *testing.T) {
	c := Color{Type: ColorNamed, Val: 3}
	b := c.appendFG(nil)
	if len(b) == 0 {
		t.Error("expected non-empty for ColorNamed FG")
	}
}

func TestP141_AppendBG_ColorNamed(t *testing.T) {
	c := Color{Type: ColorNamed, Val: 5}
	b := c.appendBG(nil)
	if len(b) == 0 {
		t.Error("expected non-empty for ColorNamed BG")
	}
}

// === HighlightCurrentMatch (92.3% → 100%) ===

func TestP141_HighlightCurrentMatch_EdgeCases(t *testing.T) {
	buf := NewBuffer(20, 3)
	buf.DrawText(0, 0, "hello world test", Style{Fg: RGB(255, 255, 255)})

	matches := []TextMatch{{X: 0, Y: 0, Length: 5}}
	HighlightCurrentMatch(buf, matches, -1, DefaultStyle, Style{Flags: Reverse})
	HighlightCurrentMatch(buf, matches, 0, DefaultStyle, Style{Flags: Reverse})
}

// === FindTextInRow (90.0% → 100%) ===

func TestP141_FindTextInRow_EdgeCases(t *testing.T) {
	buf := NewBuffer(20, 1)
	buf.DrawText(0, 0, "hello world", Style{Fg: RGB(255, 255, 255)})

	matches := FindTextInRow(buf, 0, "")
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for empty query, got %d", len(matches))
	}

	matches = FindTextInRow(buf, 0, "hello")
	if len(matches) != 1 {
		t.Errorf("expected 1 match for hello, got %d", len(matches))
	}

	matches = FindTextInRow(buf, 0, "xyz")
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for xyz, got %d", len(matches))
	}
}

// === indexFromString (94.7% → 100%) ===

func TestP141_IndexFromString_EdgeCases(t *testing.T) {
	idx := indexFromString("hello", "h", 0)
	if idx != 0 {
		t.Errorf("expected 0, got %d", idx)
	}

	idx = indexFromString("hello", "z", 0)
	if idx != -1 {
		t.Errorf("expected -1, got %d", idx)
	}

	idx = indexFromString("", "a", 0)
	if idx != -1 {
		t.Errorf("expected -1 for empty haystack, got %d", idx)
	}

	idx = indexFromString("hello", "l", 3)
	if idx != 3 {
		t.Errorf("expected 3, got %d", idx)
	}
}

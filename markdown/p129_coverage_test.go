package markdown

import (
	"testing"
)

// === NewHighlighterWithStyle (75% → 90%+) ===

func TestP129_NewHighlighterWithStyle_Empty(t *testing.T) {
	h := NewHighlighterWithStyle("")
	if h == nil {
		t.Fatal("expected non-nil highlighter")
	}
	if h.style == nil {
		t.Error("expected fallback to dracula style")
	}
}

func TestP129_NewHighlighterWithStyle_Unknown(t *testing.T) {
	h := NewHighlighterWithStyle("nonexistent-style-xyz")
	if h == nil {
		t.Fatal("expected non-nil highlighter")
	}
	if h.style == nil {
		t.Error("expected fallback to dracula style")
	}
}

func TestP129_NewHighlighterWithStyle_Monokai(t *testing.T) {
	h := NewHighlighterWithStyle("monokai")
	if h == nil {
		t.Fatal("expected non-nil highlighter")
	}
	if h.style == nil {
		t.Error("expected non-nil monokai style")
	}
}

func TestP129_NewHighlighterWithStyle_Dracula(t *testing.T) {
	h := NewHighlighterWithStyle("dracula")
	if h == nil {
		t.Fatal("expected non-nil highlighter")
	}
	if h.style == nil {
		t.Error("expected non-nil dracula style")
	}
}

func TestP129_NewHighlighterWithStyle_HighlightAfterInit(t *testing.T) {
	h := NewHighlighterWithStyle("monokai")
	lines, err := h.Highlight("fmt.Println(\"hello\")", "go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) == 0 {
		t.Error("expected at least one line")
	}
}

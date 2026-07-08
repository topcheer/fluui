package markdown

import (
	"testing"

	"github.com/alecthomas/chroma/styles"
)

// === NewHighlighterWithStyle (75% → 90%+) ===

func TestP135_NewHighlighterWithStyle_Empty(t *testing.T) {
	// Empty string → styles.Get("") returns nil → dracula fallback
	h := NewHighlighterWithStyle("")
	if h == nil {
		t.Fatal("expected non-nil highlighter")
	}
	if h.style == nil {
		t.Error("expected non-nil style (dracula fallback)")
	}
}

func TestP135_NewHighlighterWithStyle_UnknownStyle(t *testing.T) {
	// Unknown style → styles.Get returns nil → dracula fallback
	h := NewHighlighterWithStyle("nonexistent-style-xyz")
	if h == nil {
		t.Fatal("expected non-nil highlighter")
	}
	if h.style == nil {
		t.Error("expected non-nil style (dracula fallback)")
	}
}

func TestP135_NewHighlighterWithStyle_DraculaDirect(t *testing.T) {
	h := NewHighlighterWithStyle("dracula")
	if h == nil {
		t.Fatal("expected non-nil highlighter")
	}
	expected := styles.Get("dracula")
	if h.style != expected {
		t.Error("expected dracula style")
	}
}

func TestP135_NewHighlighterWithStyle_Monokai(t *testing.T) {
	h := NewHighlighterWithStyle("monokai")
	if h == nil {
		t.Fatal("expected non-nil highlighter")
	}
	if h.style == nil {
		t.Error("expected non-nil monokai style")
	}
}

func TestP135_NewHighlighterWithStyle_HighlightAfterInit(t *testing.T) {
	// Verify highlighter works after init with fallback
	h := NewHighlighterWithStyle("")
	lines, err := h.Highlight("fmt.Println(\"test\")", "go")
	if err != nil {
		t.Fatalf("highlight error: %v", err)
	}
	if len(lines) == 0 {
		t.Error("expected at least one line")
	}
}

package markdown

import (
	"testing"

	"github.com/alecthomas/chroma"
)

// P242: tokenTypeColor operator/punctuation/builtin/class + Highlight fallback/error

func TestTokenTypeColor_Operator_P242(t *testing.T) {
	style := chroma.MustNewStyle("test", chroma.StyleEntries{})
	c := tokenTypeColor(chroma.Operator, style)
	if c.Type == 0 {
		t.Error("Operator should have a color")
	}
}

func TestTokenTypeColor_Punctuation_P242(t *testing.T) {
	style := chroma.MustNewStyle("test", chroma.StyleEntries{})
	c := tokenTypeColor(chroma.Punctuation, style)
	if c.Type == 0 {
		t.Error("Punctuation should have a color")
	}
}

func TestTokenTypeColor_NameBuiltin_P242(t *testing.T) {
	style := chroma.MustNewStyle("test", chroma.StyleEntries{})
	c := tokenTypeColor(chroma.NameBuiltin, style)
	if c.Type == 0 {
		t.Error("NameBuiltin should have a color")
	}
}

func TestTokenTypeColor_GenericInserted_P242(t *testing.T) {
	style := chroma.MustNewStyle("test", chroma.StyleEntries{})
	c := tokenTypeColor(chroma.GenericInserted, style)
	if c.Type == 0 {
		t.Error("GenericInserted should have a color")
	}
}

func TestTokenTypeColor_GenericDeleted_P242(t *testing.T) {
	style := chroma.MustNewStyle("test", chroma.StyleEntries{})
	c := tokenTypeColor(chroma.GenericDeleted, style)
	if c.Type == 0 {
		t.Error("GenericDeleted should have a color")
	}
}

func TestTokenTypeColor_Default_P242(t *testing.T) {
	style := chroma.MustNewStyle("test", chroma.StyleEntries{})
	// Use a token type not covered by any case
	c := tokenTypeColor(chroma.GenericOutput, style)
	if c.Type != 0 {
		t.Error("unknown token type should return default ColorNone")
	}
}

func TestHighlight_UnknownLang_P242(t *testing.T) {
	h := NewHighlighter()
	// Unknown language → fallback lexer
	lines, err := h.Highlight("hello world", "totally-unknown-lang-xyz")
	if err != nil {
		t.Errorf("fallback should not error: %v", err)
	}
	if len(lines) == 0 {
		t.Error("should produce at least 1 line")
	}
}

func TestHighlight_EmptyLang_P242(t *testing.T) {
	h := NewHighlighter()
	lines, err := h.Highlight("hello", "")
	if err != nil {
		t.Errorf("empty lang should not error: %v", err)
	}
	if len(lines) == 0 {
		t.Error("should produce at least 1 line")
	}
}

func TestHighlightToLines_P242(t *testing.T) {
	h := NewHighlighter()
	lines, err := h.HighlightToLines("package main\nfunc foo() {}", "go")
	if err != nil {
		t.Errorf("HighlightToLines should not error: %v", err)
	}
	if len(lines) < 2 {
		t.Errorf("should produce at least 2 lines, got %d", len(lines))
	}
}

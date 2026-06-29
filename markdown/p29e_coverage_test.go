package markdown

import (
	"testing"
)

// P29e coverage tests for markdown highlighting via public API.

func TestP29e_NewHighlighterWithStyle_Invalid(t *testing.T) {
	h := NewHighlighterWithStyle("nonexistent-style-xyz")
	if h == nil {
		t.Fatal("should fallback to dracula")
	}
}

func TestP29e_NewHighlighterWithStyle_Valid(t *testing.T) {
	h := NewHighlighterWithStyle("monokai")
	if h == nil {
		t.Fatal("should return highlighter")
	}
}

func TestP29e_Highlight_Plaintext(t *testing.T) {
	h := NewHighlighter()
	lines, err := h.Highlight("hello world", "plaintext")
	if err != nil {
		t.Fatalf("Highlight plaintext: %v", err)
	}
	if len(lines) == 0 {
		t.Error("should return at least one line")
	}
}

func TestP29e_Highlight_Empty(t *testing.T) {
	h := NewHighlighter()
	lines, err := h.Highlight("", "")
	if err != nil {
		t.Fatalf("Highlight empty: %v", err)
	}
	if len(lines) == 0 {
		t.Error("should return at least one line")
	}
}

func TestP29e_Highlight_Go(t *testing.T) {
	h := NewHighlighter()
	code := "package main\nfunc main() {\n\tx := 42\n}"
	lines, err := h.Highlight(code, "go")
	if err != nil {
		t.Fatalf("Highlight go: %v", err)
	}
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines, got %d", len(lines))
	}
}

func TestP29e_Highlight_UnknownLang(t *testing.T) {
	h := NewHighlighter()
	lines, err := h.Highlight("test code", "totally-made-up-language")
	if err != nil {
		t.Fatalf("Highlight unknown lang: %v", err)
	}
	_ = lines
}

func TestP29e_Highlight_Python(t *testing.T) {
	h := NewHighlighter()
	code := "def hello():\n\tprint('world')"
	lines, err := h.Highlight(code, "python")
	if err != nil {
		t.Fatalf("Highlight python: %v", err)
	}
	if len(lines) < 2 {
		t.Errorf("expected at least 2 lines, got %d", len(lines))
	}
}

func TestP29e_Highlight_JavaScript(t *testing.T) {
	h := NewHighlighter()
	code := "const x = () => 42;"
	lines, err := h.Highlight(code, "javascript")
	if err != nil {
		t.Fatalf("Highlight js: %v", err)
	}
	if len(lines) == 0 {
		t.Error("should return at least one line")
	}
}

// === FormatOSC8 ===

func TestP29e_FormatOSC8(t *testing.T) {
	result := FormatOSC8("https://example.com", "Example")
	if result == "" {
		t.Error("FormatOSC8 should return non-empty")
	}
}

func TestP29e_FormatOSC8_EmptyURL(t *testing.T) {
	result := FormatOSC8("", "text")
	if result == "" {
		t.Error("FormatOSC8 should still return text even with empty URL")
	}
}

// === NewMarkdownRenderer edge cases ===

func TestP29e_NewMarkdownRenderer_ZeroWidth(t *testing.T) {
	r := NewMarkdownRenderer(nil, 0)
	if r == nil {
		t.Fatal("should return renderer")
	}
}

func TestP29e_NewMarkdownRenderer_DefaultTheme(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	if r == nil {
		t.Fatal("should return renderer with nil theme")
	}
}

// === Renderer: complex markdown ===

func TestP29e_Render_NestedList(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	input := "- Item 1\n  - Sub-item 1\n  - Sub-item 2\n- Item 2"
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatalf("Render nested list: %v", err)
	}
	if len(blocks) == 0 {
		t.Error("should return blocks")
	}
}

func TestP29e_Render_MultiParagraph(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	input := "First paragraph.\n\nSecond paragraph.\n\nThird."
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatalf("Render multi-paragraph: %v", err)
	}
	if len(blocks) < 3 {
		t.Errorf("expected at least 3 blocks, got %d", len(blocks))
	}
}

func TestP29e_Render_CodeBlock(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	input := "```go\nfunc main() {}\n```"
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatalf("Render code block: %v", err)
	}
	if len(blocks) == 0 {
		t.Error("should return blocks")
	}
}

func TestP29e_Render_Blockquote(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	input := "> A quote\n> More text"
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatalf("Render blockquote: %v", err)
	}
	_ = blocks
}

func TestP29e_Render_Heading(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	input := "# H1\n## H2\n### H3"
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatalf("Render headings: %v", err)
	}
	if len(blocks) < 3 {
		t.Errorf("expected at least 3 blocks, got %d", len(blocks))
	}
}

func TestP29e_Render_Table(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	input := "| A | B |\n|---|---|\n| 1 | 2 |"
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatalf("Render table: %v", err)
	}
	_ = blocks
}

func TestP29e_Render_Empty(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("")
	if err != nil {
		t.Fatalf("Render empty: %v", err)
	}
	_ = blocks
}

func TestP29e_Render_BoldItalic(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	input := "**bold** and *italic*"
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatalf("Render bold/italic: %v", err)
	}
	_ = blocks
}

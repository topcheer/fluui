package markdown

import (
	"testing"
)

func TestP69_HighlightToLines_Error(t *testing.T) {
	h := NewHighlighterWithStyle("dracula")
	// Invalid language should fall back gracefully
	lines, err := h.HighlightToLines("", "some code")
	// Should not return error (falls back to plaintext)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	_ = lines
}

func TestP69_HighlightToLines_Empty(t *testing.T) {
	h := NewHighlighter()
	lines, err := h.HighlightToLines("go", "")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(lines) > 1 {
		t.Errorf("expected at most 1 line for empty code, got %d", len(lines))
	}
}

func TestP69_Latex_advance(t *testing.T) {
	// Test latex parser advance function edge cases
	result := RenderLatexMath("\\alpha + 1")
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP69_Latex_advance_Newline(t *testing.T) {
	// Test \\ (newline)
	result := RenderLatexMath("a \\\\ b")
	_ = result // should not panic
}

func TestP69_Latex_advance_StarCommand(t *testing.T) {
	// Test \sum* star suffix
	result := RenderLatexMath("\\sum*_{i=1}^{n}")
	_ = result
}

func TestP69_FindTextInBuffer_Package(t *testing.T) {
	// Test that markdown rendering produces searchable text
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("# Hello World\n\nThis is searchable text.")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Error("expected non-empty blocks")
	}
}

func TestP69_Mermaid_MultiNode(t *testing.T) {
	// Test mermaid with multiple nodes and edges
	input := `flowchart TD
A[Start] --> B{Decision}
B -->|Yes| C[Do it]
B -->|No| D[Skip]
C --> E[End]
D --> E`

	graph := ParseMermaid(input)
	if graph == nil {
		t.Fatal("expected non-nil graph")
	}
	if len(graph.Nodes) < 5 {
		t.Errorf("expected at least 5 nodes, got %d", len(graph.Nodes))
	}
	if len(graph.Edges) < 5 {
		t.Errorf("expected at least 5 edges, got %d", len(graph.Edges))
	}
}

func TestP69_Mermaid_RenderComplex(t *testing.T) {
	input := `flowchart LR
A --> B
B --> C
A --> C`
	cells := RenderMermaidText(input, DefaultTheme())
	if len(cells) == 0 {
		t.Error("expected non-empty rendered output")
	}
}

func TestP69_RenderInlineMath_NoMatch(t *testing.T) {
	result := RenderInlineMath("just plain text with no math")
	if result != "just plain text with no math" {
		t.Errorf("expected unchanged text, got %q", result)
	}
}

func TestP69_RenderInlineMath_ParenFormat(t *testing.T) {
	result := RenderInlineMath("equation \\(x^2\\) here")
	_ = result // should convert \(...\) format
}

func TestP69_HasInlineMath_DollarFormat(t *testing.T) {
	if !HasInlineMath("equation $x^2$ here") {
		t.Error("expected true for dollar-delimited inline math")
	}
}

func TestP69_HasInlineMath_ParenFormat(t *testing.T) {
	if !HasInlineMath("equation \\(x^2\\) here") {
		t.Error("expected true for paren-delimited inline math")
	}
}

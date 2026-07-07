package markdown

import (
	"testing"
)

func TestP108_NewHighlighterWithStyle(t *testing.T) {
	h := NewHighlighterWithStyle("monokai")
	if h == nil {
		t.Fatal("expected non-nil for monokai")
	}
	h = NewHighlighterWithStyle("")
	if h == nil {
		t.Fatal("expected non-nil for empty style")
	}
	h = NewHighlighterWithStyle("nonexistent-style-xyz")
	if h == nil {
		t.Fatal("expected non-nil for unknown style")
	}
}

func TestP108_NewHighlighterWithStyle_DefaultFallback(t *testing.T) {
	h1 := NewHighlighterWithStyle("nonexistent")
	h2 := NewHighlighterWithStyle("dracula")
	l1, _ := h1.Highlight("x = 1", "go")
	l2, _ := h2.Highlight("x = 1", "go")
	if len(l1) != len(l2) {
		t.Errorf("expected same line count for unknown vs dracula: %d vs %d", len(l1), len(l2))
	}
}

func TestP108_NewHighlighterWithStyle_Dracula(t *testing.T) {
	h := NewHighlighterWithStyle("dracula")
	lines, _ := h.Highlight("func main() {}", "go")
	if len(lines) == 0 {
		t.Error("expected highlighted output")
	}
}

func TestP108_mermaidDrawEdge_AllDirections(t *testing.T) {
	for _, dir := range []string{"TD", "BT", "LR", "RL"} {
		g := ParseMermaid("graph " + dir + "\n  A-->B")
		if g == nil {
			continue
		}
		cells := RenderMermaid(g, DefaultTheme())
		if len(cells) == 0 {
			t.Errorf("direction %s: expected non-empty rendering", dir)
		}
	}
}

func TestP108_mermaidDrawEdge_AllStyles(t *testing.T) {
	for _, syntax := range []string{"A-->B", "A---B", "A-.->B", "A==>B"} {
		g := ParseMermaid("graph TD\n  " + syntax)
		if g == nil {
			continue
		}
		cells := RenderMermaid(g, DefaultTheme())
		if len(cells) == 0 {
			t.Errorf("syntax %s: expected non-empty rendering", syntax)
		}
	}
}

func TestP108_mermaidDrawEdge_WithLabel(t *testing.T) {
	g := ParseMermaid("graph TD\n  A-->|label|B")
	if g == nil {
		t.Skip("could not parse mermaid with label")
	}
	cells := RenderMermaid(g, DefaultTheme())
	if len(cells) == 0 {
		t.Error("expected non-empty rendering for labeled edge")
	}
}

func TestP108_mermaidDrawEdge_ReverseDirection(t *testing.T) {
	g := ParseMermaid("graph LR\n  B-->A")
	if g == nil {
		t.Skip("could not parse")
	}
	cells := RenderMermaid(g, DefaultTheme())
	_ = cells
}

func TestP108_mermaidDrawEdge_FromYGreaterThanToY(t *testing.T) {
	g := ParseMermaid("graph TD\n  B-->A")
	if g == nil {
		t.Skip("could not parse")
	}
	cells := RenderMermaid(g, DefaultTheme())
	_ = cells
}

func TestP108_HasInlineMath(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"$x^2$", true},
		{"\\(x^2\\)", true},
		{"no math here", false},
		{"", false},
		{"escaped \\$ dollar", false},
	}
	for _, tc := range tests {
		got := HasInlineMath(tc.input)
		if got != tc.want {
			t.Errorf("HasInlineMath(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestP108_RenderInlineMath(t *testing.T) {
	got := RenderInlineMath("The area is $\\pi r^2$")
	if got == "" {
		t.Error("expected non-empty result")
	}
	got = RenderInlineMath("Equation \\(x + 1\\)")
	if got == "" {
		t.Error("expected non-empty result")
	}
	got = RenderInlineMath("no math")
	if got != "no math" {
		t.Errorf("expected passthrough, got %q", got)
	}
}

func TestP108_RenderMathToCells(t *testing.T) {
	cells := RenderMathToCells("\\alpha + \\beta", DefaultTheme().CodeFg)
	if len(cells) == 0 {
		t.Error("expected non-empty cells")
	}
	cells = RenderMathToCells("", DefaultTheme().CodeFg)
	if len(cells) != 0 {
		t.Errorf("expected 0 cells, got %d", len(cells))
	}
}

func TestP108_FormatOSC8(t *testing.T) {
	s := FormatOSC8("text", "https://example.com")
	if s == "" {
		t.Error("expected non-empty OSC8 sequence")
	}
	s = FormatOSC8("", "")
	if s != "" {
		t.Errorf("expected empty for empty URL, got %q", s)
	}
}

func TestP108_renderFencedCode_MermaidBlock(t *testing.T) {
	md := NewMarkdownRenderer(nil, 80)
	blocks, err := md.Render("```mermaid\ngraph TD\n  A-->B\n```")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if len(blocks) == 0 {
		t.Error("expected at least 1 block")
	}
}

func TestP108_renderFencedCode_MathBlock(t *testing.T) {
	md := NewMarkdownRenderer(nil, 80)
	blocks, err := md.Render("```math\nx = \\frac{a}{b}\n```")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if len(blocks) == 0 {
		t.Error("expected at least 1 block")
	}
}

func TestP108_renderFencedCode_LatexBlock(t *testing.T) {
	md := NewMarkdownRenderer(nil, 80)
	blocks, err := md.Render("```latex\n\\sqrt{x^2 + y^2}\n```")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if len(blocks) == 0 {
		t.Error("expected at least 1 block")
	}
}

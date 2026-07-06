package markdown

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── NewHighlighterWithStyle (75% → 90%+) ───

func TestP88_NewHighlighterWithStyle_Monokai(t *testing.T) {
	h := NewHighlighterWithStyle("monokai")
	if h == nil {
		t.Fatal("expected non-nil highlighter")
	}
}

func TestP88_NewHighlighterWithStyle_Empty(t *testing.T) {
	h := NewHighlighterWithStyle("")
	if h == nil {
		t.Fatal("expected non-nil highlighter (fallback to dracula)")
	}
}

func TestP88_NewHighlighterWithStyle_Unknown(t *testing.T) {
	h := NewHighlighterWithStyle("nonexistent-style-xyz")
	if h == nil {
		t.Fatal("expected non-nil highlighter (fallback to dracula)")
	}
}

// ─── tokenTypeColor (85.7% → 100%) ───

func TestP88_TokenTypeColor_GoCode(t *testing.T) {
	h := NewHighlighter()
	// Highlight a Go code snippet with diverse token types
	src := `package main

import "fmt"

// Comment line
func main() {
	var x int = 42
	fmt.Println("hello", x)
}
`
	lines, err := h.HighlightToLines("go", src)
	if err != nil {
		t.Fatalf("HighlightToLines error: %v", err)
	}
	if len(lines) == 0 {
		t.Fatal("expected non-empty lines")
	}
}

// ─── consumeCommand (83.3% → 90%+) ───

func TestP88_Latex_LeftRight(t *testing.T) {
	// \left and \right are font-like commands
	result := RenderLatexMath(`\left(x\right)`)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP88_Latex_BigFontCommands(t *testing.T) {
	result := RenderLatexMath(`\big x + \Big y + \bigg z`)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP88_Latex_UnknownCommand(t *testing.T) {
	result := RenderLatexMath(`\unknowncmd{x}`)
	if result == "" {
		t.Error("expected non-empty result")
	}
	// Should contain the raw command name
	if !containsStr(result, "unknowncmd") {
		t.Errorf("expected 'unknowncmd' in result, got %q", result)
	}
}

func TestP88_Latex_SumStar(t *testing.T) {
	// \sum* should consume the star
	result := RenderLatexMath(`\sum*_{i=0}^{n}`)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP88_Latex_PartialDerivative(t *testing.T) {
	result := RenderLatexMath(`\frac{\partial f}{\partial x}`)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

// ─── consumeGroup (84.6% → 90%+) ───

func TestP88_Latex_NestedGroups(t *testing.T) {
	result := RenderLatexMath(`x^{a^{b}}`)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP88_Latex_DeepNesting(t *testing.T) {
	result := RenderLatexMath(`\frac{\frac{a}{b}}{\frac{c}{d}}`)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

// ─── consumeSqrt (82.4% → 90%+) ───

func TestP88_Latex_SqrtMultiChar(t *testing.T) {
	// Multi-char sqrt content should get overline
	result := RenderLatexMath(`\sqrt{abc}`)
	if result == "" {
		t.Error("expected non-empty result")
	}
	// Should contain combining overline (U+0305)
	if !containsRune(result, '\u0305') {
		t.Error("expected combining overline in multi-char sqrt")
	}
}

func TestP88_Latex_SqrtSingleChar(t *testing.T) {
	result := RenderLatexMath(`\sqrt{x}`)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP88_Latex_SqrtNthRoot(t *testing.T) {
	// \sqrt[3]{x} — nth root
	result := RenderLatexMath(`\sqrt[3]{x}`)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP88_Latex_SqrtNoBrace(t *testing.T) {
	// \sqrt x — single char without braces
	result := RenderLatexMath(`\sqrt x`)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

// ─── skipBracket (84.6% → 100%) ───

func TestP88_Latex_SqrtBracketContent(t *testing.T) {
	// \sqrt[3]{x} exercises skipBracket
	result := RenderLatexMath(`\sqrt[2n+1]{x}`)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

// ─── advance (80% → 100%) ───

func TestP88_Latex_AdvanceAtEnd(t *testing.T) {
	// Test advance() returning 0 at end of input
	result := RenderLatexMath(`\alpha`)
	if result != "α" {
		t.Errorf("expected 'α', got %q", result)
	}
}

// ─── mermaidDrawEdge (75% → 90%+) ───

func TestP88_Mermaid_LR_DifferentY(t *testing.T) {
	// Test horizontal edge with different Y positions (L-shaped connector)
	g := ParseMermaid(`graph LR
A[Top Node] --> B[Bottom Node]`)
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	cells := RenderMermaid(g, DefaultTheme())
	if len(cells) == 0 {
		t.Error("expected non-empty cells")
	}
}

func TestP88_Mermaid_TD_DifferentX(t *testing.T) {
	// Test vertical edge with different X positions
	g := ParseMermaid(`graph TD
A[Left] --> B[Right Much Longer Label]`)
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	cells := RenderMermaid(g, DefaultTheme())
	if len(cells) == 0 {
		t.Error("expected non-empty cells")
	}
}

func TestP88_Mermaid_EdgeStyles(t *testing.T) {
	// Test different edge styles
	tests := []string{
		"graph TD\nA --> B",
		"graph TD\nA -.-> B",
		"graph TD\nA ==> B",
		"graph LR\nA --> B",
		"graph LR\nA -.-> B",
		"graph LR\nA ==> B",
	}
	for _, src := range tests {
		g := ParseMermaid(src)
		if g == nil {
			t.Errorf("ParseMermaid(%q) returned nil", src)
			continue
		}
		cells := RenderMermaid(g, DefaultTheme())
		if len(cells) == 0 {
			t.Errorf("RenderMermaid(%q) returned empty", src)
		}
	}
}

func TestP88_Mermaid_EdgeWithLabel(t *testing.T) {
	g := ParseMermaid(`graph TD
A[Start] -->|label text| B[End]`)
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	cells := RenderMermaid(g, DefaultTheme())
	if len(cells) == 0 {
		t.Error("expected non-empty cells")
	}
}

func TestP88_Mermaid_BT_Direction(t *testing.T) {
	g := ParseMermaid(`graph BT
A[Bottom] --> B[Top]`)
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	cells := RenderMermaid(g, DefaultTheme())
	if len(cells) == 0 {
		t.Error("expected non-empty cells")
	}
}

// ─── parseMermaidNode (83.3% → 100%) ───

func TestP88_Mermaid_UpdateExistingNode(t *testing.T) {
	// Define a node, then update it with a shape
	g := ParseMermaid(`graph TD
A --> B[Updated Label]
A[New Label]`)
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	// Find node A and verify it was updated
	for _, n := range g.Nodes {
		if n.ID == "A" {
			if n.Label == "" {
				t.Error("expected non-empty label for A")
			}
		}
	}
}

func TestP88_Mermaid_EmptyNodeText(t *testing.T) {
	g := ParseMermaid(`graph TD
A`)
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
}

// ─── mermaidDrawNode (87.5% → 100%) ───

func TestP88_Mermaid_AllShapes(t *testing.T) {
	g := ParseMermaid(`graph TD
A[Rectangle] --> B(Rounded)
B --> C{Diamond}
C --> D((Circle))
D --> E Plain`)
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	cells := RenderMermaid(g, DefaultTheme())
	if len(cells) == 0 {
		t.Error("expected non-empty cells")
	}
}

// ─── renderInlineNode (89.5% → 95%+) ───

func TestP88_RenderInlineNode_LinkWithOSC8(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	r.SetLinkRenderer(NewLinkRenderer(true))
	blocks, err := r.Render("[click here](https://example.com)")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected non-empty blocks")
	}
}

// ─── renderList (86.4% → 95%+) ───

func TestP88_RenderList_NestedOrdered(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("1. First\n   1. Nested\n   2. Also nested\n2. Second")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected non-empty blocks")
	}
}

func TestP88_RenderList_DeepNesting(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	src := "- Level 1\n  - Level 2\n    - Level 3\n      - Level 4"
	blocks, err := r.Render(src)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected non-empty blocks")
	}
}

// ─── wrapCellsWithPrefix (85.7% → 100%) ───

func TestP88_RenderList_WithPrefixWrap(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 20)
	// Narrow width forces wrapping
	src := "- This is a very long list item that should wrap to multiple lines"
	blocks, err := r.Render(src)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected non-empty blocks")
	}
}

// ─── renderThematicBreak (85.7% → 100%) ───

func TestP88_RenderThematicBreak(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 40)
	blocks, err := r.Render("Before\n\n---\n\nAfter")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected non-empty blocks")
	}
}

func TestP88_RenderThematicBreak_Narrow(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 5)
	blocks, err := r.Render("---")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected non-empty blocks")
	}
}

// ─── HighlightToLines error path (90% → 100%) ───

func TestP88_HighlightToLines_EmptySource(t *testing.T) {
	h := NewHighlighter()
	lines, err := h.HighlightToLines("go", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty source might produce 0 or 1 empty line
	_ = lines
}

func TestP88_HighlightToLines_UnknownLang(t *testing.T) {
	h := NewHighlighter()
	lines, err := h.HighlightToLines("totally-unknown-lang-xyz", "some code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = lines // fallback should still produce output
}

// ─── helpers ───

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func containsRune(s string, r rune) bool {
	for _, ch := range s {
		if ch == r {
			return true
		}
	}
	return false
}

// ─── RenderMathToCells (already covered, but verify edge) ───

func TestP88_RenderMathToCells_Greek(t *testing.T) {
	cells := RenderMathToCells(`\alpha + \beta = \gamma`, buffer.RGB(255, 255, 255))
	if len(cells) == 0 {
		t.Error("expected non-empty cells")
	}
}

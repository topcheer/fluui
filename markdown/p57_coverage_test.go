package markdown

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// --- RenderMathToCells (was 0%) ---

func TestP57_RenderMathToCells_Basic(t *testing.T) {
	cells := RenderMathToCells(`x^2`, buffer.NamedColor(buffer.NamedCyan))
	if len(cells) == 0 {
		t.Fatal("expected non-empty cells")
	}
	if cells[0].Rune != 'x' {
		t.Errorf("expected first cell 'x', got %c", cells[0].Rune)
	}
}

func TestP57_RenderMathToCells_Empty(t *testing.T) {
	cells := RenderMathToCells("", buffer.NamedColor(buffer.NamedCyan))
	if len(cells) != 0 {
		t.Errorf("expected 0 cells for empty input, got %d", len(cells))
	}
}

func TestP57_RenderMathToCells_Greek(t *testing.T) {
	cells := RenderMathToCells(`\alpha + \beta`, buffer.NamedColor(buffer.NamedYellow))
	if len(cells) == 0 {
		t.Fatal("expected non-empty cells")
	}
	// Should contain α (alpha)
	found := false
	for _, c := range cells {
		if c.Rune == 'α' {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find alpha rune in cells")
	}
}

func TestP57_RenderMathToCells_FgColor(t *testing.T) {
	fg := buffer.RGB(0xFF, 0x00, 0x00)
	cells := RenderMathToCells("abc", fg)
	if len(cells) > 0 {
		if cells[0].Fg != fg {
			t.Error("expected foreground color to match")
		}
	}
}

// --- skipBracket (was 0%) — test via \sqrt[n]{x} ---

func TestP57_Latex_SqrtNthRoot(t *testing.T) {
	// \sqrt[3]{x} → √x (should skip the [3] bracket)
	result := RenderLatexMath(`\sqrt[3]{x}`)
	if !strings.Contains(result, "√") {
		t.Errorf("expected √, got %q", result)
	}
}

func TestP57_Latex_SqrtNthRootMulti(t *testing.T) {
	result := RenderLatexMath(`\sqrt[n]{x^n}`)
	if !strings.Contains(result, "√") {
		t.Errorf("expected √, got %q", result)
	}
}

// --- skipSpaces (was 50%) — test with leading spaces in commands ---

func TestP57_Latex_LeadingSpaces(t *testing.T) {
	// \frac with spaces between arguments
	result := RenderLatexMath(`\frac{ a }{ b }`)
	if !strings.Contains(result, "(") {
		t.Errorf("expected ( in frac output, got %q", result)
	}
}

// --- consumeCommand deeper coverage (was 63.6%) ---

func TestP57_Latex_NotSubset(t *testing.T) {
	result := RenderLatexMath(`A \not\subset B`)
	// Should produce ⊄
	if !strings.Contains(result, "⊄") {
		t.Errorf("expected ⊄, got %q", result)
	}
}

func TestP57_Latex_NotSupset(t *testing.T) {
	result := RenderLatexMath(`A \not\supset B`)
	if !strings.Contains(result, "⊅") {
		t.Errorf("expected ⊅, got %q", result)
	}
}

func TestP57_Latex_UnknownNotCommand(t *testing.T) {
	result := RenderLatexMath(`\not\foobar`)
	// Should not crash, just produce something
	_ = result
}

func TestP57_Latex_UnknownCommand(t *testing.T) {
	result := RenderLatexMath(`\xyzworld{test}`)
	// Should output the unknown command as-is
	if !strings.Contains(result, "xyzworld") {
		t.Errorf("expected unknown command preserved, got %q", result)
	}
}

func TestP57_Latex_BackslashBackslash(t *testing.T) {
	result := RenderLatexMath(`a \\ b`)
	// \\ = line break → spaces
	if !strings.Contains(result, "a") || !strings.Contains(result, "b") {
		t.Errorf("expected a and b, got %q", result)
	}
}

func TestP57_Latex_SumStar(t *testing.T) {
	// \sum* should be handled (star suffix)
	result := RenderLatexMath(`\sum*_{i=1}^n x_i`)
	if !strings.Contains(result, "Σ") {
		t.Errorf("expected Σ, got %q", result)
	}
}

func TestP57_Latex_LeftRight(t *testing.T) {
	// \left( and \right) should be consumed (font cmds)
	result := RenderLatexMath(`\left( x + y \right)`)
	if !strings.Contains(result, "x") {
		t.Errorf("expected x preserved, got %q", result)
	}
}

func TestP57_Latex_Big(t *testing.T) {
	result := RenderLatexMath(`\big( x \big)`)
	if !strings.Contains(result, "x") {
		t.Errorf("expected x preserved, got %q", result)
	}
}

// --- consumeAccent deeper coverage (was 72.7%) ---

func TestP57_Latex_AccentSingleChar(t *testing.T) {
	// \hat without braces — single char accent
	result := RenderLatexMath(`\hat x`)
	if !strings.Contains(result, "x") {
		t.Errorf("expected x, got %q", result)
	}
	if !strings.Contains(result, "\u0302") {
		t.Errorf("expected combining circumflex, got %q", result)
	}
}

func TestP57_Latex_Dot(t *testing.T) {
	result := RenderLatexMath(`\dot{x}`)
	if !strings.Contains(result, "\u0307") {
		t.Errorf("expected combining dot above, got %q", result)
	}
}

func TestP57_Latex_Ddot(t *testing.T) {
	result := RenderLatexMath(`\ddot{x}`)
	if !strings.Contains(result, "\u0308") {
		t.Errorf("expected combining diaeresis, got %q", result)
	}
}

func TestP57_Latex_Acute(t *testing.T) {
	result := RenderLatexMath(`\acute{x}`)
	if !strings.Contains(result, "\u0301") {
		t.Errorf("expected combining acute, got %q", result)
	}
}

func TestP57_Latex_Grave(t *testing.T) {
	result := RenderLatexMath(`\grave{x}`)
	if !strings.Contains(result, "\u0300") {
		t.Errorf("expected combining grave, got %q", result)
	}
}

func TestP57_Latex_Breve(t *testing.T) {
	result := RenderLatexMath(`\breve{x}`)
	if !strings.Contains(result, "\u0306") {
		t.Errorf("expected combining breve, got %q", result)
	}
}

func TestP57_Latex_Check(t *testing.T) {
	result := RenderLatexMath(`\check{x}`)
	if !strings.Contains(result, "\u030C") {
		t.Errorf("expected combining caron, got %q", result)
	}
}

// --- tokenTypeColor (was 21.4%) — exercise via Highlight with diverse code ---

func TestP57_TokenTypeColor_DiverseCode(t *testing.T) {
	h := NewHighlighter()
	// Go code with diverse token types: keyword, string, comment, number, function, etc.
	code := `// comment
package main

func add(a int, b int) int {
	return a + b // operator
}
const PI = 3.14
`
	cells, err := h.Highlight(code, "go")
	if err != nil {
		t.Fatal(err)
	}
	if len(cells) == 0 {
		t.Fatal("expected highlighted cells")
	}
}

// --- NewHighlighterWithStyle (was 75%) ---

func TestP57_NewHighlighterWithStyle(t *testing.T) {
	h := NewHighlighterWithStyle("monokai")
	if h == nil {
		t.Fatal("expected non-nil highlighter")
	}
}

func TestP57_NewHighlighterWithStyle_Empty(t *testing.T) {
	h := NewHighlighterWithStyle("")
	if h == nil {
		t.Fatal("expected non-nil for empty style (should use default)")
	}
}

func TestP57_NewHighlighterWithStyle_Unknown(t *testing.T) {
	h := NewHighlighterWithStyle("nonexistent-style-12345")
	if h == nil {
		t.Fatal("expected non-nil for unknown style (should use fallback)")
	}
}

// --- mermaidLayout BT/RL directions (was 77.4%) ---

func TestP57_Mermaid_LayoutBT(t *testing.T) {
	g := ParseMermaid("flowchart BT\n    A[Start] --> B[Middle] --> C[End]")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	cells := RenderMermaid(g, nil)
	if cells == nil {
		t.Fatal("expected non-nil cells for BT layout")
	}
}

func TestP57_Mermaid_LayoutRL(t *testing.T) {
	g := ParseMermaid("flowchart RL\n    A[Start] --> B[End]")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	cells := RenderMermaid(g, nil)
	if cells == nil {
		t.Fatal("expected non-nil cells for RL layout")
	}
}

func TestP57_Mermaid_LayoutBTReversed(t *testing.T) {
	// Multi-layer BT should reverse the layout
	g := ParseMermaid("flowchart BT\n    A --> B --> C --> D")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	cells := RenderMermaid(g, nil)
	if cells == nil {
		t.Fatal("expected non-nil cells")
	}
}

func TestP57_Mermaid_LayoutRLReversed(t *testing.T) {
	g := ParseMermaid("flowchart RL\n    A --> B --> C")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	cells := RenderMermaid(g, nil)
	if cells == nil {
		t.Fatal("expected non-nil cells")
	}
}

// --- findMermaidNode edge cases (was 75%) ---

func TestP57_Mermaid_FindNodeMissing(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A --> B")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	// Edge references non-existent node via internal findMermaidNode
	n := findMermaidNode(g, "NONEXISTENT")
	if n != nil {
		t.Error("expected nil for non-existent node")
	}
}

// --- FormatOSC8 (was 66.7%) ---

func TestP57_FormatOSC8_WithURL(t *testing.T) {
	s := FormatOSC8("Click here", "https://example.com")
	if !strings.Contains(s, "Click here") {
		t.Error("expected link text preserved")
	}
	if !strings.Contains(s, "https://example.com") {
		t.Error("expected URL in formatted link")
	}
}

func TestP57_FormatOSC8_EmptyURL(t *testing.T) {
	s := FormatOSC8("Plain text", "")
	if s != "Plain text" {
		t.Errorf("expected plain text for empty URL, got %q", s)
	}
}

func TestP57_LinkRenderer_Disabled(t *testing.T) {
	lr := NewLinkRenderer(false)
	s := lr.FormatLink("Click here", "https://example.com")
	// When disabled, should return "text (url)" format
	if !strings.Contains(s, "(https://example.com)") {
		t.Errorf("expected (url) format when disabled, got %q", s)
	}
}

func TestP57_LinkRenderer_Enabled(t *testing.T) {
	lr := NewLinkRenderer(true)
	s := lr.FormatLink("Click here", "https://example.com")
	if !strings.Contains(s, "\x1b]8") {
		t.Errorf("expected OSC8 escape when enabled, got %q", s)
	}
}

func TestP57_LinkRenderer_EmptyURL(t *testing.T) {
	lr := NewLinkRenderer(true)
	s := lr.FormatLink("Plain text", "")
	if s != "Plain text" {
		t.Errorf("expected plain text for empty URL, got %q", s)
	}
}

// --- Math block rendering in markdown ---

func TestP57_RenderMathBlock(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("```math\nE = mc^2\n```")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
}

func TestP57_RenderLatexBlock(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("```latex\n\\alpha + \\beta\n```")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
}

// --- Complex inline math with text ---

func TestP57_RenderInlineMath_ComplexNested(t *testing.T) {
	text := `The formula $\frac{a^2 + b^2}{c^2}$ gives the value.`
	result := RenderInlineMath(text)
	if !strings.Contains(result, "²") {
		t.Errorf("expected ², got %q", result)
	}
	if !strings.Contains(result, "The formula") {
		t.Errorf("expected surrounding text, got %q", result)
	}
}

// --- Superscript/subscript without Unicode fallback ---

func TestP57_Latex_SuperNoUnicode(t *testing.T) {
	// Character not in Unicode superscript table
	result := RenderLatexMath(`x^@`)
	if !strings.Contains(result, "^") {
		t.Errorf("expected fallback ^, got %q", result)
	}
}

func TestP57_Latex_SubNoUnicode(t *testing.T) {
	// 'q' has no Unicode subscript
	result := RenderLatexMath(`x_q`)
	if !strings.Contains(result, "_") {
		t.Errorf("expected fallback _, got %q", result)
	}
}

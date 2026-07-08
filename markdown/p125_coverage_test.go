package markdown

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// === stripTaskPrefix: no-trailing-space path (71.4% → 100%) ===

func TestP125_StripTaskPrefix_NoSpaceAfter_LongEnough(t *testing.T) {
	// 4+ cells: '[', 'x', ']', 'y' — should strip 3 cells and return cells[3:]
	cells := []buffer.Cell{
		{Rune: '[', Width: 1},
		{Rune: 'x', Width: 1},
		{Rune: ']', Width: 1},
		{Rune: 'y', Width: 1},
	}
	result := stripTaskPrefix(cells)
	if len(result) != 1 || result[0].Rune != 'y' {
		t.Errorf("expected 1 cell 'y', got %d cells", len(result))
	}
}

func TestP125_StripTaskPrefix_Exactly4NoSpace(t *testing.T) {
	cells := []buffer.Cell{
		{Rune: '[', Width: 1},
		{Rune: 'x', Width: 1},
		{Rune: ']', Width: 1},
		{Rune: 'a', Width: 1},
	}
	result := stripTaskPrefix(cells)
	if len(result) != 1 || result[0].Rune != 'a' {
		t.Errorf("expected 1 cell 'a', got %d cells", len(result))
	}
}

func TestP125_StripTaskPrefix_5WithSpace(t *testing.T) {
	cells := []buffer.Cell{
		{Rune: '[', Width: 1},
		{Rune: ' ', Width: 1},
		{Rune: ']', Width: 1},
		{Rune: ' ', Width: 1},
		{Rune: 'X', Width: 1},
	}
	result := stripTaskPrefix(cells)
	if len(result) != 1 || result[0].Rune != 'X' {
		t.Errorf("expected 1 cell 'X', got %d cells", len(result))
	}
}

func TestP125_StripTaskPrefix_3CellsTooShort(t *testing.T) {
	cells := []buffer.Cell{
		{Rune: '[', Width: 1},
		{Rune: 'x', Width: 1},
		{Rune: ']', Width: 1},
	}
	result := stripTaskPrefix(cells)
	if len(result) != 3 {
		t.Errorf("expected unchanged 3 cells (too short), got %d", len(result))
	}
}

func TestP125_StripTaskPrefix_NotCheckbox(t *testing.T) {
	cells := []buffer.Cell{
		{Rune: 'a', Width: 1},
		{Rune: 'b', Width: 1},
		{Rune: 'c', Width: 1},
		{Rune: 'd', Width: 1},
	}
	result := stripTaskPrefix(cells)
	if len(result) != 4 {
		t.Errorf("expected unchanged 4 cells (not checkbox), got %d", len(result))
	}
}

func TestP125_StripTaskPrefix_BracketNotAtPos2(t *testing.T) {
	cells := []buffer.Cell{
		{Rune: '[', Width: 1},
		{Rune: 'x', Width: 1},
		{Rune: 'y', Width: 1}, // not ']'
		{Rune: ' ', Width: 1},
		{Rune: 'z', Width: 1},
	}
	result := stripTaskPrefix(cells)
	if len(result) != 5 {
		t.Errorf("expected unchanged (no ] at pos 2), got %d", len(result))
	}
}

// === NewHighlighterWithStyle nil fallback (75% → 90%+) ===

func TestP125_NewHighlighterWithStyle_Empty(t *testing.T) {
	h := NewHighlighterWithStyle("")
	if h == nil {
		t.Fatal("expected non-nil highlighter for empty style")
	}
}

func TestP125_NewHighlighterWithStyle_UnknownStyle(t *testing.T) {
	h := NewHighlighterWithStyle("nonexistent-style-12345")
	if h == nil {
		t.Fatal("expected non-nil highlighter for unknown style")
	}
	// Should fall back to dracula
	cells, _ := h.Highlight("var x = 1", "go")
	if len(cells) == 0 {
		t.Error("expected at least one line of highlighted code")
	}
}

func TestP125_NewHighlighterWithStyle_Monokai(t *testing.T) {
	h := NewHighlighterWithStyle("monokai")
	if h == nil {
		t.Fatal("expected non-nil for monokai")
	}
	cells, _ := h.Highlight("fmt.Println(\"hi\")", "go")
	if len(cells) == 0 {
		t.Error("expected highlighted output")
	}
}

// === mermaid edge types (75% → 90%+) ===

func TestP125_MermaidDrawEdge_Dashed(t *testing.T) {
	text := `graph TD
A -.-> B`
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
}

func TestP125_MermaidDrawEdge_Thick(t *testing.T) {
	text := `graph LR
A ==> B`
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
}

func TestP125_MermaidDrawEdge_Plain(t *testing.T) {
	text := `graph TD
A --- B`
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
}

func TestP125_MermaidDrawEdge_WithLabel(t *testing.T) {
	text := `graph TD
A -->|label| B`
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
}

func TestP125_Mermaid_BT_Direction(t *testing.T) {
	text := `graph BT
A --> B`
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render(text)
	if err != nil {
		t.Fatal(err)
	}
	_ = blocks
}

func TestP125_Mermaid_RL_Direction(t *testing.T) {
	text := `graph RL
A --> B`
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render(text)
	if err != nil {
		t.Fatal(err)
	}
	_ = blocks
}

// === renderFencedCode with math/latex (81% → 90%+) ===

func TestP125_RenderFencedCode_Math(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render("```math\nE = mc^2\n```")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
}

func TestP125_RenderFencedCode_Latex(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render("```latex\n\\frac{a}{b}\n```")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
}

func TestP125_RenderFencedCode_Mermaid(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render("```mermaid\ngraph TD\nA --> B\n```")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
}

// === Inline math detection (89.5% → 95%+) ===

func TestP125_HasInlineMath_Dollar(t *testing.T) {
	if !HasInlineMath("The formula $x^2$ is inline") {
		t.Error("expected inline math detection")
	}
}

func TestP125_HasInlineMath_Paren(t *testing.T) {
	if !HasInlineMath("The formula \\(x^2\\) is inline") {
		t.Error("expected inline math detection")
	}
}

func TestP125_HasInlineMath_None(t *testing.T) {
	if HasInlineMath("Just regular text") {
		t.Error("should not detect inline math")
	}
}

func TestP125_RenderInlineMath_Dollar(t *testing.T) {
	result := RenderInlineMath("The value $x^2$ is squared")
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP125_RenderInlineMath_Paren(t *testing.T) {
	result := RenderInlineMath("Value \\(y^2\\) squared")
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP125_RenderInlineMath_None(t *testing.T) {
	result := RenderInlineMath("No math here")
	if result != "No math here" {
		t.Errorf("expected unchanged text, got %q", result)
	}
}

// === FormatOSC8 (66.7% → 100%) ===

func TestP125_FormatOSC8_WithURL(t *testing.T) {
	result := FormatOSC8("click here", "https://example.com")
	if result == "" {
		t.Error("expected non-empty OSC8 sequence")
	}
}

func TestP125_FormatOSC8_EmptyURL(t *testing.T) {
	result := FormatOSC8("text", "")
	if result == "" {
		t.Error("expected non-empty for empty URL")
	}
}

// === HighlightToLines error/empty paths (90% → 95%+) ===

func TestP125_HighlightToLines_EmptySource(t *testing.T) {
	h := NewHighlighter()
	lines, _ := h.HighlightToLines("", "go")
	if len(lines) != 1 {
		t.Errorf("expected 1 line for empty source, got %d", len(lines))
	}
}

// === RenderMathToCells (was 0%) ===

func TestP125_RenderMathToCells_Basic(t *testing.T) {
	cells := RenderMathToCells("x^2", buffer.Color{})
	if len(cells) == 0 {
		t.Error("expected at least some cells")
	}
}

func TestP125_RenderMathToCells_Empty(t *testing.T) {
	cells := RenderMathToCells("", buffer.Color{})
	if len(cells) != 0 {
		t.Errorf("expected 0 cells for empty, got %d", len(cells))
	}
}

func TestP125_RenderMathToCells_Greek(t *testing.T) {
	cells := RenderMathToCells("\\alpha + \\beta", buffer.Color{})
	if len(cells) == 0 {
		t.Error("expected cells for Greek letters")
	}
}

// === RenderMermaidText (was 0%) ===

func TestP125_RenderMermaidText_Simple(t *testing.T) {
	cells := RenderMermaidText("graph TD\nA --> B", DefaultTheme())
	if len(cells) == 0 {
		t.Error("expected at least some cells")
	}
}

func TestP125_RenderMermaidText_Invalid(t *testing.T) {
	cells := RenderMermaidText("invalid mermaid syntax }}}", DefaultTheme())
	_ = cells // should not panic
}

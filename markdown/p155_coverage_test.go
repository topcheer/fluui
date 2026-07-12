package markdown

import (
	"testing"
)

// Target renderInline 76.5% — specifically the fast path with inline math
func TestP155_RenderInline_SingleTextMath(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	// Single text node with inline math should trigger fast path + math conversion
	blocks, err := r.Render("$x^2 + y^2$")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP155_RenderInline_SingleTextNoMath(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	// Single text node without math — should trigger fast path textToCells
	blocks, err := r.Render("Hello world this is plain text")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP155_RenderInline_MultiChildMixed(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	// Multiple children: bold + text + code — multi-child path
	blocks, err := r.Render("**bold** plain `code` more text")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP155_RenderInline_MultiChildWithMath(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	// Multi-child with math
	blocks, err := r.Render("Before $x^2$ after")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP155_RenderInline_EmptyParagraph(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	_ = blocks
}

func TestP155_RenderInline_ParenMathFastPath(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	// \(...\) should trigger HasInlineMath + paren format
	blocks, err := r.Render("\\(x + y\\)")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

// Target Highlight 85.5% — error paths
func TestP155_Highlight_EmptySource(t *testing.T) {
	h := NewHighlighter()
	cells, err := h.Highlight("go", "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	_ = cells
}

func TestP155_Highlight_NilLexer(t *testing.T) {
	h := NewHighlighter()
	// Unknown language should fall back to fallback lexer
	cells, err := h.Highlight("", "some code")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(cells) == 0 {
		t.Error("expected cells for fallback lexer")
	}
}

func TestP155_Highlight_MultiLineCode(t *testing.T) {
	h := NewHighlighter()
	code := "package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello\")\n}"
	cells, err := h.Highlight("go", code)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(cells) == 0 {
		t.Error("expected cells")
	}
}

func TestP155_Highlight_WithNewlines(t *testing.T) {
	h := NewHighlighter()
	// Code with many newlines to test the split path
	code := "a\nb\nc\nd\ne"
	cells, err := h.Highlight("go", code)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	_ = cells
}

// Target tokenTypeColor 85.7% — more chroma categories
func TestP155_TokenTypeColor_DiverseTokens(t *testing.T) {
	h := NewHighlighter()
	// Go code with diverse token types
	codes := []struct {
		lang string
		code string
	}{
		{"go", "package main\nimport \"fmt\"\nconst PI = 3.14\ntype Foo struct{ Bar int }\nvar x = 1\n// comment\nfunc add(a, b int) int { return a + b }"},
		{"python", "class Foo:\n    def __init__(self):\n        self.x = 42\n    # comment\n    pass"},
		{"javascript", "const x = 42;\nlet y = 'str';\nfunction f() { return true; }\n// comment"},
		{"bash", "#!/bin/bash\necho hello\nx=1\nif [ $x ]; then echo yes; fi"},
		{"json", "{\"key\": \"value\", \"num\": 42, \"arr\": [1, 2, 3]}"},
	}
	for _, tc := range codes {
		cells, err := h.Highlight(tc.lang, tc.code)
		if err != nil {
			t.Errorf("highlight %s error: %v", tc.lang, err)
		}
		if len(cells) == 0 {
			t.Errorf("expected cells for %s", tc.lang)
		}
	}
}

// Target mermaidDrawNode 87.5%
func TestP155_MermaidNode_LongLabel(t *testing.T) {
	g := &MermaidGraph{
		Direction: MermaidTD,
		Nodes: []*MermaidNode{
			{ID: "L", Label: "This is a very long label that needs truncation", Shape: MermaidShapeRect, X: 0, Y: 0, W: 10, H: 3},
		},
	}
	theme := MarkdownTheme{}
	cells := RenderMermaid(g, &theme)
	if len(cells) == 0 {
		t.Fatal("expected cells")
	}
}

func TestP155_MermaidNode_NoLabel(t *testing.T) {
	g := &MermaidGraph{
		Direction: MermaidLR,
		Nodes: []*MermaidNode{
			{ID: "N", Label: "", Shape: MermaidShapeRect, X: 0, Y: 0, W: 6, H: 3},
		},
	}
	theme := MarkdownTheme{}
	cells := RenderMermaid(g, &theme)
	if len(cells) == 0 {
		t.Fatal("expected cells")
	}
}

// Target renderList 84.8% — more list types
func TestP155_RenderList_OrderedNestedDeep(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("1. A\n   1. B\n      1. C\n         1. D")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP155_RenderList_TaskWithContent(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("- [ ] **bold task**\n- [x] ~~strikethrough done~~")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP155_RenderList_NarrowWidth(t *testing.T) {
	r := NewMarkdownRenderer(nil, 15)
	blocks, err := r.Render("- This is a very long list item that wraps\n- Short\n- Another long item that also wraps around")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

// Target consumeCommand 83.3% — more LaTeX commands
func TestP155_ConsumeCommand_SumStar(t *testing.T) {
	s := "\\sum*_{i=1}^{n} x_i"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty")
	}
}

func TestP155_ConsumeCommand_ProdStar(t *testing.T) {
	s := "\\prod*_{i=1}^{n} x_i"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty")
	}
}

func TestP155_ConsumeCommand_Int(t *testing.T) {
	s := "\\int_0^1 x^2 dx"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty")
	}
}

// Target consumeGroup 84.6% — edge cases
func TestP155_ConsumeGroup_NotSubset(t *testing.T) {
	s := "x^{not_a_subset}"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty")
	}
}

func TestP155_ConsumeGroup_NestedFracInExp(t *testing.T) {
	s := "x^{\\frac{a}{b}}"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty")
	}
}

// Target skipBracket 84.6%
func TestP155_SkipBracket_NthRoot(t *testing.T) {
	s := "\\sqrt[3]{x}"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty")
	}
}

func TestP155_SkipBracket_LongIndex(t *testing.T) {
	s := "\\sqrt[abc]{x+1}"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty")
	}
}

// Target WrapText 94.0%
func TestP155_WrapText_LongLine(t *testing.T) {
	text := "This is a very long line that needs to be wrapped at a specific width to test the wrapping logic"
	wrapped := WrapText(text, 20)
	if len(wrapped) == 0 {
		t.Error("expected wrapped text")
	}
}

func TestP155_WrapText_ShortLine(t *testing.T) {
	text := "Short"
	wrapped := WrapText(text, 20)
	if len(wrapped) != 1 || wrapped[0] != text {
		t.Errorf("expected [%q], got %v", text, wrapped)
	}
}

func TestP155_WrapText_Empty(t *testing.T) {
	wrapped := WrapText("", 20)
	if len(wrapped) > 1 || (len(wrapped) == 1 && wrapped[0] != "") {
		t.Errorf("expected empty or single empty, got %v", wrapped)
	}
}

// Target renderFencedCode 95.2%
func TestP155_RenderFencedCode_RegularCode(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("```python\nprint('hello')\n```")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP155_RenderFencedCode_UnknownLang(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("```unknown_lang\nsome code\n```")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

// Target renderBlock 94.4%
func TestP155_RenderBlock_HorizontalRule(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("Before\n\n---\n\nAfter")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP155_RenderBlock_Blockquote(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("> Quote line 1\n> Quote line 2")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP155_RenderBlock_AlertWarning(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("> [!WARNING]\n> Be careful")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP155_RenderBlock_AlertCaution(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("> [!CAUTION]\n> Danger")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP155_RenderBlock_AlertTip(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("> [!TIP]\n> Helpful tip")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP155_RenderBlock_AlertImportant(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("> [!IMPORTANT]\n> Critical info")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}
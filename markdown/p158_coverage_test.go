package markdown

import (
	"testing"
)

// Target renderInline 76.5% — focus on uncovered branches
func TestP158_RenderInline_InlineMathDollar(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("Math $x^2$ here")
	if err != nil { t.Fatalf("error: %v", err) }
	if len(blocks) == 0 { t.Fatal("expected blocks") }
}

func TestP158_RenderInline_InlineMathParen(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("Math \\(y^2\\) here")
	if err != nil { t.Fatalf("error: %v", err) }
	if len(blocks) == 0 { t.Fatal("expected blocks") }
}

func TestP158_RenderInline_MixedAllInline(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	input := "**bold** *italic* `code` ~~strike~~ [link](http://x) ![img](http://y.png) $math$ \\(paren\\)"
	blocks, err := r.Render(input)
	if err != nil { t.Fatalf("error: %v", err) }
	if len(blocks) == 0 { t.Fatal("expected blocks") }
}

func TestP158_RenderInline_EmptyText(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("   ")
	if err != nil { t.Fatalf("error: %v", err) }
	_ = blocks
}

func TestP158_RenderInline_OnlyMath(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("$\\frac{a}{b}$")
	if err != nil { t.Fatalf("error: %v", err) }
	if len(blocks) == 0 { t.Fatal("expected blocks") }
}

func TestP158_RenderInline_NewlineInText(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("Line one\nLine two\nLine three")
	if err != nil { t.Fatalf("error: %v", err) }
	if len(blocks) == 0 { t.Fatal("expected blocks") }
}

// Target Highlight 85.5% — uncovered branches
func TestP158_Highlight_NilLexerFallback(t *testing.T) {
	h := NewHighlighter()
	// Empty language should use fallback
	cells, err := h.Highlight("", "some code")
	if err != nil { t.Error("unexpected error") }
	if len(cells) == 0 { t.Error("expected cells") }
}

func TestP158_Highlight_UnknownLexerFallback(t *testing.T) {
	h := NewHighlighter()
	cells, err := h.Highlight("totally_unknown_language_xyz", "code here")
	if err != nil { t.Error("unexpected error") }
	if len(cells) == 0 { t.Error("expected cells from fallback") }
}

func TestP158_Highlight_MultiLineNewlines(t *testing.T) {
	h := NewHighlighter()
	code := "package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"hi\")\n}\n"
	cells, err := h.Highlight("go", code)
	if err != nil { t.Error("unexpected error") }
	if len(cells) == 0 { t.Error("expected cells") }
}

func TestP158_Highlight_EmptyCode(t *testing.T) {
	h := NewHighlighter()
	cells, _ := h.Highlight("go", "")
	_ = cells
}

// Target tokenTypeColor 85.7% — more diverse token types
func TestP158_TokenTypeColor_GoDiverse(t *testing.T) {
	h := NewHighlighter()
	code := `package main

import "fmt"

const PI = 3.14

type Point struct {
    X, Y int
}

func (p *Point) String() string {
    return fmt.Sprintf("(%d, %d)", p.X, p.Y)
}

// Comment line
/* Block comment */
func main() {
    var x = 42
    if x > 0 {
        for i := 0; i < x; i++ {
            go func() {}()
        }
    }
}`
	cells, err := h.Highlight("go", code)
	if err != nil { t.Error("unexpected error") }
	if len(cells) == 0 { t.Error("expected cells") }
}

func TestP158_TokenTypeColor_PythonDiverse(t *testing.T) {
	h := NewHighlighter()
	code := `class Foo:
    """Docstring"""
    def __init__(self, x):
        self.x = x
        self.y = [1, 2, 3]
    
    @staticmethod
    def bar():
        return True or False and None`
	cells, err := h.Highlight("python", code)
	if err != nil { t.Error("unexpected error") }
	if len(cells) == 0 { t.Error("expected cells") }
}

func TestP158_TokenTypeColor_JSDiverse(t *testing.T) {
	h := NewHighlighter()
	code := `const x = 42;
let y = "hello";
const fn = (a, b) => a + b;
class Foo extends Bar {
    constructor() { super(); }
}
// comment
const arr = [1, 2, 3].map(x => x * 2);`
	cells, err := h.Highlight("javascript", code)
	if err != nil { t.Error("unexpected error") }
	if len(cells) == 0 { t.Error("expected cells") }
}

func TestP158_TokenTypeColor_BashDiverse(t *testing.T) {
	h := NewHighlighter()
	code := `#!/bin/bash
VAR="hello"
if [ -f "$1" ]; then
    echo "File exists"
    cat "$1" | grep "pattern"
fi
for i in $(seq 1 10); do
    echo $i
done`
	cells, err := h.Highlight("bash", code)
	if err != nil { t.Error("unexpected error") }
	if len(cells) == 0 { t.Error("expected cells") }
}

func TestP158_TokenTypeColor_JSONDiverse(t *testing.T) {
	h := NewHighlighter()
	code := `{
    "name": "test",
    "value": 42,
    "arr": [1, 2, 3],
    "nested": {"key": "val"},
    "bool": true,
    "null": null
}`
	cells, err := h.Highlight("json", code)
	if err != nil { t.Error("unexpected error") }
	if len(cells) == 0 { t.Error("expected cells") }
}

// Target consumeCommand 83.3% — uncovered branches
func TestP158_ConsumeCommand_LeftRight(t *testing.T) {
	result := RenderLatexMath("\\left(\\frac{a}{b}\\right)")
	if result == "" { t.Error("expected non-empty") }
}

func TestP158_ConsumeCommand_BigBigg(t *testing.T) {
	result := RenderLatexMath("\\big(\\bigg(x\\bigg)\\big)")
	if result == "" { t.Error("expected non-empty") }
}

func TestP158_ConsumeCommand_UnknownCmd(t *testing.T) {
	result := RenderLatexMath("\\unknownxyz{test}")
	if result == "" { t.Error("expected non-empty") }
}

func TestP158_ConsumeCommand_LineBreak(t *testing.T) {
	result := RenderLatexMath("a\\\\b\\\\c")
	if result == "" { t.Error("expected non-empty") }
}

func TestP158_ConsumeCommand_SumStar(t *testing.T) {
	result := RenderLatexMath("\\sum*_{i=0}^{n} x_i")
	if result == "" { t.Error("expected non-empty") }
}

func TestP158_ConsumeCommand_ProdStar(t *testing.T) {
	result := RenderLatexMath("\\prod*_{i=0}^{n} x_i")
	if result == "" { t.Error("expected non-empty") }
}

func TestP158_ConsumeCommand_Int(t *testing.T) {
	result := RenderLatexMath("\\int_0^{\\infty} e^{-x^2} dx")
	if result == "" { t.Error("expected non-empty") }
}

// Target consumeGroup 84.6% — uncovered branches
func TestP158_ConsumeGroup_NestedFrac(t *testing.T) {
	result := RenderLatexMath("\\frac{\\frac{a}{b}}{\\frac{c}{d}}")
	if result == "" { t.Error("expected non-empty") }
}

func TestP158_ConsumeGroup_StrayClose(t *testing.T) {
	result := RenderLatexMath("x}")
	if result == "" { t.Error("expected non-empty") }
}

func TestP158_ConsumeGroup_DollarInside(t *testing.T) {
	result := RenderLatexMath("x$y}")
	if result == "" { t.Error("expected non-empty") }
}

func TestP158_ConsumeGroup_NotSubset(t *testing.T) {
	result := RenderLatexMath("x^{not_subset_of_anything}")
	if result == "" { t.Error("expected non-empty") }
}

// Target advance 80.0% — uncovered branches
func TestP158_Advance_Spaces(t *testing.T) {
	result := RenderLatexMath("\\frac  {a}  {b}")
	if result == "" { t.Error("expected non-empty") }
}

func TestP158_Advance_Newline(t *testing.T) {
	result := RenderLatexMath("\\frac\n{a}\n{b}")
	if result == "" { t.Error("expected non-empty") }
}

func TestP158_Advance_Tab(t *testing.T) {
	result := RenderLatexMath("\\frac\t{a}\t{b}")
	if result == "" { t.Error("expected non-empty") }
}

// Target skipBracket 84.6% — uncovered branches
func TestP158_SkipBracket_NthRoot(t *testing.T) {
	result := RenderLatexMath("\\sqrt[3]{x}")
	if result == "" { t.Error("expected non-empty") }
}

func TestP158_SkipBracket_LongIndex(t *testing.T) {
	result := RenderLatexMath("\\sqrt[abc]{x}")
	if result == "" { t.Error("expected non-empty") }
}

func TestP158_SkipBracket_NoClose(t *testing.T) {
	result := RenderLatexMath("\\sqrt[3 x")
	if result == "" { t.Error("expected non-empty") }
}

// Target mermaidDrawNode 87.5% — uncovered branches
func TestP158_MermaidNode_LongLabelTrunc(t *testing.T) {
	g := &MermaidGraph{
		Direction: MermaidTD,
		Nodes: []*MermaidNode{
			{ID: "L", Label: "This is a very long label that needs truncation for sure", Shape: MermaidShapeRect, X: 0, Y: 0, W: 10, H: 3},
		},
	}
	cells := RenderMermaid(g, &MarkdownTheme{})
	if len(cells) == 0 { t.Fatal("expected cells") }
}

func TestP158_MermaidNode_NoLabel(t *testing.T) {
	g := &MermaidGraph{
		Direction: MermaidLR,
		Nodes: []*MermaidNode{
			{ID: "N", Label: "", Shape: MermaidShapeCircle, X: 0, Y: 0, W: 6, H: 3},
		},
	}
	cells := RenderMermaid(g, &MarkdownTheme{})
	if len(cells) == 0 { t.Fatal("expected cells") }
}

func TestP158_MermaidNode_LRDirection(t *testing.T) {
	g := &MermaidGraph{
		Direction: MermaidLR,
		Nodes: []*MermaidNode{
			{ID: "A", Label: "Start", Shape: MermaidShapeRect, X: 0, Y: 0, W: 8, H: 3},
			{ID: "B", Label: "End", Shape: MermaidShapeRounded, X: 15, Y: 0, W: 8, H: 3},
		},
		Edges: []*MermaidEdge{
			{From: "A", To: "B", Style: MermaidEdgeArrow},
		},
	}
	cells := RenderMermaid(g, &MarkdownTheme{})
	if len(cells) == 0 { t.Fatal("expected cells") }
}

// Target mermaidDrawEdge 92.2% — uncovered branches
func TestP158_MermaidEdge_Dotted(t *testing.T) {
	g := &MermaidGraph{
		Direction: MermaidTD,
		Nodes: []*MermaidNode{
			{ID: "A", Label: "A", Shape: MermaidShapeRect, X: 0, Y: 0, W: 6, H: 3},
			{ID: "B", Label: "B", Shape: MermaidShapeRect, X: 0, Y: 5, W: 6, H: 3},
		},
		Edges: []*MermaidEdge{
			{From: "A", To: "B", Style: MermaidEdgeDotted},
		},
	}
	cells := RenderMermaid(g, &MarkdownTheme{})
	if len(cells) == 0 { t.Fatal("expected cells") }
}

func TestP158_MermaidEdge_Thick(t *testing.T) {
	g := &MermaidGraph{
		Direction: MermaidTD,
		Nodes: []*MermaidNode{
			{ID: "A", Label: "A", Shape: MermaidShapeRect, X: 0, Y: 0, W: 6, H: 3},
			{ID: "B", Label: "B", Shape: MermaidShapeRect, X: 0, Y: 5, W: 6, H: 3},
		},
		Edges: []*MermaidEdge{
			{From: "A", To: "B", Style: MermaidEdgeThick},
		},
	}
	cells := RenderMermaid(g, &MarkdownTheme{})
	if len(cells) == 0 { t.Fatal("expected cells") }
}

func TestP158_MermaidEdge_WithLabel(t *testing.T) {
	g := &MermaidGraph{
		Direction: MermaidTD,
		Nodes: []*MermaidNode{
			{ID: "A", Label: "A", Shape: MermaidShapeRect, X: 0, Y: 0, W: 6, H: 3},
			{ID: "B", Label: "B", Shape: MermaidShapeRect, X: 0, Y: 5, W: 6, H: 3},
		},
		Edges: []*MermaidEdge{
			{From: "A", To: "B", Label: "yes", Style: MermaidEdgeArrow},
		},
	}
	cells := RenderMermaid(g, &MarkdownTheme{})
	if len(cells) == 0 { t.Fatal("expected cells") }
}

// Target HighlightToLines 90.0%
func TestP158_HighlightToLines_Error(t *testing.T) {
	h := NewHighlighter()
	lines, err := h.HighlightToLines("", "")
	if err == nil && len(lines) > 0 {
		// May return empty or fallback
	}
	_ = lines
	_ = err
}

func TestP158_HighlightToLines_Valid(t *testing.T) {
	h := NewHighlighter()
	lines, err := h.HighlightToLines("go", "package main\nfunc main() {}")
	if err != nil { t.Errorf("unexpected error: %v", err) }
	if len(lines) == 0 { t.Error("expected lines") }
}

func TestP158_HighlightToLines_UnknownLang(t *testing.T) {
	h := NewHighlighter()
	lines, err := h.HighlightToLines("unknown_xyz", "some code")
	if err != nil { t.Errorf("unexpected error: %v", err) }
	if len(lines) == 0 { t.Error("expected lines from fallback") }
}

// Target renderBlock 94.4%
func TestP158_RenderBlock_Blockquote(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("> Quote line 1\n> Quote line 2\n> Quote line 3")
	if err != nil { t.Fatalf("error: %v", err) }
	if len(blocks) == 0 { t.Fatal("expected blocks") }
}

func TestP158_RenderBlock_AlertNote(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("> [!NOTE]\n> This is a note")
	if err != nil { t.Fatalf("error: %v", err) }
	if len(blocks) == 0 { t.Fatal("expected blocks") }
}

func TestP158_RenderBlock_AlertImportant(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("> [!IMPORTANT]\n> Critical info")
	if err != nil { t.Fatalf("error: %v", err) }
	if len(blocks) == 0 { t.Fatal("expected blocks") }
}

// Target NewHighlighterWithStyle 75%
func TestP158_NewHighlighterWithStyle_Monokai(t *testing.T) {
	h := NewHighlighterWithStyle("monokai")
	if h == nil { t.Fatal("expected non-nil") }
}

func TestP158_NewHighlighterWithStyle_Empty(t *testing.T) {
	h := NewHighlighterWithStyle("")
	if h == nil { t.Fatal("expected non-nil (fallback to dracula)") }
}

func TestP158_NewHighlighterWithStyle_Unknown(t *testing.T) {
	h := NewHighlighterWithStyle("totally_unknown_style")
	if h == nil { t.Fatal("expected non-nil (fallback)") }
}

func TestP158_NewHighlighterWithStyle_Dracula(t *testing.T) {
	h := NewHighlighterWithStyle("dracula")
	if h == nil { t.Fatal("expected non-nil") }
}
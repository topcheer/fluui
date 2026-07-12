package markdown

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// Target renderInline — 76.5% coverage
func TestP154_RenderInline_Bold(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("**bold text**")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderInline_Italic(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("*italic text*")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderInline_Code(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("`inline code`")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderInline_Link(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("[click here](https://example.com)")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderInline_Image(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("![alt text](https://example.com/img.png)")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderInline_Strikethrough(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("~~deleted text~~")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderInline_Mixed(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("**bold** and *italic* and `code` and [link](http://x)")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderInline_InlineMath(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("Math: $x^2 + y^2 = z^2$")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderInline_ParenMath(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("Math: \\(x^2 + y^2\\)")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderInline_RawURL(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("Visit https://example.com today")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

// Target renderList — 84.8% coverage
func TestP154_RenderList_Nested(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("- Top\n  - Nested\n  - Another\n- Back")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderList_OrderedNested(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("1. First\n   1. Sub-first\n   2. Sub-second\n2. Second")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderList_DeepNesting(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("- A\n  - B\n    - C\n      - D")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderList_TaskList(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("- [ ] todo\n- [x] done\n- [X] also done")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderList_MixedMarkers(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("- dash\n* star\n+ plus")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderList_LongLine(t *testing.T) {
	r := NewMarkdownRenderer(nil, 20)
	longText := "This is a very long list item that should wrap across multiple lines when rendered"
	blocks, err := r.Render("- " + longText)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderList_OrderedLarge(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render("10. Tenth\n11. Eleventh\n12. Twelfth")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

// Target consumeCommand — 83.3% coverage
func TestP154_ConsumeCommand_Left(t *testing.T) {
	s := "\\left(a\\right)"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP154_ConsumeCommand_Right(t *testing.T) {
	s := "\\right)"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP154_ConsumeCommand_Big(t *testing.T) {
	s := "\\big(x\\big)"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP154_ConsumeCommand_Bigg(t *testing.T) {
	s := "\\bigg(x\\bigg)"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP154_ConsumeCommand_Unknown(t *testing.T) {
	s := "\\unknowncmd{x}"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP154_ConsumeCommand_LineBreak(t *testing.T) {
	s := "a\\\\b"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

// Target consumeGroup — 84.6% coverage
func TestP154_ConsumeGroup_Nested(t *testing.T) {
	s := "x^{a^{b}}"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP154_ConsumeGroup_FracNested(t *testing.T) {
	s := "\\frac{\\frac{a}{b}}{c}"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP154_ConsumeGroup_StrayBrace(t *testing.T) {
	s := "x}"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP154_ConsumeGroup_Dollar(t *testing.T) {
	s := "x$_y$"
	result := RenderLatexMath(s)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

// Target tokenTypeColor — 85.7% coverage
func TestP154_TokenTypeColor_Diverse(t *testing.T) {
	h := NewHighlighter()
	codes := []string{
		"package main\nfunc main() {\nvar x = 1\n// comment\nreturn x\n}",
		"import \"fmt\"\nconst PI = 3.14\ntype Foo struct{}\nfor i := 0; i < 10; i++ {}",
		"if true {}\nswitch x {\ncase 1:\ndefault:}",
		"func add(a int, b int) int { return a + b }",
		"str := \"hello\"\nnum := 42\nbool := true",
	}
	for _, code := range codes {
		cells, err := h.Highlight("go", code)
		if err != nil {
			t.Errorf("highlight error: %v", err)
		}
		if len(cells) == 0 {
			t.Error("expected cells")
		}
	}
}

func TestP154_TokenTypeColor_Python(t *testing.T) {
	h := NewHighlighter()
	cells, err := h.Highlight("python", "def foo():\n    return 42")
	if err != nil {
		t.Errorf("highlight error: %v", err)
	}
	if len(cells) == 0 {
		t.Error("expected cells")
	}
}

func TestP154_TokenTypeColor_JavaScript(t *testing.T) {
	h := NewHighlighter()
	cells, err := h.Highlight("javascript", "const x = () => 42;")
	if err != nil {
		t.Errorf("highlight error: %v", err)
	}
	if len(cells) == 0 {
		t.Error("expected cells")
	}
}

// Target mermaidDrawNode — 87.5% coverage
func TestP154_MermaidDrawNode_Rect(t *testing.T) {
	g := &MermaidGraph{
		Direction: MermaidTD,
		Nodes: []*MermaidNode{
			{ID: "A", Label: "Task A", Shape: MermaidShapeRect, X: 0, Y: 0, W: 10, H: 3},
		},
	}
	theme := MarkdownTheme{}
	cells := RenderMermaid(g, &theme)
	if len(cells) == 0 {
		t.Fatal("expected cells")
	}
}

func TestP154_MermaidDrawNode_Diamond(t *testing.T) {
	g := &MermaidGraph{
		Direction: MermaidTD,
		Nodes: []*MermaidNode{
			{ID: "D", Label: "Decide", Shape: MermaidShapeDiamond, X: 0, Y: 0, W: 12, H: 3},
		},
	}
	theme := MarkdownTheme{}
	cells := RenderMermaid(g, &theme)
	if len(cells) == 0 {
		t.Fatal("expected cells")
	}
}

func TestP154_MermaidDrawNode_Circle(t *testing.T) {
	g := &MermaidGraph{
		Direction: MermaidTD,
		Nodes: []*MermaidNode{
			{ID: "C", Label: "Start", Shape: MermaidShapeCircle, X: 0, Y: 0, W: 10, H: 3},
		},
	}
	theme := MarkdownTheme{}
	cells := RenderMermaid(g, &theme)
	if len(cells) == 0 {
		t.Fatal("expected cells")
	}
}

func TestP154_MermaidDrawNode_Rounded(t *testing.T) {
	g := &MermaidGraph{
		Direction: MermaidTD,
		Nodes: []*MermaidNode{
			{ID: "R", Label: "Process", Shape: MermaidShapeRounded, X: 0, Y: 0, W: 10, H: 3},
		},
	}
	theme := MarkdownTheme{}
	cells := RenderMermaid(g, &theme)
	if len(cells) == 0 {
		t.Fatal("expected cells")
	}
}

func TestP154_MermaidDrawNode_Plain(t *testing.T) {
	g := &MermaidGraph{
		Direction: MermaidTD,
		Nodes: []*MermaidNode{
			{ID: "P", Label: "Plain", Shape: MermaidShapePlain, X: 0, Y: 0, W: 10, H: 3},
		},
	}
	theme := MarkdownTheme{}
	cells := RenderMermaid(g, &theme)
	if len(cells) == 0 {
		t.Fatal("expected cells")
	}
}

func TestP154_Mermaid_Edges(t *testing.T) {
	g := &MermaidGraph{
		Direction: MermaidTD,
		Nodes: []*MermaidNode{
			{ID: "A", Label: "A", Shape: MermaidShapeRect, X: 0, Y: 0, W: 6, H: 3},
			{ID: "B", Label: "B", Shape: MermaidShapeRect, X: 0, Y: 5, W: 6, H: 3},
		},
		Edges: []*MermaidEdge{
			{From: "A", To: "B", Label: "next", Style: MermaidEdgeArrow},
		},
	}
	theme := MarkdownTheme{}
	cells := RenderMermaid(g, &theme)
	if len(cells) == 0 {
		t.Fatal("expected cells")
	}
}

func TestP154_RenderFencedCode_MermaidBlock(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	input := "```mermaid\ngraph TD\n  A-->B\n```"
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderFencedCode_MathBlock(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	input := "```math\nx^2 + y^2 = z^2\n```"
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_RenderFencedCode_LatexBlock(t *testing.T) {
	r := NewMarkdownRenderer(nil, 80)
	input := "```latex\n\\frac{a}{b}\n```"
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected blocks")
	}
}

func TestP154_Highlight_EmptyCode(t *testing.T) {
	h := NewHighlighter()
	cells, err := h.Highlight("go", "")
	if err != nil {
		t.Errorf("highlight error: %v", err)
	}
	// Empty code may return 0 or 1 cells
	_ = cells
}

func TestP154_Highlight_UnknownLang(t *testing.T) {
	h := NewHighlighter()
	cells, err := h.Highlight("unknown_lang_xyz", "code")
	if err != nil {
		t.Errorf("highlight error: %v", err)
	}
	if len(cells) == 0 {
		t.Error("expected cells even for unknown lang")
	}
}

// Ensure buffer import is used
var _ = buffer.BlankCell
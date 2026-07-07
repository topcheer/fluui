package markdown

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── NewHighlighterWithStyle (75%) ───

func TestP108_NewHighlighterWithStyle_EmptyString(t *testing.T) {
	h := NewHighlighterWithStyle("")
	if h == nil {
		t.Fatal("should not return nil")
	}
	// Empty style name → styles.Get("") returns nil → fallback to dracula
}

func TestP108_NewHighlighterWithStyle_UnknownStyle(t *testing.T) {
	h := NewHighlighterWithStyle("this-style-does-not-exist-12345")
	if h == nil {
		t.Fatal("should not return nil for unknown style")
	}
}

func TestP108_NewHighlighterWithStyle_Monokai(t *testing.T) {
	h := NewHighlighterWithStyle("monokai")
	if h == nil {
		t.Fatal("should not return nil for monokai")
	}
}

func TestP108_NewHighlighterWithStyle_Dracula(t *testing.T) {
	h := NewHighlighterWithStyle("dracula")
	if h == nil {
		t.Fatal("should not return nil for dracula")
	}
}

// ─── mermaidDrawEdge (75%) ───

func TestP108_MermaidDrawEdge_TD_DottedHorizontal(t *testing.T) {
	// TD direction where fromX != toX, requiring horizontal connector
	// AND using dotted edge style
	from := &MermaidNode{ID: "A", X: 0, Y: 0, W: 5, H: 2}
	to := &MermaidNode{ID: "B", X: 10, Y: 6, W: 5, H: 2}
	edge := &MermaidEdge{From: "A", To: "B", Style: MermaidEdgeDotted, Label: "test"}

	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 40)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	mermaidDrawEdge(canvas, from, to, edge, MermaidTD)
	// Should have drawn dotted lines and arrow
}

func TestP108_MermaidDrawEdge_TD_ThickHorizontal(t *testing.T) {
	from := &MermaidNode{ID: "A", X: 0, Y: 0, W: 5, H: 2}
	to := &MermaidNode{ID: "B", X: 10, Y: 6, W: 5, H: 2}
	edge := &MermaidEdge{From: "A", To: "B", Style: MermaidEdgeThick}

	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 40)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	mermaidDrawEdge(canvas, from, to, edge, MermaidTD)
}

func TestP108_MermaidDrawEdge_TD_FromXGreaterThanToX(t *testing.T) {
	// fromX > toX to trigger the else branch in horizontal connector
	from := &MermaidNode{ID: "A", X: 15, Y: 0, W: 5, H: 2}
	to := &MermaidNode{ID: "B", X: 0, Y: 6, W: 5, H: 2}
	edge := &MermaidEdge{From: "A", To: "B"}

	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 40)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	mermaidDrawEdge(canvas, from, to, edge, MermaidTD)
}

func TestP108_MermaidDrawEdge_BT_Direction(t *testing.T) {
	// Bottom-to-top direction
	from := &MermaidNode{ID: "A", X: 0, Y: 10, W: 5, H: 2}
	to := &MermaidNode{ID: "B", X: 0, Y: 0, W: 5, H: 2}
	edge := &MermaidEdge{From: "A", To: "B"}

	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 40)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	mermaidDrawEdge(canvas, from, to, edge, MermaidBT)
}

func TestP108_MermaidDrawEdge_LR_DottedVertical(t *testing.T) {
	// LR direction where fromY != toY with dotted edge
	from := &MermaidNode{ID: "A", X: 0, Y: 0, W: 5, H: 2}
	to := &MermaidNode{ID: "B", X: 20, Y: 10, W: 5, H: 2}
	edge := &MermaidEdge{From: "A", To: "B", Style: MermaidEdgeDotted}

	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 40)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	mermaidDrawEdge(canvas, from, to, edge, MermaidLR)
}

func TestP108_MermaidDrawEdge_LR_ThickVertical(t *testing.T) {
	from := &MermaidNode{ID: "A", X: 0, Y: 0, W: 5, H: 2}
	to := &MermaidNode{ID: "B", X: 20, Y: 10, W: 5, H: 2}
	edge := &MermaidEdge{From: "A", To: "B", Style: MermaidEdgeThick}

	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 40)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	mermaidDrawEdge(canvas, from, to, edge, MermaidLR)
}

func TestP108_MermaidDrawEdge_LR_FromYGreaterThanToY(t *testing.T) {
	// fromY > toY to trigger the else branch
	from := &MermaidNode{ID: "A", X: 0, Y: 15, W: 5, H: 2}
	to := &MermaidNode{ID: "B", X: 20, Y: 0, W: 5, H: 2}
	edge := &MermaidEdge{From: "A", To: "B"}

	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 40)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	mermaidDrawEdge(canvas, from, to, edge, MermaidLR)
}

func TestP108_MermaidDrawEdge_LR_WithLabel(t *testing.T) {
	from := &MermaidNode{ID: "A", X: 0, Y: 0, W: 5, H: 2}
	to := &MermaidNode{ID: "B", X: 20, Y: 5, W: 5, H: 2}
	edge := &MermaidEdge{From: "A", To: "B", Label: "yes"}

	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 40)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	mermaidDrawEdge(canvas, from, to, edge, MermaidLR)
}

func TestP108_MermaidDrawEdge_TD_StraightDown(t *testing.T) {
	// fromX == toX (straight down, no horizontal connector)
	from := &MermaidNode{ID: "A", X: 5, Y: 0, W: 5, H: 2}
	to := &MermaidNode{ID: "B", X: 5, Y: 6, W: 5, H: 2}
	edge := &MermaidEdge{From: "A", To: "B"}

	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 40)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	mermaidDrawEdge(canvas, from, to, edge, MermaidTD)
}

func TestP108_MermaidDrawEdge_LR_StraightRight(t *testing.T) {
	// fromY == toY (straight right, no vertical connector)
	from := &MermaidNode{ID: "A", X: 0, Y: 5, W: 5, H: 2}
	to := &MermaidNode{ID: "B", X: 20, Y: 5, W: 5, H: 2}
	edge := &MermaidEdge{From: "A", To: "B"}

	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 40)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	mermaidDrawEdge(canvas, from, to, edge, MermaidLR)
}

// ─── HighlightToLines edge case ───

func TestP108_HighlightToLines_ErrorLanguage(t *testing.T) {
	h := NewHighlighter()
	cells, err := h.HighlightToLines("nonexistent-language-xyz", "some code\nhere")
	_ = err
	if len(cells) == 0 {
		t.Error("should return at least empty cells")
	}
}

// ─── RenderInlineMath more cases ───

func TestP108_RenderInlineMath_ParenFormat(t *testing.T) {
	result := RenderInlineMath("hello \\(x^2\\) world")
	if result == "" {
		t.Error("should not be empty")
	}
}

func TestP108_RenderInlineMath_DollarFormat(t *testing.T) {
	result := RenderInlineMath("hello $x^2$ world")
	if result == "" {
		t.Error("should not be empty")
	}
}

func TestP108_RenderInlineMath_NoMath(t *testing.T) {
	input := "just plain text without any math"
	result := RenderInlineMath(input)
	if result != input {
		t.Errorf("expected unchanged, got %q", result)
	}
}

// ─── HasInlineMath ───

func TestP108_HasInlineMath_Dollar(t *testing.T) {
	if !HasInlineMath("hello $x^2$ world") {
		t.Error("should detect dollar math")
	}
}

func TestP108_HasInlineMath_Paren(t *testing.T) {
	if !HasInlineMath("hello \\(x^2\\) world") {
		t.Error("should detect paren math")
	}
}

func TestP108_HasInlineMath_None(t *testing.T) {
	if HasInlineMath("just plain text") {
		t.Error("should not detect math in plain text")
	}
}

// ─── RenderMathToCells ───

func TestP108_RenderMathToCells_Basic(t *testing.T) {
	cells := RenderMathToCells("x^2 + y^2 = z^2", buffer.NamedColor(buffer.NamedCyan))
	if len(cells) == 0 {
		t.Error("should return cells")
	}
}

func TestP108_RenderMathToCells_Empty(t *testing.T) {
	cells := RenderMathToCells("", buffer.NamedColor(buffer.NamedRed))
	if len(cells) != 0 {
		t.Errorf("expected 0 cells for empty, got %d", len(cells))
	}
}

func TestP108_RenderMathToCells_Greek(t *testing.T) {
	cells := RenderMathToCells("\\alpha + \\beta + \\gamma", buffer.NamedColor(buffer.NamedGreen))
	if len(cells) == 0 {
		t.Error("should return cells")
	}
}

// ─── FormatOSC8 ───

func TestP108_FormatOSC8_WithURL(t *testing.T) {
	result := FormatOSC8("click here", "https://example.com")
	if result == "" {
		t.Error("should not be empty")
	}
}

func TestP108_FormatOSC8_EmptyURL(t *testing.T) {
	result := FormatOSC8("text", "")
	if result == "" {
		t.Error("should not be empty even with empty URL")
	}
}

// ─── renderFencedCode ───

func TestP108_RenderFencedCode_MermaidBlock(t *testing.T) {
	md := NewMarkdownRenderer(nil, 80)
	blocks, err := md.Render("```mermaid\nA --> B\n```")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if len(blocks) == 0 {
		t.Error("should return at least 1 block")
	}
}

func TestP108_RenderFencedCode_MathBlock(t *testing.T) {
	md := NewMarkdownRenderer(nil, 80)
	blocks, err := md.Render("```math\nx^2 + y^2 = r^2\n```")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if len(blocks) == 0 {
		t.Error("should return at least 1 block")
	}
}

func TestP108_RenderFencedCode_LaTeXBlock(t *testing.T) {
	md := NewMarkdownRenderer(nil, 80)
	blocks, err := md.Render("```latex\n\\frac{a}{b}\n```")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if len(blocks) == 0 {
		t.Error("should return at least 1 block")
	}
}

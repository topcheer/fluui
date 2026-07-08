package markdown

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP118_NewHighlighterWithStyle_Empty(t *testing.T) {
	h := NewHighlighterWithStyle("")
	if h == nil {
		t.Fatal("expected non-nil highlighter for empty style")
	}
	// Should fall back to a default style
}

func TestP118_NewHighlighterWithStyle_Unknown(t *testing.T) {
	h := NewHighlighterWithStyle("nonexistent-style-xyz")
	if h == nil {
		t.Fatal("expected non-nil highlighter for unknown style (should fallback)")
	}
}

func TestP118_NewHighlighterWithStyle_Monokai(t *testing.T) {
	h := NewHighlighterWithStyle("monokai")
	if h == nil {
		t.Fatal("expected non-nil highlighter for monokai")
	}
}

func TestP118_MermaidDrawEdge_BTDirection(t *testing.T) {
	// Test bottom-to-top direction
	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 40)
	}
	from := &MermaidNode{ID: "A", X: 5, Y: 10, W: 6, H: 3}
	to := &MermaidNode{ID: "B", X: 5, Y: 0, W: 6, H: 3}
	edge := &MermaidEdge{From: "A", To: "B", Style: MermaidEdgeArrow}
	mermaidDrawEdge(canvas, from, to, edge, MermaidBT)
}

func TestP118_MermaidDrawEdge_RLDirection(t *testing.T) {
	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 60)
	}
	from := &MermaidNode{ID: "A", X: 30, Y: 5, W: 6, H: 3}
	to := &MermaidNode{ID: "B", X: 0, Y: 5, W: 6, H: 3}
	edge := &MermaidEdge{From: "A", To: "B", Style: MermaidEdgeArrow}
	mermaidDrawEdge(canvas, from, to, edge, MermaidRL)
}

func TestP118_MermaidDrawEdge_DottedTD(t *testing.T) {
	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 40)
	}
	from := &MermaidNode{ID: "A", X: 5, Y: 0, W: 6, H: 3}
	to := &MermaidNode{ID: "B", X: 5, Y: 10, W: 6, H: 3}
	edge := &MermaidEdge{From: "A", To: "B", Style: MermaidEdgeDotted}
	mermaidDrawEdge(canvas, from, to, edge, MermaidTD)
}

func TestP118_MermaidDrawEdge_ThickTD(t *testing.T) {
	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 40)
	}
	from := &MermaidNode{ID: "A", X: 5, Y: 0, W: 6, H: 3}
	to := &MermaidNode{ID: "B", X: 5, Y: 10, W: 6, H: 3}
	edge := &MermaidEdge{From: "A", To: "B", Style: MermaidEdgeThick}
	mermaidDrawEdge(canvas, from, to, edge, MermaidTD)
}

func TestP118_MermaidDrawEdge_HorizontalFromXGTToX(t *testing.T) {
	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 60)
	}
	from := &MermaidNode{ID: "A", X: 40, Y: 2, W: 6, H: 3}
	to := &MermaidNode{ID: "B", X: 5, Y: 8, W: 6, H: 3}
	edge := &MermaidEdge{From: "A", To: "B", Style: MermaidEdgeArrow}
	mermaidDrawEdge(canvas, from, to, edge, MermaidLR)
}

func TestP118_MermaidDrawEdge_HorizontalFromYGTToY(t *testing.T) {
	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 60)
	}
	from := &MermaidNode{ID: "A", X: 5, Y: 10, W: 6, H: 3}
	to := &MermaidNode{ID: "B", X: 30, Y: 0, W: 6, H: 3}
	edge := &MermaidEdge{From: "A", To: "B", Style: MermaidEdgeArrow}
	mermaidDrawEdge(canvas, from, to, edge, MermaidLR)
}

func TestP118_MermaidDrawEdge_WithLabel(t *testing.T) {
	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 40)
	}
	from := &MermaidNode{ID: "A", X: 5, Y: 0, W: 6, H: 3}
	to := &MermaidNode{ID: "B", X: 5, Y: 10, W: 6, H: 3}
	edge := &MermaidEdge{From: "A", To: "B", Style: MermaidEdgeArrow, Label: "yes"}
	mermaidDrawEdge(canvas, from, to, edge, MermaidTD)
}

func TestP118_MermaidDrawEdge_TD_DifferentX(t *testing.T) {
	// fromX != toX in vertical mode to exercise horizontal connector
	canvas := make([][]byte, 20)
	for i := range canvas {
		canvas[i] = make([]byte, 40)
	}
	from := &MermaidNode{ID: "A", X: 0, Y: 0, W: 6, H: 3}
	to := &MermaidNode{ID: "B", X: 15, Y: 10, W: 6, H: 3}
	edge := &MermaidEdge{From: "A", To: "B", Style: MermaidEdgeDotted}
	mermaidDrawEdge(canvas, from, to, edge, MermaidTD)
}

func TestP118_RenderMermaidText_WithEdges(t *testing.T) {
	text := `graph TD
    A[Start] --> B[Process]
    B -->|success| C[End]
    B -->|error| D[Retry]`
	cells := RenderMermaidText(text, nil)
	if cells == nil {
		t.Fatal("expected non-nil cells")
	}
}

func TestP118_HighlightToLines_Error(t *testing.T) {
	h := NewHighlighter()
	// Invalid language should still produce output (fallback lexer)
	lines, err := h.HighlightToLines("", "code")
	if err != nil {
		t.Errorf("expected no error for empty language, got %v", err)
	}
	if len(lines) == 0 {
		t.Error("expected at least 1 line")
	}
}

func TestP118_HighlightToLines_Empty(t *testing.T) {
	h := NewHighlighter()
	lines, err := h.HighlightToLines("go", "")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	// Empty source should produce at least 1 empty line
	if len(lines) == 0 {
		t.Error("expected at least 1 line for empty source")
	}
}

func TestP118_RenderInlineMath_NoMatch(t *testing.T) {
	result := RenderInlineMath("just plain text no math here")
	if result != "just plain text no math here" {
		t.Errorf("expected passthrough, got %q", result)
	}
}

func TestP118_HasInlineMath_Dollar(t *testing.T) {
	if !HasInlineMath("cost is $x^2$ today") {
		t.Error("expected true for $...$ math")
	}
	if HasInlineMath("cost is $5 today") {
		t.Error("expected false for lone dollar sign")
	}
}

func TestP118_HasInlineMath_Paren(t *testing.T) {
	if !HasInlineMath("formula is \\(x+1\\) here") {
		t.Error("expected true for \\(...\\) math")
	}
}

func TestP118_FormatOSC8_WithURL(t *testing.T) {
	result := FormatOSC8("click here", "https://example.com")
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestP118_FormatOSC8_EmptyURL(t *testing.T) {
	result := FormatOSC8("text", "")
	if result == "" {
		t.Error("expected non-empty even with empty URL")
	}
}

func TestP118_RenderMathToCells_Basic(t *testing.T) {
	fg := buffer.Color{Type: buffer.ColorNamed, Val: buffer.NamedCyan}
	cells := RenderMathToCells("\\alpha + \\beta", fg)
	if len(cells) == 0 {
		t.Error("expected non-empty cells")
	}
}

func TestP118_RenderMathToCells_Empty(t *testing.T) {
	fg := buffer.Color{Type: buffer.ColorNamed, Val: buffer.NamedCyan}
	cells := RenderMathToCells("", fg)
	// Empty input should produce no cells or minimal cells
	_ = cells // just verify no panic
}

func TestP118_RenderMathToCells_Greek(t *testing.T) {
	fg := buffer.Color{Type: buffer.ColorNamed, Val: buffer.NamedGreen}
	cells := RenderMathToCells("\\sum_{i=1}^{n} \\frac{1}{i^2} = \\frac{\\pi^2}{6}", fg)
	if len(cells) == 0 {
		t.Error("expected non-empty cells for Greek math")
	}
}

package markdown

import (
	"strings"
	"testing"
)

// --- ParseMermaid: basic parsing ---

func TestParseMermaid_Empty(t *testing.T) {
	g := ParseMermaid("")
	if g != nil {
		t.Error("expected nil for empty input")
	}
}

func TestParseMermaid_NotFlowchart(t *testing.T) {
	g := ParseMermaid("Hello World")
	if g != nil {
		t.Error("expected nil for non-flowchart input")
	}
}

func TestParseMermaid_DirectionTD(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A --> B")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if g.Direction != MermaidTD {
		t.Error("expected TD direction")
	}
}

func TestParseMermaid_DirectionLR(t *testing.T) {
	g := ParseMermaid("flowchart LR\n    A --> B")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if g.Direction != MermaidLR {
		t.Error("expected LR direction")
	}
}

func TestParseMermaid_DirectionTB(t *testing.T) {
	g := ParseMermaid("graph TB\n    A --> B")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if g.Direction != MermaidTD {
		t.Error("TB should be equivalent to TD")
	}
}

func TestParseMermaid_DirectionBT(t *testing.T) {
	g := ParseMermaid("flowchart BT\n    A --> B")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if g.Direction != MermaidBT {
		t.Error("expected BT direction")
	}
}

func TestParseMermaid_DirectionRL(t *testing.T) {
	g := ParseMermaid("flowchart RL\n    A --> B")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if g.Direction != MermaidRL {
		t.Error("expected RL direction")
	}
}

// --- Node parsing ---

func TestParseMermaid_SimpleNodes(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A --> B")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if len(g.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(g.Nodes))
	}
}

func TestParseMermaid_NodeShapes(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A[Rect] --> B(Round)\n    B --> C{Diamond}\n    C --> D((Circle))")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if len(g.Nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(g.Nodes))
	}

	// Check shapes
	if g.Nodes[0].Shape != MermaidShapeRect {
		t.Error("expected rect shape for A[Rect]")
	}
	if g.Nodes[1].Shape != MermaidShapeRounded {
		t.Error("expected rounded shape for B(Round)")
	}
	if g.Nodes[2].Shape != MermaidShapeDiamond {
		t.Error("expected diamond shape for C{Diamond}")
	}
	if g.Nodes[3].Shape != MermaidShapeCircle {
		t.Error("expected circle shape for D((Circle))")
	}
}

func TestParseMermaid_NodeLabels(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A[Start Here] --> B[End]")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if g.Nodes[0].Label != "Start Here" {
		t.Errorf("expected label 'Start Here', got %q", g.Nodes[0].Label)
	}
	if g.Nodes[1].Label != "End" {
		t.Errorf("expected label 'End', got %q", g.Nodes[1].Label)
	}
}

func TestParseMermaid_PlainNode(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if g.Nodes[0].Shape != MermaidShapePlain {
		t.Error("expected plain shape")
	}
	if g.Nodes[0].Label != "A" {
		t.Errorf("expected label 'A', got %q", g.Nodes[0].Label)
	}
}

// --- Edge parsing ---

func TestParseMermaid_EdgeArrow(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A --> B")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if len(g.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(g.Edges))
	}
	if g.Edges[0].Style != MermaidEdgeArrow {
		t.Error("expected arrow style")
	}
}

func TestParseMermaid_EdgeLine(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A --- B")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if len(g.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(g.Edges))
	}
	if g.Edges[0].Style != MermaidEdgeLine {
		t.Error("expected line style")
	}
}

func TestParseMermaid_EdgeDotted(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A -.-> B")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if len(g.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(g.Edges))
	}
	if g.Edges[0].Style != MermaidEdgeDotted {
		t.Error("expected dotted style")
	}
}

func TestParseMermaid_EdgeThick(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A ==> B")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if len(g.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(g.Edges))
	}
	if g.Edges[0].Style != MermaidEdgeThick {
		t.Error("expected thick style")
	}
}

func TestParseMermaid_EdgeLabel(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A -->|yes| B")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if len(g.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(g.Edges))
	}
	if g.Edges[0].Label != "yes" {
		t.Errorf("expected label 'yes', got %q", g.Edges[0].Label)
	}
}

func TestParseMermaid_MultipleEdges(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A --> B\n    B --> C\n    A --> C")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if len(g.Edges) != 3 {
		t.Fatalf("expected 3 edges, got %d", len(g.Edges))
	}
}

// --- RenderMermaid ---

func TestRenderMermaid_Simple(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A[Start] --> B[End]")
	cells := RenderMermaid(g, nil)
	if cells == nil {
		t.Fatal("expected non-nil cells")
	}
	if len(cells) == 0 {
		t.Fatal("expected at least 1 row")
	}
}

func TestRenderMermaid_LR(t *testing.T) {
	g := ParseMermaid("flowchart LR\n    A --> B --> C")
	cells := RenderMermaid(g, nil)
	if cells == nil {
		t.Fatal("expected non-nil cells")
	}
}

func TestRenderMermaid_AllShapes(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A[Rect] --> B(Round)\n    B --> C{Diamond}\n    C --> D((Circle))")
	cells := RenderMermaid(g, nil)
	if cells == nil {
		t.Fatal("expected non-nil cells")
	}
}

func TestRenderMermaid_EdgeLabels(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A -->|yes| B\n    A -->|no| C")
	cells := RenderMermaid(g, nil)
	if cells == nil {
		t.Fatal("expected non-nil cells")
	}
}

func TestRenderMermaid_NilGraph(t *testing.T) {
	cells := RenderMermaid(nil, nil)
	if cells != nil {
		t.Error("expected nil cells for nil graph")
	}
}

func TestRenderMermaid_EmptyGraph(t *testing.T) {
	g := &MermaidGraph{}
	cells := RenderMermaid(g, nil)
	if cells != nil {
		t.Error("expected nil cells for empty graph")
	}
}

func TestRenderMermaid_LargeGraph(t *testing.T) {
	// Stress test with many nodes
	input := "flowchart TD\n"
	for i := 0; i < 20; i++ {
		input += "    N" + itoaMermaid(i) + "[Node " + itoaMermaid(i) + "]"
		if i > 0 {
			// Just create sequential edges
		}
		input += "\n"
	}
	// Add some edges
	for i := 0; i < 19; i++ {
		input += "    N" + itoaMermaid(i) + " --> N" + itoaMermaid(i+1) + "\n"
	}
	g := ParseMermaid(input)
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	cells := RenderMermaid(g, nil)
	if cells == nil {
		t.Fatal("expected non-nil cells")
	}
}

// --- RenderMermaidText convenience ---

func TestRenderMermaidText_Valid(t *testing.T) {
	cells := RenderMermaidText("flowchart TD\n    A --> B", nil)
	if cells == nil {
		t.Fatal("expected non-nil cells")
	}
}

func TestRenderMermaidText_Invalid(t *testing.T) {
	cells := RenderMermaidText("not mermaid", nil)
	if cells != nil {
		t.Error("expected nil cells for invalid input")
	}
}

// --- parseMermaidNodeSpec ---

func TestParseMermaidNodeSpec_Plain(t *testing.T) {
	id, label, shape := parseMermaidNodeSpec("ABC")
	if id != "ABC" || label != "" || shape != MermaidShapePlain {
		t.Errorf("expected (ABC, '', Plain), got (%q, %q, %d)", id, label, shape)
	}
}

func TestParseMermaidNodeSpec_Rect(t *testing.T) {
	id, label, shape := parseMermaidNodeSpec("A[Hello]")
	if id != "A" || label != "Hello" || shape != MermaidShapeRect {
		t.Errorf("expected (A, Hello, Rect), got (%q, %q, %d)", id, label, shape)
	}
}

func TestParseMermaidNodeSpec_Rounded(t *testing.T) {
	id, label, shape := parseMermaidNodeSpec("B(World)")
	if id != "B" || label != "World" || shape != MermaidShapeRounded {
		t.Errorf("expected (B, World, Rounded), got (%q, %q, %d)", id, label, shape)
	}
}

func TestParseMermaidNodeSpec_Diamond(t *testing.T) {
	id, label, shape := parseMermaidNodeSpec("C{Choice}")
	if id != "C" || label != "Choice" || shape != MermaidShapeDiamond {
		t.Errorf("expected (C, Choice, Diamond), got (%q, %q, %d)", id, label, shape)
	}
}

func TestParseMermaidNodeSpec_Circle(t *testing.T) {
	id, label, shape := parseMermaidNodeSpec("D((Center))")
	if id != "D" || label != "Center" || shape != MermaidShapeCircle {
		t.Errorf("expected (D, Center, Circle), got (%q, %q, %d)", id, label, shape)
	}
}

// --- Complex scenarios ---

func TestParseMermaid_BranchingGraph(t *testing.T) {
	input := `flowchart TD
    A[Start] --> B{Is it?}
    B -->|Yes| C[OK]
    B -->|No| D[Error]`
	g := ParseMermaid(input)
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if len(g.Nodes) != 4 {
		t.Errorf("expected 4 nodes, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 3 {
		t.Errorf("expected 3 edges, got %d", len(g.Edges))
	}
}

func TestParseMermaid_LineComments(t *testing.T) {
	input := `flowchart TD
    %% This is a comment
    A --> B`
	g := ParseMermaid(input)
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	// Comment lines starting with %% should be skipped (or at least not break parsing)
	if len(g.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(g.Nodes))
	}
}

func TestParseMermaid_BlankLines(t *testing.T) {
	input := `flowchart TD

    A --> B

    B --> C`
	g := ParseMermaid(input)
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if len(g.Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 2 {
		t.Errorf("expected 2 edges, got %d", len(g.Edges))
	}
}

func TestParseMermaid_NodeReuse(t *testing.T) {
	// Same node appears in multiple edges — should not create duplicates
	g := ParseMermaid("flowchart TD\n    A --> B\n    A --> C")
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if len(g.Nodes) != 3 {
		t.Errorf("expected 3 nodes (A, B, C), got %d", len(g.Nodes))
	}
}

// --- Render output verification ---

func TestRenderMermaid_ContainsLabels(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A[Hello] --> B[World]")
	cells := RenderMermaid(g, nil)
	if cells == nil {
		t.Fatal("expected non-nil cells")
	}
	// Convert to text and check for labels
	var text string
	for _, row := range cells {
		for _, cell := range row {
			if cell.Rune != 0 && cell.Rune != ' ' {
				text += string(cell.Rune)
			}
		}
		text += "\n"
	}
	if !strings.Contains(text, "Hello") {
		t.Error("expected rendered output to contain 'Hello'")
	}
	if !strings.Contains(text, "World") {
		t.Error("expected rendered output to contain 'World'")
	}
}

func TestRenderMermaid_ContainsArrowChar(t *testing.T) {
	g := ParseMermaid("flowchart TD\n    A --> B")
	cells := RenderMermaid(g, nil)
	if cells == nil {
		t.Fatal("expected non-nil cells")
	}
	// Should contain vertical bars (|) for TD arrows
	var found bool
	for _, row := range cells {
		for _, cell := range row {
			if cell.Rune == '|' || cell.Rune == 'v' {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("expected arrow characters in rendered output")
	}
}

// --- Benchmark ---

func BenchmarkParseMermaid_Simple(b *testing.B) {
	input := "flowchart TD\n    A[Start] --> B{Decision}\n    B -->|Yes| C[OK]\n    B -->|No| D[Error]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseMermaid(input)
	}
}

func BenchmarkRenderMermaid_Simple(b *testing.B) {
	g := ParseMermaid("flowchart TD\n    A[Start] --> B{Decision}\n    B -->|Yes| C[OK]\n    B -->|No| D[Error]")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderMermaid(g, nil)
	}
}

// itoaMermaid converts int to string for test IDs.
func itoaMermaid(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

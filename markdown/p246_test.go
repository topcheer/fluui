package markdown

import "testing"

// P246: parseMermaidNode empty, mermaidDrawNode empty canvas, edge styles

func TestParseMermaidNode_Empty_P246(t *testing.T) {
	g := &MermaidGraph{}
	n := parseMermaidNode("", g, map[string]*MermaidNode{})
	if n != nil {
		t.Error("empty text should return nil")
	}
}

func TestMermaidDrawNode_EmptyCanvas_P246(t *testing.T) {
	n := &MermaidNode{ID: "A", Label: "A", X: 0, Y: 0}
	mermaidDrawNode(nil, n) // empty canvas → return
	mermaidDrawNode([][]byte{}, n) // zero rows → return
}

func TestRenderMermaid_DottedEdge_P246(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	r.Render("```mermaid\nA -.-> B\n```")
}

func TestRenderMermaid_ThickEdge_P246(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	r.Render("```mermaid\nA ==> B\n```")
}

func TestRenderMermaid_VerticalLayout_P246(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	r.Render("```mermaid\ngraph TD\nA --> B\nB --> C\n```")
}

func TestRenderMermaid_LabeledEdge_P246(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	r.Render("```mermaid\nA -->|hello| B\n```")
}

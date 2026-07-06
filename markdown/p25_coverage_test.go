package markdown

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P25 coverage tests for markdown renderer edge cases.
// File naming: p25_coverage_test.go

func TestP25_RenderEmptyString(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("")
	if err != nil {
		t.Fatalf("empty string error: %v", err)
	}
	if len(blocks) != 0 {
		t.Errorf("empty string should produce 0 blocks, got %d", len(blocks))
	}
}

func TestP25_RenderVeryLongLine(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	long := strings.Repeat("a", 500)
	blocks, err := r.Render(long)
	if err != nil {
		t.Fatalf("long line error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("should produce at least one block")
	}
}

func TestP25_RenderNestedLists(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	input := "- item 1\n  - nested 1\n  - nested 2\n- item 2\n"
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatalf("nested lists error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("should produce blocks for nested lists")
	}
}

func TestP25_RenderMixedMarkdown(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	input := `# Title

Some **bold** and *italic* text.

## Subsection

- list item 1
- list item 2

> A blockquote

` + "```go\nfmt.Println(\"hello\")\n```\n"
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatalf("mixed markdown error: %v", err)
	}
	if len(blocks) < 4 {
		t.Errorf("mixed markdown should produce 4+ blocks, got %d", len(blocks))
	}
}

func TestP25_RenderMalformedCodeBlock(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	// Unclosed code block
	input := "```python\nprint('hello')\n"
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatalf("malformed code block error: %v", err)
	}
	// Should still produce blocks
	if len(blocks) == 0 {
		t.Fatal("malformed code block should still produce blocks")
	}
}

func TestP25_RenderHorizontalRule(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	input := "before\n\n---\n\nafter\n"
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatalf("horizontal rule error: %v", err)
	}
	if len(blocks) < 2 {
		t.Errorf("horizontal rule should produce 2+ blocks, got %d", len(blocks))
	}
}

func TestP25_WrapCellsZeroWidth(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	// Cells with zero width should be handled gracefully
	cells := []buffer.Cell{
		buffer.NewCell('A', buffer.Style{}),
		buffer.NewCell(0, buffer.Style{}),
		buffer.NewCell('B', buffer.Style{}),
	}
	result := r.wrapCells(cells, 2)
	if len(result) == 0 {
		t.Fatal("wrapCells should produce at least one line")
	}
}

func TestP25_WrapCellsZeroWidth2(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	// Zero width parameter
	cells := []buffer.Cell{
		buffer.NewCell('A', buffer.Style{}),
		buffer.NewCell('B', buffer.Style{}),
	}
	result := r.wrapCells(cells, 0)
	if len(result) != 1 {
		t.Errorf("wrapCells with width=0 should return single line, got %d lines", len(result))
	}
}

func TestP25_WrapCellsWordBreak(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	// Test word break on space
	cells := make([]buffer.Cell, 0, 20)
	for _, c := range "hello world foo bar" {
		cells = append(cells, buffer.NewCell(c, buffer.Style{}))
	}
	result := r.wrapCells(cells, 10)
	if len(result) < 2 {
		t.Errorf("wrapCells should produce 2+ lines for long text, got %d", len(result))
	}
}

func TestP25_CellLineWidth(t *testing.T) {
	// Test internal helper
	cells := []buffer.Cell{
		buffer.NewCell('A', buffer.Style{}),
		buffer.NewCell('B', buffer.Style{}),
		buffer.NewCell('C', buffer.Style{}),
	}
	w := cellLineWidth(cells)
	if w != 3 {
		t.Errorf("cellLineWidth = %d, want 3", w)
	}
}

func TestP25_LastSpaceCellNotFound(t *testing.T) {
	// No space in cells — should return -1
	cells := []buffer.Cell{
		buffer.NewCell('A', buffer.Style{}),
		buffer.NewCell('B', buffer.Style{}),
	}
	idx, _ := lastSpaceCellAndWidth(cells)
	if idx != -1 {
		t.Errorf("lastSpaceCellAndWidth = %d, want -1 (no space)", idx)
	}
}

func TestP25_LastSpaceCellFound(t *testing.T) {
	cells := []buffer.Cell{
		buffer.NewCell('A', buffer.Style{}),
		buffer.NewCell(' ', buffer.Style{}),
		buffer.NewCell('B', buffer.Style{}),
	}
	idx, afterW := lastSpaceCellAndWidth(cells)
	if idx != 1 {
		t.Errorf("lastSpaceCellAndWidth idx = %d, want 1", idx)
	}
	if afterW != 1 { // 'B' has width 1
		t.Errorf("lastSpaceCellAndWidth afterWidth = %d, want 1", afterW)
	}
}

func TestP25_RenderLink(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	input := "[Fluui](https://github.com/fluui)\n"
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatalf("link render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("link should produce blocks")
	}
}

func TestP25_RenderMultipleParagraphs(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	input := "Para 1\n\nPara 2\n\nPara 3\n\nPara 4\n"
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatalf("paragraphs error: %v", err)
	}
	if len(blocks) < 3 {
		t.Errorf("4 paragraphs should produce 3+ blocks, got %d", len(blocks))
	}
}

func TestP25_HighlighterWithStyle(t *testing.T) {
	h := NewHighlighterWithStyle("monokai")
	if h == nil {
		t.Fatal("NewHighlighterWithStyle should return non-nil")
	}
}

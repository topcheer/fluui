package markdown

import (
	"testing"
)

func renderMD_P278(src string, t *testing.T) []*Block {
	r := NewMarkdownRenderer(nil, 80)
	blocks, err := r.Render(src)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	return blocks
}

func TestRenderList_Ordered_P278(t *testing.T) {
	blocks := renderMD_P278("1. First\n2. Second\n3. Third\n", t)
	if len(blocks) == 0 {
		t.Error("should produce blocks for ordered list")
	}
}

func TestRenderList_Unordered_P278(t *testing.T) {
	blocks := renderMD_P278("- Alpha\n- Beta\n- Gamma\n", t)
	if len(blocks) == 0 {
		t.Error("should produce blocks for unordered list")
	}
}

func TestRenderList_TaskList_P278(t *testing.T) {
	blocks := renderMD_P278("- [ ] Todo\n- [x] Done\n", t)
	if len(blocks) == 0 {
		t.Error("should produce blocks for task list")
	}
}

func TestRenderList_Nested_P278(t *testing.T) {
	blocks := renderMD_P278("- Top\n  - Nested\n  - Deep\n- Back\n", t)
	if len(blocks) == 0 {
		t.Error("should produce blocks for nested list")
	}
}

func TestRenderFootnoteList_P278(t *testing.T) {
	blocks := renderMD_P278("Text with[^1] footnote.\n\n[^1]: This is the footnote content.\n", t)
	// Footnote rendering may or may not produce separate blocks
	_ = blocks
}

func TestLatex_ConsumeGroup_P278(t *testing.T) {
	blocks := renderMD_P278("$\\frac{a+1}{b}$\n", t)
	_ = blocks
}

func TestLatex_ConsumeGroup_DeepNested_P278(t *testing.T) {
	blocks := renderMD_P278("$\\sqrt{{a+b}^{2}}$\n", t)
	_ = blocks
}

func TestRenderList_Wrapped_P278(t *testing.T) {
	r := NewMarkdownRenderer(nil, 20)
	blocks, _ := r.Render("- This is a very long item that should wrap across multiple lines\n- Short\n")
	if len(blocks) == 0 {
		t.Error("should produce blocks for wrapping list")
	}
}

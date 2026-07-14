package block

import (
	"encoding/json"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// P188: Target AssistantTextBlock.Paint 66.7% → 80%+
func TestP188_AssistantText_PaintAllMarkdown(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"headers", "# H1\n## H2\n### H3"},
		{"code_block", "```go\nfunc main() {}\n```"},
		{"ordered_list", "1. First\n2. Second\n3. Third"},
		{"unordered_list", "- Item 1\n- Item 2"},
		{"nested_list", "- Top\n  - Nested\n  - Deep"},
		{"task_list", "- [ ] Todo\n- [x] Done"},
		{"table", "| A | B |\n|---|---|\n| 1 | 2 |"},
		{"bold_italic", "**bold** and *italic*"},
		{"blockquote", "> Quote text"},
		{"alert", "> [!NOTE]\n> Note text"},
		{"link", "[click](https://example.com)"},
		{"image", "![alt](https://example.com/img.png)"},
		{"math_inline", "The formula $x^2 + y^2 = r^2$ is true"},
		{"math_block", "$$\nE = mc^2\n$$"},
		{"mermaid", "```mermaid\nA --> B\n```"},
		{"strikethrough", "~~old text~~"},
		{"horizontal_rule", "Above\n\n---\n\nBelow"},
		{"unicode", "日本語テスト αβγδ → ✓"},
		{"multiline_para", "First paragraph.\n\nSecond paragraph.\n\nThird."},
		{"mixed_content", "# Title\n\nSome **bold** text with `code`.\n\n- List item\n\n```\ncode block\n```"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewAssistantTextBlock("test")
			b.SetContent(tt.text)
			b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
			b.Measure(component.Constraints{MaxWidth: 60, MaxHeight: 20})
			b.Paint(buffer.NewBuffer(60, 20))
		})
	}
}

func TestP188_AssistantText_PaintEdgeCases(t *testing.T) {
	// empty content
	b := NewAssistantTextBlock("empty")
	b.SetContent("")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	b.Paint(buffer.NewBuffer(60, 20))

	// zero bounds
	b2 := NewAssistantTextBlock("zero")
	b2.SetContent("some text")
	b2.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	b2.Paint(buffer.NewBuffer(1, 1))

	// narrow width
	b3 := NewAssistantTextBlock("narrow")
	b3.SetContent("this is a very long line that needs wrapping")
	b3.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 20})
	b3.Paint(buffer.NewBuffer(5, 20))

	// height truncation
	b4 := NewAssistantTextBlock("truncate")
	b4.SetContent("line1\nline2\nline3\nline4\nline5")
	b4.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 2})
	b4.Paint(buffer.NewBuffer(60, 2))

	// non-zero offset
	b5 := NewAssistantTextBlock("offset")
	b5.SetContent("text content")
	b5.SetBounds(component.Rect{X: 5, Y: 3, W: 60, H: 20})
	b5.Paint(buffer.NewBuffer(70, 25))
}

func TestP188_AssistantText_CacheInvalidation(t *testing.T) {
	b := NewAssistantTextBlock("cache")
	b.SetContent("initial content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	b.Paint(buffer.NewBuffer(60, 20))
	// content change invalidates
	b.SetContent("changed content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	b.Paint(buffer.NewBuffer(60, 20))
	// width change invalidates
	b.SetContent("more content here")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 20})
	b.Paint(buffer.NewBuffer(40, 20))
}

func TestP188_AssistantText_SerializeDeserialize(t *testing.T) {
	b := NewAssistantTextBlock("ser")
	b.SetContent("# Test\n\nContent here")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	b.Paint(buffer.NewBuffer(60, 20))

	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}

	b2 := NewAssistantTextBlock("ser")
	if err := b2.DeserializeState(data); err != nil {
		t.Fatalf("deserialize: %v", err)
	}

	// invalid JSON
	if err := b2.DeserializeState(json.RawMessage(`{invalid}`)); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestP188_ThinkingBlock_PreviewLineEdges(t *testing.T) {
	b := NewThinkingBlock("think")
	b.SetContent("short")
	if p := b.PreviewLineUnlocked(10); p == "" {
		t.Error("expected non-empty preview")
	}

	b2 := NewThinkingBlock("think2")
	b2.SetContent("this is a very long thinking content that exceeds the preview length limit")
	if p := b2.PreviewLineUnlocked(10); len([]rune(p)) > 13 {
		t.Errorf("preview too long: %d runes", len([]rune(p)))
	}

	b3 := NewThinkingBlock("think3")
	b3.SetContent("")
	if p := b3.PreviewLineUnlocked(10); p != "" {
		t.Errorf("expected empty preview, got %q", p)
	}

	b4 := NewThinkingBlock("think4")
	b4.SetContent("exact")
	if p := b4.PreviewLineUnlocked(5); p != "exact" {
		t.Errorf("expected 'exact', got %q", p)
	}
}

func TestP188_BaseBlock_Paint(t *testing.T) {
	bb := BaseBlock{}
	bb.Paint(buffer.NewBuffer(10, 5))
}

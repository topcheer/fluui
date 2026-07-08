package block

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// === AssistantTextBlock.Paint (66.7% → 85%+) ===

func TestP134_AssistantText_Paint_HeightTruncation(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta(strings.Repeat("line\n", 20))
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 3})
	buf := buffer.NewBuffer(80, 3)
	b.Paint(buf)
}

func TestP134_AssistantText_Paint_WidthTruncation(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("very long text that exceeds the narrow bounds width")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 10})
	buf := buffer.NewBuffer(5, 10)
	b.Paint(buf)
}

func TestP134_AssistantText_Paint_ZeroBounds(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(0, 0)
	b.Paint(buf)
}

func TestP134_AssistantText_Paint_EmptyContent(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP134_AssistantText_Paint_MarkdownHeaders(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("# H1\n## H2\n### H3\n\nText.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 15})
	buf := buffer.NewBuffer(80, 15)
	b.Paint(buf)
}

func TestP134_AssistantText_Paint_MarkdownCodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("```go\nfmt.Println(\"hi\")\n```")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP134_AssistantText_Paint_MarkdownLists(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("- a\n- b\n  - nested\n- c\n\n1. first\n2. second")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 15})
	buf := buffer.NewBuffer(80, 15)
	b.Paint(buf)
}

func TestP134_AssistantText_Paint_MarkdownTable(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("| N | V |\n|---|---|\n| A | 1 |\n| B | 2 |")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 15})
	buf := buffer.NewBuffer(80, 15)
	b.Paint(buf)
}

func TestP134_AssistantText_Paint_MarkdownBoldItalic(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("**bold** and *italic* and `code`.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP134_AssistantText_Paint_Blockquote(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("> quote\n> line2")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP134_AssistantText_Paint_Links(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("See [link](https://example.com).")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP134_AssistantText_Paint_Unicode(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("Hello 世界 — café — 日本語")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP134_AssistantText_Paint_MultiParagraph(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("Para 1.\n\nPara 2.\n\nPara 3.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 15})
	buf := buffer.NewBuffer(80, 15)
	b.Paint(buf)
}

func TestP134_AssistantText_Paint_HorizontalRule(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("Before\n\n---\n\nAfter")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP134_AssistantText_Paint_NestedLists(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("- outer\n  - inner\n    - deep\n- back")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 15})
	buf := buffer.NewBuffer(80, 15)
	b.Paint(buf)
}

func TestP134_AssistantText_Paint_NonZeroOffset(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("content")
	b.SetBounds(component.Rect{X: 5, Y: 3, W: 40, H: 5})
	buf := buffer.NewBuffer(80, 24)
	b.Paint(buf)
}

func TestP134_AssistantText_Paint_NarrowWidth(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("long line wrapping at narrow width")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 15})
	buf := buffer.NewBuffer(10, 15)
	b.Paint(buf)
}

func TestP134_AssistantText_CacheInvalidation(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("initial content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
	b.AppendDelta(" more content")
	b.Paint(buf)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	b.Paint(buf)
}

func TestP134_AssistantText_SerializeDeserialize(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("markdown **content**")
	state, err := b.SerializeState()
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}
	b2 := NewAssistantTextBlock("test2")
	if err := b2.DeserializeState(state); err != nil {
		t.Fatalf("deserialize: %v", err)
	}
}

func TestP134_AssistantText_DeserializeInvalidJSON(t *testing.T) {
	b := NewAssistantTextBlock("test")
	if err := b.DeserializeState(json.RawMessage(`{invalid}`)); err == nil {
		t.Error("expected error")
	}
}

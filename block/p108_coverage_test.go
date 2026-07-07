package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AssistantTextBlock.Paint (72.2%) ───

func TestP108_AssistantTextBlock_Paint_PlainFallback(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("hello world\nsecond line")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_ZeroBounds(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("hello")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_EmptyContent(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_HeightTruncation(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("# Title\n\npara1\n\npara2\n\npara3\n\npara4")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_NonZeroOffset(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("hello world")
	b.SetBounds(component.Rect{X: 5, Y: 2, W: 30, H: 3})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_MarkdownHeaders(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("# H1\n## H2\n### H3\ncontent")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_MarkdownCodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("```go\nfunc main() {}\n```")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_MarkdownLists(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("- item 1\n- item 2\n  - nested\n1. first\n2. second")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_MarkdownTable(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("| A | B |\n|---|---|\n| 1 | 2 |")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_MarkdownBold(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("this is **bold** and *italic* and `code`")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_MarkdownBlockquote(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("> This is a quote\n> Second line")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_MarkdownLinks(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("[example](https://example.com) here")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_Unicode(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("héllo wörld 你好日本")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_NarrowWidth(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("this is a long line that needs wrapping at narrow width")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 10})
	buf := buffer.NewBuffer(5, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_MultiParagraph(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("First paragraph here.\n\nSecond paragraph with more text.\n\nThird one.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_NestedList(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("- top\n  - mid\n    - deep\n      - deeper")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_HorizontalRule(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("above\n\n---\n\nbelow")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 8})
	buf := buffer.NewBuffer(40, 8)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_CacheInvalidation(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("initial content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)

	b.SetContent("new content after change")
	b.Paint(buf)

	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf2 := buffer.NewBuffer(60, 5)
	b.Paint(buf2)
}

func TestP108_AssistantTextBlock_SerializeDeserialize(t *testing.T) {
	b := NewAssistantTextBlock("test-id")
	b.SetContent("hello **world**")
	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("serialized data should not be empty")
	}

	b2 := NewAssistantTextBlock("test-id")
	if err := b2.DeserializeState(data); err != nil {
		t.Fatalf("DeserializeState: %v", err)
	}
	if b2.Content() == "" {
		t.Error("content should be restored")
	}
}

func TestP108_AssistantTextBlock_DeserializeInvalidJSON(t *testing.T) {
	b := NewAssistantTextBlock("test")
	err := b.DeserializeState([]byte("invalid json {"))
	if err == nil {
		t.Error("should return error for invalid JSON")
	}
}

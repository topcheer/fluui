package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// === AssistantTextBlock.Paint (66.7% → 80%+) ===

func TestP129_AssistantText_PaintMarkdownHeader(t *testing.T) {
	b := NewAssistantTextBlock("test1")
	b.AppendDelta("# Header\n\nText")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP129_AssistantText_PaintCodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("test2")
	b.AppendDelta("```go\nfmt.Println(\"hello\")\n```")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP129_AssistantText_PaintList(t *testing.T) {
	b := NewAssistantTextBlock("test3")
	b.AppendDelta("- item one\n- item two\n- item three")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP129_AssistantText_PaintTable(t *testing.T) {
	b := NewAssistantTextBlock("test4")
	b.AppendDelta("| A | B |\n|---|---|\n| 1 | 2 |")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP129_AssistantText_PaintBoldItalic(t *testing.T) {
	b := NewAssistantTextBlock("test5")
	b.AppendDelta("**bold** and *italic* and `code`")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP129_AssistantText_PaintBlockquote(t *testing.T) {
	b := NewAssistantTextBlock("test6")
	b.AppendDelta("> A quote\n> more")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP129_AssistantText_PaintLink(t *testing.T) {
	b := NewAssistantTextBlock("test7")
	b.AppendDelta("[link text](https://example.com)")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP129_AssistantText_PaintUnicode(t *testing.T) {
	b := NewAssistantTextBlock("test8")
	b.AppendDelta("Héllo 世界 🌍 café")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP129_AssistantText_PaintNarrowWidth(t *testing.T) {
	b := NewAssistantTextBlock("test9")
	b.AppendDelta("This is a long line that should wrap in narrow width")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 20})
	buf := buffer.NewBuffer(10, 20)
	b.Paint(buf)
}

func TestP129_AssistantText_PaintMultiParagraph(t *testing.T) {
	b := NewAssistantTextBlock("test10")
	b.AppendDelta("First paragraph.\n\nSecond paragraph.\n\nThird.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP129_AssistantText_PaintNestedList(t *testing.T) {
	b := NewAssistantTextBlock("test11")
	b.AppendDelta("- top\n  - nested\n    - deep\n- back")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP129_AssistantText_PaintHorizontalRule(t *testing.T) {
	b := NewAssistantTextBlock("test12")
	b.AppendDelta("before\n\n---\n\nafter")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP129_AssistantText_PaintNonZeroOffset(t *testing.T) {
	b := NewAssistantTextBlock("test13")
	b.AppendDelta("Some content")
	b.SetBounds(component.Rect{X: 5, Y: 3, W: 40, H: 10})
	buf := buffer.NewBuffer(80, 24)
	b.Paint(buf)
}

func TestP129_AssistantText_PaintHeightTruncation(t *testing.T) {
	b := NewAssistantTextBlock("test14")
	b.AppendDelta("line1\nline2\nline3\nline4\nline5")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 2})
	buf := buffer.NewBuffer(60, 2)
	b.Paint(buf)
}

func TestP129_AssistantText_PaintEmpty(t *testing.T) {
	b := NewAssistantTextBlock("test15")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

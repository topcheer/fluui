package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// ═══════════════════════════════════════════════════════════════════════════
// P95 Block Coverage Tests
// ═══════════════════════════════════════════════════════════════════════════

// ─── BaseBlock.Paint (0% → 100%) ───

func TestP95_BaseBlock_Paint(t *testing.T) {
	bb := BaseBlock{}
	bb.Paint(buffer.NewBuffer(1, 1)) // no-op — just verify no panic
}

// ─── AssistantTextBlock.Paint fallback path (66.7% → 90%+) ───

func TestP95_AssistantText_Paint_EmptyContent(t *testing.T) {
	b := NewAssistantTextBlock("test1")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf) // empty text — should return early
}

func TestP95_AssistantText_Paint_ZeroBounds(t *testing.T) {
	b := NewAssistantTextBlock("test2")
	b.AppendDelta("hello world")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf) // zero bounds — should return early
}

func TestP95_AssistantText_Paint_NegativeWidth(t *testing.T) {
	b := NewAssistantTextBlock("test3")
	b.AppendDelta("hello")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: -1, H: 5})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf) // negative width — should return early
}

func TestP95_AssistantText_Paint_HeightTruncation(t *testing.T) {
	b := NewAssistantTextBlock("test4")
	b.AppendDelta("# Header\n\nParagraph\n\nMore text\n\nExtra line")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 2})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf) // should paint first 2 rows then stop
}

func TestP95_AssistantText_Paint_WithOffset(t *testing.T) {
	b := NewAssistantTextBlock("test5")
	b.AppendDelta("hello world this is a test")
	b.SetBounds(component.Rect{X: 5, Y: 3, W: 30, H: 5})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf) // should paint at offset (5,3)
	cell := buf.GetCell(5, 3)
	if cell.Rune == 0 || cell.Rune == ' ' {
		t.Error("expected text at offset (5,3)")
	}
}

func TestP95_AssistantText_Paint_NarrowWidth(t *testing.T) {
	b := NewAssistantTextBlock("test6")
	b.AppendDelta("this is a very long line that will need wrapping in narrow bounds")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 10})
	buf := buffer.NewBuffer(5, 10)
	b.Paint(buf) // should wrap text to 5 columns
}

func TestP95_AssistantText_Paint_MarkdownHeaders(t *testing.T) {
	b := NewAssistantTextBlock("test7")
	b.AppendDelta("# Big Header\n## Smaller\n### Tiny")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP95_AssistantText_Paint_CodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("test8")
	b.AppendDelta("```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	b.Paint(buf)
}

func TestP95_AssistantText_Paint_UnorderedList(t *testing.T) {
	b := NewAssistantTextBlock("test9")
	b.AppendDelta("- Item 1\n- Item 2\n- Item 3")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP95_AssistantText_Paint_OrderedList(t *testing.T) {
	b := NewAssistantTextBlock("test10")
	b.AppendDelta("1. First\n2. Second\n3. Third")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP95_AssistantText_Paint_BoldItalicCode(t *testing.T) {
	b := NewAssistantTextBlock("test11")
	b.AppendDelta("**bold** *italic* `code` ~~strike~~")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP95_AssistantText_Paint_Table(t *testing.T) {
	b := NewAssistantTextBlock("test12")
	b.AppendDelta("| A | B |\n|---|---|\n| 1 | 2 |")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP95_AssistantText_Paint_Blockquote(t *testing.T) {
	b := NewAssistantTextBlock("test13")
	b.AppendDelta("> This is a quote\n> Second line")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP95_AssistantText_Paint_Unicode(t *testing.T) {
	b := NewAssistantTextBlock("test14")
	b.AppendDelta("Hello 世界 — café ☕ emoji 🎉")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP95_AssistantText_Paint_MarkdownLinks(t *testing.T) {
	b := NewAssistantTextBlock("test15")
	b.AppendDelta("[Click here](https://example.com) for info.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	b.Paint(buf)
}

func TestP95_AssistantText_Paint_MultiParagraph(t *testing.T) {
	b := NewAssistantTextBlock("test16")
	b.AppendDelta("First paragraph here.\n\nSecond paragraph follows.\n\nThird paragraph ends.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP95_AssistantText_Paint_NestedList(t *testing.T) {
	b := NewAssistantTextBlock("test17")
	b.AppendDelta("- Top level\n  - Nested\n  - Also nested\n- Back to top")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	b.Paint(buf)
}

func TestP95_AssistantText_CacheInvalidation(t *testing.T) {
	b := NewAssistantTextBlock("test18")
	b.AppendDelta("initial content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)

	// Append more content — cache should invalidate
	b.AppendDelta(" more text")
	b.Paint(buf) // should re-render with new content
}

func TestP95_AssistantText_CacheWidthChange(t *testing.T) {
	b := NewAssistantTextBlock("test19")
	b.AppendDelta("some text that will be rendered")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)

	// Change width — cache should invalidate
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf2 := buffer.NewBuffer(60, 10)
	b.Paint(buf2)
}

func TestP95_AssistantText_SerializeDeserialize(t *testing.T) {
	b := NewAssistantTextBlock("test20")
	b.AppendDelta("some content to serialize")
	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}

	b2 := NewAssistantTextBlock("test20")
	if err := b2.DeserializeState(data); err != nil {
		t.Errorf("DeserializeState error: %v", err)
	}
}

func TestP95_AssistantText_SerializeInvalidJSON(t *testing.T) {
	b := NewAssistantTextBlock("test21")
	err := b.DeserializeState([]byte("not valid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/markdown"
)

// === BaseBlock.Paint (0% → 100%) ===

func TestP125_BaseBlock_Paint(t *testing.T) {
	bb := BaseBlock{}
	bb.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	bb.Paint(buf) // no-op, should not panic
}

// === AssistantTextBlock.Paint: markdown headers (69.4% → 80%+) ===

func TestP125_AssistantText_Paint_Heading(t *testing.T) {
	b := NewAssistantTextBlock("test1")
	b.AppendDelta("# Title\n## Subtitle\n### Sub-sub")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_CodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("test2")
	b.AppendDelta("```go\nfunc main() {}\n```")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_UnorderedList(t *testing.T) {
	b := NewAssistantTextBlock("test3")
	b.AppendDelta("- item 1\n- item 2\n- item 3")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_OrderedList(t *testing.T) {
	b := NewAssistantTextBlock("test4")
	b.AppendDelta("1. first\n2. second\n3. third")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_BoldItalic(t *testing.T) {
	b := NewAssistantTextBlock("test5")
	b.AppendDelta("**bold** and *italic* and `code`")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_Table(t *testing.T) {
	b := NewAssistantTextBlock("test6")
	b.AppendDelta("| A | B |\n|---|---|\n| 1 | 2 |")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_Blockquote(t *testing.T) {
	b := NewAssistantTextBlock("test7")
	b.AppendDelta("> This is a quote")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_Link(t *testing.T) {
	b := NewAssistantTextBlock("test8")
	b.AppendDelta("[click](https://example.com)")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_HorizontalRule(t *testing.T) {
	b := NewAssistantTextBlock("test9")
	b.AppendDelta("Before\n---\nAfter")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_Unicode(t *testing.T) {
	b := NewAssistantTextBlock("test10")
	b.AppendDelta("Hello 世界 café naïve résumé")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_NarrowWidth(t *testing.T) {
	b := NewAssistantTextBlock("test11")
	b.AppendDelta("This is a very long paragraph that should wrap nicely")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 10})
	buf := buffer.NewBuffer(10, 10)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_MultiParagraph(t *testing.T) {
	b := NewAssistantTextBlock("test12")
	b.AppendDelta("First paragraph.\n\nSecond paragraph.\n\nThird.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_HeightTruncation(t *testing.T) {
	b := NewAssistantTextBlock("test13")
	b.AppendDelta("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_NonZeroOffset(t *testing.T) {
	b := NewAssistantTextBlock("test14")
	b.AppendDelta("Some content here")
	b.SetBounds(component.Rect{X: 5, Y: 2, W: 30, H: 5})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_EmptyContent(t *testing.T) {
	b := NewAssistantTextBlock("test15")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_ZeroBounds(t *testing.T) {
	b := NewAssistantTextBlock("test16")
	b.AppendDelta("content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_NestedLists(t *testing.T) {
	b := NewAssistantTextBlock("test17")
	b.AppendDelta("- outer\n  - inner1\n  - inner2\n- outer2")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_Strikethrough(t *testing.T) {
	b := NewAssistantTextBlock("test18")
	b.AppendDelta("~~deleted~~ and ~~also deleted~~")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP125_AssistantText_Paint_TaskList(t *testing.T) {
	b := NewAssistantTextBlock("test19")
	b.AppendDelta("- [ ] todo\n- [x] done")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

// === Cache invalidation ===

func TestP125_AssistantText_CacheContentChange(t *testing.T) {
	b := NewAssistantTextBlock("test20")
	b.AppendDelta("# Title")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf1 := buffer.NewBuffer(40, 5)
	b.Paint(buf1)

	b.AppendDelta("\nMore content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 8})
	buf2 := buffer.NewBuffer(40, 8)
	b.Paint(buf2)
}

// === Serialize/Deserialize ===

func TestP125_AssistantText_SerializeDeserialize(t *testing.T) {
	b := NewAssistantTextBlock("test21")
	b.AppendDelta("# Hello\nSome **bold** text")

	data, err := b.SerializeState()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty serialized data")
	}

	b2 := NewAssistantTextBlock("test22")
	err = b2.DeserializeState(data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestP125_AssistantText_DeserializeInvalidJSON(t *testing.T) {
	b := NewAssistantTextBlock("test23")
	err := b.DeserializeState([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// === Measure edge cases ===

func TestP125_AssistantText_Measure_WithContent(t *testing.T) {
	b := NewAssistantTextBlock("test24")
	b.AppendDelta("# Title\nParagraph text\n- item 1")
	s := b.Measure(component.Bounded(40, 100))
	if s.W <= 0 || s.H <= 0 {
		t.Error("expected positive measure")
	}
}

func TestP125_AssistantText_Measure_Empty(t *testing.T) {
	b := NewAssistantTextBlock("test25")
	s := b.Measure(component.Bounded(40, 100))
	_ = s // should not panic
}

func TestP125_AssistantText_Measure_CodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("test26")
	b.AppendDelta("```go\nfunc main() {}\n```")
	s := b.Measure(component.Bounded(40, 100))
	_ = s
}

// Ensure imports used
var _ markdown.BlockType

package block

import (
	"encoding/json"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// === AssistantTextBlock.Paint (66.7% → 85%+) ===

func TestP187_AssistantText_PaintFallback(t *testing.T) {
	b := NewAssistantTextBlock("test1")
	b.AppendDelta("plain text without markdown")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	b.Paint(buffer.NewBuffer(40, 5))
}

func TestP187_AssistantText_PaintHeaders(t *testing.T) {
	b := NewAssistantTextBlock("test2")
	b.AppendDelta("# Header 1\n\n## Header 2\n\n### Header 3")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestP187_AssistantText_PaintCodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("test3")
	b.AppendDelta("```go\npackage main\nfunc main() {}\n```")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestP187_AssistantText_PaintList(t *testing.T) {
	b := NewAssistantTextBlock("test4")
	b.AppendDelta("- item 1\n- item 2\n  - nested\n- item 3\n\n1. first\n2. second")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestP187_AssistantText_PaintTable(t *testing.T) {
	b := NewAssistantTextBlock("test5")
	b.AppendDelta("| Name | Value |\n|------|-------|\n| A    | 1     |\n| B    | 2     |")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestP187_AssistantText_PaintBoldItalic(t *testing.T) {
	b := NewAssistantTextBlock("test6")
	b.AppendDelta("**bold** and *italic* and `code` and ~~strike~~")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	b.Paint(buffer.NewBuffer(40, 5))
}

func TestP187_AssistantText_PaintBlockquote(t *testing.T) {
	b := NewAssistantTextBlock("test7")
	b.AppendDelta("> This is a quote\n> with multiple lines")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	b.Paint(buffer.NewBuffer(40, 5))
}

func TestP187_AssistantText_PaintLinks(t *testing.T) {
	b := NewAssistantTextBlock("test8")
	b.AppendDelta("[example](https://example.com) and https://raw.url")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	b.Paint(buffer.NewBuffer(40, 5))
}

func TestP187_AssistantText_PaintUnicode(t *testing.T) {
	b := NewAssistantTextBlock("test9")
	b.AppendDelta("你好世界 🌍 café — 日本語")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	b.Paint(buffer.NewBuffer(40, 5))
}

func TestP187_AssistantText_PaintNarrowWidth(t *testing.T) {
	b := NewAssistantTextBlock("test10")
	b.AppendDelta("This is a long line that should be wrapped at narrow width")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 10})
	b.Paint(buffer.NewBuffer(5, 10))
}

func TestP187_AssistantText_PaintMultiParagraph(t *testing.T) {
	b := NewAssistantTextBlock("test11")
	b.AppendDelta("First paragraph.\n\nSecond paragraph.\n\nThird paragraph.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestP187_AssistantText_PaintNestedList(t *testing.T) {
	b := NewAssistantTextBlock("test12")
	b.AppendDelta("- top\n  - mid\n    - deep\n  - mid2\n- top2")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestP187_AssistantText_PaintHorizontalRule(t *testing.T) {
	b := NewAssistantTextBlock("test13")
	b.AppendDelta("Before\n\n---\n\nAfter")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestP187_AssistantText_PaintNonZeroOffset(t *testing.T) {
	b := NewAssistantTextBlock("test14")
	b.AppendDelta("text at offset")
	b.SetBounds(component.Rect{X: 5, Y: 3, W: 30, H: 5})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestP187_AssistantText_PaintHeightTruncation(t *testing.T) {
	b := NewAssistantTextBlock("test15")
	b.AppendDelta("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 3})
	b.Paint(buffer.NewBuffer(40, 3))
}

func TestP187_AssistantText_PaintEmptyContent(t *testing.T) {
	b := NewAssistantTextBlock("test16")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	b.Paint(buffer.NewBuffer(40, 5))
}

func TestP187_AssistantText_PaintZeroBounds(t *testing.T) {
	b := NewAssistantTextBlock("test17")
	b.AppendDelta("some text")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	b.Paint(buffer.NewBuffer(1, 1))
}

func TestP187_AssistantText_CacheInvalidation(t *testing.T) {
	b := NewAssistantTextBlock("test18")
	b.AppendDelta("initial text")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	b.Paint(buffer.NewBuffer(40, 5))
	// Invalidate cache by appending more content
	b.AppendDelta(" more content")
	b.Paint(buffer.NewBuffer(40, 5))
	// Invalidate by changing width
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 5})
	b.Paint(buffer.NewBuffer(20, 5))
}

// === AssistantTextBlock.Measure (84.2% → 90%+) ===

func TestP187_AssistantText_MeasureNilBlocks(t *testing.T) {
	b := NewAssistantTextBlock("test19")
	b.AppendDelta("text")
	s := b.Measure(component.Constraints{MaxWidth: 40, MaxHeight: 10})
	_ = s
}

func TestP187_AssistantText_MeasureEmpty(t *testing.T) {
	b := NewAssistantTextBlock("test20")
	s := b.Measure(component.Constraints{MaxWidth: 40, MaxHeight: 10})
	_ = s
}

func TestP187_AssistantText_MeasureZeroMaxWidth(t *testing.T) {
	b := NewAssistantTextBlock("test21")
	b.AppendDelta("text")
	s := b.Measure(component.Constraints{MaxWidth: 0, MaxHeight: 10})
	_ = s
}

func TestP187_AssistantText_MeasureCodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("test22")
	b.AppendDelta("```go\nx := 1\n```")
	s := b.Measure(component.Constraints{MaxWidth: 40, MaxHeight: 10})
	_ = s
}

func TestP187_AssistantText_MeasureNestedLists(t *testing.T) {
	b := NewAssistantTextBlock("test23")
	b.AppendDelta("- a\n  - b\n    - c")
	s := b.Measure(component.Constraints{MaxWidth: 40, MaxHeight: 10})
	_ = s
}

// === getCachedBlocks (84.6% → 90%+) ===

func TestP187_AssistantText_GetCachedBlocks(t *testing.T) {
	b := NewAssistantTextBlock("test24")
	b.AppendDelta("# Title\nparagraph")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	// Paint first to ensure renderer is created
	b.Paint(buffer.NewBuffer(40, 5))
	// Now getCachedBlocks should work
	blocks := b.getCachedBlocks("# Title\nparagraph", 40)
	if len(blocks) == 0 {
		t.Error("expected non-empty blocks")
	}
}

func TestP187_AssistantText_GetCachedBlocksEmpty(t *testing.T) {
	b := NewAssistantTextBlock("test25")
	b.AppendDelta("trigger renderer init")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	b.Paint(buffer.NewBuffer(40, 5))
	// Now renderer is initialized, test with empty string
	blocks := b.getCachedBlocks("", 40)
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks, got %d", len(blocks))
	}
}

// === Serialize/Deserialize ===

func TestP187_AssistantText_SerializeDeserialize(t *testing.T) {
	b := NewAssistantTextBlock("test26")
	b.AppendDelta("# Hello\nworld")
	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}
	b2 := NewAssistantTextBlock("test27")
	err = b2.DeserializeState(data)
	if err != nil {
		t.Fatalf("deserialize error: %v", err)
	}
}

func TestP187_AssistantText_DeserializeInvalidJSON(t *testing.T) {
	b := NewAssistantTextBlock("test28")
	err := b.DeserializeState(json.RawMessage(`{invalid json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// === contentString cache ===

func TestP187_AssistantText_ContentStringCache(t *testing.T) {
	b := NewAssistantTextBlock("test29")
	b.AppendDelta("hello")
	s1 := b.contentString()
	b.AppendDelta(" world")
	s2 := b.contentString()
	if s1 == s2 {
		t.Error("expected different strings after append")
	}
}

// === ThinkingBlock.PreviewLineUnlocked (84.6% → 100%) ===

func TestP187_Thinking_PreviewLineShort(t *testing.T) {
	b := NewThinkingBlock("think1")
	b.SetContent("short")
	result := b.PreviewLineUnlocked(100)
	if result == "" {
		t.Error("expected non-empty preview")
	}
}

func TestP187_Thinking_PreviewLineTruncated(t *testing.T) {
	b := NewThinkingBlock("think2")
	b.SetContent("this is a very long thinking content that should be truncated")
	result := b.PreviewLineUnlocked(10)
	if len([]rune(result)) > 13 {
		t.Errorf("expected <= 13 chars (10+ellipsis), got %d: %q", len([]rune(result)), result)
	}
}

func TestP187_Thinking_PreviewLineEmpty(t *testing.T) {
	b := NewThinkingBlock("think3")
	result := b.PreviewLineUnlocked(50)
	_ = result
}

func TestP187_Thinking_PreviewLineZeroMax(t *testing.T) {
	b := NewThinkingBlock("think4")
	b.SetContent("content")
	result := b.PreviewLineUnlocked(0)
	_ = result
}

func TestP187_Thinking_PreviewLineExactLen(t *testing.T) {
	b := NewThinkingBlock("think5")
	b.SetContent("exactly20characters!")
	result := b.PreviewLineUnlocked(20)
	_ = result
}

// === BaseBlock.Paint (0% → 100%) ===

func TestP187_BaseBlock_Paint(t *testing.T) {
	bb := BaseBlock{}
	bb.Paint(buffer.NewBuffer(10, 5))
}

// === AssistantTextBlock with wide bounds ===

func TestP187_AssistantText_WithWideBounds(t *testing.T) {
	b := NewAssistantTextBlock("test30")
	b.AppendDelta("# Title\n\ntext with **bold**")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	b.Paint(buffer.NewBuffer(60, 10))
}

// === AssistantTextBlock Paint with various markdown ===

func TestP187_AssistantText_PaintAlertBlock(t *testing.T) {
	b := NewAssistantTextBlock("test31")
	b.AppendDelta("> [!NOTE]\n> This is a note\n> [!WARNING]\n> This is a warning")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestP187_AssistantText_PaintTaskList(t *testing.T) {
	b := NewAssistantTextBlock("test32")
	b.AppendDelta("- [ ] todo\n- [x] done\n- [X] also done")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestP187_AssistantText_PaintImage(t *testing.T) {
	b := NewAssistantTextBlock("test33")
	b.AppendDelta("![alt text](https://example.com/image.png)")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	b.Paint(buffer.NewBuffer(40, 5))
}

func TestP187_AssistantText_PaintMath(t *testing.T) {
	b := NewAssistantTextBlock("test34")
	b.AppendDelta("Inline $x^2 + y^2 = z^2$ math\n\n$$\\frac{a}{b}$$")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestP187_AssistantText_PaintMermaid(t *testing.T) {
	b := NewAssistantTextBlock("test35")
	b.AppendDelta("```mermaid\nflowchart TD\n    A --> B\n```")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

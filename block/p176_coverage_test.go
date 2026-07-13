package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestP176_Container_SetVisibleRange(t *testing.T) {
	c := NewBlockContainer()
	c.SetVisibleRange(10, 20)
}

func TestP176_Container_AutoScroll(t *testing.T) {
	c := NewBlockContainer()
	if c.AutoScrollEnabled() {
		t.Error("expected auto-scroll disabled by default")
	}
	c.SetAutoScroll(true)
	if !c.AutoScrollEnabled() {
		t.Error("expected auto-scroll enabled")
	}
	c.SetAutoScroll(false)
	if c.AutoScrollEnabled() {
		t.Error("expected auto-scroll disabled")
	}
}

func TestP176_Container_ScrollToBottom(t *testing.T) {
	c := NewBlockContainer()
	c.ScrollToBottom()
}

func TestP176_Container_ScrollToBottomWithBlocks(t *testing.T) {
	c := NewBlockContainer()
	b1 := NewAssistantTextBlock("test1")
	b1.AppendDelta("content 1")
	c.AddBlock(b1)
	b2 := NewAssistantTextBlock("test2")
	b2.AppendDelta("content 2")
	c.AddBlock(b2)
	c.ScrollToBottom()
}

func TestP176_BaseBlock_Paint(t *testing.T) {
	b := BaseBlock{}
	b.Paint(nil)
}

func TestP176_AssistantTextBlock_PaintMarkdown(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("# Heading\n\nSome **bold** text with `code`.\n\n- item 1\n- item 2\n\n| Col1 | Col2 |\n|------|------|\n| a    | b    |\n\n> A quote\n\n1. First\n2. Second\n\n---\n\nFinal paragraph.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 30})
	buf := buffer.NewBuffer(80, 30)
	b.Paint(buf)
}

func TestP176_AssistantTextBlock_PaintUnicode(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("中文测试 — café — naïve — 日本語 — 한국어")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 5})
	buf := buffer.NewBuffer(80, 5)
	b.Paint(buf)
}

func TestP176_AssistantTextBlock_PaintNarrow(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("This is a long line that should be wrapped in a narrow width")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 10})
	buf := buffer.NewBuffer(5, 10)
	b.Paint(buf)
}

func TestP176_AssistantTextBlock_PaintEmpty(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 5})
	buf := buffer.NewBuffer(80, 5)
	b.Paint(buf)
}

func TestP176_AssistantTextBlock_PaintZeroBounds(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	b.Paint(buf)
}

func TestP176_AssistantTextBlock_PaintNonZeroOffset(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("line1\nline2\nline3")
	b.SetBounds(component.Rect{X: 5, Y: 3, W: 80, H: 10})
	buf := buffer.NewBuffer(90, 15)
	b.Paint(buf)
}

func TestP176_AssistantTextBlock_PaintHeightTruncation(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 3})
	buf := buffer.NewBuffer(80, 3)
	b.Paint(buf)
}

func TestP176_AssistantTextBlock_CacheInvalidation(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("initial content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 5})
	buf := buffer.NewBuffer(80, 5)
	b.Paint(buf)
	b.AppendDelta(" more content")
	b.Paint(buf)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	b.Paint(buf)
}

func TestP176_AssistantTextBlock_Serialize(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("test content")
	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty serialized data")
	}
}

func TestP176_AssistantTextBlock_Deserialize(t *testing.T) {
	b1 := NewAssistantTextBlock("test")
	b1.AppendDelta("test content")
	data, _ := b1.SerializeState()
	b2 := NewAssistantTextBlock("test2")
	err := b2.DeserializeState(data)
	if err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
}

func TestP176_AssistantTextBlock_DeserializeInvalid(t *testing.T) {
	b := NewAssistantTextBlock("test")
	err := b.DeserializeState([]byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

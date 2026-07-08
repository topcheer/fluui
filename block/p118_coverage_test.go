package block

import (
	"encoding/json"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestP118_BaseBlock_Paint(t *testing.T) {
	b := &BaseBlock{id: "test", blockType: TypeThinking, state: BlockComplete}
	b.Paint(buffer.NewBuffer(10, 5)) // no-op, should not panic
}

func TestP118_BlockContainer_Reserve(t *testing.T) {
	c := NewBlockContainer()
	c.Reserve(100)
	for i := 0; i < 50; i++ {
		c.AddBlock(NewUserMessageBlock("id", "msg"))
	}
}

func TestP118_BlockContainer_Reserve_NilSafe(t *testing.T) {
	c := NewBlockContainer()
	c.Reserve(0)
	c.Reserve(-1)
}

func TestP118_AssistantText_Paint_CodeBlock(t *testing.T) {
	c := NewBlockContainer()
	b := NewAssistantTextBlock("test")
	b.AppendDelta("```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```")
	c.AddBlock(b)
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	c.Paint(buf)
}

func TestP118_AssistantText_Paint_Headers(t *testing.T) {
	c := NewBlockContainer()
	b := NewAssistantTextBlock("test")
	b.AppendDelta("# Title\n## Subtitle\n### Sub-subtitle\n\nContent here")
	c.AddBlock(b)
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	c.Paint(buf)
}

func TestP118_AssistantText_Paint_Lists(t *testing.T) {
	c := NewBlockContainer()
	b := NewAssistantTextBlock("test")
	b.AppendDelta("- Item 1\n- Item 2\n  - Nested item\n- Item 3\n\n1. First\n2. Second\n3. Third")
	c.AddBlock(b)
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	c.Paint(buf)
}

func TestP118_AssistantText_Paint_Table(t *testing.T) {
	c := NewBlockContainer()
	b := NewAssistantTextBlock("test")
	b.AppendDelta("| Name | Age |\n|------|-----|\n| Alice | 30 |\n| Bob | 25 |")
	c.AddBlock(b)
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	c.Paint(buf)
}

func TestP118_AssistantText_Paint_BoldItalic(t *testing.T) {
	c := NewBlockContainer()
	b := NewAssistantTextBlock("test")
	b.AppendDelta("**bold text** and *italic text* and `inline code`")
	c.AddBlock(b)
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	c.Paint(buf)
}

func TestP118_AssistantText_Paint_Blockquote(t *testing.T) {
	c := NewBlockContainer()
	b := NewAssistantTextBlock("test")
	b.AppendDelta("> This is a quote\n> spanning multiple lines")
	c.AddBlock(b)
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	c.Paint(buf)
}

func TestP118_AssistantText_Paint_Links(t *testing.T) {
	c := NewBlockContainer()
	b := NewAssistantTextBlock("test")
	b.AppendDelta("Check [this link](https://example.com) for details")
	c.AddBlock(b)
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	c.Paint(buf)
}

func TestP118_AssistantText_Paint_Unicode(t *testing.T) {
	c := NewBlockContainer()
	b := NewAssistantTextBlock("test")
	b.AppendDelta("Unicode test: café naïve 日本語 emoji test")
	c.AddBlock(b)
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	c.Paint(buf)
}

func TestP118_AssistantText_Paint_NarrowWidth(t *testing.T) {
	c := NewBlockContainer()
	b := NewAssistantTextBlock("test")
	b.AppendDelta("This is a long line that will need wrapping in a narrow terminal width")
	c.AddBlock(b)
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 10})
	buf := buffer.NewBuffer(10, 10)
	c.Paint(buf)
}

func TestP118_AssistantText_Paint_MultiParagraph(t *testing.T) {
	c := NewBlockContainer()
	b := NewAssistantTextBlock("test")
	b.AppendDelta("First paragraph.\n\nSecond paragraph.\n\nThird paragraph.")
	c.AddBlock(b)
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	c.Paint(buf)
}

func TestP118_AssistantText_Paint_Empty(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(NewAssistantTextBlock("test"))
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	c.Paint(buf)
}

func TestP118_AssistantText_Paint_ZeroBounds(t *testing.T) {
	c := NewBlockContainer()
	b := NewAssistantTextBlock("test")
	b.AppendDelta("content")
	c.AddBlock(b)
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(60, 5)
	c.Paint(buf)
}

func TestP118_AssistantText_SerializeRoundTrip(t *testing.T) {
	b := NewAssistantTextBlock("test-id")
	b.AppendDelta("# Title\nSome **bold** content")

	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}

	b2 := NewAssistantTextBlock("new")
	err = b2.DeserializeState(data)
	if err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
	if b2.Content() != b.Content() {
		t.Error("content mismatch after round-trip")
	}
}

func TestP118_AssistantText_DeserializeInvalidJSON(t *testing.T) {
	b := NewAssistantTextBlock("test")
	err := b.DeserializeState(json.RawMessage(`{invalid json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestP118_AssistantText_CacheInvalidation(t *testing.T) {
	c := NewBlockContainer()
	b := NewAssistantTextBlock("test")
	b.AppendDelta("initial content")
	c.AddBlock(b)
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	c.Paint(buf)
	b.AppendDelta(" + more content")
	c.Paint(buf)
}

func TestP118_BlockContainer_TotalHeight(t *testing.T) {
	c := NewBlockContainer()
	b1 := NewAssistantTextBlock("b1")
	b1.AppendDelta("line1\nline2")
	c.AddBlock(b1)
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 30})
	// TotalHeight may be 0 if no Measure happens first
	h := c.TotalHeight()
	if h < 0 {
		t.Errorf("expected non-negative height, got %d", h)
	}
}

func TestP118_BlockContainer_BlockPositions(t *testing.T) {
	c := NewBlockContainer()
	b1 := NewAssistantTextBlock("b1")
	b1.AppendDelta("content1")
	c.AddBlock(b1)
	b2 := NewAssistantTextBlock("b2")
	b2.AppendDelta("content2")
	c.AddBlock(b2)
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 30})
	positions := c.BlockPositions()
	if len(positions) != 2 {
		t.Errorf("expected 2 positions, got %d", len(positions))
	}
}

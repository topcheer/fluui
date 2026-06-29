package block

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// P29d coverage tests for block package — Paint, Measure, Serialize.

// === AssistantTextBlock Paint edge cases ===

func TestP29d_AssistantText_Paint_Empty(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.Paint(buf) // empty content — should return early
}

func TestP29d_AssistantText_Paint_ZeroBounds(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("Hello")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(80, 24)
	b.Paint(buf) // zero bounds — should return early
}

func TestP29d_AssistantText_Paint_ClippedHeight(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("Line 1\nLine 2\nLine 3\nLine 4")
	b.Complete()
	// Set very small height — should clip
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 2})
	buf := buffer.NewBuffer(80, 24)
	b.Paint(buf) // should only paint first 2 lines
}

func TestP29d_AssistantText_Paint_Markdown(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("# Title\n\n**bold** text\n\n- item\n- item")
	b.Complete()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.Paint(buf)
	// Verify content was rendered
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 || cell.Rune == ' ' {
		t.Error("expected non-empty cell from markdown render")
	}
}

func TestP29d_AssistantText_Paint_MultiFrame(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("Hello world")
	b.Complete()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)

	// Paint multiple times — cache should work
	b.Paint(buf)
	b.Paint(buf)
	b.Paint(buf)
}

// === AssistantTextBlock Measure edge cases ===

func TestP29d_AssistantText_Measure_Empty(t *testing.T) {
	b := NewAssistantTextBlock("test")
	size := b.Measure(component.Constraints{})
	if size.H != 1 {
		t.Errorf("empty content should be H=1, got %d", size.H)
	}
}

func TestP29d_AssistantText_Measure_ZeroWidth(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("Hello world test")
	b.Complete()
	size := b.Measure(component.Constraints{MaxWidth: 0})
	// Should default to 80
	if size.W != 80 {
		t.Errorf("zero MaxWidth should default to 80, got %d", size.W)
	}
}

func TestP29d_AssistantText_Measure_Markdown(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("# Title\n\nParagraph text here.")
	b.Complete()
	size := b.Measure(component.Constraints{MaxWidth: 80})
	if size.H < 2 {
		t.Errorf("expected at least 2 lines for heading + paragraph, got %d", size.H)
	}
}

// === AssistantTextBlock cache invalidation ===

func TestP29d_AssistantText_CacheInvalidation(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("First")
	b.Complete()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf) // populates cache

	// SetContent should invalidate cache
	b.SetContent("Second content here")
	b.Paint(buf) // should re-render with new content
}

// === Serialize roundtrip ===

func TestP29d_Serialize_AssistantText(t *testing.T) {
	b := NewAssistantTextBlock("blk1")
	b.AppendDelta("Test content")
	b.Complete()

	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState: %v", err)
	}

	b2 := NewAssistantTextBlock("blk2")
	if err := b2.DeserializeState(data); err != nil {
		t.Fatalf("DeserializeState: %v", err)
	}

	// Content should match
	// Compare via Measure
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	b2.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	if b.Measure(component.Unbounded()) != b2.Measure(component.Unbounded()) {
		t.Error("serialized and deserialized blocks should have same measure")
	}
}

func TestP29d_Serialize_AssistantText_BadJSON(t *testing.T) {
	b := NewAssistantTextBlock("blk1")
	err := b.DeserializeState(json.RawMessage(`{invalid json`))
	if err == nil {
		t.Error("should error on invalid JSON")
	}
}

// === SaveContainer / LoadContainer ===

func TestP29d_SaveLoad_Roundtrip(t *testing.T) {
	c := NewBlockContainer()
	b1 := NewAssistantTextBlock("a1")
	b1.AppendDelta("Content A")
	b1.Complete()
	c.AddBlock(b1)

	b2 := NewUserMessageBlock("u1", "Hello user")
	c.AddBlock(b2)

	r := NewRegistry()
	r.Register("assistant_text", func(id string) Block {
		return NewAssistantTextBlock(id)
	})
	r.Register("user_message", func(id string) Block {
		return NewUserMessageBlock(id, "")
	})

	data, err := SaveContainer(c, r)
	if err != nil {
		t.Fatalf("SaveContainer: %v", err)
	}

	if !strings.Contains(string(data), "assistant_text") {
		t.Error("serialized data should contain type name")
	}

	c2, err := LoadContainer(data, r)
	if err != nil {
		t.Fatalf("LoadContainer: %v", err)
	}

	blocks := c2.Blocks()
	if len(blocks) != 2 {
		t.Errorf("expected 2 blocks, got %d", len(blocks))
	}
}

func TestP29d_SaveContainer_NilRegistry(t *testing.T) {
	c := NewBlockContainer()
	b := NewAssistantTextBlock("a1")
	b.AppendDelta("Test")
	b.Complete()
	c.AddBlock(b)

	data, err := SaveContainer(c, nil)
	if err != nil {
		t.Fatalf("SaveContainer with nil registry: %v", err)
	}
	if data == nil {
		t.Error("should return non-nil data")
	}
}

func TestP29d_SaveContainer_Empty(t *testing.T) {
	c := NewBlockContainer()
	r := NewRegistry()
	data, err := SaveContainer(c, r)
	if err != nil {
		t.Fatalf("SaveContainer empty: %v", err)
	}
	if !strings.Contains(string(data), `"version"`) {
		t.Error("should contain version field")
	}
}

// === Block type String() ===

func TestP29d_BlockType_String(t *testing.T) {
	tests := []struct {
		typ  BlockType
		want string
	}{
		{TypeAssistantText, "assistant_text"},
		{TypeUserMessage, "user_message"},
		{TypeThinking, "thinking"},
		{TypeToolCall, "tool_call"},
		{TypeToolResult, "tool_result"},
		{TypeError, "error"},
	}
	for _, tc := range tests {
		got := tc.typ.String()
		if got != tc.want {
			t.Errorf("BlockType(%d).String() = %q, want %q", tc.typ, got, tc.want)
		}
	}
}

// === ThinkingBlock Paint ===

func TestP29d_Thinking_Paint_Collapsed(t *testing.T) {
	b := NewThinkingBlock("think1")
	b.AppendDelta("Some thinking content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.Paint(buf) // collapsed by default
}

func TestP29d_Thinking_Paint_Expanded(t *testing.T) {
	b := NewThinkingBlock("think1")
	b.AppendDelta("Deep thoughts about the problem")
	b.Toggle() // expand
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.Paint(buf) // should paint expanded
}

func TestP29d_Thinking_Paint_Empty(t *testing.T) {
	b := NewThinkingBlock("think1")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.Paint(buf) // no content — should not panic
}

// === ErrorBlock serialize ===

func TestP29d_ErrorBlock_SerializeRoundtrip(t *testing.T) {
	b := NewErrorBlockWithMessage("err1", "test error message")
	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState: %v", err)
	}

	b2 := NewErrorBlock("err2")
	if err := b2.DeserializeState(data); err != nil {
		t.Fatalf("DeserializeState: %v", err)
	}
}

func TestP29d_ErrorBlock_DeserializeBadJSON(t *testing.T) {
	b := NewErrorBlock("err1")
	err := b.DeserializeState(json.RawMessage(`{bad json`))
	if err == nil {
		t.Error("should error on invalid JSON")
	}
}

// === ToolCallBlock serialize ===

func TestP29d_ToolCall_SerializeRoundtrip(t *testing.T) {
	b := NewToolCallBlock("tc1", "search", `{"query": "test"}`)
	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState: %v", err)
	}

	b2 := NewToolCallBlock("tc2", "", "")
	if err := b2.DeserializeState(data); err != nil {
		t.Fatalf("DeserializeState: %v", err)
	}
}

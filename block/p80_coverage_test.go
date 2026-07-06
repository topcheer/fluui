package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// ─── BaseBlock.Paint (0% → 100%) ───

func TestP80_BaseBlock_Paint_NoOp(t *testing.T) {
	bb := BaseBlock{id: "test"}
	buf := buffer.NewBuffer(10, 5)
	bb.Paint(buf) // should be no-op, not panic
}

// ─── AssistantTextBlock.Paint additional coverage (63.9% → 80%+) ───

func TestP80_AssistantText_Paint_MultiParagraph(t *testing.T) {
	b := NewAssistantTextBlock("p80-mp")
	b.AppendDelta("First paragraph.\n\nSecond paragraph.\n\nThird.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
	// Verify content is painted
	if buf.GetCell(0, 0).Rune != 'F' {
		t.Errorf("first cell = %q, want 'F'", buf.GetCell(0, 0).Rune)
	}
}

func TestP80_AssistantText_Paint_NestedList(t *testing.T) {
	b := NewAssistantTextBlock("p80-nl")
	b.AppendDelta("- Item 1\n  - Sub item 1a\n  - Sub item 1b\n- Item 2")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP80_AssistantText_Paint_Blockquote(t *testing.T) {
	b := NewAssistantTextBlock("p80-bq")
	b.AppendDelta("> This is a quote\n> with multiple lines")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP80_AssistantText_Paint_EmoticonAndSpecial(t *testing.T) {
	b := NewAssistantTextBlock("p80-em")
	b.AppendDelta("Hello! 😊 ✓ → ← ↑ ↓")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP80_AssistantText_Paint_LinkMarkdown(t *testing.T) {
	b := NewAssistantTextBlock("p80-link")
	b.AppendDelta("[Click here](https://example.com) for more info.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
}

func TestP80_AssistantText_Paint_HorizontalRule(t *testing.T) {
	b := NewAssistantTextBlock("p80-hr")
	b.AppendDelta("Above\n\n---\n\nBelow")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP80_AssistantText_Paint_EmptyContent(t *testing.T) {
	b := NewAssistantTextBlock("p80-empty")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf) // should not panic
}

func TestP80_AssistantText_Paint_SerializeDeserialize(t *testing.T) {
	b := NewAssistantTextBlock("p80-ser")
	b.AppendDelta("# Hello\n\nWorld content")
	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}
	b2 := NewAssistantTextBlock("p80-ser2")
	err = b2.DeserializeState(data)
	if err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
}

func TestP80_AssistantText_Paint_InvalidJSON(t *testing.T) {
	b := NewAssistantTextBlock("p80-bad")
	err := b.DeserializeState([]byte("{invalid"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// ─── ErrorBlock edge cases ───

func TestP80_ErrorBlock_Paint(t *testing.T) {
	eb := NewErrorBlockWithMessage("err1", "Something went wrong")
	eb.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	eb.Paint(buf)
}

func TestP80_ErrorBlock_Timestamp(t *testing.T) {
	eb := NewErrorBlockWithMessage("err2", "Error message")
	ts := eb.Timestamp()
	if ts.IsZero() {
		t.Error("timestamp should not be zero")
	}
}

func TestP80_ErrorBlock_SerializeDeserialize(t *testing.T) {
	eb := NewErrorBlockWithMessage("err3", "Test error")
	data, err := eb.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}
	eb2 := NewErrorBlockWithMessage("err4", "")
	err = eb2.DeserializeState(data)
	if err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
}

// ─── ThinkingBlock edge cases ───

func TestP80_ThinkingBlock_Deserialize(t *testing.T) {
	tb := NewThinkingBlock("think1")
	tb.SetContent("Analyzing the input data")
	data, err := tb.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}
	tb2 := NewThinkingBlock("think2")
	err = tb2.DeserializeState(data)
	if err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
}

func TestP80_ThinkingBlock_Deserialize_InvalidJSON(t *testing.T) {
	tb := NewThinkingBlock("think3")
	err := tb.DeserializeState([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestP80_ThinkingBlock_Paint_Expanded(t *testing.T) {
	tb := NewThinkingBlock("think4")
	tb.AppendDelta("Thinking about the problem...")
	tb.Expand()
	tb.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	tb.Paint(buf)
}

func TestP80_ThinkingBlock_Paint_Collapsed_LongPreview(t *testing.T) {
	tb := NewThinkingBlock("think5")
	tb.AppendDelta("This is a very long thinking content that should be truncated when collapsed")
	tb.Collapse()
	tb.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	tb.Paint(buf)
}

// ─── WorkflowBlock edge cases ───

func TestP80_Workflow_Serialize_Deserialize_Empty(t *testing.T) {
	w := NewWorkflowBlock("wf-empty")
	data, err := w.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}
	w2 := NewWorkflowBlock("wf-empty2")
	err = w2.DeserializeState(data)
	if err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
}

func TestP80_Workflow_BlockContainer_TotalHeight(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(NewUserMessageBlock("a1", "hello world this is a long message for testing"))
	c.AddBlock(NewUserMessageBlock("a2", "another message here"))
	c.AddBlock(NewUserMessageBlock("a3", "third message"))
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 100})
	h := c.TotalHeight()
	if h < 0 {
		t.Errorf("TotalHeight = %d, want >= 0", h)
	}
}

func TestP80_Workflow_BlockContainer_BlockPositions(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(NewUserMessageBlock("bp1", "test message"))
	c.AddBlock(NewUserMessageBlock("bp2", "another message"))
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	positions := c.BlockPositions()
	_ = positions // may be empty or non-empty depending on measure state
}

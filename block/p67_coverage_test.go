package block

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// === BaseBlock.InitMu ===

func TestP67_BaseBlock_InitMu(t *testing.T) {
	b := &BaseBlock{}
	b.InitMu()
	if b.mu == nil {
		t.Error("expected mutex after InitMu")
	}
	// Calling again should be safe
	b.InitMu()
}

// === BaseBlock.Paint (no-op) ===

func TestP67_BaseBlock_Paint(t *testing.T) {
	b := NewBaseBlock("test", TypeUserMessage)
	buf := buffer.NewBuffer(10, 5)
	b.Paint(buf) // should be a no-op, not panic
}

// === Registry.TypeNameForBlock ===

func TestP67_Registry_TypeNameForBlock(t *testing.T) {
	r := NewDefaultRegistry()
	tb := NewAssistantTextBlock("test")
	name := r.TypeNameForBlock(tb)
	if name == "" {
		t.Error("expected non-empty type name for AssistantTextBlock")
	}
}

func TestP67_Registry_TypeNameForBlock_Unknown(t *testing.T) {
	r := NewRegistry() // empty registry
	tb := NewAssistantTextBlock("test")
	name := r.TypeNameForBlock(tb)
	if name != "" {
		t.Error("expected empty name for unregistered block type")
	}
}

// === ErrorBlock.Timestamp ===

func TestP67_ErrorBlock_Timestamp(t *testing.T) {
	eb := NewErrorBlock("test error")
	ts := eb.Timestamp()
	if ts.IsZero() {
		t.Error("expected non-zero timestamp")
	}
	// Should be recent
	if time.Since(ts) > 1*time.Second {
		t.Error("expected recent timestamp")
	}
}

// === ErrorBlock.Measure default width ===

func TestP67_ErrorBlock_MeasureDefaultWidth(t *testing.T) {
	eb := NewErrorBlock("something went wrong")
	s := eb.Measure(component.Constraints{}) // no MaxWidth
	if s.W != 80 {
		t.Errorf("expected default W=80, got %d", s.W)
	}
}

// === ThinkingBlock.SetContent ===

func TestP67_ThinkingBlock_SetContent(t *testing.T) {
	tb := NewThinkingBlock("test")
	tb.SetContent("I think therefore I am")
	// Content should be set — verify via Measure (expanded)
	tb.Toggle() // expand
	s := tb.Measure(component.Constraints{MaxWidth: 80})
	if s.H <= 1 {
		t.Error("expected H > 1 after setting content")
	}
}

// === AssistantTextBlock.Paint edge cases (63.9% → 80%+) ===

func TestP67_AssistantText_PaintEmpty(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf) // should not panic with empty content
}

func TestP67_AssistantText_PaintShortText(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("Hello world")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
	// Should have content at (0,0)
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 || cell.Rune == ' ' {
		t.Error("expected non-empty content")
	}
}

func TestP67_AssistantText_PaintZeroBounds(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("test")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 5)
	b.Paint(buf) // should not panic
}

// === getCachedBlocks edge cases ===

func TestP67_AssistantText_CacheHitAfterFirstRender(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("# Hello\nWorld")
	b.Measure(component.Constraints{MaxWidth: 80})
	// Second measure should hit cache
	s2 := b.Measure(component.Constraints{MaxWidth: 80})
	if s2.H == 0 {
		t.Error("expected non-zero height")
	}
}

func TestP67_AssistantText_CacheInvalidatedOnContentChange(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("Line 1")
	h1 := b.Measure(component.Constraints{MaxWidth: 80}).H
	// Add enough content to guarantee more lines (headings + paragraphs)
	b.SetContent("# Title\n\nLine 1\n\nLine 2\n\nLine 3\n\nMore text")
	h2 := b.Measure(component.Constraints{MaxWidth: 80}).H
	if h2 <= h1 {
		t.Errorf("expected taller after adding content: h1=%d h2=%d", h1, h2)
	}
}

func TestP67_AssistantText_CacheInvalidatedOnWidthChange(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("A very long line that needs wrapping")
	h1 := b.Measure(component.Constraints{MaxWidth: 80}).H
	h2 := b.Measure(component.Constraints{MaxWidth: 20}).H
	// Narrower width should cause more wrapping → more lines
	if h2 < h1 {
		t.Errorf("expected >= lines for narrower width: h1=%d h2=%d", h1, h2)
	}
}

// === Block.String ===

func TestP67_Block_String(t *testing.T) {
	b := NewBaseBlock("test1", TypeUserMessage)
	s := b.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}

func TestP67_BlockState_String(t *testing.T) {
	states := []BlockState{BlockPending, BlockStreaming, BlockComplete, BlockError}
	for _, s := range states {
		str := s.String()
		if str == "" {
			t.Errorf("expected non-empty string for state %d", s)
		}
	}
	// Unknown state
	unknown := BlockState(99)
	if unknown.String() == "" {
		t.Error("expected non-empty string even for unknown state")
	}
}

// === Block.String (ToolCall) ===

func TestP67_ToolCallBlock_String(t *testing.T) {
	b := NewToolCallBlock("t1", "test_tool", `{"a":1}`)
	s := b.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}

// === SaveContainer/LoadContainer error paths ===

func TestP67_SaveContainer_NilBlocks(t *testing.T) {
	r := NewDefaultRegistry()
	c := NewBlockContainer()
	c.blocks = nil
	data, err := SaveContainer(c, r)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if data == nil {
		t.Error("expected non-nil data")
	}
}

func TestP67_LoadContainer_InvalidJSON(t *testing.T) {
	r := NewDefaultRegistry()
	_, err := LoadContainer([]byte("invalid json"), r)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

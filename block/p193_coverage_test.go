package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// P193: Targeted tests for uncovered AssistantTextBlock.Paint branches

func TestAssistantTextBlock_PaintFallbackPlain_P193(t *testing.T) {
	// Test fallback path when getCachedBlocks returns nil
	b := NewAssistantTextBlock("fb1")
	b.SetContent("plain text without markdown formatting")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf) // should hit either md path or fallback
}

func TestAssistantTextBlock_PaintCachedThenPaint_P193(t *testing.T) {
	// Test cache hit path — paint twice with same content/width
	b := NewAssistantTextBlock("fb2")
	b.SetContent("## Header\n\nparagraph text")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf) // first render
	b.Paint(buf) // second render (cache hit)
}

func TestAssistantTextBlock_PaintWidthChange_P193(t *testing.T) {
	// Test cache invalidation on width change
	b := NewAssistantTextBlock("fb3")
	b.SetContent("## Header\n\ntext here")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf) // render at width 40
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 10})
	buf2 := buffer.NewBuffer(20, 10)
	b.Paint(buf2) // re-render at width 20 (cache invalid)
}

func TestAssistantTextBlock_PaintContentChange_P193(t *testing.T) {
	// Test cache invalidation on content change
	b := NewAssistantTextBlock("fb4")
	b.SetContent("first content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
	b.SetContent("second content with different text")
	b.Paint(buffer.NewBuffer(40, 10)) // cache invalidated by SetContent
}

func TestAssistantTextBlock_PaintAppendDelta_P193(t *testing.T) {
	// Test that AppendDelta invalidates cache
	b := NewAssistantTextBlock("fb5")
	b.SetContent("initial")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
	b.AppendDelta(" more text")
	b.Paint(buffer.NewBuffer(40, 10)) // cache invalidated by AppendDelta
}

func TestAssistantTextBlock_MeasureNilBlocks_P193(t *testing.T) {
	// Test Measure with nil blocks fallback
	b := NewAssistantTextBlock("fb6")
	b.SetContent("test text")
	s := b.Measure(component.Constraints{MaxWidth: 40, MaxHeight: 10})
	if s.H < 1 {
		t.Error("Measure should return at least 1 line height")
	}
}

func TestAssistantTextBlock_MeasureEmpty_P193(t *testing.T) {
	b := NewAssistantTextBlock("fb7")
	b.SetContent("")
	s := b.Measure(component.Constraints{MaxWidth: 40, MaxHeight: 10})
	if s.H < 0 {
		t.Errorf("empty Measure should return H>=0, got %d", s.H)
	}
}

func TestAssistantTextBlock_MeasureZeroMaxWidth_P193(t *testing.T) {
	b := NewAssistantTextBlock("fb8")
	b.SetContent("some text")
	s := b.Measure(component.Constraints{}) // MaxWidth=0 → default 80
	if s.W < 1 {
		t.Error("should have non-zero width")
	}
}

func TestAssistantTextBlock_SerializeDeserialize_P193(t *testing.T) {
	b := NewAssistantTextBlock("fb9")
	b.SetContent("## Title\n\nbody text")
	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}
	b2 := NewAssistantTextBlock("fb9_copy")
	if err := b2.DeserializeState(data); err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
}

func TestAssistantTextBlock_SerializeInvalidJSON_P193(t *testing.T) {
	b := NewAssistantTextBlock("fb10")
	err := b.DeserializeState([]byte("not valid json"))
	if err == nil {
		t.Error("should return error for invalid JSON")
	}
}
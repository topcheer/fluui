package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// P213: AssistantTextBlock.Paint coverage for nil mdBlocks fallback path

func TestAssistantTextBlock_PaintEmptyBounds_P213(t *testing.T) {
	b := NewAssistantTextBlock("p213a")
	b.SetContent("test")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	b.Paint(buffer.NewBuffer(10, 5))
	// Should return early — no panic
}

func TestAssistantTextBlock_PaintEmptyContent_P213(t *testing.T) {
	b := NewAssistantTextBlock("p213b")
	b.SetContent("")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	b.Paint(buffer.NewBuffer(40, 5))
}

func TestAssistantTextBlock_PaintFallbackRender_P213(t *testing.T) {
	// Test the fallback plain-text render path (mdBlocks == nil)
	b := NewAssistantTextBlock("p213c")
	b.SetContent("short text that needs wrapping in a very narrow viewport")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 10})
	b.Paint(buffer.NewBuffer(10, 10))
}

func TestAssistantTextBlock_PaintOverflowHeight_P213(t *testing.T) {
	// Content taller than bounds — should stop at bounds.H
	b := NewAssistantTextBlock("p213d")
	b.SetContent("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 3})
	b.Paint(buffer.NewBuffer(40, 3))
}
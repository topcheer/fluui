package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// P216: block.Paint markdown render path (mdBlocks != nil) + cell overflow

func TestAssistantTextBlock_PaintMarkdownRender_P216(t *testing.T) {
	b := NewAssistantTextBlock("p216a")
	b.SetContent("# Title\n\nparagraph with **bold**\n\n- item1\n- item2")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	b.Paint(buffer.NewBuffer(60, 20))
}

func TestAssistantTextBlock_PaintWithCodeBlock_P216(t *testing.T) {
	b := NewAssistantTextBlock("p216b")
	b.SetContent("```go\nfunc main() {}\n```\n\ntext after code")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	b.Paint(buffer.NewBuffer(60, 20))
}

func TestAssistantTextBlock_PaintWidthClamp_P216(t *testing.T) {
	// Content wider than bounds — should clamp x to bounds.W
	b := NewAssistantTextBlock("p216c")
	b.SetContent("very long line that exceeds the narrow viewport width for clamping")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 10})
	b.Paint(buffer.NewBuffer(10, 10))
}

func TestAssistantTextBlock_PaintOffsetBounds_P216(t *testing.T) {
	// Non-zero bounds offset
	b := NewAssistantTextBlock("p216d")
	b.SetContent("test content")
	b.SetBounds(component.Rect{X: 5, Y: 3, W: 30, H: 5})
	b.Paint(buffer.NewBuffer(40, 10))
}
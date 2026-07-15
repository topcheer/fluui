package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// P197: Coverage for block sub-80% functions

func TestAssistantTextBlock_PaintMarkdownRender_P197(t *testing.T) {
	b := NewAssistantTextBlock("md1")
	b.SetContent("# Header\n\nSome **bold** and *italic* text.\n\n- item 1\n- item 2\n\n> Quote text\n\n| A | B |\n|---|---|\n| 1 | 2 |\n\n[link](https://example.com)\n")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 40})
	b.Paint(buffer.NewBuffer(60, 40))
}

func TestAssistantTextBlock_PaintNilMdBlocks_P197(t *testing.T) {
	b := NewAssistantTextBlock("md2")
	b.SetContent("test")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 1, H: 1})
	b.Paint(buffer.NewBuffer(1, 1))
}

func TestAssistantTextBlock_PaintStreamingState_P197(t *testing.T) {
	b := NewAssistantTextBlock("md3")
	b.SetContent("## Streaming Header\n\ntext")
	b.SetState(BlockStreaming)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestBaseBlock_Paint_P197(t *testing.T) {
	bb := BaseBlock{}
	bb.Paint(buffer.NewBuffer(10, 5))
}

func TestBaseBlock_String_P197(t *testing.T) {
	bb := NewBaseBlock("test-id", TypeThinking)
	s := bb.String()
	if s == "" {
		t.Error("String should not be empty")
	}
}
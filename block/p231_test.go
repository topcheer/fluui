package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestAssistantTextBlock_PaintEmpty_P231(t *testing.T) {
	b := NewAssistantTextBlock("test1")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestAssistantTextBlock_PaintZeroBounds_P231(t *testing.T) {
	b := NewAssistantTextBlock("test2")
	b.SetContent("hello world")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestAssistantTextBlock_PaintPlainText_P231(t *testing.T) {
	b := NewAssistantTextBlock("test3")
	b.SetContent("line1\nline2\nline3")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
}

func TestAssistantTextBlock_PaintTruncated_P231(t *testing.T) {
	b := NewAssistantTextBlock("test4")
	b.SetContent("a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk\nl")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 3})
	buf := buffer.NewBuffer(60, 3)
	b.Paint(buf)
}

func TestAssistantTextBlock_PaintMarkdown_P231(t *testing.T) {
	b := NewAssistantTextBlock("test5")
	b.SetContent("# Title\n\n**bold** and *italic*\n\n- item1\n- item2")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	b.Paint(buf)
}

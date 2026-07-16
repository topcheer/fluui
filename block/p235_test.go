package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// P235: trigger assistant_text Paint fallback path (mdBlocks==nil)
// This happens when the markdown renderer returns zero blocks.

func TestAssistantText_PaintFallbackWhitespace_P235(t *testing.T) {
	b := NewAssistantTextBlock("test1")
	// Whitespace-only content — goldmark may parse to zero blocks
	b.SetContent("   \n   \n   ")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
}

func TestAssistantText_PaintFallbackNewlines_P235(t *testing.T) {
	b := NewAssistantTextBlock("test2")
	b.SetContent("\n\n\n\n\n")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
}

func TestAssistantText_PaintFallbackHorizontalTab_P235(t *testing.T) {
	b := NewAssistantTextBlock("test3")
	// Horizontal tab — goldmark treats as whitespace
	b.SetContent("\t\n\t\n\t")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
}

func TestAssistantText_MeasureFallback_P235(t *testing.T) {
	b := NewAssistantTextBlock("test4")
	b.SetContent("\n\n\n")
	s := b.Measure(component.Unbounded())
	if s.H < 1 {
		t.Errorf("Measure should return at least 1 line, got %d", s.H)
	}
}

func TestAssistantText_PaintTruncatedCellLine_P235(t *testing.T) {
	// Cover the cell.Width==0 branch in Paint (line 292-294)
	b := NewAssistantTextBlock("test5")
	b.SetContent("```go\npackage main\n```")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	b.Paint(buf)
}

func TestAssistantText_PaintNarrowWidth_P235(t *testing.T) {
	// Cover the x >= bounds.X+bounds.W break (line 289)
	b := NewAssistantTextBlock("test6")
	b.SetContent("This is a very long line of text that should exceed narrow width bounds")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 10})
	buf := buffer.NewBuffer(5, 10)
	b.Paint(buf)
}

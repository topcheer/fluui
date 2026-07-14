package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// P191: Coverage tests for sub-80% block functions

func TestBaseBlock_Paint_P191(t *testing.T) {
	b := &BaseBlock{}
	b.Paint(buffer.NewBuffer(10, 5)) // no-op, should not crash
}

func TestAssistantTextBlock_PaintMarkdown_P191(t *testing.T) {
	b := NewAssistantTextBlock("test1")
	b.SetContent("# Header\n\nSome **bold** text.\n\n- item 1\n- item 2\n\n```\ncode block\n```\n\n| Col1 | Col2 |\n|------|------|\n| a    | b    |")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 20})
	buf := buffer.NewBuffer(40, 20)
	b.Paint(buf) // exercises markdown render path
}

func TestAssistantTextBlock_PaintBoldItalic_P191(t *testing.T) {
	b := NewAssistantTextBlock("test2")
	b.SetContent("**bold** and *italic* and `code`")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestAssistantTextBlock_PaintBlockquote_P191(t *testing.T) {
	b := NewAssistantTextBlock("test3")
	b.SetContent("> A blockquote\n> with two lines")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestAssistantTextBlock_PaintLinks_P191(t *testing.T) {
	b := NewAssistantTextBlock("test4")
	b.SetContent("[link text](https://example.com)")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestAssistantTextBlock_PaintUnicode_P191(t *testing.T) {
	b := NewAssistantTextBlock("test5")
	b.SetContent("日本語テスト — café")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestAssistantTextBlock_PaintNarrowWidth_P191(t *testing.T) {
	b := NewAssistantTextBlock("test6")
	b.SetContent("some long text that needs wrapping in narrow width")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 20})
	b.Paint(buffer.NewBuffer(10, 20))
}

func TestAssistantTextBlock_PaintMultiParagraph_P191(t *testing.T) {
	b := NewAssistantTextBlock("test7")
	b.SetContent("Para 1.\n\nPara 2.\n\nPara 3.\n\nPara 4.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 20})
	b.Paint(buffer.NewBuffer(40, 20))
}

func TestAssistantTextBlock_PaintNestedLists_P191(t *testing.T) {
	b := NewAssistantTextBlock("test8")
	b.SetContent("- top level\n  - nested\n    - deep\n- back to top")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 20})
	b.Paint(buffer.NewBuffer(40, 20))
}

func TestAssistantTextBlock_PaintHorizontalRule_P191(t *testing.T) {
	b := NewAssistantTextBlock("test9")
	b.SetContent("above\n\n---\n\nbelow")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10))
}

func TestAssistantTextBlock_PaintHeightTruncation_P191(t *testing.T) {
	b := NewAssistantTextBlock("test10")
	b.SetContent("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 3})
	b.Paint(buffer.NewBuffer(40, 3)) // only 3 rows visible
}

func TestAssistantTextBlock_PaintZeroBounds_P191(t *testing.T) {
	b := NewAssistantTextBlock("test11")
	b.SetContent("text")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	b.Paint(buffer.NewBuffer(0, 0)) // zero bounds, should return early
}

func TestAssistantTextBlock_PaintEmptyContent_P191(t *testing.T) {
	b := NewAssistantTextBlock("test12")
	b.SetContent("")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	b.Paint(buffer.NewBuffer(40, 10)) // empty content, should return early
}

func TestAssistantTextBlock_PaintNonZeroOffset_P191(t *testing.T) {
	b := NewAssistantTextBlock("test13")
	b.SetContent("offset test content")
	b.SetBounds(component.Rect{X: 5, Y: 3, W: 30, H: 10})
	b.Paint(buffer.NewBuffer(40, 20)) // non-zero X/Y offset
}
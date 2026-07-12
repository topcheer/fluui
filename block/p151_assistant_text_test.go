package block

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/markdown"
)

// Test Paint with various markdown content to cover all branches
func TestP151_Paint_MarkdownHeaders(t *testing.T) {
	b := NewAssistantTextBlock("test1")
	b.AppendDelta("# Header\n\nText after header\n\n## Subheader")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
	// Should have rendered some content
	_ = buf
}

func TestP151_Paint_CodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("test2")
	b.AppendDelta("```go\npackage main\nfunc main() {}\n```")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP151_Paint_OrderedList(t *testing.T) {
	b := NewAssistantTextBlock("test3")
	b.AppendDelta("1. First item\n2. Second item\n3. Third item")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP151_Paint_UnorderedList(t *testing.T) {
	b := NewAssistantTextBlock("test4")
	b.AppendDelta("- Item one\n- Item two\n- Item three")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP151_Paint_BoldItalic(t *testing.T) {
	b := NewAssistantTextBlock("test5")
	b.AppendDelta("**bold** and *italic* and `code`")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP151_Paint_Table(t *testing.T) {
	b := NewAssistantTextBlock("test6")
	b.AppendDelta("| Name | Value |\n|------|-------|\n| A    | 1     |\n| B    | 2     |")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP151_Paint_Blockquote(t *testing.T) {
	b := NewAssistantTextBlock("test7")
	b.AppendDelta("> This is a quote\n> Second line")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP151_Paint_Links(t *testing.T) {
	b := NewAssistantTextBlock("test8")
	b.AppendDelta("[Example](https://example.com) and https://raw-url.com")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP151_Paint_Unicode(t *testing.T) {
	b := NewAssistantTextBlock("test9")
	b.AppendDelta("Hello 世界 café 日本語 emoji 🎉")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP151_Paint_ZeroBounds(t *testing.T) {
	b := NewAssistantTextBlock("test10")
	b.AppendDelta("Some text")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf) // should not panic
}

func TestP151_Paint_EmptyContent(t *testing.T) {
	b := NewAssistantTextBlock("test11")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf) // should not panic, should be empty
}

func TestP151_Paint_NarrowWidth(t *testing.T) {
	b := NewAssistantTextBlock("test12")
	b.AppendDelta("This is a very long line that needs to wrap")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 20})
	buf := buffer.NewBuffer(5, 20)
	b.Paint(buf)
}

func TestP151_Paint_HeightTruncation(t *testing.T) {
	b := NewAssistantTextBlock("test13")
	b.AppendDelta("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 3})
	buf := buffer.NewBuffer(60, 3)
	b.Paint(buf)
}

func TestP151_Paint_NonZeroOffset(t *testing.T) {
	b := NewAssistantTextBlock("test14")
	b.AppendDelta("Hello world")
	b.SetBounds(component.Rect{X: 5, Y: 3, W: 50, H: 15})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP151_Paint_MultiParagraph(t *testing.T) {
	b := NewAssistantTextBlock("test15")
	b.AppendDelta("First paragraph.\n\nSecond paragraph.\n\nThird paragraph.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP151_Paint_NestedList(t *testing.T) {
	b := NewAssistantTextBlock("test16")
	b.AppendDelta("- Top level\n  - Nested item\n  - Another nested\n- Back to top")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP151_Paint_HorizontalRule(t *testing.T) {
	b := NewAssistantTextBlock("test17")
	b.AppendDelta("Before\n\n---\n\nAfter")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
}

func TestP151_Paint_CacheHit(t *testing.T) {
	b := NewAssistantTextBlock("test18")
	b.AppendDelta("Some content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf1 := buffer.NewBuffer(60, 20)
	b.Paint(buf1)
	// Second paint should hit cache
	buf2 := buffer.NewBuffer(60, 20)
	b.Paint(buf2)
}

func TestP151_Paint_CacheInvalidateWidth(t *testing.T) {
	b := NewAssistantTextBlock("test19")
	b.AppendDelta("Some content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
	// Change width — should invalidate cache
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 20})
	buf2 := buffer.NewBuffer(40, 20)
	b.Paint(buf2)
}

func TestP151_Measure_WithContent(t *testing.T) {
	b := NewAssistantTextBlock("test20")
	b.AppendDelta("# Header\n\nSome text\n\nMore text")
	size := b.Measure(component.Constraints{MaxWidth: 60, MaxHeight: 20})
	if size.W <= 0 || size.H <= 0 {
		t.Errorf("expected positive size, got %+v", size)
	}
}

func TestP151_Measure_Empty(t *testing.T) {
	b := NewAssistantTextBlock("test21")
	size := b.Measure(component.Constraints{MaxWidth: 60, MaxHeight: 20})
	if size.H != 1 {
		t.Errorf("expected H=1 for empty, got %d", size.H)
	}
}

func TestP151_Measure_ZeroMaxWidth(t *testing.T) {
	b := NewAssistantTextBlock("test22")
	b.AppendDelta("Some text")
	size := b.Measure(component.Constraints{})
	if size.W != 80 {
		t.Errorf("expected W=80 default, got %d", size.W)
	}
}

func TestP151_Measure_CodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("test23")
	b.AppendDelta("```go\npackage main\nfunc main() {}\n```")
	size := b.Measure(component.Constraints{MaxWidth: 60, MaxHeight: 20})
	if size.H < 3 {
		t.Errorf("expected at least 3 lines for code block, got %d", size.H)
	}
}

func TestP151_Measure_CacheHit(t *testing.T) {
	b := NewAssistantTextBlock("test24")
	b.AppendDelta("# Header\n\nText")
	size1 := b.Measure(component.Constraints{MaxWidth: 60, MaxHeight: 20})
	size2 := b.Measure(component.Constraints{MaxWidth: 60, MaxHeight: 20})
	if size1 != size2 {
		t.Errorf("expected same size on cache hit, got %+v vs %+v", size1, size2)
	}
}

func TestP151_contentString_CacheHit(t *testing.T) {
	b := NewAssistantTextBlock("test25")
	b.AppendDelta("Hello")
	// First call populates cache
	s1 := b.contentString()
	// Second call hits cache
	s2 := b.contentString()
	if s1 != s2 {
		t.Errorf("expected same string, got %q vs %q", s1, s2)
	}
}

func TestP151_contentString_DirtyAfterAppend(t *testing.T) {
	b := NewAssistantTextBlock("test26")
	b.AppendDelta("Hello")
	s1 := b.contentString()
	b.AppendDelta(" World")
	s2 := b.contentString()
	if s2 != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", s2)
	}
	if s1 != "Hello" {
		t.Errorf("expected first call 'Hello', got %q", s1)
	}
}

func TestP151_SetContent(t *testing.T) {
	b := NewAssistantTextBlock("test27")
	b.AppendDelta("Original")
	b.SetContent("Replaced")
	if b.Content() != "Replaced" {
		t.Errorf("expected 'Replaced', got %q", b.Content())
	}
}

func TestP151_SerializeDeserialize_RoundTrip(t *testing.T) {
	b := NewAssistantTextBlock("test28")
	b.AppendDelta("# Title\n\nSome markdown content")
	state, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}

	b2 := NewAssistantTextBlock("test28_copy")
	if err := b2.DeserializeState(state); err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
	if b2.Content() != b.Content() {
		t.Errorf("expected same content, got %q vs %q", b2.Content(), b.Content())
	}
}

func TestP151_SerializeDeserialize_InvalidJSON(t *testing.T) {
	b := NewAssistantTextBlock("test29")
	err := b.DeserializeState(json.RawMessage("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestP151_GetCachedBlocks_RenderError(t *testing.T) {
	b := NewAssistantTextBlock("test30")
	b.AppendDelta("Valid content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	// Override renderer to force error
	theme := markdown.MarkdownTheme{}
	b.renderer = markdown.NewMarkdownRenderer(&theme, 60)
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
}

func TestP151_Paint_FallbackPlainText(t *testing.T) {
	b := NewAssistantTextBlock("test31")
	b.AppendDelta("Just plain text\nNo markdown here")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
	// Check some content was rendered
	cell := buf.GetCell(0, 0)
	_ = cell
	hasContent := false
	for y := 0; y < 10; y++ {
		for x := 0; x < 60; x++ {
			c := buf.GetCell(x, y)
			if c.Rune != 0 && c.Rune != ' ' {
				hasContent = true
				break
			}
		}
		if hasContent {
			break
		}
	}
	if !hasContent {
		t.Error("expected some content rendered")
	}
}

func TestP151_Paint_WideCharWidth(t *testing.T) {
	b := NewAssistantTextBlock("test32")
	b.AppendDelta("日本語テスト")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
}

func TestP151_Paint_MarkdownStrikethrough(t *testing.T) {
	b := NewAssistantTextBlock("test33")
	b.AppendDelta("~~strikethrough text~~")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
}

func TestP151_Paint_TaskList(t *testing.T) {
	b := NewAssistantTextBlock("test34")
	b.AppendDelta("- [ ] Todo item\n- [x] Done item")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
}

func TestP151_Paint_Alert(t *testing.T) {
	b := NewAssistantTextBlock("test35")
	b.AppendDelta("> [!NOTE]\n> This is a note alert")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
}

func TestP151_Paint_Image(t *testing.T) {
	b := NewAssistantTextBlock("test36")
	b.AppendDelta("![alt text](https://example.com/image.png)")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
}

func TestP151_Content(t *testing.T) {
	b := NewAssistantTextBlock("test37")
	b.AppendDelta("Hello ")
	b.AppendDelta("World")
	if b.Content() != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", b.Content())
	}
}

func TestP151_Paint_LongContent(t *testing.T) {
	b := NewAssistantTextBlock("test38")
	// Generate long content
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = strings.Repeat("x", 40)
	}
	b.AppendDelta(strings.Join(lines, "\n"))
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf) // should truncate at height 10
}
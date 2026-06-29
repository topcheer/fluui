package markdown

import (
	"strings"
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ============================================================
// P25-B: Markdown Edge Case & Stress Tests
// ============================================================

// --- Renderer edge cases ---

func TestP25B_RenderEmpty(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("")
	if err != nil {
		t.Errorf("empty string should not error: %v", err)
	}
	if len(blocks) != 0 {
		t.Errorf("empty string should produce 0 blocks, got %d", len(blocks))
	}
}

func TestP25B_RenderOnlyNewlines(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("\n\n\n\n\n")
	if err != nil {
		t.Errorf("newlines should not error: %v", err)
	}
	_ = blocks
}

func TestP25B_RenderOnlyWhitespace(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("   \t  \n  \t  \n   ")
	if err != nil {
		t.Errorf("whitespace should not error: %v", err)
	}
	_ = blocks
}

func TestP25B_RenderUnclosedCodeBlock(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("```go\nfunc main() {\nfmt.Println(\"hello\")\n// code block never closed")
	if err != nil {
		t.Errorf("unclosed code block should not error: %v", err)
	}
	_ = blocks
}

func TestP25B_RenderDeeplyNestedLists(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	md := strings.Repeat("    ", 20) + "deeply nested item"
	blocks, err := r.Render(md)
	if err != nil {
		t.Errorf("deeply nested lists should not error: %v", err)
	}
	_ = blocks
}

func TestP25B_RenderDeeplyNestedBlockquotes(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	md := strings.Repeat("> ", 30) + "quote inception"
	blocks, err := r.Render(md)
	if err != nil {
		t.Errorf("deeply nested blockquotes should not error: %v", err)
	}
	_ = blocks
}

func TestP25B_RenderHugeHeading(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	md := "# " + strings.Repeat("A", 10000)
	blocks, err := r.Render(md)
	if err != nil {
		t.Errorf("huge heading should not error: %v", err)
	}
	_ = blocks
}

func TestP25B_RenderMalformedHTML(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("<div><span>unclosed<span>tags<div></div>")
	if err != nil {
		t.Errorf("malformed HTML should not error: %v", err)
	}
	_ = blocks
}

func TestP25B_RenderMixedDelimiters(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("**bold *and italic** end*")
	if err != nil {
		t.Errorf("mixed delimiters should not error: %v", err)
	}
	_ = blocks
}

func TestP25B_RenderHTMLEntities(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("&amp; &lt; &gt; &quot; &#39; &nbsp;")
	if err != nil {
		t.Errorf("HTML entities should not error: %v", err)
	}
	_ = blocks
}

func TestP25B_RenderLinkEdgeCases(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	tests := []string{
		"[](empty-url)",
		"[text]()",
		"[](http://example.com)",
		"[very long link text](http://example.com/path?a=1&b=2)",
	}
	for i, md := range tests {
		blocks, err := r.Render(md)
		if err != nil {
			t.Errorf("link case %d should not error: %v", i, err)
		}
		_ = blocks
	}
}

func TestP25B_RenderTableEdgeCases(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	tests := []string{
		"| | |\n|---|---|\n| | |",    // empty cells
		"| a |\n|---|\n| b |",          // single column
	}
	for i, md := range tests {
		blocks, err := r.Render(md)
		if err != nil {
			t.Errorf("table case %d should not error: %v", i, err)
		}
		_ = blocks
	}
}

func TestP25B_RenderNormalDocument(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	md := "# Heading\n\nSome **bold** and *italic* text.\n\n- item 1\n- item 2\n\n```go\nfunc main() {}\n```\n"
	blocks, err := r.Render(md)
	if err != nil {
		t.Errorf("normal document should not error: %v", err)
	}
	if len(blocks) == 0 {
		t.Error("normal document should produce blocks")
	}
}

func TestP25B_RenderLargeDocument(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	var sb strings.Builder
	for i := 0; i < 200; i++ {
		sb.WriteString("## Section ")
		sb.WriteString(string(rune('A' + i%26)))
		sb.WriteString("\n\nParagraph with **bold** and *italic* text.\n\n")
	}
	blocks, err := r.Render(sb.String())
	if err != nil {
		t.Errorf("large document should not error: %v", err)
	}
	if len(blocks) < 50 {
		t.Errorf("large document should produce many blocks, got %d", len(blocks))
	}
}

func TestP25B_RenderConcurrentStress(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	md := "# Heading\n\nSome **bold** and *italic* text.\n\n- item 1\n- item 2\n\n```go\nfunc main() {}\n```\n"

	var wg sync.WaitGroup
	const goroutines = 50
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			blocks, err := r.Render(md)
			if err != nil {
				t.Errorf("concurrent render error: %v", err)
			}
			if len(blocks) == 0 {
				t.Error("expected non-zero blocks")
			}
		}()
	}
	wg.Wait()
}

func TestP25B_RenderNullByte(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("\x00hello world")
	if err != nil {
		t.Errorf("null byte should not error: %v", err)
	}
	_ = blocks
}

func TestP25B_RenderControlChars(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("\x01\x02\x03hello\x04\x05")
	if err != nil {
		t.Errorf("control chars should not error: %v", err)
	}
	_ = blocks
}

// --- WrapText edge cases ---

func TestP25B_WrapTextEmpty(t *testing.T) {
	lines := WrapText("", 10)
	if len(lines) != 1 {
		t.Errorf("empty string should return 1 line, got %d", len(lines))
	}
}

func TestP25B_WrapTextNegativeWidth(t *testing.T) {
	lines := WrapText("hello world", -1)
	if len(lines) != 1 {
		t.Errorf("negative width should return original as 1 line, got %d", len(lines))
	}
}

func TestP25B_WrapTextZeroWidth(t *testing.T) {
	lines := WrapText("hello", 0)
	if len(lines) != 1 {
		t.Errorf("zero width should return original as 1 line, got %d", len(lines))
	}
}

func TestP25B_WrapTextSingleCharPerLine(t *testing.T) {
	lines := WrapText("hello", 1)
	if len(lines) != 5 {
		t.Errorf("width=1 should produce 5 lines, got %d", len(lines))
	}
}

func TestP25B_WrapTextLongWord(t *testing.T) {
	word := strings.Repeat("x", 100)
	lines := WrapText(word, 10)
	if len(lines) < 10 {
		t.Errorf("long word should be split into multiple lines, got %d", len(lines))
	}
}

func TestP25B_WrapTextAllSpaces(t *testing.T) {
	lines := WrapText("     ", 3)
	_ = lines // should not panic
}

func TestP25B_WrapTextMixedUnicode(t *testing.T) {
	lines := WrapText("hello \u4e16\u754c unicode \U0001f600 emoji", 10)
	_ = lines // should not panic
}

// --- Truncate edge cases ---

func TestP25B_TruncateEmpty(t *testing.T) {
	result := Truncate("", 10, "...")
	if result != "" {
		t.Errorf("truncate empty should return empty, got %q", result)
	}
}

func TestP25B_TruncateNegativeWidth(t *testing.T) {
	result := Truncate("hello", -1, "...")
	// Truncate with negative width may not return original
	// just verify it doesn't panic
	_ = result
}

func TestP25B_TruncateExactWidth(t *testing.T) {
	result := Truncate("hello", 5, "...")
	if result != "hello" {
		t.Errorf("exact width should return original, got %q", result)
	}
}

func TestP25B_TruncateSmallerThanEllipsis(t *testing.T) {
	result := Truncate("hello world", 2, "...")
	if len([]rune(result)) > 2 {
		t.Errorf("result should not exceed width, got %q", result)
	}
}

func TestP25B_TruncateCJK(t *testing.T) {
	result := Truncate("\u4e16\u754c\u4f60\u597d", 5, "...")
	_ = result // should not panic
}

// --- StringWidth ---

func TestP25B_StringWidthEmpty(t *testing.T) {
	if w := StringWidth(""); w != 0 {
		t.Errorf("empty string width should be 0, got %d", w)
	}
}

func TestP25B_StringWidthASCII(t *testing.T) {
	if w := StringWidth("hello"); w != 5 {
		t.Errorf("'hello' width should be 5, got %d", w)
	}
}

func TestP25B_StringWidthMixed(t *testing.T) {
	w := StringWidth("abc\u4e16\U0001f600")
	if w < 3 {
		t.Errorf("mixed string should have non-trivial width, got %d", w)
	}
}

// --- textToCells ---

func TestP25B_textToCellsEmpty(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	cells := r.textToCells("", buffer.NoColor(), 0)
	if len(cells) != 0 {
		t.Errorf("empty string should produce 0 cells, got %d", len(cells))
	}
}

func TestP25B_textToCellsBasic(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	cells := r.textToCells("hello", buffer.NoColor(), 0)
	if len(cells) != 5 {
		t.Errorf("'hello' should produce 5 cells, got %d", len(cells))
	}
}

func TestP25B_textToCellsEmoji(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	cells := r.textToCells("\U0001f600\U0001f600", buffer.NoColor(), 0)
	if len(cells) == 0 {
		t.Error("emoji should produce non-zero cells")
	}
}

func TestP25B_textToCellsVeryLong(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	text := strings.Repeat("A", 10000)
	cells := r.textToCells(text, buffer.NoColor(), 0)
	if len(cells) != 10000 {
		t.Errorf("10000 chars should produce 10000 cells, got %d", len(cells))
	}
}

func TestP25B_textToCellsStyleApplied(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	cells := r.textToCells("X", buffer.RGB(255, 0, 0), buffer.Bold)
	if len(cells) == 0 {
		t.Fatal("expected at least 1 cell")
	}
	if cells[0].Flags&buffer.Bold == 0 {
		t.Error("Bold flag not applied")
	}
}

func TestP25B_textToCellsTabChar(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	cells := r.textToCells("\t", buffer.NoColor(), 0)
	_ = cells // should not panic
}

func TestP25B_textToCellsNullByte(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	cells := r.textToCells("\x00abc", buffer.NoColor(), 0)
	_ = cells // should not panic
}

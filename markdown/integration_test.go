package markdown

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// hasColoredCell returns true if any cell in the lines has a non-default Fg color.
func hasColoredCell(lines [][]buffer.Cell) bool {
	for _, line := range lines {
		for _, c := range line {
			if !c.Fg.IsDefault() {
				return true
			}
		}
	}
	return false
}

// findCellsWithColor returns all cells whose Fg matches the given color.
func findCellsWithColor(lines [][]buffer.Cell, fg buffer.Color) []buffer.Cell {
	var result []buffer.Cell
	for _, line := range lines {
		for _, c := range line {
			if c.Fg.Equal(fg) {
				result = append(result, c)
			}
		}
	}
	return result
}

// === Integration Test 1: Full markdown document ===

func TestIntegrationRenderFullDocument(t *testing.T) {
	md := `# Project Title

This is a paragraph with some text.

- First item
- Second item
- Third item

` + "```go" + `
package main
func main() {}
` + "```" + `

> This is a quote.
`

	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render(md)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Should produce at least 5 blocks: heading, paragraph, list, code, quote
	if len(blocks) < 5 {
		t.Fatalf("expected at least 5 blocks, got %d", len(blocks))
	}

	// Verify block types in order
	expectedTypes := []BlockType{BlockHeading, BlockParagraph, BlockList, BlockCodeBlock, BlockQuote}
	for i, expected := range expectedTypes {
		if blocks[i].Type != expected {
			t.Errorf("block %d: expected type %v, got %v", i, expected, blocks[i].Type)
		}
	}

	// Heading block
	if blocks[0].Level != 1 {
		t.Errorf("heading: expected level 1, got %d", blocks[0].Level)
	}
	headingText := cellText(blocks[0].Cells[0])
	if !strings.Contains(headingText, "Project Title") {
		t.Errorf("heading text: got %q", headingText)
	}
	// H1 should have Bold + pink color
	theme := DefaultTheme()
	for _, c := range blocks[0].Cells[0] {
		if c.Rune == 'P' {
			if !c.Fg.Equal(theme.H1) {
				t.Errorf("heading: expected H1 color, got %v", c.Fg)
			}
			if c.Flags&buffer.Bold == 0 {
				t.Error("heading: expected Bold flag")
			}
			break
		}
	}

	// Paragraph block
	paraText := cellText(blocks[1].Cells[0])
	if !strings.Contains(paraText, "paragraph") {
		t.Errorf("paragraph text: got %q", paraText)
	}

	// List block — at least 3 items
	if len(blocks[2].Cells) < 3 {
		t.Errorf("list: expected at least 3 lines, got %d", len(blocks[2].Cells))
	}
	// Check bullets
	for i := 0; i < 3; i++ {
		line := cellText(blocks[2].Cells[i])
		if !strings.Contains(line, "item") {
			t.Errorf("list item %d: expected 'item' in %q", i, line)
		}
	}

	// Code block
	codeText := cellText(blocks[3].Cells[0])
	if !strings.Contains(codeText, "package main") {
		t.Errorf("code: expected 'package main' in %q", codeText)
	}

	// Quote block — should have │ prefix
	if len(blocks[4].Cells) == 0 {
		t.Fatal("quote: expected at least 1 line")
	}
	if blocks[4].Cells[0][0].Rune != '\u2502' {
		t.Errorf("quote: expected │ prefix, got %q", string(blocks[4].Cells[0][0].Rune))
	}
	quoteText := cellText(blocks[4].Cells[0])
	if !strings.Contains(quoteText, "quote") {
		t.Errorf("quote text: got %q", quoteText)
	}
}

// === Integration Test 2: Render with Highlighter ===

func TestIntegrationRenderWithHighlighter(t *testing.T) {
	md := "```go\n" +
		"package main\n\n" +
		"import \"fmt\"\n\n" +
		"func main() {\n" +
		"    fmt.Println(\"hello\")\n" +
		"}\n" +
		"```\n"

	// Without highlighter
	rPlain := NewMarkdownRenderer(DefaultTheme(), 80)
	blocksPlain, _ := rPlain.Render(md)
	if len(blocksPlain) != 1 {
		t.Fatalf("plain: expected 1 block, got %d", len(blocksPlain))
	}
	if blocksPlain[0].Type != BlockCodeBlock {
		t.Errorf("plain: expected BlockCodeBlock, got %v", blocksPlain[0].Type)
	}
	// Without highlighter, all cells should have uniform CodeFg color
	theme := DefaultTheme()
	for _, line := range blocksPlain[0].Cells {
		for _, c := range line {
			if !c.Fg.Equal(theme.CodeFg) {
				t.Errorf("plain: expected CodeFg color, got %v", c.Fg)
			}
		}
	}

	// With highlighter
	rHL := NewMarkdownRenderer(DefaultTheme(), 80)
	rHL.SetHighlighter(NewHighlighter())
	blocksHL, _ := rHL.Render(md)
	if len(blocksHL) != 1 {
		t.Fatalf("highlighted: expected 1 block, got %d", len(blocksHL))
	}
	if blocksHL[0].Type != BlockCodeBlock {
		t.Errorf("highlighted: expected BlockCodeBlock, got %v", blocksHL[0].Type)
	}

	// With highlighter, at least some cells should have non-default colors
	// (keywords, strings, etc. get distinct colors from chroma)
	if !hasColoredCell(blocksHL[0].Cells) {
		t.Error("highlighted: expected at least some cells with non-default colors")
	}

	// The content should still contain the original code text
	allText := ""
	for _, line := range blocksHL[0].Cells {
		allText += cellText(line) + "\n"
	}
	if !strings.Contains(allText, "package main") {
		t.Errorf("highlighted: expected 'package main' in rendered text")
	}
	if !strings.Contains(allText, "func main") {
		t.Errorf("highlighted: expected 'func main' in rendered text")
	}
}

// === Integration Test 3: CJK document ===

func TestIntegrationRenderCJKDocument(t *testing.T) {
	md := "# 中文标题\n\n" +
		"这是一段中文段落，包含一些测试文字。\n\n" +
		"- 第一项\n" +
		"- 第二项\n" +
		"- 第三项\n\n" +
		"```python\n" +
		"print(\"你好世界\")\n" +
		"```\n\n" +
		"> 这是一段引用文字。\n"

	r := NewMarkdownRenderer(DefaultTheme(), 40)
	blocks, err := r.Render(md)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Should produce heading + paragraph + list + code + quote
	if len(blocks) < 5 {
		t.Fatalf("expected at least 5 blocks, got %d", len(blocks))
	}

	// Heading
	if blocks[0].Type != BlockHeading {
		t.Errorf("block 0: expected heading, got %v", blocks[0].Type)
	}
	headingText := cellText(blocks[0].Cells[0])
	if !strings.Contains(headingText, "中文标题") {
		t.Errorf("heading: expected '中文标题' in %q", headingText)
	}

	// Paragraph
	if blocks[1].Type != BlockParagraph {
		t.Errorf("block 1: expected paragraph, got %v", blocks[1].Type)
	}
	paraText := cellText(blocks[1].Cells[0])
	if !strings.Contains(paraText, "中文段落") {
		t.Errorf("paragraph: expected '中文段落' in %q", paraText)
	}

	// List
	if blocks[2].Type != BlockList {
		t.Errorf("block 2: expected list, got %v", blocks[2].Type)
	}
	if len(blocks[2].Cells) < 3 {
		t.Errorf("list: expected at least 3 items, got %d", len(blocks[2].Cells))
	}
	firstItem := cellText(blocks[2].Cells[0])
	if !strings.Contains(firstItem, "第一项") {
		t.Errorf("list item 0: expected '第一项' in %q", firstItem)
	}

	// Code
	if blocks[3].Type != BlockCodeBlock {
		t.Errorf("block 3: expected code block, got %v", blocks[3].Type)
	}
	codeText := cellText(blocks[3].Cells[0])
	if !strings.Contains(codeText, "print") {
		t.Errorf("code: expected 'print' in %q", codeText)
	}

	// Quote
	if blocks[4].Type != BlockQuote {
		t.Errorf("block 4: expected quote, got %v", blocks[4].Type)
	}
	quoteText := cellText(blocks[4].Cells[0])
	if !strings.Contains(quoteText, "引用") {
		t.Errorf("quote: expected '引用' in %q", quoteText)
	}

	// Verify CJK width: each CJK character should have Width=2
	for _, line := range blocks[0].Cells {
		for _, c := range line {
			if c.Rune >= 0x4E00 && c.Rune <= 0x9FFF { // CJK Unified Ideographs
				if c.Width != 2 {
					t.Errorf("CJK rune %q: expected width 2, got %d", string(c.Rune), c.Width)
				}
			}
		}
	}
}

// === Integration Test 4: Wrapping ===

func TestIntegrationRenderWrapping(t *testing.T) {
	// Long paragraph that should wrap at width 30
	para := "Lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod " +
		"tempor incididunt ut labore et dolore magna aliqua ut enim ad minim veniam."

	r := NewMarkdownRenderer(DefaultTheme(), 30)
	blocks, err := r.Render(para + "\n")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Type != BlockParagraph {
		t.Errorf("expected paragraph, got %v", blocks[0].Type)
	}

	// Should have wrapped into multiple lines
	if len(blocks[0].Cells) < 3 {
		t.Errorf("expected at least 3 wrapped lines, got %d", len(blocks[0].Cells))
	}

	// Each line should not exceed the width (30 columns), accounting for
	// the fact that some lines may be slightly shorter due to word boundaries
	for i, line := range blocks[0].Cells {
		w := cellLineWidth(line)
		if w > 30 {
			t.Errorf("line %d: width %d exceeds 30 (text: %q)", i, w, cellText(line))
		}
	}

	// Verify content integrity: concatenating all lines should contain the original words
	allText := ""
	for _, line := range blocks[0].Cells {
		allText += cellText(line) + " "
	}
	for _, word := range []string{"Lorem", "ipsum", "dolor", "consectetur", "eiusmod"} {
		if !strings.Contains(allText, word) {
			t.Errorf("expected word %q in wrapped text", word)
		}
	}
}

// === Integration Test 5: Theme variations ===

func TestIntegrationRenderTheme(t *testing.T) {
	// Create a dark theme
	darkTheme := &MarkdownTheme{
		H1:         buffer.RGB(0xFF, 0x79, 0xC6), // pink (dracula)
		H2:         buffer.RGB(0x8B, 0xE9, 0xFD), // cyan
		CodeFg:     buffer.RGB(0xFF, 0x79, 0xC6),
		LinkFg:     buffer.RGB(0x8B, 0xE9, 0xFD),
		QuoteBar:   buffer.RGB(0x62, 0x72, 0xA4),
		Body:       buffer.NoColor(),
	}

	// Create a light theme with different colors
	lightTheme := &MarkdownTheme{
		H1:         buffer.RGB(0xC0, 0x39, 0x2B), // dark red
		H2:         buffer.RGB(0x21, 0x96, 0xF3), // blue
		CodeFg:     buffer.RGB(0xC0, 0x39, 0x2B),
		LinkFg:     buffer.RGB(0x21, 0x96, 0xF3),
		QuoteBar:   buffer.RGB(0xBD, 0xBD, 0xBD),
		Body:       buffer.NoColor(),
	}

	// Verify themes are actually different
	if darkTheme.H1.Equal(lightTheme.H1) {
		t.Fatal("dark and light H1 colors should differ")
	}
	if darkTheme.H2.Equal(lightTheme.H2) {
		t.Fatal("dark and light H2 colors should differ")
	}

	md := "# Heading\n\nText with `code`.\n"

	// Render with dark theme
	rDark := NewMarkdownRenderer(darkTheme, 80)
	blocksDark, _ := rDark.Render(md)
	if len(blocksDark) < 2 {
		t.Fatalf("dark: expected at least 2 blocks, got %d", len(blocksDark))
	}

	// Render with light theme
	rLight := NewMarkdownRenderer(lightTheme, 80)
	blocksLight, _ := rLight.Render(md)
	if len(blocksLight) < 2 {
		t.Fatalf("light: expected at least 2 blocks, got %d", len(blocksLight))
	}

	// Verify heading colors differ between themes
	var darkH1Color, lightH1Color buffer.Color
	for _, c := range blocksDark[0].Cells[0] {
		if c.Rune == 'H' {
			darkH1Color = c.Fg
			break
		}
	}
	for _, c := range blocksLight[0].Cells[0] {
		if c.Rune == 'H' {
			lightH1Color = c.Fg
			break
		}
	}
	if darkH1Color.Equal(lightH1Color) {
		t.Error("heading color should differ between dark and light themes")
	}
	if !darkH1Color.Equal(darkTheme.H1) {
		t.Errorf("dark heading: expected %v, got %v", darkTheme.H1, darkH1Color)
	}
	if !lightH1Color.Equal(lightTheme.H1) {
		t.Errorf("light heading: expected %v, got %v", lightTheme.H1, lightH1Color)
	}

	// Verify code span colors differ
	darkCodeCells := findCellsWithColor(blocksDark[1].Cells, darkTheme.CodeFg)
	lightCodeCells := findCellsWithColor(blocksLight[1].Cells, lightTheme.CodeFg)
	if len(darkCodeCells) == 0 {
		t.Error("dark theme: expected code span cells with CodeFg")
	}
	if len(lightCodeCells) == 0 {
		t.Error("light theme: expected code span cells with CodeFg")
	}
}

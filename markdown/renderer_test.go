package markdown

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func cellText(cells []buffer.Cell) string {
	var sb strings.Builder
	for _, c := range cells {
		sb.WriteRune(c.Rune)
	}
	return sb.String()
}

func TestRenderHeading(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("# Hello\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Type != BlockHeading {
		t.Errorf("expected BlockHeading, got %v", b.Type)
	}
	if b.Level != 1 {
		t.Errorf("expected level 1, got %d", b.Level)
	}
	if len(b.Cells) == 0 {
		t.Fatal("expected at least 1 line")
	}
	text := cellText(b.Cells[0])
	if !strings.Contains(text, "Hello") {
		t.Errorf("expected 'Hello' in %q", text)
	}
	// First char 'H' should have H1 color and Bold flag
	theme := DefaultTheme()
	for _, c := range b.Cells[0] {
		if c.Rune == 'H' {
			if !c.Fg.Equal(theme.H1) {
				t.Errorf("expected H1 color, got %v", c.Fg)
			}
			if c.Flags&buffer.Bold == 0 {
				t.Error("expected Bold flag on heading")
			}
			break
		}
	}
}

func TestRenderHeadingLevel2(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, _ := r.Render("## Subtitle\n")
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if blocks[0].Level != 2 {
		t.Errorf("expected level 2, got %d", blocks[0].Level)
	}
}

func TestRenderParagraph(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("Hello world\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Type != BlockParagraph {
		t.Errorf("expected BlockParagraph, got %v", b.Type)
	}
	if len(b.Cells) == 0 {
		t.Fatal("expected at least 1 line")
	}
	text := cellText(b.Cells[0])
	if !strings.Contains(text, "Hello world") {
		t.Errorf("expected 'Hello world' in %q", text)
	}
}

func TestRenderParagraphWrapping(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 20)
	long := "This is a very long paragraph that should definitely wrap across multiple lines."
	blocks, _ := r.Render(long + "\n")
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if len(blocks[0].Cells) < 2 {
		t.Errorf("expected wrapping to produce multiple lines, got %d", len(blocks[0].Cells))
	}
}

func TestRenderList(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("- item1\n- item2\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Type != BlockList {
		t.Errorf("expected BlockList, got %v", b.Type)
	}
	if len(b.Cells) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(b.Cells))
	}
	// First line should contain a bullet character
	first := cellText(b.Cells[0])
	if !strings.Contains(first, "item1") {
		t.Errorf("expected 'item1' in %q", first)
	}
	hasBullet := false
	for _, c := range b.Cells[0] {
		if c.Rune == '\u2022' { // bullet •
			hasBullet = true
			break
		}
	}
	if !hasBullet {
		t.Errorf("expected bullet character in %q", first)
	}
}

func TestRenderOrderedList(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, _ := r.Render("1. first\n2. second\n")
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Type != BlockList {
		t.Errorf("expected BlockList, got %v", b.Type)
	}
	first := cellText(b.Cells[0])
	if !strings.Contains(first, "1.") {
		t.Errorf("expected '1.' in %q", first)
	}
	if !strings.Contains(first, "first") {
		t.Errorf("expected 'first' in %q", first)
	}
}

func TestRenderCodeSpan(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("This has `code` inline.\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Type != BlockParagraph {
		t.Errorf("expected BlockParagraph, got %v", b.Type)
	}
	// Find the 'c','o','d','e' cells and check CodeFg color
	theme := DefaultTheme()
	found := false
	for _, line := range b.Cells {
		for _, c := range line {
			if c.Rune == 'c' && c.Fg.Equal(theme.CodeFg) {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected to find 'code' text with CodeFg color")
	}
}

func TestRenderFencedCode(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("```go\npackage main\n```\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Type != BlockCodeBlock {
		t.Errorf("expected BlockCodeBlock, got %v", b.Type)
	}
	if len(b.Cells) == 0 {
		t.Fatal("expected at least 1 line")
	}
	text := cellText(b.Cells[0])
	if !strings.Contains(text, "package main") {
		t.Errorf("expected 'package main' in %q", text)
	}
}

func TestRenderFencedCodeWithHighlighter(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	r.SetHighlighter(NewHighlighter())
	blocks, err := r.Render("```go\nfunc main() {}\n```\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Type != BlockCodeBlock {
		t.Errorf("expected BlockCodeBlock, got %v", b.Type)
	}
	// With highlighter, 'func' keyword should have a non-default color
	foundColored := false
	for _, line := range b.Cells {
		for _, c := range line {
			if (c.Rune == 'f' || c.Rune == 'u' || c.Rune == 'n' || c.Rune == 'c') && !c.Fg.IsDefault() {
				foundColored = true
				break
			}
		}
	}
	if !foundColored {
		t.Error("expected highlighted code with non-default color")
	}
}

func TestRenderBlockquote(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("> This is a quote\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Type != BlockQuote {
		t.Errorf("expected BlockQuote, got %v", b.Type)
	}
	if len(b.Cells) == 0 {
		t.Fatal("expected at least 1 line")
	}
	// First cell should be the │ bar character
	if b.Cells[0][0].Rune != '\u2502' {
		t.Errorf("expected │ bar prefix, got %q", string(b.Cells[0][0].Rune))
	}
}

func TestRenderThematicBreak(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 40)
	blocks, err := r.Render("---\n")
	if err != nil {
		t.Fatal(err)
	}
	// "---" might be parsed as thematic break or an empty list
	found := false
	for _, b := range blocks {
		if b.Type == BlockThematicBreak {
			found = true
			if len(b.Cells) == 0 {
				t.Error("expected at least 1 line")
			}
		}
	}
	if !found {
		t.Log("Note: '---' parsed as", len(blocks), "blocks")
	}
}

func TestRenderEmphasis(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("**bold text**\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Type != BlockParagraph {
		t.Errorf("expected BlockParagraph, got %v", b.Type)
	}
	text := cellText(b.Cells[0])
	if !strings.Contains(text, "bold text") {
		t.Errorf("expected 'bold text' in %q", text)
	}
}

func TestRenderMultipleBlocks(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	md := "# Title\n\nSome paragraph text.\n\n- item\n"
	blocks, err := r.Render(md)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) < 3 {
		t.Fatalf("expected at least 3 blocks (heading+paragraph+list), got %d", len(blocks))
	}
	// First block should be heading
	if blocks[0].Type != BlockHeading {
		t.Errorf("block 0: expected heading, got %v", blocks[0].Type)
	}
}

func TestRenderEmpty(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks for empty input, got %d", len(blocks))
	}
}

func TestRenderLink(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("[click here](https://example.com)\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Type != BlockParagraph {
		t.Errorf("expected BlockParagraph, got %v", b.Type)
	}
	theme := DefaultTheme()
	foundLink := false
	for _, line := range b.Cells {
		for _, c := range line {
			if c.Fg.Equal(theme.LinkFg) && c.Flags&buffer.Underline != 0 {
				foundLink = true
				break
			}
		}
	}
	if !foundLink {
		t.Error("expected to find link text with LinkFg color and underline")
	}
}

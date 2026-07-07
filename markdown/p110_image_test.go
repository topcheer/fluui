package markdown

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP110_Image_InlineImageWithAlt(t *testing.T) {
	md := NewMarkdownRenderer(nil, 80)
	blocks, err := md.Render("![cute cat](https://example.com/cat.png)")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	// Check that rendered text contains "image:" and alt text
	var sb strings.Builder
	for _, blk := range blocks {
		for _, line := range blk.Cells {
			for _, cell := range line {
				sb.WriteRune(cell.Rune)
			}
		}
		sb.WriteByte('\n')
	}
	text := sb.String()
	if !strings.Contains(text, "image:") {
		t.Errorf("expected 'image:' in output, got: %s", text)
	}
	if !strings.Contains(text, "cute cat") {
		t.Errorf("expected alt text 'cute cat' in output, got: %s", text)
	}
	if !strings.Contains(text, "example.com") {
		t.Errorf("expected URL in output, got: %s", text)
	}
}

func TestP110_Image_EmptyAlt(t *testing.T) {
	md := NewMarkdownRenderer(nil, 80)
	blocks, err := md.Render("![](https://example.com/image.png)")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	var sb strings.Builder
	for _, blk := range blocks {
		for _, line := range blk.Cells {
			for _, cell := range line {
				sb.WriteRune(cell.Rune)
			}
		}
	}
	text := sb.String()
	if !strings.Contains(text, "[image]") {
		t.Errorf("expected '[image]' for empty alt, got: %s", text)
	}
}

func TestP110_Image_InParagraph(t *testing.T) {
	md := NewMarkdownRenderer(nil, 80)
	blocks, err := md.Render("Here is an image: ![logo](https://example.com/logo.png) in text.")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	var sb strings.Builder
	for _, blk := range blocks {
		for _, line := range blk.Cells {
			for _, cell := range line {
				sb.WriteRune(cell.Rune)
			}
		}
		sb.WriteByte('\n')
	}
	text := sb.String()
	if !strings.Contains(text, "image:") {
		t.Errorf("expected 'image:' indicator, got: %s", text)
	}
	if !strings.Contains(text, "logo") {
		t.Errorf("expected alt text 'logo', got: %s", text)
	}
	if !strings.Contains(text, "Here is an image") {
		t.Errorf("expected surrounding text, got: %s", text)
	}
}

func TestP110_Image_ImageFgColor(t *testing.T) {
	theme := DefaultTheme()
	orange := buffer.RGB(0xFF, 0xB8, 0x6C)
	if !theme.ImageFg.Equal(orange) {
		t.Errorf("expected ImageFg to be orange, got %v", theme.ImageFg)
	}
}

func TestP110_Image_WithLinkRenderer(t *testing.T) {
	md := NewMarkdownRenderer(nil, 80)
	lr := NewLinkRenderer(true) // OSC8 enabled
	md.SetLinkRenderer(lr)
	blocks, err := md.Render("![pic](https://example.com/pic.png)")
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	// With OSC8, cells should have link metadata
	found := false
	for _, blk := range blocks {
		for _, line := range blk.Cells {
			for _, cell := range line {
				if cell.Link != nil && strings.Contains(cell.Link.URL, "example.com") {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("expected at least one cell with OSC8 link metadata")
	}
}

func TestP110_Image_MarkdownThemeFromTheme(t *testing.T) {
	mt := MarkdownThemeFromTheme(nil)
	if mt == nil {
		t.Fatal("expected non-nil for nil input")
	}
	// Should return DefaultTheme which has ImageFg set
	if mt.ImageFg.Type == 0 {
		t.Error("ImageFg should not be zero in default theme")
	}
}

func TestP110_BlockImage_Type(t *testing.T) {
	// Verify BlockImage constant exists
	var bt BlockType = BlockImage
	if bt == BlockParagraph {
		t.Error("BlockImage should differ from BlockParagraph")
	}
}

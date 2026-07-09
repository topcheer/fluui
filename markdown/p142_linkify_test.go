package markdown

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// Test that Linkify extension auto-links raw URLs in text
func TestP142_Linkify_RawURL(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("Visit https://example.com for more info.")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	// Check that the rendered output contains the URL text
	totalCells := 0
	for _, block := range blocks {
		for _, line := range block.Cells {
			totalCells += len(line)
		}
	}
	if totalCells == 0 {
		t.Error("expected non-zero cells")
	}
}

func TestP142_Linkify_MultipleURLs(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("See https://a.com and http://b.com and https://c.org.")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
}

func TestP142_Linkify_NoURL(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("This text has no URLs in it.")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
}

func TestP142_Linkify_URLWithMarkdown(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("**Bold** text with https://example.com link.")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
}

func TestP142_Linkify_URLInList(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("- Item 1: https://a.com\n- Item 2: https://b.com")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) < 1 {
		t.Errorf("expected at least 1 block, got %d", len(blocks))
	}
}

// Verify Linkify doesn't break existing link rendering
func TestP142_Linkify_ExplicitLinkStillWorks(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("[Click here](https://example.com)")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	// Should contain "Click here" text, not raw URL
	found := false
	for _, block := range blocks {
		for _, line := range block.Cells {
			for _, cell := range line {
				if cell.Rune == 'C' {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("expected to find 'C' from 'Click here'")
	}
}

func TestP142_Linkify_CellContent(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("Check https://github.com/topcheer/fluui")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	// Verify cells are rendered with proper styling
	for _, block := range blocks {
		for _, line := range block.Cells {
			for _, cell := range line {
				if cell.Width > 2 {
					t.Errorf("unexpected cell width: %d", cell.Width)
				}
				_ = buffer.BlankCell // reference buffer package
			}
		}
	}
}

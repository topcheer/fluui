package markdown

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestFormatLink_OSC8(t *testing.T) {
	lr := NewLinkRenderer(true)
	got := lr.FormatLink("click me", "https://example.com")

	// Should contain OSC8 start sequence with URL
	if !strings.Contains(got, "\x1b]8;;https://example.com\x1b\\") {
		t.Errorf("missing OSC8 start with URL, got: %q", got)
	}
	// Should contain the display text
	if !strings.Contains(got, "click me") {
		t.Errorf("missing display text, got: %q", got)
	}
	// Should contain OSC8 end sequence
	if !strings.HasSuffix(got, "\x1b]8;;\x1b\\") {
		t.Errorf("missing OSC8 end sequence, got: %q", got)
	}
}

func TestFormatLink_Fallback(t *testing.T) {
	lr := NewLinkRenderer(false) // OSC8 disabled
	got := lr.FormatLink("click me", "https://example.com")

	want := "click me (https://example.com)"
	if got != want {
		t.Errorf("FormatLink() = %q, want %q", got, want)
	}
}

func TestFormatLink_EmptyURL(t *testing.T) {
	// Empty URL should return just the text regardless of OSC8 flag
	lr := NewLinkRenderer(true)
	got := lr.FormatLink("no link", "")
	if got != "no link" {
		t.Errorf("FormatLink with empty URL = %q, want %q", got, "no link")
	}

	lr2 := NewLinkRenderer(false)
	got2 := lr2.FormatLink("no link", "")
	if got2 != "no link" {
		t.Errorf("FormatLink fallback with empty URL = %q, want %q", got2, "no link")
	}
}

func TestStripOSC8(t *testing.T) {
	lr := NewLinkRenderer(true)
	formatted := lr.FormatLink("Hello", "https://example.com")

	stripped := StripOSC8(formatted)
	if stripped != "Hello" {
		t.Errorf("StripOSC8() = %q, want %q", stripped, "Hello")
	}
}

func TestStripOSC8_NoOSC8(t *testing.T) {
	// String without OSC8 sequences should be unchanged
	s := "plain text without links"
	if got := StripOSC8(s); got != s {
		t.Errorf("StripOSC8 on plain text = %q, want %q", got, s)
	}
}

func TestStripOSC8_MultipleLinks(t *testing.T) {
	lr := NewLinkRenderer(true)
	s := lr.FormatLink("A", "https://a.com") + " and " + lr.FormatLink("B", "https://b.com")
	stripped := StripOSC8(s)
	want := "A and B"
	if stripped != want {
		t.Errorf("StripOSC8 multiple = %q, want %q", stripped, want)
	}
}

func TestExtractURLs(t *testing.T) {
	lr := NewLinkRenderer(true)
	s := lr.FormatLink("A", "https://a.com") + " " + lr.FormatLink("B", "https://b.com")

	urls := ExtractURLs(s)
	if len(urls) != 2 {
		t.Fatalf("ExtractURLs returned %d URLs, want 2", len(urls))
	}
	if urls[0] != "https://a.com" {
		t.Errorf("urls[0] = %q, want https://a.com", urls[0])
	}
	if urls[1] != "https://b.com" {
		t.Errorf("urls[1] = %q, want https://b.com", urls[1])
	}
}

func TestLinkRendererEnabled(t *testing.T) {
	lr := NewLinkRenderer(true)
	if !lr.Enabled() {
		t.Error("NewLinkRenderer(true).Enabled() = false, want true")
	}

	lr2 := NewLinkRenderer(false)
	if lr2.Enabled() {
		t.Error("NewLinkRenderer(false).Enabled() = true, want false")
	}
}

func TestFormatOSC8PackageLevel(t *testing.T) {
	got := FormatOSC8("text", "https://example.com")
	expected := "\x1b]8;;https://example.com\x1b\\text\x1b]8;;\x1b\\"
	if got != expected {
		t.Errorf("FormatOSC8() = %q, want %q", got, expected)
	}
}

// --- Integration test with full markdown render ---

func TestIntegrationLinkRendering_OSC8(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	r.SetLinkRenderer(NewLinkRenderer(true))

	blocks, err := r.Render("This is a [link](https://example.com) here.")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("no blocks returned")
	}

	// Find cells that have Link metadata
	var linkCells []buffer.Cell
	for _, line := range blocks[0].Cells {
		for _, cell := range line {
			if cell.Link != nil {
				linkCells = append(linkCells, cell)
			}
		}
	}

	if len(linkCells) == 0 {
		t.Error("expected cells with Link metadata for OSC8 link")
	}

	// Verify the URL is correct
	for _, cell := range linkCells {
		if cell.Link.URL != "https://example.com" {
			t.Errorf("Link URL = %q, want https://example.com", cell.Link.URL)
		}
	}
}

func TestIntegrationLinkRendering_Fallback(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	// Default: OSC8 disabled

	blocks, err := r.Render("This is a [link](https://example.com) here.")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("no blocks returned")
	}

	// Collect all cell text from the first block
	var allText strings.Builder
	for _, line := range blocks[0].Cells {
		for _, cell := range line {
			allText.WriteRune(cell.Rune)
		}
	}

	rendered := allText.String()

	// Should contain the URL as fallback text
	if !strings.Contains(rendered, "https://example.com") {
		t.Errorf("fallback rendering should contain URL, got: %q", rendered)
	}
	if !strings.Contains(rendered, "(https://example.com)") {
		t.Errorf("fallback should contain '(url)' format, got: %q", rendered)
	}

	// No cells should have Link metadata when OSC8 is disabled
	for _, line := range blocks[0].Cells {
		for _, cell := range line {
			if cell.Link != nil {
				t.Error("cell should not have Link metadata when OSC8 disabled")
			}
		}
	}
}

func TestIntegrationLinkRendering_EmptyURL(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	r.SetLinkRenderer(NewLinkRenderer(true))

	// Markdown with empty URL: [text]()
	blocks, err := r.Render("[text]()")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("no blocks returned")
	}

	// No cells should have Link metadata when URL is empty
	for _, line := range blocks[0].Cells {
		for _, cell := range line {
			if cell.Link != nil {
				t.Error("cell should not have Link metadata for empty URL")
			}
		}
	}
}

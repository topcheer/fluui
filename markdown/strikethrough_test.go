package markdown

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestStrikethrough_Simple(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render("This is ~~deleted~~ text")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	// Find cells with Strikethrough flag
	found := false
	for _, line := range blocks[0].Cells {
		for _, cell := range line {
			if cell.Flags&buffer.Strikethrough != 0 {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("expected at least one cell with Strikethrough flag")
	}
}

func TestStrikethrough_ContentPreserved(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render("~~deleted content~~")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	// Verify "deleted" text is present (not stripped)
	var runes []rune
	for _, line := range blocks[0].Cells {
		for _, cell := range line {
			if cell.Rune != 0 {
				runes = append(runes, cell.Rune)
			}
		}
	}
	text := string(runes)
	if !strContains(text, "deleted") {
		t.Errorf("expected 'deleted' in rendered text, got: %q", text)
	}
}

func TestStrikethrough_MixedWithBold(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render("**bold** and ~~strike~~")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	// Both bold and strike flags should be present
	boldFound, strikeFound := false, false
	for _, line := range blocks[0].Cells {
		for _, cell := range line {
			if cell.Flags&buffer.Bold != 0 {
				boldFound = true
			}
			if cell.Flags&buffer.Strikethrough != 0 {
				strikeFound = true
			}
		}
	}
	if !boldFound {
		t.Error("expected bold text")
	}
	if !strikeFound {
		t.Error("expected strikethrough text")
	}
}

func TestStrikethrough_NestedWithItalic(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render("~~*italic strike*~~")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	// Should have both italic AND strikethrough flags
	bothFound := false
	for _, line := range blocks[0].Cells {
		for _, cell := range line {
			if cell.Flags&buffer.Italic != 0 && cell.Flags&buffer.Strikethrough != 0 {
				bothFound = true
			}
		}
	}
	if !bothFound {
		t.Error("expected cells with both italic + strikethrough")
	}
}

func TestStrikethrough_NoTildes(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render("Regular text without tildes")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	for _, line := range blocks[0].Cells {
		for _, cell := range line {
			if cell.Flags&buffer.Strikethrough != 0 {
				t.Error("should not have strikethrough in regular text")
			}
		}
	}
}

func TestStrikethrough_MultipleInLine(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render("~~one~~ ~~two~~ ~~three~~")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	// Count cells with strikethrough
	count := 0
	for _, line := range blocks[0].Cells {
		for _, cell := range line {
			if cell.Flags&buffer.Strikethrough != 0 {
				count++
			}
		}
	}
	// "one" + "two" + "three" = at least 11 chars
	if count < 8 {
		t.Errorf("expected at least 8 strikethrough cells, got %d", count)
	}
}

func strContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

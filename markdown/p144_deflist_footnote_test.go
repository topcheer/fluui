package markdown

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP144_DefinitionList_Basic(t *testing.T) {
	r := NewMarkdownRenderer(nil, 60)
	input := `Apple
: A fruit

Banana
: Another fruit`
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, b := range blocks {
		if b.Type == BlockDefinitionList {
			found = true
			if len(b.Cells) < 4 {
				t.Fatalf("expected at least 4 lines, got %d", len(b.Cells))
			}
			// Term should be bold
			termLine := b.Cells[0]
			hasBold := false
			for _, cell := range termLine {
				if cell.Flags&buffer.Bold != 0 {
					hasBold = true
					break
				}
			}
			if !hasBold {
				t.Error("definition term should have bold styling")
			}
		}
	}
	if !found {
		t.Fatal("expected a BlockDefinitionList")
	}
}

func TestP144_DefinitionList_DescriptionPrefix(t *testing.T) {
	r := NewMarkdownRenderer(nil, 60)
	blocks, err := r.Render("Term\n: Description text")
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range blocks {
		if b.Type == BlockDefinitionList {
			// First description line should start with ": "
			for _, line := range b.Cells {
				if len(line) >= 1 && line[0].Rune == ':' {
					return // found the colon prefix
				}
			}
			t.Fatal("description should start with ': ' prefix")
		}
	}
	t.Fatal("no definition list block found")
}

func TestP144_DefinitionList_MultipleDescriptions(t *testing.T) {
	r := NewMarkdownRenderer(nil, 60)
	input := `Go
: A programming language
: Created at Google`
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range blocks {
		if b.Type == BlockDefinitionList {
			// Should have 1 term + 2 descriptions = 3+ lines
			if len(b.Cells) < 3 {
				t.Fatalf("expected at least 3 lines for term+2 descriptions, got %d", len(b.Cells))
			}
		}
	}
}

func TestP144_DefinitionList_LongWrapping(t *testing.T) {
	r := NewMarkdownRenderer(nil, 20)
	input := `Short
: This is a very long description that should wrap across multiple lines`
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range blocks {
		if b.Type == BlockDefinitionList {
			if len(b.Cells) < 3 {
				t.Fatalf("expected wrapping, got %d lines", len(b.Cells))
			}
		}
	}
}

func TestP144_Footnote_Basic(t *testing.T) {
	r := NewMarkdownRenderer(nil, 60)
	input := `This has a footnote[^1].

[^1]: This is the footnote content.`
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatal(err)
	}
	foundFn := false
	for _, b := range blocks {
		if b.Type == BlockFootnote {
			foundFn = true
			if len(b.Cells) == 0 {
				t.Fatal("footnote block should have content")
			}
			// Should contain the footnote text
			firstLine := b.Cells[0]
			hasRef := false
			for _, cell := range firstLine {
				if cell.Rune == '[' {
					hasRef = true
					break
				}
			}
			if !hasRef {
				t.Error("footnote line should start with [1] reference")
			}
		}
	}
	if !foundFn {
		t.Fatal("expected a BlockFootnote")
	}
}

func TestP144_Footnote_ReferenceInText(t *testing.T) {
	r := NewMarkdownRenderer(nil, 60)
	input := `See this[^1].

[^1]: Note.`
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatal(err)
	}
	// The paragraph should contain the footnote reference [1]
	for _, b := range blocks {
		if b.Type == BlockParagraph {
			for _, line := range b.Cells {
				for _, cell := range line {
					if cell.Rune == '[' && cell.Flags&buffer.Dim != 0 {
						return // found dim-styled footnote reference
					}
				}
			}
		}
	}
	t.Error("expected footnote reference [1] in paragraph text")
}

func TestP144_Footnote_Multiple(t *testing.T) {
	r := NewMarkdownRenderer(nil, 60)
	input := `First[^1] and second[^2].

[^1]: Note one.
[^2]: Note two.`
	blocks, err := r.Render(input)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range blocks {
		if b.Type == BlockFootnote {
			if len(b.Cells) < 2 {
				t.Fatalf("expected at least 2 footnote lines, got %d", len(b.Cells))
			}
		}
	}
}

func TestP144_DefinitionList_NotRegularList(t *testing.T) {
	r := NewMarkdownRenderer(nil, 60)
	// Regular dash list should NOT be a definition list
	blocks, err := r.Render("- item 1\n- item 2")
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range blocks {
		if b.Type == BlockDefinitionList {
			t.Fatal("dash list should not be parsed as definition list")
		}
	}
}

package markdown

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestHighlightGo(t *testing.T) {
	h := NewHighlighter()
	source := "package main\nfunc hello() {}"
	cells, err := h.Highlight(source, "go")
	if err != nil {
		t.Fatalf("Highlight error: %v", err)
	}
	if len(cells) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(cells))
	}

	// Verify that some cells have non-default color (highlighting applied).
	hasColor := false
	for _, row := range cells {
		for _, cell := range row {
			if !cell.Fg.Equal(buffer.Color{}) {
				hasColor = true
				break
			}
		}
		if hasColor {
			break
		}
	}
	if !hasColor {
		t.Error("expected at least one cell with syntax highlighting color")
	}

	// Verify "package" keyword has color.
	keywordColored := false
	if len(cells) > 0 {
		for _, cell := range cells[0] {
			// Find the start of "package" (after any whitespace).
			if cell.Rune == 'p' && !cell.Fg.Equal(buffer.Color{}) {
				keywordColored = true
				break
			}
		}
	}
	if !keywordColored {
		t.Error("expected 'package' keyword to have highlighting color")
	}
}

func TestHighlightUnknownLang(t *testing.T) {
	h := NewHighlighter()
	// Unknown language should not panic.
	cells, err := h.Highlight("hello world", "totally-unknown-lang-xyz")
	if err != nil {
		t.Fatalf("Highlight with unknown lang error: %v", err)
	}
	if cells == nil {
		t.Error("expected non-nil cells")
	}
}

func TestHighlightEmptySource(t *testing.T) {
	h := NewHighlighter()
	cells, err := h.Highlight("", "go")
	if err != nil {
		t.Fatalf("Highlight empty source error: %v", err)
	}
	// Should return at least one (empty) line.
	if len(cells) == 0 {
		t.Error("expected at least 1 line for empty source")
	}
}

func TestHighlightMultiline(t *testing.T) {
	h := NewHighlighter()
	source := "package main\n\nfunc main() {\n\tprintln(42)\n}"
	cells, err := h.Highlight(source, "go")
	if err != nil {
		t.Fatalf("Highlight error: %v", err)
	}
	// Source has 5 lines (separated by \n).
	if len(cells) != 5 {
		t.Errorf("expected 5 lines, got %d", len(cells))
	}
}

func TestHighlightCJK(t *testing.T) {
	h := NewHighlighter()
	// CJK characters in code comments or strings.
	source := "// 你好世界\nprint(\"你好\")"
	cells, err := h.Highlight(source, "python")
	if err != nil {
		t.Fatalf("Highlight CJK error: %v", err)
	}
	// Verify CJK characters have width 2.
	foundCJK := false
	for _, row := range cells {
		for _, cell := range row {
			if cell.Width == 2 {
				foundCJK = true
				break
			}
		}
		if foundCJK {
			break
		}
	}
	if !foundCJK {
		t.Error("expected at least one width-2 CJK cell")
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"go", "go"},
		{"python", "python"},
		{"Go", "go"},           // lowercase
		{"", "plaintext"},       // empty → plaintext
		{"  ", "plaintext"},     // whitespace only → plaintext
		{"go hl_lines=8-10", "go"},  // info string with attributes
		{"javascript linenos", "javascript"},
		{"rust", "rust"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := DetectLanguage(tt.input)
			if got != tt.want {
				t.Errorf("DetectLanguage(%q): got %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHighlightToLines(t *testing.T) {
	h := NewHighlighter()
	source := "package main"
	lines, err := h.HighlightToLines(source, "go")
	if err != nil {
		t.Fatalf("HighlightToLines error: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0] != "package main" {
		t.Errorf("got %q, want %q", lines[0], "package main")
	}
}

func TestNewHighlighterWithStyle(t *testing.T) {
	h := NewHighlighterWithStyle("monokai")
	if h == nil {
		t.Fatal("expected non-nil Highlighter")
	}
	if h.style == nil {
		t.Error("expected non-nil style")
	}

	// Invalid style should fall back to dracula.
	h2 := NewHighlighterWithStyle("nonexistent-style")
	if h2 == nil || h2.style == nil {
		t.Error("expected fallback to dracula for invalid style")
	}
}

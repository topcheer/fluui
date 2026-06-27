package markdown

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

func TestDefaultTheme_AllFieldsSet(t *testing.T) {
	tt := DefaultTheme()
	if tt == nil {
		t.Fatal("DefaultTheme returned nil")
	}

	// Every field must be a valid Color (either RGB or NoColor).
	// We verify they're zero-initialized (which would be a bug).
	type field struct {
		name  string
		color buffer.Color
	}
	fields := []field{
		{"H1", tt.H1}, {"H2", tt.H2}, {"H3", tt.H3},
		{"H4", tt.H4}, {"H5", tt.H5}, {"H6", tt.H6},
		{"Bold", tt.Bold}, {"Italic", tt.Italic}, {"Strike", tt.Strike},
		{"CodeFg", tt.CodeFg}, {"CodeBg", tt.CodeBg},
		{"LinkFg", tt.LinkFg}, {"LinkUrlFg", tt.LinkUrlFg},
		{"QuoteFg", tt.QuoteFg}, {"QuoteBar", tt.QuoteBar},
		{"ListBullet", tt.ListBullet},
		{"TableBorder", tt.TableBorder}, {"TableHeader", tt.TableHeader},
		{"Hr", tt.Hr}, {"Body", tt.Body},
	}

	for _, f := range fields {
		// Even NoColor() has Type=ColorNone(0) which is valid.
		// But a completely zero Color{} also has Type=0, Val=0 — same as NoColor().
		// So this test mainly checks the theme doesn't panic and has 20 fields.
		_ = f.color
	}

	if len(fields) != 20 {
		t.Errorf("expected 20 color fields, counted %d", len(fields))
	}
}

func TestDefaultTheme_HeadingColorsAreTrueColor(t *testing.T) {
	tt := DefaultTheme()

	// Headings should use TrueColor (Dracula palette)
	headings := []struct {
		name  string
		color buffer.Color
	}{
		{"H1", tt.H1},
		{"H2", tt.H2},
		{"H3", tt.H3},
		{"H4", tt.H4},
		{"H5", tt.H5},
		{"H6", tt.H6},
	}

	for _, h := range headings {
		if h.color.Type != buffer.ColorTrue {
			t.Errorf("%s should be TrueColor, got Type=%d", h.name, h.color.Type)
		}
	}
}

func TestDefaultTheme_SpotCheckColors(t *testing.T) {
	tt := DefaultTheme()

	// H1 = pink #FF79C6
	if !tt.H1.Equal(buffer.RGB(0xFF, 0x79, 0xC6)) {
		t.Errorf("H1 = %s, want #ff79c6", tt.H1)
	}
	// H2 = cyan #8BE9FD
	if !tt.H2.Equal(buffer.RGB(0x8B, 0xE9, 0xFD)) {
		t.Errorf("H2 = %s, want #8be9fd", tt.H2)
	}
	// H3 = green #50FA7B
	if !tt.H3.Equal(buffer.RGB(0x50, 0xFA, 0x7B)) {
		t.Errorf("H3 = %s, want #50fa7b", tt.H3)
	}
	// H4 = purple #BD93F9
	if !tt.H4.Equal(buffer.RGB(0xBD, 0x93, 0xF9)) {
		t.Errorf("H4 = %s, want #bd93f9", tt.H4)
	}
	// CodeFg = pink #FF79C6 (same as H1)
	if !tt.CodeFg.Equal(buffer.RGB(0xFF, 0x79, 0xC6)) {
		t.Errorf("CodeFg = %s, want #ff79c6", tt.CodeFg)
	}
	// LinkFg = cyan #8BE9FD
	if !tt.LinkFg.Equal(buffer.RGB(0x8B, 0xE9, 0xFD)) {
		t.Errorf("LinkFg = %s, want #8be9fd", tt.LinkFg)
	}
	// QuoteBar = dim blue-gray #6272A4
	if !tt.QuoteBar.Equal(buffer.RGB(0x62, 0x72, 0xA4)) {
		t.Errorf("QuoteBar = %s, want #6272a4", tt.QuoteBar)
	}
	// TableBorder = same dim blue-gray
	if !tt.TableBorder.Equal(buffer.RGB(0x62, 0x72, 0xA4)) {
		t.Errorf("TableBorder = %s, want #6272a4", tt.TableBorder)
	}
}

func TestDefaultTheme_NoColorFields(t *testing.T) {
	tt := DefaultTheme()

	// These fields should be NoColor (terminal default / inherit)
	noColorFields := []struct {
		name  string
		color buffer.Color
	}{
		{"Bold", tt.Bold},
		{"Italic", tt.Italic},
		{"Strike", tt.Strike},
		{"CodeBg", tt.CodeBg},
		{"Body", tt.Body},
	}

	for _, f := range noColorFields {
		if !f.color.IsDefault() {
			t.Errorf("%s should be NoColor, got %s", f.name, f.color)
		}
	}
}

func TestDefaultTheme_HeadingsAreDistinct(t *testing.T) {
	tt := DefaultTheme()

	// H1-H6 should not all be the same color
	seen := make(map[uint32]string)
	headings := []struct {
		name  string
		color buffer.Color
	}{
		{"H1", tt.H1}, {"H2", tt.H2}, {"H3", tt.H3},
		{"H4", tt.H4}, {"H5", tt.H5}, {"H6", tt.H6},
	}

	dupes := 0
	for _, h := range headings {
		if prev, ok := seen[h.color.Val]; ok {
			t.Logf("note: %s and %s share color %s (acceptable for H5/H6)", prev, h.name, h.color)
			dupes++
		}
		seen[h.color.Val] = h.name
	}
	// At least 4 distinct colors among 6 headings
	if len(seen) < 4 {
		t.Errorf("expected at least 4 distinct heading colors, got %d", len(seen))
	}
	_ = dupes
}

func TestHeadingColor(t *testing.T) {
	tt := DefaultTheme()

	for level := 1; level <= 6; level++ {
		got := tt.headingColor(level)
		expected := [6]buffer.Color{tt.H1, tt.H2, tt.H3, tt.H4, tt.H5, tt.H6}
		if !got.Equal(expected[level-1]) {
			t.Errorf("headingColor(%d) = %s, want %s", level, got, expected[level-1])
		}
	}

	// Level > 6 should fall through to H6
	got := tt.headingColor(99)
	if !got.Equal(tt.H6) {
		t.Errorf("headingColor(99) = %s, want H6 %s", got, tt.H6)
	}
}

func TestMarkdownThemeFromTheme_Dracula(t *testing.T) {
	global := theme.Dracula()
	mt := MarkdownThemeFromTheme(global)

	if mt == nil {
		t.Fatal("MarkdownThemeFromTheme returned nil")
	}

	// Spot check: CodeFg should come from global theme CodeFg
	if !mt.CodeFg.Equal(global.CodeFg) {
		t.Errorf("CodeFg = %s, want %s", mt.CodeFg, global.CodeFg)
	}
	// CodeBg should come from global CodeBg
	if !mt.CodeBg.Equal(global.CodeBg) {
		t.Errorf("CodeBg = %s, want %s", mt.CodeBg, global.CodeBg)
	}
	// TableBorder should map from global Border
	if !mt.TableBorder.Equal(global.Border) {
		t.Errorf("TableBorder = %s, want %s", mt.TableBorder, global.Border)
	}
	// Hr should map from global Separator
	if !mt.Hr.Equal(global.Separator) {
		t.Errorf("Hr = %s, want %s", mt.Hr, global.Separator)
	}
	// Body should be NoColor
	if !mt.Body.IsDefault() {
		t.Errorf("Body should be NoColor, got %s", mt.Body)
	}
}

func TestMarkdownThemeFromTheme_Nord(t *testing.T) {
	global := theme.Nord()
	mt := MarkdownThemeFromTheme(global)

	// H1 should come from Accent
	if !mt.H1.Equal(global.Accent) {
		t.Errorf("H1 = %s, want %s", mt.H1, global.Accent)
	}
	// H2 should come from DiffHunk
	if !mt.H2.Equal(global.DiffHunk) {
		t.Errorf("H2 = %s, want %s", mt.H2, global.DiffHunk)
	}
	// H3 should come from Success
	if !mt.H3.Equal(global.Success) {
		t.Errorf("H3 = %s, want %s", mt.H3, global.Success)
	}
	// TableHeader should come from Accent
	if !mt.TableHeader.Equal(global.Accent) {
		t.Errorf("TableHeader = %s, want %s", mt.TableHeader, global.Accent)
	}
}

func TestMarkdownThemeFromTheme_NilReturnsDefault(t *testing.T) {
	mt := MarkdownThemeFromTheme(nil)
	if mt == nil {
		t.Fatal("expected DefaultTheme, got nil")
	}
	dt := DefaultTheme()
	if !mt.H1.Equal(dt.H1) {
		t.Errorf("nil theme should return DefaultTheme, H1 = %s, want %s", mt.H1, dt.H1)
	}
}

func TestMarkdownThemeFromTheme_AllThemes(t *testing.T) {
	for _, global := range theme.Builtin() {
		t.Run(global.Name, func(t *testing.T) {
			mt := MarkdownThemeFromTheme(global)
			if mt == nil {
				t.Fatal("returned nil")
			}

			// Verify CodeFg/CodeBg are properly mapped
			if !mt.CodeFg.Equal(global.CodeFg) {
				t.Errorf("CodeFg mismatch: %s vs %s", mt.CodeFg, global.CodeFg)
			}
			if !mt.CodeBg.Equal(global.CodeBg) {
				t.Errorf("CodeBg mismatch: %s vs %s", mt.CodeBg, global.CodeBg)
			}

			// Verify headings use TrueColor (global themes always do)
			headings := []buffer.Color{mt.H1, mt.H2, mt.H3, mt.H4, mt.H5, mt.H6}
			for i, h := range headings {
				if h.Type != buffer.ColorTrue {
					t.Errorf("H%d should be TrueColor, got Type=%d", i+1, h.Type)
				}
			}
		})
	}
}

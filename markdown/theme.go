package markdown

import (
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// MarkdownTheme holds all colors used by the markdown renderer.
type MarkdownTheme struct {
	H1          buffer.Color
	H2          buffer.Color
	H3          buffer.Color
	H4          buffer.Color
	H5          buffer.Color
	H6          buffer.Color
	Bold        buffer.Color
	Italic      buffer.Color
	Strike      buffer.Color
	CodeFg      buffer.Color
	CodeBg      buffer.Color
	LinkFg      buffer.Color
	LinkUrlFg   buffer.Color
	QuoteFg     buffer.Color
	QuoteBar    buffer.Color
	ListBullet  buffer.Color
	TableBorder buffer.Color
	TableHeader buffer.Color
	Hr          buffer.Color
	Body        buffer.Color
	ImageFg     buffer.Color
	NoteFg      buffer.Color // GitHub [!NOTE] alert
	TipFg       buffer.Color // GitHub [!TIP] alert
	ImportantFg buffer.Color // GitHub [!IMPORTANT] alert
	WarningFg   buffer.Color // GitHub [!WARNING] alert
	CautionFg   buffer.Color // GitHub [!CAUTION] alert
}

// DefaultTheme returns a sensible dark-terminal markdown theme.
func DefaultTheme() *MarkdownTheme {
	return &MarkdownTheme{
		H1:          buffer.RGB(0xFF, 0x79, 0xC6), // pink
		H2:          buffer.RGB(0x8B, 0xE9, 0xFD), // cyan
		H3:          buffer.RGB(0x50, 0xFA, 0x7B), // green
		H4:          buffer.RGB(0xBD, 0x93, 0xF9), // purple
		H5:          buffer.RGB(0xFF, 0xB8, 0x6C), // orange
		H6:          buffer.RGB(0xF1, 0xFA, 0x8C), // yellow
		Bold:        buffer.NoColor(),            // inherit terminal default
		Italic:      buffer.NoColor(),
		Strike:      buffer.NoColor(),
		CodeFg:      buffer.RGB(0xFF, 0x79, 0xC6), // pink
		CodeBg:      buffer.NoColor(),
		LinkFg:      buffer.RGB(0x8B, 0xE9, 0xFD), // cyan
		LinkUrlFg:   buffer.RGB(0x62, 0x72, 0xA4), // dim blue
		QuoteFg:     buffer.RGB(0x62, 0x72, 0xA4), // dim
		QuoteBar:    buffer.RGB(0x62, 0x72, 0xA4),
		ListBullet:  buffer.RGB(0xFF, 0x79, 0xC6), // pink
		TableBorder: buffer.RGB(0x62, 0x72, 0xA4),
		TableHeader: buffer.RGB(0xBD, 0x93, 0xF9),
		Hr:          buffer.RGB(0x62, 0x72, 0xA4),
		Body:        buffer.NoColor(),
		ImageFg:     buffer.RGB(0xFF, 0xB8, 0x6C), // orange
		NoteFg:      buffer.RGB(0x8B, 0xE9, 0xFD), // cyan (info)
		TipFg:       buffer.RGB(0x50, 0xFA, 0x7B), // green (success)
		ImportantFg: buffer.RGB(0xBD, 0x93, 0xF9), // purple (highlight)
		WarningFg:   buffer.RGB(0xFF, 0xB8, 0x6C), // orange (warning)
		CautionFg:   buffer.RGB(0xFF, 0x55, 0x55), // red (danger)
	}
}

// MarkdownThemeFromTheme creates a MarkdownTheme by mapping colors from
// a global theme.Theme. This lets the markdown renderer adapt to the
// user's active color scheme (Dracula, Nord, Gruvbox, etc.).
func MarkdownThemeFromTheme(t *theme.Theme) *MarkdownTheme {
	if t == nil {
		return DefaultTheme()
	}
	return &MarkdownTheme{
		H1:          t.Accent,       // accent as primary heading
		H2:          t.DiffHunk,     // cyan-ish for H2
		H3:          t.Success,      // green-ish for H3
		H4:          t.DiffFile,     // purple-ish for H4
		H5:          t.Warning,      // yellow/orange for H5
		H6:          t.Warning,      // same family for H6
		Bold:        buffer.NoColor(),
		Italic:      buffer.NoColor(),
		Strike:      buffer.NoColor(),
		CodeFg:      t.CodeFg,
		CodeBg:      t.CodeBg,
		LinkFg:      t.Accent,
		LinkUrlFg:   t.Muted,
		QuoteFg:     t.Muted,
		QuoteBar:    t.Border,
		ListBullet:  t.Accent,
		TableBorder: t.Border,
		TableHeader: t.Accent,
		Hr:          t.Separator,
		Body:        buffer.NoColor(),
		ImageFg:     t.Warning,
	}
}

// headingColor returns the theme color for the given heading level.
func (t *MarkdownTheme) headingColor(level int) buffer.Color {
	switch level {
	case 1:
		return t.H1
	case 2:
		return t.H2
	case 3:
		return t.H3
	case 4:
		return t.H4
	case 5:
		return t.H5
	default:
		return t.H6
	}
}

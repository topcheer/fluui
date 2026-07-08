package markdown

import (
	"strings"

	"github.com/alecthomas/chroma"
	_ "github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/topcheer/fluui/internal/buffer"
)

// Highlighter wraps chroma lexer/style for code highlighting.
type Highlighter struct {
	style    *chroma.Style
	colorCache map[chroma.TokenType]buffer.Color // cache tokenTypeColor results
}

// NewHighlighter creates a Highlighter with the dracula theme.
func NewHighlighter() *Highlighter {
	return &Highlighter{
		style:      styles.Get("dracula"),
		colorCache: make(map[chroma.TokenType]buffer.Color),
	}
}

// NewHighlighterWithStyle creates a Highlighter with a named chroma style.
// Falls back to dracula if the style is not found.
func NewHighlighterWithStyle(styleName string) *Highlighter {
	s := styles.Get(styleName)
	if s == nil {
		s = styles.Get("dracula")
	}
	return &Highlighter{style: s, colorCache: make(map[chroma.TokenType]buffer.Color)}
}

// tokenTypeColor maps chroma token types to buffer.Color values.
// Based on common syntax highlighting color schemes.
func tokenTypeColor(tt chroma.TokenType, style *chroma.Style) buffer.Color {
	// Try to get color from the chroma style entries.
	entry := style.Get(tt)
	if entry.Colour.IsSet() {
		return buffer.RGB(entry.Colour.Red(), entry.Colour.Green(), entry.Colour.Blue())
	}

	// Fallback: manual mapping for common token types.
	switch {
	case tt == chroma.Keyword || tt.InCategory(chroma.Keyword):
		return buffer.RGB(0xFF, 0x79, 0xC6) // pink/purple
	case tt == chroma.String || tt.InCategory(chroma.String):
		return buffer.RGB(0xF1, 0xFA, 0x8C) // yellow-green
	case tt == chroma.Comment || tt.InCategory(chroma.Comment):
		return buffer.RGB(0x62, 0x72, 0xA4) // gray-blue
	case tt == chroma.Number || tt.InCategory(chroma.Number):
		return buffer.RGB(0xBD, 0x93, 0xF9) // purple
	case tt == chroma.NameFunction || tt.InCategory(chroma.NameFunction):
		return buffer.RGB(0x50, 0xFA, 0x7B) // green
	case tt == chroma.Operator || tt == chroma.Punctuation:
		return buffer.RGB(0xFF, 0x79, 0xC6) // pink
	case tt == chroma.NameBuiltin || tt == chroma.NameClass || tt == chroma.NameDecorator:
		return buffer.RGB(0x8B, 0xE9, 0xFD) // cyan
	case tt == chroma.GenericInserted:
		return buffer.RGB(0x50, 0xFA, 0x7B) // green
	case tt == chroma.GenericDeleted:
		return buffer.RGB(0xFF, 0x55, 0x55) // red
	default:
		return buffer.Color{} // default/no color
	}
}

// Highlight converts source code into highlighted Cell lines.
// Returns a [][]buffer.Cell where each inner slice is one line.
func (h *Highlighter) Highlight(source string, lang string) ([][]buffer.Cell, error) {
	var lexer chroma.Lexer

	// Get lexer for the language.
	if lang != "" && lang != "plaintext" {
		lexer = lexers.Get(lang)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Tokenize.
	iterator, err := lexer.Tokenise(nil, source)
	if err != nil {
		return nil, err
	}

	// Build cell lines.
	var lines [][]buffer.Cell
	var currentLine []buffer.Cell

	for token := iterator(); token != chroma.EOF; token = iterator() {
		// Cached color lookup (avoids style.Get map lookup per token).
		color, ok := h.colorCache[token.Type]
		if !ok {
			color = tokenTypeColor(token.Type, h.style)
			h.colorCache[token.Type] = color
		}

		// Fast path: token has no newline.
		if strings.IndexByte(token.Value, '\n') < 0 {
			// Ultra-fast path: pure ASCII token (common for code).
			if isAllASCII(token.Value) {
				for i := 0; i < len(token.Value); i++ {
					currentLine = append(currentLine, buffer.Cell{
						Rune:  rune(token.Value[i]),
						Width: 1,
						Fg:    color,
					})
				}
			} else {
				for _, r := range token.Value {
					currentLine = append(currentLine, buffer.Cell{
						Rune:  r,
						Width: uint8(buffer.RuneWidth(r)),
						Fg:    color,
					})
				}
			}
			continue
		}

		// Slow path: token contains newlines, split.
		parts := strings.Split(token.Value, "\n")
		for i, part := range parts {
			if i > 0 {
				lines = append(lines, currentLine)
				currentLine = nil
			}

			for _, r := range part {
				currentLine = append(currentLine, buffer.Cell{
					Rune:  r,
					Width: uint8(buffer.RuneWidth(r)),
					Fg:    color,
				})
			}
		}
	}

	// Flush the last line.
	lines = append(lines, currentLine)

	// Handle empty source.
	if len(lines) == 0 {
		lines = append(lines, []buffer.Cell{})
	}

	return lines, nil
}

// DetectLanguage extracts the language identifier from a fenced code block info string.
// Example: "go hl_lines=8-10" → "go"
func DetectLanguage(info string) string {
	info = strings.TrimSpace(info)
	if info == "" {
		return "plaintext"
	}
	// Language is the first word before any space.
	if idx := strings.IndexByte(info, ' '); idx > 0 {
		return strings.ToLower(info[:idx])
	}
	return strings.ToLower(info)
}

// HighlightToLines is a convenience wrapper that returns highlighted text
// as a slice of display strings (for debugging or simple rendering).
func (h *Highlighter) HighlightToLines(source string, lang string) ([]string, error) {
	cells, err := h.Highlight(source, lang)
	if err != nil {
		return nil, err
	}

	lines := make([]string, len(cells))
	for i, row := range cells {
		var sb strings.Builder
		for _, cell := range row {
			sb.WriteRune(cell.Rune)
		}
		lines[i] = sb.String()
	}
	return lines, nil
}

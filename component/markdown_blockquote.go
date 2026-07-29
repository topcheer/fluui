package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownBlockquote: Render Markdown > Blockquotes ───
//
// MarkdownBlockquote parses markdown blockquote text (lines prefixed with >)
// and renders them with an accent left border bar, indentation, and styled
// text. Supports nested blockquotes (multiple > prefixes).
//
// Usage:
//
//	bq := NewMarkdownBlockquote()
//	bq.SetMarkdown("> This is a quote.\n> Second line.\n>> Nested quote.")
//	bq.Paint(buf)

// BlockquoteStyle holds styling for MarkdownBlockquote.
type BlockquoteStyle struct {
	Text     buffer.Style
	Accent   buffer.Style // left border bar
	Nested   buffer.Style // nested quote text
	Border   buffer.Style
}

// DefaultBlockquoteStyle returns sensible defaults.
func DefaultBlockquoteStyle() BlockquoteStyle {
	text := buffer.Style{Fg: buffer.RGB(203, 213, 225), Flags: buffer.Italic} // slate-300 italic
	accent := buffer.Style{Fg: buffer.RGB(129, 140, 248)}                     // indigo-400
	nested := buffer.Style{Fg: buffer.RGB(148, 163, 184), Flags: buffer.Italic} // slate-400 italic
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}                       // slate-600
	return BlockquoteStyle{Text: text, Accent: accent, Nested: nested, Border: border}
}

// BlockquoteLine represents a parsed blockquote line.
type BlockquoteLine struct {
	Text   string
	Indent int // nesting level (number of > prefixes)
}

// MarkdownBlockquote renders markdown blockquote content.
type MarkdownBlockquote struct {
	BaseComponent
	mu sync.Mutex

	source string
	style  BlockquoteStyle

	// cached parsed lines
	cachedLines []BlockquoteLine
}

// NewMarkdownBlockquote creates a MarkdownBlockquote with defaults.
func NewMarkdownBlockquote() *MarkdownBlockquote {
	bq := &MarkdownBlockquote{
		style: DefaultBlockquoteStyle(),
	}
	bq.SetID(GenerateID("blockquote"))
	return bq
}

// SetMarkdown sets the raw markdown source and parses blockquote lines.
func (bq *MarkdownBlockquote) SetMarkdown(source string) *MarkdownBlockquote {
	bq.mu.Lock()
	bq.source = source
	bq.parseLocked()
	bq.mu.Unlock()
	return bq
}

// Markdown returns the raw markdown source.
func (bq *MarkdownBlockquote) Markdown() string {
	bq.mu.Lock()
	defer bq.mu.Unlock()
	return bq.source
}

// SetStyle sets the custom style.
func (bq *MarkdownBlockquote) SetStyle(s BlockquoteStyle) *MarkdownBlockquote {
	bq.mu.Lock()
	bq.style = s
	bq.mu.Unlock()
	return bq
}

// LineCount returns the number of parsed lines.
func (bq *MarkdownBlockquote) LineCount() int {
	bq.mu.Lock()
	defer bq.mu.Unlock()
	return len(bq.cachedLines)
}

// parseLocked parses > prefixed lines into BlockquoteLine entries.
func (bq *MarkdownBlockquote) parseLocked() {
	bq.cachedLines = bq.cachedLines[:0]
	lines := strings.Split(bq.source, "\n")
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		if !strings.HasPrefix(trimmed, ">") {
			// Non-blockquote line — include as indent 0 if it's part of the block
			if trimmed != "" {
				bq.cachedLines = append(bq.cachedLines, BlockquoteLine{
					Text:   trimmed,
					Indent: 0,
				})
			}
			continue
		}
		// Count nesting level
		indent := 0
		rest := trimmed
		for strings.HasPrefix(rest, ">") {
			indent++
			rest = rest[1:]
			rest = strings.TrimPrefix(rest, " ")
		}
		bq.cachedLines = append(bq.cachedLines, BlockquoteLine{
			Text:   rest,
			Indent: indent,
		})
	}
}

// Measure returns the preferred size.
func (bq *MarkdownBlockquote) Measure(cs Constraints) Size {
	bq.mu.Lock()
	lineCount := len(bq.cachedLines)
	bq.mu.Unlock()

	w := 50
	h := lineCount + 2 // borders
	if h < 3 {
		h = 3
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the blockquote into the buffer.
func (bq *MarkdownBlockquote) Paint(buf *buffer.Buffer) {
	bq.mu.Lock()
	defer bq.mu.Unlock()

	b := bq.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 {
		w = 50
	}
	if h < 3 {
		h = 3
	}

	// Draw border
	bs := bq.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 {
				ch = '┌'
			} else if row == 0 && col == w-1 {
				ch = '┐'
			} else if row == h-1 && col == 0 {
				ch = '└'
			} else if row == h-1 && col == w-1 {
				ch = '┘'
			} else if row == 0 || row == h-1 {
				ch = '─'
			} else if col == 0 || col == w-1 {
				ch = '│'
			}
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	// Draw each line
	for idx, line := range bq.cachedLines {
		rowY := y + 1 + idx
		if rowY >= y+h-1 || rowY >= buf.Height {
			break
		}

		col := x + 1

		// Draw accent bar(s) based on indent level
		accentStyle := bq.style.Accent
		for i := 0; i < line.Indent; i++ {
			if col >= x+w-1 || col >= buf.Width {
				break
			}
			buf.SetCell(col, rowY, buffer.Cell{Rune: '▎', Fg: accentStyle.Fg, Bg: accentStyle.Bg, Flags: accentStyle.Flags, Width: 1})
			col++
			// Indentation space
			if col >= x+w-1 || col >= buf.Width {
				break
			}
			buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: accentStyle.Fg, Bg: accentStyle.Bg, Flags: accentStyle.Flags, Width: 1})
			col++
		}

		// Choose text style based on indent
		var textStyle buffer.Style
		if line.Indent > 1 {
			textStyle = bq.style.Nested
		} else {
			textStyle = bq.style.Text
		}

		// Draw text
		for _, r := range line.Text {
			if col >= x+w-1 || col >= buf.Width {
				break
			}
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (bq *MarkdownBlockquote) Children() []Component { return nil }

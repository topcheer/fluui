package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownFootnote: Render Markdown Footnotes ───
//
// MarkdownFootnote parses [^N] footnote references and [^N]: text definitions,
// rendering superscript markers inline and footnote definitions at the bottom.
//
// Usage:
//
//	mf := NewMarkdownFootnote()
//	mf.SetMarkdown("See this[^1] for details.\n\n[^1]: https://example.com")
//	mf.Paint(buf)

// FootnoteRef represents a parsed footnote reference.
type FootnoteRef struct {
	ID   string // e.g. "1"
	Text string // definition text
}

// MarkdownFootnoteStyle holds styling.
type MarkdownFootnoteStyle struct {
	Text       buffer.Style
	RefMarker  buffer.Style // superscript [^N]
	Definition buffer.Style // footnote definition line
	Border     buffer.Style
}

// DefaultMarkdownFootnoteStyle returns defaults.
func DefaultMarkdownFootnoteStyle() MarkdownFootnoteStyle {
	text := buffer.Style{Fg: buffer.RGB(226, 232, 240)}
	ref := buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Bold}
	def := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return MarkdownFootnoteStyle{text, ref, def, border}
}

// MarkdownFootnote renders markdown footnotes.
type MarkdownFootnote struct {
	BaseComponent
	mu sync.Mutex

	source  string
	style   MarkdownFootnoteStyle
	cached  []FootnoteRef // parsed footnote definitions
	cachedTextSegments []string // text with ref markers replaced
}

// NewMarkdownFootnote creates a MarkdownFootnote.
func NewMarkdownFootnote() *MarkdownFootnote {
	mf := &MarkdownFootnote{style: DefaultMarkdownFootnoteStyle()}
	mf.SetID(GenerateID("footnote"))
	return mf
}

// SetMarkdown sets the source and parses footnotes.
func (mf *MarkdownFootnote) SetMarkdown(source string) *MarkdownFootnote {
	mf.mu.Lock()
	mf.source = source
	mf.parseLocked()
	mf.mu.Unlock()
	return mf
}

// Markdown returns the raw source.
func (mf *MarkdownFootnote) Markdown() string {
	mf.mu.Lock()
	defer mf.mu.Unlock()
	return mf.source
}

// SetStyle sets the custom style.
func (mf *MarkdownFootnote) SetStyle(s MarkdownFootnoteStyle) *MarkdownFootnote {
	mf.mu.Lock()
	mf.style = s
	mf.mu.Unlock()
	return mf
}

// FootnoteCount returns the number of footnote definitions.
func (mf *MarkdownFootnote) FootnoteCount() int {
	mf.mu.Lock()
	defer mf.mu.Unlock()
	return len(mf.cached)
}

// parseLocked parses footnote refs and definitions. Caller holds lock.
func (mf *MarkdownFootnote) parseLocked() {
	mf.cached = mf.cached[:0]
	mf.cachedTextSegments = mf.cachedTextSegments[:0]
	if mf.source == "" {
		return
	}
	lines := strings.Split(mf.source, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Check for definition: [^N]: text
		if strings.HasPrefix(trimmed, "[^") {
			closeIdx := strings.Index(trimmed, "]:")
			if closeIdx > 2 {
				id := trimmed[2:closeIdx]
				text := strings.TrimSpace(trimmed[closeIdx+2:])
				mf.cached = append(mf.cached, FootnoteRef{ID: id, Text: text})
				continue
			}
		}
		mf.cachedTextSegments = append(mf.cachedTextSegments, line)
	}
}

// Measure returns the preferred size.
func (mf *MarkdownFootnote) Measure(cs Constraints) Size {
	mf.mu.Lock()
	lineCount := len(mf.cachedTextSegments) + len(mf.cached) + 2
	mf.mu.Unlock()
	w := 50
	h := lineCount + 2
	if h < 3 { h = 3 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders footnotes into the buffer.
func (mf *MarkdownFootnote) Paint(buf *buffer.Buffer) {
	mf.mu.Lock()
	defer mf.mu.Unlock()

	b := mf.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 { w = 50 }
	if h < 3 { h = 3 }

	bs := mf.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	textStyle := mf.style.Text
	refStyle := mf.style.RefMarker
	defStyle := mf.style.Definition

	rowY := y + 1
	col := x + 1

	// Render text segments (with inline [^N] markers)
	for _, line := range mf.cachedTextSegments {
		if rowY >= y+h-1 || rowY >= buf.Height { break }
		col = x + 1
		for _, r := range line {
			if col >= x+w-1 || col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
			col++
		}
		rowY++
	}

	// Separator before footnotes
	if len(mf.cached) > 0 && rowY < y+h-1 && rowY < buf.Height {
		for c := x + 1; c < x+w-1 && c < buf.Width; c++ {
			buf.SetCell(c, rowY, buffer.Cell{Rune: '─', Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
		}
		rowY++
	}

	// Render footnote definitions
	for _, fn := range mf.cached {
		if rowY >= y+h-1 || rowY >= buf.Height { break }
		col = x + 1
		// [N] prefix
		markerText := "[" + fn.ID + "] "
		for _, r := range markerText {
			if col >= x+w-1 || col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: refStyle.Fg, Bg: refStyle.Bg, Flags: refStyle.Flags, Width: 1})
			col++
		}
		for _, r := range fn.Text {
			if col >= x+w-1 || col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: defStyle.Fg, Bg: defStyle.Bg, Flags: defStyle.Flags, Width: 1})
			col++
		}
		rowY++
	}
}

// Children returns nil.
func (mf *MarkdownFootnote) Children() []Component { return nil }

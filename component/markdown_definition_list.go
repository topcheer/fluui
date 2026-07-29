package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownDefinitionList: Render Definition Lists ───
//
// MarkdownDefinitionList parses Term\n: Definition pairs and renders them
// with indented descriptions under each term.
//
// Usage:
//
//	dl := NewMarkdownDefinitionList()
//	dl.SetMarkdown("Go\n: A compiled language\nRust\n: A systems language")
//	dl.Paint(buf)

// DefinitionEntry represents a term-definition pair.
type DefinitionEntry struct {
	Term       string
	Definition string
}

// DefinitionListStyle holds styling.
type DefinitionListStyle struct {
	Term       buffer.Style
	Definition buffer.Style
	Border     buffer.Style
}

// DefaultDefinitionListStyle returns defaults.
func DefaultDefinitionListStyle() DefinitionListStyle {
	term := buffer.Style{Fg: buffer.RGB(167, 139, 250), Flags: buffer.Bold}
	def := buffer.Style{Fg: buffer.RGB(203, 213, 225)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return DefinitionListStyle{Term: term, Definition: def, Border: border}
}

// MarkdownDefinitionList renders markdown definition lists.
type MarkdownDefinitionList struct {
	BaseComponent
	mu sync.Mutex

	source string
	style  DefinitionListStyle
	cached []DefinitionEntry
}

// NewMarkdownDefinitionList creates a MarkdownDefinitionList.
func NewMarkdownDefinitionList() *MarkdownDefinitionList {
	dl := &MarkdownDefinitionList{style: DefaultDefinitionListStyle()}
	dl.SetID(GenerateID("deflist"))
	return dl
}

// SetMarkdown sets source and parses definition entries.
func (dl *MarkdownDefinitionList) SetMarkdown(source string) *MarkdownDefinitionList {
	dl.mu.Lock()
	dl.source = source
	dl.parseLocked()
	dl.mu.Unlock()
	return dl
}

// Markdown returns the raw source.
func (dl *MarkdownDefinitionList) Markdown() string {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	return dl.source
}

// SetStyle sets custom style.
func (dl *MarkdownDefinitionList) SetStyle(s DefinitionListStyle) *MarkdownDefinitionList {
	dl.mu.Lock()
	dl.style = s
	dl.mu.Unlock()
	return dl
}

// TermCount returns the number of definition entries.
func (dl *MarkdownDefinitionList) TermCount() int {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	return len(dl.cached)
}

// parseLocked parses definition list entries. Caller holds lock.
func (dl *MarkdownDefinitionList) parseLocked() {
	dl.cached = dl.cached[:0]
	if dl.source == "" {
		return
	}
	lines := strings.Split(dl.source, "\n")
	var currentTerm string

	for _, line := range lines {
		trimmed := line
		// Check for definition marker: leading ": "
		if strings.HasPrefix(strings.TrimSpace(trimmed), ": ") {
			def := strings.TrimSpace(trimmed)[2:]
			if currentTerm != "" {
				dl.cached = append(dl.cached, DefinitionEntry{Term: currentTerm, Definition: def})
				currentTerm = ""
			}
			continue
		}
		// Check for definition marker with tab indent: "\t: "
		if strings.HasPrefix(trimmed, "\t:") || strings.HasPrefix(trimmed, "    :") {
			def := strings.TrimSpace(trimmed)
			def = strings.TrimPrefix(def, ":")
			def = strings.TrimSpace(def)
			if currentTerm != "" {
				dl.cached = append(dl.cached, DefinitionEntry{Term: currentTerm, Definition: def})
				currentTerm = ""
			}
			continue
		}
		if strings.TrimSpace(trimmed) != "" {
			currentTerm = strings.TrimSpace(trimmed)
		}
	}
}

// Measure returns the preferred size.
func (dl *MarkdownDefinitionList) Measure(cs Constraints) Size {
	dl.mu.Lock()
	count := len(dl.cached)
	dl.mu.Unlock()
	w := 50
	h := count*2 + 2 // term + definition per entry + borders
	if h < 3 { h = 3 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the definition list into the buffer.
func (dl *MarkdownDefinitionList) Paint(buf *buffer.Buffer) {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	b := dl.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 { w = 50 }
	if h < 3 { h = 3 }

	bs := dl.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	termStyle := dl.style.Term
	defStyle := dl.style.Definition
	rowY := y + 1

	for _, entry := range dl.cached {
		if rowY >= y+h-1 || rowY >= buf.Height { break }

		// Term line
		col := x + 1
		for _, r := range entry.Term {
			if col >= x+w-1 || col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: termStyle.Fg, Bg: termStyle.Bg, Flags: termStyle.Flags, Width: 1})
			col++
		}
		rowY++
		if rowY >= y+h-1 || rowY >= buf.Height { break }

		// Definition line (indented)
		col = x + 3
		// Indent prefix
		for i := 0; i < 2; i++ {
			if col < x+w-1 && col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: defStyle.Fg, Bg: defStyle.Bg, Flags: defStyle.Flags, Width: 1})
			}
			col++
		}
		for _, r := range entry.Definition {
			if col >= x+w-1 || col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: defStyle.Fg, Bg: defStyle.Bg, Flags: defStyle.Flags, Width: 1})
			col++
		}
		rowY++
	}
}

// Children returns nil.
func (dl *MarkdownDefinitionList) Children() []Component { return nil }

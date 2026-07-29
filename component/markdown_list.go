package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownList: Render Markdown Ordered and Unordered Lists ───
//
// MarkdownList parses markdown list items (-, *, + for unordered; 1. for
// ordered) with indentation-based nesting, rendering them with bullet chars
// (•, ◦, ▪) at different levels.
//
// Usage:
//
//	ml := NewMarkdownList()
//	ml.SetMarkdown("- Item one\n  - Nested item\n- Item two")
//	ml.Paint(buf)

// MarkdownListStyle holds styling for MarkdownList.
type MarkdownListStyle struct {
	Text     buffer.Style
	Bullet   [3]buffer.Style // level 0, 1, 2+ bullet styles
	Number   buffer.Style
	Border   buffer.Style
}

// DefaultMarkdownListStyle returns sensible defaults.
func DefaultMarkdownListStyle() MarkdownListStyle {
	text := buffer.Style{Fg: buffer.RGB(226, 232, 240)}    // slate-200
	b0 := buffer.Style{Fg: buffer.RGB(96, 165, 250)}       // blue-400
	b1 := buffer.Style{Fg: buffer.RGB(167, 139, 250)}      // violet-400
	b2 := buffer.Style{Fg: buffer.RGB(244, 114, 182)}      // pink-400
	num := buffer.Style{Fg: buffer.RGB(251, 146, 60)}      // orange-400
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}    // slate-600
	return MarkdownListStyle{Text: text, Bullet: [3]buffer.Style{b0, b1, b2}, Number: num, Border: border}
}

// ListLineType classifies a parsed list line.
type ListLineType int

const (
	listUnordered ListLineType = iota
	listOrdered
	listBlank
)

// ListLine represents a parsed list item.
type ListLine struct {
	Text     string
	Type     ListLineType
	Indent   int    // nesting level (0-based)
	Number   int    // for ordered lists
}

// MarkdownList renders markdown ordered and unordered lists.
type MarkdownList struct {
	BaseComponent
	mu sync.Mutex

	source string
	style  MarkdownListStyle

	// cached parsed lines
	cachedLines   []ListLine
	cachedListType string // "ordered", "unordered", "mixed", ""
}

// NewMarkdownList creates a MarkdownList with defaults.
func NewMarkdownList() *MarkdownList {
	ml := &MarkdownList{
		style: DefaultMarkdownListStyle(),
	}
	ml.SetID(GenerateID("mdlist"))
	return ml
}

// SetMarkdown sets the raw markdown source and parses list items.
func (ml *MarkdownList) SetMarkdown(source string) *MarkdownList {
	ml.mu.Lock()
	ml.source = source
	ml.parseLocked()
	ml.mu.Unlock()
	return ml
}

// Markdown returns the raw markdown source.
func (ml *MarkdownList) Markdown() string {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	return ml.source
}

// SetStyle sets the custom style.
func (ml *MarkdownList) SetStyle(s MarkdownListStyle) *MarkdownList {
	ml.mu.Lock()
	ml.style = s
	ml.mu.Unlock()
	return ml
}

// ItemCount returns the number of list items (excluding blanks).
func (ml *MarkdownList) ItemCount() int {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	count := 0
	for _, line := range ml.cachedLines {
		if line.Type != listBlank {
			count++
		}
	}
	return count
}

// ListType returns the list type: "ordered", "unordered", "mixed", or "".
func (ml *MarkdownList) ListType() string {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	return ml.cachedListType
}

// parseLocked parses markdown list items. Caller must hold lock.
func (ml *MarkdownList) parseLocked() {
	ml.cachedLines = ml.cachedLines[:0]
	ml.cachedListType = ""
	if ml.source == "" {
		return
	}

	hasOrdered := false
	hasUnordered := false
	orderedCounters := map[int]int{} // per-indent-level counter

	lines := strings.Split(ml.source, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			ml.cachedLines = append(ml.cachedLines, ListLine{Type: listBlank})
			continue
		}

		// Count leading spaces (indent level)
		indent := 0
		spaces := 0
		for _, ch := range line {
			if ch == ' ' {
				spaces++
				if spaces >= 2 {
					indent++
					spaces = 0
				}
			} else if ch == '\t' {
				indent++
			} else {
				break
			}
		}

		content := strings.TrimSpace(line)

		// Check for unordered list markers
		if strings.HasPrefix(content, "- ") || strings.HasPrefix(content, "* ") || strings.HasPrefix(content, "+ ") {
			text := content[2:]
			ml.cachedLines = append(ml.cachedLines, ListLine{
				Text:   text,
				Type:   listUnordered,
				Indent: indent,
			})
			hasUnordered = true
			continue
		}

		// Check for ordered list markers (N. or N))
		dotIdx := strings.Index(content, ". ")
		parenIdx := strings.Index(content, ") ")
		markerEnd := -1
		markerType := -1

		if dotIdx > 0 && dotIdx <= 4 {
			markerEnd = dotIdx
			markerType = 0 // dot
		}
		if parenIdx > 0 && parenIdx <= 4 && (markerEnd < 0 || parenIdx < markerEnd) {
			markerEnd = parenIdx
			markerType = 1 // paren
		}

		if markerEnd > 0 {
			numStr := content[:markerEnd]
			num := 0
			for _, ch := range numStr {
				if ch >= '0' && ch <= '9' {
					num = num*10 + int(ch-'0')
				} else {
					num = 0
					break
				}
			}
			if num > 0 {
				textStart := markerEnd + 2
				if markerType == 1 {
					textStart = markerEnd + 2
				}
				text := content[textStart:]

				// Auto-number if number is 1, otherwise use explicit
				if num == 1 {
					orderedCounters[indent] = 1
				} else {
					orderedCounters[indent] = num
				}
				num = orderedCounters[indent]
				orderedCounters[indent]++

				ml.cachedLines = append(ml.cachedLines, ListLine{
					Text:   text,
					Type:   listOrdered,
					Indent: indent,
					Number: num,
				})
				hasOrdered = true
				continue
			}
		}

		// Not a list item — treat as text with indent
		ml.cachedLines = append(ml.cachedLines, ListLine{
			Text:   content,
			Type:   listUnordered,
			Indent: indent,
		})
		hasUnordered = true
	}

	// Determine list type
	if hasOrdered && hasUnordered {
		ml.cachedListType = "mixed"
	} else if hasOrdered {
		ml.cachedListType = "ordered"
	} else if hasUnordered {
		ml.cachedListType = "unordered"
	}
}

// Measure returns the preferred size.
func (ml *MarkdownList) Measure(cs Constraints) Size {
	ml.mu.Lock()
	lineCount := len(ml.cachedLines)
	ml.mu.Unlock()

	w := 50
	h := lineCount + 2
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

// Paint renders the list into the buffer.
func (ml *MarkdownList) Paint(buf *buffer.Buffer) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	b := ml.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 {
		w = 50
	}
	if h < 3 {
		h = 3
	}

	// Draw border
	bs := ml.style.Border
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
	textStyle := ml.style.Text
	for idx, line := range ml.cachedLines {
		rowY := y + 1 + idx
		if rowY >= y+h-1 || rowY >= buf.Height {
			break
		}
		if line.Type == listBlank {
			continue
		}

		col := x + 1

		// Indentation (2 spaces per level)
		for i := 0; i < line.Indent*2; i++ {
			if col >= x+w-1 || col >= buf.Width {
				break
			}
			buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
			col++
		}

		if line.Type == listOrdered {
			// Draw number + dot
			numStyle := ml.style.Number
			numStr := itoa(line.Number) + ". "
			for _, r := range numStr {
				if col >= x+w-1 || col >= buf.Width {
					break
				}
				buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: numStyle.Fg, Bg: numStyle.Bg, Flags: numStyle.Flags, Width: 1})
				col++
			}
		} else {
			// Draw bullet based on indent level
			bulletIdx := line.Indent
			if bulletIdx > 2 {
				bulletIdx = 2
			}
			bulletStyle := ml.style.Bullet[bulletIdx]
			bulletChars := [3]rune{'•', '◦', '▪'}
			bulletChar := bulletChars[bulletIdx]
			if col < x+w-1 && col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: bulletChar, Fg: bulletStyle.Fg, Bg: bulletStyle.Bg, Flags: bulletStyle.Flags, Width: 1})
			}
			col++
			if col < x+w-1 && col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: bulletStyle.Fg, Bg: bulletStyle.Bg, Flags: bulletStyle.Flags, Width: 1})
			}
			col++
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
func (ml *MarkdownList) Children() []Component { return nil }

package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownInlineCode: Render Inline Code Spans and Fenced Blocks ───
//
// MarkdownInlineCode parses markdown text containing backtick-delimited inline
// code (`code`) and fenced code blocks (```lang ... ```), rendering them with
// distinct background styling. Useful in markdown preview and AI response
// rendering.
//
// Usage:
//
//	mic := NewMarkdownInlineCode()
//	mic.SetMarkdown("Use `fmt.Println` to print.\n```go\nfunc main() {}\n```")
//	mic.Paint(buf)

// InlineCodeStyle holds styling for MarkdownInlineCode.
type InlineCodeStyle struct {
	Text       buffer.Style
	InlineCode buffer.Style // `code` spans
	CodeBlock  buffer.Style // fenced ``` blocks
	BlockLabel buffer.Style // language label
	Border     buffer.Style
}

// DefaultInlineCodeStyle returns sensible defaults.
func DefaultInlineCodeStyle() InlineCodeStyle {
	text := buffer.Style{Fg: buffer.RGB(226, 232, 240)}    // slate-200
	inline := buffer.Style{Fg: buffer.RGB(251, 146, 60), Bg: buffer.RGB(30, 41, 59)} // orange-400 on slate-800
	block := buffer.Style{Fg: buffer.RGB(134, 239, 172), Bg: buffer.RGB(15, 23, 42)} // green-300 on slate-900
	label := buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Bold}  // blue-400 bold
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}   // slate-600
	return InlineCodeStyle{Text: text, InlineCode: inline, CodeBlock: block, BlockLabel: label, Border: border}
}

// CodeSegmentType classifies a rendered segment.
type CodeSegmentType int

const (
	segText       CodeSegmentType = iota
	segInlineCode
	segCodeBlock
)

// CodeSegment represents a parsed line or segment.
type CodeSegment struct {
	Text     string
	Type     CodeSegmentType
	Language string // for fenced blocks
}

// MarkdownInlineCode renders inline code spans and fenced code blocks.
type MarkdownInlineCode struct {
	BaseComponent
	mu sync.Mutex

	source string
	style  InlineCodeStyle

	// cached parsed segments
	cachedSegments []CodeSegment
}

// NewMarkdownInlineCode creates a MarkdownInlineCode with defaults.
func NewMarkdownInlineCode() *MarkdownInlineCode {
	mic := &MarkdownInlineCode{
		style: DefaultInlineCodeStyle(),
	}
	mic.SetID(GenerateID("inlinecode"))
	return mic
}

// SetMarkdown sets the raw markdown source and parses code spans/blocks.
func (mic *MarkdownInlineCode) SetMarkdown(source string) *MarkdownInlineCode {
	mic.mu.Lock()
	mic.source = source
	mic.parseLocked()
	mic.mu.Unlock()
	return mic
}

// Markdown returns the raw markdown source.
func (mic *MarkdownInlineCode) Markdown() string {
	mic.mu.Lock()
	defer mic.mu.Unlock()
	return mic.source
}

// SetStyle sets the custom style.
func (mic *MarkdownInlineCode) SetStyle(s InlineCodeStyle) *MarkdownInlineCode {
	mic.mu.Lock()
	mic.style = s
	mic.mu.Unlock()
	return mic
}

// InlineCodeCount returns the number of inline code spans.
func (mic *MarkdownInlineCode) InlineCodeCount() int {
	mic.mu.Lock()
	defer mic.mu.Unlock()
	count := 0
	for _, seg := range mic.cachedSegments {
		if seg.Type == segInlineCode {
			count++
		}
	}
	return count
}

// CodeBlockCount returns the number of fenced code blocks.
func (mic *MarkdownInlineCode) CodeBlockCount() int {
	mic.mu.Lock()
	defer mic.mu.Unlock()
	count := 0
	for _, seg := range mic.cachedSegments {
		if seg.Type == segCodeBlock {
			count++
		}
	}
	return count
}

// parseLocked parses markdown into segments with code detection.
func (mic *MarkdownInlineCode) parseLocked() {
	mic.cachedSegments = mic.cachedSegments[:0]
	if mic.source == "" {
		return
	}

	lines := strings.Split(mic.source, "\n")
	inFence := false
	fenceLang := ""

	for _, line := range lines {
		// Check for fenced code block start/end
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			if !inFence {
				inFence = true
				fenceLang = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "```"))
				// Add label segment
				mic.cachedSegments = append(mic.cachedSegments, CodeSegment{
					Text:     fenceLang,
					Type:     segCodeBlock,
					Language: fenceLang,
				})
				continue
			} else {
				inFence = false
				fenceLang = ""
				continue
			}
		}

		if inFence {
			mic.cachedSegments = append(mic.cachedSegments, CodeSegment{
				Text:     line,
				Type:     segCodeBlock,
				Language: fenceLang,
			})
			continue
		}

		// Parse inline code spans within the line
		mic.parseInlineSpansLocked(line)
	}
}

// parseInlineSpansLocked parses backtick-delimited inline code in a line.
func (mic *MarkdownInlineCode) parseInlineSpansLocked(line string) {
	remaining := line
	for {
		idx := strings.Index(remaining, "`")
		if idx < 0 {
			if remaining != "" {
				mic.cachedSegments = append(mic.cachedSegments, CodeSegment{Text: remaining, Type: segText})
			}
			break
		}

		// Text before backtick
		if idx > 0 {
			mic.cachedSegments = append(mic.cachedSegments, CodeSegment{Text: remaining[:idx], Type: segText})
		} else if idx == 0 && len(mic.cachedSegments) > 0 && mic.cachedSegments[len(mic.cachedSegments)-1].Type == segText {
			// Merge with previous text segment if starts at beginning
		}

		afterTick := remaining[idx+1:]
		endIdx := strings.Index(afterTick, "`")
		if endIdx < 0 {
			// No closing backtick — treat as text
			mic.cachedSegments = append(mic.cachedSegments, CodeSegment{Text: remaining[idx:], Type: segText})
			break
		}

		codeContent := afterTick[:endIdx]
		mic.cachedSegments = append(mic.cachedSegments, CodeSegment{Text: codeContent, Type: segInlineCode})
		remaining = afterTick[endIdx+1:]
	}
}

// Measure returns the preferred size.
func (mic *MarkdownInlineCode) Measure(cs Constraints) Size {
	mic.mu.Lock()
	segCount := len(mic.cachedSegments)
	mic.mu.Unlock()

	w := 50
	h := segCount + 2
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

// Paint renders the inline code content into the buffer.
func (mic *MarkdownInlineCode) Paint(buf *buffer.Buffer) {
	mic.mu.Lock()
	defer mic.mu.Unlock()

	b := mic.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 {
		w = 50
	}
	if h < 3 {
		h = 3
	}

	// Draw border
	bs := mic.style.Border
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

	// Draw each segment
	rowY := y + 1
	for _, seg := range mic.cachedSegments {
		if rowY >= y+h-1 || rowY >= buf.Height {
			break
		}

		switch seg.Type {
		case segInlineCode:
			// Draw with background style + surrounding spaces
			codeStyle := mic.style.InlineCode
			col := x + 1
			// Leading space with code bg
			if col < x+w-1 && col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: codeStyle.Fg, Bg: codeStyle.Bg, Flags: codeStyle.Flags, Width: 1})
			}
			col++
			for _, r := range seg.Text {
				if col >= x+w-1 || col >= buf.Width {
					break
				}
				buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: codeStyle.Fg, Bg: codeStyle.Bg, Flags: codeStyle.Flags, Width: 1})
				col++
			}
			if col < x+w-1 && col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: codeStyle.Fg, Bg: codeStyle.Bg, Flags: codeStyle.Flags, Width: 1})
			}
		case segCodeBlock:
			// Check if this is the language label (first block segment)
			if seg.Language != "" && seg.Text == seg.Language {
				labelStyle := mic.style.BlockLabel
				col := x + 1
				for _, r := range seg.Text {
					if col >= x+w-1 || col >= buf.Width {
						break
					}
					buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
					col++
				}
			} else {
				// Code block line
				blockStyle := mic.style.CodeBlock
				col := x + 2
				for _, r := range seg.Text {
					if col >= x+w-1 || col >= buf.Width {
						break
					}
					buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: blockStyle.Fg, Bg: blockStyle.Bg, Flags: blockStyle.Flags, Width: 1})
					col++
				}
			}
		default:
			// Normal text
			textStyle := mic.style.Text
			col := x + 1
			for _, r := range seg.Text {
				if col >= x+w-1 || col >= buf.Width {
					break
				}
				buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
				col++
			}
		}
		rowY++
	}
}

// Children returns nil.
func (mic *MarkdownInlineCode) Children() []Component { return nil }

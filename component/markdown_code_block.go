package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownCodeBlock: Fenced Code Block with Language Label ───
//
// MarkdownCodeBlock renders a fenced code block with a language label,
// optional line numbers, and a copy indicator. Designed for AI response
// rendering where code snippets need clear visual boundaries.
//
// Usage:
//
//	cb := NewMarkdownCodeBlock()
//	cb.SetLanguage("go")
//	cb.SetLines([]string{"func main() {", "    fmt.Println(\"hi\")", "}"})
//	cb.Paint(buf)

// MarkdownCodeBlockStyle holds styling.
type MarkdownCodeBlockStyle struct {
	Header    buffer.Style
	Code      buffer.Style
	LineNumber buffer.Style
	Border    buffer.Style
	CopyIcon  buffer.Style
}

// DefaultMarkdownCodeBlockStyle returns defaults.
func DefaultMarkdownCodeBlockStyle() MarkdownCodeBlockStyle {
	return MarkdownCodeBlockStyle{
		Header:     buffer.Style{Fg: buffer.RGB(148, 163, 184), Flags: buffer.Bold},
		Code:       buffer.Style{Fg: buffer.RGB(226, 232, 240)},
		LineNumber: buffer.Style{Fg: buffer.RGB(71, 85, 105)},
		Border:     buffer.Style{Fg: buffer.RGB(51, 65, 85)},
		CopyIcon:   buffer.Style{Fg: buffer.RGB(59, 130, 246)},
	}
}

const codeBlockMaxLines = 64

// MarkdownCodeBlock renders a fenced code block.
type MarkdownCodeBlock struct {
	BaseComponent
	mu sync.Mutex

	language   string
	lines      [codeBlockMaxLines]string
	count      int
	showLineNum bool
	width      int
	style      MarkdownCodeBlockStyle
	// cached
	langLabel  string
}

// NewMarkdownCodeBlock creates a MarkdownCodeBlock.
func NewMarkdownCodeBlock() *MarkdownCodeBlock {
	cb := &MarkdownCodeBlock{width: 40, showLineNum: true, style: DefaultMarkdownCodeBlockStyle()}
	cb.SetID(GenerateID("mdblock"))
	cb.recomputeLocked()
	return cb
}

// SetLanguage sets the code language label.
func (cb *MarkdownCodeBlock) SetLanguage(lang string) *MarkdownCodeBlock {
	cb.mu.Lock()
	cb.language = lang
	cb.recomputeLocked()
	cb.mu.Unlock()
	return cb
}

// SetLines sets the code lines.
func (cb *MarkdownCodeBlock) SetLines(lines []string) *MarkdownCodeBlock {
	cb.mu.Lock()
	cb.count = 0
	for _, l := range lines {
		if cb.count >= codeBlockMaxLines { break }
		cb.lines[cb.count] = l
		cb.count++
	}
	cb.mu.Unlock()
	return cb
}

// SetShowLineNumbers toggles line number display.
func (cb *MarkdownCodeBlock) SetShowLineNumbers(show bool) *MarkdownCodeBlock {
	cb.mu.Lock()
	cb.showLineNum = show
	cb.mu.Unlock()
	return cb
}

func (cb *MarkdownCodeBlock) recomputeLocked() {
	if cb.language == "" {
		cb.langLabel = "code"
	} else {
		cb.langLabel = cb.language
	}
}

// LineCount returns the number of code lines.
func (cb *MarkdownCodeBlock) LineCount() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.count
}

// SetWidth sets the block width.
func (cb *MarkdownCodeBlock) SetWidth(w int) *MarkdownCodeBlock {
	cb.mu.Lock()
	if w < 10 { w = 10 }
	cb.width = w
	cb.mu.Unlock()
	return cb
}

// SetStyle sets custom style.
func (cb *MarkdownCodeBlock) SetStyle(s MarkdownCodeBlockStyle) *MarkdownCodeBlock {
	cb.mu.Lock()
	cb.style = s
	cb.mu.Unlock()
	return cb
}

// Measure returns preferred size.
func (cb *MarkdownCodeBlock) Measure(cs Constraints) Size {
	cb.mu.Lock()
	h := cb.count + 2 // header + border
	cb.mu.Unlock()
	w := cb.width
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: h}
}

// Paint renders the code block.
func (cb *MarkdownCodeBlock) Paint(buf *buffer.Buffer) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	b := cb.Bounds()
	x, y := b.X, b.Y
	w := b.W
	if w < 10 { w = cb.width }

	headerStyle := cb.style.Header
	codeStyle := cb.style.Code
	lineNumStyle := cb.style.LineNumber
	borderStyle := cb.style.Border
	copyStyle := cb.style.CopyIcon

	// Header row: language label + copy icon
	col := x
	for _, r := range cb.langLabel {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: headerStyle.Fg, Bg: headerStyle.Bg, Flags: headerStyle.Flags, Width: 1})
		col++
	}
	// Right-align copy icon
	copyLabel := "⎘"
	copyStart := x + w - 1 - len(copyLabel)
	for c := col; c < copyStart && c < buf.Width; c++ {
		buf.SetCell(c, y, buffer.Cell{Rune: '─', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
	}
	for i, r := range copyLabel {
		cx := copyStart + i
		if cx >= buf.Width { break }
		buf.SetCell(cx, y, buffer.Cell{Rune: r, Fg: copyStyle.Fg, Bg: copyStyle.Bg, Flags: copyStyle.Flags, Width: 1})
	}

	// Code lines
	for i := 0; i < cb.count; i++ {
		yy := y + 1 + i
		if yy >= buf.Height { break }
		col = x

		// Line number
		if cb.showLineNum {
			ln := itoa(i + 1)
			for _, r := range ln {
				if col >= buf.Width { break }
				buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: lineNumStyle.Fg, Bg: lineNumStyle.Bg, Flags: lineNumStyle.Flags, Width: 1})
				col++
			}
			if col < buf.Width {
				buf.SetCell(col, yy, buffer.Cell{Rune: ' ', Fg: lineNumStyle.Fg, Bg: lineNumStyle.Bg, Flags: lineNumStyle.Flags, Width: 1})
				col++
			}
		}

		// Code text
		for _, r := range cb.lines[i] {
			if col >= buf.Width { break }
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: codeStyle.Fg, Bg: codeStyle.Bg, Flags: codeStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (cb *MarkdownCodeBlock) Children() []Component { return nil }

package component

import (
	"fmt"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/markdown"
	"github.com/topcheer/fluui/theme"
)

// CodeBlock is a component that renders syntax-highlighted source code with
// optional line numbers and a title bar. It wraps the markdown.Highlighter
// (chroma-based) so any of the 300+ languages supported by chroma work out
// of the box.
//
// Features:
//   - Syntax highlighting via chroma (Go, Python, Rust, JS, SQL, YAML, etc.)
//   - Optional line numbers (toggleable)
//   - Optional title bar (filename / language label)
//   - Vertical scrolling for long code
//   - Thread-safe via sync.RWMutex
type CodeBlock struct {
	BaseComponent
	mu sync.RWMutex

	// source is the raw code text.
	source string
	// language identifies the syntax (e.g. "go", "python", "rust").
	language string
	// title is shown in the title bar (typically a filename).
	title string

	// display options
	showLineNumbers bool
	showTitle       bool

	// highlighter is reused across paints; lazily initialised.
	highlighter *markdown.Highlighter

	// rendered lines (highlighted)
	lines [][]buffer.Cell

	// scroll state
	scrollOffset int

	// theme for title bar styling
	currentTheme *theme.Theme
}

// NewCodeBlock creates a CodeBlock with the given language and source.
// Syntax highlighting is enabled by default. Line numbers are off by default.
func NewCodeBlock(language, source string) *CodeBlock {
	cb := &CodeBlock{
		source:           source,
		language:         language,
		showLineNumbers:  false,
		showTitle:        false,
		highlighter:      markdown.NewHighlighter(),
	}
	cb.SetID(GenerateID("codeblock"))
	cb.rehighlight()
	return cb
}

// SetSource updates the code text and re-highlights.
func (cb *CodeBlock) SetSource(source string) {
	cb.mu.Lock()
	cb.source = source
	cb.rehighlightLocked()
	cb.scrollOffset = 0
	cb.mu.Unlock()
}

// SetLanguage changes the syntax language and re-highlights.
func (cb *CodeBlock) SetLanguage(lang string) {
	cb.mu.Lock()
	cb.language = lang
	cb.rehighlightLocked()
	cb.scrollOffset = 0
	cb.mu.Unlock()
}

// SetTitle sets the title bar text. Set to "" and call SetShowTitle(false) to hide.
func (cb *CodeBlock) SetTitle(title string) {
	cb.mu.Lock()
	cb.title = title
	if title != "" {
		cb.showTitle = true
	}
	cb.mu.Unlock()
}

// SetShowTitle toggles the title bar visibility.
func (cb *CodeBlock) SetShowTitle(show bool) {
	cb.mu.Lock()
	cb.showTitle = show
	cb.mu.Unlock()
}

// SetShowLineNumbers toggles line number rendering.
func (cb *CodeBlock) SetShowLineNumbers(show bool) {
	cb.mu.Lock()
	cb.showLineNumbers = show
	cb.mu.Unlock()
}

// SetHighlighter replaces the default chroma highlighter with a custom one.
func (cb *CodeBlock) SetHighlighter(h *markdown.Highlighter) {
	cb.mu.Lock()
	cb.highlighter = h
	cb.rehighlightLocked()
	cb.mu.Unlock()
}

// SetTheme sets the theme for title bar styling.
func (cb *CodeBlock) SetTheme(t *theme.Theme) {
	cb.mu.Lock()
	cb.currentTheme = t
	cb.mu.Unlock()
}

// ScrollUp moves the viewport up by n lines (clamped at 0).
func (cb *CodeBlock) ScrollUp(n int) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.scrollOffset -= n
	if cb.scrollOffset < 0 {
		cb.scrollOffset = 0
	}
}

// ScrollDown moves the viewport down by n lines (clamped at content bottom).
func (cb *CodeBlock) ScrollDown(n int) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	maxScroll := cb.maxScrollOffsetLocked()
	cb.scrollOffset += n
	if cb.scrollOffset > maxScroll {
		cb.scrollOffset = maxScroll
	}
}

// ScrollTo sets the absolute scroll offset (clamped).
func (cb *CodeBlock) ScrollTo(offset int) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	maxScroll := cb.maxScrollOffsetLocked()
	if offset < 0 {
		offset = 0
	}
	if offset > maxScroll {
		offset = maxScroll
	}
	cb.scrollOffset = offset
}

// ScrollOffset returns the current vertical scroll position.
func (cb *CodeBlock) ScrollOffset() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.scrollOffset
}

// LineCount returns the total number of code lines.
func (cb *CodeBlock) LineCount() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return len(cb.lines)
}

// Source returns the current raw source code.
func (cb *CodeBlock) Source() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.source
}

// Language returns the current language identifier.
func (cb *CodeBlock) Language() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.language
}

// --- Component interface ---

// Measure returns the natural size: width = longest line + gutter, height = line count (+ 1 for title).
func (cb *CodeBlock) Measure(cs Constraints) Size {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	maxLineW := 0
	for _, line := range cb.lines {
		w := cellWidth(line)
		if w > maxLineW {
			maxLineW = w
		}
	}

	width := maxLineW + cb.gutterWidthLocked()
	height := len(cb.lines)
	if cb.showTitle {
		height++
	}

	if cs.HasWidth() && width > cs.MaxWidth {
		width = cs.MaxWidth
	}
	if cs.HasHeight() && height > cs.MaxHeight {
		height = cs.MaxHeight
	}
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	return Size{W: width, H: height}
}

// SetBounds sets the component's allocated area and triggers re-highlight if needed.
func (cb *CodeBlock) SetBounds(r Rect) {
	cb.BaseComponent.SetBounds(r)
}

// Paint renders the code block into buf.
func (cb *CodeBlock) Paint(buf *buffer.Buffer) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	bounds := cb.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	row := bounds.Y
	maxRow := bounds.Y + bounds.H

	// Draw title bar
	if cb.showTitle && row < maxRow {
		cb.paintTitleLocked(buf, bounds, row)
		row++
	}

	// Draw code lines
	gutterW := cb.gutterWidthLocked()
	codeX := bounds.X + gutterW
	codeW := bounds.W - gutterW
	if codeW < 0 {
		codeW = 0
	}

	visibleLines := maxRow - row
	endIdx := cb.scrollOffset + visibleLines
	if endIdx > len(cb.lines) {
		endIdx = len(cb.lines)
	}

	for i := cb.scrollOffset; i < endIdx && row < maxRow; i++ {
		// Paint gutter (line number)
		if cb.showLineNumbers && gutterW > 0 {
			ln := fmt.Sprintf("%*d", gutterW-1, i+1)
			for j, r := range ln {
				if bounds.X+j < bounds.X+bounds.W {
					buf.SetCell(bounds.X+j, row, buffer.Cell{
						Rune: r,
						Fg:   cb.gutterColorLocked(),
					})
				}
			}
			// Separator
			if codeX-1 < bounds.X+bounds.W {
				buf.SetCell(bounds.X+gutterW-1, row, buffer.Cell{
					Rune: ' ',
				})
			}
		}

		// Paint highlighted code
		line := cb.lines[i]
		col := codeX
		for _, cell := range line {
			if col >= bounds.X+bounds.W {
				break
			}
			if col >= 0 && row >= 0 {
				buf.SetCell(col, row, cell)
			}
			if cell.Rune == '\t' {
				col += 4
			} else {
				col++
			}
		}

		// Fill remaining with blank background
		for c := col; c < bounds.X+bounds.W; c++ {
			buf.SetCell(c, row, buffer.Cell{Rune: ' '})
		}

		row++
	}

	// Fill remaining rows below content
	for ; row < maxRow; row++ {
		for c := bounds.X; c < bounds.X+bounds.W; c++ {
			buf.SetCell(c, row, buffer.Cell{Rune: ' '})
		}
	}
}

// --- internal helpers ---

func (cb *CodeBlock) rehighlight() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.rehighlightLocked()
}

func (cb *CodeBlock) rehighlightLocked() {
	if cb.highlighter == nil {
		cb.highlighter = markdown.NewHighlighter()
	}
	lines, err := cb.highlighter.Highlight(cb.source, cb.language)
	if err != nil || lines == nil {
		// Fallback: plain text, one cell per rune
		lines = cb.plainLinesLocked()
	}
	cb.lines = lines
}

func (cb *CodeBlock) plainLinesLocked() [][]buffer.Cell {
	rawLines := strings.Split(cb.source, "\n")
	result := make([][]buffer.Cell, len(rawLines))
	for i, line := range rawLines {
		cells := make([]buffer.Cell, 0, len(line))
		for _, r := range line {
			cells = append(cells, buffer.Cell{Rune: r})
		}
		result[i] = cells
	}
	return result
}

func (cb *CodeBlock) gutterWidthLocked() int {
	if !cb.showLineNumbers {
		return 0
	}
	// width of largest line number + separator space
	lineCount := len(cb.lines)
	digits := len(fmt.Sprintf("%d", lineCount))
	if digits < 2 {
		digits = 2
	}
	return digits + 1 // +1 for separator
}

func (cb *CodeBlock) maxScrollOffsetLocked() int {
	bounds := cb.Bounds()
	visibleH := bounds.H
	if cb.showTitle {
		visibleH--
	}
	if visibleH < 1 {
		visibleH = 1
	}
	max := len(cb.lines) - visibleH
	if max < 0 {
		max = 0
	}
	return max
}

func (cb *CodeBlock) gutterColorLocked() buffer.Color {
	if cb.currentTheme != nil {
		// Use theme muted/dim color if available
		return buffer.RGB(0x62, 0x72, 0xA4) // dracula comment color
	}
	return buffer.RGB(0x62, 0x72, 0xA4)
}

func (cb *CodeBlock) paintTitleLocked(buf *buffer.Buffer, bounds Rect, row int) {
	titleText := cb.title
	if titleText == "" {
		titleText = cb.language
	}
	if titleText == "" {
		titleText = "code"
	}

	label := fmt.Sprintf(" %s ", titleText)
	fg := buffer.RGB(0x8B, 0xE9, 0xFD)   // cyan
	bg := buffer.RGB(0x28, 0x2A, 0x36)   // dark bg

	// Paint title text
	col := bounds.X
	for _, r := range label {
		if col >= bounds.X+bounds.W {
			break
		}
		buf.SetCell(col, row, buffer.Cell{
			Rune: r,
			Fg:   fg,
			Bg:   bg,
		})
		col++
	}
	// Fill rest of title bar
	for ; col < bounds.X+bounds.W; col++ {
		buf.SetCell(col, row, buffer.Cell{
			Rune: ' ',
			Bg:   bg,
		})
	}
}

func cellWidth(line []buffer.Cell) int {
	w := 0
	for _, c := range line {
		if c.Rune == '\t' {
			w += 4
		} else {
			w++
		}
	}
	return w
}



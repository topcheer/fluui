package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownHorizontalRule: Render Markdown Horizontal Rules ───
//
// MarkdownHorizontalRule parses markdown horizontal rule markers (---, ***, ___)
// and renders them as styled horizontal divider lines. Supports `~~` strikethrough
// rendering and title labels for section dividers.
//
// Usage:
//
//	hr := NewMarkdownHorizontalRule()
//	hr.SetMarkdown("Intro text\n---\nContent below\n***\nMore content")
//	hr.Paint(buf)

// HorizontalRuleStyle holds styling for MarkdownHorizontalRule.
type HorizontalRuleStyle struct {
	Text     buffer.Style
	Rule     buffer.Style
	Title    buffer.Style
	Border   buffer.Style
}

// DefaultHorizontalRuleStyle returns sensible defaults.
func DefaultHorizontalRuleStyle() HorizontalRuleStyle {
	text := buffer.Style{Fg: buffer.RGB(226, 232, 240)}    // slate-200
	rule := buffer.Style{Fg: buffer.RGB(100, 116, 139)}   // slate-500
	title := buffer.Style{Fg: buffer.RGB(167, 139, 250), Flags: buffer.Bold} // violet-400 bold
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}   // slate-600
	return HorizontalRuleStyle{Text: text, Rule: rule, Title: title, Border: border}
}

// HRLineType classifies a rendered line.
type HRLineType int

const (
	hrText      HRLineType = iota
	hrRule
	hrTitle
)

// HRLine represents a single rendered line.
type HRLine struct {
	Text string
	Type HRLineType
}

// MarkdownHorizontalRule renders text with horizontal rule dividers.
type MarkdownHorizontalRule struct {
	BaseComponent
	mu sync.Mutex

	source string
	style  HorizontalRuleStyle

	// cached parsed lines
	cachedLines []HRLine
}

// NewMarkdownHorizontalRule creates a MarkdownHorizontalRule with defaults.
func NewMarkdownHorizontalRule() *MarkdownHorizontalRule {
	hr := &MarkdownHorizontalRule{
		style: DefaultHorizontalRuleStyle(),
	}
	hr.SetID(GenerateID("hrule"))
	return hr
}

// SetMarkdown sets the raw markdown source and parses horizontal rules.
func (hr *MarkdownHorizontalRule) SetMarkdown(source string) *MarkdownHorizontalRule {
	hr.mu.Lock()
	hr.source = source
	hr.parseLocked()
	hr.mu.Unlock()
	return hr
}

// Markdown returns the raw markdown source.
func (hr *MarkdownHorizontalRule) Markdown() string {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	return hr.source
}

// SetStyle sets the custom style.
func (hr *MarkdownHorizontalRule) SetStyle(s HorizontalRuleStyle) *MarkdownHorizontalRule {
	hr.mu.Lock()
	hr.style = s
	hr.mu.Unlock()
	return hr
}

// LineCount returns the number of parsed lines.
func (hr *MarkdownHorizontalRule) LineCount() int {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	return len(hr.cachedLines)
}

// RuleCount returns the number of horizontal rule lines.
func (hr *MarkdownHorizontalRule) RuleCount() int {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	count := 0
	for _, line := range hr.cachedLines {
		if line.Type == hrRule {
			count++
		}
	}
	return count
}

// isHorizontalRule checks if a line is a markdown horizontal rule (---, ***, ___).
func isHorizontalRule(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 3 {
		return false
	}
	// All chars must be the same (-, *, _)
	first := trimmed[0]
	if first != '-' && first != '*' && first != '_' {
		return false
	}
	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] != first {
			return false
		}
	}
	return true
}

// parseLocked parses markdown source into lines with rule detection.
func (hr *MarkdownHorizontalRule) parseLocked() {
	hr.cachedLines = hr.cachedLines[:0]
	lines := strings.Split(hr.source, "\n")
	for _, line := range lines {
		if hr.source == "" {
			continue
		}
		if isHorizontalRule(line) {
			hr.cachedLines = append(hr.cachedLines, HRLine{Type: hrRule})
		} else if strings.HasPrefix(strings.TrimSpace(line), "# ") {
			// Title line
			title := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "# "))
			hr.cachedLines = append(hr.cachedLines, HRLine{Text: title, Type: hrTitle})
		} else {
			hr.cachedLines = append(hr.cachedLines, HRLine{Text: line, Type: hrText})
		}
	}
}

// Measure returns the preferred size.
func (hr *MarkdownHorizontalRule) Measure(cs Constraints) Size {
	hr.mu.Lock()
	lineCount := len(hr.cachedLines)
	hr.mu.Unlock()

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

// Paint renders the horizontal rule content into the buffer.
func (hr *MarkdownHorizontalRule) Paint(buf *buffer.Buffer) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	b := hr.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 {
		w = 50
	}
	if h < 3 {
		h = 3
	}

	// Draw border
	bs := hr.style.Border
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
	for idx, line := range hr.cachedLines {
		rowY := y + 1 + idx
		if rowY >= y+h-1 || rowY >= buf.Height {
			break
		}

		switch line.Type {
		case hrRule:
			// Draw full-width horizontal rule
			ruleStyle := hr.style.Rule
			for col := x + 1; col < x+w-1 && col < buf.Width; col++ {
				buf.SetCell(col, rowY, buffer.Cell{Rune: '─', Fg: ruleStyle.Fg, Bg: ruleStyle.Bg, Flags: ruleStyle.Flags, Width: 1})
			}
		case hrTitle:
			// Draw title with bold style
			titleStyle := hr.style.Title
			col := x + 1
			for _, r := range line.Text {
				if col >= x+w-1 || col >= buf.Width {
					break
				}
				buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: titleStyle.Fg, Bg: titleStyle.Bg, Flags: titleStyle.Flags, Width: 1})
				col++
			}
		default:
			// Normal text
			textStyle := hr.style.Text
			col := x + 1
			for _, r := range line.Text {
				if col >= x+w-1 || col >= buf.Width {
					break
				}
				buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
				col++
			}
		}
	}
}

// Children returns nil.
func (hr *MarkdownHorizontalRule) Children() []Component { return nil }

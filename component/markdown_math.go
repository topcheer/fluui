package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownMath: Inline/Block Math Equation Display ───
//
// MarkdownMath renders LaTeX-style math expressions using Unicode math
// symbols. Supports both inline ($...$) and block ($$...$$) modes.
// Performs simple symbol substitution for common math operators.
//
// Usage:
//
//	mm := NewMarkdownMath()
//	mm.SetExpression("E = mc^2")
//	mm.SetBlock(true)
//	mm.Paint(buf)

// MarkdownMathStyle holds styling.
type MarkdownMathStyle struct {
	Text      buffer.Style
	Variable  buffer.Style
	Operator  buffer.Style
	Delimiter buffer.Style
}

// DefaultMarkdownMathStyle returns defaults.
func DefaultMarkdownMathStyle() MarkdownMathStyle {
	return MarkdownMathStyle{
		Text:      buffer.Style{Fg: buffer.RGB(226, 232, 240)},
		Variable:  buffer.Style{Fg: buffer.RGB(147, 197, 253), Flags: buffer.Italic},
		Operator:  buffer.Style{Fg: buffer.RGB(252, 165, 165)},
		Delimiter: buffer.Style{Fg: buffer.RGB(168, 85, 247)},
	}
}

var mathSymbolMap = map[rune]rune{
	'*': '×',
	'-': '−',
	'~': '∼',
	'>': '⟩',
	'<': '⟨',
	'.': '·',
}

// MarkdownMath renders a math expression.
type MarkdownMath struct {
	BaseComponent
	mu sync.Mutex

	expression string
	isBlock    bool
	style      MarkdownMathStyle
	// cached
	runes []rune
}

// NewMarkdownMath creates a MarkdownMath.
func NewMarkdownMath() *MarkdownMath {
	mm := &MarkdownMath{style: DefaultMarkdownMathStyle()}
	mm.SetID(GenerateID("math"))
	return mm
}

// SetExpression sets the math expression text.
func (mm *MarkdownMath) SetExpression(expr string) *MarkdownMath {
	mm.mu.Lock()
	mm.expression = expr
	mm.runes = mm.transformLocked([]rune(expr))
	mm.mu.Unlock()
	return mm
}

// SetBlock toggles block display mode (centered, with delimiters).
func (mm *MarkdownMath) SetBlock(block bool) *MarkdownMath {
	mm.mu.Lock()
	mm.isBlock = block
	mm.mu.Unlock()
	return mm
}

func (mm *MarkdownMath) transformLocked(input []rune) []rune {
	result := make([]rune, 0, len(input))
	for _, r := range input {
		if replacement, ok := mathSymbolMap[r]; ok {
			result = append(result, replacement)
		} else {
			result = append(result, r)
		}
	}
	return result
}

// Expression returns the original expression.
func (mm *MarkdownMath) Expression() string {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	return mm.expression
}

// SetStyle sets custom style.
func (mm *MarkdownMath) SetStyle(s MarkdownMathStyle) *MarkdownMath {
	mm.mu.Lock()
	mm.style = s
	mm.mu.Unlock()
	return mm
}

// Measure returns preferred size.
func (mm *MarkdownMath) Measure(cs Constraints) Size {
	mm.mu.Lock()
	w := len(mm.runes) + 4
	mm.mu.Unlock()
	if w < 4 {
		w = 4
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the math expression.
func (mm *MarkdownMath) Paint(buf *buffer.Buffer) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	b := mm.Bounds()
	x, y := b.X, b.Y

	textStyle := mm.style.Text
	delimStyle := mm.style.Delimiter

	col := x

	if mm.isBlock {
		// Block delimiters
		for _, r := range "⟦ " {
			if col >= buf.Width {
				return
			}
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: delimStyle.Fg, Bg: delimStyle.Bg, Flags: delimStyle.Flags, Width: 1})
			col++
		}
	}

	for _, r := range mm.runes {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
		col++
	}

	if mm.isBlock {
		for _, r := range " ⟧" {
			if col >= buf.Width {
				return
			}
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: delimStyle.Fg, Bg: delimStyle.Bg, Flags: delimStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (mm *MarkdownMath) Children() []Component { return nil }

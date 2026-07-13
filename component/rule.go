package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Rule: Horizontal/Vertical Separator Line ───
//
// Rule is a simple separator line widget (like HTML <hr>).
// Inspired by Textual's Rule and Ratatui's Block divider.
//
// Usage:
//	rule := NewRule()
//	rule.SetOrientation(HorizontalRule) // or VerticalRule
//	rule.SetChar('─')                   // custom line char
//	rule.SetStyle(buffer.Style{Fg: buffer.RGB(98,114,164)})

type RuleOrientation int

const (
	HorizontalRule RuleOrientation = iota
	VerticalRule
)

// Rule draws a horizontal or vertical separator line.
type Rule struct {
	mu          sync.RWMutex
	BaseComponent
	orientation RuleOrientation
	char        rune
	style       buffer.Style
}

// NewRule creates a horizontal rule with default style.
func NewRule() *Rule {
	return &Rule{
		orientation: HorizontalRule,
		char:        '─',
		style:       buffer.Style{Fg: buffer.RGB(98, 114, 164)}, // Dracula comment color
	}
}

// NewVerticalRule creates a vertical rule.
func NewVerticalRule() *Rule {
	r := NewRule()
	r.orientation = VerticalRule
	r.char = '│'
	return r
}

// SetOrientation sets horizontal or vertical.
func (r *Rule) SetOrientation(o RuleOrientation) {
	r.mu.Lock()
	r.orientation = o
	r.mu.Unlock()
}

// SetChar sets the line character.
func (r *Rule) SetChar(c rune) {
	r.mu.Lock()
	r.char = c
	r.mu.Unlock()
}

// SetStyle sets the line color/style.
func (r *Rule) SetStyle(s buffer.Style) {
	r.mu.Lock()
	r.style = s
	r.mu.Unlock()
}

func (r *Rule) Measure(constraints Constraints) Size {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.orientation == HorizontalRule {
		return Size{W: constraints.MaxWidth, H: 1}
	}
	return Size{W: 1, H: constraints.MaxHeight}
}

func (r *Rule) Paint(buf *buffer.Buffer) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bounds := r.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	cell := buffer.Cell{
		Rune:  r.char,
		Width: 1,
		Fg:    r.style.Fg,
		Bg:    r.style.Bg,
		Flags: r.style.Flags,
	}

	if r.orientation == HorizontalRule {
		for x := 0; x < bounds.W; x++ {
			buf.SetCell(bounds.X+x, bounds.Y, cell)
		}
	} else {
		for y := 0; y < bounds.H; y++ {
			buf.SetCell(bounds.X, bounds.Y+y, cell)
		}
	}
}
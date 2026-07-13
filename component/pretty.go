package component

import (
	"fmt"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// PrettyStyle holds visual styling for the Pretty component.
type PrettyStyle struct {
	KeyFg       buffer.Color
	StringFg    buffer.Color
	NumberFg    buffer.Color
	BoolFg      buffer.Color
	NilFg       buffer.Color
	PunctuationFg buffer.Color
}

// DefaultPrettyStyle returns a Dracula-themed pretty-print style.
func DefaultPrettyStyle() PrettyStyle {
	return PrettyStyle{
		KeyFg:       buffer.NamedColor(buffer.NamedCyan),
		StringFg:    buffer.NamedColor(buffer.NamedGreen),
		NumberFg:    buffer.NamedColor(buffer.NamedMagenta),
		BoolFg:      buffer.NamedColor(buffer.NamedYellow),
		NilFg:       buffer.NamedColor(buffer.NamedBrightBlack),
		PunctuationFg: buffer.NamedColor(buffer.NamedBrightBlack),
	}
}

// Pretty displays a pretty-formatted Go data structure (like fmt.Sprintf("%#v", v))
// with syntax highlighting. Similar to Textual's Pretty widget.
type Pretty struct {
	BaseComponent

	data  string
	style PrettyStyle

	mu sync.RWMutex
}

// NewPretty creates a pretty display from any Go value.
func NewPretty(v any) *Pretty {
	p := &Pretty{
		data:  prettyFormat(v),
		style: DefaultPrettyStyle(),
	}
	return p
}

// NewPrettyString creates a pretty display from a pre-formatted string.
func NewPrettyString(s string) *Pretty {
	return &Pretty{
		data:  s,
		style: DefaultPrettyStyle(),
	}
}

// SetStyle sets the visual style.
func (p *Pretty) SetStyle(s PrettyStyle) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.style = s
}

// Measure returns the desired size based on content lines.
func (p *Pretty) Measure(cs Constraints) Size {
	p.mu.RLock()
	defer p.mu.RUnlock()

	w := cs.MaxWidth
	if w <= 0 {
		w = 40
	}
	lines := strings.Count(p.data, "\n") + 1
	if lines < 1 {
		lines = 1
	}
	return Size{W: w, H: lines}
}

// Paint renders the pretty-printed data into the buffer.
func (p *Pretty) Paint(buf *buffer.Buffer) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if buf == nil {
		return
	}

	bounds := p.Bounds()
	x, y := bounds.X, bounds.Y
	w := bounds.W
	if w <= 0 {
		return
	}

	lines := strings.Split(p.data, "\n")
	for row, line := range lines {
		if row >= bounds.H {
			break
		}
		cx := x
		col := p.style.PunctuationFg
		inString := false
		for _, r := range line {
			if cx >= x+w {
				break
			}
			// Simple syntax highlighting
			if r == '"' {
				inString = !inString
				col = p.style.StringFg
			} else if inString {
				col = p.style.StringFg
			} else if r >= '0' && r <= '9' {
				col = p.style.NumberFg
			} else if r == '{' || r == '}' || r == '[' || r == ']' || r == ':' || r == ',' || r == ' ' {
				col = p.style.PunctuationFg
			} else if r == 't' || r == 'f' {
				col = p.style.BoolFg
			} else if r == 'n' {
				col = p.style.NilFg
			} else {
				col = p.style.KeyFg
			}
			buf.SetCell(cx, y+row, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    col,
			})
			cx++
		}
	}
}

// HandleKey is a no-op for Pretty.
func (p *Pretty) HandleKey(_ *term.KeyEvent) bool { return false }

// Children returns nil.
func (p *Pretty) Children() []Component { return nil }

// prettyFormat converts a Go value to a pretty-printed string.
func prettyFormat(v any) string {
	if v == nil {
		return "nil"
	}
	s := fmt.Sprintf("%#v", v)
	// Indent for readability
	s = strings.ReplaceAll(s, ", ", ",\n  ")
	s = strings.ReplaceAll(s, "{", "{\n  ")
	s = strings.ReplaceAll(s, "}", "\n}")
	return s
}

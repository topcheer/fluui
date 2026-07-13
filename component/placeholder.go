package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// PlaceholderStyle holds visual styling for the Placeholder component.
type PlaceholderStyle struct {
	Fg    buffer.Color
	Bg    buffer.Color
	BorderFg buffer.Color
}

// DefaultPlaceholderStyle returns a Dracula-themed placeholder style.
func DefaultPlaceholderStyle() PlaceholderStyle {
	return PlaceholderStyle{
		Fg:       buffer.NamedColor(buffer.NamedBrightBlack),
		Bg:       buffer.Color{},
		BorderFg: buffer.NamedColor(buffer.NamedBrightBlack),
	}
}

// Placeholder displays a centered label in a bordered box.
// It is used during UI development to reserve space before the real
// component is ready, similar to Textual's Placeholder widget.
type Placeholder struct {
	BaseComponent

	label string
	style PlaceholderStyle

	mu sync.RWMutex
}

// NewPlaceholder creates a placeholder with the given label.
func NewPlaceholder(label string) *Placeholder {
	return &Placeholder{
		label: label,
		style: DefaultPlaceholderStyle(),
	}
}

// SetLabel sets the placeholder label.
func (p *Placeholder) SetLabel(s string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.label = s
}

// Label returns the current label.
func (p *Placeholder) Label() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.label
}

// SetStyle sets the visual style.
func (p *Placeholder) SetStyle(s PlaceholderStyle) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.style = s
}

// Measure returns the desired size based on label length.
func (p *Placeholder) Measure(cs Constraints) Size {
	p.mu.RLock()
	defer p.mu.RUnlock()

	w := len([]rune(p.label)) + 4 // border + padding
	if w < 10 {
		w = 10
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 3} // border + 1 content line
}

// Paint renders the placeholder into the buffer.
func (p *Placeholder) Paint(buf *buffer.Buffer) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if buf == nil {
		return
	}

	bounds := p.Bounds()
	x, y := bounds.X, bounds.Y
	w, h := bounds.W, bounds.H
	if w <= 0 || h <= 0 {
		return
	}

	// Draw border (single line)
	if h >= 1 {
		// Top border
		for i := 0; i < w; i++ {
			r := '─'
			if i == 0 {
				r = '┌'
			} else if i == w-1 {
				r = '┐'
			}
			buf.SetCell(x+i, y, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    p.style.BorderFg,
			})
		}
		// Bottom border
		if h > 1 {
			for i := 0; i < w; i++ {
				r := '─'
				if i == 0 {
					r = '└'
				} else if i == w-1 {
					r = '┘'
				}
				buf.SetCell(x+i, y+h-1, buffer.Cell{
					Rune:  r,
					Width: 1,
					Fg:    p.style.BorderFg,
				})
			}
		}
		// Side borders
		for j := 1; j < h-1; j++ {
			buf.SetCell(x, y+j, buffer.Cell{
				Rune:  '│',
				Width: 1,
				Fg:    p.style.BorderFg,
			})
			buf.SetCell(x+w-1, y+j, buffer.Cell{
				Rune:  '│',
				Width: 1,
				Fg:    p.style.BorderFg,
			})
		}
	}

	// Center label on the middle line
	if h >= 2 && w > 2 {
		labelRunes := []rune(p.label)
		labelLen := len(labelRunes)
		innerW := w - 2
		if labelLen > innerW {
			labelRunes = labelRunes[:innerW]
			labelLen = innerW
		}
		startX := x + 1 + (innerW-labelLen)/2
		midY := y + h/2
		for i, r := range labelRunes {
			buf.SetCell(startX+i, midY, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    p.style.Fg,
			})
		}
	}
}

// HandleKey is a no-op for Placeholder.
func (p *Placeholder) HandleKey(_ *term.KeyEvent) bool { return false }

// Children returns nil.
func (p *Placeholder) Children() []Component { return nil }

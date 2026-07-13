package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// FooterStyle holds visual styling for the Footer component.
type FooterStyle struct {
	KeyFg       buffer.Color
	DescriptionFg buffer.Color
	Separator   string
	Bg          buffer.Color
}

// DefaultFooterStyle returns a Dracula-themed footer style.
func DefaultFooterStyle() FooterStyle {
	return FooterStyle{
		KeyFg:         buffer.NamedColor(buffer.NamedYellow),
		DescriptionFg: buffer.NamedColor(buffer.NamedWhite),
		Separator:     " ",
		Bg:            buffer.NamedColor(buffer.NamedBrightBlack),
	}
}

// FooterHint represents a single keybinding hint displayed in the footer.
type FooterHint struct {
	Keys        string // e.g., "Ctrl+S"
	Description string // e.g., "Save"
}

// Footer displays keybinding hints at the bottom of the screen,
// similar to Textual's Footer widget. It shows key shortcuts and their
// descriptions in a compact format.
type Footer struct {
	BaseComponent

	hints []FooterHint
	style FooterStyle

	mu sync.RWMutex
}

// NewFooter creates an empty footer.
func NewFooter() *Footer {
	return &Footer{
		style: DefaultFooterStyle(),
	}
}

// SetHints sets the keybinding hints to display.
func (f *Footer) SetHints(hints []FooterHint) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.hints = hints
}

// Hints returns a copy of the current hints.
func (f *Footer) Hints() []FooterHint {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make([]FooterHint, len(f.hints))
	copy(out, f.hints)
	return out
}

// AddHint adds a single hint.
func (f *Footer) AddHint(keys, desc string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.hints = append(f.hints, FooterHint{Keys: keys, Description: desc})
}

// ClearHints removes all hints.
func (f *Footer) ClearHints() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.hints = nil
}

// SetStyle sets the visual style.
func (f *Footer) SetStyle(s FooterStyle) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.style = s
}

// Measure returns the desired size (always 1 row high).
func (f *Footer) Measure(cs Constraints) Size {
	w := cs.MaxWidth
	if w <= 0 {
		w = 40
	}
	return Size{W: w, H: 1}
}

// Paint renders the footer into the buffer.
func (f *Footer) Paint(buf *buffer.Buffer) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if buf == nil {
		return
	}

	bounds := f.Bounds()
	x, y := bounds.X, bounds.Y
	w := bounds.W
	if w <= 0 {
		return
	}

	// Fill background
	for i := 0; i < w; i++ {
		buf.SetCell(x+i, y, buffer.Cell{
			Rune:  ' ',
			Width: 1,
			Fg:    f.style.DescriptionFg,
			Bg:    f.style.Bg,
		})
	}

	// Draw hints: "key desc  key desc  key desc"
	cx := x
	for i, hint := range f.hints {
		if i > 0 {
			// Add separator (2 spaces)
			if cx+2 > x+w {
				break
			}
			cx += 2
		}

		// Draw key in KeyFg
		keyRunes := []rune(hint.Keys)
		for _, r := range keyRunes {
			if cx >= x+w {
				return
			}
			buf.SetCell(cx, y, buffer.Cell{
				Rune:   r,
				Width:  1,
				Fg:     f.style.KeyFg,
				Bg:     f.style.Bg,
				Flags:  buffer.Bold,
			})
			cx++
		}

		// Space between key and description
		if cx < x+w && hint.Description != "" {
			buf.SetCell(cx, y, buffer.Cell{
				Rune:  ' ',
				Width: 1,
				Fg:    f.style.DescriptionFg,
				Bg:    f.style.Bg,
			})
			cx++
		}

		// Draw description
		descRunes := []rune(hint.Description)
		for _, r := range descRunes {
			if cx >= x+w {
				return
			}
			buf.SetCell(cx, y, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    f.style.DescriptionFg,
				Bg:    f.style.Bg,
			})
			cx++
		}
	}
}

// HandleKey is a no-op for Footer.
func (f *Footer) HandleKey(_ *term.KeyEvent) bool { return false }

// Children returns nil.
func (f *Footer) Children() []Component { return nil }

package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// HeaderStyle holds visual styling for the Header component.
type HeaderStyle struct {
	TitleFg    buffer.Color
	SubtitleFg buffer.Color
	Bg         buffer.Color
	BorderFg   buffer.Color
}

// DefaultHeaderStyle returns a Dracula-themed header style.
func DefaultHeaderStyle() HeaderStyle {
	return HeaderStyle{
		TitleFg:    buffer.NamedColor(buffer.NamedCyan),
		SubtitleFg: buffer.NamedColor(buffer.NamedWhite),
		Bg:         buffer.NamedColor(buffer.NamedBrightBlack),
		BorderFg:   buffer.NamedColor(buffer.NamedBrightBlack),
	}
}

// Header displays an application title and optional subtitle at the top
// of the screen, similar to Textual's Header widget.
type Header struct {
	BaseComponent

	title    string
	subtitle string
	style    HeaderStyle

	mu sync.RWMutex
}

// NewHeader creates a header with the given title.
func NewHeader(title string) *Header {
	return &Header{
		title: title,
		style: DefaultHeaderStyle(),
	}
}

// SetTitle sets the header title.
func (h *Header) SetTitle(s string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.title = s
}

// Title returns the current title.
func (h *Header) Title() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.title
}

// SetSubtitle sets the header subtitle (shown right-aligned).
func (h *Header) SetSubtitle(s string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.subtitle = s
}

// Subtitle returns the current subtitle.
func (h *Header) Subtitle() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.subtitle
}

// SetStyle sets the visual style.
func (h *Header) SetStyle(s HeaderStyle) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.style = s
}

// Measure returns the desired size (always 1 row high).
func (h *Header) Measure(cs Constraints) Size {
	w := cs.MaxWidth
	if w <= 0 {
		w = 40
	}
	return Size{W: w, H: 1}
}

// Paint renders the header into the buffer.
func (h *Header) Paint(buf *buffer.Buffer) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if buf == nil {
		return
	}

	bounds := h.Bounds()
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
			Fg:    h.style.TitleFg,
			Bg:    h.style.Bg,
		})
	}

	// Draw title on left
	titleRunes := []rune(h.title)
	if len(titleRunes) > w {
		titleRunes = titleRunes[:w]
	}
	for i, r := range titleRunes {
		buf.SetCell(x+i, y, buffer.Cell{
			Rune:   r,
			Width:  1,
			Fg:     h.style.TitleFg,
			Bg:     h.style.Bg,
			Flags:  buffer.Bold,
		})
	}

	// Draw subtitle right-aligned
	if h.subtitle != "" {
		subRunes := []rune(h.subtitle)
		subLen := len(subRunes)
		startX := x + w - subLen
		if startX < x+len(titleRunes)+1 {
			startX = x + len(titleRunes) + 1
		}
		if startX+subLen > x+w {
			subLen = x + w - startX
			if subLen <= 0 {
				return
			}
			subRunes = subRunes[:subLen]
		}
		for i, r := range subRunes {
			if startX+i < x+w {
				buf.SetCell(startX+i, y, buffer.Cell{
					Rune:  r,
					Width: 1,
					Fg:    h.style.SubtitleFg,
					Bg:    h.style.Bg,
				})
			}
		}
	}
}

// HandleKey is a no-op for Header.
func (h *Header) HandleKey(_ *term.KeyEvent) bool { return false }

// Children returns nil.
func (h *Header) Children() []Component { return nil }

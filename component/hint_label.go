package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// HintLabel renders a small italic help text with padding.
// Useful for displaying contextual hints inline.
//
// Thread-safe.
type HintLabel struct {
	BaseComponent
	mu       sync.RWMutex
	text     string
	customFg buffer.Color
	customBg buffer.Color
}

// NewHintLabel creates a hint label with the given text.
func NewHintLabel(text string) *HintLabel {
	return &HintLabel{
		BaseComponent: BaseComponent{id: GenerateID("hint")},
		text:          text,
	}
}

// Text returns the hint text.
func (h *HintLabel) Text() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.text
}

// SetText updates the hint text.
func (h *HintLabel) SetText(s string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.text = s
}

// SetColors overrides foreground and background. Pass zero values for auto.
func (h *HintLabel) SetColors(fg, bg buffer.Color) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.customFg = fg
	h.customBg = bg
}

// Measure returns preferred size: text width + 2 (padding).
func (h *HintLabel) Measure(cs Constraints) Size {
	h.mu.RLock()
	defer h.mu.RUnlock()
	w := buffer.StringWidth(h.text) + 2
	hh := 1
	if cs.HasWidth() && w > cs.MaxWidth { w = cs.MaxWidth }
	if w < 1 { w = 1 }
	if hh < 1 { hh = 1 }
	return Size{W: w, H: hh}
}

// Paint draws the hint label. Zero allocations.
func (h *HintLabel) Paint(buf *buffer.Buffer) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	bounds := h.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	tt := theme.Get()
	fg := h.customFg
	if fg.Type == buffer.ColorNone { fg = tt.Fg }
	bg := h.customBg
	if bg.Type == buffer.ColorNone { bg = tt.Muted }

	style := buffer.Style{Fg: fg, Bg: bg, Flags: buffer.Italic}
	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	if x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: fg, Bg: bg})
		x++
	}

	msgW := buffer.StringWidth(h.text)
	avail := maxX - x - 1
	if avail < 0 { avail = 0 }
	if msgW > avail {
		drawn := 0
		for _, r := range h.text {
			if drawn >= avail-1 && avail > 0 {
				buf.SetCell(x, y, buffer.Cell{Rune: '\u2026', Width: 1, Fg: fg, Bg: bg, Flags: buffer.Italic})
				x++
				break
			}
			rw := buffer.RuneWidth(r)
			if x+rw > maxX-1 { break }
			buf.SetCell(x, y, buffer.Cell{Rune: r, Width: uint8(rw), Fg: fg, Bg: bg, Flags: buffer.Italic})
			if rw == 2 && x+1 < maxX {
				buf.SetCell(x+1, y, buffer.Cell{Rune: 0, Width: 0, Bg: bg})
			}
			x += rw
			drawn++
		}
	} else {
		x = buf.DrawText(x, y, h.text, style)
	}
	for x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: fg, Bg: bg})
		x++
	}
}

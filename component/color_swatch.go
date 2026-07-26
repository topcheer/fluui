package component

import (
	"strconv"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// ColorSwatch renders a color sample with optional hex label.
// Useful for theme pickers, color palettes, and AI-generated color outputs.
//
// Thread-safe.
type ColorSwatch struct {
	BaseComponent
	mu       sync.RWMutex
	color    buffer.Color
	label    string // hex label, empty = auto from color
	showHex  bool
}

// NewColorSwatch creates a swatch displaying the given color.
func NewColorSwatch(c buffer.Color) *ColorSwatch {
	return &ColorSwatch{
		BaseComponent: BaseComponent{id: GenerateID("swatch")},
		color:         c,
		showHex:       true,
	}
}

// Color returns the swatch color.
func (s *ColorSwatch) Color() buffer.Color {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.color
}

// SetColor updates the swatch color.
func (s *ColorSwatch) SetColor(c buffer.Color) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.color = c
}

// Label returns the custom label (empty = auto hex).
func (s *ColorSwatch) Label() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.label
}

// SetLabel overrides the display label. Pass "" for auto hex.
func (s *ColorSwatch) SetLabel(l string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.label = l
}

// ShowHex returns whether the hex code is displayed.
func (s *ColorSwatch) ShowHex() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.showHex
}

// SetShowHex toggles hex label display.
func (s *ColorSwatch) SetShowHex(b bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.showHex = b
}

// hexStringLocked returns the hex representation of the color (zero alloc via stack buffer).
func (s *ColorSwatch) hexStringLocked() string {
	c := s.color
	switch c.Type {
	case buffer.ColorTrue:
		r := uint8(c.Val >> 16)
		g := uint8(c.Val >> 8)
		b := uint8(c.Val)
		var buf [7]byte
		out := buf[:0]
		out = append(out, '#')
		const hexChars = "0123456789abcdef"
		out = append(out, hexChars[r>>4], hexChars[r&0xf])
		out = append(out, hexChars[g>>4], hexChars[g&0xf])
		out = append(out, hexChars[b>>4], hexChars[b&0xf])
		return string(out)
	case buffer.Color256:
		return "c" + strconv.Itoa(int(c.Val))
	case buffer.ColorNamed:
		return "n" + strconv.Itoa(int(c.Val))
	default:
		return "none"
	}
}

// Measure returns the preferred size: swatch(2) + space + label width.
func (s *ColorSwatch) Measure(cs Constraints) Size {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w := 3 // "██ " (2-cell swatch + space)
	if s.showHex {
		if s.label != "" {
			w += len(s.label)
		} else {
			w += 7 // "#RRGGBB"
		}
	}
	h := 1

	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return Size{W: w, H: h}
}

// Paint draws the swatch. Zero allocations.
func (s *ColorSwatch) Paint(buf *buffer.Buffer) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bounds := s.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	t := theme.Get()
	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	// Draw 2-cell color block using full-block char █
	swStyle := buffer.Style{Fg: s.color, Bg: t.Bg}
	for i := 0; i < 2 && x < maxX; i++ {
		buf.SetCell(x, y, buffer.Cell{Rune: '\u2588', Width: 1, Fg: swStyle.Fg, Bg: swStyle.Bg})
		x++
	}

	// Space separator
	if x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1})
		x++
	}

	// Hex label
	if s.showHex && x < maxX {
		labelStyle := buffer.Style{Fg: t.Muted}
		if s.label != "" {
			x = buf.DrawText(x, y, s.label, labelStyle)
		} else {
			// Draw hex directly via stack buffer + DrawBytes (zero alloc)
			s.drawHexLocked(buf, x, y, maxX, labelStyle)
		}
	}
}

// drawHexLocked draws the hex representation of the color directly to buffer.
func (s *ColorSwatch) drawHexLocked(buf *buffer.Buffer, x, y, maxX int, style buffer.Style) {
	c := s.color
	const hexChars = "0123456789abcdef"
	switch c.Type {
	case buffer.ColorTrue:
		r := uint8(c.Val >> 16)
		g := uint8(c.Val >> 8)
		b := uint8(c.Val)
		var hb [7]byte
		h := hb[:0]
		h = append(h, '#')
		h = append(h, hexChars[r>>4], hexChars[r&0xf])
		h = append(h, hexChars[g>>4], hexChars[g&0xf])
		h = append(h, hexChars[b>>4], hexChars[b&0xf])
		buf.DrawBytes(x, y, h, style)
	case buffer.Color256:
		var hb [8]byte
		h := hb[:0]
		h = append(h, 'c')
		h = strconv.AppendInt(h, int64(c.Val), 10)
		buf.DrawBytes(x, y, h, style)
	case buffer.ColorNamed:
		var hb [8]byte
		h := hb[:0]
		h = append(h, 'n')
		h = strconv.AppendInt(h, int64(c.Val), 10)
		buf.DrawBytes(x, y, h, style)
	default:
		var hb [4]byte
		h := hb[:0]
		h = append(h, 'n', 'o', 'n', 'e')
		buf.DrawBytes(x, y, h, style)
	}
}

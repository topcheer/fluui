package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ColorWheel: HSL Color Wheel Selector ───
//
// ColorWheel renders a compact color wheel showing hue positions.
// The current hue is highlighted. Useful for color pickers and
// theme customization panels.
//
// Usage:
//
//	cw := NewColorWheel()
//	cw.SetHue(180) // cyan
//	cw.Paint(buf)

// ColorWheelStyle holds styling.
type ColorWheelStyle struct {
	Selected buffer.Style
	Marker   buffer.Style
	Label    buffer.Style
}

// DefaultColorWheelStyle returns defaults.
func DefaultColorWheelStyle() ColorWheelStyle {
	return ColorWheelStyle{
		Selected: buffer.Style{Fg: buffer.RGB(251, 191, 36), Flags: buffer.Bold},
		Marker:   buffer.Style{Fg: buffer.RGB(71, 85, 105)},
		Label:    buffer.Style{Fg: buffer.RGB(148, 163, 184)},
	}
}

// 12 hue positions mapped to Unicode block characters and color names.
var hueColors = [12]uint32{
	0xFF0000, 0xFF8000, 0xFFFF00, 0x80FF00,
	0x00FF00, 0x00FF80, 0x00FFFF, 0x0080FF,
	0x0000FF, 0x8000FF, 0xFF00FF, 0xFF0080,
}

var hueLabels = [12]string{"Rd", "Or", "Yl", "LG", "Gr", "CG", "Cy", "AB", "Bu", "Vi", "Mg", "Pk"}

// ColorWheel renders a compact hue color wheel.
type ColorWheel struct {
	BaseComponent
	mu sync.Mutex

	hue   int // 0-359
	style ColorWheelStyle
	// cached
	selIdx int
	hueStr string
}

// NewColorWheel creates a ColorWheel.
func NewColorWheel() *ColorWheel {
	cw := &ColorWheel{hue: 0, style: DefaultColorWheelStyle()}
	cw.SetID(GenerateID("colorwheel"))
	cw.recomputeLocked()
	return cw
}

// SetHue sets the current hue (0-359).
func (cw *ColorWheel) SetHue(h int) *ColorWheel {
	cw.mu.Lock()
	if h < 0 {
		h += 360 * ((-h / 360) + 1)
	}
	cw.hue = h % 360
	cw.recomputeLocked()
	cw.mu.Unlock()
	return cw
}

func (cw *ColorWheel) recomputeLocked() {
	cw.selIdx = cw.hue * 12 / 360
	if cw.selIdx > 11 {
		cw.selIdx = 11
	}
	cw.hueStr = itoa(cw.hue) + "°"
}

// Hue returns the current hue.
func (cw *ColorWheel) Hue() int {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	return cw.hue
}

// SetStyle sets custom style.
func (cw *ColorWheel) SetStyle(s ColorWheelStyle) *ColorWheel {
	cw.mu.Lock()
	cw.style = s
	cw.mu.Unlock()
	return cw
}

// Measure returns preferred size.
func (cw *ColorWheel) Measure(cs Constraints) Size {
	w := 16
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the color wheel.
func (cw *ColorWheel) Paint(buf *buffer.Buffer) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	b := cw.Bounds()
	x, y := b.X, b.Y

	markerStyle := cw.style.Marker
	selStyle := cw.style.Selected
	labelStyle := cw.style.Label

	col := x
	for i := 0; i < 12; i++ {
		if col >= buf.Width {
			break
		}
		var st buffer.Style
		if i == cw.selIdx {
			st = selStyle
		} else {
			r := uint8(hueColors[i] >> 16)
			g := uint8(hueColors[i] >> 8)
			b2 := uint8(hueColors[i])
			st = buffer.Style{Fg: buffer.RGB(r, g, b2)}
		}
		var r rune
		if i == cw.selIdx {
			r = '◉'
		} else {
			r = '●'
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		col++
	}

	// Hue value label
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, r := range hueLabels[cw.selIdx] {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: selStyle.Fg, Bg: selStyle.Bg, Flags: selStyle.Flags, Width: 1})
		col++
	}
	_ = markerStyle
}

// Children returns nil.
func (cw *ColorWheel) Children() []Component { return nil }

package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── CompassDial: Compass Dial Indicator ───
//
// CompassDial renders a compass rose showing a heading direction.
// The needle points to the current heading with N/E/S/W markers.
// Useful for navigation and orientation widgets.
//
// Usage:
//
//	cd := NewCompassDial()
//	cd.SetHeading(90) // East
//	cd.Paint(buf)

// CompassDialStyle holds styling.
type CompassDialStyle struct {
	Needle   buffer.Style
	Cardinal buffer.Style
	Hub      buffer.Style
}

// DefaultCompassDialStyle returns defaults.
func DefaultCompassDialStyle() CompassDialStyle {
	return CompassDialStyle{
		Needle:   buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Cardinal: buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Hub:      buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
	}
}

var compassNeedleChars = [8]rune{'↑', '↗', '→', '↘', '↓', '↙', '←', '↖'}

// CompassDial renders a compact compass indicator.
type CompassDial struct {
	BaseComponent
	mu sync.Mutex

	heading int // 0-359
	style   CompassDialStyle
	// cached
	needleIdx  int
	headingStr string
}

// NewCompassDial creates a CompassDial.
func NewCompassDial() *CompassDial {
	cd := &CompassDial{style: DefaultCompassDialStyle()}
	cd.SetID(GenerateID("compass"))
	cd.recomputeLocked()
	return cd
}

// SetHeading sets the heading in degrees (0-359, 0=North).
func (cd *CompassDial) SetHeading(h int) *CompassDial {
	cd.mu.Lock()
	for h < 0 {
		h += 360
	}
	cd.heading = h % 360
	cd.recomputeLocked()
	cd.mu.Unlock()
	return cd
}

func (cd *CompassDial) recomputeLocked() {
	cd.needleIdx = cd.heading * 8 / 360
	if cd.needleIdx > 7 {
		cd.needleIdx = 7
	}
	cd.headingStr = itoa(cd.heading) + "°"
}

// Heading returns the current heading.
func (cd *CompassDial) Heading() int {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	return cd.heading
}

// SetStyle sets custom style.
func (cd *CompassDial) SetStyle(s CompassDialStyle) *CompassDial {
	cd.mu.Lock()
	cd.style = s
	cd.mu.Unlock()
	return cd
}

// Measure returns preferred size.
func (cd *CompassDial) Measure(cs Constraints) Size {
	w := 10
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the compass dial.
func (cd *CompassDial) Paint(buf *buffer.Buffer) {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	b := cd.Bounds()
	x, y := b.X, b.Y

	cardinalStyle := cd.style.Cardinal
	needleStyle := cd.style.Needle
	hubStyle := cd.style.Hub

	col := x

	// Cardinal markers: N E S W
	for _, r := range "N·E·S·W" {
		if col >= buf.Width {
			break
		}
		var st buffer.Style
		if r == 'N' || r == 'E' || r == 'S' || r == 'W' {
			st = hubStyle
		} else {
			st = cardinalStyle
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		col++
	}

	// Needle
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: cardinalStyle.Fg, Bg: cardinalStyle.Bg, Flags: cardinalStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: compassNeedleChars[cd.needleIdx], Fg: needleStyle.Fg, Bg: needleStyle.Bg, Flags: needleStyle.Flags, Width: 1})
		col++
	}

	// Heading value
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: cardinalStyle.Fg, Bg: cardinalStyle.Bg, Flags: cardinalStyle.Flags, Width: 1})
		col++
	}
	for _, r := range cd.headingStr {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: hubStyle.Fg, Bg: hubStyle.Bg, Flags: hubStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (cd *CompassDial) Children() []Component { return nil }

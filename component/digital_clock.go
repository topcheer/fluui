package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── DigitalClock: 7-Segment Digital Clock Display ───
//
// DigitalClock renders a time display using large block characters
// in a 7-segment style. Shows hours:minutes:seconds with optional
// AM/PM indicator.
//
// Usage:
//
//	dc := NewDigitalClock()
//	dc.SetTime(14, 30, 45) // 14:30:45
//	dc.SetFormat24h(true)
//	dc.Paint(buf)

// DigitalClockStyle holds styling.
type DigitalClockStyle struct {
	Digit  buffer.Style
	Colon  buffer.Style
	Label  buffer.Style
	Suffix buffer.Style
}

// DefaultDigitalClockStyle returns defaults.
func DefaultDigitalClockStyle() DigitalClockStyle {
	return DigitalClockStyle{
		Digit:  buffer.Style{Fg: buffer.RGB(34, 197, 246), Flags: buffer.Bold},
		Colon:  buffer.Style{Fg: buffer.RGB(251, 191, 36)},
		Label:  buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Suffix: buffer.Style{Fg: buffer.RGB(148, 163, 184)},
	}
}

// DigitalClock renders a digital clock display.
type DigitalClock struct {
	BaseComponent
	mu sync.Mutex

	hour   int
	minute int
	second int
	format24 bool
	style  DigitalClockStyle
	// cached
	timeStr   string
	suffixStr string
}

// NewDigitalClock creates a DigitalClock.
func NewDigitalClock() *DigitalClock {
	dc := &DigitalClock{format24: true, style: DefaultDigitalClockStyle()}
	dc.SetID(GenerateID("digiclock"))
	dc.recomputeLocked()
	return dc
}

// SetTime sets the time components.
func (dc *DigitalClock) SetTime(h, m, s int) *DigitalClock {
	dc.mu.Lock()
	if h < 0 { h = 0 }
	if h > 23 { h = 23 }
	if m < 0 { m = 0 }
	if m > 59 { m = 59 }
	if s < 0 { s = 0 }
	if s > 59 { s = 59 }
	dc.hour = h
	dc.minute = m
	dc.second = s
	dc.recomputeLocked()
	dc.mu.Unlock()
	return dc
}

// SetFormat24h toggles 24-hour format.
func (dc *DigitalClock) SetFormat24h(f bool) *DigitalClock {
	dc.mu.Lock()
	dc.format24 = f
	dc.recomputeLocked()
	dc.mu.Unlock()
	return dc
}

func (dc *DigitalClock) recomputeLocked() {
	if dc.format24 {
		dc.timeStr = pad2(dc.hour) + ":" + pad2(dc.minute) + ":" + pad2(dc.second)
		dc.suffixStr = ""
	} else {
		h := dc.hour
		suf := "AM"
		if h == 0 {
			h = 12
		} else if h >= 12 {
			suf = "PM"
			if h > 12 { h -= 12 }
		}
		dc.timeStr = pad2(h) + ":" + pad2(dc.minute) + ":" + pad2(dc.second)
		dc.suffixStr = suf
	}
}

// pad2 formats a number as 2-digit with leading zero.
func pad2(n int) string {
	if n < 10 {
		return "0" + itoa(n)
	}
	return itoa(n)
}

// Hour returns the current hour.
func (dc *DigitalClock) Hour() int {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	return dc.hour
}

// SetStyle sets custom style.
func (dc *DigitalClock) SetStyle(s DigitalClockStyle) *DigitalClock {
	dc.mu.Lock()
	dc.style = s
	dc.mu.Unlock()
	return dc
}

// Measure returns preferred size.
func (dc *DigitalClock) Measure(cs Constraints) Size {
	w := 8
	if !dc.format24 { w = 11 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// Paint renders the digital clock.
func (dc *DigitalClock) Paint(buf *buffer.Buffer) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	b := dc.Bounds()
	x, y := b.X, b.Y

	digitStyle := dc.style.Digit
	colonStyle := dc.style.Colon
	suffixStyle := dc.style.Suffix

	col := x
	for _, r := range dc.timeStr {
		if col >= buf.Width { break }
		var st buffer.Style
		if r == ':' {
			st = colonStyle
		} else {
			st = digitStyle
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		col++
	}

	// AM/PM suffix
	if dc.suffixStr != "" {
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: suffixStyle.Fg, Bg: suffixStyle.Bg, Flags: suffixStyle.Flags, Width: 1})
			col++
		}
		for _, r := range dc.suffixStr {
			if col >= buf.Width { break }
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: suffixStyle.Fg, Bg: suffixStyle.Bg, Flags: suffixStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (dc *DigitalClock) Children() []Component { return nil }

package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── SpinnerDots: Animated Loading Dots ───
//
// SpinnerDots renders an animated row of dots where one dot is highlighted
// at a time, creating a loading animation effect. Advance the frame with
// Advance().
//
// Usage:
//
//	sd := NewSpinnerDots()
//	sd.SetLabel("Loading")
//	sd.SetDotCount(5)
//	sd.Advance()
//	sd.Paint(buf)

// SpinnerDotsStyle holds styling.
type SpinnerDotsStyle struct {
	Active   buffer.Style
	Inactive buffer.Style
	Label    buffer.Style
}

// DefaultSpinnerDotsStyle returns defaults.
func DefaultSpinnerDotsStyle() SpinnerDotsStyle {
	active := buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Bold}
	inactive := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	return SpinnerDotsStyle{Active: active, Inactive: inactive, Label: label}
}

// SpinnerDots renders animated loading dots.
type SpinnerDots struct {
	BaseComponent
	mu sync.Mutex

	label    string
	dotCount int
	current  int
	style    SpinnerDotsStyle
}

// NewSpinnerDots creates a SpinnerDots.
func NewSpinnerDots() *SpinnerDots {
	sd := &SpinnerDots{dotCount: 3, label: "Loading", style: DefaultSpinnerDotsStyle()}
	sd.SetID(GenerateID("spinner"))
	return sd
}

// SetLabel sets the display label.
func (sd *SpinnerDots) SetLabel(l string) *SpinnerDots {
	sd.mu.Lock()
	sd.label = l
	sd.mu.Unlock()
	return sd
}

// Label returns the label.
func (sd *SpinnerDots) Label() string {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.label
}

// SetDotCount sets the number of dots.
func (sd *SpinnerDots) SetDotCount(n int) *SpinnerDots {
	sd.mu.Lock()
	if n < 1 { n = 1 }
	if n > 20 { n = 20 }
	sd.dotCount = n
	if sd.current >= n { sd.current = 0 }
	sd.mu.Unlock()
	return sd
}

// DotCount returns the number of dots.
func (sd *SpinnerDots) DotCount() int {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.dotCount
}

// Advance moves the active dot to the next position (wraps around).
func (sd *SpinnerDots) Advance() *SpinnerDots {
	sd.mu.Lock()
	sd.current = (sd.current + 1) % sd.dotCount
	sd.mu.Unlock()
	return sd
}

// Reset resets the active dot to position 0.
func (sd *SpinnerDots) Reset() *SpinnerDots {
	sd.mu.Lock()
	sd.current = 0
	sd.mu.Unlock()
	return sd
}

// Current returns the current active dot index.
func (sd *SpinnerDots) Current() int {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.current
}

// SetStyle sets custom style.
func (sd *SpinnerDots) SetStyle(s SpinnerDotsStyle) *SpinnerDots {
	sd.mu.Lock()
	sd.style = s
	sd.mu.Unlock()
	return sd
}

// Measure returns the preferred size.
func (sd *SpinnerDots) Measure(cs Constraints) Size {
	sd.mu.Lock()
	w := len(sd.label) + sd.dotCount*2 + 4
	sd.mu.Unlock()
	if w < 10 { w = 10 }
	h := 1
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: h}
}

// Paint renders the spinner dots into the buffer.
func (sd *SpinnerDots) Paint(buf *buffer.Buffer) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	b := sd.Bounds()
	x, y := b.X, b.Y

	labelStyle := sd.style.Label
	activeStyle := sd.style.Active
	inactiveStyle := sd.style.Inactive

	col := x

	// Label
	for _, r := range sd.label {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	// Space
	if col >= buf.Width { return }
	buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
	col++

	// Dots
	for i := 0; i < sd.dotCount; i++ {
		if col >= buf.Width { return }
		var ch rune
		var style buffer.Style
		if i == sd.current {
			ch = '●'
			style = activeStyle
		} else {
			ch = '○'
			style = inactiveStyle
		}
		buf.SetCell(col, y, buffer.Cell{Rune: ch, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
		col++
		// Space between dots
		if i < sd.dotCount-1 {
			if col >= buf.Width { return }
			buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: inactiveStyle.Fg, Bg: inactiveStyle.Bg, Flags: inactiveStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (sd *SpinnerDots) Children() []Component { return nil }

package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// DigitsStyle holds visual styling for the Digits component.
type DigitsStyle struct {
	Fg    buffer.Color
	Bg    buffer.Color
	Dim   buffer.Color // dim segments for off state
}

// DefaultDigitsStyle returns a green-on-black 7-segment style.
func DefaultDigitsStyle() DigitsStyle {
	return DigitsStyle{
		Fg:  buffer.NamedColor(buffer.NamedBrightGreen),
		Bg:  buffer.Color{Type: buffer.ColorNone},
		Dim: buffer.NamedColor(buffer.NamedBrightBlack),
	}
}

// Each digit 0-9 is rendered as a 3x5 grid of segments.
// The 7-segment display pattern:
//
//	 ___
//	|   |
//	|___|
//	|   |
//	|___|
//
// Segments: a=top, b=top-right, c=bottom-right, d=bottom, e=bottom-left, f=top-left, g=middle

var digitPatterns = [10][5]string{
	// 0: top, sides, bottom (no middle)
	{"███",
		"█ █",
		"█ █",
		"█ █",
		"███"},
	// 1: right side only
	{"  █",
		"  █",
		"  █",
		"  █",
		"  █"},
	// 2: top, top-right, middle, bottom-left, bottom
	{"███",
		"  █",
		"███",
		"█  ",
		"███"},
	// 3: top, right, middle, right, bottom
	{"███",
		"  █",
		"███",
		"  █",
		"███"},
	// 4: left-top, right-top, middle, right-bottom
	{"█ █",
		"█ █",
		"███",
		"  █",
		"  █"},
	// 5: top, top-left, middle, bottom-right, bottom
	{"███",
		"█  ",
		"███",
		"  █",
		"███"},
	// 6: top, top-left, middle, bottom-left, bottom-right, bottom
	{"███",
		"█  ",
		"███",
		"█ █",
		"███"},
	// 7: top, right side
	{"███",
		"  █",
		"  █",
		"  █",
		"  █"},
	// 8: all segments
	{"███",
		"█ █",
		"███",
		"█ █",
		"███"},
	// 9: top, both top sides, middle, bottom-right, bottom
	{"███",
		"█ █",
		"███",
		"  █",
		"███"},
}

var colonPattern = [5]string{
	"   ",
	"  █",
	"   ",
	"  █",
	"   ",
}

var minusPattern = [5]string{
	"   ",
	"   ",
	"███",
	"   ",
	"   ",
}

// Digits displays numbers using large 7-segment-style characters.
type Digits struct {
	BaseComponent

	value    string
	style    DigitsStyle
	showDim  bool // show dim background segments
	digitGap int  // gap between digits

	mu sync.RWMutex
}

// NewDigits creates a Digits component displaying the given value.
func NewDigits(value string) *Digits {
	return &Digits{
		value:    value,
		style:    DefaultDigitsStyle(),
		showDim:  false,
		digitGap: 1,
	}
}

// SetValue sets the displayed value.
func (d *Digits) SetValue(s string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.value = s
}

// Value returns the current displayed value.
func (d *Digits) Value() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.value
}

// SetStyle sets the visual style.
func (d *Digits) SetStyle(s DigitsStyle) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.style = s
}

// SetShowDim toggles showing dim background segments.
func (d *Digits) SetShowDim(b bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.showDim = b
}

// SetDigitGap sets the gap between digits (default 1).
func (d *Digits) SetDigitGap(g int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if g < 0 {
		g = 0
	}
	d.digitGap = g
}

// Measure returns the desired size for the current value.
func (d *Digits) Measure(cs Constraints) Size {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.measureLocked()
}

func (d *Digits) measureLocked() Size {
	chars := []rune(d.value)
	if len(chars) == 0 {
		return Size{W: 0, H: 5}
	}
	w := len(chars)*3 + (len(chars)-1)*d.digitGap
	return Size{W: w, H: 5}
}

// Paint renders the digits into the buffer.
func (d *Digits) Paint(buf *buffer.Buffer) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if buf == nil || d.value == "" {
		return
	}

	bounds := d.Bounds()
	x, y := bounds.X, bounds.Y

	for _, ch := range d.value {
		var pattern [5]string
		switch {
		case ch >= '0' && ch <= '9':
			pattern = digitPatterns[ch-'0']
		case ch == ':':
			pattern = colonPattern
		case ch == '-':
			pattern = minusPattern
		case ch == ' ':
			pattern = [5]string{"   ", "   ", "   ", "   ", "   "}
		default:
			// Unknown char — skip
			x += 3 + d.digitGap
			continue
		}

		for row, line := range pattern {
			for col, r := range line {
				cx := x + col
				cy := y + row
				if cx < bounds.X+bounds.W && cy < bounds.Y+bounds.H {
					if r == '█' {
						buf.SetCell(cx, cy, buffer.Cell{
							Rune:   '█',
							Width:  1,
							Fg:     d.style.Fg,
							Bg:     d.style.Bg,
							Flags:  0,
						})
					} else if d.showDim {
						buf.SetCell(cx, cy, buffer.Cell{
							Rune:   '·',
							Width:  1,
							Fg:     d.style.Dim,
							Bg:     d.style.Bg,
							Flags:  0,
						})
					}
				}
			}
		}
		x += 3 + d.digitGap
	}
}

// HandleKey is a no-op for Digits.
func (d *Digits) HandleKey(_ interface{}) bool { return false }

// Children returns nil.
func (d *Digits) Children() []Component { return nil }

// SetValueInt sets the value from an integer.
func (d *Digits) SetValueInt(n int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if n < 0 {
		d.value = "-" + itoa(-n)
	} else {
		d.value = itoa(n)
	}
}

// SetValueFormatted sets a formatted display (e.g., "12:30:45").
func (d *Digits) SetValueFormatted(s string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	// Filter to only valid characters
	var b strings.Builder
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == ':' || r == '-' || r == ' ' {
			b.WriteRune(r)
		}
	}
	d.value = b.String()
}

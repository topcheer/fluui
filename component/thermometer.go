package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Thermometer: Vertical Temperature Display ───
//
// Thermometer renders a vertical thermometer showing a temperature value
// with a fill level, min/max scale, and unit label. Color-codes based
// on temperature range.
//
// Usage:
//
//	th := NewThermometer()
//	th.SetTemperature(23, -10, 40) // 23°C, range -10 to 40
//	th.Paint(buf)

// ThermometerStyle holds styling.
type ThermometerStyle struct {
	Cold  buffer.Style
	Warm  buffer.Style
	Hot   buffer.Style
	Label buffer.Style
	Value buffer.Style
	Bulb  buffer.Style
}

// DefaultThermometerStyle returns defaults.
func DefaultThermometerStyle() ThermometerStyle {
	return ThermometerStyle{
		Cold:  buffer.Style{Fg: buffer.RGB(59, 130, 246)},
		Warm:  buffer.Style{Fg: buffer.RGB(245, 158, 11)},
		Hot:   buffer.Style{Fg: buffer.RGB(239, 68, 68)},
		Label: buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Value: buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
		Bulb:  buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
	}
}

// Thermometer renders a vertical temperature gauge.
type Thermometer struct {
	BaseComponent
	mu sync.Mutex

	temp int
	min  int
	max  int
	height int
	style  ThermometerStyle
	// cached
	tempStr  string
	minStr   string
	maxStr   string
	fillRows int
	curStyle buffer.Style
}

// NewThermometer creates a Thermometer.
func NewThermometer() *Thermometer {
	th := &Thermometer{height: 8, min: 0, max: 100, style: DefaultThermometerStyle()}
	th.SetID(GenerateID("thermo"))
	th.recomputeLocked()
	return th
}

// SetTemperature sets current temp, min scale, and max scale.
func (th *Thermometer) SetTemperature(temp, min, max int) *Thermometer {
	th.mu.Lock()
	if min >= max { max = min + 1 }
	if temp < min { temp = min }
	if temp > max { temp = max }
	th.temp = temp
	th.min = min
	th.max = max
	th.recomputeLocked()
	th.mu.Unlock()
	return th
}

func (th *Thermometer) recomputeLocked() {
	th.tempStr = itoa(th.temp) + "C"
	th.minStr = itoa(th.min)
	th.maxStr = itoa(th.max)

	// Fill level (0 to height-1)
	range_ := th.max - th.min
	if range_ == 0 { range_ = 1 }
	th.fillRows = (th.temp - th.min) * (th.height - 1) / range_

	// Color coding
	pct := (th.temp - th.min) * 100 / range_
	if pct >= 70 {
		th.curStyle = th.style.Hot
	} else if pct >= 40 {
		th.curStyle = th.style.Warm
	} else {
		th.curStyle = th.style.Cold
	}
}

// Temperature returns the current temperature.
func (th *Thermometer) Temperature() int {
	th.mu.Lock()
	defer th.mu.Unlock()
	return th.temp
}

// SetHeight sets the thermometer height.
func (th *Thermometer) SetHeight(h int) *Thermometer {
	th.mu.Lock()
	if h < 4 { h = 4 }
	th.height = h
	th.recomputeLocked()
	th.mu.Unlock()
	return th
}

// SetStyle sets custom style.
func (th *Thermometer) SetStyle(s ThermometerStyle) *Thermometer {
	th.mu.Lock()
	th.style = s
	th.recomputeLocked()
	th.mu.Unlock()
	return th
}

// Measure returns preferred size.
func (th *Thermometer) Measure(cs Constraints) Size {
	w := 8
	h := th.height + 1
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the thermometer.
func (th *Thermometer) Paint(buf *buffer.Buffer) {
	th.mu.Lock()
	defer th.mu.Unlock()

	b := th.Bounds()
	x, y := b.X, b.Y

	fillStyle := th.curStyle
	labelStyle := th.style.Label
	valueStyle := th.style.Value
	bulbStyle := th.style.Bulb

	// Draw vertical tube
	for row := 0; row < th.height; row++ {
		yy := y + row
		if yy >= buf.Height { break }

		// Fill from bottom: rows below fillRows are filled
		tubeRow := th.height - 1 - row // 0 at bottom
		var ch rune
		var st buffer.Style
		if tubeRow < th.fillRows {
			ch = '█'
			st = fillStyle
		} else {
			ch = '░'
			st = labelStyle
		}
		// Left wall
		if x < buf.Width {
			buf.SetCell(x, yy, buffer.Cell{Rune: '│', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		}
		// Fill column
		if x+1 < buf.Width {
			buf.SetCell(x+1, yy, buffer.Cell{Rune: ch, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		}
		// Right wall
		if x+2 < buf.Width {
			buf.SetCell(x+2, yy, buffer.Cell{Rune: '│', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		}
	}

	// Bulb at bottom
	bulbY := y + th.height
	if bulbY < buf.Height {
		if x+1 < buf.Width {
			buf.SetCell(x+1, bulbY, buffer.Cell{Rune: '◯', Fg: bulbStyle.Fg, Bg: bulbStyle.Bg, Flags: bulbStyle.Flags, Width: 1})
		}
	}

	// Temperature value to the right
	valCol := x + 4
	for _, r := range th.tempStr {
		if valCol >= buf.Width { break }
		buf.SetCell(valCol, y, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		valCol++
	}

	// Max label at top
	maxCol := x + 4
	for _, r := range th.maxStr {
		if maxCol >= buf.Width { break }
		buf.SetCell(maxCol, y+1, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		maxCol++
	}

	// Min label at bottom
	minY := y + th.height - 1
	minCol := x + 4
	for _, r := range th.minStr {
		if minCol >= buf.Width { break }
		buf.SetCell(minCol, minY, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		minCol++
	}
}

// Children returns nil.
func (th *Thermometer) Children() []Component { return nil }

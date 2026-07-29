package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── RadialGauge: Circular Radial Progress Gauge ───
//
// RadialGauge renders a circular progress indicator using arc block
// characters. Shows percentage in the center. Useful for dashboards
// and KPI displays.
//
// Usage:
//
//	rg := NewRadialGauge()
//	rg.SetValue(65) // 65%
//	rg.Paint(buf)

// RadialGaugeStyle holds styling.
type RadialGaugeStyle struct {
	Filled  buffer.Style
	Empty   buffer.Style
	Center  buffer.Style
	Label   buffer.Style
}

// DefaultRadialGaugeStyle returns defaults.
func DefaultRadialGaugeStyle() RadialGaugeStyle {
	return RadialGaugeStyle{
		Filled: buffer.Style{Fg: buffer.RGB(59, 130, 246), Flags: buffer.Bold},
		Empty:  buffer.Style{Fg: buffer.RGB(71, 85, 105)},
		Center: buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
		Label:  buffer.Style{Fg: buffer.RGB(148, 163, 184)},
	}
}

// RadialGauge renders a circular progress gauge.
type RadialGauge struct {
	BaseComponent
	mu sync.Mutex

	value int // 0-100
	label string
	style RadialGaugeStyle
	// cached
	pctStr string
	ring   [12]rune // 12-position ring representation
}

// 12-position radial segments using block characters
var radialGaugeSegs = [...][12]rune{
	{}, // 0%
	{'▏', '░', '░', '░', '░', '░', '░', '░', '░', '░', '░', '░'},
	{'▎', '▏', '░', '░', '░', '░', '░', '░', '░', '░', '░', '░'},
	{'▍', '▎', '▏', '░', '░', '░', '░', '░', '░', '░', '░', '░'},
	{'▌', '▍', '▎', '▏', '░', '░', '░', '░', '░', '░', '░', '░'},
	{'▋', '▌', '▍', '▎', '▏', '░', '░', '░', '░', '░', '░', '░'},
	{'▊', '▋', '▌', '▍', '▎', '▏', '░', '░', '░', '░', '░', '░'},
	{'▉', '▊', '▋', '▌', '▍', '▎', '▏', '░', '░', '░', '░', '░'},
	{'█', '▉', '▊', '▋', '▌', '▍', '▎', '▏', '░', '░', '░', '░'},
}

// NewRadialGauge creates a RadialGauge.
func NewRadialGauge() *RadialGauge {
	rg := &RadialGauge{style: DefaultRadialGaugeStyle()}
	rg.SetID(GenerateID("radial"))
	rg.recomputeLocked()
	return rg
}

// SetValue sets the gauge value (0-100).
func (rg *RadialGauge) SetValue(v int) *RadialGauge {
	rg.mu.Lock()
	if v < 0 { v = 0 }
	if v > 100 { v = 100 }
	rg.value = v
	rg.recomputeLocked()
	rg.mu.Unlock()
	return rg
}

// SetLabel sets an optional label below the percentage.
func (rg *RadialGauge) SetLabel(s string) *RadialGauge {
	rg.mu.Lock()
	rg.label = s
	rg.mu.Unlock()
	return rg
}

func (rg *RadialGauge) recomputeLocked() {
	rg.pctStr = itoa(rg.value) + "%"

	// Select ring segment based on value (8 levels for 12 chars)
	level := rg.value * 8 / 100
	if level > 7 { level = 7 }
	if level < 0 { level = 0 }
	if rg.value == 0 {
		for i := range rg.ring {
			rg.ring[i] = '○'
		}
	} else if rg.value >= 100 {
		for i := range rg.ring {
			rg.ring[i] = '●'
		}
	} else {
		seg := radialGaugeSegs[level]
		copy(rg.ring[:], seg[:])
	}
}

// Value returns the current value.
func (rg *RadialGauge) Value() int {
	rg.mu.Lock()
	defer rg.mu.Unlock()
	return rg.value
}

// SetStyle sets custom style.
func (rg *RadialGauge) SetStyle(s RadialGaugeStyle) *RadialGauge {
	rg.mu.Lock()
	rg.style = s
	rg.mu.Unlock()
	return rg
}

// Measure returns preferred size.
func (rg *RadialGauge) Measure(cs Constraints) Size {
	w := 16
	h := 4
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the radial gauge.
func (rg *RadialGauge) Paint(buf *buffer.Buffer) {
	rg.mu.Lock()
	defer rg.mu.Unlock()

	b := rg.Bounds()
	x, y := b.X, b.Y

	filledStyle := rg.style.Filled
	emptyStyle := rg.style.Empty
	centerStyle := rg.style.Center
	labelStyle := rg.style.Label

	// Row 0: top arc (4 chars)
	for i := 0; i < 4; i++ {
		idx := i + 2
		cx := x + 1 + i
		if cx >= buf.Width { break }
		ch := rg.ring[idx%12]
		var st buffer.Style
		if ch == '░' || ch == '○' { st = emptyStyle } else { st = filledStyle }
		buf.SetCell(cx, y, buffer.Cell{Rune: ch, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
	}

	// Row 1-2: sides + center
	for row := 1; row <= 2; row++ {
		// Left side
		leftCh := rg.ring[(14-row)%12]
		var leftSt buffer.Style
		if leftCh == '░' || leftCh == '○' { leftSt = emptyStyle } else { leftSt = filledStyle }
		if x < buf.Width {
			buf.SetCell(x, y+row, buffer.Cell{Rune: leftCh, Fg: leftSt.Fg, Bg: leftSt.Bg, Flags: leftSt.Flags, Width: 1})
		}
		// Right side
		rightCh := rg.ring[(2+row)%12]
		var rightSt buffer.Style
		if rightCh == '░' || rightCh == '○' { rightSt = emptyStyle } else { rightSt = filledStyle }
		cx := x + 4
		if cx < buf.Width {
			buf.SetCell(cx, y+row, buffer.Cell{Rune: rightCh, Fg: rightSt.Fg, Bg: rightSt.Bg, Flags: rightSt.Flags, Width: 1})
		}
	}

	// Center: percentage text
	centerCol := x + 6
	for _, r := range rg.pctStr {
		if centerCol >= buf.Width { break }
		buf.SetCell(centerCol, y+1, buffer.Cell{Rune: r, Fg: centerStyle.Fg, Bg: centerStyle.Bg, Flags: centerStyle.Flags, Width: 1})
		centerCol++
	}

	// Row 3: bottom arc
	for i := 0; i < 4; i++ {
		idx := (8 + i) % 12
		cx := x + 1 + i
		if cx >= buf.Width { break }
		ch := rg.ring[idx]
		var st buffer.Style
		if ch == '░' || ch == '○' { st = emptyStyle } else { st = filledStyle }
		buf.SetCell(cx, y+3, buffer.Cell{Rune: ch, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
	}

	// Optional label
	if rg.label != "" {
		col := x + 6
		for _, r := range rg.label {
			if col >= buf.Width { break }
			buf.SetCell(col, y+2, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (rg *RadialGauge) Children() []Component { return nil }

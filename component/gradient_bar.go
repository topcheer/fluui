package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── GradientBar: Gradient-Colored Progress Bar ───
//
// GradientBar renders a progress bar that smoothly transitions through
// a color gradient based on the fill percentage. Low values are green,
// mid values are yellow, and high values are red.
//
// Usage:
//
//	gb := NewGradientBar()
//	gb.SetValue(65, 100) // 65%
//	gb.Paint(buf)

// GradientBarStyle holds styling.
type GradientBarStyle struct {
	Label buffer.Style
	Value buffer.Style
}

// DefaultGradientBarStyle returns defaults.
func DefaultGradientBarStyle() GradientBarStyle {
	return GradientBarStyle{
		Label: buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Value: buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
	}
}

// GradientBar renders a gradient progress bar.
type GradientBar struct {
	BaseComponent
	mu sync.Mutex

	value  int
	max    int
	width  int
	style  GradientBarStyle
	// cached
	pctStr  string
	barFill int
}

// NewGradientBar creates a GradientBar.
func NewGradientBar() *GradientBar {
	gb := &GradientBar{width: 24, max: 100, style: DefaultGradientBarStyle()}
	gb.SetID(GenerateID("gradbar"))
	gb.recomputeLocked()
	return gb
}

// SetValue sets current value and maximum.
func (gb *GradientBar) SetValue(value, max int) *GradientBar {
	gb.mu.Lock()
	if value < 0 { value = 0 }
	if max < 1 { max = 1 }
	if value > max { value = max }
	gb.value = value
	gb.max = max
	gb.recomputeLocked()
	gb.mu.Unlock()
	return gb
}

func (gb *GradientBar) recomputeLocked() {
	pct := gb.value * 100 / gb.max
	gb.pctStr = itoa(pct) + "%"

	const barMax = 20
	gb.barFill = gb.value * barMax / gb.max
}

// Percent returns the fill percentage.
func (gb *GradientBar) Percent() int {
	gb.mu.Lock()
	defer gb.mu.Unlock()
	return gb.value * 100 / gb.max
}

// SetWidth sets the bar width.
func (gb *GradientBar) SetWidth(w int) *GradientBar {
	gb.mu.Lock()
	if w < 10 { w = 10 }
	gb.width = w
	gb.mu.Unlock()
	return gb
}

// SetStyle sets custom style.
func (gb *GradientBar) SetStyle(s GradientBarStyle) *GradientBar {
	gb.mu.Lock()
	gb.style = s
	gb.mu.Unlock()
	return gb
}

// Measure returns preferred size.
func (gb *GradientBar) Measure(cs Constraints) Size {
	w := gb.width + 8
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// gradientColor returns an interpolated RGB color based on percentage.
// 0% = green(34,197,94), 50% = yellow(234,179,8), 100% = red(239,68,68)
func gradientBarColor(pct int) (r, g, b int) {
	if pct <= 50 {
		// green to yellow
		t := pct * 2 // 0-100
		r = 34 + (234-34)*t/100
		g = 197 + (179-197)*t/100
		b = 94 + (8-94)*t/100
	} else {
		// yellow to red
		t := (pct - 50) * 2 // 0-100
		r = 234 + (239-234)*t/100
		g = 179 + (68-179)*t/100
		b = 8 + (68-8)*t/100
	}
	return
}

// Paint renders the gradient bar.
func (gb *GradientBar) Paint(buf *buffer.Buffer) {
	gb.mu.Lock()
	defer gb.mu.Unlock()

	b := gb.Bounds()
	x, y := b.X, b.Y

	labelStyle := gb.style.Label
	valueStyle := gb.style.Value

	col := x

	// Filled portion with gradient colors
	for i := 0; i < gb.barFill; i++ {
		if col >= buf.Width { break }
		// Color based on position in bar
		segPct := i * 100 / 20 // position percentage (0-100)
		r, g, bl := gradientBarColor(segPct)
		style_ := buffer.Style{Fg: buffer.RGB(uint8(r), uint8(g), uint8(bl))}
		buf.SetCell(col, y, buffer.Cell{Rune: '█', Fg: style_.Fg, Bg: style_.Bg, Flags: style_.Flags, Width: 1})
		col++
	}

	// Empty portion
	for i := gb.barFill; i < 20; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '░', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Percentage label
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, r := range gb.pctStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (gb *GradientBar) Children() []Component { return nil }

package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MiniGauge: Compact Single-Line Gauge ───
//
// MiniGauge renders a 1-line progress gauge: [######·····] 50%
// with threshold colors (green/yellow/red) and optional label.
//
// Usage:
//
//	mg := NewMiniGauge()
//	mg.SetLabel("CPU")
//	mg.SetValue(75)
//	mg.SetMax(100)
//	mg.SetWidth(20)
//	mg.Paint(buf)

// MiniGaugeStyle holds styling for MiniGauge.
type MiniGaugeStyle struct {
	Normal  buffer.Style
	Warning buffer.Style
	Critical buffer.Style
	Label   buffer.Style
	Percent buffer.Style
}

// DefaultMiniGaugeStyle returns defaults.
func DefaultMiniGaugeStyle() MiniGaugeStyle {
	normal := buffer.Style{Fg: buffer.RGB(34, 197, 94)}
	warn := buffer.Style{Fg: buffer.RGB(234, 179, 8)}
	crit := buffer.Style{Fg: buffer.RGB(239, 68, 68)}
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	pct := buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold}
	return MiniGaugeStyle{Normal: normal, Warning: warn, Critical: crit, Label: label, Percent: pct}
}

// MiniGauge renders a compact single-line gauge.
type MiniGauge struct {
	BaseComponent
	mu sync.Mutex

	value   float64
	max     float64
	width   int
	label   string
	style   MiniGaugeStyle
	// cached
	pctStr   [8]byte
	pctLen   int
}

// NewMiniGauge creates a MiniGauge with defaults.
func NewMiniGauge() *MiniGauge {
	mg := &MiniGauge{max: 100, width: 15, style: DefaultMiniGaugeStyle()}
	mg.SetID(GenerateID("minigauge"))
	return mg
}

// SetValue sets the current value (cached pct string).
func (mg *MiniGauge) SetValue(v float64) *MiniGauge {
	mg.mu.Lock()
	mg.value = v
	pct := 0.0
	if mg.max > 0 {
		pct = v / mg.max * 100
		if pct > 100 { pct = 100 }
		if pct < 0 { pct = 0 }
	}
	mg.pctLen = 0
	digits := itoa(int(pct))
	for i := 0; i < len(digits); i++ {
		mg.pctStr[mg.pctLen] = digits[i]
		mg.pctLen++
	}
	mg.pctStr[mg.pctLen] = '%'
	mg.pctLen++
	mg.mu.Unlock()
	return mg
}

// Value returns the current value.
func (mg *MiniGauge) Value() float64 {
	mg.mu.Lock()
	defer mg.mu.Unlock()
	return mg.value
}

// SetMax sets the maximum value.
func (mg *MiniGauge) SetMax(m float64) *MiniGauge {
	mg.mu.Lock()
	mg.max = m
	mg.mu.Unlock()
	return mg
}

// SetWidth sets the gauge bar width in characters.
func (mg *MiniGauge) SetWidth(w int) *MiniGauge {
	mg.mu.Lock()
	mg.width = w
	mg.mu.Unlock()
	return mg
}

// SetLabel sets the display label.
func (mg *MiniGauge) SetLabel(l string) *MiniGauge {
	mg.mu.Lock()
	mg.label = l
	mg.mu.Unlock()
	return mg
}

// SetStyle sets the custom style.
func (mg *MiniGauge) SetStyle(s MiniGaugeStyle) *MiniGauge {
	mg.mu.Lock()
	mg.style = s
	mg.mu.Unlock()
	return mg
}

// Measure returns the preferred size.
func (mg *MiniGauge) Measure(cs Constraints) Size {
	mg.mu.Lock()
	w := mg.width + len(mg.label) + 12
	mg.mu.Unlock()
	h := 1
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: h}
}

// Paint renders the mini gauge into the buffer.
func (mg *MiniGauge) Paint(buf *buffer.Buffer) {
	mg.mu.Lock()
	defer mg.mu.Unlock()

	b := mg.Bounds()
	x, y := b.X, b.Y

	// Determine threshold color
	ratio := 0.0
	if mg.max > 0 {
		ratio = mg.value / mg.max
		if ratio > 1 { ratio = 1 }
		if ratio < 0 { ratio = 0 }
	}
	var fillStyle buffer.Style
	switch {
	case ratio >= 0.85:
		fillStyle = mg.style.Critical
	case ratio >= 0.6:
		fillStyle = mg.style.Warning
	default:
		fillStyle = mg.style.Normal
	}

	col := x
	labelStyle := mg.style.Label

	// Label
	for _, r := range mg.label {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	if len(mg.label) > 0 {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Filled portion
	filled := int(ratio * float64(mg.width))
	for i := 0; i < mg.width; i++ {
		if col >= buf.Width { return }
		var ch rune
		var style buffer.Style
		if i < filled {
			ch = '▰'
			style = fillStyle
		} else {
			ch = '▱'
			style = labelStyle
		}
		buf.SetCell(col, y, buffer.Cell{Rune: ch, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
		col++
	}

	// Space + percentage
	if col >= buf.Width { return }
	buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
	col++
	pctStyle := mg.style.Percent
	for i := 0; i < mg.pctLen; i++ {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: rune(mg.pctStr[i]), Fg: pctStyle.Fg, Bg: pctStyle.Bg, Flags: pctStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (mg *MiniGauge) Children() []Component { return nil }

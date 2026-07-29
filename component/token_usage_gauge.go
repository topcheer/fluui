package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── TokenUsageGauge: Streaming Token Budget Gauge ───
//
// TokenUsageGauge renders an arc gauge showing used/budget token ratio with
// a pulse indicator for current generation rate.
//
// Usage:
//
//	tug := NewTokenUsageGauge()
//	tug.SetBudget(128000)
//	tug.SetUsed(45000)
//	tug.SetRate(85.5)
//	tug.Paint(buf)

// TokenUsageGaugeStyle holds styling.
type TokenUsageGaugeStyle struct {
	Normal   buffer.Style
	Warning  buffer.Style
	Critical buffer.Style
	Label    buffer.Style
	Value    buffer.Style
	Pulse    buffer.Style
	Border   buffer.Style
}

// DefaultTokenUsageGaugeStyle returns defaults.
func DefaultTokenUsageGaugeStyle() TokenUsageGaugeStyle {
	normal := buffer.Style{Fg: buffer.RGB(34, 197, 94)}
	warn := buffer.Style{Fg: buffer.RGB(234, 179, 8)}
	crit := buffer.Style{Fg: buffer.RGB(239, 68, 68)}
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	value := buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold}
	pulse := buffer.Style{Fg: buffer.RGB(96, 165, 250)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return TokenUsageGaugeStyle{Normal: normal, Warning: warn, Critical: crit, Label: label, Value: value, Pulse: pulse, Border: border}
}

// TokenUsageGauge renders a token budget gauge with pulse.
type TokenUsageGauge struct {
	BaseComponent
	mu sync.Mutex

	budget int
	used   int
	rate   float64 // tokens/sec
	pulse  int
	style  TokenUsageGaugeStyle
	// cached
	pctStr  [8]byte
	pctLen  int
	rateStr string
}

// NewTokenUsageGauge creates a TokenUsageGauge.
func NewTokenUsageGauge() *TokenUsageGauge {
	tug := &TokenUsageGauge{budget: 128000, style: DefaultTokenUsageGaugeStyle()}
	tug.SetID(GenerateID("tokenusage"))
	tug.rateStr = "0 tok/s"
	return tug
}

// SetBudget sets the token budget.
func (tug *TokenUsageGauge) SetBudget(n int) *TokenUsageGauge {
	tug.mu.Lock()
	tug.budget = n
	tug.mu.Unlock()
	return tug
}

// SetUsed sets used tokens (caches percentage display).
func (tug *TokenUsageGauge) SetUsed(n int) *TokenUsageGauge {
	tug.mu.Lock()
	tug.used = n
	pct := 0.0
	if tug.budget > 0 {
		pct = float64(n) / float64(tug.budget) * 100
		if pct > 100 { pct = 100 }
	}
	tug.pctLen = 0
	digits := itoa(int(pct))
	for i := 0; i < len(digits); i++ {
		tug.pctStr[tug.pctLen] = digits[i]
		tug.pctLen++
	}
	tug.pctStr[tug.pctLen] = '%'
	tug.pctLen++
	tug.mu.Unlock()
	return tug
}

// SetRate sets generation rate (caches display string).
func (tug *TokenUsageGauge) SetRate(r float64) *TokenUsageGauge {
	tug.mu.Lock()
	tug.rate = r
	tug.rateStr = itoa(int(r)) + " tok/s"
	tug.mu.Unlock()
	return tug
}

// AdvancePulse increments the pulse counter for animation.
func (tug *TokenUsageGauge) AdvancePulse() *TokenUsageGauge {
	tug.mu.Lock()
	tug.pulse = (tug.pulse + 1) % 4
	tug.mu.Unlock()
	return tug
}

// SetStyle sets custom style.
func (tug *TokenUsageGauge) SetStyle(s TokenUsageGaugeStyle) *TokenUsageGauge {
	tug.mu.Lock()
	tug.style = s
	tug.mu.Unlock()
	return tug
}

// Measure returns preferred size.
func (tug *TokenUsageGauge) Measure(cs Constraints) Size {
	w := 30
	h := 5
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the token usage gauge into the buffer.
func (tug *TokenUsageGauge) Paint(buf *buffer.Buffer) {
	tug.mu.Lock()
	defer tug.mu.Unlock()

	b := tug.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 15 { w = 30 }
	if h < 3 { h = 5 }

	// Determine threshold style
	ratio := 0.0
	if tug.budget > 0 {
		ratio = float64(tug.used) / float64(tug.budget)
		if ratio > 1 { ratio = 1 }
	}
	var fillStyle buffer.Style
	switch {
	case ratio >= 0.85: fillStyle = tug.style.Critical
	case ratio >= 0.6: fillStyle = tug.style.Warning
	default: fillStyle = tug.style.Normal
	}

	// Arc gauge (simulated as horizontal bar with brackets)
	col := x + 1
	gaugeW := w - 2

	// Left bracket
	if col < buf.Width {
		buf.SetCell(col, y+1, buffer.Cell{Rune: '[', Fg: tug.style.Border.Fg, Bg: tug.style.Border.Bg, Width: 1})
	}
	col++

	filled := int(ratio * float64(gaugeW-2))
	for i := 0; i < gaugeW-2; i++ {
		if col >= buf.Width { break }
		var ch rune
		var style buffer.Style
		if i < filled {
			ch = '█'
			style = fillStyle
		} else {
			ch = '·'
			style = tug.style.Label
		}
		buf.SetCell(col, y+1, buffer.Cell{Rune: ch, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y+1, buffer.Cell{Rune: ']', Fg: tug.style.Border.Fg, Bg: tug.style.Border.Bg, Width: 1})
	}

	// Percentage centered on gauge row
	pctStart := x + (w-tug.pctLen)/2
	for i := 0; i < tug.pctLen; i++ {
		cx := pctStart + i
		if cx >= buf.Width { break }
		buf.SetCell(cx, y+1, buffer.Cell{Rune: rune(tug.pctStr[i]), Fg: fillStyle.Fg, Bg: fillStyle.Bg, Flags: fillStyle.Flags, Width: 1})
	}

	// Used/budget on row 2
	labelStyle := tug.style.Label
	valueStyle := tug.style.Value
	col = x + 1
	usedStr := itoa(tug.used) + " / " + itoa(tug.budget) + " tok"
	for _, r := range usedStr {
		if col >= x+w-1 || col >= buf.Width { break }
		buf.SetCell(col, y+2, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}

	// Rate with pulse on row 3
	pulseStyle := tug.style.Pulse
	col = x + 1
	pulseChars := [4]rune{'◐', '◓', '◑', '◒'}
	pulseChar := pulseChars[tug.pulse%4]
	if col < buf.Width {
		buf.SetCell(col, y+3, buffer.Cell{Rune: pulseChar, Fg: pulseStyle.Fg, Bg: pulseStyle.Bg, Flags: pulseStyle.Flags, Width: 1})
	}
	col++
	if col < buf.Width {
		buf.SetCell(col, y+3, buffer.Cell{Rune: ' ', Fg: pulseStyle.Fg, Bg: pulseStyle.Bg, Flags: pulseStyle.Flags, Width: 1})
	}
	col++
	for _, r := range tug.rateStr {
		if col >= x+w-1 || col >= buf.Width { break }
		buf.SetCell(col, y+3, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (tug *TokenUsageGauge) Children() []Component { return nil }

package component

import (
	"strconv"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── TokenMeter: AI Context Window Usage Meter ───
//
// TokenMeter renders a compact bar showing how much of the AI model's context
// window is being used. Green when low, yellow at 50%, red near limit.
//
// Usage:
//
//	tm := NewTokenMeter(128000)
//	tm.SetUsed(45000)
//	tm.Paint(buf) // renders "███░░░░░░░ 35% (45K/128K)"

// TokenMeterStyle holds visual styles.
type TokenMeterStyle struct {
	Safe   buffer.Style // < 50%
	Warn   buffer.Style // 50-75%
	Crit   buffer.Style // > 75%
	Label  buffer.Style
}

// DefaultTokenMeterStyle returns sensible defaults.
func DefaultTokenMeterStyle() TokenMeterStyle {
	return TokenMeterStyle{
		Safe:  buffer.Style{Fg: buffer.RGB(16, 163, 127)},  // green
		Warn:  buffer.Style{Fg: buffer.RGB(255, 175, 64)},  // orange
		Crit:  buffer.Style{Fg: buffer.RGB(220, 80, 80)},   // red
		Label: buffer.Style{Fg: buffer.White},
	}
}

// TokenMeter renders an AI context window usage meter.
type TokenMeter struct {
	BaseComponent
	mu       sync.RWMutex
	max      int
	used     int
	style    TokenMeterStyle
	showPct  bool
	showAbs  bool
}

// NewTokenMeter creates a meter with the given max context size.
func NewTokenMeter(maxTokens int) *TokenMeter {
	tm := &TokenMeter{
		max:     maxTokens,
		style:   DefaultTokenMeterStyle(),
		showPct: true,
		showAbs: true,
	}
	tm.SetID(GenerateID("tokenmeter"))
	return tm
}

// Max returns the maximum token count.
func (tm *TokenMeter) Max() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.max
}

// SetMax sets the maximum token count.
func (tm *TokenMeter) SetMax(n int) *TokenMeter {
	tm.mu.Lock()
	tm.max = n
	tm.mu.Unlock()
	return tm
}

// Used returns the used token count.
func (tm *TokenMeter) Used() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.used
}

// SetUsed sets the used token count.
func (tm *TokenMeter) SetUsed(n int) *TokenMeter {
	tm.mu.Lock()
	tm.used = n
	tm.mu.Unlock()
	return tm
}

// Percent returns the usage percentage (0-100).
func (tm *TokenMeter) Percent() float64 {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	if tm.max <= 0 {
		return 0
	}
	return float64(tm.used) / float64(tm.max) * 100
}

// Remaining returns the remaining tokens.
func (tm *TokenMeter) Remaining() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	r := tm.max - tm.used
	if r < 0 {
		return 0
	}
	return r
}

// IsCritical returns true if usage > 75%.
func (tm *TokenMeter) IsCritical() bool {
	return tm.Percent() > 75
}

// IsWarning returns true if usage > 50%.
func (tm *TokenMeter) IsWarning() bool {
	return tm.Percent() > 50
}

// SetShowPct toggles percentage display.
func (tm *TokenMeter) SetShowPct(show bool) *TokenMeter {
	tm.mu.Lock()
	tm.showPct = show
	tm.mu.Unlock()
	return tm
}

// SetShowAbs toggles absolute count display.
func (tm *TokenMeter) SetShowAbs(show bool) *TokenMeter {
	tm.mu.Lock()
	tm.showAbs = show
	tm.mu.Unlock()
	return tm
}

// SetStyle sets the visual style.
func (tm *TokenMeter) SetStyle(s TokenMeterStyle) *TokenMeter {
	tm.mu.Lock()
	tm.style = s
	tm.mu.Unlock()
	return tm
}

// Measure computes the desired size.
func (tm *TokenMeter) Measure(cs Constraints) Size {
	w := 40
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the token meter.
func (tm *TokenMeter) Paint(buf *buffer.Buffer) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	b := tm.bounds
	if b.W < 4 || b.H < 1 {
		return
	}

	pct := 0.0
	if tm.max > 0 {
		pct = float64(tm.used) / float64(tm.max)
	}

	var style buffer.Style
	switch {
	case pct > 0.75:
		style = tm.style.Crit
	case pct > 0.50:
		style = tm.style.Warn
	default:
		style = tm.style.Safe
	}

	// Bar width: 15 cells, or half the bounds
	barW := 15
	if b.W/2 < barW {
		barW = b.W / 2
	}
	if barW < 3 {
		barW = 3
	}

	filled := int(pct * float64(barW))
	if filled > barW {
		filled = barW
	}

	x := b.X
	for i := 0; i < barW; i++ {
		var r rune
		if i < filled {
			r = '█'
		} else {
			r = '░'
		}
		buf.SetCell(x, b.Y, buffer.Cell{Rune: r, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
		x++
	}

	// Percentage
	if tm.showPct && x < b.X+b.W {
		buf.SetCell(x, b.Y, buffer.Cell{Rune: ' ', Width: 1})
		x++
		pctStr := strconv.FormatFloat(pct*100, 'f', 0, 64) + "%"
		for _, r := range pctStr {
			if x >= b.X+b.W {
				break
			}
			buf.SetCell(x, b.Y, buffer.Cell{Rune: r, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
			x++
		}
	}

	// Absolute counts
	if tm.showAbs && x < b.X+b.W {
		absStr := " (" + formatTokenK(tm.used) + "/" + formatTokenK(tm.max) + ")"
		for _, r := range absStr {
			if x >= b.X+b.W {
				break
			}
			buf.SetCell(x, b.Y, buffer.Cell{Rune: r, Fg: tm.style.Label.Fg, Bg: tm.style.Label.Bg, Width: 1})
			x++
		}
	}
}

// formatTokenK formats tokens as "45K" or "1.2M" (compact).
func formatTokenK(n int) string {
	if n >= 1000000 {
		return strconv.FormatFloat(float64(n)/1e6, 'f', 1, 64) + "M"
	}
	if n >= 1000 {
		return strconv.Itoa(n/1000) + "K"
	}
	return strconv.Itoa(n)
}

// Children returns nil.
func (tm *TokenMeter) Children() []Component { return nil }

package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── BatteryGauge: Device Battery Level Display ───
//
// BatteryGauge renders a battery icon with fill level and charging state.
// Color shifts from green (high) to red (low). Useful for system monitors.
//
// Usage:
//
//	bg := NewBatteryGauge()
//	bg.SetLevel(75)
//	bg.SetCharging(true)
//	bg.Paint(buf)

// BatteryGaugeStyle holds styling.
type BatteryGaugeStyle struct {
	High     buffer.Style
	Medium   buffer.Style
	Low      buffer.Style
	Charging buffer.Style
	Shell    buffer.Style
}

// DefaultBatteryGaugeStyle returns defaults.
func DefaultBatteryGaugeStyle() BatteryGaugeStyle {
	return BatteryGaugeStyle{
		High:     buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Medium:   buffer.Style{Fg: buffer.RGB(245, 158, 11), Flags: buffer.Bold},
		Low:      buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Charging: buffer.Style{Fg: buffer.RGB(59, 130, 246), Flags: buffer.Bold},
		Shell:    buffer.Style{Fg: buffer.RGB(148, 163, 184)},
	}
}

// BatteryGauge renders a battery level indicator.
type BatteryGauge struct {
	BaseComponent
	mu sync.Mutex

	level    int // 0-100
	charging bool
	width    int
	style    BatteryGaugeStyle
	// cached
	fillW    int
	curStyle buffer.Style
	levelStr string
}

// NewBatteryGauge creates a BatteryGauge.
func NewBatteryGauge() *BatteryGauge {
	bg := &BatteryGauge{width: 12, style: DefaultBatteryGaugeStyle()}
	bg.SetID(GenerateID("battery"))
	bg.recomputeLocked()
	return bg
}

// SetLevel sets the battery level (0-100).
func (bg *BatteryGauge) SetLevel(n int) *BatteryGauge {
	bg.mu.Lock()
	if n < 0 {
		n = 0
	}
	if n > 100 {
		n = 100
	}
	bg.level = n
	bg.recomputeLocked()
	bg.mu.Unlock()
	return bg
}

// SetCharging toggles charging state.
func (bg *BatteryGauge) SetCharging(c bool) *BatteryGauge {
	bg.mu.Lock()
	bg.charging = c
	bg.recomputeLocked()
	bg.mu.Unlock()
	return bg
}

func (bg *BatteryGauge) recomputeLocked() {
	bg.fillW = bg.level * bg.width / 100
	bg.levelStr = itoa(bg.level) + "%"

	if bg.charging {
		bg.curStyle = bg.style.Charging
	} else if bg.level >= 50 {
		bg.curStyle = bg.style.High
	} else if bg.level >= 20 {
		bg.curStyle = bg.style.Medium
	} else {
		bg.curStyle = bg.style.Low
	}
}

// Level returns the current level.
func (bg *BatteryGauge) Level() int {
	bg.mu.Lock()
	defer bg.mu.Unlock()
	return bg.level
}

// SetWidth sets the battery width (the fill area width).
func (bg *BatteryGauge) SetWidth(w int) *BatteryGauge {
	bg.mu.Lock()
	if w < 5 {
		w = 5
	}
	bg.width = w
	bg.recomputeLocked()
	bg.mu.Unlock()
	return bg
}

// SetStyle sets custom style.
func (bg *BatteryGauge) SetStyle(s BatteryGaugeStyle) *BatteryGauge {
	bg.mu.Lock()
	bg.style = s
	bg.recomputeLocked()
	bg.mu.Unlock()
	return bg
}

// Measure returns preferred size.
func (bg *BatteryGauge) Measure(cs Constraints) Size {
	w := bg.width + 6 // shell + terminal + space + label
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the battery gauge.
func (bg *BatteryGauge) Paint(buf *buffer.Buffer) {
	bg.mu.Lock()
	defer bg.mu.Unlock()

	b := bg.Bounds()
	x, y := b.X, b.Y

	fillStyle := bg.curStyle
	shellStyle := bg.style.Shell

	col := x

	// Left shell
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: '[', Fg: shellStyle.Fg, Bg: shellStyle.Bg, Flags: shellStyle.Flags, Width: 1})
		col++
	}

	// Fill area
	for i := 0; i < bg.fillW; i++ {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: '█', Fg: fillStyle.Fg, Bg: fillStyle.Bg, Flags: fillStyle.Flags, Width: 1})
		col++
	}
	for i := bg.fillW; i < bg.width; i++ {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: '░', Fg: shellStyle.Fg, Bg: shellStyle.Bg, Flags: shellStyle.Flags, Width: 1})
		col++
	}

	// Terminal (nub)
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ']', Fg: shellStyle.Fg, Bg: shellStyle.Bg, Flags: shellStyle.Flags, Width: 1})
		col++
	}

	// Charging icon or level label
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: shellStyle.Fg, Bg: shellStyle.Bg, Flags: shellStyle.Flags, Width: 1})
		col++
	}
	if bg.charging {
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: '⚡', Fg: fillStyle.Fg, Bg: fillStyle.Bg, Flags: fillStyle.Flags, Width: 1})
			col++
		}
	} else {
		for _, r := range bg.levelStr {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: fillStyle.Fg, Bg: fillStyle.Bg, Flags: fillStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (bg *BatteryGauge) Children() []Component { return nil }

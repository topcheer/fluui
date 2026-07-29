package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ResponseTimer: AI Response Duration Timer ───
//
// ResponseTimer renders a compact display showing the elapsed time of an
// AI response, including TTFB (time to first byte) and total duration.
// Useful for profiling model response performance.
//
// Usage:
//
//	rt := NewResponseTimer()
//	rt.SetDurations(120, 3500) // 120ms TTFB, 3500ms total
//	rt.Paint(buf)

// ResponseTimerStyle holds styling.
type ResponseTimerStyle struct {
	Label buffer.Style
	Value buffer.Style
	Unit  buffer.Style
	Icon  buffer.Style
}

// DefaultResponseTimerStyle returns defaults.
func DefaultResponseTimerStyle() ResponseTimerStyle {
	return ResponseTimerStyle{
		Label: buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Value: buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
		Unit:  buffer.Style{Fg: buffer.RGB(100, 116, 139)},
		Icon:  buffer.Style{Fg: buffer.RGB(59, 130, 246)},
	}
}

// ResponseTimer renders an AI response timing display.
type ResponseTimer struct {
	BaseComponent
	mu sync.Mutex

	tfbMs     int
	totalMs   int
	style     ResponseTimerStyle
	// cached
tfbStr   string
	totalStr string
}

// NewResponseTimer creates a ResponseTimer.
func NewResponseTimer() *ResponseTimer {
	rt := &ResponseTimer{style: DefaultResponseTimerStyle()}
	rt.SetID(GenerateID("resptimer"))
	rt.recomputeLocked()
	return rt
}

// SetDurations sets time-to-first-byte and total duration in milliseconds.
func (rt *ResponseTimer) SetDurations(tfb, total int) *ResponseTimer {
	rt.mu.Lock()
	if tfb < 0 { tfb = 0 }
	if total < 0 { total = 0 }
	rt.tfbMs = tfb
	rt.totalMs = total
	rt.recomputeLocked()
	rt.mu.Unlock()
	return rt
}

func (rt *ResponseTimer) recomputeLocked() {
	rt.tfbStr = itoa(rt.tfbMs)
	rt.totalStr = itoa(rt.totalMs)
}

// TotalDuration returns total duration in ms.
func (rt *ResponseTimer) TotalDuration() int {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	return rt.totalMs
}

// TTFB returns time-to-first-byte in ms.
func (rt *ResponseTimer) TTFB() int {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	return rt.tfbMs
}

// SetStyle sets custom style.
func (rt *ResponseTimer) SetStyle(s ResponseTimerStyle) *ResponseTimer {
	rt.mu.Lock()
	rt.style = s
	rt.mu.Unlock()
	return rt
}

// Measure returns preferred size.
func (rt *ResponseTimer) Measure(cs Constraints) Size {
	w := 28
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// Paint renders the response timer in a single row.
func (rt *ResponseTimer) Paint(buf *buffer.Buffer) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	b := rt.Bounds()
	x, y := b.X, b.Y

	labelStyle := rt.style.Label
	valueStyle := rt.style.Value
	unitStyle := rt.style.Unit
	iconStyle := rt.style.Icon

	col := x

	// ⏱ icon
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: '◷', Fg: iconStyle.Fg, Bg: iconStyle.Bg, Flags: iconStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// TTFB label
	for _, r := range "TTFB" {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	// TTFB value
	for _, r := range rt.tfbStr {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}
	// ms unit
	for _, r := range "ms" {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: unitStyle.Fg, Bg: unitStyle.Bg, Flags: unitStyle.Flags, Width: 1})
		col++
	}
	// separator
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Total label
	for _, r := range "Total" {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	// Total value
	for _, r := range rt.totalStr {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}
	for _, r := range "ms" {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: unitStyle.Fg, Bg: unitStyle.Bg, Flags: unitStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (rt *ResponseTimer) Children() []Component { return nil }

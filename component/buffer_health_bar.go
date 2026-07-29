package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── BufferHealthBar: Render Buffer Utilization Indicator ───
//
// BufferHealthBar renders a compact bar showing render buffer utilization.
// Color-codes green/yellow/red based on fill percentage. Useful for
// monitoring render performance in real-time.
//
// Usage:
//
//	bh := NewBufferHealthBar()
//	bh.SetUtilization(65, 100) // 65% of buffer used
//	bh.Paint(buf)

// BufferHealthStyle holds styling.
type BufferHealthStyle struct {
	Normal  buffer.Style
	High    buffer.Style
	Critical buffer.Style
	Label   buffer.Style
	Pct     buffer.Style
}

// DefaultBufferHealthStyle returns defaults.
func DefaultBufferHealthStyle() BufferHealthStyle {
	return BufferHealthStyle{
		Normal:   buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		High:     buffer.Style{Fg: buffer.RGB(245, 158, 11)},
		Critical: buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Label:    buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Pct:      buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
	}
}

// BufferHealthBar renders a buffer utilization indicator.
type BufferHealthBar struct {
	BaseComponent
	mu sync.Mutex

	used   int
	total  int
	width  int
	style  BufferHealthStyle
	// cached
	pctStr  string
	barFill int
	curStyle buffer.Style
}

// NewBufferHealthBar creates a BufferHealthBar.
func NewBufferHealthBar() *BufferHealthBar {
	bh := &BufferHealthBar{width: 24, total: 100, style: DefaultBufferHealthStyle()}
	bh.SetID(GenerateID("bufhealth"))
	bh.recomputeLocked()
	return bh
}

// SetUtilization sets used and total buffer units.
func (bh *BufferHealthBar) SetUtilization(used, total int) *BufferHealthBar {
	bh.mu.Lock()
	if used < 0 { used = 0 }
	if total < 1 { total = 1 }
	if used > total { used = total }
	bh.used = used
	bh.total = total
	bh.recomputeLocked()
	bh.mu.Unlock()
	return bh
}

func (bh *BufferHealthBar) recomputeLocked() {
	pct := bh.used * 100 / bh.total
	bh.pctStr = itoa(pct) + "%"

	const barMax = 16
	bh.barFill = bh.used * barMax / bh.total

	if pct >= 85 {
		bh.curStyle = bh.style.Critical
	} else if pct >= 60 {
		bh.curStyle = bh.style.High
	} else {
		bh.curStyle = bh.style.Normal
	}
}

// Percent returns utilization percentage.
func (bh *BufferHealthBar) Percent() int {
	bh.mu.Lock()
	defer bh.mu.Unlock()
	return bh.used * 100 / bh.total
}

// SetWidth sets the bar width.
func (bh *BufferHealthBar) SetWidth(w int) *BufferHealthBar {
	bh.mu.Lock()
	if w < 10 { w = 10 }
	bh.width = w
	bh.mu.Unlock()
	return bh
}

// SetStyle sets custom style.
func (bh *BufferHealthBar) SetStyle(s BufferHealthStyle) *BufferHealthBar {
	bh.mu.Lock()
	bh.style = s
	bh.recomputeLocked()
	bh.mu.Unlock()
	return bh
}

// Measure returns preferred size.
func (bh *BufferHealthBar) Measure(cs Constraints) Size {
	w := bh.width + 12
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// Paint renders the buffer health bar.
func (bh *BufferHealthBar) Paint(buf *buffer.Buffer) {
	bh.mu.Lock()
	defer bh.mu.Unlock()

	b := bh.Bounds()
	x, y := b.X, b.Y

	labelStyle := bh.style.Label
	barStyle := bh.curStyle
	pctStyle := bh.style.Pct

	col := x

	// Label
	for _, r := range "Buf " {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Filled bar
	for i := 0; i < bh.barFill; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '█', Fg: barStyle.Fg, Bg: barStyle.Bg, Flags: barStyle.Flags, Width: 1})
		col++
	}
	// Empty bar
	for i := bh.barFill; i < 16; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '░', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Space + percentage
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, r := range bh.pctStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: pctStyle.Fg, Bg: pctStyle.Bg, Flags: pctStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (bh *BufferHealthBar) Children() []Component { return nil }

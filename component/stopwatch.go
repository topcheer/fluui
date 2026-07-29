package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Stopwatch: Elapsed Time Stopwatch Display ───
//
// Stopwatch renders a monospace elapsed time display with centisecond
// precision. Shows hours:minutes:seconds.centiseconds format.
//
// Usage:
//
//	sw := NewStopwatch()
//	sw.SetElapsed(3661500) // 1h 1m 1.5s in milliseconds
//	sw.Paint(buf)

// StopwatchStyle holds styling.
type StopwatchStyle struct {
	Digits buffer.Style
	Dot    buffer.Style
	Label  buffer.Style
}

// DefaultStopwatchStyle returns defaults.
func DefaultStopwatchStyle() StopwatchStyle {
	return StopwatchStyle{
		Digits: buffer.Style{Fg: buffer.RGB(34, 197, 246), Flags: buffer.Bold},
		Dot:    buffer.Style{Fg: buffer.RGB(251, 191, 36), Flags: buffer.Bold},
		Label:  buffer.Style{Fg: buffer.RGB(148, 163, 184)},
	}
}

// Stopwatch renders an elapsed time display.
type Stopwatch struct {
	BaseComponent
	mu sync.Mutex

	elapsedMs int
	running   bool
	style     StopwatchStyle
	// cached
	timeStr string
}

// NewStopwatch creates a Stopwatch.
func NewStopwatch() *Stopwatch {
	sw := &Stopwatch{style: DefaultStopwatchStyle()}
	sw.SetID(GenerateID("stopwatch"))
	sw.recomputeLocked()
	return sw
}

// SetElapsed sets the elapsed time in milliseconds.
func (sw *Stopwatch) SetElapsed(ms int) *Stopwatch {
	sw.mu.Lock()
	if ms < 0 {
		ms = 0
	}
	sw.elapsedMs = ms
	sw.recomputeLocked()
	sw.mu.Unlock()
	return sw
}

// SetRunning toggles the running indicator.
func (sw *Stopwatch) SetRunning(r bool) *Stopwatch {
	sw.mu.Lock()
	sw.running = r
	sw.mu.Unlock()
	return sw
}

func (sw *Stopwatch) recomputeLocked() {
	totalCs := sw.elapsedMs / 10
	cs := totalCs % 100
	totalSec := totalCs / 100
	sec := totalSec % 60
	totalMin := totalSec / 60
	min := totalMin % 60
	hr := totalMin / 60
	sw.timeStr = pad2(hr) + ":" + pad2(min) + ":" + pad2(sec) + "." + pad2(cs)
}

// ElapsedMs returns elapsed time in milliseconds.
func (sw *Stopwatch) ElapsedMs() int {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.elapsedMs
}

// SetStyle sets custom style.
func (sw *Stopwatch) SetStyle(s StopwatchStyle) *Stopwatch {
	sw.mu.Lock()
	sw.style = s
	sw.mu.Unlock()
	return sw
}

// Measure returns preferred size.
func (sw *Stopwatch) Measure(cs Constraints) Size {
	w := 14
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the stopwatch.
func (sw *Stopwatch) Paint(buf *buffer.Buffer) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	b := sw.Bounds()
	x, y := b.X, b.Y

	digitStyle := sw.style.Digits
	dotStyle := sw.style.Dot
	labelStyle := sw.style.Label

	col := x

	// Running indicator
	if sw.running {
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: '▶', Fg: digitStyle.Fg, Bg: digitStyle.Bg, Flags: digitStyle.Flags, Width: 1})
			col++
		}
	} else {
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: '⏸', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Time string
	for _, r := range sw.timeStr {
		if col >= buf.Width {
			break
		}
		var st buffer.Style
		if r == ':' || r == '.' {
			st = dotStyle
		} else {
			st = digitStyle
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (sw *Stopwatch) Children() []Component { return nil }

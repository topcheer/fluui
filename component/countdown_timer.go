package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── CountdownTimer: Countdown Timer Display ───
//
// CountdownTimer renders a countdown timer showing remaining time with
// optional urgency color-coding when time is running low.
//
// Usage:
//
//	ct := NewCountdownTimer()
//	ct.SetRemaining(65000) // 65 seconds remaining
//	ct.SetUrgencyThreshold(30000) // turn urgent at 30s
//	ct.Paint(buf)

// CountdownTimerStyle holds styling.
type CountdownTimerStyle struct {
	Normal  buffer.Style
	Urgent  buffer.Style
	Expired buffer.Style
	Dot     buffer.Style
	Label   buffer.Style
}

// DefaultCountdownTimerStyle returns defaults.
func DefaultCountdownTimerStyle() CountdownTimerStyle {
	return CountdownTimerStyle{
		Normal:  buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Urgent:  buffer.Style{Fg: buffer.RGB(245, 158, 11), Flags: buffer.Bold},
		Expired: buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Dot:     buffer.Style{Fg: buffer.RGB(251, 191, 36)},
		Label:   buffer.Style{Fg: buffer.RGB(148, 163, 184)},
	}
}

// CountdownTimer renders a countdown timer.
type CountdownTimer struct {
	BaseComponent
	mu sync.Mutex

	remainingMs int
	urgencyMs   int
	style       CountdownTimerStyle
	// cached
	timeStr  string
	curStyle buffer.Style
}

// NewCountdownTimer creates a CountdownTimer.
func NewCountdownTimer() *CountdownTimer {
	ct := &CountdownTimer{remainingMs: 60000, urgencyMs: 10000, style: DefaultCountdownTimerStyle()}
	ct.SetID(GenerateID("countdown"))
	ct.recomputeLocked()
	return ct
}

// SetRemaining sets remaining time in milliseconds.
func (ct *CountdownTimer) SetRemaining(ms int) *CountdownTimer {
	ct.mu.Lock()
	if ms < 0 {
		ms = 0
	}
	ct.remainingMs = ms
	ct.recomputeLocked()
	ct.mu.Unlock()
	return ct
}

// SetUrgencyThreshold sets the urgency threshold in milliseconds.
func (ct *CountdownTimer) SetUrgencyThreshold(ms int) *CountdownTimer {
	ct.mu.Lock()
	if ms < 0 {
		ms = 0
	}
	ct.urgencyMs = ms
	ct.recomputeLocked()
	ct.mu.Unlock()
	return ct
}

func (ct *CountdownTimer) recomputeLocked() {
	totalSec := ct.remainingMs / 1000
	sec := totalSec % 60
	min := totalSec / 60
	ct.timeStr = pad2(min) + ":" + pad2(sec)

	if ct.remainingMs == 0 {
		ct.curStyle = ct.style.Expired
	} else if ct.remainingMs <= ct.urgencyMs {
		ct.curStyle = ct.style.Urgent
	} else {
		ct.curStyle = ct.style.Normal
	}
}

// RemainingMs returns remaining time in milliseconds.
func (ct *CountdownTimer) RemainingMs() int {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	return ct.remainingMs
}

// SetStyle sets custom style.
func (ct *CountdownTimer) SetStyle(s CountdownTimerStyle) *CountdownTimer {
	ct.mu.Lock()
	ct.style = s
	ct.recomputeLocked()
	ct.mu.Unlock()
	return ct
}

// Measure returns preferred size.
func (ct *CountdownTimer) Measure(cs Constraints) Size {
	w := 10
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the countdown timer.
func (ct *CountdownTimer) Paint(buf *buffer.Buffer) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	b := ct.Bounds()
	x, y := b.X, b.Y

	timeStyle := ct.curStyle
	dotStyle := ct.style.Dot
	labelStyle := ct.style.Label

	col := x

	// Timer icon
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: '⏱', Fg: timeStyle.Fg, Bg: timeStyle.Bg, Flags: timeStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Time string
	for _, r := range ct.timeStr {
		if col >= buf.Width {
			break
		}
		var st buffer.Style
		if r == ':' {
			st = dotStyle
		} else {
			st = timeStyle
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (ct *CountdownTimer) Children() []Component { return nil }

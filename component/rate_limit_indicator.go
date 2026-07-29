package component

import (
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── RateLimitIndicator: API Rate Limit Status Display ───
//
// RateLimitIndicator renders remaining API requests, reset time, and
// retry-after countdown. Common in AI tooling dashboards that need to
// visualize API quota usage and rate limit headers.
//
// Usage:
//
//	rl := NewRateLimitIndicator()
//	rl.SetLimit(5000)
//	rl.SetRemaining(3200)
//	rl.SetResetTime(time.Now().Add(30 * time.Minute))
//	rl.Paint(buf)

// RateLimitStyle holds styling for RateLimitIndicator.
type RateLimitStyle struct {
	Normal   buffer.Style
	Warning  buffer.Style
	Critical buffer.Style
	Label    buffer.Style
}

// DefaultRateLimitStyle returns sensible defaults.
func DefaultRateLimitStyle() RateLimitStyle {
	normal := buffer.Style{Fg: buffer.RGB(34, 197, 94)}   // green-500
	warning := buffer.Style{Fg: buffer.RGB(234, 179, 8)}  // yellow-500
	critical := buffer.Style{Fg: buffer.RGB(239, 68, 68)} // red-500
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)}  // slate-400
	return RateLimitStyle{Normal: normal, Warning: warning, Critical: critical, Label: label}
}

// RateLimitIndicator displays API rate limit status.
type RateLimitIndicator struct {
	BaseComponent
	mu sync.Mutex

	limit      int
	remaining  int
	resetTime  time.Time
	retryAfter time.Duration

	style RateLimitStyle
}

// NewRateLimitIndicator creates a RateLimitIndicator with defaults.
func NewRateLimitIndicator() *RateLimitIndicator {
	rl := &RateLimitIndicator{
		limit:     5000,
		remaining: 5000,
		style:     DefaultRateLimitStyle(),
	}
	rl.SetID(GenerateID("ratelimit"))
	return rl
}

// SetLimit sets the total request limit.
func (rl *RateLimitIndicator) SetLimit(n int) *RateLimitIndicator {
	rl.mu.Lock()
	rl.limit = n
	rl.mu.Unlock()
	return rl
}

// Limit returns the total request limit.
func (rl *RateLimitIndicator) Limit() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.limit
}

// SetRemaining sets remaining requests.
func (rl *RateLimitIndicator) SetRemaining(n int) *RateLimitIndicator {
	rl.mu.Lock()
	rl.remaining = n
	rl.mu.Unlock()
	return rl
}

// Remaining returns the remaining requests.
func (rl *RateLimitIndicator) Remaining() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.remaining
}

// SetResetTime sets when the rate limit resets.
func (rl *RateLimitIndicator) SetResetTime(t time.Time) *RateLimitIndicator {
	rl.mu.Lock()
	rl.resetTime = t
	rl.mu.Unlock()
	return rl
}

// ResetTime returns the reset time.
func (rl *RateLimitIndicator) ResetTime() time.Time {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.resetTime
}

// SetRetryAfter sets a retry-after duration (e.g., from 429 response).
func (rl *RateLimitIndicator) SetRetryAfter(d time.Duration) *RateLimitIndicator {
	rl.mu.Lock()
	rl.retryAfter = d
	rl.mu.Unlock()
	return rl
}

// RetryAfter returns the retry-after duration.
func (rl *RateLimitIndicator) RetryAfter() time.Duration {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.retryAfter
}

// UsagePercent returns the percentage of limit used (0-100).
func (rl *RateLimitIndicator) UsagePercent() float64 {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.limit <= 0 {
		return 0
	}
	used := rl.limit - rl.remaining
	if used < 0 {
		used = 0
	}
	pct := float64(used) / float64(rl.limit) * 100
	if pct > 100 {
		pct = 100
	}
	return pct
}

// IsRateLimited returns true if currently rate limited (retryAfter > 0 or remaining == 0).
func (rl *RateLimitIndicator) IsRateLimited() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.retryAfter > 0 || rl.remaining <= 0
}

// SetStyle sets the custom style.
func (rl *RateLimitIndicator) SetStyle(s RateLimitStyle) *RateLimitIndicator {
	rl.mu.Lock()
	rl.style = s
	rl.mu.Unlock()
	return rl
}

// Measure returns the preferred size.
func (rl *RateLimitIndicator) Measure(cs Constraints) Size {
	w := 30
	h := 1
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: h}
}

// Paint renders the rate limit status into the buffer.
func (rl *RateLimitIndicator) Paint(buf *buffer.Buffer) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b := rl.Bounds()
	x, y := b.X, b.Y
	col := x

	// Determine color based on remaining percentage
	remainingPct := 1.0
	if rl.limit > 0 {
		remainingPct = float64(rl.remaining) / float64(rl.limit)
	}
	var valueStyle buffer.Style
	switch {
	case rl.retryAfter > 0 || rl.remaining <= 0:
		valueStyle = rl.style.Critical
	case remainingPct < 0.2:
		valueStyle = rl.style.Critical
	case remainingPct < 0.5:
		valueStyle = rl.style.Warning
	default:
		valueStyle = rl.style.Normal
	}

	// Status icon + remaining count
	var icon rune
	if rl.retryAfter > 0 || rl.remaining <= 0 {
		icon = '⛔'
	} else if remainingPct < 0.2 {
		icon = '⚠'
	} else {
		icon = '✓'
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: icon, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}

	// " remaining/limit"
	text := " " + itoa(rl.remaining) + "/" + itoa(rl.limit)
	for _, r := range text {
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		}
		col++
	}

	// Retry-after or reset countdown
	if rl.retryAfter > 0 {
		retryStr := " retry:" + formatInspectorDuration(rl.retryAfter)
		for _, r := range retryStr {
			if col < buf.Width {
				buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: rl.style.Critical.Fg, Bg: rl.style.Critical.Bg, Flags: rl.style.Critical.Flags, Width: 1})
			}
			col++
		}
	} else if !rl.resetTime.IsZero() {
		until := time.Until(rl.resetTime)
		if until > 0 {
			resetStr := " reset:" + formatInspectorDuration(until)
			for _, r := range resetStr {
				if col < buf.Width {
					buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: rl.style.Label.Fg, Bg: rl.style.Label.Bg, Flags: rl.style.Label.Flags, Width: 1})
				}
				col++
			}
		}
	}
}

// Children returns nil.
func (rl *RateLimitIndicator) Children() []Component { return nil }

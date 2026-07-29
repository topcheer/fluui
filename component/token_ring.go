package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── TokenRing: Circular Token Usage Ring Display ───
//
// TokenRing renders a circular progress indicator showing token budget
// usage. Uses arc characters to represent the fill percentage in a
// compact single-line format suitable for status bars.
//
// Usage:
//
//	r := NewTokenRing()
//	r.SetUsage(750, 1000) // 750 used out of 1000 limit
//	r.Paint(buf)

// TokenRingStyle holds styling.
type TokenRingStyle struct {
	Fill    buffer.Style
	Empty   buffer.Style
	Label   buffer.Style
	Percent buffer.Style
}

// DefaultTokenRingStyle returns defaults.
func DefaultTokenRingStyle() TokenRingStyle {
	return TokenRingStyle{
		Fill:    buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Empty:   buffer.Style{Fg: buffer.RGB(71, 85, 105)},
		Label:   buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Percent: buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
	}
}

// TokenRing renders a circular token usage indicator.
type TokenRing struct {
	BaseComponent
	mu sync.Mutex

	used   int
	limit  int
	style  TokenRingStyle
	// cached
	pctStr    string
	usedStr   string
	limitStr  string
	ringChars [4]rune // 4-char compact ring: e.g., [◐◑◓●]
}

// 4-position ring characters for quadrants
var ringQuadrantChars = [...][4]rune{
	{'○', '○', '○', '○'}, // 0%
	{'◐', '○', '○', '○'}, // 1-25%
	{'◐', '◑', '○', '○'}, // 26-50%
	{'◐', '◑', '◓', '○'}, // 51-75%
	{'◐', '◑', '◓', '●'}, // 76-100%
}

// NewTokenRing creates a TokenRing.
func NewTokenRing() *TokenRing {
	r := &TokenRing{limit: 1000, style: DefaultTokenRingStyle()}
	r.SetID(GenerateID("tokenring"))
	r.recomputeLocked()
	return r
}

// SetUsage sets used tokens and limit.
func (r *TokenRing) SetUsage(used, limit int) *TokenRing {
	r.mu.Lock()
	if used < 0 { used = 0 }
	if limit < 1 { limit = 1 }
	if used > limit { used = limit }
	r.used = used
	r.limit = limit
	r.recomputeLocked()
	r.mu.Unlock()
	return r
}

func (r *TokenRing) recomputeLocked() {
	pct := r.used * 100 / r.limit
	r.pctStr = itoa(pct) + "%"
	r.usedStr = itoa(r.used)
	r.limitStr = itoa(r.limit)

	// Select ring quadrant representation
	var idx int
	if pct == 0 {
		idx = 0
	} else if pct <= 25 {
		idx = 1
	} else if pct <= 50 {
		idx = 2
	} else if pct <= 75 {
		idx = 3
	} else {
		idx = 4
	}
	r.ringChars = ringQuadrantChars[idx]
}

// Percent returns usage percentage.
func (r *TokenRing) Percent() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.used * 100 / r.limit
}

// SetStyle sets custom style.
func (r *TokenRing) SetStyle(s TokenRingStyle) *TokenRing {
	r.mu.Lock()
	r.style = s
	r.mu.Unlock()
	return r
}

// Measure returns preferred size.
func (r *TokenRing) Measure(cs Constraints) Size {
	w := 20
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// Paint renders the token ring.
func (r *TokenRing) Paint(buf *buffer.Buffer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	b := r.Bounds()
	x, y := b.X, b.Y

	fillStyle := r.style.Fill
	emptyStyle := r.style.Empty
	labelStyle := r.style.Label
	pctStyle := r.style.Percent

	col := x

	// Ring characters (4 chars showing fill level)
	for i := 0; i < 4; i++ {
		if col >= buf.Width { break }
		ch := r.ringChars[i]
		var style_ buffer.Style
		if ch == '○' {
			style_ = emptyStyle
		} else {
			style_ = fillStyle
		}
		buf.SetCell(col, y, buffer.Cell{Rune: ch, Fg: style_.Fg, Bg: style_.Bg, Flags: style_.Flags, Width: 1})
		col++
	}

	// Space
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Percentage
	for _, rr := range r.pctStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: rr, Fg: pctStyle.Fg, Bg: pctStyle.Bg, Flags: pctStyle.Flags, Width: 1})
		col++
	}

	// Space + label
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, rr := range r.usedStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: rr, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: '/', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, rr := range r.limitStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: rr, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (r *TokenRing) Children() []Component { return nil }

package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ActivityRing: Circular Activity Pulse Indicator ───
//
// ActivityRing renders a pulsing ring that shows recent activity count.
// The ring fills proportionally to the ratio of current vs max events.
// Useful for real-time monitoring dashboards.
//
// Usage:
//
//	ar := NewActivityRing()
//	ar.SetCount(75, 100) // 75 of 100 events
//	ar.Paint(buf)

// ActivityRingStyle holds styling.
type ActivityRingStyle struct {
	Filled buffer.Style
	Empty  buffer.Style
	Center buffer.Style
}

// DefaultActivityRingStyle returns defaults.
func DefaultActivityRingStyle() ActivityRingStyle {
	return ActivityRingStyle{
		Filled: buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Empty:  buffer.Style{Fg: buffer.RGB(30, 41, 59)},
		Center: buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
	}
}

var ringArcs = [9][8]rune{
	{'○', '○', '○', '○', '○', '○', '○', '○'},
	{'◔', '○', '○', '○', '○', '○', '○', '○'},
	{'◔', '◔', '○', '○', '○', '○', '○', '○'},
	{'◑', '◔', '○', '○', '○', '○', '○', '○'},
	{'◑', '◑', '◔', '○', '○', '○', '○', '○'},
	{'◕', '◑', '◔', '○', '○', '○', '○', '○'},
	{'◕', '◕', '◑', '◔', '○', '○', '○', '○'},
	{'●', '◕', '◑', '◔', '○', '○', '○', '○'},
	{'●', '●', '●', '●', '●', '●', '●', '●'},
}

// ActivityRing renders a circular activity pulse indicator.
type ActivityRing struct {
	BaseComponent
	mu sync.Mutex

	count int
	max   int
	label string
	style ActivityRingStyle
	// cached
	level    int
	countStr string
}

// NewActivityRing creates an ActivityRing.
func NewActivityRing() *ActivityRing {
	ar := &ActivityRing{max: 100, style: DefaultActivityRingStyle()}
	ar.SetID(GenerateID("actring"))
	ar.recomputeLocked()
	return ar
}

// SetCount sets current count and max.
func (ar *ActivityRing) SetCount(count, max int) *ActivityRing {
	ar.mu.Lock()
	if count < 0 {
		count = 0
	}
	if max < 1 {
		max = 1
	}
	if count > max {
		count = max
	}
	ar.count = count
	ar.max = max
	ar.recomputeLocked()
	ar.mu.Unlock()
	return ar
}

// SetLabel sets an optional center label.
func (ar *ActivityRing) SetLabel(s string) *ActivityRing {
	ar.mu.Lock()
	ar.label = s
	ar.mu.Unlock()
	return ar
}

func (ar *ActivityRing) recomputeLocked() {
	ratio := ar.count * 8 / ar.max
	if ratio > 8 {
		ratio = 8
	}
	ar.level = ratio
	ar.countStr = itoa(ar.count)
}

// Count returns the current count.
func (ar *ActivityRing) Count() int {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	return ar.count
}

// SetStyle sets custom style.
func (ar *ActivityRing) SetStyle(s ActivityRingStyle) *ActivityRing {
	ar.mu.Lock()
	ar.style = s
	ar.mu.Unlock()
	return ar
}

// Measure returns preferred size.
func (ar *ActivityRing) Measure(cs Constraints) Size {
	w := 8
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the activity ring.
func (ar *ActivityRing) Paint(buf *buffer.Buffer) {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	b := ar.Bounds()
	x, y := b.X, b.Y

	filledStyle := ar.style.Filled
	emptyStyle := ar.style.Empty
	centerStyle := ar.style.Center

	arc := ringArcs[ar.level]
	col := x
	for i := 0; i < 8; i++ {
		if col >= buf.Width {
			break
		}
		var st buffer.Style
		if arc[i] == '○' {
			st = emptyStyle
		} else {
			st = filledStyle
		}
		buf.SetCell(col, y, buffer.Cell{Rune: arc[i], Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		col++
	}

	// Center label after ring
	if ar.label != "" {
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: centerStyle.Fg, Bg: centerStyle.Bg, Flags: centerStyle.Flags, Width: 1})
			col++
		}
		for _, r := range ar.label {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: centerStyle.Fg, Bg: centerStyle.Bg, Flags: centerStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (ar *ActivityRing) Children() []Component { return nil }

package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── SignalBars: Wireless Signal Strength Indicator ───
//
// SignalBars renders a compact signal strength indicator with ascending
// bars. Shows 0-5 bars based on signal level, with optional dBm label.
//
// Usage:
//
//	sb := NewSignalBars()
//	sb.SetLevel(4) // 4 out of 5 bars
//	sb.SetDbm(-67)
//	sb.Paint(buf)

// SignalBarsStyle holds styling.
type SignalBarsStyle struct {
	Strong buffer.Style
	Medium buffer.Style
	Weak   buffer.Style
	Label  buffer.Style
}

// DefaultSignalBarsStyle returns defaults.
func DefaultSignalBarsStyle() SignalBarsStyle {
	return SignalBarsStyle{
		Strong: buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Medium: buffer.Style{Fg: buffer.RGB(245, 158, 11)},
		Weak:   buffer.Style{Fg: buffer.RGB(239, 68, 68)},
		Label:  buffer.Style{Fg: buffer.RGB(148, 163, 184)},
	}
}

const signalBarsMax = 5

// bar heights for each of 5 ascending bars
var signalBarHeights = [5]int{1, 2, 3, 4, 5}

// SignalBars renders a signal strength indicator.
type SignalBars struct {
	BaseComponent
	mu sync.Mutex

	level int // 0-5
	dbm   int
	style SignalBarsStyle
	// cached
	dbmStr   string
	curStyle buffer.Style
}

// NewSignalBars creates a SignalBars.
func NewSignalBars() *SignalBars {
	sb := &SignalBars{style: DefaultSignalBarsStyle()}
	sb.SetID(GenerateID("signal"))
	sb.recomputeLocked()
	return sb
}

// SetLevel sets signal level (0-5 bars).
func (sb *SignalBars) SetLevel(n int) *SignalBars {
	sb.mu.Lock()
	if n < 0 { n = 0 }
	if n > signalBarsMax { n = signalBarsMax }
	sb.level = n
	sb.recomputeLocked()
	sb.mu.Unlock()
	return sb
}

// SetDbm sets the dBm value for display.
func (sb *SignalBars) SetDbm(dbm int) *SignalBars {
	sb.mu.Lock()
	sb.dbm = dbm
	sb.recomputeLocked()
	sb.mu.Unlock()
	return sb
}

func (sb *SignalBars) recomputeLocked() {
	sb.dbmStr = itoa(sb.dbm) + "dBm"

	if sb.level >= 4 {
		sb.curStyle = sb.style.Strong
	} else if sb.level >= 2 {
		sb.curStyle = sb.style.Medium
	} else {
		sb.curStyle = sb.style.Weak
	}
}

// Level returns the current signal level.
func (sb *SignalBars) Level() int {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return sb.level
}

// SetStyle sets custom style.
func (sb *SignalBars) SetStyle(s SignalBarsStyle) *SignalBars {
	sb.mu.Lock()
	sb.style = s
	sb.recomputeLocked()
	sb.mu.Unlock()
	return sb
}

// Measure returns preferred size.
func (sb *SignalBars) Measure(cs Constraints) Size {
	w := 14
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: signalBarsMax}
}

// Paint renders the signal bars.
func (sb *SignalBars) Paint(buf *buffer.Buffer) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	b := sb.Bounds()
	x, y := b.X, b.Y

	barStyle := sb.curStyle
	labelStyle := sb.style.Label
	maxH := signalBarsMax

	for i := 0; i < signalBarsMax; i++ {
		barH := signalBarHeights[i]
		active := i < sb.level
		cx := x + i
		if cx >= buf.Width { break }

		for row := 0; row < maxH; row++ {
			yy := y + maxH - 1 - row
			if yy >= buf.Height { continue }

			var ch rune
			var st buffer.Style
			if active && row < barH {
				ch = '█'
				st = barStyle
			} else {
				ch = '░'
				st = labelStyle
			}
			buf.SetCell(cx, yy, buffer.Cell{Rune: ch, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		}
	}

	// dBm label on the right
	labelCol := x + signalBarsMax + 1
	for _, r := range sb.dbmStr {
		if labelCol >= buf.Width { break }
		buf.SetCell(labelCol, y+maxH-1, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		labelCol++
	}
}

// Children returns nil.
func (sb *SignalBars) Children() []Component { return nil }

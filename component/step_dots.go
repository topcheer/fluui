package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── StepDots: Dot-Based Progress Stepper ───
//
// StepDots renders a compact progress indicator using dots (○/●) with
// optional connecting lines. Each dot represents a step; filled dots
// indicate completed steps. More space-efficient than numbered steppers.
//
// Usage:
//
//	sd := NewStepDots()
//	sd.SetTotal(5)
//	sd.SetCurrent(3) // 3 of 5 done
//	sd.Paint(buf)

// StepDotsStyle holds styling.
type StepDotsStyle struct {
	Done      buffer.Style
	Current   buffer.Style
	Pending   buffer.Style
	Connector buffer.Style
}

// DefaultStepDotsStyle returns defaults.
func DefaultStepDotsStyle() StepDotsStyle {
	return StepDotsStyle{
		Done:      buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Current:   buffer.Style{Fg: buffer.RGB(251, 191, 36), Flags: buffer.Bold},
		Pending:   buffer.Style{Fg: buffer.RGB(71, 85, 105)},
		Connector: buffer.Style{Fg: buffer.RGB(51, 65, 85)},
	}
}

const stepDotsMax = 20

// StepDots renders a dot-based progress stepper.
type StepDots struct {
	BaseComponent
	mu sync.Mutex

	total     int
	current   int
	connected bool
	style     StepDotsStyle
}

// NewStepDots creates a StepDots.
func NewStepDots() *StepDots {
	sd := &StepDots{total: 5, style: DefaultStepDotsStyle()}
	sd.SetID(GenerateID("stepdots"))
	return sd
}

// SetTotal sets the total number of steps.
func (sd *StepDots) SetTotal(n int) *StepDots {
	sd.mu.Lock()
	if n < 1 {
		n = 1
	}
	if n > stepDotsMax {
		n = stepDotsMax
	}
	sd.total = n
	sd.mu.Unlock()
	return sd
}

// SetCurrent sets the current step (0-indexed; values >= total mean all done).
func (sd *StepDots) SetCurrent(n int) *StepDots {
	sd.mu.Lock()
	if n < 0 {
		n = 0
	}
	sd.current = n
	sd.mu.Unlock()
	return sd
}

// SetConnected toggles connecting lines between dots.
func (sd *StepDots) SetConnected(c bool) *StepDots {
	sd.mu.Lock()
	sd.connected = c
	sd.mu.Unlock()
	return sd
}

// Current returns the current step index.
func (sd *StepDots) Current() int {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.current
}

// SetStyle sets custom style.
func (sd *StepDots) SetStyle(s StepDotsStyle) *StepDots {
	sd.mu.Lock()
	sd.style = s
	sd.mu.Unlock()
	return sd
}

// Measure returns preferred size.
func (sd *StepDots) Measure(cs Constraints) Size {
	sd.mu.Lock()
	w := sd.total
	if sd.connected {
		w = sd.total*2 - 1
	}
	sd.mu.Unlock()
	if w < 1 {
		w = 1
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the step dots.
func (sd *StepDots) Paint(buf *buffer.Buffer) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	b := sd.Bounds()
	x, y := b.X, b.Y

	doneStyle := sd.style.Done
	curStyle := sd.style.Current
	pendStyle := sd.style.Pending
	connStyle := sd.style.Connector

	col := x
	for i := 0; i < sd.total; i++ {
		if col >= buf.Width {
			break
		}

		var r rune
		var st buffer.Style
		if i < sd.current {
			r = '●'
			st = doneStyle
		} else if i == sd.current {
			r = '◑'
			st = curStyle
		} else {
			r = '○'
			st = pendStyle
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		col++

		// Connector line
		if sd.connected && i < sd.total-1 {
			if col >= buf.Width {
				break
			}
			var connR rune
			if i < sd.current {
				connR = '━'
			} else {
				connR = '─'
			}
			buf.SetCell(col, y, buffer.Cell{Rune: connR, Fg: connStyle.Fg, Bg: connStyle.Bg, Flags: connStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (sd *StepDots) Children() []Component { return nil }

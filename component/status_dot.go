package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── StatusDot: Animated Status Indicator Dot ───
//
// StatusDot renders a single colored dot with a label. Supports 5 states:
// idle, active, success, warning, error. Useful as inline status indicators
// in lists, tables, and dashboards.
//
// Usage:
//
//	sd := NewStatusDot()
//	sd.SetState(StatusActive)
//	sd.SetLabel("Processing...")
//	sd.Paint(buf)

type StatusDotState int

const (
	StatusIdle    StatusDotState = 0
	StatusActive  StatusDotState = 1
	StatusSuccess StatusDotState = 2
	StatusWarning StatusDotState = 3
	StatusError   StatusDotState = 4
)

var statusDotIcons = [...]rune{'○', '◉', '✓', '⚠', '✗'}
var statusDotLabels = [...]string{"Idle", "Active", "Success", "Warning", "Error"}

type StatusDotStyle struct {
	Idle    buffer.Style
	Active  buffer.Style
	Success buffer.Style
	Warning buffer.Style
	Error   buffer.Style
	Label   buffer.Style
}

func DefaultStatusDotStyle() StatusDotStyle {
	return StatusDotStyle{
		Idle:    buffer.Style{Fg: buffer.RGB(100, 116, 139)},
		Active:  buffer.Style{Fg: buffer.RGB(59, 130, 246), Flags: buffer.Bold},
		Success: buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Warning: buffer.Style{Fg: buffer.RGB(245, 158, 11)},
		Error:   buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Label:   buffer.Style{Fg: buffer.RGB(226, 232, 240)},
	}
}

// StatusDot renders an animated status indicator.
type StatusDot struct {
	BaseComponent
	mu sync.Mutex

	state       StatusDotState
	label       string
	showDefault bool
	style       StatusDotStyle
	// cached
	curStyle   buffer.Style
	stateLabel string
}

// NewStatusDot creates a StatusDot.
func NewStatusDot() *StatusDot {
	sd := &StatusDot{state: StatusIdle, showDefault: true, style: DefaultStatusDotStyle()}
	sd.SetID(GenerateID("statusdot"))
	sd.recomputeLocked()
	return sd
}

// SetState sets the dot state.
func (sd *StatusDot) SetState(s StatusDotState) *StatusDot {
	sd.mu.Lock()
	if int(s) < 0 || int(s) >= len(statusDotIcons) {
		s = StatusIdle
	}
	sd.state = s
	sd.recomputeLocked()
	sd.mu.Unlock()
	return sd
}

// SetLabel sets a custom label.
func (sd *StatusDot) SetLabel(l string) *StatusDot {
	sd.mu.Lock()
	sd.label = l
	sd.mu.Unlock()
	return sd
}

// SetShowDefaultLabel toggles showing the default state label when no custom label is set.
func (sd *StatusDot) SetShowDefaultLabel(show bool) *StatusDot {
	sd.mu.Lock()
	sd.showDefault = show
	sd.mu.Unlock()
	return sd
}

func (sd *StatusDot) recomputeLocked() {
	switch sd.state {
	case StatusIdle:
		sd.curStyle = sd.style.Idle
	case StatusActive:
		sd.curStyle = sd.style.Active
	case StatusSuccess:
		sd.curStyle = sd.style.Success
	case StatusWarning:
		sd.curStyle = sd.style.Warning
	case StatusError:
		sd.curStyle = sd.style.Error
	}
	sd.stateLabel = statusDotLabels[sd.state]
}

// State returns the current state.
func (sd *StatusDot) State() StatusDotState {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.state
}

// SetStyle sets custom style.
func (sd *StatusDot) SetStyle(s StatusDotStyle) *StatusDot {
	sd.mu.Lock()
	sd.style = s
	sd.recomputeLocked()
	sd.mu.Unlock()
	return sd
}

// Measure returns preferred size.
func (sd *StatusDot) Measure(cs Constraints) Size {
	w := 14
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the status dot.
func (sd *StatusDot) Paint(buf *buffer.Buffer) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	b := sd.Bounds()
	x, y := b.X, b.Y

	dotStyle := sd.curStyle
	labelStyle := sd.style.Label

	col := x

	// Dot icon
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: statusDotIcons[sd.state], Fg: dotStyle.Fg, Bg: dotStyle.Bg, Flags: dotStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Label: custom or default state label
	label := sd.label
	if label == "" && sd.showDefault {
		label = sd.stateLabel
	}

	var st buffer.Style
	if label == sd.stateLabel {
		st = dotStyle
	} else {
		st = labelStyle
	}
	for _, r := range label {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (sd *StatusDot) Children() []Component { return nil }

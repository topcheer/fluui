package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── StepProgress: Horizontal Multi-Step Progress Indicator ───
//
// StepProgress renders a horizontal sequence of numbered steps with connecting
// lines, showing completed (✓), current (●), and upcoming (○) states.
// Common in wizards, checkout flows, and multi-step forms.
//
// Usage:
//
//	sp := NewStepProgress()
//	sp.AddStep("Account")
//	sp.AddStep("Profile")
//	sp.AddStep("Confirm")
//	sp.SetCurrentStep(1)
//	sp.Paint(buf)

// StepProgressStyle holds styling for StepProgress.
type StepProgressStyle struct {
	Completed buffer.Style
	Current   buffer.Style
	Upcoming  buffer.Style
	Connector buffer.Style // connecting line
	Label     buffer.Style
}

// DefaultStepProgressStyle returns sensible defaults.
func DefaultStepProgressStyle() StepProgressStyle {
	completed := buffer.Style{Fg: buffer.RGB(34, 197, 94)}    // green-500
	current := buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Bold} // blue-400 bold
	upcoming := buffer.Style{Fg: buffer.RGB(71, 85, 105)}     // slate-600
	connector := buffer.Style{Fg: buffer.RGB(100, 116, 139)}  // slate-500
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)}      // slate-400
	return StepProgressStyle{Completed: completed, Current: current, Upcoming: upcoming, Connector: connector, Label: label}
}

// StepState describes a step's display state.
type StepState int

const (
	StepUpcoming  StepState = iota
	StepCurrent
	StepCompleted
)

// StepProgress displays a horizontal multi-step progress bar.
type StepProgress struct {
	BaseComponent
	mu sync.Mutex

	steps       []string
	current     int
	allDone     bool
	style       StepProgressStyle
}

// NewStepProgress creates a StepProgress with defaults.
func NewStepProgress() *StepProgress {
	sp := &StepProgress{
		style: DefaultStepProgressStyle(),
	}
	sp.SetID(GenerateID("stepprog"))
	return sp
}

// AddStep adds a labeled step to the progress indicator.
func (sp *StepProgress) AddStep(label string) *StepProgress {
	sp.mu.Lock()
	sp.steps = append(sp.steps, label)
	sp.mu.Unlock()
	return sp
}

// SetCurrentStep sets the current active step (0-indexed, clamped).
func (sp *StepProgress) SetCurrentStep(idx int) *StepProgress {
	sp.mu.Lock()
	sp.allDone = false
	if len(sp.steps) > 0 {
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sp.steps) {
			idx = len(sp.steps) - 1
		}
		sp.current = idx
	}
	sp.mu.Unlock()
	return sp
}

// CurrentStep returns the current step index.
func (sp *StepProgress) CurrentStep() int {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	return sp.current
}

// StepCount returns the total number of steps.
func (sp *StepProgress) StepCount() int {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	return len(sp.steps)
}

// Complete marks all steps as completed.
func (sp *StepProgress) Complete() *StepProgress {
	sp.mu.Lock()
	sp.allDone = true
	sp.mu.Unlock()
	return sp
}

// Reset resets to the first step.
func (sp *StepProgress) Reset() *StepProgress {
	sp.mu.Lock()
	sp.allDone = false
	sp.current = 0
	sp.mu.Unlock()
	return sp
}

// SetStyle sets the custom style.
func (sp *StepProgress) SetStyle(s StepProgressStyle) *StepProgress {
	sp.mu.Lock()
	sp.style = s
	sp.mu.Unlock()
	return sp
}

// stepStateLocked returns the state for step at given index.
func (sp *StepProgress) stepStateLocked(idx int) StepState {
	if sp.allDone {
		return StepCompleted
	}
	if idx < sp.current {
		return StepCompleted
	}
	if idx == sp.current {
		return StepCurrent
	}
	return StepUpcoming
}

// Measure returns the preferred size.
func (sp *StepProgress) Measure(cs Constraints) Size {
	sp.mu.Lock()
	count := len(sp.steps)
	sp.mu.Unlock()

	w := count*6 + 4 // ~6 chars per step (circle + connector + label space)
	if w < 20 {
		w = 20
	}
	h := 2 // step row + label row
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the step progress indicator into the buffer.
func (sp *StepProgress) Paint(buf *buffer.Buffer) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	b := sp.Bounds()
	x, y := b.X, b.Y

	if len(sp.steps) == 0 {
		return
	}

	col := x
	connectorStyle := sp.style.Connector

	for idx, label := range sp.steps {
		state := sp.stepStateLocked(idx)

		var stepStyle buffer.Style
		var stepChar rune
		switch state {
		case StepCompleted:
			stepStyle = sp.style.Completed
			stepChar = '✓'
		case StepCurrent:
			stepStyle = sp.style.Current
			stepChar = '●'
		default:
			stepStyle = sp.style.Upcoming
			stepChar = '○'
		}

		// Draw step circle/check
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: stepChar, Fg: stepStyle.Fg, Bg: stepStyle.Bg, Flags: stepStyle.Flags, Width: 1})
		}
		col++

		// Draw label on same row (abbreviated to fit)
		labelStyle := sp.style.Label
		if state == StepCurrent {
			labelStyle = sp.style.Current
		}
		for _, r := range label {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}

		// Draw connector line to next step (if not last)
		if idx < len(sp.steps)-1 {
			connLen := 3 // 3 dashes between steps
			for i := 0; i < connLen; i++ {
				if col >= buf.Width {
					break
				}
				var lineStyle buffer.Style
				var lineChar rune
				if state == StepCompleted {
					lineStyle = sp.style.Completed
					lineChar = '━'
				} else {
					lineStyle = connectorStyle
					lineChar = '─'
				}
				buf.SetCell(col, y, buffer.Cell{Rune: lineChar, Fg: lineStyle.Fg, Bg: lineStyle.Bg, Flags: lineStyle.Flags, Width: 1})
				col++
			}
		}
	}
}

// Children returns nil.
func (sp *StepProgress) Children() []Component { return nil }

package component

import (
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// StepperStep represents a single step in a Stepper.
type StepperStep struct {
	Title       string
	Description string
}

// Stepper renders a horizontal multi-step progress indicator.
// Common in wizards, onboarding flows, and multi-stage AI pipelines.
// Shows completed steps with a checkmark, current step highlighted,
// and future steps dimmed.
//
// Thread-safe. Zero-alloc Paint.
type Stepper struct {
	BaseComponent
	mu sync.Mutex

	steps    []StepperStep
	current  int // 0-based current step index
	vertical bool
}

// NewStepper creates a horizontal stepper with the given steps.
func NewStepper(steps []StepperStep) *Stepper {
	return &Stepper{
		BaseComponent: BaseComponent{id: GenerateID("stepper")},
		steps:         steps,
		current:       0,
	}
}

// CurrentStep returns the index of the active step.
func (s *Stepper) CurrentStep() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current
}

// SetCurrent sets the active step, clamped to valid range.
func (s *Stepper) SetCurrent(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx < 0 {
		idx = 0
	}
	if idx >= len(s.steps) {
		idx = len(s.steps) - 1
	}
	s.current = idx
}

// Next advances to the next step. Returns false if already at last step.
func (s *Stepper) Next() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current >= len(s.steps)-1 {
		return false
	}
	s.current++
	return true
}

// Prev moves to the previous step. Returns false if already at first step.
func (s *Stepper) Prev() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current <= 0 {
		return false
	}
	s.current--
	return true
}

// SetSteps replaces all steps.
func (s *Stepper) SetSteps(steps []StepperStep) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.steps = steps
	if s.current >= len(steps) {
		s.current = len(steps) - 1
	}
	if s.current < 0 {
		s.current = 0
	}
}

// StepCount returns the number of steps.
func (s *Stepper) StepCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.steps)
}

// IsComplete returns true if all steps are done (current is past the last).
func (s *Stepper) IsComplete() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current >= len(s.steps)-1
}

// SetVertical toggles vertical orientation.
func (s *Stepper) SetVertical(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vertical = v
}

// IsVertical returns whether vertical mode is active.
func (s *Stepper) IsVertical() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.vertical
}

// Measure returns the desired size.
func (s *Stepper) Measure(cs Constraints) Size {
	s.mu.Lock()
	vertical := s.vertical
	steps := s.steps
	s.mu.Unlock()

	if vertical {
		h := len(steps) * 2
		if h < 1 {
			h = 1
		}
		maxW := cs.MaxWidth
		if maxW <= 0 {
			maxW = 40
		}
		return Size{W: maxW, H: h}
	}
	w := 0
	for _, step := range steps {
		w += utf8.RuneCountInString(step.Title) + 6 // circle + spaces
	}
	if w < 1 {
		w = 1
	}
	maxH := cs.MaxHeight
	if maxH <= 0 {
		maxH = 1
	}
	return Size{W: w, H: maxH}
}

// Paint renders the stepper.
func (s *Stepper) Paint(buf *buffer.Buffer) {
	s.mu.Lock()
	steps := s.steps
	current := s.current
	vertical := s.vertical
	s.mu.Unlock()

	if len(steps) == 0 {
		return
	}

	b := s.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	th := theme.Get()
	doneStyle := buffer.Style{Fg: th.Success}
	activeStyle := buffer.Style{Fg: th.Accent}
	pendingStyle := buffer.Style{Fg: th.Muted}
	connectorStyle := buffer.Style{Fg: th.Border}

	if vertical {
		s.paintVertical(buf, b, steps, current, doneStyle, activeStyle, pendingStyle, connectorStyle)
	} else {
		s.paintHorizontal(buf, b, steps, current, doneStyle, activeStyle, pendingStyle, connectorStyle)
	}
}

func (s *Stepper) paintHorizontal(buf *buffer.Buffer, b Rect, steps []StepperStep, current int,
	doneStyle, activeStyle, pendingStyle, connectorStyle buffer.Style) {

	x := b.X
	y := b.Y

	for i, step := range steps {
		var marker string
		var style buffer.Style
		if i < current {
			marker = "\u2714"
			style = doneStyle
		} else if i == current {
			marker = "\u25cf"
			style = activeStyle
		} else {
			marker = "\u25cb"
			style = pendingStyle
		}

		buf.DrawText(x, y, marker, style)
		x += 2

		title := step.Title
		availW := b.X + b.W - x - 3
		titleW := utf8.RuneCountInString(title)
		if titleW > availW && availW > 2 {
			title = truncateRunes(title, availW-1) + "\u2026"
		}
		x = buf.DrawText(x, y, title, style)

		if i < len(steps)-1 {
			x += 1
			connW := 3
			if x+connW > b.X+b.W {
				break
			}
			buf.DrawText(x, y, "\u2500\u2500\u2500", connectorStyle)
			x += connW
		}
	}
}

func (s *Stepper) paintVertical(buf *buffer.Buffer, b Rect, steps []StepperStep, current int,
	doneStyle, activeStyle, pendingStyle, connectorStyle buffer.Style) {

	x := b.X
	y := b.Y

	for i, step := range steps {
		if y >= b.Y+b.H {
			break
		}

		var marker string
		var style buffer.Style
		if i < current {
			marker = "\u2714"
			style = doneStyle
		} else if i == current {
			marker = "\u25cf"
			style = activeStyle
		} else {
			marker = "\u25cb"
			style = pendingStyle
		}

		buf.DrawText(x, y, marker, style)
		buf.DrawText(x+2, y, step.Title, style)
		y++

		if i < len(steps)-1 && y < b.Y+b.H {
			buf.DrawText(x, y, "\u2502", connectorStyle)
			y++
		}
	}
}

package component

import (
	"sync"
	"time"

	"github.com/topcheer/fluui/animation"
	"github.com/topcheer/fluui/internal/buffer"
)

// StatusIndicator is a component that shows a spinner animation alongside
// a text message, typically used to indicate a running background task.
type StatusIndicator struct {
	BaseComponent

	mu sync.RWMutex

	spinner *animation.Spinner
	message string
	running bool

	// accumulated time for spinner updates
	elapsed time.Duration

	style     buffer.Style
	spinStyle buffer.Style
}

// NewStatusIndicator creates a StatusIndicator with the default "dots" spinner.
func NewStatusIndicator() *StatusIndicator {
	s := &StatusIndicator{
		spinner:   animation.NewSpinner("dots"),
		running:   false,
		style:     buffer.DefaultStyle,
		spinStyle: buffer.DefaultStyle.AddFlags(buffer.Bold),
	}
	s.SetID(GenerateID("status"))
	return s
}

// SetMessage sets the status message displayed after the spinner.
func (s *StatusIndicator) SetMessage(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = msg
}

// Message returns the current status message.
func (s *StatusIndicator) Message() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.message
}

// Start activates the spinner animation.
func (s *StatusIndicator) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = true
}

// Stop deactivates the spinner animation.
func (s *StatusIndicator) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
}

// IsRunning returns whether the spinner is currently active.
func (s *StatusIndicator) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// SetSpinnerStyle sets the style for the spinner character.
func (s *StatusIndicator) SetSpinnerStyle(st buffer.Style) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.spinStyle = st
}

// SetStyle sets the style for the message text.
func (s *StatusIndicator) SetStyle(st buffer.Style) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.style = st
}

// SetSpinnerStyle allows choosing a named spinner style (dots, arc, arrow, bouncing).
func (s *StatusIndicator) SetSpinnerStyleName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.spinner = animation.NewSpinner(name)
}

// Update advances the spinner animation by delta. Returns true if the
// spinner is still running (always true while running).
func (s *StatusIndicator) Update(delta time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return false
	}
	s.spinner.Update(delta)
	return true
}

// CurrentFrame returns the current spinner character, or a space if stopped.
func (s *StatusIndicator) CurrentFrame() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.running {
		return " "
	}
	return s.spinner.Current()
}

// Measure returns the desired size: width = spinner(2) + message, height = 1.
func (s *StatusIndicator) Measure(cs Constraints) Size {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w := 2 + buffer.StringWidth(s.message)
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if w < 2 {
		w = 2
	}
	return Size{W: w, H: 1}
}

// Paint renders the spinner + message into the buffer.
func (s *StatusIndicator) Paint(buf *buffer.Buffer) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b := s.bounds
	if b.W <= 0 || b.H <= 0 {
		return
	}

	// Draw spinner character.
	spinChar := " "
	if s.running {
		spinChar = s.spinner.Current()
	}
	if spinChar != "" && b.W >= 1 {
		buf.DrawText(b.X, b.Y, spinChar, s.spinStyle)
	}

	// Draw message after spinner + 1 space gap.
	msgX := b.X + 2
	if s.message != "" && msgX < b.X+b.W {
		maxW := b.W - 2
		msg := s.message
		if buffer.StringWidth(msg) > maxW {
			// Truncate to fit
			runes := []rune(msg)
			truncW := 0
			endIdx := 0
			for i, r := range runes {
				rw := buffer.RuneWidth(r)
				if truncW+rw > maxW {
					break
				}
				truncW += rw
				endIdx = i + 1
			}
			msg = string(runes[:endIdx])
		}
		buf.DrawTextClamped(msgX, b.Y, msg, s.style)
	}
}

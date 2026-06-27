package animation

import "time"

// SpinnerFrames holds predefined spinner frame sequences.
var SpinnerFrames = map[string][]string{
	"dots":     {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	"arc":      {"◜", "◠", "◝", "◞", "◡", "◟"},
	"arrow":    {"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"},
	"bouncing": {"⠁", "⠂", "⠄", "⠂"},
}

// Spinner is a looping rotation animation. It never completes.
type Spinner struct {
	frames   []string
	current  int
	interval time.Duration
	elapsed  time.Duration
}

// NewSpinner creates a spinner using the named style. If the style is
// unknown, "dots" is used as a fallback.
func NewSpinner(style string) *Spinner {
	frames, ok := SpinnerFrames[style]
	if !ok {
		frames = SpinnerFrames["dots"]
	}
	return &Spinner{
		frames:   frames,
		interval: 100 * time.Millisecond,
	}
}

// Current returns the frame string currently shown.
func (s *Spinner) Current() string {
	if len(s.frames) == 0 {
		return ""
	}
	return s.frames[s.current]
}

// Update advances the spinner by delta. It always returns false because a
// spinner never finishes.
func (s *Spinner) Update(delta time.Duration) bool {
	s.elapsed += delta
	for s.elapsed >= s.interval {
		s.elapsed -= s.interval
		s.current = (s.current + 1) % len(s.frames)
	}
	return false
}

// Done returns a channel that is never closed — a spinner runs forever.
func (s *Spinner) Done() <-chan struct{} {
	return make(chan struct{})
}

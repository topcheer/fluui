package animation

import "time"

// FadeIn is a one-shot fade-in animation whose progress goes from 0 to 1.
type FadeIn struct {
	progress float64
	duration time.Duration
	elapsed  time.Duration
	done     chan struct{}
}

// NewFadeIn creates a FadeIn animation that completes after d.
func NewFadeIn(d time.Duration) *FadeIn {
	return &FadeIn{
		duration: d,
		done:     make(chan struct{}),
	}
}

// Update advances the fade-in by delta. Returns true once the animation
// has reached its full duration.
func (f *FadeIn) Update(delta time.Duration) bool {
	select {
	case <-f.done:
		return true
	default:
	}

	f.elapsed += delta
	if f.duration <= 0 {
		f.progress = 1
	} else {
		f.progress = float64(f.elapsed) / float64(f.duration)
	}
	if f.progress >= 1 {
		f.progress = 1
		close(f.done)
		return true
	}
	return false
}

// Progress returns the current fade progress in the range [0, 1].
func (f *FadeIn) Progress() float64 {
	return f.progress
}

// Done returns a channel that is closed when the fade-in completes.
func (f *FadeIn) Done() <-chan struct{} {
	return f.done
}

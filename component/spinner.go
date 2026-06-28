package component

import (
	"sync"
	"time"

	"github.com/topcheer/fluui/animation"
	"github.com/topcheer/fluui/internal/buffer"
)

// SpinnerStyle holds visual styling for the Spinner component.
type SpinnerStyle struct {
	Frame  buffer.Style
	Label  buffer.Style
	Prefix buffer.Style // optional prefix before label (e.g. status icon)
}

// DefaultSpinnerStyle returns a default spinner style.
func DefaultSpinnerStyle() SpinnerStyle {
	return SpinnerStyle{
		Frame:  buffer.Style{Fg: buffer.Color256Val(39)},
		Label:  buffer.Style{Fg: buffer.Color256Val(250)},
		Prefix: buffer.Style{Fg: buffer.Color256Val(241)},
	}
}

// Spinner is a loading animation component.
// It wraps animation.Spinner and renders an animated frame glyph + label text.
type Spinner struct {
	BaseComponent
	mu sync.RWMutex

	anim    *animation.Spinner
	label   string
	prefix  string
	running bool
	style   SpinnerStyle

	// frameStyle tracks the animation style name (e.g. "dots", "arc").
	frameStyle string
	frameIdx   int
	frames     []string
}

// NewSpinner creates a new Spinner with the given label and frame style.
func NewSpinner(label string) *Spinner {
	s := &Spinner{
		label:      label,
		running:    true,
		style:      DefaultSpinnerStyle(),
		frames:     animation.SpinnerFrames["dots"],
		frameStyle: "dots",
	}
	if len(s.frames) == 0 {
		s.frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	}
	s.anim = animation.NewSpinner("dots")
	s.SetID(GenerateID("spinner"))
	return s
}

// ─── Configuration ──────────────────────────────────────────────

// SetLabel updates the label text.
func (s *Spinner) SetLabel(label string) {
	s.mu.Lock()
	s.label = label
	s.mu.Unlock()
}

// Label returns the current label.
func (s *Spinner) Label() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.label
}

// SetPrefix sets optional prefix text (e.g. status icon).
func (s *Spinner) SetPrefix(prefix string) {
	s.mu.Lock()
	s.prefix = prefix
	s.mu.Unlock()
}

// Prefix returns the current prefix.
func (s *Spinner) Prefix() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.prefix
}

// SetStyle updates the visual style.
func (s *Spinner) SetStyle(style SpinnerStyle) {
	s.mu.Lock()
	s.style = style
	s.mu.Unlock()
}

// Style returns the current style.
func (s *Spinner) Style() SpinnerStyle {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.style
}

// SetFrameStyle changes the spinner frame set (e.g. "dots", "arc", "arrow").
func (s *Spinner) SetFrameStyle(frameStyle string) {
	s.mu.Lock()
	if frames, ok := animation.SpinnerFrames[frameStyle]; ok && len(frames) > 0 {
		s.frames = frames
		s.anim = animation.NewSpinner(frameStyle)
		s.frameIdx = 0
		s.frameStyle = frameStyle
	}
	s.mu.Unlock()
}

// FrameStyle returns the current frame set name.
// Since animation.Spinner doesn't expose its style name, we track it internally.
func (s *Spinner) FrameStyle() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.frameStyle
}

// ─── Animation control ──────────────────────────────────────────

// Start activates the spinner animation.
func (s *Spinner) Start() {
	s.mu.Lock()
	s.running = true
	s.mu.Unlock()
}

// Stop deactivates the spinner animation.
func (s *Spinner) Stop() {
	s.mu.Lock()
	s.running = false
	s.mu.Unlock()
}

// Running returns whether the spinner is currently animating.
func (s *Spinner) Running() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// Update advances the animation by delta. Returns true if the frame changed.
func (s *Spinner) Update(delta time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running || len(s.frames) == 0 {
		return false
	}
	if s.anim != nil {
		changed := s.anim.Update(delta)
		if changed {
			s.frameIdx = (s.frameIdx + 1) % len(s.frames)
		}
		return changed
	}
	return false
}

// CurrentFrame returns the current frame glyph.
func (s *Spinner) CurrentFrame() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.frames) == 0 {
		return "⠋"
	}
	return s.frames[s.frameIdx%len(s.frames)]
}

// SetFrameIndex sets the frame index directly (useful for tests).
func (s *Spinner) SetFrameIndex(idx int) {
	s.mu.Lock()
	if len(s.frames) > 0 {
		s.frameIdx = idx % len(s.frames)
		if s.frameIdx < 0 {
			s.frameIdx += len(s.frames)
		}
	}
	s.mu.Unlock()
}

// FrameIndex returns the current frame index.
func (s *Spinner) FrameIndex() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.frameIdx
}

// FrameCount returns the number of frames in the current frame set.
func (s *Spinner) FrameCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.frames)
}

// ─── Component interface ────────────────────────────────────────

// Measure calculates the preferred size.
func (s *Spinner) Measure(cs Constraints) Size {
	s.mu.RLock()
	defer s.mu.RUnlock()

	width := 3 // frame glyph + space
	if s.prefix != "" {
		width += len([]rune(s.prefix)) + 1
	}
	width += len([]rune(s.label))

	if cs.MaxWidth > 0 && width > cs.MaxWidth {
		width = cs.MaxWidth
	}
	if width < 3 {
		width = 3
	}

	return Size{W: width, H: 1}
}

// Paint renders the spinner into the buffer.
func (s *Spinner) Paint(buf *buffer.Buffer) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b := s.bounds
	if b.W < 1 || b.H < 1 {
		return
	}

	style := s.style
	x := b.X

	// Draw frame glyph
	if b.W >= 2 {
		frame := "·"
		if s.running && len(s.frames) > 0 {
			frame = s.frames[s.frameIdx%len(s.frames)]
		}
		for _, r := range frame {
			buf.SetCell(x, b.Y, buffer.NewCell(r, style.Frame))
			x++
		}
		// Space after frame
		buf.SetCell(x, b.Y, buffer.NewCell(' ', style.Frame))
		x++
	}

	// Draw prefix
	if s.prefix != "" && x < b.X+b.W {
		for _, r := range s.prefix {
			if x >= b.X+b.W {
				break
			}
			buf.SetCell(x, b.Y, buffer.NewCell(r, style.Prefix))
			x++
		}
		if x < b.X+b.W {
			buf.SetCell(x, b.Y, buffer.NewCell(' ', style.Prefix))
			x++
		}
	}

	// Draw label
	for _, r := range s.label {
		if x >= b.X+b.W {
			break
		}
		buf.SetCell(x, b.Y, buffer.NewCell(r, style.Label))
		x++
	}
}

// Children returns nil — Spinner has no children.
func (s *Spinner) Children() []Component { return nil }

// String returns a debug description.
func (s *Spinner) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return "Spinner(label=" + s.label + ",frameStyle=" + s.frameStyle + ")"
}

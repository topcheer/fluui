package component

import (
	"fmt"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── SliderRange: Dual-Thumb Range Slider ───
//
// SliderRange is a horizontal slider with two handles (low and high) for
// selecting a sub-range within [min, max]. Common in filter panels, price
// selectors, and data thresholds.
//
// Usage:
//
//	sr := NewSliderRange()                       // 0–100, low=0, high=100
//	sr.SetLow(20).SetHigh(80).SetStep(5)
//	sr.SetOnChange(func(lo, hi float64) { ... })
//	low, high := sr.Low(), sr.High()             // 20, 80

// RangeThumb identifies which handle is active.
type RangeThumb int

const (
	ThumbLow  RangeThumb = 0
	ThumbHigh RangeThumb = 1
)

// SliderRangeStyle holds visual styles for the range slider.
type SliderRangeStyle struct {
	Track     buffer.Style // unfilled track (outside low..high)
	Filled    buffer.Style // filled track (inside low..high)
	HandleLow buffer.Style // low handle
	HandleHi  buffer.Style // high handle
	Label     buffer.Style
	ValueText buffer.Style
}

// DefaultSliderRangeStyle returns sensible default styles.
func DefaultSliderRangeStyle() SliderRangeStyle {
	return SliderRangeStyle{
		Track:     buffer.Style{Fg: buffer.RGB(80, 80, 80)},
		Filled:    buffer.Style{Fg: buffer.Cyan},
		HandleLow: buffer.Style{Fg: buffer.Yellow, Flags: buffer.Bold},
		HandleHi:  buffer.Style{Fg: buffer.Green, Flags: buffer.Bold},
		Label:     buffer.Style{Fg: buffer.White},
		ValueText: buffer.Style{Fg: buffer.White, Flags: buffer.Bold},
	}
}

// SliderRange is a dual-thumb horizontal range slider.
type SliderRange struct {
	BaseComponent
	mu         sync.RWMutex
	min, max   float64
	low, high  float64
	step       float64
	active     RangeThumb
	style      SliderRangeStyle
	label      string
	showValues bool
	OnChange   func(low, high float64)
}

// NewSliderRange creates a range slider from 0 to 100, low=0, high=100, step=1.
func NewSliderRange() *SliderRange {
	s := &SliderRange{
		min:        0,
		max:        100,
		low:        0,
		high:       100,
		step:       1,
		active:     ThumbLow,
		style:      DefaultSliderRangeStyle(),
		showValues: true,
	}
	s.SetID(GenerateID("sliderrange"))
	return s
}

// NewSliderRangeWithBounds creates a range slider with explicit min, max, low, high, step.
func NewSliderRangeWithBounds(min, max, low, high, step float64) *SliderRange {
	s := NewSliderRange()
	s.min = min
	s.max = max
	s.step = step
	s.low = clampFloat(low, min, max)
	s.high = clampFloat(high, s.low, max)
	return s
}

// Low returns the current low value.
func (s *SliderRange) Low() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.low
}

// High returns the current high value.
func (s *SliderRange) High() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.high
}

// SetLow sets the low value, clamped to [min, high].
func (s *SliderRange) SetLow(v float64) *SliderRange {
	s.mu.Lock()
	old := s.low
	s.low = clampFloat(v, s.min, s.high)
	changed := s.low != old
	cb := s.OnChange
	s.mu.Unlock()
	if changed && cb != nil {
		cb(s.low, s.high)
	}
	return s
}

// SetHigh sets the high value, clamped to [low, max].
func (s *SliderRange) SetHigh(v float64) *SliderRange {
	s.mu.Lock()
	old := s.high
	s.high = clampFloat(v, s.low, s.max)
	changed := s.high != old
	cb := s.OnChange
	s.mu.Unlock()
	if changed && cb != nil {
		cb(s.low, s.high)
	}
	return s
}

// Min returns the minimum value.
func (s *SliderRange) Min() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.min
}

// Max returns the maximum value.
func (s *SliderRange) Max() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.max
}

// SetRange sets the min and max. Values are clamped if needed.
func (s *SliderRange) SetRange(min, max float64) *SliderRange {
	s.mu.Lock()
	s.min = min
	s.max = max
	if s.low < min {
		s.low = min
	}
	if s.high > max {
		s.high = max
	}
	if s.low > s.high {
		s.low, s.high = s.high, s.low
	}
	s.mu.Unlock()
	return s
}

// Step returns the step size.
func (s *SliderRange) Step() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.step
}

// SetStep sets the step size.
func (s *SliderRange) SetStep(step float64) *SliderRange {
	s.mu.Lock()
	if step > 0 {
		s.step = step
	}
	s.mu.Unlock()
	return s
}

// ActiveThumb returns the currently active thumb (ThumbLow or ThumbHigh).
func (s *SliderRange) ActiveThumb() RangeThumb {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active
}

// SetActiveThumb sets which handle responds to keyboard adjustments.
func (s *SliderRange) SetActiveThumb(t RangeThumb) *SliderRange {
	s.mu.Lock()
	s.active = t
	s.mu.Unlock()
	return s
}

// SetStyle sets the visual style.
func (s *SliderRange) SetStyle(style SliderRangeStyle) *SliderRange {
	s.mu.Lock()
	s.style = style
	s.mu.Unlock()
	return s
}

// Style returns the current style.
func (s *SliderRange) Style() SliderRangeStyle {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.style
}

// SetLabel sets an optional label shown above the track.
func (s *SliderRange) SetLabel(label string) *SliderRange {
	s.mu.Lock()
	s.label = label
	s.mu.Unlock()
	return s
}

// Label returns the label.
func (s *SliderRange) Label() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.label
}

// SetShowValues toggles whether low/high values are displayed.
func (s *SliderRange) SetShowValues(show bool) *SliderRange {
	s.mu.Lock()
	s.showValues = show
	s.mu.Unlock()
	return s
}

// ShowValues returns whether values are displayed.
func (s *SliderRange) ShowValues() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.showValues
}

// SetOnChange sets the callback fired when low or high changes.
func (s *SliderRange) SetOnChange(fn func(low, high float64)) *SliderRange {
	s.mu.Lock()
	s.OnChange = fn
	s.mu.Unlock()
	return s
}

// Increment increases the active thumb by one step.
func (s *SliderRange) Increment() {
	s.adjustActive(s.step)
}

// Decrement decreases the active thumb by one step.
func (s *SliderRange) Decrement() {
	s.adjustActive(-s.step)
}

func (s *SliderRange) adjustActive(delta float64) {
	s.mu.Lock()
	oldLow, oldHigh := s.low, s.high
	if s.active == ThumbLow {
		s.low = clampFloat(s.low+delta, s.min, s.high)
	} else {
		s.high = clampFloat(s.high+delta, s.low, s.max)
	}
	changed := s.low != oldLow || s.high != oldHigh
	cb := s.OnChange
	s.mu.Unlock()
	if changed && cb != nil {
		cb(s.low, s.high)
	}
}

// HandleKey processes keyboard input for the range slider.
func (s *SliderRange) HandleKey(key *term.KeyEvent) bool {
	if key == nil {
		return false
	}
	switch key.Key {
	case term.KeyLeft, term.KeyDown:
		s.Decrement()
		return true
	case term.KeyRight, term.KeyUp:
		s.Increment()
		return true
	case term.KeyTab:
		s.mu.Lock()
		if s.active == ThumbLow {
			s.active = ThumbHigh
		} else {
			s.active = ThumbLow
		}
		s.mu.Unlock()
		return true
	case term.KeyHome:
		s.mu.Lock()
		oldLow, oldHigh := s.low, s.high
		if s.active == ThumbLow {
			s.low = s.min
		} else {
			s.high = s.low
		}
		changed := s.low != oldLow || s.high != oldHigh
		cb := s.OnChange
		s.mu.Unlock()
		if changed && cb != nil {
			cb(s.low, s.high)
		}
		return true
	case term.KeyEnd:
		s.mu.Lock()
		oldLow, oldHigh := s.low, s.high
		if s.active == ThumbLow {
			s.low = s.high
		} else {
			s.high = s.max
		}
		changed := s.low != oldLow || s.high != oldHigh
		cb := s.OnChange
		s.mu.Unlock()
		if changed && cb != nil {
			cb(s.low, s.high)
		}
		return true
	default:
		if key.Key == term.KeyUnknown && key.Rune != 0 {
			switch key.Rune {
			case 'h':
				s.Decrement()
				return true
			case 'l':
				s.Increment()
				return true
			case 'k':
				s.Increment()
				return true
			case 'j':
				s.Decrement()
				return true
			}
		}
	}
	return false
}

// Measure computes the desired size.
func (s *SliderRange) Measure(cs Constraints) Size {
	h := 1
	if s.label != "" || s.showValues {
		h = 2
	}
	w := 30
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// SetBounds sets the component's bounds.
func (s *SliderRange) SetBounds(r Rect) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BaseComponent.SetBounds(r)
}

// ratioLowLocked returns the fill ratio of the low handle (caller holds lock).
func (s *SliderRange) ratioLowLocked() float64 {
	if s.max == s.min {
		return 0
	}
	return (s.low - s.min) / (s.max - s.min)
}

// ratioHighLocked returns the fill ratio of the high handle (caller holds lock).
func (s *SliderRange) ratioHighLocked() float64 {
	if s.max == s.min {
		return 1
	}
	return (s.high - s.min) / (s.max - s.min)
}

// Paint renders the range slider into the buffer.
func (s *SliderRange) Paint(buf *buffer.Buffer) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b := s.bounds
	if b.W <= 0 || b.H <= 0 {
		return
	}

	// Draw label/value row
	labelRow := 0
	if s.label != "" || s.showValues {
		y := b.Y
		if s.label != "" {
			buf.DrawText(b.X, y, s.label, s.style.Label)
		}
		if s.showValues {
			valText := formatSliderValue(s.low) + " - " + formatSliderValue(s.high)
			textW := buffer.StringWidth(valText)
			buf.DrawText(b.X+b.W-textW, y, valText, s.style.ValueText)
		}
		labelRow = 1
	}

	trackY := b.Y + labelRow
	if trackY >= b.Y+b.H {
		trackY = b.Y + b.H - 1
	}

	trackW := b.W
	if trackW <= 0 {
		return
	}

	lowX := int(float64(trackW) * s.ratioLowLocked())
	highX := int(float64(trackW) * s.ratioHighLocked())
	if lowX >= trackW {
		lowX = trackW - 1
	}
	if highX >= trackW {
		highX = trackW - 1
	}
	if lowX > highX {
		lowX = highX
	}

	// Draw track: outside range = track style, inside range = filled style
	for x := 0; x < trackW; x++ {
		var cell buffer.Cell
		if x >= lowX && x <= highX {
			cell = buffer.Cell{Rune: ' ', Fg: s.style.Filled.Fg, Bg: s.style.Filled.Fg}
		} else {
			cell = buffer.Cell{Rune: ' ', Fg: s.style.Track.Fg, Bg: s.style.Track.Fg}
		}
		buf.SetCell(b.X+x, trackY, cell)
	}

	// Draw handles
	buf.SetCell(b.X+lowX, trackY, buffer.Cell{Rune: '◀', Fg: s.style.HandleLow.Fg, Bg: s.style.HandleLow.Bg})
	buf.SetCell(b.X+highX, trackY, buffer.Cell{Rune: '▶', Fg: s.style.HandleHi.Fg, Bg: s.style.HandleHi.Bg})
}

// Children returns nil (leaf component).
func (s *SliderRange) Children() []Component {
	return nil
}

// String returns a string representation.
func (s *SliderRange) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fmt.Sprintf("SliderRange[%.1f-%.1f of %.1f-%.1f]", s.low, s.high, s.min, s.max)
}

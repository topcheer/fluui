package component

import (
	"fmt"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// SliderOrientation specifies horizontal or vertical slider.
type SliderOrientation int

const (
	SliderHorizontal SliderOrientation = 0
	SliderVertical   SliderOrientation = 1
)

// SliderStyle holds visual styles for the slider.
type SliderStyle struct {
	Track     buffer.Style
	Filled    buffer.Style
	Handle    buffer.Style
	Label     buffer.Style
	ValueText buffer.Style
}

// DefaultSliderStyle returns a sensible default style.
func DefaultSliderStyle() SliderStyle {
	return SliderStyle{
		Track:     buffer.Style{Fg: buffer.RGB(80, 80, 80)},
		Filled:    buffer.Style{Fg: buffer.Cyan},
		Handle:    buffer.Style{Fg: buffer.Yellow, Flags: buffer.Bold},
		Label:     buffer.Style{Fg: buffer.White},
		ValueText: buffer.Style{Fg: buffer.White, Flags: buffer.Bold},
	}
}

// Slider is a draggable slider for selecting a numeric value within a range.
// Supports horizontal and vertical orientations, keyboard navigation,
// and configurable step size.
type Slider struct {
	BaseComponent
	mu          sync.RWMutex
	min, max    float64
	value       float64
	step        float64
	orientation SliderOrientation
	style       SliderStyle
	label       string
	showValue   bool
	// OnChange is called whenever the value changes.
	OnChange func(value float64)
	// dragging tracks whether the handle is being dragged by the mouse.
	dragging bool
}

// NewSlider creates a new horizontal slider from 0 to 100, value 0, step 1.
func NewSlider() *Slider {
	s := &Slider{
		min:         0,
		max:         100,
		value:       0,
		step:        1,
		orientation: SliderHorizontal,
		style:       DefaultSliderStyle(),
		showValue:   true,
	}
	s.SetID(GenerateID("slider"))
	return s
}

// NewSliderWithRange creates a slider with custom min, max, and initial value.
func NewSliderWithRange(min, max, value, step float64) *Slider {
	s := NewSlider()
	s.min = min
	s.max = max
	s.value = value
	s.step = step
	return s
}

// Value returns the current value.
func (s *Slider) Value() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// SetValue sets the slider value (clamped to min/max).
func (s *Slider) SetValue(v float64) {
	s.mu.Lock()
	old := s.value
	s.value = clampFloat(v, s.min, s.max)
	changed := s.value != old
	cb := s.OnChange
	s.mu.Unlock()
	if changed && cb != nil {
		cb(s.value)
	}
}

// Min returns the minimum value.
func (s *Slider) Min() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.min
}

// Max returns the maximum value.
func (s *Slider) Max() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.max
}

// SetRange sets the min and max values. Value is clamped if needed.
func (s *Slider) SetRange(min, max float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.min = min
	s.max = max
	s.value = clampFloat(s.value, min, max)
}

// Step returns the step size.
func (s *Slider) Step() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.step
}

// SetStep sets the step size.
func (s *Slider) SetStep(step float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.step = step
}

// Orientation returns the slider orientation.
func (s *Slider) Orientation() SliderOrientation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.orientation
}

// SetOrientation sets the slider orientation.
func (s *Slider) SetOrientation(o SliderOrientation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orientation = o
}

// SetStyle sets the slider style.
func (s *Slider) SetStyle(style SliderStyle) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.style = style
}

// Style returns the current style.
func (s *Slider) Style() SliderStyle {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.style
}

// SetLabel sets an optional text label.
func (s *Slider) SetLabel(label string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.label = label
}

// Label returns the current label.
func (s *Slider) Label() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.label
}

// SetShowValue toggles whether the numeric value is displayed.
func (s *Slider) SetShowValue(show bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.showValue = show
}

// ShowValue returns whether the value is displayed.
func (s *Slider) ShowValue() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.showValue
}

// SetOnChange sets the callback fired when the value changes.
func (s *Slider) SetOnChange(fn func(value float64)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.OnChange = fn
}

// Increment increases the value by one step.
func (s *Slider) Increment() {
	s.mu.Lock()
	old := s.value
	s.value = clampFloat(s.value+s.step, s.min, s.max)
	changed := s.value != old
	cb := s.OnChange
	s.mu.Unlock()
	if changed && cb != nil {
		cb(s.value)
	}
}

// Decrement decreases the value by one step.
func (s *Slider) Decrement() {
	s.mu.Lock()
	old := s.value
	s.value = clampFloat(s.value-s.step, s.min, s.max)
	changed := s.value != old
	cb := s.OnChange
	s.mu.Unlock()
	if changed && cb != nil {
		cb(s.value)
	}
}

// IncrementBy increases the value by n steps.
func (s *Slider) IncrementBy(n int) {
	s.mu.Lock()
	old := s.value
	s.value = clampFloat(s.value+float64(n)*s.step, s.min, s.max)
	changed := s.value != old
	cb := s.OnChange
	s.mu.Unlock()
	if changed && cb != nil {
		cb(s.value)
	}
}

// Ratio returns the value as a fraction of the range (0.0 to 1.0).
func (s *Slider) Ratio() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.max == s.min {
		return 0
	}
	return (s.value - s.min) / (s.max - s.min)
}

// HandleKey processes keyboard input.
// Left/Down (or h/j): decrease by step
// Right/Up (or l/k): increase by step
// Home: set to min
// End: set to max
func (s *Slider) HandleKey(key *term.KeyEvent) bool {
	if key == nil {
		return false
	}
	switch key.Key {
	case term.KeyLeft:
		s.Decrement()
		return true
	case term.KeyRight:
		s.Increment()
		return true
	case term.KeyUp:
		s.Increment()
		return true
	case term.KeyDown:
		s.Decrement()
		return true
	case term.KeyHome:
		s.mu.Lock()
		old := s.value
		s.value = s.min
		cb := s.OnChange
		changed := s.value != old
		s.mu.Unlock()
		if changed && cb != nil {
			cb(s.value)
		}
		return true
	case term.KeyEnd:
		s.mu.Lock()
		old := s.value
		s.value = s.max
		cb := s.OnChange
		changed := s.value != old
		s.mu.Unlock()
		if changed && cb != nil {
			cb(s.value)
		}
		return true
	default:
		// Vim-style h/j/k/l
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
func (s *Slider) Measure(cs Constraints) Size {
	s.mu.RLock()
	defer s.mu.RUnlock()

	h := 1
	if s.label != "" || s.showValue {
		h = 2
	}
	if s.orientation == SliderVertical {
		// Vertical slider: narrow width, taller height
		w := 3
		if cs.MaxWidth > 0 && w > cs.MaxWidth {
			w = cs.MaxWidth
		}
		vh := 5
		if cs.MaxHeight > 0 && vh > cs.MaxHeight {
			vh = cs.MaxHeight
		}
		if vh < 3 {
			vh = 3
		}
		return Size{W: w, H: vh}
	}

	// Horizontal slider: wide, 1-2 rows
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
func (s *Slider) SetBounds(r Rect) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.BaseComponent.SetBounds(r)
}

// Paint renders the slider into the buffer.
func (s *Slider) Paint(buf *buffer.Buffer) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b := s.bounds
	if b.W <= 0 || b.H <= 0 {
		return
	}

	// Draw label/value row if present
	labelRow := 0
	if s.label != "" || s.showValue {
		y := b.Y
		if s.label != "" {
			buf.DrawText(b.X, y, s.label, s.style.Label)
		}
		if s.showValue {
			valText := formatSliderValue(s.value)
			textW := buffer.StringWidth(valText)
			buf.DrawText(b.X+b.W-textW, y, valText, s.style.ValueText)
		}
		labelRow = 1
	}

	if s.orientation == SliderVertical {
		s.paintVertical(buf, b, labelRow)
	} else {
		s.paintHorizontal(buf, b, labelRow)
	}
}

func (s *Slider) paintHorizontal(buf *buffer.Buffer, b Rect, labelRow int) {
	trackY := b.Y + labelRow
	if trackY >= b.Y+b.H {
		trackY = b.Y + b.H - 1
	}

	trackW := b.W
	if trackW <= 0 {
		return
	}

	// Fill ratio
	filledW := int(float64(trackW) * s.ratioLocked())
	if filledW > trackW {
		filledW = trackW
	}

	// Draw track
	for x := 0; x < trackW; x++ {
		ch := buffer.Cell{Rune: ' ', Fg: s.style.Track.Fg, Bg: s.style.Track.Fg}
		if x < filledW {
			ch = buffer.Cell{Rune: ' ', Fg: s.style.Filled.Fg, Bg: s.style.Filled.Fg}
		}
		buf.SetCell(b.X+x, trackY, ch)
	}

	// Draw handle
	handleX := b.X + filledW
	if handleX >= b.X+trackW {
		handleX = b.X + trackW - 1
	}
	buf.SetCell(handleX, trackY, buffer.Cell{Rune: '█', Fg: s.style.Handle.Fg, Bg: s.style.Handle.Bg})
}

func (s *Slider) paintVertical(buf *buffer.Buffer, b Rect, labelRow int) {
	trackH := b.H - labelRow
	if trackH <= 0 {
		trackH = 1
	}

	trackX := b.X
	filledH := int(float64(trackH) * s.ratioLocked())
	if filledH > trackH {
		filledH = trackH
	}

	// Draw vertical track (bottom to top)
	for y := 0; y < trackH; y++ {
		posY := b.Y + b.H - 1 - y
		ch := buffer.Cell{Rune: ' ', Fg: s.style.Track.Fg, Bg: s.style.Track.Fg}
		if y < filledH {
			ch = buffer.Cell{Rune: ' ', Fg: s.style.Filled.Fg, Bg: s.style.Filled.Fg}
		}
		buf.SetCell(trackX, posY, ch)
	}

	// Draw handle at top of filled portion
	handleY := b.Y + b.H - 1 - filledH
	if handleY < b.Y+labelRow {
		handleY = b.Y + labelRow
	}
	buf.SetCell(trackX, handleY, buffer.Cell{Rune: '█', Fg: s.style.Handle.Fg, Bg: s.style.Handle.Bg})
}

// ratioLocked returns the fill ratio (caller must hold lock).
func (s *Slider) ratioLocked() float64 {
	if s.max == s.min {
		return 0
	}
	return (s.value - s.min) / (s.max - s.min)
}

// Children returns nil (leaf component).
func (s *Slider) Children() []Component {
	return nil
}

// String returns a string representation.
func (s *Slider) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fmt.Sprintf("Slider[%.1f/%.1f-%.1f]", s.value, s.min, s.max)
}

// SetFromRatio sets the value based on a 0.0-1.0 ratio.
func (s *Slider) SetFromRatio(ratio float64) {
	s.mu.Lock()
	old := s.value
	ratio = clampFloat(ratio, 0, 1)
	s.value = clampFloat(s.min+ratio*(s.max-s.min), s.min, s.max)
	changed := s.value != old
	cb := s.OnChange
	s.mu.Unlock()
	if changed && cb != nil {
		cb(s.value)
	}
}

// clampFloat clamps v to the [min, max] range.
func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// formatSliderValue formats a float value for display.
func formatSliderValue(v float64) string {
	// Show decimals only if needed
	if v == float64(int(v)) {
		return fmt.Sprintf("%.0f", v)
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", v), "0"), ".")
}

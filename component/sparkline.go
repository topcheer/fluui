package component

import (
	"math"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// Sparkline bar characters from low to high (8 levels).
var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// ColorMode controls how sparkline bars are colored.
type ColorMode int

const (
	// ColorSingle uses a single color for all bars.
	ColorSingle ColorMode = iota
	// ColorGradient maps each bar's height to a color from low (green) to high (red).
	ColorGradient
	// ColorValue maps each bar's value to a color threshold.
	ColorValue
)

// Sparkline is a compact inline chart that visualizes a time series
// using Unicode block characters (▁▂▃▄▅▆▇█).
type Sparkline struct {
	BaseComponent
	mu sync.RWMutex

	data     []float64
	max      float64 // auto-computed if AutoScale is true
	min      float64 // auto-computed if AutoScale is true
	autoScale bool   // if true, recompute min/max on each render

	colorMode ColorMode
	fgColor   buffer.Color // for ColorSingle mode
	style     buffer.Style

	// Y-axis range overrides (0 = auto)
	YMin float64
	YMax float64

	// Label shown to the right of the sparkline
	label string

	// ShowMinMax appends "min..max" after the label
	showMinMax bool

	// scrollX for horizontal scrolling when data exceeds width
	scrollX int
}

// NewSparkline creates a Sparkline with default settings.
func NewSparkline() *Sparkline {
	s := &Sparkline{
		autoScale: true,
		colorMode: ColorSingle,
		fgColor:   buffer.NamedColor(buffer.NamedGreen),
		style:     buffer.Style{},
	}
	s.SetID(GenerateID("sparkline"))
	return s
}

// SetData replaces all data points.
func (s *Sparkline) SetData(data []float64) {
	s.mu.Lock()
	s.data = make([]float64, len(data))
	copy(s.data, data)
	if s.autoScale {
		s.recomputeRange()
	}
	s.mu.Unlock()
}

// Data returns a copy of the current data points.
func (s *Sparkline) Data() []float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]float64, len(s.data))
	copy(result, s.data)
	return result
}

// Push appends a single data point, optionally trimming to maxPoints.
func (s *Sparkline) Push(value float64, maxPoints int) {
	s.mu.Lock()
	s.data = append(s.data, value)
	if maxPoints > 0 && len(s.data) > maxPoints {
		s.data = s.data[len(s.data)-maxPoints:]
	}
	if s.autoScale {
		s.recomputeRange()
	}
	s.mu.Unlock()
}

// Clear removes all data points.
func (s *Sparkline) Clear() {
	s.mu.Lock()
	s.data = nil
	s.min = 0
	s.max = 0
	s.scrollX = 0
	s.mu.Unlock()
}

// SetColorMode sets the coloring strategy.
func (s *Sparkline) SetColorMode(mode ColorMode) {
	s.mu.Lock()
	s.colorMode = mode
	s.mu.Unlock()
}

// SetColor sets the foreground color (used for ColorSingle mode).
func (s *Sparkline) SetColor(c buffer.Color) {
	s.mu.Lock()
	s.fgColor = c
	s.mu.Unlock()
}

// SetStyle sets the base text style for labels.
func (s *Sparkline) SetStyle(st buffer.Style) {
	s.mu.Lock()
	s.style = st
	s.mu.Unlock()
}

// SetAutoScale controls whether min/max are recomputed automatically.
func (s *Sparkline) SetAutoScale(auto bool) {
	s.mu.Lock()
	s.autoScale = auto
	if auto {
		s.recomputeRange()
	}
	s.mu.Unlock()
}

// SetLabel sets the text label shown to the right of the sparkline.
func (s *Sparkline) SetLabel(label string) {
	s.mu.Lock()
	s.label = label
	s.mu.Unlock()
}

// Label returns the current label.
func (s *Sparkline) Label() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.label
}

// SetShowMinMax controls whether min..max values are displayed.
func (s *Sparkline) SetShowMinMax(show bool) {
	s.mu.Lock()
	s.showMinMax = show
	s.mu.Unlock()
}

// SetScrollX sets the horizontal scroll offset.
func (s *Sparkline) SetScrollX(x int) {
	s.mu.Lock()
	if x < 0 {
		x = 0
	}
	s.scrollX = x
	s.mu.Unlock()
}

// ScrollY returns the current horizontal scroll offset.
func (s *Sparkline) ScrollX() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scrollX
}

// Min returns the current minimum value.
func (s *Sparkline) Min() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.min
}

// Max returns the current maximum value.
func (s *Sparkline) Max() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.max
}

// Count returns the number of data points.
func (s *Sparkline) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// recomputeRange recalculates min/max from data. Caller must hold lock.
func (s *Sparkline) recomputeRange() {
	if len(s.data) == 0 {
		s.min = 0
		s.max = 0
		return
	}
	s.min = s.data[0]
	s.max = s.data[0]
	for _, v := range s.data {
		if v < s.min {
			s.min = v
		}
		if v > s.max {
			s.max = v
		}
	}
	// If YMin/YMax overrides are set, apply them
	if s.YMin != 0 || s.YMax != 0 {
		if s.YMin < s.min {
			s.min = s.YMin
		}
		if s.YMax > s.max {
			s.max = s.YMax
		}
	}
	// Prevent division by zero when all values are equal
	if s.max == s.min {
		s.max = s.min + 1
	}
}

// valueToBar maps a value to a sparkline bar character index (0-7).
func (s *Sparkline) valueToBar(val float64) int {
	if len(s.data) == 0 || s.max <= s.min {
		return 0
	}
	ratio := (val - s.min) / (s.max - s.min)
	idx := int(ratio * float64(len(sparkChars)))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sparkChars) {
		idx = len(sparkChars) - 1
	}
	return idx
}

// gradientColor returns a color based on bar height (0-7).
// Low values = green, mid = yellow, high = red.
func sparkGradientColor(level int) buffer.Color {
	switch {
	case level < 2:
		return buffer.NamedColor(buffer.NamedGreen)
	case level < 4:
		return buffer.NamedColor(buffer.NamedYellow)
	case level < 6:
		return buffer.NamedColor(buffer.NamedBrightYellow)
	default:
		return buffer.NamedColor(buffer.NamedRed)
	}
}

// sparkValueColor returns a color based on absolute value thresholds.
func sparkValueColor(val, min, max float64) buffer.Color {
	if max <= min {
		return buffer.NamedColor(buffer.NamedGreen)
	}
	ratio := (val - min) / (max - min)
	switch {
	case ratio < 0.25:
		return buffer.NamedColor(buffer.NamedGreen)
	case ratio < 0.5:
		return buffer.NamedColor(buffer.NamedYellow)
	case ratio < 0.75:
		return buffer.NamedColor(buffer.NamedBrightYellow)
	default:
		return buffer.NamedColor(buffer.NamedRed)
	}
}

// formatFloat formats a float for display, trimming unnecessary precision.
func formatSparkFloat(f float64) string {
	if f == math.Floor(f) && !math.IsInf(f, 0) {
		return formatInt(int(f))
	}
	return formatFloat(f)
}

func formatInt(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var digits [20]byte
	pos := len(digits)
	for i > 0 {
		pos--
		digits[pos] = byte('0' + i%10)
		i /= 10
	}
	s := string(digits[pos:])
	if neg {
		s = "-" + s
	}
	return s
}

func formatFloat(f float64) string {
	// Simple float formatting with 1 decimal precision
	whole := int(f)
	frac := int(math.Abs((f - float64(whole)) * 10))
	return formatInt(whole) + "." + string(rune('0'+frac))
}

// --- Component interface ---

// Measure returns the desired size: width = data count + label width, height = 1.
func (s *Sparkline) Measure(cs Constraints) Size {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w := len(s.data)
	if s.label != "" {
		w += 1 + len(s.label) // space + label
	}
	if s.showMinMax {
		w += 1 + formatSparkFloatLen(s.min) + 4 + formatSparkFloatLen(s.max) // " min..max"
	}
	h := 1
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if w < 1 {
		w = 1
	}
	return Size{W: w, H: h}
}

func formatSparkFloatLen(f float64) int {
	return len(formatSparkFloat(f))
}

// Paint renders the sparkline into the buffer.
func (s *Sparkline) Paint(buf *buffer.Buffer) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bounds := s.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	// Draw sparkline bars
	start := s.scrollX
	end := start + bounds.W
	if end > len(s.data) {
		end = len(s.data)
	}

	for i := start; i < end; i++ {
		if x >= maxX {
			break
		}
		level := s.valueToBar(s.data[i])
		ch := sparkChars[level]

		var fg buffer.Color
		switch s.colorMode {
		case ColorGradient:
			fg = sparkGradientColor(level)
		case ColorValue:
			fg = sparkValueColor(s.data[i], s.min, s.max)
		default:
			fg = s.fgColor
		}

		buf.SetCell(x, y, buffer.Cell{
			Rune:  ch,
			Width: 1,
			Fg:    fg,
			Bg:    s.style.Bg,
			Flags: s.style.Flags,
		})
		x++
	}

	// Draw label if there's room
	if s.label != "" && x < maxX {
		// Space before label
		buf.SetCell(x, y, buffer.Cell{
			Rune:  ' ',
			Width: 1,
			Fg:    s.style.Fg,
			Bg:    s.style.Bg,
			Flags: s.style.Flags,
		})
		x++
		for _, r := range s.label {
			if x >= maxX {
				break
			}
			buf.SetCell(x, y, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    s.style.Fg,
				Bg:    s.style.Bg,
				Flags: s.style.Flags,
			})
			x++
		}
	}

	// Draw min..max if enabled
	if s.showMinMax && x+4 < maxX {
		minStr := formatSparkFloat(s.min)
		maxStr := formatSparkFloat(s.max)
		suffix := " " + minStr + ".." + maxStr
		for _, r := range suffix {
			if x >= maxX {
				break
			}
			buf.SetCell(x, y, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    s.style.Fg,
				Bg:    s.style.Bg,
				Flags: s.style.Flags,
			})
			x++
		}
	}
}

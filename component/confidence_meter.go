package component

import (
	"strconv"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// ConfidenceThreshold defines how confidence values map to colors.
type ConfidenceThreshold struct {
	Min   float64 // 0.0-1.0
	Color buffer.Color
}

// ConfidenceMeter renders an AI model confidence indicator as a colored
// progress bar with percentage. Designed for displaying model certainty
// in AI chat UIs, tool call results, and classification outputs.
//
// Default thresholds:
//   - < 0.4: red (low confidence)
//   - 0.4-0.7: yellow (medium)
//   - > 0.7: green (high)
//
// Thread-safe.
type ConfidenceMeter struct {
	BaseComponent
	mu sync.RWMutex

	value     float64 // 0.0-1.0
	label     string
	barWidth  int
	showPct   bool
	showLabel bool
	thresholds []ConfidenceThreshold
	customColor buffer.Color // zero = auto from thresholds
}

// defaultConfidenceThresholds returns the standard red/yellow/green thresholds.
func defaultConfidenceThresholds() []ConfidenceThreshold {
	t := theme.Get()
	return []ConfidenceThreshold{
		{Min: 0.7, Color: t.Success},
		{Min: 0.4, Color: t.Warning},
		{Min: 0.0, Color: t.Error},
	}
}

// NewConfidenceMeter creates a confidence meter with the given value (0.0-1.0).
func NewConfidenceMeter(value float64) *ConfidenceMeter {
	return &ConfidenceMeter{
		BaseComponent: BaseComponent{id: GenerateID("confidence")},
		value:         clampFloat(value, 0, 1),
		barWidth:      12,
		showPct:       true,
		showLabel:     true,
		thresholds:    defaultConfidenceThresholds(),
	}
}

// Value returns the current confidence value (0.0-1.0).
func (c *ConfidenceMeter) Value() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

// SetValue sets the confidence value. Clamped to [0.0, 1.0].
func (c *ConfidenceMeter) SetValue(v float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = clampFloat(v, 0, 1)
}

// Label returns the display label.
func (c *ConfidenceMeter) Label() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.label
}

// SetLabel sets the display label (e.g., "Confidence", "Certainty").
func (c *ConfidenceMeter) SetLabel(s string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.label = s
}

// BarWidth returns the bar width.
func (c *ConfidenceMeter) BarWidth() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.barWidth
}

// SetBarWidth sets the bar width (default 12).
func (c *ConfidenceMeter) SetBarWidth(w int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if w < 1 {
		w = 1
	}
	c.barWidth = w
}

// SetShowPct toggles percentage display.
func (c *ConfidenceMeter) SetShowPct(b bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.showPct = b
}

// ShowPct returns whether percentage is displayed.
func (c *ConfidenceMeter) ShowPct() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.showPct
}

// SetShowLabel toggles label display.
func (c *ConfidenceMeter) SetShowLabel(b bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.showLabel = b
}

// ShowLabel returns whether label is displayed.
func (c *ConfidenceMeter) ShowLabel() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.showLabel
}

// SetColor overrides the bar color. Pass buffer.Color{} (zero) for auto.
func (c *ConfidenceMeter) SetColor(col buffer.Color) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.customColor = col
}

// resolveColorLocked returns the effective bar color.
func (c *ConfidenceMeter) resolveColorLocked() buffer.Color {
	if c.customColor.Type != buffer.ColorNone {
		return c.customColor
	}
	for _, th := range c.thresholds {
		if c.value >= th.Min {
			return th.Color
		}
	}
	return theme.Get().Muted
}

// Measure returns the preferred size (always 1 row).
func (c *ConfidenceMeter) Measure(cs Constraints) Size {
	c.mu.RLock()
	defer c.mu.RUnlock()

	w := c.measureWidthLocked()
	h := 1

	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.HasHeight() && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return Size{W: w, H: h}
}

func (c *ConfidenceMeter) measureWidthLocked() int {
	w := 0
	if c.showLabel && c.label != "" {
		w += len(c.label) + 1 // "Label "
	}
	w += c.barWidth
	if c.showPct {
		w += 4 // " NN%"
	}
	return w
}

// Paint draws the confidence meter. Zero allocations.
func (c *ConfidenceMeter) Paint(buf *buffer.Buffer) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	bounds := c.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W
	muted := buffer.Style{Fg: theme.Get().Muted}

	// Draw label
	if c.showLabel && c.label != "" {
		x = buf.DrawText(x, y, c.label+" ", muted)
	}

	// Draw bar
	color := c.resolveColorLocked()
	filled := int(c.value * float64(c.barWidth))
	if filled > c.barWidth {
		filled = c.barWidth
	}
	if c.value > 0 && filled == 0 {
		filled = 1
	}

	emptyColor := theme.Get().Muted
	barStyle := buffer.Style{Fg: color}
	emptyStyle := buffer.Style{Fg: emptyColor}

	for i := 0; i < c.barWidth && x < maxX; i++ {
		if i < filled {
			buf.SetCell(x, y, buffer.Cell{Rune: '\u2588', Width: 1, Fg: barStyle.Fg})
		} else {
			buf.SetCell(x, y, buffer.Cell{Rune: '\u2591', Width: 1, Fg: emptyStyle.Fg})
		}
		x++
	}

	// Draw percentage
	if c.showPct && x < maxX {
		if x < maxX {
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1})
			x++
		}
		if x < maxX {
			var pb [8]byte
			pbs := pb[:0]
			pct := int(c.value * 100)
			pbs = strconv.AppendInt(pbs, int64(pct), 10)
			pbs = append(pbs, '%')
			pctStyle := buffer.Style{Fg: color, Flags: buffer.Bold}
			x = buf.DrawBytes(x, y, pbs, pctStyle)
		}
	}
}


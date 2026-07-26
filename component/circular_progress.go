package component

import (
	"strconv"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// CircularProgressStyle controls the visual rendering style.
type CircularProgressStyle int

const (
	// ProgressStyleRing renders a ring using Unicode arcs.
	ProgressStyleRing CircularProgressStyle = iota
	// ProgressStyleDots renders using dot characters (●○).
	ProgressStyleDots
	// ProgressStyleBlock renders using block characters (▰▱).
	ProgressStyleBlock
)

// CircularProgress renders a circular/ring progress indicator.
// Useful for showing completion percentage in a compact space.
// Renders as a 1-row inline indicator: ◴ 75% or ●●●○○ 60%.
//
// Thread-safe.
type CircularProgress struct {
	BaseComponent
	mu       sync.RWMutex
	value    float64 // 0.0–1.0
	label    string
	style    CircularProgressStyle
	barW     int
	customFg buffer.Color
}

// NewCircularProgress creates a circular progress with value (0.0–1.0).
func NewCircularProgress(value float64) *CircularProgress {
	return &CircularProgress{
		BaseComponent: BaseComponent{id: GenerateID("cprogress")},
		value:         clampFloat(value, 0, 1),
		style:         ProgressStyleRing,
		barW:          5,
	}
}

func (c *CircularProgress) Value() float64 { c.mu.RLock(); defer c.mu.RUnlock(); return c.value }
func (c *CircularProgress) SetValue(v float64) { c.mu.Lock(); defer c.mu.Unlock(); c.value = clampFloat(v, 0, 1) }

func (c *CircularProgress) Label() string { c.mu.RLock(); defer c.mu.RUnlock(); return c.label }
func (c *CircularProgress) SetLabel(s string) { c.mu.Lock(); defer c.mu.Unlock(); c.label = s }

func (c *CircularProgress) Style() CircularProgressStyle { c.mu.RLock(); defer c.mu.RUnlock(); return c.style }
func (c *CircularProgress) SetStyle(s CircularProgressStyle) { c.mu.Lock(); defer c.mu.Unlock(); c.style = s }

func (c *CircularProgress) BarWidth() int { c.mu.RLock(); defer c.mu.RUnlock(); return c.barW }
func (c *CircularProgress) SetBarWidth(w int) {
	c.mu.Lock(); defer c.mu.Unlock()
	if w < 1 { w = 1 }
	c.barW = w
}

func (c *CircularProgress) SetColor(col buffer.Color) { c.mu.Lock(); defer c.mu.Unlock(); c.customFg = col }

func (c *CircularProgress) resolveColorLocked() buffer.Color {
	if c.customFg.Type != buffer.ColorNone { return c.customFg }
	t := theme.Get()
	if c.value >= 0.8 { return t.Success }
	if c.value >= 0.4 { return t.Warning }
	return t.Error
}

// Measure returns the preferred size.
func (c *CircularProgress) Measure(cs Constraints) Size {
	c.mu.RLock()
	defer c.mu.RUnlock()
	w := 0
	if c.label != "" { w += len(c.label) + 1 }
	w += 2 // icon/space
	w += c.barW
	w += 4 // " NN%"
	h := 1
	if cs.HasWidth() && w > cs.MaxWidth { w = cs.MaxWidth }
	if w < 1 { w = 1 }
	if h < 1 { h = 1 }
	return Size{W: w, H: h}
}

// Paint draws the circular progress. Zero allocations.
func (c *CircularProgress) Paint(buf *buffer.Buffer) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	bounds := c.bounds
	if bounds.W <= 0 || bounds.H <= 0 { return }

	t := theme.Get()
	color := c.resolveColorLocked()
	muted := buffer.Style{Fg: t.Muted}
	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	// Label
	if c.label != "" {
		x = buf.DrawText(x, y, c.label+" ", muted)
	}

	// Style-specific rendering
	switch c.style {
	case ProgressStyleDots:
		filled := int(c.value * float64(c.barW))
		if c.value > 0 && filled == 0 { filled = 1 }
		for i := 0; i < c.barW && x < maxX; i++ {
			r := '○'
			fg := muted.Fg
			if i < filled { r = '●'; fg = color }
			buf.SetCell(x, y, buffer.Cell{Rune: r, Width: 1, Fg: fg})
			x++
		}
	case ProgressStyleBlock:
		filled := int(c.value * float64(c.barW))
		if c.value > 0 && filled == 0 { filled = 1 }
		for i := 0; i < c.barW && x < maxX; i++ {
			r := '▱'
			fg := muted.Fg
			if i < filled { r = '▰'; fg = color }
			buf.SetCell(x, y, buffer.Cell{Rune: r, Width: 1, Fg: fg})
			x++
		}
	default: // ProgressStyleRing
		// Use quadrant arc based on percentage
		var icon string
		switch {
		case c.value >= 0.875: icon = "OKIE" // use ○ + checkmark-like
		case c.value >= 0.75: icon = "◔" // 3/4
		case c.value >= 0.625: icon = "◔"
		case c.value >= 0.5: icon = "◑" // half
		case c.value >= 0.375: icon = "◑"
		case c.value >= 0.25: icon = "◒" // quarter
		case c.value >= 0.125: icon = "◓"
		case c.value > 0: icon = "◔"
		default: icon = "○" // empty
		}
		if c.value >= 1.0 { icon = "✓" }
		style := buffer.Style{Fg: color, Flags: buffer.Bold}
		x = buf.DrawText(x, y, icon+" ", style)
	}

	// Percentage
	if x < maxX {
		var pb [8]byte
		pbs := pb[:0]
		pbs = strconv.AppendInt(pbs, int64(c.value*100), 10)
		pbs = append(pbs, '%')
		pctStyle := buffer.Style{Fg: color, Flags: buffer.Bold}
		buf.DrawBytes(x, y, pbs, pctStyle)
	}
}

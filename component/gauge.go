package component

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// GaugeOrientation determines the direction of a linear gauge.
type GaugeOrientation int

const (
	// GaugeHorizontal renders the bar left-to-right.
	GaugeHorizontal GaugeOrientation = iota
	// GaugeVertical renders the bar bottom-to-top.
	GaugeVertical
)

// Threshold defines a color band for the gauge.
// When the current ratio falls within [Low, High), the associated color is used.
type Threshold struct {
	Low   float64 // 0.0–1.0 inclusive
	High  float64 // 0.0–1.0 exclusive (or 1.0 for the top band)
	Color buffer.Color
}

// Gauge is a component that renders a gauge (linear or radial) showing a
// value relative to a min/max range. Linear gauges can be horizontal or
// vertical. Radial gauges render as an arc.
//
// Gauges support color thresholds: different colors for different value bands
// (e.g., green < 0.6, yellow < 0.85, red ≥ 0.85).
type Gauge struct {
	BaseComponent

	mu sync.RWMutex

	value    float64 // current value
	min      float64
	max      float64
	label    string
	showVal  bool          // show "value/max" suffix
	orient   GaugeOrientation
	radial   bool          // render as arc instead of bar
	style    buffer.Style  // base style for the track
	fillChar rune          // character used for the filled portion
	emptChar rune          // character used for the unfilled portion

	// thresholds define color bands. If empty, a single gradient is used.
	thresholds []Threshold
}

// NewGauge creates a horizontal linear gauge with range [0, 100], value 0.
func NewGauge() *Gauge {
	g := &Gauge{
		value:    0,
		min:      0,
		max:      100,
		orient:   GaugeHorizontal,
		radial:   false,
		style:    buffer.DefaultStyle,
		fillChar: '█',
		emptChar: '░',
		showVal:  true,
	}
	g.SetID(GenerateID("gauge"))
	return g
}

// SetValue sets the current gauge value (clamped to [min, max]).
func (g *Gauge) SetValue(v float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = clamp(v, g.min, g.max)
}

// Value returns the current value.
func (g *Gauge) Value() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.value
}

// SetRange sets the min and max of the gauge. If min >= max, the call is ignored.
func (g *Gauge) SetRange(min, max float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if min >= max {
		return
	}
	g.min = min
	g.max = max
	// Re-clamp the current value.
	g.value = clamp(g.value, min, max)
}

// SetLabel sets the gauge's optional text label.
func (g *Gauge) SetLabel(label string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.label = label
}

// SetOrientation sets the bar direction (horizontal or vertical).
func (g *Gauge) SetOrientation(o GaugeOrientation) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.orient = o
	g.radial = false
}

// SetRadial switches to radial (arc) rendering mode.
func (g *Gauge) SetRadial(r bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.radial = r
}

// SetShowValue toggles whether "value/max" text is shown.
func (g *Gauge) SetShowValue(show bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.showVal = show
}

// SetStyle sets the base track style.
func (g *Gauge) SetStyle(s buffer.Style) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.style = s
}

// SetThresholds sets the color bands. Pass nil to revert to gradient coloring.
func (g *Gauge) SetThresholds(t []Threshold) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.thresholds = t
}

// SetFillChar overrides the default fill character.
func (g *Gauge) SetFillChar(r rune) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.fillChar = r
}

// SetEmptyChar overrides the default empty character.
func (g *Gauge) SetEmptyChar(r rune) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.emptChar = r
}

// Ratio returns the fill ratio (0.0–1.0), clamped.
func (g *Gauge) Ratio() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.max <= g.min {
		return 0
	}
	return clamp((g.value-g.min)/(g.max-g.min), 0, 1)
}

// Measure returns the desired size. Horizontal: full width × 1 (+ label).
// Vertical: 1 × full height (+ label line). Radial: min 7×5.
func (g *Gauge) Measure(cs Constraints) Size {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.radial {
		w := 7
		h := 5
		if g.label != "" {
			h += 2
		}
		if cs.MaxWidth > 0 && w > cs.MaxWidth {
			w = cs.MaxWidth
		}
		if cs.MaxHeight > 0 && h > cs.MaxHeight {
			h = cs.MaxHeight
		}
		return Size{W: w, H: h}
	}

	if g.orient == GaugeVertical {
		w := 1
		h := cs.MaxHeight
		if h <= 0 {
			h = 10
		}
		totalH := h
		if g.label != "" {
			totalH++ // label above
		}
		if g.showVal {
			totalH++ // value below
		}
		return Size{W: w, H: totalH}
	}

	w := cs.MaxWidth
	if w <= 0 {
		w = 40
	}
	h := 1
	if g.label != "" {
		h++
	}
	return Size{W: w, H: h}
}

// Paint renders the gauge into the buffer.
func (g *Gauge) Paint(buf *buffer.Buffer) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	b := g.bounds
	if b.W <= 0 || b.H <= 0 {
		return
	}

	if g.radial {
		g.paintRadial(buf, b)
		return
	}

	if g.orient == GaugeVertical {
		g.paintVertical(buf, b)
		return
	}
	g.paintHorizontal(buf, b)
}

// paintHorizontal renders a left-to-right bar.
func (g *Gauge) paintHorizontal(buf *buffer.Buffer, b Rect) {
	ratio := g.ratioLocked()
	barY := b.Y
	if g.label != "" {
		buf.DrawText(b.X, b.Y, g.label, g.style)
		barY = b.Y + 1
	}

	availableW := b.W
	var valStr string
	if g.showVal {
		valStr = formatGaugeValue(g.value, g.min, g.max)
		valW := buffer.StringWidth(valStr) + 1
		availableW -= valW
		if availableW < 1 {
			availableW = 1
		}
	}

	filledCount := int(math.Round(ratio * float64(availableW)))
	fillColor := g.colorForRatio(ratio)
	emptyColor := buffer.RGB(60, 60, 60)

	for i := 0; i < availableW; i++ {
		if i < filledCount {
			buf.SetCell(b.X+i, barY, buffer.Cell{Rune: g.fillChar, Width: 1, Fg: fillColor})
		} else {
			buf.SetCell(b.X+i, barY, buffer.Cell{Rune: g.emptChar, Width: 1, Fg: emptyColor})
		}
	}

	if valStr != "" {
		buf.DrawText(b.X+availableW+1, barY, valStr, g.style)
	}
}

// paintVertical renders a bottom-to-top bar.
func (g *Gauge) paintVertical(buf *buffer.Buffer, b Rect) {
	ratio := g.ratioLocked()
	barX := b.X
	barTop := b.Y
	barH := b.H
	if g.label != "" {
		buf.DrawText(b.X, b.Y, g.label, g.style)
		barTop = b.Y + 1
		barH--
	}
	if g.showVal {
		barH--
	}
	if barH <= 0 {
		barH = 1
	}

	filledCount := int(math.Round(ratio * float64(barH)))
	fillColor := g.colorForRatio(ratio)
	emptyColor := buffer.RGB(60, 60, 60)

	bottom := barTop + barH - 1
	for i := 0; i < barH; i++ {
		y := bottom - i
		if i < filledCount {
			buf.SetCell(barX, y, buffer.Cell{Rune: g.fillChar, Width: 1, Fg: fillColor})
		} else {
			buf.SetCell(barX, y, buffer.Cell{Rune: g.emptChar, Width: 1, Fg: emptyColor})
		}
	}

	if g.showVal {
		valStr := formatGaugeValue(g.value, g.min, g.max)
		buf.DrawText(b.X, barTop+barH, valStr, g.style)
	}
}

// paintRadial renders a circular arc gauge.
// Uses partial block characters (▘▝▖▗▀▄▌▐▚▞█) to approximate a ring.
func (g *Gauge) paintRadial(buf *buffer.Buffer, b Rect) {
	ratio := g.ratioLocked()
	fillColor := g.colorForRatio(ratio)
	emptyColor := buffer.RGB(60, 60, 60)
	labelY := b.Y

	if g.label != "" {
		buf.DrawText(b.X, b.Y, g.label, g.style)
		labelY += 2
	}

	// ASCII radial gauge: 7 cols wide, 3 rows tall (half-block resolution).
	// Full circle composed of 12 segments (4 per row), each segment is a
	// fraction of the circle. As the ratio increases, segments fill clockwise.
	cx := b.X + 3
	cy := labelY + 1
	radius := 1.5

	segments := radialSegments()
	totalSegs := len(segments)
	fillSegs := int(math.Round(ratio * float64(totalSegs)))

	for i, seg := range segments {
		var r rune
		if i < fillSegs {
			r = seg.fill
			buf.SetCell(cx+seg.dx, cy+seg.dy, buffer.Cell{Rune: r, Width: 1, Fg: fillColor})
		} else {
			r = seg.empty
			buf.SetCell(cx+seg.dx, cy+seg.dy, buffer.Cell{Rune: r, Width: 1, Fg: emptyColor})
		}
	}

	// Draw the value text in the center-ish area.
	if g.showVal {
		pct := fmt.Sprintf("%3.0f%%", ratio*100)
		textW := buffer.StringWidth(pct)
		buf.DrawText(cx-textW/2+1, cy, pct, g.style)
	}

	_ = radius
}

type radialSeg struct {
	dx, dy int
	fill   rune
	empty  rune
}

func radialSegments() []radialSeg {
	// 12 segments arranged in a 7-wide × 3-tall grid using half-blocks.
	return []radialSeg{
		{0, -1, '▗', ' '},
		{1, -1, '▄', ' '},
		{2, -1, '▖', ' '},
		{3, -1, '▐', ' '},
		{3, 0, '▐', ' '},
		{3, 1, '▟', ' '},
		{2, 1, '▀', ' '},
		{1, 1, '▀', ' '},
		{0, 1, '▙', ' '},
		{0, 0, '▌', ' '},
		{0, -1, '▝', ' '},
		{1, -1, '▜', ' '},
	}
}

// ratioLocked computes the ratio without acquiring the lock (caller holds it).
func (g *Gauge) ratioLocked() float64 {
	if g.max <= g.min {
		return 0
	}
	return clamp((g.value-g.min)/(g.max-g.min), 0, 1)
}

// colorForRatio returns the fill color for the given ratio. If thresholds are
// set, the matching band is used; otherwise a red→yellow→green gradient.
func (g *Gauge) colorForRatio(ratio float64) buffer.Color {
	if len(g.thresholds) > 0 {
		for _, t := range g.thresholds {
			if ratio >= t.Low && ratio < t.High {
				return t.Color
			}
			if ratio >= 1.0 && t.High >= 1.0 {
				return t.Color
			}
		}
		// Fallback to last threshold.
		return g.thresholds[len(g.thresholds)-1].Color
	}
	return gradientColor(ratio)
}

// gradientColor returns a green→yellow→red gradient.
// Low values are green, high values are red (ideal for load gauges).
func gradientColor(ratio float64) buffer.Color {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	var r, g uint8
	if ratio < 0.5 {
		g = 255
		r = uint8(ratio * 2 * 255)
	} else {
		r = 255
		g = uint8((1.0 - (ratio-0.5)*2) * 255)
	}
	return buffer.RGB(r, g, 0)
}

// formatGaugeValue formats the numeric display.
func formatGaugeValue(value, min, max float64) string {
	if max == 100 && min == 0 {
		return fmt.Sprintf("%.0f%%", value)
	}
	return fmt.Sprintf("%.1f/%.0f", value, max)
}

// clamp restricts v to [lo, hi].
func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// DefaultThresholds returns a common 3-band threshold set:
// green [0, 0.6), yellow [0.6, 0.85), red [0.85, 1.0].
func DefaultThresholds() []Threshold {
	return []Threshold{
		{Low: 0.0, High: 0.6, Color: buffer.RGB(80, 200, 80)},
		{Low: 0.6, High: 0.85, Color: buffer.RGB(220, 180, 40)},
		{Low: 0.85, High: 1.01, Color: buffer.RGB(220, 80, 80)},
	}
}

// String returns a debug representation of the gauge state.
func (g *Gauge) String() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return fmt.Sprintf("Gauge{value:%.1f, min:%.1f, max:%.1f, ratio:%.2f}", g.value, g.min, g.max, g.ratioLocked())
}

// Ensure imports are used.
var _ = strings.TrimSpace

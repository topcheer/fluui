package component

import (
	"strconv"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// MetricBar renders a horizontal labeled progress bar with min/max range.
// Useful for dashboards showing resource usage, model metrics, or
// benchmark results. Unlike Gauge, it shows the numeric value inline.
//
// Thread-safe.
type MetricBar struct {
	BaseComponent
	mu     sync.RWMutex
	label  string
	value  float64
	min    float64
	max    float64
	unit   string
	barW   int
	customFg buffer.Color
}

// NewMetricBar creates a metric bar with label, value, min, max.
func NewMetricBar(label string, value, min, max float64) *MetricBar {
	return &MetricBar{
		BaseComponent: BaseComponent{id: GenerateID("metricbar")},
		label:         label,
		value:         value,
		min:           min,
		max:           max,
		barW:          10,
	}
}

func (m *MetricBar) Label() string { m.mu.RLock(); defer m.mu.RUnlock(); return m.label }
func (m *MetricBar) SetLabel(s string) { m.mu.Lock(); defer m.mu.Unlock(); m.label = s }

func (m *MetricBar) Value() float64 { m.mu.RLock(); defer m.mu.RUnlock(); return m.value }
func (m *MetricBar) SetValue(v float64) { m.mu.Lock(); defer m.mu.Unlock(); m.value = v }

func (m *MetricBar) Min() float64 { m.mu.RLock(); defer m.mu.RUnlock(); return m.min }
func (m *MetricBar) Max() float64 { m.mu.RLock(); defer m.mu.RUnlock(); return m.max }
func (m *MetricBar) SetRange(min, max float64) { m.mu.Lock(); defer m.mu.Unlock(); m.min = min; m.max = max }

func (m *MetricBar) Unit() string { m.mu.RLock(); defer m.mu.RUnlock(); return m.unit }
func (m *MetricBar) SetUnit(s string) { m.mu.Lock(); defer m.mu.Unlock(); m.unit = s }

func (m *MetricBar) BarWidth() int { m.mu.RLock(); defer m.mu.RUnlock(); return m.barW }
func (m *MetricBar) SetBarWidth(w int) {
	m.mu.Lock(); defer m.mu.Unlock()
	if w < 1 { w = 1 }
	m.barW = w
}

func (m *MetricBar) SetColor(c buffer.Color) { m.mu.Lock(); defer m.mu.Unlock(); m.customFg = c }

// pctLocked returns 0.0–1.0 fraction.
func (m *MetricBar) pctLocked() float64 {
	rng := m.max - m.min
	if rng <= 0 { return 0 }
	p := (m.value - m.min) / rng
	if p < 0 { p = 0 }
	if p > 1 { p = 1 }
	return p
}

func (m *MetricBar) resolveColorLocked() buffer.Color {
	if m.customFg.Type != buffer.ColorNone { return m.customFg }
	p := m.pctLocked()
	t := theme.Get()
	if p >= 0.8 { return t.Error }
	if p >= 0.6 { return t.Warning }
	return t.Success
}

// Measure returns preferred size.
func (m *MetricBar) Measure(cs Constraints) Size {
	m.mu.RLock(); defer m.mu.RUnlock()
	w := buffer.StringWidth(m.label) + 1 // "label "
	w += m.barW + 1                      // bar + space
	// value text: approximate
	w += 8 // " 123.4/x "
	if m.unit != "" { w += len(m.unit) }
	h := 1
	if cs.HasWidth() && w > cs.MaxWidth { w = cs.MaxWidth }
	if w < 1 { w = 1 }
	return Size{W: w, H: h}
}

// Paint draws the metric bar. Zero allocations.
func (m *MetricBar) Paint(buf *buffer.Buffer) {
	m.mu.RLock(); defer m.mu.RUnlock()

	bounds := m.bounds
	if bounds.W <= 0 || bounds.H <= 0 { return }

	t := theme.Get()
	muted := buffer.Style{Fg: t.Muted}
	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	// Label
	if m.label != "" {
		x = buf.DrawText(x, y, m.label+" ", muted)
	}

	// Bar
	pct := m.pctLocked()
	filled := int(pct * float64(m.barW))
	if pct > 0 && filled == 0 { filled = 1 }
	if filled > m.barW { filled = m.barW }
	barColor := m.resolveColorLocked()

	for i := 0; i < m.barW && x < maxX; i++ {
		if i < filled {
			buf.SetCell(x, y, buffer.Cell{Rune: '\u2588', Width: 1, Fg: barColor})
		} else {
			buf.SetCell(x, y, buffer.Cell{Rune: '\u2591', Width: 1, Fg: t.Muted})
		}
		x++
	}

	// Value text
	if x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1})
		x++
	}
	if x < maxX {
		var vb [32]byte
		vbs := vb[:0]
		vbs = strconv.AppendFloat(vbs, m.value, 'f', 1, 64)
		vbs = append(vbs, '/')
		vbs = strconv.AppendFloat(vbs, m.max, 'f', 1, 64)
		if m.unit != "" { vbs = append(vbs, m.unit...) }
		valStyle := buffer.Style{Fg: barColor, Flags: buffer.Bold}
		buf.DrawBytes(x, y, vbs, valStyle)
	}
}

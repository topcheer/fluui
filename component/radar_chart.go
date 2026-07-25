package component

import (
	"math"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// RadarAxis represents one dimension of a radar chart.
type RadarAxis struct {
	Label string
	Value float64 // 0..Max
	Max   float64 // maximum value for this axis
}

// RadarChart renders a radar/spider chart showing multi-dimensional data.
// Useful for model comparison, skill assessments, and metric dashboards.
//
// Thread-safe.
type RadarChart struct {
	BaseComponent
	mu    sync.Mutex
	axes  []RadarAxis
}

// NewRadarChart creates a radar chart with the given axes.
func NewRadarChart(axes []RadarAxis) *RadarChart {
	return &RadarChart{
		BaseComponent: BaseComponent{id: GenerateID("radar")},
		axes:          axes,
	}
}

// SetAxes replaces all axes.
func (r *RadarChart) SetAxes(axes []RadarAxis) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.axes = axes
}

// Axes returns a copy.
func (r *RadarChart) Axes() []RadarAxis {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]RadarAxis, len(r.axes))
	copy(out, r.axes)
	return out
}

// AxisCount returns the number of axes.
func (r *RadarChart) AxisCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.axes)
}

// Measure returns the desired size.
func (r *RadarChart) Measure(cs Constraints) Size {
	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 30
	}
	maxH := cs.MaxHeight
	if maxH <= 0 {
		maxH = 15
	}
	return Size{W: maxW, H: maxH}
}

// Paint renders the radar chart using half-block characters.
func (r *RadarChart) Paint(buf *buffer.Buffer) {
	r.mu.Lock()
	axes := r.axes
	r.mu.Unlock()

	if len(axes) < 3 {
		return
	}

	b := r.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	th := theme.Get()
	ringStyle := buffer.Style{Fg: th.Border}
	lineStyle := buffer.Style{Fg: th.BorderMuted}
	dataStyle := buffer.Style{Fg: th.Accent}
	labelStyle := buffer.Style{Fg: th.Muted}

	cx := b.X + b.W/2
	cy := b.Y + b.H/2
	radius := b.H / 2
	if b.W/2 < radius {
		radius = b.W / 2
	}
	if radius < 2 {
		radius = 2
	}

	n := len(axes)

	// Draw concentric rings (25%, 50%, 75%, 100%)
	for ringPct := 0.25; ringPct <= 1.0; ringPct += 0.25 {
		rr := float64(radius) * ringPct
		for i := 0; i < n; i++ {
			a1 := float64(i) / float64(n) * 2 * math.Pi - math.Pi/2
			a2 := float64(i+1) / float64(n) * 2 * math.Pi - math.Pi/2
			x1 := cx + int(math.Cos(a1)*rr)
			y1 := cy + int(math.Sin(a1)*rr/2)
			x2 := cx + int(math.Cos(a2)*rr)
			y2 := cy + int(math.Sin(a2)*rr/2)
			drawLine(buf, x1, y1, x2, y2, ringStyle, b)
		}
	}

	// Draw axis lines from center to edge
	for i := 0; i < n; i++ {
		a := float64(i) / float64(n) * 2 * math.Pi - math.Pi/2
		ex := cx + int(math.Cos(a)*float64(radius))
		ey := cy + int(math.Sin(a)*float64(radius)/2)
		drawLine(buf, cx, cy, ex, ey, lineStyle, b)

		// Label at edge
		lx := cx + int(math.Cos(a)*float64(radius+1))
		ly := cy + int(math.Sin(a)*float64(radius+1)/2)
		if lx >= b.X && lx < b.X+b.W && ly >= b.Y && ly < b.Y+b.H {
			label := axes[i].Label
			if len(label) > 3 {
				label = label[:3]
			}
			if len(label) > 0 {
				buf.DrawText(lx, ly, label, labelStyle)
			}
		}
	}

	// Draw data polygon
	for i := 0; i < n; i++ {
		max := axes[i].Max
		if max <= 0 {
			max = 1
		}
		val := axes[i].Value / max
		if val > 1 {
			val = 1
		}
		if val < 0 {
			val = 0
		}

		a1 := float64(i) / float64(n) * 2 * math.Pi - math.Pi/2
		a2 := float64((i+1)%n) / float64(n) * 2 * math.Pi - math.Pi/2

		r1 := float64(radius) * val
		r2 := float64(radius) * (func() float64 {
			max2 := axes[(i+1)%n].Max
			if max2 <= 0 {
				max2 = 1
			}
			v2 := axes[(i+1)%n].Value / max2
			if v2 > 1 {
				v2 = 1
			}
			if v2 < 0 {
				v2 = 0
			}
			return v2
		}())

		x1 := cx + int(math.Cos(a1)*r1)
		y1 := cy + int(math.Sin(a1)*r1/2)
		x2 := cx + int(math.Cos(a2)*r2)
		y2 := cy + int(math.Sin(a2)*r2/2)
		drawLine(buf, x1, y1, x2, y2, dataStyle, b)
	}
}

// drawLine draws a line between two points using Bresenham's algorithm.
func drawLine(buf *buffer.Buffer, x0, y0, x1, y1 int, style buffer.Style, b Rect) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	if x0 >= x1 {
		sx = -1
	}
	sy := 1
	if y0 >= y1 {
		sy = -1
	}
	err := dx - dy
	for {
		if x0 >= b.X && x0 < b.X+b.W && y0 >= b.Y && y0 < b.Y+b.H {
			buf.SetCell(x0, y0, buffer.Cell{Rune: '\u00b7', Width: 1, Fg: style.Fg, Bg: style.Bg})
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

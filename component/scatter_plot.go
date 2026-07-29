package component

import (
	"math"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ScatterPlot: 2D Point Scatter Visualization ───
//
// ScatterPlot renders a set of (x, y) data points as characters within a
// bounded area. Points are drawn as '·', with auto or manual axis scaling.
// Useful for data analysis, correlation visualization, and scientific plotting.
//
// Usage:
//
//	sp := NewScatterPlot()
//	sp.SetXRange(0, 100)
//	sp.SetYRange(0, 100)
//	sp.AddPoint(10, 20)
//	sp.AddPoint(50, 80)
//	sp.Paint(buf)

// ScatterPlotStyle holds styling for ScatterPlot.
type ScatterPlotStyle struct {
	Point      buffer.Style
	Axis       buffer.Style
	Grid       buffer.Style
	Border     buffer.Style
	PointChar  rune
}

// DefaultScatterPlotStyle returns sensible defaults.
func DefaultScatterPlotStyle() ScatterPlotStyle {
	point := buffer.Style{Fg: buffer.RGB(96, 165, 250)}   // blue-400
	axis := buffer.Style{Fg: buffer.RGB(100, 116, 139)}   // slate-500
	grid := buffer.Style{Fg: buffer.RGB(51, 65, 85)}      // slate-700
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}   // slate-600
	return ScatterPlotStyle{
		Point:     point,
		Axis:      axis,
		Grid:      grid,
		Border:    border,
		PointChar: '·',
	}
}

// ScatterPoint is a single (x, y) data point.
type ScatterPoint struct {
	X, Y float64
}

// ScatterPlot renders 2D data points as characters.
type ScatterPlot struct {
	BaseComponent
	mu sync.Mutex

	points    []ScatterPoint
	xMin      float64
	xMax      float64
	yMin      float64
	yMax      float64
	autoScale bool

	style ScatterPlotStyle
}

// NewScatterPlot creates a ScatterPlot with defaults.
func NewScatterPlot() *ScatterPlot {
	sp := &ScatterPlot{
		autoScale: true,
		style:     DefaultScatterPlotStyle(),
	}
	sp.SetID(GenerateID("scatter"))
	return sp
}

// AddPoint adds a data point.
func (sp *ScatterPlot) AddPoint(x, y float64) *ScatterPlot {
	sp.mu.Lock()
	sp.points = append(sp.points, ScatterPoint{X: x, Y: y})
	sp.mu.Unlock()
	return sp
}

// SetPoints replaces all data points.
func (sp *ScatterPlot) SetPoints(pts []ScatterPoint) *ScatterPlot {
	sp.mu.Lock()
	sp.points = pts
	sp.mu.Unlock()
	return sp
}

// Points returns a copy of the current points.
func (sp *ScatterPlot) Points() []ScatterPoint {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	result := make([]ScatterPoint, len(sp.points))
	copy(result, sp.points)
	return result
}

// PointCount returns the number of points.
func (sp *ScatterPlot) PointCount() int {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	return len(sp.points)
}

// SetXRange sets the manual X axis range. Disables auto-scaling for X.
func (sp *ScatterPlot) SetXRange(min, max float64) *ScatterPlot {
	sp.mu.Lock()
	sp.xMin = min
	sp.xMax = max
	sp.autoScale = false
	sp.mu.Unlock()
	return sp
}

// SetYRange sets the manual Y axis range. Disables auto-scaling for Y.
func (sp *ScatterPlot) SetYRange(min, max float64) *ScatterPlot {
	sp.mu.Lock()
	sp.yMin = min
	sp.yMax = max
	sp.autoScale = false
	sp.mu.Unlock()
	return sp
}

// SetAutoScale enables or disables auto-scaling.
func (sp *ScatterPlot) SetAutoScale(v bool) *ScatterPlot {
	sp.mu.Lock()
	sp.autoScale = v
	sp.mu.Unlock()
	return sp
}

// XRange returns the current X axis range.
func (sp *ScatterPlot) XRange() (float64, float64) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if sp.autoScale {
		return sp.computeRangeLocked(true)
	}
	return sp.xMin, sp.xMax
}

// YRange returns the current Y axis range.
func (sp *ScatterPlot) YRange() (float64, float64) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if sp.autoScale {
		return sp.computeRangeLocked(false)
	}
	return sp.yMin, sp.yMax
}

// computeRangeLocked returns the computed min/max for X (isX=true) or Y.
// Caller must hold the lock.
func (sp *ScatterPlot) computeRangeLocked(isX bool) (float64, float64) {
	if len(sp.points) == 0 {
		return 0, 1
	}
	minVal := math.Inf(1)
	maxVal := math.Inf(-1)
	for _, p := range sp.points {
		v := p.X
		if !isX {
			v = p.Y
		}
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	if minVal == maxVal {
		maxVal = minVal + 1
	}
	return minVal, maxVal
}

// SetStyle sets the custom style.
func (sp *ScatterPlot) SetStyle(s ScatterPlotStyle) *ScatterPlot {
	sp.mu.Lock()
	sp.style = s
	sp.mu.Unlock()
	return sp
}

// Measure returns the preferred size.
func (sp *ScatterPlot) Measure(cs Constraints) Size {
	w := 40
	h := 20
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the scatter plot into the buffer.
func (sp *ScatterPlot) Paint(buf *buffer.Buffer) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	b := sp.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 {
		w = 40
	}
	if h < 5 {
		h = 20
	}

	// Determine ranges (access fields directly, no locking methods)
	xMin, xMax := sp.xMin, sp.xMax
	yMin, yMax := sp.yMin, sp.yMax
	if sp.autoScale {
		xMin, xMax = sp.computeRangeLocked(true)
		yMin, yMax = sp.computeRangeLocked(false)
	}

	xRange := xMax - xMin
	yRange := yMax - yMin
	if xRange == 0 {
		xRange = 1
	}
	if yRange == 0 {
		yRange = 1
	}

	// Plot area dimensions (leave room for border)
	plotX := x + 1
	plotY := y + 1
	plotW := w - 2
	plotH := h - 2
	if plotW < 2 {
		plotW = 2
	}
	if plotH < 2 {
		plotH = 2
	}

	// Draw border
	bs := sp.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 {
				ch = '┌'
			} else if row == 0 && col == w-1 {
				ch = '┐'
			} else if row == h-1 && col == 0 {
				ch = '└'
			} else if row == h-1 && col == w-1 {
				ch = '┘'
			} else if row == 0 || row == h-1 {
				ch = '─'
			} else if col == 0 || col == w-1 {
				ch = '│'
			}
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	// Draw X axis at bottom of plot area
	axisStyle := sp.style.Axis
	axisRow := plotY + plotH - 1
	for col := 0; col < plotW; col++ {
		cx := plotX + col
		if cx < buf.Width && axisRow < buf.Height {
			buf.SetCell(cx, axisRow, buffer.Cell{Rune: '─', Fg: axisStyle.Fg, Bg: axisStyle.Bg, Flags: axisStyle.Flags, Width: 1})
		}
	}

	// Draw Y axis at left of plot area
	for row := 0; row < plotH; row++ {
		cy := plotY + row
		if plotX < buf.Width && cy < buf.Height {
			buf.SetCell(plotX, cy, buffer.Cell{Rune: '│', Fg: axisStyle.Fg, Bg: axisStyle.Bg, Flags: axisStyle.Flags, Width: 1})
		}
	}

	// Draw grid (dotted, every ~5 cells)
	gridStyle := sp.style.Grid
	for col := 5; col < plotW-1; col += 5 {
		for row := 1; row < plotH-1; row++ {
			cx := plotX + col
			cy := plotY + row
			if cx < buf.Width && cy < buf.Height {
				buf.SetCell(cx, cy, buffer.Cell{Rune: '┊', Fg: gridStyle.Fg, Bg: gridStyle.Bg, Flags: gridStyle.Flags, Width: 1})
			}
		}
	}

	// Draw data points
	pointStyle := sp.style.Point
	pointChar := sp.style.PointChar
	if pointChar == 0 {
		pointChar = '·'
	}
	for _, pt := range sp.points {
		// Map data coordinates to screen coordinates
		// X: left→right, Y: bottom→top (inverted since screen Y grows downward)
		screenX := plotX + 1 + int((pt.X-xMin)/xRange*float64(plotW-2))
		// Invert Y: higher data value = lower screen row
		screenY := plotY + plotH - 2 - int((pt.Y-yMin)/yRange*float64(plotH-2))

		// Clamp to plot area
		if screenX < plotX+1 {
			screenX = plotX + 1
		}
		if screenX >= plotX+plotW {
			screenX = plotX + plotW - 1
		}
		if screenY < plotY {
			screenY = plotY
		}
		if screenY >= plotY+plotH-1 {
			screenY = plotY + plotH - 2
		}

		if screenX < buf.Width && screenY < buf.Height {
			buf.SetCell(screenX, screenY, buffer.Cell{Rune: pointChar, Fg: pointStyle.Fg, Bg: pointStyle.Bg, Flags: pointStyle.Flags, Width: 1})
		}
	}
}

// Children returns nil.
func (sp *ScatterPlot) Children() []Component { return nil }

package component

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// --- Types ---

// ChartPoint is a 2D data point for chart series.
type ChartPoint struct {
	X, Y float64
}

// ChartMarker controls how data points are visually marked on the chart.
type ChartMarker int

const (
	// ChartMarkerNone draws no markers, only lines.
	ChartMarkerNone ChartMarker = iota
	// ChartMarkerDot draws a filled circle at each data point.
	ChartMarkerDot
	// ChartMarkerPlus draws a + at each data point.
	ChartMarkerPlus
	// ChartMarkerStar draws a * at each data point.
	ChartMarkerStar
)

// ChartSeries is a named, colored data series rendered on the chart.
type ChartSeries struct {
	// Name is shown in the legend.
	Name string
	// Data is the sequence of (X, Y) points, sorted by X ascending.
	Data []ChartPoint
	// Color is the foreground color for this series' line and markers.
	Color buffer.Color
	// Marker controls point markers (ChartMarkerNone for line-only).
	Marker ChartMarker
}

// ChartAxisConfig configures an axis (X or Y) of the chart.
type ChartAxisConfig struct {
	// Title is displayed beside the axis.
	Title string
	// Min is the minimum value. If 0 and Max is 0, auto-scaled from data.
	Min float64
	// Max is the maximum value. If 0 and Min is 0, auto-scaled from data.
	Max float64
	// LabelCount is the number of tick labels to render (0 = auto, typically 4-5).
	LabelCount int
}

// LineChart is a multi-series line chart that renders data points connected
// by line segments on a 2D grid with axes, optional grid lines, and a legend.
//
// Features:
//   - Multiple overlapping series with distinct colors
//   - Auto-scaling Y-axis (and X-axis) from data
//   - Configurable grid lines (dotted style)
//   - Y-axis labels on the left, X-axis labels on the bottom
//   - Legend with series names and color indicators
//   - Data point markers (dot, plus, star, or none)
//   - Box-drawing characters for line segments (─ │ ╱ ╲)
//   - Thread-safe via sync.RWMutex
type LineChart struct {
	BaseComponent
	mu sync.RWMutex

	title      string
	series     []ChartSeries
	showGrid   bool
	showAxes   bool
	showLegend bool
	showTitle  bool

	xAxis ChartAxisConfig
	yAxis ChartAxisConfig

	autoScale bool // if true, recompute ranges on each paint

	gridStyle  buffer.Style
	axisStyle  buffer.Style
	titleStyle buffer.Style

	currentTheme *theme.Theme
}

// NewLineChart creates a LineChart with sensible defaults.
func NewLineChart() *LineChart {
	return &LineChart{
		showGrid:    true,
		showAxes:    true,
		showLegend:  true,
		showTitle:   true,
		autoScale:   true,
		xAxis:       ChartAxisConfig{LabelCount: 5},
		yAxis:       ChartAxisConfig{LabelCount: 5},
		gridStyle:   buffer.Style{Fg: buffer.NamedColor(buffer.NamedBrightBlack)},
		axisStyle:   buffer.Style{Fg: buffer.NamedColor(buffer.NamedWhite)},
		titleStyle:  buffer.Style{Fg: buffer.NamedColor(buffer.NamedCyan), Flags: buffer.Bold},
	}
}

// --- Configuration ---

// SetTitle sets the chart title shown at the top.
func (lc *LineChart) SetTitle(t string) {
	lc.mu.Lock()
	lc.title = t
	lc.mu.Unlock()
}

// Title returns the current chart title.
func (lc *LineChart) Title() string {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.title
}

// AddSeries appends a data series to the chart.
func (lc *LineChart) AddSeries(s ChartSeries) {
	lc.mu.Lock()
	lc.series = append(lc.series, s)
	lc.mu.Unlock()
}

// SetSeries replaces all series with the given slice.
func (lc *LineChart) SetSeries(series []ChartSeries) {
	lc.mu.Lock()
	lc.series = series
	lc.mu.Unlock()
}

// ClearSeries removes all series.
func (lc *LineChart) ClearSeries() {
	lc.mu.Lock()
	lc.series = nil
	lc.mu.Unlock()
}

// SeriesCount returns the number of series.
func (lc *LineChart) SeriesCount() int {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return len(lc.series)
}

// Series returns the series at the given index.
func (lc *LineChart) Series(i int) ChartSeries {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	if i < 0 || i >= len(lc.series) {
		return ChartSeries{}
	}
	return lc.series[i]
}

// SetShowGrid toggles dotted grid lines.
func (lc *LineChart) SetShowGrid(b bool) {
	lc.mu.Lock()
	lc.showGrid = b
	lc.mu.Unlock()
}

// SetShowAxes toggles axis labels and borders.
func (lc *LineChart) SetShowAxes(b bool) {
	lc.mu.Lock()
	lc.showAxes = b
	lc.mu.Unlock()
}

// SetShowLegend toggles the series legend.
func (lc *LineChart) SetShowLegend(b bool) {
	lc.mu.Lock()
	lc.showLegend = b
	lc.mu.Unlock()
}

// SetShowTitle toggles the title bar.
func (lc *LineChart) SetShowTitle(b bool) {
	lc.mu.Lock()
	lc.showTitle = b
	lc.mu.Unlock()
}

// SetXAxis configures the X-axis.
func (lc *LineChart) SetXAxis(cfg ChartAxisConfig) {
	lc.mu.Lock()
	lc.xAxis = cfg
	lc.mu.Unlock()
}

// SetYAxis configures the Y-axis.
func (lc *LineChart) SetYAxis(cfg ChartAxisConfig) {
	lc.mu.Lock()
	lc.yAxis = cfg
	lc.mu.Unlock()
}

// SetAutoScale enables/disables automatic range computation.
func (lc *LineChart) SetAutoScale(b bool) {
	lc.mu.Lock()
	lc.autoScale = b
	lc.mu.Unlock()
}

// SetGridStyle sets the style for grid lines.
func (lc *LineChart) SetGridStyle(s buffer.Style) {
	lc.mu.Lock()
	lc.gridStyle = s
	lc.mu.Unlock()
}

// SetTheme applies a theme for default colors.
func (lc *LineChart) SetTheme(t *theme.Theme) {
	lc.mu.Lock()
	lc.currentTheme = t
	lc.mu.Unlock()
}

// --- Range computation ---

// computeRanges returns (xMin, xMax, yMin, yMax) based on data or fixed config.
func (lc *LineChart) computeRanges() (xMin, xMax, yMin, yMax float64) {
	xMin, xMax = lc.xAxis.Min, lc.xAxis.Max
	yMin, yMax = lc.yAxis.Min, lc.yAxis.Max

	if lc.autoScale || (xMin == 0 && xMax == 0) || (yMin == 0 && yMax == 0) {
		dXMin, dXMax, dYMin, dYMax := math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1)
		hasData := false
		for _, s := range lc.series {
			for _, p := range s.Data {
				hasData = true
				if p.X < dXMin {
					dXMin = p.X
				}
				if p.X > dXMax {
					dXMax = p.X
				}
				if p.Y < dYMin {
					dYMin = p.Y
				}
				if p.Y > dYMax {
					dYMax = p.Y
				}
			}
		}
		if !hasData {
			return 0, 1, 0, 1
		}
		if xMin == 0 && xMax == 0 {
			xMin, xMax = dXMin, dXMax
		}
		if yMin == 0 && yMax == 0 {
			yMin, yMax = dYMin, dYMax
		}
	}

	// Prevent division by zero
	if xMax == xMin {
		xMax = xMin + 1
	}
	if yMax == yMin {
		yMax = yMin + 1
	}
	return
}

// --- Component interface ---

// Measure returns the desired size, defaulting to 40x12.
func (lc *LineChart) Measure(cs Constraints) Size {
	w, h := 40, 12
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	if w < 10 {
		w = 10
	}
	if h < 5 {
		h = 5
	}
	return Size{W: w, H: h}
}

// Paint renders the chart into the buffer.
func (lc *LineChart) Paint(buf *buffer.Buffer) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	bounds := lc.bounds
	if bounds.W < 5 || bounds.H < 3 {
		return
	}

	xMin, xMax, yMin, yMax := lc.computeRanges()

	// Layout: title (1 row) + legend (1 row) + chart area + x-axis labels (1 row)
	row := bounds.Y
	chartTop := row
	chartBottom := bounds.Y + bounds.H - 1

	// Y-axis label width
	yLabelW := 0
	if lc.showAxes {
		yLabelW = maxInt(len(formatAxisVal(yMin)), len(formatAxisVal(yMax))) + 1
	}

	// Title
	if lc.showTitle && lc.title != "" {
		drawText(buf, bounds.X, row, bounds.W, lc.title, lc.titleStyle)
		row++
	}

	// Legend
	if lc.showLegend && len(lc.series) > 0 {
		x := bounds.X
		for _, s := range lc.series {
			marker := seriesMarkerRune(s.Marker)
			entry := fmt.Sprintf("%c %s", marker, s.Name)
			if x+len(entry) > bounds.X+bounds.W {
				break
			}
			for _, r := range entry {
				buf.SetCell(x, row, buffer.Cell{
					Rune:  r,
					Width: 1,
					Fg:    s.Color,
					Bg:    lc.axisStyle.Bg,
				})
				x++
			}
			x += 2 // spacing between entries
		}
		row++
	}

	// Chart area
	if lc.showAxes {
		chartTop = row
		chartBottom = bounds.Y + bounds.H - 2 // leave 1 row for x-axis labels
	} else {
		chartTop = row
		chartBottom = bounds.Y + bounds.H - 1
	}

	chartX := bounds.X
	chartW := bounds.W
	if lc.showAxes {
		chartX = bounds.X + yLabelW
		chartW = bounds.W - yLabelW
	}
	chartH := chartBottom - chartTop + 1
	if chartW < 2 || chartH < 2 {
		return
	}

	// Draw grid
	if lc.showGrid {
		lc.drawGrid(buf, chartX, chartTop, chartW, chartH, xMin, xMax, yMin, yMax)
	}

	// Draw axes
	if lc.showAxes {
		lc.drawAxes(buf, chartX, chartTop, chartW, chartH, xMin, xMax, yMin, yMax)
	}

	// Draw each series
	for _, s := range lc.series {
		lc.drawSeries(buf, s, chartX, chartTop, chartW, chartH, xMin, xMax, yMin, yMax)
	}
}

// --- Internal rendering ---

// drawGrid draws dotted grid lines.
func (lc *LineChart) drawGrid(buf *buffer.Buffer, cx, cy, cw, ch int, xMin, xMax, yMin, yMax float64) {
	// Horizontal grid lines (Y-axis divisions)
	yLabels := lc.yAxis.LabelCount
	if yLabels <= 0 {
		yLabels = 5
	}
	for i := 0; i <= yLabels; i++ {
		y := cy + ch - 1 - (i*ch)/yLabels
		if y < cy || y > cy+ch-1 {
			continue
		}
		for x := cx; x < cx+cw; x++ {
			setGridCell(buf, x, y, lc.gridStyle, '─', '·')
		}
	}

	// Vertical grid lines (X-axis divisions)
	xLabels := lc.xAxis.LabelCount
	if xLabels <= 0 {
		xLabels = 5
	}
	for i := 0; i <= xLabels; i++ {
		x := cx + (i*cw)/xLabels
		if x < cx || x >= cx+cw {
			continue
		}
		for y := cy; y < cy+ch; y++ {
			setGridCell(buf, x, y, lc.gridStyle, '│', '·')
		}
	}
}

// setGridCell draws a grid character without overwriting existing series cells.
func setGridCell(buf *buffer.Buffer, x, y int, style buffer.Style, lineChar, dotChar rune) {
	existing := buf.GetCell(x, y)
	// Only draw grid if the cell is empty (space or zero rune)
	if existing.Rune != ' ' && existing.Rune != 0 {
		return
	}
	// Use dot character for a subtle grid
	buf.SetCell(x, y, buffer.Cell{
		Rune:  dotChar,
		Width: 1,
		Fg:    style.Fg,
		Bg:    style.Bg,
	})
}

// drawAxes draws axis labels and borders.
func (lc *LineChart) drawAxes(buf *buffer.Buffer, cx, cy, cw, ch int, xMin, xMax, yMin, yMax float64) {
	bg := lc.axisStyle.Bg

	// Y-axis labels (left side)
	yLabels := lc.yAxis.LabelCount
	if yLabels <= 0 {
		yLabels = 5
	}
	for i := 0; i <= yLabels; i++ {
		y := cy + ch - 1 - (i*ch)/yLabels
		if y < cy || y > cy+ch-1 {
			continue
		}
		val := yMin + (yMax-yMin)*float64(i)/float64(yLabels)
		label := formatAxisVal(val)
		// Right-align the label
		labelX := cx - len(label) - 1
		for j, r := range label {
			buf.SetCell(labelX+j, y, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    lc.axisStyle.Fg,
				Bg:    bg,
			})
		}
	}

	// X-axis labels (bottom row)
	xLabels := lc.xAxis.LabelCount
	if xLabels <= 0 {
		xLabels = 5
	}
	labelRow := cy + ch
	for i := 0; i <= xLabels; i++ {
		x := cx + (i*cw)/xLabels
		if x < cx || x >= cx+cw {
			continue
		}
		val := xMin + (xMax-xMin)*float64(i)/float64(xLabels)
		label := formatAxisVal(val)
		// Center the label under the tick
		startX := x - len(label)/2
		if startX < cx {
			startX = cx
		}
		for j, r := range label {
			if startX+j >= cx+cw {
				break
			}
			buf.SetCell(startX+j, labelRow, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    lc.axisStyle.Fg,
				Bg:    bg,
			})
		}
	}
}

// drawSeries draws a single data series as connected line segments.
func (lc *LineChart) drawSeries(buf *buffer.Buffer, s ChartSeries, cx, cy, cw, ch int, xMin, xMax, yMin, yMax float64) {
	if len(s.Data) == 0 {
		return
	}

	// Map data points to screen coordinates
	pts := make([][2]int, len(s.Data))
	for i, p := range s.Data {
		pts[i] = lc.dataToScreen(p.X, p.Y, cx, cy, cw, ch, xMin, xMax, yMin, yMax)
	}

	// Draw line segments between consecutive points
	for i := 0; i < len(pts)-1; i++ {
		lc.drawLineSegment(buf, pts[i][0], pts[i][1], pts[i+1][0], pts[i+1][1], s.Color, lc.axisStyle.Bg)
	}

	// Draw markers at each data point
	if s.Marker != ChartMarkerNone {
		for _, pt := range pts {
			drawMarker(buf, pt[0], pt[1], s.Marker, s.Color, lc.axisStyle.Bg)
		}
	}
}

// dataToScreen maps a data coordinate to screen cell coordinates.
func (lc *LineChart) dataToScreen(dx, dy float64, cx, cy, cw, ch int, xMin, xMax, yMin, yMax float64) [2]int {
	sx := cx
	if xMax > xMin {
		sx = cx + int((dx-xMin)/(xMax-xMin)*float64(cw-1))
	}
	sy := cy + ch - 1
	if yMax > yMin {
		sy = cy + ch - 1 - int((dy-yMin)/(yMax-yMin)*float64(ch-1))
	}
	// Clamp to chart area
	if sx < cx {
		sx = cx
	}
	if sx >= cx+cw {
		sx = cx + cw - 1
	}
	if sy < cy {
		sy = cy
	}
	if sy >= cy+ch {
		sy = cy + ch - 1
	}
	return [2]int{sx, sy}
}

// drawLineSegment draws a line between two screen points using Bresenham.
func (lc *LineChart) drawLineSegment(buf *buffer.Buffer, x1, y1, x2, y2 int, color buffer.Color, bg buffer.Color) {
	dx := absInt(x2 - x1)
	dy := absInt(y2 - y1)

	// Determine the overall character based on slope
	var slopeChar rune
	if dx == 0 && dy == 0 {
		slopeChar = '·'
	} else if dx == 0 {
		slopeChar = '│'
	} else if dy == 0 {
		slopeChar = '─'
	} else if dx >= 2*dy {
		slopeChar = '─'
	} else if dy >= 2*dx {
		slopeChar = '│'
	} else {
		// Diagonal — check direction
		if (y2 < y1 && x2 > x1) || (y2 > y1 && x2 < x1) {
			slopeChar = '╱' // ascending (up-right or down-left)
		} else {
			slopeChar = '╲' // descending
		}
	}

	// Bresenham line algorithm
	sx := signInt(x2 - x1)
	sy := signInt(y2 - y1)
	err := dx - dy

	x, y := x1, y1
	for {
		buf.SetCell(x, y, buffer.Cell{
			Rune:  slopeChar,
			Width: 1,
			Fg:    color,
			Bg:    bg,
		})
		if x == x2 && y == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

// drawMarker draws a data point marker.
func drawMarker(buf *buffer.Buffer, x, y int, marker ChartMarker, color buffer.Color, bg buffer.Color) {
	var r rune
	switch marker {
	case ChartMarkerDot:
		r = '●'
	case ChartMarkerPlus:
		r = '+'
	case ChartMarkerStar:
		r = '*'
	default:
		return
	}
	buf.SetCell(x, y, buffer.Cell{
		Rune:  r,
		Width: 1,
		Fg:    color,
		Bg:    bg,
	})
}

// seriesMarkerRune returns the rune used in the legend for a marker type.
func seriesMarkerRune(m ChartMarker) rune {
	switch m {
	case ChartMarkerDot:
		return '●'
	case ChartMarkerPlus:
		return '+'
	case ChartMarkerStar:
		return '*'
	default:
		return '─'
	}
}

// drawText writes a string into the buffer with the given style.
func drawText(buf *buffer.Buffer, x, y, maxW int, text string, style buffer.Style) {
	col := x
	for _, r := range text {
		if col >= x+maxW {
			break
		}
		buf.SetCell(col, y, buffer.Cell{
			Rune:  r,
			Width: 1,
			Fg:    style.Fg,
			Bg:    style.Bg,
			Flags: style.Flags,
		})
		col++
	}
}

// formatAxisVal formats a float for axis labels.
func formatAxisVal(f float64) string {
	if f == 0 {
		return "0"
	}
	if math.IsInf(f, 0) {
		return "∞"
	}
	// Integer values
	if f == math.Floor(f) && math.Abs(f) < 1e6 {
		return fmt.Sprintf("%d", int64(f))
	}
	// Small decimals
	if math.Abs(f) >= 100 {
		return fmt.Sprintf("%.0f", f)
	}
	if math.Abs(f) >= 10 {
		return fmt.Sprintf("%.1f", f)
	}
	return fmt.Sprintf("%.2f", f)
}

// --- Math helpers ---

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func signInt(x int) int {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// minInt returns the smaller of two ints.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// formatSeriesValue formats a Y value for display in tooltips/labels.
func formatSeriesValue(f float64) string {
	return formatAxisVal(f)
}

// --- Utility methods ---

// SeriesNames returns the names of all series.
func (lc *LineChart) SeriesNames() []string {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	names := make([]string, len(lc.series))
	for i, s := range lc.series {
		names[i] = s.Name
	}
	return names
}

// PointCount returns the total number of data points across all series.
func (lc *LineChart) PointCount() int {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	total := 0
	for _, s := range lc.series {
		total += len(s.Data)
	}
	return total
}

// Range returns the current data range (xMin, xMax, yMin, yMax).
func (lc *LineChart) Range() (xMin, xMax, yMin, yMax float64) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.computeRanges()
}

// String returns a brief description for debugging.
func (lc *LineChart) String() string {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	var sb strings.Builder
	fmt.Fprintf(&sb, "LineChart(%q, %d series, %d pts)", lc.title, len(lc.series), func() int {
		total := 0
		for _, s := range lc.series {
			total += len(s.Data)
		}
		return total
	}())
	return sb.String()
}

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

// BarOrientation controls whether bars grow upward (vertical) or rightward (horizontal).
type BarOrientation int

const (
	// BarVertical draws bars growing upward from bottom.
	BarVertical BarOrientation = iota
	// BarHorizontal draws bars growing rightward from left.
	BarHorizontal
)

// BarData is a single category with a label and value for a bar chart.
type BarData struct {
	// Label is the category name shown below (vertical) or beside (horizontal) the bar.
	Label string
	// Value is the numeric height/length of the bar.
	Value float64
}

// BarSeries is a named, colored series of bars for grouped bar charts.
type BarSeries struct {
	// Name is shown in the legend.
	Name string
	// Data is the list of bars in this series.
	Data []BarData
	// Color is the bar fill color.
	Color buffer.Color
}

// BarChart is a multi-series bar chart that renders vertical or horizontal bars
// with categories, auto-scaling, optional grid lines, value labels, and a legend.
//
// Features:
//   - Multiple grouped series with distinct colors
//   - Vertical and horizontal orientations
//   - Auto-scaling from data
//   - Configurable grid lines
//   - Category labels and value labels
//   - Legend with series names
//   - Unicode block characters for partial-height bars
//   - Thread-safe via sync.RWMutex
type BarChart struct {
	BaseComponent
	mu sync.RWMutex

	title      string
	series     []BarSeries
	orientation BarOrientation

	showGrid    bool
	showAxes    bool
	showLegend  bool
	showTitle   bool
	showValues  bool

	maxVal   float64 // 0 = auto-scale
	gap      int     // gap between bar groups (columns)

	gridStyle  buffer.Style
	axisStyle  buffer.Style
	titleStyle buffer.Style

	currentTheme *theme.Theme
}

// NewBarChart creates a BarChart with sensible defaults.
func NewBarChart() *BarChart {
	return &BarChart{
		orientation: BarVertical,
		showGrid:    true,
		showAxes:    true,
		showLegend:  true,
		showTitle:   true,
		showValues:  false,
		gap:         1,
		gridStyle:   buffer.Style{Fg: buffer.NamedColor(buffer.NamedBrightBlack)},
		axisStyle:   buffer.Style{Fg: buffer.NamedColor(buffer.NamedWhite)},
		titleStyle:  buffer.Style{Fg: buffer.NamedColor(buffer.NamedCyan), Flags: buffer.Bold},
	}
}

// --- Configuration ---

// SetTitle sets the chart title shown at the top.
func (bc *BarChart) SetTitle(t string) {
	bc.mu.Lock()
	bc.title = t
	bc.mu.Unlock()
}

// Title returns the current chart title.
func (bc *BarChart) Title() string {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.title
}

// SetOrientation sets vertical or horizontal bar layout.
func (bc *BarChart) SetOrientation(o BarOrientation) {
	bc.mu.Lock()
	bc.orientation = o
	bc.mu.Unlock()
}

// Orientation returns the current orientation.
func (bc *BarChart) Orientation() BarOrientation {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.orientation
}

// AddSeries appends a data series to the chart.
func (bc *BarChart) AddSeries(s BarSeries) {
	bc.mu.Lock()
	bc.series = append(bc.series, s)
	bc.mu.Unlock()
}

// SetSeries replaces all series with the given slice.
func (bc *BarChart) SetSeries(series []BarSeries) {
	bc.mu.Lock()
	bc.series = series
	bc.mu.Unlock()
}

// ClearSeries removes all series.
func (bc *BarChart) ClearSeries() {
	bc.mu.Lock()
	bc.series = nil
	bc.mu.Unlock()
}

// SeriesCount returns the number of series.
func (bc *BarChart) SeriesCount() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return len(bc.series)
}

// Series returns the series at the given index.
func (bc *BarChart) Series(i int) BarSeries {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	if i < 0 || i >= len(bc.series) {
		return BarSeries{}
	}
	return bc.series[i]
}

// SetShowGrid toggles dotted grid lines.
func (bc *BarChart) SetShowGrid(b bool) {
	bc.mu.Lock()
	bc.showGrid = b
	bc.mu.Unlock()
}

// SetShowAxes toggles axis labels and borders.
func (bc *BarChart) SetShowAxes(b bool) {
	bc.mu.Lock()
	bc.showAxes = b
	bc.mu.Unlock()
}

// SetShowLegend toggles the series legend.
func (bc *BarChart) SetShowLegend(b bool) {
	bc.mu.Lock()
	bc.showLegend = b
	bc.mu.Unlock()
}

// SetShowTitle toggles the title bar.
func (bc *BarChart) SetShowTitle(b bool) {
	bc.mu.Lock()
	bc.showTitle = b
	bc.mu.Unlock()
}

// SetShowValues toggles value labels on bars.
func (bc *BarChart) SetShowValues(b bool) {
	bc.mu.Lock()
	bc.showValues = b
	bc.mu.Unlock()
}

// SetMaxVal sets a fixed maximum value for the axis (0 = auto-scale).
func (bc *BarChart) SetMaxVal(v float64) {
	bc.mu.Lock()
	bc.maxVal = v
	bc.mu.Unlock()
}

// MaxVal returns the configured max value (0 = auto).
func (bc *BarChart) MaxVal() float64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.maxVal
}

// SetGap sets the gap (in columns) between bar groups.
func (bc *BarChart) SetGap(g int) {
	bc.mu.Lock()
	if g < 0 {
		g = 0
	}
	bc.gap = g
	bc.mu.Unlock()
}

// Gap returns the gap between bar groups.
func (bc *BarChart) Gap() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.gap
}

// SetTheme applies a color theme.
func (bc *BarChart) SetTheme(t *theme.Theme) {
	bc.mu.Lock()
	bc.currentTheme = t
	bc.mu.Unlock()
}

// SetGridStyle sets the style for grid lines.
func (bc *BarChart) SetGridStyle(s buffer.Style) {
	bc.mu.Lock()
	bc.gridStyle = s
	bc.mu.Unlock()
}

// SetAxisStyle sets the style for axes and labels.
func (bc *BarChart) SetAxisStyle(s buffer.Style) {
	bc.mu.Lock()
	bc.axisStyle = s
	bc.mu.Unlock()
}

// SetTitleStyle sets the style for the title.
func (bc *BarChart) SetTitleStyle(s buffer.Style) {
	bc.mu.Lock()
	bc.titleStyle = s
	bc.mu.Unlock()
}

// --- Component interface ---

// Measure returns the desired size for the chart.
func (bc *BarChart) Measure(cs Constraints) Size {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	w := 60
	h := 15
	if bc.orientation == BarHorizontal {
		w = 60
		h = 20
	}

	numCategories := 0
	for _, s := range bc.series {
		if len(s.Data) > numCategories {
			numCategories = len(s.Data)
		}
	}
	if numCategories > 0 {
		numSeries := len(bc.series)
		if numSeries == 0 {
			numSeries = 1
		}
		if bc.orientation == BarVertical {
			minW := numCategories*(numSeries+bc.gap) + 10
			if minW > w {
				w = minW
			}
		}
	}

	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.HasHeight() && h > cs.MaxHeight {
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

// SetBounds sets the position and size for the chart.
func (bc *BarChart) SetBounds(r Rect) {
	bc.BaseComponent.SetBounds(r)
}

// Paint renders the chart into the buffer.
func (bc *BarChart) Paint(buf *buffer.Buffer) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	bounds := bc.Bounds()
	if bounds.W < 3 || bounds.H < 3 || len(bc.series) == 0 {
		return
	}

	// Compute max value
	maxVal := bc.maxVal
	if maxVal <= 0 {
		for _, s := range bc.series {
			for _, d := range s.Data {
				if d.Value > maxVal {
					maxVal = d.Value
				}
			}
		}
	}
	if maxVal <= 0 {
		maxVal = 1
	}

	if bc.orientation == BarVertical {
		bc.paintVertical(buf, bounds, maxVal)
	} else {
		bc.paintHorizontal(buf, bounds, maxVal)
	}
}

// --- Vertical rendering ---

func (bc *BarChart) paintVertical(buf *buffer.Buffer, bounds Rect, maxVal float64) {
	bg := bc.axisStyle.Bg
	numSeries := len(bc.series)
	numCategories := 0
	for _, s := range bc.series {
		if len(s.Data) > numCategories {
			numCategories = len(s.Data)
		}
	}
	if numCategories == 0 {
		return
	}

	row := bounds.Y

	// Title
	if bc.showTitle && bc.title != "" {
		titleRunes := []rune(bc.title)
		titleX := bounds.X + (bounds.W-len(titleRunes))/2
		if titleX < bounds.X {
			titleX = bounds.X
		}
		for i, r := range titleRunes {
			if titleX+i >= bounds.X+bounds.W {
				break
			}
			buf.SetCell(titleX+i, row, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    bc.titleStyle.Fg,
				Bg:    bg,
				Flags: bc.titleStyle.Flags,
			})
		}
		row++
	}

	// Legend
	if bc.showLegend && numSeries > 1 {
		x := bounds.X
		for _, s := range bc.series {
			entry := fmt.Sprintf("█ %s", s.Name)
			if x+len(entry) > bounds.X+bounds.W {
				break
			}
			for _, r := range entry {
				buf.SetCell(x, row, buffer.Cell{
					Rune:  r,
					Width: 1,
					Fg:    s.Color,
					Bg:    bg,
				})
				x++
			}
			x += 2
		}
		row++
	}

	// Y-axis label width
	yLabelW := len(formatBarVal(maxVal)) + 2

	// Chart area
	chartTop := row
	chartBottom := bounds.Y + bounds.H - 2 // 1 row for x-labels, 1 for axis
	if chartBottom <= chartTop {
		chartBottom = bounds.Y + bounds.H - 1
	}
	chartX := bounds.X
	chartW := bounds.W
	if bc.showAxes {
		chartX = bounds.X + yLabelW
		chartW = bounds.W - yLabelW
	}
	chartH := chartBottom - chartTop + 1
	if chartW < 2 || chartH < 2 {
		return
	}

	// Draw grid
	if bc.showGrid {
		bc.drawVerticalGrid(buf, chartX, chartTop, chartW, chartH, maxVal)
	}

	// Draw Y-axis labels
	if bc.showAxes {
		numLabels := 5
		for i := 0; i <= numLabels; i++ {
			y := chartTop + (chartH-1) - (i*(chartH-1))/numLabels
			if y < chartTop || y > chartBottom {
				continue
			}
			val := maxVal * float64(i) / float64(numLabels)
			label := formatBarVal(val)
			labelX := chartX - len(label) - 1
			for j, r := range label {
				buf.SetCell(labelX+j, y, buffer.Cell{
					Rune:  r,
					Width: 1,
					Fg:    bc.axisStyle.Fg,
					Bg:    bg,
				})
			}
		}
	}

	// Draw bars
	groupW := numSeries + bc.gap
	if groupW < 1 {
		groupW = 1
	}
	totalW := numCategories * groupW
	if totalW > chartW {
		// Truncate categories that don't fit
		numCategories = chartW / groupW
	}

	for catIdx := 0; catIdx < numCategories; catIdx++ {
		groupStart := chartX + catIdx*groupW
		for sIdx, s := range bc.series {
			if catIdx >= len(s.Data) {
				continue
			}
			barX := groupStart + sIdx
			if barX >= chartX+chartW {
				break
			}
			val := s.Data[catIdx].Value
			if val <= 0 {
				continue
			}
			bc.drawVerticalBar(buf, barX, chartTop, chartH, val, maxVal, s.Color, bg)
		}

		// Category label
		if bc.showAxes {
			label := ""
			if catIdx < len(bc.series[0].Data) {
				label = bc.series[0].Data[catIdx].Label
			}
			if len(label) > groupW {
				label = label[:groupW]
			}
			labelX := groupStart
			for j, r := range label {
				if labelX+j >= chartX+chartW {
					break
				}
				buf.SetCell(labelX+j, chartBottom+1, buffer.Cell{
					Rune:  r,
					Width: 1,
					Fg:    bc.axisStyle.Fg,
					Bg:    bg,
				})
			}
		}
	}

	// X-axis line
	if bc.showAxes {
		for x := chartX; x < chartX+chartW; x++ {
			c := buf.GetCell(x, chartBottom)
			if c.Rune == ' ' || c.Rune == 0 {
				buf.SetCell(x, chartBottom, buffer.Cell{
					Rune:  '─',
					Width: 1,
					Fg:    bc.axisStyle.Fg,
					Bg:    bg,
				})
			}
		}
	}
}

// drawVerticalBar draws a single vertical bar at column x.
func (bc *BarChart) drawVerticalBar(buf *buffer.Buffer, x, top, height int, val, maxVal float64, color buffer.Color, bg buffer.Color) {
	ratio := val / maxVal
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	totalUnits := float64(height) * 8.0 // 8 sub-bar levels per cell
	barUnits := ratio * totalUnits
	fullCells := int(barUnits) / 8
	partial := int(barUnits) % 8

	bottom := top + height - 1

	// Fill full cells
	for i := 0; i < fullCells; i++ {
		y := bottom - i
		if y < top {
			break
		}
		buf.SetCell(x, y, buffer.Cell{
			Rune:  '█',
			Width: 1,
			Fg:    color,
			Bg:    bg,
		})
	}

	// Partial cell
	if partial > 0 {
		y := bottom - fullCells
		if y >= top {
			partialChars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇'}
			idx := partial - 1
			if idx >= len(partialChars) {
				idx = len(partialChars) - 1
			}
			buf.SetCell(x, y, buffer.Cell{
				Rune:  partialChars[idx],
				Width: 1,
				Fg:    color,
				Bg:    bg,
			})
		}
	}

	// Value label on top
	if bc.showValues {
		label := formatBarVal(val)
		y := bottom - fullCells
		if y >= top {
			y--
		}
		if y >= top {
			labelRunes := []rune(label)
			startX := x
			for j := range labelRunes {
				if startX+j >= x+1 {
					break
				}
				buf.SetCell(startX+j, y, buffer.Cell{
					Rune:  labelRunes[j],
					Width: 1,
					Fg:    color,
					Bg:    bg,
				})
			}
		}
	}
}

// drawVerticalGrid draws horizontal grid lines for vertical bars.
func (bc *BarChart) drawVerticalGrid(buf *buffer.Buffer, cx, cy, cw, ch int, maxVal float64) {
	numLines := 5
	for i := 0; i <= numLines; i++ {
		y := cy + (ch - 1) - (i*(ch-1))/numLines
		if y < cy || y > cy+ch-1 {
			continue
		}
		for x := cx; x < cx+cw; x++ {
			c := buf.GetCell(x, y)
			if c.Rune != ' ' && c.Rune != 0 {
				continue
			}
			buf.SetCell(x, y, buffer.Cell{
				Rune:  '·',
				Width: 1,
				Fg:    bc.gridStyle.Fg,
				Bg:    bc.gridStyle.Bg,
			})
		}
	}
}

// --- Horizontal rendering ---

func (bc *BarChart) paintHorizontal(buf *buffer.Buffer, bounds Rect, maxVal float64) {
	bg := bc.axisStyle.Bg
	numSeries := len(bc.series)

	// Determine max label width
	labelW := 0
	for _, s := range bc.series {
		for _, d := range s.Data {
			if len(d.Label) > labelW {
				labelW = len(d.Label)
			}
		}
	}
	if labelW > 20 {
		labelW = 20
	}
	if labelW < 5 {
		labelW = 5
	}

	row := bounds.Y

	// Title
	if bc.showTitle && bc.title != "" {
		titleRunes := []rune(bc.title)
		titleX := bounds.X + (bounds.W-len(titleRunes))/2
		if titleX < bounds.X {
			titleX = bounds.X
		}
		for i, r := range titleRunes {
			if titleX+i >= bounds.X+bounds.W {
				break
			}
			buf.SetCell(titleX+i, row, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    bc.titleStyle.Fg,
				Bg:    bg,
				Flags: bc.titleStyle.Flags,
			})
		}
		row++
	}

	// Legend
	if bc.showLegend && numSeries > 1 {
		x := bounds.X
		for _, s := range bc.series {
			entry := fmt.Sprintf("█ %s", s.Name)
			if x+len(entry) > bounds.X+bounds.W {
				break
			}
			for _, r := range entry {
				buf.SetCell(x, row, buffer.Cell{
					Rune:  r,
					Width: 1,
					Fg:    s.Color,
					Bg:    bg,
				})
				x++
			}
			x += 2
		}
		row++
	}

	// Collect all bars across series
	type flatBar struct {
		label string
		val   float64
		color buffer.Color
	}
	var bars []flatBar
	for _, s := range bc.series {
		for _, d := range s.Data {
			bars = append(bars, flatBar{label: d.Label, val: d.Value, color: s.Color})
		}
	}

	availH := bounds.Y + bounds.H - row
	if availH < 1 {
		return
	}
	// Truncate if needed
	if len(bars) > availH {
		bars = bars[:availH]
	}

	barStartX := bounds.X + labelW + 2
	barW := bounds.X + bounds.W - barStartX
	if barW < 2 {
		barW = bounds.W - labelW - 2
		if barW < 2 {
			return
		}
	}

	// Draw horizontal grid (vertical lines at intervals)
	if bc.showGrid {
		numLines := 5
		for i := 0; i <= numLines; i++ {
			x := barStartX + (i * barW) / numLines
			if x < barStartX || x >= barStartX+barW {
				continue
			}
			for y := row; y < row+len(bars); y++ {
				c := buf.GetCell(x, y)
				if c.Rune != ' ' && c.Rune != 0 {
					continue
				}
				buf.SetCell(x, y, buffer.Cell{
					Rune:  '·',
					Width: 1,
					Fg:    bc.gridStyle.Fg,
					Bg:    bc.gridStyle.Bg,
				})
			}
		}
	}

	// Draw each horizontal bar
	for i, b := range bars {
		y := row + i
		// Label
		label := b.label
		if len(label) > labelW {
			label = label[:labelW]
		}
		for j, r := range label {
			buf.SetCell(bounds.X+j, y, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    bc.axisStyle.Fg,
				Bg:    bg,
			})
		}

		// Bar
		ratio := b.val / maxVal
		if ratio < 0 {
			ratio = 0
		}
		if ratio > 1 {
			ratio = 1
		}
		totalUnits := float64(barW) * 8.0
		barUnits := ratio * totalUnits
		fullCells := int(barUnits) / 8
		partial := int(barUnits) % 8

		for j := 0; j < fullCells; j++ {
			x := barStartX + j
			if x >= barStartX+barW {
				break
			}
			buf.SetCell(x, y, buffer.Cell{
				Rune:  '█',
				Width: 1,
				Fg:    b.color,
				Bg:    bg,
			})
		}
		if partial > 0 {
			x := barStartX + fullCells
			if x < barStartX+barW {
				partialChars := []rune{'▏', '▎', '▍', '▌', '▋', '▊', '▉'}
				idx := partial - 1
				if idx >= len(partialChars) {
					idx = len(partialChars) - 1
				}
				buf.SetCell(x, y, buffer.Cell{
					Rune:  partialChars[idx],
					Width: 1,
					Fg:    b.color,
					Bg:    bg,
				})
			}
		}

		// Value label
		if bc.showValues {
			valLabel := formatBarVal(b.val)
			labelX := barStartX + fullCells + 1
			if partial > 0 {
				labelX++
			}
			for j, r := range valLabel {
				if labelX+j >= barStartX+barW {
					break
				}
				buf.SetCell(labelX+j, y, buffer.Cell{
					Rune:  r,
					Width: 1,
					Fg:    bc.axisStyle.Fg,
					Bg:    bg,
				})
			}
		}
	}

	// X-axis labels at bottom
	if bc.showAxes {
		axisY := row + len(bars)
		if axisY < bounds.Y+bounds.H {
			numLabels := 5
			for i := 0; i <= numLabels; i++ {
				x := barStartX + (i*barW)/numLabels
				if x < barStartX || x >= barStartX+barW {
					continue
				}
				val := maxVal * float64(i) / float64(numLabels)
				label := formatBarVal(val)
				for j, r := range label {
					if x+j >= barStartX+barW {
						break
					}
					buf.SetCell(x+j, axisY, buffer.Cell{
						Rune:  r,
						Width: 1,
						Fg:    bc.axisStyle.Fg,
						Bg:    bg,
					})
				}
			}
		}
	}
}

// Children returns nil (leaf component).
func (bc *BarChart) Children() []Component { return nil }

// --- Helpers ---

// formatBarVal formats a numeric value for axis/label display.
func formatBarVal(v float64) string {
	if v == 0 {
		return "0"
	}
	abs := math.Abs(v)
	if abs >= 1e9 {
		return fmt.Sprintf("%.1fB", v/1e9)
	}
	if abs >= 1e6 {
		return fmt.Sprintf("%.1fM", v/1e6)
	}
	if abs >= 1e3 {
		return fmt.Sprintf("%.1fK", v/1e3)
	}
	if abs >= 100 {
		return fmt.Sprintf("%.0f", v)
	}
	if abs >= 10 {
		return fmt.Sprintf("%.1f", v)
	}
	return fmt.Sprintf("%.2f", v)
}

// barAutoColors returns a palette of distinct colors for auto-coloring series.
func barAutoColors(n int) []buffer.Color {
	palette := []buffer.Color{
		buffer.NamedColor(buffer.NamedCyan),
		buffer.NamedColor(buffer.NamedGreen),
		buffer.NamedColor(buffer.NamedYellow),
		buffer.NamedColor(buffer.NamedMagenta),
		buffer.NamedColor(buffer.NamedRed),
		buffer.NamedColor(buffer.NamedBlue),
		buffer.NamedColor(buffer.NamedBrightCyan),
		buffer.NamedColor(buffer.NamedBrightGreen),
	}
	result := make([]buffer.Color, n)
	for i := 0; i < n; i++ {
		result[i] = palette[i%len(palette)]
	}
	return result
}

// String returns a debug string for the BarChart.
func (bc *BarChart) String() string {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	var sb strings.Builder
	fmt.Fprintf(&sb, "BarChart{title=%q, series=%d, orientation=%v", bc.title, len(bc.series), bc.orientation)
	if bc.maxVal > 0 {
		fmt.Fprintf(&sb, ", maxVal=%.2f", bc.maxVal)
	}
	sb.WriteString("}")
	return sb.String()
}

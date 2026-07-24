package component

import (
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// HeatmapCell represents a single cell in the heatmap grid.
type HeatmapCell struct {
	Value int    // intensity value (0 = empty)
	Label string // optional tooltip/label
}

// Heatmap renders a grid of colored cells with varying intensity,
// similar to a GitHub contribution graph. Useful for visualizing
// data density, activity patterns, or token usage over time.
//
// Features:
//   - Configurable rows × cols grid
//   - Intensity-based coloring (5 levels: empty, low, medium, high, peak)
//   - Customizable color palette
//   - Column labels (e.g., day names)
//   - Thread-safe
type Heatmap struct {
	BaseComponent
	mu sync.Mutex

	rows    int
	cols    int
	data    [][]HeatmapCell // [row][col]
	colLabels []string       // optional labels for columns
	maxVal  int
}

// NewHeatmap creates a heatmap with the given dimensions.
// Data cells default to zero intensity.
func NewHeatmap(rows, cols int) *Heatmap {
	if rows < 1 {
		rows = 1
	}
	if cols < 1 {
		cols = 1
	}
	data := make([][]HeatmapCell, rows)
	for i := range data {
		data[i] = make([]HeatmapCell, cols)
	}
	return &Heatmap{
		BaseComponent: BaseComponent{id: GenerateID("heatmap")},
		rows:          rows,
		cols:          cols,
		data:          data,
	}
}

// SetData replaces the entire data grid. The grid is resized to match.
func (h *Heatmap) SetData(data [][]HeatmapCell) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.data = data
	h.rows = len(data)
	if h.rows > 0 {
		h.cols = len(data[0])
	}
	h.recomputeMaxLocked()
}

// SetCell sets a single cell's value.
func (h *Heatmap) SetCell(row, col int, value int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if row < 0 || row >= h.rows || col < 0 || col >= h.cols {
		return
	}
	h.data[row][col].Value = value
	if value > h.maxVal {
		h.maxVal = value
	}
}

// Data returns a copy of the current data grid.
func (h *Heatmap) Data() [][]HeatmapCell {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([][]HeatmapCell, len(h.data))
	for i := range h.data {
		out[i] = make([]HeatmapCell, len(h.data[i]))
		copy(out[i], h.data[i])
	}
	return out
}

// SetColLabels sets optional labels for columns (e.g., day names).
func (h *Heatmap) SetColLabels(labels []string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.colLabels = labels
}

// Dimensions returns the current rows × cols.
func (h *Heatmap) Dimensions() (int, int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.rows, h.cols
}

// MaxValue returns the maximum value in the grid.
func (h *Heatmap) MaxValue() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.maxVal
}

// recomputeMaxLocked recalculates max value (caller holds lock).
func (h *Heatmap) recomputeMaxLocked() {
	h.maxVal = 0
	for _, row := range h.data {
		for _, cell := range row {
			if cell.Value > h.maxVal {
				h.maxVal = cell.Value
			}
		}
	}
}

// Measure returns the desired size.
func (h *Heatmap) Measure(cs Constraints) Size {
	h.mu.Lock()
	defer h.mu.Unlock()

	w := h.cols * 2 // each cell is 2 chars wide
	if len(h.colLabels) > 0 {
		w = h.cols * 2
	}
	maxW := cs.MaxWidth
	if maxW > 0 && w > maxW {
		w = maxW
	}
	h2 := h.rows
	if len(h.colLabels) > 0 {
		h2++ // space for labels
	}
	if h2 < 1 {
		h2 = 1
	}
	maxH := cs.MaxHeight
	if maxH > 0 && h2 > maxH {
		h2 = maxH
	}
	return Size{W: w, H: h2}
}

// intensityLevel maps a value to one of 5 intensity levels (0-4).
func intensityLevel(value, max int) int {
	if max <= 0 || value <= 0 {
		return 0
	}
	ratio := float64(value) / float64(max)
	switch {
	case ratio > 0.75:
		return 4
	case ratio > 0.50:
		return 3
	case ratio > 0.25:
		return 2
	default:
		return 1
	}
}

// Paint renders the heatmap.
func (h *Heatmap) Paint(buf *buffer.Buffer) {
	h.mu.Lock()
	data := h.data
	colLabels := h.colLabels
	maxVal := h.maxVal
	h.mu.Unlock()

	if len(data) == 0 {
		return
	}

	b := h.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	th := theme.Get()
	// Intensity colors: 0=none, 1=dim, 2=medium, 3=bright, 4=peak
	levels := [5]buffer.Color{
		th.Border, // empty
		th.Muted,  // low
		th.Accent, // medium
		th.Accent, // high (brighter accent)
		th.Accent, // peak (brightest)
	}

	y := b.Y
	labelRow := len(colLabels) > 0
	if labelRow {
		mutedStyle := buffer.Style{Fg: th.Muted}
		x := b.X
		for i, label := range colLabels {
			if i >= len(data[0]) {
				break
			}
			lw := utf8.RuneCountInString(label)
			if lw > 2 {
				label = truncateRunes(label, 1)
			}
			buf.DrawText(x, y, label, mutedStyle)
			x += 2
		}
		y++
	}

	for _, row := range data {
		if y >= b.Y+b.H {
			break
		}
		x := b.X
		for _, cell := range row {
			if x+1 >= b.X+b.W {
				break
			}
			level := intensityLevel(cell.Value, maxVal)
			bg := levels[level]
			if level > 0 {
				buf.SetCell(x, y, buffer.Cell{
					Rune:  ' ',
					Width: 1,
					Bg:    bg,
				})
				buf.SetCell(x+1, y, buffer.Cell{
					Rune:  ' ',
					Width: 1,
					Bg:    bg,
				})
			} else {
				buf.SetCell(x, y, buffer.Cell{
					Rune:  ' ',
					Width: 1,
					Bg:    levels[0],
				})
				buf.SetCell(x+1, y, buffer.Cell{
					Rune:  ' ',
					Width: 1,
					Bg:    levels[0],
				})
			}
			x += 2
		}
		y++
	}
}

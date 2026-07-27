package component

import (
	"strconv"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── HeatmapGrid: Calendar-Style Heatmap (GitHub Contributions) ───
//
// HeatmapGrid renders a grid of colored cells where intensity represents
// value. Common for activity tracking, contribution graphs, and density maps.
//
// Usage:
//
//	hg := NewHeatmapGrid(7, 20) // 7 rows × 20 weeks
//	hg.Set(0, 0, 5)  // Monday week 1: 5 commits
//	hg.Set(0, 1, 12) // Monday week 2: 12 commits
//	hg.SetMaxValue(20)
//	hg.SetBounds(Rect{X:0, Y:0, W:50, H:10})
//	hg.Paint(buf)

// HeatmapGridStyle holds visual styles.
type HeatmapGridStyle struct {
	Level0 buffer.Style // empty
	Level1 buffer.Style // low
	Level2 buffer.Style // medium-low
	Level3 buffer.Style // medium-high
	Level4 buffer.Style // high
	Empty  buffer.Style // no data
}

// DefaultHeatmapGridStyle returns GitHub-style green palette.
func DefaultHeatmapGridStyle() HeatmapGridStyle {
	return HeatmapGridStyle{
		Level0: buffer.Style{Fg: buffer.RGB(30, 30, 30)},
		Level1: buffer.Style{Fg: buffer.RGB(14, 68, 41)},
		Level2: buffer.Style{Fg: buffer.RGB(0, 109, 50)},
		Level3: buffer.Style{Fg: buffer.RGB(38, 166, 65)},
		Level4: buffer.Style{Fg: buffer.RGB(57, 211, 83)},
		Empty:  buffer.Style{Fg: buffer.RGB(20, 20, 20)},
	}
}

// HeatmapGrid renders a calendar-style heatmap grid.
type HeatmapGrid struct {
	BaseComponent
	mu       sync.RWMutex
	rows     int
	cols     int
	data     map[int]float64 // key = row*1000 + col
	maxValue float64
	style    HeatmapGridStyle
	cellW    int // width per cell (default 2)
	cellGap  int // gap between cells (default 1)
}

// NewHeatmapGrid creates a heatmap with the given dimensions.
func NewHeatmapGrid(rows, cols int) *HeatmapGrid {
	if rows < 1 {
		rows = 1
	}
	if cols < 1 {
		cols = 1
	}
	hg := &HeatmapGrid{
		rows:    rows,
		cols:    cols,
		data:    make(map[int]float64),
		maxValue: 10,
		style:   DefaultHeatmapGridStyle(),
		cellW:   2,
		cellGap: 1,
	}
	hg.SetID(GenerateID("heatmap"))
	return hg
}

// Set sets the value at a specific row/col position.
func (hg *HeatmapGrid) Set(row, col int, value float64) *HeatmapGrid {
	hg.mu.Lock()
	if row >= 0 && row < hg.rows && col >= 0 && col < hg.cols {
		hg.data[row*10000+col] = value
		if value > hg.maxValue {
			hg.maxValue = value
		}
	}
	hg.mu.Unlock()
	return hg
}

// Get returns the value at a position (0 if not set).
func (hg *HeatmapGrid) Get(row, col int) float64 {
	hg.mu.RLock()
	defer hg.mu.RUnlock()
	return hg.data[row*10000+col]
}

// SetMaxValue sets the maximum value for intensity scaling.
func (hg *HeatmapGrid) SetMaxValue(v float64) *HeatmapGrid {
	hg.mu.Lock()
	if v > 0 {
		hg.maxValue = v
	}
	hg.mu.Unlock()
	return hg
}

// MaxValue returns the max value.
func (hg *HeatmapGrid) MaxValue() float64 {
	hg.mu.RLock()
	defer hg.mu.RUnlock()
	return hg.maxValue
}

// Rows returns the number of rows.
func (hg *HeatmapGrid) Rows() int {
	hg.mu.RLock()
	defer hg.mu.RUnlock()
	return hg.rows
}

// Cols returns the number of columns.
func (hg *HeatmapGrid) Cols() int {
	hg.mu.RLock()
	defer hg.mu.RUnlock()
	return hg.cols
}

// CellCount returns rows * cols.
func (hg *HeatmapGrid) CellCount() int {
	hg.mu.RLock()
	defer hg.mu.RUnlock()
	return hg.rows * hg.cols
}

// FilledCount returns the number of cells with non-zero values.
func (hg *HeatmapGrid) FilledCount() int {
	hg.mu.RLock()
	defer hg.mu.RUnlock()
	count := 0
	for _, v := range hg.data {
		if v > 0 {
			count++
		}
	}
	return count
}

// TotalValue returns the sum of all cell values.
func (hg *HeatmapGrid) TotalValue() float64 {
	hg.mu.RLock()
	defer hg.mu.RUnlock()
	total := 0.0
	for _, v := range hg.data {
		total += v
	}
	return total
}

// Clear resets all data.
func (hg *HeatmapGrid) Clear() *HeatmapGrid {
	hg.mu.Lock()
	hg.data = make(map[int]float64)
	hg.mu.Unlock()
	return hg
}

// SetCellSize sets cell width and gap.
func (hg *HeatmapGrid) SetCellSize(w, gap int) *HeatmapGrid {
	hg.mu.Lock()
	if w >= 1 {
		hg.cellW = w
	}
	if gap >= 0 {
		hg.cellGap = gap
	}
	hg.mu.Unlock()
	return hg
}

// SetStyle sets the visual style.
func (hg *HeatmapGrid) SetStyle(s HeatmapGridStyle) *HeatmapGrid {
	hg.mu.Lock()
	hg.style = s
	hg.mu.Unlock()
	return hg
}

// Measure computes the desired size.
func (hg *HeatmapGrid) Measure(cs Constraints) Size {
	hg.mu.RLock()
	defer hg.mu.RUnlock()
	w := hg.cols*(hg.cellW+hg.cellGap) + 2
	h := hg.rows + 2
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// levelForValue returns the intensity level 0-4 for a value.
func levelForValue(value, max float64) int {
	if value <= 0 || max <= 0 {
		return 0
	}
	ratio := value / max
	switch {
	case ratio > 0.75:
		return 4
	case ratio > 0.5:
		return 3
	case ratio > 0.25:
		return 2
	default:
		return 1
	}
}

// Paint renders the heatmap.
func (hg *HeatmapGrid) Paint(buf *buffer.Buffer) {
	hg.mu.Lock()
	defer hg.mu.Unlock()

	b := hg.bounds
	if b.W < 4 || b.H < 2 {
		return
	}

	stride := hg.cellW + hg.cellGap
	for row := 0; row < hg.rows; row++ {
		y := b.Y + row
		if y >= b.Y+b.H {
			break
		}
		for col := 0; col < hg.cols; col++ {
			x := b.X + col*stride
			if x+hg.cellW > b.X+b.W {
				break
			}

			value := hg.data[row*10000+col]
			level := levelForValue(value, hg.maxValue)

			var style buffer.Style
			switch level {
			case 0:
				style = hg.style.Level0
			case 1:
				style = hg.style.Level1
			case 2:
				style = hg.style.Level2
			case 3:
				style = hg.style.Level3
			case 4:
				style = hg.style.Level4
			}

			for dx := 0; dx < hg.cellW; dx++ {
				ax := x + dx
				if ax >= b.X+b.W {
					break
				}
				buf.SetCell(ax, y, buffer.Cell{Rune: '█', Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
			}
		}
	}
}

// Children returns nil.
func (hg *HeatmapGrid) Children() []Component { return nil }

// Legend returns the legend text for each level.
func (hg *HeatmapGrid) Legend() string {
	return "Less " + strconv.Itoa(int(hg.maxValue*0.25)) + " " +
		strconv.Itoa(int(hg.maxValue*0.5)) + " " +
		strconv.Itoa(int(hg.maxValue*0.75)) + " " +
		strconv.Itoa(int(hg.maxValue)) + " More"
}

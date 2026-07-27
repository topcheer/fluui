package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── GanttChart: Project Timeline / Task Gantt Chart ───
//
// GanttChart renders a horizontal timeline of tasks with start/end dates,
// useful for project management dashboards. Each task is a colored bar
// positioned along a time axis.
//
// Usage:
//
//	gc := NewGanttChart()
//	gc.SetTimeline("2024-01-01", "2024-03-31")
//	gc.AddTask(GanttTask{Label: "Design", Start: 0, End: 14, Color: buffer.Cyan})
//	gc.AddTask(GanttTask{Label: "Build", Start: 10, End: 40, Color: buffer.Green})
//	gc.AddTask(GanttTask{Label: "Test", Start: 35, End: 55, Color: buffer.Yellow})

// GanttTask represents a single task bar on the chart.
type GanttTask struct {
	Label    string
	Start    int        // start position in timeline units (days, weeks, etc.)
	End      int        // end position (exclusive)
	Color    buffer.Color
	Progress float64    // 0.0 to 1.0 (drawn as filled portion of the bar)
}

// GanttChartStyle holds visual styles.
type GanttChartStyle struct {
	Header     buffer.Style
	Grid       buffer.Style
	TaskLabel  buffer.Style
	BarBg      buffer.Style
	BarOutline buffer.Style
}

// DefaultGanttChartStyle returns sensible defaults.
func DefaultGanttChartStyle() GanttChartStyle {
	return GanttChartStyle{
		Header:     buffer.Style{Fg: buffer.White, Flags: buffer.Bold},
		Grid:       buffer.Style{Fg: buffer.RGB(60, 60, 60)},
		TaskLabel:  buffer.Style{Fg: buffer.White},
		BarBg:      buffer.Style{Fg: buffer.RGB(40, 40, 40)},
		BarOutline: buffer.Style{Fg: buffer.RGB(100, 100, 100)},
	}
}

// GanttChart renders a Gantt timeline chart.
type GanttChart struct {
	BaseComponent
	mu       sync.RWMutex
	tasks    []GanttTask
	maxUnit  int        // total timeline span
	style    GanttChartStyle
	labelW   int        // width reserved for task labels (default 12)
}

// NewGanttChart creates a Gantt chart with default settings.
func NewGanttChart() *GanttChart {
	gc := &GanttChart{
		maxUnit: 60,
		style:   DefaultGanttChartStyle(),
		labelW:  12,
	}
	gc.SetID(GenerateID("gantt"))
	return gc
}

// AddTask adds a task to the chart.
func (gc *GanttChart) AddTask(t GanttTask) *GanttChart {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	if t.End > gc.maxUnit {
		gc.maxUnit = t.End
	}
	if t.Color.Type == 0 {
		t.Color = buffer.Cyan
	}
	gc.tasks = append(gc.tasks, t)
	return gc
}

// SetTasks replaces all tasks.
func (gc *GanttChart) SetTasks(tasks []GanttTask) *GanttChart {
	gc.mu.Lock()
	defer gc.mu.Unlock()
	gc.tasks = tasks
	for _, t := range tasks {
		if t.End > gc.maxUnit {
			gc.maxUnit = t.End
		}
	}
	return gc
}

// Tasks returns the current tasks.
func (gc *GanttChart) Tasks() []GanttTask {
	gc.mu.RLock()
	defer gc.mu.RUnlock()
	return gc.tasks
}

// TaskCount returns the number of tasks.
func (gc *GanttChart) TaskCount() int {
	gc.mu.RLock()
	defer gc.mu.RUnlock()
	return len(gc.tasks)
}

// SetMaxUnit sets the total timeline span.
func (gc *GanttChart) SetMaxUnit(n int) *GanttChart {
	gc.mu.Lock()
	gc.maxUnit = n
	gc.mu.Unlock()
	return gc
}

// MaxUnit returns the total timeline span.
func (gc *GanttChart) MaxUnit() int {
	gc.mu.RLock()
	defer gc.mu.RUnlock()
	return gc.maxUnit
}

// SetLabelWidth sets the width reserved for task labels.
func (gc *GanttChart) SetLabelWidth(w int) *GanttChart {
	gc.mu.Lock()
	if w >= 4 {
		gc.labelW = w
	}
	gc.mu.Unlock()
	return gc
}

// LabelWidth returns the label column width.
func (gc *GanttChart) LabelWidth() int {
	gc.mu.RLock()
	defer gc.mu.RUnlock()
	return gc.labelW
}

// SetStyle sets the visual style.
func (gc *GanttChart) SetStyle(s GanttChartStyle) *GanttChart {
	gc.mu.Lock()
	gc.style = s
	gc.mu.Unlock()
	return gc
}

// Style returns the current style.
func (gc *GanttChart) Style() GanttChartStyle {
	gc.mu.RLock()
	defer gc.mu.RUnlock()
	return gc.style
}

// Clear removes all tasks.
func (gc *GanttChart) Clear() *GanttChart {
	gc.mu.Lock()
	gc.tasks = gc.tasks[:0]
	gc.mu.Unlock()
	return gc
}

// unitToX converts a timeline unit to an x-offset within the chart area.
func (gc *GanttChart) unitToX(unit, chartW int) int {
	if gc.maxUnit <= 0 {
		return 0
	}
	return unit * chartW / gc.maxUnit
}

// Measure computes the desired size.
func (gc *GanttChart) Measure(cs Constraints) Size {
	gc.mu.RLock()
	defer gc.mu.RUnlock()
	w := gc.labelW + 40
	h := len(gc.tasks) + 2 // header + tasks
	if h < 3 {
		h = 3
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the Gantt chart.
func (gc *GanttChart) Paint(buf *buffer.Buffer) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	b := gc.bounds
	if b.W < 10 || b.H < 2 {
		return
	}

	labelW := gc.labelW
	if labelW > b.W-4 {
		labelW = b.W - 4
	}
	chartW := b.W - labelW - 1 // 1 for separator
	if chartW < 2 {
		return
	}

	row := b.Y

	// Header: timeline scale
	for x := 0; x < chartW; x++ {
		ch := ' '
		if x%10 == 0 && x < gc.maxUnit {
			ch = '|'
		} else if x%5 == 0 {
			ch = '·'
		}
		buf.SetCell(b.X+labelW+1+x, row, buffer.Cell{Rune: ch, Fg: gc.style.Grid.Fg, Bg: gc.style.Grid.Bg, Width: 1})
	}
	row++

	// Separator
	for x := 0; x < b.W; x++ {
		buf.SetCell(b.X+x, row, buffer.Cell{Rune: '─', Fg: gc.style.Grid.Fg, Bg: gc.style.Grid.Bg, Width: 1})
	}
	row++

	// Task rows
	for _, task := range gc.tasks {
		if row >= b.Y+b.H {
			break
		}

		// Label
		label := task.Label
		if len(label) > labelW {
			label = label[:labelW-1] + "~"
		}
		for i, r := range label {
			if i >= labelW {
				break
			}
			buf.SetCell(b.X+i, row, buffer.Cell{Rune: r, Fg: gc.style.TaskLabel.Fg, Bg: gc.style.TaskLabel.Bg, Flags: gc.style.TaskLabel.Flags, Width: 1})
		}
		// Pad remaining label area
		for i := len(label); i < labelW; i++ {
			buf.SetCell(b.X+i, row, buffer.Cell{Rune: ' ', Bg: gc.style.TaskLabel.Bg, Width: 1})
		}

		// Separator
		buf.SetCell(b.X+labelW, row, buffer.Cell{Rune: '│', Fg: gc.style.Grid.Fg, Bg: gc.style.Grid.Bg, Width: 1})

		// Bar
		startX := gc.unitToX(task.Start, chartW)
		endX := gc.unitToX(task.End, chartW)
		if endX <= startX {
			endX = startX + 1
		}
		barLen := endX - startX
		progressLen := int(float64(barLen) * task.Progress)
		if progressLen > barLen {
			progressLen = barLen
		}

		for i := 0; i < chartW; i++ {
			absX := b.X + labelW + 1 + i
			if i < startX || i >= endX {
				// Grid background
				if i%5 == 0 {
					buf.SetCell(absX, row, buffer.Cell{Rune: '·', Fg: gc.style.Grid.Fg, Bg: gc.style.Grid.Bg, Width: 1})
				} else {
					buf.SetCell(absX, row, buffer.Cell{Rune: ' ', Bg: gc.style.BarBg.Bg, Width: 1})
				}
			} else {
				// Bar area
				localIdx := i - startX
				if localIdx < progressLen {
					buf.SetCell(absX, row, buffer.Cell{Rune: '█', Fg: task.Color, Bg: gc.style.BarBg.Bg, Width: 1})
				} else {
					buf.SetCell(absX, row, buffer.Cell{Rune: '░', Fg: gc.style.BarOutline.Fg, Bg: gc.style.BarBg.Bg, Width: 1})
				}
			}
		}

		row++
	}
}

// String returns a debug representation.
func (gc *GanttChart) String() string {
	gc.mu.RLock()
	defer gc.mu.RUnlock()
	var sb strings.Builder
	sb.WriteString("GanttChart[")
	for _, t := range gc.tasks {
		sb.WriteString(t.Label)
		sb.WriteString(" ")
	}
	sb.WriteString("]")
	return sb.String()
}

// Children returns nil.
func (gc *GanttChart) Children() []Component { return nil }

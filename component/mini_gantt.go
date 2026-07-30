package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MiniGantt: Compact Gantt Chart ───
//
// MiniGantt renders a compact horizontal Gantt chart with tasks shown as
// colored bars on a time axis. Each task has a start position, duration,
// and label. Useful for project tracking in dashboards.
//
// Usage:
//
//	mg := NewMiniGantt()
//	mg.SetRange(0, 100)
//	mg.AddTask("Design", 0, 20, buffer.RGB(59, 130, 246))
//	mg.AddTask("Build", 15, 50, buffer.RGB(34, 197, 94))
//	mg.AddTask("Test", 60, 30, buffer.RGB(245, 158, 11))
//	mg.Paint(buf)

// MiniGanttTask holds a single task.
type MiniGanttTask struct {
	Label    string
	Start    int
	Duration int
	Color    buffer.Color
}

// MiniGanttStyle holds styling.
type MiniGanttStyle struct {
	Axis  buffer.Style
	Label buffer.Style
	Empty buffer.Style
}

// DefaultMiniGanttStyle returns defaults.
func DefaultMiniGanttStyle() MiniGanttStyle {
	return MiniGanttStyle{
		Axis:  buffer.Style{Fg: buffer.RGB(71, 85, 105)},
		Label: buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Empty: buffer.Style{Fg: buffer.RGB(30, 41, 59)},
	}
}

const miniGanttMaxTasks = 10

// MiniGantt renders a compact Gantt chart.
type MiniGantt struct {
	BaseComponent
	mu sync.Mutex

	tasks    [miniGanttMaxTasks]MiniGanttTask
	count    int
	rangeMin int
	rangeMax int
	width    int
	style    MiniGanttStyle
}

// NewMiniGantt creates a MiniGantt.
func NewMiniGantt() *MiniGantt {
	mg := &MiniGantt{rangeMin: 0, rangeMax: 100, width: 30, style: DefaultMiniGanttStyle()}
	mg.SetID(GenerateID("minigantt"))
	return mg
}

// SetRange sets the time range.
func (mg *MiniGantt) SetRange(minV, maxV int) *MiniGantt {
	mg.mu.Lock()
	if maxV <= minV {
		maxV = minV + 1
	}
	mg.rangeMin = minV
	mg.rangeMax = maxV
	mg.mu.Unlock()
	return mg
}

// AddTask adds a task to the chart.
func (mg *MiniGantt) AddTask(label string, start, duration int, color buffer.Color) *MiniGantt {
	mg.mu.Lock()
	if mg.count < miniGanttMaxTasks {
		if duration < 1 {
			duration = 1
		}
		if start < 0 {
			start = 0
		}
		mg.tasks[mg.count] = MiniGanttTask{Label: label, Start: start, Duration: duration, Color: color}
		mg.count++
	}
	mg.mu.Unlock()
	return mg
}

// Clear removes all tasks.
func (mg *MiniGantt) Clear() *MiniGantt {
	mg.mu.Lock()
	mg.count = 0
	mg.mu.Unlock()
	return mg
}

// TaskCount returns the number of tasks.
func (mg *MiniGantt) TaskCount() int {
	mg.mu.Lock()
	defer mg.mu.Unlock()
	return mg.count
}

// SetWidth sets the chart width.
func (mg *MiniGantt) SetWidth(w int) *MiniGantt {
	mg.mu.Lock()
	if w < 10 {
		w = 10
	}
	mg.width = w
	mg.mu.Unlock()
	return mg
}

// SetStyle sets custom style.
func (mg *MiniGantt) SetStyle(s MiniGanttStyle) *MiniGantt {
	mg.mu.Lock()
	mg.style = s
	mg.mu.Unlock()
	return mg
}

// Measure returns preferred size.
func (mg *MiniGantt) Measure(cs Constraints) Size {
	mg.mu.Lock()
	h := mg.count + 1 // tasks + axis
	mg.mu.Unlock()
	if h < 1 {
		h = 1
	}
	w := mg.width
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: h}
}

// Paint renders the mini Gantt chart.
func (mg *MiniGantt) Paint(buf *buffer.Buffer) {
	mg.mu.Lock()
	defer mg.mu.Unlock()

	b := mg.Bounds()
	x, y := b.X, b.Y
	w := mg.width

	axisStyle := mg.style.Axis
	labelStyle := mg.style.Label
	emptyStyle := mg.style.Empty

	rangeSpan := mg.rangeMax - mg.rangeMin

	// Axis row
	for i := 0; i < w; i++ {
		if x+i >= buf.Width {
			break
		}
		var r rune
		if i == 0 {
			r = '0'
		} else if i == w/2 {
			r = '├'
		} else if i == w-1 {
			r = '┤'
		} else {
			r = '─'
		}
		buf.SetCell(x+i, y, buffer.Cell{Rune: r, Fg: axisStyle.Fg, Bg: axisStyle.Bg, Flags: axisStyle.Flags, Width: 1})
	}

	// Task rows
	for t := 0; t < mg.count; t++ {
		yy := y + 1 + t
		if yy >= buf.Height {
			break
		}
		task := mg.tasks[t]
		barStyle := buffer.Style{Fg: task.Color}

		startCol := x + (task.Start-mg.rangeMin)*w/rangeSpan
		endCol := x + (task.Start+task.Duration-mg.rangeMin)*w/rangeSpan
		if endCol > x+w {
			endCol = x + w
		}
		if startCol < x {
			startCol = x
		}

		for col := x; col < x+w; col++ {
			if col >= buf.Width {
				break
			}
			if col >= startCol && col < endCol {
				buf.SetCell(col, yy, buffer.Cell{Rune: '█', Fg: barStyle.Fg, Bg: barStyle.Bg, Flags: barStyle.Flags, Width: 1})
			} else {
				buf.SetCell(col, yy, buffer.Cell{Rune: '░', Fg: emptyStyle.Fg, Bg: emptyStyle.Bg, Flags: emptyStyle.Flags, Width: 1})
			}
		}

		// Task label at the start of the bar
		col := startCol
		for _, r := range task.Label {
			if col >= buf.Width || col >= endCol {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: barStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (mg *MiniGantt) Children() []Component { return nil }

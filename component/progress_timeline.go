package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ProgressTimeline: Milestone Progress Tracker ───
//
// ProgressTimeline renders a horizontal timeline of milestones with
// completed/in-progress/pending states. Common in project tracking and
// release roadmap dashboards.
//
// Usage:
//
//	pt := NewProgressTimeline()
//	pt.AddMilestone(TimelineMilestone{Label: "Design", Status: MilestoneDone})
//	pt.AddMilestone(TimelineMilestone{Label: "Build", Status: MilestoneActive})
//	pt.AddMilestone(TimelineMilestone{Label: "Ship", Status: MilestonePending})
//	pt.SetBounds(Rect{X:0, Y:0, W:60, H:5})
//	pt.Paint(buf)

// MilestoneStatus represents the state of a milestone.
type MilestoneStatus int

const (
	MilestonePending MilestoneStatus = iota
	MilestoneActive
	MilestoneDone
)

// TimelineMilestone represents a single point on the timeline.
type TimelineMilestone struct {
	Label  string
	Status MilestoneStatus
}

// ProgressTimelineStyle holds visual styles.
type ProgressTimelineStyle struct {
	Done    buffer.Style
	Active  buffer.Style
	Pending buffer.Style
	Line    buffer.Style
	Label   buffer.Style
}

// DefaultProgressTimelineStyle returns sensible defaults.
func DefaultProgressTimelineStyle() ProgressTimelineStyle {
	return ProgressTimelineStyle{
		Done:    buffer.Style{Fg: buffer.RGB(16, 163, 127), Flags: buffer.Bold},
		Active:  buffer.Style{Fg: buffer.RGB(255, 175, 64), Flags: buffer.Bold},
		Pending: buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Line:    buffer.Style{Fg: buffer.RGB(60, 60, 60)},
		Label:   buffer.Style{Fg: buffer.White},
	}
}

// ProgressTimeline renders a milestone progress tracker.
type ProgressTimeline struct {
	BaseComponent
	mu        sync.RWMutex
	milestones []TimelineMilestone
	style      ProgressTimelineStyle
}

// NewProgressTimeline creates an empty timeline.
func NewProgressTimeline() *ProgressTimeline {
	pt := &ProgressTimeline{
		style: DefaultProgressTimelineStyle(),
	}
	pt.SetID(GenerateID("timeline"))
	return pt
}

// AddMilestone adds a milestone to the timeline.
func (pt *ProgressTimeline) AddMilestone(m TimelineMilestone) *ProgressTimeline {
	pt.mu.Lock()
	pt.milestones = append(pt.milestones, m)
	pt.mu.Unlock()
	return pt
}

// SetMilestones replaces all milestones.
func (pt *ProgressTimeline) SetMilestones(milestones []TimelineMilestone) *ProgressTimeline {
	pt.mu.Lock()
	pt.milestones = milestones
	pt.mu.Unlock()
	return pt
}

// Milestones returns the current milestones.
func (pt *ProgressTimeline) Milestones() []TimelineMilestone {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.milestones
}

// MilestoneCount returns the number of milestones.
func (pt *ProgressTimeline) MilestoneCount() int {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return len(pt.milestones)
}

// DoneCount returns the number of completed milestones.
func (pt *ProgressTimeline) DoneCount() int {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	count := 0
	for _, m := range pt.milestones {
		if m.Status == MilestoneDone {
			count++
		}
	}
	return count
}

// Progress returns the completion ratio (0.0 to 1.0).
func (pt *ProgressTimeline) Progress() float64 {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	n := len(pt.milestones)
	if n == 0 {
		return 0
	}
	done := 0
	for _, m := range pt.milestones {
		if m.Status == MilestoneDone {
			done++
		}
	}
	return float64(done) / float64(n)
}

// Clear removes all milestones.
func (pt *ProgressTimeline) Clear() *ProgressTimeline {
	pt.mu.Lock()
	pt.milestones = pt.milestones[:0]
	pt.mu.Unlock()
	return pt
}

// SetStyle sets the visual style.
func (pt *ProgressTimeline) SetStyle(s ProgressTimelineStyle) *ProgressTimeline {
	pt.mu.Lock()
	pt.style = s
	pt.mu.Unlock()
	return pt
}

// Style returns the current style.
func (pt *ProgressTimeline) Style() ProgressTimelineStyle {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	return pt.style
}

// Measure computes the desired size.
func (pt *ProgressTimeline) Measure(cs Constraints) Size {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	w := len(pt.milestones)*8 + 4
	if w < 20 {
		w = 20
	}
	h := 3
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the progress timeline.
func (pt *ProgressTimeline) Paint(buf *buffer.Buffer) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	b := pt.bounds
	if b.W < 6 || b.H < 2 || len(pt.milestones) == 0 {
		return
	}

	n := len(pt.milestones)
	spacing := b.W / n
	if spacing < 3 {
		spacing = 3
	}

	centerY := b.Y + b.H/2

	// Draw connecting line
	for i := 0; i < n-1; i++ {
		x0 := b.X + i*spacing + spacing/2 + 1
		x1 := b.X + (i+1)*spacing + spacing/2 - 1
		var lineStyle buffer.Style
		if pt.milestones[i].Status == MilestoneDone {
			lineStyle = pt.style.Done
		} else {
			lineStyle = pt.style.Line
		}
		for x := x0; x <= x1 && x < b.X+b.W; x++ {
			buf.SetCell(x, centerY, buffer.Cell{Rune: '─', Fg: lineStyle.Fg, Bg: lineStyle.Bg, Width: 1})
		}
	}

	// Draw milestones
	for i, m := range pt.milestones {
		x := b.X + i*spacing + spacing/2
		if x >= b.X+b.W {
			break
		}

		var glyph rune
		var style buffer.Style
		switch m.Status {
		case MilestoneDone:
			glyph = '●'
			style = pt.style.Done
		case MilestoneActive:
			glyph = '◐'
			style = pt.style.Active
		default:
			glyph = '○'
			style = pt.style.Pending
		}

		// Marker
		buf.SetCell(x, centerY, buffer.Cell{Rune: glyph, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})

		// Label below
		labelY := centerY + 1
		if labelY >= b.Y+b.H {
			labelY = b.Y + b.H - 1
		}
		labelRunes := []rune(m.Label)
		startX := x - len(labelRunes)/2
		if startX < b.X {
			startX = b.X
		}
		for j, r := range labelRunes {
			ax := startX + j
			if ax >= b.X+b.W {
				break
			}
			buf.SetCell(ax, labelY, buffer.Cell{Rune: r, Fg: pt.style.Label.Fg, Bg: pt.style.Label.Bg, Flags: pt.style.Label.Flags, Width: 1})
		}
	}
}

// Children returns nil.
func (pt *ProgressTimeline) Children() []Component { return nil }

package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── CompactTimeline: Horizontal Event Timeline ───
//
// CompactTimeline renders a horizontal timeline of events on a single row.
// Each event is positioned by timestamp relative to the time range.
// Uses dots/markers with color-coded types.
//
// Usage:
//
//	ct := NewCompactTimeline()
//	ct.SetRange(0, 100)
//	ct.AddEvent(25, 0)  // start, type=info
//	ct.AddEvent(60, 1)  // middle, type=warn
//	ct.AddEvent(95, 2)  // end, type=error
//	ct.Paint(buf)

// CompactTimelineStyle holds styling.
type CompactTimelineStyle struct {
	Info  buffer.Style
	Warn  buffer.Style
	Error buffer.Style
	Line  buffer.Style
}

// DefaultCompactTimelineStyle returns defaults.
func DefaultCompactTimelineStyle() CompactTimelineStyle {
	return CompactTimelineStyle{
		Info:  buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Bold},
		Warn:  buffer.Style{Fg: buffer.RGB(245, 158, 11), Flags: buffer.Bold},
		Error: buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Line:  buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

var timelineEventIcons = [...]rune{'●', '⚠', '✗'}

const compactTimelineMaxEvents = 20

// compactTimelineEvent holds a single event.
type compactTimelineEvent struct {
	position  int
	eventType int
}

// CompactTimeline renders a horizontal event timeline.
type CompactTimeline struct {
	BaseComponent
	mu sync.Mutex

	rangeMin int
	rangeMax int
	events   [compactTimelineMaxEvents]compactTimelineEvent
	count    int
	width    int
	style    CompactTimelineStyle
}

// NewCompactTimeline creates a CompactTimeline.
func NewCompactTimeline() *CompactTimeline {
	ct := &CompactTimeline{rangeMin: 0, rangeMax: 100, width: 40, style: DefaultCompactTimelineStyle()}
	ct.SetID(GenerateID("ctimeline"))
	return ct
}

// SetRange sets the min and max positions for the timeline.
func (ct *CompactTimeline) SetRange(minV, maxV int) *CompactTimeline {
	ct.mu.Lock()
	if maxV <= minV {
		maxV = minV + 1
	}
	ct.rangeMin = minV
	ct.rangeMax = maxV
	ct.mu.Unlock()
	return ct
}

// AddEvent adds an event at a position with a type (0=info, 1=warn, 2=error).
func (ct *CompactTimeline) AddEvent(position, eventType int) *CompactTimeline {
	ct.mu.Lock()
	if ct.count < compactTimelineMaxEvents {
		if eventType < 0 {
			eventType = 0
		}
		if eventType > 2 {
			eventType = 2
		}
		ct.events[ct.count] = compactTimelineEvent{position: position, eventType: eventType}
		ct.count++
	}
	ct.mu.Unlock()
	return ct
}

// Clear removes all events.
func (ct *CompactTimeline) Clear() *CompactTimeline {
	ct.mu.Lock()
	ct.count = 0
	ct.mu.Unlock()
	return ct
}

// EventCount returns the number of events.
func (ct *CompactTimeline) EventCount() int {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	return ct.count
}

// SetWidth sets the timeline width.
func (ct *CompactTimeline) SetWidth(w int) *CompactTimeline {
	ct.mu.Lock()
	if w < 10 {
		w = 10
	}
	ct.width = w
	ct.mu.Unlock()
	return ct
}

// SetStyle sets custom style.
func (ct *CompactTimeline) SetStyle(s CompactTimelineStyle) *CompactTimeline {
	ct.mu.Lock()
	ct.style = s
	ct.mu.Unlock()
	return ct
}

// Measure returns preferred size.
func (ct *CompactTimeline) Measure(cs Constraints) Size {
	w := ct.width
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the compact timeline.
func (ct *CompactTimeline) Paint(buf *buffer.Buffer) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	b := ct.Bounds()
	x, y := b.X, b.Y
	w := ct.width

	lineStyle := ct.style.Line
	infoStyle := ct.style.Info
	warnStyle := ct.style.Warn
	errorStyle := ct.style.Error

	// Draw base line
	rangeSpan := ct.rangeMax - ct.rangeMin
	for i := 0; i < w; i++ {
		if x+i >= buf.Width {
			break
		}
		buf.SetCell(x+i, y, buffer.Cell{Rune: '─', Fg: lineStyle.Fg, Bg: lineStyle.Bg, Flags: lineStyle.Flags, Width: 1})
	}

	// Draw events
	for i := 0; i < ct.count; i++ {
		ev := ct.events[i]
		normPos := ev.position - ct.rangeMin
		if normPos < 0 || normPos > rangeSpan {
			continue
		}
		col := x + normPos*w/rangeSpan
		if col >= buf.Width {
			continue
		}

		var st buffer.Style
		switch ev.eventType {
		case 0:
			st = infoStyle
		case 1:
			st = warnStyle
		case 2:
			st = errorStyle
		}
		buf.SetCell(col, y, buffer.Cell{Rune: timelineEventIcons[ev.eventType], Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
	}
}

// Children returns nil.
func (ct *CompactTimeline) Children() []Component { return nil }

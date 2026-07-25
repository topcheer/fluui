package component

import (
	"sync"
	"time"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// TimelineEvent represents a single event in a timeline.
type TimelineEvent struct {
	Time        time.Time
	Title       string
	Description string
	Type        TimelineEventType
}

// TimelineEventType controls the visual style of an event.
type TimelineEventType int

const (
	TimelineEventDefault TimelineEventType = iota
	TimelineEventSuccess
	TimelineEventWarning
	TimelineEventError
	TimelineEventInfo
)

// Timeline renders a vertical sequence of timestamped events,
// like a git log, CI pipeline, or AI reasoning chain. Each event
// shows a timestamp, marker, title, and optional description.
//
// Thread-safe. Zero-alloc Paint.
type Timeline struct {
	BaseComponent
	mu      sync.Mutex
	events  []TimelineEvent
	maxView int // max events to show (0 = all)
}

// NewTimeline creates a timeline with the given events.
func NewTimeline(events []TimelineEvent) *Timeline {
	return &Timeline{
		BaseComponent: BaseComponent{id: GenerateID("timeline")},
		events:        events,
	}
}

// AddEvent appends a single event.
func (t *Timeline) AddEvent(e TimelineEvent) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if e.Time.IsZero() {
		e.Time = time.Now()
	}
	t.events = append(t.events, e)
}

// SetEvents replaces all events.
func (t *Timeline) SetEvents(events []TimelineEvent) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = events
}

// Events returns a copy of the current events.
func (t *Timeline) Events() []TimelineEvent {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]TimelineEvent, len(t.events))
	copy(out, t.events)
	return out
}

// EventCount returns the number of events.
func (t *Timeline) EventCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.events)
}

// SetMaxView limits how many events to display (0 = unlimited).
func (t *Timeline) SetMaxView(n int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.maxView = n
}

// Clear removes all events.
func (t *Timeline) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = nil
}

// Measure returns the desired size.
func (t *Timeline) Measure(cs Constraints) Size {
	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 50
	}
	h := t.EventCount() * 2
	if h < 1 {
		h = 1
	}
	maxH := cs.MaxHeight
	if maxH > 0 && h > maxH {
		h = maxH
	}
	return Size{W: maxW, H: h}
}

// eventMarker returns the rune and color for an event type.
func eventMarker(et TimelineEventType) (string, buffer.Color) {
	th := theme.Get()
	switch et {
	case TimelineEventSuccess:
		return "\u2714", th.Success // ✔
	case TimelineEventWarning:
		return "\u26a0", th.Warning // ⚠
	case TimelineEventError:
		return "\u2716", th.Error // ✖
	case TimelineEventInfo:
		return "\u2139", th.Accent // ℹ
	default:
		return "\u25cf", th.Accent // ●
	}
}

// Paint renders the timeline.
func (t *Timeline) Paint(buf *buffer.Buffer) {
	t.mu.Lock()
	events := t.events
	maxView := t.maxView
	t.mu.Unlock()

	if len(events) == 0 {
		return
	}

	b := t.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	th := theme.Get()
	timeStyle := buffer.Style{Fg: th.Muted}
	titleStyle := buffer.Style{Fg: th.Fg}
	descStyle := buffer.Style{Fg: th.Muted}
	connectorStyle := buffer.Style{Fg: th.Border}

	// Limit visible events
	visible := events
	if maxView > 0 && len(visible) > maxView {
		visible = visible[len(visible)-maxView:]
	}

	x := b.X
	y := b.Y

	for i, ev := range visible {
		if y >= b.Y+b.H {
			break
		}

		// Timestamp
		var tbuf [16]byte
		tb := ev.Time.AppendFormat(tbuf[:0], "15:04:05")
		buf.DrawText(x, y, string(tb), timeStyle)
		tx := x + 9

		// Marker
		marker, mcolor := eventMarker(ev.Type)
		markerStyle := buffer.Style{Fg: mcolor}
		buf.DrawText(tx, y, marker, markerStyle)
		tx += 2

		// Title
		availW := b.X + b.W - tx
		title := ev.Title
		if utf8.RuneCountInString(title) > availW && availW > 2 {
			title = truncateRunes(title, availW-1) + "\u2026"
		}
		buf.DrawText(tx, y, title, titleStyle)
		y++

		// Description (if present and space available)
		if ev.Description != "" && y < b.Y+b.H {
			dx := tx
			desc := ev.Description
			descAvailW := b.X + b.W - dx
			if utf8.RuneCountInString(desc) > descAvailW && descAvailW > 2 {
				desc = truncateRunes(desc, descAvailW-1) + "\u2026"
			}
			buf.DrawText(dx, y, desc, descStyle)
			y++
		}

		// Vertical connector (except last event)
		if i < len(visible)-1 && y < b.Y+b.H {
			buf.DrawText(tx, y, "\u2502", connectorStyle)
			y++
		}
	}
}

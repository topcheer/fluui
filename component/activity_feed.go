package component

import (
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ActivityFeed: Vertical Activity Timeline ───
//
// ActivityFeed renders a vertical timeline of activity entries with actor,
// action, and relative timestamp. Entries are shown newest-first with
// configurable max display count.
//
// Usage:
//
//	af := NewActivityFeed()
//	af.AddEntry("alice", "pushed commit", time.Now().Add(-5*time.Minute))
//	af.AddEntry("bob", "opened PR", time.Now().Add(-1*time.Hour))
//	af.Paint(buf)

// ActivityEntry represents a single activity feed entry.
type ActivityEntry struct {
	Actor     string
	Action    string
	Timestamp time.Time
	// cached display fields
	RelTime string
}

// ActivityFeedStyle holds styling.
type ActivityFeedStyle struct {
	Actor    buffer.Style
	Action   buffer.Style
	Time     buffer.Style
	Dot      buffer.Style
	Border   buffer.Style
}

// DefaultActivityFeedStyle returns defaults.
func DefaultActivityFeedStyle() ActivityFeedStyle {
	actor := buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Bold}
	action := buffer.Style{Fg: buffer.RGB(226, 232, 240)}
	tm := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	dot := buffer.Style{Fg: buffer.RGB(34, 197, 94)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return ActivityFeedStyle{Actor: actor, Action: action, Time: tm, Dot: dot, Border: border}
}

// formatRelativeTime returns a human-readable relative time string.
func formatRelativeTime(t, now time.Time) string {
	d := now.Sub(t)
	if d < time.Minute {
		return "now"
	}
	if d < time.Hour {
		return itoa(int(d.Minutes())) + "m"
	}
	if d < 24*time.Hour {
		return itoa(int(d.Hours())) + "h"
	}
	return itoa(int(d.Hours()/24)) + "d"
}

// ActivityFeed renders a vertical activity timeline.
type ActivityFeed struct {
	BaseComponent
	mu sync.Mutex

	entries    []ActivityEntry
	maxEntries int
	style      ActivityFeedStyle
}

// NewActivityFeed creates an ActivityFeed.
func NewActivityFeed() *ActivityFeed {
	af := &ActivityFeed{maxEntries: 20, style: DefaultActivityFeedStyle()}
	af.SetID(GenerateID("activity"))
	return af
}

// AddEntry adds an activity entry with relative time computed at insertion.
func (af *ActivityFeed) AddEntry(actor, action string, timestamp time.Time) *ActivityFeed {
	af.mu.Lock()
	af.entries = append(af.entries, ActivityEntry{
		Actor:     actor,
		Action:    action,
		Timestamp: timestamp,
		RelTime:   formatRelativeTime(timestamp, time.Now()),
	})
	af.mu.Unlock()
	return af
}

// EntryCount returns the number of entries.
func (af *ActivityFeed) EntryCount() int {
	af.mu.Lock()
	defer af.mu.Unlock()
	return len(af.entries)
}

// SetMaxEntries sets the maximum entries to display.
func (af *ActivityFeed) SetMaxEntries(n int) *ActivityFeed {
	af.mu.Lock()
	af.maxEntries = n
	af.mu.Unlock()
	return af
}

// SetStyle sets custom style.
func (af *ActivityFeed) SetStyle(s ActivityFeedStyle) *ActivityFeed {
	af.mu.Lock()
	af.style = s
	af.mu.Unlock()
	return af
}

// Clear removes all entries.
func (af *ActivityFeed) Clear() *ActivityFeed {
	af.mu.Lock()
	af.entries = af.entries[:0]
	af.mu.Unlock()
	return af
}

// Measure returns the preferred size.
func (af *ActivityFeed) Measure(cs Constraints) Size {
	af.mu.Lock()
	count := len(af.entries)
	mx := af.maxEntries
	af.mu.Unlock()
	if count > mx { count = mx }
	w := 50
	h := count + 2
	if h < 3 { h = 3 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the activity feed into the buffer.
func (af *ActivityFeed) Paint(buf *buffer.Buffer) {
	af.mu.Lock()
	defer af.mu.Unlock()

	b := af.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 { w = 50 }
	if h < 3 { h = 3 }

	bs := af.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	actorStyle := af.style.Actor
	actionStyle := af.style.Action
	timeStyle := af.style.Time
	dotStyle := af.style.Dot

	displayCount := len(af.entries)
	if displayCount > af.maxEntries { displayCount = af.maxEntries }

	for i := 0; i < displayCount; i++ {
		entry := af.entries[len(af.entries)-1-i] // newest first
		rowY := y + 1 + i
		if rowY >= y+h-1 || rowY >= buf.Height { break }

		col := x + 1

		// Timeline dot
		if col < buf.Width {
			buf.SetCell(col, rowY, buffer.Cell{Rune: '●', Fg: dotStyle.Fg, Bg: dotStyle.Bg, Flags: dotStyle.Flags, Width: 1})
		}
		col++
		if col < buf.Width {
			buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: dotStyle.Fg, Bg: dotStyle.Bg, Flags: dotStyle.Flags, Width: 1})
		}
		col++

		// Actor
		for _, r := range entry.Actor {
			if col >= x+w-8 || col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: actorStyle.Fg, Bg: actorStyle.Bg, Flags: actorStyle.Flags, Width: 1})
			col++
		}

		// Space
		if col < buf.Width {
			buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: actionStyle.Fg, Bg: actionStyle.Bg, Flags: actionStyle.Flags, Width: 1})
		}
		col++

		// Action
		for _, r := range entry.Action {
			if col >= x+w-6 || col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: actionStyle.Fg, Bg: actionStyle.Bg, Flags: actionStyle.Flags, Width: 1})
			col++
		}

		// Relative time (right-aligned)
		timeLen := len(entry.RelTime)
		timeStart := x + w - 2 - timeLen
		if timeStart < col { timeStart = col }
		for i2, r := range entry.RelTime {
			cx := timeStart + i2
			if cx >= x+w-1 || cx >= buf.Width { break }
			buf.SetCell(cx, rowY, buffer.Cell{Rune: r, Fg: timeStyle.Fg, Bg: timeStyle.Bg, Flags: timeStyle.Flags, Width: 1})
		}
	}
}

// Children returns nil.
func (af *ActivityFeed) Children() []Component { return nil }

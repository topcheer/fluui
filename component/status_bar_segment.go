package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── StatusBarSegment: Configurable Status Bar Items ───
//
// StatusBarSegment renders a horizontal status bar with colored background
// segments, each showing a label+value pair (like VS Code status bar items).
//
// Usage:
//
//	sb := NewStatusBarSegment()
//	sb.AddSegment("Branch", "main", buffer.RGB(255,255,255), buffer.RGB(34,197,94))
//	sb.AddSegment("Errors", "0", buffer.RGB(255,255,255), buffer.RGB(239,68,68))
//	sb.Paint(buf)

// StatusSegment represents a single status bar item.
type StatusSegment struct {
	Label string
	Value string
	Fg    buffer.Color
	Bg    buffer.Color
}

// StatusBarSegmentStyle holds border styling.
type StatusBarSegmentStyle struct {
	Separator buffer.Style
	Border    buffer.Style
}

// DefaultStatusBarSegmentStyle returns defaults.
func DefaultStatusBarSegmentStyle() StatusBarSegmentStyle {
	sep := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return StatusBarSegmentStyle{Separator: sep, Border: border}
}

// StatusBarSegment displays configurable status bar items.
type StatusBarSegment struct {
	BaseComponent
	mu sync.Mutex

	segments []StatusSegment
	style    StatusBarSegmentStyle
}

// NewStatusBarSegment creates a StatusBarSegment.
func NewStatusBarSegment() *StatusBarSegment {
	sb := &StatusBarSegment{style: DefaultStatusBarSegmentStyle()}
	sb.SetID(GenerateID("statusseg"))
	return sb
}

// AddSegment adds a status bar item with label, value, and colors.
func (sb *StatusBarSegment) AddSegment(label, value string, fg, bg buffer.Color) *StatusBarSegment {
	sb.mu.Lock()
	sb.segments = append(sb.segments, StatusSegment{Label: label, Value: value, Fg: fg, Bg: bg})
	sb.mu.Unlock()
	return sb
}

// SegmentCount returns the number of segments.
func (sb *StatusBarSegment) SegmentCount() int {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	return len(sb.segments)
}

// Clear removes all segments.
func (sb *StatusBarSegment) Clear() *StatusBarSegment {
	sb.mu.Lock()
	sb.segments = sb.segments[:0]
	sb.mu.Unlock()
	return sb
}

// SetStyle sets the custom style.
func (sb *StatusBarSegment) SetStyle(s StatusBarSegmentStyle) *StatusBarSegment {
	sb.mu.Lock()
	sb.style = s
	sb.mu.Unlock()
	return sb
}

// Measure returns the preferred size.
func (sb *StatusBarSegment) Measure(cs Constraints) Size {
	sb.mu.Lock()
	count := len(sb.segments)
	sb.mu.Unlock()
	w := count * 15
	if w < 20 { w = 20 }
	h := 1
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: h}
}

// Paint renders the status bar segments into the buffer.
func (sb *StatusBarSegment) Paint(buf *buffer.Buffer) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	b := sb.Bounds()
	x, y := b.X, b.Y

	col := x
	for _, seg := range sb.segments {
		// Label with background
		for _, r := range seg.Label {
			if col >= buf.Width { return }
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: seg.Fg, Bg: seg.Bg, Width: 1})
			col++
		}
		// Space between label and value
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: seg.Fg, Bg: seg.Bg, Width: 1})
		col++
		// Value
		for _, r := range seg.Value {
			if col >= buf.Width { return }
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: seg.Fg, Bg: seg.Bg, Width: 1})
			col++
		}
		// Trailing space
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: seg.Fg, Bg: seg.Bg, Width: 1})
		col++
	}
}

// Children returns nil.
func (sb *StatusBarSegment) Children() []Component { return nil }

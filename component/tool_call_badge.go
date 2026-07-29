package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ToolCallBadge: AI Tool Call Status Badge ───
//
// ToolCallBadge renders a compact badge showing the count and status of
// AI tool/function calls. Displays success/failure counts with colored
// indicators suitable for placement in headers or status bars.
//
// Usage:
//
//	badge := NewToolCallBadge()
//	badge.SetCalls(12, 2) // 12 succeeded, 2 failed
//	badge.Paint(buf)

// ToolCallBadgeStyle holds styling.
type ToolCallBadgeStyle struct {
	Success buffer.Style
	Failure buffer.Style
	Label   buffer.Style
	Count   buffer.Style
	Bracket buffer.Style
}

// DefaultToolCallBadgeStyle returns defaults.
func DefaultToolCallBadgeStyle() ToolCallBadgeStyle {
	return ToolCallBadgeStyle{
		Success: buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Failure: buffer.Style{Fg: buffer.RGB(239, 68, 68)},
		Label:   buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Count:   buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
		Bracket: buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

// ToolCallBadge renders a tool call status badge.
type ToolCallBadge struct {
	BaseComponent
	mu sync.Mutex

	success int
	failure int
	style   ToolCallBadgeStyle
	// cached
	totalStr   string
	successStr string
	failureStr string
	labelText  string // "Tools" or "Tool" based on count
}

// NewToolCallBadge creates a ToolCallBadge.
func NewToolCallBadge() *ToolCallBadge {
	b := &ToolCallBadge{style: DefaultToolCallBadgeStyle()}
	b.SetID(GenerateID("toolbadge"))
	b.recomputeLocked()
	return b
}

// SetCalls sets success and failure counts.
func (b *ToolCallBadge) SetCalls(success, failure int) *ToolCallBadge {
	b.mu.Lock()
	if success < 0 { success = 0 }
	if failure < 0 { failure = 0 }
	b.success = success
	b.failure = failure
	b.recomputeLocked()
	b.mu.Unlock()
	return b
}

func (b *ToolCallBadge) recomputeLocked() {
	total := b.success + b.failure
	b.totalStr = itoa(total)
	b.successStr = itoa(b.success)
	b.failureStr = itoa(b.failure)
	if total == 1 {
		b.labelText = "Tool"
	} else {
		b.labelText = "Tools"
	}
}

// TotalCalls returns the total number of calls.
func (b *ToolCallBadge) TotalCalls() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.success + b.failure
}

// SuccessRate returns success percentage (0-100).
func (b *ToolCallBadge) SuccessRate() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	total := b.success + b.failure
	if total == 0 { return 0 }
	return b.success * 100 / total
}

// SetStyle sets custom style.
func (b *ToolCallBadge) SetStyle(s ToolCallBadgeStyle) *ToolCallBadge {
	b.mu.Lock()
	b.style = s
	b.mu.Unlock()
	return b
}

// Measure returns preferred size.
func (b *ToolCallBadge) Measure(cs Constraints) Size {
	w := 22 // approximate: [Tools 999/999]
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// Paint renders the tool call badge in a single row.
func (b *ToolCallBadge) Paint(buf *buffer.Buffer) {
	b.mu.Lock()
	defer b.mu.Unlock()

	bx := b.Bounds()
	x, y := bx.X, bx.Y

	bracketStyle := b.style.Bracket
	labelStyle := b.style.Label
	successStyle := b.style.Success
	failureStyle := b.style.Failure
	col := x

	// [ prefix
	buf.SetCell(col, y, buffer.Cell{Rune: '[', Fg: bracketStyle.Fg, Bg: bracketStyle.Bg, Flags: bracketStyle.Flags, Width: 1})
	col++

	// Label
	for _, r := range b.labelText {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	// space
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	// success count
	for _, r := range b.successStr {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: successStyle.Fg, Bg: successStyle.Bg, Flags: successStyle.Flags, Width: 1})
		col++
	}
	// separator /
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: '/', Fg: bracketStyle.Fg, Bg: bracketStyle.Bg, Flags: bracketStyle.Flags, Width: 1})
		col++
	}
	// failure count
	for _, r := range b.failureStr {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: failureStyle.Fg, Bg: failureStyle.Bg, Flags: failureStyle.Flags, Width: 1})
		col++
	}
	// ] suffix
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ']', Fg: bracketStyle.Fg, Bg: bracketStyle.Bg, Flags: bracketStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (b *ToolCallBadge) Children() []Component { return nil }

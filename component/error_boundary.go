package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ErrorBoundary: Error Display Boundary ───
//
// ErrorBoundary renders an error panel with an icon, message, and optional
// stack trace preview. Designed to gracefully display component errors
// without crashing the entire application.
//
// Usage:
//
//	eb := NewErrorBoundary()
//	eb.SetError("Failed to render component", "widget.go:42")
//	eb.Paint(buf)

// ErrorBoundaryStyle holds styling.
type ErrorBoundaryStyle struct {
	Icon    buffer.Style
	Title   buffer.Style
	Message buffer.Style
	Detail  buffer.Style
	Border  buffer.Style
}

// DefaultErrorBoundaryStyle returns defaults.
func DefaultErrorBoundaryStyle() ErrorBoundaryStyle {
	return ErrorBoundaryStyle{
		Icon:    buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Title:   buffer.Style{Fg: buffer.RGB(252, 165, 165), Flags: buffer.Bold},
		Message: buffer.Style{Fg: buffer.RGB(226, 232, 240)},
		Detail:  buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Border:  buffer.Style{Fg: buffer.RGB(127, 29, 29)},
	}
}

// ErrorBoundary renders an error display panel.
type ErrorBoundary struct {
	BaseComponent
	mu sync.Mutex

	message string
	detail  string
	width   int
	style   ErrorBoundaryStyle
	// cached
	msgRunes  []rune
	detRunes  []rune
}

// NewErrorBoundary creates an ErrorBoundary.
func NewErrorBoundary() *ErrorBoundary {
	eb := &ErrorBoundary{width: 40, style: DefaultErrorBoundaryStyle()}
	eb.SetID(GenerateID("errbound"))
	return eb
}

// SetError sets the error message and optional detail (e.g., file:line).
func (eb *ErrorBoundary) SetError(message, detail string) *ErrorBoundary {
	eb.mu.Lock()
	eb.message = message
	eb.detail = detail
	eb.msgRunes = []rune(message)
	eb.detRunes = []rune(detail)
	eb.mu.Unlock()
	return eb
}

// Clear clears the error.
func (eb *ErrorBoundary) Clear() *ErrorBoundary {
	eb.mu.Lock()
	eb.message = ""
	eb.detail = ""
	eb.msgRunes = nil
	eb.detRunes = nil
	eb.mu.Unlock()
	return eb
}

// HasError returns whether there is an active error.
func (eb *ErrorBoundary) HasError() bool {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	return eb.message != ""
}

// SetWidth sets the display width.
func (eb *ErrorBoundary) SetWidth(w int) *ErrorBoundary {
	eb.mu.Lock()
	if w < 10 { w = 10 }
	eb.width = w
	eb.mu.Unlock()
	return eb
}

// SetStyle sets custom style.
func (eb *ErrorBoundary) SetStyle(s ErrorBoundaryStyle) *ErrorBoundary {
	eb.mu.Lock()
	eb.style = s
	eb.mu.Unlock()
	return eb
}

// Measure returns preferred size.
func (eb *ErrorBoundary) Measure(cs Constraints) Size {
	w := eb.width
	h := 4
	if eb.detail != "" { h = 5 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: h}
}

// Paint renders the error boundary.
func (eb *ErrorBoundary) Paint(buf *buffer.Buffer) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	b := eb.Bounds()
	x, y := b.X, b.Y
	w := b.W
	if w < 10 { w = 40 }
	h := b.H
	if h < 3 { h = 4 }

	if eb.message == "" { return }

	borderStyle := eb.style.Border
	iconStyle := eb.style.Icon
	msgStyle := eb.style.Message
	detStyle := eb.style.Detail

	// Top border with title
	col := x
	buf.SetCell(col, y, buffer.Cell{Rune: '╭', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
	col++
	for range " ERROR " {
		if col >= x+w-1 { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '─', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
		col++
	}
	// Icon and title on border
	col = x + 1
	iconTitle := "✕ ERROR"
	for _, r := range iconTitle {
		if col >= x+w-1 { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: iconStyle.Fg, Bg: iconStyle.Bg, Flags: iconStyle.Flags, Width: 1})
		col++
	}
	// Right corner
	if x+w-1 < buf.Width {
		buf.SetCell(x+w-1, y, buffer.Cell{Rune: '╮', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
	}

	// Side borders + message on row 1
	for row := 1; row < h-1; row++ {
		yy := y + row
		if x < buf.Width {
			buf.SetCell(x, yy, buffer.Cell{Rune: '│', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
		}
		if x+w-1 < buf.Width {
			buf.SetCell(x+w-1, yy, buffer.Cell{Rune: '│', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
		}
	}

	// Message text on row 1
	col = x + 2
	for _, r := range eb.msgRunes {
		if col >= x+w-2 || col >= buf.Width { break }
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: msgStyle.Fg, Bg: msgStyle.Bg, Flags: msgStyle.Flags, Width: 1})
		col++
	}

	// Detail on row 2 (if present)
	if len(eb.detRunes) > 0 {
		col = x + 2
		for _, r := range eb.detRunes {
			if col >= x+w-2 || col >= buf.Width { break }
			buf.SetCell(col, y+2, buffer.Cell{Rune: r, Fg: detStyle.Fg, Bg: detStyle.Bg, Flags: detStyle.Flags, Width: 1})
			col++
	}
	}

	// Bottom border
	buf.SetCell(x, y+h-1, buffer.Cell{Rune: '╰', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
	for col := x + 1; col < x+w-1; col++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y+h-1, buffer.Cell{Rune: '─', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
	}
	if x+w-1 < buf.Width {
		buf.SetCell(x+w-1, y+h-1, buffer.Cell{Rune: '╯', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
	}
}

// Children returns nil.
func (eb *ErrorBoundary) Children() []Component { return nil }

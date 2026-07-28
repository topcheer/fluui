package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── StopReasonBadge: AI Stream Completion Reason ───
//
// StopReasonBadge renders a compact badge showing why an AI response stream
// ended: "stop" (natural end), "length" (max tokens), "tool_call" (function
// call), "content_filter" (filtered), or "error".
//
// Usage:
//
//	sr := NewStopReasonBadge(StopReasonStop)
//	sr.Paint(buf) // renders "✓ stop" in green

// StopReason represents why an AI stream ended.
type StopReason int

const (
	StopReasonNone           StopReason = iota
	StopReasonStop                      // natural completion
	StopReasonLength                    // hit max_tokens
	StopReasonToolCall                  // function/tool call
	StopReasonContentFilter             // content filtered
	StopReasonError                     // error occurred
)

// StopReasonStyle holds visual styles.
type StopReasonStyle struct {
	Stop      buffer.Style
	Length    buffer.Style
	ToolCall  buffer.Style
	Filter    buffer.Style
	Error     buffer.Style
}

// DefaultStopReasonStyle returns sensible defaults.
func DefaultStopReasonStyle() StopReasonStyle {
	return StopReasonStyle{
		Stop:     buffer.Style{Fg: buffer.RGB(16, 163, 127)},   // green
		Length:   buffer.Style{Fg: buffer.RGB(255, 175, 64)},   // orange
		ToolCall: buffer.Style{Fg: buffer.RGB(100, 149, 237)},  // blue
		Filter:   buffer.Style{Fg: buffer.RGB(220, 80, 80)},    // red
		Error:    buffer.Style{Fg: buffer.RGB(220, 80, 80), Flags: buffer.Bold}, // red bold
	}
}

// StopReasonBadge renders an AI stream stop reason.
type StopReasonBadge struct {
	BaseComponent
	mu     sync.RWMutex
	reason StopReason
	style  StopReasonStyle
}

// NewStopReasonBadge creates a badge with the given reason.
func NewStopReasonBadge(reason StopReason) *StopReasonBadge {
	sr := &StopReasonBadge{
		reason: reason,
		style:  DefaultStopReasonStyle(),
	}
	sr.SetID(GenerateID("stopreason"))
	return sr
}

// Reason returns the current stop reason.
func (sr *StopReasonBadge) Reason() StopReason {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.reason
}

// SetReason sets the stop reason.
func (sr *StopReasonBadge) SetReason(r StopReason) *StopReasonBadge {
	sr.mu.Lock()
	sr.reason = r
	sr.mu.Unlock()
	return sr
}

// ReasonText returns the human-readable label.
func (sr *StopReasonBadge) ReasonText() string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	switch sr.reason {
	case StopReasonStop:
		return "stop"
	case StopReasonLength:
		return "max tokens"
	case StopReasonToolCall:
		return "tool call"
	case StopReasonContentFilter:
		return "filtered"
	case StopReasonError:
		return "error"
	default:
		return "streaming"
	}
}

// ReasonIcon returns the icon character.
func (sr *StopReasonBadge) ReasonIcon() rune {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	switch sr.reason {
	case StopReasonStop:
		return '✓'
	case StopReasonLength:
		return '⋯'
	case StopReasonToolCall:
		return '⚡'
	case StopReasonContentFilter:
		return '⊘'
	case StopReasonError:
		return '✗'
	default:
		return '●'
	}
}

// SetStyle sets the visual style.
func (sr *StopReasonBadge) SetStyle(s StopReasonStyle) *StopReasonBadge {
	sr.mu.Lock()
	sr.style = s
	sr.mu.Unlock()
	return sr
}

// Measure computes the desired size.
func (sr *StopReasonBadge) Measure(cs Constraints) Size {
	return Size{W: 16, H: 1}
}

// Paint renders the badge.
func (sr *StopReasonBadge) Paint(buf *buffer.Buffer) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	b := sr.bounds
	if b.W < 3 || b.H < 1 {
		return
	}

	var style buffer.Style
	switch sr.reason {
	case StopReasonStop:
		style = sr.style.Stop
	case StopReasonLength:
		style = sr.style.Length
	case StopReasonToolCall:
		style = sr.style.ToolCall
	case StopReasonContentFilter:
		style = sr.style.Filter
	case StopReasonError:
		style = sr.style.Error
	default:
		style = buffer.Style{Fg: buffer.RGB(150, 150, 150)}
	}

	x := b.X
	// Icon
	buf.SetCell(x, b.Y, buffer.Cell{Rune: sr.ReasonIconLocked(), Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
	x++
	// Space
	if x < b.X+b.W {
		buf.SetCell(x, b.Y, buffer.Cell{Rune: ' ', Width: 1})
		x++
	}
	// Label
	label := sr.reasonTextLocked()
	for _, r := range label {
		if x >= b.X+b.W {
			break
		}
		buf.SetCell(x, b.Y, buffer.Cell{Rune: r, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
		x++
	}
}

// ReasonIconLocked returns icon (caller holds lock).
func (sr *StopReasonBadge) ReasonIconLocked() rune {
	switch sr.reason {
	case StopReasonStop:
		return '✓'
	case StopReasonLength:
		return '⋯'
	case StopReasonToolCall:
		return '⚡'
	case StopReasonContentFilter:
		return '⊘'
	case StopReasonError:
		return '✗'
	default:
		return '●'
	}
}

// reasonTextLocked returns label (caller holds lock).
func (sr *StopReasonBadge) reasonTextLocked() string {
	switch sr.reason {
	case StopReasonStop:
		return "stop"
	case StopReasonLength:
		return "max tokens"
	case StopReasonToolCall:
		return "tool call"
	case StopReasonContentFilter:
		return "filtered"
	case StopReasonError:
		return "error"
	default:
		return "streaming"
	}
}

// Children returns nil.
func (sr *StopReasonBadge) Children() []Component { return nil }

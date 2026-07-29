package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AccessibilityFocusRing: Focus Ring for Keyboard Navigation ───
//
// AccessibilityFocusRing renders a visual focus indicator (dashed/solid border)
// around a component area, essential for keyboard-accessible TUI applications.
// Supports different ring styles (solid, dashed, thick) and ARIA-like labels.
//
// Usage:
//
//	afr := NewAccessibilityFocusRing()
//	afr.SetLabel("Submit Button")
//	afr.SetFocused(true)
//	afr.SetRingStyle(AccRingDashed)
//	afr.Paint(buf)

// AccRingStyle_type describes the ring border style.
type AccRingStyleType int

const (
	AccRingSolid  AccRingStyleType = iota
	AccRingDashed
	AccRingThick
)

// AccessibilityFocusRingStyle holds styling.
type AccessibilityFocusRingStyle struct {
	Focused   buffer.Style
	Unfocused buffer.Style
	Label     buffer.Style
}

// DefaultAccessibilityFocusRingStyle returns defaults.
func DefaultAccessibilityFocusRingStyle() AccessibilityFocusRingStyle {
	focused := buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Bold}
	unfocused := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	label := buffer.Style{Fg: buffer.RGB(250, 204, 21), Flags: buffer.Bold}
	return AccessibilityFocusRingStyle{Focused: focused, Unfocused: unfocused, Label: label}
}

// AccessibilityFocusRing renders a focus ring indicator.
type AccessibilityFocusRing struct {
	BaseComponent
	mu sync.Mutex

	focused   bool
	label     string
	ringStyle AccRingStyleType
	style     AccessibilityFocusRingStyle
}

// NewAccessibilityFocusRing creates an AccessibilityFocusRing.
func NewAccessibilityFocusRing() *AccessibilityFocusRing {
	afr := &AccessibilityFocusRing{ringStyle: AccRingSolid, style: DefaultAccessibilityFocusRingStyle()}
	afr.SetID(GenerateID("accfocus"))
	return afr
}

// SetFocused sets whether the ring is focused.
func (afr *AccessibilityFocusRing) SetFocused(f bool) *AccessibilityFocusRing {
	afr.mu.Lock()
	afr.focused = f
	afr.mu.Unlock()
	return afr
}

// Focused returns whether the ring is focused.
func (afr *AccessibilityFocusRing) Focused() bool {
	afr.mu.Lock()
	defer afr.mu.Unlock()
	return afr.focused
}

// SetLabel sets the accessibility label text.
func (afr *AccessibilityFocusRing) SetLabel(l string) *AccessibilityFocusRing {
	afr.mu.Lock()
	afr.label = l
	afr.mu.Unlock()
	return afr
}

// Label returns the label.
func (afr *AccessibilityFocusRing) Label() string {
	afr.mu.Lock()
	defer afr.mu.Unlock()
	return afr.label
}

// SetRingStyle sets the ring border style type.
func (afr *AccessibilityFocusRing) SetRingStyle(s AccRingStyleType) *AccessibilityFocusRing {
	afr.mu.Lock()
	afr.ringStyle = s
	afr.mu.Unlock()
	return afr
}

// RingStyle returns the ring style type.
func (afr *AccessibilityFocusRing) RingStyle() AccRingStyleType {
	afr.mu.Lock()
	defer afr.mu.Unlock()
	return afr.ringStyle
}

// SetStyle sets custom style.
func (afr *AccessibilityFocusRing) SetStyle(s AccessibilityFocusRingStyle) *AccessibilityFocusRing {
	afr.mu.Lock()
	afr.style = s
	afr.mu.Unlock()
	return afr
}

// Measure returns preferred size.
func (afr *AccessibilityFocusRing) Measure(cs Constraints) Size {
	w := 20
	h := 3
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the focus ring into the buffer.
func (afr *AccessibilityFocusRing) Paint(buf *buffer.Buffer) {
	afr.mu.Lock()
	defer afr.mu.Unlock()

	b := afr.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 5 { w = 20 }
	if h < 3 { h = 3 }

	// Determine style
	var style buffer.Style
	if afr.focused {
		style = afr.style.Focused
	} else {
		style = afr.style.Unfocused
	}

	// Determine border chars based on ring style
	var hChar, vChar, tlChar, trChar, blChar, brChar rune
	switch afr.ringStyle {
	case AccRingDashed:
		hChar, vChar = '┄', '┊'
		tlChar, trChar, blChar, brChar = '┌', '┐', '└', '┘'
	case AccRingThick:
		hChar, vChar = '━', '┃'
		tlChar, trChar, blChar, brChar = '┏', '┓', '┗', '┛'
	default:
		hChar, vChar = '─', '│'
		tlChar, trChar, blChar, brChar = '┌', '┐', '└', '┘'
	}

	// Draw border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = tlChar } else if row == 0 && col == w-1 { ch = trChar } else if row == h-1 && col == 0 { ch = blChar } else if row == h-1 && col == w-1 { ch = brChar } else if row == 0 || row == h-1 { ch = hChar } else if col == 0 || col == w-1 { ch = vChar }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
			}
		}
	}

	// Label (if focused and label non-empty)
	if afr.focused && afr.label != "" {
		labelStyle := afr.style.Label
		// Draw label on top border, overwriting some border chars
		labelLen := len(afr.label)
		labelStart := x + 2
		// Leading space
		if labelStart-1 >= 0 && labelStart-1 < buf.Width && y < buf.Height {
			buf.SetCell(labelStart-1, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		}
		for i, r := range afr.label {
			cx := labelStart + i
			if cx >= x+w-1 || cx >= buf.Width { break }
			buf.SetCell(cx, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		}
		// Trailing space
		trailX := labelStart + labelLen
		if trailX < x+w-1 && trailX < buf.Width {
			buf.SetCell(trailX, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		}
		_ = labelLen
	}
}

// Children returns nil.
func (afr *AccessibilityFocusRing) Children() []Component { return nil }

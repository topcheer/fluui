package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// KBDVariant controls the visual style of a KBD keycap.
type KBDVariant int

const (
	// KBDInverse renders the key in inverse video (swap fg/bg).
	KBDInverse KBDVariant = iota
	// KBDBracket renders the key in square brackets: [Ctrl].
	KBDBracket
	// KBDBordered renders the key with box-drawing border chars.
	KBDBordered
)

// kbdBorderedChars are the border characters for KBDBordered style.
// Using light box-drawing characters: ┌─┐│└┘
const (
	kbdTL = '\u250c' // ┌
	kbdTR = '\u2510' // ┐
	kbdBL = '\u2514' // └
	kbdBR = '\u2518' // ┘
	kbdH  = '\u2500' // ─
	kbdV  = '\u2502' // │
)

// KBD renders a keyboard keycap, useful for displaying keyboard shortcuts
// in help screens, command hints, and documentation. Supports three visual
// styles: inverse video, bracketed, and bordered.
//
// Thread-safe.
type KBD struct {
	BaseComponent
	mu      sync.RWMutex
	text    string
	variant KBDVariant
}

// NewKBD creates a keycap with the given text (e.g., "Ctrl+C", "Enter").
func NewKBD(text string) *KBD {
	return &KBD{
		BaseComponent: BaseComponent{id: GenerateID("kbd")},
		text:          text,
		variant:       KBDInverse,
	}
}

// Text returns the keycap display text.
func (k *KBD) Text() string {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.text
}

// SetText updates the keycap display text.
func (k *KBD) SetText(s string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.text = s
}

// Variant returns the current visual style.
func (k *KBD) Variant() KBDVariant {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.variant
}

// SetVariant updates the visual style.
func (k *KBD) SetVariant(v KBDVariant) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.variant = v
}

// Measure returns the preferred size.
// Inverse/Bracket: single row, content+2 wide.
// Bordered: 3 rows tall, content+2 wide.
func (k *KBD) Measure(cs Constraints) Size {
	k.mu.RLock()
	defer k.mu.RUnlock()

	w := buffer.StringWidth(k.text) + 2 // padding/border
	h := 1
	if k.variant == KBDBordered {
		h = 3
	}

	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.HasHeight() && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return Size{W: w, H: h}
}

// Paint draws the keycap. Zero allocations.
func (k *KBD) Paint(buf *buffer.Buffer) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	bounds := k.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	t := theme.Get()

	switch k.variant {
	case KBDBracket:
		k.paintBracketLocked(buf, bounds, t)
	case KBDBordered:
		k.paintBorderedLocked(buf, bounds, t)
	default:
		k.paintInverseLocked(buf, bounds, t)
	}
}

// paintInverseLocked draws key text in inverse video with 1-cell padding.
func (k *KBD) paintInverseLocked(buf *buffer.Buffer, bounds Rect, t *theme.Theme) {
	style := buffer.Style{
		Fg:    t.Bg,
		Bg:    t.Fg,
		Flags: buffer.Bold,
	}
	padStyle := buffer.Style{Fg: t.Bg, Bg: t.Fg}

	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	// Left padding
	if x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: padStyle.Fg, Bg: padStyle.Bg})
		x++
	}
	// Text
	if x < maxX {
		x = buf.DrawText(x, y, k.text, style)
	}
	// Right padding
	if x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: padStyle.Fg, Bg: padStyle.Bg})
	}
}

// paintBracketLocked draws key text in square brackets.
func (k *KBD) paintBracketLocked(buf *buffer.Buffer, bounds Rect, t *theme.Theme) {
	style := buffer.Style{Fg: t.Accent, Flags: buffer.Bold}

	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	// [
	if x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: '[', Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
		x++
	}
	// Text
	if x < maxX {
		x = buf.DrawText(x, y, k.text, style)
	}
	// ]
	if x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: ']', Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
	}
}

// paintBorderedLocked draws key text inside a box-drawing border (3 rows).
func (k *KBD) paintBorderedLocked(buf *buffer.Buffer, bounds Rect, t *theme.Theme) {
	borderStyle := buffer.Style{Fg: t.Muted}
	textStyle := buffer.Style{Fg: t.Accent, Flags: buffer.Bold}

	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W
	maxY := bounds.Y + bounds.H

	// Content width inside border
	innerW := bounds.W - 2 // minus left/right border
	if innerW < 0 {
		innerW = 0
	}

	// Top border: ┌───┐
	if y < maxY {
		buf.SetCell(x, y, buffer.Cell{Rune: kbdTL, Width: 1, Fg: borderStyle.Fg})
		for i := 1; i < bounds.W-1 && x+i < maxX; i++ {
			buf.SetCell(x+i, y, buffer.Cell{Rune: kbdH, Width: 1, Fg: borderStyle.Fg})
		}
		if bounds.W > 1 && x+bounds.W-1 < maxX {
			buf.SetCell(x+bounds.W-1, y, buffer.Cell{Rune: kbdTR, Width: 1, Fg: borderStyle.Fg})
		}
	}

	// Middle row: │text│
	if y+1 < maxY {
		midY := y + 1
		buf.SetCell(x, midY, buffer.Cell{Rune: kbdV, Width: 1, Fg: borderStyle.Fg})
		// Draw text starting at x+1
		tx := x + 1
		if tx < maxX {
			tx = buf.DrawText(tx, midY, k.text, textStyle)
		}
		// Pad remaining inner space
		for tx < x+bounds.W-1 && tx < maxX {
			buf.SetCell(tx, midY, buffer.Cell{Rune: ' ', Width: 1, Fg: borderStyle.Fg})
			tx++
		}
		// Right border
		if x+bounds.W-1 < maxX {
			buf.SetCell(x+bounds.W-1, midY, buffer.Cell{Rune: kbdV, Width: 1, Fg: borderStyle.Fg})
		}
	}

	// Bottom border: └───┘
	if y+2 < maxY {
		botY := y + 2
		buf.SetCell(x, botY, buffer.Cell{Rune: kbdBL, Width: 1, Fg: borderStyle.Fg})
		for i := 1; i < bounds.W-1 && x+i < maxX; i++ {
			buf.SetCell(x+i, botY, buffer.Cell{Rune: kbdH, Width: 1, Fg: borderStyle.Fg})
		}
		if bounds.W > 1 && x+bounds.W-1 < maxX {
			buf.SetCell(x+bounds.W-1, botY, buffer.Cell{Rune: kbdBR, Width: 1, Fg: borderStyle.Fg})
		}
	}
}

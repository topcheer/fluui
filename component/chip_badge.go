package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ChipBadge: Removable Tag Chips ───
//
// ChipBadge renders a horizontal row of removable tag chips, each showing
// a label with a close (x) button. Common in filter bars and tag inputs.
//
// Usage:
//
//	cb := NewChipBadge()
//	cb.AddChip("go")
//	cb.AddChip("tui")
//	cb.AddChip("ai")
//	cb.RemoveChip(0)
//	cb.Paint(buf)

// ChipEntry represents a single chip.
type ChipEntry struct {
	Label    string
	Selected bool
}

// ChipBadgeStyle holds styling.
type ChipBadgeStyle struct {
	Normal   buffer.Style
	Selected buffer.Style
	Remove   buffer.Style // the x button
	Border   buffer.Style
}

// DefaultChipBadgeStyle returns defaults.
func DefaultChipBadgeStyle() ChipBadgeStyle {
	normal := buffer.Style{Fg: buffer.RGB(148, 163, 184), Bg: buffer.RGB(30, 41, 59)}
	sel := buffer.Style{Fg: buffer.RGB(96, 165, 250), Bg: buffer.RGB(30, 58, 138), Flags: buffer.Bold}
	remove := buffer.Style{Fg: buffer.RGB(239, 68, 68)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return ChipBadgeStyle{Normal: normal, Selected: sel, Remove: remove, Border: border}
}

// ChipBadge displays a row of removable tag chips.
type ChipBadge struct {
	BaseComponent
	mu sync.Mutex

	chips []ChipEntry
	style ChipBadgeStyle
}

// NewChipBadge creates a ChipBadge.
func NewChipBadge() *ChipBadge {
	cb := &ChipBadge{style: DefaultChipBadgeStyle()}
	cb.SetID(GenerateID("chipbadge"))
	return cb
}

// AddChip adds a tag chip.
func (cb *ChipBadge) AddChip(label string) *ChipBadge {
	cb.mu.Lock()
	cb.chips = append(cb.chips, ChipEntry{Label: label})
	cb.mu.Unlock()
	return cb
}

// RemoveChip removes a chip by index.
func (cb *ChipBadge) RemoveChip(index int) *ChipBadge {
	cb.mu.Lock()
	if index >= 0 && index < len(cb.chips) {
		cb.chips = append(cb.chips[:index], cb.chips[index+1:]...)
	}
	cb.mu.Unlock()
	return cb
}

// ToggleSelected toggles the selected state of a chip.
func (cb *ChipBadge) ToggleSelected(index int) *ChipBadge {
	cb.mu.Lock()
	if index >= 0 && index < len(cb.chips) {
		cb.chips[index].Selected = !cb.chips[index].Selected
	}
	cb.mu.Unlock()
	return cb
}

// ChipCount returns the number of chips.
func (cb *ChipBadge) ChipCount() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return len(cb.chips)
}

// Clear removes all chips.
func (cb *ChipBadge) Clear() *ChipBadge {
	cb.mu.Lock()
	cb.chips = cb.chips[:0]
	cb.mu.Unlock()
	return cb
}

// SetStyle sets custom style.
func (cb *ChipBadge) SetStyle(s ChipBadgeStyle) *ChipBadge {
	cb.mu.Lock()
	cb.style = s
	cb.mu.Unlock()
	return cb
}

// Measure returns the preferred size.
func (cb *ChipBadge) Measure(cs Constraints) Size {
	cb.mu.Lock()
	count := len(cb.chips)
	cb.mu.Unlock()
	w := count * 10
	if w < 20 { w = 20 }
	h := 1
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: h}
}

// Paint renders the chip badges into the buffer.
func (cb *ChipBadge) Paint(buf *buffer.Buffer) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	b := cb.Bounds()
	x, y := b.X, b.Y

	col := x
	for _, chip := range cb.chips {
		var style buffer.Style
		if chip.Selected {
			style = cb.style.Selected
		} else {
			style = cb.style.Normal
		}

		// Opening bracket [
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: '[', Fg: style.Fg, Bg: style.Bg, Width: 1})
		col++

		// Label
		for _, r := range chip.Label {
			if col >= buf.Width { return }
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
			col++
		}

		// Space
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: style.Fg, Bg: style.Bg, Width: 1})
		col++

		// Remove x
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: 'x', Fg: cb.style.Remove.Fg, Bg: style.Bg, Width: 1})
		col++

		// Closing bracket ]
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: ']', Fg: style.Fg, Bg: style.Bg, Width: 1})
		col++

		// Separator space
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: style.Fg, Bg: style.Bg, Width: 1})
		col++
	}
}

// Children returns nil.
func (cb *ChipBadge) Children() []Component { return nil }

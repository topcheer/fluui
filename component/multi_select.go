package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MultiSelect: Multi-Selection List with Checkboxes ───
//
// MultiSelect renders a scrollable list of options with checkboxes (☐/☑),
// supporting toggle, select all, and deselect all operations.
//
// Usage:
//
//	ms := NewMultiSelect()
//	ms.AddOption("Apple")
//	ms.AddOption("Banana")
//	ms.AddOption("Cherry")
//	ms.Toggle(0)
//	ms.Paint(buf)

// MultiSelectStyle holds styling.
type MultiSelectStyle struct {
	Normal    buffer.Style
	Selected  buffer.Style
	Highlight buffer.Style // current cursor row
	Border    buffer.Style
}

// DefaultMultiSelectStyle returns defaults.
func DefaultMultiSelectStyle() MultiSelectStyle {
	normal := buffer.Style{Fg: buffer.RGB(226, 232, 240)}
	sel := buffer.Style{Fg: buffer.RGB(34, 197, 94)}
	hl := buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Bold}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return MultiSelectStyle{Normal: normal, Selected: sel, Highlight: hl, Border: border}
}

// MultiSelectOption represents a single selectable option.
type MultiSelectOption struct {
	Label    string
	Selected bool
}

// MultiSelect renders a multi-selection list with checkboxes.
type MultiSelect struct {
	BaseComponent
	mu sync.Mutex

	options []MultiSelectOption
	cursor  int
	style   MultiSelectStyle
}

// NewMultiSelect creates a MultiSelect.
func NewMultiSelect() *MultiSelect {
	ms := &MultiSelect{style: DefaultMultiSelectStyle()}
	ms.SetID(GenerateID("multiselect"))
	return ms
}

// AddOption adds a selectable option.
func (ms *MultiSelect) AddOption(label string) *MultiSelect {
	ms.mu.Lock()
	ms.options = append(ms.options, MultiSelectOption{Label: label})
	ms.mu.Unlock()
	return ms
}

// Toggle toggles the selection state of an option by index.
func (ms *MultiSelect) Toggle(index int) *MultiSelect {
	ms.mu.Lock()
	if index >= 0 && index < len(ms.options) {
		ms.options[index].Selected = !ms.options[index].Selected
	}
	ms.mu.Unlock()
	return ms
}

// SelectAll selects all options.
func (ms *MultiSelect) SelectAll() *MultiSelect {
	ms.mu.Lock()
	for i := range ms.options {
		ms.options[i].Selected = true
	}
	ms.mu.Unlock()
	return ms
}

// DeselectAll deselects all options.
func (ms *MultiSelect) DeselectAll() *MultiSelect {
	ms.mu.Lock()
	for i := range ms.options {
		ms.options[i].Selected = false
	}
	ms.mu.Unlock()
	return ms
}

// SelectedIndices returns indices of all selected options.
func (ms *MultiSelect) SelectedIndices() []int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	var result []int
	for i, opt := range ms.options {
		if opt.Selected {
			result = append(result, i)
		}
	}
	return result
}

// SelectedCount returns the number of selected options.
func (ms *MultiSelect) SelectedCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	count := 0
	for _, opt := range ms.options {
		if opt.Selected { count++ }
	}
	return count
}

// OptionCount returns the total number of options.
func (ms *MultiSelect) OptionCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return len(ms.options)
}

// SetCursor sets the cursor position (clamped).
func (ms *MultiSelect) SetCursor(idx int) *MultiSelect {
	ms.mu.Lock()
	if idx < 0 { idx = 0 }
	if idx >= len(ms.options) { idx = len(ms.options) - 1 }
	ms.cursor = idx
	ms.mu.Unlock()
	return ms
}

// Cursor returns the current cursor index.
func (ms *MultiSelect) Cursor() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.cursor
}

// MoveCursorUp moves cursor up by one.
func (ms *MultiSelect) MoveCursorUp() *MultiSelect {
	ms.mu.Lock()
	if ms.cursor > 0 { ms.cursor-- }
	ms.mu.Unlock()
	return ms
}

// MoveCursorDown moves cursor down by one.
func (ms *MultiSelect) MoveCursorDown() *MultiSelect {
	ms.mu.Lock()
	if ms.cursor < len(ms.options)-1 { ms.cursor++ }
	ms.mu.Unlock()
	return ms
}

// SetStyle sets custom style.
func (ms *MultiSelect) SetStyle(s MultiSelectStyle) *MultiSelect {
	ms.mu.Lock()
	ms.style = s
	ms.mu.Unlock()
	return ms
}

// Measure returns the preferred size.
func (ms *MultiSelect) Measure(cs Constraints) Size {
	ms.mu.Lock()
	count := len(ms.options)
	ms.mu.Unlock()
	w := 30
	h := count + 2
	if h < 3 { h = 3 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the multi-select list into the buffer.
func (ms *MultiSelect) Paint(buf *buffer.Buffer) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	b := ms.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 { w = 30 }
	if h < 3 { h = 3 }

	bs := ms.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	for idx, opt := range ms.options {
		rowY := y + 1 + idx
		if rowY >= y+h-1 || rowY >= buf.Height { break }

		isCursor := idx == ms.cursor
		col := x + 1

		// Cursor indicator
		if isCursor {
			if col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: '>', Fg: ms.style.Highlight.Fg, Bg: ms.style.Highlight.Bg, Flags: ms.style.Highlight.Flags, Width: 1})
			}
		} else {
			if col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: ms.style.Normal.Fg, Bg: ms.style.Normal.Bg, Width: 1})
			}
		}
		col++

		// Checkbox
		var cbChar rune
		var cbStyle buffer.Style
		if opt.Selected {
			cbChar = '☑'
			cbStyle = ms.style.Selected
		} else {
			cbChar = '☐'
			cbStyle = ms.style.Normal
		}
		if isCursor { cbStyle = ms.style.Highlight }
		if col < buf.Width {
			buf.SetCell(col, rowY, buffer.Cell{Rune: cbChar, Fg: cbStyle.Fg, Bg: cbStyle.Bg, Flags: cbStyle.Flags, Width: 1})
		}
		col++

		// Space
		if col < buf.Width {
			buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: cbStyle.Fg, Bg: cbStyle.Bg, Flags: cbStyle.Flags, Width: 1})
		}
		col++

		// Label
		var labelStyle buffer.Style
		if isCursor {
			labelStyle = ms.style.Highlight
		} else {
			labelStyle = ms.style.Normal
		}
		for _, r := range opt.Label {
			if col >= x+w-1 || col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (ms *MultiSelect) Children() []Component { return nil }

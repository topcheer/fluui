package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ShortcutList: Contextual Keyboard Shortcut Help ───
//
// ShortcutList renders a list of keyboard shortcuts with descriptions.
// Each entry shows a key binding and its action description.
// Useful for discoverability in TUI applications.
//
// Usage:
//
//	h := NewShortcutList()
//	h.AddBinding("Q", "Quit application")
//	h.AddBinding("Ctrl+S", "Save current file")
//	h.Paint(buf)

// ShortcutListStyle holds styling.
type ShortcutListStyle struct {
	Key     buffer.Style
	Sep     buffer.Style
	Desc    buffer.Style
	Title   buffer.Style
	Bracket buffer.Style
}

// DefaultShortcutListStyle returns defaults.
func DefaultShortcutListStyle() ShortcutListStyle {
	return ShortcutListStyle{
		Key:     buffer.Style{Fg: buffer.RGB(147, 197, 253), Flags: buffer.Bold},
		Sep:     buffer.Style{Fg: buffer.RGB(71, 85, 105)},
		Desc:    buffer.Style{Fg: buffer.RGB(203, 213, 225)},
		Title:   buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
		Bracket: buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

const helpMaxBindings = 20

// helpEntry stores a single key binding.
type helpEntry struct {
	key  string
	desc string
}

// ShortcutList renders a keyboard shortcut help panel.
type ShortcutList struct {
	BaseComponent
	mu sync.Mutex

	entries [helpMaxBindings]helpEntry
	count   int
	width   int
	style   ShortcutListStyle
}

// NewShortcutList creates a ShortcutList.
func NewShortcutList() *ShortcutList {
	h := &ShortcutList{width: 36, style: DefaultShortcutListStyle()}
	h.SetID(GenerateID("help"))
	return h
}

// AddBinding adds a keyboard shortcut entry.
func (h *ShortcutList) AddBinding(key, desc string) *ShortcutList {
	h.mu.Lock()
	if h.count < helpMaxBindings {
		h.entries[h.count] = helpEntry{key: key, desc: desc}
		h.count++
	}
	h.mu.Unlock()
	return h
}

// Clear removes all bindings.
func (h *ShortcutList) Clear() *ShortcutList {
	h.mu.Lock()
	h.count = 0
	h.mu.Unlock()
	return h
}

// Count returns the number of bindings.
func (h *ShortcutList) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.count
}

// SetWidth sets the display width.
func (h *ShortcutList) SetWidth(w int) *ShortcutList {
	h.mu.Lock()
	if w < 15 { w = 15 }
	h.width = w
	h.mu.Unlock()
	return h
}

// SetStyle sets custom style.
func (h *ShortcutList) SetStyle(s ShortcutListStyle) *ShortcutList {
	h.mu.Lock()
	h.style = s
	h.mu.Unlock()
	return h
}

// Measure returns preferred size.
func (h *ShortcutList) Measure(cs Constraints) Size {
	w := h.width
	height := h.count + 1 // +1 for title
	if height < 2 { height = 2 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && height > cs.MaxHeight { height = cs.MaxHeight }
	return Size{W: w, H: height}
}

// Paint renders the help overlay.
func (h *ShortcutList) Paint(buf *buffer.Buffer) {
	h.mu.Lock()
	defer h.mu.Unlock()

	b := h.Bounds()
	x, y := b.X, b.Y
	w := b.W
	if w < 15 { w = 36 }

	keyStyle := h.style.Key
	sepStyle := h.style.Sep
	descStyle := h.style.Desc
	titleStyle := h.style.Title

	// Title row
	title := " Shortcuts"
	col := x
	for _, r := range title {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: titleStyle.Fg, Bg: titleStyle.Bg, Flags: titleStyle.Flags, Width: 1})
		col++
	}

	// Bindings
	keyColWidth := 12 // fixed width for key column
	if keyColWidth > w-4 { keyColWidth = w - 4 }

	for i := 0; i < h.count; i++ {
		yy := y + 1 + i
		if yy >= buf.Height { break }

		entry := h.entries[i]

		// Key in brackets
		col = x
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: '[', Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
			col++
		}
		for _, r := range entry.key {
			if col >= buf.Width { break }
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: keyStyle.Fg, Bg: keyStyle.Bg, Flags: keyStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: ']', Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
			col++
		}

		// Pad to description column
		descStart := x + keyColWidth + 1
		for c := col; c < descStart && c < buf.Width; c++ {
			buf.SetCell(c, yy, buffer.Cell{Rune: ' ', Fg: descStyle.Fg, Bg: descStyle.Bg, Flags: descStyle.Flags, Width: 1})
		}

		// Description
		col = descStart
		for _, r := range entry.desc {
			if col >= buf.Width { break }
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: descStyle.Fg, Bg: descStyle.Bg, Flags: descStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (h *ShortcutList) Children() []Component { return nil }

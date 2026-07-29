package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── KeyHintBar: Keyboard Shortcut Badge Display ───
//
// KeyHintBar renders keyboard shortcuts as styled badges like [Q] Quit [S] Save.
// Useful in status bars, help panels, and command palettes.
//
// Usage:
//
//	kb := NewKeyHintBar()
//	kb.AddHint("Q", "Quit")
//	kb.AddHint("S", "Save")
//	kb.AddHint("/", "Search")
//	kb.Paint(buf)

// KeyHint represents a single keyboard shortcut entry.
type KeyHint struct {
	Key         string
	Description string
}

// KeyHintBarStyle holds styling for KeyHintBar.
type KeyHintBarStyle struct {
	Key      buffer.Style
	KeyBg    buffer.Style // background bracket style
	Desc     buffer.Style
	Separator buffer.Style
	Border   buffer.Style
}

// DefaultKeyHintBarStyle returns defaults.
func DefaultKeyHintBarStyle() KeyHintBarStyle {
	key := buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold}
	keyBg := buffer.Style{Fg: buffer.RGB(71, 85, 105), Bg: buffer.RGB(30, 41, 59)}
	desc := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	sep := buffer.Style{Fg: buffer.RGB(51, 65, 85)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return KeyHintBarStyle{Key: key, KeyBg: keyBg, Desc: desc, Separator: sep, Border: border}
}

// KeyHintBar displays keyboard shortcut badges.
type KeyHintBar struct {
	BaseComponent
	mu sync.Mutex

	hints  []KeyHint
	style  KeyHintBarStyle
}

// NewKeyHintBar creates a KeyHintBar with defaults.
func NewKeyHintBar() *KeyHintBar {
	kb := &KeyHintBar{style: DefaultKeyHintBarStyle()}
	kb.SetID(GenerateID("keyhint"))
	return kb
}

// AddHint adds a keyboard shortcut.
func (kb *KeyHintBar) AddHint(key, description string) *KeyHintBar {
	kb.mu.Lock()
	kb.hints = append(kb.hints, KeyHint{Key: key, Description: description})
	kb.mu.Unlock()
	return kb
}

// SetHints replaces all hints.
func (kb *KeyHintBar) SetHints(hints []KeyHint) *KeyHintBar {
	kb.mu.Lock()
	kb.hints = hints
	kb.mu.Unlock()
	return kb
}

// HintCount returns the number of hints.
func (kb *KeyHintBar) HintCount() int {
	kb.mu.Lock()
	defer kb.mu.Unlock()
	return len(kb.hints)
}

// Clear removes all hints.
func (kb *KeyHintBar) Clear() *KeyHintBar {
	kb.mu.Lock()
	kb.hints = kb.hints[:0]
	kb.mu.Unlock()
	return kb
}

// SetStyle sets the custom style.
func (kb *KeyHintBar) SetStyle(s KeyHintBarStyle) *KeyHintBar {
	kb.mu.Lock()
	kb.style = s
	kb.mu.Unlock()
	return kb
}

// Measure returns the preferred size.
func (kb *KeyHintBar) Measure(cs Constraints) Size {
	kb.mu.Lock()
	count := len(kb.hints)
	kb.mu.Unlock()
	w := count * 12 // approximate width per hint
	if w < 20 { w = 20 }
	h := 1
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: h}
}

// Paint renders the key hint bar into the buffer.
func (kb *KeyHintBar) Paint(buf *buffer.Buffer) {
	kb.mu.Lock()
	defer kb.mu.Unlock()

	b := kb.Bounds()
	x, y := b.X, b.Y

	keyStyle := kb.style.Key
	keyBgStyle := kb.style.KeyBg
	descStyle := kb.style.Desc
	sepStyle := kb.style.Separator

	col := x
	for idx, hint := range kb.hints {
		// Separator before each hint (except first)
		if idx > 0 {
			for _, r := range [2]rune{' ', '·'} {
				if col >= buf.Width { return }
				buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
				col++
			}
			if col >= buf.Width { return }
			buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
			col++
		}

		// Key bracket [
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: '[', Fg: keyBgStyle.Fg, Bg: keyBgStyle.Bg, Flags: keyBgStyle.Flags, Width: 1})
		col++
		// Key text
		for _, r := range hint.Key {
			if col >= buf.Width { return }
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: keyStyle.Fg, Bg: keyBgStyle.Bg, Flags: keyStyle.Flags, Width: 1})
			col++
		}
		// Key bracket ]
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: ']', Fg: keyBgStyle.Fg, Bg: keyBgStyle.Bg, Flags: keyBgStyle.Flags, Width: 1})
		col++
		// Space
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: descStyle.Fg, Bg: descStyle.Bg, Flags: descStyle.Flags, Width: 1})
		col++
		// Description
		for _, r := range hint.Description {
			if col >= buf.Width { return }
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: descStyle.Fg, Bg: descStyle.Bg, Flags: descStyle.Flags, Width: 1})
			col++
		}
		// Space after
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: descStyle.Fg, Bg: descStyle.Bg, Flags: descStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (kb *KeyHintBar) Children() []Component { return nil }

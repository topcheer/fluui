package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// SelectOption represents a single option in a Select dropdown.
type SelectOption struct {
	Label string
	Value string
}

// Select is a dropdown selection component. It displays the currently
// selected value and expands into a popup list when activated.
// Inspired by Textual's Select widget.
type Select struct {
	BaseComponent
	mu          sync.RWMutex
	options     []SelectOption
	selected    int    // index of selected option, -1 = none
	open        bool   // popup visible
	cursor      int    // highlighted option in popup
	width       int
	popupHeight int    // max visible options in popup (0 = auto)
	style       SelectStyle
	onChange    func(value string, index int)
}

// SelectStyle defines the visual style of a Select component.
type SelectStyle struct {
	Normal  buffer.Style // collapsed appearance
	Active  buffer.Style // when focused/active
	Popup   buffer.Style // popup background
	Selected buffer.Style // selected option in popup
	Cursor   buffer.Style // cursor-highlighted option
}

// DefaultSelectStyle returns a Dracula-themed style for Select.
func DefaultSelectStyle() SelectStyle {
	fg := buffer.RGB(248, 248, 242)
	bg := buffer.RGB(40, 42, 54)
	return SelectStyle{
		Normal:   buffer.Style{Fg: fg, Bg: bg},
		Active:   buffer.Style{Fg: fg, Bg: buffer.RGB(68, 71, 90)},
		Popup:    buffer.Style{Fg: fg, Bg: buffer.RGB(30, 31, 41)},
		Selected: buffer.Style{Fg: buffer.RGB(80, 250, 123), Bg: bg},
		Cursor:   buffer.Style{Fg: buffer.RGB(248, 248, 242), Bg: buffer.RGB(68, 71, 90)},
	}
}

// NewSelect creates a Select dropdown with the given options.
func NewSelect(options []SelectOption) *Select {
	return &Select{
		options:    options,
		selected:   -1,
		cursor:     0,
		popupHeight: 8,
		style:      DefaultSelectStyle(),
	}
}

// Value returns the value of the selected option, or "" if none.
func (s *Select) Value() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.selected < 0 || s.selected >= len(s.options) {
		return ""
	}
	return s.options[s.selected].Value
}

// Label returns the label of the selected option, or "" if none.
func (s *Select) Label() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.selected < 0 || s.selected >= len(s.options) {
		return ""
	}
	return s.options[s.selected].Label
}

// SelectedIndex returns the index of the selected option, or -1.
func (s *Select) SelectedIndex() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.selected
}

// SetSelectedIndex sets the selected option by index.
func (s *Select) SetSelectedIndex(idx int) {
	s.mu.Lock()
	if idx >= 0 && idx < len(s.options) {
		s.selected = idx
	} else {
		s.selected = -1
	}
	onChange := s.onChange
	var val string
	if s.selected >= 0 && s.selected < len(s.options) {
		val = s.options[s.selected].Value
	}
	s.mu.Unlock()
	if onChange != nil {
		onChange(val, idx)
	}
}

// SetOptions replaces all options. Always resets selection.
func (s *Select) SetOptions(opts []SelectOption) {
	s.mu.Lock()
	s.options = opts
	s.selected = -1 // always reset when options change
	if s.cursor >= len(opts) {
		s.cursor = 0
	}
	s.mu.Unlock()
}

// Options returns a copy of all options.
func (s *Select) Options() []SelectOption {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SelectOption, len(s.options))
	copy(out, s.options)
	return out
}

// IsOpen returns true if the popup is visible.
func (s *Select) IsOpen() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.open
}

// Open shows the popup.
func (s *Select) Open() {
	s.mu.Lock()
	s.open = true
	if s.selected >= 0 {
		s.cursor = s.selected
	} else {
		s.cursor = 0
	}
	s.mu.Unlock()
}

// Close hides the popup.
func (s *Select) Close() {
	s.mu.Lock()
	s.open = false
	s.mu.Unlock()
}

// Toggle opens/closes the popup.
func (s *Select) Toggle() {
	s.mu.Lock()
	s.open = !s.open
	if s.open && s.selected >= 0 {
		s.cursor = s.selected
	}
	s.mu.Unlock()
}

// SetOnChange sets a callback fired when selection changes.
func (s *Select) SetOnChange(fn func(value string, index int)) {
	s.mu.Lock()
	s.onChange = fn
	s.mu.Unlock()
}

// SetStyle customizes the visual appearance.
func (s *Select) SetStyle(style SelectStyle) {
	s.mu.Lock()
	s.style = style
	s.mu.Unlock()
}

// SetWidth sets the display width.
func (s *Select) SetWidth(w int) {
	s.mu.Lock()
	s.width = w
	s.mu.Unlock()
}

// SetPopupHeight sets the max visible options in the popup.
func (s *Select) SetPopupHeight(h int) {
	s.mu.Lock()
	s.popupHeight = h
	s.mu.Unlock()
}

// HandleKey processes keyboard input.
func (s *Select) HandleKey(k *term.KeyEvent) bool {
	s.mu.Lock()
	var onChange func(string, int)
	var changeVal string
	var changeIdx int
	consumed := false

	if !s.open {
		// When closed: Enter/Space opens popup
		if k.Key == term.KeyEnter || k.Key == term.KeySpace || (k.Rune == ' ' && k.Modifiers == 0) {
			s.open = true
			if s.selected >= 0 {
				s.cursor = s.selected
			} else {
				s.cursor = 0
			}
			consumed = true
		}
		s.mu.Unlock()
		return consumed
	}

	// When open: navigate popup
	switch k.Key {
	case term.KeyUp:
		if s.cursor > 0 {
			s.cursor--
		}
		consumed = true
	case term.KeyDown:
		if s.cursor < len(s.options)-1 {
			s.cursor++
		}
		consumed = true
	case term.KeyEnter:
		if s.cursor >= 0 && s.cursor < len(s.options) {
			s.selected = s.cursor
			s.open = false
			onChange = s.onChange
			changeIdx = s.selected
			if s.selected >= 0 && s.selected < len(s.options) {
				changeVal = s.options[s.selected].Value
			}
		}
		consumed = true
	case term.KeyEscape:
		s.open = false
		consumed = true
	case term.KeyHome:
		s.cursor = 0
		consumed = true
	case term.KeyEnd:
		if len(s.options) > 0 {
			s.cursor = len(s.options) - 1
		}
		consumed = true
	}

	if !consumed {
		// Vim keys
		switch k.Rune {
		case 'j':
			if k.Modifiers == 0 {
				if s.cursor < len(s.options)-1 {
					s.cursor++
				}
				consumed = true
			}
		case 'k':
			if k.Modifiers == 0 {
				if s.cursor > 0 {
					s.cursor--
				}
				consumed = true
			}
		case 'g':
			if k.Modifiers == 0 {
				s.cursor = 0
				consumed = true
			}
		case 'G':
			if k.Modifiers == 0 {
				if len(s.options) > 0 {
					s.cursor = len(s.options) - 1
				}
				consumed = true
			}
		}
	}

	s.mu.Unlock()

	// Fire callback outside lock
	if onChange != nil {
		onChange(changeVal, changeIdx)
	}

	return consumed
}

// HandleMouse processes mouse clicks.
func (s *Select) HandleMouse(x, y int, action string) bool {
	s.mu.Lock()
	var onChange func(string, int)
	var changeVal string
	var changeIdx int
	consumed := false

	if action != "down" {
		s.mu.Unlock()
		return false
	}

	b := s.Bounds()
	if !s.open {
		// Click on the select bar toggles it
		if x >= b.X && x < b.X+b.W && y >= b.Y && y < b.Y+b.H {
			s.open = true
			if s.selected >= 0 {
				s.cursor = s.selected
			}
			consumed = true
		}
		s.mu.Unlock()
		return consumed
	}

	// When open: check if click is on a popup option
	popupY := b.Y + b.H
	popupH := s.popupHeight
	if popupH > len(s.options) {
		popupH = len(s.options)
	}
	if y >= popupY && y < popupY+popupH && x >= b.X && x < b.X+b.W {
		idx := y - popupY
		if idx >= 0 && idx < len(s.options) {
			s.selected = idx
			s.open = false
			onChange = s.onChange
			changeIdx = s.selected
			if s.selected >= 0 && s.selected < len(s.options) {
				changeVal = s.options[s.selected].Value
			}
			consumed = true
		}
	} else {
		// Click outside closes popup
		s.open = false
		consumed = true
	}

	s.mu.Unlock()

	// Fire callback outside lock
	if onChange != nil {
		onChange(changeVal, changeIdx)
	}

	return consumed
}

// Measure returns the desired size: 1 row, min 10 cols.
func (s *Select) Measure(cs Constraints) Size {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w := s.width
	if w < 10 {
		w = 10
	}
	// Find longest label
	for _, opt := range s.options {
		lw := len([]rune(opt.Label))
		if lw+4 > w { // +4 for arrow and padding
			w = lw + 4
		}
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the Select component.
func (s *Select) Paint(buf *buffer.Buffer) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b := s.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	// Draw collapsed bar
	label := "(none)"
	if s.selected >= 0 && s.selected < len(s.options) {
		label = s.options[s.selected].Label
	}

	style := s.style.Normal
	if s.open {
		style = s.style.Active
	}

	// Draw label + arrow
	runeLabel := []rune(label)
	maxW := b.W - 2 // space for " ▼"
	for i := 0; i < maxW && i < len(runeLabel); i++ {
		buf.SetCell(b.X+i, b.Y, buffer.Cell{
			Rune:   runeLabel[i],
			Width:  1,
			Fg:     style.Fg,
			Bg:     style.Bg,
			Flags:  style.Flags,
		})
	}

	// Draw arrow indicator
	if b.W >= 2 {
		buf.SetCell(b.X+b.W-2, b.Y, buffer.Cell{
			Rune:  ' ',
			Width: 1,
			Fg:    style.Fg,
			Bg:    style.Bg,
		})
		arrow := '▼'
		if s.open {
			arrow = '▲'
		}
		buf.SetCell(b.X+b.W-1, b.Y, buffer.Cell{
			Rune:  arrow,
			Width: 1,
			Fg:    style.Fg,
			Bg:    style.Bg,
		})
	}

	// Draw popup if open
	if !s.open || len(s.options) == 0 {
		return
	}

	popupY := b.Y + b.H
	popupH := s.popupHeight
	if popupH > len(s.options) {
		popupH = len(s.options)
	}

	for i := 0; i < popupH; i++ {
		idx := i
		opt := s.options[idx]
		optStyle := s.style.Popup

		if idx == s.selected {
			optStyle = s.style.Selected
		}
		if idx == s.cursor {
			optStyle = s.style.Cursor
		}

		runeLabel := []rune(opt.Label)
		for x := 0; x < b.W; x++ {
			cell := buffer.Cell{Rune: ' ', Width: 1, Fg: optStyle.Fg, Bg: optStyle.Bg, Flags: optStyle.Flags}
			if x < len(runeLabel) {
				cell.Rune = runeLabel[x]
			}
			buf.SetCell(b.X+x, popupY+i, cell)
		}
	}
}
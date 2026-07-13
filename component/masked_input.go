package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── MaskedInput: Text Input with Template Mask ───
//
// MaskedInput enforces a template pattern on user input.
// Inspired by Textual's MaskedInput.
//
// Usage:
//	mi := NewMaskedInput("##/##/####")
//	// User types: 12312025 → display: "12/31/2025"
//	// Mask chars: # = any digit, A = any letter, * = any char, L = letter upper, l = letter lower
//	mi.Value() // "12/31/2025"

// MaskedInput is a text input that enforces a template mask.
type MaskedInput struct {
	mu       sync.RWMutex
	BaseComponent
	mask     string
	value    []rune // raw input chars (without mask literals)
	cursor   int    // cursor position in value
	fg       buffer.Color
	bg       buffer.Color
	focused  bool
	onChange func(string)
}

// NewMaskedInput creates a masked input with the given template.
// # = digit, A = letter (any case), L = uppercase letter, l = lowercase letter,
// * = any character. All other chars are literal (shown as-is, not editable).
func NewMaskedInput(mask string) *MaskedInput {
	return &MaskedInput{
		mask: mask,
		fg:   buffer.NamedColor(buffer.NamedWhite),
		bg:   buffer.Color{Type: buffer.ColorNone},
	}
}

// Value returns the full display string (input + mask literals).
func (m *MaskedInput) Value() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.displayLocked()
}

// RawValue returns only the user-entered characters (no mask literals).
func (m *MaskedInput) RawValue() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return string(m.value)
}

// SetRawValue sets the input directly, filtering through mask.
func (m *MaskedInput) SetRawValue(s string) {
	m.mu.Lock()
	m.value = nil
	m.cursor = 0
	m.mu.Unlock()
	for _, r := range s {
		m.insertChar(r)
	}
	m.notifyChange()
}

// SetFocus sets focus state.
func (m *MaskedInput) SetFocus(f bool) {
	m.mu.Lock()
	m.focused = f
	m.mu.Unlock()
}

// IsFocused returns focus state.
func (m *MaskedInput) IsFocused() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.focused
}

// SetOnChange sets callback.
func (m *MaskedInput) SetOnChange(fn func(string)) {
	m.mu.Lock()
	m.onChange = fn
	m.mu.Unlock()
}

// SetFg sets foreground color.
func (m *MaskedInput) SetFg(c buffer.Color) {
	m.mu.Lock()
	m.fg = c
	m.mu.Unlock()
}

// SetBg sets background color.
func (m *MaskedInput) SetBg(c buffer.Color) {
	m.mu.Lock()
	m.bg = c
	m.mu.Unlock()
}

// Mask returns the mask template.
func (m *MaskedInput) Mask() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mask
}

// Cursor returns cursor position in the value.
func (m *MaskedInput) Cursor() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cursor
}

// IsComplete returns true if all mask positions are filled.
func (m *MaskedInput) IsComplete() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.value) >= m.countMaskSlots()
}

// HandleKey processes keyboard input.
func (m *MaskedInput) HandleKey(ev *term.KeyEvent) bool {
	if ev == nil {
		return false
	}

	switch ev.Key {
	case term.KeyBackspace:
		m.mu.Lock()
		if m.cursor > 0 {
			m.cursor--
			m.value = append(m.value[:m.cursor], m.value[m.cursor+1:]...)
		}
		m.mu.Unlock()
		m.notifyChange()
		return true
	case term.KeyLeft:
		m.mu.Lock()
		if m.cursor > 0 {
			m.cursor--
		}
		m.mu.Unlock()
		return true
	case term.KeyRight:
		m.mu.Lock()
		if m.cursor < len(m.value) {
			m.cursor++
		}
		m.mu.Unlock()
		return true
	case term.KeyHome:
		m.mu.Lock()
		m.cursor = 0
		m.mu.Unlock()
		return true
	case term.KeyEnd:
		m.mu.Lock()
		m.cursor = len(m.value)
		m.mu.Unlock()
		return true
	}

	// Printable char
	if ev.Rune != 0 {
		if m.charMatchesMask(ev.Rune) {
			m.insertChar(ev.Rune)
		}
		return true
	}

	return false
}

func (m *MaskedInput) insertChar(r rune) {
	m.mu.Lock()
	if len(m.value) >= m.countMaskSlotsLocked() {
		m.mu.Unlock()
		return
	}
	// Apply case conversion for L/l mask chars
	slotType := m.slotTypeLocked(len(m.value))
	if !maskAcceptsChar(slotType, r) {
		m.mu.Unlock()
		return
	}
	switch slotType {
	case 'L':
		upper := strings.ToUpper(string(r))
		if len(upper) > 0 {
			r = rune(upper[0])
		}
	case 'l':
		lower := strings.ToLower(string(r))
		if len(lower) > 0 {
			r = rune(lower[0])
		}
	}
	if m.cursor >= len(m.value) {
		m.value = append(m.value, r)
	} else {
		m.value = append(m.value[:m.cursor], append([]rune{r}, m.value[m.cursor:]...)...)
	}
	m.cursor++
	m.mu.Unlock()
	m.notifyChange()
}

func (m *MaskedInput) notifyChange() {
	m.mu.RLock()
	cb := m.onChange
	val := m.displayLocked()
	m.mu.RUnlock()
	if cb != nil {
		cb(val)
	}
}

func (m *MaskedInput) charMatchesMask(r rune) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	slots := m.countMaskSlotsLocked()
	if len(m.value) >= slots {
		return false
	}
	maskChar := m.slotTypeLocked(len(m.value))
	return maskAcceptsChar(maskChar, r)
}

func (m *MaskedInput) slotTypeLocked(idx int) rune {
	count := 0
	for _, mc := range m.mask {
		if isMaskSlot(mc) {
			if count == idx {
				return mc
			}
			count++
		}
	}
	return 0
}

func (m *MaskedInput) countMaskSlotsLocked() int {
	count := 0
	for _, mc := range m.mask {
		if isMaskSlot(mc) {
			count++
		}
	}
	return count
}

func (m *MaskedInput) countMaskSlots() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.countMaskSlotsLocked()
}

// displayLocked builds the display string from mask + value.
func (m *MaskedInput) displayLocked() string {
	var sb strings.Builder
	vi := 0
	for _, mc := range m.mask {
		if isMaskSlot(mc) {
			if vi < len(m.value) {
				sb.WriteRune(m.value[vi])
				vi++
			} else {
				sb.WriteRune(mc) // show mask char as placeholder
			}
		} else {
			sb.WriteRune(mc) // literal
		}
	}
	return sb.String()
}

func isMaskSlot(c rune) bool {
	return c == '#' || c == 'A' || c == 'L' || c == 'l' || c == '*'
}

func maskAcceptsChar(maskChar, input rune) bool {
	switch maskChar {
	case '#':
		return input >= '0' && input <= '9'
	case 'A':
		return (input >= 'a' && input <= 'z') || (input >= 'A' && input <= 'Z')
	case 'L':
		return (input >= 'a' && input <= 'z') || (input >= 'A' && input <= 'Z')
	case 'l':
		return (input >= 'a' && input <= 'z') || (input >= 'A' && input <= 'Z')
	case '*':
		return true
	}
	return false
}

func (m *MaskedInput) Measure(constraints Constraints) Size {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w := len([]rune(m.mask))
	if w > constraints.MaxWidth && constraints.MaxWidth > 0 {
		w = constraints.MaxWidth
	}
	return Size{W: w, H: 1}
}

func (m *MaskedInput) Paint(buf *buffer.Buffer) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bounds := m.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	display := m.displayLocked()

	// Determine cursor position in display
	cursorDisplayX := 0
	vi := 0
	for i, mc := range m.mask {
		if isMaskSlot(mc) {
			if vi == m.cursor {
				cursorDisplayX = i
				break
			}
			if vi < len(m.value) {
				vi++
			}
		}
	}
	if m.cursor >= len(m.value) {
		cursorDisplayX = len([]rune(display))
		if cursorDisplayX > bounds.W {
			cursorDisplayX = bounds.W
		}
	}

	for i, r := range display {
		if i >= bounds.W {
			break
		}
		isSlot := isMaskSlot(rune(m.mask[i]))
		isFilled := false
		if isSlot {
			// Check if this slot is filled
			slotIdx := 0
			for j := 0; j < i; j++ {
				if isMaskSlot(rune(m.mask[j])) {
					slotIdx++
				}
			}
			isFilled = slotIdx < len(m.value)
		}

		fg := m.fg
		bg := m.bg
		var flags buffer.StyleFlags

		// Unfilled mask slots are dim
		if isSlot && !isFilled {
			flags |= buffer.Dim
		}

		// Cursor indicator
		if i == cursorDisplayX && m.focused {
			flags |= buffer.Reverse
		}

		buf.SetCell(bounds.X+i, bounds.Y, buffer.Cell{
			Rune:  r,
			Width: 1,
			Fg:    fg,
			Bg:    bg,
			Flags: flags,
		})
	}
}
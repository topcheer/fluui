package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// RadioGroupStyle holds the visual styles for different parts of a RadioGroup.
type RadioGroupStyle struct {
	Normal   buffer.Style // unselected item text
	Selected buffer.Style // cursor-highlighted item
	Active   buffer.Style // the selected (checked) radio item
	Disabled buffer.Style // disabled item
}

// DefaultRadioGroupStyle returns a sensible default style.
func DefaultRadioGroupStyle() RadioGroupStyle {
	return RadioGroupStyle{
		Normal:   buffer.Style{},
		Selected: buffer.Style{Flags: buffer.Reverse},
		Active:   buffer.Style{Fg: buffer.Green},
		Disabled: buffer.Style{Flags: buffer.Dim},
	}
}

// RadioGroup is a component that renders a list of mutually-exclusive options.
// Only one option can be selected at a time. Users navigate with Up/Down (or
// j/k), select with Enter or Space.
//
// Concurrent safe via sync.RWMutex.
type RadioGroup struct {
	BaseComponent
	mu sync.RWMutex

	labels []string
	cursor int   // currently highlighted item index
	active int   // currently selected item index (-1 if none)
	disabled map[int]bool
	style   RadioGroupStyle

	OnChange func(label string, index int) // fired on selection change
}

// NewRadioGroup creates a RadioGroup with the given option labels.
func NewRadioGroup(labels []string) *RadioGroup {
	rg := &RadioGroup{
		labels:   labels,
		active:   -1,
		disabled: make(map[int]bool),
		style:    DefaultRadioGroupStyle(),
	}
	rg.SetID(GenerateID("radiogroup"))
	return rg
}

// --- Item queries ---

// Labels returns a copy of the option labels.
func (r *RadioGroup) Labels() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]string, len(r.labels))
	copy(result, r.labels)
	return result
}

// ItemCount returns the number of options.
func (r *RadioGroup) ItemCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.labels)
}

// SelectedIndex returns the index of the selected option, or -1 if none.
func (r *RadioGroup) SelectedIndex() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.active
}

// SelectedLabel returns the label of the selected option, or "" if none.
func (r *RadioGroup) SelectedLabel() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.active < 0 || r.active >= len(r.labels) {
		return ""
	}
	return r.labels[r.active]
}

// IsDisabled returns whether the option at idx is disabled.
func (r *RadioGroup) IsDisabled(idx int) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.disabled[idx]
}

// SetDisabled enables or disables an option.
func (r *RadioGroup) SetDisabled(idx int, disabled bool) {
	r.mu.Lock()
	r.disabled[idx] = disabled
	// If the active item is being disabled, clear selection
	if disabled && r.active == idx {
		r.active = -1
	}
	r.mu.Unlock()
}

// --- Cursor navigation ---

// Cursor returns the current cursor index.
func (r *RadioGroup) Cursor() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cursor
}

// SetCursor sets the cursor, skipping disabled items.
func (r *RadioGroup) SetCursor(idx int) {
	r.mu.Lock()
	r.setNavigableCursor(idx)
	r.mu.Unlock()
}

// MoveUp moves the cursor up, wrapping around.
func (r *RadioGroup) MoveUp() {
	r.mu.Lock()
	r.moveCursor(-1)
	r.mu.Unlock()
}

// MoveDown moves the cursor down, wrapping around.
func (r *RadioGroup) MoveDown() {
	r.mu.Lock()
	r.moveCursor(1)
	r.mu.Unlock()
}

// setNavigableCursor sets cursor, skipping disabled. Caller must hold lock.
func (r *RadioGroup) setNavigableCursor(idx int) {
	n := len(r.labels)
	if n == 0 {
		r.cursor = 0
		return
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	orig := idx
	for i := 0; i < n; i++ {
		test := (orig + i) % n
		if !r.disabled[test] {
			r.cursor = test
			return
		}
	}
	r.cursor = orig
}

// moveCursor moves by delta, wrapping around and skipping disabled. Caller must hold lock.
func (r *RadioGroup) moveCursor(delta int) {
	n := len(r.labels)
	if n == 0 {
		return
	}
	for i := 0; i < n; i++ {
		r.cursor = (r.cursor + delta + n) % n
		if !r.disabled[r.cursor] {
			return
		}
	}
}

// --- Selection ---

// Select chooses the item at the cursor as the active option.
func (r *RadioGroup) Select() {
	r.mu.Lock()
	if r.cursor < len(r.labels) && !r.disabled[r.cursor] {
		r.active = r.cursor
	}
	label := ""
	idx := r.active
	if idx >= 0 && idx < len(r.labels) {
		label = r.labels[idx]
	}
	cb := r.OnChange
	r.mu.Unlock()
	if cb != nil && idx >= 0 {
		cb(label, idx)
	}
}

// SetSelected sets the active option by index.
func (r *RadioGroup) SetSelected(idx int) {
	r.mu.Lock()
	if idx >= 0 && idx < len(r.labels) && !r.disabled[idx] {
		r.active = idx
	}
	label := ""
	active := r.active
	if active >= 0 && active < len(r.labels) {
		label = r.labels[active]
	}
	cb := r.OnChange
	r.mu.Unlock()
	if cb != nil && active >= 0 {
		cb(label, active)
	}
}

// --- Style ---

// SetStyle sets the radio group style.
func (r *RadioGroup) SetStyle(s RadioGroupStyle) {
	r.mu.Lock()
	r.style = s
	r.mu.Unlock()
}

// Style returns the current style.
func (r *RadioGroup) Style() RadioGroupStyle {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.style
}

// --- Keyboard ---

// HandleKey processes a key event. Returns true if consumed.
func (r *RadioGroup) HandleKey(key *term.KeyEvent) bool {
	if key == nil {
		return false
	}
	switch {
	case key.Key == term.KeyUp || (key.Rune == 'k' && key.Key == term.KeyUnknown):
		r.MoveUp()
		return true
	case key.Key == term.KeyDown || (key.Rune == 'j' && key.Key == term.KeyUnknown):
		r.MoveDown()
		return true
	case key.Key == term.KeyEnter:
		r.Select()
		return true
	case key.Key == term.KeySpace:
		r.Select()
		return true
	default:
		return false
	}
}

// --- Component interface ---

// Measure returns the desired size based on the longest label.
func (r *RadioGroup) Measure(cs Constraints) Size {
	r.mu.RLock()
	defer r.mu.RUnlock()

	maxLen := 0
	for _, label := range r.labels {
		l := runeLen(label) + 6 // "  (o) " prefix
		if l > maxLen {
			maxLen = l
		}
	}
	w := maxLen
	h := len(r.labels)
	if h == 0 {
		h = 1
	}
	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.HasHeight() && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the radio group into the buffer.
func (r *RadioGroup) Paint(buf *buffer.Buffer) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bounds := r.Bounds()
	if bounds.W == 0 || bounds.H == 0 {
		return
	}

	for row, label := range r.labels {
		if row >= bounds.H {
			break
		}

		isCursor := row == r.cursor
		isActive := row == r.active
		isDisabled := r.disabled[row]

		var style buffer.Style
		switch {
			case isDisabled:
			style = r.style.Disabled
			case isCursor:
			style = r.style.Selected
			default:
			style = r.style.Normal
		}

		// Draw radio indicator
		var indicator string
		if isActive {
			indicator = "(o) "
		} else {
			indicator = "( ) "
		}

		col := bounds.X
		for _, rn := range indicator {
			if col-bounds.X >= bounds.W {
				break
			}
			cellStyle := style
			if isActive {
				cellStyle = r.style.Active
			}
			buf.SetCell(col, bounds.Y+row, buffer.Cell{Rune: rn, Width: 1, Fg: cellStyle.Fg, Bg: cellStyle.Bg, Flags: cellStyle.Flags})
			col++
		}

		// Paint label
		for _, rn := range label {
			if col-bounds.X >= bounds.W {
				break
			}
			buf.SetCell(col, bounds.Y+row, buffer.Cell{Rune: rn, Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
			col++
		}
	}
}

// String returns a string representation.
func (r *RadioGroup) String() string {
	return "RadioGroup"
}

package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// CheckboxItem represents a single selectable item in a Checkbox list.
type CheckboxItem struct {
	Label    string
	Checked  bool
	Disabled bool
}

// CheckboxStyle holds the visual styles for different parts of a Checkbox list.
type CheckboxStyle struct {
	Normal   buffer.Style // unchecked item text
	Selected buffer.Style // cursor-highlighted item (not necessarily checked)
	Checked  buffer.Style // checkmark indicator style
	Disabled buffer.Style // disabled item
}

// DefaultCheckboxStyle returns a sensible default style.
func DefaultCheckboxStyle() CheckboxStyle {
	return CheckboxStyle{
		Normal:   buffer.Style{},
		Selected: buffer.Style{Flags: buffer.Reverse},
		Checked:  buffer.Style{Fg: buffer.Green},
		Disabled: buffer.Style{Flags: buffer.Dim},
	}
}

// Checkbox is a component that renders a list of toggleable checkboxes.
// Users navigate with Up/Down (or j/k), toggle with Space, select-all with
// Ctrl+A, and deselect-all with Ctrl+D.
//
// Concurrent safe via sync.RWMutex.
type Checkbox struct {
	BaseComponent
	mu sync.RWMutex

	items  []CheckboxItem
	cursor int // currently highlighted item index
	style  CheckboxStyle

	OnChange func(checked []CheckboxItem) // fired on any toggle
}

// NewCheckbox creates a Checkbox component with the given item labels.
func NewCheckbox(labels []string) *Checkbox {
	items := make([]CheckboxItem, len(labels))
	for i, l := range labels {
		items[i] = CheckboxItem{Label: l}
	}
	cb := &Checkbox{
		items: items,
		style: DefaultCheckboxStyle(),
	}
	cb.SetID(GenerateID("checkbox"))
	return cb
}

// --- Item management ---

// Items returns a copy of the current items.
func (c *Checkbox) Items() []CheckboxItem {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]CheckboxItem, len(c.items))
	copy(result, c.items)
	return result
}

// SetItems replaces all items.
func (c *Checkbox) SetItems(items []CheckboxItem) {
	c.mu.Lock()
	c.items = make([]CheckboxItem, len(items))
	copy(c.items, items)
	if c.cursor >= len(c.items) {
		c.cursor = 0
	}
	c.mu.Unlock()
}

// AddItem appends a new checkbox item.
func (c *Checkbox) AddItem(label string) {
	c.mu.Lock()
	c.items = append(c.items, CheckboxItem{Label: label})
	c.mu.Unlock()
}

// ItemCount returns the number of items.
func (c *Checkbox) ItemCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// CheckedItems returns only the checked items (copy).
func (c *Checkbox) CheckedItems() []CheckboxItem {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result []CheckboxItem
	for _, item := range c.items {
		if item.Checked {
			result = append(result, item)
		}
	}
	return result
}

// CheckedLabels returns labels of all checked items.
func (c *Checkbox) CheckedLabels() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result []string
	for _, item := range c.items {
		if item.Checked {
			result = append(result, item.Label)
		}
	}
	return result
}

// --- Cursor navigation ---

// Cursor returns the current cursor index.
func (c *Checkbox) Cursor() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cursor
}

// SetCursor sets the cursor index, clamped to valid range and skipping disabled items.
func (c *Checkbox) SetCursor(idx int) {
	c.mu.Lock()
	c.setNavigableCursor(idx)
	c.mu.Unlock()
}

// MoveUp moves the cursor up, wrapping around.
func (c *Checkbox) MoveUp() {
	c.mu.Lock()
	c.moveCursor(-1)
	c.mu.Unlock()
}

// MoveDown moves the cursor down, wrapping around.
func (c *Checkbox) MoveDown() {
	c.mu.Lock()
	c.moveCursor(1)
	c.mu.Unlock()
}

// setNavigableCursor sets cursor to idx, skipping disabled items. Caller must hold lock.
func (c *Checkbox) setNavigableCursor(idx int) {
	n := len(c.items)
	if n == 0 {
		c.cursor = 0
		return
	}
	// Clamp
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	// Skip disabled (search forward, then backward)
	orig := idx
	for i := 0; i < n; i++ {
		test := (orig + i) % n
		if !c.items[test].Disabled {
			c.cursor = test
			return
		}
	}
	// All disabled
	c.cursor = orig
}

// moveCursor moves by delta, wrapping around and skipping disabled items. Caller must hold lock.
func (c *Checkbox) moveCursor(delta int) {
	n := len(c.items)
	if n == 0 {
		return
	}
	for i := 0; i < n; i++ {
		c.cursor = (c.cursor + delta + n) % n
		if !c.items[c.cursor].Disabled {
			return
		}
	}
}

// --- Toggle operations ---

// Toggle flips the checked state of the item at the cursor.
func (c *Checkbox) Toggle() {
	c.mu.Lock()
	if c.cursor < len(c.items) && !c.items[c.cursor].Disabled {
		c.items[c.cursor].Checked = !c.items[c.cursor].Checked
	}
	items := make([]CheckboxItem, len(c.items))
	copy(items, c.items)
	c.mu.Unlock()
	if c.OnChange != nil {
		c.OnChange(items)
	}
}

// SetChecked sets the checked state of the item at idx.
func (c *Checkbox) SetChecked(idx int, checked bool) {
	c.mu.Lock()
	if idx >= 0 && idx < len(c.items) && !c.items[idx].Disabled {
		c.items[idx].Checked = checked
	}
	items := make([]CheckboxItem, len(c.items))
	copy(items, c.items)
	c.mu.Unlock()
	if c.OnChange != nil {
		c.OnChange(items)
	}
}

// IsChecked returns whether the item at idx is checked.
func (c *Checkbox) IsChecked(idx int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if idx < 0 || idx >= len(c.items) {
		return false
	}
	return c.items[idx].Checked
}

// CheckAll checks all enabled items.
func (c *Checkbox) CheckAll() {
	c.mu.Lock()
	for i := range c.items {
		if !c.items[i].Disabled {
			c.items[i].Checked = true
		}
	}
	items := make([]CheckboxItem, len(c.items))
	copy(items, c.items)
	c.mu.Unlock()
	if c.OnChange != nil {
		c.OnChange(items)
	}
}

// UncheckAll unchecks all items.
func (c *Checkbox) UncheckAll() {
	c.mu.Lock()
	for i := range c.items {
		c.items[i].Checked = false
	}
	items := make([]CheckboxItem, len(c.items))
	copy(items, c.items)
	c.mu.Unlock()
	if c.OnChange != nil {
		c.OnChange(items)
	}
}

// --- Style ---

// SetStyle sets the checkbox style.
func (c *Checkbox) SetStyle(s CheckboxStyle) {
	c.mu.Lock()
	c.style = s
	c.mu.Unlock()
}

// Style returns the current style.
func (c *Checkbox) Style() CheckboxStyle {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.style
}

// --- Keyboard ---

// HandleKey processes a key event. Returns true if the key was consumed.
func (c *Checkbox) HandleKey(key *term.KeyEvent) bool {
	if key == nil {
		return false
	}
	switch {
	case key.Key == term.KeyUp || (key.Rune == 'k' && key.Key == term.KeyUnknown):
		c.MoveUp()
		return true
	case key.Key == term.KeyDown || (key.Rune == 'j' && key.Key == term.KeyUnknown):
		c.MoveDown()
		return true
	case key.Key == term.KeySpace:
		c.Toggle()
		return true
	case key.Key == term.KeyUnknown && key.Rune == 'a' && key.Modifiers&term.ModCtrl != 0:
		c.CheckAll()
		return true
	case key.Key == term.KeyUnknown && key.Rune == 'd' && key.Modifiers&term.ModCtrl != 0:
		c.UncheckAll()
		return true
	default:
		return false
	}
}

// --- Component interface ---

// Measure returns the desired size based on the longest label.
func (c *Checkbox) Measure(cs Constraints) Size {
	c.mu.RLock()
	defer c.mu.RUnlock()

	maxLen := 0
	for _, item := range c.items {
		l := runeLen(item.Label) + 6 // "  [x] " prefix
		if l > maxLen {
			maxLen = l
		}
	}
	w := maxLen
	h := len(c.items)
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

// Paint renders the checkbox list into the buffer.
func (c *Checkbox) Paint(buf *buffer.Buffer) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	bounds := c.Bounds()
	if bounds.W == 0 || bounds.H == 0 {
		return
	}

	for row, item := range c.items {
		if row >= bounds.H {
			break
		}

		isCursor := row == c.cursor

		// Choose style
		var style buffer.Style
		switch {
			case item.Disabled:
			style = c.style.Disabled
			case isCursor:
			style = c.style.Selected
			default:
			style = c.style.Normal
		}

		// Draw indicator
		var indicator string
		if item.Checked {
			indicator = "[x] "
		} else {
			indicator = "[ ] "
		}

		// Paint indicator
		col := bounds.X
		for _, r := range indicator {
			if col-bounds.X >= bounds.W {
				break
			}
			cellStyle := style
			if item.Checked {
				cellStyle = c.style.Checked
			}
			buf.SetCell(col, bounds.Y+row, buffer.Cell{Rune: r, Width: 1, Fg: cellStyle.Fg, Bg: cellStyle.Bg, Flags: cellStyle.Flags})
			col++
		}

		// Paint label
		for _, r := range item.Label {
			if col-bounds.X >= bounds.W {
				break
			}
			buf.SetCell(col, bounds.Y+row, buffer.Cell{Rune: r, Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
			col++
		}
	}
}

// String returns a string representation.
func (c *Checkbox) String() string {
	return "Checkbox"
}

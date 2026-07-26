package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// ComboBox combines a text input with a dropdown list. As the user types,
// items are filtered. Common in search-driven forms and AI model selectors.
//
// Thread-safe.
type ComboBox struct {
	BaseComponent
	mu       sync.RWMutex
	items    []string
	filtered []string
	query    string
	selected int
	expanded bool
}

// NewComboBox creates a combo box with the given items.
func NewComboBox(items []string) *ComboBox {
	cb := &ComboBox{
		BaseComponent: BaseComponent{id: GenerateID("combobox")},
		items:         items,
	}
	cb.applyFilterLocked()
	return cb
}

func (c *ComboBox) Items() []string { c.mu.RLock(); defer c.mu.RUnlock(); return c.items }

func (c *ComboBox) SetItems(items []string) {
	c.mu.Lock(); defer c.mu.Unlock()
	c.items = items
	c.applyFilterLocked()
}

func (c *ComboBox) Query() string { c.mu.RLock(); defer c.mu.RUnlock(); return c.query }

func (c *ComboBox) SetQuery(q string) {
	c.mu.Lock(); defer c.mu.Unlock()
	c.query = q
	c.applyFilterLocked()
	c.expanded = len(q) > 0 && len(c.filtered) > 0
}

func (c *ComboBox) Filtered() []string { c.mu.RLock(); defer c.mu.RUnlock(); return c.filtered }

func (c *ComboBox) Selected() int { c.mu.RLock(); defer c.mu.RUnlock(); return c.selected }

func (c *ComboBox) SelectedItem() string {
	c.mu.RLock(); defer c.mu.RUnlock()
	if c.selected < 0 || c.selected >= len(c.filtered) { return "" }
	return c.filtered[c.selected]
}

func (c *ComboBox) Expanded() bool { c.mu.RLock(); defer c.mu.RUnlock(); return c.expanded }

func (c *ComboBox) SetExpanded(b bool) { c.mu.Lock(); defer c.mu.Unlock(); c.expanded = b }

func (c *ComboBox) Collapse() { c.mu.Lock(); defer c.mu.Unlock(); c.expanded = false }

func (c *ComboBox) MoveUp() {
	c.mu.Lock(); defer c.mu.Unlock()
	if c.selected > 0 { c.selected-- }
}

func (c *ComboBox) MoveDown() {
	c.mu.Lock(); defer c.mu.Unlock()
	if c.selected < len(c.filtered)-1 { c.selected++ }
}

func (c *ComboBox) SelectCurrent() {
	c.mu.Lock(); defer c.mu.Unlock()
	if c.selected >= 0 && c.selected < len(c.filtered) {
		c.query = c.filtered[c.selected]
		c.applyFilterLocked()
	}
	c.expanded = false
}

func (c *ComboBox) applyFilterLocked() {
	if c.query == "" {
		c.filtered = c.items
		if c.selected >= len(c.items) { c.selected = 0 }
		return
	}
	c.filtered = c.filtered[:0]
	q := c.query
	for i := 0; i < len(q); i++ {
		if q[i] >= 'A' && q[i] <= 'Z' {
			b := []byte(q)
			for j := range b { if b[j] >= 'A' && b[j] <= 'Z' { b[j] += 32 } }
			q = string(b)
			break
		}
	}
	for _, item := range c.items {
		if containsCI(item, q) { c.filtered = append(c.filtered, item) }
	}
	if c.selected >= len(c.filtered) && len(c.filtered) > 0 { c.selected = 0 }
}

func (c *ComboBox) Measure(cs Constraints) Size {
	c.mu.RLock(); defer c.mu.RUnlock()
	w := len(c.query)
	for _, item := range c.filtered { if iw := len(item); iw > w { w = iw } }
	w += 2
	h := 1
	if c.expanded { h = 1 + len(c.filtered); if h > 10 { h = 10 } }
	if cs.HasWidth() && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.HasHeight() && h > cs.MaxHeight { h = cs.MaxHeight }
	if w < 1 { w = 1 }
	if h < 1 { h = 1 }
	return Size{W: w, H: h}
}

func (c *ComboBox) Paint(buf *buffer.Buffer) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	bounds := c.bounds
	if bounds.W <= 0 || bounds.H <= 0 { return }

	tt := theme.Get()
	normal := buffer.Style{Fg: tt.Fg}
	selected := buffer.Style{Fg: tt.Bg, Bg: tt.Accent, Flags: buffer.Bold}
	muted := buffer.Style{Fg: tt.Muted}

	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	// Input line
	if c.query != "" {
		x = buf.DrawText(x, y, c.query, normal)
	} else {
		x = buf.DrawText(x, y, "type to search...", muted)
	}
	// Cursor
	if x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: '\u2588', Width: 1, Fg: tt.Accent})
	}

	// Filtered dropdown
	if c.expanded && bounds.H > 1 {
		itemY := y + 1
		for i, item := range c.filtered {
			if itemY >= bounds.Y+bounds.H { break }
			style := normal
			if i == c.selected { style = selected }
			for col := 0; col < bounds.W && bounds.X+col < maxX; col++ {
				if col < len(item) {
					buf.SetCell(bounds.X+col, itemY, buffer.Cell{
						Rune: rune(item[col]), Width: 1,
						Fg: style.Fg, Bg: style.Bg, Flags: style.Flags,
					})
				} else {
					buf.SetCell(bounds.X+col, itemY, buffer.Cell{Rune: ' ', Width: 1, Bg: style.Bg})
				}
			}
			itemY++
		}
	}
}

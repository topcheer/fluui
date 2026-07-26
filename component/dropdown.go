package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// DropdownItem represents a single selectable option in a Dropdown.
type DropdownItem struct {
	Label string
	Value string
}

// Dropdown renders a compact selection widget with a collapsed label
// that expands to show a list of options. Common in forms and settings.
//
// Thread-safe.
type Dropdown struct {
	BaseComponent
	mu        sync.RWMutex
	items     []DropdownItem
	selected  int
	expanded  bool
	maxHeight int
	label     string // optional label prefix
}

// NewDropdown creates a dropdown with the given items.
func NewDropdown(items []DropdownItem) *Dropdown {
	return &Dropdown{
		BaseComponent: BaseComponent{id: GenerateID("dropdown")},
		items:         items,
		maxHeight:     8,
	}
}

func (d *Dropdown) Items() []DropdownItem { d.mu.RLock(); defer d.mu.RUnlock(); return d.items }

func (d *Dropdown) SetItems(items []DropdownItem) {
	d.mu.Lock(); defer d.mu.Unlock()
	d.items = items
	if d.selected >= len(items) { d.selected = 0 }
}

func (d *Dropdown) Selected() int { d.mu.RLock(); defer d.mu.RUnlock(); return d.selected }

func (d *Dropdown) SetSelected(i int) {
	d.mu.Lock(); defer d.mu.Unlock()
	if i >= 0 && i < len(d.items) { d.selected = i }
}

func (d *Dropdown) SelectedItem() DropdownItem {
	d.mu.RLock(); defer d.mu.RUnlock()
	if d.selected < 0 || d.selected >= len(d.items) { return DropdownItem{} }
	return d.items[d.selected]
}

func (d *Dropdown) Expanded() bool { d.mu.RLock(); defer d.mu.RUnlock(); return d.expanded }

func (d *Dropdown) SetExpanded(b bool) { d.mu.Lock(); defer d.mu.Unlock(); d.expanded = b }

func (d *Dropdown) Toggle() { d.mu.Lock(); defer d.mu.Unlock(); d.expanded = !d.expanded }

func (d *Dropdown) Label() string { d.mu.RLock(); defer d.mu.RUnlock(); return d.label }

func (d *Dropdown) SetLabel(s string) { d.mu.Lock(); defer d.mu.Unlock(); d.label = s }

func (d *Dropdown) MaxHeight() int { d.mu.RLock(); defer d.mu.RUnlock(); return d.maxHeight }

func (d *Dropdown) SetMaxHeight(h int) {
	d.mu.Lock(); defer d.mu.Unlock()
	if h < 1 { h = 1 }
	d.maxHeight = h
}

func (d *Dropdown) MoveUp() {
	d.mu.Lock(); defer d.mu.Unlock()
	if d.selected > 0 { d.selected-- }
}

func (d *Dropdown) MoveDown() {
	d.mu.Lock(); defer d.mu.Unlock()
	if d.selected < len(d.items)-1 { d.selected++ }
}

func (d *Dropdown) Measure(cs Constraints) Size {
	d.mu.RLock(); defer d.mu.RUnlock()
	w := 5 // "[ X ]▼"
	for _, item := range d.items {
		if iw := len(item.Label); iw > w { w = iw }
	}
	w += 3 // brackets + arrow
	if d.label != "" { w += len(d.label) + 1 }
	h := 1
	if d.expanded {
		h = 1 + len(d.items)
		if h > d.maxHeight { h = d.maxHeight }
	}
	if cs.HasWidth() && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.HasHeight() && h > cs.MaxHeight { h = cs.MaxHeight }
	if w < 1 { w = 1 }
	if h < 1 { h = 1 }
	return Size{W: w, H: h}
}

func (d *Dropdown) Paint(buf *buffer.Buffer) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	bounds := d.bounds
	if bounds.W <= 0 || bounds.H <= 0 { return }

	tt := theme.Get()
	muted := buffer.Style{Fg: tt.Muted}
	normal := buffer.Style{Fg: tt.Fg}
	selected := buffer.Style{Fg: tt.Bg, Bg: tt.Accent, Flags: buffer.Bold}
	accent := buffer.Style{Fg: tt.Accent}

	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	// Label
	if d.label != "" {
		x = buf.DrawText(x, y, d.label+" ", muted)
	}

	// Collapsed header: [SelectedItem] ▼
	headerText := ""
	if d.selected >= 0 && d.selected < len(d.items) {
		headerText = d.items[d.selected].Label
	}
	style := normal
	hdr := "[ " + headerText + " ]"
	x = buf.DrawText(x, y, hdr, style)
	if x < maxX {
		x = buf.DrawText(x, y, "\u25bc", accent) // ▼
	}

	// Expanded items
	if d.expanded && bounds.H > 1 {
		itemY := y + 1
		for i, item := range d.items {
			if itemY >= bounds.Y+bounds.H { break }
			itemStyle := normal
			if i == d.selected { itemStyle = selected }
			ix := bounds.X
			if d.label != "" { ix = bounds.X + len(d.label) + 1 }
			for col := 0; col < bounds.W-(ix-bounds.X) && ix+col < maxX; col++ {
				if col < len(item.Label) {
					buf.SetCell(ix+col, itemY, buffer.Cell{
						Rune: rune(item.Label[col]), Width: 1,
						Fg: itemStyle.Fg, Bg: itemStyle.Bg, Flags: itemStyle.Flags,
					})
				} else {
					buf.SetCell(ix+col, itemY, buffer.Cell{Rune: ' ', Width: 1, Bg: itemStyle.Bg})
				}
			}
			itemY++
		}
	}
}

package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── MenuItem ──────────────────────────────────────────────────

// MenuItem represents a single entry in a context menu.
type MenuItem struct {
	ID        string       // unique identifier within the menu
	Label     string       // display text
	Shortcut  string       // optional shortcut hint (e.g. "Ctrl+C")
	Icon      string       // optional icon prefix
	Enabled   bool         // whether the item is interactive
	Separator bool         // if true, render as a horizontal divider
	Submenu   *ContextMenu // optional nested submenu
	Action    func()       // callback fired when activated
}

// NewMenuItem creates a standard, enabled menu item.
func NewMenuItem(id, label string) *MenuItem {
	return &MenuItem{ID: id, Label: label, Enabled: true}
}

// NewSeparator returns a separator item.
func NewSeparator() *MenuItem {
	return &MenuItem{ID: "sep", Separator: true}
}

// SetShortcut sets the keyboard shortcut hint text.
func (mi *MenuItem) SetShortcut(s string) *MenuItem { mi.Shortcut = s; return mi }

// SetIcon sets the icon prefix.
func (mi *MenuItem) SetIcon(icon string) *MenuItem { mi.Icon = icon; return mi }

// SetEnabled toggles the enabled state.
func (mi *MenuItem) SetEnabled(en bool) *MenuItem { mi.Enabled = en; return mi }

// SetSubmenu attaches a nested submenu.
func (mi *MenuItem) SetSubmenu(cm *ContextMenu) *MenuItem { mi.Submenu = cm; return mi }

// SetAction sets the activation callback.
func (mi *MenuItem) SetAction(fn func()) *MenuItem { mi.Action = fn; return mi }

// HasSubmenu returns true if the item has a nested submenu.
func (mi *MenuItem) HasSubmenu() bool { return mi.Submenu != nil }

// ─── Style ─────────────────────────────────────────────────────

// ContextMenuStyle holds the visual styling for a context menu.
type ContextMenuStyle struct {
	Border    buffer.Style
	Normal    buffer.Style
	Selected  buffer.Style
	Disabled  buffer.Style
	Separator buffer.Style
	Shortcut  buffer.Style
}

// DefaultContextMenuStyle returns a sensible default.
func DefaultContextMenuStyle() ContextMenuStyle {
	return ContextMenuStyle{
		Border:    buffer.DefaultStyle,
		Normal:    buffer.DefaultStyle,
		Selected:  buffer.Style{Flags: buffer.Reverse},
		Disabled:  buffer.Style{Flags: buffer.Dim},
		Separator: buffer.Style{Flags: buffer.Dim},
		Shortcut:  buffer.Style{Flags: buffer.Dim},
	}
}

// ─── ContextMenu ───────────────────────────────────────────────

// ContextMenu is a popup menu that displays a list of actions.
// It implements the Component interface and can be embedded in an overlay.
type ContextMenu struct {
	BaseComponent
	mu sync.RWMutex

	items   []*MenuItem
	cursor  int // index of highlighted item
	style   ContextMenuStyle
	visible bool
	x, y    int // anchor (top-left of menu in screen coordinates)
	width   int // computed menu width

	// OnClose is called when the menu is dismissed (Esc or outside click).
	OnClose func()
	// OnSelect is called when an item is activated (optional override).
	OnSelect func(item *MenuItem)
}

// NewContextMenu creates an empty context menu with a generated ID.
func NewContextMenu() *ContextMenu {
	cm := &ContextMenu{
		style:  DefaultContextMenuStyle(),
		cursor: 0,
	}
	cm.SetID(GenerateID("contextmenu"))
	return cm
}

// ─── Item management ───────────────────────────────────────────

// AddItem appends a menu item and returns it for chaining.
func (cm *ContextMenu) AddItem(item *MenuItem) *MenuItem {
	cm.mu.Lock()
	cm.items = append(cm.items, item)
	cm.mu.Unlock()
	return item
}

// AddLabel is a convenience that appends a simple enabled item.
func (cm *ContextMenu) AddLabel(id, label string) *MenuItem {
	return cm.AddItem(NewMenuItem(id, label))
}

// AddSeparator appends a separator item.
func (cm *ContextMenu) AddSeparator() {
	cm.mu.Lock()
	cm.items = append(cm.items, NewSeparator())
	cm.mu.Unlock()
}

// Remove removes an item by ID. Returns true if found.
func (cm *ContextMenu) Remove(id string) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	for i, it := range cm.items {
		if it.ID == id {
			cm.items = append(cm.items[:i], cm.items[i+1:]...)
			if cm.cursor >= len(cm.items) {
				cm.cursor = len(cm.items) - 1
			}
			return true
		}
	}
	return false
}

// Clear removes all items.
func (cm *ContextMenu) Clear() {
	cm.mu.Lock()
	cm.items = nil
	cm.cursor = 0
	cm.mu.Unlock()
}

// Items returns a copy of all items.
func (cm *ContextMenu) Items() []*MenuItem {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	out := make([]*MenuItem, len(cm.items))
	copy(out, cm.items)
	return out
}

// ItemCount returns the total number of items (including separators).
func (cm *ContextMenu) ItemCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.items)
}

// ItemAt returns the item at index, or nil.
func (cm *ContextMenu) ItemAt(idx int) *MenuItem {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if idx < 0 || idx >= len(cm.items) {
		return nil
	}
	return cm.items[idx]
}

// Find returns the first item with the given ID, or nil.
func (cm *ContextMenu) Find(id string) *MenuItem {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	for _, it := range cm.items {
		if it.ID == id {
			return it
		}
	}
	return nil
}

// ─── Cursor / navigation ───────────────────────────────────────

// Cursor returns the current cursor index.
func (cm *ContextMenu) Cursor() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.cursor
}

// SetCursor sets the cursor, skipping separators and disabled items.
func (cm *ContextMenu) SetCursor(idx int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.setCursorLocked(idx)
}

func (cm *ContextMenu) setCursorLocked(idx int) {
	if len(cm.items) == 0 {
		cm.cursor = 0
		return
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(cm.items) {
		idx = len(cm.items) - 1
	}
	// Forward search for first navigable item
	for i := idx; i < len(cm.items); i++ {
		if cm.navigableLocked(i) {
			cm.cursor = i
			return
		}
	}
	// Backward search
	for i := idx; i >= 0; i-- {
		if cm.navigableLocked(i) {
			cm.cursor = i
			return
		}
	}
}

// navigableLocked returns true if item at idx can be highlighted.
func (cm *ContextMenu) navigableLocked(idx int) bool {
	if idx < 0 || idx >= len(cm.items) {
		return false
	}
	it := cm.items[idx]
	return !it.Separator && it.Enabled
}

// MoveUp moves the cursor up by one, wrapping around.
func (cm *ContextMenu) MoveUp() { cm.moveCursor(-1) }

// MoveDown moves the cursor down by one, wrapping around.
func (cm *ContextMenu) MoveDown() { cm.moveCursor(1) }

func (cm *ContextMenu) moveCursor(delta int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if len(cm.items) == 0 {
		return
	}
	dir := 1
	if delta < 0 {
		dir = -1
	}
	pos := cm.cursor
	for steps := 0; steps <= len(cm.items); steps++ {
		pos += dir
		if pos >= len(cm.items) {
			pos = 0
		}
		if pos < 0 {
			pos = len(cm.items) - 1
		}
		if cm.navigableLocked(pos) {
			cm.cursor = pos
			return
		}
	}
}

// CurrentItem returns the item under the cursor, or nil.
func (cm *ContextMenu) CurrentItem() *MenuItem {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if cm.cursor < 0 || cm.cursor >= len(cm.items) {
		return nil
	}
	return cm.items[cm.cursor]
}

// ─── Visibility / positioning ──────────────────────────────────

// Show makes the menu visible at the given screen coordinates.
func (cm *ContextMenu) Show(x, y int) {
	cm.mu.Lock()
	cm.visible = true
	cm.x = x
	cm.y = y
	cm.cursor = 0
	for i, it := range cm.items {
		if !it.Separator && it.Enabled {
			cm.cursor = i
			break
		}
	}
	cm.mu.Unlock()
}

// Hide makes the menu invisible and fires OnClose.
func (cm *ContextMenu) Hide() {
	cm.mu.Lock()
	cm.visible = false
	// Close any open submenu
	if cm.cursor >= 0 && cm.cursor < len(cm.items) {
		if sub := cm.items[cm.cursor].Submenu; sub != nil {
			sub.Hide()
		}
	}
	cb := cm.OnClose
	cm.mu.Unlock()
	if cb != nil {
		cb()
	}
}

// Visible returns whether the menu is currently shown.
func (cm *ContextMenu) Visible() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.visible
}

// Position returns the menu's anchor (x, y).
func (cm *ContextMenu) Position() (int, int) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.x, cm.y
}

// SetPosition sets the menu's anchor (x, y).
func (cm *ContextMenu) SetPosition(x, y int) {
	cm.mu.Lock()
	cm.x = x
	cm.y = y
	cm.mu.Unlock()
}

// ─── Activation ────────────────────────────────────────────────

// Activate triggers the action of the item under the cursor.
// If the item has a submenu, the submenu is shown instead.
// Returns the activated item, or nil.
func (cm *ContextMenu) Activate() *MenuItem {
	cm.mu.Lock()
	if cm.cursor < 0 || cm.cursor >= len(cm.items) {
		cm.mu.Unlock()
		return nil
	}
	item := cm.items[cm.cursor]
	if item.Separator || !item.Enabled {
		cm.mu.Unlock()
		return nil
	}
	action := item.Action
	onSelect := cm.OnSelect
	sub := item.Submenu
	width := cm.width
	startX, startRow := cm.x, cm.cursor
	cm.mu.Unlock()

	if sub != nil {
		sub.Show(startX+width+1, startRow+1)
		return item
	}
	if action != nil {
		action()
	}
	if onSelect != nil {
		onSelect(item)
	}
	return item
}

// ─── Style ─────────────────────────────────────────────────────

// SetStyle overrides the menu's style.
func (cm *ContextMenu) SetStyle(s ContextMenuStyle) {
	cm.mu.Lock()
	cm.style = s
	cm.mu.Unlock()
}

// Style returns the current style.
func (cm *ContextMenu) Style() ContextMenuStyle {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.style
}

// ─── Keyboard handling ─────────────────────────────────────────

// HandleKey processes keyboard navigation for the menu.
// Returns true if the key was consumed.
func (cm *ContextMenu) HandleKey(key *term.KeyEvent) bool {
	cm.mu.RLock()
	if !cm.visible {
		cm.mu.RUnlock()
		return false
	}
	// If a submenu is visible, forward the key
	if cm.cursor >= 0 && cm.cursor < len(cm.items) {
		if sub := cm.items[cm.cursor].Submenu; sub != nil && sub.Visible() {
			cm.mu.RUnlock()
			if sub.HandleKey(key) {
				return true
			}
			// Submenu didn't consume — check Left to close submenu
			if key.Key == term.KeyLeft {
				sub.Hide()
				return true
			}
			return false
		}
	}
	cm.mu.RUnlock()

	switch key.Key {
	case term.KeyUp:
		cm.MoveUp()
		return true
	case term.KeyDown:
		cm.MoveDown()
		return true
	case term.KeyEnter:
		cm.Activate()
		return true
	case term.KeyEscape:
		cm.Hide()
		return true
	case term.KeyRight:
		// Open submenu if current item has one
		if item := cm.CurrentItem(); item != nil && item.HasSubmenu() {
			cm.Activate() // Activate opens submenu
			return true
		}
		return false
	default:
		return false
	}
}

// ─── Component interface ───────────────────────────────────────

// Measure computes the desired size.
func (cm *ContextMenu) Measure(cs Constraints) Size {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	maxW := 0
	for _, it := range cm.items {
		w := measureItemWidth(it)
		if w > maxW {
			maxW = w
		}
	}
	width := maxW + 4 // 2 content padding + 2 borders
	height := len(cm.items) + 2
	if height < 3 {
		height = 3
	}
	if width < 10 {
		width = 10
	}

	// Apply constraints
	if cs.MaxWidth > 0 && width > cs.MaxWidth {
		width = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && height > cs.MaxHeight {
		height = cs.MaxHeight
	}
	cm.width = width
	return Size{W: width, H: height}
}

// measureItemWidth computes the rendered width of a single item.
func measureItemWidth(it *MenuItem) int {
	if it.Separator {
		return 3
	}
	w := 0
	if it.Icon != "" {
		w += len([]rune(it.Icon)) + 1
	}
	w += len([]rune(it.Label))
	if it.Shortcut != "" {
		w += len([]rune(it.Shortcut)) + 2
	}
	return w
}

// Paint renders the menu into the buffer at its anchor position.
func (cm *ContextMenu) Paint(buf *buffer.Buffer) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.visible || len(cm.items) == 0 {
		return
	}

	x, y := cm.x, cm.y
	w := cm.width
	if w < 10 {
		w = 10
	}
	h := len(cm.items)

	// Cell helper
	put := func(px, py int, r rune, s buffer.Style) {
		buf.SetCell(px, py, buffer.NewCell(r, s))
	}

	bs := cm.style.Border

	// Top border
	put(x, y, '┌', bs)
	for i := x + 1; i < x+w-1; i++ {
		put(i, y, '─', bs)
	}
	put(x+w-1, y, '┐', bs)

	// Content rows
	for row, it := range cm.items {
		py := y + 1 + row

		// Side borders
		put(x, py, '│', bs)
		put(x+w-1, py, '│', bs)

		if it.Separator {
			for i := x + 1; i < x+w-1; i++ {
				put(i, py, '─', cm.style.Separator)
			}
			continue
		}

		// Determine cell style
		st := cm.style.Normal
		if row == cm.cursor {
			st = cm.style.Selected
		}
		if !it.Enabled {
			st = cm.style.Disabled
		}

		px := x + 2 // 1-char padding inside left border

		// Icon
		if it.Icon != "" {
			for _, r := range it.Icon {
				put(px, py, r, st)
				px++
			}
			px++ // space after icon
		}

		// Label
		for _, r := range it.Label {
			put(px, py, r, st)
			px++
		}

		// Fill background for selected
		if row == cm.cursor {
			for i := px; i < x+w-1; i++ {
				put(i, py, ' ', st)
			}
		}

		// Shortcut (right-aligned, after fill so it stays visible)
		if it.Shortcut != "" {
			sx := x + w - 2 - len([]rune(it.Shortcut))
			if sx > x+1 {
				for _, r := range it.Shortcut {
					put(sx, py, r, cm.style.Shortcut)
					sx++
				}
			}
		}
	}

	// Bottom border
	botY := y + h + 1
	put(x, botY, '└', bs)
	for i := x + 1; i < x+w-1; i++ {
		put(i, botY, '─', bs)
	}
	put(x+w-1, botY, '┘', bs)
}

// Children returns visible submenus (for rendering), or nil.
func (cm *ContextMenu) Children() []Component {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	var children []Component
	if cm.cursor >= 0 && cm.cursor < len(cm.items) {
		if sub := cm.items[cm.cursor].Submenu; sub != nil && sub.Visible() {
			children = append(children, sub)
		}
	}
	return children
}

// ─── Mouse support ─────────────────────────────────────────────

// HitTest returns the item index at screen coordinates (mx, my), or -1 if outside.
func (cm *ContextMenu) HitTest(mx, my int) int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if !cm.visible || len(cm.items) == 0 {
		return -1
	}
	w := cm.width
	if w < 10 {
		w = 10
	}
	// Content is inside borders: x < mx < x+w, y < my <= y+h
	if mx <= cm.x || mx >= cm.x+w {
		return -1
	}
	if my <= cm.y || my > cm.y+len(cm.items) {
		return -1
	}
	idx := my - cm.y - 1
	if idx < 0 || idx >= len(cm.items) {
		return -1
	}
	return idx
}

// ClickAt handles a mouse click at (mx, my).
// Returns true if the click was inside the menu.
func (cm *ContextMenu) ClickAt(mx, my int) bool {
	idx := cm.HitTest(mx, my)
	if idx < 0 {
		return false
	}
	cm.mu.RLock()
	item := cm.items[idx]
	cm.mu.RUnlock()
	if item.Separator || !item.Enabled {
		return true // consumed but no action
	}
	cm.SetCursor(idx)
	cm.Activate()
	return true
}

// ─── Display helpers ───────────────────────────────────────────

// String returns a text representation of the menu.
func (cm *ContextMenu) String() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	var b strings.Builder
	b.WriteString("ContextMenu[")
	b.WriteString(cm.ID())
	b.WriteString("]{")
	for i, it := range cm.items {
		if i > 0 {
			b.WriteString(", ")
		}
		if it.Separator {
			b.WriteString("---")
		} else {
			b.WriteString(it.Label)
			if it.Shortcut != "" {
				b.WriteString("(")
				b.WriteString(it.Shortcut)
				b.WriteString(")")
			}
		}
	}
	b.WriteString("}")
	return b.String()
}

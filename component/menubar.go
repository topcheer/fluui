package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// MenuEntry represents a single selectable item in a dropdown menu.
type MenuEntry struct {
	ID       string // unique identifier
	Label    string // display text
	Shortcut string // optional shortcut hint (e.g. "Ctrl+S")
	Separator bool   // if true, renders as a separator line
	Disabled bool   // if true, item is greyed out and not selectable
	Submenu  []MenuEntry // optional submenu items
}

// Menu represents a top-level menu with a title and dropdown items.
type Menu struct {
	ID     string
	Title  string
	Items  []MenuEntry
}

// MenuBarStyle holds visual styles for the menu bar.
type MenuBarStyle struct {
	Bar       buffer.Style // bar background style
	MenuTitle buffer.Style // normal menu title style
	Active    buffer.Style // active/hovered menu title style
	Open      buffer.Style // open menu title style
	Dropdown  buffer.Style // dropdown background style
	Item      buffer.Style // normal item style
	ItemHover buffer.Style // hovered item style
	ItemDisabled buffer.Style // disabled item style
	Separator buffer.Style // separator line style
	Shortcut  buffer.Style // shortcut key hint style
}

// DefaultMenuBarStyle returns a Dracula-themed MenuBarStyle.
func DefaultMenuBarStyle() MenuBarStyle {
	white := buffer.RGB(0xf8, 0xf8, 0xf2)
	barBg := buffer.RGB(0x28, 0x2a, 0x36)
	dim := buffer.RGB(0x62, 0x72, 0xa4)
	purple := buffer.RGB(0xbd, 0x93, 0xf9)
	green := buffer.RGB(0x50, 0xfa, 0x7b)
	dropBg := buffer.RGB(0x38, 0x3a, 0x4a)

	return MenuBarStyle{
		Bar:          buffer.Style{Fg: white, Bg: barBg},
		MenuTitle:    buffer.Style{Fg: white, Bg: barBg},
		Active:       buffer.Style{Fg: purple, Bg: barBg, Flags: buffer.Bold},
		Open:         buffer.Style{Fg: white, Bg: purple},
		Dropdown:     buffer.Style{Fg: white, Bg: dropBg},
		Item:         buffer.Style{Fg: white, Bg: dropBg},
		ItemHover:    buffer.Style{Fg: white, Bg: purple},
		ItemDisabled: buffer.Style{Fg: dim, Bg: dropBg},
		Separator:    buffer.Style{Fg: dim, Bg: dropBg},
		Shortcut:     buffer.Style{Fg: green, Bg: dropBg},
	}
}

// MenuBar is a top-level menu bar with dropdown menus.
// It implements the Component interface and is safe for concurrent use.
type MenuBar struct {
	BaseComponent
	mu sync.RWMutex

	menus     []Menu
	activeIdx int    // which top-level menu is highlighted (-1 = none)
	openIdx   int    // which menu is open (-1 = none)
	itemIdx   int    // selected item within open menu (-1 = none)
	style     MenuBarStyle

	// OnAction is called when a menu item is selected.
	// Receives the menu ID and item ID.
	OnAction func(menuID, itemID string)

	// internal layout cache
	dropX     int   // x position of open dropdown
	dropW     int   // width of open dropdown
	dropH     int   // height of open dropdown
	menuXs    []int // x positions of each top-level menu title
}

// NewMenuBar creates a MenuBar with the given menus and default styling.
func NewMenuBar(menus []Menu) *MenuBar {
	return &MenuBar{
		menus:     menus,
		activeIdx: -1,
		openIdx:   -1,
		itemIdx:   -1,
		style:     DefaultMenuBarStyle(),
	}
}

// SetStyle sets the visual style.
func (m *MenuBar) SetStyle(s MenuBarStyle) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.style = s
}

// Style returns the current visual style.
func (m *MenuBar) Style() MenuBarStyle {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.style
}

// Menus returns a copy of the current menus.
func (m *MenuBar) Menus() []Menu {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Menu, len(m.menus))
	copy(out, m.menus)
	return out
}

// SetMenus replaces the menu list.
func (m *MenuBar) SetMenus(menus []Menu) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.menus = menus
	m.activeIdx = -1
	m.openIdx = -1
	m.itemIdx = -1
}

// IsOpen returns true if any dropdown menu is currently open.
func (m *MenuBar) IsOpen() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.openIdx >= 0
}

// OpenMenu opens the dropdown for the given menu index.
func (m *MenuBar) OpenMenu(idx int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if idx >= 0 && idx < len(m.menus) {
		m.openIdx = idx
		m.itemIdx = m.firstSelectableLocked(idx)
	}
}

// CloseMenu closes any open dropdown.
func (m *MenuBar) CloseMenu() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.openIdx = -1
	m.itemIdx = -1
}

// ActiveMenu returns the index of the currently highlighted top-level menu (-1 if none).
func (m *MenuBar) ActiveMenu() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeIdx
}

// SelectedItem returns the selected item index within the open menu (-1 if none).
func (m *MenuBar) SelectedItem() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.itemIdx
}

// HandleKey processes keyboard input.
// Returns true if the key was consumed.
func (m *MenuBar) HandleKey(k *term.KeyEvent) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.menus) == 0 {
		return false
	}

	key := k.Key
	r := k.Rune

	// Alt+letter: open menu whose title starts with that letter
	if k.Modifiers&term.ModAlt != 0 && r != 0 {
		upper := strings.ToUpper(string(r))
		for i, menu := range m.menus {
			if len(menu.Title) > 0 && strings.HasPrefix(strings.ToUpper(menu.Title), upper) {
				m.openIdx = i
				m.activeIdx = i
				m.itemIdx = m.firstSelectableLocked(i)
				return true
			}
		}
	}

	// If a menu is open, handle navigation within it
	if m.openIdx >= 0 {
		switch key {
		case term.KeyEscape:
			m.openIdx = -1
			m.itemIdx = -1
			return true
		case term.KeyLeft:
			// Move to previous top-level menu
			m.openIdx--
			if m.openIdx < 0 {
				m.openIdx = len(m.menus) - 1
			}
			m.activeIdx = m.openIdx
			m.itemIdx = m.firstSelectableLocked(m.openIdx)
			return true
		case term.KeyRight:
			// Move to next top-level menu
			m.openIdx++
			if m.openIdx >= len(m.menus) {
				m.openIdx = 0
			}
			m.activeIdx = m.openIdx
			m.itemIdx = m.firstSelectableLocked(m.openIdx)
			return true
		case term.KeyUp:
			// Move selection up, skipping separators and disabled
			m.itemIdx = m.prevSelectableLocked(m.openIdx, m.itemIdx)
			return true
		case term.KeyDown:
			// Move selection down, skipping separators and disabled
			m.itemIdx = m.nextSelectableLocked(m.openIdx, m.itemIdx)
			return true
		case term.KeyEnter:
			// Execute selected item
			if m.openIdx >= 0 && m.itemIdx >= 0 && m.itemIdx < len(m.menus[m.openIdx].Items) {
				item := m.menus[m.openIdx].Items[m.itemIdx]
				menuID := m.menus[m.openIdx].ID
				if !item.Disabled && !item.Separator {
					cb := m.OnAction
					m.openIdx = -1
					m.itemIdx = -1
					m.mu.Unlock()
					if cb != nil {
						cb(menuID, item.ID)
					}
					m.mu.Lock()
				}
			}
			return true
		}
	} else {
		// No menu open — handle menu bar navigation
		switch key {
		case term.KeyLeft:
			if m.activeIdx > 0 {
				m.activeIdx--
			} else {
				m.activeIdx = len(m.menus) - 1
			}
			return true
		case term.KeyRight:
			if m.activeIdx < len(m.menus)-1 {
				m.activeIdx++
			} else {
				m.activeIdx = 0
			}
			return true
		case term.KeyDown, term.KeyEnter:
			if m.activeIdx >= 0 {
				m.openIdx = m.activeIdx
				m.itemIdx = m.firstSelectableLocked(m.openIdx)
				return true
			}
		}
	}

	return false
}

// firstSelectableLocked returns the first non-separator, non-disabled item index.
func (m *MenuBar) firstSelectableLocked(menuIdx int) int {
	if menuIdx < 0 || menuIdx >= len(m.menus) {
		return -1
	}
	for i, item := range m.menus[menuIdx].Items {
		if !item.Separator && !item.Disabled {
			return i
		}
	}
	return -1
}

// nextSelectableLocked moves selection down, skipping separators/disabled.
func (m *MenuBar) nextSelectableLocked(menuIdx, from int) int {
	if menuIdx < 0 || menuIdx >= len(m.menus) {
		return -1
	}
	items := m.menus[menuIdx].Items
	n := len(items)
	if n == 0 {
		return -1
	}
	if from < 0 {
		from = -1 // will be incremented to 0 on first iteration
	}
	for i := 1; i <= n; i++ {
		idx := (from + i) % n
		if idx < 0 {
			idx += n
		}
		if !items[idx].Separator && !items[idx].Disabled {
			return idx
		}
	}
	return from
}

// prevSelectableLocked moves selection up, skipping separators/disabled.
func (m *MenuBar) prevSelectableLocked(menuIdx, from int) int {
	if menuIdx < 0 || menuIdx >= len(m.menus) {
		return -1
	}
	items := m.menus[menuIdx].Items
	n := len(items)
	if n == 0 {
		return -1
	}
	if from < 0 {
		from = 0
	}
	for i := 1; i <= n; i++ {
		idx := (from - i + n) % n
		if idx < 0 {
			idx += n
		}
		if !items[idx].Separator && !items[idx].Disabled {
			return idx
		}
	}
	return from
}

// Measure returns the desired size (always height=1 for the bar).
func (m *MenuBar) Measure(cs Constraints) Size {
	m.mu.RLock()
	defer m.mu.RUnlock()

	w := 0
	for _, menu := range m.menus {
		w += 2 + len(menu.Title) + 2 // padding + title + gap
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if w < 0 {
		w = 0
	}
	return Size{W: w, H: 1}
}

// SetBounds sets the position and size.
func (m *MenuBar) SetBounds(r Rect) {
	m.mu.Lock()
	m.BaseComponent.SetBounds(r)
	// Compute menu title positions
	m.menuXs = make([]int, len(m.menus))
	x := 0
	for i, menu := range m.menus {
		m.menuXs[i] = x
		x += 2 + len(menu.Title) + 1 // left-pad(1) + title + right-pad(1) + gap
		if x > 0 {
			x-- // compact gap
		}
	}
	// Compute dropdown dimensions for open menu
	if m.openIdx >= 0 && m.openIdx < len(m.menus) {
		m.computeDropDimsLocked()
	}
	m.mu.Unlock()
}

// computeDropDimsLocked computes dropdown width and height.
func (m *MenuBar) computeDropDimsLocked() {
	if m.openIdx < 0 || m.openIdx >= len(m.menus) {
		m.dropW = 0
		m.dropH = 0
		return
	}
	menu := m.menus[m.openIdx]
	maxW := 0
	for _, item := range menu.Items {
		w := 4 + len(item.Label) // padding + label
		if item.Shortcut != "" {
			w += 2 + len(item.Shortcut) // gap + shortcut
		}
		if w > maxW {
			maxW = w
		}
	}
	if maxW < 12 {
		maxW = 12
	}
	m.dropW = maxW
	m.dropH = len(menu.Items)
	// Position dropdown below the active menu title
	if m.openIdx < len(m.menuXs) {
		m.dropX = m.menuXs[m.openIdx]
	} else {
		m.dropX = 0
	}
}

// Paint renders the menu bar into the buffer.
func (m *MenuBar) Paint(buf *buffer.Buffer) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bounds := m.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	// Fill bar background
	for x := 0; x < bounds.W; x++ {
		buf.SetCell(bounds.X+x, bounds.Y, buffer.Cell{
			Rune:  ' ',
			Width: 1,
			Fg:    m.style.Bar.Fg,
			Bg:    m.style.Bar.Bg,
		})
	}

	// Draw menu titles
	x := bounds.X
	for i, menu := range m.menus {
		title := " " + menu.Title + " "
		titleW := len(title)

		var style buffer.Style
		if m.openIdx == i {
			style = m.style.Open
		} else if m.activeIdx == i {
			style = m.style.Active
		} else {
			style = m.style.MenuTitle
		}

		for j, ch := range title {
			if x+j < bounds.X+bounds.W {
				buf.SetCell(x+j, bounds.Y, buffer.Cell{
					Rune:  ch,
					Width: 1,
					Fg:    style.Fg,
					Bg:    style.Bg,
					Flags: style.Flags,
				})
			}
		}
		x += titleW + 1 // +1 gap between menus
	}

	// Draw dropdown if a menu is open
	if m.openIdx >= 0 && m.openIdx < len(m.menus) {
		m.paintDropdownLocked(buf, bounds)
	}
}

// paintDropdownLocked renders the dropdown menu.
func (m *MenuBar) paintDropdownLocked(buf *buffer.Buffer, bounds Rect) {
	menu := m.menus[m.openIdx]
	dropX := bounds.X + m.dropX
	dropY := bounds.Y + 1

	for i, item := range menu.Items {
		y := dropY + i
		if y >= bounds.Y+bounds.H+8 { // allow overflow but cap
			break
		}

		if item.Separator {
			// Draw separator line
			for dx := 0; dx < m.dropW; dx++ {
				x := dropX + dx
				if x >= 0 && x < buf.Width && y >= 0 && y < buf.Height {
					buf.SetCell(x, y, buffer.Cell{
						Rune:  '─',
						Width: 1,
						Fg:    m.style.Separator.Fg,
						Bg:    m.style.Separator.Bg,
					})
				}
			}
			continue
		}

		// Determine item style
		var style buffer.Style
		if i == m.itemIdx {
			style = m.style.ItemHover
		} else if item.Disabled {
			style = m.style.ItemDisabled
		} else {
			style = m.style.Item
		}

		// Draw label with padding
		label := "  " + item.Label
		for j, ch := range label {
			x := dropX + j
			if x >= 0 && x < buf.Width && y >= 0 && y < buf.Height {
				buf.SetCell(x, y, buffer.Cell{
					Rune:  ch,
					Width: 1,
					Fg:    style.Fg,
					Bg:    style.Bg,
					Flags: style.Flags,
				})
			}
		}

		// Pad remaining width with background
		for j := len(label); j < m.dropW; j++ {
			x := dropX + j
			if x >= 0 && x < buf.Width && y >= 0 && y < buf.Height {
				buf.SetCell(x, y, buffer.Cell{
					Rune:  ' ',
					Width: 1,
					Fg:    style.Fg,
					Bg:    style.Bg,
				})
			}
		}

		// Draw shortcut on right side
		if item.Shortcut != "" {
			scX := dropX + m.dropW - len(item.Shortcut) - 1
			for j, ch := range item.Shortcut {
				x := scX + j
				if x >= 0 && x < buf.Width && y >= 0 && y < buf.Height {
					buf.SetCell(x, y, buffer.Cell{
						Rune:  ch,
						Width: 1,
						Fg:    m.style.Shortcut.Fg,
						Bg:    style.Bg,
					})
				}
			}
		}
	}
}

// HandleMouse processes a mouse click. Returns true if consumed.
func (m *MenuBar) HandleMouse(x, y int, action term.MouseAction) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	bounds := m.Bounds()

	// Click in menu bar area
	if y == bounds.Y && x >= bounds.X && x < bounds.X+bounds.W {
		// Find which menu was clicked
		cx := bounds.X
		for i, menu := range m.menus {
			titleW := 2 + len(menu.Title) // " title "
			if x >= cx && x < cx+titleW {
				if action == term.MouseDown {
					if m.openIdx == i {
						m.openIdx = -1
						m.itemIdx = -1
					} else {
						m.openIdx = i
						m.activeIdx = i
						m.itemIdx = m.firstSelectableLocked(i)
					}
				} else {
					m.activeIdx = i
				}
				return true
			}
			cx += titleW + 1
		}
		// Click on empty area — close menu
		m.openIdx = -1
		m.itemIdx = -1
		return true
	}

	// Click in dropdown area
	if m.openIdx >= 0 {
		m.computeDropDimsLocked()
		dropX := bounds.X + m.dropX
		dropY := bounds.Y + 1
		if x >= dropX && x < dropX+m.dropW && y >= dropY && y < dropY+m.dropH {
			itemIdx := y - dropY
			if itemIdx >= 0 && itemIdx < len(m.menus[m.openIdx].Items) {
				item := m.menus[m.openIdx].Items[itemIdx]
				if action == term.MouseDown && !item.Separator && !item.Disabled {
					menuID := m.menus[m.openIdx].ID
					cb := m.OnAction
					m.openIdx = -1
					m.itemIdx = -1
					m.mu.Unlock()
					if cb != nil {
						cb(menuID, item.ID)
					}
					m.mu.Lock()
				} else if !item.Separator {
					m.itemIdx = itemIdx
				}
				return true
			}
		}
	}

	// Click outside — close menu
	if action == term.MouseDown && m.openIdx >= 0 {
		m.openIdx = -1
		m.itemIdx = -1
		return true
	}

	return false
}

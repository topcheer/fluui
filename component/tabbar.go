package component

import (
	"fmt"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// Tab represents a single tab in a TabBar.
type Tab struct {
	ID      string
	Title   string
	Closable bool
	Active  bool
}

// TabBarStyle defines visual styling for the tab bar.
type TabBarStyle struct {
	Normal     buffer.Style
	Active     buffer.Style
	Hover      buffer.Style
	CloseBtn   buffer.Style
	Separator  buffer.Style
	Background buffer.Style
}

// DefaultTabBarStyle returns a Dracula-inspired tab bar style.
func DefaultTabBarStyle() TabBarStyle {
	fg := buffer.RGB(248, 248, 242)
	dim := buffer.RGB(98, 114, 164)
	active := buffer.RGB(189, 147, 249)   // purple
	closeFg := buffer.RGB(255, 85, 85)     // red
	return TabBarStyle{
		Normal:     buffer.Style{Fg: dim},
		Active:     buffer.Style{Fg: active, Flags: buffer.Bold},
		Hover:      buffer.Style{Fg: fg},
		CloseBtn:   buffer.Style{Fg: closeFg},
		Separator:  buffer.Style{Fg: dim},
		Background: buffer.Style{Fg: fg, Bg: buffer.RGB(40, 42, 54)},
	}
}

// TabBar is a horizontal tab bar component supporting add/close/switch.
type TabBar struct {
	BaseComponent
	mu       sync.RWMutex
	tabs     []Tab
	active   int
	style    TabBarStyle
	hoverIdx int
	showNew  bool
	maxTitle int
}

// NewTabBar creates a TabBar with default styling.
func NewTabBar() *TabBar {
	return &TabBar{
		style:    DefaultTabBarStyle(),
		hoverIdx: -1,
		showNew:  true,
		maxTitle: 20,
	}
}

// --- Tab management ---

// AddTab appends a new tab and returns its index.
func (tb *TabBar) AddTab(id, title string) int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.tabs = append(tb.tabs, Tab{ID: id, Title: title, Closable: true})
	if len(tb.tabs) == 1 {
		tb.active = 0
	}
	return len(tb.tabs) - 1
}

// InsertTab inserts a tab at the given index.
func (tb *TabBar) InsertTab(idx int, id, title string) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if idx < 0 {
		idx = 0
	}
	if idx > len(tb.tabs) {
		idx = len(tb.tabs)
	}
	tb.tabs = append(tb.tabs, Tab{})
	copy(tb.tabs[idx+1:], tb.tabs[idx:])
	tb.tabs[idx] = Tab{ID: id, Title: title, Closable: true}
	if tb.active >= idx {
		tb.active++
	}
}

// RemoveTab removes the tab with the given ID.
func (tb *TabBar) RemoveTab(id string) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	idx := tb.indexOfLocked(id)
	if idx < 0 {
		return
	}
	tb.tabs = append(tb.tabs[:idx], tb.tabs[idx+1:]...)
	if len(tb.tabs) == 0 {
		tb.active = 0
		return
	}
	if tb.active >= len(tb.tabs) {
		tb.active = len(tb.tabs) - 1
	} else if tb.active > idx {
		tb.active--
	}
}

// CloseActive closes the active tab.
func (tb *TabBar) CloseActive() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if len(tb.tabs) == 0 {
		return
	}
	idx := tb.active
	tb.tabs = append(tb.tabs[:idx], tb.tabs[idx+1:]...)
	if len(tb.tabs) == 0 {
		tb.active = 0
		return
	}
	if tb.active >= len(tb.tabs) {
		tb.active = len(tb.tabs) - 1
	}
}

// Tabs returns a copy of the tab list.
func (tb *TabBar) Tabs() []Tab {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	out := make([]Tab, len(tb.tabs))
	copy(out, tb.tabs)
	return out
}

// TabCount returns the number of tabs.
func (tb *TabBar) TabCount() int {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return len(tb.tabs)
}

// TabAt returns the tab at the given index, or nil if out of range.
func (tb *TabBar) TabAt(idx int) *Tab {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	if idx < 0 || idx >= len(tb.tabs) {
		return nil
	}
	return &tb.tabs[idx]
}

// FindTab returns the tab with the given ID, or nil.
func (tb *TabBar) FindTab(id string) *Tab {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	idx := tb.indexOfLocked(id)
	if idx < 0 {
		return nil
	}
	return &tb.tabs[idx]
}

// --- Active tab ---

// ActiveIndex returns the index of the active tab.
func (tb *TabBar) ActiveIndex() int {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.active
}

// SetActive sets the active tab by index.
func (tb *TabBar) SetActive(idx int) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if idx < 0 || idx >= len(tb.tabs) {
		return
	}
	tb.active = idx
}

// SetActiveByID sets the active tab by ID.
func (tb *TabBar) SetActiveByID(id string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	idx := tb.indexOfLocked(id)
	if idx < 0 {
		return false
	}
	tb.active = idx
	return true
}

// NextTab switches to the next tab (wraps around).
func (tb *TabBar) NextTab() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if len(tb.tabs) == 0 {
		return
	}
	tb.active = (tb.active + 1) % len(tb.tabs)
}

// PrevTab switches to the previous tab (wraps around).
func (tb *TabBar) PrevTab() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if len(tb.tabs) == 0 {
		return
	}
	tb.active = (tb.active - 1 + len(tb.tabs)) % len(tb.tabs)
}

// --- Configuration ---

// SetStyle sets the tab bar style.
func (tb *TabBar) SetStyle(s TabBarStyle) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.style = s
}

// Style returns the current tab bar style.
func (tb *TabBar) Style() TabBarStyle {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.style
}

// SetShowNewButton toggles the "+" new-tab button.
func (tb *TabBar) SetShowNewButton(v bool) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.showNew = v
}

// ShowNewButton returns whether the new-tab button is shown.
func (tb *TabBar) ShowNewButton() bool {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.showNew
}

// SetMaxTitleWidth sets the maximum visible width of tab titles.
func (tb *TabBar) SetMaxTitleWidth(w int) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.maxTitle = w
}

// MaxTitleWidth returns the maximum title width.
func (tb *TabBar) MaxTitleWidth() int {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.maxTitle
}

// --- Component interface ---

// Measure returns the preferred size for the tab bar.
func (tb *TabBar) Measure(c Constraints) Size {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	w := 0
	for _, t := range tb.tabs {
		w += tb.tabWidthLocked(t) + 1 // +1 for separator
	}
	if tb.showNew {
		w += 3 // " + "
	}
	if w < c.MaxWidth && c.MaxWidth > 0 {
		w = c.MaxWidth
	}
	return Size{W: w, H: 1}
}

// SetBounds sets the component bounds.
func (tb *TabBar) SetBounds(b Rect) {
	tb.mu.Lock()
	tb.BaseComponent.SetBounds(b)
	tb.mu.Unlock()
}

// Bounds returns the component bounds.
func (tb *TabBar) Bounds() Rect {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.BaseComponent.Bounds()
}

// Paint renders the tab bar to the buffer.
func (tb *TabBar) Paint(buf *buffer.Buffer) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	b := tb.BaseComponent.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	x := b.X
	for i, tab := range tb.tabs {
		tw := tb.tabWidthLocked(tab)
		style := tb.style.Normal
		if i == tb.active {
			style = tb.style.Active
		} else if i == tb.hoverIdx {
			style = tb.style.Hover
		}

		// Draw title
		title := truncateTabTitle(tab.Title, tb.maxTitle)
		for j, r := range title {
			if x+j >= b.X+b.W {
				break
			}
			buf.SetCell(x+j, b.Y, buffer.NewCell(r, style))
		}
		x += tw

		// Draw close button
		if tab.Closable {
			if x < b.X+b.W {
				buf.SetCell(x, b.Y, buffer.NewCell(' ', style))
				x++
			}
			if x < b.X+b.W {
				buf.SetCell(x, b.Y, buffer.NewCell('x', tb.style.CloseBtn))
				x++
			}
		}

		// Draw separator
		if i < len(tb.tabs)-1 && x < b.X+b.W {
			buf.SetCell(x, b.Y, buffer.NewCell('│', tb.style.Separator))
			x++
		}
	}

	// Draw new-tab button
	if tb.showNew && x < b.X+b.W {
		buf.SetCell(x, b.Y, buffer.NewCell(' ', tb.style.Normal))
		x++
		if x < b.X+b.W {
			buf.SetCell(x, b.Y, buffer.NewCell('+', tb.style.Active))
		}
	}
}

// Children returns nil (tab bar has no child components).
func (tb *TabBar) Children() []Component {
	return nil
}

// --- Hit testing ---

// HitTest returns the tab index at the given coordinates, or -1.
func (tb *TabBar) HitTest(x, y int) int {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	b := tb.BaseComponent.Bounds()
	if y < b.Y || y >= b.Y+b.H || x < b.X {
		return -1
	}
	cx := b.X
	for i, tab := range tb.tabs {
		tw := tb.tabWidthLocked(tab)
		if x >= cx && x < cx+tw {
			return i
		}
		cx += tw
		// separator
		if i < len(tb.tabs)-1 {
			if x == cx {
				return -1
			}
			cx++
		}
	}
	// New button
	if tb.showNew && x >= cx && x < cx+2 {
		return -2 // special: new button
	}
	return -1
}

// IsCloseButton tests if coordinates are on a tab's close button.
func (tb *TabBar) IsCloseButton(x, y int) (tabIdx int, ok bool) {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	b := tb.BaseComponent.Bounds()
	if y < b.Y || y >= b.Y+b.H || x < b.X {
		return -1, false
	}
	cx := b.X
	for i, tab := range tb.tabs {
		tw := tb.tabWidthLocked(tab)
		if !tab.Closable {
			cx += tw
			if i < len(tb.tabs)-1 {
				cx++
			}
			continue
		}
		// Close button is at the 'x' position (tw-1)
		closeX := cx + tw - 1
		if x == closeX && y == b.Y {
			return i, true
		}
		cx += tw
		if i < len(tb.tabs)-1 {
			cx++
		}
	}
	return -1, false
}

// --- Helpers ---

func (tb *TabBar) tabWidthLocked(tab Tab) int {
	title := truncateTabTitle(tab.Title, tb.maxTitle)
	w := len([]rune(title))
	if tab.Closable {
		w += 2 // space + x
	}
	return w
}

func (tb *TabBar) indexOfLocked(id string) int {
	for i, t := range tb.tabs {
		if t.ID == id {
			return i
		}
	}
	return -1
}

func truncateTabTitle(title string, maxW int) string {
	if maxW <= 0 {
		return title
	}
	runes := []rune(title)
	if len(runes) <= maxW {
		return title
	}
	return string(runes[:maxW-1]) + "…"
}

// String returns a debug representation.
func (tb *TabBar) String() string {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	titles := make([]string, len(tb.tabs))
	for i, t := range tb.tabs {
		if i == tb.active {
			titles[i] = "[" + t.Title + "]"
		} else {
			titles[i] = t.Title
		}
	}
	return fmt.Sprintf("TabBar{tabs:%d active:%d [%s]}", len(tb.tabs), tb.active, strings.Join(titles, " "))
}

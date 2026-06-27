package component

import (
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
)

// truncateRunes truncates s to at most n runes.
func truncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

// ─── Data Types ─────────────────────────────────────────────

// HelpEntry represents a single keyboard shortcut displayed in the help overlay.
type HelpEntry struct {
	Keys        string // e.g. "Ctrl+F", "g g", "Ctrl+P"
	Description string // e.g. "Search in conversation"
}

// HelpGroup represents a category of shortcuts (Navigation, Edit, Search, etc.)
type HelpGroup struct {
	Name    string
	Entries []HelpEntry
}

// ─── HelpOverlay ────────────────────────────────────────────

// HelpOverlay is a popup component that displays a searchable keyboard
// shortcut reference. It groups shortcuts by category and allows filtering
// by a search query. It implements the Component interface and is designed
// to be used as an overlay (triggered by '?', dismissed by Esc).
type HelpOverlay struct {
	BaseComponent
	mu sync.RWMutex

	groups   []HelpGroup
	query    string
	cursor   int // cursor position in query string
	scrollY  int // vertical scroll offset (in lines)
	selected int // highlighted row index (for keyboard navigation)

	// Visual configuration
	title     string
	maxWidth  int
	maxHeight int

	// style configuration
	style HelpStyle

	// filtering cache (always pre-computed; safe for RLock readers)
	filteredGroups []HelpGroup
}

// HelpStyle holds the visual styling for the HelpOverlay.
type HelpStyle struct {
	Border       buffer.Style // border frame
	Title        buffer.Style // title text
	GroupHeader  buffer.Style // group header text
	Key          buffer.Style // shortcut key text
	Description  buffer.Style // description text
	SearchPrompt buffer.Style // "/" search prompt
	Highlight    buffer.Style // highlighted/selected row
	Normal       buffer.Style // normal background
}

// DefaultHelpStyle returns a reasonable default style with Dracula-like colors.
func DefaultHelpStyle() HelpStyle {
	normal := buffer.DefaultStyle
	return HelpStyle{
		Border:       normal,
		Title:        normal.AddFlags(buffer.Bold),
		GroupHeader:  normal.AddFlags(buffer.Bold).WithFg(buffer.RGB(139, 233, 253)), // cyan
		Key:          normal.WithFg(buffer.RGB(241, 250, 140)),                        // yellow
		Description:  normal.WithFg(buffer.RGB(248, 248, 242)),                        // white
		SearchPrompt: normal.WithFg(buffer.RGB(139, 233, 253)),                        // cyan
		Highlight:    normal.AddFlags(buffer.Reverse),
		Normal:       normal,
	}
}

// NewHelpOverlay creates a HelpOverlay with the given groups.
func NewHelpOverlay(groups []HelpGroup) *HelpOverlay {
	h := &HelpOverlay{
		groups:   groups,
		title:    " Keyboard Shortcuts ",
		maxWidth: 70,
		maxHeight: 30,
		style:    DefaultHelpStyle(),
	}
	h.SetID(GenerateID("help"))
	h.computeFilteredLocked()
	return h
}

// ─── Public API ─────────────────────────────────────────────

// SetGroups replaces the shortcut groups.
func (h *HelpOverlay) SetGroups(groups []HelpGroup) {
	h.mu.Lock()
	h.groups = groups
	h.scrollY = 0
	h.selected = 0
	h.computeFilteredLocked()
	h.mu.Unlock()
}

// Groups returns a copy of the shortcut groups.
func (h *HelpOverlay) Groups() []HelpGroup {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]HelpGroup, len(h.groups))
	copy(result, h.groups)
	return result
}

// SetQuery sets the search filter query.
func (h *HelpOverlay) SetQuery(q string) {
	h.mu.Lock()
	h.query = q
	h.cursor = len([]rune(q))
	h.scrollY = 0
	h.selected = 0
	h.computeFilteredLocked()
	h.mu.Unlock()
}

// Query returns the current search query.
func (h *HelpOverlay) Query() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.query
}

// AppendQuery appends text to the search query.
func (h *HelpOverlay) AppendQuery(s string) {
	h.mu.Lock()
	runes := []rune(h.query)
	runes = append(runes, []rune(s)...)
	h.query = string(runes)
	h.cursor = len(runes)
	h.scrollY = 0
	h.selected = 0
	h.computeFilteredLocked()
	h.mu.Unlock()
}

// BackspaceQuery removes the last rune from the query.
func (h *HelpOverlay) BackspaceQuery() bool {
	h.mu.Lock()
	runes := []rune(h.query)
	if len(runes) == 0 {
		h.mu.Unlock()
		return false
	}
	h.query = string(runes[:len(runes)-1])
	h.cursor = len([]rune(h.query))
	h.scrollY = 0
	h.selected = 0
	h.computeFilteredLocked()
	h.mu.Unlock()
	return true
}

// ClearQuery clears the search filter.
func (h *HelpOverlay) ClearQuery() {
	h.mu.Lock()
	h.query = ""
	h.cursor = 0
	h.scrollY = 0
	h.selected = 0
	h.computeFilteredLocked()
	h.mu.Unlock()
}

// ScrollUp moves the viewport up by n lines.
func (h *HelpOverlay) ScrollUp(n int) {
	h.mu.Lock()
	h.scrollY -= n
	if h.scrollY < 0 {
		h.scrollY = 0
	}
	h.mu.Unlock()
}

// ScrollDown moves the viewport down by n lines.
func (h *HelpOverlay) ScrollDown(n int) {
	h.mu.Lock()
	h.scrollY += n
	h.ensureScrollValidLocked()
	h.mu.Unlock()
}

// ScrollY returns the current scroll offset.
func (h *HelpOverlay) ScrollY() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.scrollY
}

// SelectedIndex returns the currently highlighted row index.
func (h *HelpOverlay) SelectedIndex() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.selected
}

// SetSelected sets the highlighted row index.
func (h *HelpOverlay) SetSelected(idx int) {
	h.mu.Lock()
	h.selected = idx
	if h.selected < 0 {
		h.selected = 0
	}
	h.ensureSelectedVisibleLocked()
	h.mu.Unlock()
}

// SelectNext moves the selection down by one.
func (h *HelpOverlay) SelectNext() {
	h.mu.Lock()
	h.selected++
	h.ensureSelectedValidLocked()
	h.ensureSelectedVisibleLocked()
	h.mu.Unlock()
}

// SelectPrev moves the selection up by one.
func (h *HelpOverlay) SelectPrev() {
	h.mu.Lock()
	h.selected--
	if h.selected < 0 {
		h.selected = 0
	}
	h.ensureSelectedVisibleLocked()
	h.mu.Unlock()
}

// SetTitle sets the overlay title.
func (h *HelpOverlay) SetTitle(title string) {
	h.mu.Lock()
	h.title = title
	h.mu.Unlock()
}

// Title returns the overlay title.
func (h *HelpOverlay) Title() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.title
}

// SetStyle overrides the default visual style.
func (h *HelpOverlay) SetStyle(s HelpStyle) {
	h.mu.Lock()
	h.style = s
	h.mu.Unlock()
}

// Style returns the current style.
func (h *HelpOverlay) Style() HelpStyle {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.style
}

// SetMaxWidth sets the maximum width constraint.
func (h *HelpOverlay) SetMaxWidth(w int) {
	h.mu.Lock()
	h.maxWidth = w
	h.mu.Unlock()
}

// SetMaxHeight sets the maximum height constraint.
func (h *HelpOverlay) SetMaxHeight(h2 int) {
	h.mu.Lock()
	h.maxHeight = h2
	h.mu.Unlock()
}

// TotalRows returns the total number of renderable rows (group headers + entries)
// after filtering. Each group contributes 1 header + N entries.
func (h *HelpOverlay) TotalRows() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.totalRowsLocked()
}

// totalRowsLocked computes total rows. Caller must hold at least RLock.
func (h *HelpOverlay) totalRowsLocked() int {
	total := 0
	for _, g := range h.filteredGroups {
		total++ // group header
		total += len(g.Entries)
	}
	return total
}

// FilteredGroups returns a copy of the filtered groups based on the current query.
func (h *HelpOverlay) FilteredGroups() []HelpGroup {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]HelpGroup, len(h.filteredGroups))
	copy(result, h.filteredGroups)
	return result
}

// HasResults returns true if the current filter has any matching entries.
func (h *HelpOverlay) HasResults() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, g := range h.filteredGroups {
		if len(g.Entries) > 0 {
			return true
		}
	}
	return false
}

// ─── Component Interface ────────────────────────────────────

// Measure computes the desired size of the help overlay.
func (h *HelpOverlay) Measure(cs Constraints) Size {
	h.mu.RLock()
	defer h.mu.RUnlock()

	width := h.maxWidth
	height := h.maxHeight

	if cs.MaxWidth > 0 && width > cs.MaxWidth {
		width = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && height > cs.MaxHeight {
		height = cs.MaxHeight
	}

	if width < 20 {
		width = 20
	}
	if height < 5 {
		height = 5
	}

	return Size{W: width, H: height}
}

// SetBounds sets the component's bounds and recalculates scroll.
func (h *HelpOverlay) SetBounds(r Rect) {
	h.mu.Lock()
	h.BaseComponent.SetBounds(r)
	h.ensureScrollValidLocked()
	h.ensureSelectedVisibleLocked()
	h.mu.Unlock()
}

// Paint renders the help overlay into the buffer.
func (h *HelpOverlay) Paint(buf *buffer.Buffer) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	bounds := h.BaseComponent.Bounds()
	if bounds.W < 3 || bounds.H < 3 {
		return
	}

	// Draw border
	h.drawBorderLocked(buf, bounds)

	// Content area (inside border)
	contentX := bounds.X + 1
	contentY := bounds.Y + 1
	contentW := bounds.W - 2
	contentH := bounds.H - 2

	// Row 0: search bar (if contentH > 0)
	row := 0
	if contentH > 0 {
		h.drawSearchBarLocked(buf, contentX, contentY, contentW)
		row = 1
	}

	// Rows 1..contentH-1: group headers and entries
	if contentH <= 1 {
		return
	}

	visibleH := contentH - 1 // subtract search bar
	maxKeyW := h.maxKeyWidthLocked()
	if maxKeyW > contentW/2 {
		maxKeyW = contentW / 2
	}

	drawn := 0
	for _, g := range h.filteredGroups {
		if drawn >= visibleH {
			break
		}

		// Group header
		if row >= h.scrollY && drawn < visibleH {
			y := contentY + 1 + drawn
			if y < bounds.Y+bounds.H-1 {
				headerText := truncateRunes(g.Name, contentW)
				buf.DrawTextClamped(contentX, y, headerText, h.style.GroupHeader)
			}
		}
		drawn++

		// Entries
		for _, entry := range g.Entries {
			if drawn >= visibleH {
				break
			}
			if row+1 >= h.scrollY && drawn < visibleH {
				y := contentY + 1 + drawn
				if y < bounds.Y+bounds.H-1 {
					h.drawEntryLocked(buf, contentX, y, contentW, maxKeyW, entry)
				}
			}
			drawn++
		}
	}

	// If no results, show a message
	if !h.hasResultsLocked() && contentH > 1 {
		msg := "No matching shortcuts found"
		y := contentY + 2
		if y < bounds.Y+bounds.H-1 {
			buf.DrawTextClamped(contentX, y, msg, h.style.Description)
		}
	}
}

// Children returns nil (leaf component).
func (h *HelpOverlay) Children() []Component {
	return nil
}

// ─── Internal helpers (all *Locked — caller must hold lock) ─

// computeFilteredLocked recomputes the filtered groups from scratch.
// Caller MUST hold the write lock (h.mu.Lock).
func (h *HelpOverlay) computeFilteredLocked() {
	if h.query == "" {
		h.filteredGroups = make([]HelpGroup, len(h.groups))
		copy(h.filteredGroups, h.groups)
		return
	}

	q := strings.ToLower(h.query)
	h.filteredGroups = h.filteredGroups[:0]
	for _, g := range h.groups {
		var matched []HelpEntry
		for _, e := range g.Entries {
			if strings.Contains(strings.ToLower(e.Keys), q) ||
				strings.Contains(strings.ToLower(e.Description), q) ||
				strings.Contains(strings.ToLower(g.Name), q) {
				matched = append(matched, e)
			}
		}
		if len(matched) > 0 {
			h.filteredGroups = append(h.filteredGroups, HelpGroup{
				Name:    g.Name,
				Entries: matched,
			})
		}
	}
}

// hasResultsLocked checks if filtered groups have entries. Caller must hold lock.
func (h *HelpOverlay) hasResultsLocked() bool {
	for _, g := range h.filteredGroups {
		if len(g.Entries) > 0 {
			return true
		}
	}
	return false
}

// maxKeyWidthLocked returns the width of the longest key string.
func (h *HelpOverlay) maxKeyWidthLocked() int {
	maxW := 0
	for _, g := range h.filteredGroups {
		for _, e := range g.Entries {
			w := utf8.RuneCountInString(e.Keys)
			if w > maxW {
				maxW = w
			}
		}
	}
	return maxW
}

// ensureSelectedValidLocked clamps the selected index to valid range.
// Caller must hold the write lock.
func (h *HelpOverlay) ensureSelectedValidLocked() {
	total := h.totalRowsLocked()
	if h.selected >= total {
		h.selected = total - 1
	}
	if h.selected < 0 {
		h.selected = 0
	}
}

// ensureSelectedVisibleLocked adjusts scrollY to keep the selected row visible.
// Caller must hold the write lock.
func (h *HelpOverlay) ensureSelectedVisibleLocked() {
	bounds := h.BaseComponent.Bounds()
	visibleH := bounds.H - 3 // border(1) + search(1) + border(1)
	if visibleH <= 0 {
		return
	}
	if h.selected < h.scrollY {
		h.scrollY = h.selected
	}
	if h.selected >= h.scrollY+visibleH {
		h.scrollY = h.selected - visibleH + 1
	}
}

// ensureScrollValidLocked clamps scrollY to valid range.
// Caller must hold the write lock.
func (h *HelpOverlay) ensureScrollValidLocked() {
	if h.scrollY < 0 {
		h.scrollY = 0
	}
	bounds := h.BaseComponent.Bounds()
	visibleH := bounds.H - 3
	if visibleH <= 0 {
		h.scrollY = 0
		return
	}
	total := h.totalRowsLocked()
	if h.scrollY > total-visibleH && total > visibleH {
		h.scrollY = total - visibleH
	}
}

// drawBorderLocked draws the Unicode border frame with title.
// Caller must hold at least RLock.
func (h *HelpOverlay) drawBorderLocked(buf *buffer.Buffer, bounds Rect) {
	x, y, w, hh := bounds.X, bounds.Y, bounds.W, bounds.H

	// Top border with title
	buf.SetCell(x, y, buffer.NewCell('┌', h.style.Border))
	buf.SetCell(x+w-1, y, buffer.NewCell('┐', h.style.Border))
	for i := 1; i < w-1; i++ {
		buf.SetCell(x+i, y, buffer.NewCell('─', h.style.Border))
	}
	// Title on top border
	titleRunes := []rune(h.title)
	titleStart := x + 2
	for i, r := range titleRunes {
		if titleStart+i < x+w-1 {
			buf.SetCell(titleStart+i, y, buffer.NewCell(r, h.style.Title))
		}
	}

	// Side borders
	for i := 1; i < hh-1; i++ {
		buf.SetCell(x, y+i, buffer.NewCell('│', h.style.Border))
		buf.SetCell(x+w-1, y+i, buffer.NewCell('│', h.style.Border))
	}

	// Bottom border
	buf.SetCell(x, y+hh-1, buffer.NewCell('└', h.style.Border))
	buf.SetCell(x+w-1, y+hh-1, buffer.NewCell('┘', h.style.Border))
	for i := 1; i < w-1; i++ {
		buf.SetCell(x+i, y+hh-1, buffer.NewCell('─', h.style.Border))
	}

	// Scroll indicator on bottom-right
	total := h.totalRowsLocked()
	visibleH := hh - 3
	if total > visibleH && visibleH > 0 {
		sx := x + w - 2
		r := '↓'
		buf.SetCell(sx, y+hh-1, buffer.NewCell(r, h.style.Border))
	}
}

// drawSearchBarLocked draws the search input row.
// Caller must hold at least RLock.
func (h *HelpOverlay) drawSearchBarLocked(buf *buffer.Buffer, x, y, w int) {
	prompt := "/ "
	promptRunes := []rune(prompt)
	for i, r := range promptRunes {
		if x+i < x+w {
			buf.SetCell(x+i, y, buffer.NewCell(r, h.style.SearchPrompt))
		}
	}

	queryStart := x + len(promptRunes)
	queryRunes := []rune(h.query)
	for i, r := range queryRunes {
		if queryStart+i < x+w-1 {
			buf.SetCell(queryStart+i, y, buffer.NewCell(r, h.style.Normal))
		}
	}

	// Cursor indicator (block character)
	if queryStart+len(queryRunes) < x+w-1 {
		buf.SetCell(queryStart+len(queryRunes), y, buffer.NewCell('▏', h.style.Highlight))
	}

	// Hint on the right side
	if w > 30 {
		hint := "ESC to close"
		hintX := x + w - utf8.RuneCountInString(hint) - 1
		if hintX > queryStart+len(queryRunes)+1 {
			hr := []rune(hint)
			for i, r := range hr {
				buf.SetCell(hintX+i, y, buffer.NewCell(r, h.style.Description))
			}
		}
	}
}

// drawEntryLocked draws a single shortcut entry (key + description).
// Caller must hold at least RLock.
func (h *HelpOverlay) drawEntryLocked(buf *buffer.Buffer, x, y, w, maxKeyW int, entry HelpEntry) {
	keyRunes := []rune(entry.Keys)
	// Pad key to maxKeyW + 2 spaces
	for i := 0; i < maxKeyW+2 && x+i < x+w; i++ {
		if i < len(keyRunes) {
			buf.SetCell(x+i, y, buffer.NewCell(keyRunes[i], h.style.Key))
		} else {
			buf.SetCell(x+i, y, buffer.NewCell(' ', h.style.Normal))
		}
	}

	// Description
	descX := x + maxKeyW + 2
	descW := w - maxKeyW - 2
	if descW > 0 {
		desc := truncateRunes(entry.Description, descW)
		descRunes := []rune(desc)
		for i, r := range descRunes {
			if descX+i < x+w {
				buf.SetCell(descX+i, y, buffer.NewCell(r, h.style.Description))
			}
		}
	}
}

// String returns a string representation for debugging.
func (h *HelpOverlay) String() string {
	return fmt.Sprintf("HelpOverlay{groups: %d, query: %q, rows: %d}", len(h.groups), h.query, h.totalRowsLocked())
}

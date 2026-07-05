package component

import (
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ListItem represents a single entry in a ListView.
type ListItem struct {
	Label    string
	Value    any
	Disabled bool
	Icon     rune // optional, 0 = no icon
}

// ListViewStyle holds visual styles for the ListView.
type ListViewStyle struct {
	Normal   buffer.Style // regular item text
	Selected buffer.Style // cursor-highlighted item
	Disabled buffer.Style // disabled item
	Icon     buffer.Style // icon rune
}

// DefaultListViewStyle returns a sensible default style.
func DefaultListViewStyle() ListViewStyle {
	return ListViewStyle{
		Normal:   buffer.Style{},
		Selected: buffer.Style{Flags: buffer.Reverse},
		Disabled: buffer.Style{Flags: buffer.Dim},
		Icon:     buffer.Style{Fg: buffer.Cyan},
	}
}

// ListView is a single-column selectable list with cursor navigation.
//
// Key bindings:
//   - Up/k: move cursor up
//   - Down/j: move cursor down
//   - Home/g: jump to first item
//   - End/G: jump to last item
//   - PageUp: scroll up by viewport height
//   - PageDown: scroll down by viewport height
//   - Enter: select current item (fires OnSelect)
//
// The viewport auto-scrolls to keep the cursor visible.
//
// Concurrent safe via sync.RWMutex.
type ListView struct {
	BaseComponent
	mu sync.RWMutex

	items    []ListItem
	cursor   int // currently highlighted item
	scrollTo int // first visible item index
	style    ListViewStyle

	OnSelect  func(item ListItem, index int)
	OnChange  func(cursor int) // fired on cursor move
	OnKey     func(key *term.KeyEvent) bool
}

// NewListView creates a ListView with the given labels.
func NewListView(labels []string) *ListView {
	items := make([]ListItem, len(labels))
	for i, l := range labels {
		items[i] = ListItem{Label: l, Value: i}
	}
	lv := &ListView{
		items:  items,
		style:  DefaultListViewStyle(),
	}
	lv.SetID(GenerateID("listview"))
	return lv
}

// --- Item management ---

// Items returns a copy of all items.
func (lv *ListView) Items() []ListItem {
	lv.mu.RLock()
	defer lv.mu.RUnlock()
	result := make([]ListItem, len(lv.items))
	copy(result, lv.items)
	return result
}

// SetItems replaces all items.
func (lv *ListView) SetItems(items []ListItem) {
	lv.mu.Lock()
	lv.items = make([]ListItem, len(items))
	copy(lv.items, items)
	if lv.cursor >= len(lv.items) {
		lv.cursor = max0(len(lv.items) - 1)
	}
	if lv.cursor < 0 {
		lv.cursor = 0
	}
	lv.scrollTo = 0
	lv.mu.Unlock()
}

// AddItem appends a new item.
func (lv *ListView) AddItem(label string, value any) {
	lv.mu.Lock()
	lv.items = append(lv.items, ListItem{Label: label, Value: value})
	lv.mu.Unlock()
}

// AddItemWithIcon appends a new item with an icon.
func (lv *ListView) AddItemWithIcon(label string, value any, icon rune) {
	lv.mu.Lock()
	lv.items = append(lv.items, ListItem{Label: label, Value: value, Icon: icon})
	lv.mu.Unlock()
}

// RemoveItem removes the item at the given index.
func (lv *ListView) RemoveItem(idx int) {
	lv.mu.Lock()
	if idx < 0 || idx >= len(lv.items) {
		lv.mu.Unlock()
		return
	}
	lv.items = append(lv.items[:idx], lv.items[idx+1:]...)
	if lv.cursor >= len(lv.items) {
		lv.cursor = max0(len(lv.items) - 1)
	}
	if lv.cursor < 0 {
		lv.cursor = 0
	}
	lv.mu.Unlock()
}

// ItemCount returns the number of items.
func (lv *ListView) ItemCount() int {
	lv.mu.RLock()
	defer lv.mu.RUnlock()
	return len(lv.items)
}

// Labels returns labels of all items.
func (lv *ListView) Labels() []string {
	lv.mu.RLock()
	defer lv.mu.RUnlock()
	result := make([]string, len(lv.items))
	for i, item := range lv.items {
		result[i] = item.Label
	}
	return result
}

// --- Style ---

// SetStyle sets the ListView style.
func (lv *ListView) SetStyle(s ListViewStyle) {
	lv.mu.Lock()
	lv.style = s
	lv.mu.Unlock()
}

// Style returns the current style.
func (lv *ListView) Style() ListViewStyle {
	lv.mu.RLock()
	defer lv.mu.RUnlock()
	return lv.style
}

// --- Cursor navigation ---

// Cursor returns the current cursor index.
func (lv *ListView) Cursor() int {
	lv.mu.RLock()
	defer lv.mu.RUnlock()
	return lv.cursor
}

// SetCursor moves the cursor to the given index (clamped, skipping disabled items).
func (lv *ListView) SetCursor(idx int) {
	lv.mu.Lock()
	lv.setNavigableCursor(idx)
	idx = lv.cursor
	lv.adjustScroll()
	lv.mu.Unlock()
	if lv.OnChange != nil {
		lv.OnChange(idx)
	}
}

// MoveUp moves the cursor up by one (wraps around).
func (lv *ListView) MoveUp() {
	lv.mu.Lock()
	if len(lv.items) == 0 {
		lv.mu.Unlock()
		return
	}
	idx := lv.cursor - 1
	if idx < 0 {
		idx = len(lv.items) - 1
	}
	lv.setNavigableCursor(idx)
	idx = lv.cursor
	lv.adjustScroll()
	lv.mu.Unlock()
	if lv.OnChange != nil {
		lv.OnChange(idx)
	}
}

// MoveDown moves the cursor down by one (wraps around).
func (lv *ListView) MoveDown() {
	lv.mu.Lock()
	if len(lv.items) == 0 {
		lv.mu.Unlock()
		return
	}
	idx := lv.cursor + 1
	if idx >= len(lv.items) {
		idx = 0
	}
	lv.setNavigableCursor(idx)
	idx = lv.cursor
	lv.adjustScroll()
	lv.mu.Unlock()
	if lv.OnChange != nil {
		lv.OnChange(idx)
	}
}

// MoveTop moves cursor to the first navigable item.
func (lv *ListView) MoveTop() {
	lv.SetCursor(0)
}

// MoveBottom moves cursor to the last navigable item.
func (lv *ListView) MoveBottom() {
	lv.mu.RLock()
	last := len(lv.items) - 1
	lv.mu.RUnlock()
	lv.SetCursor(last)
}

// PageUp moves the cursor up by the viewport height.
func (lv *ListView) PageUp() {
	bounds := lv.Bounds()
	h := bounds.H
	if h <= 0 {
		h = 10
	}
	lv.mu.Lock()
	lv.setNavigableCursor(lv.cursor - h)
	idx := lv.cursor
	lv.adjustScroll()
	lv.mu.Unlock()
	if lv.OnChange != nil {
		lv.OnChange(idx)
	}
}

// PageDown moves the cursor down by the viewport height.
func (lv *ListView) PageDown() {
	bounds := lv.Bounds()
	h := bounds.H
	if h <= 0 {
		h = 10
	}
	lv.mu.Lock()
	lv.setNavigableCursor(lv.cursor + h)
	idx := lv.cursor
	lv.adjustScroll()
	lv.mu.Unlock()
	if lv.OnChange != nil {
		lv.OnChange(idx)
	}
}

// SelectedItem returns the currently highlighted item, or zero value if empty.
func (lv *ListView) SelectedItem() (ListItem, bool) {
	lv.mu.RLock()
	defer lv.mu.RUnlock()
	if len(lv.items) == 0 || lv.cursor < 0 || lv.cursor >= len(lv.items) {
		return ListItem{}, false
	}
	return lv.items[lv.cursor], true
}

// Select triggers the OnSelect callback with the current item.
func (lv *ListView) Select() {
	lv.mu.RLock()
	if len(lv.items) == 0 || lv.cursor < 0 || lv.cursor >= len(lv.items) {
		lv.mu.RUnlock()
		return
	}
	item := lv.items[lv.cursor]
	idx := lv.cursor
	lv.mu.RUnlock()
	if lv.OnSelect != nil {
		lv.OnSelect(item, idx)
	}
}

// --- Internal helpers ---

// setNavigableCursor sets cursor to idx, clamping and skipping disabled items.
// Caller must hold lv.mu.
func (lv *ListView) setNavigableCursor(idx int) {
	n := len(lv.items)
	if n == 0 {
		lv.cursor = 0
		return
	}
	// Wrap around for negative or overflow
	if idx < 0 {
		idx = n - 1
	} else if idx >= n {
		idx = 0
	}
	// Try to find a navigable (non-disabled) item
	start := idx
	for lv.items[idx].Disabled {
		idx++
		if idx >= n {
			idx = 0
		}
		if idx == start {
			// All items disabled, stay put
			break
		}
	}
	lv.cursor = idx
}

// adjustScroll ensures the cursor is within the visible viewport.
// Caller must hold lv.mu.
func (lv *ListView) adjustScroll() {
	bounds := lv.Bounds()
	h := bounds.H
	if h <= 0 {
		h = 10
	}
	if lv.cursor < lv.scrollTo {
		lv.scrollTo = lv.cursor
	}
	if lv.cursor >= lv.scrollTo+h {
		lv.scrollTo = lv.cursor - h + 1
	}
	// Clamp scroll
	if lv.scrollTo < 0 {
		lv.scrollTo = 0
	}
	n := len(lv.items)
	if lv.scrollTo+h > n && n > h {
		lv.scrollTo = n - h
	}
	if lv.scrollTo < 0 {
		lv.scrollTo = 0
	}
}

// ScrollOffset returns the first visible item index.
func (lv *ListView) ScrollOffset() int {
	lv.mu.RLock()
	defer lv.mu.RUnlock()
	return lv.scrollTo
}

// --- Key handling ---

// HandleKey processes keyboard input for the ListView.
func (lv *ListView) HandleKey(key *term.KeyEvent) bool {
	if key == nil {
		return false
	}

	// Custom key handler first
	lv.mu.RLock()
	customHandler := lv.OnKey
	lv.mu.RUnlock()
	if customHandler != nil && customHandler(key) {
		return true
	}

	switch {
	case key.Key == term.KeyUp || (key.Rune == 'k' && key.Key == term.KeyUnknown):
		lv.MoveUp()
		return true
	case key.Key == term.KeyDown || (key.Rune == 'j' && key.Key == term.KeyUnknown):
		lv.MoveDown()
		return true
	case key.Key == term.KeyHome || (key.Rune == 'g' && key.Key == term.KeyUnknown && key.Modifiers == 0):
		lv.MoveTop()
		return true
	case key.Rune == 'G' || (key.Rune == 'g' && key.Key == term.KeyUnknown && key.Modifiers&term.ModShift != 0):
		lv.MoveBottom()
		return true
	case key.Key == term.KeyEnd:
		lv.MoveBottom()
		return true
	case key.Key == term.KeyPageUp:
		lv.PageUp()
		return true
	case key.Key == term.KeyPageDown:
		lv.PageDown()
		return true
	case key.Key == term.KeyEnter:
		lv.Select()
		return true
	default:
		return false
	}
}

// --- Component interface ---

// Measure returns the desired size based on the longest label.
func (lv *ListView) Measure(cs Constraints) Size {
	lv.mu.RLock()
	defer lv.mu.RUnlock()

	maxLen := 0
	for _, item := range lv.items {
		l := utf8.RuneCountInString(item.Label) + 2
		if item.Icon != 0 {
			l += 2 // icon + space
		}
		if l > maxLen {
			maxLen = l
		}
	}
	w := maxLen
	h := len(lv.items)
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

// Paint renders the list into the buffer.
func (lv *ListView) Paint(buf *buffer.Buffer) {
	lv.mu.RLock()
	defer lv.mu.RUnlock()

	bounds := lv.Bounds()
	if bounds.W == 0 || bounds.H == 0 {
		return
	}

	n := len(lv.items)
	if n == 0 {
		return
	}

	// Calculate visible range
	h := bounds.H
	start := lv.scrollTo
	if start < 0 {
		start = 0
	}
	end := start + h
	if end > n {
		end = n
	}

	for row := 0; row < h && start+row < end; row++ {
		itemIdx := start + row
		item := lv.items[itemIdx]
		isCursor := itemIdx == lv.cursor

		var style buffer.Style
		switch {
		case item.Disabled:
			style = lv.style.Disabled
		case isCursor:
			style = lv.style.Selected
		default:
			style = lv.style.Normal
		}

		col := bounds.X

		// Draw cursor indicator
		if isCursor {
			buf.SetCell(col, bounds.Y+row, buffer.Cell{Rune: '>', Width: 1, Fg: lv.style.Icon.Fg, Bg: style.Bg, Flags: style.Flags})
		} else {
			buf.SetCell(col, bounds.Y+row, buffer.Cell{Rune: ' ', Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
		}
		col++

		// Draw icon if present
		if item.Icon != 0 {
			buf.SetCell(col, bounds.Y+row, buffer.Cell{Rune: item.Icon, Width: 1, Fg: lv.style.Icon.Fg, Bg: style.Bg, Flags: style.Flags})
			col++
		}

		// Draw separator space
		col++

		// Draw label (truncate to fit)
		availW := bounds.W - (col - bounds.X)
		if availW <= 0 {
			continue
		}
		runes := []rune(item.Label)
		for i := 0; i < availW && i < len(runes); i++ {
			buf.SetCell(col, bounds.Y+row, buffer.Cell{Rune: runes[i], Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
			col++
		}

		// Fill rest of line with background
		for col-bounds.X < bounds.W {
			buf.SetCell(col, bounds.Y+row, buffer.Cell{Rune: ' ', Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
			col++
		}
	}

	// Draw scrollbar if content overflows
	if n > h && h > 1 {
		scrollCol := bounds.X + bounds.W - 1
		trackStyle := buffer.Style{Fg: lv.style.Disabled.Fg, Flags: buffer.Dim}
		for i := 0; i < h; i++ {
			buf.SetCell(scrollCol, bounds.Y+i, buffer.Cell{Rune: '|', Width: 1, Fg: trackStyle.Fg, Bg: trackStyle.Bg, Flags: trackStyle.Flags})
		}
		// Thumb position
		thumbRatio := float64(lv.scrollTo) / float64(max0(n-h))
		thumbPos := int(thumbRatio * float64(h-1))
		buf.SetCell(scrollCol, bounds.Y+thumbPos, buffer.Cell{Rune: '*', Width: 1, Fg: lv.style.Icon.Fg, Bg: trackStyle.Bg, Flags: trackStyle.Flags})
	}
}

// --- Filter ---

// Filter returns indices of items matching the query (case-insensitive substring).
func (lv *ListView) Filter(query string) []int {
	lv.mu.RLock()
	defer lv.mu.RUnlock()
	if query == "" {
		result := make([]int, len(lv.items))
		for i := range lv.items {
			result[i] = i
		}
		return result
	}
	q := strings.ToLower(query)
	var result []int
	for i, item := range lv.items {
		if strings.Contains(strings.ToLower(item.Label), q) {
			result = append(result, i)
		}
	}
	return result
}

// SetFilter narrows visible items to those matching the query.
// Returns the number of matching items. Cursor is reset to the first match.
func (lv *ListView) SetFilter(query string) int {
	indices := lv.Filter(query)
	lv.mu.Lock()
	if len(indices) > 0 {
		lv.cursor = indices[0]
		lv.scrollTo = 0
	}
	lv.mu.Unlock()
	return len(indices)
}

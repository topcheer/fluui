package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/fuzzy"
	"github.com/topcheer/fluui/internal/term"
)

// ============================================================
// AutoComplete — Input completion popup component (P18-B)
//
// Provides fuzzy-matched completion suggestions in a popup.
// Integrates with InputLine and TextArea via the OnSelect callback.
// ============================================================

// CompletionItem represents a single suggestion in the popup.
type CompletionItem struct {
	Label       string // display text shown in the popup
	Value       string // text to insert when selected
	Description string // optional helper text
	Category    string // optional group label
}

// AutoCompleteStyle holds the visual styling for the popup.
type AutoCompleteStyle struct {
	Border      buffer.Style
	Normal      buffer.Style
	Selected    buffer.Style
	Highlight   buffer.Style // matched characters
	Description buffer.Style
	Category    buffer.Style
}

// DefaultAutoCompleteStyle returns a sensible default style.
func DefaultAutoCompleteStyle() AutoCompleteStyle {
	return AutoCompleteStyle{
		Border:      buffer.Style{Fg: buffer.Color256Val(240)},
		Normal:      buffer.Style{Fg: buffer.Color256Val(252)},
		Selected:    buffer.Style{Fg: buffer.Color256Val(255), Bg: buffer.Color256Val(60), Flags: buffer.Bold},
		Highlight:   buffer.Style{Fg: buffer.Color256Val(81), Flags: buffer.Bold},
		Description: buffer.Style{Fg: buffer.Color256Val(243)},
		Category:    buffer.Style{Fg: buffer.Color256Val(215)},
	}
}

// filteredItem pairs a completion with its fuzzy match result.
type filteredItem struct {
	item     CompletionItem
	result   fuzzy.Result
	segments []fuzzy.Segment
}

// AutoComplete is a completion popup that shows fuzzy-matched suggestions.
type AutoComplete struct {
	BaseComponent
	mu sync.RWMutex

	items      []CompletionItem
	filtered   []filteredItem
	query      string
	cursor     int
	visible    bool
	maxVisible int
	scrollY    int
	x, y       int
	width      int
	matcher    *fuzzy.Matcher
	style      AutoCompleteStyle

	OnSelect  func(item CompletionItem)
	OnDismiss func()
}

// NewAutoComplete creates a new AutoComplete popup.
func NewAutoComplete() *AutoComplete {
	ac := &AutoComplete{
		items:      make([]CompletionItem, 0),
		cursor:     0,
		visible:    false,
		maxVisible: 10,
		scrollY:    0,
		matcher:    fuzzy.NewMatcher(),
		style:      DefaultAutoCompleteStyle(),
	}
	ac.SetID(GenerateID("autocomplete"))
	return ac
}

// ─── Items ───────────────────────────────────────────────────

// SetItems replaces the full set of candidate completions.
func (ac *AutoComplete) SetItems(items []CompletionItem) {
	ac.mu.Lock()
	ac.items = make([]CompletionItem, len(items))
	copy(ac.items, items)
	ac.applyFilterLocked()
	ac.mu.Unlock()
}

// AddItem appends a single completion item.
func (ac *AutoComplete) AddItem(item CompletionItem) {
	ac.mu.Lock()
	ac.items = append(ac.items, item)
	ac.applyFilterLocked()
	ac.mu.Unlock()
}

// Items returns a copy of all candidate items.
func (ac *AutoComplete) Items() []CompletionItem {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	out := make([]CompletionItem, len(ac.items))
	copy(out, ac.items)
	return out
}

// ItemCount returns the total number of candidates.
func (ac *AutoComplete) ItemCount() int {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return len(ac.items)
}

// Clear removes all items and hides the popup.
func (ac *AutoComplete) Clear() {
	ac.mu.Lock()
	ac.items = ac.items[:0]
	ac.filtered = ac.filtered[:0]
	ac.query = ""
	ac.cursor = 0
	ac.scrollY = 0
	ac.visible = false
	ac.mu.Unlock()
}

// ─── Query & Filtering ───────────────────────────────────────

// SetQuery updates the filter text and re-computes filtered results.
func (ac *AutoComplete) SetQuery(query string) {
	ac.mu.Lock()
	ac.query = query
	ac.applyFilterLocked()
	ac.mu.Unlock()
}

// Query returns the current filter text.
func (ac *AutoComplete) Query() string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.query
}

// FilteredCount returns the number of items matching the current query.
func (ac *AutoComplete) FilteredCount() int {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return len(ac.filtered)
}

// HasResults returns true if there are filtered items to show.
func (ac *AutoComplete) HasResults() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return len(ac.filtered) > 0
}

// applyFilterLocked re-computes the filtered list from items + query.
// Uses fuzzy.Rank() which returns []Result with OriginalIndex and Highlight().
func (ac *AutoComplete) applyFilterLocked() {
	if len(ac.items) == 0 {
		ac.filtered = ac.filtered[:0]
		ac.cursor = 0
		ac.scrollY = 0
		return
	}

	candidates := make([]string, len(ac.items))
	for i, it := range ac.items {
		candidates[i] = it.Label
	}

	results := ac.matcher.Rank(ac.query, candidates)

	ac.filtered = ac.filtered[:0]
	for i := range results {
		idx := results[i].OriginalIndex
		if idx < 0 || idx >= len(ac.items) {
			continue
		}
		ac.filtered = append(ac.filtered, filteredItem{
			item:     ac.items[idx],
			result:   results[i],
			segments: results[i].Highlight(),
		})
	}

	if ac.cursor >= len(ac.filtered) {
		ac.cursor = 0
	}
	ac.clampScrollLocked()
}

// ─── Visibility & Position ───────────────────────────────────

// Show makes the popup visible at the given screen coordinates.
func (ac *AutoComplete) Show(x, y int) {
	ac.mu.Lock()
	ac.visible = true
	ac.x = x
	ac.y = y
	ac.mu.Unlock()
}

// Hide makes the popup invisible and fires OnDismiss.
func (ac *AutoComplete) Hide() {
	ac.mu.Lock()
	wasVisible := ac.visible
	ac.visible = false
	cb := ac.OnDismiss
	ac.mu.Unlock()

	if wasVisible && cb != nil {
		cb()
	}
}

// Visible returns whether the popup is currently shown.
func (ac *AutoComplete) Visible() bool {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.visible
}

// Position returns the current (x, y) position.
func (ac *AutoComplete) Position() (int, int) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.x, ac.y
}

// SetPosition updates the popup position.
func (ac *AutoComplete) SetPosition(x, y int) {
	ac.mu.Lock()
	ac.x = x
	ac.y = y
	ac.mu.Unlock()
}

// ─── Cursor Navigation ───────────────────────────────────────

// Cursor returns the index of the highlighted item.
func (ac *AutoComplete) Cursor() int {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.cursor
}

// CurrentItem returns the highlighted completion item, or nil if none.
func (ac *AutoComplete) CurrentItem() *CompletionItem {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	if ac.cursor < 0 || ac.cursor >= len(ac.filtered) {
		return nil
	}
	item := ac.filtered[ac.cursor].item
	return &item
}

// SetCursor sets the cursor index, clamping to valid range.
func (ac *AutoComplete) SetCursor(idx int) {
	ac.mu.Lock()
	if len(ac.filtered) == 0 {
		ac.cursor = 0
		ac.mu.Unlock()
		return
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(ac.filtered) {
		idx = len(ac.filtered) - 1
	}
	ac.cursor = idx
	ac.clampScrollLocked()
	ac.mu.Unlock()
}

// MoveUp moves the cursor up by one, wrapping to the bottom.
func (ac *AutoComplete) MoveUp() {
	ac.mu.Lock()
	if len(ac.filtered) == 0 {
		ac.mu.Unlock()
		return
	}
	if ac.cursor <= 0 {
		ac.cursor = len(ac.filtered) - 1
	} else {
		ac.cursor--
	}
	ac.clampScrollLocked()
	ac.mu.Unlock()
}

// MoveDown moves the cursor down by one, wrapping to the top.
func (ac *AutoComplete) MoveDown() {
	ac.mu.Lock()
	if len(ac.filtered) == 0 {
		ac.mu.Unlock()
		return
	}
	ac.cursor++
	if ac.cursor >= len(ac.filtered) {
		ac.cursor = 0
	}
	ac.clampScrollLocked()
	ac.mu.Unlock()
}

// clampScrollLocked ensures cursor is within the visible scroll window.
func (ac *AutoComplete) clampScrollLocked() {
	if ac.cursor < ac.scrollY {
		ac.scrollY = ac.cursor
	}
	if ac.cursor >= ac.scrollY+ac.maxVisible {
		ac.scrollY = ac.cursor - ac.maxVisible + 1
	}
	if ac.scrollY < 0 {
		ac.scrollY = 0
	}
}

// ScrollY returns the current scroll offset.
func (ac *AutoComplete) ScrollY() int {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.scrollY
}

// ─── Selection ───────────────────────────────────────────────

// Select triggers OnSelect for the current item and hides.
func (ac *AutoComplete) Select() {
	ac.mu.Lock()
	if ac.cursor < 0 || ac.cursor >= len(ac.filtered) {
		ac.mu.Unlock()
		return
	}
	item := ac.filtered[ac.cursor].item
	cb := ac.OnSelect
	ac.visible = false
	ac.mu.Unlock()

	if cb != nil {
		cb(item)
	}
}

// ─── Configuration ───────────────────────────────────────────

// SetMaxVisible sets the maximum number of visible items before scrolling.
func (ac *AutoComplete) SetMaxVisible(n int) {
	ac.mu.Lock()
	if n < 1 {
		n = 1
	}
	ac.maxVisible = n
	ac.clampScrollLocked()
	ac.mu.Unlock()
}

// MaxVisible returns the maximum visible items.
func (ac *AutoComplete) MaxVisible() int {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.maxVisible
}

// SetOnSelect sets the callback fired when an item is selected.
func (ac *AutoComplete) SetOnSelect(fn func(item CompletionItem)) {
	ac.mu.Lock()
	ac.OnSelect = fn
	ac.mu.Unlock()
}

// SetOnDismiss sets the callback fired when the popup is dismissed.
func (ac *AutoComplete) SetOnDismiss(fn func()) {
	ac.mu.Lock()
	ac.OnDismiss = fn
	ac.mu.Unlock()
}

// SetStyle sets the visual style.
func (ac *AutoComplete) SetStyle(s AutoCompleteStyle) {
	ac.mu.Lock()
	ac.style = s
	ac.mu.Unlock()
}

// Style returns the current style.
func (ac *AutoComplete) Style() AutoCompleteStyle {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.style
}

// FilteredItems returns the currently filtered completion items.
func (ac *AutoComplete) FilteredItems() []CompletionItem {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	out := make([]CompletionItem, len(ac.filtered))
	for i, fi := range ac.filtered {
		out[i] = fi.item
	}
	return out
}

// SetCaseSensitive controls fuzzy matching case sensitivity.
func (ac *AutoComplete) SetCaseSensitive(v bool) {
	ac.mu.Lock()
	ac.matcher.SetCaseSensitive(v)
	ac.applyFilterLocked()
	ac.mu.Unlock()
}

// ─── Keyboard ────────────────────────────────────────────────

// HandleKey processes a key event. Returns true if the key was consumed.
func (ac *AutoComplete) HandleKey(key *term.KeyEvent) bool {
	if key == nil {
		return false
	}
	switch key.Key {
	case term.KeyUp:
		ac.MoveUp()
		return true
	case term.KeyDown:
		ac.MoveDown()
		return true
	case term.KeyTab, term.KeyEnter:
		ac.Select()
		return true
	case term.KeyEscape:
		ac.Hide()
		return true
	}
	return false
}

// ─── Component Interface ─────────────────────────────────────

func (ac *AutoComplete) visibleItemCountLocked() int {
	n := len(ac.filtered)
	if n > ac.maxVisible {
		return ac.maxVisible
	}
	return n
}

// Measure computes the desired size based on the longest filtered item label.
func (ac *AutoComplete) Measure(cs Constraints) Size {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	maxW := 0
	for _, fi := range ac.filtered {
		w := len([]rune(fi.item.Label))
		if fi.item.Description != "" {
			w += 1 + len([]rune(fi.item.Description))
		}
		if fi.item.Category != "" {
			w += 1 + len([]rune(fi.item.Category))
		}
		if w > maxW {
			maxW = w
		}
	}

	width := maxW + 4
	height := ac.visibleItemCountLocked() + 2
	if height < 3 {
		height = 3
	}
	if width < 15 {
		width = 15
	}

	if cs.MaxWidth > 0 && width > cs.MaxWidth {
		width = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && height > cs.MaxHeight {
		height = cs.MaxHeight
	}

	ac.width = width
	return Size{W: width, H: height}
}

// Paint renders the popup into the buffer.
func (ac *AutoComplete) Paint(buf *buffer.Buffer) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	if !ac.visible || len(ac.filtered) == 0 {
		return
	}

	x, y := ac.x, ac.y
	w := ac.width
	visibleN := ac.visibleItemCountLocked()
	h := visibleN + 2

	// Top border
	buf.DrawText(x, y, "\u250c", ac.style.Border)
	for i := 1; i < w-1 && x+i < buf.Width; i++ {
		buf.DrawText(x+i, y, "\u2500", ac.style.Border)
	}
	buf.DrawText(x+w-1, y, "\u2510", ac.style.Border)

	// Items
	for row := 0; row < visibleN; row++ {
		idx := ac.scrollY + row
		if idx >= len(ac.filtered) {
			break
		}
		fi := ac.filtered[idx]
		itemY := y + 1 + row

		buf.DrawText(x, itemY, "\u2502", ac.style.Border)

		style := ac.style.Normal
		isSelected := idx == ac.cursor
		if isSelected {
			style = ac.style.Selected
		}

		// Fill background for selected
		if isSelected {
			for i := 1; i < w-1 && x+i < buf.Width; i++ {
				buf.SetCell(x+i, itemY, buffer.NewCell(' ', style))
			}
		}

		contentX := x + 1
		maxContentW := w - 3

		// Category prefix
		if fi.item.Category != "" {
			catText := fi.item.Category + ":"
			catRunes := []rune(catText)
			if len(catRunes) > maxContentW/2 {
				catText = string(catRunes[:maxContentW/2])
			}
			catStyle := ac.style.Category
			if isSelected {
				catStyle = style
			}
			contentX = buf.DrawTextClamped(contentX, itemY, catText, catStyle)
			contentX = buf.DrawTextClamped(contentX, itemY, " ", style)
		}

		// Label with highlight segments
		labelStartX := contentX
		for _, seg := range fi.segments {
			segStyle := style
			if seg.Matched && !isSelected {
				segStyle = ac.style.Highlight
			}
			for _, r := range seg.Text {
				if contentX-labelStartX >= maxContentW {
					break
				}
				if contentX >= buf.Width {
					break
				}
				buf.SetCell(contentX, itemY, buffer.NewCell(r, segStyle))
				contentX++
			}
		}

		// Description (if room)
		if fi.item.Description != "" {
			remaining := maxContentW - (contentX - labelStartX)
			if remaining > 3 {
				buf.DrawTextClamped(contentX+1, itemY, fi.item.Description, ac.style.Description)
			}
		}

		buf.DrawText(x+w-1, itemY, "\u2502", ac.style.Border)
	}

	// Bottom border
	bottomY := y + h - 1
	buf.DrawText(x, bottomY, "\u2514", ac.style.Border)
	for i := 1; i < w-1 && x+i < buf.Width; i++ {
		buf.DrawText(x+i, bottomY, "\u2500", ac.style.Border)
	}
	buf.DrawText(x+w-1, bottomY, "\u2518", ac.style.Border)
}

// Children returns nil (popup has no children components).
func (ac *AutoComplete) Children() []Component { return nil }

// String returns a debug representation.
func (ac *AutoComplete) String() string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return "AutoComplete(items=" + itoa(len(ac.items)) + ",filtered=" + itoa(len(ac.filtered)) + ")"
}

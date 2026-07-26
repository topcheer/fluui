package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// FilteredList wraps a ListView with a search filter. As the user types,
// items are filtered in real-time by substring match (case-insensitive).
// Common in file pickers, command palettes, and AI tool selectors.
//
// Thread-safe.
type FilteredList struct {
	BaseComponent
	mu         sync.RWMutex
	items      []string
	filtered   []string
	query      string
	selected   int
	maxVisible int
}

// NewFilteredList creates a filtered list with the given items.
func NewFilteredList(items []string) *FilteredList {
	fl := &FilteredList{
		BaseComponent: BaseComponent{id: GenerateID("filtered")},
		items:         items,
		maxVisible:    10,
	}
	fl.applyFilterLocked()
	return fl
}

func (f *FilteredList) Items() []string { f.mu.RLock(); defer f.mu.RUnlock(); return f.items }

func (f *FilteredList) SetItems(items []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.items = items
	f.applyFilterLocked()
}

func (f *FilteredList) Query() string { f.mu.RLock(); defer f.mu.RUnlock(); return f.query }

func (f *FilteredList) SetQuery(q string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.query = q
	f.applyFilterLocked()
}

func (f *FilteredList) Filtered() []string { f.mu.RLock(); defer f.mu.RUnlock(); return f.filtered }

func (f *FilteredList) Selected() int { f.mu.RLock(); defer f.mu.RUnlock(); return f.selected }

func (f *FilteredList) SetSelected(i int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.selected = i
	f.clampSelectedLocked()
}

func (f *FilteredList) MaxVisible() int { f.mu.RLock(); defer f.mu.RUnlock(); return f.maxVisible }
func (f *FilteredList) SetMaxVisible(n int) { f.mu.Lock(); defer f.mu.Unlock(); if n < 1 { n = 1 }; f.maxVisible = n }

func (f *FilteredList) MoveUp() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.selected > 0 { f.selected-- }
}

func (f *FilteredList) MoveDown() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.selected < len(f.filtered)-1 { f.selected++ }
}

func (f *FilteredList) SelectedItem() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.selected < 0 || f.selected >= len(f.filtered) { return "" }
	return f.filtered[f.selected]
}

// applyFilterLocked updates f.filtered based on current query.
// Case-insensitive substring match.
func (f *FilteredList) applyFilterLocked() {
	if f.query == "" {
		f.filtered = f.items
		f.clampSelectedLocked()
		return
	}
	f.filtered = f.filtered[:0]
	q := f.query
	for i := 0; i < len(q); i++ {
		if q[i] >= 'A' && q[i] <= 'Z' {
			b := []byte(q)
			for j := range b { if b[j] >= 'A' && b[j] <= 'Z' { b[j] += 32 } }
			q = string(b)
			break
		}
	}
	for _, item := range f.items {
		if containsCI(item, q) {
			f.filtered = append(f.filtered, item)
		}
	}
	f.clampSelectedLocked()
}

func (f *FilteredList) clampSelectedLocked() {
	if f.selected < 0 { f.selected = 0 }
	if f.selected >= len(f.filtered) && len(f.filtered) > 0 {
		f.selected = len(f.filtered) - 1
	}
	if len(f.filtered) == 0 { f.selected = 0 }
}

// containsCI checks if s contains substr (case-insensitive, zero alloc).
func containsCI(s, substr string) bool {
	if len(substr) == 0 { return true }
	if len(substr) > len(s) { return false }
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := s[i+j]
			c2 := substr[j]
			if c1 >= 'A' && c1 <= 'Z' { c1 += 32 }
			// c2 already lowercased by applyFilterLocked
			if c1 != c2 { match = false; break }
		}
		if match { return true }
	}
	return false
}

func (f *FilteredList) Measure(cs Constraints) Size {
	f.mu.RLock()
	defer f.mu.RUnlock()
	w := 10
	for _, item := range f.filtered {
		if iw := len(item); iw > w { w = iw }
	}
	h := len(f.filtered)
	if h > f.maxVisible { h = f.maxVisible }
	if h < 1 { h = 1 }
	if cs.HasWidth() && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.HasHeight() && h > cs.MaxHeight { h = cs.MaxHeight }
	if w < 1 { w = 1 }
	if h < 1 { h = 1 }
	return Size{W: w, H: h}
}

func (f *FilteredList) Paint(buf *buffer.Buffer) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	bounds := f.bounds
	if bounds.W <= 0 || bounds.H <= 0 { return }

	tt := theme.Get()
	normal := buffer.Style{Fg: tt.Fg}
	selected := buffer.Style{Fg: tt.Bg, Bg: tt.Accent, Flags: buffer.Bold}
	muted := buffer.Style{Fg: tt.Muted}

	for i := 0; i < bounds.H; i++ {
		if i >= len(f.filtered) { break }
		item := f.filtered[i]
		style := normal
		if i == f.selected { style = selected }
		for col := 0; col < bounds.W; col++ {
			if col < len(item) {
				buf.SetCell(bounds.X+col, bounds.Y+i, buffer.Cell{
					Rune: rune(item[col]), Width: 1,
					Fg: style.Fg, Bg: style.Bg, Flags: style.Flags,
				})
			} else {
				buf.SetCell(bounds.X+col, bounds.Y+i, buffer.Cell{
					Rune: ' ', Width: 1, Fg: style.Fg, Bg: style.Bg,
				})
			}
		}
	}

	// Empty state
	if len(f.filtered) == 0 && bounds.H > 0 {
		buf.DrawText(bounds.X, bounds.Y, "No matches", muted)
	}
}

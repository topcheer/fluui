package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// SearchBar renders a search input with a magnifying glass icon and placeholder.
// Emits search queries for filtering lists, file trees, or command palettes.
//
// Thread-safe.
type SearchBar struct {
	BaseComponent
	mu          sync.RWMutex
	placeholder string
	query       string
	focused     bool
}

// NewSearchBar creates a search bar with the given placeholder.
func NewSearchBar(placeholder string) *SearchBar {
	return &SearchBar{
		BaseComponent: BaseComponent{id: GenerateID("searchbar")},
		placeholder:   placeholder,
	}
}

func (s *SearchBar) Query() string { s.mu.RLock(); defer s.mu.RUnlock(); return s.query }
func (s *SearchBar) SetQuery(q string) { s.mu.Lock(); defer s.mu.Unlock(); s.query = q }

func (s *SearchBar) Placeholder() string { s.mu.RLock(); defer s.mu.RUnlock(); return s.placeholder }
func (s *SearchBar) SetPlaceholder(p string) { s.mu.Lock(); defer s.mu.Unlock(); s.placeholder = p }

func (s *SearchBar) Focused() bool { s.mu.RLock(); defer s.mu.RUnlock(); return s.focused }
func (s *SearchBar) SetFocus(b bool) { s.mu.Lock(); defer s.mu.Unlock(); s.focused = b }

func (s *SearchBar) Clear() { s.mu.Lock(); defer s.mu.Unlock(); s.query = "" }

func (s *SearchBar) Measure(cs Constraints) Size {
	w := cs.MaxWidth
	if w <= 0 { w = 40 }
	h := 1
	if cs.HasWidth() && w > cs.MaxWidth { w = cs.MaxWidth }
	if w < 1 { w = 1 }
	return Size{W: w, H: h}
}

func (s *SearchBar) Paint(buf *buffer.Buffer) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bounds := s.bounds
	if bounds.W <= 0 || bounds.H <= 0 { return }

	tt := theme.Get()
	muted := buffer.Style{Fg: tt.Muted}
	normal := buffer.Style{Fg: tt.Fg}
	accent := buffer.Style{Fg: tt.Accent}

	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	// Search icon
	x = buf.DrawText(x, y, "\u2315 ", accent) // ⌕
	if x >= maxX { return }

	if s.query != "" {
		x = buf.DrawText(x, y, s.query, normal)
	} else {
		x = buf.DrawText(x, y, s.placeholder, muted)
	}

	// Cursor when focused
	if s.focused && x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: '\u2588', Width: 1, Fg: tt.Accent})
	}
}

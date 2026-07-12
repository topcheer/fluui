package component

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// SessionItem represents a single chat session in the sidebar.
type SessionItem struct {
	ID          string
	Title       string
	Workspace   string
	LastMessage string
	LastTime    time.Time
	Pinned      bool
	Llocked     bool
	Busy        bool
	HasError    bool
	UnreadCount int
}

// SessionSidebar is a collapsible sidebar showing chat sessions grouped by workspace.
// Supports search filtering, keyboard navigation, and context menu actions.
type SessionSidebar struct {
	BaseComponent

	mu           sync.RWMutex
	items        []SessionItem
	filtered     []int
	filterText   string
	selected     int
	collapsed    bool
	scrollOffset int
	width        int
	searchFocus  bool

	groups          []string
	expandedGroups  map[string]bool

	OnSelect func(id string)
	OnRename func(id string)
	OnDelete func(id string)
	OnLock   func(id string)
	OnPin    func(id string)
	OnExport func(id string)
}

// NewSessionSidebar creates a new session sidebar with default settings.
func NewSessionSidebar() *SessionSidebar {
	return &SessionSidebar{
		width:         30,
		expandedGroups: make(map[string]bool),
		selected:      -1,
	}
}

// SetItems replaces all session items and rebuilds groups/filters.
func (s *SessionSidebar) SetItems(items []SessionItem) {
	s.mu.Lock()
	s.items = items
	s.rebuildGroups()
	s.applyFilterLocked()
	s.mu.Unlock()
}

// SetFilter sets the search filter text.
func (s *SessionSidebar) SetFilter(text string) {
	s.mu.Lock()
	s.filterText = text
	s.applyFilterLocked()
	s.mu.Unlock()
}

// SelectedItem returns the currently selected session item, or nil.
func (s *SessionSidebar) SelectedItem() *SessionItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.selected < 0 || s.selected >= len(s.filtered) {
		return nil
	}
	idx := s.filtered[s.selected]
	if idx >= len(s.items) {
		return nil
	}
	return &s.items[idx]
}

// SetCollapsed sets the collapsed state.
func (s *SessionSidebar) SetCollapsed(c bool) { s.mu.Lock(); s.collapsed = c; s.mu.Unlock() }

// ToggleCollapsed toggles collapsed state.
func (s *SessionSidebar) ToggleCollapsed() { s.mu.Lock(); s.collapsed = !s.collapsed; s.mu.Unlock() }

// IsCollapsed returns the collapsed state.
func (s *SessionSidebar) IsCollapsed() bool { s.mu.RLock(); defer s.mu.RUnlock(); return s.collapsed }

// SetWidth sets the expanded width.
func (s *SessionSidebar) SetWidth(w int) { s.mu.Lock(); s.width = w; s.mu.Unlock() }

// rebuildGroups builds the list of unique workspace names from items.
func (s *SessionSidebar) rebuildGroups() {
	seen := make(map[string]bool)
	s.groups = s.groups[:0]
	for _, item := range s.items {
		if !seen[item.Workspace] {
			seen[item.Workspace] = true
			s.groups = append(s.groups, item.Workspace)
			if _, ok := s.expandedGroups[item.Workspace]; !ok {
				s.expandedGroups[item.Workspace] = true // expanded by default
			}
		}
	}
	sort.Strings(s.groups)
}

// applyFilterLocked rebuilds the filtered index list. Caller must hold lock.
func (s *SessionSidebar) applyFilterLocked() {
	s.filtered = s.filtered[:0]
	query := strings.ToLower(s.filterText)

	for i, item := range s.items {
		if query != "" {
			title := strings.ToLower(item.Title)
			msg := strings.ToLower(item.LastMessage)
			if !strings.Contains(title, query) && !strings.Contains(msg, query) {
				continue
			}
		}

		// Skip if group is collapsed and not searching
		if query == "" {
			expanded := s.expandedGroups[item.Workspace]
			if !expanded {
				continue
			}
		}

		s.filtered = append(s.filtered, i)
	}

	// Sort: pinned first, then by last time descending
	sort.SliceStable(s.filtered, func(a, b int) bool {
		ia, ib := s.filtered[a], s.filtered[b]
		if s.items[ia].Pinned != s.items[ib].Pinned {
			return s.items[ia].Pinned
		}
		return s.items[ia].LastTime.After(s.items[ib].LastTime)
	})

	// Clamp selection
	if s.selected >= len(s.filtered) {
		s.selected = len(s.filtered) - 1
	}
	if s.selected < 0 && len(s.filtered) > 0 {
		s.selected = 0
	}
}

// Measure returns the desired size.
func (s *SessionSidebar) Measure(cs Constraints) Size {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.collapsed {
		return Size{W: 1, H: cs.MaxHeight}
	}
	h := cs.MaxHeight
	if h <= 0 {
		h = 24
	}
	return Size{W: s.width, H: h}
}

// SetBounds sets position and size.
func (s *SessionSidebar) SetBounds(r Rect) {
	s.BaseComponent.SetBounds(r)
}

// Paint renders the sidebar.
func (s *SessionSidebar) Paint(buf *buffer.Buffer) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bounds := s.Bounds()
	if s.collapsed {
		// Draw narrow indicator strip
		for y := bounds.Y; y < bounds.Y+bounds.H && y < buf.Height; y++ {
			if bounds.X < buf.Width {
				buf.SetCell(bounds.X, y, buffer.Cell{Rune: '\u258c', Width: 1, Fg: buffer.RGB(98, 114, 164)})
			}
		}
		return
	}

	fg := buffer.RGB(248, 248, 242)   // dracula foreground
	muted := buffer.RGB(139, 143, 159) // dracula comment
	accent := buffer.RGB(189, 147, 249) // dracula purple
	errColor := buffer.RGB(255, 85, 85)
	busyColor := buffer.RGB(241, 250, 140)  // dracula yellow
	success := buffer.RGB(80, 250, 123)     // dracula green

	y := bounds.Y
	maxY := bounds.Y + bounds.H
	x := bounds.X
	w := bounds.W

	// Draw search box
	if y < maxY && y < buf.Height {
		searchText := s.filterText
		if searchText == "" {
			searchText = "[search...]"
		}
		style := muted
		if s.searchFocus {
			style = fg
		}
		sbDrawTextLeft(buf, x, y, "[\U0001F50D "+searchText+"]", style, w)
		// Clear rest of line
		for i := x + sbVisibleStrLen("[\U0001F50D "+searchText+"]"); i < x+w && i < buf.Width; i++ {
			buf.SetCell(i, y, buffer.Cell{Rune: ' ', Width: 1, Fg: style})
		}
		y++
		// Separator
		if y < maxY && y < buf.Height {
			for i := x; i < x+w && i < buf.Width; i++ {
				buf.SetCell(i, y, buffer.Cell{Rune: '\u2500', Width: 1, Fg: muted})
			}
			y++
		}
	}

	// Draw groups
	selIdx := s.selected
	for _, gname := range s.groups {
		if y >= maxY {
			break
		}

		expanded := s.expandedGroups[gname]
		// Count items in this group
		count := 0
		for _, idx := range s.filtered {
			if s.items[idx].Workspace == gname {
				count++
			}
		}
		if count == 0 && s.filterText == "" {
			continue
		}

		// Group header
		arrow := "\u25BC" // ▼
		if !expanded {
			arrow = "\u25B6" // ▶
		}
		header := fmt.Sprintf("%s %s (%d)", arrow, gname, count)
		if y < buf.Height {
			sbDrawTextLeft(buf, x, y, header, accent, w)
			for i := x + sbVisibleStrLen(header); i < x+w && i < buf.Width; i++ {
				buf.SetCell(i, y, buffer.Cell{Rune: ' ', Width: 1, Fg: accent})
			}
		}
		y++

		if !expanded {
			continue
		}

		// Draw items in this group
		for fIdx, itemIdx := range s.filtered {
			item := s.items[itemIdx]
			if item.Workspace != gname {
				continue
			}
			if y >= maxY {
				break
			}

			isSelected := fIdx == selIdx
			bg := buffer.Color{} // transparent
			textColor := fg
			if isSelected {
				bg = buffer.RGB(68, 71, 90) // dracula current line
			}

			// Status indicator
			var indicator string
			indicatorColor := muted
			if item.HasError {
				indicator = "\u2717" // ✕
				indicatorColor = errColor
			} else if item.Busy {
				indicator = "\u25CF" // ●
				indicatorColor = busyColor
			} else if item.Pinned {
				indicator = "\U0001F4CC" // 📌
				indicatorColor = accent
			} else {
				indicator = "\u25CB" // ○
				indicatorColor = success
			}

			// Title line
			titleMax := w - 12 // leave room for time
			if titleMax < 5 {
				titleMax = 5
			}
			title := sbTruncateStr(item.Title, titleMax)
			timeStr := formatTimeAgo(item.LastTime)

			line := fmt.Sprintf("  %s %s", indicator, title)
			_ = line // drawn cell by cell below
			if y < buf.Height {
				// Draw indicator
				drawX := x + 1
				for _, r := range indicator {
					if drawX < x+w && drawX < buf.Width {
						buf.SetCell(drawX, y, buffer.Cell{Rune: r, Width: 1, Fg: indicatorColor, Bg: bg})
						drawX++
					}
				}
				// Draw title
				for _, r := range title {
					if drawX < x+w && drawX < buf.Width {
						buf.SetCell(drawX, y, buffer.Cell{Rune: r, Width: 1, Fg: textColor, Bg: bg})
						drawX++
					}
				}
				// Draw time on the right
				tx := x + w - len(timeStr)
				if tx < drawX {
					tx = drawX
				}
				for _, r := range timeStr {
					if tx < x+w && tx < buf.Width {
						buf.SetCell(tx, y, buffer.Cell{Rune: r, Width: 1, Fg: muted, Bg: bg})
						tx++
					}
				}
				// Clear remaining
				for i := drawX; i < tx && i < x+w && i < buf.Width; i++ {
					buf.SetCell(i, y, buffer.Cell{Rune: ' ', Width: 1, Bg: bg})
				}
			}
			y++

			// Last message preview (if room)
			if y < maxY && item.LastMessage != "" {
				previewMax := w - 4
				if previewMax < 5 {
					previewMax = 5
				}
				preview := sbTruncateStr(item.LastMessage, previewMax)
				if y < buf.Height {
					drawX := x + 4
					for _, r := range preview {
						if drawX < x+w && drawX < buf.Width {
							buf.SetCell(drawX, y, buffer.Cell{Rune: r, Width: 1, Fg: muted, Bg: bg})
							drawX++
						}
					}
				}
				y++
			}
		}
	}
}

// HandleKey processes keyboard input for the sidebar.
func (s *SessionSidebar) HandleKey(k *term.KeyEvent) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.searchFocus {
		switch {
		case k.Key == term.KeyEscape:
			s.searchFocus = false
			s.filterText = ""
			s.applyFilterLocked()
			return true
		case k.Key == term.KeyEnter:
			s.searchFocus = false
			return true
		case k.Rune == '\u007f' || k.Key == term.KeyBackspace: // backspace
			if len(s.filterText) > 0 {
				s.filterText = s.filterText[:len(s.filterText)-1]
				s.applyFilterLocked()
			}
			return true
		case k.Rune >= 0x20 && k.Rune <= 0x7e: // printable ASCII
			s.filterText += string(k.Rune)
			s.applyFilterLocked()
			return true
		default:
			return true // consume all keys when search is focused
		}
	}

	switch {
	case k.Rune == '/':
		s.searchFocus = true
		return true
	case k.Rune == 'j' || k.Key == term.KeyDown:
		if s.selected < len(s.filtered)-1 {
			s.selected++
		}
		return true
	case k.Rune == 'k' || k.Key == term.KeyUp:
		if s.selected > 0 {
			s.selected--
		}
		return true
	case k.Key == term.KeyEnter:
		if s.selected >= 0 && s.selected < len(s.filtered) {
			item := s.items[s.filtered[s.selected]]
			if s.OnSelect != nil {
				s.mu.Unlock()
				s.OnSelect(item.ID)
				s.mu.Lock()
			}
		}
		return true
	case k.Rune == 'g':
		s.selected = 0
		return true
	case k.Rune == 'G':
		s.selected = len(s.filtered) - 1
		return true
	default:
		return false
	}
}

// HandleMouse handles mouse clicks for group expand/collapse and item selection.
func (s *SessionSidebar) HandleMouse(m *term.MouseEvent) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Mouse left click on group header toggles expand
	// This is a simplified version — full implementation would track click positions
	_ = m
	return false
}

// ToggleGroup expands/collapses a workspace group.
func (s *SessionSidebar) ToggleGroup(workspace string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expandedGroups[workspace] = !s.expandedGroups[workspace]
	s.applyFilterLocked()
}

// SetGroupExpanded sets the expanded state of a group.
func (s *SessionSidebar) SetGroupExpanded(workspace string, expanded bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expandedGroups[workspace] = expanded
	s.applyFilterLocked()
}

// --- helpers ---

func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 48*time.Hour:
		return "1d"
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	default:
		return t.Format("01/02")
	}
}

func sbTruncateStr(s string, maxLen int) string {
	if sbVisibleStrLen(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	if maxLen <= 1 {
		return "\u2026"
	}
	result := make([]rune, 0, maxLen)
	width := 0
	for _, r := range runes {
		if width+1 >= maxLen {
			break
		}
		result = append(result, r)
		width++
	}
	result = append(result, '\u2026') // …
	return string(result)
}

func sbDrawTextLeft(buf *buffer.Buffer, x, y int, text string, color buffer.Color, maxWidth int) {
	dx := x
	for _, r := range text {
		if dx >= x+maxWidth || dx >= buf.Width {
			break
		}
		buf.SetCell(dx, y, buffer.Cell{Rune: r, Width: 1, Fg: color})
		dx++
	}
}

func sbVisibleStrLen(s string) int {
	// Simple rune count (good enough for our use case)
	count := 0
	for range s {
		count++
	}
	return count
}

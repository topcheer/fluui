package app

import (
	"fmt"
	"strings"

	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// SearchMatch represents a single search result within a block.
type SearchMatch struct {
	BlockID   string
	BlockType string
	Snippet   string // surrounding text for display
	LineNum   int    // line within the block (0-based)
}

// SearchMode manages search state within ChatApp.
// When active, it shows a search bar at the bottom of the screen and
// highlights matches in the conversation content.
type SearchMode struct {
	active  bool
	query   string
	matches []SearchMatch
	current int // current match index (0-based), -1 if none
}

// NewSearchMode creates an inactive SearchMode.
func NewSearchMode() *SearchMode {
	return &SearchMode{
		current: -1,
	}
}

// IsActive reports whether search mode is active.
func (s *SearchMode) IsActive() bool {
	return s.active
}

// Query returns the current search query.
func (s *SearchMode) Query() string {
	return s.query
}

// MatchCount returns the total number of matches.
func (s *SearchMode) MatchCount() int {
	return len(s.matches)
}

// CurrentIndex returns the current match index (1-based for display).
// Returns 0 if no matches.
func (s *SearchMode) CurrentIndex() int {
	if s.current < 0 || s.current >= len(s.matches) {
		return 0
	}
	return s.current + 1
}

// CurrentMatch returns the current match, or nil if none.
func (s *SearchMode) CurrentMatch() *SearchMatch {
	if s.current < 0 || s.current >= len(s.matches) {
		return nil
	}
	return &s.matches[s.current]
}

// CurrentMatches returns all matches (for rendering highlight overlays).
func (s *SearchMode) CurrentMatches() []SearchMatch {
	return s.matches
}

// StartSearch enters search mode.
func (s *SearchMode) StartSearch() {
	s.active = true
	s.query = ""
	s.matches = nil
	s.current = -1
}

// CloseSearch exits search mode and resets state.
func (s *SearchMode) CloseSearch() {
	s.active = false
	s.query = ""
	s.matches = nil
	s.current = -1
}

// UpdateQuery recomputes matches from the given blocks using the new query.
// Search is case-insensitive. An empty query clears all matches.
func (s *SearchMode) UpdateQuery(q string, blocks []block.Block) {
	s.query = q
	if q == "" {
		s.matches = nil
		s.current = -1
		return
	}

	lowerQuery := strings.ToLower(q)
	s.matches = s.matches[:0] // reuse capacity

	for _, b := range blocks {
		text, ok := extractBlockText(b)
		if !ok || text == "" {
			continue
		}
		s.searchInBlock(b, text, lowerQuery)
	}

	if len(s.matches) > 0 {
		s.current = 0
	} else {
		s.current = -1
	}
}

// searchInBlock finds all occurrences of query within a block's text
// and appends them to s.matches.
func (s *SearchMode) searchInBlock(b block.Block, text, lowerQuery string) {
	lowerText := strings.ToLower(text)
	blockID := b.ID()
	blockType := b.Type().String()

	lineNum := 0
	searchStart := 0
	for {
		idx := strings.Index(lowerText[searchStart:], lowerQuery)
		if idx < 0 {
			break
		}
		absIdx := searchStart + idx

		// Count newlines before this match to get line number
		lineNum = strings.Count(text[:absIdx], "\n")

		// Build snippet (up to 40 chars around match)
		snippet := buildSnippet(text, absIdx, len(lowerQuery))

		s.matches = append(s.matches, SearchMatch{
			BlockID:   blockID,
			BlockType: blockType,
			Snippet:   snippet,
			LineNum:   lineNum,
		})

		searchStart = absIdx + len(lowerQuery)
		if searchStart >= len(lowerText) {
			break
		}
	}
}

// buildSnippet extracts a context window around the match position.
func buildSnippet(text string, matchIdx, matchLen int) string {
	const maxSnippet = 40

	start := matchIdx - maxSnippet/2
	if start < 0 {
		start = 0
	}
	end := matchIdx + matchLen + maxSnippet/2
	if end > len(text) {
		end = len(text)
	}

	snippet := text[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(text) {
		snippet = snippet + "..."
	}

	// Replace newlines with spaces for single-line display
	snippet = strings.ReplaceAll(snippet, "\n", " ")
	return snippet
}

// NextMatch advances to the next match. Wraps around to the beginning.
func (s *SearchMode) NextMatch() {
	if len(s.matches) == 0 {
		return
	}
	s.current = (s.current + 1) % len(s.matches)
}

// PrevMatch goes to the previous match. Wraps around to the end.
func (s *SearchMode) PrevMatch() {
	if len(s.matches) == 0 {
		return
	}
	s.current = (s.current - 1 + len(s.matches)) % len(s.matches)
}

// HandleKey processes a key event while search mode is active.
// Returns true if the key was consumed by search.
func (s *SearchMode) HandleKey(key *term.KeyEvent) bool {
	if !s.active {
		return false
	}

	// Escape: close search
	if key.Key == term.KeyEscape {
		s.CloseSearch()
		return true
	}

	// Enter: go to next match (Shift+Enter = previous)
	if key.Key == term.KeyEnter {
		if key.Modifiers&term.ModShift != 0 {
			s.PrevMatch()
		} else {
			s.NextMatch()
		}
		return true
	}

	// Ctrl+F again: also goes to next match (convenience)
	if key.Modifiers&term.ModCtrl != 0 && (key.Rune == 'f' || key.Rune == 'F') {
		if key.Modifiers&term.ModShift != 0 {
			s.PrevMatch()
		} else {
			s.NextMatch()
		}
		return true
	}

	// Backspace: remove last char from query
	if key.Key == term.KeyBackspace {
		if len(s.query) > 0 {
			// Convert to rune slice for proper Unicode handling
			runes := []rune(s.query)
			s.query = string(runes[:len(runes)-1])
		}
		return true
	}

	// Printable character: append to query
	if key.Rune != 0 && key.Rune >= 0x20 && key.Modifiers&term.ModCtrl == 0 {
		s.query += string(key.Rune)
		return true
	}

	return false
}

// StatusText returns the display text for the search bar.
// Example: "3/12 matches" or "no matches" or "Search: query".
func (s *SearchMode) StatusText() string {
	if !s.active {
		return ""
	}
	if s.query == "" {
		return "Search: "
	}
	total := len(s.matches)
	if total == 0 {
		return fmt.Sprintf("Search: %s (no matches)", s.query)
	}
	return fmt.Sprintf("Search: %s (%d/%d matches)", s.query, s.CurrentIndex(), total)
}

// RenderSearchBar draws the search bar at the bottom of the buffer.
// It occupies one line, showing the search query and match status.
func (s *SearchMode) RenderSearchBar(buf *buffer.Buffer, width, y int) {
	if !s.active || width <= 0 || y < 0 || y >= buf.Height {
		return
	}

	t := theme.Get()
	bgCell := buffer.Cell{Rune: ' ', Width: 1, Bg: t.SearchBarBg}
	for x := 0; x < width && x < buf.Width; x++ {
		buf.SetCell(x, y, bgCell)
	}

	// Draw status text
	status := s.StatusText()
	if status == "" {
		return
	}

	promptStyle := buffer.Style{
		Fg:    t.SearchBarFg,
		Bg:    t.SearchBarBg,
		Flags: 0,
	}
	// Draw match count in accent color if there are matches
	if s.query != "" && len(s.matches) > 0 {
		// Split into query part and count part
		countText := fmt.Sprintf(" (%d/%d matches)", s.CurrentIndex(), len(s.matches))
		queryText := "Search: " + s.query

		// Draw query text
		x := 0
		for _, r := range queryText {
			if x >= width-1 {
				break
			}
			buf.SetCell(x, y, buffer.Cell{Rune: r, Width: 1, Fg: promptStyle.Fg, Bg: promptStyle.Bg})
			x++
		}

		// Draw count in accent color
		countStyle := buffer.Style{
			Fg:    t.SearchMatch,
			Bg:    t.SearchBarBg,
			Flags: buffer.Bold,
		}
		for _, r := range countText {
			if x >= width-1 {
				break
			}
			buf.SetCell(x, y, buffer.Cell{Rune: r, Width: 1, Fg: countStyle.Fg, Bg: countStyle.Bg})
			x++
		}
		return
	}

	// No matches or empty query — draw in single style
	x := 0
	noMatchStyle := promptStyle
	if s.query != "" && len(s.matches) == 0 {
		noMatchStyle = buffer.Style{Fg: t.SearchNoMatch, Bg: t.SearchBarBg}
	}
	for _, r := range status {
		if x >= width-1 {
			break
		}
		buf.SetCell(x, y, buffer.Cell{Rune: r, Width: 1, Fg: noMatchStyle.Fg, Bg: noMatchStyle.Bg})
		x++
	}
}

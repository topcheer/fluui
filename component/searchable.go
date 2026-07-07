package component

import (
	"sort"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// FilterResult represents a single filtered item with match metadata.
type FilterResult struct {
	// Index is the original position in the input slice.
	Index int

	// Text is the original item text.
	Text string

	// Score ranks the match quality (higher = better).
	// Exact substring matches score highest, prefix matches next,
	// then substring, then case-insensitive substring.
	Score int

	// MatchStart is the byte offset of the match start within Text.
	MatchStart int

	// MatchEnd is the byte offset just past the match end within Text.
	MatchEnd int
}

// SearchFilter provides incremental, case-insensitive filtering for
// arbitrary string slices. It is designed to be embedded in components
// like Table, List, or any data-driven widget that needs live filtering.
//
// SearchFilter is safe for concurrent use.
type SearchFilter struct {
	mu            sync.RWMutex
	query         string
	caseSensitive bool
}

// NewSearchFilter creates a SearchFilter with an empty query
// (case-insensitive by default).
func NewSearchFilter() *SearchFilter {
	return &SearchFilter{}
}

// SetCaseSensitive toggles case-sensitive matching.
func (sf *SearchFilter) SetCaseSensitive(cs bool) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.caseSensitive = cs
}

// CaseSensitive reports whether matching is case-sensitive.
func (sf *SearchFilter) CaseSensitive() bool {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.caseSensitive
}

// SetQuery updates the filter query.
func (sf *SearchFilter) SetQuery(q string) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.query = q
}

// Query returns the current filter query.
func (sf *SearchFilter) Query() string {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.query
}

// IsActive reports whether a non-empty query is set.
func (sf *SearchFilter) IsActive() bool {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.query != ""
}

// MatchRange describes a text range within a cell buffer that should be
// highlighted as a search match.
type MatchRange struct {
	X      int // start column
	Y      int // row
	Length int // number of cells to highlight
}

// Filter applies the current query to the given items and returns
// matching results sorted by score (descending), then by original index.
//
// With an empty query, all items are returned unfiltered (score 0,
// full-range match).
//
// Scoring:
//   - Exact match: 1000
//   - Prefix match: 500
//   - Substring match: 100
//   - Case-insensitive substring match: 50
func (sf *SearchFilter) Filter(items []string) []FilterResult {
	sf.mu.RLock()
	query := sf.query
	caseSensitive := sf.caseSensitive
	sf.mu.RUnlock()

	if query == "" {
		results := make([]FilterResult, len(items))
		for i, s := range items {
			results[i] = FilterResult{
				Index:      i,
				Text:       s,
				Score:      0,
				MatchStart: 0,
				MatchEnd:   len(s),
			}
		}
		return results
	}

	var results []FilterResult
	cmpQuery := query
	if !caseSensitive {
		cmpQuery = strings.ToLower(query)
	}

	for i, text := range items {
		start, score := scoreMatch(text, cmpQuery, caseSensitive)
		if score > 0 {
			results = append(results, FilterResult{
				Index:      i,
				Text:       text,
				Score:      score,
				MatchStart: start,
				MatchEnd:   start + len(query),
			})
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Index < results[j].Index
	})

	return results
}

// FilterIndices is a convenience that returns just the original indices
// of matching items, preserving Filter's sort order.
func (sf *SearchFilter) FilterIndices(items []string) []int {
	results := sf.Filter(items)
	indices := make([]int, len(results))
	for i, r := range results {
		indices[i] = r.Index
	}
	return indices
}

// HighlightSegments splits text into segments based on a FilterResult,
// marking the matched portion for rendering with a distinct style.
type HighlightSegment struct {
	Text    string
	Matched bool
}

// Segments returns the text split into matched and non-matched portions
// according to the FilterResult's MatchStart/MatchEnd.
func (r FilterResult) Segments() []HighlightSegment {
	if r.MatchStart == r.MatchEnd || r.MatchStart < 0 || r.MatchEnd > len(r.Text) {
		return []HighlightSegment{{Text: r.Text, Matched: false}}
	}
	var segs []HighlightSegment
	if r.MatchStart > 0 {
		segs = append(segs, HighlightSegment{Text: r.Text[:r.MatchStart], Matched: false})
	}
	segs = append(segs, HighlightSegment{Text: r.Text[r.MatchStart:r.MatchEnd], Matched: true})
	if r.MatchEnd < len(r.Text) {
		segs = append(segs, HighlightSegment{Text: r.Text[r.MatchEnd:], Matched: false})
	}
	return segs
}

// PaintHighlight renders a row of text into the buffer, applying the
// given highlight style to the matched portion of the FilterResult.
// This is the bridge between filtering and visual rendering.
func PaintHighlight(buf *buffer.Buffer, x, y int, result FilterResult, normalStyle, matchStyle buffer.Style) int {
	curX := x
	for _, seg := range result.Segments() {
		style := normalStyle
		if seg.Matched {
			style = matchStyle
		}
		for _, r := range seg.Text {
			if curX >= buf.Width {
				return curX
			}
			w := buffer.RuneWidth(r)
			buf.SetCell(curX, y, buffer.Cell{
				Rune:  r,
				Width: uint8(w),
				Fg:    style.Fg,
				Bg:    style.Bg,
				Flags: style.Flags,
			})
			if w == 0 {
				// Combining char: don't advance cursor
				continue
			}
			curX++
		}
	}
	return curX
}

// scoreMatch finds the best match position for query in text and returns
// the byte offset and a score. Returns (0, 0) if no match.
func scoreMatch(text, query string, caseSensitive bool) (int, int) {
	if caseSensitive {
		// Exact match
		if text == query {
			return 0, 1000
		}
		// Prefix
		if strings.HasPrefix(text, query) {
			return 0, 500
		}
		// Substring
		idx := strings.Index(text, query)
		if idx >= 0 {
			return idx, 100
		}
		return 0, 0
	}

	// Case-insensitive
	lowerText := strings.ToLower(text)

	// Exact (case-insensitive)
	if lowerText == query {
		return 0, 1000
	}
	// Prefix (case-insensitive)
	if strings.HasPrefix(lowerText, query) {
		return 0, 500
	}
	// Substring (case-insensitive)
	idx := strings.Index(lowerText, query)
	if idx >= 0 {
		return idx, 50
	}
	return 0, 0
}

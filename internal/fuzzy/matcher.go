// Package fuzzy provides subsequence-based fuzzy string matching with scoring.
//
// The matcher uses a scoring algorithm that rewards:
//   - Exact substring matches (highest score)
//   - Consecutive character matches (compounding bonus)
//   - Start-of-word matches (word boundary bonus)
//   - Early position matches (proximity to start)
//   - Shorter target strings (density bonus)
//
// It returns match positions for highlighting matched characters in the UI.
package fuzzy

import (
	"sort"
	"strings"
	"sync"
	"unicode"
)

// Result represents a single fuzzy match result.
type Result struct {
	// Item is the original matched string.
	Item string

	// Score is the match quality (higher = better).
	Score float64

	// Positions are the rune indices in Item that matched the query.
	// Used for highlighting matched characters in the UI.
	Positions []int

	// OriginalIndex is the position of this item in the input candidate slice.
	OriginalIndex int
}

// Segment represents a portion of text for highlight rendering.
type Segment struct {
	Text    string
	Matched bool
}

// Matcher is a reusable fuzzy matching engine with configurable scoring.
// It is safe for concurrent use.
type Matcher struct {
	mu            sync.RWMutex
	caseSensitive bool

	// Scoring weights
	wSubstring   float64
	wStartOfWord float64
	wConsecutive float64
	wPosition    float64
	wDensity     float64
	wCaseMatch   float64
}

// NewMatcher creates a Matcher with default weights (case-insensitive).
func NewMatcher() *Matcher {
	return &Matcher{
		wSubstring:   20,
		wStartOfWord: 10,
		wConsecutive: 5,
		wPosition:    2,
		wDensity:     1,
		wCaseMatch:   2,
	}
}

// SetCaseSensitive controls whether matching respects letter case.
// Returns the matcher for chaining.
func (m *Matcher) SetCaseSensitive(v bool) *Matcher {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.caseSensitive = v
	return m
}

// CaseSensitive returns whether the matcher is case-sensitive.
func (m *Matcher) CaseSensitive() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.caseSensitive
}

// Match checks if query is a fuzzy subsequence match of target, returning
// the score and matched character positions. Returns nil if no match.
// An empty query matches everything with score 0.
func (m *Matcher) Match(query, target string) *Result {
	if query == "" {
		return &Result{Item: target, Score: 0, OriginalIndex: -1}
	}

	m.mu.RLock()
	cs := m.caseSensitive
	wSub := m.wSubstring
	wSow := m.wStartOfWord
	wCon := m.wConsecutive
	wPos := m.wPosition
	wDen := m.wDensity
	wCase := m.wCaseMatch
	m.mu.RUnlock()

	qRunes := []rune(query)
	tRunes := []rune(target)

	// Normalize for comparison
	qNorm := make([]rune, len(qRunes))
	tNorm := make([]rune, len(tRunes))
	for i, r := range qRunes {
		if cs {
			qNorm[i] = r
		} else {
			qNorm[i] = unicode.ToLower(r)
		}
	}
	for i, r := range tRunes {
		if cs {
			tNorm[i] = r
		} else {
			tNorm[i] = unicode.ToLower(r)
		}
	}

	// Exact substring check — big bonus
	subStr := string(qNorm)
	tgtStr := string(tNorm)
	subIdx := strings.Index(tgtStr, subStr)
	if subIdx >= 0 {
		charIdx := utf8RuneIndex(tgtStr, subIdx)
		positions := make([]int, len(qRunes))
		for i := range qRunes {
			positions[i] = charIdx + i
		}
		score := wSub + float64(len(qRunes))
		if charIdx == 0 {
			score += 10
		}
		// Word boundary bonus
		if charIdx == 0 || isBoundary(tRunes[charIdx-1]) {
			score += wSow
		}
		// Exact full-string match bonus
		if len(qRunes) == len(tRunes) {
			score += 5
		}
		// Density bonus: shorter target with same match = higher density.
		score += wDen * float64(len(qRunes)) / float64(len(tRunes))
		return &Result{Item: target, Score: score, Positions: positions, OriginalIndex: -1}
	}

	// Subsequence matching with scoring
	positions := make([]int, 0, len(qRunes))
	score := 0.0
	qi := 0
	consecutive := 0

	for ti := 0; ti < len(tRunes) && qi < len(qRunes); ti++ {
		if tNorm[ti] == qNorm[qi] {
			positions = append(positions, ti)
			score++

			// Consecutive bonus
			if qi > 0 && positions[qi-1] == ti-1 {
				consecutive++
				score += float64(consecutive) * wCon
			} else {
				consecutive = 0
			}

			// Word boundary bonus
			if ti == 0 || isBoundary(tRunes[ti-1]) {
				score += wSow
			}

			// Position bonus (earlier = better)
			score += wPos / float64(ti+1)

			// Exact case match bonus
			if qRunes[qi] == tRunes[ti] {
				score += wCase
			}

			qi++
		} else {
			consecutive = 0
		}
	}

	if qi < len(qRunes) {
		return nil
	}

	// Density bonus
	score += wDen * float64(len(qRunes)) / float64(len(tRunes))

	return &Result{Item: target, Score: score, Positions: positions, OriginalIndex: -1}
}

// Score returns just the match score, or -1 if no match.
func (m *Matcher) Score(query, target string) float64 {
	r := m.Match(query, target)
	if r == nil {
		return -1
	}
	return r.Score
}

// Rank matches query against a candidate list and returns sorted results.
// Non-matching candidates are excluded. Results are sorted by score desc,
// then alphabetically.
func (m *Matcher) Rank(query string, candidates []string) []Result {
	results := make([]Result, 0, len(candidates))
	for i, c := range candidates {
		r := m.Match(query, c)
		if r != nil {
			r.OriginalIndex = i
			results = append(results, *r)
		}
	}
	sortResults(results)
	return results
}

// RankTopN returns at most n top results.
func (m *Matcher) RankTopN(query string, candidates []string, n int) []Result {
	results := m.Rank(query, candidates)
	if len(results) > n {
		results = results[:n]
	}
	return results
}

// Filter returns only matching candidates (unsorted).
func (m *Matcher) Filter(query string, candidates []string) []string {
	matched := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if m.Match(query, c) != nil {
			matched = append(matched, c)
		}
	}
	return matched
}

// IsMatch returns true if query is a fuzzy match of target.
// Faster than Match (no scoring).
func (m *Matcher) IsMatch(query, target string) bool {
	if query == "" {
		return true
	}
	m.mu.RLock()
	cs := m.caseSensitive
	m.mu.RUnlock()

	if !cs {
		query = strings.ToLower(query)
		target = strings.ToLower(target)
	}

	qi := 0
	for ti := 0; ti < len(target) && qi < len(query); ti++ {
		if target[ti] == query[qi] {
			qi++
		}
	}
	return qi == len(query)
}

// Highlight splits the result into matched/unmatched segments for UI rendering.
func (r *Result) Highlight() []Segment {
	if r == nil || len(r.Positions) == 0 {
		if r != nil {
			return []Segment{{Text: r.Item, Matched: false}}
		}
		return nil
	}

	runes := []rune(r.Item)
	posSet := make(map[int]bool, len(r.Positions))
	for _, p := range r.Positions {
		posSet[p] = true
	}

	segments := make([]Segment, 0, len(r.Positions)+1)
	var buf strings.Builder
	currentMatched := false

	for i, r := range runes {
		matched := posSet[i]
		if i == 0 {
			currentMatched = matched
			buf.WriteRune(r)
			continue
		}
		if matched != currentMatched {
			segments = append(segments, Segment{Text: buf.String(), Matched: currentMatched})
			buf.Reset()
			currentMatched = matched
		}
		buf.WriteRune(r)
	}
	if buf.Len() > 0 {
		segments = append(segments, Segment{Text: buf.String(), Matched: currentMatched})
	}
	return segments
}

// --- Package-level convenience functions (use default Matcher) ---

var defaultMatcher = NewMatcher()

// Score is a package-level convenience for scoring a single match.
// Returns -1 if no match.
func Score(query, target string) float64 {
	return defaultMatcher.Score(query, target)
}

// Matches checks if query fuzzy-matches target.
func Matches(query, target string) bool {
	return defaultMatcher.IsMatch(query, target)
}

// BestMatch returns the highest-scoring match from candidates, or nil.
func BestMatch(candidates []string, query string) *Result {
	results := defaultMatcher.Rank(query, candidates)
	if len(results) == 0 {
		return nil
	}
	return &results[0]
}

// TopN returns the top N ranked matches.
func TopN(candidates []string, query string, n int) []Result {
	return defaultMatcher.RankTopN(query, candidates, n)
}

// IsAlnum reports whether r is an ASCII alphanumeric character.
func IsAlnum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// --- Helpers ---

// isBoundary returns true if r is a word separator.
func isBoundary(r rune) bool {
	return r == ' ' || r == '_' || r == '-' || r == '/' || r == '.'
}

// utf8RuneIndex converts a byte offset to a rune index.
func utf8RuneIndex(s string, byteOff int) int {
	return len([]rune(s[:byteOff]))
}

// sortResults sorts by score descending, then alphabetically.
func sortResults(results []Result) {
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Item < results[j].Item
	})
}

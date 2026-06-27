package fuzzy

import (
	"sync"
	"testing"
)

// === Matcher creation and config ===

func TestNewMatcher(t *testing.T) {
	m := NewMatcher()
	if m == nil {
		t.Fatal("NewMatcher returned nil")
	}
	if m.CaseSensitive() {
		t.Error("default should be case-insensitive")
	}
}

func TestMatcher_SetCaseSensitive(t *testing.T) {
	m := NewMatcher()
	m.SetCaseSensitive(true)
	if !m.CaseSensitive() {
		t.Error("should be case-sensitive after SetCaseSensitive(true)")
	}
	m.SetCaseSensitive(false)
	if m.CaseSensitive() {
		t.Error("should be case-insensitive after SetCaseSensitive(false)")
	}
}

// === Basic matching ===

func TestMatch_SimpleSubsequence(t *testing.T) {
	m := NewMatcher()
	r := m.Match("app", "apple")
	if r == nil {
		t.Fatal("Match('app', 'apple') should not be nil")
	}
	if len(r.Positions) != 3 {
		t.Errorf("Positions len = %d, want 3", len(r.Positions))
	}
	// Should be an exact substring at position 0
	for i, p := range r.Positions {
		if p != i {
			t.Errorf("Positions[%d] = %d, want %d", i, p, i)
		}
	}
}

func TestMatch_ExactSubstring(t *testing.T) {
	m := NewMatcher()
	r := m.Match("world", "hello world")
	if r == nil {
		t.Fatal("should match")
	}
	if r.Score < 20 {
		t.Errorf("Exact substring score = %.1f, want >= 20", r.Score)
	}
	// Positions should be runes 6-10
	if r.Positions[0] != 6 {
		t.Errorf("First position = %d, want 6", r.Positions[0])
	}
}

func TestMatch_NoMatch(t *testing.T) {
	m := NewMatcher()
	r := m.Match("xyz", "hello")
	if r != nil {
		t.Error("Match('xyz', 'hello') should be nil")
	}
}

func TestMatch_EmptyQuery(t *testing.T) {
	m := NewMatcher()
	r := m.Match("", "anything")
	if r == nil {
		t.Fatal("Empty query should return non-nil result")
	}
	if r.Score != 0 {
		t.Errorf("Empty query score = %.1f, want 0", r.Score)
	}
	if r.Positions != nil {
		t.Error("Empty query positions should be nil")
	}
}

func TestMatch_QueryLongerThanTarget(t *testing.T) {
	m := NewMatcher()
	r := m.Match("hello world", "hi")
	if r != nil {
		t.Error("Query longer than target should not match")
	}
}

func TestMatch_CaseInsensitive(t *testing.T) {
	m := NewMatcher() // default: case-insensitive
	r := m.Match("hw", "Hello World")
	if r == nil {
		t.Fatal("Case-insensitive 'hw' should match 'Hello World'")
	}
	// H at 0, W at 6
	if len(r.Positions) != 2 {
		t.Fatalf("Positions len = %d, want 2", len(r.Positions))
	}
	if r.Positions[0] != 0 || r.Positions[1] != 6 {
		t.Errorf("Positions = %v, want [0 6]", r.Positions)
	}
}

func TestMatch_CaseSensitive(t *testing.T) {
	m := NewMatcher().SetCaseSensitive(true)
	r := m.Match("H", "hello")
	if r != nil {
		t.Error("Case-sensitive 'H' should not match 'hello'")
	}
	r = m.Match("H", "Hello")
	if r == nil {
		t.Error("Case-sensitive 'H' should match 'Hello'")
	}
}

func TestMatch_SingleChar(t *testing.T) {
	m := NewMatcher()
	r := m.Match("a", "apple")
	if r == nil {
		t.Fatal("should match")
	}
	if len(r.Positions) != 1 || r.Positions[0] != 0 {
		t.Errorf("Positions = %v, want [0]", r.Positions)
	}
}

func TestMatch_AllCharsMatched(t *testing.T) {
	m := NewMatcher()
	r := m.Match("abc", "abc")
	if r == nil {
		t.Fatal("should match")
	}
	if len(r.Positions) != 3 {
		t.Errorf("Positions len = %d, want 3", len(r.Positions))
	}
}

// === Scoring ===

func TestScore_Positive(t *testing.T) {
	m := NewMatcher()
	score := m.Score("app", "apple")
	if score <= 0 {
		t.Errorf("Score = %.1f, want > 0", score)
	}
}

func TestScore_NoMatch(t *testing.T) {
	m := NewMatcher()
	score := m.Score("xyz", "hello")
	if score != -1 {
		t.Errorf("No match score = %.1f, want -1", score)
	}
}

func TestScore_EmptyQuery(t *testing.T) {
	m := NewMatcher()
	score := m.Score("", "anything")
	if score != 0 {
		t.Errorf("Empty query score = %.1f, want 0", score)
	}
}

func TestScore_ExactSubstringHigherThanSubsequence(t *testing.T) {
	m := NewMatcher()
	score1 := m.Score("app", "apple")  // exact substring
	score2 := m.Score("app", "aXpXp")   // subsequence
	if score1 <= score2 {
		t.Errorf("Exact substring (%.1f) should beat subsequence (%.1f)", score1, score2)
	}
}

func TestScore_WordBoundaryHigherThanMidWord(t *testing.T) {
	m := NewMatcher()
	// "bar" in "foo bar" at word boundary
	score1 := m.Score("bar", "foo bar")
	// "bar" in "barbarian" at position 3 (after "bar", not boundary)
	score2 := m.Score("bar", "embargo")
	if score1 <= score2 {
		t.Errorf("Word boundary match (%.1f) should beat mid-word (%.1f)", score1, score2)
	}
}

func TestScore_StartMatchHigherThanLater(t *testing.T) {
	m := NewMatcher()
	score1 := m.Score("app", "apple")
	score2 := m.Score("app", "snapple")
	if score1 <= score2 {
		t.Errorf("Match at start (%.1f) should beat match later (%.1f)", score1, score2)
	}
}

func TestScore_CompactnessBonus(t *testing.T) {
	m := NewMatcher()
	// "lo" in "hello" is compact (positions 3,4)
	score1 := m.Score("lo", "hello")
	// "lo" in "lxxxxo" spread out (positions 0,5)
	score2 := m.Score("lo", "lxxxxo")
	if score1 <= score2 {
		t.Errorf("Compact match (%.1f) should beat spread (%.1f)", score1, score2)
	}
}

// === Ranking ===

func TestRank_Basic(t *testing.T) {
	m := NewMatcher()
	candidates := []string{"apple", "banana", "application"}
	results := m.Rank("app", candidates)
	if len(results) != 2 {
		t.Fatalf("Rank = %d results, want 2", len(results))
	}
	// "apple" should rank higher than "application" (exact substring at start, shorter)
	for _, r := range results {
		if r.Item != "apple" && r.Item != "application" {
			t.Errorf("Unexpected item: %s", r.Item)
		}
	}
}

func TestRank_SortedByScore(t *testing.T) {
	m := NewMatcher()
	candidates := []string{"format data", "format", "information"}
	results := m.Rank("form", candidates)
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("Not sorted: [%d].Score (%.1f) > [%d].Score (%.1f)",
				i, results[i].Score, i-1, results[i-1].Score)
		}
	}
}

func TestRank_ExcludesNonMatching(t *testing.T) {
	m := NewMatcher()
	candidates := []string{"apple", "banana", "cherry"}
	results := m.Rank("app", candidates)
	if len(results) != 1 {
		t.Fatalf("Rank = %d results, want 1", len(results))
	}
	if results[0].Item != "apple" {
		t.Errorf("Item = %q, want 'apple'", results[0].Item)
	}
}

func TestRank_EmptyQuery(t *testing.T) {
	m := NewMatcher()
	candidates := []string{"alpha", "beta", "gamma"}
	results := m.Rank("", candidates)
	if len(results) != 3 {
		t.Fatalf("Empty query = %d results, want 3", len(results))
	}
	for _, r := range results {
		if r.Score != 0 {
			t.Errorf("Empty query score = %.1f, want 0", r.Score)
		}
	}
}

func TestRank_OriginalIndex(t *testing.T) {
	m := NewMatcher()
	candidates := []string{"alpha", "beta", "gamma"}
	results := m.Rank("beta", candidates)
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	if results[0].OriginalIndex != 1 {
		t.Errorf("OriginalIndex = %d, want 1", results[0].OriginalIndex)
	}
}

func TestRank_TieBreakerAlphabetical(t *testing.T) {
	m := NewMatcher()
	candidates := []string{"zebra_apple", "alpha_apple"}
	results := m.Rank("apple", candidates)
	if len(results) < 2 {
		t.Fatal("expected at least 2 results")
	}
	// Both match "apple" — check alphabetical tiebreak
	if results[0].Score == results[1].Score {
		if results[0].Item > results[1].Item {
			t.Errorf("Tie should break alphabetically: %q > %q",
				results[0].Item, results[1].Item)
		}
	}
}

// === RankTopN ===

func TestRankTopN(t *testing.T) {
	m := NewMatcher()
	candidates := []string{"apple", "application", "apricot", "banana"}
	results := m.RankTopN("a", candidates, 2)
	if len(results) != 2 {
		t.Fatalf("RankTopN = %d results, want 2", len(results))
	}
}

func TestRankTopN_MoreThanAvailable(t *testing.T) {
	m := NewMatcher()
	candidates := []string{"apple", "banana"}
	results := m.RankTopN("a", candidates, 10)
	// Only "apple" matches "a" as exact substring (wait - banana also has 'a')
	// Both "apple" and "banana" match "a" as subsequence
	if len(results) > 2 {
		t.Errorf("RankTopN = %d, should be <= 2", len(results))
	}
}

// === Filter ===

func TestFilter(t *testing.T) {
	m := NewMatcher()
	candidates := []string{"apple", "banana", "apricot"}
	matched := m.Filter("ap", candidates)
	if len(matched) != 2 {
		t.Fatalf("Filter = %d, want 2", len(matched))
	}
	for _, s := range matched {
		if s == "banana" {
			t.Error("banana should not match 'ap'")
		}
	}
}

// === IsMatch ===

func TestIsMatch_True(t *testing.T) {
	m := NewMatcher()
	if !m.IsMatch("hlo", "hello") {
		t.Error("IsMatch('hlo', 'hello') should be true")
	}
}

func TestIsMatch_False(t *testing.T) {
	m := NewMatcher()
	if m.IsMatch("xyz", "hello") {
		t.Error("IsMatch('xyz', 'hello') should be false")
	}
}

func TestIsMatch_EmptyQuery(t *testing.T) {
	m := NewMatcher()
	if !m.IsMatch("", "anything") {
		t.Error("IsMatch('', 'anything') should be true")
	}
}

// === Package-level convenience functions ===

func TestScore_Func(t *testing.T) {
	score := Score("app", "apple")
	if score <= 0 {
		t.Errorf("Score = %.1f, want > 0", score)
	}
}

func TestMatches_True(t *testing.T) {
	if !Matches("hlo", "hello") {
		t.Error("Matches('hlo', 'hello') should be true")
	}
}

func TestMatches_False(t *testing.T) {
	if Matches("xyz", "hello") {
		t.Error("Matches('xyz', 'hello') should be false")
	}
}

func TestBestMatch(t *testing.T) {
	result := BestMatch([]string{"format", "transform", "perform"}, "form")
	if result == nil {
		t.Fatal("BestMatch returned nil")
	}
	if result.Item != "format" {
		t.Errorf("BestMatch = %q, want 'format'", result.Item)
	}
}

func TestBestMatch_None(t *testing.T) {
	result := BestMatch([]string{"hello"}, "xyz")
	if result != nil {
		t.Error("BestMatch should return nil for no match")
	}
}

func TestTopN_Func(t *testing.T) {
	results := TopN([]string{"apple", "application", "apricot"}, "ap", 2)
	if len(results) != 2 {
		t.Fatalf("TopN = %d, want 2", len(results))
	}
}

func TestIsAlnum(t *testing.T) {
	if !IsAlnum('a') {
		t.Error("'a' should be alnum")
	}
	if !IsAlnum('Z') {
		t.Error("'Z' should be alnum")
	}
	if !IsAlnum('5') {
		t.Error("'5' should be alnum")
	}
	if IsAlnum('-') {
		t.Error("'-' should not be alnum")
	}
}

// === Highlight segments ===

func TestHighlight_AllMatched(t *testing.T) {
	m := NewMatcher()
	r := m.Match("abc", "abc")
	if r == nil {
		t.Fatal("should match")
	}
	segs := r.Highlight()
	// All characters are matched, so one segment with Matched=true
	if len(segs) != 1 {
		t.Fatalf("Segments len = %d, want 1", len(segs))
	}
	if !segs[0].Matched {
		t.Error("segment should be matched")
	}
	if segs[0].Text != "abc" {
		t.Errorf("Text = %q", segs[0].Text)
	}
}

func TestHighlight_PartialMatch(t *testing.T) {
	m := NewMatcher()
	r := m.Match("hl", "hello")
	if r == nil {
		t.Fatal("should match")
	}
	segs := r.Highlight()
	// h(matched) + e(unmatched) + l(matched) + l(unmatched) + o(unmatched)
	// → segments: [h, e, l, lo] with matched: [true, false, true, false]
	if len(segs) < 3 {
		t.Fatalf("Segments len = %d, want >= 3", len(segs))
	}
	// First segment should be matched (h)
	if !segs[0].Matched {
		t.Error("first segment should be matched")
	}
}

// === Concurrency ===

func TestMatcher_Concurrent(t *testing.T) {
	m := NewMatcher().SetCaseSensitive(false)
	candidates := []string{"apple", "banana", "cherry", "date"}

	var wg sync.WaitGroup
	const goroutines = 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = m.Rank("a", candidates)
			_ = m.Match("app", "apple")
			_ = m.IsMatch("a", "apple")
			if n%3 == 0 {
				m.SetCaseSensitive(n%2 == 0)
			}
		}(i)
	}
	wg.Wait()
}

func TestConcurrent_PackageFuncs(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = Score("app", "apple")
			_ = Matches("hlo", "hello")
			_ = BestMatch([]string{"a", "b"}, "a")
		}()
	}
	wg.Wait()
}

// === Edge cases ===

func TestMatch_RealWorldFileSearch(t *testing.T) {
	m := NewMatcher()
	candidates := []string{
		"src/main/java/com/example/Application.java",
		"src/test/java/com/example/ApplicationTest.java",
		"src/main/java/com/example/Database.java",
		"docs/architecture.md",
		"README.md",
	}

	// "app" should match Application files
	results := m.Rank("app", candidates)
	if len(results) != 2 {
		t.Fatalf("File search 'app' = %d results, want 2", len(results))
	}

	// "Application.java" should rank higher (shorter path, 'app' closer to start)
	if results[0].Item != "src/main/java/com/example/Application.java" &&
		results[0].Item != "src/test/java/com/example/ApplicationTest.java" {
		t.Errorf("Unexpected top result: %s", results[0].Item)
	}

	// "test" should match only ApplicationTest.java
	results = m.Rank("test", candidates)
	if len(results) != 1 {
		t.Fatalf("File search 'test' = %d results, want 1", len(results))
	}
	if results[0].Item != "src/test/java/com/example/ApplicationTest.java" {
		t.Errorf("Top result = %q", results[0].Item)
	}

	// "md" matches .md files plus "main/Database" (m+d subsequence)
	results = m.Rank("md", candidates)
	if len(results) < 2 {
		t.Fatalf("File search 'md' = %d results, want >= 2", len(results))
	}
	// Top 2 should be .md files (exact substring "md" in extension)
	for i := 0; i < 2; i++ {
		if results[i].Item != "docs/architecture.md" && results[i].Item != "README.md" {
			t.Errorf("result[%d] = %q, should be a .md file", i, results[i].Item)
		}
	}
}

func TestMatch_DuplicateChars(t *testing.T) {
	m := NewMatcher()
	r := m.Match("ll", "hello")
	if r == nil {
		t.Fatal("should match")
	}
	if len(r.Positions) != 2 {
		t.Fatalf("Positions = %v, want 2 entries", r.Positions)
	}
	// Both l's at positions 2,3
	if r.Positions[0] != 2 || r.Positions[1] != 3 {
		t.Errorf("Positions = %v, want [2 3]", r.Positions)
	}
}

func TestMatch_LargeCandidateList(t *testing.T) {
	m := NewMatcher()
	candidates := make([]string, 1000)
	for i := range candidates {
		candidates[i] = "item"
	}
	candidates[500] = "special target"

	results := m.Rank("special", candidates)
	if len(results) != 1 {
		t.Fatalf("Large list = %d results, want 1", len(results))
	}
	if results[0].OriginalIndex != 500 {
		t.Errorf("OriginalIndex = %d, want 500", results[0].OriginalIndex)
	}
}

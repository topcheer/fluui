package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// --- SearchFilter basics ---

func TestSearchFilter_EmptyQuery(t *testing.T) {
	sf := NewSearchFilter()
	items := []string{"apple", "banana", "cherry"}

	results := sf.Filter(items)
	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}
	for i, r := range results {
		if r.Index != i {
			t.Errorf("result[%d].Index = %d, want %d", i, r.Index, i)
		}
		if r.Score != 0 {
			t.Errorf("result[%d].Score = %d, want 0 (empty query)", i, r.Score)
		}
	}
}

func TestSearchFilter_SubstringMatch(t *testing.T) {
	sf := NewSearchFilter()
	sf.SetQuery("an")
	items := []string{"banana", "apple", "cherry", "mango"}

	results := sf.Filter(items)
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2 (banana, mango)", len(results))
	}
	// Both match as substrings with score 50. Stable sort by original index:
	// banana (index 0) comes before mango (index 3).
	if results[0].Text != "banana" {
		t.Errorf("results[0].Text = %q, want 'banana'", results[0].Text)
	}
	if results[1].Text != "mango" {
		t.Errorf("results[1].Text = %q, want 'mango'", results[1].Text)
	}
}

func TestSearchFilter_ExactMatch(t *testing.T) {
	sf := NewSearchFilter()
	sf.SetQuery("apple")
	items := []string{"apple", "snapple", "apples", "Apple"}

	results := sf.Filter(items)
	// "apple" exact (1000), "Apple" exact case-insensitive (1000),
	// "apples" prefix (500), "snapple" substring (50)
	if len(results) != 4 {
		t.Fatalf("len(results) = %d, want 4", len(results))
	}
	if results[0].Text != "apple" {
		t.Errorf("results[0] = %q, want 'apple' (exact match, score 1000)", results[0].Text)
	}
}

func TestSearchFilter_CaseInsensitive(t *testing.T) {
	sf := NewSearchFilter()
	sf.SetQuery("GO")
	items := []string{"golang", "Golang", "GOLANG", "python"}

	results := sf.Filter(items)
	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3 (golang, Golang, GOLANG)", len(results))
	}
}

func TestSearchFilter_CaseSensitive(t *testing.T) {
	sf := NewSearchFilter()
	sf.SetCaseSensitive(true)
	sf.SetQuery("GO")
	items := []string{"golang", "GOLANG", "Golang", "GO"}

	results := sf.Filter(items)
	// Case-sensitive "GO" matches: "GO" (exact, 1000) and "GOLANG" (prefix, 500)
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2 (GO and GOLANG)", len(results))
	}
	if results[0].Text != "GO" {
		t.Errorf("results[0].Text = %q, want 'GO' (exact match)", results[0].Text)
	}
}

func TestSearchFilter_NoMatches(t *testing.T) {
	sf := NewSearchFilter()
	sf.SetQuery("xyz")
	items := []string{"apple", "banana", "cherry"}

	results := sf.Filter(items)
	if len(results) != 0 {
		t.Fatalf("len(results) = %d, want 0", len(results))
	}
}

func TestSearchFilter_FilterIndices(t *testing.T) {
	sf := NewSearchFilter()
	sf.SetQuery("a")
	items := []string{"apple", "banana", "cherry"}

	indices := sf.FilterIndices(items)
	// "apple" (prefix, 500), "banana" (substring, 50), "cherry" no match
	if len(indices) != 2 {
		t.Fatalf("len(indices) = %d, want 2", len(indices))
	}
	if indices[0] != 0 || indices[1] != 1 {
		t.Errorf("indices = %v, want [0, 1]", indices)
	}
}

func TestSearchFilter_IsActive(t *testing.T) {
	sf := NewSearchFilter()
	if sf.IsActive() {
		t.Error("IsActive should be false with empty query")
	}
	sf.SetQuery("test")
	if !sf.IsActive() {
		t.Error("IsActive should be true with non-empty query")
	}
	sf.SetQuery("")
	if sf.IsActive() {
		t.Error("IsActive should be false after clearing query")
	}
}

func TestSearchFilter_QueryGetter(t *testing.T) {
	sf := NewSearchFilter()
	sf.SetQuery("hello")
	if sf.Query() != "hello" {
		t.Errorf("Query() = %q, want 'hello'", sf.Query())
	}
}

func TestSearchFilter_CaseSensitiveGetter(t *testing.T) {
	sf := NewSearchFilter()
	if sf.CaseSensitive() {
		t.Error("CaseSensitive should be false by default")
	}
	sf.SetCaseSensitive(true)
	if !sf.CaseSensitive() {
		t.Error("CaseSensitive should be true after SetCaseSensitive(true)")
	}
}

func TestSearchFilter_PrefixBeatsSubstring(t *testing.T) {
	sf := NewSearchFilter()
	sf.SetQuery("ba")
	items := []string{"banana", "cabana"}

	results := sf.Filter(items)
	// "banana" is prefix (500), "cabana" is substring (50)
	if results[0].Text != "banana" {
		t.Errorf("results[0] = %q, want 'banana' (prefix > substring)", results[0].Text)
	}
}

func TestSearchFilter_EmptyItems(t *testing.T) {
	sf := NewSearchFilter()
	sf.SetQuery("test")
	results := sf.Filter([]string{})
	if len(results) != 0 {
		t.Errorf("len(results) = %d, want 0", len(results))
	}
}

// --- FilterResult Segments ---

func TestFilterResult_Segments(t *testing.T) {
	r := FilterResult{
		Text:       "hello world",
		MatchStart: 6,
		MatchEnd:   11,
	}
	segs := r.Segments()
	if len(segs) != 2 {
		t.Fatalf("len(segs) = %d, want 2", len(segs))
	}
	if segs[0].Text != "hello " || segs[0].Matched {
		t.Errorf("segs[0] = %+v, want text='hello ' matched=false", segs[0])
	}
	if segs[1].Text != "world" || !segs[1].Matched {
		t.Errorf("segs[1] = %+v, want text='world' matched=true", segs[1])
	}
}

func TestFilterResult_SegmentsAtStart(t *testing.T) {
	r := FilterResult{
		Text:       "hello world",
		MatchStart: 0,
		MatchEnd:   5,
	}
	segs := r.Segments()
	if len(segs) != 2 {
		t.Fatalf("len(segs) = %d, want 2", len(segs))
	}
	if !segs[0].Matched || segs[0].Text != "hello" {
		t.Errorf("segs[0] = %+v, want matched 'hello'", segs[0])
	}
	if segs[1].Matched || segs[1].Text != " world" {
		t.Errorf("segs[1] = %+v, want non-matched ' world'", segs[1])
	}
}

func TestFilterResult_SegmentsFullMatch(t *testing.T) {
	r := FilterResult{
		Text:       "hello",
		MatchStart: 0,
		MatchEnd:   5,
	}
	segs := r.Segments()
	if len(segs) != 1 {
		t.Fatalf("len(segs) = %d, want 1", len(segs))
	}
	if !segs[0].Matched || segs[0].Text != "hello" {
		t.Errorf("segs[0] = %+v, want matched 'hello'", segs[0])
	}
}

func TestFilterResult_SegmentsNoMatch(t *testing.T) {
	r := FilterResult{
		Text:       "hello",
		MatchStart: 0,
		MatchEnd:   0,
	}
	segs := r.Segments()
	if len(segs) != 1 {
		t.Fatalf("len(segs) = %d, want 1", len(segs))
	}
	if segs[0].Matched {
		t.Error("segs[0] should not be matched when MatchStart==MatchEnd")
	}
}

// --- PaintHighlight ---

func TestPaintHighlight_Basic(t *testing.T) {
	buf := buffer.NewBuffer(30, 1)
	r := FilterResult{
		Text:       "hello world",
		MatchStart: 6,
		MatchEnd:   11,
	}
	normalStyle := buffer.Style{}
	matchStyle := buffer.Style{Flags: buffer.Reverse}

	PaintHighlight(buf, 0, 0, r, normalStyle, matchStyle)

	// Cells 6-10 should have Reverse flag
	for x := 6; x <= 10; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Flags&buffer.Reverse == 0 {
			t.Errorf("cell[%d] should have Reverse flag", x)
		}
	}
	// Cells 0-5 should NOT have Reverse
	for x := 0; x <= 5; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Flags&buffer.Reverse != 0 {
			t.Errorf("cell[%d] should NOT have Reverse flag", x)
		}
	}
}

func TestPaintHighlight_ClampToBuffer(t *testing.T) {
	buf := buffer.NewBuffer(5, 1)
	r := FilterResult{
		Text:       "hello world",
		MatchStart: 0,
		MatchEnd:   11,
	}
	PaintHighlight(buf, 0, 0, r, buffer.Style{}, buffer.Style{Flags: buffer.Bold})
	// Should not panic, should fill at most 5 chars
	for x := 0; x < 5; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Rune == 0 {
			t.Errorf("cell[%d] should have a rune", x)
		}
	}
}

// --- Concurrent access ---

func TestSearchFilter_ConcurrentAccess(t *testing.T) {
	sf := NewSearchFilter()
	items := []string{"alpha", "beta", "gamma", "delta", "epsilon"}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			sf.SetQuery("a")
			_ = sf.Filter(items)
		}()
		go func() {
			defer wg.Done()
			_ = sf.Query()
			_ = sf.IsActive()
		}()
	}
	wg.Wait()
}

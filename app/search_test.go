package app

import (
	"testing"

	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestSearchStartAndClose(t *testing.T) {
	s := NewSearchMode()

	// Initially inactive
	if s.IsActive() {
		t.Error("SearchMode should be inactive initially")
	}

	// Start search
	s.StartSearch()
	if !s.IsActive() {
		t.Error("SearchMode should be active after StartSearch")
	}
	if s.Query() != "" {
		t.Error("Query should be empty after StartSearch")
	}

	// Close search
	s.CloseSearch()
	if s.IsActive() {
		t.Error("SearchMode should be inactive after CloseSearch")
	}
	if s.MatchCount() != 0 {
		t.Error("MatchCount should be 0 after CloseSearch")
	}
}

func TestSearchFindsInUserMessages(t *testing.T) {
	s := NewSearchMode()
	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "Hello world"))
	c.AddBlock(block.NewUserMessageBlock("u2", "Goodbye moon"))

	s.StartSearch()
	s.UpdateQuery("hello", c.Blocks())

	if s.MatchCount() != 1 {
		t.Errorf("MatchCount = %d, want 1", s.MatchCount())
	}
	match := s.CurrentMatch()
	if match == nil {
		t.Fatal("CurrentMatch is nil")
	}
	if match.BlockID != "u1" {
		t.Errorf("BlockID = %s, want u1", match.BlockID)
	}
	if match.BlockType != "user_message" {
		t.Errorf("BlockType = %s, want user_message", match.BlockType)
	}
}

func TestSearchFindsInAssistantText(t *testing.T) {
	s := NewSearchMode()
	c := block.NewBlockContainer()
	c.AddBlock(block.NewAssistantTextBlock("a1"))
	atb := c.Blocks()[0].(*block.AssistantTextBlock)
	atb.AppendDelta("The quick brown fox")

	s.StartSearch()
	s.UpdateQuery("fox", c.Blocks())

	if s.MatchCount() != 1 {
		t.Errorf("MatchCount = %d, want 1", s.MatchCount())
	}
	match := s.CurrentMatch()
	if match.BlockID != "a1" {
		t.Errorf("BlockID = %s, want a1", match.BlockID)
	}
}

func TestSearchCaseInsensitive(t *testing.T) {
	s := NewSearchMode()
	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "Go is Awesome GO gO"))

	s.StartSearch()

	// Search with different cases
	for _, q := range []string{"go", "GO", "Go", "gO"} {
		s.UpdateQuery(q, c.Blocks())
		if s.MatchCount() != 3 {
			t.Errorf("Query %q: MatchCount = %d, want 3", q, s.MatchCount())
		}
	}
}

func TestSearchNoResults(t *testing.T) {
	s := NewSearchMode()
	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "Hello world"))

	s.StartSearch()
	s.UpdateQuery("nonexistent", c.Blocks())

	if s.MatchCount() != 0 {
		t.Errorf("MatchCount = %d, want 0", s.MatchCount())
	}
	if s.CurrentMatch() != nil {
		t.Error("CurrentMatch should be nil with no results")
	}
	if s.CurrentIndex() != 0 {
		t.Errorf("CurrentIndex = %d, want 0", s.CurrentIndex())
	}
}

func TestSearchNextPrevCycling(t *testing.T) {
	s := NewSearchMode()
	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "foo bar foo"))
	c.AddBlock(block.NewUserMessageBlock("u2", "foo baz"))

	s.StartSearch()
	s.UpdateQuery("foo", c.Blocks())

	if s.MatchCount() != 3 {
		t.Fatalf("MatchCount = %d, want 3", s.MatchCount())
	}

	// Start at match 0
	if s.CurrentIndex() != 1 {
		t.Errorf("CurrentIndex = %d, want 1 (first)", s.CurrentIndex())
	}

	// Next → match 1
	s.NextMatch()
	if s.CurrentIndex() != 2 {
		t.Errorf("After Next: CurrentIndex = %d, want 2", s.CurrentIndex())
	}

	// Next → match 2
	s.NextMatch()
	if s.CurrentIndex() != 3 {
		t.Errorf("After Next: CurrentIndex = %d, want 3", s.CurrentIndex())
	}

	// Next → wrap to match 0
	s.NextMatch()
	if s.CurrentIndex() != 1 {
		t.Errorf("After wrap: CurrentIndex = %d, want 1", s.CurrentIndex())
	}

	// Prev → wrap to last match
	s.PrevMatch()
	if s.CurrentIndex() != 3 {
		t.Errorf("After wrap Prev: CurrentIndex = %d, want 3", s.CurrentIndex())
	}

	// Prev → match 1
	s.PrevMatch()
	if s.CurrentIndex() != 2 {
		t.Errorf("After Prev: CurrentIndex = %d, want 2", s.CurrentIndex())
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	s := NewSearchMode()
	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "Hello world"))

	s.StartSearch()
	s.UpdateQuery("", c.Blocks())

	if s.MatchCount() != 0 {
		t.Errorf("Empty query MatchCount = %d, want 0", s.MatchCount())
	}
	if s.CurrentMatch() != nil {
		t.Error("CurrentMatch should be nil for empty query")
	}
}

func TestSearchAcrossMultipleBlockTypes(t *testing.T) {
	s := NewSearchMode()
	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "search here"))
	c.AddBlock(block.NewThinkingBlock("t1"))
	tb := c.Blocks()[1].(*block.ThinkingBlock)
	tb.AppendDelta("search thinking")
	c.AddBlock(block.NewAssistantTextBlock("a1"))
	atb := c.Blocks()[2].(*block.AssistantTextBlock)
	atb.AppendDelta("search answer")

	s.StartSearch()
	s.UpdateQuery("search", c.Blocks())

	if s.MatchCount() != 3 {
		t.Errorf("MatchCount = %d, want 3 (across 3 block types)", s.MatchCount())
	}

	// Verify matches span different block types
	types := make(map[string]bool)
	for _, m := range s.CurrentMatches() {
		types[m.BlockType] = true
	}
	if len(types) != 3 {
		t.Errorf("Expected 3 different block types, got %d: %v", len(types), types)
	}
}

func TestSearchHandleKey_Escape(t *testing.T) {
	s := NewSearchMode()
	s.StartSearch()

	key := &term.KeyEvent{Key: term.KeyEscape}
	consumed := s.HandleKey(key)
	if !consumed {
		t.Error("HandleKey should consume Escape")
	}
	if s.IsActive() {
		t.Error("Search should be inactive after Escape")
	}
}

func TestSearchHandleKey_Enter(t *testing.T) {
	s := NewSearchMode()
	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "foo foo foo"))

	s.StartSearch()
	s.UpdateQuery("foo", c.Blocks())

	if s.CurrentIndex() != 1 {
		t.Fatalf("Initial CurrentIndex = %d, want 1", s.CurrentIndex())
	}

	// Enter → next match
	key := &term.KeyEvent{Key: term.KeyEnter}
	s.HandleKey(key)
	if s.CurrentIndex() != 2 {
		t.Errorf("After Enter: CurrentIndex = %d, want 2", s.CurrentIndex())
	}
}

func TestSearchHandleKey_ShiftEnter(t *testing.T) {
	s := NewSearchMode()
	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "foo foo foo"))

	s.StartSearch()
	s.UpdateQuery("foo", c.Blocks())
	s.NextMatch() // move to match 2

	if s.CurrentIndex() != 2 {
		t.Fatalf("Initial CurrentIndex = %d, want 2", s.CurrentIndex())
	}

	// Shift+Enter → previous match
	key := &term.KeyEvent{Key: term.KeyEnter, Modifiers: term.ModShift}
	s.HandleKey(key)
	if s.CurrentIndex() != 1 {
		t.Errorf("After Shift+Enter: CurrentIndex = %d, want 1", s.CurrentIndex())
	}
}

func TestSearchHandleKey_Backspace(t *testing.T) {
	s := NewSearchMode()
	s.StartSearch()

	// Type "abc"
	for _, r := range "abc" {
		s.HandleKey(&term.KeyEvent{Rune: r})
	}
	if s.Query() != "abc" {
		t.Fatalf("Query = %q, want 'abc'", s.Query())
	}

	// Backspace → "ab"
	s.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if s.Query() != "ab" {
		t.Errorf("After backspace: Query = %q, want 'ab'", s.Query())
	}

	// Backspace → "a"
	s.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if s.Query() != "a" {
		t.Errorf("After backspace: Query = %q, want 'a'", s.Query())
	}

	// Backspace → ""
	s.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if s.Query() != "" {
		t.Errorf("After backspace: Query = %q, want ''", s.Query())
	}

	// Backspace on empty → stays empty (no panic)
	s.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if s.Query() != "" {
		t.Errorf("After backspace on empty: Query = %q, want ''", s.Query())
	}
}

func TestSearchHandleKey_Printable(t *testing.T) {
	s := NewSearchMode()
	s.StartSearch()

	s.HandleKey(&term.KeyEvent{Rune: 'h'})
	s.HandleKey(&term.KeyEvent{Rune: 'i'})

	if s.Query() != "hi" {
		t.Errorf("Query = %q, want 'hi'", s.Query())
	}
}

func TestSearchHandleKey_NotActive(t *testing.T) {
	s := NewSearchMode()
	// Don't start search
	consumed := s.HandleKey(&term.KeyEvent{Rune: 'a'})
	if consumed {
		t.Error("HandleKey should return false when not active")
	}
}

func TestSearchStatusText(t *testing.T) {
	s := NewSearchMode()

	// Inactive
	if s.StatusText() != "" {
		t.Errorf("Inactive StatusText = %q, want empty", s.StatusText())
	}

	// Active, empty query
	s.StartSearch()
	if s.StatusText() != "Search: " {
		t.Errorf("Empty query StatusText = %q, want 'Search: '", s.StatusText())
	}

	// Active, no matches
	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "Hello"))
	s.UpdateQuery("xyz", c.Blocks())
	status := s.StatusText()
	if status != "Search: xyz (no matches)" {
		t.Errorf("No-match StatusText = %q", status)
	}

	// Active, with matches
	s.UpdateQuery("hello", c.Blocks())
	status = s.StatusText()
	if status != "Search: hello (1/1 matches)" {
		t.Errorf("Match StatusText = %q, want 'Search: hello (1/1 matches)'", status)
	}
}

func TestSearchRenderSearchBar(t *testing.T) {
	s := NewSearchMode()
	s.StartSearch()

	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "hello"))
	s.UpdateQuery("hello", c.Blocks())

	buf := buffer.NewBuffer(60, 5)
	s.RenderSearchBar(buf, 60, 4)

	// Should not panic. Verify some cells were written on line 4.
	written := false
	for x := 0; x < 60; x++ {
		cell := buf.GetCell(x, 4)
		if cell.Rune != 0 && cell.Rune != ' ' {
			written = true
			break
		}
	}
	if !written {
		t.Error("RenderSearchBar wrote no visible characters")
	}
}

func TestSearchRenderSearchBar_Inactive(t *testing.T) {
	s := NewSearchMode()
	buf := buffer.NewBuffer(60, 5)
	// Should not panic when inactive
	s.RenderSearchBar(buf, 60, 0)
}

func TestSearchSnippet(t *testing.T) {
	s := NewSearchMode()
	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "short"))

	s.StartSearch()
	s.UpdateQuery("short", c.Blocks())

	if s.MatchCount() != 1 {
		t.Fatalf("MatchCount = %d, want 1", s.MatchCount())
	}

	match := s.CurrentMatch()
	if !contains(match.Snippet, "short") {
		t.Errorf("Snippet should contain 'short', got %q", match.Snippet)
	}
}

func TestSearchSnippetLong(t *testing.T) {
	s := NewSearchMode()
	c := block.NewBlockContainer()
	longText := "This is a very long text that has the search term somewhere in the middle of it"
	c.AddBlock(block.NewUserMessageBlock("u1", longText))

	s.StartSearch()
	s.UpdateQuery("search", c.Blocks())

	if s.MatchCount() != 1 {
		t.Fatalf("MatchCount = %d, want 1", s.MatchCount())
	}

	match := s.CurrentMatch()
	if !contains(match.Snippet, "search") {
		t.Errorf("Snippet should contain 'search', got %q", match.Snippet)
	}
	// Snippet should have ellipsis on both sides
	if match.Snippet[:3] != "..." {
		t.Errorf("Snippet should start with '...', got %q", match.Snippet[:3])
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (indexOf(s, substr) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

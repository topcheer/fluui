package app

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestCommandPaletteRegister(t *testing.T) {
	cp := NewCommandPalette()
	cp.Register(Command{ID: "test1", Title: "Test Command", Category: "Test"})

	if len(cp.Commands()) != 1 {
		t.Errorf("Commands() len = %d, want 1", len(cp.Commands()))
	}
	if cp.Commands()[0].ID != "test1" {
		t.Errorf("ID = %s, want test1", cp.Commands()[0].ID)
	}
}

func TestCommandPaletteRegisterMany(t *testing.T) {
	cp := NewCommandPalette()
	cp.RegisterMany([]Command{
		{ID: "a", Title: "A", Category: "X"},
		{ID: "b", Title: "B", Category: "X"},
		{ID: "c", Title: "C", Category: "X"},
	})

	if len(cp.Commands()) != 3 {
		t.Errorf("Commands() len = %d, want 3", len(cp.Commands()))
	}
}

func TestCommandPaletteOpenClose(t *testing.T) {
	cp := NewCommandPalette()

	if cp.IsActive() {
		t.Error("Should be inactive initially")
	}

	cp.Open()
	if !cp.IsActive() {
		t.Error("Should be active after Open")
	}
	if cp.Query() != "" {
		t.Error("Query should be empty after Open")
	}

	cp.Close()
	if cp.IsActive() {
		t.Error("Should be inactive after Close")
	}
}

func TestFuzzyMatchSubsequence(t *testing.T) {
	cp := NewCommandPalette()
	cp.RegisterMany([]Command{
		{ID: "1", Title: "Change Theme", Category: "Theme"},
		{ID: "2", Title: "Clear Conversation", Category: "Edit"},
		{ID: "3", Title: "Help", Category: "View"},
	})

	cp.Open()

	// "ct" should match "Change Theme" (subsequence)
	cp.query = "ct"
	cp.Filter()

	if cp.FilteredCount() == 0 {
		t.Error("'ct' should match at least 'Change Theme'")
	}

	found := false
	for _, cmd := range cp.FilteredCommands() {
		if cmd.Title == "Change Theme" {
			found = true
			break
		}
	}
	if !found {
		t.Error("'ct' should match 'Change Theme'")
	}
}

func TestFuzzyMatchCaseInsensitive(t *testing.T) {
	cp := NewCommandPalette()
	cp.Register(Command{ID: "1", Title: "Change Theme", Category: "Theme"})

	cp.Open()

	for _, q := range []string{"CT", "ct", "Ct", "cT"} {
		cp.query = q
		cp.Filter()
		if cp.FilteredCount() != 1 {
			t.Errorf("Query %q: FilteredCount = %d, want 1", q, cp.FilteredCount())
		}
	}
}

func TestFuzzyMatchNoResults(t *testing.T) {
	cp := NewCommandPalette()
	cp.Register(Command{ID: "1", Title: "Change Theme", Category: "Theme"})

	cp.Open()
	cp.query = "zzzzz"
	cp.Filter()

	if cp.FilteredCount() != 0 {
		t.Errorf("FilteredCount = %d, want 0 for no match", cp.FilteredCount())
	}
}

func TestFuzzyMatchEmptyQuery(t *testing.T) {
	cp := NewCommandPalette()
	cp.RegisterMany([]Command{
		{ID: "1", Title: "A", Category: "X"},
		{ID: "2", Title: "B", Category: "X"},
		{ID: "3", Title: "C", Category: "X"},
	})

	cp.Open()
	cp.query = ""
	cp.Filter()

	// Empty query should show all commands
	if cp.FilteredCount() != 3 {
		t.Errorf("Empty query FilteredCount = %d, want 3", cp.FilteredCount())
	}
}

func TestFuzzyScoreRanking(t *testing.T) {
	cp := NewCommandPalette()
	cp.RegisterMany([]Command{
		{ID: "1", Title: "Theme Changer", Category: "Theme"},  // subsequence match
		{ID: "2", Title: "Change Theme", Category: "Theme"},    // exact substring
		{ID: "3", Title: "Something Else", Category: "View"},   // no match
	})

	cp.Open()
	cp.query = "change"
	cp.Filter()

	// Both "Change Theme" and "Theme Changer" should match "change" as subsequence
	if cp.FilteredCount() != 2 {
		t.Fatalf("FilteredCount = %d, want 2", cp.FilteredCount())
	}

	// "Change Theme" should rank higher (exact substring at start)
	first := cp.FilteredCommands()[0]
	if first.Title != "Change Theme" {
		t.Errorf("Top result = %q, want 'Change Theme'", first.Title)
	}
}

func TestFuzzyScoreExactSubstring(t *testing.T) {
	score := fuzzyScore("change theme", "change")
	if score < 20 {
		t.Errorf("Exact substring score = %d, should be >= 20", score)
	}
}

func TestFuzzyScoreSubsequence(t *testing.T) {
	score := fuzzyScore("change theme", "ct")
	if score < 0 {
		t.Errorf("Subsequence 'ct' in 'change theme' should match, got %d", score)
	}
}

func TestFuzzyScoreNoMatch(t *testing.T) {
	score := fuzzyScore("hello", "xyz")
	if score >= 0 {
		t.Errorf("'xyz' in 'hello' should not match, got score %d", score)
	}
}

func TestFuzzyScoreConsecutiveBonus(t *testing.T) {
	// "ange" (consecutive) should score higher than "ae" (non-consecutive)
	consecutive := fuzzyScore("change", "ange")
	scattered := fuzzyScore("change", "ae")
	if consecutive <= scattered {
		t.Errorf("Consecutive (%d) should score higher than scattered (%d)", consecutive, scattered)
	}
}

func TestHandleKey_Escape(t *testing.T) {
	cp := NewCommandPalette()
	cp.Register(Command{ID: "1", Title: "Test", Category: "X"})
	cp.Open()

	consumed := cp.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Error("Escape should be consumed")
	}
	if cp.IsActive() {
		t.Error("Should be inactive after Escape")
	}
}

func TestHandleKey_ArrowNavigation(t *testing.T) {
	cp := NewCommandPalette()
	cp.RegisterMany([]Command{
		{ID: "1", Title: "A", Category: "X"},
		{ID: "2", Title: "B", Category: "X"},
		{ID: "3", Title: "C", Category: "X"},
	})
	cp.Open()

	if cp.SelectedIndex() != 0 {
		t.Fatalf("Initial SelectedIndex = %d, want 0", cp.SelectedIndex())
	}

	// Down → index 1
	cp.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if cp.SelectedIndex() != 1 {
		t.Errorf("After Down: SelectedIndex = %d, want 1", cp.SelectedIndex())
	}

	// Down → index 2
	cp.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if cp.SelectedIndex() != 2 {
		t.Errorf("After Down: SelectedIndex = %d, want 2", cp.SelectedIndex())
	}

	// Down → wrap to 0
	cp.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if cp.SelectedIndex() != 0 {
		t.Errorf("After wrap Down: SelectedIndex = %d, want 0", cp.SelectedIndex())
	}

	// Up → wrap to last (2)
	cp.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if cp.SelectedIndex() != 2 {
		t.Errorf("After wrap Up: SelectedIndex = %d, want 2", cp.SelectedIndex())
	}

	// Up → index 1
	cp.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if cp.SelectedIndex() != 1 {
		t.Errorf("After Up: SelectedIndex = %d, want 1", cp.SelectedIndex())
	}
}

func TestHandleKey_EnterExecutes(t *testing.T) {
	cp := NewCommandPalette()
	executed := false
	cp.Register(Command{
		ID:     "1",
		Title:  "Test Action",
		Action: func() { executed = true },
	})
	cp.Open()

	cp.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	if !executed {
		t.Error("Action should have been executed on Enter")
	}
	if cp.IsActive() {
		t.Error("Palette should close after executing action")
	}
}

func TestHandleKey_PrintableUpdatesQuery(t *testing.T) {
	cp := NewCommandPalette()
	cp.Register(Command{ID: "1", Title: "hello", Category: "X"})
	cp.Open()

	cp.HandleKey(&term.KeyEvent{Rune: 'h'})
	cp.HandleKey(&term.KeyEvent{Rune: 'i'})

	if cp.Query() != "hi" {
		t.Errorf("Query = %q, want 'hi'", cp.Query())
	}
}

func TestHandleKey_Backspace(t *testing.T) {
	cp := NewCommandPalette()
	cp.Open()

	// Type "abc"
	cp.HandleKey(&term.KeyEvent{Rune: 'a'})
	cp.HandleKey(&term.KeyEvent{Rune: 'b'})
	cp.HandleKey(&term.KeyEvent{Rune: 'c'})

	// Backspace → "ab"
	cp.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if cp.Query() != "ab" {
		t.Errorf("After backspace: Query = %q, want 'ab'", cp.Query())
	}
}

func TestHandleKey_NotActive(t *testing.T) {
	cp := NewCommandPalette()
	consumed := cp.HandleKey(&term.KeyEvent{Rune: 'a'})
	if consumed {
		t.Error("HandleKey should return false when not active")
	}
}

func TestFilterUpdatesOnQueryChange(t *testing.T) {
	cp := NewCommandPalette()
	cp.RegisterMany([]Command{
		{ID: "1", Title: "Change Theme", Category: "Theme"},
		{ID: "2", Title: "Clear All", Category: "Edit"},
		{ID: "3", Title: "Close", Category: "View"},
	})
	cp.Open()

	// Type "c" — should match all three
	cp.HandleKey(&term.KeyEvent{Rune: 'c'})
	if cp.FilteredCount() != 3 {
		t.Errorf("After 'c': FilteredCount = %d, want 3", cp.FilteredCount())
	}

	// Type "l" → "cl" — should match "Clear All" and "Close"
	cp.HandleKey(&term.KeyEvent{Rune: 'l'})
	if cp.FilteredCount() != 2 {
		t.Errorf("After 'cl': FilteredCount = %d, want 2", cp.FilteredCount())
	}

	// Type "e" → "cle" — should match "Clear All" and "Close" (c-l-e subsequence)
	cp.HandleKey(&term.KeyEvent{Rune: 'e'})
	if cp.FilteredCount() != 2 {
		t.Errorf("After 'cle': FilteredCount = %d, want 2 (Clear All + Close)", cp.FilteredCount())
	}

	// "Clear All" should rank higher (exact substring "cle")
	first := cp.FilteredCommands()[0]
	if first.Title != "Clear All" {
		t.Errorf("Top result = %q, want 'Clear All'", first.Title)
	}
}

func TestDefaultCommands(t *testing.T) {
	cmds := DefaultCommands()
	if len(cmds) < 8 {
		t.Errorf("DefaultCommands returned %d, want >= 8", len(cmds))
	}

	// Verify some expected commands exist
	ids := make(map[string]bool)
	for _, c := range cmds {
		ids[c.ID] = true
	}
	for _, expected := range []string{"theme.dracula", "theme.nord", "search.toggle", "conv.clear", "help", "quit"} {
		if !ids[expected] {
			t.Errorf("DefaultCommands missing ID %q", expected)
		}
	}
}

func TestPaintNoPanic(t *testing.T) {
	cp := NewCommandPalette()
	cp.RegisterMany([]Command{
		{ID: "1", Title: "Test A", Category: "X", Hint: "Ctrl+A"},
		{ID: "2", Title: "Test B", Category: "Y"},
	})
	cp.Open()

	buf := buffer.NewBuffer(60, 20)
	// Should not panic
	cp.Paint(buf, 60, 20)
}

func TestPaintInactive(t *testing.T) {
	cp := NewCommandPalette()
	buf := buffer.NewBuffer(60, 20)
	// Should not panic when inactive
	cp.Paint(buf, 60, 20)
}

func TestPaintEmpty(t *testing.T) {
	cp := NewCommandPalette()
	cp.Open()
	cp.query = "zzz"
	cp.Filter()

	buf := buffer.NewBuffer(60, 20)
	// Should not panic with no results
	cp.Paint(buf, 60, 20)
}

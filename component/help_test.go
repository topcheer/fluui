package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Test fixtures ──────────────────────────────────────────

func testGroups() []HelpGroup {
	return []HelpGroup{
		{
			Name: "Navigation",
			Entries: []HelpEntry{
				{Keys: "Ctrl+F", Description: "Find in conversation"},
				{Keys: "Ctrl+G", Description: "Go to top"},
				{Keys: "Ctrl+End", Description: "Jump to bottom"},
			},
		},
		{
			Name: "Editing",
			Entries: []HelpEntry{
				{Keys: "Ctrl+U", Description: "Clear input line"},
				{Keys: "Ctrl+W", Description: "Delete previous word"},
				{Keys: "Ctrl+A", Description: "Move cursor to start"},
			},
		},
		{
			Name: "AI",
			Entries: []HelpEntry{
				{Keys: "Enter", Description: "Send message"},
				{Keys: "Ctrl+C", Description: "Stop streaming"},
			},
		},
	}
}

// ─── Constructor tests ──────────────────────────────────────

func TestNewHelpOverlay_Defaults(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	if h == nil {
		t.Fatal("NewHelpOverlay returned nil")
	}
	if h.ID() == "" {
		t.Error("ID should not be empty")
	}
	if h.Title() == "" {
		t.Error("Title should not be empty")
	}
	if h.Query() != "" {
		t.Errorf("Query() = %q, want empty", h.Query())
	}
}

func TestNewHelpOverlay_Empty(t *testing.T) {
	h := NewHelpOverlay(nil)
	if h == nil {
		t.Fatal("NewHelpOverlay returned nil")
	}
	if h.TotalRows() != 0 {
		t.Errorf("TotalRows() = %d, want 0 for nil groups", h.TotalRows())
	}
}

func TestNewHelpOverlay_HasUniqueID(t *testing.T) {
	h1 := NewHelpOverlay(nil)
	h2 := NewHelpOverlay(nil)
	if h1.ID() == h2.ID() {
		t.Error("two HelpOverlays should have different IDs")
	}
}

// ─── Groups tests ───────────────────────────────────────────

func TestHelpOverlay_SetGroups(t *testing.T) {
	h := NewHelpOverlay(nil)
	h.SetGroups(testGroups())
	if len(h.Groups()) != 3 {
		t.Errorf("Groups() = %d groups, want 3", len(h.Groups()))
	}
}

func TestHelpOverlay_GroupsCopy(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	g := h.Groups()
	g[0].Name = "modified"
	g2 := h.Groups()
	if g2[0].Name == "modified" {
		t.Error("Groups() should return a copy, not internal reference")
	}
}

func TestHelpOverlay_GroupsEmpty(t *testing.T) {
	h := NewHelpOverlay(nil)
	if len(h.Groups()) != 0 {
		t.Errorf("Groups() = %d, want 0", len(h.Groups()))
	}
	if h.TotalRows() != 0 {
		t.Errorf("TotalRows() = %d, want 0", h.TotalRows())
	}
}

// ─── Query / Search tests ───────────────────────────────────

func TestHelpOverlay_SetQuery(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetQuery("find")
	if h.Query() != "find" {
		t.Errorf("Query() = %q, want 'find'", h.Query())
	}
}

func TestHelpOverlay_SetQuery_FiltersResults(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetQuery("find")
	// Only "Find in conversation" should match.
	total := h.TotalRows()
	if total == 0 {
		t.Error("TotalRows() should be > 0 for matching query")
	}
	// Should only have 1 group header + 1 entry = 2 rows.
	if total != 2 {
		t.Errorf("TotalRows() = %d, want 2 (1 header + 1 entry)", total)
	}
}

func TestHelpOverlay_SetQuery_NoMatch(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetQuery("zzzznomatch")
	if h.HasResults() {
		t.Error("HasResults() should be false for no-match query")
	}
}

func TestHelpOverlay_ClearQuery(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetQuery("find")
	if !h.HasResults() {
		t.Fatal("Should have results for 'find'")
	}
	h.ClearQuery()
	if h.Query() != "" {
		t.Errorf("Query() = %q, want empty after ClearQuery", h.Query())
	}
	if h.TotalRows() == 0 {
		t.Error("TotalRows() should be > 0 after clearing query")
	}
}

func TestHelpOverlay_AppendQuery(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.AppendQuery("find")
	if h.Query() != "find" {
		t.Errorf("Query() = %q, want 'find'", h.Query())
	}
}

func TestHelpOverlay_BackspaceQuery(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetQuery("find")
	ok := h.BackspaceQuery()
	if !ok {
		t.Error("BackspaceQuery() should return true when query is non-empty")
	}
	if h.Query() != "fin" {
		t.Errorf("Query() = %q, want 'fin'", h.Query())
	}
}

func TestHelpOverlay_BackspaceQuery_Empty(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	ok := h.BackspaceQuery()
	if ok {
		t.Error("BackspaceQuery() should return false when query is empty")
	}
	if h.Query() != "" {
		t.Errorf("Query() = %q, want empty", h.Query())
	}
}

func TestHelpOverlay_BackspaceQuery_Multibyte(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetQuery("搜索")
	h.BackspaceQuery()
	if h.Query() != "搜" {
		t.Errorf("Query() after multibyte backspace = %q, want '搜'", h.Query())
	}
}

func TestHelpOverlay_QueryCaseInsensitive(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetQuery("FIND")
	if !h.HasResults() {
		t.Error("HasResults() should be true for case-insensitive match 'FIND'")
	}
}

func TestHelpOverlay_QueryMatchesGroupName(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetQuery("navigation")
	// All Navigation items should match because group name matches.
	// Should have 1 header + 3 entries = 4 rows.
	if h.TotalRows() != 4 {
		t.Errorf("TotalRows() = %d, want 4 (1 header + 3 entries)", h.TotalRows())
	}
}

// ─── TotalRows tests ────────────────────────────────────────

func TestHelpOverlay_TotalRows(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	// 3 groups × (1 header + entries): Nav(1+3) + Edit(1+3) + AI(1+2) = 4+4+3 = 11.
	if h.TotalRows() != 11 {
		t.Errorf("TotalRows() = %d, want 11", h.TotalRows())
	}
}

func TestHelpOverlay_TotalRows_Empty(t *testing.T) {
	h := NewHelpOverlay(nil)
	if h.TotalRows() != 0 {
		t.Errorf("TotalRows() = %d, want 0", h.TotalRows())
	}
}

// ─── HasResults tests ───────────────────────────────────────

func TestHelpOverlay_HasResults(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	if !h.HasResults() {
		t.Error("HasResults() should be true with entries")
	}
}

func TestHelpOverlay_HasResults_Empty(t *testing.T) {
	h := NewHelpOverlay(nil)
	if h.HasResults() {
		t.Error("HasResults() should be false with no entries")
	}
}

// ─── FilteredGroups tests ───────────────────────────────────

func TestHelpOverlay_FilteredGroups(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	// Without query, should return all groups.
	fg := h.FilteredGroups()
	if len(fg) != 3 {
		t.Errorf("FilteredGroups() = %d, want 3", len(fg))
	}
}

func TestHelpOverlay_FilteredGroups_WithQuery(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetQuery("find")
	fg := h.FilteredGroups()
	// Only Navigation group with 1 entry should match.
	if len(fg) != 1 {
		t.Errorf("FilteredGroups() = %d, want 1", len(fg))
	}
	if len(fg[0].Entries) != 1 {
		t.Errorf("Entries in first filtered group = %d, want 1", len(fg[0].Entries))
	}
}

func TestHelpOverlay_FilteredGroupsCopy(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	fg := h.FilteredGroups()
	if len(fg) > 0 {
		fg[0].Name = "modified"
		fg2 := h.FilteredGroups()
		if fg2[0].Name == "modified" {
			t.Error("FilteredGroups() should return a copy")
		}
	}
}

// ─── Selection tests ────────────────────────────────────────

func TestHelpOverlay_SelectNext(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 30})
	h.SelectNext()
	if h.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex() = %d, want 1 after SelectNext", h.SelectedIndex())
	}
}

func TestHelpOverlay_SelectPrev(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 30})
	h.SelectNext()
	h.SelectNext()
	h.SelectPrev()
	if h.SelectedIndex() != 1 {
		t.Errorf("SelectedIndex() = %d, want 1 after SelectPrev", h.SelectedIndex())
	}
}

func TestHelpOverlay_SetSelected(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 30})
	h.SetSelected(5)
	if h.SelectedIndex() != 5 {
		t.Errorf("SelectedIndex() = %d, want 5", h.SelectedIndex())
	}
}

func TestHelpOverlay_SetSelected_Negative(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetSelected(-1)
	if h.SelectedIndex() != 0 {
		t.Errorf("SelectedIndex() = %d, should clamp to 0", h.SelectedIndex())
	}
}

// ─── Scroll tests ───────────────────────────────────────────

func TestHelpOverlay_ScrollDown(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 30})
	old := h.scrollY
	h.ScrollDown(3)
	if h.scrollY != old+3 {
		t.Errorf("scrollY after ScrollDown(3) = %d, want %d", h.scrollY, old+3)
	}
}

func TestHelpOverlay_ScrollUp(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 30})
	h.ScrollDown(5)
	h.ScrollUp(3)
	// scrollY should be 2.
	if h.scrollY != 2 {
		t.Errorf("scrollY = %d, want 2", h.scrollY)
	}
}

func TestHelpOverlay_ScrollUp_Clamped(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 30})
	h.ScrollUp(10) // should clamp to 0
	if h.scrollY < 0 {
		t.Errorf("scrollY = %d, should not go below 0", h.scrollY)
	}
}

// ─── Measure tests ──────────────────────────────────────────

func TestHelpOverlay_Measure(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	size := h.Measure(Constraints{})
	if size.W < 20 {
		t.Errorf("W = %d, should be >= 20", size.W)
	}
	if size.H < 5 {
		t.Errorf("H = %d, should be >= 5", size.H)
	}
}

func TestHelpOverlay_Measure_MaxConstraints(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	size := h.Measure(Constraints{MaxWidth: 40, MaxHeight: 15})
	if size.W > 40 {
		t.Errorf("W = %d, should be <= 40", size.W)
	}
	if size.H > 15 {
		t.Errorf("H = %d, should be <= 15", size.H)
	}
}

func TestHelpOverlay_Measure_Empty(t *testing.T) {
	h := NewHelpOverlay(nil)
	size := h.Measure(Constraints{})
	if size.W < 20 {
		t.Errorf("W = %d, should be >= 20", size.W)
	}
	if size.H < 5 {
		t.Errorf("H = %d, should be >= 5", size.H)
	}
}

// ─── Paint tests ────────────────────────────────────────────

func TestHelpOverlay_Paint_NoPanic(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 30})
	buf := buffer.NewBuffer(70, 35)
	h.Paint(buf)
}

func TestHelpOverlay_Paint_ZeroBounds(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	buf := buffer.NewBuffer(70, 35)
	h.Paint(buf) // should not panic with zero bounds
}

func TestHelpOverlay_Paint_SmallBounds(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 3})
	buf := buffer.NewBuffer(70, 35)
	h.Paint(buf) // should not panic
}

func TestHelpOverlay_Paint_WithQuery(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 30})
	h.SetQuery("find")
	buf := buffer.NewBuffer(70, 35)
	h.Paint(buf)
}

func TestHelpOverlay_Paint_RendersTitle(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 30})
	buf := buffer.NewBuffer(70, 35)
	h.Paint(buf)
	// Check row 0 has non-space characters.
	hasChar := false
	for x := 0; x < 60; x++ {
		if buf.GetCell(x, 0).Rune != ' ' {
			hasChar = true
			break
		}
	}
	if !hasChar {
		t.Error("Paint should render border/title on row 0")
	}
}

func TestHelpOverlay_Paint_RendersEntries(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 30})
	buf := buffer.NewBuffer(70, 35)
	h.Paint(buf)
	// Check 'Ctrl' appears in buffer.
	found := false
	for y := 0; y < 35; y++ {
		for x := 0; x < 60; x++ {
			c := buf.GetCell(x, y)
			if c.Rune == 'C' && x+1 < 60 {
				if buf.GetCell(x+1, y).Rune == 't' {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("Paint should render 'Ctrl' entries in the buffer")
	}
}

func TestHelpOverlay_Paint_Empty(t *testing.T) {
	h := NewHelpOverlay(nil)
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 30})
	buf := buffer.NewBuffer(70, 35)
	h.Paint(buf)
}

func TestHelpOverlay_Paint_NoResults(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetQuery("zzzz")
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 30})
	buf := buffer.NewBuffer(70, 35)
	h.Paint(buf)
}

// ─── Children ───────────────────────────────────────────────

func TestHelpOverlay_Children(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	if h.Children() != nil {
		t.Error("Children() should return nil (leaf component)")
	}
}

// ─── SetBounds ──────────────────────────────────────────────

func TestHelpOverlay_SetBounds(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	r := Rect{X: 5, Y: 3, W: 60, H: 25}
	h.SetBounds(r)
	b := h.Bounds()
	if b.X != 5 || b.Y != 3 || b.W != 60 || b.H != 25 {
		t.Errorf("Bounds = %+v, want %+v", b, r)
	}
}

// ─── Style tests ────────────────────────────────────────────

func TestHelpOverlay_SetStyle(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	custom := DefaultHelpStyle()
	custom.Title = buffer.Style{Fg: buffer.RGB(255, 0, 0)}
	h.SetStyle(custom)
	s := h.Style()
	if s.Title.Fg.R() != 255 {
		t.Error("SetStyle should update Title.Fg.R to 255")
	}
}

func TestDefaultHelpStyle(t *testing.T) {
	s := DefaultHelpStyle()
	// All styles should have non-nil values.
	_ = s.Border
	_ = s.Title
	_ = s.Key
	_ = s.Description
}

// ─── SetMaxWidth ────────────────────────────────────────────

func TestHelpOverlay_SetMaxWidth(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetMaxWidth(50)
	size := h.Measure(Constraints{})
	if size.W > 50 {
		t.Errorf("W = %d, should be <= 50 after SetMaxWidth", size.W)
	}
}

// ─── SetTitle ───────────────────────────────────────────────

func TestHelpOverlay_SetTitle(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetTitle("Custom Title")
	if h.Title() != "Custom Title" {
		t.Errorf("Title() = %q, want 'Custom Title'", h.Title())
	}
}

// ─── Concurrency ────────────────────────────────────────────

func TestHelpOverlay_ConcurrentAccess(t *testing.T) {
	h := NewHelpOverlay(testGroups())
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 30})

	var wg sync.WaitGroup

	// Concurrent writers.
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				h.SetQuery("find")
				h.ClearQuery()
				h.AppendQuery("a")
				h.BackspaceQuery()
			}
		}()
	}

	// Concurrent readers.
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_ = h.Query()
				_ = h.HasResults()
				_ = h.TotalRows()
				_ = h.Groups()
				_ = h.FilteredGroups()
				_ = h.SelectedIndex()
			}
		}()
	}

	// Concurrent cursor movers.
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				h.SelectNext()
				h.SelectPrev()
				h.SetSelected(0)
				h.ScrollDown(1)
				h.ScrollUp(1)
			}
		}()
	}

	// Concurrent painters.
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				buf := buffer.NewBuffer(70, 35)
				h.Paint(buf)
			}
		}()
	}

	wg.Wait()
}

func TestHelpOverlay_ConcurrentSetGroups(t *testing.T) {
	h := NewHelpOverlay(nil)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.SetGroups(testGroups())
		}()
	}

	// Concurrent readers.
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = h.Groups()
			_ = h.TotalRows()
		}()
	}

	wg.Wait()
}

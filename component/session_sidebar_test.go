package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestSessionSidebar_Basic(t *testing.T) {
	s := NewSessionSidebar()
	s.SetItems([]SessionItem{
		{ID: "1", Title: "Session A", Workspace: "ws1", LastMessage: "Hello", LastTime: time.Now()},
		{ID: "2", Title: "Session B", Workspace: "ws1", LastMessage: "World", LastTime: time.Now()},
	})
	if s.SelectedItem() == nil {
		t.Fatal("should have a selected item")
	}
}

func TestSessionSidebar_Filter(t *testing.T) {
	s := NewSessionSidebar()
	s.SetItems([]SessionItem{
		{ID: "1", Title: "Fix bug", Workspace: "ws1"},
		{ID: "2", Title: "Add feature", Workspace: "ws1"},
	})
	s.SetFilter("fix")
	item := s.SelectedItem()
	if item == nil || item.Title != "Fix bug" {
		t.Fatalf("expected 'Fix bug', got %v", item)
	}
}

func TestSessionSidebar_Collapse(t *testing.T) {
	s := NewSessionSidebar()
	if s.IsCollapsed() {
		t.Fatal("should start expanded")
	}
	s.ToggleCollapsed()
	if !s.IsCollapsed() {
		t.Fatal("should be collapsed after toggle")
	}
	s.SetCollapsed(false)
	if s.IsCollapsed() {
		t.Fatal("should be expanded after SetCollapsed(false)")
	}
}

func TestSessionSidebar_GroupToggle(t *testing.T) {
	s := NewSessionSidebar()
	s.SetItems([]SessionItem{
		{ID: "1", Title: "A", Workspace: "group1"},
		{ID: "2", Title: "B", Workspace: "group2"},
	})
	s.ToggleGroup("group1")
	item := s.SelectedItem()
	if item != nil && item.Workspace == "group1" {
		t.Fatal("group1 should be collapsed, items hidden")
	}
}

func TestSessionSidebar_PinnedSorting(t *testing.T) {
	s := NewSessionSidebar()
	s.SetItems([]SessionItem{
		{ID: "1", Title: "Normal", Workspace: "ws", Pinned: false, LastTime: time.Now()},
		{ID: "2", Title: "Pinned", Workspace: "ws", Pinned: true, LastTime: time.Now()},
	})
	item := s.SelectedItem()
	if item == nil || item.Title != "Pinned" {
		t.Fatalf("pinned should be first, got %v", item)
	}
}

func TestSessionSidebar_Paint(t *testing.T) {
	s := NewSessionSidebar()
	s.SetItems([]SessionItem{
		{ID: "1", Title: "Test", Workspace: "ws", LastMessage: "Hi", LastTime: time.Now(), Busy: true},
	})
	s.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	s.Paint(buf)
	// Should not panic, should have some content
}

func TestSessionSidebar_PaintCollapsed(t *testing.T) {
	s := NewSessionSidebar()
	s.SetItems([]SessionItem{
		{ID: "1", Title: "Test", Workspace: "ws"},
	})
	s.SetCollapsed(true)
	s.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	s.Paint(buf)
	// Collapsed state: draw indicator strip at column 0
	c := buf.GetCell(0, 0)
	if c.Rune == 0 {
		t.Fatal("collapsed sidebar should draw indicator")
	}
}

func TestSessionSidebar_KeyboardNav(t *testing.T) {
	s := NewSessionSidebar()
	s.SetItems([]SessionItem{
		{ID: "1", Title: "A", Workspace: "ws"},
		{ID: "2", Title: "B", Workspace: "ws"},
		{ID: "3", Title: "C", Workspace: "ws"},
	})

	// j to move down
	s.HandleKey(KeyEvent('j'))
	s.HandleKey(KeyEvent('j'))
	item := s.SelectedItem()
	if item == nil || item.ID != "3" {
		t.Fatalf("expected item 3 after 2 j, got %v", item)
	}

	// k to move up
	s.HandleKey(KeyEvent('k'))
	item = s.SelectedItem()
	if item == nil || item.ID != "2" {
		t.Fatalf("expected item 2 after k, got %v", item)
	}
}

func TestSessionSidebar_SearchMode(t *testing.T) {
	s := NewSessionSidebar()
	s.SetItems([]SessionItem{
		{ID: "1", Title: "Bug fix", Workspace: "ws"},
		{ID: "2", Title: "Feature", Workspace: "ws"},
	})

	// Press / to focus search
	s.HandleKey(KeyEvent('/'))
	// Type 'b'
	s.HandleKey(KeyEvent('b'))
	// Press enter
	s.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	item := s.SelectedItem()
	if item == nil || item.Title != "Bug fix" {
		t.Fatalf("expected 'Bug fix' after search, got %v", item)
	}
}

func TestSessionSidebar_TimeFormat(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
	}{
		{"now", time.Now()},
		{"5min", time.Now().Add(-5 * time.Minute)},
		{"2hour", time.Now().Add(-2 * time.Hour)},
		{"3day", time.Now().Add(-3 * 24 * time.Hour)},
		{"old", time.Now().Add(-30 * 24 * time.Hour)},
	}
	for _, tc := range tests {
		result := formatTimeAgo(tc.t)
		if result == "" {
			t.Errorf("%s: expected non-empty time string", tc.name)
		}
	}
}

// --- helpers ---

func KeyEvent(r rune) *term.KeyEvent {
	return &term.KeyEvent{Rune: r}
}

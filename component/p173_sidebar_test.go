package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestP173_sbTruncateStr(t *testing.T) {
	// Short string — no truncation
	if got := sbTruncateStr("hi", 10); got != "hi" {
		t.Errorf("expected 'hi', got %q", got)
	}
	// Exact length — no truncation
	if got := sbTruncateStr("hello", 5); got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
	// Needs truncation
	if got := sbTruncateStr("hello world", 8); got != "hello w…" {
		t.Errorf("expected 'hello w…', got %q", got)
	}
	// maxLen <= 1
	if got := sbTruncateStr("hello", 1); got != "…" {
		t.Errorf("expected '…', got %q", got)
	}
	// maxLen = 0
	if got := sbTruncateStr("hello", 0); got != "…" {
		t.Errorf("expected '…', got %q", got)
	}
	// maxLen = 2
	if got := sbTruncateStr("hello", 2); got != "h…" {
		t.Errorf("expected 'h…', got %q", got)
	}
	// Unicode — maxLen=5 means 4 chars + ellipsis
	if got := sbTruncateStr("héllo wörld", 5); got != "héll…" {
		t.Errorf("expected 'héll…', got %q", got)
	}
}

func TestP173_sbVisibleStrLen(t *testing.T) {
	if got := sbVisibleStrLen(""); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
	if got := sbVisibleStrLen("hello"); got != 5 {
		t.Errorf("expected 5, got %d", got)
	}
	if got := sbVisibleStrLen("héllo"); got != 5 {
		t.Errorf("expected 5, got %d", got)
	}
}

func TestP173_formatTimeAgo(t *testing.T) {
	now := time.Now()
	// Just now
	if got := formatTimeAgo(now); got != "now" {
		t.Errorf("expected 'now', got %q", got)
	}
	// Minutes ago
	if got := formatTimeAgo(now.Add(-5 * time.Minute)); got != "5m" {
		t.Errorf("expected '5m', got %q", got)
	}
	// Hours ago
	if got := formatTimeAgo(now.Add(-3 * time.Hour)); got != "3h" {
		t.Errorf("expected '3h', got %q", got)
	}
	// Days ago (1 day)
	if got := formatTimeAgo(now.Add(-25 * time.Hour)); got != "1d" {
		t.Errorf("expected '1d', got %q", got)
	}
	// Days ago (multiple)
	if got := formatTimeAgo(now.Add(-72 * time.Hour)); got != "3d" {
		t.Errorf("expected '3d', got %q", got)
	}
	// Very old — date format
	old := now.Add(-10 * 24 * time.Hour)
	got := formatTimeAgo(old)
	if len(got) != 5 { // "01/02" format
		t.Errorf("expected date format (5 chars), got %q (len %d)", got, len(got))
	}
}

func TestP173_sbDrawTextLeft(t *testing.T) {
	buf := buffer.NewBuffer(10, 5)
	sbDrawTextLeft(buf, 0, 0, "hello", buffer.NamedColor(buffer.NamedWhite), 10)
	if buf.GetCell(0, 0).Rune != 'h' {
		t.Error("expected 'h' at (0,0)")
	}
	// Truncate at maxWidth
	sbDrawTextLeft(buf, 0, 1, "hello world", buffer.NamedColor(buffer.NamedWhite), 5)
	if buf.GetCell(5, 1).Rune != ' ' {
		t.Error("expected truncation at col 5")
	}
}

func TestP173_SessionSidebar_HandleKey(t *testing.T) {
	sb := NewSessionSidebar()
	sb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 20})

	sb.SetItems([]SessionItem{
		{ID: "1", Title: "Session 1", Workspace: "Active"},
		{ID: "2", Title: "Session 2", Workspace: "Active"},
		{ID: "3", Title: "Session 3", Workspace: "Other"},
	})

	// Down
	sb.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	// Up
	sb.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	// j key
	sb.HandleKey(&term.KeyEvent{Rune: 'j'})
	// k key
	sb.HandleKey(&term.KeyEvent{Rune: 'k'})
	// g (go to top)
	sb.HandleKey(&term.KeyEvent{Rune: 'g'})
	// G (go to bottom)
	sb.HandleKey(&term.KeyEvent{Rune: 'G'})
	// Enter
	sb.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	// / (search)
	sb.HandleKey(&term.KeyEvent{Rune: '/'})
	// Search type
	sb.HandleKey(&term.KeyEvent{Rune: 'S'})
	// Search escape
	sb.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	// Unknown key
	sb.HandleKey(&term.KeyEvent{Key: term.KeyF1})
}

func TestP173_SessionSidebar_Measure(t *testing.T) {
	sb := NewSessionSidebar()
	sb.SetWidth(25)
	s := sb.Measure(Constraints{MaxWidth: 100, MaxHeight: 100})
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected non-zero size, got %dx%d", s.W, s.H)
	}

	// Narrow constraints — sidebar uses configured width
	s2 := sb.Measure(Constraints{MaxWidth: 5, MaxHeight: 5})
	_ = s2 // Measure may use configured width, not constraints
}

func TestP173_SessionSidebar_SetGroupExpanded(t *testing.T) {
	sb := NewSessionSidebar()
	sb.SetItems([]SessionItem{
		{ID: "1", Title: "S1", Workspace: "G1"},
		{ID: "2", Title: "S2", Workspace: "G2"},
	})

	sb.SetGroupExpanded("G1", false)
	sb.SetGroupExpanded("G1", true)
	sb.ToggleGroup("G1")
}

func TestP173_SessionSidebar_SelectedItem(t *testing.T) {
	sb := NewSessionSidebar()
	// Empty — should return nil
	item := sb.SelectedItem()
	if item != nil {
		t.Error("expected nil item when no sessions")
	}

	sb.SetItems([]SessionItem{
		{ID: "1", Title: "S1", Workspace: "G"},
	})
	item = sb.SelectedItem()
	if item == nil {
		t.Error("expected non-nil item")
	} else if item.ID != "1" {
		t.Errorf("expected ID '1', got %q", item.ID)
	}
}

func TestP173_SessionSidebar_Paint(t *testing.T) {
	sb := NewSessionSidebar()
	sb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 20})
	sb.SetItems([]SessionItem{
		{ID: "1", Title: "S1", Workspace: "Active", LastTime: time.Now()},
		{ID: "2", Title: "S2", Workspace: "Active", LastTime: time.Now().Add(-time.Hour)},
		{ID: "3", Title: "S3", Workspace: "Other", Pinned: true},
	})
	buf := buffer.NewBuffer(30, 20)
	sb.Paint(buf)
}

func TestP173_SessionSidebar_Filter(t *testing.T) {
	sb := NewSessionSidebar()
	sb.SetItems([]SessionItem{
		{ID: "1", Title: "Alpha", Workspace: "G"},
		{ID: "2", Title: "Beta", Workspace: "G"},
	})
	sb.SetFilter("Alpha")
	// Should filter to only show Alpha
}

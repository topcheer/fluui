package app

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

func TestInputHistoryEmpty(t *testing.T) {
	il := NewInputLine("> ")

	// Up/Down on empty history should be a no-op (not panic).
	il.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if il.Text() != "" {
		t.Fatalf("expected empty text after Up on empty history, got %q", il.Text())
	}

	il.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if il.Text() != "" {
		t.Fatalf("expected empty text after Down on empty history, got %q", il.Text())
	}
}

func TestInputHistoryNavigation(t *testing.T) {
	il := NewInputLine("> ")
	il.SetHistory([]string{"first", "second", "third"})

	// Up → most recent ("third")
	il.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if il.Text() != "third" {
		t.Fatalf("expected 'third', got %q", il.Text())
	}

	// Up → "second"
	il.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if il.Text() != "second" {
		t.Fatalf("expected 'second', got %q", il.Text())
	}

	// Up → "first" (oldest)
	il.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if il.Text() != "first" {
		t.Fatalf("expected 'first', got %q", il.Text())
	}

	// Up again → still "first" (clamped)
	il.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if il.Text() != "first" {
		t.Fatalf("expected 'first' (clamped), got %q", il.Text())
	}

	// Down → "second"
	il.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if il.Text() != "second" {
		t.Fatalf("expected 'second', got %q", il.Text())
	}

	// Down → "third"
	il.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if il.Text() != "third" {
		t.Fatalf("expected 'third', got %q", il.Text())
	}

	// Down → restore draft (original empty)
	il.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if il.Text() != "" {
		t.Fatalf("expected draft restored (empty), got %q", il.Text())
	}
}

func TestInputHistoryDraftRestore(t *testing.T) {
	il := NewInputLine("> ")
	il.SetHistory([]string{"old message"})

	// Type something (not yet submitted).
	for _, r := range "draft text" {
		il.HandleKey(&term.KeyEvent{Rune: r})
	}

	// Up → shows "old message"
	il.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if il.Text() != "old message" {
		t.Fatalf("expected 'old message', got %q", il.Text())
	}

	// Down → restores draft
	il.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if il.Text() != "draft text" {
		t.Fatalf("expected draft 'draft text' restored, got %q", il.Text())
	}
}

func TestInputHistorySubmitAdds(t *testing.T) {
	var submitted []string
	il := NewInputLineWithHandler("> ", func(text string) {
		submitted = append(submitted, text)
	})

	// Type and submit first message.
	for _, r := range "hello" {
		il.HandleKey(&term.KeyEvent{Rune: r})
	}
	il.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	// Type and submit second message.
	for _, r := range "world" {
		il.HandleKey(&term.KeyEvent{Rune: r})
	}
	il.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	// History should contain both.
	h := il.History()
	if len(h) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(h))
	}
	if h[0] != "hello" {
		t.Fatalf("expected h[0]='hello', got %q", h[0])
	}
	if h[1] != "world" {
		t.Fatalf("expected h[1]='world', got %q", h[1])
	}

	// Up → most recent "world"
	il.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if il.Text() != "world" {
		t.Fatalf("expected 'world', got %q", il.Text())
	}

	// Up → "hello"
	il.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if il.Text() != "hello" {
		t.Fatalf("expected 'hello', got %q", il.Text())
	}
}

func TestInputHistoryDuplicateConsecutive(t *testing.T) {
	il := NewInputLine("> ")

	il.AddHistory("same")
	il.AddHistory("same")
	il.AddHistory("same")

	if len(il.history) != 1 {
		t.Fatalf("expected 1 entry (dedup consecutive), got %d", len(il.history))
	}
}

func TestInputHistoryAddHistory(t *testing.T) {
	il := NewInputLine("> ")

	il.AddHistory("a")
	il.AddHistory("b")
	il.AddHistory("c")

	h := il.History()
	if len(h) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(h))
	}
	if h[2] != "c" {
		t.Fatalf("expected last='c', got %q", h[2])
	}
}

func TestInputHistorySetHistory(t *testing.T) {
	il := NewInputLine("> ")
	il.AddHistory("old")

	il.SetHistory([]string{"x", "y"})

	h := il.History()
	if len(h) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(h))
	}
	if h[0] != "x" || h[1] != "y" {
		t.Fatalf("expected ['x','y'], got %v", h)
	}
}

func TestInputHistoryTypingExitsBrowse(t *testing.T) {
	il := NewInputLine("> ")
	il.SetHistory([]string{"old1", "old2"})

	// Browse to history entry.
	il.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if il.Text() != "old2" {
		t.Fatalf("expected 'old2', got %q", il.Text())
	}

	// Type a character — should exit history mode.
	il.HandleKey(&term.KeyEvent{Rune: 'X'})
	if il.Text() != "old2X" {
		t.Fatalf("expected 'old2X', got %q", il.Text())
	}

	// Down should NOT restore draft anymore (we exited browse mode).
	il.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if il.Text() != "old2X" {
		t.Fatalf("Down should be no-op after exiting browse, got %q", il.Text())
	}
}

func TestInputHistoryClearExitsBrowse(t *testing.T) {
	il := NewInputLine("> ")
	il.SetHistory([]string{"old"})

	il.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if il.Text() != "old" {
		t.Fatalf("expected 'old', got %q", il.Text())
	}

	// Ctrl+U clears and exits browse mode.
	il.HandleKey(&term.KeyEvent{Rune: 'u', Modifiers: term.ModCtrl})
	if il.Text() != "" {
		t.Fatalf("expected empty after Ctrl+U, got %q", il.Text())
	}

	// Down should be no-op (not in browse mode).
	il.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if il.Text() != "" {
		t.Fatalf("Down after Ctrl+U should be no-op, got %q", il.Text())
	}
}

func TestInputHistoryDoesNotAffectCursor(t *testing.T) {
	il := NewInputLine("> ")
	il.SetHistory([]string{"hello world"})

	il.HandleKey(&term.KeyEvent{Key: term.KeyUp})

	// After loading history entry, cursor should be at end.
	if il.Cursor() != len("hello world") {
		t.Fatalf("expected cursor at %d, got %d", len("hello world"), il.Cursor())
	}
}

func TestInputHistoryMaxSize(t *testing.T) {
	il := NewInputLine("> ")
	il.maxHistory = 3

	il.AddHistory("a")
	il.AddHistory("b")
	il.AddHistory("c")
	il.AddHistory("d") // should evict "a"

	h := il.History()
	if len(h) != 3 {
		t.Fatalf("expected 3 entries (capped), got %d", len(h))
	}
	if h[0] != "b" {
		t.Fatalf("expected h[0]='b' (oldest after trim), got %q", h[0])
	}
	if h[2] != "d" {
		t.Fatalf("expected h[2]='d' (newest), got %q", h[2])
	}
}

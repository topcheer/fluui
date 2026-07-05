package term

import (
	"testing"
)

// --- Focus Tracking parser tests ---

func TestP46_FocusIn(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[I"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event for FocusIn, got %d", len(evs))
	}
	if evs[0].Type != EventFocus {
		t.Errorf("expected EventFocus, got %d", evs[0].Type)
	}
	if !evs[0].Focused {
		t.Error("expected Focused=true for FocusIn")
	}
}

func TestP46_FocusOut(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[O"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event for FocusOut, got %d", len(evs))
	}
	if evs[0].Type != EventFocus {
		t.Errorf("expected EventFocus, got %d", evs[0].Type)
	}
	if evs[0].Focused {
		t.Error("expected Focused=false for FocusOut")
	}
}

func TestP46_FocusInWithOtherEvents(t *testing.T) {
	p := NewParser()
	// Mix: key event, focus in, key event, focus out
	evs := p.Feed([]byte("a\x1b[Ib\x1b[Oc"))
	if len(evs) != 5 {
		t.Fatalf("expected 5 events, got %d", len(evs))
	}
	// 0: key 'a'
	if evs[0].Type != EventKey || evs[0].Key == nil || evs[0].Key.Rune != 'a' {
		t.Errorf("expected key 'a' first, got %+v", evs[0])
	}
	// 1: focus in
	if evs[1].Type != EventFocus || !evs[1].Focused {
		t.Errorf("expected FocusIn second, got %+v", evs[1])
	}
	// 2: key 'b'
	if evs[2].Type != EventKey || evs[2].Key == nil || evs[2].Key.Rune != 'b' {
		t.Errorf("expected key 'b' third, got %+v", evs[2])
	}
	// 3: focus out
	if evs[3].Type != EventFocus || evs[3].Focused {
		t.Errorf("expected FocusOut fourth, got %+v", evs[3])
	}
	// 4: key 'c'
	if evs[4].Type != EventKey || evs[4].Key == nil || evs[4].Key.Rune != 'c' {
		t.Errorf("expected key 'c' fifth, got %+v", evs[4])
	}
}

func TestP46_FocusInFocusOutCycle(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[I\x1b[O\x1b[I\x1b[O"))
	if len(evs) != 4 {
		t.Fatalf("expected 4 events, got %d", len(evs))
	}
	if !evs[0].Focused {
		t.Error("expected first event Focused=true")
	}
	if evs[1].Focused {
		t.Error("expected second event Focused=false")
	}
	if !evs[2].Focused {
		t.Error("expected third event Focused=true")
	}
	if evs[3].Focused {
		t.Error("expected fourth event Focused=false")
	}
}

func TestP46_FocusNotConfusedWithArrowKeys(t *testing.T) {
	p := NewParser()
	// Up arrow = ESC[A — should NOT be parsed as focus
	evs := p.Feed([]byte("\x1b[A"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Type == EventFocus {
		t.Error("Up arrow should not be parsed as focus event")
	}
	if evs[0].Key == nil || evs[0].Key.Key != KeyUp {
		t.Errorf("expected KeyUp, got %+v", evs[0])
	}
}

// --- Focus tracking enable/disable constants ---

func TestP46_FocusEnableConstants(t *testing.T) {
	if EnableFocus != "\x1b[?1004h" {
		t.Errorf("expected EnableFocus to be ESC[?1004h, got %q", EnableFocus)
	}
	if DisableFocus != "\x1b[?1004l" {
		t.Errorf("expected DisableFocus to be ESC[?1004l, got %q", DisableFocus)
	}
}

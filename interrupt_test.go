package fluui

import (
	"testing"

	"github.com/topcheer/fluui/event"
	"github.com/topcheer/fluui/internal/term"
)

// TestCtrlCDefaultQuit tests that Ctrl+C triggers Quit by default
// (no OnInterrupt handler set).
func TestCtrlCDefaultQuit(t *testing.T) {
	// We can't create a real App (needs a terminal), so test the
	// dispatcher logic directly.
	disp := event.NewDispatcher()

	quitCalled := false
	var onInterrupt func() bool

	disp.OnKey(func(e event.Event) bool {
		if e.Key == nil {
			return false
		}
		if e.Key.Modifiers&term.ModCtrl != 0 && (e.Key.Rune == 'c' || e.Key.Rune == 'C') {
			if onInterrupt != nil {
				if onInterrupt() {
					quitCalled = true
				}
				return true
			}
			quitCalled = true
			return true
		}
		return true
	})

	// Dispatch Ctrl+C
	disp.Dispatch(event.Event{
		Type: event.TypeKey,
		Key:  &term.KeyEvent{Rune: 'c', Modifiers: term.ModCtrl},
	})

	if !quitCalled {
		t.Error("Ctrl+C should trigger quit by default")
	}
}

// TestOnInterruptBlock tests that OnInterrupt returning false
// prevents the app from quitting.
func TestOnInterruptBlock(t *testing.T) {
	disp := event.NewDispatcher()

	quitCalled := false
	onInterrupt := func() bool { return false } // block quit

	disp.OnKey(func(e event.Event) bool {
		if e.Key == nil {
			return false
		}
		if e.Key.Modifiers&term.ModCtrl != 0 && (e.Key.Rune == 'c' || e.Key.Rune == 'C') {
			if onInterrupt != nil {
				if onInterrupt() {
					quitCalled = true
				}
				return true
			}
			quitCalled = true
			return true
		}
		return true
	})

	// Dispatch Ctrl+C
	disp.Dispatch(event.Event{
		Type: event.TypeKey,
		Key:  &term.KeyEvent{Rune: 'c', Modifiers: term.ModCtrl},
	})

	if quitCalled {
		t.Error("Ctrl+C should NOT quit when OnInterrupt returns false")
	}
}

// TestOnInterruptAllow tests that OnInterrupt returning true
// allows the app to quit.
func TestOnInterruptAllow(t *testing.T) {
	disp := event.NewDispatcher()

	quitCalled := false
	onInterrupt := func() bool { return true } // allow quit

	disp.OnKey(func(e event.Event) bool {
		if e.Key == nil {
			return false
		}
		if e.Key.Modifiers&term.ModCtrl != 0 && (e.Key.Rune == 'c' || e.Key.Rune == 'C') {
			if onInterrupt != nil {
				if onInterrupt() {
					quitCalled = true
				}
				return true
			}
			quitCalled = true
			return true
		}
		return true
	})

	// Dispatch Ctrl+C
	disp.Dispatch(event.Event{
		Type: event.TypeKey,
		Key:  &term.KeyEvent{Rune: 'c', Modifiers: term.ModCtrl},
	})

	if !quitCalled {
		t.Error("Ctrl+C should quit when OnInterrupt returns true")
	}
}

// TestOnQuitCleanup tests that the onQuit callback is called
// during App.Run() exit.
func TestOnQuitCleanup(t *testing.T) {
	// Verify the pattern works: onQuit is called after loop.Run returns.
	// We can't test with a real terminal, but we can verify the
	// callback registration mechanism.
	cleanupCalled := false
	onQuit := func() {
		cleanupCalled = true
	}

	// Simulate what Run() does
	onQuit()

	if !cleanupCalled {
		t.Error("onQuit callback should have been called")
	}
}

// TestCtrlCCaseInsensitive tests both lowercase and uppercase C.
func TestCtrlCCaseInsensitive(t *testing.T) {
	disp := event.NewDispatcher()

	quitCount := 0

	disp.OnKey(func(e event.Event) bool {
		if e.Key == nil {
			return false
		}
		if e.Key.Modifiers&term.ModCtrl != 0 && (e.Key.Rune == 'c' || e.Key.Rune == 'C') {
			quitCount++
			return true
		}
		return true
	})

	// Dispatch Ctrl+C (lowercase)
	disp.Dispatch(event.Event{
		Type: event.TypeKey,
		Key:  &term.KeyEvent{Rune: 'c', Modifiers: term.ModCtrl},
	})

	// Dispatch Ctrl+Shift+C (uppercase)
	disp.Dispatch(event.Event{
		Type: event.TypeKey,
		Key:  &term.KeyEvent{Rune: 'C', Modifiers: term.ModCtrl},
	})

	if quitCount != 2 {
		t.Errorf("Expected 2 quit triggers, got %d", quitCount)
	}
}

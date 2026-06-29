package event

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

// ============================================================
// P26-B: Event → Dispatcher → Handler End-to-End Integration Tests
// ============================================================
// These tests verify the full event chain:
// Key/Mouse/Paste/Resize Event → Dispatcher.BindKey →
// Handler callback → state mutation → dirty flag.
// ============================================================

// --- Key → BindKey → handler fired ---

func TestP26B_KeyBindingHandlerFired(t *testing.T) {
	d := NewDispatcher()
	var fired int32

	d.BindKey(CtrlRune('q'), func(e Event) bool {
		atomic.StoreInt32(&fired, 1)
		return true
	})

	// Simulate Ctrl+Q keypress
	keyEvent := &term.KeyEvent{Rune: 'q', Modifiers: term.ModCtrl}
	e := Event{Type: TypeKey, Key: keyEvent}
	consumed := d.Dispatch(e)

	if !consumed {
		t.Error("handler should have consumed the event")
	}
	if atomic.LoadInt32(&fired) != 1 {
		t.Error("handler was not fired")
	}
}

// --- Key → default handler when no binding matches ---

func TestP26B_DefaultKeyHandlerFallback(t *testing.T) {
	d := NewDispatcher()
	var defaultFired int32

	d.BindKey(CtrlRune('q'), func(e Event) bool {
		return true // quit handler
	})

	d.OnKey(func(e Event) bool {
		atomic.StoreInt32(&defaultFired, 1)
		return true
	})

	// Simulate 'a' keypress (not bound)
	keyEvent := &term.KeyEvent{Rune: 'a'}
	e := Event{Type: TypeKey, Key: keyEvent}
	consumed := d.Dispatch(e)

	if !consumed {
		t.Error("default handler should have consumed")
	}
	if atomic.LoadInt32(&defaultFired) != 1 {
		t.Error("default handler was not fired")
	}
}

// --- Binding takes priority over default handler ---

func TestP26B_BindingBeforeDefault(t *testing.T) {
	d := NewDispatcher()
	var bindingFired, defaultFired int32

	d.BindKey(PlainRune('x'), func(e Event) bool {
		atomic.StoreInt32(&bindingFired, 1)
		return true
	})

	d.OnKey(func(e Event) bool {
		atomic.StoreInt32(&defaultFired, 1)
		return true
	})

	keyEvent := &term.KeyEvent{Rune: 'x'}
	e := Event{Type: TypeKey, Key: keyEvent}
	d.Dispatch(e)

	if atomic.LoadInt32(&bindingFired) != 1 {
		t.Error("binding handler not fired")
	}
	if atomic.LoadInt32(&defaultFired) != 0 {
		t.Error("default handler should NOT fire when binding matches")
	}
}

// --- Mouse event → handler ---

func TestP26B_MouseEventDispatched(t *testing.T) {
	d := NewDispatcher()
	var mouseHandled int32

	d.OnMouse(func(e Event) bool {
		atomic.StoreInt32(&mouseHandled, 1)
		if e.Mouse == nil {
			t.Error("mouse event is nil")
		}
		return true
	})

	mouseEvent := &term.MouseEvent{X: 10, Y: 5}
	e := Event{Type: TypeMouse, Mouse: mouseEvent}
	consumed := d.Dispatch(e)

	if !consumed {
		t.Error("mouse handler should have consumed")
	}
	if atomic.LoadInt32(&mouseHandled) != 1 {
		t.Error("mouse handler not fired")
	}
}

// --- Resize event → handler ---

func TestP26B_ResizeEventDispatched(t *testing.T) {
	d := NewDispatcher()
	var resizeHandled int32
	var newW, newH int

	d.OnResize(func(e Event) bool {
		atomic.StoreInt32(&resizeHandled, 1)
		newW = e.Width
		newH = e.Height
		return true
	})

	e := Event{Type: TypeResize, Width: 120, Height: 40}
	consumed := d.Dispatch(e)

	if !consumed {
		t.Error("resize handler should have consumed")
	}
	if atomic.LoadInt32(&resizeHandled) != 1 {
		t.Error("resize handler not fired")
	}
	if newW != 120 || newH != 40 {
		t.Errorf("resize dims wrong: got %dx%d, want 120x40", newW, newH)
	}
}

// --- Paste event → handler ---

func TestP26B_PasteEventDispatched(t *testing.T) {
	d := NewDispatcher()
	var pasteHandled int32
	var pastedText string

	d.OnPaste(func(e Event) bool {
		atomic.StoreInt32(&pasteHandled, 1)
		pastedText = e.Paste
		return true
	})

	e := Event{Type: TypePaste, Paste: "pasted content"}
	consumed := d.Dispatch(e)

	if !consumed {
		t.Error("paste handler should have consumed")
	}
	if atomic.LoadInt32(&pasteHandled) != 1 {
		t.Error("paste handler not fired")
	}
	if pastedText != "pasted content" {
		t.Errorf("paste content wrong: got %q", pastedText)
	}
}

// --- Unhandled event returns false ---

func TestP26B_UnhandledEventReturnsFalse(t *testing.T) {
	d := NewDispatcher()

	// No handlers registered — all events should return false
	e := Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'a'}}
	if d.Dispatch(e) {
		t.Error("unhandled key should return false")
	}

	e = Event{Type: TypeMouse, Mouse: &term.MouseEvent{X: 0, Y: 0}}
	if d.Dispatch(e) {
		t.Error("unhandled mouse should return false")
	}

	e = Event{Type: TypeResize, Width: 80, Height: 24}
	if d.Dispatch(e) {
		t.Error("unhandled resize should return false")
	}

	e = Event{Type: TypePaste, Paste: "test"}
	if d.Dispatch(e) {
		t.Error("unhandled paste should return false")
	}
}

// --- Multiple key bindings ---

func TestP26B_MultipleKeyBindings(t *testing.T) {
	d := NewDispatcher()
	var quitFired, saveFired, openFired int32

	d.BindKey(CtrlRune('q'), func(e Event) bool {
		atomic.StoreInt32(&quitFired, 1)
		return true
	})
	d.BindKey(CtrlRune('s'), func(e Event) bool {
		atomic.StoreInt32(&saveFired, 1)
		return true
	})
	d.BindKey(CtrlRune('o'), func(e Event) bool {
		atomic.StoreInt32(&openFired, 1)
		return true
	})

	// Dispatch Ctrl+S
	d.Dispatch(Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 's', Modifiers: term.ModCtrl}})

	if atomic.LoadInt32(&quitFired) != 0 {
		t.Error("quit should not fire")
	}
	if atomic.LoadInt32(&saveFired) != 1 {
		t.Error("save should fire")
	}
	if atomic.LoadInt32(&openFired) != 0 {
		t.Error("open should not fire")
	}
}

// --- Key with modifiers (Shift, Alt, Ctrl) ---

func TestP26B_KeyModifiersRouting(t *testing.T) {
	d := NewDispatcher()
	var ctrlFired, altFired, shiftFired int32

	d.BindKey(KeyShortcut{Rune: 'a', Modifiers: term.ModCtrl}, func(e Event) bool {
		atomic.StoreInt32(&ctrlFired, 1)
		return true
	})
	d.BindKey(KeyShortcut{Rune: 'a', Modifiers: term.ModAlt}, func(e Event) bool {
		atomic.StoreInt32(&altFired, 1)
		return true
	})

	// Ctrl+A
	d.Dispatch(Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'a', Modifiers: term.ModCtrl}})
	if atomic.LoadInt32(&ctrlFired) != 1 {
		t.Error("Ctrl+A handler not fired")
	}

	// Alt+A
	d.Dispatch(Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'a', Modifiers: term.ModAlt}})
	if atomic.LoadInt32(&altFired) != 1 {
		t.Error("Alt+A handler not fired")
	}

	// Shift+A (no handler registered)
	d.Dispatch(Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'a', Modifiers: term.ModShift}})
	_ = shiftFired
}

// --- Arrow keys routing ---

func TestP26B_ArrowKeyRouting(t *testing.T) {
	d := NewDispatcher()
	var upFired, downFired, leftFired, rightFired int32

	d.BindKey(Plain(term.KeyUp), func(e Event) bool {
		atomic.StoreInt32(&upFired, 1)
		return true
	})
	d.BindKey(Plain(term.KeyDown), func(e Event) bool {
		atomic.StoreInt32(&downFired, 1)
		return true
	})
	d.BindKey(Plain(term.KeyLeft), func(e Event) bool {
		atomic.StoreInt32(&leftFired, 1)
		return true
	})
	d.BindKey(Plain(term.KeyRight), func(e Event) bool {
		atomic.StoreInt32(&rightFired, 1)
		return true
	})

	// Dispatch each arrow key
	d.Dispatch(Event{Type: TypeKey, Key: &term.KeyEvent{Key: term.KeyUp}})
	d.Dispatch(Event{Type: TypeKey, Key: &term.KeyEvent{Key: term.KeyDown}})
	d.Dispatch(Event{Type: TypeKey, Key: &term.KeyEvent{Key: term.KeyLeft}})
	d.Dispatch(Event{Type: TypeKey, Key: &term.KeyEvent{Key: term.KeyRight}})

	if atomic.LoadInt32(&upFired) != 1 {
		t.Error("Up handler not fired")
	}
	if atomic.LoadInt32(&downFired) != 1 {
		t.Error("Down handler not fired")
	}
	if atomic.LoadInt32(&leftFired) != 1 {
		t.Error("Left handler not fired")
	}
	if atomic.LoadInt32(&rightFired) != 1 {
		t.Error("Right handler not fired")
	}
}

// --- Full event loop simulation: key → handler → state → quit ---

func TestP26B_FullEventLoopSimulation(t *testing.T) {
	d := NewDispatcher()

	var state int32 // 0=running, 1=quit
	d.BindKey(CtrlRune('c'), func(e Event) bool {
		atomic.StoreInt32(&state, 1) // signal quit
		return true
	})
	d.BindKey(PlainRune('a'), func(e Event) bool {
		// normal key, doesn't change state
		return true
	})

	// Simulate a sequence of events
	keys := []Event{
		{Type: TypeKey, Key: &term.KeyEvent{Rune: 'a'}},
		{Type: TypeKey, Key: &term.KeyEvent{Rune: 'b'}},
		{Type: TypeKey, Key: &term.KeyEvent{Key: term.KeyUp}},
		{Type: TypeKey, Key: &term.KeyEvent{Rune: 'c', Modifiers: term.ModCtrl}}, // Ctrl+C → quit
	}

	for _, e := range keys {
		if d.Dispatch(e) {
			if atomic.LoadInt32(&state) == 1 {
				break // quit signaled
			}
		}
	}

	if atomic.LoadInt32(&state) != 1 {
		t.Error("quit state was not reached after Ctrl+C")
	}
}

// --- Concurrent dispatch stress ---

func TestP26B_ConcurrentDispatchStress(t *testing.T) {
	d := NewDispatcher()
	var counter int32

	d.OnKey(func(e Event) bool {
		atomic.AddInt32(&counter, 1)
		return true
	})

	var wg sync.WaitGroup
	const goroutines = 50
	const eventsPerG = 100

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < eventsPerG; j++ {
				d.Dispatch(Event{
					Type: TypeKey,
					Key:  &term.KeyEvent{Rune: 'x'},
				})
			}
		}()
	}
	wg.Wait()

	expected := int32(goroutines * eventsPerG)
	if atomic.LoadInt32(&counter) != expected {
		t.Errorf("counter %d, expected %d", counter, expected)
	}
}

// --- Event type coverage: all event types in sequence ---

func TestP26B_AllEventTypesInSequence(t *testing.T) {
	d := NewDispatcher()
	var keyCount, mouseCount, resizeCount, pasteCount int32

	d.OnKey(func(e Event) bool { atomic.AddInt32(&keyCount, 1); return true })
	d.OnMouse(func(e Event) bool { atomic.AddInt32(&mouseCount, 1); return true })
	d.OnResize(func(e Event) bool { atomic.AddInt32(&resizeCount, 1); return true })
	d.OnPaste(func(e Event) bool { atomic.AddInt32(&pasteCount, 1); return true })

	events := []Event{
		{Type: TypeKey, Key: &term.KeyEvent{Rune: 'a'}},
		{Type: TypeMouse, Mouse: &term.MouseEvent{X: 1, Y: 1}},
		{Type: TypeResize, Width: 80, Height: 24},
		{Type: TypePaste, Paste: "text"},
		{Type: TypeKey, Key: &term.KeyEvent{Rune: 'b'}},
		{Type: TypeMouse, Mouse: &term.MouseEvent{X: 2, Y: 2}},
		{Type: TypeResize, Width: 120, Height: 40},
	}

	for _, e := range events {
		d.Dispatch(e)
	}

	if atomic.LoadInt32(&keyCount) != 2 {
		t.Errorf("key count %d, expected 2", keyCount)
	}
	if atomic.LoadInt32(&mouseCount) != 2 {
		t.Errorf("mouse count %d, expected 2", mouseCount)
	}
	if atomic.LoadInt32(&resizeCount) != 2 {
		t.Errorf("resize count %d, expected 2", resizeCount)
	}
	if atomic.LoadInt32(&pasteCount) != 1 {
		t.Errorf("paste count %d, expected 1", pasteCount)
	}
}

// --- KeyShortcut Match edge cases ---

func TestP26B_KeyShortcutMatchNil(t *testing.T) {
	d := NewDispatcher()
	var fired int32

	d.BindKey(PlainRune('a'), func(e Event) bool {
		atomic.StoreInt32(&fired, 1)
		return true
	})

	// Key event with nil Key pointer
	e := Event{Type: TypeKey, Key: nil}
	d.Dispatch(e)

	if atomic.LoadInt32(&fired) != 0 {
		t.Error("handler should not fire for nil key event")
	}
}

// --- FromTermEvent conversion chain ---

func TestP26B_FromTermEventChain(t *testing.T) {
	d := NewDispatcher()
	var keyHandled, resizeHandled int32

	d.OnKey(func(e Event) bool { atomic.StoreInt32(&keyHandled, 1); return true })
	d.OnResize(func(e Event) bool { atomic.StoreInt32(&resizeHandled, 1); return true })

	// Key event
	te1 := term.Event{Type: term.EventKey, Key: &term.KeyEvent{Rune: 'x'}}
	e1 := FromTermEvent(te1)
	d.Dispatch(e1)
	if atomic.LoadInt32(&keyHandled) != 1 {
		t.Error("key not dispatched from term event")
	}

	// Resize event
	te2 := term.Event{Type: term.EventResize, Width: 100, Height: 30}
	e2 := FromTermEvent(te2)
	d.Dispatch(e2)
	if atomic.LoadInt32(&resizeHandled) != 1 {
		t.Error("resize not dispatched from term event")
	}
}

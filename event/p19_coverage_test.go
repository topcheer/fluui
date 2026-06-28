package event

import (
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/term"
)

// === Loop.Run() coverage via real Terminal ===
// These tests use term.Open() which requires /dev/tty.
// On CI/headless, we skip. On interactive, we verify Run/Quit lifecycle.

func TestLoop_RunQuit_Lifecycle(t *testing.T) {
	if testing.Short() || os.Getenv("RUN_TERMINAL_TESTS") == "" {
		t.Skip("skipping terminal test; set RUN_TERMINAL_TESTS=1 to run")
	}

	tm, err := term.Open()
	if err != nil {
		t.Skipf("no terminal available: %v", err)
	}
	defer tm.Close()

	d := NewDispatcher()
	loop := NewLoop(tm, d)

	rendered := atomic.Int32{}
	loop.OnRender(func() bool {
		rendered.Add(1)
		return true
	})

	// Quit from a goroutine after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		loop.Quit()
	}()

	err = loop.Run()
	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}
	if rendered.Load() < 1 {
		t.Error("expected at least one render")
	}
	if loop.running.Load() {
		t.Error("loop should not be running after Run returns")
	}
}

func TestLoop_RunDispatchesCustomEvents(t *testing.T) {
	if testing.Short() || os.Getenv("RUN_TERMINAL_TESTS") == "" {
		t.Skip("skipping terminal test; set RUN_TERMINAL_TESTS=1 to run")
	}

	tm, err := term.Open()
	if err != nil {
		t.Skipf("no terminal available: %v", err)
	}
	defer tm.Close()

	d := NewDispatcher()
	loop := NewLoop(tm, d)

	received := atomic.Int32{}
	d.OnKey(func(e Event) bool {
		received.Add(1)
		return true
	})

	// Send a custom event then quit
	go func() {
		time.Sleep(20 * time.Millisecond)
		loop.Send(Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'a'}})
		time.Sleep(20 * time.Millisecond)
		loop.Quit()
	}()

	_ = loop.Run()
	if received.Load() < 1 {
		t.Error("expected custom event to be dispatched")
	}
}

func TestLoop_RunRendersWhenDirty(t *testing.T) {
	if testing.Short() || os.Getenv("RUN_TERMINAL_TESTS") == "" {
		t.Skip("skipping terminal test; set RUN_TERMINAL_TESTS=1 to run")
	}

	tm, err := term.Open()
	if err != nil {
		t.Skipf("no terminal available: %v", err)
	}
	defer tm.Close()

	d := NewDispatcher()
	loop := NewLoop(tm, d)

	renderCount := atomic.Int32{}
	loop.OnRender(func() bool {
		renderCount.Add(1)
		return true
	})

	go func() {
		// Mark dirty multiple times
		for i := 0; i < 5; i++ {
			time.Sleep(10 * time.Millisecond)
			loop.MarkDirty()
		}
		time.Sleep(20 * time.Millisecond)
		loop.Quit()
	}()

	_ = loop.Run()
	if renderCount.Load() < 2 {
		t.Errorf("expected multiple renders, got %d", renderCount.Load())
	}
}

func TestLoop_QuitIsIdempotent_P19(t *testing.T) {
	d := NewDispatcher()
	loop := NewLoop(nil, d)

	// Quit should not panic when called multiple times
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			loop.Quit()
		}()
	}
	wg.Wait()

	select {
	case <-loop.doneCh:
		// expected
	default:
		t.Error("doneCh should be closed after Quit")
	}
}

func TestLoop_SendAfterQuit(t *testing.T) {
	d := NewDispatcher()
	loop := NewLoop(nil, d)

	loop.Quit()

	// Send after quit should not block
	done := make(chan struct{})
	go func() {
		defer close(done)
		loop.Send(Event{Type: TypeKey})
	}()

	select {
	case <-done:
		// good
	case <-time.After(100 * time.Millisecond):
		t.Error("Send after Quit should not block")
	}
}

func TestLoop_MarkDirtyAfterQuit(t *testing.T) {
	d := NewDispatcher()
	loop := NewLoop(nil, d)

	loop.Quit()
	// Should not panic
	loop.MarkDirty()
}

func TestLoop_DirtyFlag(t *testing.T) {
	d := NewDispatcher()
	loop := NewLoop(nil, d)

	if loop.dirty.Load() {
		t.Error("dirty should start false")
	}

	loop.MarkDirty()
	if !loop.dirty.Load() {
		t.Error("dirty should be true after MarkDirty")
	}
}

func TestLoop_RunningFlag(t *testing.T) {
	d := NewDispatcher()
	loop := NewLoop(nil, d)

	if loop.running.Load() {
		t.Error("running should start false")
	}
}

func TestLoop_OnRender_P19(t *testing.T) {
	d := NewDispatcher()
	loop := NewLoop(nil, d)

	called := false
	loop.OnRender(func() bool {
		called = true
		return true
	})

	if loop.onRender == nil {
		t.Error("onRender should be set")
	}
	if !called {
		// We can't call onRender directly in the test, but verify it was set
		result := loop.onRender()
		if !result || !called {
			t.Error("onRender callback should work")
		}
	}
}

func TestLoop_RenderIfDirty_WithCallback(t *testing.T) {
	d := NewDispatcher()
	loop := NewLoop(nil, d)

	rendered := atomic.Int32{}
	loop.OnRender(func() bool {
		rendered.Add(1)
		return true
	})

	loop.MarkDirty()
	loop.renderIfDirty()
	if rendered.Load() != 1 {
		t.Error("expected render to fire when dirty")
	}

	// Second call should NOT render (dirty was cleared)
	loop.renderIfDirty()
	if rendered.Load() != 1 {
		t.Error("expected no second render (not dirty)")
	}
}

func TestLoop_RenderIfDirty_NoCallback_P19(t *testing.T) {
	d := NewDispatcher()
	loop := NewLoop(nil, d)
	loop.MarkDirty()

	// Should not panic with nil onRender
	loop.renderIfDirty()
}

func TestLoop_RenderIfDirty_NotDirty_P19(t *testing.T) {
	d := NewDispatcher()
	loop := NewLoop(nil, d)

	rendered := atomic.Int32{}
	loop.OnRender(func() bool {
		rendered.Add(1)
		return true
	})

	// Not dirty — should not render
	loop.renderIfDirty()
	if rendered.Load() != 0 {
		t.Error("should not render when not dirty")
	}
}

// === Custom event channel coverage ===

func TestLoop_CustomChannelBuffer(t *testing.T) {
	d := NewDispatcher()
	loop := NewLoop(nil, d)

	// customCh has buffer 64
	for i := 0; i < 64; i++ {
		select {
		case loop.customCh <- Event{Type: TypeKey}:
		default:
			t.Errorf("customCh should accept %d events", i)
		}
	}
}

// === Dispatcher coverage ===

func TestDispatcher_OnResize(t *testing.T) {
	d := NewDispatcher()

	called := false
	d.OnResize(func(e Event) bool {
		called = true
		if e.Width != 100 || e.Height != 40 {
			t.Errorf("expected 100x40, got %dx%d", e.Width, e.Height)
		}
		return true
	})

	result := d.Dispatch(Event{Type: TypeResize, Width: 100, Height: 40})
	if !result {
		t.Error("Dispatch should return true")
	}
	if !called {
		t.Error("OnResize handler should be called")
	}
}

func TestDispatcher_OnResize_NoHandler(t *testing.T) {
	d := NewDispatcher()
	result := d.Dispatch(Event{Type: TypeResize, Width: 80, Height: 24})
	if result {
		t.Error("Dispatch without handler should return false")
	}
}

func TestDispatcher_OnKey_NoHandler(t *testing.T) {
	d := NewDispatcher()
	result := d.Dispatch(Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'x'}})
	if result {
		t.Error("Dispatch without handler should return false")
	}
}

func TestDispatcher_OnMouse_NoHandler(t *testing.T) {
	d := NewDispatcher()
	result := d.Dispatch(Event{Type: TypeMouse, Mouse: &term.MouseEvent{X: 1, Y: 1}})
	if result {
		t.Error("Dispatch without handler should return false")
	}
}

func TestDispatcher_OnPaste(t *testing.T) {
	d := NewDispatcher()

	called := false
	d.OnPaste(func(e Event) bool {
		called = true
		return true
	})

	result := d.Dispatch(Event{Type: TypePaste, Paste: "test"})
	if !result || !called {
		t.Error("OnPaste handler should be called and return true")
	}
}

func TestDispatcher_OnPaste_NoHandler(t *testing.T) {
	d := NewDispatcher()
	result := d.Dispatch(Event{Type: TypePaste, Paste: "test"})
	if result {
		t.Error("should return false without handler")
	}
}

func TestDispatcher_MultipleHandlers(t *testing.T) {
	d := NewDispatcher()

	count := atomic.Int32{}
	d.OnKey(func(e Event) bool {
		count.Add(1)
		return true
	})

	// Only the last registered handler runs (no chain)
	d.Dispatch(Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'a'}})
	if count.Load() != 1 {
		t.Errorf("expected 1 call, got %d", count.Load())
	}
}

// === Event type coverage ===

func TestFromTermEvent_KeyEvent(t *testing.T) {
	key := &term.KeyEvent{Rune: 'x', Modifiers: term.ModCtrl}
	ev := FromTermEvent(term.Event{Type: term.EventKey, Key: key})
	if ev.Type != TypeKey || ev.Key == nil || ev.Key.Rune != 'x' {
		t.Error("FromTermEvent should convert KeyEvent")
	}
}

func TestFromTermEvent_MouseEvent(t *testing.T) {
	me := &term.MouseEvent{X: 5, Y: 10, Button: term.MouseLeft}
	ev := FromTermEvent(term.Event{Type: term.EventMouse, Mouse: me})
	if ev.Type != TypeMouse || ev.Mouse == nil || ev.Mouse.X != 5 {
		t.Error("FromTermEvent should convert MouseEvent")
	}
}

func TestFromTermEvent_ResizeEvent(t *testing.T) {
	ev := FromTermEvent(term.Event{Type: term.EventResize, Width: 120, Height: 50})
	if ev.Type != TypeResize || ev.Width != 120 || ev.Height != 50 {
		t.Error("FromTermEvent should convert ResizeEvent")
	}
}

func TestFromTermEvent_UnknownType(t *testing.T) {
	ev := FromTermEvent(term.Event{})
	if ev.Type != EventType(0) {
		t.Error("Unknown event should map to zero value")
	}
}

func TestEventTypes(t *testing.T) {
	if TypeKey != EventType(0) {
		t.Error("TypeKey should be 0")
	}
	types := []EventType{TypeKey, TypeMouse, TypeResize, TypePaste}
	for _, et := range types {
		if int(et) < 0 {
			t.Errorf("event type %d should be >= 0", et)
		}
	}
}

// === Concurrent stress tests ===

func TestLoop_ConcurrentSendAndQuit(t *testing.T) {
	d := NewDispatcher()
	loop := NewLoop(nil, d)

	var wg sync.WaitGroup
	// Concurrent Senders
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				loop.Send(Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'a'}})
			}
		}()
	}
	// Concurrent Quitters
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			loop.Quit()
		}()
	}
	wg.Wait()
}

func TestLoop_ConcurrentMarkDirtyAndOnRender(t *testing.T) {
	d := NewDispatcher()
	loop := NewLoop(nil, d)

	rendered := atomic.Int32{}
	loop.OnRender(func() bool {
		rendered.Add(1)
		return true
	})

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				loop.MarkDirty()
				if j%100 == 0 {
					loop.renderIfDirty()
				}
			}
		}()
	}
	wg.Wait()
	if rendered.Load() == 0 {
		t.Error("expected some renders")
	}
}

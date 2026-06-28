package event

import (
	"sync"
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/term"
)

// === NewLoop coverage ===

func TestNewLoop_NilTerminal(t *testing.T) {
	d := NewDispatcher()
	loop := NewLoop(nil, d)

	if loop.dispatcher != d {
		t.Error("Dispatcher not set correctly")
	}
	if loop.parser == nil {
		t.Error("Parser should be initialized")
	}
	if loop.rawCh == nil {
		t.Error("rawCh should be initialized")
	}
	if loop.customCh == nil {
		t.Error("customCh should be initialized")
	}
	if loop.doneCh == nil {
		t.Error("doneCh should be initialized")
	}
	if loop.fps != 60 {
		t.Errorf("fps = %d, want 60", loop.fps)
	}
}

func TestNewLoop_Defaults(t *testing.T) {
	d := NewDispatcher()
	loop := NewLoop(nil, d)

	// Verify all fields are properly initialized
	if cap(loop.rawCh) != 16 {
		t.Errorf("rawCh capacity = %d, want 16", cap(loop.rawCh))
	}
	if cap(loop.customCh) != 64 {
		t.Errorf("customCh capacity = %d, want 64", cap(loop.customCh))
	}
}

// === OnRender coverage ===

func TestLoop_OnRender(t *testing.T) {
	loop := &Loop{
		customCh:   make(chan Event, 64),
		rawCh:      make(chan []byte, 16),
		doneCh:     make(chan struct{}),
		fps:        60,
		dispatcher: NewDispatcher(),
	}

	if loop.onRender != nil {
		t.Error("onRender should be nil by default")
	}

	loop.OnRender(func() bool {
		return true
	})

	if loop.onRender == nil {
		t.Error("onRender should be set after OnRender()")
	}
}

// === renderIfDirty coverage ===

func TestLoop_RenderIfDirty_NotDirty(t *testing.T) {
	loop := &Loop{
		customCh:   make(chan Event, 64),
		rawCh:      make(chan []byte, 16),
		doneCh:     make(chan struct{}),
		fps:        60,
		dispatcher: NewDispatcher(),
	}

	renderCalled := false
	loop.OnRender(func() bool {
		renderCalled = true
		return true
	})

	// dirty is false by default → render should not be called
	loop.renderIfDirty()

	if renderCalled {
		t.Error("renderIfDirty should not call render when not dirty")
	}
}

func TestLoop_RenderIfDirty_Dirty(t *testing.T) {
	loop := &Loop{
		customCh:   make(chan Event, 64),
		rawCh:      make(chan []byte, 16),
		doneCh:     make(chan struct{}),
		fps:        60,
		dispatcher: NewDispatcher(),
	}

	renderCount := 0
	var mu sync.Mutex
	loop.OnRender(func() bool {
		mu.Lock()
		renderCount++
		mu.Unlock()
		return true
	})

	// Mark dirty → render should be called
	loop.MarkDirty()
	loop.renderIfDirty()

	mu.Lock()
	if renderCount != 1 {
		t.Errorf("renderCount = %d, want 1", renderCount)
	}
	mu.Unlock()

	// After render, dirty should be false → calling again should not render
	loop.renderIfDirty()
	mu.Lock()
	if renderCount != 1 {
		t.Errorf("renderCount = %d, want 1 (dirty cleared)", renderCount)
	}
	mu.Unlock()
}

func TestLoop_RenderIfDirty_NoCallback(t *testing.T) {
	loop := &Loop{
		customCh:   make(chan Event, 64),
		rawCh:      make(chan []byte, 16),
		doneCh:     make(chan struct{}),
		fps:        60,
		dispatcher: NewDispatcher(),
	}

	// Set dirty but no render callback
	loop.MarkDirty()
	// Should not panic
	loop.renderIfDirty()
}

// === Quit idempotency ===

func TestLoop_QuitIdempotent(t *testing.T) {
	loop := &Loop{
		customCh:   make(chan Event, 64),
		rawCh:      make(chan []byte, 16),
		doneCh:     make(chan struct{}),
		fps:        60,
		dispatcher: NewDispatcher(),
	}

	// First Quit should close doneCh
	loop.Quit()
	select {
	case <-loop.doneCh:
		// Good
	default:
		t.Error("doneCh should be closed after Quit()")
	}

	// Second Quit should be a no-op (not panic)
	loop.Quit()
	loop.Quit()
}

// === MarkDirty + Send interaction ===

func TestLoop_MarkDirtySetsFlag(t *testing.T) {
	loop := &Loop{
		customCh:   make(chan Event, 64),
		rawCh:      make(chan []byte, 16),
		doneCh:     make(chan struct{}),
		fps:        60,
		dispatcher: NewDispatcher(),
	}

	if loop.dirty.Load() {
		t.Error("dirty should be false initially")
	}

	loop.MarkDirty()
	if !loop.dirty.Load() {
		t.Error("dirty should be true after MarkDirty()")
	}
}

// === Send after Quit ===

func TestLoop_SendAfterQuitDoesNotBlock(t *testing.T) {
	loop := &Loop{
		customCh:   make(chan Event, 64),
		rawCh:      make(chan []byte, 16),
		doneCh:     make(chan struct{}),
		fps:        60,
		dispatcher: NewDispatcher(),
	}

	loop.Quit()

	// Send after quit should return quickly (not block)
	done := make(chan struct{})
	go func() {
		loop.Send(Event{Type: TypeKey})
		close(done)
	}()

	select {
	case <-done:
		// Good — didn't block
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Send after Quit should not block")
	}
}

// === Send normal operation ===

func TestLoop_SendDeliversEvent(t *testing.T) {
	loop := &Loop{
		customCh:   make(chan Event, 64),
		rawCh:      make(chan []byte, 16),
		doneCh:     make(chan struct{}),
		fps:        60,
		dispatcher: NewDispatcher(),
	}

	ev := Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'x'}}
	loop.Send(ev)

	select {
	case received := <-loop.customCh:
		if received.Type != TypeKey {
			t.Errorf("Received type = %v, want TypeKey", received.Type)
		}
		if received.Key == nil || received.Key.Rune != 'x' {
			t.Errorf("Received key = %v, want rune 'x'", received.Key)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Send should deliver event to customCh")
	}
}

// === Concurrent MarkDirty + renderIfDirty ===

func TestLoop_ConcurrentMarkDirtyAndRender(t *testing.T) {
	loop := &Loop{
		customCh:   make(chan Event, 64),
		rawCh:      make(chan []byte, 16),
		doneCh:     make(chan struct{}),
		fps:        60,
		dispatcher: NewDispatcher(),
	}

	var renderCount int64
	var renderMu sync.Mutex
	loop.OnRender(func() bool {
		renderMu.Lock()
		renderCount++
		renderMu.Unlock()
		return true
	})

	done := make(chan struct{})

	// Writer: MarkDirty
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				loop.MarkDirty()
			}
		}
	}()

	// Reader: renderIfDirty
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				loop.renderIfDirty()
			}
		}
	}()

	time.Sleep(50 * time.Millisecond)
	close(done)
}

// === Multiple events sent ===

func TestLoop_SendMultipleEvents(t *testing.T) {
	loop := &Loop{
		customCh:   make(chan Event, 64),
		rawCh:      make(chan []byte, 16),
		doneCh:     make(chan struct{}),
		fps:        60,
		dispatcher: NewDispatcher(),
	}

	for i := 0; i < 10; i++ {
		loop.Send(Event{Type: TypeKey, Key: &term.KeyEvent{Rune: rune('a' + i)}})
	}

	if len(loop.customCh) != 10 {
		t.Errorf("customCh length = %d, want 10", len(loop.customCh))
	}
}

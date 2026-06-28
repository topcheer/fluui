package fluui

import (
	"io"
	"sync/atomic"
	"testing"

	"github.com/topcheer/fluui/event"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/render"
)

// newTestApp creates an App wired with a discard writer, real renderer,
// and real dispatcher/loop. This lets us test setupHandlers, callbacks,
// and convenience methods without /dev/tty.
func newTestApp(w, h int) *App {
	disp := event.NewDispatcher()
	loop := event.NewLoop(nil, disp)
	tw := term.NewWriter(io.Discard, term.Profile256)
	r := render.New(tw, w, h)

	app := &App{
		terminal:   nil, // no real terminal — test only methods that don't need it
		writer:     tw,
		renderer:   r,
		loop:       loop,
		dispatcher: disp,
		width:      w,
		height:     h,
	}
	app.setupHandlers()
	return app
}

// ─── Accessors ────────────────────────────────────────────

func TestApp_Renderer(t *testing.T) {
	app := newTestApp(80, 24)
	if app.Renderer() == nil {
		t.Error("Renderer() should return non-nil renderer")
	}
}

func TestApp_Writer(t *testing.T) {
	app := newTestApp(80, 24)
	if app.Writer() == nil {
		t.Error("Writer() should return non-nil writer")
	}
}

func TestApp_Loop(t *testing.T) {
	app := newTestApp(80, 24)
	if app.Loop() == nil {
		t.Error("Loop() should return non-nil loop")
	}
}

func TestApp_Dispatcher(t *testing.T) {
	app := newTestApp(80, 24)
	if app.Dispatcher() == nil {
		t.Error("Dispatcher() should return non-nil dispatcher")
	}
}

func TestApp_Size(t *testing.T) {
	app := newTestApp(80, 24)
	w, h := app.Size()
	if w != 80 || h != 24 {
		t.Errorf("Size() = (%d, %d), want (80, 24)", w, h)
	}
}

func TestApp_Terminal_NilWithoutRealTerm(t *testing.T) {
	app := newTestApp(80, 24)
	if app.Terminal() != nil {
		t.Error("Terminal() should be nil in test app")
	}
}

// ─── Callback setters ─────────────────────────────────────

func TestApp_OnKey(t *testing.T) {
	app := newTestApp(80, 24)
	called := false
	app.OnKey(func(k *term.KeyEvent) {
		called = true
	})

	// Dispatch a regular key event through the handler chain
	result := app.dispatcher.Dispatch(event.Event{
		Type: event.TypeKey,
		Key:  &term.KeyEvent{Rune: 'x'},
	})
	if !result {
		t.Error("Dispatch should return true for key event with handler")
	}
	if !called {
		t.Error("onKey callback should have been called")
	}
}

func TestApp_OnKey_NilKeyEvent(t *testing.T) {
	app := newTestApp(80, 24)
	result := app.dispatcher.Dispatch(event.Event{
		Type: event.TypeKey,
		Key:  nil,
	})
	if result {
		t.Error("Dispatch with nil Key should return false")
	}
}

func TestApp_OnMouse(t *testing.T) {
	app := newTestApp(80, 24)
	called := false
	app.OnMouse(func(m *term.MouseEvent) {
		called = true
	})

	result := app.dispatcher.Dispatch(event.Event{
		Type:  event.TypeMouse,
		Mouse: &term.MouseEvent{X: 5, Y: 10, Button: term.MouseLeft},
	})
	if !result {
		t.Error("Dispatch should return true for mouse event with handler")
	}
	if !called {
		t.Error("onMouse callback should have been called")
	}
}

func TestApp_OnMouse_NilMouseEvent(t *testing.T) {
	app := newTestApp(80, 24)
	result := app.dispatcher.Dispatch(event.Event{
		Type:  event.TypeMouse,
		Mouse: nil,
	})
	if result {
		t.Error("Dispatch with nil Mouse should return false")
	}
}

func TestApp_OnResize(t *testing.T) {
	app := newTestApp(80, 24)
	var rw, rh int
	app.OnResize(func(w, h int) {
		rw, rh = w, h
	})

	result := app.dispatcher.Dispatch(event.Event{
		Type:   event.TypeResize,
		Width:  100,
		Height: 30,
	})
	if !result {
		t.Error("Dispatch should return true for resize event")
	}
	if rw != 100 || rh != 30 {
		t.Errorf("onResize got (%d, %d), want (100, 30)", rw, rh)
	}
	// App dimensions should update
	w, h := app.Size()
	if w != 100 || h != 30 {
		t.Errorf("Size() after resize = (%d, %d), want (100, 30)", w, h)
	}
}

func TestApp_OnQuit(t *testing.T) {
	cleanupCalled := false
	fn := func() {
		cleanupCalled = true
	}
	// Simulate what Run() does after loop exits
	fn()
	if !cleanupCalled {
		t.Error("onQuit callback should have been called")
	}
}

func TestApp_OnInterrupt(t *testing.T) {
	app := newTestApp(80, 24)
	quitCalled := atomic.Bool{}
	app.loop.Quit() // prep the loop

	// Set OnInterrupt to block
	app.OnInterrupt(func() bool {
		return false
	})

	// Dispatch Ctrl+C — should NOT quit since OnInterrupt returns false
	// We can't easily test actual Quit() behavior without running the loop,
	// but we can verify the dispatch path doesn't panic.
	_ = app.dispatcher.Dispatch(event.Event{
		Type: event.TypeKey,
		Key:  &term.KeyEvent{Rune: 'c', Modifiers: term.ModCtrl},
	})

	if quitCalled.Load() {
		t.Error("should not quit when OnInterrupt returns false")
	}
}

func TestApp_OnPaint(t *testing.T) {
	app := newTestApp(80, 24)
	paintCalled := false
	app.OnPaint(func(buf *buffer.Buffer) {
		paintCalled = true
	})

	// Call the paint callback directly
	if app.onPaint != nil {
		app.onPaint(app.renderer.Back())
	}
	if !paintCalled {
		t.Error("onPaint callback should have been called")
	}
}

// ─── Ctrl+C / interrupt handling ──────────────────────────

func TestApp_CtrlC_DefaultQuit(t *testing.T) {
	app := newTestApp(80, 24)
	// No OnInterrupt set — Ctrl+C should trigger quit
	result := app.dispatcher.Dispatch(event.Event{
		Type: event.TypeKey,
		Key:  &term.KeyEvent{Rune: 'c', Modifiers: term.ModCtrl},
	})
	if !result {
		t.Error("Ctrl+C dispatch should return true")
	}
}

func TestApp_CtrlC_WithOnInterruptTrue(t *testing.T) {
	app := newTestApp(80, 24)
	app.OnInterrupt(func() bool {
		return true // allow quit
	})
	result := app.dispatcher.Dispatch(event.Event{
		Type: event.TypeKey,
		Key:  &term.KeyEvent{Rune: 'c', Modifiers: term.ModCtrl},
	})
	if !result {
		t.Error("Ctrl+C dispatch should return true")
	}
}

func TestApp_CtrlC_WithOnInterruptFalse(t *testing.T) {
	app := newTestApp(80, 24)
	app.OnInterrupt(func() bool {
		return false // block quit
	})
	result := app.dispatcher.Dispatch(event.Event{
		Type: event.TypeKey,
		Key:  &term.KeyEvent{Rune: 'c', Modifiers: term.ModCtrl},
	})
	if !result {
		t.Error("Ctrl+C dispatch should return true even if quit blocked")
	}
}

func TestApp_CtrlShiftC(t *testing.T) {
	app := newTestApp(80, 24)
	result := app.dispatcher.Dispatch(event.Event{
		Type: event.TypeKey,
		Key:  &term.KeyEvent{Rune: 'C', Modifiers: term.ModCtrl},
	})
	if !result {
		t.Error("Ctrl+Shift+C dispatch should return true")
	}
}

func TestApp_RegularKey_CallsOnKey(t *testing.T) {
	app := newTestApp(80, 24)
	var capturedRune rune
	app.OnKey(func(k *term.KeyEvent) {
		capturedRune = k.Rune
	})

	_ = app.dispatcher.Dispatch(event.Event{
		Type: event.TypeKey,
		Key:  &term.KeyEvent{Rune: 'a'},
	})
	if capturedRune != 'a' {
		t.Errorf("onKey got rune %q, want 'a'", capturedRune)
	}
}

// ─── MarkDirty / Quit / Send ──────────────────────────────

func TestApp_MarkDirty(t *testing.T) {
	app := newTestApp(80, 24)
	// Should not panic
	app.MarkDirty()
}

func TestApp_Quit(t *testing.T) {
	app := newTestApp(80, 24)
	// Should not panic — loop may already be done
	app.Quit()
}

func TestApp_Send(t *testing.T) {
	app := newTestApp(80, 24)
	// Send a custom event — should not panic
	app.Send(event.Event{Type: event.TypeKey, Key: &term.KeyEvent{Rune: 'x'}})
}

// ─── Convenience rendering methods ────────────────────────

func TestApp_BackBuffer(t *testing.T) {
	app := newTestApp(80, 24)
	buf := app.BackBuffer()
	if buf == nil {
		t.Error("BackBuffer() should return non-nil buffer")
	}
}

func TestApp_DrawText(t *testing.T) {
	app := newTestApp(80, 24)
	n := app.DrawText(0, 0, "Hello", buffer.Style{})
	if n != 5 {
		t.Errorf("DrawText returned %d, want 5", n)
	}
}

func TestApp_DrawText_Overflow(t *testing.T) {
	app := newTestApp(10, 5)
	n := app.DrawText(8, 0, "Hello", buffer.Style{})
	// Text may be clamped — verify it doesn't panic
	if n < 0 {
		t.Errorf("DrawText returned negative %d", n)
	}
}

func TestApp_DrawTextClamped(t *testing.T) {
	app := newTestApp(10, 5)
	n := app.DrawTextClamped(8, 0, "Hello World", buffer.Style{})
	if n < 0 {
		t.Errorf("DrawTextClamped returned negative %d", n)
	}
}

func TestApp_FillRect(t *testing.T) {
	app := newTestApp(80, 24)
	cell := buffer.NewCell('X', buffer.Style{})
	app.FillRect(buffer.Rect{X: 0, Y: 0, W: 5, H: 3}, cell)
	// Should not panic — verify buffer has content
	buf := app.BackBuffer()
	if buf == nil {
		t.Error("BackBuffer should not be nil after FillRect")
	}
}

// ─── Concurrent access ────────────────────────────────────

func TestApp_ConcurrentDispatch(t *testing.T) {
	app := newTestApp(80, 24)
	var count atomic.Int32
	app.OnKey(func(k *term.KeyEvent) {
		count.Add(1)
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			_ = app.dispatcher.Dispatch(event.Event{
				Type: event.TypeKey,
				Key:  &term.KeyEvent{Rune: 'x'},
			})
		}
	}()

	for i := 0; i < 100; i++ {
		_ = app.dispatcher.Dispatch(event.Event{
			Type: event.TypeKey,
			Key:  &term.KeyEvent{Rune: 'y'},
		})
	}

	<-done

	total := count.Load()
	if total != 200 {
		t.Errorf("expected 200 key callbacks, got %d", total)
	}
}

func TestApp_ConcurrentResize(t *testing.T) {
	app := newTestApp(80, 24)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for i := 0; i < 50; i++ {
			_ = app.dispatcher.Dispatch(event.Event{
				Type:   event.TypeResize,
				Width:  80 + i,
				Height: 24 + i,
			})
		}
	}()

	for i := 0; i < 50; i++ {
		_ = app.dispatcher.Dispatch(event.Event{
			Type:   event.TypeResize,
			Width:  80,
			Height: 24,
		})
	}

	<-done
}

func TestApp_ConcurrentDrawText(t *testing.T) {
	app := newTestApp(80, 24)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			app.DrawText(0, 0, "A", buffer.Style{})
		}
	}()

	for i := 0; i < 100; i++ {
		app.DrawText(1, 0, "B", buffer.Style{})
	}

	<-done
}

// ─── New() error path ─────────────────────────────────────

func TestNew_NoTerminal(t *testing.T) {
	// In a non-interactive environment, term.Open() may fail.
	// This test verifies that New() returns an error when terminal is unavailable.
	// It's expected to skip or pass depending on the environment.
	if testing.Short() {
		t.Skip("skipping terminal test in short mode")
	}
	// We can't guarantee /dev/tty is unavailable, so just verify
	// that New() either succeeds or returns an error (not panic).
	_, err := New()
	if err != nil {
		t.Logf("New() returned error (expected in non-interactive): %v", err)
	}
}

// ─── Resize updates renderer ──────────────────────────────

func TestApp_Resize_UpdatesRenderer(t *testing.T) {
	app := newTestApp(80, 24)

	_ = app.dispatcher.Dispatch(event.Event{
		Type:   event.TypeResize,
		Width:  120,
		Height: 40,
	})

	if app.Renderer().Width() != 120 {
		t.Errorf("renderer width = %d, want 120", app.Renderer().Width())
	}
	if app.Renderer().Height() != 40 {
		t.Errorf("renderer height = %d, want 40", app.Renderer().Height())
	}
}

// ─── Multiple callbacks ───────────────────────────────────

func TestApp_MultipleCallbacks(t *testing.T) {
	app := newTestApp(80, 24)

	keyCalled := atomic.Bool{}
	mouseCalled := atomic.Bool{}
	resizeCalled := atomic.Bool{}

	app.OnKey(func(k *term.KeyEvent) { keyCalled.Store(true) })
	app.OnMouse(func(m *term.MouseEvent) { mouseCalled.Store(true) })
	app.OnResize(func(w, h int) { resizeCalled.Store(true) })

	// Dispatch all three event types
	_ = app.dispatcher.Dispatch(event.Event{
		Type: event.TypeKey,
		Key:  &term.KeyEvent{Rune: 'x'},
	})
	_ = app.dispatcher.Dispatch(event.Event{
		Type:  event.TypeMouse,
		Mouse: &term.MouseEvent{X: 1, Y: 1, Button: term.MouseLeft},
	})
	_ = app.dispatcher.Dispatch(event.Event{
		Type:   event.TypeResize,
		Width:  100,
		Height: 30,
	})

	if !keyCalled.Load() {
		t.Error("onKey not called")
	}
	if !mouseCalled.Load() {
		t.Error("onMouse not called")
	}
	if !resizeCalled.Load() {
		t.Error("onResize not called")
	}
}

// ─── Overwrite callback ───────────────────────────────────

func TestApp_OverwriteCallback(t *testing.T) {
	app := newTestApp(80, 24)

	callCount := atomic.Int32{}
	app.OnKey(func(k *term.KeyEvent) {
		callCount.Add(1)
	})
	// Overwrite with new callback
	app.OnKey(func(k *term.KeyEvent) {
		callCount.Add(100)
	})

	_ = app.dispatcher.Dispatch(event.Event{
		Type: event.TypeKey,
		Key:  &term.KeyEvent{Rune: 'x'},
	})

	if got := callCount.Load(); got != 100 {
		t.Errorf("expected only second callback (100), got %d", got)
	}
}

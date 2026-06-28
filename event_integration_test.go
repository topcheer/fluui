package fluui

import (
	"bytes"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/topcheer/fluui/event"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/render"
)

// newIntegrationApp creates an App with a blocking terminal (pipe reader)
// and a bytes.Buffer writer for verifying output. The app is ready for
// integration tests that exercise the full event chain.
func newIntegrationApp(w, h int) (*App, io.WriteCloser, *bytes.Buffer) {
	var buf bytes.Buffer
	r, pipeW := io.Pipe()
	tm := term.NewTestTerminal(r, &buf, w, h)
	tw := term.NewWriter(tm, tm.ColorProfile())
	rend := render.New(tw, w, h)
	disp := event.NewDispatcher()
	loop := event.NewLoop(tm, disp)

	app := &App{
		terminal:   tm,
		writer:     tw,
		renderer:   rend,
		loop:       loop,
		dispatcher: disp,
		width:      w,
		height:     h,
	}
	app.setupHandlers()
	return app, pipeW, &buf
}

// ─── Key Event Chain Integration ──────────────────────────
// Tests: raw bytes → parser → dispatcher → handler → state change → render

func TestIntegration_KeyEvent_OnKeyCallback(t *testing.T) {
	app, pipeW, _ := newIntegrationApp(80, 24)
	keyReceived := atomic.Bool{}
	var receivedKey term.KeyEvent

	app.OnKey(func(key *term.KeyEvent) {
		keyReceived.Store(true)
		receivedKey = *key
	})

	done := make(chan error, 1)
	go func() { done <- app.Run() }()

	// Write a plain 'a' key to the terminal
	pipeW.Write([]byte("a"))

	deadline := time.After(2 * time.Second)
	for !keyReceived.Load() {
		select {
		case <-deadline:
			app.Quit()
			pipeW.Close()
			<-done
			t.Fatal("timeout: key event not received")
		default:
			time.Sleep(time.Millisecond)
		}
	}

	app.Quit()
	pipeW.Close()
	<-done

	if receivedKey.Rune != 'a' {
		t.Errorf("received rune = %q, want 'a'", receivedKey.Rune)
	}
}

func TestIntegration_KeyEvent_CtrlC_Quit(t *testing.T) {
	app, pipeW, _ := newIntegrationApp(80, 24)

	done := make(chan error, 1)
	go func() { done <- app.Run() }()

	// Write Ctrl+C (byte 0x03)
	pipeW.Write([]byte{0x03})

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		app.Quit()
		pipeW.Close()
		t.Fatal("Ctrl+C did not cause Run() to exit")
	}

	pipeW.Close()
}

func TestIntegration_KeyEvent_CtrlC_InterruptHandler(t *testing.T) {
	app, pipeW, _ := newIntegrationApp(80, 24)
	interruptCalled := atomic.Bool{}

	app.OnInterrupt(func() bool {
		interruptCalled.Store(true)
		return false // don't quit
	})

	done := make(chan error, 1)
	go func() { done <- app.Run() }()

	// Write Ctrl+C
	pipeW.Write([]byte{0x03})

	deadline := time.After(2 * time.Second)
	for !interruptCalled.Load() {
		select {
		case <-deadline:
			app.Quit()
			pipeW.Close()
			<-done
			t.Fatal("timeout: interrupt handler not called")
		default:
			time.Sleep(time.Millisecond)
		}
	}

	// App should still be running since we returned false
	app.Quit()
	pipeW.Close()
	<-done

	if !interruptCalled.Load() {
		t.Error("interrupt handler should have been called")
	}
}

func TestIntegration_KeyEvent_MarksDirty(t *testing.T) {
	app, pipeW, _ := newIntegrationApp(80, 24)
	paintCount := atomic.Int32{}

	app.OnKey(func(key *term.KeyEvent) {
		// Key handler registered — setupHandlers will call MarkDirty
	})
	app.OnPaint(func(buf *buffer.Buffer) {
		paintCount.Add(1)
	})

	done := make(chan error, 1)
	go func() { done <- app.Run() }()

	// Wait for initial render, then capture count
	time.Sleep(50 * time.Millisecond)
	initialCount := paintCount.Load()

	// Send a key event — should trigger dirty → render
	pipeW.Write([]byte("x"))

	deadline := time.After(2 * time.Second)
	for paintCount.Load() <= initialCount {
		select {
		case <-deadline:
			app.Quit()
			pipeW.Close()
			<-done
			t.Fatalf("timeout: render not triggered after key event (initial=%d, current=%d)", initialCount, paintCount.Load())
		default:
			time.Sleep(time.Millisecond)
		}
	}

	app.Quit()
	pipeW.Close()
	<-done
}

func TestIntegration_KeyEvent_BoundKeyShortcut(t *testing.T) {
	app, pipeW, _ := newIntegrationApp(80, 24)
	handlerCalled := atomic.Bool{}

	// Bind Ctrl+Q to a handler
	app.dispatcher.BindKey(event.CtrlRune('q'), func(e event.Event) bool {
		handlerCalled.Store(true)
		app.Quit()
		return true
	})

	done := make(chan error, 1)
	go func() { done <- app.Run() }()

	// Write Ctrl+Q (byte 0x11)
	pipeW.Write([]byte{0x11})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		app.Quit()
		pipeW.Close()
		t.Fatal("timeout: Ctrl+Q did not fire")
	}

	pipeW.Close()

	if !handlerCalled.Load() {
		t.Error("bound Ctrl+Q handler should have been called")
	}
}

// ─── Resize Event Chain Integration ───────────────────────
// Tests: resize event → dispatcher → handler → renderer resize → render

func TestIntegration_Resize_UpdatesDimensions(t *testing.T) {
	app, pipeW, _ := newIntegrationApp(80, 24)
	resizeCalled := atomic.Bool{}

	app.OnResize(func(w, h int) {
		resizeCalled.Store(true)
	})

	done := make(chan error, 1)
	go func() { done <- app.Run() }()

	// Send resize event via loop
	app.Send(event.Event{
		Type:   event.TypeResize,
		Width:  120,
		Height: 40,
	})

	deadline := time.After(2 * time.Second)
	for !resizeCalled.Load() {
		select {
		case <-deadline:
			app.Quit()
			pipeW.Close()
			<-done
			t.Fatal("timeout: resize callback not called")
		default:
			time.Sleep(time.Millisecond)
		}
	}

	app.Quit()
	pipeW.Close()
	<-done

	w, h := app.Size()
	if w != 120 || h != 40 {
		t.Errorf("Size() = (%d, %d), want (120, 40)", w, h)
	}
}

func TestIntegration_Resize_RendererUpdated(t *testing.T) {
	app, pipeW, _ := newIntegrationApp(80, 24)

	done := make(chan error, 1)
	go func() { done <- app.Run() }()

	// Send resize event
	app.Send(event.Event{
		Type:   event.TypeResize,
		Width:  100,
		Height: 30,
	})

	time.Sleep(100 * time.Millisecond)
	app.Quit()
	pipeW.Close()
	<-done

	// Verify app dimensions updated
	w, h := app.Size()
	if w != 100 {
		t.Errorf("app width = %d, want 100", w)
	}
	if h != 30 {
		t.Errorf("app height = %d, want 30", h)
	}
}

// ─── Render Cycle Integration ─────────────────────────────
// Tests: dirty → render → OnPaint → terminal output

func TestIntegration_Render_OnPaintReceivesBuffer(t *testing.T) {
	app, pipeW, _ := newIntegrationApp(80, 24)
	paintCount := atomic.Int32{}

	app.OnPaint(func(buf *buffer.Buffer) {
		paintCount.Add(1)
	})

	done := make(chan error, 1)
	go func() { done <- app.Run() }()

	// Wait for at least one render
	deadline := time.After(2 * time.Second)
	for paintCount.Load() == 0 {
		select {
		case <-deadline:
			app.Quit()
			pipeW.Close()
			<-done
			t.Fatal("timeout: no render occurred")
		default:
			time.Sleep(time.Millisecond)
		}
	}

	app.Quit()
	pipeW.Close()
	<-done

	if paintCount.Load() == 0 {
		t.Error("OnPaint should have been called at least once")
	}
}

func TestIntegration_Render_OutputWrittenToTerminal(t *testing.T) {
	app, pipeW, buf := newIntegrationApp(40, 5)

	// Draw something specific on paint
	app.OnPaint(func(b *buffer.Buffer) {
		b.DrawText(0, 0, "INTEGRATION", buffer.Style{})
	})

	done := make(chan error, 1)
	go func() { done <- app.Run() }()

	// Wait for render
	time.Sleep(150 * time.Millisecond)
	app.Quit()
	pipeW.Close()
	<-done

	// Terminal output has escape sequences between chars, so check for
	// individual letters that make up 'INTEGRATION'
	output := buf.String()
	for _, ch := range "INTEGRATION" {
		if !strings.ContainsRune(output, ch) {
			shown := output
			if len(shown) > 100 {
				shown = shown[:100]
			}
			t.Errorf("terminal output should contain %q, got %q (len=%d)", ch, shown, len(output))
		}
	}
}

// ─── Concurrent Event Processing Integration ──────────────
// Tests: multiple events processed concurrently without data races

func TestIntegration_ConcurrentKeys(t *testing.T) {
	app, pipeW, _ := newIntegrationApp(80, 24)
	keyCount := atomic.Int32{}

	app.OnKey(func(key *term.KeyEvent) {
		keyCount.Add(1)
	})

	done := make(chan error, 1)
	go func() { done <- app.Run() }()

	// Send 20 keys rapidly
	go func() {
		for i := 0; i < 20; i++ {
			pipeW.Write([]byte{byte('a' + i%26)})
			time.Sleep(2 * time.Millisecond)
		}
	}()

	// Wait for processing
	time.Sleep(500 * time.Millisecond)
	app.Quit()
	pipeW.Close()
	<-done

	if keyCount.Load() < 10 {
		t.Errorf("expected at least 10 keys processed, got %d", keyCount.Load())
	}
}

func TestIntegration_ConcurrentKeysAndResize(t *testing.T) {
	app, pipeW, _ := newIntegrationApp(80, 24)
	keyCount := atomic.Int32{}
	resizeCount := atomic.Int32{}

	app.OnKey(func(key *term.KeyEvent) {
		keyCount.Add(1)
	})
	app.OnResize(func(w, h int) {
		resizeCount.Add(1)
	})

	done := make(chan error, 1)
	go func() { done <- app.Run() }()

	// Interleave key events and resize events
	go func() {
		for i := 0; i < 10; i++ {
			pipeW.Write([]byte{'k'})
			time.Sleep(5 * time.Millisecond)
			app.Send(event.Event{
				Type:   event.TypeResize,
				Width:  80 + i,
				Height: 24,
			})
			time.Sleep(5 * time.Millisecond)
		}
	}()

	time.Sleep(500 * time.Millisecond)
	app.Quit()
	pipeW.Close()
	<-done

	if keyCount.Load() == 0 {
		t.Error("no keys were processed")
	}
	if resizeCount.Load() == 0 {
		t.Error("no resize events were processed")
	}
}

// ─── Full Lifecycle Integration ───────────────────────────
// Tests: Run → process events → Quit → clean shutdown

func TestIntegration_FullLifecycle(t *testing.T) {
	app, pipeW, buf := newIntegrationApp(80, 24)

	quitCalled := atomic.Bool{}
	app.OnQuit(func() {
		quitCalled.Store(true)
	})

	paintCount := atomic.Int32{}
	app.OnPaint(func(b *buffer.Buffer) {
		paintCount.Add(1)
	})

	done := make(chan error, 1)
	go func() { done <- app.Run() }()

	// 1. Initial render happens
	time.Sleep(100 * time.Millisecond)

	// 2. Send some keys
	pipeW.Write([]byte("hello"))

	// 3. Wait for processing
	time.Sleep(100 * time.Millisecond)

	// 4. Clean shutdown
	app.Quit()
	pipeW.Close()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not exit cleanly")
	}

	// Verify lifecycle
	if !quitCalled.Load() {
		t.Error("OnQuit should have been called")
	}
	if paintCount.Load() == 0 {
		t.Error("OnPaint should have been called at least once")
	}
	if buf.Len() == 0 {
		t.Error("terminal output should not be empty")
	}
}

func TestIntegration_CloseAfterRun(t *testing.T) {
	app, pipeW, _ := newIntegrationApp(80, 24)

	done := make(chan error, 1)
	go func() { done <- app.Run() }()

	time.Sleep(100 * time.Millisecond)
	app.Quit()
	pipeW.Close()
	<-done

	// Close after Run returns should not panic
	err := app.Close()
	if err != nil {
		t.Errorf("Close() after Run() error: %v", err)
	}
}

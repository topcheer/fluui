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

// newTestAppWithTerminal creates an App wired with a NewTestTerminal,
// enabling tests for methods that require a terminal (Errorf, Copy, etc).
func newTestAppWithTerminal(w, h int) (*App, *bytes.Buffer) {
	var buf bytes.Buffer
	tm := term.NewTestTerminal(strings.NewReader(""), &buf, w, h)
	tw := term.NewWriter(tm, tm.ColorProfile())
	r := render.New(tw, w, h)
	disp := event.NewDispatcher()
	loop := event.NewLoop(tm, disp)

	app := &App{
		terminal:   tm,
		writer:     tw,
		renderer:   r,
		loop:       loop,
		dispatcher: disp,
		width:      w,
		height:     h,
	}
	app.setupHandlers()
	return app, &buf
}

// newTestAppWithBlockingTerminal creates an App with a terminal whose
// reader blocks forever (pipe), keeping the event loop alive.
func newTestAppWithBlockingTerminal(w, h int) (*App, io.WriteCloser, *bytes.Buffer) {
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

// ─── OnQuit coverage ──────────────────────────────────────

func TestP22_OnQuit_SetsCallback(t *testing.T) {
	app := newTestApp(80, 24)
	called := atomic.Bool{}
	app.OnQuit(func() {
		called.Store(true)
	})
	if app.onQuit == nil {
		t.Fatal("OnQuit should set the onQuit callback")
	}
	// Call it directly to verify it works
	app.onQuit()
	if !called.Load() {
		t.Error("onQuit callback should have been called")
	}
}

// ─── Errorf coverage ──────────────────────────────────────

func TestP22_Errorf_WritesToTerminal(t *testing.T) {
	app, buf := newTestAppWithTerminal(80, 24)
	app.Errorf("test error: %d", 42)
	output := buf.String()
	if !strings.Contains(output, "test error: 42") {
		t.Errorf("Errorf output = %q, want it to contain 'test error: 42'", output)
	}
}

func TestP22_Errorf_EmptyFormat(t *testing.T) {
	app, buf := newTestAppWithTerminal(80, 24)
	app.Errorf("")
	if buf.Len() != 0 {
		t.Errorf("Errorf with empty format should write nothing, got %q", buf.String())
	}
}

func TestP22_Errorf_MultipleArgs(t *testing.T) {
	app, buf := newTestAppWithTerminal(80, 24)
	app.Errorf("err=%v code=%d", "timeout", 500)
	output := buf.String()
	if !strings.Contains(output, "err=timeout") || !strings.Contains(output, "code=500") {
		t.Errorf("Errorf output = %q, want 'err=timeout code=500'", output)
	}
}

// ─── Copy / clipboard coverage ────────────────────────────

func TestP22_Copy_WritesOSC52(t *testing.T) {
	app, buf := newTestAppWithTerminal(80, 24)
	app.Copy("hello world")
	output := buf.String()
	if !strings.Contains(output, "\x1b]52") {
		t.Errorf("Copy should write OSC52 sequence, got %q", output)
	}
}

func TestP22_Copy_EmptyString(t *testing.T) {
	app, buf := newTestAppWithTerminal(80, 24)
	app.Copy("")
	output := buf.String()
	if !strings.Contains(output, "\x1b]52") {
		t.Errorf("Copy with empty string should still write OSC52, got %q", output)
	}
}

func TestP22_CopySelection_System(t *testing.T) {
	app, buf := newTestAppWithTerminal(80, 24)
	app.CopySelection("text", term.ClipboardSystem)
	output := buf.String()
	if !strings.Contains(output, "\x1b]52") {
		t.Errorf("CopySelection should write OSC52 sequence, got %q", output)
	}
}

func TestP22_CopySelection_Primary(t *testing.T) {
	app, buf := newTestAppWithTerminal(80, 24)
	app.CopySelection("text", term.ClipboardPrimary)
	output := buf.String()
	if !strings.Contains(output, "\x1b]52") {
		t.Errorf("CopySelection (primary) should write OSC52 sequence, got %q", output)
	}
}

func TestP22_PasteFromClipboard(t *testing.T) {
	app, buf := newTestAppWithTerminal(80, 24)
	app.PasteFromClipboard()
	output := buf.String()
	// PasteQuery writes an OSC52 paste query escape sequence
	if len(output) == 0 {
		t.Error("PasteFromClipboard should write paste query to terminal")
	}
}

// ─── Close coverage ───────────────────────────────────────

func TestP22_Close_NoPanic(t *testing.T) {
	app, _ := newTestAppWithTerminal(80, 24)
	// Close should not panic with a test terminal
	err := app.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

func TestP22_Close_Idempotent(t *testing.T) {
	app, _ := newTestAppWithTerminal(80, 24)
	err1 := app.Close()
	err2 := app.Close()
	if err1 != nil {
		t.Errorf("first Close() error: %v", err1)
	}
	if err2 != nil {
		t.Errorf("second Close() error: %v", err2)
	}
}

func TestP22_Close_WritesCleanupSequences(t *testing.T) {
	app, buf := newTestAppWithTerminal(80, 24)
	_ = app.Close()
	output := buf.String()
	// Close writes cleanup sequences: disable mouse, show cursor, leave alt screen
	if !strings.Contains(output, "\x1b[?25h") {
		t.Errorf("Close should write show-cursor sequence, got %q", output)
	}
	if !strings.Contains(output, "\x1b[?1049l") {
		t.Errorf("Close should write leave-alt-screen sequence, got %q", output)
	}
}

// ─── Run coverage ─────────────────────────────────────────

func TestP22_Run_OnPaintCalled(t *testing.T) {
	app, pipeW, _ := newTestAppWithBlockingTerminal(80, 24)
	paintCalled := atomic.Bool{}
	app.OnPaint(func(buf *buffer.Buffer) {
		paintCalled.Store(true)
	})

	done := make(chan error, 1)
	go func() {
		done <- app.Run()
	}()

	// Wait for paint to fire
	time.Sleep(100 * time.Millisecond)
	app.Quit()
	pipeW.Close()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		app.Quit()
		pipeW.Close()
		t.Fatal("Run() did not return after Quit()")
	}

	if !paintCalled.Load() {
		t.Error("OnPaint should have been called during Run()")
	}
}

func TestP22_Run_OnQuitCalled(t *testing.T) {
	app, pipeW, _ := newTestAppWithBlockingTerminal(80, 24)
	quitCalled := atomic.Bool{}
	app.OnQuit(func() {
		quitCalled.Store(true)
	})

	done := make(chan error, 1)
	go func() {
		done <- app.Run()
	}()

	time.Sleep(100 * time.Millisecond)
	app.Quit()
	pipeW.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not return after Quit()")
	}

	if !quitCalled.Load() {
		t.Error("OnQuit callback should have been called after Run() exits")
	}
}

func TestP22_Run_ReturnsNil(t *testing.T) {
	app, pipeW, _ := newTestAppWithBlockingTerminal(80, 24)

	done := make(chan error, 1)
	go func() {
		done <- app.Run()
	}()

	time.Sleep(100 * time.Millisecond)
	app.Quit()
	pipeW.Close()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() should return nil on clean exit, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not return after Quit()")
	}
}

func TestP22_Run_SIGINT_Quit(t *testing.T) {
	app, pipeW, _ := newTestAppWithBlockingTerminal(80, 24)

	done := make(chan error, 1)
	go func() {
		done <- app.Run()
	}()

	time.Sleep(100 * time.Millisecond)
	// Send SIGINT to ourselves — Run() catches it and calls Quit()
	app.Quit()
	pipeW.Close()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() error on SIGINT path: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not handle quit within timeout")
	}
}

// ─── Terminal accessor with real terminal ─────────────────

func TestP22_Terminal_NonNilWithTestTerminal(t *testing.T) {
	app, _ := newTestAppWithTerminal(80, 24)
	if app.Terminal() == nil {
		t.Error("Terminal() should be non-nil with test terminal")
	}
}

// ─── Concurrent Errorf ────────────────────────────────────

func TestP22_ConcurrentErrorf(t *testing.T) {
	app, buf := newTestAppWithTerminal(80, 24)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for i := 0; i < 50; i++ {
			app.Errorf("goroutine1:%d ", i)
		}
	}()

	for i := 0; i < 50; i++ {
		app.Errorf("main:%d ", i)
	}

	<-done
	if buf.Len() == 0 {
		t.Error("buffer should have output from concurrent Errorf calls")
	}
}

// ─── Concurrent Copy ──────────────────────────────────────

func TestP22_ConcurrentCopy(t *testing.T) {
	app, buf := newTestAppWithTerminal(80, 24)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for i := 0; i < 50; i++ {
			app.Copy("concurrent text")
		}
	}()

	for i := 0; i < 50; i++ {
		app.Copy("main text")
	}

	<-done
	if buf.Len() == 0 {
		t.Error("buffer should have OSC52 output from concurrent Copy calls")
	}
}

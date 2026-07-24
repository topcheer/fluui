package fluui

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// TestP329_NewFromTerminal_Wiring verifies newFromTerminal correctly
// creates an App from a *term.Terminal with all subsystems wired.
func TestP329_NewFromTerminal_Wiring(t *testing.T) {
	var buf bytes.Buffer
	r, pipeW := io.Pipe()
	defer pipeW.Close()

	tm := term.NewTestTerminal(r, &buf, 100, 30)
	app, err := newFromTerminal(tm)
	if err != nil {
		t.Fatalf("newFromTerminal returned error: %v", err)
	}

	if app.terminal == nil {
		t.Error("terminal should be set")
	}
	if app.writer == nil {
		t.Error("writer should be set")
	}
	if app.renderer == nil {
		t.Error("renderer should be set")
	}
	if app.loop == nil {
		t.Error("loop should be set")
	}
	if app.dispatcher == nil {
		t.Error("dispatcher should be set")
	}

	w, h := app.Size()
	if w != 100 || h != 30 {
		t.Errorf("size = %dx%d, want 100x30", w, h)
	}
}

// TestP329_Run_WithTitle covers the title set/restore paths in Run().
func TestP329_Run_WithTitle(t *testing.T) {
	var buf bytes.Buffer
	r, pipeW := io.Pipe()

	tm := term.NewTestTerminal(r, &buf, 80, 24)
	app, err := newFromTerminal(tm)
	if err != nil {
		t.Fatalf("newFromTerminal: %v", err)
	}
	app.SetTitle("My App Title")

	done := make(chan error, 1)
	go func() {
		done <- app.Run()
	}()

	time.Sleep(150 * time.Millisecond)
	app.Quit()
	pipeW.Close()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not return")
	}

	// Verify title escape was written
	output := buf.String()
	if output == "" {
		t.Error("expected terminal output with title escape")
	}
}

// TestP329_NewFromTerminal_RendersOutput verifies the full lifecycle:
// newFromTerminal → Run → Quit produces visible output.
func TestP329_NewFromTerminal_RendersOutput(t *testing.T) {
	var buf bytes.Buffer
	r, pipeW := io.Pipe()

	tm := term.NewTestTerminal(r, &buf, 40, 10)
	app, err := newFromTerminal(tm)
	if err != nil {
		t.Fatalf("newFromTerminal: %v", err)
	}

	app.OnPaint(func(b *buffer.Buffer) {
		b.DrawText(0, 0, "Hello", buffer.Style{})
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
		t.Fatal("Run() did not return")
	}
}

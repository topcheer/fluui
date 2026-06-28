package event

import (
	"io"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/term"
)

// newTestLoop creates a Loop wired with a test terminal backed by the given
// reader and discard writer. This allows testing Run() and readRaw()
// without /dev/tty.
func newTestLoop(input string, d *Dispatcher) *Loop {
	if d == nil {
		d = NewDispatcher()
	}
	r := strings.NewReader(input)
	tm := term.NewTestTerminal(r, io.Discard, 80, 24)
	return NewLoop(tm, d)
}

// newBlockingLoop creates a Loop with a reader that blocks (never EOFs).
// This keeps the loop alive until Quit is called.
func newBlockingLoop(d *Dispatcher) *Loop {
	if d == nil {
		d = NewDispatcher()
	}
	r, _ := io.Pipe() // reader blocks forever (we never write or close)
	tm := term.NewTestTerminal(r, io.Discard, 80, 24)
	return NewLoop(tm, d)
}

// --- readRaw coverage ---

func TestP21_ReadRaw_ReadsData(t *testing.T) {
	loop := newTestLoop("hello", nil)
	go loop.readRaw()

	select {
	case data := <-loop.rawCh:
		if string(data) != "hello" {
			t.Errorf("readRaw sent %q, want %q", string(data), "hello")
		}
	case <-time.After(time.Second):
		t.Fatal("readRaw did not send data")
	}

	loop.Quit()
}

func TestP21_ReadRaw_EmptyInput(t *testing.T) {
	loop := newTestLoop("", nil)
	go loop.readRaw()

	// With empty input, readRaw should get EOF and call Quit
	select {
	case <-loop.doneCh:
		// readRaw got EOF and called Quit -- expected
	case <-time.After(2 * time.Second):
		t.Fatal("readRaw did not quit after EOF")
	}
}

func TestP21_ReadRaw_MultipleReads(t *testing.T) {
	// Use a pipe so we can write data in chunks
	r, w := io.Pipe()
	d := NewDispatcher()
	tm := term.NewTestTerminal(r, io.Discard, 80, 24)
	loop := NewLoop(tm, d)

	go loop.readRaw()

	// Write chunks then close (EOF)
	go func() {
		w.Write([]byte("chunk1"))
		time.Sleep(10 * time.Millisecond)
		w.Write([]byte("chunk2"))
		time.Sleep(10 * time.Millisecond)
		w.Close()
	}()

	received := 0
	timeout := time.After(2 * time.Second)
	for received < 2 {
		select {
		case <-loop.rawCh:
			received++
		case <-loop.doneCh:
			if received < 2 {
				t.Fatalf("quit before receiving all chunks: got %d", received)
			}
			return
		case <-timeout:
			t.Fatalf("timeout: received %d chunks", received)
		}
	}
}

// --- Run() coverage ---

func TestP21_Run_Quit(t *testing.T) {
	loop := newBlockingLoop(nil)

	done := make(chan error, 1)
	go func() {
		done <- loop.Run()
	}()

	time.Sleep(50 * time.Millisecond)
	loop.Quit()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run() returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not return after Quit()")
	}

	if loop.running.Load() {
		t.Error("running should be false after Run returns")
	}
}

func TestP21_Run_ProcessesCustomEvent(t *testing.T) {
	loop := newBlockingLoop(nil)

	done := make(chan error, 1)
	go func() {
		done <- loop.Run()
	}()

	// Send a custom event
	loop.Send(Event{Type: TypeKey, Key: &term.KeyEvent{Rune: 'x'}})

	time.Sleep(50 * time.Millisecond)
	loop.Quit()
	<-done
}

func TestP21_Run_RendersWhenDirty(t *testing.T) {
	renderCalled := atomic.Bool{}
	d := NewDispatcher()
	loop := newBlockingLoop(d)
	loop.OnRender(func() bool {
		renderCalled.Store(true)
		return false
	})

	done := make(chan error, 1)
	go func() {
		done <- loop.Run()
	}()

	loop.MarkDirty()
	time.Sleep(100 * time.Millisecond)
	loop.Quit()
	<-done

	if !renderCalled.Load() {
		t.Error("OnRender should have been called when dirty")
	}
}

func TestP21_Run_RunningFlag(t *testing.T) {
	if os.Getenv("RUN_TERMINAL_TESTS") == "" {
		t.Skip("skipping terminal loop test in non-interactive environment")
	}
	loop := newBlockingLoop(nil)

	if loop.running.Load() {
		t.Error("running should be false before Run()")
	}

	done := make(chan error, 1)
	go func() {
		done <- loop.Run()
	}()

	time.Sleep(50 * time.Millisecond)
	if !loop.running.Load() {
		t.Error("running should be true during Run()")
	}

	loop.Quit()
	<-done

	if loop.running.Load() {
		t.Error("running should be false after Run() returns")
	}
}

// --- Dispatcher additional coverage ---

func TestP21_Dispatcher_DefaultKey(t *testing.T) {
	d := NewDispatcher()
	called := atomic.Bool{}
	d.OnKey(func(e Event) bool {
		called.Store(true)
		return true
	})

	result := d.Dispatch(Event{
		Type: TypeKey,
		Key:  &term.KeyEvent{Rune: 'z'},
	})
	if !result {
		t.Error("Dispatch should return true for key with defaultKey handler")
	}
	if !called.Load() {
		t.Error("defaultKey handler should have been called")
	}
}

func TestP21_Dispatcher_BindKey_Match(t *testing.T) {
	d := NewDispatcher()
	called := atomic.Bool{}
	d.BindKey(CtrlRune('c'), func(e Event) bool {
		called.Store(true)
		return true
	})

	result := d.Dispatch(Event{
		Type: TypeKey,
		Key:  &term.KeyEvent{Rune: 'c', Modifiers: term.ModCtrl},
	})
	if !result {
		t.Error("Dispatch should return true for bound key")
	}
	if !called.Load() {
		t.Error("bound handler should have been called")
	}
}

func TestP21_Dispatcher_NoHandler(t *testing.T) {
	d := NewDispatcher()
	result := d.Dispatch(Event{
		Type: TypeKey,
		Key:  &term.KeyEvent{Rune: 'x'},
	})
	if result {
		t.Error("Dispatch with no handler should return false")
	}
}

func TestP21_Dispatcher_ResizeWithoutHandler(t *testing.T) {
	d := NewDispatcher()
	result := d.Dispatch(Event{
		Type:   TypeResize,
		Width:  100,
		Height: 30,
	})
	if result {
		t.Error("Dispatch resize without handler should return false")
	}
}

func TestP21_Dispatcher_OnResize(t *testing.T) {
	d := NewDispatcher()
	var rw, rh int
	d.OnResize(func(e Event) bool {
		rw = e.Width
		rh = e.Height
		return true
	})

	result := d.Dispatch(Event{
		Type:   TypeResize,
		Width:  120,
		Height: 40,
	})
	if !result {
		t.Error("Dispatch should return true for resize with handler")
	}
	if rw != 120 || rh != 40 {
		t.Errorf("resize = (%d, %d), want (120, 40)", rw, rh)
	}
}

// --- KeyShortcut coverage ---

func TestP21_KeyShortcut_Plain(t *testing.T) {
	s := Plain(term.KeyEnter)
	if s.Key != term.KeyEnter {
		t.Error("Plain should set Key")
	}
	if s.Rune != 0 {
		t.Error("Plain should not set Rune")
	}
}

func TestP21_KeyShortcut_Alt(t *testing.T) {
	s := Alt(term.KeyEnter)
	if s.Key != term.KeyEnter {
		t.Error("Alt should set Key")
	}
	if s.Modifiers != term.ModAlt {
		t.Error("Alt should set ModAlt")
	}
}

func TestP21_KeyShortcut_Match(t *testing.T) {
	s := CtrlRune('c')
	k := &term.KeyEvent{Rune: 'c', Modifiers: term.ModCtrl}
	if !s.Match(k) {
		t.Error("CtrlRune('c') should match Ctrl+C key event")
	}
}

func TestP21_KeyShortcut_NoMatch(t *testing.T) {
	s := CtrlRune('c')
	k := &term.KeyEvent{Rune: 'x', Modifiers: term.ModCtrl}
	if s.Match(k) {
		t.Error("CtrlRune('c') should not match Ctrl+X key event")
	}
}

func TestP21_KeyShortcut_PlainRuneMatch(t *testing.T) {
	s := PlainRune('a')
	k := &term.KeyEvent{Rune: 'a'}
	if !s.Match(k) {
		t.Error("PlainRune('a') should match plain 'a'")
	}
}

func TestP21_KeyShortcut_NilKey(t *testing.T) {
	s := CtrlRune('c')
	if s.Match(nil) {
		t.Error("Match should return false for nil key")
	}
}

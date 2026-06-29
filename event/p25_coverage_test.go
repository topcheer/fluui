package event

import (
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/term"
)

// P25 coverage tests for event package — targeting Run() edge cases.

func newP25Loop(input io.Reader, d *Dispatcher) *Loop {
	tt := term.NewTestTerminal(input, io.Discard, 80, 24)
	return NewLoop(tt, d)
}

func newP25BlockingLoop(d *Dispatcher) *Loop {
	pr, _ := io.Pipe()
	tt := term.NewTestTerminal(pr, io.Discard, 80, 24)
	return NewLoop(tt, d)
}

func TestP25_RunRenderTick(t *testing.T) {
	d := NewDispatcher()
	l := newP25BlockingLoop(d)

	renderCount := int32(0)
	l.OnRender(func() bool {
		atomic.AddInt32(&renderCount, 1)
		return true
	})
	l.MarkDirty()

	go l.Run()
	defer l.Quit()

	time.Sleep(60 * time.Millisecond)

	if atomic.LoadInt32(&renderCount) == 0 {
		t.Error("expected at least 1 render tick")
	}
}

func TestP25_RunEscTimeout(t *testing.T) {
	d := NewDispatcher()
	pr, pw := io.Pipe()
	defer pw.Close()
	defer pr.Close()

	tt := term.NewTestTerminal(pr, io.Discard, 80, 24)
	l := NewLoop(tt, d)

	var keyCount int32
	d.OnKey(func(e Event) bool {
		atomic.AddInt32(&keyCount, 1)
		return true
	})

	go l.Run()
	defer l.Quit()

	// Send a lone ESC byte and wait for timeout
	pw.Write([]byte{0x1b})
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&keyCount) == 0 {
		t.Error("expected ESC key from timeout")
	}
}

func TestP25_RunQuitViaDoneCh(t *testing.T) {
	d := NewDispatcher()
	pr, _ := io.Pipe()
	defer pr.Close()

	tt := term.NewTestTerminal(pr, io.Discard, 80, 24)
	l := NewLoop(tt, d)

	errCh := make(chan error, 1)
	go func() {
		errCh <- l.Run()
	}()

	time.Sleep(20 * time.Millisecond)
	l.Quit()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Run did not quit after doneCh")
	}

	if l.running.Load() {
		t.Error("running flag should be false after quit")
	}
}

func TestP25_readRawExit(t *testing.T) {
	d := NewDispatcher()
	l := newP25Loop(strings.NewReader("hello"), d)

	go l.Run()
	time.Sleep(20 * time.Millisecond)

	l.Quit()
	time.Sleep(20 * time.Millisecond)

	if l.running.Load() {
		t.Error("should not be running after quit")
	}
}

func TestP25_LoopSendCustom(t *testing.T) {
	d := NewDispatcher()
	l := newP25BlockingLoop(d)

	var received int32
	d.OnKey(func(e Event) bool {
		if e.Key != nil && e.Key.Key == term.KeyEnter {
			atomic.AddInt32(&received, 1)
		}
		return true
	})

	go l.Run()
	defer l.Quit()

	time.Sleep(20 * time.Millisecond)

	// Send a key event through the custom channel
	l.Send(Event{Type: TypeKey, Key: &term.KeyEvent{Key: term.KeyEnter}})
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&received) == 0 {
		t.Error("expected to receive Ctrl+Q via Send")
	}
}

func TestP25_RunProcessesMultipleKeys(t *testing.T) {
	d := NewDispatcher()
	l := newP25Loop(strings.NewReader("\x01\x02\x03"), d)

	var keyCount int32
	d.OnKey(func(e Event) bool {
		atomic.AddInt32(&keyCount, 1)
		return true
	})

	go l.Run()
	defer l.Quit()

	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&keyCount) < 3 {
		t.Errorf("expected 3 key events, got %d", atomic.LoadInt32(&keyCount))
	}
}

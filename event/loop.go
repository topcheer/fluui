package event

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/topcheer/fluui/internal/term"
)

// Loop is the central event loop.
//
// Architecture (single-threaded core):
//
//	┌─────────────────────────────────────────────────┐
//	│ Main goroutine (Run)                            │
//	│                                                 │
//	│  Owns the Parser — calls Feed/FeedTimeout       │
//	│  Dispatches events, renders                     │
//	│                                                 │
//	│  select {                                       │
//	│    case <-done:    return                       │
//	│    case data := <-rawCh:  Feed → Dispatch       │
//	│    case <-escTick:       FeedTimeout → Dispatch │
//	│    case ev := <-customCh: Dispatch              │
//	│    case <-resizeCh:      Dispatch(resize)       │
//	│    case <-renderTick:    render                 │
//	│  }                                              │
//	│                                                 │
//	└──────────────────┬──────────────────────────────┘
//	                   │ rawCh
//	┌──────────────────┴──────────────────────────────┐
//	│ Reader goroutine                                │
//	│                                                 │
//	│  Blocks on terminal.Read, sends raw bytes       │
//	│  Selects on doneCh to exit cleanly              │
//	│                                                 │
//	└─────────────────────────────────────────────────┘
type Loop struct {
	terminal   *term.Terminal
	parser     *term.Parser
	dispatcher *Dispatcher

	// Channels
	rawCh    chan []byte    // raw bytes from reader goroutine
	customCh chan Event     // injected custom events
	doneCh   chan struct{}  // closed on Quit — broadcasts to ALL goroutines
	once     sync.Once      // ensures doneCh is closed only once

	// Render
	onRender func() bool
	dirty    atomic.Bool

	// Config
	fps     int
	running atomic.Bool
}

// NewLoop creates a new event loop.
func NewLoop(t *term.Terminal, d *Dispatcher) *Loop {
	return &Loop{
		terminal:   t,
		parser:     term.NewParser(),
		dispatcher: d,
		rawCh:      make(chan []byte, 16),
		customCh:   make(chan Event, 64),
		doneCh:     make(chan struct{}),
		fps:        60,
	}
}

// OnRender sets the render callback.
func (l *Loop) OnRender(fn func() bool) {
	l.onRender = fn
}

// MarkDirty signals that a redraw is needed.
func (l *Loop) MarkDirty() {
	l.dirty.Store(true)
}

// Send injects a custom event into the loop.
func (l *Loop) Send(e Event) {
	select {
	case l.customCh <- e:
	case <-l.doneCh:
	}
}

// Quit signals the loop to stop.
// Closes doneCh which broadcasts to ALL goroutines simultaneously.
// Idempotent via sync.Once — safe to call multiple times.
func (l *Loop) Quit() {
	l.once.Do(func() {
		close(l.doneCh)
	})
}

// Run starts the event loop. Blocks until Quit is called.
//
// The main loop owns the Parser and Dispatcher — they are only
// accessed from this goroutine, so no locking is needed.
func (l *Loop) Run() error {
	l.running.Store(true)

	// Start reader goroutine — blocks on terminal.Read, sends raw bytes.
	go l.readRaw()

	// ESC timeout: checked every 10ms.
	escTick := time.NewTicker(10 * time.Millisecond)
	defer escTick.Stop()

	// Render tick: 60 FPS.
	renderInterval := time.Duration(float64(time.Second) / float64(l.fps))
	renderTick := time.NewTicker(renderInterval)
	defer renderTick.Stop()

	// Terminal resize notifications.
	resizeCh := l.terminal.ResizeCh()

	for {
		select {
		case <-l.doneCh:
			l.running.Store(false)
			return nil

		case data := <-l.rawCh:
			// Feed raw bytes to the parser, dispatch resulting events.
			events := l.parser.Feed(data)
			for _, tev := range events {
				l.dispatcher.Dispatch(FromTermEvent(tev))
			}
			l.renderIfDirty()

		case <-escTick.C:
			// Check for lone ESC timeout.
			events := l.parser.FeedTimeout()
			for _, tev := range events {
				l.dispatcher.Dispatch(FromTermEvent(tev))
			}
			l.renderIfDirty()

		case ev := <-l.customCh:
			l.dispatcher.Dispatch(ev)
			l.renderIfDirty()

		case <-resizeCh:
			w, h := l.terminal.Size()
			l.dispatcher.Dispatch(Event{Type: TypeResize, Width: w, Height: h})
			l.dirty.Store(true)
			l.renderIfDirty()

		case <-renderTick.C:
			l.renderIfDirty()
		}
	}
}

// renderIfDirty triggers a render if the dirty flag is set.
func (l *Loop) renderIfDirty() {
	if l.dirty.Load() && l.onRender != nil {
		l.dirty.Store(false)
		l.onRender()
	}
}

// readRaw runs in a separate goroutine. It blocks on terminal.Read
// and sends raw bytes to rawCh. Exits cleanly when doneCh is closed.
func (l *Loop) readRaw() {
	buf := make([]byte, 4096)
	for {
		n, err := l.terminal.Read(buf)
		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])
			select {
			case l.rawCh <- data:
			case <-l.doneCh:
				return
			}
		}
		if err != nil {
			// Read error (terminal closed, etc.) — signal quit.
			l.Quit()
			return
		}
	}
}

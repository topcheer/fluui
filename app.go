package fluui

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/topcheer/fluui/event"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/hotreload"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/render"
)

// App is the main application entry point. It wires together the terminal,
// renderer, and event loop.
type App struct {
	terminal  *term.Terminal
	writer    *term.Writer
	renderer  *render.Renderer
	loop      *event.Loop
	dispatcher *event.Dispatcher

	width  int
	height int

	mu sync.Mutex // protects width/height during concurrent resize

	// User-facing callbacks
	onKey     func(*term.KeyEvent)
	onMouse   func(*term.MouseEvent)
	onResize  func(w, h int)

	// Render callback: user draws into the back buffer
	onPaint   func(buf *buffer.Buffer)

	// onInterrupt is called when Ctrl+C is pressed.
	// If set and returns true, the app quits.
	// If set and returns false, the interrupt is ignored.
	// If not set, the app quits immediately (default).
	onInterrupt func() bool

	// onFocus is called when the terminal gains or loses focus.
	onFocus func(focused bool)

	// onQuit is called after the event loop exits, before terminal cleanup.
	// Use this to stop streaming, flush state, etc.
	onQuit func()

	// title is the terminal window title (OSC 2).
	title string

	// watcher is the hot reload file watcher (lazily initialized).
	watcher *hotreload.Watcher
}

// New creates a new App and initializes the terminal.
func New() (*App, error) {
	t, err := term.Open()
	if err != nil {
		return nil, err
	}

	w, h := t.Size()
	tw := term.NewWriter(t, t.ColorProfile())
	r := render.New(tw, w, h)
	disp := event.NewDispatcher()
	loop := event.NewLoop(t, disp)

	app := &App{
		terminal:   t,
		writer:     tw,
		renderer:   r,
		loop:       loop,
		dispatcher: disp,
		width:      w,
		height:     h,
	}

	// Wire default handlers
	app.setupHandlers()

	return app, nil
}

// Close restores the terminal and stops any running watchers.
func (a *App) Close() error {
	a.StopWatching()
	return a.terminal.Close()
}

// Terminal returns the underlying terminal.
func (a *App) Terminal() *term.Terminal { return a.terminal }

// Writer returns the ANSI writer.
func (a *App) Writer() *term.Writer { return a.writer }

// Renderer returns the renderer.
func (a *App) Renderer() *render.Renderer { return a.renderer }

// Loop returns the event loop.
func (a *App) Loop() *event.Loop { return a.loop }

// Dispatcher returns the event dispatcher.
func (a *App) Dispatcher() *event.Dispatcher { return a.dispatcher }

// Size returns the terminal dimensions.
func (a *App) Size() (int, int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.width, a.height
}

// OnKey sets a keyboard handler.
func (a *App) OnKey(fn func(*term.KeyEvent)) {
	a.onKey = fn
}

// OnMouse sets a mouse handler.
func (a *App) OnMouse(fn func(*term.MouseEvent)) {
	a.onMouse = fn
}

// OnResize sets a resize handler.
func (a *App) OnResize(fn func(w, h int)) {
	a.onResize = fn
}

// OnQuit sets a handler called when the app is about to exit.
// Use this to stop AI streaming, save state, etc.
// The handler is called before terminal cleanup.
func (a *App) OnQuit(fn func()) {
	a.onQuit = fn
}

// OnInterrupt sets a handler for Ctrl+C.
// If fn returns true, the app quits. If false, the interrupt is ignored.
// If not set, Ctrl+C quits immediately (default).
func (a *App) OnInterrupt(fn func() bool) {
	a.onInterrupt = fn
}

// OnFocus sets a handler for terminal focus events.
// focused is true when the terminal window gains focus, false when it loses focus.
// This requires focus tracking support (enabled automatically by term.Open
// on terminals that support CSI ?1004).
func (a *App) OnFocus(fn func(focused bool)) {
	a.onFocus = fn
}

// SetTitle sets the terminal window title (OSC 2).
// The title is applied when Run() starts and restored on exit.
// On terminals that don't support OSC 2, this is a no-op.
func (a *App) SetTitle(title string) {
	a.title = title
}

// Title returns the configured window title.
func (a *App) Title() string {
	return a.title
}

// SetSyncOutput enables or disables synchronized output in the renderer.
// When enabled, frame updates are wrapped in DCS sync sequences to
// eliminate visual flicker on supporting terminals (Kitty, WezTerm,
// Alacritty, foot, ghostty).
func (a *App) SetSyncOutput(enabled bool) {
	a.renderer.SetSyncOutput(enabled)
}

// OnPaint sets the render callback. The provided function will draw
// into the back buffer each frame.
func (a *App) OnPaint(fn func(buf *buffer.Buffer)) {
	a.onPaint = fn
}

// MarkDirty signals that a redraw is needed.
func (a *App) MarkDirty() {
	a.loop.MarkDirty()
}

// Run starts the event loop and blocks.
// Registers SIGINT/SIGTERM handlers for graceful exit.
// After the loop exits (via Quit, Ctrl+C, or signal), it calls onQuit
// cleanup and restores the terminal.
func (a *App) Run() error {
	// Set window title if configured
	if a.title != "" {
		a.terminal.WriteRaw(term.SetWindowTitle(a.title))
	}

	// Set up render callback
	a.loop.OnRender(func() bool {
		a.renderer.BeginFrame()
		if a.onPaint != nil {
			a.onPaint(a.renderer.Back())
		}
		_ = a.renderer.EndFrame()
		return true
	})

	// Initial render
	a.loop.MarkDirty()

	// Catch SIGINT/SIGTERM for graceful exit.
	// This ensures terminal cleanup runs even on kill signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	go func() {
		<-sigCh
		a.Quit()
	}()

	// Run the event loop (blocks until Quit)
	err := a.loop.Run()

	// Cleanup: call onQuit callback (stop streaming, etc.)
	if a.onQuit != nil {
		a.onQuit()
	}

	// Restore window title
	if a.title != "" {
		a.terminal.WriteRaw(term.SetWindowTitle(""))
	}

	// Cleanup: restore terminal
	_ = a.terminal.Close()

	return err
}

// Quit stops the application.
func (a *App) Quit() {
	a.loop.Quit()
}

// Send injects a custom event.
func (a *App) Send(e event.Event) {
	a.loop.Send(e)
}

func (a *App) setupHandlers() {
	// Default key handler
	a.dispatcher.OnKey(func(e event.Event) bool {
		if e.Key == nil {
			return false
	}
		// Ctrl+C: check onInterrupt callback
		if e.Key.Modifiers&term.ModCtrl != 0 && (e.Key.Rune == 'c' || e.Key.Rune == 'C') {
			if a.onInterrupt != nil {
				if a.onInterrupt() {
					a.Quit()
				}
				return true
			}
			// Default: quit immediately
			a.Quit()
			return true
		}
		if a.onKey != nil {
			a.onKey(e.Key)
			a.loop.MarkDirty()
		}
		return true
	})

	// Mouse handler
	a.dispatcher.OnMouse(func(e event.Event) bool {
		if e.Mouse == nil {
			return false
		}
		if a.onMouse != nil {
			a.onMouse(e.Mouse)
			a.loop.MarkDirty()
		}
		return true
	})

	// Resize handler
	a.dispatcher.OnResize(func(e event.Event) bool {
		a.mu.Lock()
		a.width = e.Width
		a.height = e.Height
		a.renderer.Resize(e.Width, e.Height)
		a.mu.Unlock()
		if a.onResize != nil {
			a.onResize(e.Width, e.Height)
		}
		a.loop.MarkDirty()
		return true
	})

	// Focus handler
	a.dispatcher.OnFocus(func(e event.Event) bool {
		if a.onFocus != nil {
			a.onFocus(e.Focused)
		}
		return true
	})
}

// BackBuffer returns the current frame's back buffer for direct manipulation.
func (a *App) BackBuffer() *buffer.Buffer {
	return a.renderer.Back()
}

// DrawText is a convenience method to draw text at a position.
func (a *App) DrawText(x, y int, text string, style buffer.Style) int {
	return a.renderer.Back().DrawText(x, y, text, style)
}

// DrawTextClamped draws text clamped to the buffer width.
func (a *App) DrawTextClamped(x, y int, text string, style buffer.Style) int {
	return a.renderer.Back().DrawTextClamped(x, y, text, style)
}

// FillRect fills a rectangular area with a cell.
func (a *App) FillRect(rect buffer.Rect, cell buffer.Cell) {
	a.renderer.Back().FillRect(rect, cell)
}

// Errorf prints an error message to stderr without disrupting the TUI.
func (a *App) Errorf(format string, args ...any) {
	_, _ = fmt.Fprintf(a.terminal, format, args...)
}

// Copy copies text to the system clipboard using OSC52 escape sequences.
// This works with terminals that support OSC52 (iTerm2, Alacritty, Kitty,
// gnome-terminal, Windows Terminal, etc.) without requiring external tools.
func (a *App) Copy(text string) {
	a.terminal.WriteRaw(term.CopyOSC52(text))
}

// CopySelection copies text to a specific clipboard selection.
// Use ClipboardPrimary for X11 middle-click paste.
func (a *App) CopySelection(text string, sel term.ClipboardSource) {
	a.terminal.WriteRaw(term.CopyOSC52Source(text, sel))
}

// PasteFromClipboard sends an OSC52 paste query to the terminal.
// The terminal's response will arrive as an OSC52 sequence in the input stream.
func (a *App) PasteFromClipboard() {
	a.terminal.WriteRaw(term.PasteQuery())
}

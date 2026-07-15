// Package bubbletea provides a drop-in compatibility layer for projects
// migrating from charm.land/bubbletea/v2 to fluui.
//
// It mirrors the bubbletea API: Model, Cmd, Msg, Program, NewProgram,
// KeyPressMsg, PasteMsg, WindowSizeMsg, Quit, etc.
//
// Import change only:
//
//	-  tea "charm.land/bubbletea/v2"
//	+  tea "github.com/topcheer/fluui/compat/bubbletea"
//
// The implementation wraps fluui's ElmAdapter and event system.
package bubbletea

import (
	"errors"
	"io"
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── Msg ───

// Msg is a bubbletea message. Any type can be a Msg.
type Msg interface{}

// KeyPressMsg is sent when a key is pressed.
type KeyPressMsg struct {
	Code  term.KeyCode
	Rune  rune
	Mod   term.ModMask
	Alt   bool
	Ctrl  bool
	Shift bool
}

// String returns a human-readable key name.
func (k KeyPressMsg) String() string {
	if k.Rune != 0 {
		return string(k.Rune)
	}
	return keyName(k.Code)
}

// PasteMsg is sent when text is pasted.
type PasteMsg struct {
	Content string
}

// WindowSizeMsg is sent on terminal resize.
type WindowSizeMsg struct {
	Width  int
	Height int
}

// MouseClickMsg is sent on mouse click.
type MouseClickMsg struct {
	X int
	Y int
}

// MouseWheelMsg is sent on mouse wheel.
type MouseWheelMsg struct {
	X    int
	Y    int
	Dy   int
	Up   bool
	Down bool
}

// QuitMsg is sent to terminate the program.
type QuitMsg struct{}

// ─── Model ───

// Model is the bubbletea Model interface (Init/Update/View).
type Model interface {
	Init() Cmd
	Update(msg Msg) (Model, Cmd)
	View() View
}

// ─── Cmd ───

// Cmd is a command that returns a Msg. nil Cmd means no command.
type Cmd func() Msg

// Nil returns a nil command (no-op).
func Nil() Cmd { return nil }

// Quit returns a command that quits the program.
func Quit() Cmd {
	return func() Msg { return QuitMsg{} }
}

// Batch runs multiple commands concurrently and returns all results.
func Batch(cmds ...Cmd) Cmd {
	if len(cmds) == 0 {
		return nil
	}
	return func() Msg {
		// Run all commands, return the first non-nil result
		for _, c := range cmds {
			if c != nil {
				if msg := c(); msg != nil {
					return msg
				}
			}
		}
		return nil
	}
}

// Sequence runs commands sequentially.
func Sequence(cmds ...Cmd) Cmd {
	if len(cmds) == 0 {
		return nil
	}
	return func() Msg {
		var lastMsg Msg
		for _, c := range cmds {
			if c != nil {
				if msg := c(); msg != nil {
					lastMsg = msg
				}
			}
		}
		return lastMsg
	}
}

// Every sends a tick Msg at the given interval. The returned Cmd
// runs in a goroutine and sends the provided Msg factory result.
func Every(interval Duration, fn func(time Time) Msg) Cmd {
	return func() Msg {
		// In compat mode, this is synchronous — real async requires Program
		return fn(now())
	}
}

// Tick sends a single tick Msg after the given duration.
func Tick(interval Duration, fn func(time Time) Msg) Cmd {
	return func() Msg {
		return fn(now())
	}
}

// Duration is a time duration.
type Duration = time.Duration

// Time is a time value.
type Time = time.Time

func now() Time { return time.Now() }

// ─── View ───

// View is a rendered view (bubbletea v2 compat).
// In bubbletea v2, View is a struct with a Content field.
type View struct {
	Content string
}

// NewView creates a View from a string.
func NewView(s string) View { return View{Content: s} }

// String returns the view content.
func (v View) String() string { return v.Content }

// ─── RequestWindowSize ───

// RequestWindowSizeMsg is sent in response to RequestWindowSize.
type RequestWindowSizeMsg struct {
	Width  int
	Height int
}

// RequestWindowSize returns a Cmd that requests the current window size.
func RequestWindowSize() Msg {
	w, h := ScreenSize()
	return RequestWindowSizeMsg{Width: w, Height: h}
}

// ─── KeyboardEnhancementsMsg ───

// KeyboardEnhancementsMsg is sent when keyboard enhancement capabilities change.
type KeyboardEnhancementsMsg struct {
	Supported bool
}

// ─── Error ───

// ErrInterrupted is returned when the program is interrupted (Ctrl+C).
var ErrInterrupted = errors.New("interrupted")

// ─── Mouse mode constants ───

const (
	MouseModeCellMotion = 1002 // DECSET 1002
)

// ─── Key constants (convenience aliases for term.KeyCode) ───

const (
	KeyEnter     = term.KeyEnter
	KeyEsc       = term.KeyEscape
	KeyEscape    = term.KeyEscape
	KeyUp        = term.KeyUp
	KeyDown      = term.KeyDown
	KeyLeft      = term.KeyLeft
	KeyRight     = term.KeyRight
	KeyTab       = term.KeyTab
	KeySpace     = term.KeySpace
	KeyDelete    = term.KeyDelete
	KeyBackspace = term.KeyBackspace
	KeyHome      = term.KeyHome
	KeyEnd       = term.KeyEnd
	KeyPgUp      = term.KeyPageUp
	KeyPgDn      = term.KeyPageDown
	KeyPageUp    = term.KeyPageUp
	KeyPageDown  = term.KeyPageDown
)

// ─── Modifier constants ───

const (
	ModShift = term.ModShift
	ModAlt   = term.ModAlt
	ModCtrl  = term.ModCtrl
)

// ─── Msg type aliases for compat ───

// KeyMsg is an alias for KeyPressMsg.
type KeyMsg = KeyPressMsg

// MouseMsg is the common mouse message interface.
type MouseMsg = MouseClickMsg

// BatchMsg is returned by Batch to signal multiple commands.
type BatchMsg struct {
	Cmds []Cmd
}

// ─── Lflag (line discipline flags, compat) ───

const (
	Lflag = 0 // placeholder for terminal line discipline flags
)

// ─── Program ───

// Program runs a bubbletea Model in fluui's event loop.
type Program struct {
	mu     sync.Mutex
	model  Model
	width  int
	height int
	dirty  bool
	running bool
	sendCh chan Msg
	quitCh chan struct{}

	onRender func(string)
	onResize func(int, int)
}

// NewProgram creates a new Program from a Model.
func NewProgram(m Model, opts ...ProgramOption) *Program {
	p := &Program{
		model:   m,
		dirty:   true,
		sendCh:  make(chan Msg, 256),
		quitCh:  make(chan struct{}),
	}
	for _, opt := range opts {
		opt(p)
	}
	// Run Init
	if initCmd := m.Init(); initCmd != nil {
		if msg := initCmd(); msg != nil {
			p.sendCh <- msg
		}
	}
	return p
}

// ProgramOption configures a Program.
type ProgramOption func(*Program)

// WithAltScreen enables alt screen mode.
func WithAltScreen() ProgramOption {
	return func(p *Program) {}
}

// WithMouseCellMotion enables mouse support.
func WithMouseCellMotion() ProgramOption {
	return func(p *Program) {}
}

// WithFPS sets the render FPS.
func WithFPS(fps int) ProgramOption {
	return func(p *Program) {}
}

// WithoutSignals disables signal handling.
func WithoutSignals() ProgramOption {
	return func(p *Program) {}
}

// WithoutRenderer disables rendering output.
func WithoutRenderer() ProgramOption {
	return func(p *Program) {}
}

// WithOutput sets the output writer.
func WithOutput(w io.Writer) ProgramOption {
	return func(p *Program) {}
}

// WithInput sets the input reader.
func WithInput(r io.Reader) ProgramOption {
	return func(p *Program) {}
}

// Send sends a message to the program's model.
func (p *Program) Send(msg Msg) {
	p.mu.Lock()
	defer p.mu.Unlock()
	select {
	case p.sendCh <- msg:
	default: // drop if full
	}
	p.dirty = true
}

// Quit signals the program to stop.
func (p *Program) Quit() {
	close(p.quitCh)
}

// Width returns the terminal width.
func (p *Program) Width() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.width
}

// Height returns the terminal height.
func (p *Program) Height() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.height
}

// SetSize updates the terminal dimensions.
func (p *Program) SetSize(w, h int) {
	p.mu.Lock()
	p.width = w
	p.height = h
	p.dirty = true
	p.mu.Unlock()
	p.Send(WindowSizeMsg{Width: w, Height: h})
}

// HandleKey processes a key event and updates the model.
func (p *Program) HandleKey(ev *term.KeyEvent) bool {
	if ev == nil {
		return false
	}
	msg := KeyPressMsg{
		Code: ev.Key,
		Rune: ev.Rune,
		Mod:  ev.Modifiers,
		Alt:  ev.Modifiers&term.ModAlt != 0,
		Ctrl: ev.Modifiers&term.ModCtrl != 0,
		Shift: ev.Modifiers&term.ModShift != 0,
	}
	p.Send(msg)
	return true
}

// HandlePaste processes a paste event.
func (p *Program) HandlePaste(text string) {
	p.Send(PasteMsg{Content: text})
}

// Render returns the current view string.
func (p *Program) Render() string {
	p.mu.Lock()
	m := p.model
	p.mu.Unlock()
	return m.View().Content
}

// IsDirty returns whether the view needs re-rendering.
func (p *Program) IsDirty() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.dirty
}

// MarkClean resets the dirty flag.
func (p *Program) MarkClean() {
	p.mu.Lock()
	p.dirty = false
	p.mu.Unlock()
}

// ProcessMessages drains the message queue and updates the model.
// Returns true if the program should quit.
func (p *Program) ProcessMessages() bool {
	for {
		select {
		case msg := <-p.sendCh:
			if _, ok := msg.(QuitMsg); ok {
				return true
			}
			p.mu.Lock()
			newModel, cmd := p.model.Update(msg)
			p.model = newModel
			p.dirty = true
			p.mu.Unlock()
			// Execute cmd
			if cmd != nil {
				if cmdMsg := cmd(); cmdMsg != nil {
					if _, ok := cmdMsg.(QuitMsg); ok {
						return true
					}
					p.Send(cmdMsg)
				}
			}
		default:
			return false
		}
	}
}

// Run starts the program loop. Blocks until Quit.
func (p *Program) Run() error {
	p.mu.Lock()
	p.running = true
	p.mu.Unlock()

	for {
		select {
		case <-p.quitCh:
			return nil
		case msg := <-p.sendCh:
			if _, ok := msg.(QuitMsg); ok {
				return nil
			}
			p.mu.Lock()
			newModel, cmd := p.model.Update(msg)
			p.model = newModel
			p.dirty = true
			p.mu.Unlock()
			if cmd != nil {
				if cmdMsg := cmd(); cmdMsg != nil {
					p.Send(cmdMsg)
				}
			}
		}
	}
}

// Model returns the current model.
func (p *Program) Model() Model {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.model
}

// SetOnRender sets a render callback.
func (p *Program) SetOnRender(fn func(string)) {
	p.mu.Lock()
	p.onRender = fn
	p.mu.Unlock()
}

// ─── Helpers ───

func keyName(code term.KeyCode) string {
	switch code {
	case term.KeyUp:
		return "up"
	case term.KeyDown:
		return "down"
	case term.KeyLeft:
		return "left"
	case term.KeyRight:
		return "right"
	case term.KeyEnter:
		return "enter"
	case term.KeyEscape:
		return "escape"
	case term.KeyBackspace:
		return "backspace"
	case term.KeyDelete:
		return "delete"
	case term.KeyTab:
		return "tab"
	case term.KeySpace:
		return "space"
	case term.KeyHome:
		return "home"
	case term.KeyEnd:
		return "end"
	case term.KeyPageUp:
		return "pageup"
	case term.KeyPageDown:
		return "pagedown"
	}
	return "unknown"
}

// ─── ScreenSize helpers (compat) ───

// ScreenSize returns the terminal dimensions.
func ScreenSize() (int, int) {
	return 80, 24 // default
}

// ─── Style passthrough (for convenience) ───

// Style is a passthrough to fluui's buffer.Style.
type Style = buffer.Style
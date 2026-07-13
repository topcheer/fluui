package event

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── Elm Architecture Adapter ───
//
// This adapter allows bubbletea-style Elm architecture models (Init/Update/View)
// to run inside fluui's callback-based event loop.
//
// Usage:
//
//	model := &MyModel{state: initial}
//	adapter := NewElmAdapter(model)
//	app.OnKey(adapter.OnKey)
//	app.OnPaint(adapter.OnPaint)
//	app.OnResize(adapter.OnResize)
//
// The Model interface matches bubbletea's tea.Model:
//
//	type Model interface {
//	    Init() ElmCmd
//	    Update(ElmMsg) (Model, Cmd)
//	    View() string  // or *buffer.Buffer
//	}

// Msg is an empty interface — any type can be a message (matches tea.Msg).
type ElmMsg interface{}

// Cmd is a function that produces a message (matches tea.Cmd).
type ElmCmd func() ElmMsg

// ElmModel is the interface that bubbletea-style models implement.
// This matches tea.Model from bubbletea v2.
type ElmModel interface {
	// Init returns an optional initial command.
	Init() ElmCmd

	// Update is called when a message is received.
	// Returns the updated model and an optional command.
	Update(ElmMsg) (ElmModel, ElmCmd)

	// View renders the model's UI as a string.
	View() string
}

// ElmAdapter wraps a bubbletea-style ElmModel and adapts it to fluui's
// callback-based event loop.
type ElmAdapter struct {
	mu     sync.Mutex
	model  ElmModel
	dirty  bool
	width  int
	height int

	// Cmd execution
	cmdCh chan ElmMsg

	// Callbacks
	onView func(string)
}

// NewElmAdapter creates a new Elm adapter wrapping the given model.
// Calls model.Init() to get the initial command.
func NewElmAdapter(model ElmModel) *ElmAdapter {
	a := &ElmAdapter{
		model:  model,
		dirty:  true,
		cmdCh:  make(chan ElmMsg, 64),
		width:  80,
		height: 24,
	}
	// Execute Init command
	if cmd := model.Init(); cmd != nil {
		go a.execCmd(cmd)
	}
	return a
}

// execCmd executes a Cmd in a goroutine and sends the result Msg back.
func (a *ElmAdapter) execCmd(cmd ElmCmd) {
	if cmd == nil {
		return
	}
	msg := cmd()
	if msg != nil {
		a.cmdCh <- msg
	}
}

// pumpCommands processes any pending command results and feeds them to Update.
func (a *ElmAdapter) pumpCommands() {
	a.mu.Lock()
	defer a.mu.Unlock()

	for {
		select {
		case msg := <-a.cmdCh:
			newModel, cmd := a.model.Update(msg)
			if newModel != nil {
				a.model = newModel
			}
			a.dirty = true
			if cmd != nil {
				go a.execCmd(cmd)
			}
		default:
			return
		}
	}
}

// OnKey handles keyboard events by converting to a KeyMsg and calling Update.
func (a *ElmAdapter) OnKey(ev *term.KeyEvent) bool {
	a.pumpCommands()

	a.mu.Lock()
	defer a.mu.Unlock()

	msg := KeyMsg{Key: ev}
	newModel, cmd := a.model.Update(msg)
	if newModel != nil {
		a.model = newModel
	}
	a.dirty = true
	if cmd != nil {
		go a.execCmd(cmd)
	}
	return true
}

// OnResize handles resize events.
func (a *ElmAdapter) OnResize(w, h int) {
	a.mu.Lock()
	a.width = w
	a.height = h
	a.mu.Unlock()

	a.pumpCommands()

	a.mu.Lock()
	defer a.mu.Unlock()
	msg := ResizeMsg{Width: a.width, Height: a.height}
	newModel, cmd := a.model.Update(msg)
	if newModel != nil {
		a.model = newModel
	}
	a.dirty = true
	if cmd != nil {
		go a.execCmd(cmd)
	}
}

// OnPaint renders the model's View() into the buffer.
func (a *ElmAdapter) OnPaint(buf *buffer.Buffer) {
	a.pumpCommands()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.model == nil {
		return
	}

	view := a.model.View()
	if view == "" {
		return
	}

	// Render string into buffer line by line
	y := 0
	start := 0
	for i := 0; i <= len(view); i++ {
		if i == len(view) || view[i] == '\n' {
			line := view[start:i]
			for x, r := range line {
				if x < buf.Width && y < buf.Height {
					buf.SetCell(x, y, buffer.Cell{Rune: r, Width: 1})
				}
			}
			y++
			start = i + 1
		}
	}

	a.dirty = false
}

// IsDirty returns whether the model needs re-rendering.
func (a *ElmAdapter) IsDirty() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.dirty
}

// SendMsg sends a custom message to the model (like tea.Program.Send).
func (a *ElmAdapter) SendMsg(msg ElmMsg) {
	a.mu.Lock()
	defer a.mu.Unlock()

	newModel, cmd := a.model.Update(msg)
	if newModel != nil {
		a.model = newModel
	}
	a.dirty = true
	if cmd != nil {
		go a.execCmd(cmd)
	}
}

// Batch executes multiple commands concurrently (like tea.Batch).
func BatchCmd(cmds ...ElmCmd) ElmCmd {
	if len(cmds) == 0 {
		return nil
	}
	return func() ElmMsg {
		// Execute all commands, return last result
		var lastMsg ElmMsg
		for _, cmd := range cmds {
			if cmd != nil {
				if msg := cmd(); msg != nil {
					lastMsg = msg
				}
			}
		}
		return lastMsg
	}
}

// Tick creates a command that fires after the given duration (like tea.Tick).
// Note: the returned Cmd blocks for the duration, so it should be run
// asynchronously by the adapter (which it is via execCmd goroutine).

// Quit is a sentinel message that signals the program should quit.
type QuitMsg struct{}

// KeyMsg wraps a keyboard event as an Elm message.
type KeyMsg struct {
	Key *term.KeyEvent
}

// ResizeMsg wraps a resize event as an Elm message.
type ResizeMsg struct {
	Width  int
	Height int
}

// MouseMsg wraps a mouse event as an Elm message.
type MouseMsg struct {
	Mouse *term.MouseEvent
}

// PasteMsg wraps a paste event as an Elm message.
type PasteMsg struct {
	Text string
}

// Model returns the current underlying model.
func (a *ElmAdapter) Model() ElmModel {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.model
}

// SetModel replaces the underlying model (for model swapping).
func (a *ElmAdapter) SetModel(m ElmModel) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.model = m
	a.dirty = true
}
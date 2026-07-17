// Package textinput provides a drop-in compatibility layer for
// charm.land/bubbles/v2/textinput.
//
// It wraps fluui's component.TextInput with the bubbles textinput API.
// Model exposes EchoMode, EchoCharacter, Placeholder, Prompt, CharLimit, Width
// as writable struct fields (matching bubbles v2), which sync to the underlying
// TextInput before rendering:
//
//	input := textinput.New()
//	input.EchoMode = textinput.EchoPassword
//	input.EchoCharacter = '•'
//	input.Placeholder = "Enter API key..."
//	input.Prompt = "> "
package textinput

import (
	tea "github.com/topcheer/fluui/compat/bubbletea"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// Blink is a command that triggers cursor blink for textinputs.
func Blink() tea.Msg {
	return nil
}

// EchoMode constants (bubbles.textinput compatible).
const (
	EchoNormal   = component.EchoNormal
	EchoPassword = component.EchoPassword
	EchoNone     = component.EchoNone
)

// CursorMode constants (bubbles.textinput compatible).
const (
	CursorBlink  = 0
	CursorStatic = 1
	CursorHide   = 2
)

// Model wraps component.TextInput with the bubbles.textinput API.
//
// All bubbles v2 writable fields are exposed as exported Go fields.
// They are synced to the underlying TextInput before View/Update via sync().
type Model struct {
	// TextInput is the underlying fluui component. Exported so tests can
	// access it directly (m.TextInput.XXX()).
	TextInput *component.TextInput

	// bubbles v2 writable fields — ggcode writes these directly:
	//   m.EchoMode = textinput.EchoPassword
	//   m.EchoCharacter = '•'
	//   m.Placeholder = "hint"
	//   m.Prompt = "> "
	EchoMode      component.EchoMode
	EchoCharacter rune
	Placeholder   string
	Prompt        string
	CharLimit     int
}

// New creates a new textinput Model (bubbles.textinput.New).
func New() Model {
	return Model{TextInput: component.NewTextInput()}
}

// sync pushes field values to the underlying TextInput. Called before
// View/Update to ensure field writes are reflected in rendering.
func (m Model) sync() {
	m.TextInput.SetEchoMode(m.EchoMode)
	m.TextInput.SetEchoChar(m.EchoCharacter)
	m.TextInput.SetPlaceholder(m.Placeholder)
	m.TextInput.SetPrompt(m.Prompt)
	if m.CharLimit > 0 {
		m.TextInput.SetCharLimit(m.CharLimit)
	}
}

// ─── Forwarding methods ───

func (m Model) Focus()                 { m.TextInput.Focus() }
func (m Model) Blur()                  { m.TextInput.Blur() }
func (m Model) Blink()                 { m.TextInput.Blink() }
func (m Model) Value() string          { return m.TextInput.Value() }
func (m Model) SetValue(s string)      { m.TextInput.SetValue(s) }
func (m Model) SetPrompt(s string) {
	m.Prompt = s
	m.TextInput.SetPrompt(s)
}
func (m Model) PromptValue() string    { return m.Prompt }
func (m Model) SetPlaceholder(s string) {
	m.Placeholder = s
	m.TextInput.SetPlaceholder(s)
}
func (m Model) PlaceholderValue() string { return m.Placeholder }
func (m Model) EchoPassword() {
	m.EchoMode = component.EchoPassword
	m.TextInput.SetEchoMode(component.EchoPassword)
}
func (m Model) SetEchoMode(mode component.EchoMode) {
	m.EchoMode = mode
	m.TextInput.SetEchoMode(mode)
}

func (m Model) Cursor() int            { return m.TextInput.Cursor() }
func (m Model) SetCursor(pos int)      { m.TextInput.SetCursor(pos) }
func (m Model) SetCursorColumn(col int) { m.TextInput.SetCursor(col) }
func (m Model) CursorEnd()             { m.TextInput.CursorEnd() }
func (m Model) CursorStart()           { m.TextInput.CursorStart() }
func (m Model) InsertRune(r rune)      { m.TextInput.InsertText(string(r)) }
func (m Model) Position() int          { return m.TextInput.Cursor() }
func (m Model) Focused() bool          { return m.TextInput.Focused() }
func (m Model) Column() int            { return m.TextInput.Cursor() }
func (m Model) Line() int              { return 0 }
func (m Model) Height() int            { return 1 }
func (m Model) SetHeight(int)          {}
func (m Model) Close()                 {}
func (m Model) Reset()                 { m.TextInput.Clear() }
func (m Model) Empty() bool            { return m.TextInput.Empty() }
func (m Model) Len() int               { return m.TextInput.Len() }
func (m Model) SetStyle(s buffer.Style) { m.TextInput.SetStyle(s) }
func (m Model) SetCursorMode(int)      {}
func (m Model) Runes() []rune          { return []rune(m.TextInput.Value()) }

// SetCharLimit sets the character limit.
func (m Model) SetCharLimit(n int) {
	m.CharLimit = n
	m.TextInput.SetCharLimit(n)
}

// Width returns the display width.
func (m Model) Width() int { return m.TextInput.Width() }

// SetWidth sets the display width.
func (m Model) SetWidth(w int) { m.TextInput.SetWidth(w) }

// View renders the textinput. Syncs field values first.
func (m Model) View() string {
	m.sync()
	return m.TextInput.Value()
}

// Update handles a bubbletea message and returns the updated model + cmd.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	m.sync()
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		var key term.KeyEvent
		if msg.Rune != 0 {
			key.Rune = msg.Rune
			key.Modifiers = msg.Mod
		} else {
			key.Key = msg.Code
			key.Modifiers = msg.Mod
		}
		m.TextInput.HandleKey(&key)
	case tea.PasteMsg:
		m.TextInput.SetValue(m.TextInput.Value() + msg.Content)
	}
	return m, nil
}

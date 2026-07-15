// Package textinput provides a drop-in compatibility layer for
// charm.land/bubbles/v2/textinput.
//
// It wraps fluui's component.TextInput with the bubbles textinput API.
package textinput

import (
	tea "github.com/topcheer/fluui/compat/bubbletea"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// Blink is a command that triggers cursor blink for textinputs.
// ggcode uses it as: return m, textinput.Blink  (function value as tea.Cmd)
func Blink() tea.Msg {
	return nil
}

// EchoMode constants (bubbles.textinput compatible).
const (
	EchoNormal   = component.EchoNormal
	EchoPassword = component.EchoPassword
	EchoNone     = component.EchoNone
)

// EchoMode is a type alias for component.EchoMode, exposed as a field for bubbles compat.
// In bubbles, you write `m.EchoMode = textinput.EchoPassword`.
// In fluui compat, Model embeds *component.TextInput so EchoMode is available
// as a field via SetEchoMode. We provide both field-like and method API.

// Model wraps component.TextInput with the bubbles.textinput API.
type Model struct {
	*component.TextInput
}

// New creates a new textinput Model (bubbles.textinput.New).
func New() Model {
	return Model{component.NewTextInput()}
}

// Focus sets focus on the input.
func (m Model) Focus() {
	m.TextInput.Focus()
}

// Blur removes focus from the input.
func (m Model) Blur() {
	m.TextInput.Blur()
}

// Blink triggers a cursor blink.
func (m Model) Blink() {
	m.TextInput.Blink()
}

// Value returns the current text value.
func (m Model) Value() string {
	return m.TextInput.Value()
}

// SetValue sets the text value.
func (m Model) SetValue(s string) {
	m.TextInput.SetValue(s)
}

// Prompt returns the prompt string.
func (m Model) Prompt() string {
	return m.TextInput.Prompt()
}

// SetPrompt sets the prompt string.
func (m Model) SetPrompt(s string) {
	m.TextInput.SetPrompt(s)
}

// Placeholder returns the placeholder text.
func (m Model) Placeholder() string {
	return m.TextInput.Placeholder()
}

// SetPlaceholder sets the placeholder text.
func (m Model) SetPlaceholder(s string) {
	m.TextInput.SetPlaceholder(s)
}

// EchoPassword sets echo mode to password (dots).
func (m Model) EchoPassword() {
	m.TextInput.SetEchoMode(component.EchoPassword)
}

// SetEchoMode sets the echo mode.
func (m Model) SetEchoMode(mode component.EchoMode) {
	m.TextInput.SetEchoMode(mode)
}

// CharLimit returns the character limit.
func (m Model) CharLimit() int {
	return m.TextInput.CharLimit()
}

// SetCharLimit sets the character limit.
func (m Model) SetCharLimit(n int) {
	m.TextInput.SetCharLimit(n)
}

// Width returns the display width.
func (m Model) Width() int {
	return m.TextInput.Width()
}

// SetWidth sets the display width.
func (m Model) SetWidth(w int) {
	m.TextInput.SetWidth(w)
}

// Cursor returns the cursor position.
func (m Model) Cursor() int {
	return m.TextInput.Cursor()
}

// SetCursor sets the cursor position.
func (m Model) SetCursor(pos int) {
	m.TextInput.SetCursor(pos)
}

// CursorEnd moves cursor to end.
func (m Model) CursorEnd() {
	m.TextInput.CursorEnd()
}

// CursorStart moves cursor to start.
func (m Model) CursorStart() {
	m.TextInput.CursorStart()
}

// View renders the textinput content as a string (bubbles v2 compatible).
func (m Model) View() string {
	return m.TextInput.Value()
}

// InsertRune inserts a rune at the current cursor position.
func (m Model) InsertRune(r rune) {
	m.TextInput.InsertText(string(r))
}

// Position returns cursor position (alias for Cursor).
func (m Model) Position() int {
	return m.TextInput.Cursor()
}

// Focused returns whether the input is focused.
func (m Model) Focused() bool {
	return m.TextInput.Focused()
}

// Update handles a bubbletea message and returns the updated model + cmd.
// This mirrors bubbles v2: m, cmd := m.Update(msg)
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
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

// Reset clears the input.
func (m Model) Reset() {
	m.TextInput.Clear()
}

// Empty returns true if the input has no text.
func (m Model) Empty() bool {
	return m.TextInput.Empty()
}

// Len returns the length of the text.
func (m Model) Len() int {
	return m.TextInput.Len()
}

// SetStyle sets the text style.
func (m Model) SetStyle(style buffer.Style) {
	m.TextInput.SetStyle(style)
}

// CursorMode constants (bubbles.textinput compatible).
const (
	CursorBlink   = 0
	CursorStatic  = 1
	CursorHide    = 2
)

// SetCursorMode sets the cursor display mode.
func (m Model) SetCursorMode(mode int) {
	// fluui TextInput doesn't distinguish cursor modes yet; no-op for compat
}

// Runes returns the input as a slice of runes.
func (m Model) Runes() []rune {
	return []rune(m.TextInput.Value())
}
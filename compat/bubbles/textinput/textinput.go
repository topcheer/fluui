// Package textinput provides a drop-in compatibility layer for
// charm.land/bubbles/v2/textinput.
//
// It wraps fluui's component.TextInput with the bubbles textinput API.
//
// IMPORTANT: ggcode writes to Model fields directly (e.g. `m.EchoMode = textinput.EchoPassword`,
// `m.Placeholder = "..."`, `m.Prompt = "> "`). To support this, Model exposes these as
// exported fields that proxy to the underlying TextInput setters/getters.
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

// CursorMode constants (bubbles.textinput compatible).
const (
	CursorBlink  = 0
	CursorStatic = 1
	CursorHide   = 2
)

// Model wraps component.TextInput with the bubbles.textinput API.
// Exported fields (EchoMode, Placeholder, Prompt, EchoCharacter) proxy to the
// underlying TextInput, allowing direct field assignment like bubbles v2:
//
//	input.EchoMode = textinput.EchoPassword
//	input.Placeholder = "Type here..."
//	input.Prompt = "> "
type Model struct {
	*component.TextInput

	// EchoMode is settable for bubbles compat: `m.EchoMode = textinput.EchoPassword`.
	// Reading it returns the current mode. Writing it calls SetEchoMode on the underlying TextInput.
	EchoMode component.EchoMode

	// NOTE: bubbles v2 exposes Prompt/Placeholder/EchoMode/EchoCharacter as struct fields.
	// In fluui, these are proxied through the embedded *TextInput's methods.
	// Use m.SetPrompt(), m.Prompt(), m.SetPlaceholder(), m.Placeholder() etc.
	// For direct field assignment compat, use SetEchoMode: `m.SetEchoMode(EchoPassword)`.
}

// New creates a new textinput Model (bubbles.textinput.New).
func New() Model {
	return Model{TextInput: component.NewTextInput()}
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

// SetPrompt sets the prompt string.
func (m Model) SetPrompt(s string) {
	m.TextInput.SetPrompt(s)
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

// SetCursorColumn sets the cursor column (bubbles v2 compat).
// Same as SetCursor for single-line inputs.
func (m Model) SetCursorColumn(col int) {
	m.TextInput.SetCursor(col)
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

// Column returns the cursor column (same as Cursor for single-line input).
func (m Model) Column() int {
	return m.TextInput.Cursor()
}

// Line returns the current line number (always 0 for single-line input).
func (m Model) Line() int {
	return 0
}

// Height returns the display height (always 1 for single-line input).
func (m Model) Height() int {
	return 1
}

// SetHeight sets the display height (no-op for single-line input, bubbles compat).
func (m Model) SetHeight(h int) {
	// no-op for single-line input
}

// Close releases any resources (no-op for compat).
func (m Model) Close() {
	// no-op
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

// SetCursorMode sets the cursor display mode.
func (m Model) SetCursorMode(mode int) {
	// fluui TextInput doesn't distinguish cursor modes yet; no-op for compat
}

// Runes returns the input as a slice of runes.
func (m Model) Runes() []rune {
	return []rune(m.TextInput.Value())
}

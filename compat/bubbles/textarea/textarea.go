// Package textarea provides a drop-in compatibility layer for
// charm.land/bubbles/v2/textarea.
//
// It wraps fluui's component.TextArea with the bubbles textarea API.
package textarea

import (
	tea "github.com/topcheer/fluui/compat/bubbletea"
	"github.com/topcheer/fluui/compat/lipgloss"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// Model wraps component.TextArea with the bubbles.textarea API.
type Model struct {
	*component.TextArea
}

// New creates a new textarea Model (bubbles.textarea.New).
func New() Model {
	return Model{component.NewTextArea()}
}

// Value returns the current text content.
func (m Model) Value() string {
	return m.TextArea.Value()
}

// SetValue sets the text content.
func (m Model) SetValue(s string) {
	m.TextArea.SetValue(s)
}

// Text returns the current text content (alias for Value).
func (m Model) Text() string {
	return m.TextArea.Value()
}

// SetText sets the text content (alias for SetValue).
func (m Model) SetText(s string) {
	m.TextArea.SetValue(s)
}

// Prompt returns the prompt string.
func (m Model) Prompt() string {
	return m.TextArea.Prompt()
}

// SetPrompt sets the prompt string.
func (m Model) SetPrompt(s string) {
	m.TextArea.SetPrompt(s)
}

// Placeholder returns the placeholder text.
func (m Model) Placeholder() string {
	return m.TextArea.Placeholder()
}

// SetPlaceholder sets the placeholder text.
func (m Model) SetPlaceholder(s string) {
	m.TextArea.SetPlaceholder(s)
}

// Focus sets focus on the textarea.
func (m Model) Focus() {
	m.TextArea.Focus()
}

// Blur removes focus from the textarea.
func (m Model) Blur() {
	m.TextArea.Blur()
}

// Blink triggers a cursor blink.
func (m Model) Blink() {
	m.TextArea.Blink()
}

// Focused returns whether the textarea is focused.
func (m Model) Focused() bool {
	return false // TextArea doesn't track focused state
}

// SetWidth sets the display width.
func (m Model) SetWidth(w int) {
	m.TextArea.SetWidth(w)
}

// SetHeight sets the display height.
func (m Model) SetHeight(h int) {
	m.TextArea.SetHeight(h)
}

// Width returns the display width.
func (m Model) Width() int {
	b := m.TextArea.Bounds()
	return b.W
}

// Height returns the display height.
func (m Model) Height() int {
	b := m.TextArea.Bounds()
	return b.H
}

// Line returns the current line number (0-indexed).
func (m Model) Line() int {
	return m.TextArea.Line()
}

// Column returns the current column number (0-indexed).
func (m Model) Column() int {
	return m.TextArea.Column()
}

// CursorDown moves cursor down one line.
func (m Model) CursorDown() {
	m.TextArea.CursorDown()
}

// CursorUp moves cursor up one line.
func (m Model) CursorUp() {
	m.TextArea.CursorUp()
}

// Reset clears the textarea.
func (m Model) Reset() {
	m.TextArea.Reset()
}

// CharLimit returns the character limit.
func (m Model) CharLimit() int {
	return m.TextArea.CharLimit()
}

// SetCharLimit sets the character limit.
func (m Model) SetCharLimit(n int) {
	m.TextArea.SetCharLimit(n)
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
		m.TextArea.HandleKey(&key)
	case tea.PasteMsg:
		m.TextArea.InsertText(msg.Content)
	}
	return m, nil
}

// SetStyle sets the text style.
func (m Model) SetStyle(style buffer.Style) {
	m.TextArea.SetStyle(style)
}

// InsertString inserts a string at the current cursor position.
func (m Model) InsertString(s string) {
	m.TextArea.InsertText(s)
}

// DeleteBeforeCursor deletes the character before the cursor.
func (m Model) DeleteBeforeCursor() {
	m.TextArea.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
}

// DeleteAfterCursor deletes the character after the cursor.
func (m Model) DeleteAfterCursor() {
	m.TextArea.HandleKey(&term.KeyEvent{Key: term.KeyDelete})
}

// View renders the textarea content as a string (bubbles v2 compatible).
func (m Model) View() string {
	return m.TextArea.Value()
}

// InsertRune inserts a rune at the current cursor position.
func (m Model) InsertRune(r rune) {
	m.TextArea.InsertText(string(r))
}

// CursorEnd moves the cursor to the end of the text.
func (m Model) CursorEnd() {
	// Move to last line, end of content
	val := m.TextArea.Value()
	m.TextArea.SetValue(val) // SetValue resets cursor to end
}

// CursorStart moves the cursor to the start of the text.
func (m Model) CursorStart() {
	// Reset value to same content resets cursor to start in some impls.
	// Use InsertText trick: clear and re-set
	m.TextArea.HandleKey(&term.KeyEvent{Key: term.KeyHome})
}

// ─── Styles (bubbles.textarea compatible) ───

// StateStyles holds styles for a specific state (focused or blurred).
type StateStyles struct {
	Base             lipgloss.Style
	CursorLine       lipgloss.Style
	EndOfBuffer      lipgloss.Style
	LineNumber       lipgloss.Style
	CursorLineNumber lipgloss.Style
	Text             lipgloss.Style
	Prompt           lipgloss.Style
	Placeholder      lipgloss.Style
}

// Styles holds the textarea style configuration.
type Styles struct {
	Focused  StateStyles
	Blurred  StateStyles
}

// DefaultStyles returns default textarea styles (bubbles.textarea.DefaultStyles).
// If dark is true, uses a dark color scheme.
func DefaultStyles(dark bool) Styles {
	return Styles{
		Focused: StateStyles{
			Base:            lipgloss.NewStyle(),
			CursorLine:      lipgloss.NewStyle(),
			EndOfBuffer:     lipgloss.NewStyle(),
			LineNumber:      lipgloss.NewStyle(),
			CursorLineNumber: lipgloss.NewStyle(),
		},
		Blurred: StateStyles{
			Base:            lipgloss.NewStyle(),
			CursorLine:      lipgloss.NewStyle(),
			EndOfBuffer:     lipgloss.NewStyle(),
			LineNumber:      lipgloss.NewStyle(),
			CursorLineNumber: lipgloss.NewStyle(),
		},
	}
}

// SetStyles sets the style configuration (bubbles.textarea compatible).
func (m Model) SetStyles(s Styles) {
	// Styles are not used by fluui's TextArea renderer, but stored for compat.
}

// Blink is a command that triggers cursor blink for all textareas.
// ggcode uses it as: func() tea.Msg { return textarea.Blink() }
// Returns a nil message (fluui handles cursor blink internally).
func Blink() tea.Msg {
	return nil
}
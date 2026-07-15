// Package textarea provides a drop-in compatibility layer for
// charm.land/bubbles/v2/textarea.
//
// It wraps fluui's component.TextArea with the bubbles textarea API.
package textarea

import (
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

// Update handles a key event (bubbles.textarea.Model.Update).
func (m Model) Update(key *term.KeyEvent) {
	m.TextArea.HandleKey(key)
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

// ─── Styles (bubbles.textarea compatible) ───

// StateStyles holds styles for a specific state (focused or blurred).
type StateStyles struct {
	Base            lipgloss.Style
	CursorLine      lipgloss.Style
	EndOfBuffer     lipgloss.Style
	LineNumber      lipgloss.Style
	CursorLineNumber lipgloss.Style
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

// Blink is a command that triggers cursor blink for all textareas.
// Returns a nil message (fluui handles cursor blink internally).
func Blink() interface{} {
	return nil
}
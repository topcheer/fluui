package textarea

import (
	"testing"

	tea "github.com/topcheer/fluui/compat/bubbletea"
	"github.com/topcheer/fluui/compat/lipgloss"
	"github.com/topcheer/fluui/internal/term"
)

// P203: Tests for bubbles v2 compatible Update signature + Blink + Styles

func TestUpdate_KeyPressMsg_P203(t *testing.T) {
	m := New()
	m.SetValue("hello")
	// Cursor is at position 0, so 'a' gets inserted at beginning
	m2, cmd := m.Update(tea.KeyPressMsg{Rune: 'a'})
	if cmd != nil {
		t.Error("expected nil cmd for simple keypress")
	}
	// TextArea inserts at cursor position
	if m2.Value() != "ahello" {
		t.Errorf("expected 'ahello', got %q", m2.Value())
	}
}

func TestUpdate_SpecialKey_P203(t *testing.T) {
	m := New()
	m.SetValue("test")
	// Simulate pressing backspace
	m2, _ := m.Update(tea.KeyPressMsg{Code: term.KeyBackspace})
	_ = m2
}

func TestUpdate_PasteMsg_P203(t *testing.T) {
	m := New()
	m2, _ := m.Update(tea.PasteMsg{Content: "pasted text"})
	if m2.Value() != "pasted text" {
		t.Errorf("expected 'pasted text', got %q", m2.Value())
	}
}

func TestUpdate_UnknownMsg_P203(t *testing.T) {
	m := New()
	m.SetValue("unchanged")
	// Unknown message type should be a no-op
	m2, cmd := m.Update(tea.QuitMsg{})
	if cmd != nil {
		t.Error("expected nil cmd for unknown msg")
	}
	if m2.Value() != "unchanged" {
		t.Error("value should not change for unknown msg")
	}
}

func TestBlink_P203(t *testing.T) {
	result := Blink()
	if result != nil {
		t.Error("Blink should return nil Msg")
	}
}

func TestSetStyles_P203(t *testing.T) {
	m := New()
	s := DefaultStyles(true)
	// Verify all fields that ggcode sets
	s.Focused.Base = lipgloss.NewStyle()
	s.Focused.CursorLine = lipgloss.NewStyle()
	s.Focused.EndOfBuffer = lipgloss.NewStyle()
	s.Focused.LineNumber = lipgloss.NewStyle()
	s.Focused.CursorLineNumber = lipgloss.NewStyle()
	s.Focused.Text = lipgloss.NewStyle().Bold(true)
	s.Focused.Prompt = lipgloss.NewStyle().Bold(true)
	s.Focused.Placeholder = lipgloss.NewStyle().Bold(true)
	s.Blurred.Base = lipgloss.NewStyle()
	s.Blurred.Text = lipgloss.NewStyle().Bold(true)
	s.Blurred.Prompt = lipgloss.NewStyle()
	s.Blurred.Placeholder = lipgloss.NewStyle()
	m.SetStyles(s)
}
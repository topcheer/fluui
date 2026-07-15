package textinput

import (
	"testing"

	tea "github.com/topcheer/fluui/compat/bubbletea"
	"github.com/topcheer/fluui/internal/term"
)

// P203: Tests for bubbles v2 compatible Update signature + Blink

func TestUpdate_KeyPressMsg_P203(t *testing.T) {
	m := New()
	m.SetValue("hello")
	m2, cmd := m.Update(tea.KeyPressMsg{Rune: 'a'})
	if cmd != nil {
		t.Error("expected nil cmd for simple keypress")
	}
	if m2.Value() != "helloa" {
		t.Errorf("expected 'helloa', got %q", m2.Value())
	}
}

func TestUpdate_SpecialKey_P203(t *testing.T) {
	m := New()
	m.SetValue("test")
	m2, _ := m.Update(tea.KeyPressMsg{Code: term.KeyBackspace})
	_ = m2
}

func TestUpdate_PasteMsg_P203(t *testing.T) {
	m := New()
	m2, _ := m.Update(tea.PasteMsg{Content: "pasted"})
	if m2.Value() != "pasted" {
		t.Errorf("expected 'pasted', got %q", m2.Value())
	}
}

func TestUpdate_UnknownMsg_P203(t *testing.T) {
	m := New()
	m.SetValue("unchanged")
	m2, cmd := m.Update(tea.QuitMsg{})
	if cmd != nil {
		t.Error("expected nil cmd for unknown msg")
	}
	if m2.Value() != "unchanged" {
		t.Error("value should not change for unknown msg")
	}
}

func TestBlink_P203(t *testing.T) {
	// textinput.Blink is used as a tea.Cmd function value (not called)
	// Verify it has the right type
	var cmd tea.Cmd = Blink
	if cmd == nil {
		t.Error("Blink should not be nil")
	}
	result := cmd()
	if result != nil {
		t.Error("Blink() should return nil Msg")
	}
}
package textinput

import (
	"testing"

	"github.com/topcheer/fluui/compat/bubbletea"
	"github.com/topcheer/fluui/component"
)

// P285: cover uncovered methods after P284 Model redesign

func TestModel_SetPrompt_P285(t *testing.T) {
	m := New()
	m.SetPrompt(">> ")
	// Value receiver: field write is lost, but TextInput is synced
	if m.TextInput.Prompt() != ">> " {
		t.Errorf("SetPrompt should sync to TextInput, got %q", m.TextInput.Prompt())
	}
}

func TestModel_PromptValue_P285(t *testing.T) {
	m := New()
	m.Prompt = "? "
	if m.PromptValue() != "? " {
		t.Errorf("PromptValue should return field, got %q", m.PromptValue())
	}
}

func TestModel_SetPlaceholder_P285(t *testing.T) {
	m := New()
	m.SetPlaceholder("hint")
	// Value receiver: field write is lost, but TextInput is synced
	if m.TextInput.Placeholder() != "hint" {
		t.Errorf("SetPlaceholder should sync to TextInput, got %q", m.TextInput.Placeholder())
	}
}

func TestModel_PlaceholderValue_P285(t *testing.T) {
	m := New()
	m.Placeholder = "x"
	if m.PlaceholderValue() != "x" {
		t.Errorf("PlaceholderValue should return field, got %q", m.PlaceholderValue())
	}
}

func TestModel_SetCharLimit_P285(t *testing.T) {
	m := New()
	m.SetCharLimit(500)
	// Value receiver: field write is lost, but TextInput is synced
	if m.TextInput.CharLimit() != 500 {
		t.Errorf("SetCharLimit should sync to TextInput, got %d", m.TextInput.CharLimit())
	}
}

func TestModel_SetHeight_P285(t *testing.T) {
	m := New()
	m.SetHeight(3) // no-op for single-line, shouldn't panic
}

func TestModel_Close_P285(t *testing.T) {
	m := New()
	m.Close()
}

func TestModel_SetCursorMode_P285(t *testing.T) {
	m := New()
	m.SetCursorMode(CursorHide)
	m.SetCursorMode(CursorStatic)
}

func TestModel_sync_WithCharLimit_P285(t *testing.T) {
	m := New()
	m.CharLimit = 256
	m.EchoMode = component.EchoPassword
	m.EchoCharacter = '*'
	m.Placeholder = "pwd"
	m.Prompt = "> "
	// View calls sync() which should push all fields to TextInput
	m.View()
	if m.TextInput.CharLimit() != 256 {
		t.Errorf("sync should push CharLimit, got %d", m.TextInput.CharLimit())
	}
	if m.TextInput.EchoMode() != component.EchoPassword {
		t.Error("sync should push EchoMode")
	}
	if m.TextInput.EchoChar() != '*' {
		t.Error("sync should push EchoChar")
	}
	if m.TextInput.Placeholder() != "pwd" {
		t.Error("sync should push Placeholder")
	}
	if m.TextInput.Prompt() != "> " {
		t.Error("sync should push Prompt")
	}
}

func TestModel_sync_ZeroCharLimit_P285(t *testing.T) {
	m := New()
	m.CharLimit = 0 // zero = no limit, should NOT call SetCharLimit
	m.View()
	// Default char limit is 0 (no limit)
	if m.TextInput.CharLimit() != 0 {
		t.Errorf("zero CharLimit should not set limit, got %d", m.TextInput.CharLimit())
	}
}

func TestModel_Update_SyncsFields_P285(t *testing.T) {
	m := New()
	m.EchoMode = component.EchoPassword
	m.Placeholder = "secret"
	// Update should call sync() first
	updated, _ := m.Update(bubbletea.KeyPressMsg{Rune: 'x'})
	if updated.TextInput.EchoMode() != component.EchoPassword {
		t.Error("Update should sync EchoMode before processing")
	}
}

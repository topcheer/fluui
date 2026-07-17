package textinput

import (
	"testing"

	"github.com/topcheer/fluui/compat/bubbletea"
	"github.com/topcheer/fluui/component"
)

func TestModel_EchoModeField_P282(t *testing.T) {
	m := New()
	m.SetEchoMode(EchoPassword)
	if m.TextInput.EchoMode() != component.EchoPassword {
		t.Errorf("SetEchoMode should set EchoPassword")
	}
}

func TestModel_PlaceholderField_P282(t *testing.T) {
	m := New()
	m.Placeholder = "Type here..."
	if m.Placeholder != "Type here..." {
		t.Errorf("Placeholder field should work, got %q", m.Placeholder)
	}
}

func TestModel_PromptField_P282(t *testing.T) {
	m := New()
	m.Prompt = "> "
	if m.Prompt != "> " {
		t.Errorf("Prompt field should work, got %q", m.Prompt)
	}
}

func TestModel_EchoChar_P282(t *testing.T) {
	m := New()
	m.TextInput.SetEchoChar('#')
	if m.TextInput.EchoChar() != '#' {
		t.Errorf("EchoChar should be '#', got %q", m.TextInput.EchoChar())
	}
}

func TestModel_SetCursorColumn_P282(t *testing.T) {
	m := New()
	m.SetValue("hello")
	m.SetCursorColumn(3)
	if m.Cursor() != 3 {
		t.Errorf("SetCursorColumn should set cursor to 3, got %d", m.Cursor())
	}
}

func TestModel_Column_P282(t *testing.T) {
	m := New()
	m.SetValue("test")
	m.SetCursor(2)
	if m.Column() != 2 {
		t.Errorf("Column should equal cursor pos, got %d", m.Column())
	}
}

func TestModel_Line_P282(t *testing.T) {
	m := New()
	if m.Line() != 0 {
		t.Errorf("Line should always be 0, got %d", m.Line())
	}
}

func TestModel_Height_P282(t *testing.T) {
	m := New()
	if m.Height() != 1 {
		t.Errorf("Height should always be 1, got %d", m.Height())
	}
}

func TestModel_SetHeight_P282(t *testing.T) {
	m := New()
	m.SetHeight(5)
}

func TestModel_Close_P282(t *testing.T) {
	m := New()
	m.Close()
}

func TestModel_Update_KeyPressMsg_P282(t *testing.T) {
	m := New()
	updated, cmd := m.Update(bubbletea.KeyPressMsg{Rune: 'x'})
	if cmd != nil {
		t.Error("cmd should be nil for key press")
	}
	if updated.Value() != "x" {
		t.Errorf("Update should handle keypress, value = %q", updated.Value())
	}
}

func TestModel_Update_PasteMsg_P282(t *testing.T) {
	m := New()
	m.SetValue("hello")
	updated, _ := m.Update(bubbletea.PasteMsg{Content: " world"})
	if updated.Value() != "hello world" {
		t.Errorf("Paste should append, got %q", updated.Value())
	}
}

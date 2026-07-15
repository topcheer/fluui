package textinput

import (
	"testing"
)

// P199: Comprehensive coverage for textinput compat

func TestModelLifecycle_P199(t *testing.T) {
	m := New()
	m.Focus()
	m.Blur()
	m.Blink()
}

func TestModelValue_P199(t *testing.T) {
	m := New()
	m.SetValue("hello")
	if m.Value() != "hello" {
		t.Errorf("expected 'hello', got %q", m.Value())
	}
}

func TestModelPrompt_P199(t *testing.T) {
	m := New()
	m.SetPrompt("> ")
	if m.Prompt() != "> " {
		t.Errorf("expected '> ', got %q", m.Prompt())
	}
}

func TestModelPlaceholder_P199(t *testing.T) {
	m := New()
	m.SetPlaceholder("enter...")
	if m.Placeholder() != "enter..." {
		t.Errorf("expected 'enter...', got %q", m.Placeholder())
	}
}

func TestModelEchoMode_P199(t *testing.T) {
	m := New()
	m.SetEchoMode(EchoPassword)
	m.EchoPassword()
}

func TestModelCharLimit_P199(t *testing.T) {
	m := New()
	m.SetCharLimit(100)
	if m.CharLimit() != 100 {
		t.Errorf("expected 100, got %d", m.CharLimit())
	}
}

func TestModelWidth_P199(t *testing.T) {
	m := New()
	m.SetWidth(40)
	if m.Width() != 40 {
		t.Errorf("expected 40, got %d", m.Width())
	}
}

func TestModelCursor_P199(t *testing.T) {
	m := New()
	m.SetValue("hello")
	m.SetCursor(2)
	if m.Cursor() != 2 {
		t.Errorf("expected 2, got %d", m.Cursor())
	}
	m.CursorEnd()
	m.CursorStart()
	_ = m.Position()
}

func TestModelFocused_P199(t *testing.T) {
	m := New()
	m.Focus()
	if !m.Focused() {
		t.Error("should be focused after Focus()")
	}
	m.Blur()
}

func TestModelReset_P199(t *testing.T) {
	m := New()
	m.SetValue("test")
	m.Reset()
	if !m.Empty() {
		t.Error("should be empty after Reset")
	}
}

func TestModelLen_P199(t *testing.T) {
	m := New()
	m.SetValue("hello")
	if m.Len() != 5 {
		t.Errorf("expected 5, got %d", m.Len())
	}
}

func TestModelRunes_P199(t *testing.T) {
	m := New()
	m.SetValue("abc")
	r := m.Runes()
	if len(r) != 3 {
		t.Errorf("expected 3 runes, got %d", len(r))
	}
}

func TestModelCursorMode_P199(t *testing.T) {
	m := New()
	m.SetCursorMode(CursorHide)
}
package textinput

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	tea "github.com/topcheer/fluui/compat/bubbletea"
)

// Coverage tests for bubbles.textinput compatibility stubs.

func TestModel_SetHeight_P431(t *testing.T) {
	m := New()
	m.SetHeight(3) // no-op for single-line input
	if m.Height() != 1 {
		t.Errorf("Height = %d, want 1 (always 1)", m.Height())
	}
}

func TestModel_Close_P431(t *testing.T) {
	m := New()
	m.Close() // no-op, should not panic
}

func TestModel_SetCursorMode_P431(t *testing.T) {
	m := New()
	m.SetCursorMode(1) // no-op
	// Verify no panic
}

func TestModel_SetStyle_P431(t *testing.T) {
	m := New()
	st := buffer.Style{Fg: buffer.Red}
	m.SetStyle(st)
	// SetStyle is a no-op passthrough — just verify no panic
}

func TestModel_Runes_P431(t *testing.T) {
	m := New()
	m.SetValue("hello")
	runes := m.Runes()
	if len(runes) != 5 || string(runes) != "hello" {
		t.Errorf("Runes = %v, want [h e l l o]", runes)
	}
}

func TestModel_Line_P431(t *testing.T) {
	m := New()
	if m.Line() != 0 {
		t.Errorf("Line = %d, want 0 (always 0)", m.Line())
	}
}

func TestModel_CursorStart_P431(t *testing.T) {
	m := New()
	m.CursorStart() // should not panic
}

func TestModel_InsertRune_P431(t *testing.T) {
	m := New()
	m.InsertRune('x')
	if m.Value() != "x" {
		t.Errorf("Value = %q, want %q", m.Value(), "x")
	}
}

func TestModel_Position_P431(t *testing.T) {
	m := New()
	m.SetValue("hello")
	// SetValue moves cursor to end
	if m.Position() != 5 {
		t.Errorf("Position = %d, want 5 (after SetValue)", m.Position())
	}
}

func TestModel_Column_P431(t *testing.T) {
	m := New()
	m.SetValue("abc")
	m.SetCursor(2)
	if m.Column() != 2 {
		t.Errorf("Column = %d, want 2", m.Column())
	}
}

func TestModel_Focused_P431(t *testing.T) {
	m := New()
	if m.Focused() {
		t.Error("new model should not be focused")
	}
	m.Focus()
	if !m.Focused() {
		t.Error("should be focused after Focus()")
	}
}

func TestModel_Reset_P431(t *testing.T) {
	m := New()
	m.SetValue("hello")
	m.Reset()
	if m.Value() != "" {
		t.Errorf("Value = %q after reset, want empty", m.Value())
	}
}

func TestModel_Empty_P431(t *testing.T) {
	m := New()
	if !m.Empty() {
		t.Error("new model should be empty")
	}
	m.SetValue("x")
	if m.Empty() {
		t.Error("should not be empty after SetValue")
	}
}

func TestModel_Len_P431(t *testing.T) {
	m := New()
	m.SetValue("hello")
	if m.Len() != 5 {
		t.Errorf("Len = %d, want 5", m.Len())
	}
}

func TestModel_View_P431(t *testing.T) {
	m := New()
	m.SetValue("hello")
	v := m.View()
	if v == "" {
		t.Error("View should not be empty")
	}
}

func TestModel_Update_P431(t *testing.T) {
	m := New()
	// Send a key press message
	newModel, cmd := m.Update(tea.KeyPressMsg{Rune: 'a'})
	if cmd != nil {
		_ = cmd // command is optional
	}
	_ = newModel
}

func TestModel_SetCharLimit_P431(t *testing.T) {
	m := New()
	m.SetCharLimit(10)
	// SetCharLimit propagates to TextInput; Model.CharLimit may not reflect on value receiver
	if m.TextInput.CharLimit() != 10 {
		t.Errorf("TextInput CharLimit = %d, want 10", m.TextInput.CharLimit())
	}
}

func TestModel_WidthSetWidth_P431(t *testing.T) {
	m := New()
	m.SetWidth(20)
	if m.Width() != 20 {
		t.Errorf("Width = %d, want 20", m.Width())
	}
}

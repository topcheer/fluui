package textarea

import (
	"testing"

	"github.com/topcheer/fluui/compat/lipgloss"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// P199: Comprehensive coverage for textarea compat

func TestModelLifecycle_P199(t *testing.T) {
	m := New()
	m.Focus()
	m.Blur()
	m.Blink()
	_ = m.Focused()
}

func TestModelValue_P199(t *testing.T) {
	m := New()
	m.SetValue("hello world")
	if m.Value() != "hello world" {
		t.Errorf("expected 'hello world', got %q", m.Value())
	}
}

func TestModelText_P199(t *testing.T) {
	m := New()
	m.SetText("line1\nline2")
	if m.Text() != "line1\nline2" {
		t.Errorf("expected multiline, got %q", m.Text())
	}
}

func TestModelPrompt_P199(t *testing.T) {
	m := New()
	m.SetPrompt("> ")
	_ = m.Prompt()
}

func TestModelPlaceholder_P199(t *testing.T) {
	m := New()
	m.SetPlaceholder("type here...")
	_ = m.Placeholder()
}

func TestModelDimensions_P199(t *testing.T) {
	m := New()
	m.SetWidth(40)
	m.SetHeight(10)
}

func TestModelCursor_P199(t *testing.T) {
	m := New()
	m.SetValue("line1\nline2")
	_ = m.Line()
	_ = m.Column()
	m.CursorDown()
	m.CursorUp()
}

func TestModelReset_P199(t *testing.T) {
	m := New()
	m.SetValue("test")
	m.Reset()
}

func TestModelCharLimit_P199(t *testing.T) {
	m := New()
	m.SetCharLimit(1000)
	_ = m.CharLimit()
}

func TestModelUpdate_P199(t *testing.T) {
	m := New()
	m.Update(&term.KeyEvent{Key: term.KeyEnter})
	m.Update(&term.KeyEvent{Rune: 'a'})
}

func TestModelInsertDelete_P199(t *testing.T) {
	m := New()
	m.SetValue("hello")
	m.InsertString(" world")
	m.DeleteBeforeCursor()
	m.DeleteAfterCursor()
}

func TestModelSetStyle_P199(t *testing.T) {
	m := New()
	m.SetStyle(buffer.Style{})
}

func TestDefaultStyles_P199(t *testing.T) {
	s := DefaultStyles(true)
	s.Focused.Base = lipgloss.NewStyle()
	s.Blurred.Base = lipgloss.NewStyle()
	_ = s
}

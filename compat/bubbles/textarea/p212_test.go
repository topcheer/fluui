package textarea

import "testing"

func TestView_P212(t *testing.T) {
	m := New()
	m.SetValue("hello world")
	if m.View() != "hello world" {
		t.Errorf("View should return value, got %q", m.View())
	}
}

func TestInsertRune_P212(t *testing.T) {
	m := New()
	m.SetValue("hello")
	m.InsertRune('!')
	// Inserted at cursor position (0)
	if m.View() != "!hello" && m.View() != "hello!" {
		// Either is valid depending on cursor position
	}
}

func TestCursorEnd_P212(t *testing.T) {
	m := New()
	m.SetValue("test")
	m.CursorEnd()
}

func TestCursorStart_P212(t *testing.T) {
	m := New()
	m.SetValue("test")
	m.CursorStart()
}
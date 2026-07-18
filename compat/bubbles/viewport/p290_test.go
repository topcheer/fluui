package viewport

import (
	"strings"
	"testing"

	tea "github.com/topcheer/fluui/compat/bubbletea"
)

func TestModel_View_YOffsetBeyondLines_P290(t *testing.T) {
	m := New(WithWidth(10), WithHeight(3))
	m.SetContent("line1\nline2")
	m.SetYOffset(100)
	_ = m.View()
}

func TestModel_View_HeightBeyondLines_P290(t *testing.T) {
	m := New(WithWidth(10), WithHeight(10))
	m.SetContent("only one line")
	m.SetYOffset(0)
	_ = m.View()
}

func TestModel_ScrollIndicator_NotAtBottom_P290(t *testing.T) {
	m := New(WithWidth(10), WithHeight(3))
	content := make([]string, 20)
	for i := range content {
		content[i] = "line"
	}
	m.SetContent(strings.Join(content, "\n"))
	m.SetYOffset(0)
	if m.AtBottom() {
		t.Error("should not be at bottom")
	}
}

func TestModel_ScrollIndicator_AutoFollow_P290(t *testing.T) {
	m := New(WithWidth(10), WithHeight(3))
	m.SetContent("a\nb\nc\nd\ne")
	m.autoFollow = true
	m.SetYOffset(0)
}

func TestModel_Update_MouseWheelMsg_P290(t *testing.T) {
	m := New(WithWidth(20), WithHeight(5))
	m.SetContent(strings.Repeat("line\n", 20))
	u1, _ := m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	_ = u1
	u2, _ := m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	_ = u2
}

func TestModel_ScrollDown_PastBottom_P290(t *testing.T) {
	m := New(WithWidth(10), WithHeight(3))
	m.SetContent("line1\nline2\nline3")
	m.ScrollDown(100)
	if !m.AtBottom() {
		t.Error("should be at bottom")
	}
}

func TestModel_SetSize_P290(t *testing.T) {
	m := New(WithWidth(10), WithHeight(5))
	m.SetContent(strings.Repeat("x\n", 20))
	m.SetSize(30, 10)
	if m.Width() != 30 || m.Height() != 10 {
		t.Errorf("expected 30x10, got %dx%d", m.Width(), m.Height())
	}
}

func TestModel_SetWidth_P290(t *testing.T) {
	m := New(WithWidth(10), WithHeight(5))
	m.SetWidth(50)
	if m.Width() != 50 {
		t.Errorf("expected 50, got %d", m.Width())
	}
}

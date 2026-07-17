package viewport

import (
	"strings"
	"testing"

	tea "github.com/topcheer/fluui/compat/bubbletea"
)

func TestView_YOffsetBeyondLines_P263(t *testing.T) {
	m := New(WithWidth(10), WithHeight(5))
	m.SetContent(strings.Repeat("line\n", 3))
	m.SetYOffset(100)
	v := m.View()
	if v == "" {
		t.Error("view should not be empty even with yOffset beyond lines")
	}
}

func TestScrollIndicator_NoneNeeded_P263(t *testing.T) {
	m := New(WithWidth(10), WithHeight(5))
	m.SetContent("short")
	_ = m.ScrollIndicatorStyle()
}

func TestScrollIndicator_Shown_P263(t *testing.T) {
	m := New(WithWidth(10), WithHeight(2))
	m.SetContent(strings.Repeat("line\n", 20))
	s := m.ScrollIndicatorStyle()
	_ = s
}

func TestUpdate_KeyDown_P263(t *testing.T) {
	m := New(WithWidth(10), WithHeight(3))
	m.SetContent(strings.Repeat("line\n", 20))
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
}

func TestUpdate_KeyPgUp_P263(t *testing.T) {
	m := New(WithWidth(10), WithHeight(3))
	m.SetContent(strings.Repeat("line\n", 20))
	m.SetYOffset(10)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyPgUp})
}

func TestUpdate_KeyPgDn_P263(t *testing.T) {
	m := New(WithWidth(10), WithHeight(3))
	m.SetContent(strings.Repeat("line\n", 20))
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyPgDn})
}

func TestUpdate_KeyHome_P263(t *testing.T) {
	m := New(WithWidth(10), WithHeight(3))
	m.SetContent(strings.Repeat("line\n", 20))
	m.SetYOffset(5)
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyHome})
	if m.YOffset() != 0 {
		t.Error("home should scroll to top")
	}
}

func TestUpdate_KeyEnd_P263(t *testing.T) {
	m := New(WithWidth(10), WithHeight(3))
	m.SetContent(strings.Repeat("line\n", 20))
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
}

func TestUpdate_MouseWheel_P263(t *testing.T) {
	m := New(WithWidth(10), WithHeight(3))
	m.SetContent(strings.Repeat("line\n", 20))
	m, _ = m.Update(tea.MouseWheelMsg{Up: true})
	m, _ = m.Update(tea.MouseWheelMsg{Down: true})
}

func TestClampYOffset_Negative_P263(t *testing.T) {
	m := New(WithWidth(10), WithHeight(3))
	m.SetContent(strings.Repeat("line\n", 20))
	m.SetYOffset(-5)
	if m.YOffset() < 0 {
		t.Error("yOffset should be clamped to >= 0")
	}
}

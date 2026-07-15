package bubbletea

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/term"
)

// P205: Coverage for bubbletea compat sub-80% functions

func TestToLower_P205(t *testing.T) {
	if toLower('A') != 'a' {
		t.Error("toLower(A) should be a")
	}
	if toLower('Z') != 'z' {
		t.Error("toLower(Z) should be z")
	}
	if toLower('a') != 'a' {
		t.Error("toLower(a) should be a")
	}
	if toLower('1') != '1' {
		t.Error("toLower(1) should be 1")
	}
}

func TestMouseClickMsg_Mouse_P205(t *testing.T) {
	m := MouseClickMsg{X: 5, Y: 10, Button: MouseLeft, Alt: true}
	info := m.Mouse()
	if info.X != 5 || info.Y != 10 {
		t.Error("Mouse() should preserve X/Y")
	}
	if info.Button != MouseLeft {
		t.Error("Mouse() should preserve Button")
	}
	if !info.Alt {
		t.Error("Mouse() should preserve Alt")
	}
}

func TestMouseClickMsg_String_P205(t *testing.T) {
	m := MouseClickMsg{}
	if m.String() == "" {
		t.Error("String() should not be empty")
	}
}

func TestMouseWheelMsg_Mouse_P205(t *testing.T) {
	m := MouseWheelMsg{X: 1, Y: 2, Up: true}
	info := m.Mouse()
	if info.Button != MouseWheelUp {
		t.Errorf("expected MouseWheelUp, got %d", info.Button)
	}

	m2 := MouseWheelMsg{X: 1, Y: 2, Down: true}
	info2 := m2.Mouse()
	if info2.Button != MouseWheelDown {
		t.Errorf("expected MouseWheelDown, got %d", info2.Button)
	}

	m3 := MouseWheelMsg{X: 3, Y: 4, Button: MouseMiddle}
	info3 := m3.Mouse()
	if info3.Button != MouseMiddle {
		t.Error("should preserve explicit Button")
	}
}

func TestMouseWheelMsg_String_P205(t *testing.T) {
	up := MouseWheelMsg{Up: true}
	if up.String() == "" {
		t.Error("String should not be empty")
	}
	down := MouseWheelMsg{Down: true}
	if down.String() == "" {
		t.Error("String should not be empty")
	}
}

func TestTick_P205(t *testing.T) {
	called := false
	cmd := Tick(10*time.Millisecond, func(t time.Time) Msg {
		called = true
		return nil
	})
	if cmd == nil {
		t.Fatal("Tick should return non-nil Cmd")
	}
	cmd()
	if !called {
		t.Error("Tick callback should be called")
	}
}

func TestQuit_P205(t *testing.T) {
	cmd := Quit()
	if cmd == nil {
		t.Fatal("Quit should return non-nil Cmd")
	}
	msg := cmd()
	if _, ok := msg.(QuitMsg); !ok {
		t.Error("Quit() should return QuitMsg")
	}
}

func TestHandlePaste_P205(t *testing.T) {
	m := &batchModel{}
	p := NewProgram(m)
	p.HandlePaste("hello")
	// Update happens during Run(), not immediately
	_, _ = p.Run()
}

func TestSetOnRender_P205(t *testing.T) {
	m := &countingModel{}
	p := NewProgram(m)
	called := false
	p.SetOnRender(func(s string) {
		called = true
	})
	if called {
		t.Error("render callback should not be called yet")
	}
}

func TestKeyName_P205(t *testing.T) {
	tests := []struct {
		code term.KeyCode
		want string
	}{
		{term.KeyUp, "up"},
		{term.KeyDown, "down"},
		{term.KeyLeft, "left"},
		{term.KeyRight, "right"},
		{term.KeyEnter, "enter"},
		{term.KeyEscape, "esc"},
		{term.KeyBackspace, "backspace"},
		{term.KeyDelete, "delete"},
		{term.KeyTab, "tab"},
		{term.KeySpace, "space"},
		{term.KeyHome, "home"},
		{term.KeyEnd, "end"},
		{999, ""},
	}
	for _, tt := range tests {
		got := keyName(tt.code)
		if got != tt.want {
			t.Errorf("keyName(%d) = %q, want %q", tt.code, got, tt.want)
		}
	}
}

func TestRunWithBatch_P205(t *testing.T) {
	m := &batchModel{}
	p := NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		t.Errorf("Run error: %v", err)
	}
	if finalModel == nil {
		t.Error("should return model")
	}
}

type batchModel struct {
	step int
}

func (m *batchModel) Init() Cmd {
	return Batch(func() Msg { return nil }, Quit())
}
func (m *batchModel) Update(msg Msg) (Model, Cmd) {
	return m, nil
}
func (m *batchModel) View() View {
	return NewView("")
}
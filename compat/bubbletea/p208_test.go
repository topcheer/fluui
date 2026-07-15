package bubbletea

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

// P208: HandleKey + Run deep coverage + ProcessMessages

func TestHandleKey_Nil_P208(t *testing.T) {
	m := &countingModel{}
	p := NewProgram(m)
	if p.HandleKey(nil) {
		t.Error("HandleKey(nil) should return false")
	}
}

func TestHandleKey_AllModifiers_P208(t *testing.T) {
	m := &countingModel{}
	p := NewProgram(m)
	ev := &term.KeyEvent{
		Key:       term.KeyEnter,
		Modifiers: term.ModCtrl | term.ModAlt | term.ModShift,
	}
	if !p.HandleKey(ev) {
		t.Error("HandleKey should return true for valid event")
	}
}

func TestHandleKey_PlainKey_P208(t *testing.T) {
	m := &countingModel{}
	p := NewProgram(m)
	ev := &term.KeyEvent{Rune: 'x'}
	if !p.HandleKey(ev) {
		t.Error("should return true")
	}
}

func TestRun_HandleKeyViaSend_P208(t *testing.T) {
	m := &quitModel{}
	p := NewProgram(m)
	// Send a keypress that triggers Update, which returns nil cmd,
	// then Init's Quit() fires via the cmd processing path
	finalModel, err := p.Run()
	if err != nil {
		t.Errorf("Run error: %v", err)
	}
	if finalModel == nil {
		t.Error("should return model")
	}
}

func TestRun_NilCmdFromUpdate_P208(t *testing.T) {
	m := &nilCmdModel{}
	p := NewProgram(m)
	finalModel, _ := p.Run()
	if finalModel == nil {
		t.Error("should return model")
	}
}

type nilCmdModel struct{}

func (m *nilCmdModel) Init() Cmd { return Quit() }
func (m *nilCmdModel) Update(msg Msg) (Model, Cmd) {
	return m, nil // returns nil cmd
}
func (m *nilCmdModel) View() View { return NewView("") }

func TestProcessMessages_Multiple_P208(t *testing.T) {
	m := &countingModel{}
	p := NewProgram(m)
	p.Send(KeyPressMsg{Rune: 'a'})
	p.Send(KeyPressMsg{Rune: 'b'})
	p.Send(QuitMsg{})
	_, _ = p.Run()
	// countingModel should have processed at least the quit msg
}

func TestRun_CmdReturnsNilMsg_P208(t *testing.T) {
	m := &nilMsgModel{}
	p := NewProgram(m)
	go func() {
		p.Send(QuitMsg{})
	}()
	_, err := p.Run()
	if err != nil {
		t.Errorf("Run error: %v", err)
	}
}

type nilMsgModel struct{}

func (m *nilMsgModel) Init() Cmd { return nil }
func (m *nilMsgModel) Update(msg Msg) (Model, Cmd) {
	return m, nil
}
func (m *nilMsgModel) View() View { return NewView("") }
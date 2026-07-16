package bubbletea

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

// P219: ProcessMessages + Run deep coverage

// cmdReturningQuit — model whose Update returns a Cmd that returns QuitMsg
type cmdQuitModel struct{}

func (m *cmdQuitModel) Init() Cmd { return nil }
func (m *cmdQuitModel) Update(msg Msg) (Model, Cmd) {
	return m, Quit()
}
func (m *cmdQuitModel) View() View { return NewView("") }

func TestProcessMessages_CmdReturnsQuit_P219(t *testing.T) {
	m := &cmdQuitModel{}
	p := NewProgram(m)
	// Send a message that triggers Update→cmd→QuitMsg
	p.Send(KeyPressMsg{Code: term.KeyEnter})
	result := p.ProcessMessages()
	if !result {
		t.Error("ProcessMessages should return true when cmd returns QuitMsg")
	}
}

func TestProcessMessages_Empty_P219(t *testing.T) {
	m := &cmdQuitModel{}
	p := NewProgram(m)
	// No messages queued → should return false immediately
	if p.ProcessMessages() {
		t.Error("ProcessMessages should return false when no messages")
	}
}

func TestProcessMessages_CmdReturnsNilMsg_P219(t *testing.T) {
	m := &nilCmdModel{}
	p := NewProgram(m)
	p.Send(KeyPressMsg{Rune: 'x'})
	// ProcessMessages processes the key, cmd returns nil msg, then hits default
	_ = p.ProcessMessages()
	// Either true or false is valid depending on timing — just don't hang
}

func TestRun_QuitChannel_P219(t *testing.T) {
	m := &noopModel{}
	p := NewProgram(m)
	// Close quitCh directly to trigger the <-p.quitCh branch
	go func() {
		p.quitCh <- struct{}{}
	}()
	model, err := p.Run()
	if err != nil {
		t.Errorf("Run error: %v", err)
	}
	if model == nil {
		t.Error("should return model")
	}
}

func TestRun_CmdReturnsQuit_P219(t *testing.T) {
	m := &cmdQuitModel{}
	p := NewProgram(m)
	// Send a key → Update returns Quit() → Run should exit
	go func() {
		p.Send(KeyPressMsg{Rune: 'a'})
	}()
	_, err := p.Run()
	if err != nil {
		t.Errorf("Run error: %v", err)
	}
}

func TestRun_CmdReturnsNilMsg_P219(t *testing.T) {
	m := &nilCmdModel{}
	p := NewProgram(m)
	go func() {
		p.Send(KeyPressMsg{Rune: 'b'})
		// Then send quit to exit Run
		p.Send(QuitMsg{})
	}()
	_, err := p.Run()
	if err != nil {
		t.Errorf("Run error: %v", err)
	}
}

type noopModel struct{}

func (m *noopModel) Init() Cmd                  { return nil }
func (m *noopModel) Update(msg Msg) (Model, Cmd) { return m, nil }
func (m *noopModel) View() View                  { return NewView("") }
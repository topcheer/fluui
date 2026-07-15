package bubbletea

import (
	"testing"
)

// P201: Tests for Run() returning (Model, error)

func TestProgramRunReturnsModel_P201(t *testing.T) {
	m := &countingModel{}
	p := NewProgram(m)
	// Send quit immediately
	go func() {
		p.Send(QuitMsg{})
	}()
	finalModel, err := p.Run()
	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}
	if finalModel == nil {
		t.Error("Run() should return non-nil Model")
	}
	// Verify type assertion works (ggcode pattern)
	if cm, ok := finalModel.(*countingModel); ok {
		_ = cm
	} else {
		t.Error("type assertion should succeed")
	}
}

func TestProgramRunViaQuitCmd_P201(t *testing.T) {
	m := &quitModel{}
	p := NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		t.Errorf("Run() error: %v", err)
	}
	if finalModel == nil {
		t.Error("should return model")
	}
}

// countingModel tracks messages for testing.
type countingModel struct {
	count int
}

func (m *countingModel) Init() Cmd { return nil }
func (m *countingModel) Update(msg Msg) (Model, Cmd) {
	m.count++
	return m, nil
}
func (m *countingModel) View() View {
	return NewView("test")
}

// quitModel returns Quit() from Init.
type quitModel struct{}

func (m *quitModel) Init() Cmd { return Quit() }
func (m *quitModel) Update(msg Msg) (Model, Cmd) {
	return m, nil
}
func (m *quitModel) View() View {
	return NewView("")
}
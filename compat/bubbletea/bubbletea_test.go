package bubbletea

import (
	"testing"
	"time"
)

// ─── Tests ───

func TestNewProgram(t *testing.T) {
	m := &testModel{text: "hello"}
	p := NewProgram(m)
	if p == nil {
		t.Fatal("NewProgram returned nil")
	}
	if p.Render() != "hello" {
		t.Error("Render should return model's View()")
	}
}

func TestProgramSend(t *testing.T) {
	m := &testModel{text: "init"}
	p := NewProgram(m)
	p.Send(KeyPressMsg{Rune: 'x'})
	p.ProcessMessages()
	if m.lastMsg == nil {
		t.Error("model should have received message")
	}
}

func TestProgramQuit(t *testing.T) {
	m := &testModel{text: "hello"}
	p := NewProgram(m)
	p.Send(QuitMsg{})
	if !p.ProcessMessages() {
		t.Error("ProcessMessages should return true on Quit")
	}
}

func TestProgramHandleKey(t *testing.T) {
	m := &testModel{text: "hello"}
	p := NewProgram(m)
	p.HandleKey(nil) // should not crash
}

func TestProgramSetSize(t *testing.T) {
	m := &testModel{text: "hello"}
	p := NewProgram(m)
	p.SetSize(100, 30)
	if p.Width() != 100 || p.Height() != 30 {
		t.Error("SetSize should update dimensions")
	}
}

func TestBatch(t *testing.T) {
	called := 0
	c1 := func() Msg { called++; return nil }
	c2 := func() Msg { called++; return nil }
	cmd := Batch(c1, c2)
	if cmd != nil {
		cmd()
	}
	if called != 2 {
		t.Errorf("expected 2 calls, got %d", called)
	}
}

func TestBatchEmpty(t *testing.T) {
	cmd := Batch()
	if cmd != nil {
		t.Error("empty Batch should return nil")
	}
}

func TestSequence(t *testing.T) {
	called := 0
	c1 := func() Msg { called++; return nil }
	c2 := func() Msg { called++; return KeyPressMsg{Rune: 'x'} }
	cmd := Sequence(c1, c2)
	if cmd != nil {
		msg := cmd()
		if msg == nil {
			t.Error("last non-nil msg should be returned")
		}
	}
	if called != 2 {
		t.Errorf("expected 2 calls, got %d", called)
	}
}

func TestQuit(t *testing.T) {
	cmd := Quit()
	if cmd == nil {
		t.Fatal("Quit should return non-nil Cmd")
	}
	msg := cmd()
	if _, ok := msg.(QuitMsg); !ok {
		t.Error("Quit() should return QuitMsg")
	}
}

func TestKeyPressMsgString(t *testing.T) {
	k := KeyPressMsg{Rune: 'a'}
	if k.String() != "a" {
		t.Error("rune key should return its character")
	}
	k2 := KeyPressMsg{Code: 0x41} // KeyUp
	_ = k2.String() // should not crash
}

func TestProgramIsDirty(t *testing.T) {
	m := &testModel{text: "hello"}
	p := NewProgram(m)
	if !p.IsDirty() {
		t.Error("should be dirty after init")
	}
	p.MarkClean()
	if p.IsDirty() {
		t.Error("should not be dirty after MarkClean")
	}
}

func TestProgramModel(t *testing.T) {
	m := &testModel{text: "hello"}
	p := NewProgram(m)
	if p.Model() != m {
		t.Error("Model() should return the model")
	}
}

func TestEveryTick(t *testing.T) {
	cmd := Every(100, func(t Time) Msg {
		return KeyPressMsg{Rune: 't'}
	})
	if cmd == nil {
		t.Fatal("Every should return non-nil")
	}
	msg := cmd()
	if msg == nil {
		t.Error("Every should return a msg")
	}
}

func TestWithOptions(t *testing.T) {
	m := &testModel{text: "hello"}
	p := NewProgram(m, WithAltScreen(), WithMouseCellMotion(), WithFPS(60))
	if p == nil {
		t.Fatal("options should not crash")
	}
}

// ─── Test Model ───

type testModel struct {
	text    string
	lastMsg Msg
}

func (m *testModel) Init() Cmd { return nil }

func (m *testModel) Update(msg Msg) (Model, Cmd) {
	m.lastMsg = msg
	if _, ok := msg.(KeyPressMsg); ok {
		kp := msg.(KeyPressMsg)
		if kp.Rune != 0 {
			m.text += string(kp.Rune)
		}
	}
	return m, nil
}

func (m *testModel) View() View {
	return NewView(m.text)
}

// Ensure package compiles with time import
var _ = time.Second
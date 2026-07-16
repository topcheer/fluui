package event

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// P233: cover defensive branches in elm_adapter.go
// Lines 86-88 (execCmd nil cmd), 108-110 (pumpCommands cmd goroutine),
// 153-155 (OnKey cmd goroutine), 165-167 (model nil guard),
// 210-212 (SendMsg cmd goroutine), 225-227 (BatchCmd nil cmd skip)

// nilReturningModel returns a cmd that returns nil msg, and returns nil newModel
type nilReturningModel struct {
	updateCount int
	returnsNil  bool
}

func (m *nilReturningModel) Init() ElmCmd { return nil }
func (m *nilReturningModel) Update(msg ElmMsg) (ElmModel, ElmCmd) {
	m.updateCount++
	if m.returnsNil {
		return nil, nil
	}
	// Return cmd that returns nil msg
	return m, func() ElmMsg { return nil }
}
func (m *nilReturningModel) View() string { return "" }

func TestElmAdapter_ExecCmdNilCmd_P233(t *testing.T) {
	// execCmd(nil) should return early — covered by Init returning nil
	m := &counterModel{}
	a := NewElmAdapter(m)
	// SendMsg with a msg that triggers nil cmd — execCmd(nil) path
	a.SendMsg(ElmMsg(customMsg{val: "test"}))
}

func TestElmAdapter_NilNewModelFromUpdate_P233(t *testing.T) {
	// When Update returns nil model, adapter should keep old model
	m := &nilReturningModel{returnsNil: true}
	a := NewElmAdapter(m)
	a.SendMsg(ElmMsg(customMsg{}))
	// Model should still be the original (not nil)
	if a.Model() == nil {
		t.Error("model should not be nil after nil-returning Update")
	}
}

func TestElmAdapter_OnKeyCmdExecuted_P233(t *testing.T) {
	// OnKey should trigger execCmd when Update returns a cmd
	m := &counterModel{}
	a := NewElmAdapter(m)
	a.OnKey(&term.KeyEvent{Key: term.KeyEnter})
	time.Sleep(50 * time.Millisecond) // wait for goroutine
}

func TestElmAdapter_OnPaintNilModel_P233(t *testing.T) {
	// Set model to nil after construction, then OnPaint should not crash
	m := &counterModel{}
	a := NewElmAdapter(m)
	a.SetModel(nil)
	buf := buffer.NewBuffer(10, 5)
	a.OnPaint(buf) // model is nil — should not crash
}

func TestElmAdapter_BatchCmdWithNilEntry_P233(t *testing.T) {
	// BatchCmd should skip nil cmds within the slice
	calls := 0
	cmd1 := func() ElmMsg { calls++; return nil }
	cmdNil := (func() ElmMsg)(nil)
	cmd2 := func() ElmMsg { calls++; return customMsg{val: "last"} }
	batch := BatchCmd(cmd1, cmdNil, cmd2)
	if batch == nil {
		t.Fatal("expected non-nil batch")
	}
	msg := batch()
	if calls != 2 {
		t.Errorf("expected 2 calls (nil skipped), got %d", calls)
	}
	if msg == nil {
		t.Error("expected non-nil msg from last cmd")
	}
}

func TestElmAdapter_BatchCmdAllNil_P233(t *testing.T) {
	cmdNil := (func() ElmMsg)(nil)
	cmdNil2 := (func() ElmMsg)(nil)
	batch := BatchCmd(cmdNil, cmdNil2)
	if batch == nil {
		t.Fatal("expected non-nil batch")
	}
	msg := batch()
	if msg != nil {
		t.Error("expected nil msg when all cmds are nil")
	}
}

func TestElmAdapter_SendMsgCmdExecuted_P233(t *testing.T) {
	// SendMsg should trigger execCmd when Update returns a cmd.
	// The cmd runs in a goroutine, sends result to cmdCh.
	// Then we call OnPaint which triggers pumpCommands to process it.
	m := &counterModel{}
	a := NewElmAdapter(m)
	// KeyMsg with KeyEnter triggers a cmd that returns customMsg{val:"done"}
	a.SendMsg(ElmMsg(KeyMsg{Key: &term.KeyEvent{Key: term.KeyEnter}}))
	// Wait for goroutine to put msg in cmdCh
	time.Sleep(100 * time.Millisecond)
	// Trigger pumpCommands via OnPaint
	buf := buffer.NewBuffer(10, 1)
	a.OnPaint(buf)
	time.Sleep(50 * time.Millisecond)
	// customMsg should have been processed
	m2 := a.Model().(*counterModel)
	if m2.text != "done" {
		t.Errorf("expected 'done', got %q", m2.text)
	}
}

package event

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

type stubElmModel struct {
	view string
	cmd  ElmCmd
}

func (m *stubElmModel) Init() ElmCmd { return nil }
func (m *stubElmModel) Update(msg ElmMsg) (ElmModel, ElmCmd) {
	return m, m.cmd
}
func (m *stubElmModel) View() string { return m.view }

func TestElmAdapter_ExecCmd_Nil_P280(t *testing.T) {
	a := NewElmAdapter(&stubElmModel{view: "test"})
	a.execCmd(nil)
}

func TestElmAdapter_ExecCmd_WithMsg_P280(t *testing.T) {
	a := NewElmAdapter(&stubElmModel{view: "test"})
	a.execCmd(func() ElmMsg { return "test-msg" })
	time.Sleep(50 * time.Millisecond)
}

func TestElmAdapter_PumpCommands_P280(t *testing.T) {
	a := NewElmAdapter(&stubElmModel{view: "test"})
	a.cmdCh <- "msg1"
	time.Sleep(50 * time.Millisecond)
	a.pumpCommands()
}

func TestElmAdapter_OnPaint_P280(t *testing.T) {
	a := NewElmAdapter(&stubElmModel{view: "hello world"})
	buf := buffer.NewBuffer(40, 5)
	a.OnPaint(buf)
}

func TestElmAdapter_OnResize_P280(t *testing.T) {
	m := &stubElmModel{view: "test", cmd: func() ElmMsg { return "resize-done" }}
	a := NewElmAdapter(m)
	a.OnResize(80, 24)
	time.Sleep(50 * time.Millisecond)
}

package event

import (
	"sync"
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

type counterModel struct {
	count int
	text  string
}

func (m *counterModel) Init() ElmCmd { return nil }
func (m *counterModel) Update(msg ElmMsg) (ElmModel, ElmCmd) {
	switch msg := msg.(type) {
	case KeyMsg:
		if msg.Key != nil && msg.Key.Rune == 'i' {
			m.count++
			return m, nil
		}
		if msg.Key != nil && msg.Key.Key == term.KeyEnter {
			return m, func() ElmMsg { return customMsg{val: "done"} }
		}
	case customMsg:
		m.text = msg.val
		return m, nil
	}
	return m, nil
}
func (m *counterModel) View() string {
	return "count: " + itoaP(m.count) + " text: " + m.text
}

type customMsg struct{ val string }
type emptyViewModel struct{}

func (m *emptyViewModel) Init() ElmCmd                    { return nil }
func (m *emptyViewModel) Update(ElmMsg) (ElmModel, ElmCmd) { return m, nil }
func (m *emptyViewModel) View() string                   { return "" }

func TestP167_ElmAdapter_Basic(t *testing.T) {
	m := &counterModel{}
	a := NewElmAdapter(m)
	if a.Model() == nil {
		t.Error("expected non-nil model")
	}
}

func TestP167_ElmAdapter_OnKey(t *testing.T) {
	m := &counterModel{}
	a := NewElmAdapter(m)
	handled := a.OnKey(&term.KeyEvent{Rune: 'i', Key: term.KeyUnknown})
	if !handled {
		t.Error("expected handled")
	}
	m2 := a.Model().(*counterModel)
	if m2.count != 1 {
		t.Errorf("expected count 1, got %d", m2.count)
	}
}

func TestP167_ElmAdapter_OnKey_Multiple(t *testing.T) {
	m := &counterModel{}
	a := NewElmAdapter(m)
	for i := 0; i < 5; i++ {
		a.OnKey(&term.KeyEvent{Rune: 'i', Key: term.KeyUnknown})
	}
	m2 := a.Model().(*counterModel)
	if m2.count != 5 {
		t.Errorf("expected count 5, got %d", m2.count)
	}
}

func TestP167_ElmAdapter_CmdExecution(t *testing.T) {
	m := &counterModel{}
	a := NewElmAdapter(m)
	a.OnKey(&term.KeyEvent{Key: term.KeyEnter})
	time.Sleep(50 * time.Millisecond)
	a.pumpCommands()
	m2 := a.Model().(*counterModel)
	if m2.text != "done" {
		t.Errorf("expected text 'done', got %q", m2.text)
	}
}

func TestP167_ElmAdapter_OnResize(t *testing.T) {
	m := &counterModel{}
	a := NewElmAdapter(m)
	a.OnResize(120, 40)
	if a.width != 120 || a.height != 40 {
		t.Errorf("expected 120x40, got %dx%d", a.width, a.height)
	}
}

func TestP167_ElmAdapter_OnPaint(t *testing.T) {
	m := &counterModel{count: 42, text: "hello"}
	a := NewElmAdapter(m)
	buf := buffer.NewBuffer(40, 5)
	a.OnPaint(buf)
	hasContent := false
	for y := 0; y < 5; y++ {
		for x := 0; x < 40; x++ {
			if buf.GetCell(x, y).Rune != 0 {
				hasContent = true
				break
			}
		}
	}
	if !hasContent {
		t.Error("expected rendered content")
	}
}

func TestP167_ElmAdapter_IsDirty(t *testing.T) {
	m := &counterModel{}
	a := NewElmAdapter(m)
	if !a.IsDirty() {
		t.Error("expected dirty initially")
	}
	a.OnKey(&term.KeyEvent{Rune: 'i', Key: term.KeyUnknown})
	if !a.IsDirty() {
		t.Error("expected dirty after key")
	}
}

func TestP167_ElmAdapter_SendMsg(t *testing.T) {
	m := &counterModel{}
	a := NewElmAdapter(m)
	a.SendMsg(ElmMsg(customMsg{val: "custom"}))
	m2 := a.Model().(*counterModel)
	if m2.text != "custom" {
		t.Errorf("expected 'custom', got %q", m2.text)
	}
}

func TestP167_ElmAdapter_SetModel(t *testing.T) {
	m1 := &counterModel{count: 1}
	a := NewElmAdapter(m1)
	m2 := &counterModel{count: 99}
	a.SetModel(m2)
	m3 := a.Model().(*counterModel)
	if m3.count != 99 {
		t.Errorf("expected 99, got %d", m3.count)
	}
}

func TestP167_ElmAdapter_OnPaint_Empty(t *testing.T) {
	m := &emptyViewModel{}
	a := NewElmAdapter(m)
	buf := buffer.NewBuffer(10, 5)
	a.OnPaint(buf)
}

func TestP167_ElmAdapter_Concurrent(t *testing.T) {
	m := &counterModel{}
	a := NewElmAdapter(m)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.OnKey(&term.KeyEvent{Rune: 'i', Key: term.KeyUnknown})
		}()
	}
	wg.Wait()
	m2 := a.Model().(*counterModel)
	if m2.count != 20 {
		t.Errorf("expected 20, got %d", m2.count)
	}
}

func TestP167_BatchCmd(t *testing.T) {
	calls := 0
	cmd1 := func() ElmMsg { calls++; return nil }
	cmd2 := func() ElmMsg { calls++; return nil }
	batch := BatchCmd(cmd1, cmd2)
	if batch == nil {
		t.Fatal("expected non-nil batch")
	}
	batch()
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestP167_BatchCmd_Empty(t *testing.T) {
	batch := BatchCmd()
	if batch != nil {
		t.Error("expected nil for empty batch")
	}
}

func itoaP(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
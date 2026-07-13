package event

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestP180_NewElmAdapterWithInit(t *testing.T) {
	// Test the Init() path that NewElmAdapter calls
	called := false
	model := &testElmModel{
		initFn: func() ElmCmd {
			called = true
			return nil
		},
	}
	adapter := NewElmAdapter(model)
	if adapter == nil {
		t.Fatal("expected non-nil adapter")
	}
	if !called {
		t.Error("expected Init to be called")
	}
}

func TestP180_NewElmAdapterWithInitCmd(t *testing.T) {
	var executed atomic.Bool
	model := &testElmModel{
		initFn: func() ElmCmd {
			return func() ElmMsg {
				executed.Store(true)
				return nil
			}
		},
	}
	adapter := NewElmAdapter(model)
	_ = adapter
	time.Sleep(100 * time.Millisecond)
	if !executed.Load() {
		t.Error("expected init cmd to execute")
	}
}

// testElmModel is a minimal ElmModel for testing
type testElmModel struct {
	initFn   func() ElmCmd
	updateFn func(ElmMsg) (ElmModel, ElmCmd)
	viewFn   func() string
}

func (m *testElmModel) Init() ElmCmd {
	if m.initFn != nil {
		return m.initFn()
	}
	return nil
}

func (m *testElmModel) Update(msg ElmMsg) (ElmModel, ElmCmd) {
	if m.updateFn != nil {
		return m.updateFn(msg)
	}
	return m, nil
}

func (m *testElmModel) View() string {
	if m.viewFn != nil {
		return m.viewFn()
	}
	return ""
}

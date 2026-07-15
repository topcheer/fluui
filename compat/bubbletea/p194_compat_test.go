package bubbletea

import (
	"testing"
	"time"
)

func TestNewView(t *testing.T) {
	v := NewView("hello")
	if v.String() != "hello" {
		t.Errorf("expected 'hello', got %q", v.String())
	}
}

func TestRequestWindowSize(t *testing.T) {
	msg := RequestWindowSize()
	ws, ok := msg.(RequestWindowSizeMsg)
	if !ok {
		t.Fatalf("expected RequestWindowSizeMsg, got %T", msg)
	}
	if ws.Width <= 0 || ws.Height <= 0 {
		t.Errorf("expected positive dimensions, got %dx%d", ws.Width, ws.Height)
	}
}

func TestKeyboardEnhancementsMsg(t *testing.T) {
	m := KeyboardEnhancementsMsg{Supported: true}
	if !m.Supported {
		t.Error("expected Supported=true")
	}
}

func TestErrInterrupted(t *testing.T) {
	if ErrInterrupted == nil {
		t.Error("ErrInterrupted should not be nil")
	}
	if ErrInterrupted.Error() != "interrupted" {
		t.Errorf("expected 'interrupted', got %q", ErrInterrupted.Error())
	}
}

func TestMouseModeCellMotion(t *testing.T) {
	if MouseModeCellMotion != 1002 {
		t.Errorf("expected 1002, got %d", MouseModeCellMotion)
	}
}

func TestKeyConstants(t *testing.T) {
	tests := []struct {
		name string
		val  int
	}{
		{"KeyEnter", int(KeyEnter)},
		{"KeyEsc", int(KeyEsc)},
		{"KeyEscape", int(KeyEscape)},
		{"KeyUp", int(KeyUp)},
		{"KeyDown", int(KeyDown)},
		{"KeyTab", int(KeyTab)},
	}
	for _, tt := range tests {
		if tt.val == 0 {
			t.Errorf("%s should not be 0", tt.name)
		}
	}
	if KeyEsc != KeyEscape {
		t.Error("KeyEsc should equal KeyEscape")
	}
}

func TestModConstants(t *testing.T) {
	if ModShift == 0 && ModAlt == 0 && ModCtrl == 0 {
		t.Error("at least some modifiers should be non-zero")
	}
}

func TestDurationAndTime(t *testing.T) {
	d := Duration(5 * time.Second)
	if d != 5*time.Second {
		t.Error("Duration should equal time.Duration")
	}
	now := Time(time.Now())
	if now.IsZero() {
		t.Error("Time should work as time.Time")
	}
}

func TestWithoutSignals(t *testing.T) {
	opt := WithoutSignals()
	if opt == nil {
		t.Error("WithoutSignals should return non-nil option")
	}
}

func TestWithoutRenderer(t *testing.T) {
	opt := WithoutRenderer()
	if opt == nil {
		t.Error("WithoutRenderer should return non-nil option")
	}
}

func TestWithOutput(t *testing.T) {
	opt := WithOutput(nil)
	if opt == nil {
		t.Error("WithOutput should return non-nil option")
	}
}

func TestWithInput(t *testing.T) {
	opt := WithInput(nil)
	if opt == nil {
		t.Error("WithInput should return non-nil option")
	}
}

func TestKeyMsgAlias(t *testing.T) {
	var k KeyMsg = KeyPressMsg{Rune: 'a'}
	if k.Rune != 'a' {
		t.Error("KeyMsg alias should work")
	}
}

func TestBatchMsg(t *testing.T) {
	bm := BatchMsg{Cmds: []Cmd{Nil(), Quit()}}
	if len(bm.Cmds) != 2 {
		t.Errorf("expected 2 cmds, got %d", len(bm.Cmds))
	}
}

func TestLflag(t *testing.T) {
	// Just verify the constant exists
	_ = Lflag
}
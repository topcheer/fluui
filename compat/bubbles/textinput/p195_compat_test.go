package textinput

import (
	"testing"
)

func TestEchoConstants(t *testing.T) {
	if EchoPassword == EchoNormal {
		t.Error("EchoPassword should differ from EchoNormal")
	}
	if EchoNone == EchoNormal {
		t.Error("EchoNone should differ from EchoNormal")
	}
}

func TestCursorModeConstants(t *testing.T) {
	if CursorBlink != 0 {
		t.Error("CursorBlink should be 0")
	}
}

func TestSetEchoMode(t *testing.T) {
	m := New()
	m.SetEchoMode(EchoPassword)
	m.EchoPassword() // method version
}

func TestRunes(t *testing.T) {
	m := New()
	m.SetValue("hello")
	r := m.Runes()
	if len(r) != 5 {
		t.Errorf("expected 5 runes, got %d", len(r))
	}
}

func TestSetCursorMode(t *testing.T) {
	m := New()
	m.SetCursorMode(CursorHide)
}
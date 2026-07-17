package xterm

import (
	"os"
	"testing"
)

func TestIsTerminal_Stdin_P284(t *testing.T) {
	// Stdin in test context may or may not be a terminal — just verify no panic
	result := IsTerminal(os.Stdin.Fd())
	_ = result
}

func TestIsTerminal_InvalidFD_P284(t *testing.T) {
	// Invalid fd should return false, not panic
	result := IsTerminal(99999)
	if result {
		t.Error("invalid fd should not be a terminal")
	}
}

func TestGetSize_Stdin_P284(t *testing.T) {
	w, h, err := GetSize(os.Stdin.Fd())
	// May or may not succeed in test context — verify no panic
	_ = w
	_ = h
	_ = err
}

func TestGetSize_InvalidFD_P284(t *testing.T) {
	_, _, err := GetSize(99999)
	if err == nil {
		t.Error("invalid fd should return error")
	}
}

func TestGetState_InvalidFD_P284(t *testing.T) {
	_, err := GetState(99999)
	if err == nil {
		t.Error("invalid fd should return error")
	}
}

func TestRestore_InvalidFD_P284(t *testing.T) {
	st := &State{}
	err := Restore(99999, st)
	if err == nil {
		t.Error("invalid fd should return error")
	}
}

package xterm

import (
	"os"
	"testing"
)

func TestGetSize_SuccessPath_P286(t *testing.T) {
	fd := os.Stdin.Fd()
	w, h, err := GetSize(fd)
	if err == nil {
		if w <= 0 {
			t.Error("width should be positive on terminal")
		}
		if h <= 0 {
			t.Error("height should be positive on terminal")
		}
	}
	if err != nil {
		if w != 80 || h != 24 {
			t.Errorf("non-terminal should return 80x24, got %dx%d", w, h)
		}
	}
}

func TestGetState_SuccessPath_P286(t *testing.T) {
	fd := os.Stdin.Fd()
	state, err := GetState(fd)
	if err == nil && state == nil {
		t.Error("state should be non-nil on terminal")
	}
}

func TestMakeRaw_InvalidFD_P286(t *testing.T) {
	_, err := MakeRaw(99999)
	if err == nil {
		t.Error("invalid fd should return error")
	}
}

func TestRestore_SuccessPath_P286(t *testing.T) {
	fd := os.Stdin.Fd()
	state, err := GetState(fd)
	if err == nil && state != nil {
		err := Restore(fd, state)
		if err != nil {
			t.Errorf("restore should succeed on terminal: %v", err)
		}
	}
}

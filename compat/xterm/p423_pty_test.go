//go:build darwin || linux

package xterm

import "testing"

// TestMakeRaw_Restore_Pty exercises the full MakeRaw + Restore success path
// using a real pseudo-terminal, covering the bit-manipulation logic that is
// unreachable with a non-terminal fd.
func TestMakeRaw_Restore_Pty(t *testing.T) {
	slave, cleanup := openPtyPair(t)
	defer cleanup()
	fd := slave.Fd()

	state, err := MakeRaw(fd)
	if err != nil {
		t.Fatalf("MakeRaw failed on pty: %v", err)
	}
	if state == nil {
		t.Fatal("state should be non-nil")
	}

	// Restoring the saved state should succeed.
	if err := Restore(fd, state); err != nil {
		t.Errorf("Restore failed on pty: %v", err)
	}
}

// TestGetState_Pty covers the GetState success path on a real terminal.
func TestGetState_Pty(t *testing.T) {
	slave, cleanup := openPtyPair(t)
	defer cleanup()

	state, err := GetState(slave.Fd())
	if err != nil {
		t.Fatalf("GetState failed on pty: %v", err)
	}
	if state == nil {
		t.Fatal("state should be non-nil")
	}
}

// TestGetSize_Pty covers the GetSize success path on a real terminal.
func TestGetSize_Pty(t *testing.T) {
	slave, cleanup := openPtyPair(t)
	defer cleanup()

	// A freshly opened pty may report 0x0; the important coverage target is
	// the no-error success path (line: return int(ws.Col), int(ws.Row), nil).
	w, h, err := GetSize(slave.Fd())
	if err != nil {
		t.Fatalf("GetSize failed on pty: %v", err)
	}
	if w < 0 || h < 0 {
		t.Errorf("dimensions should not be negative, got %dx%d", w, h)
	}
}

// TestIsTerminal_Pty covers the IsTerminal true-branch on a real terminal.
func TestIsTerminal_Pty(t *testing.T) {
	slave, cleanup := openPtyPair(t)
	defer cleanup()

	if !IsTerminal(slave.Fd()) {
		t.Error("slave pty should be reported as a terminal")
	}
}

// TestMakeRaw_GetState_NilRestore exercises Restore with a zero-value State
// on a real terminal (verifies the ioctl path doesn't panic).
func TestRestore_ZeroState_Pty(t *testing.T) {
	slave, cleanup := openPtyPair(t)
	defer cleanup()

	st := &State{}
	_ = Restore(slave.Fd(), st) // may or may not error depending on the termios values; just verify no panic
}

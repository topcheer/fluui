//go:build linux

package xterm

import (
	"fmt"
	"os"
	"testing"
	"unsafe"

	"golang.org/x/sys/unix"
)

// openPtyPair opens a pseudo-terminal pair using /dev/ptmx and returns the
// slave end as an *os.File plus a cleanup function. It skips the test when the
// host cannot allocate a pty (e.g. restricted sandbox).
func openPtyPair(t *testing.T) (*os.File, func()) {
	t.Helper()

	masterFd, err := unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		t.Skipf("cannot open /dev/ptmx: %v", err)
	}
	cleanup := func() { _ = unix.Close(masterFd) }

	// unlockpt
	var unlock int32 = 0
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(masterFd), uintptr(unix.TIOCSPTLCK), uintptr(unsafe.Pointer(&unlock))); errno != 0 {
		cleanup()
		t.Skipf("unlockpt failed: %v", errno)
	}
	// get slave number via TIOCGPTN
	var ptn int32
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(masterFd), uintptr(unix.TIOCGPTN), uintptr(unsafe.Pointer(&ptn))); errno != 0 {
		cleanup()
		t.Skipf("get pty number failed: %v", errno)
	}
	slaveName := fmt.Sprintf("/dev/pts/%d", ptn)

	// Set a default window size so GetSize returns non-zero dimensions.
	ws := &unix.Winsize{Row: 30, Col: 80, Xpixel: 0, Ypixel: 0}
	_ = unix.IoctlSetWinsize(masterFd, unix.TIOCSWINSZ, ws)

	slaveFd, err := unix.Open(slaveName, unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		cleanup()
		t.Skipf("cannot open slave %s: %v", slaveName, err)
	}

	return os.NewFile(uintptr(slaveFd), slaveName), func() {
		_ = unix.Close(slaveFd)
		cleanup()
	}
}

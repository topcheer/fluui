//go:build darwin

package xterm

import (
	"bytes"
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

	// grantpt
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(masterFd), uintptr(unix.TIOCPTYGRANT), 0); errno != 0 {
		cleanup()
		t.Skipf("grantpt failed: %v", errno)
	}
	// unlockpt
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(masterFd), uintptr(unix.TIOCPTYUNLK), 0); errno != 0 {
		cleanup()
		t.Skipf("unlockpt failed: %v", errno)
	}
	// ptsname via TIOCPTYGNAME (fills a 128-byte buffer)
	buf := make([]byte, 128)
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(masterFd), uintptr(unix.TIOCPTYGNAME), uintptr(unsafe.Pointer(&buf[0]))); errno != 0 {
		cleanup()
		t.Skipf("ptsname failed: %v", errno)
	}
	n := bytes.IndexByte(buf, 0)
	if n < 0 {
		n = len(buf)
	}
	slaveName := string(buf[:n])

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

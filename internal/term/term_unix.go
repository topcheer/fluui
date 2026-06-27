//go:build !windows

package term

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Terminal manages the raw terminal state, including alt screen, raw mode,
// mouse tracking, and input/output.
type Terminal struct {
	w        io.Writer
	r        io.Reader
	fd       int
	oldState *unix.Termios
	width    int
	height   int
	profile  ColorProfile
	closed   bool
	mu       sync.Mutex

	// resize notification
	resizeCh   chan struct{}
	sigCh      chan os.Signal
	restoreFns []func()
}

// Open initializes the terminal for TUI rendering:
// - Enters raw mode (disables echo, canonical mode, signals)
// - Switches to alt screen
// - Enables mouse tracking (SGR mode)
// - Enables bracketed paste
// - Hides the cursor
// - Starts resize detection
func Open() (*Terminal, error) {
	// Use /dev/tty for both input and output.
	f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		// Fallback to stdin/stdout if /dev/tty is not available.
		f, err = os.OpenFile(os.Stdout.Name(), os.O_RDWR, 0)
		if err != nil {
			return nil, fmt.Errorf("fluui: cannot open terminal: %w", err)
		}
	}

	t := &Terminal{
		w:        f,
		r:        f,
		fd:       int(f.Fd()),
		profile:  detectColorProfile(),
		resizeCh: make(chan struct{}, 1),
		sigCh:    make(chan os.Signal, 1),
	}

	// Get initial size.
	t.width, t.height = t.getSize()

	// Enter raw mode.
	if err := t.enterRawMode(); err != nil {
		f.Close()
		return nil, err
	}

	// Write init sequences.
	t.WriteRaw(
		"\x1b[?1049h" + // alt screen
			"\x1b[?25l" + // hide cursor
			"\x1b[?2004h" + // bracketed paste
			"\x1b[?1006h" + // SGR mouse mode
			"\x1b[?1002h" + // cell motion mouse tracking
			"\x1b[2J" + // clear screen
			"\x1b[H", // cursor home
	)

	// Start resize detection.
	signal.Notify(t.sigCh, syscall.SIGWINCH)
	go t.watchResize()

	return t, nil
}

// Close restores the terminal to its original state.
func (t *Terminal) Close() error {
	t.mu.Lock()

	if t.closed {
		t.mu.Unlock()
		return nil
	}
	t.closed = true

	// Write cleanup sequences DIRECTLY to the underlying writer.
	_, _ = t.w.Write([]byte(
		"\x1b[?1002l" + // disable mouse tracking
			"\x1b[?1006l" + // disable SGR mouse
			"\x1b[?2004l" + // disable bracketed paste
			"\x1b[?25h" + // show cursor
			"\x1b[?1049l", // leave alt screen
	))

	t.mu.Unlock()

	// Restore terminal state (no lock needed during shutdown).
	for _, fn := range t.restoreFns {
		fn()
	}

	if t.oldState != nil {
		unix.IoctlSetTermios(t.fd, ioctlSetTermios, t.oldState)
	}

	signal.Stop(t.sigCh)

	// Close the tty file.
	if closer, ok := t.w.(io.Closer); ok {
		closer.Close()
	}

	return nil
}

// Write writes raw bytes to the terminal.
func (t *Terminal) Write(b []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.w.Write(b)
}

// WriteRaw writes a string directly to the terminal.
func (t *Terminal) WriteRaw(s string) {
	t.Write([]byte(s))
}

// Read reads raw bytes from the terminal (input).
func (t *Terminal) Read(b []byte) (int, error) {
	return t.r.Read(b)
}

// Size returns the current terminal dimensions (width, height).
func (t *Terminal) Size() (int, int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.width, t.height
}

// ColorProfile returns the detected color profile.
func (t *Terminal) ColorProfile() ColorProfile {
	return t.profile
}

// SupportsMouse reports whether mouse tracking is active.
func (t *Terminal) SupportsMouse() bool {
	return true
}

// ResizeCh returns a channel that receives a signal when the terminal resizes.
func (t *Terminal) ResizeCh() <-chan struct{} {
	return t.resizeCh
}

func (t *Terminal) watchResize() {
	for range t.sigCh {
		t.mu.Lock()
		t.width, t.height = t.getSize()
		t.mu.Unlock()

		select {
		case t.resizeCh <- struct{}{}:
		default:
		}
	}
}

func (t *Terminal) getSize() (int, int) {
	ws := &winsize{}
	ret, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(t.fd),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)
	if int(ret) == -1 || errno != 0 {
		return 80, 24
	}
	return int(ws.Col), int(ws.Row)
}

func (t *Terminal) enterRawMode() error {
	oldState, err := unix.IoctlGetTermios(t.fd, ioctlGetTermios)
	if err != nil {
		return fmt.Errorf("fluui: cannot get terminal attributes: %w", err)
	}
	t.oldState = oldState

	newState := *oldState
	newState.Iflag &^= unix.BRKINT | unix.ICRNL | unix.INPCK | unix.ISTRIP | unix.IXON
	newState.Oflag &^= unix.OPOST
	newState.Cflag |= unix.CS8
	newState.Lflag &^= unix.ECHO | unix.ICANON | unix.IEXTEN | unix.ISIG
	newState.Cc[unix.VMIN] = 1
	newState.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(t.fd, ioctlSetTermios, &newState); err != nil {
		return fmt.Errorf("fluui: cannot set terminal attributes: %w", err)
	}

	return nil
}

func detectColorProfile() ColorProfile {
	termVar := os.Getenv("TERM")
	colorterm := os.Getenv("COLORTERM")

	if colorterm == "truecolor" || colorterm == "24bit" {
		return ProfileTrue
	}
	if contains(termVar, "256color") {
		return Profile256
	}
	if contains(termVar, "color") || contains(termVar, "ansi") {
		return ProfileANSI16
	}
	return ProfileANSI16
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// winsize matches the C struct winsize.
type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

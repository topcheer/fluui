// Package xterm provides a drop-in compatibility layer for projects
// migrating from github.com/charmbracelet/x/term to fluui.
//
// It mirrors the charmbracelet/x/term API for the functions used by ggcode:
//
//	-  "github.com/charmbracelet/x/term"
//	+  xterm "github.com/topcheer/fluui/compat/xterm"
//
// Usage:
//
//	w, h, err := xterm.GetSize(os.Stdin.Fd())
//	isTTY := xterm.IsTerminal(os.Stdout.Fd())
package xterm

import "golang.org/x/sys/unix"

// State represents the terminal state, compatible with term.State from charmbracelet/x/term.
type State struct {
	termios unix.Termios
}

// GetSize returns the dimensions of the terminal connected to fd.
// Returns (width, height, error) matching charmbracelet/x/term.GetSize.
func GetSize(fd uintptr) (int, int, error) {
	ws, err := unix.IoctlGetWinsize(int(fd), unix.TIOCGWINSZ)
	if err != nil {
		return 80, 24, err
	}
	return int(ws.Col), int(ws.Row), nil
}

// IsTerminal returns true if the given file descriptor is a terminal.
func IsTerminal(fd uintptr) bool {
	_, err := unix.IoctlGetTermios(int(fd), ioctlReadTermios)
	return err == nil
}

// MakeRaw puts the terminal connected to fd into raw mode and returns the
// previous state so it can be restored.
func MakeRaw(fd uintptr) (*State, error) {
	oldState, err := unix.IoctlGetTermios(int(fd), ioctlReadTermios)
	if err != nil {
		return nil, err
	}

	newState := *oldState
	newState.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP |
		unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	newState.Oflag &^= unix.OPOST
	newState.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	newState.Cflag &^= unix.CSIZE | unix.PARENB
	newState.Cflag |= unix.CS8
	newState.Cc[unix.VMIN] = 1
	newState.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(int(fd), ioctlWriteTermios, &newState); err != nil {
		return nil, err
	}

	return &State{termios: *oldState}, nil
}

// Restore restores the terminal connected to fd to a previous state.
func Restore(fd uintptr, state *State) error {
	return unix.IoctlSetTermios(int(fd), ioctlWriteTermios, &state.termios)
}

// GetState returns the current terminal state for fd.
func GetState(fd uintptr) (*State, error) {
	termios, err := unix.IoctlGetTermios(int(fd), ioctlReadTermios)
	if err != nil {
		return nil, err
	}
	return &State{termios: *termios}, nil
}

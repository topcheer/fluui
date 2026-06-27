//go:build windows

package term

import (
	"errors"
	"io"
)

// Terminal is a stub on Windows. Full Windows support is not yet implemented.
type Terminal struct {
	w       io.Writer
	r       io.Reader
	width   int
	height  int
	profile ColorProfile
	closed  bool
}

var errWindows = errors.New("fluui: Windows is not yet supported")

// Open returns an error on Windows.
func Open() (*Terminal, error) {
	return nil, errWindows
}

// Close returns an error on Windows.
func (t *Terminal) Close() error {
	return errWindows
}

func (t *Terminal) Write(b []byte) (int, error)       { return 0, errWindows }
func (t *Terminal) WriteRaw(s string)                  {}
func (t *Terminal) Read(b []byte) (int, error)        { return 0, errWindows }
func (t *Terminal) Size() (int, int)                   { return 80, 24 }
func (t *Terminal) ColorProfile() ColorProfile         { return ProfileANSI16 }
func (t *Terminal) SupportsMouse() bool                { return false }
func (t *Terminal) ResizeCh() <-chan struct{}         { return nil }

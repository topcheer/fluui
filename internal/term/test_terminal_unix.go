//go:build !windows

package term

import (
	"io"
	"os"
)

// NewTestTerminal creates a Terminal for testing purposes.
// It does NOT enter raw mode or start resize detection.
// The reader provides input bytes; the writer receives output.
// This enables testing event.Loop.Run and event.Loop.readRaw
// without /dev/tty.
func NewTestTerminal(r io.Reader, w io.Writer, width, height int) *Terminal {
	return &Terminal{
		r:        r,
		w:        w,
		fd:       -1,
		width:    width,
		height:   height,
		profile:  Profile256,
		resizeCh: make(chan struct{}, 1),
		sigCh:    make(chan os.Signal, 1),
	}
}

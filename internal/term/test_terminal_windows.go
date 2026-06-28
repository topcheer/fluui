//go:build windows

package term

import (
	"io"
)

// NewTestTerminal creates a Terminal for testing with the given reader/writer.
// On Windows, Terminal is a stub — this is mainly for compilation.
func NewTestTerminal(r io.Reader, w io.Writer, width, height int) *Terminal {
	return &Terminal{
		r:       r,
		w:       w,
		width:   width,
		height:  height,
		profile: Profile256,
	}
}

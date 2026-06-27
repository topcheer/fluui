package term

// ColorProfile describes the color capabilities of the terminal.
type ColorProfile int

const (
	ProfileNone     ColorProfile = iota // no color support
	ProfileANSI16                       // 16 named colors
	Profile256                          // xterm 256 colors
	ProfileTrue                         // 24-bit true color
)

package buffer

import (
	"fmt"
	"strconv"
	"strings"
)

// ColorType determines how a Color value is interpreted.
type ColorType uint8

const (
	// ColorNone means "use the terminal's default color".
	ColorNone ColorType = iota
	// ColorNamed is one of the 16 standard ANSI named colors.
	ColorNamed
	// Color256 is an xterm 256-color palette entry (0-255).
	Color256
	// ColorTrue is a 24-bit TrueColor (0xRRGGBB).
	ColorTrue
)

// Named colors matching the standard 16-color ANSI palette.
const (
	NamedBlack   = 0
	NamedRed     = 1
	NamedGreen   = 2
	NamedYellow  = 3
	NamedBlue    = 4
	NamedMagenta = 5
	NamedCyan    = 6
	NamedWhite   = 7
	// Bright variants (8-15)
	NamedBrightBlack   = 8
	NamedBrightRed     = 9
	NamedBrightGreen   = 10
	NamedBrightYellow  = 11
	NamedBrightBlue    = 12
	NamedBrightMagenta = 13
	NamedBrightCyan    = 14
	NamedBrightWhite   = 15
)

// Color represents a terminal color.
type Color struct {
	Type ColorType
	Val  uint32
}

// NoColor returns a Color that uses the terminal default.
func NoColor() Color { return Color{Type: ColorNone} }

// NamedColor returns a 16-color palette Color.
func NamedColor(n int) Color {
	return Color{Type: ColorNamed, Val: uint32(n)}
}

// Color256 returns a 256-color palette Color.
func Color256Val(n uint8) Color {
	return Color{Type: Color256, Val: uint32(n)}
}

// RGB returns a TrueColor Color.
func RGB(r, g, b uint8) Color {
	return Color{Type: ColorTrue, Val: uint32(r)<<16 | uint32(g)<<8 | uint32(b)}
}

// Hex parses a hex string like "#ff6600" into a TrueColor.
func Hex(hex string) Color {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return NoColor()
	}
	n, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return NoColor()
	}
	return Color{Type: ColorTrue, Val: uint32(n)}
}

// Predefined convenience colors.
var (
	Default    = NoColor()
	Black      = NamedColor(NamedBlack)
	Red        = NamedColor(NamedRed)
	Green      = NamedColor(NamedGreen)
	Yellow     = NamedColor(NamedYellow)
	Blue       = NamedColor(NamedBlue)
	Magenta    = NamedColor(NamedMagenta)
	Cyan       = NamedColor(NamedCyan)
	White      = NamedColor(NamedWhite)
	BrightRed  = NamedColor(NamedBrightRed)
	BrightGreen = NamedColor(NamedBrightGreen)
)

// Equal reports whether two colors are identical.
func (c Color) Equal(o Color) bool {
	return c.Type == o.Type && c.Val == o.Val
}

// IsDefault reports whether the color is unset (terminal default).
func (c Color) IsDefault() bool {
	return c.Type == ColorNone
}

// String returns a human-readable description.
func (c Color) String() string {
	switch c.Type {
	case ColorNone:
		return "default"
	case ColorNamed:
		return fmt.Sprintf("named(%d)", c.Val)
	case Color256:
		return fmt.Sprintf("256(%d)", c.Val)
	case ColorTrue:
		return fmt.Sprintf("#%06x", c.Val)
	}
	return "?"
}

// R, G, B extract the components of a TrueColor. Returns 0 for other types.
func (c Color) R() uint8 { return uint8(c.Val >> 16) }
func (c Color) G() uint8 { return uint8(c.Val >> 8) }
func (c Color) B() uint8 { return uint8(c.Val) }

// ANSI returns the SGR parameter(s) for this color as a foreground.
func (c Color) FGSequence() string {
	switch c.Type {
	case ColorNone:
		return "39"
	case ColorNamed:
		if c.Val >= 8 {
			return fmt.Sprintf("%d", 90+c.Val-8)
		}
		return fmt.Sprintf("%d", 30+c.Val)
	case Color256:
		return fmt.Sprintf("38;5;%d", c.Val)
	case ColorTrue:
		return fmt.Sprintf("38;2;%d;%d;%d", c.R(), c.G(), c.B())
	}
	return "39"
}

// BGSequence returns the SGR parameter(s) for this color as a background.
func (c Color) BGSequence() string {
	switch c.Type {
	case ColorNone:
		return "49"
	case ColorNamed:
		if c.Val >= 8 {
			return fmt.Sprintf("%d", 100+c.Val-8)
		}
		return fmt.Sprintf("%d", 40+c.Val)
	case Color256:
		return fmt.Sprintf("48;5;%d", c.Val)
	case ColorTrue:
		return fmt.Sprintf("48;2;%d;%d;%d", c.R(), c.G(), c.B())
	}
	return "49"
}

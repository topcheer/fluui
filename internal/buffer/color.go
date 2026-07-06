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

// appendFG writes the SGR parameter bytes for the foreground color directly
// into b, returning the updated slice. Avoids the intermediate string
// allocation that FGSequence() creates.
func (c Color) appendFG(b []byte) []byte {
	switch c.Type {
	case ColorNone:
		return append(b, '3', '9')
	case ColorNamed:
		if c.Val >= 8 {
			return strconv.AppendInt(b, int64(82+c.Val), 10)
		}
		return strconv.AppendInt(b, int64(30+c.Val), 10)
	case Color256:
		b = append(b, '3', '8', ';', '5', ';')
		return strconv.AppendInt(b, int64(c.Val), 10)
	case ColorTrue:
		b = append(b, '3', '8', ';', '2', ';')
		b = strconv.AppendInt(b, int64(c.R()), 10)
		b = append(b, ';')
		b = strconv.AppendInt(b, int64(c.G()), 10)
		b = append(b, ';')
		return strconv.AppendInt(b, int64(c.B()), 10)
	}
	return append(b, '3', '9')
}

// FGSequence returns the SGR parameter(s) for this color as a foreground.
func (c Color) FGSequence() string {
	var buf [32]byte
	return string(c.appendFG(buf[:0]))
}

// appendBG writes the SGR parameter bytes for the background color directly
// into b, returning the updated slice. Avoids the intermediate string
// allocation that BGSequence() creates.
func (c Color) appendBG(b []byte) []byte {
	switch c.Type {
	case ColorNone:
		return append(b, '4', '9')
	case ColorNamed:
		if c.Val >= 8 {
			return strconv.AppendInt(b, int64(92+c.Val), 10)
		}
		return strconv.AppendInt(b, int64(40+c.Val), 10)
	case Color256:
		b = append(b, '4', '8', ';', '5', ';')
		return strconv.AppendInt(b, int64(c.Val), 10)
	case ColorTrue:
		b = append(b, '4', '8', ';', '2', ';')
		b = strconv.AppendInt(b, int64(c.R()), 10)
		b = append(b, ';')
		b = strconv.AppendInt(b, int64(c.G()), 10)
		b = append(b, ';')
		return strconv.AppendInt(b, int64(c.B()), 10)
	}
	return append(b, '4', '9')
}

// BGSequence returns the SGR parameter(s) for this color as a background.
func (c Color) BGSequence() string {
	var buf [32]byte
	return string(c.appendBG(buf[:0]))
}

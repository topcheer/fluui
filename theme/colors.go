package theme

import "github.com/topcheer/fluui/internal/buffer"

// C is a shorthand constructor for TrueColor colors.
func C(r, g, b uint8) Color {
	return buffer.RGB(r, g, b)
}

// Hex parses a hex color string like "#ff6600".
func Hex(s string) Color {
	return buffer.Hex(s)
}

// NoColor returns a color representing the terminal default.
func NoColor() Color {
	return buffer.NoColor()
}

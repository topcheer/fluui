package component

import (
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// ─── AdaptiveColor: lipgloss/compat.AdaptiveColor equivalent ───
//
// AdaptiveColor provides light/dark mode-aware colors, matching
// lipgloss v2's compat.AdaptiveColor API for ggcode migration.
//
// Usage:
//
//	color := AdaptiveColor{Light: "#d0d0d0", Dark: "#a0a0a0"}
//	style := NewStyle().Foreground(color.Resolve())

// AdaptiveColor holds light and dark mode color values.
type AdaptiveColor struct {
	Light string
	Dark  string
}

// Resolve returns the appropriate buffer.Color based on the current theme.
// If the current theme is dark, returns Dark; otherwise returns Light.
func (a AdaptiveColor) Resolve() buffer.Color {
	th := theme.Active
	if th != nil && isDarkBackground(th) {
		return parseColorString(a.Dark)
	}
	return parseColorString(a.Light)
}

// ResolveToStyle returns a buffer.Style with Fg set to the resolved color.
func (a AdaptiveColor) ResolveToStyle() buffer.Style {
	return buffer.Style{Fg: a.Resolve()}
}

// ResolveToBgStyle returns a buffer.Style with Bg set to the resolved color.
func (a AdaptiveColor) ResolveToBgStyle() buffer.Style {
	return buffer.Style{Bg: a.Resolve()}
}

// String returns a human-readable representation.
func (a AdaptiveColor) String() string {
	return "AdaptiveColor{Light:" + a.Light + ", Dark:" + a.Dark + "}"
}

// parseColorString parses a color string (hex, named, or numeric) into buffer.Color.
func parseColorString(s string) buffer.Color {
	if s == "" {
		return buffer.Color{}
	}
	// Hex color
	if len(s) > 0 && s[0] == '#' {
		c, ok := parseHexColor(s)
		if ok {
			return c
		}
	}
	// Try named color
	switch s {
	case "red", "1":
		return buffer.NamedColor(buffer.NamedRed)
	case "green", "2":
		return buffer.NamedColor(buffer.NamedGreen)
	case "yellow", "3":
		return buffer.NamedColor(buffer.NamedYellow)
	case "blue", "4":
		return buffer.NamedColor(buffer.NamedBlue)
	case "magenta", "5":
		return buffer.NamedColor(buffer.NamedMagenta)
	case "cyan", "6":
		return buffer.NamedColor(buffer.NamedCyan)
	case "white", "7":
		return buffer.NamedColor(buffer.NamedWhite)
	case "black", "0":
		return buffer.NamedColor(buffer.NamedBlack)
	}
	// Try numeric ANSI 256
	n := parseUint(s)
	if n >= 0 && n <= 255 {
		return buffer.Color256Val(uint8(n))
	}
	return buffer.Color{}
}
// isDarkBackground returns true if the theme's background color is dark.
func isDarkBackground(th *theme.Theme) bool {
	if th == nil || th.Bg.Type == 0 {
		return true // default to dark
	}
	// If RGB color, compute luminance
	if th.Bg.Type == 3 { // ColorTrue (RGB)
		r := int((th.Bg.Val >> 16) & 0xFF)
		g := int((th.Bg.Val >> 8) & 0xFF)
		b := int(th.Bg.Val & 0xFF)
		luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
		return luminance < 128
	}
	// For named/256 colors, assume dark (most TUI themes are dark)
	return true
}

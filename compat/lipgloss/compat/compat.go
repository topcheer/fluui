// Package compat provides adaptive color support for projects migrating
// from charm.land/lipgloss/v2/compat to fluui.
//
// AdaptiveColor picks between a light and dark variant based on the terminal's
// reported color scheme. Since fluui always detects dark terminals by default
// (the common case for developer workstations), the Dark variant is returned
// unless explicitly overridden.
//
// Usage (drop-in replacement):
//
//	-  "charm.land/lipgloss/v2/compat"
//	+  "github.com/topcheer/fluui/compat/lipgloss/compat"
//
//	compat.AdaptiveColor{Light: lipgloss.Color("240"), Dark: lipgloss.Color("250")}
package compat

import "github.com/topcheer/fluui/compat/lipgloss"

// AdaptiveColor holds a light and dark variant of a color.
// The Dark variant is used when the terminal reports a dark background
// (the default for most developer terminals).
type AdaptiveColor struct {
	Light lipgloss.Color
	Dark  lipgloss.Color
}

// colorScheme tracks the current terminal scheme. Defaults to dark.
var colorScheme = "dark"

// SetColorScheme sets the global terminal color scheme.
// Valid values: "light" or "dark". Any other value defaults to "dark".
func SetColorScheme(scheme string) {
	if scheme == "light" {
		colorScheme = "light"
	} else {
		colorScheme = "dark"
	}
}

// ColorScheme returns the current terminal color scheme ("light" or "dark").
func ColorScheme() string {
	return colorScheme
}

// Resolve returns the appropriate color variant for the current terminal scheme.
// If the terminal scheme is "light", returns the Light color; otherwise Dark.
func (ac AdaptiveColor) Resolve() lipgloss.Color {
	if colorScheme == "light" {
		return ac.Light
	}
	return ac.Dark
}

// Background returns the background-appropriate color (same as Resolve for now).
func (ac AdaptiveColor) Background() lipgloss.Color {
	return ac.Resolve()
}

// Foreground returns the foreground-appropriate color (same as Resolve for now).
func (ac AdaptiveColor) Foreground() lipgloss.Color {
	return ac.Resolve()
}

// String returns the string value of the resolved color.
func (ac AdaptiveColor) String() string {
	return string(ac.Resolve())
}

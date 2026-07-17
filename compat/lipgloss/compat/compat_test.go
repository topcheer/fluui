package compat

import (
	"testing"

	"github.com/topcheer/fluui/compat/lipgloss"
)

func TestAdaptiveColor_DefaultDark_P281(t *testing.T) {
	ac := AdaptiveColor{Light: lipgloss.Color("240"), Dark: lipgloss.Color("250")}
	if got := ac.Resolve(); got != lipgloss.Color("250") {
		t.Errorf("default dark scheme should return Dark color, got %q", got)
	}
}

func TestAdaptiveColor_LightScheme_P281(t *testing.T) {
	SetColorScheme("light")
	defer SetColorScheme("dark")
	ac := AdaptiveColor{Light: lipgloss.Color("240"), Dark: lipgloss.Color("248")}
	if got := ac.Resolve(); got != lipgloss.Color("240") {
		t.Errorf("light scheme should return Light color, got %q", got)
	}
}

func TestAdaptiveColor_InvalidScheme_P281(t *testing.T) {
	SetColorScheme("purple")
	defer SetColorScheme("dark")
	if ColorScheme() != "dark" {
		t.Error("invalid scheme should default to dark")
	}
}

func TestAdaptiveColor_String_P281(t *testing.T) {
	ac := AdaptiveColor{Light: lipgloss.Color("240"), Dark: lipgloss.Color("250")}
	if ac.String() != "250" {
		t.Errorf("expected '250', got %q", ac.String())
	}
}

func TestAdaptiveColor_BackgroundForeground_P281(t *testing.T) {
	ac := AdaptiveColor{Light: lipgloss.Color("240"), Dark: lipgloss.Color("250")}
	if ac.Background() != lipgloss.Color("250") {
		t.Error("Background should equal Resolve in dark mode")
	}
	if ac.Foreground() != lipgloss.Color("250") {
		t.Error("Foreground should equal Resolve in dark mode")
	}
}

func TestSetColorScheme_Roundtrip_P281(t *testing.T) {
	SetColorScheme("light")
	if ColorScheme() != "light" {
		t.Error("scheme should be light")
	}
	SetColorScheme("dark")
	if ColorScheme() != "dark" {
		t.Error("scheme should be dark")
	}
}

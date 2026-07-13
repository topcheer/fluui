package component

import (
	"testing"

	"github.com/topcheer/fluui/theme"
)

func TestP165_AdaptiveColor_Dark(t *testing.T) {
	theme.Active = theme.Dracula() // dark theme
	ac := AdaptiveColor{Light: "#d0d0d0", Dark: "#a0a0a0"}
	c := ac.Resolve()
	if c.Type == 0 {
		t.Error("expected non-zero color")
	}
}

func TestP165_AdaptiveColor_Light(t *testing.T) {
	// Force light theme by using a bright bg
	th := &theme.Theme{Name: "light", Bg: theme.Color{Type: 3, Val: 0xFFFFFF}}
	theme.Active = th
	defer func() { theme.Active = theme.Dracula() }()
	ac := AdaptiveColor{Light: "#d0d0d0", Dark: "#a0a0a0"}
	c := ac.Resolve()
	if c.Type == 0 {
		t.Error("expected non-zero color")
	}
}

func TestP165_AdaptiveColor_Empty(t *testing.T) {
	ac := AdaptiveColor{}
	c := ac.Resolve()
	if c.Type != 0 {
		t.Error("expected zero color for empty")
	}
}

func TestP165_AdaptiveColor_Named(t *testing.T) {
	theme.Active = theme.Dracula()
	ac := AdaptiveColor{Light: "red", Dark: "blue"}
	c := ac.Resolve()
	if c.Type == 0 {
		t.Error("expected non-zero color")
	}
}

func TestP165_AdaptiveColor_Numeric(t *testing.T) {
	theme.Active = theme.Dracula()
	ac := AdaptiveColor{Light: "240", Dark: "250"}
	c := ac.Resolve()
	if c.Type == 0 {
		t.Error("expected non-zero color")
	}
}

func TestP165_AdaptiveColor_ResolveToStyle(t *testing.T) {
	theme.Active = theme.Dracula()
	ac := AdaptiveColor{Light: "#ff0000", Dark: "#00ff00"}
	s := ac.ResolveToStyle()
	if s.Fg.Type == 0 {
		t.Error("expected non-zero fg")
	}
}

func TestP165_AdaptiveColor_ResolveToBgStyle(t *testing.T) {
	theme.Active = theme.Dracula()
	ac := AdaptiveColor{Light: "#ff0000", Dark: "#00ff00"}
	s := ac.ResolveToBgStyle()
	if s.Bg.Type == 0 {
		t.Error("expected non-zero bg")
	}
}

func TestP165_AdaptiveColor_String(t *testing.T) {
	ac := AdaptiveColor{Light: "red", Dark: "blue"}
	s := ac.String()
	if s == "" {
		t.Error("expected non-empty string")
	}
}

func TestP165_isDarkBackground(t *testing.T) {
	// Dark theme
	if !isDarkBackground(theme.Dracula()) {
		t.Error("expected Dracula to be dark")
	}
	// Light theme (white bg)
	light := &theme.Theme{Name: "light", Bg: theme.Color{Type: 3, Val: 0xFFFFFF}}
	if isDarkBackground(light) {
		t.Error("expected white bg to be light")
	}
	// Nil
	if !isDarkBackground(nil) {
		t.Error("expected nil to default dark")
	}
}
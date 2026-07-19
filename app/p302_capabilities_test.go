package app

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

func TestP302_DefaultCapabilities(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	shell := NewAppShell(root)
	caps := shell.Capabilities()
	// Should have default capabilities
	if !caps.SGRMouse {
		t.Error("default caps should have SGRMouse")
	}
	if !caps.BracketedPaste {
		t.Error("default caps should have BracketedPaste")
	}
}

func TestP302_DetectCapabilities_Kitty(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	shell := NewAppShell(root)
	shell.DetectCapabilities("xterm-kitty", "kitty", "truecolor")
	caps := shell.Capabilities()
	if !caps.TrueColor {
		t.Error("kitty should have TrueColor")
	}
	if !caps.KittyGraphics {
		t.Error("kitty should have KittyGraphics")
	}
	if !caps.OSC8 {
		t.Error("kitty should have OSC8")
	}
}

func TestP302_DetectCapabilities_WezTerm(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	shell := NewAppShell(root)
	shell.DetectCapabilities("wezterm", "WezTerm", "truecolor")
	caps := shell.Capabilities()
	if !caps.Sixel {
		t.Error("wezterm should have Sixel")
	}
}

func TestP302_DetectCapabilities_Default(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	shell := NewAppShell(root)
	shell.DetectCapabilities("xterm-256color", "", "")
	caps := shell.Capabilities()
	if !caps.Color256 {
		t.Error("should have Color256 from TERM=xterm-256color")
	}
	// Without TERM_PROGRAM, advanced features should not be auto-detected
	if caps.TrueColor {
		t.Error("should NOT auto-detect TrueColor without COLORTERM")
	}
}

func TestP302_SetCapabilities(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	shell := NewAppShell(root)
	custom := term.TerminalCapabilities{
		TrueColor:  true,
		OSC8:       true,
		OSC133:     true,
	}
	shell.SetCapabilities(custom)
	caps := shell.Capabilities()
	if !caps.TrueColor {
		t.Error("should have TrueColor after manual set")
	}
	if !caps.OSC8 {
		t.Error("should have OSC8 after manual set")
	}
}

func TestP302_CapabilitiesThreadSafe(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	shell := NewAppShell(root)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			shell.DetectCapabilities("xterm-256color", "", "")
		}
		close(done)
	}()
	for i := 0; i < 100; i++ {
		_ = shell.Capabilities()
	}
	<-done
}

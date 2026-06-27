package termcompat

import (
	"testing"
)

func TestIdentifyTerminal_ITerm(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM_PROGRAM": "iTerm.app",
		"TERM":         "xterm-256color",
	})
	if caps.Name != "iTerm.app" {
		t.Errorf("Name = %q, want %q", caps.Name, "iTerm.app")
	}
	assertTrue(t, "HasOSC52", caps.HasOSC52)
	assertTrue(t, "HasTrueColor", caps.HasTrueColor)
	assertTrue(t, "HasBracketedPaste", caps.HasBracketedPaste)
	assertTrue(t, "HasSGRMouse", caps.HasSGRMouse)
	assertTrue(t, "Has256Color", caps.Has256Color)
}

func TestIdentifyTerminal_WezTerm(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM_PROGRAM": "WezTerm",
		"TERM":         "wezterm",
	})
	if caps.Name != "WezTerm" {
		t.Errorf("Name = %q, want %q", caps.Name, "WezTerm")
	}
	assertTrue(t, "HasOSC52", caps.HasOSC52)
	assertTrue(t, "HasTrueColor", caps.HasTrueColor)
	assertTrue(t, "Has256Color", caps.Has256Color)
	assertTrue(t, "HasBracketedPaste", caps.HasBracketedPaste)
	assertTrue(t, "HasSGRMouse", caps.HasSGRMouse)
}

func TestIdentifyTerminal_Kitty(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM_PROGRAM": "kitty",
		"TERM":         "xterm-kitty",
	})
	if caps.Name != "kitty" {
		t.Errorf("Name = %q, want %q", caps.Name, "kitty")
	}
	assertTrue(t, "HasOSC52", caps.HasOSC52)
	assertTrue(t, "HasTrueColor", caps.HasTrueColor)
}

func TestIdentifyTerminal_Alacritty(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM_PROGRAM": "Alacritty",
		"TERM":         "alacritty",
	})
	if caps.Name != "Alacritty" {
		t.Errorf("Name = %q, want %q", caps.Name, "Alacritty")
	}
	assertTrue(t, "HasOSC52", caps.HasOSC52)
	assertTrue(t, "HasTrueColor", caps.HasTrueColor)
}

func TestIdentifyTerminal_Ghostty(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM_PROGRAM": "ghostty",
		"TERM":         "xterm-ghostty",
	})
	if caps.Name != "Ghostty" {
		t.Errorf("Name = %q, want %q", caps.Name, "Ghostty")
	}
	assertTrue(t, "HasOSC52", caps.HasOSC52)
	assertTrue(t, "HasTrueColor", caps.HasTrueColor)
}

func TestIdentifyTerminal_AppleTerminal(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM_PROGRAM": "Apple_Terminal",
		"TERM":         "xterm-256color",
	})
	if caps.Name != "Apple Terminal" {
		t.Errorf("Name = %q, want %q", caps.Name, "Apple Terminal")
	}
	assertTrue(t, "HasOSC52", caps.HasOSC52)
	assertTrue(t, "HasTrueColor", caps.HasTrueColor)
}

func TestIdentifyTerminal_VSCode(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM_PROGRAM": "vscode",
		"TERM":         "xterm-256color",
	})
	if caps.Name != "VSCode" {
		t.Errorf("Name = %q, want %q", caps.Name, "VSCode")
	}
	assertTrue(t, "HasOSC52", caps.HasOSC52)
}

func TestIdentifyTerminal_WindowsTerminal(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"WT_SESSION": "abc123",
		"TERM":       "xterm-256color",
	})
	if caps.Name != "Windows Terminal" {
		t.Errorf("Name = %q, want %q", caps.Name, "Windows Terminal")
	}
	assertTrue(t, "HasOSC52", caps.HasOSC52)
}

func TestIdentifyTerminal_GNOMETerminal(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM": "gnome-256color",
	})
	assertFalse(t, "HasOSC52 for GNOME", caps.HasOSC52)
	assertTrue(t, "HasTrueColor", caps.HasTrueColor)
	assertTrue(t, "Has256Color", caps.Has256Color)
}

func TestIdentifyTerminal_Hyper(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM_PROGRAM": "Hyper",
		"TERM":         "xterm-256color",
	})
	assertFalse(t, "HasOSC52 for Hyper", caps.HasOSC52)
	assertTrue(t, "HasTrueColor", caps.HasTrueColor)
}

func TestIdentifyTerminal_Tmux(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM_PROGRAM": "tmux",
		"TERM":         "tmux-256color",
		"TMUX":         "/tmp/tmux-1000/default,1234,0",
	})
	assertTrue(t, "InsideTmux", caps.InsideTmux)
	assertTrue(t, "HasBracketedPaste in tmux", caps.HasBracketedPaste)
	assertTrue(t, "HasSGRMouse in tmux", caps.HasSGRMouse)
}

func TestIdentifyTerminal_Screen(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM": "screen-256color",
		"STY":  "12345.pts-0.host",
	})
	assertTrue(t, "InsideScreen", caps.InsideScreen)
	assertFalse(t, "HasOSC52 in screen", caps.HasOSC52)
	assertFalse(t, "HasTrueColor in screen", caps.HasTrueColor)
	assertTrue(t, "Has256Color in screen", caps.Has256Color)
}

func TestIdentifyTerminal_TmuxPassthrough(t *testing.T) {
	// Running iTerm inside tmux
	caps := DetectFromEnv(mockEnv{
		"TERM_PROGRAM": "iTerm.app",
		"TERM":         "tmux-256color",
		"TMUX":         "/tmp/tmux-1000/default,1234,0",
	})
	assertTrue(t, "InsideTmux", caps.InsideTmux)
	assertTrue(t, "HasOSC52 via iTerm", caps.HasOSC52)
	assertTrue(t, "HasTrueColor passthrough", caps.HasTrueColor)
}

func TestIdentifyTerminal_Unknown(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM": "xterm-256color",
	})
	assertFalse(t, "HasOSC52 for unknown", caps.HasOSC52)
	assertTrue(t, "Has256Color from TERM", caps.Has256Color)
	assertTrue(t, "HasBracketedPaste default", caps.HasBracketedPaste)
	assertTrue(t, "HasSGRMouse default", caps.HasSGRMouse)
}

func TestIdentifyTerminal_Dumb(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM": "dumb",
	})
	assertFalse(t, "HasBracketedPaste for dumb", caps.HasBracketedPaste)
	assertFalse(t, "HasSGRMouse for dumb", caps.HasSGRMouse)
	assertFalse(t, "HasOSC52 for dumb", caps.HasOSC52)
	assertFalse(t, "Has256Color for dumb", caps.Has256Color)
}

func TestIdentifyTerminal_EmptyEnv(t *testing.T) {
	caps := DetectFromEnv(mockEnv{})
	assertFalse(t, "HasOSC52", caps.HasOSC52)
	assertFalse(t, "HasTrueColor", caps.HasTrueColor)
	assertFalse(t, "HasBracketedPaste", caps.HasBracketedPaste)
	assertFalse(t, "HasSGRMouse", caps.HasSGRMouse)
}

func TestCapabilities_TrueColorFromColorTerm(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM":      "xterm-256color",
		"COLORTERM": "truecolor",
	})
	assertTrue(t, "HasTrueColor from COLORTERM", caps.HasTrueColor)
}

func TestCapabilities_TrueColorFrom24bit(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM":      "xterm-256color",
		"COLORTERM": "24bit",
	})
	assertTrue(t, "HasTrueColor from 24bit", caps.HasTrueColor)
}

func TestCapabilities_256ColorFromTerm(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM": "myterm-256color",
	})
	// Unknown terminal: should detect 256 from TERM but not true color
	assertFalse(t, "HasTrueColor for unknown 256", caps.HasTrueColor)
	assertTrue(t, "Has256Color from TERM", caps.Has256Color)
}

func TestString_KnownTerminal(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM_PROGRAM": "iTerm.app",
		"TERM":         "xterm-256color",
		"COLORTERM":    "truecolor",
	})
	s := caps.String()
	if !contains(s, "iTerm") {
		t.Errorf("String() should contain terminal name, got %q", s)
	}
	if !contains(s, "truecolor") {
		t.Errorf("String() should contain truecolor, got %q", s)
	}
	if !contains(s, "osc52") {
		t.Errorf("String() should contain osc52, got %q", s)
	}
}

func TestString_TmuxIndicator(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM_PROGRAM": "iTerm.app",
		"TERM":         "tmux-256color",
		"TMUX":         "/tmp/tmux-1000/default,1234,0",
	})
	s := caps.String()
	if !contains(s, "(tmux)") {
		t.Errorf("String() should contain (tmux), got %q", s)
	}
}

func TestString_ScreenIndicator(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM": "screen-256color",
		"STY":  "12345",
	})
	s := caps.String()
	if !contains(s, "(screen)") {
		t.Errorf("String() should contain (screen), got %q", s)
	}
}

func TestString_Dumb16Color(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM": "dumb",
	})
	s := caps.String()
	if !contains(s, "16color") {
		t.Errorf("String() for dumb should contain 16color, got %q", s)
	}
}

func TestColorDepth(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM_PROGRAM": "iTerm.app",
		"TERM":         "xterm-256color",
		"COLORTERM":    "truecolor",
	})
	if caps.ColorDepth() != 24 {
		t.Errorf("ColorDepth() = %d, want 24", caps.ColorDepth())
	}
}

func TestShouldUseOSC52(t *testing.T) {
	// iTerm: yes
	caps := DetectFromEnv(mockEnv{
		"TERM_PROGRAM": "iTerm.app",
		"TERM":         "xterm-256color",
	})
	if !caps.ShouldUseOSC52() {
		t.Error("iTerm should use OSC52")
	}

	// GNOME Terminal: no
	caps = DetectFromEnv(mockEnv{
		"TERM": "gnome-256color",
	})
	if caps.ShouldUseOSC52() {
		t.Error("GNOME Terminal should not use OSC52")
	}

	// GNU screen: no
	caps = DetectFromEnv(mockEnv{
		"TERM": "screen-256color",
		"STY":  "12345",
	})
	if caps.ShouldUseOSC52() {
		t.Error("GNU screen should not use OSC52")
	}
}

// --- Helpers ---

type mockEnv map[string]string

func (m mockEnv) Get(key string) string { return m[key] }

func assertTrue(t *testing.T, name string, v bool) {
	t.Helper()
	if !v {
		t.Errorf("%s: expected true, got false", name)
	}
}

func assertFalse(t *testing.T, name string, v bool) {
	t.Helper()
	if v {
		t.Errorf("%s: expected false, got true", name)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

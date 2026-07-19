package term

import "testing"

func TestP297_DefaultCapabilities(t *testing.T) {
	caps := DefaultCapabilities()
	if !caps.Color256 {
		t.Error("expected Color256 in defaults")
	}
	if !caps.SGRMouse {
		t.Error("expected SGRMouse in defaults")
	}
	if !caps.BracketedPaste {
		t.Error("expected BracketedPaste in defaults")
	}
	if caps.TrueColor {
		t.Error("TrueColor should NOT be in defaults")
	}
}

func TestP297_DetectCapabilities_Kitty(t *testing.T) {
	caps := DetectCapabilities("", "", "\x1b[>kitty(0.32)\x07")
	if !caps.OSC52 {
		t.Error("kitty should have OSC52")
	}
	if !caps.KittyGraphics {
		t.Error("kitty should have KittyGraphics")
	}
	if !caps.Iterm2Images {
		t.Error("kitty should have Iterm2Images")
	}
	if caps.Sixel {
		t.Error("kitty should NOT have Sixel")
	}
	if !caps.OSC133 {
		t.Error("kitty should have OSC133")
	}
	if !caps.TrueColor {
		t.Error("kitty should have TrueColor")
	}
	if !caps.KittyKeyboard {
		t.Error("kitty should have KittyKeyboard")
	}
	if !caps.Sync {
		t.Error("kitty should have Sync")
	}
	if caps.TerminalName != "kitty" {
		t.Errorf("TerminalName = %q", caps.TerminalName)
	}
}

func TestP297_DetectCapabilities_WezTerm(t *testing.T) {
	caps := DetectCapabilities("", "", "\x1b[>WezTerm(20240203)\x07")
	if !caps.Sixel {
		t.Error("wezterm should have Sixel")
	}
	if !caps.KittyGraphics {
		t.Error("wezterm should have KittyGraphics")
	}
	if !caps.OSC8 {
		t.Error("wezterm should have OSC8")
	}
	if caps.TerminalName != "WezTerm" {
		t.Errorf("TerminalName = %q", caps.TerminalName)
	}
}

func TestP297_DetectCapabilities_iTerm(t *testing.T) {
	caps := DetectCapabilities("", "", "\x1b[>iTerm.app(3.5)\x07")
	if !caps.Iterm2Images {
		t.Error("iTerm should have Iterm2Images")
	}
	if caps.Sixel {
		t.Error("iTerm should NOT have Sixel")
	}
	if caps.KittyGraphics {
		t.Error("iTerm should NOT have KittyGraphics")
	}
}

func TestP297_DetectCapabilities_Alacritty(t *testing.T) {
	caps := DetectCapabilities("", "", "\x1b[>alacritty(0.13)\x07")
	if !caps.OSC8 {
		t.Error("alacritty should have OSC8")
	}
	if caps.KittyKeyboard {
		t.Error("alacritty should NOT have KittyKeyboard")
	}
}

func TestP297_DetectCapabilities_Tmux(t *testing.T) {
	caps := DetectCapabilities("", "", "\x1b[>tmux(3.3a)\x07")
	if !caps.OSC52 {
		t.Error("tmux should have OSC52")
	}
	if !caps.OSC8 {
		t.Error("tmux should have OSC8")
	}
}

func TestP297_DetectCapabilities_XTerm(t *testing.T) {
	caps := DetectCapabilities("", "", "\x1b[>xterm(372)\x07")
	if !caps.OSC8 {
		t.Error("xterm should have OSC8")
	}
	if caps.Sixel {
		t.Error("xterm default should NOT have Sixel")
	}
}

func TestP297_DetectCapabilities_Foot(t *testing.T) {
	caps := DetectCapabilities("", "", "\x1b[>foot(1.16)\x07")
	if !caps.Sixel {
		t.Error("foot should have Sixel")
	}
	if !caps.KittyKeyboard {
		t.Error("foot should have KittyKeyboard")
	}
}

func TestP297_DetectCapabilities_Unknown(t *testing.T) {
	caps := DetectCapabilities("", "", "")
	// Should fall back to defaults
	if !caps.Color256 {
		t.Error("unknown should have Color256 (default)")
	}
	if caps.TerminalName != "" {
		t.Errorf("TerminalName = %q, want empty", caps.TerminalName)
	}
}

func TestP297_DetectCapabilitiesFromEnv_iTerm(t *testing.T) {
	caps := DetectCapabilitiesFromEnv("xterm-256color", "iTerm.app", "truecolor")
	if !caps.Iterm2Images {
		t.Error("should detect iTerm2Images from TERM_PROGRAM")
	}
	if !caps.TrueColor {
		t.Error("should detect TrueColor from COLORTERM")
	}
}

func TestP297_DetectCapabilitiesFromEnv_Kitty(t *testing.T) {
	caps := DetectCapabilitiesFromEnv("xterm-kitty", "kitty", "truecolor")
	if !caps.KittyGraphics {
		t.Error("should detect KittyGraphics from TERM_PROGRAM")
	}
}

func TestP297_DetectCapabilitiesFromEnv_WezTerm(t *testing.T) {
	caps := DetectCapabilitiesFromEnv("wezterm", "WezTerm", "truecolor")
	if !caps.Sixel {
		t.Error("should detect Sixel from TERM_PROGRAM")
	}
}

func TestP297_DetectCapabilitiesFromEnv_VSCode(t *testing.T) {
	caps := DetectCapabilitiesFromEnv("xterm-256color", "vscode", "")
	if !caps.OSC8 {
		t.Error("vscode should have OSC8")
	}
	if !caps.TrueColor {
		t.Error("vscode should have TrueColor")
	}
}

func TestP297_DetectCapabilitiesFromEnv_SixelInTerm(t *testing.T) {
	caps := DetectCapabilitiesFromEnv("xterm-sixel", "", "")
	if !caps.Sixel {
		t.Error("should detect Sixel from TERM=xterm-sixel")
	}
}

func TestP297_DetectCapabilitiesFromEnv_Default(t *testing.T) {
	caps := DetectCapabilitiesFromEnv("xterm-256color", "", "")
	if !caps.Color256 {
		t.Error("should have Color256 from TERM=xterm-256color")
	}
	if !caps.SGRMouse {
		t.Error("should have SGRMouse (default)")
	}
}

func TestP297_Merge(t *testing.T) {
	base := DefaultCapabilities()
	enhanced := TerminalCapabilities{
		TrueColor:   true,
		OSC8:        true,
		Sixel:       true,
	}
	base.Merge(enhanced)
	if !base.TrueColor {
		t.Error("should have TrueColor after merge")
	}
	if !base.OSC8 {
		t.Error("should have OSC8 after merge")
	}
	if !base.Sixel {
		t.Error("should have Sixel after merge")
	}
	if !base.SGRMouse {
		t.Error("should retain SGRMouse after merge")
	}
}

func TestP297_HasAny(t *testing.T) {
	minimal := DefaultCapabilities()
	if minimal.HasAny() {
		t.Error("default caps should have HasAny=false (only baseline)")
	}
	rich := TerminalCapabilities{TrueColor: true}
	if !rich.HasAny() {
		t.Error("TrueColor should trigger HasAny")
	}
}

func TestP297_ParseVersionForDetect(t *testing.T) {
	tests := []struct {
		input string
		name  string
		ver   string
		ok    bool
	}{
		{"\x1b[>kitty(0.32)\x07", "kitty", "0.32", true},
		{"\x1b[>WezTerm(20240203)\x07", "WezTerm", "20240203", true},
		{"\x1b[>xterm;372q", "xterm", "372", true},
		{"\x1b[>foot(1.16)\x1b\\", "foot", "1.16", true},
		{"", "", "", false},
		{"short", "", "", false},
	}
	for _, tt := range tests {
		name, ver, ok := parseVersionForDetect(tt.input)
		if ok != tt.ok {
			t.Errorf("parseVersionForDetect(%q) ok=%v, want %v", tt.input, ok, tt.ok)
			continue
		}
		if ok && (name != tt.name || ver != tt.ver) {
			t.Errorf("parseVersionForDetect(%q) = (%q,%q), want (%q,%q)",
				tt.input, name, ver, tt.name, tt.ver)
		}
	}
}

package term

import "testing"

// P324: Push DetectCapabilities and ParseDA2Response past 90%

// === DetectCapabilities: contour, foot, alacritty, tmux, xterm, unknown DA2 ===

func TestP324_DetectCapabilities_Contour(t *testing.T) {
	caps := DetectCapabilities("", "", "\x1b[>contour(0.3)\x07")
	if !caps.Sixel {
		t.Error("contour should have Sixel")
	}
	if !caps.KittyGraphics {
		t.Error("contour should have KittyGraphics")
	}
}

func TestP324_DetectCapabilities_Foot(t *testing.T) {
	caps := DetectCapabilities("", "", "\x1b[>foot(1.16)\x07")
	if !caps.KittyKeyboard {
		t.Error("foot should have KittyKeyboard")
	}
}

func TestP324_DetectCapabilities_Alacritty(t *testing.T) {
	caps := DetectCapabilities("", "", "\x1b[>alacritty(0.13)\x07")
	if caps.KittyKeyboard {
		t.Error("alacritty should NOT have KittyKeyboard")
	}
	if !caps.OSC8 {
		t.Error("alacritty should have OSC8")
	}
}

func TestP324_DetectCapabilities_XTerm(t *testing.T) {
	caps := DetectCapabilities("", "", "\x1b[>xterm(372)\x07")
	if caps.Sixel {
		t.Error("xterm default should NOT have Sixel")
	}
	if !caps.OSC8 {
		t.Error("xterm should have OSC8")
	}
}

func TestP324_DetectCapabilities_Tmux(t *testing.T) {
	caps := DetectCapabilities("", "", "\x1b[>tmux(3.3a)\x07")
	if !caps.OSC52 {
		t.Error("tmux should have OSC52")
	}
	if !caps.OSC133 {
		t.Error("tmux should have OSC133")
	}
}

func TestP324_DetectCapabilities_DefaultFallback(t *testing.T) {
	// Unknown terminal with DA2 type 65 (VT500)
	caps := DetectCapabilities("", "\x1b[>65;0;0c", "")
	if !caps.TrueColor {
		t.Error("VT500 type should detect TrueColor")
	}
}

func TestP324_DetectCapabilities_UnknownNoDA2(t *testing.T) {
	// Completely unknown, no DA2
	caps := DetectCapabilities("", "", "")
	if caps.TerminalName != "" {
		t.Errorf("name = %q", caps.TerminalName)
	}
	if !caps.SGRMouse {
		t.Error("should have defaults")
	}
}

func TestP324_DetectCapabilities_ITerm(t *testing.T) {
	caps := DetectCapabilities("", "", "\x1b[>iTerm.app(3.5)\x07")
	if !caps.Iterm2Images {
		t.Error("iTerm should have Iterm2Images")
	}
	if caps.KittyGraphics {
		t.Error("iTerm should NOT have KittyGraphics")
	}
}

func TestP324_DetectCapabilities_WezTerm(t *testing.T) {
	caps := DetectCapabilities("", "", "\x1b[>WezTerm(20240203)\x07")
	if !caps.Sixel {
		t.Error("wezterm should have Sixel")
	}
	if !caps.KittyGraphics {
		t.Error("wezterm should have KittyGraphics")
	}
}

// === ParseDA2Response edge cases ===

func TestP324_ParseDA2Response_TwoFields(t *testing.T) {
	// Only type + version, no ROM field
	resp, ok := ParseDA2Response("\x1b[>1;276c")
	if !ok {
		t.Error("should parse")
	}
	if resp.TerminalType != 1 {
		t.Errorf("type = %d", resp.TerminalType)
	}
	if resp.Version != 276 {
		t.Errorf("version = %d", resp.Version)
	}
	if resp.ROMCartridges != 0 {
		t.Errorf("rom = %d", resp.ROMCartridges)
	}
}

func TestP324_ParseDA2Response_OneField(t *testing.T) {
	// Only type
	resp, ok := ParseDA2Response("\x1b[>62c")
	if !ok {
		t.Error("should parse")
	}
	if resp.TerminalType != 62 {
		t.Errorf("type = %d", resp.TerminalType)
	}
}

func TestP324_ParseDA2Response_WrongTerminator(t *testing.T) {
	_, ok := ParseDA2Response("\x1b[>1;276;0x")
	if ok {
		t.Error("should fail with wrong terminator")
	}
}

func TestP324_ParseDA2Response_WrongPrefix(t *testing.T) {
	_, ok := ParseDA2Response("\x1b[?1;276;0c")
	if ok {
		t.Error("should fail with ? prefix")
	}
}

func TestP324_ParseDA2Response_EmptyBody(t *testing.T) {
	_, ok := ParseDA2Response("\x1b[>c")
	if ok {
		t.Error("should fail with empty body")
	}
}

package term

import (
	"strings"
)

// TerminalCapabilities describes which terminal protocols are available
// on the current terminal. Use DetectCapabilities or DetectCapabilitiesFromEnv
// to populate it, then check individual fields before emitting protocol sequences.
type TerminalCapabilities struct {
	// Clipboard
	OSC52 bool // OSC 52 clipboard (set/read selection)

	// Images
	Sixel        bool // Sixel graphics
	KittyGraphics bool // Kitty graphics protocol
	Iterm2Images  bool // iTerm2 inline images

	// Hyperlinks
	OSC8   bool // clickable hyperlinks
	OSC133 bool // shell integration / semantic prompts

	// Color
	TrueColor bool // 24-bit color
	Color256  bool // 256-color palette

	// Input
	KittyKeyboard  bool // Kitty keyboard protocol
	SGRMouse       bool // SGR mouse mode (CSI ?1006)
	BracketedPaste bool // bracketed paste mode (CSI ?2004)
	FocusTracking  bool // focus in/out events (CSI ?1004)

	// Rendering
	Sync          bool // synchronized output (BSU/ESU)
	WindowTitles  bool // OSC 0/1/2 window title

	// DSR
	CursorPosition bool // DSR cursor position report
	TerminalSize   bool // DSR terminal size report

	// Metadata
	TerminalName string // detected terminal name
}

// DefaultCapabilities returns a conservative baseline that 99%+ of
// terminals support. Use this when detection fails or for safe defaults.
func DefaultCapabilities() TerminalCapabilities {
	return TerminalCapabilities{
		Color256:       true,
		SGRMouse:       true,
		BracketedPaste: true,
		WindowTitles:   true,
	}
}

// DetectCapabilities inspects DA1/DA2/XTVersion response strings and
// returns capabilities based on known terminal signatures.
//
// Parameters:
//   - da1: raw DA1 response string (may be empty if unavailable)
//   - da2: raw DA2 response string (may be empty)
//   - version: raw XTVERSION response string (may be empty)
func DetectCapabilities(da1, da2, version string) TerminalCapabilities {
	caps := DefaultCapabilities()

	// Parse XTVersion for terminal name detection
	name, verStr, _ := parseVersionForDetect(version)

	// Parse DA2 for terminal type
	da2Resp, _ := ParseDA2Response(da2)

	caps.TerminalName = name
	_ = verStr // version available if needed for fine-grained checks

	// Match known terminals
	termLower := strings.ToLower(name)

	switch {
	case strings.Contains(termLower, "kitty"):
		caps.OSC52 = true
		caps.Sixel = false // Kitty doesn't support Sixel, uses own protocol
		caps.KittyGraphics = true
		caps.Iterm2Images = true // Kitty also supports iTerm2 inline images
		caps.OSC8 = true
		caps.OSC133 = true
		caps.TrueColor = true
		caps.KittyKeyboard = true
		caps.FocusTracking = true
		caps.Sync = true
		caps.CursorPosition = true
		caps.TerminalSize = true

	case strings.Contains(termLower, "wezterm"):
		caps.OSC52 = true
		caps.Sixel = true
		caps.KittyGraphics = true
		caps.Iterm2Images = true
		caps.OSC8 = true
		caps.OSC133 = true
		caps.TrueColor = true
		caps.KittyKeyboard = true
		caps.FocusTracking = true
		caps.Sync = true
		caps.CursorPosition = true
		caps.TerminalSize = true

	case strings.Contains(termLower, "iterm") || strings.Contains(termLower, "itreml"):
		caps.OSC52 = true
		caps.Sixel = false // iTerm2 doesn't natively support Sixel
		caps.KittyGraphics = false
		caps.Iterm2Images = true
		caps.OSC8 = true
		caps.OSC133 = true
		caps.TrueColor = true
		caps.FocusTracking = true
		caps.Sync = true
		caps.CursorPosition = true
		caps.TerminalSize = true

	case strings.Contains(termLower, "foot"):
		caps.OSC52 = true
		caps.Sixel = true
		caps.KittyGraphics = false
		caps.Iterm2Images = false
		caps.OSC8 = true
		caps.OSC133 = true
		caps.TrueColor = true
		caps.KittyKeyboard = true
		caps.FocusTracking = true
		caps.Sync = true
		caps.CursorPosition = true
		caps.TerminalSize = true

	case strings.Contains(termLower, "alacritty"):
		caps.OSC52 = true
		caps.Sixel = false
		caps.KittyGraphics = false
		caps.Iterm2Images = false
		caps.OSC8 = true
		caps.OSC133 = true
		caps.TrueColor = true
		caps.KittyKeyboard = false
		caps.FocusTracking = true
		caps.Sync = true
		caps.CursorPosition = true
		caps.TerminalSize = true

	case strings.Contains(termLower, "contour"):
		caps.OSC52 = true
		caps.Sixel = true
		caps.KittyGraphics = true
		caps.Iterm2Images = true
		caps.OSC8 = true
		caps.OSC133 = true
		caps.TrueColor = true
		caps.KittyKeyboard = true
		caps.FocusTracking = true
		caps.Sync = true
		caps.CursorPosition = true
		caps.TerminalSize = true

	case strings.Contains(termLower, "xterm"):
		caps.OSC52 = true
		caps.OSC8 = true
		caps.OSC133 = true
		caps.TrueColor = true
		caps.FocusTracking = true
		caps.Sync = true
		caps.CursorPosition = true
		// Sixel support depends on build flags, default off
		caps.Sixel = false

	case strings.Contains(termLower, "tmux"):
		caps.OSC52 = true
		caps.OSC8 = true
		caps.OSC133 = true
		caps.TrueColor = true
		caps.FocusTracking = true
		caps.Sync = true
		caps.CursorPosition = true

	default:
		// Use DA2 terminal type for fallback detection
		switch da2Resp.TerminalType {
		case 65: // VT500 series (often indicates modern terminal)
			caps.TrueColor = true
			caps.OSC8 = true
		}
	}

	return caps
}

// DetectCapabilitiesFromEnv detects capabilities from environment variables.
// This is a convenience for CI/automation where interactive queries aren't possible.
//
// Parameters:
//   - term: value of $TERM (e.g. "xterm-256color")
//   - termProgram: value of $TERM_PROGRAM (e.g. "iTerm.app", "WezTerm")
//   - colorTerm: value of $COLORTERM (e.g. "truecolor", "24bit")
func DetectCapabilitiesFromEnv(term, termProgram, colorTerm string) TerminalCapabilities {
	caps := DefaultCapabilities()

	// TrueColor detection from COLORTERM
	ctLower := strings.ToLower(colorTerm)
	if strings.Contains(ctLower, "truecolor") || strings.Contains(ctLower, "24bit") {
		caps.TrueColor = true
	}

	// TERM_PROGRAM based detection
	tpLower := strings.ToLower(termProgram)
	switch {
	case strings.Contains(tpLower, "iterm"):
		caps.Merge(DetectCapabilities("", "", "\x1b[>iTerm.app(3.5)\x07"))
	case strings.Contains(tpLower, "wezterm"):
		caps.Merge(DetectCapabilities("", "", "\x1b[>WezTerm(20240203)\x07"))
	case strings.Contains(tpLower, "kitty"):
		caps.Merge(DetectCapabilities("", "", "\x1b[>kitty(0.32)\x07"))
	case strings.Contains(tpLower, "ghostty"):
		caps.Merge(DetectCapabilities("", "", "\x1b[>ghostty(1.0)\x07"))
	case strings.Contains(tpLower, "vscode"):
		caps.OSC52 = true
		caps.OSC8 = true
		caps.OSC133 = true
		caps.TrueColor = true
		caps.Sync = true
	}

	// TERM based fallback
	termLower := strings.ToLower(term)
	if strings.Contains(termLower, "256color") {
		caps.Color256 = true
	}
	if strings.Contains(termLower, "sixel") {
		caps.Sixel = true
	}

	return caps
}

// Merge merges another capabilities set into this one (logical OR for each field).
func (c *TerminalCapabilities) Merge(other TerminalCapabilities) {
	c.OSC52 = c.OSC52 || other.OSC52
	c.Sixel = c.Sixel || other.Sixel
	c.KittyGraphics = c.KittyGraphics || other.KittyGraphics
	c.Iterm2Images = c.Iterm2Images || other.Iterm2Images
	c.OSC8 = c.OSC8 || other.OSC8
	c.OSC133 = c.OSC133 || other.OSC133
	c.TrueColor = c.TrueColor || other.TrueColor
	c.Color256 = c.Color256 || other.Color256
	c.KittyKeyboard = c.KittyKeyboard || other.KittyKeyboard
	c.SGRMouse = c.SGRMouse || other.SGRMouse
	c.BracketedPaste = c.BracketedPaste || other.BracketedPaste
	c.FocusTracking = c.FocusTracking || other.FocusTracking
	c.Sync = c.Sync || other.Sync
	c.WindowTitles = c.WindowTitles || other.WindowTitles
	c.CursorPosition = c.CursorPosition || other.CursorPosition
	c.TerminalSize = c.TerminalSize || other.TerminalSize
	if other.TerminalName != "" {
		c.TerminalName = other.TerminalName
	}
}

// HasAny returns true if at least one advanced capability is present
// (beyond the default baseline).
func (c TerminalCapabilities) HasAny() bool {
	return c.OSC52 || c.Sixel || c.KittyGraphics || c.Iterm2Images ||
		c.OSC8 || c.OSC133 || c.TrueColor || c.KittyKeyboard || c.Sync
}

// parseVersionForDetect is a lightweight parser that returns name and version
// from an XTVERSION response string, without the full strict validation.
func parseVersionForDetect(version string) (name, ver string, ok bool) {
	if len(version) < 4 || version[0] != 0x1b || version[1] != '[' || version[2] != '>' {
		return "", "", false
	}
	rest := version[3:]
	// Strip terminators
	for len(rest) > 0 {
		last := rest[len(rest)-1]
		if last == 0x07 {
			rest = rest[:len(rest)-1]
		} else if last == '\\' && len(rest) >= 2 && rest[len(rest)-2] == 0x1b {
			rest = rest[:len(rest)-2]
		} else if last == 'q' {
			rest = rest[:len(rest)-1]
		} else {
			break
		}
	}
	// Parse name(version) or name;version
	openParen := indexOfByte(rest, '(')
	if openParen >= 0 {
		name = rest[:openParen]
		closeParen := indexOfByte(rest[openParen+1:], ')')
		if closeParen >= 0 {
			ver = rest[openParen+1 : openParen+1+closeParen]
		}
		return name, ver, true
	}
	semiPos := indexOfByte(rest, ';')
	if semiPos >= 0 {
		return rest[:semiPos], rest[semiPos+1:], true
	}
	return rest, "", true
}

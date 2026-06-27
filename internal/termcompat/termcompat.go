// Package termcompat detects terminal capabilities and provides fallback
// strategies for features that may not be supported.
//
// It reads standard environment variables like $TERM_PROGRAM, $TERM, and
// $COLORTERM to determine which features the current terminal supports,
// such as OSC52 clipboard, true color, bracketed paste, and SGR mouse.
package termcompat

import (
	"os"
	"strings"
)

// Capabilities describes the feature set of the current terminal.
type Capabilities struct {
	// Name is the detected terminal name (e.g., "iTerm.app", "WezTerm").
	Name string

	// HasOSC52 indicates support for OSC52 clipboard sequences.
	HasOSC52 bool

	// HasTrueColor indicates 24-bit color support.
	HasTrueColor bool

	// Has256Color indicates 256-color support.
	Has256Color bool

	// HasBracketedPaste indicates bracketed paste mode support.
	HasBracketedPaste bool

	// HasSGRMouse indicates SGR (1006) mouse mode support.
	HasSGRMouse bool

	// HasURLHyperlinks indicates OSC8 hyperlink support.
	HasURLHyperlinks bool

	// InsideTmux indicates the terminal is inside tmux.
	InsideTmux bool

	// InsideScreen indicates the terminal is inside GNU screen.
	InsideScreen bool

	// IsSSH indicates the session is over SSH.
	IsSSH bool
}

// EnvGetter abstracts environment variable access for testability.
type EnvGetter interface {
	Get(key string) string
}

// envGetterImpl implements EnvGetter using os.Getenv.
type envGetterImpl struct{}

func (envGetterImpl) Get(key string) string { return os.Getenv(key) }

// terminalEntry holds known capabilities for a specific terminal.
type terminalEntry struct {
	OSC52          bool
	TrueColor      bool
	BracketedPaste bool
	SGRMouse       bool
}

// Detect queries the current environment and returns terminal capabilities.
func Detect() Capabilities {
	return DetectFromEnv(envGetterImpl{})
}

// DetectFromEnv detects capabilities from the provided environment getter.
// This allows tests to inject mock environment variables.
func DetectFromEnv(getter EnvGetter) Capabilities {
	termProgram := getter.Get("TERM_PROGRAM")
	term := getter.Get("TERM")
	colorTerm := getter.Get("COLORTERM")

	caps := Capabilities{
		Name: identifyTerminal(termProgram, term, getter),
	}

	// Detect multiplexers.
	caps.InsideTmux = getter.Get("TMUX") != ""
	caps.InsideScreen = getter.Get("STY") != "" ||
		term == "screen" || strings.HasPrefix(term, "screen-")

	// Detect SSH.
	caps.IsSSH = getter.Get("SSH_CONNECTION") != "" ||
		getter.Get("SSH_CLIENT") != "" ||
		getter.Get("SSH_TTY") != ""

	// True color detection: check COLORTERM first, then terminal database.
	if colorTerm == "truecolor" || colorTerm == "24bit" {
		caps.HasTrueColor = true
	}

	// 256 color detection.
	if strings.Contains(term, "256color") {
		caps.Has256Color = true
	}

	// Look up the terminal in the database for capability flags.
	entry, known := terminalDB[caps.Name]
	if known {
		if !caps.HasTrueColor {
			caps.HasTrueColor = entry.TrueColor
		}
		if entry.TrueColor {
			caps.HasTrueColor = true
		}
		// Terminals in the DB with true color also support 256.
		if entry.TrueColor || entry.BracketedPaste {
			caps.Has256Color = true
		}
		caps.HasOSC52 = entry.OSC52
		caps.HasBracketedPaste = entry.BracketedPaste
		caps.HasSGRMouse = entry.SGRMouse
	}

	// Conservative fallback for unknown terminals.
	if !known {
		// Most modern terminals support bracketed paste and SGR mouse.
		caps.HasBracketedPaste = isModernTerminal(term)
		caps.HasSGRMouse = isModernTerminal(term)

		// OSC52: conservative — only enable if explicitly detected.
		caps.HasOSC52 = false

		// 256 color: assume yes for most terminals.
		if !caps.Has256Color && !caps.HasTrueColor {
			caps.Has256Color = isModernTerminal(term)
		}
	}

	// tmux passthrough: tmux supports OSC52 if configured, and passes through
	// truecolor. We enable OSC52 conservatively — tmux may strip it unless
	// allow-passthrough is on, but modern tmux (3.3+) enables it by default.
	if caps.InsideTmux {
		// tmux supports SGR mouse and bracketed paste natively.
		caps.HasSGRMouse = true
		caps.HasBracketedPaste = true
		// True color passes through tmux 2.2+.
		if known && entry.TrueColor {
			caps.HasTrueColor = true
		}
	}

	// GNU screen is very limited.
	if caps.InsideScreen {
		caps.HasOSC52 = false
		caps.HasTrueColor = false
		// Screen supports 256 colors in modern versions.
		caps.Has256Color = true
	}

	return caps
}

// String returns a human-readable description of the capabilities.
func (c Capabilities) String() string {
	var sb strings.Builder
	sb.WriteString(c.Name)
	if c.InsideTmux {
		sb.WriteString(" (tmux)")
	}
	if c.InsideScreen {
		sb.WriteString(" (screen)")
	}
	sb.WriteByte(':')

	if c.HasTrueColor {
		sb.WriteString(" truecolor")
	} else if c.Has256Color {
		sb.WriteString(" 256color")
	} else {
		sb.WriteString(" 16color")
	}

	if c.HasOSC52 {
		sb.WriteString(" osc52")
	}
	if c.HasBracketedPaste {
		sb.WriteString(" paste")
	}
	if c.HasSGRMouse {
		sb.WriteString(" mouse")
	}

	return sb.String()
}

// OSC52Mode controls how OSC52 clipboard sequences are handled.
type OSC52Mode int

const (
	// OSC52Enabled means OSC52 is fully supported and can be used.
	OSC52Enabled OSC52Mode = iota
	// OSC52Disabled means OSC52 is not supported — fall back to other methods.
	OSC52Disabled
)

// OSC52Support returns the recommended OSC52 handling mode.
func (c Capabilities) OSC52Support() OSC52Mode {
	if c.HasOSC52 {
		return OSC52Enabled
	}
	return OSC52Disabled
}

// ColorDepth returns the best supported color depth.
func (c Capabilities) ColorDepth() int {
	if c.HasTrueColor {
		return 24
	}
	if c.Has256Color {
		return 8
	}
	return 4
}

// ShouldUseOSC52 returns true if OSC52 should be used for clipboard operations.
// It returns false for terminals known to not support it.
func (c Capabilities) ShouldUseOSC52() bool {
	return c.HasOSC52 && !c.InsideScreen
}

// identifyTerminal determines the terminal name from environment variables.
func identifyTerminal(termProgram, term string, getter EnvGetter) string {
	// TERM_PROGRAM is the most reliable signal on macOS/Linux.
	if termProgram != "" {
		switch {
		case strings.Contains(termProgram, "iTerm"):
			return "iTerm.app"
		case strings.Contains(termProgram, "WezTerm"):
			return "WezTerm"
		case strings.Contains(termProgram, "vscode"):
			return "VSCode"
		case strings.Contains(termProgram, "Apple_Terminal"):
			return "Apple Terminal"
		case strings.Contains(termProgram, "Hyper"):
			return "Hyper"
		case strings.Contains(termProgram, "Alacritty"):
			return "Alacritty"
		case strings.Contains(termProgram, "kitty"):
			return "kitty"
		case strings.Contains(termProgram, "ghostty"):
			return "Ghostty"
		case strings.Contains(termProgram, "tmux"):
			return "tmux"
		default:
			return termProgram
		}
	}

	// Windows Terminal sets WT_SESSION — check before $TERM fallback
	// because Windows Terminal may use xterm-256color as its TERM.
	if getter.Get("WT_SESSION") != "" {
		return "Windows Terminal"
	}

	// Fall back to $TERM matching.
	switch {
	case strings.HasPrefix(term, "alacritty"):
		return "Alacritty"
	case strings.HasPrefix(term, "xterm-kitty"):
		return "kitty"
	case strings.HasPrefix(term, "wezterm"):
		return "WezTerm"
	case strings.Contains(term, "gnome") || strings.Contains(term, "xterm-256"):
		return "GNOME Terminal"
	case strings.HasPrefix(term, "screen"):
		return "GNU screen"
	case strings.HasPrefix(term, "tmux"):
		return "tmux"
	}

	// Unknown terminal.
	if term != "" {
		return "unknown (" + term + ")"
	}
	return "unknown"
}

// isModernTerminal returns true for terminals that are likely to support
// modern features like bracketed paste and SGR mouse.
func isModernTerminal(term string) bool {
	// Dumb terminals definitely don't.
	if term == "dumb" || term == "" {
		return false
	}
	// Most xterm-compatible terminals are modern enough.
	return true
}

// terminalDB is the known terminal capability database.
var terminalDB = map[string]terminalEntry{
	"iTerm.app": {
		OSC52:          true,
		TrueColor:      true,
		BracketedPaste: true,
		SGRMouse:       true,
	},
	"WezTerm": {
		OSC52:          true,
		TrueColor:      true,
		BracketedPaste: true,
		SGRMouse:       true,
	},
	"kitty": {
		OSC52:          true,
		TrueColor:      true,
		BracketedPaste: true,
		SGRMouse:       true,
	},
	"Alacritty": {
		OSC52:          true,
		TrueColor:      true,
		BracketedPaste: true,
		SGRMouse:       true,
	},
	"Ghostty": {
		OSC52:          true,
		TrueColor:      true,
		BracketedPaste: true,
		SGRMouse:       true,
	},
	"GNOME Terminal": {
		OSC52:          false, // GNOME Terminal doesn't support OSC52
		TrueColor:      true,
		BracketedPaste: true,
		SGRMouse:       true,
	},
	"Apple Terminal": {
		OSC52:          true,
		TrueColor:      true,
		BracketedPaste: true,
		SGRMouse:       true,
	},
	"VSCode": {
		OSC52:          true,
		TrueColor:      true,
		BracketedPaste: true,
		SGRMouse:       true,
	},
	"Windows Terminal": {
		OSC52:          true,
		TrueColor:      true,
		BracketedPaste: true,
		SGRMouse:       true,
	},
	"Hyper": {
		OSC52:          false,
		TrueColor:      true,
		BracketedPaste: true,
		SGRMouse:       true,
	},
	"tmux": {
		OSC52:          true,
		TrueColor:      true, // passes through outer terminal's capability
		BracketedPaste: true,
		SGRMouse:       true,
	},
	"GNU screen": {
		OSC52:          false,
		TrueColor:      false,
		BracketedPaste: true,
		SGRMouse:       true,
	},
}

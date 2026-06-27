package termcompat

// TerminalMatrix defines a test configuration for a specific terminal
// environment. It captures the environment variables and expected
// capabilities for use in compatibility testing and documentation.
type TerminalMatrix struct {
	// Name is the human-readable terminal name.
	Name string

	// TermProgram is the value of $TERM_PROGRAM.
	TermProgram string

	// Term is the value of $TERM.
	Term string

	// ColorTerm is the value of $COLORTERM ("truecolor", "24bit", or "").
	ColorTerm string

	// ExtraEnv holds additional environment variables (e.g. TMUX, WT_SESSION).
	ExtraEnv map[string]string

	// ExpectedCaps are the capabilities we expect DetectFromEnv to return.
	ExpectedCaps Capabilities
}

// AllTerminals returns the full terminal compatibility test matrix.
// Each entry represents a real-world terminal configuration.
//
// This matrix serves two purposes:
//  1. Tests verify that DetectFromEnv produces correct capabilities for each.
//  2. Documentation showing which features are supported per terminal.
func AllTerminals() []TerminalMatrix {
	return []TerminalMatrix{
		{
			Name:        "iTerm.app",
			TermProgram: "iTerm.app",
			Term:        "xterm-256color",
			ColorTerm:   "truecolor",
			ExpectedCaps: Capabilities{
				Name:             "iTerm.app",
				HasOSC52:         true,
				HasTrueColor:     true,
				Has256Color:      true,
				HasBracketedPaste: true,
				HasSGRMouse:      true,
			},
		},
		{
			Name:        "WezTerm",
			TermProgram: "WezTerm",
			Term:        "wezterm",
			ColorTerm:   "truecolor",
			ExpectedCaps: Capabilities{
				Name:             "WezTerm",
				HasOSC52:         true,
				HasTrueColor:     true,
				Has256Color:      true,
				HasBracketedPaste: true,
				HasSGRMouse:      true,
			},
		},
		{
			Name:        "kitty",
			TermProgram: "kitty",
			Term:        "xterm-kitty",
			ColorTerm:   "truecolor",
			ExpectedCaps: Capabilities{
				Name:             "kitty",
				HasOSC52:         true,
				HasTrueColor:     true,
				Has256Color:      true,
				HasBracketedPaste: true,
				HasSGRMouse:      true,
			},
		},
		{
			Name:        "Alacritty",
			TermProgram: "Alacritty",
			Term:        "alacritty",
			ColorTerm:   "truecolor",
			ExpectedCaps: Capabilities{
				Name:             "Alacritty",
				HasOSC52:         true,
				HasTrueColor:     true,
				Has256Color:      true,
				HasBracketedPaste: true,
				HasSGRMouse:      true,
			},
		},
		{
			Name:        "GNOME Terminal",
			TermProgram: "",
			Term:        "xterm-256color",
			ColorTerm:   "",
			ExpectedCaps: Capabilities{
				Name:             "GNOME Terminal",
				HasOSC52:         false, // GNOME Terminal does NOT support OSC52
				HasTrueColor:     true,
				Has256Color:      true,
				HasBracketedPaste: true,
				HasSGRMouse:      true,
			},
		},
		{
			Name:        "Apple Terminal",
			TermProgram: "Apple_Terminal",
			Term:        "xterm-256color",
			ColorTerm:   "",
			ExpectedCaps: Capabilities{
				Name:             "Apple Terminal",
				HasOSC52:         true,
				HasTrueColor:     true,
				Has256Color:      true,
				HasBracketedPaste: true,
				HasSGRMouse:      true,
			},
		},
		{
			Name:        "tmux",
			TermProgram: "tmux",
			Term:        "tmux-256color",
			ColorTerm:   "",
			ExtraEnv:    map[string]string{"TMUX": "/tmp/tmux-1000/default,1234,0"},
			ExpectedCaps: Capabilities{
				Name:             "tmux",
				HasOSC52:         true,
				HasTrueColor:     true,
				Has256Color:      true,
				HasBracketedPaste: true,
				HasSGRMouse:      true,
				InsideTmux:       true,
			},
		},
		{
			Name:        "GNU screen",
			TermProgram: "",
			Term:        "screen-256color",
			ColorTerm:   "",
			ExtraEnv:    map[string]string{"STY": "12345.pts-0.host"},
			ExpectedCaps: Capabilities{
				Name:             "GNU screen",
				HasOSC52:         false, // Screen blocks OSC52
				HasTrueColor:     false, // Screen does not support truecolor
				Has256Color:      true,
				HasBracketedPaste: true,
				HasSGRMouse:      true,
				InsideScreen:     true,
			},
		},
		{
			Name:        "Windows Terminal",
			TermProgram: "",
			Term:        "xterm-256color",
			ColorTerm:   "truecolor",
			ExtraEnv:    map[string]string{"WT_SESSION": "abc-123-def"},
			ExpectedCaps: Capabilities{
				Name:             "Windows Terminal",
				HasOSC52:         true,
				HasTrueColor:     true,
				Has256Color:      true,
				HasBracketedPaste: true,
				HasSGRMouse:      true,
			},
		},
		{
			Name:        "VSCode Terminal",
			TermProgram: "vscode",
			Term:        "xterm-256color",
			ColorTerm:   "truecolor",
			ExpectedCaps: Capabilities{
				Name:             "VSCode",
				HasOSC52:         true,
				HasTrueColor:     true,
				Has256Color:      true,
				HasBracketedPaste: true,
				HasSGRMouse:      true,
			},
		},
		{
			Name:        "Hyper",
			TermProgram: "Hyper",
			Term:        "xterm-256color",
			ColorTerm:   "truecolor",
			ExpectedCaps: Capabilities{
				Name:             "Hyper",
				HasOSC52:         false, // Hyper does NOT support OSC52
				HasTrueColor:     true,
				Has256Color:      true,
				HasBracketedPaste: true,
				HasSGRMouse:      true,
			},
		},
		{
			Name:        "Ghostty",
			TermProgram: "ghostty",
			Term:        "xterm-ghostty",
			ColorTerm:   "truecolor",
			ExpectedCaps: Capabilities{
				Name:             "Ghostty",
				HasOSC52:         true,
				HasTrueColor:     true,
				Has256Color:      true,
				HasBracketedPaste: true,
				HasSGRMouse:      true,
			},
		},
	}
}

// FeatureSupport describes whether a specific feature is supported
// by a given terminal. Used for documentation and test assertions.
type FeatureSupport struct {
	Terminal       string
	TrueColor      bool
	Color256       bool
	OSC52          bool
	SGRMouse       bool
	BracketedPaste bool
	Unicode        bool // all listed terminals support basic Unicode
	CJK            bool // all listed terminals support CJK via font fallback
}

// CapabilityMatrix returns the feature support table for all known terminals.
// This is useful for generating documentation and verifying test expectations.
func CapabilityMatrix() []FeatureSupport {
	terminals := AllTerminals()
	matrix := make([]FeatureSupport, 0, len(terminals))
	for _, t := range terminals {
		matrix = append(matrix, FeatureSupport{
			Terminal:       t.ExpectedCaps.Name,
			TrueColor:      t.ExpectedCaps.HasTrueColor,
			Color256:       t.ExpectedCaps.Has256Color,
			OSC52:          t.ExpectedCaps.ShouldUseOSC52(),
			SGRMouse:       t.ExpectedCaps.HasSGRMouse,
			BracketedPaste: t.ExpectedCaps.HasBracketedPaste,
			Unicode:        true,
			CJK:            true,
		})
	}
	return matrix
}

// MatrixEnvGetter implements EnvGetter for a TerminalMatrix entry.
// It simulates the environment variables of that terminal.
type MatrixEnvGetter struct {
	Matrix TerminalMatrix
}

// Get returns the environment variable value for the matrix entry.
func (m MatrixEnvGetter) Get(key string) string {
	switch key {
	case "TERM_PROGRAM":
		return m.Matrix.TermProgram
	case "TERM":
		return m.Matrix.Term
	case "COLORTERM":
		return m.Matrix.ColorTerm
	default:
		if v, ok := m.Matrix.ExtraEnv[key]; ok {
			return v
		}
		return ""
	}
}

// DetectForMatrix runs DetectFromEnv against the given matrix entry
// and returns the detected capabilities.
func DetectForMatrix(m TerminalMatrix) Capabilities {
	return DetectFromEnv(MatrixEnvGetter{Matrix: m})
}

// SSHTerminalMatrix returns a matrix entry for an SSH session
// connected through an OSC52-capable terminal.
func SSHTerminalMatrix() TerminalMatrix {
	return TerminalMatrix{
		Name:        "iTerm.app (SSH)",
		TermProgram: "iTerm.app",
		Term:        "xterm-256color",
		ColorTerm:   "truecolor",
		ExtraEnv: map[string]string{
			"SSH_CONNECTION": "192.168.1.100 54321 10.0.0.1 22",
			"SSH_CLIENT":      "192.168.1.100 54321 22",
		},
		ExpectedCaps: Capabilities{
			Name:             "iTerm.app",
			HasOSC52:         true,
			HasTrueColor:     true,
			Has256Color:      true,
			HasBracketedPaste: true,
			HasSGRMouse:      true,
			IsSSH:            true,
		},
	}
}

// NestedMultiplexerMatrix returns a matrix entry for tmux inside SSH.
func NestedMultiplexerMatrix() TerminalMatrix {
	return TerminalMatrix{
		Name:        "iTerm.app (SSH → tmux)",
		TermProgram: "tmux",
		Term:        "tmux-256color",
		ColorTerm:   "",
		ExtraEnv: map[string]string{
			"TMUX":            "/tmp/tmux-1000/default,1234,0",
			"SSH_CONNECTION":  "192.168.1.100 54321 10.0.0.1 22",
			"SSH_CLIENT":      "192.168.1.100 54321 22",
		},
		ExpectedCaps: Capabilities{
			Name:             "tmux",
			HasOSC52:         true,
			HasTrueColor:     true,
			Has256Color:      true,
			HasBracketedPaste: true,
			HasSGRMouse:      true,
			InsideTmux:       true,
			IsSSH:            true,
		},
	}
}

// ScreenInsideTmuxMatrix returns a matrix entry for screen inside tmux
// (an unusual but possible nesting scenario).
func ScreenInsideTmuxMatrix() TerminalMatrix {
	return TerminalMatrix{
		Name:        "GNU screen (inside tmux)",
		TermProgram: "",
		Term:        "screen-256color",
		ColorTerm:   "",
		ExtraEnv: map[string]string{
			"STY": "12345.pts-0.host",
			"TMUX": "/tmp/tmux-1000/default,1234,0",
		},
		ExpectedCaps: Capabilities{
			Name:             "GNU screen",
			HasOSC52:         false,
			HasTrueColor:     false,
			Has256Color:      true,
			HasBracketedPaste: true,
			HasSGRMouse:      true,
			InsideTmux:       true,
			InsideScreen:     true,
		},
	}
}

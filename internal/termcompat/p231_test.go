package termcompat

import "testing"

// P231: identifyTerminal comprehensive branch coverage

func TestIdentifyTerminal_AllPrograms_P231(t *testing.T) {
	cases := []struct {
		termProgram string
		want        string
	}{
		{"iTerm.app", "iTerm.app"},
		{"WezTerm", "WezTerm"},
		{"vscode", "VSCode"},
		{"Apple_Terminal", "Apple Terminal"},
		{"Hyper", "Hyper"},
		{"Alacritty", "Alacritty"},
		{"kitty", "kitty"},
		{"ghostty", "Ghostty"},
		{"tmux", "tmux"},
		{"UnknownApp", "UnknownApp"}, // default case
	}
	for _, tc := range cases {
		got := identifyTerminal(tc.termProgram, "", &mockEnvGetter{})
		if got != tc.want {
			t.Errorf("identifyTerminal(%q,\"\") = %q, want %q", tc.termProgram, got, tc.want)
		}
	}
}

func TestIdentifyTerminal_WTSession_P231(t *testing.T) {
	got := identifyTerminal("", "xterm-256color", &mockEnvGetter{vars: map[string]string{"WT_SESSION": "12345"}})
	if got != "Windows Terminal" {
		t.Errorf("WT_SESSION should detect Windows Terminal, got %q", got)
	}
}

func TestIdentifyTerminal_TermFallback_P231(t *testing.T) {
	cases := []struct {
		term string
		want string
	}{
		{"alacritty", "Alacritty"},
		{"xterm-kitty", "kitty"},
		{"wezterm", "WezTerm"},
		{"gnome", "GNOME Terminal"},
		{"xterm-256color", "GNOME Terminal"},
		{"screen-256color", "GNU screen"},
		{"tmux-256color", "tmux"},
	}
	for _, tc := range cases {
		got := identifyTerminal("", tc.term, &mockEnvGetter{})
		if got != tc.want {
			t.Errorf("identifyTerminal(\"\",%q) = %q, want %q", tc.term, got, tc.want)
		}
	}
}

func TestIdentifyTerminal_Unknown_P231(t *testing.T) {
	got := identifyTerminal("", "some-unknown-term", &mockEnvGetter{})
	if got != "unknown (some-unknown-term)" {
		t.Errorf("unknown terminal, got %q", got)
	}
}

func TestIdentifyTerminal_Empty_P231(t *testing.T) {
	got := identifyTerminal("", "", &mockEnvGetter{})
	if got != "unknown" {
		t.Errorf("empty terminal should be unknown, got %q", got)
	}
}

func TestIsModernTerminal_P231(t *testing.T) {
	if isModernTerminal("dumb") {
		t.Error("dumb should not be modern")
	}
	if isModernTerminal("") {
		t.Error("empty should not be modern")
	}
	if !isModernTerminal("xterm-256color") {
		t.Error("xterm-256color should be modern")
	}
}

// mockEnvGetter implements EnvGetter for testing
type mockEnvGetter struct {
	vars map[string]string
}

func (m *mockEnvGetter) Get(key string) string {
	if m.vars == nil {
		return ""
	}
	return m.vars[key]
}

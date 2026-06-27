package termcompat

import (
	"strings"
	"testing"
)

// --- AllTerminals matrix tests ---

func TestMatrix_AllTerminals(t *testing.T) {
	terminals := AllTerminals()

	if len(terminals) < 12 {
		t.Fatalf("expected at least 12 terminal entries, got %d", len(terminals))
	}

	// Verify each entry has a name and TERM value.
	seen := map[string]bool{}
	for _, tm := range terminals {
		if tm.Name == "" {
			t.Fatal("found terminal entry with empty Name")
		}
		if tm.Term == "" {
			t.Fatalf("terminal %q has empty TERM", tm.Name)
		}
		if seen[tm.Name] {
			t.Fatalf("duplicate terminal name: %q", tm.Name)
		}
		seen[tm.Name] = true
	}

	// Verify key terminals are present.
	required := []string{
		"iTerm.app", "WezTerm", "kitty", "Alacritty",
		"GNOME Terminal", "Apple Terminal", "tmux", "GNU screen",
		"Windows Terminal",
	}
	for _, name := range required {
		if !seen[name] {
			t.Fatalf("missing required terminal: %q", name)
		}
	}
}

func TestMatrix_AllTerminals_DetectMatch(t *testing.T) {
	terminals := AllTerminals()

	for _, tm := range terminals {
		t.Run(tm.Name, func(t *testing.T) {
			caps := DetectForMatrix(tm)

			if caps.Name != tm.ExpectedCaps.Name {
				t.Errorf("Name: got %q, want %q", caps.Name, tm.ExpectedCaps.Name)
			}
			if caps.HasOSC52 != tm.ExpectedCaps.HasOSC52 {
				t.Errorf("HasOSC52: got %v, want %v", caps.HasOSC52, tm.ExpectedCaps.HasOSC52)
			}
			if caps.HasTrueColor != tm.ExpectedCaps.HasTrueColor {
				t.Errorf("HasTrueColor: got %v, want %v", caps.HasTrueColor, tm.ExpectedCaps.HasTrueColor)
			}
			if caps.Has256Color != tm.ExpectedCaps.Has256Color {
				t.Errorf("Has256Color: got %v, want %v", caps.Has256Color, tm.ExpectedCaps.Has256Color)
			}
			if caps.HasBracketedPaste != tm.ExpectedCaps.HasBracketedPaste {
				t.Errorf("HasBracketedPaste: got %v, want %v", caps.HasBracketedPaste, tm.ExpectedCaps.HasBracketedPaste)
			}
			if caps.HasSGRMouse != tm.ExpectedCaps.HasSGRMouse {
				t.Errorf("HasSGRMouse: got %v, want %v", caps.HasSGRMouse, tm.ExpectedCaps.HasSGRMouse)
			}
			if caps.InsideTmux != tm.ExpectedCaps.InsideTmux {
				t.Errorf("InsideTmux: got %v, want %v", caps.InsideTmux, tm.ExpectedCaps.InsideTmux)
			}
			if caps.InsideScreen != tm.ExpectedCaps.InsideScreen {
				t.Errorf("InsideScreen: got %v, want %v", caps.InsideScreen, tm.ExpectedCaps.InsideScreen)
			}
		})
	}
}

// --- Feature-specific matrix tests ---

func TestMatrix_TrueColor(t *testing.T) {
	// Terminals with COLORTERM=truecolor should detect true color.
	trueColorTerminals := []string{"iTerm.app", "WezTerm", "kitty", "Alacritty", "Ghostty"}
	terminalMap := map[string]bool{}
	for _, tm := range AllTerminals() {
		terminalMap[tm.Name] = tm.ExpectedCaps.HasTrueColor
	}

	for _, name := range trueColorTerminals {
		if !terminalMap[name] {
			t.Fatalf("expected %s to support true color", name)
		}
	}

	// GNU screen should NOT support true color.
	screen := findTerminal("GNU screen")
	if screen.ExpectedCaps.HasTrueColor {
		t.Fatal("expected GNU screen to NOT support true color")
	}
}

func TestMatrix_OSC52(t *testing.T) {
	// Known OSC52 supporters.
	osc52Yes := []string{"iTerm.app", "WezTerm", "kitty", "Alacritty", "Apple Terminal", "Windows Terminal", "tmux"}
	osc52No := []string{"GNOME Terminal", "GNU screen", "Hyper"}

	terminalMap := map[string]bool{}
	for _, tm := range AllTerminals() {
		terminalMap[tm.Name] = tm.ExpectedCaps.HasOSC52
	}

	for _, name := range osc52Yes {
		if !terminalMap[name] {
			t.Fatalf("expected %s to support OSC52", name)
		}
	}
	for _, name := range osc52No {
		if terminalMap[name] {
			t.Fatalf("expected %s to NOT support OSC52", name)
		}
	}
}

func TestMatrix_OSC52_ShouldUse(t *testing.T) {
	// ShouldUseOSC52 should return false for GNU screen.
	screen := findTerminal("GNU screen")
	if screen.ExpectedCaps.ShouldUseOSC52() {
		t.Fatal("expected ShouldUseOSC52()=false for GNU screen")
	}

	// ShouldUseOSC52 should return true for iTerm.app.
	iterm := findTerminal("iTerm.app")
	if !iterm.ExpectedCaps.ShouldUseOSC52() {
		t.Fatal("expected ShouldUseOSC52()=true for iTerm.app")
	}
}

func TestMatrix_Mouse(t *testing.T) {
	// All listed terminals should support SGR mouse except none.
	for _, tm := range AllTerminals() {
		if !tm.ExpectedCaps.HasSGRMouse {
			t.Fatalf("expected %s to support SGR mouse", tm.Name)
		}
	}
}

func TestMatrix_BracketedPaste(t *testing.T) {
	// All listed terminals should support bracketed paste.
	for _, tm := range AllTerminals() {
		if !tm.ExpectedCaps.HasBracketedPaste {
			t.Fatalf("expected %s to support bracketed paste", tm.Name)
		}
	}
}

func TestMatrix_256Color(t *testing.T) {
	// All listed terminals should support at least 256 colors.
	for _, tm := range AllTerminals() {
		if !tm.ExpectedCaps.Has256Color {
			t.Fatalf("expected %s to support 256 colors", tm.Name)
		}
	}
}

// --- Special scenario tests ---

func TestMatrix_Tmux(t *testing.T) {
	tmux := findTerminal("tmux")
	if !tmux.ExpectedCaps.InsideTmux {
		t.Fatal("expected tmux entry to have InsideTmux=true")
	}

	caps := DetectForMatrix(tmux)
	if !caps.InsideTmux {
		t.Fatal("expected detected caps to have InsideTmux=true")
	}

	// tmux should pass through OSC52.
	if !caps.HasOSC52 {
		t.Fatal("expected tmux to support OSC52")
	}

	// tmux should support true color (passes through outer terminal).
	if !caps.HasTrueColor {
		t.Fatal("expected tmux to pass through true color")
	}
}

func TestMatrix_Screen(t *testing.T) {
	screen := findTerminal("GNU screen")
	if !screen.ExpectedCaps.InsideScreen {
		t.Fatal("expected screen entry to have InsideScreen=true")
	}

	caps := DetectForMatrix(screen)
	if !caps.InsideScreen {
		t.Fatal("expected detected caps to have InsideScreen=true")
	}

	// Screen should NOT support OSC52.
	if caps.HasOSC52 {
		t.Fatal("expected GNU screen to NOT support OSC52")
	}

	// Screen should NOT support true color.
	if caps.HasTrueColor {
		t.Fatal("expected GNU screen to NOT support true color")
	}

	// Screen should support 256 colors.
	if !caps.Has256Color {
		t.Fatal("expected GNU screen to support 256 colors")
	}

	// ShouldUseOSC52 must return false for screen.
	if caps.ShouldUseOSC52() {
		t.Fatal("expected ShouldUseOSC52()=false for GNU screen")
	}
}

func TestMatrix_WindowsTerminal(t *testing.T) {
	wt := findTerminal("Windows Terminal")

	caps := DetectForMatrix(wt)
	if caps.Name != "Windows Terminal" {
		t.Fatalf("expected Name='Windows Terminal', got %q", caps.Name)
	}
	if !caps.HasOSC52 {
		t.Fatal("expected Windows Terminal to support OSC52")
	}
	if !caps.HasTrueColor {
		t.Fatal("expected Windows Terminal to support true color")
	}
}

func TestMatrix_SSHPassthrough(t *testing.T) {
	ssh := SSHTerminalMatrix()

	caps := DetectForMatrix(ssh)

	if !caps.IsSSH {
		t.Fatal("expected IsSSH=true for SSH session")
	}

	// SSH should NOT disable OSC52 — the terminal decides.
	if !caps.HasOSC52 {
		t.Fatal("expected OSC52=true for SSH session with iTerm.app")
	}

	// ShouldUseOSC52 should still work over SSH.
	if !caps.ShouldUseOSC52() {
		t.Fatal("expected ShouldUseOSC52()=true for SSH with capable terminal")
	}
}

func TestMatrix_NestedMultiplexer(t *testing.T) {
	// SSH → tmux: both flags should be set.
	nested := NestedMultiplexerMatrix()

	caps := DetectForMatrix(nested)

	if !caps.IsSSH {
		t.Fatal("expected IsSSH=true")
	}
	if !caps.InsideTmux {
		t.Fatal("expected InsideTmux=true")
	}

	// OSC52 should still work — tmux passes it through.
	if !caps.HasOSC52 {
		t.Fatal("expected OSC52=true for SSH→tmux")
	}
}

func TestMatrix_ScreenInsideTmux(t *testing.T) {
	// Edge case: screen inside tmux (unusual nesting).
	m := ScreenInsideTmuxMatrix()

	caps := DetectForMatrix(m)

	if !caps.InsideScreen {
		t.Fatal("expected InsideScreen=true")
	}

	// Screen forces OSC52 off.
	if caps.HasOSC52 {
		t.Fatal("expected OSC52=false when inside screen")
	}

	// Screen forces true color off.
	if caps.HasTrueColor {
		t.Fatal("expected true color=false when inside screen")
	}

	// ShouldUseOSC52 must return false.
	if caps.ShouldUseOSC52() {
		t.Fatal("expected ShouldUseOSC52()=false for screen inside tmux")
	}
}

// --- Capability fallback tests ---

func TestCapability_Fallback_UnknownTerminal(t *testing.T) {
	// Unknown terminal should still get basic capabilities via fallback.
	caps := DetectFromEnv(mockEnv{
		"TERM": "myterm-256color",
	})

	if caps.Name == "" {
		t.Fatal("expected non-empty name for unknown terminal")
	}

	// Should have 256 color via fallback.
	if !caps.Has256Color {
		t.Fatal("expected 256 color fallback for unknown modern terminal")
	}

	// Should NOT have OSC52 (conservative).
	if caps.HasOSC52 {
		t.Fatal("expected OSC52=false for unknown terminal (conservative)")
	}
}

func TestCapability_Fallback_DumbTerminal(t *testing.T) {
	caps := DetectFromEnv(mockEnv{
		"TERM": "dumb",
	})

	if caps.HasBracketedPaste {
		t.Fatal("expected bracketed paste=false for dumb terminal")
	}
	if caps.HasSGRMouse {
		t.Fatal("expected SGR mouse=false for dumb terminal")
	}
}

func TestCapability_Fallback_EmptyEnv(t *testing.T) {
	caps := DetectFromEnv(mockEnv{})

	// Empty env → unknown terminal, minimal capabilities.
	if caps.HasBracketedPaste {
		t.Fatal("expected bracketed paste=false for empty env")
	}
	if caps.HasSGRMouse {
		t.Fatal("expected SGR mouse=false for empty env")
	}
}

// --- CapabilityMatrix documentation tests ---

func TestCapabilityMatrix(t *testing.T) {
	matrix := CapabilityMatrix()

	if len(matrix) < 12 {
		t.Fatalf("expected at least 12 entries in capability matrix, got %d", len(matrix))
	}

	// Verify each entry has a terminal name.
	for _, fs := range matrix {
		if fs.Terminal == "" {
			t.Fatal("found entry with empty Terminal name")
		}
		// All terminals support Unicode and CJK (at least via font fallback).
		if !fs.Unicode {
			t.Fatalf("expected %s to support Unicode", fs.Terminal)
		}
		if !fs.CJK {
			t.Fatalf("expected %s to support CJK", fs.Terminal)
		}
	}
}

func TestCapabilityMatrix_OSC52Consistency(t *testing.T) {
	matrix := CapabilityMatrix()

	// Verify GNOME Terminal and screen show OSC52=false in the matrix.
	for _, fs := range matrix {
		if fs.Terminal == "GNOME Terminal" && fs.OSC52 {
			t.Fatal("expected GNOME Terminal OSC52=false in matrix")
		}
		if fs.Terminal == "GNU screen" && fs.OSC52 {
			t.Fatal("expected GNU screen OSC52=false in matrix")
		}
	}
}

// --- MatrixEnvGetter tests ---

func TestMatrixEnvGetter(t *testing.T) {
	tm := TerminalMatrix{
		Name:        "test",
		TermProgram: "TestTerm",
		Term:        "test-256color",
		ColorTerm:   "truecolor",
		ExtraEnv:    map[string]string{"FOO": "bar"},
	}

	g := MatrixEnvGetter{Matrix: tm}

	if g.Get("TERM_PROGRAM") != "TestTerm" {
		t.Fatal("expected TERM_PROGRAM=TestTerm")
	}
	if g.Get("TERM") != "test-256color" {
		t.Fatal("expected TERM=test-256color")
	}
	if g.Get("COLORTERM") != "truecolor" {
		t.Fatal("expected COLORTERM=truecolor")
	}
	if g.Get("FOO") != "bar" {
		t.Fatal("expected FOO=bar from ExtraEnv")
	}
	if g.Get("NONEXISTENT") != "" {
		t.Fatal("expected empty string for nonexistent env var")
	}
}

// --- Color depth tests across matrix ---

func TestMatrix_ColorDepth(t *testing.T) {
	for _, tm := range AllTerminals() {
		caps := DetectForMatrix(tm)
		depth := caps.ColorDepth()

		switch tm.Name {
		case "GNU screen":
			if depth != 8 {
				t.Fatalf("expected %s depth=8 (256 color), got %d", tm.Name, depth)
			}
		default:
			if depth != 24 {
				t.Fatalf("expected %s depth=24 (true color), got %d", tm.Name, depth)
			}
		}
	}
}

// --- String representation tests ---

func TestMatrix_StringRepresentation(t *testing.T) {
	for _, tm := range AllTerminals() {
		caps := DetectForMatrix(tm)
		s := caps.String()

		if !strings.Contains(s, tm.ExpectedCaps.Name) {
			t.Fatalf("expected String() to contain %q, got %q", tm.ExpectedCaps.Name, s)
		}

		// tmux entry should show (tmux) indicator.
		if tm.ExpectedCaps.InsideTmux {
			if !strings.Contains(s, "(tmux)") {
				t.Fatalf("expected String() to contain '(tmux)' for %s, got %q", tm.Name, s)
			}
		}

		// Screen entry should show (screen) indicator.
		if tm.ExpectedCaps.InsideScreen {
			if !strings.Contains(s, "(screen)") {
				t.Fatalf("expected String() to contain '(screen)' for %s, got %q", tm.Name, s)
			}
		}
	}
}

// --- Helpers ---

func findTerminal(name string) TerminalMatrix {
	for _, tm := range AllTerminals() {
		if tm.Name == name {
			return tm
		}
	}
	panic("terminal not found: " + name)
}

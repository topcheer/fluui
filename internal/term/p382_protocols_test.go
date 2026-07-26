package term

import "testing"

// P382: Tests for OSC 133 Shell Integration and Bracketed Paste

func TestP382_ShellPromptStart(t *testing.T) {
	got := ShellPromptStart()
	want := "\x1b]133;A\x07"
	if got != want {
		t.Errorf("ShellPromptStart() = %q, want %q", got, want)
	}
}

func TestP382_ShellCommandStart(t *testing.T) {
	got := ShellCommandStart()
	want := "\x1b]133;B\x07"
	if got != want {
		t.Errorf("ShellCommandStart() = %q, want %q", got, want)
	}
}

func TestP382_ShellOutputStart(t *testing.T) {
	got := ShellOutputStart()
	want := "\x1b]133;C\x07"
	if got != want {
		t.Errorf("ShellOutputStart() = %q, want %q", got, want)
	}
}

func TestP382_ShellOutputEnd_WithCode(t *testing.T) {
	got := ShellOutputEnd(0)
	want := "\x1b]133;D;0\x07"
	if got != want {
		t.Errorf("ShellOutputEnd(0) = %q, want %q", got, want)
	}

	got = ShellOutputEnd(42)
	want = "\x1b]133;D;42\x07"
	if got != want {
		t.Errorf("ShellOutputEnd(42) = %q, want %q", got, want)
	}
}

func TestP382_ShellOutputEnd_NoCode(t *testing.T) {
	got := ShellOutputEnd(-1)
	want := "\x1b]133;D\x07"
	if got != want {
		t.Errorf("ShellOutputEnd(-1) = %q, want %q", got, want)
	}
}

func TestP382_ShellIntegration(t *testing.T) {
	got := ShellIntegration(0)
	// Should contain all 4 markers
	if len(got) < 4*7 {
		t.Errorf("ShellIntegration too short: %d bytes", len(got))
	}
	// Verify it starts with prompt start
	if got[:len(ShellPromptStart())] != ShellPromptStart() {
		t.Error("ShellIntegration should start with prompt marker")
	}
}

func TestP382_EnableBracketedPaste(t *testing.T) {
	got := EnableBracketedPaste
	want := "\x1b[?2004h"
	if got != want {
		t.Errorf("EnableBracketedPaste = %q, want %q", got, want)
	}
}

func TestP382_DisableBracketedPaste(t *testing.T) {
	got := DisableBracketedPaste
	want := "\x1b[?2004l"
	if got != want {
		t.Errorf("DisableBracketedPaste = %q, want %q", got, want)
	}
}

func TestP382_IsBracketedPaste(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		ok    bool
	}{
		{"valid paste", "\x1b[200~hello world\x1b[201~", "hello world", true},
		{"empty paste", "\x1b[200~\x1b[201~", "", true},
		{"no paste", "hello world", "hello world", false},
		{"partial prefix", "\x1b[200~hello", "\x1b[200~hello", false},
		{"empty input", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := IsBracketedPaste(tt.input)
			if ok != tt.ok {
				t.Errorf("IsBracketedPaste(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if got != tt.want {
				t.Errorf("IsBracketedPaste(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestP382_QueryPaletteColorIndex(t *testing.T) {
	got := QueryPaletteColorIndex(1)
	want := "\x1b]4;1;?\x07"
	if got != want {
		t.Errorf("QueryPaletteColorIndex(1) = %q, want %q", got, want)
	}
}

func TestP382_SetPaletteColor(t *testing.T) {
	got := SetPaletteColor(0, 255, 128, 0)
	// Should be ESC]4;0;rgb:ff/80/00\x07
	if got[:7] != "\x1b]4;0;r" {
		t.Errorf("SetPaletteColor prefix wrong: %q", got[:10])
	}
	if got[len(got)-1] != 0x07 {
		t.Errorf("SetPaletteColor should end with BEL")
	}
}

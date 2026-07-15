package bubbletea

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

// P202: Tests for KeyPressMsg.String() — critical for ggcode compat
// ggcode uses msg.String() comparisons like "ctrl+c", "esc", "alt+up", etc.

func TestKeyPressMsg_StringCtrlC_P202(t *testing.T) {
	k := KeyPressMsg{Rune: 'c', Ctrl: true}
	if k.String() != "ctrl+c" {
		t.Errorf("expected 'ctrl+c', got %q", k.String())
	}
}

func TestKeyPressMsg_StringCtrlX_P202(t *testing.T) {
	k := KeyPressMsg{Rune: 'x', Ctrl: true}
	if k.String() != "ctrl+x" {
		t.Errorf("expected 'ctrl+x', got %q", k.String())
	}
}

func TestKeyPressMsg_StringCtrlR_P202(t *testing.T) {
	k := KeyPressMsg{Rune: 'r', Ctrl: true}
	if k.String() != "ctrl+r" {
		t.Errorf("expected 'ctrl+r', got %q", k.String())
	}
}

func TestKeyPressMsg_StringEsc_P202(t *testing.T) {
	k := KeyPressMsg{Code: term.KeyEscape}
	if k.String() != "esc" {
		t.Errorf("expected 'esc', got %q", k.String())
	}
}

func TestKeyPressMsg_StringEnter_P202(t *testing.T) {
	k := KeyPressMsg{Code: term.KeyEnter}
	if k.String() != "enter" {
		t.Errorf("expected 'enter', got %q", k.String())
	}
}

func TestKeyPressMsg_StringAltUp_P202(t *testing.T) {
	k := KeyPressMsg{Code: term.KeyUp, Alt: true}
	if k.String() != "alt+up" {
		t.Errorf("expected 'alt+up', got %q", k.String())
	}
}

func TestKeyPressMsg_StringAltK_P202(t *testing.T) {
	k := KeyPressMsg{Rune: 'k', Alt: true}
	if k.String() != "alt+k" {
		t.Errorf("expected 'alt+k', got %q", k.String())
	}
}

func TestKeyPressMsg_StringAltJ_P202(t *testing.T) {
	k := KeyPressMsg{Rune: 'j', Alt: true}
	if k.String() != "alt+j" {
		t.Errorf("expected 'alt+j', got %q", k.String())
	}
}

func TestKeyPressMsg_StringAltDown_P202(t *testing.T) {
	k := KeyPressMsg{Code: term.KeyDown, Alt: true}
	if k.String() != "alt+down" {
		t.Errorf("expected 'alt+down', got %q", k.String())
	}
}

func TestKeyPressMsg_StringCtrlShiftC_P202(t *testing.T) {
	// ctrl+shift+c — Ctrl takes priority in prefix
	k := KeyPressMsg{Rune: 'c', Ctrl: true, Shift: true}
	s := k.String()
	// Should start with "ctrl+"
	if len(s) < 5 || s[:5] != "ctrl+" {
		t.Errorf("expected 'ctrl+' prefix, got %q", s)
	}
}

func TestKeyPressMsg_StringPlainRune_P202(t *testing.T) {
	k := KeyPressMsg{Rune: 'a'}
	if k.String() != "a" {
		t.Errorf("expected 'a', got %q", k.String())
	}
}

func TestKeyPressMsg_StringPlainJ_P202(t *testing.T) {
	k := KeyPressMsg{Rune: 'j'}
	if k.String() != "j" {
		t.Errorf("expected 'j', got %q", k.String())
	}
}

func TestKeyPressMsg_StringPlainK_P202(t *testing.T) {
	k := KeyPressMsg{Rune: 'k'}
	if k.String() != "k" {
		t.Errorf("expected 'k', got %q", k.String())
	}
}

func TestKeyPressMsg_StringCtrlBackslash_P202(t *testing.T) {
	k := KeyPressMsg{Rune: '\\', Ctrl: true}
	if k.String() != "ctrl+\\" {
		t.Errorf("expected 'ctrl+\\', got %q", k.String())
	}
}

func TestKeyPressMsg_StringCtrlG_P202(t *testing.T) {
	k := KeyPressMsg{Rune: 'g', Ctrl: true}
	if k.String() != "ctrl+g" {
		t.Errorf("expected 'ctrl+g', got %q", k.String())
	}
}

func TestKeyPressMsg_StringCtrlL_P202(t *testing.T) {
	k := KeyPressMsg{Rune: 'l', Ctrl: true}
	if k.String() != "ctrl+l" {
		t.Errorf("expected 'ctrl+l', got %q", k.String())
	}
}

func TestKeyPressMsg_StringCtrlT_P202(t *testing.T) {
	k := KeyPressMsg{Rune: 't', Ctrl: true}
	if k.String() != "ctrl+t" {
		t.Errorf("expected 'ctrl+t', got %q", k.String())
	}
}

func TestKeyPressMsg_StringShiftTab_P202(t *testing.T) {
	k := KeyPressMsg{Code: term.KeyTab, Shift: true}
	if k.String() != "shift+tab" {
		t.Errorf("expected 'shift+tab', got %q", k.String())
	}
}

func TestKeyPressMsg_StringTab_P202(t *testing.T) {
	k := KeyPressMsg{Code: term.KeyTab}
	if k.String() != "tab" {
		t.Errorf("expected 'tab', got %q", k.String())
	}
}

func TestKeyPressMsg_StringUp_P202(t *testing.T) {
	k := KeyPressMsg{Code: term.KeyUp}
	if k.String() != "up" {
		t.Errorf("expected 'up', got %q", k.String())
	}
}

func TestKeyPressMsg_StringDown_P202(t *testing.T) {
	k := KeyPressMsg{Code: term.KeyDown}
	if k.String() != "down" {
		t.Errorf("expected 'down', got %q", k.String())
	}
}
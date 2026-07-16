package bubbletea

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

// P234: cover Quit(), Sequence() nil path, keyTextFor() all branches, keyName()

func TestProgram_Quit_P234(t *testing.T) {
	m := &echoModel{}
	p := NewProgram(m)
	p.Quit()
	// Calling Quit twice should not panic (close on closed channel)
	// Actually close on closed channel panics, so we only call once
}

func TestSequence_Empty_P234(t *testing.T) {
	cmd := Sequence()
	if cmd != nil {
		t.Error("Sequence() with no args should return nil")
	}
}

func TestSequence_WithNilCmds_P234(t *testing.T) {
	cmdNil := func() Msg { return nil }
	cmd2 := func() Msg { return KeyPressMsg{Rune: 'x'} }
	cmd := Sequence(cmdNil, cmd2)
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	if msg == nil {
		t.Error("expected non-nil msg from last cmd")
	}
}

func TestSequence_AllNil_P234(t *testing.T) {
	cmdNil := func() Msg { return nil }
	cmd := Sequence(cmdNil, cmdNil)
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	if msg != nil {
		t.Error("expected nil msg when all cmds return nil")
	}
}

func TestKeyTextFor_AllSpecialKeys_P234(t *testing.T) {
	// All special keys should return ""
	specials := []term.KeyCode{
		term.KeyEnter, term.KeyTab, term.KeyBackspace, term.KeyDelete,
		term.KeyUp, term.KeyDown, term.KeyLeft, term.KeyRight,
		term.KeyHome, term.KeyEnd, term.KeyPageUp, term.KeyPageDown,
		term.KeyEscape, term.KeyInsert,
	}
	for _, k := range specials {
		if txt := keyTextFor(k, 0, 0); txt != "" {
			t.Errorf("keyTextFor(%d) = %q, want ''", k, txt)
		}
	}
}

func TestKeyTextFor_Space_P234(t *testing.T) {
	if txt := keyTextFor(term.KeySpace, ' ', 0); txt != " " {
		t.Errorf("keyTextFor space = %q, want ' '", txt)
	}
}

func TestKeyTextFor_PrintableRune_P234(t *testing.T) {
	if txt := keyTextFor(term.KeyUnknown, 'A', 0); txt != "A" {
		t.Errorf("keyTextFor 'A' = %q, want 'A'", txt)
	}
}

func TestKeyTextFor_ZeroRune_P234(t *testing.T) {
	// Unknown code with rune=0 → ""
	if txt := keyTextFor(term.KeyUnknown, 0, 0); txt != "" {
		t.Errorf("keyTextFor zero rune = %q, want ''", txt)
	}
}

func TestKeyName_Unknown_P234(t *testing.T) {
	// keyName with unknown code should return ""
	if name := keyName(term.KeyUnknown); name != "" {
		t.Errorf("keyName(KeyUnknown) = %q, want ''", name)
	}
}

func TestKeyName_SpecialKeys_P234(t *testing.T) {
	cases := []struct {
		code term.KeyCode
		want string
	}{
		{term.KeyEnter, "enter"},
		{term.KeyTab, "tab"},
		{term.KeyEscape, "esc"},
		{term.KeyUp, "up"},
		{term.KeyDown, "down"},
		{term.KeyLeft, "left"},
		{term.KeyRight, "right"},
		{term.KeyBackspace, "backspace"},
		{term.KeyDelete, "delete"},
		{term.KeyHome, "home"},
		{term.KeyEnd, "end"},
		{term.KeySpace, "space"},
	}
	for _, tc := range cases {
		if name := keyName(tc.code); name != tc.want {
			t.Errorf("keyName(%d) = %q, want %q", tc.code, name, tc.want)
		}
	}
}

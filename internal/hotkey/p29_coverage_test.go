package hotkey

import (
	"testing"
	"unicode"

	"github.com/topcheer/fluui/internal/term"
)

// P29 coverage tests for hotkey Match and related methods.

func TestP29_KeyCombo_Match_NilKey(t *testing.T) {
	c := KeyCombo{Key: term.KeyEnter}
	if c.Match(nil) {
		t.Error("Match(nil) should return false")
	}
}

func TestP29_KeyCombo_Match_KeyCodeMatch(t *testing.T) {
	c := KeyCombo{Key: term.KeyEnter}
	k := &term.KeyEvent{Key: term.KeyEnter}
	if !c.Match(k) {
		t.Error("should match same key code")
	}
}

func TestP29_KeyCombo_Match_KeyCodeWithMods(t *testing.T) {
	c := KeyCombo{Key: term.KeyEnter, Modifiers: term.ModCtrl}
	k := &term.KeyEvent{Key: term.KeyEnter, Modifiers: term.ModCtrl}
	if !c.Match(k) {
		t.Error("should match key + modifiers")
	}
}

func TestP29_KeyCombo_Match_KeyCodeModMismatch(t *testing.T) {
	c := KeyCombo{Key: term.KeyEnter, Modifiers: term.ModCtrl}
	k := &term.KeyEvent{Key: term.KeyEnter, Modifiers: term.ModAlt}
	if c.Match(k) {
		t.Error("should not match with different modifiers")
	}
}

func TestP29_KeyCombo_Match_RuneWithMods(t *testing.T) {
	c := KeyCombo{Rune: 'f', Modifiers: term.ModCtrl}
	k := &term.KeyEvent{Rune: 'F', Modifiers: term.ModCtrl}
	if !c.Match(k) {
		t.Error("should match rune case-insensitive with modifiers")
	}
}

func TestP29_KeyCombo_Match_RuneNoMods(t *testing.T) {
	c := KeyCombo{Rune: 'g'}
	k := &term.KeyEvent{Rune: 'g'}
	if !c.Match(k) {
		t.Error("should match exact rune with no modifiers")
	}
}

func TestP29_KeyCombo_Match_RuneCaseSensitiveNoMods(t *testing.T) {
	c := KeyCombo{Rune: 'G'}
	k := &term.KeyEvent{Rune: 'g'}
	if c.Match(k) {
		t.Error("should NOT match different case with no modifiers")
	}
}

func TestP29_KeyCombo_Match_SpaceByKey(t *testing.T) {
	c := KeyCombo{Key: term.KeySpace}
	k := &term.KeyEvent{Rune: ' '}
	if !c.Match(k) {
		t.Error("should match KeySpace vs rune space")
	}
}

func TestP29_KeyCombo_Match_NoMatch(t *testing.T) {
	c := KeyCombo{Key: term.KeyEnter}
	k := &term.KeyEvent{Key: term.KeyTab}
	if c.Match(k) {
		t.Error("should not match different keys")
	}
}

func TestP29_KeyCombo_Equal_WithMods(t *testing.T) {
	c1 := KeyCombo{Rune: 'f', Modifiers: term.ModCtrl}
	c2 := KeyCombo{Rune: 'F', Modifiers: term.ModCtrl}
	if !c1.Equal(c2) {
		t.Error("Equal should be case-insensitive with modifiers")
	}
}

func TestP29_KeyCombo_Equal_NoMods(t *testing.T) {
	c1 := KeyCombo{Rune: 'g'}
	c2 := KeyCombo{Rune: 'G'}
	if c1.Equal(c2) {
		t.Error("Equal should be case-sensitive without modifiers")
	}
}

func TestP29_KeyCombo_Equal_DifferentKey(t *testing.T) {
	c1 := KeyCombo{Key: term.KeyEnter}
	c2 := KeyCombo{Key: term.KeyTab}
	if c1.Equal(c2) {
		t.Error("different keys should not be equal")
	}
}

func TestP29_KeyCombo_String_All(t *testing.T) {
	tests := []struct {
		combo KeyCombo
		want  string
	}{
		{KeyCombo{Rune: 'g'}, "g"},
		{KeyCombo{Rune: 'G'}, "G"},
		{KeyCombo{Key: term.KeyEnter}, "Enter"},
		{KeyCombo{Key: term.KeyTab}, "Tab"},
		{KeyCombo{Key: term.KeyEscape}, "Escape"},
	}
	for _, tc := range tests {
		got := tc.combo.String()
		if got != tc.want {
			t.Errorf("KeyCombo.String() = %q, want %q", got, tc.want)
		}
	}
}

func TestP29_KeyCombo_String_CtrlCombo(t *testing.T) {
	c := KeyCombo{Rune: 'f', Modifiers: term.ModCtrl}
	s := c.String()
	if s != "Ctrl+f" {
		t.Errorf("expected 'Ctrl+f', got %q", s)
	}
}

func TestP29_KeyCombo_String_AltCombo(t *testing.T) {
	c := KeyCombo{Rune: 'x', Modifiers: term.ModAlt}
	s := c.String()
	if s != "Alt+x" {
		t.Errorf("expected 'Alt+x', got %q", s)
	}
}

func TestP29_KeyCombo_String_CtrlShiftCombo(t *testing.T) {
	c := KeyCombo{Rune: 'c', Modifiers: term.ModCtrl | term.ModShift}
	s := c.String()
	if s != "Ctrl+Shift+c" {
		t.Errorf("expected 'Ctrl+Shift+c', got %q", s)
	}
}

func TestP29_ParseScope(t *testing.T) {
	// Valid scope
	got, err := ParseScope("global")
	if err != nil || got != ScopeGlobal {
		t.Errorf("ParseScope(\"global\") = %v, %v, want ScopeGlobal", got, err)
	}

	// Invalid scope
	_, err = ParseScope("unknown")
	if err == nil {
		t.Error("expected error for unknown scope")
	}

	// Empty
	_, err = ParseScope("")
	if err != nil {
		t.Error("empty scope may be valid or invalid — checking behavior")
	}
}

func TestP29_keyEventToCombo(t *testing.T) {
	// Test with nil-safe behavior
	k := &term.KeyEvent{Key: term.KeyEnter}
	c := keyEventToCombo(k)
	if c.Key != term.KeyEnter {
		t.Errorf("expected KeyEnter, got %v", c.Key)
	}
}

func TestP29_keyEventToCombo_Rune(t *testing.T) {
	k := &term.KeyEvent{Rune: 'x'}
	c := keyEventToCombo(k)
	if c.Rune != 'x' {
		t.Errorf("expected rune 'x', got %v", c.Rune)
	}
}

func TestP29_keyEventToCombo_WithMods(t *testing.T) {
	k := &term.KeyEvent{Rune: 'f', Modifiers: term.ModCtrl}
	c := keyEventToCombo(k)
	if c.Rune != unicode.ToLower('f') {
		t.Errorf("expected lowercase rune, got %v", c.Rune)
	}
	if c.Modifiers != term.ModCtrl {
		t.Errorf("expected Ctrl modifier, got %v", c.Modifiers)
	}
}

func TestP29_ParseCombo_Single(t *testing.T) {
	c, err := ParseCombo("g")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Rune != 'g' {
		t.Errorf("expected rune 'g', got %v", c.Rune)
	}
}

func TestP29_ParseCombo_CtrlF(t *testing.T) {
	c, err := ParseCombo("Ctrl+F")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Rune != 'f' {
		t.Errorf("expected rune 'f', got %v", c.Rune)
	}
	if c.Modifiers != term.ModCtrl {
		t.Errorf("expected Ctrl, got %v", c.Modifiers)
	}
}

func TestP29_ParseCombo_Invalid(t *testing.T) {
	_, err := ParseCombo("")
	if err == nil {
		t.Error("expected error for empty combo")
	}
}

func TestP29_ParseCombo_ShiftOnly(t *testing.T) {
	c, err := ParseCombo("Shift+Tab")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Key != term.KeyTab {
		t.Errorf("expected KeyTab, got %v", c.Key)
	}
	if c.Modifiers != term.ModShift {
		t.Errorf("expected Shift, got %v", c.Modifiers)
	}
}

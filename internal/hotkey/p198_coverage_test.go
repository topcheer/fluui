package hotkey

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

// P198: Coverage for keyEventToCombo nil path

func TestKeyEventToCombo_Nil_P198(t *testing.T) {
	kc := keyEventToCombo(nil)
	if kc.Key != 0 || kc.Rune != 0 || kc.Modifiers != 0 {
		t.Error("nil key should produce zero KeyCombo")
	}
}

func TestKeyEventToCombo_Valid_P198(t *testing.T) {
	ev := &term.KeyEvent{
		Key:       term.KeyEnter,
		Rune:      'a',
		Modifiers: term.ModCtrl | term.ModShift,
	}
	kc := keyEventToCombo(ev)
	if kc.Key != term.KeyEnter {
		t.Errorf("expected KeyEnter, got %v", kc.Key)
	}
	if kc.Rune != 'a' {
		t.Errorf("expected 'a', got %q", kc.Rune)
	}
	if kc.Modifiers != term.ModCtrl|term.ModShift {
		t.Error("modifiers mismatch")
	}
}
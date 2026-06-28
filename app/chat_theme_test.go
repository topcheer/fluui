package app

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// --- Theme management tests ---

func TestChatApp_ThemeCount(t *testing.T) {
	a := NewChatApp(80, 24)
	count := a.ThemeCount()
	if count < 3 {
		t.Errorf("ThemeCount = %d, want >= 3", count)
	}
}

func TestChatApp_ThemeList(t *testing.T) {
	a := NewChatApp(80, 24)
	list := a.ThemeList()
	if len(list) < 3 {
		t.Errorf("ThemeList length = %d, want >= 3", len(list))
	}
	// Check that all names are non-empty
	for i, name := range list {
		if name == "" {
			t.Errorf("ThemeList[%d] is empty", i)
		}
	}
}

func TestChatApp_ThemeIndex(t *testing.T) {
	a := NewChatApp(80, 24)
	idx := a.ThemeIndex()
	if idx < 0 {
		t.Errorf("ThemeIndex = %d, want >= 0", idx)
	}
}

func TestChatApp_ThemeName(t *testing.T) {
	a := NewChatApp(80, 24)
	name := a.ThemeName()
	if name == "" {
		t.Error("ThemeName should not be empty")
	}
}

func TestChatApp_SetThemeByIndex(t *testing.T) {
	a := NewChatApp(80, 24)
	original := a.ThemeIndex()

	a.SetThemeByIndex(1)
	if a.ThemeIndex() != 1 {
		t.Errorf("ThemeIndex = %d, want 1", a.ThemeIndex())
	}

	// Restore
	a.SetThemeByIndex(original)
}

func TestChatApp_SetThemeByIndex_OutOfRange(t *testing.T) {
	a := NewChatApp(80, 24)
	original := a.ThemeIndex()

	a.SetThemeByIndex(-1)
	a.SetThemeByIndex(999)

	// Index should not change for invalid indices
	if a.ThemeIndex() != original {
		t.Errorf("ThemeIndex changed to %d for invalid indices", a.ThemeIndex())
	}
}

func TestChatApp_SetThemeByName(t *testing.T) {
	a := NewChatApp(80, 24)
	list := a.ThemeList()
	if len(list) < 2 {
		t.Fatal("need at least 2 themes")
	}

	// Set by name should succeed
	if !a.SetThemeByName(list[1]) {
		t.Errorf("SetThemeByName(%q) returned false", list[1])
	}
	if a.ThemeName() != list[1] {
		t.Errorf("ThemeName = %q, want %q", a.ThemeName(), list[1])
	}
}

func TestChatApp_SetThemeByName_NotFound(t *testing.T) {
	a := NewChatApp(80, 24)
	if a.SetThemeByName("NonExistentTheme123") {
		t.Error("SetThemeByName should return false for non-existent theme")
	}
}

func TestChatApp_CycleTheme(t *testing.T) {
	a := NewChatApp(80, 24)
	count := a.ThemeCount()
	original := a.ThemeIndex()

	a.CycleTheme()
	expected := (original + 1) % count
	if a.ThemeIndex() != expected {
		t.Errorf("after CycleTheme: ThemeIndex = %d, want %d", a.ThemeIndex(), expected)
	}
}

func TestChatApp_CycleThemeBack(t *testing.T) {
	a := NewChatApp(80, 24)
	count := a.ThemeCount()
	original := a.ThemeIndex()

	a.CycleThemeBack()
	expected := (original - 1 + count) % count
	if a.ThemeIndex() != expected {
		t.Errorf("after CycleThemeBack: ThemeIndex = %d, want %d", a.ThemeIndex(), expected)
	}
}

func TestChatApp_CycleTheme_WrapAround(t *testing.T) {
	a := NewChatApp(80, 24)
	count := a.ThemeCount()

	// Cycle forward past the last theme
	for i := 0; i < count+1; i++ {
		a.CycleTheme()
	}
	// Should have wrapped around
	idx := a.ThemeIndex()
	if idx < 0 || idx >= count {
		t.Errorf("after wrap: ThemeIndex = %d, should be in [0, %d)", idx, count)
	}
}

func TestChatApp_HandleThemeKey_CtrlRightBracket(t *testing.T) {
	a := NewChatApp(80, 24)
	original := a.ThemeIndex()

	consumed := a.handleThemeKey(&term.KeyEvent{
		Rune:       ']',
		Modifiers:  term.ModCtrl,
	})
	if !consumed {
		t.Error("Ctrl+] should be consumed")
	}
	if a.ThemeIndex() == original {
		t.Error("theme should have changed after Ctrl+]")
	}
}

func TestChatApp_HandleThemeKey_CtrlBackslash(t *testing.T) {
	a := NewChatApp(80, 24)
	original := a.ThemeIndex()

	consumed := a.handleThemeKey(&term.KeyEvent{
		Rune:       '\\',
		Modifiers:  term.ModCtrl,
	})
	if !consumed {
		t.Error("Ctrl+\\ should be consumed")
	}
	if a.ThemeIndex() == original {
		t.Error("theme should have changed after Ctrl+\\")
	}
}

func TestChatApp_HandleThemeKey_NotConsumed(t *testing.T) {
	a := NewChatApp(80, 24)

	// Non-theme keys should not be consumed
	tests := []term.KeyEvent{
		{Rune: 'x', Modifiers: term.ModCtrl}, // Ctrl+X, not a theme key
		{Rune: ']', Modifiers: 0},            // ] without Ctrl
		{Key: term.KeyEnter},                  // Enter key
	}

	for _, key := range tests {
		if a.handleThemeKey(&key) {
			t.Errorf("key %+v should not be consumed by handleThemeKey", key)
		}
	}
}

func TestChatApp_SetThemeByIndex_Multiple(t *testing.T) {
	a := NewChatApp(80, 24)
	count := a.ThemeCount()

	for i := 0; i < count; i++ {
		a.SetThemeByIndex(i)
		if a.ThemeIndex() != i {
			t.Errorf("after SetThemeByIndex(%d): ThemeIndex = %d", i, a.ThemeIndex())
		}
	}
}

func TestChatApp_ThemeList_MatchesBuiltin(t *testing.T) {
	a := NewChatApp(80, 24)
	list := a.ThemeList()
	builtins := theme.Builtin()
	if len(list) != len(builtins) {
		t.Fatalf("ThemeList length = %d, Builtin length = %d", len(list), len(builtins))
	}
	for i, name := range list {
		if name != builtins[i].Name {
			t.Errorf("ThemeList[%d] = %q, want %q", i, name, builtins[i].Name)
		}
	}
}

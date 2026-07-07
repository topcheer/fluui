package component

import (
	"strings"
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

func newKeyEvent(runeVal rune, key term.KeyCode, mods term.ModMask) *term.KeyEvent {
	return &term.KeyEvent{Rune: runeVal, Key: key, Modifiers: mods}
}

// ─── Registration Tests ───

func TestKeybindingMgr_Register(t *testing.T) {
	km := NewKeybindingManager()
	called := false
	err := km.Register("save", "ctrl+s", "Save file", func() bool {
		called = true
		return true
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if km.BindingCount() != 1 {
		t.Errorf("expected 1 binding, got %d", km.BindingCount())
	}

	// Match should trigger handler
	cmd, handled := km.Match(newKeyEvent('s', 0, term.ModCtrl))
	if !handled {
		t.Error("expected handled=true")
	}
	if cmd != "save" {
		t.Errorf("expected command 'save', got %q", cmd)
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestKeybindingMgr_RegisterIn(t *testing.T) {
	km := NewKeybindingManager()
	err := km.RegisterIn("editor", "format", "ctrl+f", "Format code", func() bool { return true })
	if err != nil {
		t.Fatalf("RegisterIn failed: %v", err)
	}

	// In global context, ctrl+f should NOT match (it's in "editor" context)
	_, handled := km.Match(newKeyEvent('f', 0, term.ModCtrl))
	if handled {
		t.Error("ctrl+f should not match in global context")
	}

	// Push editor context
	km.PushContext("editor")
	_, handled = km.Match(newKeyEvent('f', 0, term.ModCtrl))
	if !handled {
		t.Error("ctrl+f should match in editor context")
	}
}

func TestKeybindingMgr_ConflictDetection(t *testing.T) {
	km := NewKeybindingManager()
	err := km.Register("save", "ctrl+s", "Save", func() bool { return true })
	if err != nil {
		t.Fatalf("first Register failed: %v", err)
	}

	// Same keys, same context -> conflict
	err = km.Register("save-as", "ctrl+s", "Save As", func() bool { return true })
	if err == nil {
		t.Error("expected conflict error for duplicate keys")
	}
}

func TestKeybindingMgr_NoConflictDifferentContext(t *testing.T) {
	km := NewKeybindingManager()
	err := km.Register("save", "ctrl+s", "Save", func() bool { return true })
	if err != nil {
		t.Fatalf("first Register failed: %v", err)
	}

	// Same keys, different context -> no conflict
	err = km.RegisterIn("modal", "submit", "ctrl+s", "Submit form", func() bool { return true })
	if err != nil {
		t.Errorf("RegisterIn in different context should succeed: %v", err)
	}
}

func TestKeybindingMgr_Unregister(t *testing.T) {
	km := NewKeybindingManager()
	_ = km.Register("quit", "ctrl+q", "Quit", func() bool { return true })

	if !km.Unregister("quit") {
		t.Error("Unregister returned false for existing binding")
	}
	if km.BindingCount() != 0 {
		t.Errorf("expected 0 bindings after unregister, got %d", km.BindingCount())
	}

	// Unregister non-existent
	if km.Unregister("nonexistent") {
		t.Error("Unregister returned true for non-existent binding")
	}
}

func TestKeybindingMgr_EnableDisable(t *testing.T) {
	km := NewKeybindingManager()
	_ = km.Register("copy", "ctrl+c", "Copy", func() bool { return true })

	km.Disable("copy")
	_, handled := km.Match(newKeyEvent('c', 0, term.ModCtrl))
	if handled {
		t.Error("disabled binding should not match")
	}

	km.Enable("copy")
	_, handled = km.Match(newKeyEvent('c', 0, term.ModCtrl))
	if !handled {
		t.Error("re-enabled binding should match")
	}
}

// ─── Context Management Tests ───

func TestKeybindingMgr_PushPopContext(t *testing.T) {
	km := NewKeybindingManager()
	_ = km.Register("global-quit", "ctrl+q", "Quit", func() bool { return true })
	_ = km.RegisterIn("modal", "modal-esc", "esc", "Close modal", func() bool { return true })

	if km.ActiveContext() != "global" {
		t.Errorf("expected 'global', got %q", km.ActiveContext())
	}

	km.PushContext("modal")
	if km.ActiveContext() != "modal" {
		t.Errorf("expected 'modal', got %q", km.ActiveContext())
	}

	// In modal context, esc should work
	_, handled := km.Match(newKeyEvent(0, term.KeyEscape, 0))
	if !handled {
		t.Error("esc should match in modal context")
	}

	popped := km.PopContext()
	if popped != "modal" {
		t.Errorf("expected 'modal', got %q", popped)
	}
	if km.ActiveContext() != "global" {
		t.Errorf("expected 'global' after pop, got %q", km.ActiveContext())
	}
}

func TestKeybindingMgr_NestedContexts(t *testing.T) {
	km := NewKeybindingManager()
	km.PushContext("level1")
	km.PushContext("level2")
	km.PushContext("level3")

	if km.ActiveContext() != "level3" {
		t.Errorf("expected 'level3', got %q", km.ActiveContext())
	}
	km.PopContext()
	if km.ActiveContext() != "level2" {
		t.Errorf("expected 'level2', got %q", km.ActiveContext())
	}
	km.PopContext()
	if km.ActiveContext() != "level1" {
		t.Errorf("expected 'level1', got %q", km.ActiveContext())
	}
	km.PopContext()
	if km.ActiveContext() != "global" {
		t.Errorf("expected 'global', got %q", km.ActiveContext())
	}
}

func TestKeybindingMgr_PopEmpty(t *testing.T) {
	km := NewKeybindingManager()
	popped := km.PopContext()
	if popped != "" {
		t.Errorf("expected empty string, got %q", popped)
	}
}

// ─── Chord Tests ───

func TestKeybindingMgr_Chord(t *testing.T) {
	km := NewKeybindingManager()
	called := false
	_ = km.Register("save-all", "ctrl+x ctrl+s", "Save all files", func() bool {
		called = true
		return true
	})

	// First key: ctrl+x — should NOT trigger, but set chord prefix
	_, handled := km.Match(newKeyEvent('x', 0, term.ModCtrl))
	if handled {
		t.Error("first key of chord should not be 'handled'")
	}
	if !km.IsChordActive() {
		t.Error("chord should be active after first key")
	}
	if km.ChordPrefix() != "ctrl+x" {
		t.Errorf("expected chord prefix 'ctrl+x', got %q", km.ChordPrefix())
	}

	// Second key: ctrl+s — should trigger handler
	_, handled = km.Match(newKeyEvent('s', 0, term.ModCtrl))
	if !handled {
		t.Error("second key of chord should be handled")
	}
	if called {
		// good
	} else {
		t.Error("chord handler was not called")
	}
	if km.IsChordActive() {
		t.Error("chord should be inactive after completion")
	}
}

func TestKeybindingMgr_ChordCancelled(t *testing.T) {
	km := NewKeybindingManager()
	_ = km.Register("save-all", "ctrl+x ctrl+s", "Save all", func() bool { return true })

	// Start chord
	_, _ = km.Match(newKeyEvent('x', 0, term.ModCtrl))
	if !km.IsChordActive() {
		t.Error("chord should be active")
	}

	// Press non-matching key
	km.Match(newKeyEvent('a', 0, 0))
	if km.IsChordActive() {
		t.Error("chord should be cancelled after non-matching key")
	}
}

func TestKeybindingMgr_CancelChord(t *testing.T) {
	km := NewKeybindingManager()
	_ = km.Register("chord", "ctrl+x ctrl+s", "Chord", func() bool { return true })

	_, _ = km.Match(newKeyEvent('x', 0, term.ModCtrl))
	km.CancelChord()
	if km.IsChordActive() {
		t.Error("chord should be cancelled")
	}
}

// ─── Matching Edge Cases ───

func TestKeybindingMgr_MatchNil(t *testing.T) {
	km := NewKeybindingManager()
	_, handled := km.Match(nil)
	if handled {
		t.Error("nil key should not be handled")
	}
}

func TestKeybindingMgr_MatchNoBindings(t *testing.T) {
	km := NewKeybindingManager()
	_, handled := km.Match(newKeyEvent('a', 0, 0))
	if handled {
		t.Error("unbound key should not be handled")
	}
}

func TestKeybindingMgr_HandleKey(t *testing.T) {
	km := NewKeybindingManager()
	_ = km.Register("test", "a", "Test", func() bool { return true })

	if !km.HandleKey(newKeyEvent('a', 0, 0)) {
		t.Error("HandleKey should return true")
	}
	if km.HandleKey(newKeyEvent('b', 0, 0)) {
		t.Error("HandleKey should return false for unbound key")
	}
}

// ─── Introspection Tests ───

func TestKeybindingMgr_Bindings(t *testing.T) {
	km := NewKeybindingManager()
	_ = km.Register("a", "ctrl+a", "A", func() bool { return true })
	_ = km.Register("b", "ctrl+b", "B", func() bool { return true })

	bindings := km.Bindings()
	if len(bindings) != 2 {
		t.Errorf("expected 2 bindings, got %d", len(bindings))
	}
}

func TestKeybindingMgr_FindByCommand(t *testing.T) {
	km := NewKeybindingManager()
	_ = km.Register("save", "ctrl+s", "Save file", func() bool { return true })

	b := km.FindByCommand("save")
	if b == nil {
		t.Fatal("expected binding for 'save'")
	}
	if b.Keys != "ctrl+s" {
		t.Errorf("expected keys 'ctrl+s', got %q", b.Keys)
	}
	if b.Help != "Save file" {
		t.Errorf("expected help 'Save file', got %q", b.Help)
	}

	if km.FindByCommand("nonexistent") != nil {
		t.Error("expected nil for non-existent command")
	}
}

func TestKeybindingMgr_FindByKeys(t *testing.T) {
	km := NewKeybindingManager()
	_ = km.Register("save", "ctrl+s", "Save", func() bool { return true })

	b := km.FindByKeys("ctrl+s")
	if b == nil {
		t.Fatal("expected binding for ctrl+s")
	}
	if b.Command != "save" {
		t.Errorf("expected command 'save', got %q", b.Command)
	}

	// Also test case-insensitive
	b = km.FindByKeys("Ctrl+S")
	if b == nil {
		t.Error("FindByKeys should be case-insensitive")
	}
}

// ─── Help Text Tests ───

func TestKeybindingMgr_HelpText(t *testing.T) {
	km := NewKeybindingManager()
	_ = km.Register("save", "ctrl+s", "Save file", func() bool { return true })
	_ = km.Register("quit", "ctrl+q", "Quit application", func() bool { return true })
	_ = km.RegisterIn("editor", "format", "ctrl+f", "Format code", func() bool { return true })

	help := km.HelpText()
	if !strings.Contains(help, "ctrl+s") {
		t.Error("help text should contain 'ctrl+s'")
	}
	if !strings.Contains(help, "Save file") {
		t.Error("help text should contain 'Save file'")
	}
}

func TestKeybindingMgr_HelpTextEmpty(t *testing.T) {
	km := NewKeybindingManager()
	help := km.HelpText()
	if help == "" {
		t.Error("help text should not be empty even with no bindings")
	}
}

func TestKeybindingMgr_HelpTextWithContext(t *testing.T) {
	km := NewKeybindingManager()
	_ = km.Register("save", "ctrl+s", "Save", func() bool { return true })
	_ = km.RegisterIn("editor", "format", "ctrl+f", "Format", func() bool { return true })

	km.PushContext("editor")
	help := km.HelpText()
	if !strings.Contains(help, "editor:") {
		t.Error("help text should contain 'editor:' context header")
	}
}

// ─── Conflict Check Tests ───

func TestKeybindingMgr_CheckConflicts_None(t *testing.T) {
	km := NewKeybindingManager()
	_ = km.Register("save", "ctrl+s", "Save", func() bool { return true })
	_ = km.Register("quit", "ctrl+q", "Quit", func() bool { return true })

	conflicts := km.CheckConflicts()
	if len(conflicts) != 0 {
		t.Errorf("expected 0 conflicts, got %d: %v", len(conflicts), conflicts)
	}
}

func TestKeybindingMgr_CheckConflicts_DetectDisabled(t *testing.T) {
	km := NewKeybindingManager()
	_ = km.Register("save", "ctrl+s", "Save", func() bool { return true })
	// Manually add a disabled duplicate (Register would catch it)
	km.mu.Lock()
	km.bindings = append(km.bindings, &KeyBinding{
		Command: "save2",
		Keys:    "ctrl+s",
		Help:    "Save2",
		Context: "global",
		Enabled: false,
	})
	km.mu.Unlock()

	conflicts := km.CheckConflicts()
	// Disabled bindings should NOT show as conflicts
	if len(conflicts) != 0 {
		t.Errorf("disabled binding should not conflict, got %d", len(conflicts))
	}
}

// ─── Key Description Tests ───

func TestKeybindingMgr_KeyEventToDesc(t *testing.T) {
	tests := []struct {
		name string
		k    *term.KeyEvent
		want string
	}{
		{"ctrl+s", newKeyEvent('s', 0, term.ModCtrl), "ctrl+s"},
		{"alt+x", newKeyEvent('x', 0, term.ModAlt), "alt+x"},
		{"shift+a", newKeyEvent('a', 0, term.ModShift), "shift+a"},
		{"ctrl+shift+s", newKeyEvent('s', 0, term.ModCtrl|term.ModShift), "ctrl+shift+s"},
		{"plain_a", newKeyEvent('a', 0, 0), "a"},
		{"enter", newKeyEvent(0, term.KeyEnter, 0), "enter"},
		{"tab", newKeyEvent(0, term.KeyTab, 0), "tab"},
		{"esc", newKeyEvent(0, term.KeyEscape, 0), "esc"},
		{"up", newKeyEvent(0, term.KeyUp, 0), "up"},
		{"down", newKeyEvent(0, term.KeyDown, 0), "down"},
		{"left", newKeyEvent(0, term.KeyLeft, 0), "left"},
		{"right", newKeyEvent(0, term.KeyRight, 0), "right"},
		{"home", newKeyEvent(0, term.KeyHome, 0), "home"},
		{"end", newKeyEvent(0, term.KeyEnd, 0), "end"},
		{"pageup", newKeyEvent(0, term.KeyPageUp, 0), "pageup"},
		{"pagedown", newKeyEvent(0, term.KeyPageDown, 0), "pagedown"},
		{"delete", newKeyEvent(0, term.KeyDelete, 0), "delete"},
		{"backspace", newKeyEvent(0, term.KeyBackspace, 0), "backspace"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := keyEventToDesc(tt.k)
			if got != tt.want {
				t.Errorf("keyEventToDesc() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestKeybindingMgr_NormalizeKeys(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Ctrl+S", "ctrl+s"},
		{"CTRL+X CTRL+S", "ctrl+x ctrl+s"},
		{"ctrl+a", "ctrl+a"},
		{"  ctrl+s  ", "ctrl+s"},
	}
	for _, tt := range tests {
		got := normalizeKeys(tt.input)
		if got != tt.want {
			t.Errorf("normalizeKeys(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestKeybindingMgr_ParseKeyDesc(t *testing.T) {
	tests := []struct {
		desc    string
		wantKey term.KeyCode
		wantMod term.ModMask
		wantRune rune
	}{
		{"ctrl+s", 0, term.ModCtrl, 's'},
		{"alt+x", 0, term.ModAlt, 'x'},
		{"shift+a", 0, term.ModShift, 'a'},
		{"ctrl+shift+a", 0, term.ModCtrl | term.ModShift, 'a'},
		{"enter", term.KeyEnter, 0, 0},
		{"esc", term.KeyEscape, 0, 0},
		{"escape", term.KeyEscape, 0, 0},
		{"up", term.KeyUp, 0, 0},
		{"tab", term.KeyTab, 0, 0},
		{"space", term.KeySpace, 0, ' '},
		{"delete", term.KeyDelete, 0, 0},
		{"del", term.KeyDelete, 0, 0},
		{"backspace", term.KeyBackspace, 0, 0},
		{"a", 0, 0, 'a'},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			keyCode, mods, runeVal := ParseKeyDesc(tt.desc)
			if keyCode != tt.wantKey {
				t.Errorf("keyCode = %d, want %d", keyCode, tt.wantKey)
			}
			if mods != tt.wantMod {
				t.Errorf("mods = %d, want %d", mods, tt.wantMod)
			}
			if runeVal != tt.wantRune {
				t.Errorf("rune = %q, want %q", runeVal, tt.wantRune)
			}
		})
	}
}

// ─── Concurrency Tests ───

func TestKeybindingMgr_Concurrent(t *testing.T) {
	km := NewKeybindingManager()

	var wg sync.WaitGroup
	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = km.Register(
				"cmd"+string(rune('a'+n)),
				"ctrl+"+string(rune('a'+n)),
				"Command "+string(rune('a'+n)),
				func() bool { return true },
			)
		}(i)
	}
	// Concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = km.Bindings()
			_ = km.HelpText()
			_ = km.BindingCount()
		}()
	}
	// Concurrent context operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			km.PushContext("ctx" + string(rune('a'+n)))
			km.PopContext()
		}(i)
	}
	wg.Wait()
	// If we get here without panic/race, test passes
}

func TestKeybindingMgr_ConcurrentMatch(t *testing.T) {
	km := NewKeybindingManager()
	_ = km.Register("save", "ctrl+s", "Save", func() bool { return true })

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			km.Match(newKeyEvent('s', 0, term.ModCtrl))
		}()
	}
	wg.Wait()
}

package hotkey

import (
	"sync"
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/term"
)

// === Manager creation ===

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.Count() != 0 {
		t.Errorf("new manager Count() = %d, want 0", m.Count())
	}
}

func TestManager_SetDefaultGroup(t *testing.T) {
	m := NewManager()
	m.SetDefaultGroup("Navigation")
	err := m.Register("nav.up", MustParseSequence("k"))
	if err != nil {
		t.Fatal(err)
	}
	b, ok := m.Get("nav.up")
	if !ok {
		t.Fatal("binding not found")
	}
	if b.Group != "Navigation" {
		t.Errorf("Group = %q, want 'Navigation'", b.Group)
	}
}

// === Registration ===

func TestRegister_SingleKey(t *testing.T) {
	m := NewManager()
	err := m.Register("test.action", MustParseSequence("Ctrl+F"),
		WithDescription("Find"),
		WithGroup("Search"),
	)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if m.Count() != 1 {
		t.Errorf("Count = %d, want 1", m.Count())
	}
}

func TestRegister_DuplicateAction(t *testing.T) {
	m := NewManager()
	m.Register("test.action", MustParseSequence("a"))
	err := m.Register("test.action", MustParseSequence("b"))
	if err == nil {
		t.Error("expected error for duplicate action")
	}
}

func TestRegister_ConflictSameSequence(t *testing.T) {
	m := NewManager()
	m.Register("action1", MustParseSequence("Ctrl+F"))
	err := m.Register("action2", MustParseSequence("Ctrl+F"))
	if err == nil {
		t.Error("expected conflict error for same key sequence")
	}
}

func TestRegister_ConflictDifferentModifiers(t *testing.T) {
	m := NewManager()
	m.Register("action1", MustParseSequence("Ctrl+F"))
	err := m.Register("action2", MustParseSequence("Alt+F"))
	if err != nil {
		t.Errorf("different modifiers should not conflict: %v", err)
	}
}

func TestRegister_AllowOverride(t *testing.T) {
	m := NewManager()
	m.SetAllowOverride(true)
	m.Register("action1", MustParseSequence("Ctrl+F"))
	err := m.Register("action2", MustParseSequence("Ctrl+F"))
	if err != nil {
		t.Errorf("override should not error: %v", err)
	}
	if _, ok := m.Get("action1"); ok {
		t.Error("action1 should have been overridden")
	}
	b, ok := m.Get("action2")
	if !ok {
		t.Fatal("action2 should exist")
	}
	if b.Action != "action2" {
		t.Errorf("action = %q, want 'action2'", b.Action)
	}
}

func TestRegister_MultiKeySequence(t *testing.T) {
	m := NewManager()
	err := m.Register("goto.top", MustParseSequence("g g"))
	if err != nil {
		t.Fatalf("multi-key register failed: %v", err)
	}
	b, ok := m.Get("goto.top")
	if !ok {
		t.Fatal("binding not found")
	}
	if b.Sequence.Len() != 2 {
		t.Errorf("sequence len = %d, want 2", b.Sequence.Len())
	}
}

func TestUnregister(t *testing.T) {
	m := NewManager()
	m.Register("test.action", MustParseSequence("a"))
	err := m.Unregister("test.action")
	if err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}
	if m.Count() != 0 {
		t.Errorf("Count = %d, want 0", m.Count())
	}
}

func TestUnregister_NotFound(t *testing.T) {
	m := NewManager()
	err := m.Unregister("nonexistent")
	if err == nil {
		t.Error("expected error for unregistering nonexistent action")
	}
}

// === Enable/Disable ===

func TestEnableDisable(t *testing.T) {
	m := NewManager()
	m.Register("test.action", MustParseSequence("a"))

	m.Disable("test.action")
	b, _ := m.Get("test.action")
	if b.Enabled {
		t.Error("should be disabled")
	}

	// Disabled binding should not match
	_, result := m.Match(&term.KeyEvent{Rune: 'a'})
	if result != MatchNone {
		t.Errorf("disabled binding should not match, got %d", result)
	}

	m.Enable("test.action")
	b, _ = m.Get("test.action")
	if !b.Enabled {
		t.Error("should be enabled")
	}
}

func TestRegister_WithDisabled(t *testing.T) {
	m := NewManager()
	m.Register("test.action", MustParseSequence("a"), WithDisabled())
	b, _ := m.Get("test.action")
	if b.Enabled {
		t.Error("should be disabled")
	}
}

// === Matching: single key ===

func TestMatch_SingleKey(t *testing.T) {
	m := NewManager()
	m.Register("insert.a", MustParseSequence("a"))

	action, result := m.Match(&term.KeyEvent{Rune: 'a'})
	if result != MatchComplete {
		t.Errorf("result = %d, want MatchComplete", result)
	}
	if action != "insert.a" {
		t.Errorf("action = %q, want 'insert.a'", action)
	}
}

func TestMatch_CtrlKey(t *testing.T) {
	m := NewManager()
	m.Register("find", MustParseSequence("Ctrl+F"))

	action, result := m.Match(&term.KeyEvent{Rune: 'f', Modifiers: term.ModCtrl})
	if result != MatchComplete {
		t.Errorf("result = %d, want MatchComplete", result)
	}
	if action != "find" {
		t.Errorf("action = %q, want 'find'", action)
	}
}

func TestMatch_NoMatch(t *testing.T) {
	m := NewManager()
	m.Register("find", MustParseSequence("Ctrl+F"))

	action, result := m.Match(&term.KeyEvent{Rune: 'x'})
	if result != MatchNone {
		t.Errorf("result = %d, want MatchNone", result)
	}
	if action != "" {
		t.Errorf("action = %q, want empty", action)
	}
}

func TestMatch_SpecialKey(t *testing.T) {
	m := NewManager()
	m.Register("enter", MustParseSequence("Enter"))

	action, result := m.Match(&term.KeyEvent{Key: term.KeyEnter})
	if result != MatchComplete {
		t.Errorf("result = %d, want MatchComplete", result)
	}
	if action != "enter" {
		t.Errorf("action = %q, want 'enter'", action)
	}
}

// === Matching: multi-key sequences ===

func TestMatch_MultiKeySequence(t *testing.T) {
	m := NewManager()
	m.Register("goto.top", MustParseSequence("g g"),
		WithDescription("Go to top"),
		WithGroup("Navigation"),
	)

	// First 'g' — partial match
	_, result := m.Match(&term.KeyEvent{Rune: 'g'})
	if result != MatchPartial {
		t.Errorf("first key result = %d, want MatchPartial", result)
	}

	// Second 'g' — complete match
	action, result := m.Match(&term.KeyEvent{Rune: 'g'})
	if result != MatchComplete {
		t.Errorf("second key result = %d, want MatchComplete", result)
	}
	if action != "goto.top" {
		t.Errorf("action = %q, want 'goto.top'", action)
	}
}

func TestMatch_MultiKeySequence_WrongKey(t *testing.T) {
	m := NewManager()
	m.Register("goto.top", MustParseSequence("g g"))

	// First 'g' — partial
	m.Match(&term.KeyEvent{Rune: 'g'})

	// Wrong key 'x' — should reset
	action, result := m.Match(&term.KeyEvent{Rune: 'x'})
	if result != MatchNone {
		t.Errorf("wrong key result = %d, want MatchNone", result)
	}
	if action != "" {
		t.Errorf("action = %q, want empty", action)
	}
	if m.HasPending() {
		t.Error("pending should be cleared after wrong key")
	}
}

func TestMatch_MultiKeySequence_ResetPending(t *testing.T) {
	m := NewManager()
	m.Register("goto.top", MustParseSequence("g g"))

	m.Match(&term.KeyEvent{Rune: 'g'})
	if !m.HasPending() {
		t.Error("should have pending")
	}

	m.ResetPending()
	if m.HasPending() {
		t.Error("should not have pending after reset")
	}
}

func TestMatch_MultiKeySequence_Timeout(t *testing.T) {
	m := NewManager()
	m.SetSequenceTimeout(10 * time.Millisecond)
	m.Register("goto.top", MustParseSequence("g g"))

	m.Match(&term.KeyEvent{Rune: 'g'})
	if !m.HasPending() {
		t.Fatal("should have pending")
	}

	time.Sleep(20 * time.Millisecond)

	// After timeout, second 'g' should start a new sequence, not complete
	action, result := m.Match(&term.KeyEvent{Rune: 'g'})
	// The timeout causes the first 'g' to be forgotten.
	// The second 'g' starts a new partial sequence.
	if result != MatchPartial {
		t.Errorf("after timeout result = %d, want MatchPartial (starting new sequence)", result)
	}
	if action != "" {
		t.Errorf("action = %q, want empty", action)
	}
}

func TestMatch_PendingKeys(t *testing.T) {
	m := NewManager()
	m.Register("goto.top", MustParseSequence("g g"))

	m.Match(&term.KeyEvent{Rune: 'g'})

	keys := m.PendingKeys()
	if len(keys) != 1 {
		t.Fatalf("pending keys = %d, want 1", len(keys))
	}
	if keys[0].Rune != 'g' {
		t.Errorf("pending key rune = %q, want 'g'", string(keys[0].Rune))
	}
}

// === Scope ===

func TestScope_String(t *testing.T) {
	tests := []struct {
		scope Scope
		want  string
	}{
		{ScopeGlobal, "Global"},
		{ScopeLocal, "Local"},
		{ScopeModal, "Modal"},
	}
	for _, tc := range tests {
		if got := tc.scope.String(); got != tc.want {
			t.Errorf("Scope(%d).String() = %q, want %q", tc.scope, got, tc.want)
		}
	}
}

func TestParseScope(t *testing.T) {
	tests := []struct {
		input string
		want  Scope
	}{
		{"global", ScopeGlobal},
		{"local", ScopeLocal},
		{"modal", ScopeModal},
		{"", ScopeGlobal},
	}
	for _, tc := range tests {
		got, err := ParseScope(tc.input)
		if err != nil {
			t.Errorf("ParseScope(%q) error: %v", tc.input, err)
		}
		if got != tc.want {
			t.Errorf("ParseScope(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestRegister_WithScope(t *testing.T) {
	m := NewManager()
	m.Register("global.action", MustParseSequence("Ctrl+F"), WithScope(ScopeGlobal))
	m.Register("local.action", MustParseSequence("Ctrl+S"), WithScope(ScopeLocal))

	bindings := m.BindingsByScope(ScopeGlobal)
	if len(bindings) != 1 {
		t.Errorf("global bindings = %d, want 1", len(bindings))
	}
	bindings = m.BindingsByScope(ScopeLocal)
	if len(bindings) != 1 {
		t.Errorf("local bindings = %d, want 1", len(bindings))
	}
}

// === Querying ===

func TestBindings(t *testing.T) {
	m := NewManager()
	m.Register("action1", MustParseSequence("a"))
	m.Register("action2", MustParseSequence("b"))

	all := m.Bindings()
	if len(all) != 2 {
		t.Errorf("Bindings() = %d, want 2", len(all))
	}
}

func TestBindingsByGroup(t *testing.T) {
	m := NewManager()
	m.Register("nav.up", MustParseSequence("k"), WithGroup("Navigation"))
	m.Register("nav.down", MustParseSequence("j"), WithGroup("Navigation"))
	m.Register("edit.copy", MustParseSequence("Ctrl+C"), WithGroup("Edit"))

	navBindings := m.BindingsByGroup("Navigation")
	if len(navBindings) != 2 {
		t.Errorf("Navigation bindings = %d, want 2", len(navBindings))
	}

	editBindings := m.BindingsByGroup("Edit")
	if len(editBindings) != 1 {
		t.Errorf("Edit bindings = %d, want 1", len(editBindings))
	}
}

func TestGroups(t *testing.T) {
	m := NewManager()
	m.Register("nav.up", MustParseSequence("k"), WithGroup("Navigation"))
	m.Register("edit.copy", MustParseSequence("Ctrl+C"), WithGroup("Edit"))

	groups := m.Groups()
	if len(groups) != 2 {
		t.Fatalf("Groups() = %d, want 2", len(groups))
	}
}

func TestGet(t *testing.T) {
	m := NewManager()
	m.Register("test.action", MustParseSequence("Ctrl+F"),
		WithDescription("Find"),
		WithGroup("Search"),
	)

	b, ok := m.Get("test.action")
	if !ok {
		t.Fatal("binding not found")
	}
	if b.Description != "Find" {
		t.Errorf("Description = %q, want 'Find'", b.Description)
	}
	if b.Group != "Search" {
		t.Errorf("Group = %q, want 'Search'", b.Group)
	}
}

func TestGet_NotFound(t *testing.T) {
	m := NewManager()
	_, ok := m.Get("nonexistent")
	if ok {
		t.Error("should not be found")
	}
}

// === Parsing ===

func TestParseCombo_PlainRune(t *testing.T) {
	c, err := ParseCombo("a")
	if err != nil {
		t.Fatal(err)
	}
	if c.Rune != 'a' {
		t.Errorf("Rune = %q, want 'a'", string(c.Rune))
	}
	if c.Modifiers != 0 {
		t.Errorf("Modifiers = %d, want 0", c.Modifiers)
	}
}

func TestParseCombo_CtrlRune(t *testing.T) {
	c, err := ParseCombo("Ctrl+F")
	if err != nil {
		t.Fatal(err)
	}
	if c.Rune != 'f' {
		t.Errorf("Rune = %q, want 'f'", string(c.Rune))
	}
	if c.Modifiers != term.ModCtrl {
		t.Errorf("Modifiers = %d, want %d", c.Modifiers, term.ModCtrl)
	}
}

func TestParseCombo_SpecialKey(t *testing.T) {
	c, err := ParseCombo("Enter")
	if err != nil {
		t.Fatal(err)
	}
	if c.Key != term.KeyEnter {
		t.Errorf("Key = %d, want KeyEnter", c.Key)
	}
}

func TestParseCombo_CtrlSpecialKey(t *testing.T) {
	c, err := ParseCombo("Ctrl+Enter")
	if err != nil {
		t.Fatal(err)
	}
	if c.Key != term.KeyEnter {
		t.Errorf("Key = %d, want KeyEnter", c.Key)
	}
	if c.Modifiers != term.ModCtrl {
		t.Errorf("Modifiers = %d, want ModCtrl", c.Modifiers)
	}
}

func TestParseCombo_Empty(t *testing.T) {
	_, err := ParseCombo("")
	if err == nil {
		t.Error("expected error for empty combo")
	}
}

func TestParseCombo_Invalid(t *testing.T) {
	_, err := ParseCombo("Ctrl+")
	if err == nil {
		t.Error("expected error for 'Ctrl+'")
	}
}

func TestParseCombo_AltModifier(t *testing.T) {
	c, err := ParseCombo("Alt+x")
	if err != nil {
		t.Fatal(err)
	}
	if c.Modifiers != term.ModAlt {
		t.Errorf("Modifiers = %d, want ModAlt", c.Modifiers)
	}
}

func TestParseCombo_ShiftModifier(t *testing.T) {
	c, err := ParseCombo("Shift+Up")
	if err != nil {
		t.Fatal(err)
	}
	if c.Modifiers != term.ModShift {
		t.Errorf("Modifiers = %d, want ModShift", c.Modifiers)
	}
	if c.Key != term.KeyUp {
		t.Errorf("Key = %d, want KeyUp", c.Key)
	}
}

func TestParseCombo_AllModifiers(t *testing.T) {
	c, err := ParseCombo("Ctrl+Alt+Shift+F")
	if err != nil {
		t.Fatal(err)
	}
	expected := term.ModCtrl | term.ModAlt | term.ModShift
	if c.Modifiers != expected {
		t.Errorf("Modifiers = %d, want %d", c.Modifiers, expected)
	}
}

func TestParseSequence_SingleCombo(t *testing.T) {
	seq, err := ParseSequence("Ctrl+F")
	if err != nil {
		t.Fatal(err)
	}
	if !seq.IsSingle() {
		t.Error("should be single")
	}
	if seq.Len() != 1 {
		t.Errorf("Len = %d, want 1", seq.Len())
	}
}

func TestParseSequence_MultiCombo(t *testing.T) {
	seq, err := ParseSequence("g g")
	if err != nil {
		t.Fatal(err)
	}
	if !seq.IsMulti() {
		t.Error("should be multi")
	}
	if seq.Len() != 2 {
		t.Errorf("Len = %d, want 2", seq.Len())
	}
}

func TestParseSequence_Empty(t *testing.T) {
	_, err := ParseSequence("")
	if err == nil {
		t.Error("expected error for empty sequence")
	}
}

func TestMustParseSequence(t *testing.T) {
	seq := MustParseSequence("Ctrl+F")
	if seq.Len() != 1 {
		t.Errorf("Len = %d, want 1", seq.Len())
	}
}

func TestMustParseSequence_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	MustParseSequence("")
}

// === KeyCombo/String ===

func TestKeyCombo_String(t *testing.T) {
	tests := []struct {
		combo KeyCombo
		want  string
	}{
		{KeyCombo{Rune: 'a'}, "a"},
		{KeyCombo{Rune: 'f', Modifiers: term.ModCtrl}, "Ctrl+f"},
		{KeyCombo{Key: term.KeyEnter}, "Enter"},
		{KeyCombo{Key: term.KeyUp, Modifiers: term.ModShift}, "Shift+Up"},
		{KeyCombo{Key: term.KeyF1}, "F1"},
	}
	for _, tc := range tests {
		got := tc.combo.String()
		if got != tc.want {
			t.Errorf("KeyCombo.String() = %q, want %q", got, tc.want)
		}
	}
}

func TestKeySequence_String(t *testing.T) {
	seq := KeySequence{Combos: []KeyCombo{
		{Rune: 'g'},
		{Rune: 'g'},
	}}
	if seq.String() != "g g" {
		t.Errorf("String() = %q, want 'g g'", seq.String())
	}
}

func TestKeyCombo_Equal(t *testing.T) {
	a := KeyCombo{Rune: 'a', Modifiers: term.ModCtrl}
	b := KeyCombo{Rune: 'a', Modifiers: term.ModCtrl}
	c := KeyCombo{Rune: 'b', Modifiers: term.ModCtrl}
	if !a.Equal(b) {
		t.Error("a should equal b")
	}
	if a.Equal(c) {
		t.Error("a should not equal c")
	}
}

func TestKeySequence_Equal(t *testing.T) {
	a := MustParseSequence("g g")
	b := MustParseSequence("g g")
	c := MustParseSequence("g x")
	if !a.Equal(b) {
		t.Error("a should equal b")
	}
	if a.Equal(c) {
		t.Error("a should not equal c")
	}
}

func TestKeySequence_HasPrefix(t *testing.T) {
	full := MustParseSequence("g g t")
	prefix1 := MustParseSequence("g")
	prefix2 := MustParseSequence("g g")
	nonPrefix := MustParseSequence("x")

	if !full.HasPrefix(prefix1) {
		t.Error("'g g t' should have prefix 'g'")
	}
	if !full.HasPrefix(prefix2) {
		t.Error("'g g t' should have prefix 'g g'")
	}
	if full.HasPrefix(nonPrefix) {
		t.Error("'g g t' should not have prefix 'x'")
	}
}

// === Conflict detection ===

func TestHasConflict(t *testing.T) {
	m := NewManager()
	m.Register("action1", MustParseSequence("Ctrl+F"))

	conflict := m.HasConflict(MustParseSequence("Ctrl+F"))
	if conflict == nil {
		t.Error("should detect conflict")
	}
	if conflict.Action != "action1" {
		t.Errorf("conflicting action = %q, want 'action1'", conflict.Action)
	}

	noConflict := m.HasConflict(MustParseSequence("Ctrl+S"))
	if noConflict != nil {
		t.Error("should not detect conflict")
	}
}

// === Serialization ===

func TestExportImport(t *testing.T) {
	m1 := NewManager()
	m1.Register("find", MustParseSequence("Ctrl+F"),
		WithDescription("Find"),
		WithGroup("Search"),
	)
	m1.Register("goto.top", MustParseSequence("g g"),
		WithDescription("Go to top"),
		WithGroup("Navigation"),
	)

	cfg := m1.ExportConfig()
	if len(cfg.Bindings) != 2 {
		t.Fatalf("ExportConfig = %d bindings, want 2", len(cfg.Bindings))
	}

	m2 := NewManager()
	err := m2.ImportConfig(cfg)
	if err != nil {
		t.Fatalf("ImportConfig failed: %v", err)
	}
	if m2.Count() != 2 {
		t.Errorf("Count = %d, want 2", m2.Count())
	}

	// Verify the imported binding works
	action, result := m2.Match(&term.KeyEvent{Rune: 'f', Modifiers: term.ModCtrl})
	if result != MatchComplete {
		t.Errorf("imported binding match result = %d, want MatchComplete", result)
	}
	if action != "find" {
		t.Errorf("imported binding action = %q, want 'find'", action)
	}
}

func TestExportImport_PreservesMetadata(t *testing.T) {
	m1 := NewManager()
	m1.Register("find", MustParseSequence("Ctrl+F"),
		WithDescription("Find in conversation"),
		WithGroup("Search"),
	)

	cfg := m1.ExportConfig()
	m2 := NewManager()
	m2.ImportConfig(cfg)

	b, _ := m2.Get("find")
	if b.Description != "Find in conversation" {
		t.Errorf("Description = %q", b.Description)
	}
	if b.Group != "Search" {
		t.Errorf("Group = %q", b.Group)
	}
}

// === Concurrency ===

func TestManager_ConcurrentAccess(t *testing.T) {
	m := NewManager()

	var wg sync.WaitGroup
	actions := []string{"a", "b", "c", "d", "e"}
	keys := []string{"a", "b", "c", "d", "e"}

	// Concurrent registrations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			m.Register(actions[idx], MustParseSequence(keys[idx]))
		}(i)
	}

	// Concurrent queries
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Bindings()
			m.Groups()
			m.Count()
		}()
	}

	// Concurrent matching
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			m.Match(&term.KeyEvent{Rune: rune(keys[idx%len(keys)][0])})
		}(i)
	}

	wg.Wait()
}

func TestManager_ConcurrentMatch(t *testing.T) {
	m := NewManager()
	m.Register("goto.top", MustParseSequence("g g"))

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Match(&term.KeyEvent{Rune: 'g'})
			m.ResetPending()
		}()
	}
	wg.Wait()
}

// === Real-world scenario ===

func TestManager_RealWorldScenario(t *testing.T) {
	m := NewManager()

	// Navigation
	m.Register("nav.up", MustParseSequence("k"), WithGroup("Navigation"), WithDescription("Move up"))
	m.Register("nav.down", MustParseSequence("j"), WithGroup("Navigation"), WithDescription("Move down"))
	m.Register("goto.top", MustParseSequence("g g"), WithGroup("Navigation"), WithDescription("Go to top"))
	m.Register("goto.bottom", MustParseSequence("G"), WithGroup("Navigation"), WithDescription("Go to bottom"))

	// Search
	m.Register("search.find", MustParseSequence("Ctrl+F"), WithGroup("Search"), WithDescription("Find"))
	m.Register("search.next", MustParseSequence("n"), WithGroup("Search"), WithDescription("Next match"))

	// Edit
	m.Register("edit.copy", MustParseSequence("Ctrl+C"), WithGroup("Edit"), WithDescription("Copy"))
	m.Register("edit.paste", MustParseSequence("Ctrl+V"), WithGroup("Edit"), WithDescription("Paste"))

	// App
	m.Register("app.quit", MustParseSequence("Ctrl+Q"), WithGroup("Application"), WithDescription("Quit"))
	m.Register("app.help", MustParseSequence("?"), WithGroup("Application"), WithDescription("Help"))

	if m.Count() != 10 {
		t.Fatalf("Count = %d, want 10", m.Count())
	}

	// Verify groups
	groups := m.Groups()
	if len(groups) != 4 {
		t.Errorf("Groups = %d, want 4", len(groups))
	}

	// Test matching
	action, result := m.Match(&term.KeyEvent{Rune: 'j'})
	if result != MatchComplete || action != "nav.down" {
		t.Errorf("match 'j' = %q/%d, want 'nav.down'/MatchComplete", action, result)
	}

	// Test multi-key
	_, result = m.Match(&term.KeyEvent{Rune: 'g'})
	if result != MatchPartial {
		t.Errorf("first 'g' = %d, want MatchPartial", result)
	}
	action, result = m.Match(&term.KeyEvent{Rune: 'g'})
	if result != MatchComplete || action != "goto.top" {
		t.Errorf("'g g' = %q/%d, want 'goto.top'/MatchComplete", action, result)
	}

	// Test Ctrl+F
	action, result = m.Match(&term.KeyEvent{Rune: 'f', Modifiers: term.ModCtrl})
	if result != MatchComplete || action != "search.find" {
		t.Errorf("Ctrl+F = %q/%d, want 'search.find'/MatchComplete", action, result)
	}
}

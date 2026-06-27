package app

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

// --- SlashCommandProvider tests ---

func TestSlashCommandProvider_PartialMatch(t *testing.T) {
	p := NewSlashCommandProvider()
	items := p.Candidates("/th")

	if len(items) != 1 {
		t.Fatalf("expected 1 candidate for '/th', got %d: %+v", len(items), items)
	}
	if items[0].Label != "/theme" {
		t.Fatalf("expected '/theme', got %q", items[0].Label)
	}
	if items[0].Insert != "/theme " {
		t.Fatalf("expected insert '/theme ', got %q", items[0].Insert)
	}
	if items[0].Detail != "Change theme (e.g. /theme dracula)" {
		t.Fatalf("unexpected detail: %q", items[0].Detail)
	}
}

func TestSlashCommandProvider_EmptyPrefix(t *testing.T) {
	p := NewSlashCommandProvider()
	items := p.Candidates("/")

	if len(items) < 8 {
		t.Fatalf("expected at least 8 candidates for '/', got %d", len(items))
	}

	// Verify all start with /.
	for _, item := range items {
		if item.Label[0] != '/' {
			t.Fatalf("expected label starting with /, got %q", item.Label)
		}
	}
}

func TestSlashCommandProvider_ExactMatch(t *testing.T) {
	p := NewSlashCommandProvider()
	items := p.Candidates("/clear")

	if len(items) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(items))
	}
	if items[0].Label != "/clear" {
		t.Fatalf("expected '/clear', got %q", items[0].Label)
	}
}

func TestSlashCommandProvider_NoMatch(t *testing.T) {
	p := NewSlashCommandProvider()
	items := p.Candidates("/xyz")

	if len(items) != 0 {
		t.Fatalf("expected 0 candidates for '/xyz', got %d", len(items))
	}
}

func TestSlashCommandProvider_NotSlashPrefix(t *testing.T) {
	p := NewSlashCommandProvider()
	items := p.Candidates("hello")

	if items != nil {
		t.Fatalf("expected nil for non-slash prefix, got %d items", len(items))
	}
}

func TestSlashCommandProvider_AddRemove(t *testing.T) {
	p := NewSlashCommandProvider()

	p.AddCommand("custom", "A custom command")
	items := p.Candidates("/cu")
	if len(items) != 1 {
		t.Fatalf("expected 1 candidate after AddCommand, got %d", len(items))
	}
	if items[0].Label != "/custom" {
		t.Fatalf("expected '/custom', got %q", items[0].Label)
	}

	p.RemoveCommand("custom")
	items = p.Candidates("/cu")
	if len(items) != 0 {
		t.Fatalf("expected 0 candidates after RemoveCommand, got %d", len(items))
	}
}

func TestSlashCommandProvider_Commands(t *testing.T) {
	p := NewSlashCommandProvider()
	cmds := p.Commands()

	if len(cmds) < 8 {
		t.Fatalf("expected at least 8 commands, got %d", len(cmds))
	}

	// Verify sorted.
	for i := 1; i < len(cmds); i++ {
		if cmds[i-1] > cmds[i] {
			t.Fatalf("expected sorted, got %q > %q", cmds[i-1], cmds[i])
		}
	}
}

// --- MentionProvider tests ---

func TestMentionProvider_PartialMatch(t *testing.T) {
	p := NewMentionProvider()
	items := p.Candidates("@f")

	if len(items) != 1 {
		t.Fatalf("expected 1 candidate for '@f', got %d", len(items))
	}
	if items[0].Label != "@file" {
		t.Fatalf("expected '@file', got %q", items[0].Label)
	}
	if items[0].Insert != "@file " {
		t.Fatalf("expected insert '@file ', got %q", items[0].Insert)
	}
}

func TestMentionProvider_EmptyPrefix(t *testing.T) {
	p := NewMentionProvider()
	items := p.Candidates("@")

	if len(items) < 4 {
		t.Fatalf("expected at least 4 candidates for '@', got %d", len(items))
	}
}

func TestMentionProvider_NoMatch(t *testing.T) {
	p := NewMentionProvider()
	items := p.Candidates("@xyz")
	if len(items) != 0 {
		t.Fatalf("expected 0 candidates, got %d", len(items))
	}
}

func TestMentionProvider_NotAtPrefix(t *testing.T) {
	p := NewMentionProvider()
	items := p.Candidates("hello")
	if items != nil {
		t.Fatalf("expected nil for non-@ prefix")
	}
}

func TestMentionProvider_AddRemove(t *testing.T) {
	p := NewMentionProvider()

	p.AddMention("custom", "Custom mention")
	items := p.Candidates("@cu")
	if len(items) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(items))
	}

	p.RemoveMention("custom")
	items = p.Candidates("@cu")
	if len(items) != 0 {
		t.Fatalf("expected 0 after remove, got %d", len(items))
	}
}

// --- MultiProvider tests ---

func TestMultiProvider(t *testing.T) {
	slash := NewSlashCommandProvider()
	mention := NewMentionProvider()
	multi := NewMultiProvider(slash, mention)

	// "/" prefix should only match slash commands.
	items := multi.Candidates("/")
	for _, item := range items {
		if item.Label[0] != '/' {
			t.Fatalf("expected all labels starting with /, got %q", item.Label)
		}
	}

	// "@" prefix should only match mentions.
	items = multi.Candidates("@")
	for _, item := range items {
		if item.Label[0] != '@' {
			t.Fatalf("expected all labels starting with @, got %q", item.Label)
		}
	}
}

// --- CompletionManager tests ---

func TestCompletionManager_Start(t *testing.T) {
	p := NewSlashCommandProvider()
	cm := NewCompletionManager(p)

	if cm.Active() {
		t.Fatal("expected inactive initially")
	}

	ok := cm.Start("/th")
	if !ok {
		t.Fatal("expected Start to succeed")
	}
	if !cm.Active() {
		t.Fatal("expected active after Start")
	}

	item, found := cm.Selected()
	if !found {
		t.Fatal("expected selected item")
	}
	if item.Label != "/theme" {
		t.Fatalf("expected '/theme', got %q", item.Label)
	}
}

func TestCompletionManager_NoCandidates(t *testing.T) {
	p := NewSlashCommandProvider()
	cm := NewCompletionManager(p)

	ok := cm.Start("/xyz")
	if ok {
		t.Fatal("expected Start to fail with no candidates")
	}
	if cm.Active() {
		t.Fatal("expected inactive when no candidates")
	}
}

func TestCompletionManager_CycleNext(t *testing.T) {
	p := NewSlashCommandProvider()
	cm := NewCompletionManager(p)

	cm.Start("/")

	item, ok := cm.Selected()
	if !ok {
		t.Fatal("expected selected item")
	}
	firstLabel := item.Label

	// Cycle forward.
	item, ok = cm.CycleNext()
	if !ok {
		t.Fatal("expected CycleNext to succeed")
	}
	secondLabel := item.Label

	if firstLabel == secondLabel {
		t.Fatal("expected different item after CycleNext")
	}
}

func TestCompletionManager_CyclePrev(t *testing.T) {
	p := NewSlashCommandProvider()
	cm := NewCompletionManager(p)

	cm.Start("/")

	// Cycle forward first.
	cm.CycleNext()

	// Cycle back.
	item, ok := cm.CyclePrev()
	if !ok {
		t.Fatal("expected CyclePrev to succeed")
	}

	// Should be back at first item.
	firstItem, _ := cm.Selected()
	if item.Label != firstItem.Label {
		t.Fatalf("expected to cycle back to first item, got %q vs %q", item.Label, firstItem.Label)
	}
}

func TestCompletionManager_CycleWrap(t *testing.T) {
	p := NewSlashCommandProvider()
	cm := NewCompletionManager(p)

	cm.Start("/")

	// Cycle forward many times to test wrap.
	seen := map[string]bool{}
	for i := 0; i < 20; i++ {
		item, ok := cm.CycleNext()
		if !ok {
			t.Fatal("expected CycleNext to succeed")
		}
		seen[item.Label] = true
	}

	// Should have seen multiple distinct items.
	if len(seen) < 3 {
		t.Fatalf("expected to cycle through multiple items, only saw %d", len(seen))
	}
}

func TestCompletionManager_Accept(t *testing.T) {
	p := NewSlashCommandProvider()
	cm := NewCompletionManager(p)

	cm.Start("/th")

	item, ok := cm.Accept()
	if !ok {
		t.Fatal("expected Accept to succeed")
	}
	if item.Insert != "/theme " {
		t.Fatalf("expected insert '/theme ', got %q", item.Insert)
	}
	if cm.Active() {
		t.Fatal("expected inactive after Accept")
	}
}

func TestCompletionManager_Cancel(t *testing.T) {
	p := NewSlashCommandProvider()
	cm := NewCompletionManager(p)

	cm.Start("/")
	cm.Cancel()

	if cm.Active() {
		t.Fatal("expected inactive after Cancel")
	}
	if cm.Prefix() != "" {
		t.Fatal("expected empty prefix after Cancel")
	}
}

func TestCompletionManager_SetProvider(t *testing.T) {
	p1 := NewSlashCommandProvider()
	cm := NewCompletionManager(p1)

	cm.Start("/")
	p2 := NewMentionProvider()
	cm.SetProvider(p2)

	if cm.Active() {
		t.Fatal("expected inactive after SetProvider")
	}
	if cm.Provider() != p2 {
		t.Fatal("expected provider to be updated")
	}
}

// --- ExtractCompletionPrefix tests ---

func TestExtractCompletionPrefix_Slash(t *testing.T) {
	prefix := ExtractCompletionPrefix("hello /th", len("hello /th"))
	if prefix != "/th" {
		t.Fatalf("expected '/th', got %q", prefix)
	}
}

func TestExtractCompletionPrefix_AtSign(t *testing.T) {
	prefix := ExtractCompletionPrefix("@fi", len("@fi"))
	if prefix != "@fi" {
		t.Fatalf("expected '@fi', got %q", prefix)
	}
}

func TestExtractCompletionPrefix_NoTrigger(t *testing.T) {
	prefix := ExtractCompletionPrefix("hello world", len("hello world"))
	if prefix != "" {
		t.Fatalf("expected empty prefix for plain text, got %q", prefix)
	}
}

func TestExtractCompletionPrefix_EmptyText(t *testing.T) {
	prefix := ExtractCompletionPrefix("", 0)
	if prefix != "" {
		t.Fatalf("expected empty prefix, got %q", prefix)
	}
}

func TestExtractCompletionPrefix_MidWord(t *testing.T) {
	// Cursor at position 5 in "/theme" (after "/them").
	prefix := ExtractCompletionPrefix("/them", 5)
	if prefix != "/them" {
		t.Fatalf("expected '/them', got %q", prefix)
	}
}

// --- InputLine integration tests ---

func TestInputLine_TabTriggersCompletion(t *testing.T) {
	input := NewInputLine("> ")
	cm := NewCompletionManager(NewSlashCommandProvider())
	input.SetCompletionManager(cm)

	// Type "/th".
	for _, r := range "/th" {
		input.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: r})
	}

	// Press Tab.
	consumed := input.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if !consumed {
		t.Fatal("expected Tab to be consumed")
	}

	if !cm.Active() {
		t.Fatal("expected completion to be active after Tab")
	}

	item, ok := cm.Selected()
	if !ok || item.Label != "/theme" {
		t.Fatalf("expected selected '/theme', got %+v", item)
	}
}

func TestInputLine_TabCyclesCandidates(t *testing.T) {
	input := NewInputLine("> ")
	cm := NewCompletionManager(NewSlashCommandProvider())
	input.SetCompletionManager(cm)

	// Type "/".
	input.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: '/'})

	// Tab → start completion.
	input.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	first, _ := cm.Selected()

	// Tab → cycle to next.
	input.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	second, _ := cm.Selected()

	if first.Label == second.Label {
		t.Fatal("expected different item after cycling")
	}
}

func TestInputLine_ShiftTabCyclesBack(t *testing.T) {
	input := NewInputLine("> ")
	cm := NewCompletionManager(NewSlashCommandProvider())
	input.SetCompletionManager(cm)

	input.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: '/'})

	// Tab → start.
	input.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	item1, _ := cm.Selected()

	// Tab → cycle forward.
	input.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	item2, _ := cm.Selected()
	_ = item2

	// Shift+Tab → cycle back.
	input.HandleKey(&term.KeyEvent{Key: term.KeyBacktab})
	item3, _ := cm.Selected()

	if item3.Label != item1.Label {
		t.Fatalf("expected Shift+Tab to cycle back to %q, got %q", item1.Label, item3.Label)
	}
}

func TestInputLine_EscapeCancels(t *testing.T) {
	input := NewInputLine("> ")
	cm := NewCompletionManager(NewSlashCommandProvider())
	input.SetCompletionManager(cm)

	input.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: '/'})
	input.HandleKey(&term.KeyEvent{Key: term.KeyTab})

	if !cm.Active() {
		t.Fatal("expected completion active")
	}

	consumed := input.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Fatal("expected Escape to be consumed")
	}
	if cm.Active() {
		t.Fatal("expected completion to be cancelled")
	}
}

func TestInputLine_TabNoCompletion_NilManager(t *testing.T) {
	input := NewInputLine("> ")

	// No completion manager set.
	consumed := input.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if consumed {
		t.Fatal("expected Tab to be ignored when no completion manager")
	}
}

func TestInputLine_TabNoCandidates(t *testing.T) {
	input := NewInputLine("> ")
	cm := NewCompletionManager(NewSlashCommandProvider())
	input.SetCompletionManager(cm)

	// Type "/xyz" (no matching candidates).
	for _, r := range "/xyz" {
		input.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: r})
	}

	input.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	// Tab with no candidates — should not activate.
	if cm.Active() {
		t.Fatal("expected completion not active with no candidates")
	}
}

func TestInputLine_EnterClearsCompletion(t *testing.T) {
	input := NewInputLine("> ")
	cm := NewCompletionManager(NewSlashCommandProvider())
	input.SetCompletionManager(cm)

	input.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: '/'})
	input.HandleKey(&term.KeyEvent{Key: term.KeyTab})

	if !cm.Active() {
		t.Fatal("expected completion active")
	}

	// Enter should submit and clear.
	input.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	if cm.Active() {
		t.Fatal("expected completion to be cleared after Enter")
	}
}

func TestInputLine_BackspaceClearsCompletion(t *testing.T) {
	input := NewInputLine("> ")
	cm := NewCompletionManager(NewSlashCommandProvider())
	input.SetCompletionManager(cm)

	input.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: '/'})
	input.HandleKey(&term.KeyEvent{Key: term.KeyTab})

	if !cm.Active() {
		t.Fatal("expected completion active")
	}

	input.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})

	if cm.Active() {
		t.Fatal("expected completion to be cleared after Backspace")
	}
}

func TestInputLine_ArrowKeysClearCompletion(t *testing.T) {
	input := NewInputLine("> ")
	cm := NewCompletionManager(NewSlashCommandProvider())
	input.SetCompletionManager(cm)

	input.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: '/'})
	input.HandleKey(&term.KeyEvent{Key: term.KeyTab})

	// Left arrow should cancel completion.
	input.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if cm.Active() {
		t.Fatal("expected completion cleared after Left")
	}
}

func TestInputLine_ApplyCompletion(t *testing.T) {
	input := NewInputLine("> ")
	cm := NewCompletionManager(NewSlashCommandProvider())
	input.SetCompletionManager(cm)

	// Type "/cl".
	for _, r := range "/cl" {
		input.HandleKey(&term.KeyEvent{Key: term.KeyUnknown, Rune: r})
	}

	// Tab → should find /clear.
	input.HandleKey(&term.KeyEvent{Key: term.KeyTab})

	item, ok := cm.Selected()
	if !ok || item.Label != "/clear" {
		t.Fatalf("expected '/clear', got %+v", item)
	}
}

// --- ReplacePrefix utility tests ---

func TestReplacePrefix(t *testing.T) {
	text := "hello /th world"
	cursor := len("hello /th") // cursor right after "/th"

	newText, newCursor := ReplacePrefix(text, cursor, "/th", "/theme ")

	expected := "hello /theme  world"
	if newText != expected {
		t.Fatalf("expected %q, got %q", expected, newText)
	}
	expectedCursor := len("hello /theme ")
	if newCursor != expectedCursor {
		t.Fatalf("expected cursor %d, got %d", expectedCursor, newCursor)
	}
}

func TestReplacePrefix_NoMatch(t *testing.T) {
	text := "hello world"
	cursor := 5

	newText, newCursor := ReplacePrefix(text, cursor, "xxx", "yyy")
	if newText != text || newCursor != cursor {
		t.Fatal("expected no change when prefix doesn't match")
	}
}

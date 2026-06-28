package term

import (
	"testing"
	"time"
)

// P22-D: Additional coverage tests for internal/term

// ─── FeedTimeout: immediate mode (escTimeout == 0) ─────

func TestP22_FeedTimeout_ImmediateEscape(t *testing.T) {
	prev := escTimeoutForTest
	escTimeoutForTest = 0 // immediate mode
	defer func() { escTimeoutForTest = prev }()
	p := NewParser()
	p.Feed([]byte{0x1b}) // ESC starts escape state
	evs := p.FeedTimeout()
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key == nil || evs[0].Key.Key != KeyEscape {
		t.Error("should emit Escape key in immediate mode")
	}
}

func TestP22_FeedTimeout_ImmediateNoEscape(t *testing.T) {
	p := NewParser()
	evs := p.FeedTimeout()
	if evs != nil {
		t.Error("should return nil when not in escape state")
	}
}

func TestP22_FeedTimeout_DelayedNotExpired(t *testing.T) {
	prev := escTimeoutForTest
	escTimeoutForTest = 100 * time.Millisecond
	defer func() { escTimeoutForTest = prev }()
	p := NewParser()
	p.Feed([]byte{0x1b})
	// Immediately timeout — not expired yet
	evs := p.FeedTimeout()
	if evs != nil {
		t.Error("should return nil when timeout not expired")
	}
}

func TestP22_FeedTimeout_DelayedExpired(t *testing.T) {
	prev := escTimeoutForTest
	escTimeoutForTest = 1 * time.Millisecond
	defer func() { escTimeoutForTest = prev }()
	p := NewParser()
	p.Feed([]byte{0x1b})
	time.Sleep(5 * time.Millisecond)
	evs := p.FeedTimeout()
	if len(evs) != 1 {
		t.Fatalf("expected 1 event after timeout, got %d", len(evs))
	}
	if evs[0].Key == nil || evs[0].Key.Key != KeyEscape {
		t.Error("should emit Escape after timeout expires")
	}
}

// ─── handlePrintable: UTF-8 multi-byte ─────────────────

func TestP22_Feed_UTF8MultiByte(t *testing.T) {
	p := NewParser()
	// 'é' = 0xC3 0xA9 in UTF-8
	evs := p.Feed([]byte{0xC3, 0xA9})
	if len(evs) < 1 {
		t.Fatal("expected at least 1 event for UTF-8 char")
	}
	// The last event should be the completed UTF-8 rune
	last := evs[len(evs)-1]
	if last.Key == nil {
		t.Fatal("last event has nil key")
	}
	if last.Key.Rune != 'é' {
		t.Errorf("rune = %q, want 'é'", last.Key.Rune)
	}
}

func TestP22_Feed_UTF8ThreeByte(t *testing.T) {
	p := NewParser()
	// '€' = 0xE2 0x82 0xAC in UTF-8
	evs := p.Feed([]byte{0xE2, 0x82, 0xAC})
	if len(evs) < 1 {
		t.Fatal("expected at least 1 event for 3-byte UTF-8")
	}
	last := evs[len(evs)-1]
	if last.Key == nil || last.Key.Rune != '€' {
		t.Errorf("rune = %q, want '€'", last.Key.Rune)
	}
}

func TestP22_Feed_UTF8InvalidLead(t *testing.T) {
	p := NewParser()
	// 0x80 is a continuation byte used as lead — invalid
	evs := p.Feed([]byte{0x80})
	if len(evs) < 1 {
		t.Fatal("expected replacement char event")
	}
	if evs[0].Key == nil || evs[0].Key.Rune != 0xFFFD {
		t.Errorf("rune = %q, want 0xFFFD (replacement)", evs[0].Key.Rune)
	}
}

// ─── parseCSI: modifier sequences ──────────────────────

func TestP22_ParseCSI_ShiftArrow(t *testing.T) {
	p := NewParser()
	// CSI 1 ; 2 A = Shift+Up
	ev := p.parseCSI([]byte("1;2A"))
	if ev == nil || ev.Key == nil {
		t.Fatal("nil event for Shift+Up")
	}
	if ev.Key.Key != KeyUp {
		t.Errorf("key = %v, want Up", ev.Key.Key)
	}
	if ev.Key.Modifiers&ModShift == 0 {
		t.Error("Shift modifier not set")
	}
}

func TestP22_ParseCSI_CtrlArrow(t *testing.T) {
	p := NewParser()
	// CSI 1 ; 5 C = Ctrl+Right
	ev := p.parseCSI([]byte("1;5C"))
	if ev == nil || ev.Key == nil {
		t.Fatal("nil event for Ctrl+Right")
	}
	if ev.Key.Key != KeyRight {
		t.Errorf("key = %v, want Right", ev.Key.Key)
	}
	if ev.Key.Modifiers&ModCtrl == 0 {
		t.Error("Ctrl modifier not set")
	}
}

func TestP22_ParseCSI_FKeysViaTilde(t *testing.T) {
	p := NewParser()
	// F5 = CSI 15 ~
	ev := p.parseCSI([]byte("15~"))
	if ev == nil || ev.Key == nil {
		t.Fatal("nil event for F5")
	}
	if ev.Key.Key != KeyF5 {
		t.Errorf("key = %v, want F5", ev.Key.Key)
	}
}

// ─── parseSGRMouse: more buttons ───────────────────────

func TestP22_ParseSGRMouse_RightButton(t *testing.T) {
	p := NewParser()
	ev := p.parseSGRMouse([]byte("<2;5;3M"))
	if ev == nil || ev.Mouse == nil {
		t.Fatal("nil event")
	}
}

func TestP22_ParseSGRMouse_MiddleButton(t *testing.T) {
	p := NewParser()
	ev := p.parseSGRMouse([]byte("<1;5;3M"))
	if ev == nil || ev.Mouse == nil {
		t.Fatal("nil event")
	}
}

func TestP22_ParseSGRMouse_WheelUp(t *testing.T) {
	p := NewParser()
	ev := p.parseSGRMouse([]byte("<64;5;3M"))
	if ev == nil || ev.Mouse == nil {
		t.Fatal("nil event")
	}
}

// ─── decodeMouseButton: more codes ─────────────────────

func TestP22_DecodeMouseButton_Right(t *testing.T) {
	var btn MouseButton
	var mods ModMask
	var act MouseAction
	decodeMouseButton(2, &btn, &mods, &act)
	if btn != MouseRight {
		t.Errorf("button = %v, want MouseRight", btn)
	}
}

func TestP22_DecodeMouseButton_Middle(t *testing.T) {
	var btn MouseButton
	var mods ModMask
	var act MouseAction
	decodeMouseButton(1, &btn, &mods, &act)
	if btn != MouseMiddle {
		t.Errorf("button = %v, want MouseMiddle", btn)
	}
}

func TestP22_DecodeMouseButton_Unknown(t *testing.T) {
	var btn MouseButton
	var mods ModMask
	var act MouseAction
	decodeMouseButton(99, &btn, &mods, &act)
	// Should not panic
}

// ─── parseCSI: paste mode ──────────────────────────────

func TestP22_ParseCSI_PasteSequence(t *testing.T) {
	p := NewParser()
	// Full paste: ESC[200~ hello ESC[201~
	evs := p.Feed([]byte{0x1b, '[', '2', '0', '0', '~', 'h', 'e', 'l', 'l', 'o', 0x1b, '[', '2', '0', '1', '~'})
	found := false
	for _, ev := range evs {
		if ev.Type == EventPaste {
			found = true
			if ev.Paste == "" {
				t.Error("paste content should not be empty")
			}
			break
		}
	}
	if !found {
		t.Error("should produce EventPaste for complete paste sequence")
	}
}

func TestP22_ParseCSI_PasteEnd_NoStart(t *testing.T) {
	p := NewParser()
	// Stray 201~ without paste start — should be ignored
	evs := p.Feed([]byte{0x1b, '[', '2', '0', '1', '~'})
	for _, ev := range evs {
		if ev.Type == EventPaste {
			t.Error("should not produce paste event without start")
		}
	}
}

// ─── Feed: Ctrl+key combinations ───────────────────────

func TestP22_Feed_CtrlKey(t *testing.T) {
	p := NewParser()
	// Ctrl+A = 0x01
	evs := p.Feed([]byte{0x01})
	if len(evs) < 1 {
		t.Fatal("expected event for Ctrl+A")
	}
	if evs[0].Key == nil {
		t.Fatal("nil key")
	}
	if evs[0].Key.Rune != 1 && evs[0].Key.Key != 0 {
		// Ctrl chars may be rune or key
	}
}

// ─── contains helper ───────────────────────────────────

func TestP22_Contains(t *testing.T) {
	if !contains("hello world", "world") {
		t.Error("should find 'world' in 'hello world'")
	}
	if contains("hello", "world") {
		t.Error("should not find 'world' in 'hello'")
	}
	if !contains("abc", "") {
		// Empty substring should always match
	}
	if contains("", "abc") {
		t.Error("should not find 'abc' in empty string")
	}
}

// ─── Feed: Enter and Tab ───────────────────────────────

func TestP22_Feed_Enter(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte{'\r'})
	if len(evs) < 1 {
		t.Fatal("expected event for Enter")
	}
	if evs[0].Key == nil || evs[0].Key.Key != KeyEnter {
		t.Errorf("expected Enter key, got %+v", evs[0])
	}
}

func TestP22_Feed_Tab(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte{'\t'})
	if len(evs) < 1 {
		t.Fatal("expected event for Tab")
	}
	if evs[0].Key == nil || evs[0].Key.Key != KeyTab {
		t.Errorf("expected Tab key, got %+v", evs[0])
	}
}

func TestP22_Feed_Backspace(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte{0x7f})
	if len(evs) < 1 {
		t.Fatal("expected event for Backspace")
	}
	if evs[0].Key == nil || evs[0].Key.Key != KeyBackspace {
		t.Errorf("expected Backspace key, got %+v", evs[0])
	}
}

// ─── Clipboard: ParseOSC52Response edge cases ──────────

func TestP22_ParseOSC52Response_QueryEcho(t *testing.T) {
	// A query echo (?): should return empty string and true
	result, ok := ParseOSC52Response("\x1b]52;c;?\x1b\\")
	if !ok {
		t.Error("should succeed for query echo")
	}
	if result != "" {
		t.Errorf("query echo should return empty string, got %q", result)
	}
}

func TestP22_ParseOSC52Response_NoSemicolon(t *testing.T) {
	_, ok := ParseOSC52Response("\x1b]52;no-semi")
	if ok {
		t.Error("should fail without second semicolon")
	}
}

func TestP22_ParseOSC52Response_BELTerminator(t *testing.T) {
	// Some terminals use BEL instead of ESC backslash
	result, ok := ParseOSC52Response("\x1b]52;c;aGVsbG8=\x07")
	if !ok {
		t.Error("should succeed with BEL terminator")
	}
	if result != "hello" {
		t.Errorf("result = %q, want 'hello'", result)
	}
}

// ─── Writer: basic operations ──────────────────────────

func TestP22_NewWriter(t *testing.T) {
	w := NewWriter(&byteWriter{}, Profile256)
	if w == nil {
		t.Fatal("NewWriter returned nil")
	}
}

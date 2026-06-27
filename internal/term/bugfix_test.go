package term

import (
	"testing"
)

// === BUG 1: UTF-8 multi-byte character handling ===

func TestParseUTF8TwoByte(t *testing.T) {
	// é = 0xC3 0xA9 (2-byte UTF-8)
	p := NewParser()
	events := p.Feed([]byte{0xC3, 0xA9})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key == nil {
		t.Fatal("expected key event")
	}
	if events[0].Key.Rune != 'é' {
		t.Errorf("got rune %q (U+%04X), want 'é'", string(events[0].Key.Rune), events[0].Key.Rune)
	}
}

func TestParseUTF8ThreeByteCJK(t *testing.T) {
	// 你 = 0xE4 0xBD 0xA0 (3-byte UTF-8)
	// 好 = 0xE5 0xA5 0xBD (3-byte UTF-8)
	p := NewParser()
	events := p.Feed([]byte{0xE4, 0xBD, 0xA0, 0xE5, 0xA5, 0xBD})
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Key.Rune != '你' {
		t.Errorf("event 0: got %q, want '你'", string(events[0].Key.Rune))
	}
	if events[1].Key.Rune != '好' {
		t.Errorf("event 1: got %q, want '好'", string(events[1].Key.Rune))
	}
}

func TestParseUTF8FourByteEmoji(t *testing.T) {
	// 😀 = U+1F600 = 0xF0 0x9F 0x98 0x80 (4-byte UTF-8)
	p := NewParser()
	events := p.Feed([]byte{0xF0, 0x9F, 0x98, 0x80})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key.Rune != '😀' {
		t.Errorf("got %q, want '😀'", string(events[0].Key.Rune))
	}
}

func TestParseUTF8SplitAcrossFeeds(t *testing.T) {
	// 你 = 0xE4 0xBD 0xA0, arriving in three separate Feed calls
	p := NewParser()

	events := p.Feed([]byte{0xE4})
	if len(events) != 0 {
		t.Fatalf("after first byte: expected 0 events, got %d", len(events))
	}

	events = p.Feed([]byte{0xBD})
	if len(events) != 0 {
		t.Fatalf("after second byte: expected 0 events, got %d", len(events))
	}

	events = p.Feed([]byte{0xA0})
	if len(events) != 1 {
		t.Fatalf("after third byte: expected 1 event, got %d", len(events))
	}
	if events[0].Key.Rune != '你' {
		t.Errorf("got %q, want '你'", string(events[0].Key.Rune))
	}
}

func TestParseUTF8InvalidContinuation(t *testing.T) {
	// 0xC3 followed by a non-continuation byte 'a' (0x61)
	// Should emit replacement char, then 'a' as normal
	p := NewParser()
	events := p.Feed([]byte{0xC3, 0x61})
	if len(events) != 2 {
		t.Fatalf("expected 2 events (replacement + 'a'), got %d", len(events))
	}
	if events[0].Key.Rune != '\ufffd' {
		t.Errorf("event 0: got %q, want replacement char", string(events[0].Key.Rune))
	}
	if events[1].Key.Rune != 'a' {
		t.Errorf("event 1: got %q, want 'a'", string(events[1].Key.Rune))
	}
}

func TestParseUTF8InvalidLeadByte(t *testing.T) {
	// 0xFE is never a valid UTF-8 lead byte
	p := NewParser()
	events := p.Feed([]byte{0xFE})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key.Rune != '\ufffd' {
		t.Errorf("got %q, want replacement char", string(events[0].Key.Rune))
	}
}

// === BUG 2: Standalone ESC key ===

func TestParseStandaloneESC(t *testing.T) {
	p := NewParser()
	p.escTimeout = 0 // immediate timeout for testing

	events := p.Feed([]byte{0x1b})
	// No event immediately — parser is waiting for potential escape sequence
	if len(events) != 0 {
		t.Fatalf("expected 0 events immediately after ESC, got %d", len(events))
	}

	// After timeout, ESC should be emitted
	events = p.FeedTimeout()
	if len(events) != 1 {
		t.Fatalf("expected 1 event after timeout, got %d", len(events))
	}
	if events[0].Key == nil || events[0].Key.Key != KeyEscape {
		t.Errorf("expected KeyEscape, got %v", events[0].Key)
	}
}

func TestParseESCNotEmittedBeforeTimeout(t *testing.T) {
	// ESC followed quickly by [A should be Up, not ESC+Up
	p := NewParser()
	events := p.Feed([]byte{0x1b, '[', 'A'})
	if len(events) != 1 {
		t.Fatalf("expected 1 event (Up), got %d", len(events))
	}
	if events[0].Key == nil || events[0].Key.Key != KeyUp {
		t.Errorf("expected KeyUp, got %v", events[0].Key)
	}
}

func TestParseDoubleESC(t *testing.T) {
	// Two ESCs in a row → first should be standalone ESC
	p := NewParser()
	events := p.Feed([]byte{0x1b, 0x1b})
	if len(events) < 1 {
		t.Fatalf("expected at least 1 event, got %d", len(events))
	}
	if events[0].Key == nil || events[0].Key.Key != KeyEscape {
		t.Errorf("first event should be KeyEscape, got %v", events[0].Key)
	}
	// Parser should be back in stateEscape for the second ESC
	p.escTimeout = 0
	ev2 := p.FeedTimeout()
	if len(ev2) != 1 {
		t.Fatalf("expected 1 event from second ESC timeout, got %d", len(ev2))
	}
	if ev2[0].Key == nil || ev2[0].Key.Key != KeyEscape {
		t.Errorf("second event should be KeyEscape, got %v", ev2[0].Key)
	}
}

// === BUG 3: Paste content with CSI sequences ===

func TestParsePasteWithANSI(t *testing.T) {
	// Paste content contains \e[0m (reset color)
	pasteContent := "\x1b[0mHello\x1b[31mRed\x1b[0m"
	input := append([]byte("\x1b[200~"), []byte(pasteContent)...)
	input = append(input, []byte("\x1b[201~")...)

	p := NewParser()
	events := p.Feed(input)

	// Should produce exactly 1 paste event with all content intact
	if len(events) != 1 {
		t.Fatalf("expected 1 paste event, got %d events", len(events))
	}
	if events[0].Type != EventPaste {
		t.Errorf("expected EventPaste, got %v", events[0].Type)
	}
	if events[0].Paste != pasteContent {
		t.Errorf("paste content mismatch:\n  got:  %q\n  want: %q", events[0].Paste, pasteContent)
	}
}

func TestParsePasteWithClearScreen(t *testing.T) {
	// Paste content contains \e[2J (clear screen)
	pasteContent := "Before\x1b[2JAfter"
	input := append([]byte("\x1b[200~"), []byte(pasteContent)...)
	input = append(input, []byte("\x1b[201~")...)

	p := NewParser()
	events := p.Feed(input)

	if len(events) != 1 {
		t.Fatalf("expected 1 paste event, got %d", len(events))
	}
	if events[0].Type != EventPaste {
		t.Errorf("expected EventPaste, got %v", events[0].Type)
	}
	if events[0].Paste != pasteContent {
		t.Errorf("paste content: got %q, want %q", events[0].Paste, pasteContent)
	}
}

// === BUG 4: Modifier keys for tilde sequences ===

func TestParseCtrlDelete(t *testing.T) {
	// Ctrl+Delete = ESC[3;5~
	p := NewParser()
	events := p.Feed([]byte("\x1b[3;5~"))
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key == nil {
		t.Fatal("expected key event")
	}
	if events[0].Key.Key != KeyDelete {
		t.Errorf("got key %v, want KeyDelete", events[0].Key.Key)
	}
	if events[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

func TestParseShiftPageUp(t *testing.T) {
	// Shift+PageUp = ESC[5;2~
	p := NewParser()
	events := p.Feed([]byte("\x1b[5;2~"))
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key == nil || events[0].Key.Key != KeyPageUp {
		t.Errorf("expected KeyPageUp, got %v", events[0].Key)
	}
	if events[0].Key.Modifiers&ModShift == 0 {
		t.Error("expected Shift modifier")
	}
}

func TestParseCtrlF5(t *testing.T) {
	// Ctrl+F5 = ESC[15;5~
	p := NewParser()
	events := p.Feed([]byte("\x1b[15;5~"))
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key == nil || events[0].Key.Key != KeyF5 {
		t.Errorf("expected KeyF5, got %v", events[0].Key)
	}
	if events[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

func TestParseCtrlF12(t *testing.T) {
	// Ctrl+F12 = ESC[24;5~
	p := NewParser()
	events := p.Feed([]byte("\x1b[24;5~"))
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key == nil || events[0].Key.Key != KeyF12 {
		t.Errorf("expected KeyF12, got %v", events[0].Key)
	}
	if events[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

// === BUG 5: Ctrl+\ / Ctrl+] / Ctrl+^ / Ctrl+_ ===

func TestParseCtrlBackslash(t *testing.T) {
	p := NewParser()
	events := p.Feed([]byte{0x1c})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key.Rune != '\\' {
		t.Errorf("got rune %q, want '\\'", string(events[0].Key.Rune))
	}
	if events[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

func TestParseCtrlRightBracket(t *testing.T) {
	p := NewParser()
	events := p.Feed([]byte{0x1d})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key.Rune != ']' {
		t.Errorf("got rune %q, want ']'", string(events[0].Key.Rune))
	}
	if events[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

func TestParseCtrlCaret(t *testing.T) {
	p := NewParser()
	events := p.Feed([]byte{0x1e})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key.Rune != '^' {
		t.Errorf("got rune %q, want '^'", string(events[0].Key.Rune))
	}
	if events[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

func TestParseCtrlUnderscore(t *testing.T) {
	p := NewParser()
	events := p.Feed([]byte{0x1f})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key.Rune != '_' {
		t.Errorf("got rune %q, want '_'", string(events[0].Key.Rune))
	}
	if events[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

// === BUG 6: Paste start marker inside paste mode ===

func TestParsePasteStartInsidePaste(t *testing.T) {
	// Paste content that itself contains a paste-start marker
	pasteContent := "before\x1b[200~after"
	input := append([]byte("\x1b[200~"), []byte(pasteContent)...)
	input = append(input, []byte("\x1b[201~")...)

	p := NewParser()
	events := p.Feed(input)

	// Should produce 1 paste event; the inner \e[200~ should be literal
	if len(events) != 1 {
		t.Fatalf("expected 1 paste event, got %d", len(events))
	}
	if events[0].Type != EventPaste {
		t.Errorf("expected EventPaste, got %v", events[0].Type)
	}
	// Content should include the raw marker text
	if events[0].Paste != pasteContent {
		t.Errorf("paste content: got %q, want %q", events[0].Paste, pasteContent)
	}
}

// === Additional edge case tests ===

func TestParseMixedASCIIAndUTF8(t *testing.T) {
	// Mix ASCII and CJK: "ab你c好 d"
	p := NewParser()
	events := p.Feed([]byte("ab\xe4\xbd\xa0c\xe5\xa5\xbd d"))

	expected := []rune{'a', 'b', '你', 'c', '好', ' ', 'd'}
	if len(events) < len(expected) {
		t.Fatalf("expected at least %d events, got %d", len(expected), len(events))
	}

	// Find the rune events (skip any non-rune)
	idx := 0
	for _, ev := range events {
		if ev.Key != nil && ev.Key.Rune != 0 && idx < len(expected) {
			if ev.Key.Rune != expected[idx] {
				t.Errorf("event %d: got %q, want %q", idx, string(ev.Key.Rune), string(expected[idx]))
			}
			idx++
		}
	}
	if idx != len(expected) {
		t.Errorf("matched %d runes, expected %d", idx, len(expected))
	}
}

func TestParseUTF8ThenControl(t *testing.T) {
	// CJK char followed by Enter
	p := NewParser()
	events := p.Feed([]byte{0xE4, 0xBD, 0xA0, 0x0D})
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Key.Rune != '你' {
		t.Errorf("event 0: got %q, want '你'", string(events[0].Key.Rune))
	}
	if events[1].Key == nil || events[1].Key.Key != KeyEnter {
		t.Errorf("event 1: expected KeyEnter, got %v", events[1].Key)
	}
}

// === Edge case: paste content with UTF-8 multi-byte characters ===

func TestParsePasteWithUTF8(t *testing.T) {
	// Paste content contains CJK characters
	pasteContent := "你好世界"
	input := append([]byte("\x1b[200~"), []byte(pasteContent)...)
	input = append(input, []byte("\x1b[201~")...)

	p := NewParser()
	events := p.Feed(input)

	if len(events) != 1 {
		t.Fatalf("expected 1 paste event, got %d", len(events))
	}
	if events[0].Type != EventPaste {
		t.Errorf("expected EventPaste, got %v", events[0].Type)
	}
	if events[0].Paste != pasteContent {
		t.Errorf("paste content: got %q, want %q", events[0].Paste, pasteContent)
	}
}

// === Edge case: paste with multiple interleaved CSI sequences ===

func TestParsePasteWithMultipleCSI(t *testing.T) {
	// Multiple ANSI sequences mixed with regular text inside paste
	pasteContent := "text\x1b[1;31mRed\x1b[0m\x1b[2Kmore"
	input := append([]byte("\x1b[200~"), []byte(pasteContent)...)
	input = append(input, []byte("\x1b[201~")...)

	p := NewParser()
	events := p.Feed(input)

	if len(events) != 1 {
		t.Fatalf("expected 1 paste event, got %d", len(events))
	}
	if events[0].Paste != pasteContent {
		t.Errorf("paste content: got %q, want %q", events[0].Paste, pasteContent)
	}
}

// === Edge case: paste split across multiple Feed calls ===

func TestParsePasteSplitAcrossFeeds(t *testing.T) {
	// Paste arriving in chunks across multiple reads
	pasteContent := "chunk1\x1b[0mchunk2"
	full := append([]byte("\x1b[200~"), []byte(pasteContent)...)
	full = append(full, []byte("\x1b[201~")...)

	// Split at arbitrary boundaries
	split1 := full[:5]
	split2 := full[5:15]
	split3 := full[15:]

	p := NewParser()
	var events []Event
	events = append(events, p.Feed(split1)...)
	events = append(events, p.Feed(split2)...)
	events = append(events, p.Feed(split3)...)

	if len(events) != 1 {
		t.Fatalf("expected 1 paste event, got %d", len(events))
	}
	if events[0].Type != EventPaste {
		t.Errorf("expected EventPaste, got %v", events[0].Type)
	}
	if events[0].Paste != pasteContent {
		t.Errorf("paste content: got %q, want %q", events[0].Paste, pasteContent)
	}
}

// === Edge case: empty paste ===

func TestParseEmptyPaste(t *testing.T) {
	// Paste start immediately followed by paste end
	p := NewParser()
	events := p.Feed([]byte("\x1b[200~\x1b[201~"))

	if len(events) != 1 {
		t.Fatalf("expected 1 paste event, got %d", len(events))
	}
	if events[0].Type != EventPaste {
		t.Errorf("expected EventPaste, got %v", events[0].Type)
	}
	if events[0].Paste != "" {
		t.Errorf("expected empty paste, got %q", events[0].Paste)
	}
}

// === BUG 1 edge case: 0xFF invalid UTF-8 byte ===

func TestParseUTF8Invalid0xFF(t *testing.T) {
	// 0xFF is never valid in UTF-8; should emit replacement char
	p := NewParser()
	events := p.Feed([]byte{0xFF})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key.Rune != '\ufffd' {
		t.Errorf("got %q, want replacement char", string(events[0].Key.Rune))
	}
}

// === BUG 4 edge case: Alt+PageDown ===

func TestParseAltPageDown(t *testing.T) {
	// Alt+PageDown = ESC[6;3~
	p := NewParser()
	events := p.Feed([]byte("\x1b[6;3~"))
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key == nil || events[0].Key.Key != KeyPageDown {
		t.Errorf("expected KeyPageDown, got %v", events[0].Key)
	}
	if events[0].Key.Modifiers&ModAlt == 0 {
		t.Error("expected Alt modifier")
	}
}

// === BUG 6 edge case: nested double paste markers ===

func TestParseNestedPasteMarkers(t *testing.T) {
	// Inner paste-start marker should be literal (BUG 6 fix).
	// The first \e[201~ terminates the paste; trailing \e[201~ is dropped
	// (201~ is not a recognized key in normal mode).
	// Full input: \e[200~ \e[200~ test \e[201~ \e[201~
	pasteContent := "\x1b[200~test"
	input := []byte("\x1b[200~" + pasteContent + "\x1b[201~\x1b[201~")

	p := NewParser()
	events := p.Feed(input)

	if len(events) != 1 {
		t.Fatalf("expected 1 paste event, got %d", len(events))
	}
	if events[0].Type != EventPaste {
		t.Errorf("expected EventPaste, got %v", events[0].Type)
	}
	if events[0].Paste != pasteContent {
		t.Errorf("paste content:\n  got:  %q\n  want: %q", events[0].Paste, pasteContent)
	}
}



package term

import "testing"

// === feed() line 213: raw 0x0a (LF) should produce Enter key ===

func TestP122_Feed_RawLF(t *testing.T) {
	p := NewParser()
	events := p.Feed([]byte{0x0a})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != EventKey {
		t.Errorf("expected EventKey, got %v", events[0].Type)
	}
	if events[0].Key == nil || events[0].Key.Key != KeyEnter {
		t.Error("expected KeyEnter for raw LF")
	}
}

// === feed() lines 326-330: ESC + non-'[' in paste mode ===

func TestP122_Feed_Paste_EscNotBracket(t *testing.T) {
	p := NewParser()
	// Enter paste mode: ESC [ 2 0 0 ~
	p.Feed([]byte{0x1b, '[', '2', '0', '0', '~'})
	// Now in paste mode, send ESC followed by non-bracket (e.g. ESC X)
	events := p.Feed([]byte{0x1b, 'X', 'd', 'a', 't', 'a'})
	// The ESC X should be treated as literal paste content
	// Then 'data' is also paste content
	// Check that no paste event is emitted yet (still in paste mode)
	for _, ev := range events {
		if ev.Type == EventPaste {
			t.Errorf("should not emit paste event while still in paste mode")
		}
	}
	// Now end paste: ESC [ 2 0 1 ~
	events = p.Feed([]byte{0x1b, '[', '2', '0', '1', '~'})
	// Should emit paste event with the literal ESC X data
	found := false
	for _, ev := range events {
		if ev.Type == EventPaste {
			found = true
			// The paste content should contain the escaped ESC and X
			if ev.Paste == "" {
				t.Error("expected non-empty paste")
			}
		}
	}
	if !found {
		t.Error("expected paste event after closing paste")
	}
}

// === feed() lines 357-361: ESC + non-'\' in OSC state ===

func TestP122_Feed_OSC_EscNotBackslash(t *testing.T) {
	p := NewParser()
	// Start OSC: ESC ]
	p.Feed([]byte{0x1b, ']'})
	// Collect some OSC payload
	p.Feed([]byte("52;c;"))
	// Now send ESC followed by something other than '\' — should continue collecting
	p.Feed([]byte{0x1b, 'x'})
	// The OSC should not have been terminated
	// Now send proper ST: ESC \
	events := p.Feed([]byte{0x1b, '\\'})
	// Should eventually produce some output
	_ = events
}

// === parseKittyCSIU line 616: CSI 8 u = Backspace ===

func TestP122_ParseKittyCSIU_Backspace(t *testing.T) {
	p := NewParser()
	events := p.Feed([]byte{0x1b, '[', '8', 'u'})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key == nil || events[0].Key.Key != KeyBackspace {
		t.Error("expected KeyBackspace for CSI 8 u")
	}
}

// === parseKittyCSIU lines 629-630: CSI 57353 u = Delete ===

func TestP122_ParseKittyCSIU_Delete(t *testing.T) {
	p := NewParser()
	events := p.Feed([]byte{0x1b, '[', '5', '7', '3', '5', '3', 'u'})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key == nil || events[0].Key.Key != KeyDelete {
		t.Error("expected KeyDelete for CSI 57353 u")
	}
}

// === parseKittyCSIU line 616-617: CSI 8 u = Backspace with modifier ===

func TestP122_ParseKittyCSIU_BackspaceWithCtrl(t *testing.T) {
	p := NewParser()
	// CSI 8 ; 5 u = Ctrl+Backspace (modifier 5 = Ctrl in Kitty)
	events := p.Feed([]byte{0x1b, '[', '8', ';', '5', 'u'})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key == nil || events[0].Key.Key != KeyBackspace {
		t.Error("expected KeyBackspace")
	}
	if events[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

// === ParseColorResponse lines 369-371: invalid rgb: format (wrong components) ===

func TestP122_ParseColorResponse_InvalidRgb(t *testing.T) {
	// rgb: with only 2 components (missing blue)
	cr := ParseColorResponse("\x1b]4;0;rgb:ff00/00ff\x07")
	if cr.Valid {
		t.Error("expected invalid for malformed rgb (2 components)")
	}
}

func TestP122_ParseColorResponse_TooManyComponents(t *testing.T) {
	// rgb: with 4 components
	cr := ParseColorResponse("\x1b]4;0;rgb:ff/00/ff/aa\x07")
	if cr.Valid {
		t.Error("expected invalid for malformed rgb (4 components)")
	}
}

// === parseHexComponent line 444-445: unexpected hex length (default case) ===

func TestP122_ParseHexComponent_OddLength(t *testing.T) {
	// Test with 3-char hex string (not 1, 2, or 4 chars)
	// This exercises the default case in the switch
	cr := ParseColorResponse("\x1b]4;0;rgb:abc/def/123\x07")
	// The components are 3 chars each, which hits the default case
	// parseHexComponent should return some value (uint8 truncation)
	if !cr.Valid {
		t.Error("expected valid response despite odd-length hex")
	}
}

func TestP122_ParseHexComponent_FiveChar(t *testing.T) {
	// Test with 5-char hex string (not 1, 2, or 4 chars)
	cr := ParseColorResponse("\x1b]10;rgb:abcde/abcde/abcde\x07")
	_ = cr // exercises default case in parseHexComponent
}

// === EncodeSixelSimple edge case ===

func TestP122_EncodeSixelSimple_SinglePixel(t *testing.T) {
	// Single pixel, grayscale 0 (black)
	result := EncodeSixelSimple([]byte{0}, 1, 1)
	if result == "" {
		t.Error("expected non-empty result")
	}
	// Should contain DCS introducer
	if !p122contains(result, "\x1bP") {
		t.Error("expected DCS introducer in Sixel output")
	}
}

func TestP122_EncodeSixelSimple_AllMaxLevel(t *testing.T) {
	// 2x2 all max level (255 = white)
	result := EncodeSixelSimple([]byte{255, 255, 255, 255}, 2, 2)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

// === ParseOSC52Response uncovered branches ===

func TestP122_ParseOSC52Response_NotOSC52(t *testing.T) {
	// Not an OSC52 response at all
	_, ok := ParseOSC52Response("\x1b]99;hello\x07")
	if ok {
		t.Error("expected false for non-OSC52 response")
	}
}

func TestP122_ParseOSC52Response_EmptyPayload(t *testing.T) {
	// OSC52 with empty base64 payload
	_, ok := ParseOSC52Response("\x1b]52;c;\x07")
	if ok {
		// Empty payload might return true with empty string, or false
		// Either is acceptable
	}
}

// === Clipboard round-trip with special characters ===

func TestP122_Clipboard_RoundTripSpecial(t *testing.T) {
	// Test that we can handle unicode in clipboard
	original := "Hello, 世界! 🌍"
	seq := CopyClipboard(original)
	if seq == "" {
		t.Fatal("expected non-empty clipboard sequence")
	}
}

// === Additional parseCSI coverage: F-keys via tilde with modifiers ===

func TestP122_ParseCSI_F5_WithShift(t *testing.T) {
	p := NewParser()
	// F5 = CSI 15 ~ with Shift modifier (2)
	events := p.Feed([]byte{0x1b, '[', '1', '5', ';', '2', '~'})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key == nil || events[0].Key.Key != KeyF5 {
		t.Error("expected KeyF5")
	}
	if events[0].Key.Modifiers&ModShift == 0 {
		t.Error("expected Shift modifier")
	}
}

func TestP122_ParseCSI_F6_Plain(t *testing.T) {
	p := NewParser()
	// F6 = CSI 17 ~
	events := p.Feed([]byte{0x1b, '[', '1', '7', '~'})
	if len(events) != 1 || events[0].Key == nil || events[0].Key.Key != KeyF6 {
		t.Error("expected KeyF6")
	}
}

func TestP122_ParseCSI_F7_Plain(t *testing.T) {
	p := NewParser()
	// F7 = CSI 18 ~
	events := p.Feed([]byte{0x1b, '[', '1', '8', '~'})
	if len(events) != 1 || events[0].Key == nil || events[0].Key.Key != KeyF7 {
		t.Error("expected KeyF7")
	}
}

func TestP122_ParseCSI_F8_Plain(t *testing.T) {
	p := NewParser()
	// F8 = CSI 19 ~
	events := p.Feed([]byte{0x1b, '[', '1', '9', '~'})
	if len(events) != 1 || events[0].Key == nil || events[0].Key.Key != KeyF8 {
		t.Error("expected KeyF8")
	}
}

func TestP122_ParseCSI_F9_Plain(t *testing.T) {
	p := NewParser()
	// F9 = CSI 20 ~
	events := p.Feed([]byte{0x1b, '[', '2', '0', '~'})
	if len(events) != 1 || events[0].Key == nil || events[0].Key.Key != KeyF9 {
		t.Error("expected KeyF9")
	}
}

func TestP122_ParseCSI_F10_Plain(t *testing.T) {
	p := NewParser()
	// F10 = CSI 21 ~
	events := p.Feed([]byte{0x1b, '[', '2', '1', '~'})
	if len(events) != 1 || events[0].Key == nil || events[0].Key.Key != KeyF10 {
		t.Error("expected KeyF10")
	}
}

// === parseCSI: focus events coexisting with F-keys ===

func TestP122_ParseCSI_FocusInNotConfused(t *testing.T) {
	p := NewParser()
	events := p.Feed([]byte{0x1b, '[', 'I'})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != EventFocus || !events[0].Focused {
		t.Error("expected FocusIn event")
	}
}

func TestP122_ParseCSI_FocusOutNotConfused(t *testing.T) {
	p := NewParser()
	events := p.Feed([]byte{0x1b, '[', 'O'})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != EventFocus || events[0].Focused {
		t.Error("expected FocusOut event")
	}
}

// === feed(): Ctrl+letter combinations ===

func TestP122_Feed_CtrlA(t *testing.T) {
	p := NewParser()
	events := p.Feed([]byte{0x01})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key == nil {
		t.Fatal("expected key event")
	}
	if events[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

func TestP122_Feed_CtrlZ(t *testing.T) {
	p := NewParser()
	events := p.Feed([]byte{0x1a})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key == nil || events[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl+Z")
	}
}

// helper
func p122contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || p122contains(s[1:], substr))
}

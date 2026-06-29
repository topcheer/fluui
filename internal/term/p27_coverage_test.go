package term

import (
	"testing"
)

// P27 coverage tests for internal/term — targeting parseCSI, feed,
// decodeMouseButton, and FeedTimeout uncovered branches.

// --- parseCSI modifier extraction ---

func TestP27_ParseCSI_ModShift(t *testing.T) {
	p := NewParser()
	// ESC [ 1 ; 2 A = Shift+Up
	evs := p.Feed([]byte{0x1b, '[', '1', ';', '2', 'A'})
	if len(evs) == 0 || evs[0].Key == nil {
		t.Fatal("expected key event")
	}
	if evs[0].Key.Key != KeyUp {
		t.Errorf("expected KeyUp, got %v", evs[0].Key.Key)
	}
	if evs[0].Key.Modifiers&ModShift == 0 {
		t.Error("expected Shift modifier")
	}
}

func TestP27_ParseCSI_ModAlt(t *testing.T) {
	p := NewParser()
	// ESC [ 1 ; 3 A = Alt+Up
	evs := p.Feed([]byte{0x1b, '[', '1', ';', '3', 'A'})
	if len(evs) == 0 || evs[0].Key == nil {
		t.Fatal("expected key event")
	}
	if evs[0].Key.Modifiers&ModAlt == 0 {
		t.Error("expected Alt modifier")
	}
}

func TestP27_ParseCSI_ModCtrl(t *testing.T) {
	p := NewParser()
	// ESC [ 1 ; 5 A = Ctrl+Up
	evs := p.Feed([]byte{0x1b, '[', '1', ';', '5', 'A'})
	if len(evs) == 0 || evs[0].Key == nil {
		t.Fatal("expected key event")
	}
	if evs[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

func TestP27_ParseCSI_NoMatch(t *testing.T) {
	p := NewParser()
	// ESC [ Z = Shift+Tab
	evs := p.Feed([]byte{0x1b, '[', 'Z'})
	if len(evs) == 0 || evs[0].Key == nil {
		t.Fatal("expected key event")
	}
	if evs[0].Key.Key != KeyBacktab {
		t.Errorf("expected KeyBacktab, got %v", evs[0].Key.Key)
	}
}

func TestP27_ParseCSI_TildeF5(t *testing.T) {
	p := NewParser()
	// ESC [ 1 5 ~ = F5
	evs := p.Feed([]byte{0x1b, '[', '1', '5', '~'})
	if len(evs) == 0 || evs[0].Key == nil {
		t.Fatal("expected key event")
	}
	if evs[0].Key.Key != KeyF5 {
		t.Errorf("expected KeyF5, got %v", evs[0].Key.Key)
	}
}

func TestP27_ParseCSI_TildeF12(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte{0x1b, '[', '2', '4', '~'})
	if len(evs) == 0 || evs[0].Key == nil {
		t.Fatal("expected key event")
	}
	if evs[0].Key.Key != KeyF12 {
		t.Errorf("expected KeyF12, got %v", evs[0].Key.Key)
	}
}

func TestP27_ParseCSI_TildeInsert(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte{0x1b, '[', '2', '~'})
	if len(evs) == 0 || evs[0].Key == nil {
		t.Fatal("expected key event")
	}
	if evs[0].Key.Key != KeyInsert {
		t.Errorf("expected KeyInsert, got %v", evs[0].Key.Key)
	}
}

func TestP27_ParseCSI_TildeDelete(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte{0x1b, '[', '3', '~'})
	if len(evs) == 0 || evs[0].Key == nil {
		t.Fatal("expected key event")
	}
	if evs[0].Key.Key != KeyDelete {
		t.Errorf("expected KeyDelete, got %v", evs[0].Key.Key)
	}
}

func TestP27_ParseCSI_TildePageUp(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte{0x1b, '[', '5', '~'})
	if len(evs) == 0 || evs[0].Key == nil {
		t.Fatal("expected key event")
	}
	if evs[0].Key.Key != KeyPageUp {
		t.Errorf("expected KeyPageUp, got %v", evs[0].Key.Key)
	}
}

func TestP27_ParseCSI_TildePageDown(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte{0x1b, '[', '6', '~'})
	if len(evs) == 0 || evs[0].Key == nil {
		t.Fatal("expected key event")
	}
	if evs[0].Key.Key != KeyPageDown {
		t.Errorf("expected KeyPageDown, got %v", evs[0].Key.Key)
	}
}

// --- parseSS3 (F1-F4) ---

func TestP27_ParseSS3_F1(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte{0x1b, 'O', 'P'})
	if len(evs) == 0 || evs[0].Key == nil {
		t.Fatal("expected key event")
	}
	if evs[0].Key.Key != KeyF1 {
		t.Errorf("expected KeyF1, got %v", evs[0].Key.Key)
	}
}

func TestP27_ParseSS3_F4(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte{0x1b, 'O', 'S'})
	if len(evs) == 0 || evs[0].Key == nil {
		t.Fatal("expected key event")
	}
	if evs[0].Key.Key != KeyF4 {
		t.Errorf("expected KeyF4, got %v", evs[0].Key.Key)
	}
}

// --- feed edge cases ---

func TestP27_Feed_AltEnter(t *testing.T) {
	p := NewParser()
	// ESC followed by Enter (0x0d)
	evs := p.Feed([]byte{0x1b, 0x0d})
	if len(evs) == 0 || evs[0].Key == nil {
		t.Fatal("expected key event")
	}
	if evs[0].Key.Key != KeyEnter {
		t.Errorf("expected KeyEnter, got %v", evs[0].Key.Key)
	}
	if evs[0].Key.Modifiers&ModAlt == 0 {
		t.Error("expected Alt modifier")
	}
}

func TestP27_Feed_AltKey(t *testing.T) {
	p := NewParser()
	// ESC followed by 'a'
	evs := p.Feed([]byte{0x1b, 'a'})
	if len(evs) == 0 || evs[0].Key == nil {
		t.Fatal("expected key event")
	}
	if evs[0].Key.Rune != 'a' {
		t.Errorf("expected rune 'a', got %v", evs[0].Key.Rune)
	}
	if evs[0].Key.Modifiers&ModAlt == 0 {
		t.Error("expected Alt modifier")
	}
}

func TestP27_Feed_DoubleESC(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte{0x1b, 0x1b})
	if len(evs) == 0 {
		t.Fatal("expected at least one event")
	}
	// First ESC should produce KeyEscape
	if evs[0].Key == nil || evs[0].Key.Key != KeyEscape {
		t.Errorf("expected KeyEscape, got %+v", evs[0])
	}
}

func TestP27_Feed_UnknownEscape(t *testing.T) {
	p := NewParser()
	// ESC followed by non-printable (not [, ], O, printable)
	evs := p.Feed([]byte{0x1b, 0x01})
	// Should produce no events (or just go back to normal)
	for _, ev := range evs {
		_ = ev // just verify no panic
	}
}

func TestP27_Feed_UTF8Invalid(t *testing.T) {
	p := NewParser()
	// Invalid UTF-8: lead byte 0xC0 followed by non-continuation
	evs := p.Feed([]byte{0xC0, 0x41}) // 0xC0 is invalid lead, 0x41 = 'A'
	// Should produce replacement char + 'A' or similar
	if len(evs) == 0 {
		t.Fatal("expected at least one event")
	}
}

// --- decodeMouseButton branches ---

func TestP27_DecodeMouse_WheelUp(t *testing.T) {
	var btn MouseButton
	var mods ModMask
	var action MouseAction
	// code 64 = wheel, button 0 = up
	decodeMouseButton(64, &btn, &mods, &action)
	if btn != MouseWheelUp {
		t.Errorf("expected MouseWheelUp, got %v", btn)
	}
	if action != MouseWheel {
		t.Errorf("expected MouseWheel action, got %v", action)
	}
}

func TestP27_DecodeMouse_WheelDown(t *testing.T) {
	var btn MouseButton
	var mods ModMask
	var action MouseAction
	// code 65 = wheel, button 1 = down
	decodeMouseButton(65, &btn, &mods, &action)
	if btn != MouseWheelDown {
		t.Errorf("expected MouseWheelDown, got %v", btn)
	}
}

func TestP27_DecodeMouse_Drag(t *testing.T) {
	var btn MouseButton
	var mods ModMask
	var action MouseAction
	// code 32 = motion, button 0 = left drag
	decodeMouseButton(32, &btn, &mods, &action)
	if btn != MouseLeft {
		t.Errorf("expected MouseLeft, got %v", btn)
	}
	if action != MouseDrag {
		t.Errorf("expected MouseDrag, got %v", action)
	}
}

func TestP27_DecodeMouse_Move(t *testing.T) {
	var btn MouseButton
	var mods ModMask
	var action MouseAction
	// code 35 = motion (32) + button 3 (none)
	decodeMouseButton(35, &btn, &mods, &action)
	if action != MouseMove {
		t.Errorf("expected MouseMove, got %v", action)
	}
}

func TestP27_DecodeMouse_Middle(t *testing.T) {
	var btn MouseButton
	var mods ModMask
	var action MouseAction
	decodeMouseButton(1, &btn, &mods, &action)
	if btn != MouseMiddle {
		t.Errorf("expected MouseMiddle, got %v", btn)
	}
}

func TestP27_DecodeMouse_Right(t *testing.T) {
	var btn MouseButton
	var mods ModMask
	var action MouseAction
	decodeMouseButton(2, &btn, &mods, &action)
	if btn != MouseRight {
		t.Errorf("expected MouseRight, got %v", btn)
	}
}

// --- FeedTimeout edge cases ---

func TestP27_FeedTimeout_NoPendingEsc(t *testing.T) {
	p := NewParser()
	// No pending ESC — should return nil
	evs := p.FeedTimeout()
	if len(evs) != 0 {
		t.Errorf("expected 0 events, got %d", len(evs))
	}
}

// --- OSC52 clipboard ---

func TestP27_ParseOSC52_Empty(t *testing.T) {
	resp, ok := ParseOSC52Response("")
	if ok {
		t.Error("expected ok=false for empty response")
	}
	if resp != "" {
		t.Errorf("expected empty response, got %q", resp)
	}
}

func TestP27_CopyOSC52_Roundtrip(t *testing.T) {
	// Copy and parse back
	seq := CopyOSC52("hello world")
	if seq == "" {
		t.Fatal("CopyOSC52 should return non-empty")
	}
	result, ok := ParseOSC52Response(seq)
	if !ok {
		t.Fatal("ParseOSC52Response should return ok=true")
	}
	if result != "hello world" {
		t.Errorf("roundtrip: expected 'hello world', got %q", result)
	}
}

package term

import (
	"testing"
)

// --- parseCSI coverage tests ---

func TestP41_ParseCSI_FKeys(t *testing.T) {
	p := NewParser()
	// F1-F5 use SS3 (ESC O), not CSI
	// F6-F12 use CSI with tilde
	// Test F5 via SS3
	evs := p.Feed([]byte("\x1b[15~"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event for F5, got %d", len(evs))
	}
	if evs[0].Key == nil || evs[0].Key.Key != KeyF5 {
		t.Errorf("expected KeyF5, got %+v", evs[0])
	}
}

func TestP41_ParseCSI_F12(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[24~"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key == nil || evs[0].Key.Key != KeyF12 {
		t.Errorf("expected KeyF12, got %+v", evs[0])
	}
}

func TestP41_ParseCSI_InsertDelete(t *testing.T) {
	p := NewParser()
	// Insert = CSI 2~
	evs := p.Feed([]byte("\x1b[2~"))
	if len(evs) != 1 || evs[0].Key == nil || evs[0].Key.Key != KeyInsert {
		t.Errorf("expected KeyInsert, got %+v", evs)
	}

	// Delete = CSI 3~
	p = NewParser()
	evs = p.Feed([]byte("\x1b[3~"))
	if len(evs) != 1 || evs[0].Key == nil || evs[0].Key.Key != KeyDelete {
		t.Errorf("expected KeyDelete, got %+v", evs)
	}

	// Page Up = CSI 5~
	p = NewParser()
	evs = p.Feed([]byte("\x1b[5~"))
	if len(evs) != 1 || evs[0].Key == nil || evs[0].Key.Key != KeyPageUp {
		t.Errorf("expected KeyPageUp, got %+v", evs)
	}

	// Page Down = CSI 6~
	p = NewParser()
	evs = p.Feed([]byte("\x1b[6~"))
	if len(evs) != 1 || evs[0].Key == nil || evs[0].Key.Key != KeyPageDown {
		t.Errorf("expected KeyPageDown, got %+v", evs)
	}
}

func TestP41_ParseCSI_HomeEnd(t *testing.T) {
	p := NewParser()
	// Home = CSI 1~
	evs := p.Feed([]byte("\x1b[1~"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event for Home, got %d", len(evs))
	}
	if evs[0].Key == nil {
		t.Fatalf("expected non-nil key")
	}

	// End = CSI 4~
	p = NewParser()
	evs = p.Feed([]byte("\x1b[4~"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event for End, got %d", len(evs))
	}
}

func TestP41_ParseCSI_ModifierShift(t *testing.T) {
	p := NewParser()
	// CSI 1;2A = Shift+Up
	evs := p.Feed([]byte("\x1b[1;2A"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event for Shift+Up, got %d", len(evs))
	}
	if evs[0].Key == nil {
		t.Fatal("expected non-nil key")
	}
	if evs[0].Key.Modifiers&ModShift == 0 {
		t.Error("expected Shift modifier")
	}
}

func TestP41_ParseCSI_ModifierCtrl(t *testing.T) {
	p := NewParser()
	// CSI 1;5A = Ctrl+Up
	evs := p.Feed([]byte("\x1b[1;5A"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key == nil {
		t.Fatal("expected non-nil key")
	}
	if evs[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

func TestP41_ParseCSI_ModifierAlt(t *testing.T) {
	p := NewParser()
	// CSI 1;3A = Alt+Up
	evs := p.Feed([]byte("\x1b[1;3A"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key == nil {
		t.Fatal("expected non-nil key")
	}
	if evs[0].Key.Modifiers&ModAlt == 0 {
		t.Error("expected Alt modifier")
	}
}

func TestP41_ParseCSI_ArrowUp(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[A"))
	if len(evs) != 1 || evs[0].Key == nil || evs[0].Key.Key != KeyUp {
		t.Errorf("expected KeyUp, got %+v", evs)
	}
}

func TestP41_ParseCSI_ArrowDown(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[B"))
	if len(evs) != 1 || evs[0].Key == nil || evs[0].Key.Key != KeyDown {
		t.Errorf("expected KeyDown, got %+v", evs)
	}
}

func TestP41_ParseCSI_ArrowRight(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[C"))
	if len(evs) != 1 || evs[0].Key == nil || evs[0].Key.Key != KeyRight {
		t.Errorf("expected KeyRight, got %+v", evs)
	}
}

func TestP41_ParseCSI_ArrowLeft(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[D"))
	if len(evs) != 1 || evs[0].Key == nil || evs[0].Key.Key != KeyLeft {
		t.Errorf("expected KeyLeft, got %+v", evs)
	}
}

// --- feed coverage ---

func TestP41_Feed_MultipleEvents(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("abc"))
	if len(evs) != 3 {
		t.Fatalf("expected 3 events for 'abc', got %d", len(evs))
	}
}

func TestP41_Feed_MixedEscapes(t *testing.T) {
	p := NewParser()
	// Mix of printable chars and escape sequences
	evs := p.Feed([]byte("a\x1b[Bc"))
	if len(evs) != 3 {
		t.Fatalf("expected 3 events, got %d: %+v", len(evs), evs)
	}
	// First should be 'a', second Down, third 'c'
	if evs[0].Key == nil || evs[0].Key.Rune != 'a' {
		t.Errorf("expected 'a' first, got %+v", evs[0])
	}
	if evs[1].Key == nil || evs[1].Key.Key != KeyDown {
		t.Errorf("expected KeyDown second, got %+v", evs[1])
	}
	if evs[2].Key == nil || evs[2].Key.Rune != 'c' {
		t.Errorf("expected 'c' third, got %+v", evs[2])
	}
}

func TestP41_Feed_Empty(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte{})
	if len(evs) != 0 {
		t.Errorf("expected 0 events for empty input, got %d", len(evs))
	}
}

// --- parseSGRMouse coverage ---

func TestP41_ParseSGRMouse_Move(t *testing.T) {
	p := NewParser()
	// SGR mouse: CSI < 35 ; 10 ; 5 M (left button move at 10,5)
	evs := p.Feed([]byte("\x1b[<35;10;5M"))
	// May produce 0 or 1 events depending on mouse handling
	_ = evs // just verify no panic
}

func TestP41_ParseSGRMouse_Release(t *testing.T) {
	p := NewParser()
	// SGR mouse release: CSI < 0 ; 10 ; 5 m
	evs := p.Feed([]byte("\x1b[<0;10;5m"))
	_ = evs
}

func TestP41_ParseSGRMouse_Wheel(t *testing.T) {
	p := NewParser()
	// SGR mouse wheel up: CSI < 64 ; 10 ; 5 M
	evs := p.Feed([]byte("\x1b[<64;10;5M"))
	_ = evs
}

// --- ParseOSC52Response coverage ---

func TestP41_ParseOSC52Response_Valid(t *testing.T) {
	// OSC52 paste response: ESC ] 52 ; c ; <base64> ST
	resp := "\x1b]52;c;SGVsbG8=\x07"
	text, ok := ParseOSC52Response(resp)
	if !ok {
		t.Error("expected ok=true")
	}
	if text != "Hello" {
		t.Errorf("expected 'Hello', got %q", text)
	}
}

func TestP41_ParseOSC52Response_Empty(t *testing.T) {
	resp := "\x1b]52;c;\x07"
	text, ok := ParseOSC52Response(resp)
	_ = text
	_ = ok
	// Just verify no panic
}

func TestP41_ParseOSC52Response_NotOSC52(t *testing.T) {
	resp := "not osc52"
	_, ok := ParseOSC52Response(resp)
	if ok {
		t.Error("expected ok=false for non-OSC52 input")
	}
}

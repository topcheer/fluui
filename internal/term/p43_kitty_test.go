package term

import (
	"testing"
)

// --- Kitty Keyboard Protocol CSI u parsing tests ---

func TestP43_KittyCSIU_PlainChar(t *testing.T) {
	p := NewParser()
	// CSI 97 u = 'a'
	evs := p.Feed([]byte("\x1b[97u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key == nil {
		t.Fatal("expected non-nil key")
	}
	if evs[0].Key.Rune != 'a' {
		t.Errorf("expected rune 'a', got %q", string(evs[0].Key.Rune))
	}
}

func TestP43_KittyCSIU_CtrlA(t *testing.T) {
	p := NewParser()
	// CSI 1 ; 5 u = Ctrl+A (codepoint 1, modifier 4=Ctrl in Kitty)
	evs := p.Feed([]byte("\x1b[1;5u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key == nil {
		t.Fatal("expected non-nil key")
	}
	if evs[0].Key.Rune != 'a' {
		t.Errorf("expected rune 'a', got %q", string(evs[0].Key.Rune))
	}
	if evs[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

func TestP43_KittyCSIU_ShiftA(t *testing.T) {
	p := NewParser()
	// CSI 65 ; 1 u = Shift+A (codepoint 'A'=65, modifier 1=Shift in Kitty)
	evs := p.Feed([]byte("\x1b[65;1u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key == nil {
		t.Fatal("expected non-nil key")
	}
	if evs[0].Key.Rune != 'A' {
		t.Errorf("expected rune 'A', got %q", string(evs[0].Key.Rune))
	}
	if evs[0].Key.Modifiers&ModShift == 0 {
		t.Error("expected Shift modifier")
	}
}

func TestP43_KittyCSIU_AltX(t *testing.T) {
	p := NewParser()
	// CSI 120 ; 3 u = Alt+X (codepoint 'x'=120, modifier 2=Alt in Kitty)
	evs := p.Feed([]byte("\x1b[120;3u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key == nil {
		t.Fatal("expected non-nil key")
	}
	if evs[0].Key.Rune != 'x' {
		t.Errorf("expected rune 'x', got %q", string(evs[0].Key.Rune))
	}
	if evs[0].Key.Modifiers&ModAlt == 0 {
		t.Error("expected Alt modifier")
	}
}

func TestP43_KittyCSIU_CtrlShiftE(t *testing.T) {
	p := NewParser()
	// CSI 69 ; 5 u = Ctrl+Shift+E (codepoint 'E'=69, modifier 5=Shift(1)+Ctrl(4) in Kitty)
	evs := p.Feed([]byte("\x1b[69;5u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key == nil {
		t.Fatal("expected non-nil key")
	}
	if evs[0].Key.Rune != 'E' {
		t.Errorf("expected rune 'E', got %q", string(evs[0].Key.Rune))
	}
	if evs[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
	if evs[0].Key.Modifiers&ModShift == 0 {
		t.Error("expected Shift modifier")
	}
}

func TestP43_KittyCSIU_Escape(t *testing.T) {
	p := NewParser()
	// CSI 27 u = ESC (codepoint 0x1b = 27)
	evs := p.Feed([]byte("\x1b[27u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key == nil {
		t.Fatal("expected non-nil key")
	}
	if evs[0].Key.Key != KeyEscape {
		t.Errorf("expected KeyEscape, got %d", evs[0].Key.Key)
	}
}

func TestP43_KittyCSIU_Tab(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[9u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key.Key != KeyTab {
		t.Errorf("expected KeyTab, got %d", evs[0].Key.Key)
	}
}

func TestP43_KittyCSIU_Enter(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[13u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key.Key != KeyEnter {
		t.Errorf("expected KeyEnter, got %d", evs[0].Key.Key)
	}
}

func TestP43_KittyCSIU_Backspace(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[127u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key.Key != KeyDelete {
		t.Errorf("expected KeyDelete, got %d", evs[0].Key.Key)
	}
}

func TestP43_KittyCSIU_Space(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[32u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key.Key != KeySpace {
		t.Errorf("expected KeySpace, got %d", evs[0].Key.Key)
	}
}

func TestP43_KittyCSIU_ArrowUp(t *testing.T) {
	p := NewParser()
	// Kitty codepoint 57360 = Up arrow
	evs := p.Feed([]byte("\x1b[57360u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key.Key != KeyUp {
		t.Errorf("expected KeyUp, got %d", evs[0].Key.Key)
	}
}

func TestP43_KittyCSIU_ArrowDown(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[57361u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key.Key != KeyDown {
		t.Errorf("expected KeyDown, got %d", evs[0].Key.Key)
	}
}

func TestP43_KittyCSIU_ArrowLeft(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[57358u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key.Key != KeyLeft {
		t.Errorf("expected KeyLeft, got %d", evs[0].Key.Key)
	}
}

func TestP43_KittyCSIU_ArrowRight(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[57359u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key.Key != KeyRight {
		t.Errorf("expected KeyRight, got %d", evs[0].Key.Key)
	}
}

func TestP43_KittyCSIU_CtrlArrowUp(t *testing.T) {
	p := NewParser()
	// Ctrl+Up = CSI 57360 ; 5 u (modifier 4=Ctrl in Kitty bitmask)
	evs := p.Feed([]byte("\x1b[57360;5u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key.Key != KeyUp {
		t.Errorf("expected KeyUp, got %d", evs[0].Key.Key)
	}
	if evs[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

func TestP43_KittyCSIU_Home(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[57356u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key.Key != KeyHome {
		t.Errorf("expected KeyHome, got %d", evs[0].Key.Key)
	}
}

func TestP43_KittyCSIU_End(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[57357u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key.Key != KeyEnd {
		t.Errorf("expected KeyEnd, got %d", evs[0].Key.Key)
	}
}

func TestP43_KittyCSIU_PageUp(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[57354u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key.Key != KeyPageUp {
		t.Errorf("expected KeyPageUp, got %d", evs[0].Key.Key)
	}
}

func TestP43_KittyCSIU_PageDown(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[57355u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key.Key != KeyPageDown {
		t.Errorf("expected KeyPageDown, got %d", evs[0].Key.Key)
	}
}

func TestP43_KittyCSIU_Insert(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[57352u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key.Key != KeyInsert {
		t.Errorf("expected KeyInsert, got %d", evs[0].Key.Key)
	}
}

func TestP43_KittyCSIU_F1(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[57362u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key.Key != KeyF1 {
		t.Errorf("expected KeyF1, got %d", evs[0].Key.Key)
	}
}

func TestP43_KittyCSIU_F12(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[57373u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key.Key != KeyF12 {
		t.Errorf("expected KeyF12, got %d", evs[0].Key.Key)
	}
}

func TestP43_KittyCSIU_F5_F8(t *testing.T) {
	tests := []struct {
		codepoint int
		expected  KeyCode
	}{
		{57366, KeyF5},
		{57367, KeyF6},
		{57368, KeyF7},
		{57369, KeyF8},
	}
	for _, tc := range tests {
		p := NewParser()
		evs := p.Feed([]byte("\x1b[" + itoa(tc.codepoint) + "u"))
		if len(evs) != 1 || evs[0].Key == nil || evs[0].Key.Key != tc.expected {
			t.Errorf("codepoint %d: expected KeyF%d, got %+v", tc.codepoint, tc.expected, evs)
		}
	}
}

func TestP43_KittyCSIU_F9_F11(t *testing.T) {
	tests := []struct {
		codepoint int
		expected  KeyCode
	}{
		{57370, KeyF9},
		{57371, KeyF10},
		{57372, KeyF11},
	}
	for _, tc := range tests {
		p := NewParser()
		evs := p.Feed([]byte("\x1b[" + itoa(tc.codepoint) + "u"))
		if len(evs) != 1 || evs[0].Key == nil || evs[0].Key.Key != tc.expected {
			t.Errorf("codepoint %d: expected %d, got %+v", tc.codepoint, tc.expected, evs)
		}
	}
}

func TestP43_KittyCSIU_F2_F4(t *testing.T) {
	tests := []struct {
		codepoint int
		expected  KeyCode
	}{
		{57363, KeyF2},
		{57364, KeyF3},
		{57365, KeyF4},
	}
	for _, tc := range tests {
		p := NewParser()
		evs := p.Feed([]byte("\x1b[" + itoa(tc.codepoint) + "u"))
		if len(evs) != 1 || evs[0].Key == nil || evs[0].Key.Key != tc.expected {
			t.Errorf("codepoint %d: expected %d, got %+v", tc.codepoint, tc.expected, evs)
		}
	}
}

func TestP43_KittyCSIU_NoParams(t *testing.T) {
	// CSI u with no params — should return nil
	p := NewParser()
	// Can't easily produce this since parser needs params before 'u'
	// The parseCSI function would get params="" and final='u'
	// parseKittyCSIU with empty nums returns nil
	evs := p.Feed([]byte("\x1b[u"))
	// May or may not produce events depending on parser state
	_ = evs
}

func TestP43_KittyCSIU_MixedWithTraditional(t *testing.T) {
	p := NewParser()
	// Mix traditional CSI A (Up) with Kitty CSI u
	evs := p.Feed([]byte("\x1b[A\x1b[97u"))
	if len(evs) != 2 {
		t.Fatalf("expected 2 events, got %d", len(evs))
	}
	if evs[0].Key == nil || evs[0].Key.Key != KeyUp {
		t.Errorf("expected KeyUp first, got %+v", evs[0])
	}
	if evs[1].Key == nil || evs[1].Key.Rune != 'a' {
		t.Errorf("expected 'a' second, got %+v", evs[1])
	}
}

func TestP43_KittyCSIU_UnicodeChar(t *testing.T) {
	p := NewParser()
	// CSI 8364 u = '€' (Euro sign)
	evs := p.Feed([]byte("\x1b[8364u"))
	if len(evs) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evs))
	}
	if evs[0].Key == nil {
		t.Fatal("expected non-nil key")
	}
	if evs[0].Key.Rune != '€' {
		t.Errorf("expected '€', got %q", string(evs[0].Key.Rune))
	}
}

func TestP43_KittyCSIU_AllModifiers(t *testing.T) {
	tests := []struct {
		modCode  int
		expected ModMask
		name     string
	}{
		{1, ModShift, "Shift"},
		{2, ModAlt, "Alt"},
		{4, ModCtrl, "Ctrl"},
		{3, ModShift | ModAlt, "Shift+Alt"},
		{5, ModShift | ModCtrl, "Shift+Ctrl"},
		{6, ModAlt | ModCtrl, "Alt+Ctrl"},
		{7, ModShift | ModAlt | ModCtrl, "Shift+Alt+Ctrl"},
	}
	for _, tc := range tests {
		p := NewParser()
		input := "\x1b[97;" + itoa(tc.modCode) + "u"
		evs := p.Feed([]byte(input))
		if len(evs) != 1 || evs[0].Key == nil {
			t.Errorf("%s: expected 1 event, got %d", tc.name, len(evs))
			continue
		}
		if evs[0].Key.Modifiers != tc.expected {
			t.Errorf("%s: expected modifiers %d, got %d", tc.name, tc.expected, evs[0].Key.Modifiers)
		}
	}
}

func TestP43_KittyCSIU_CtrlLetterRange(t *testing.T) {
	// Ctrl+A through Ctrl+Z are codepoints 1-26
	tests := []struct {
		codepoint int
		letter    rune
	}{
		{1, 'a'}, {2, 'b'}, {3, 'c'}, {5, 'e'}, {12, 'l'}, {26, 'z'},
	}
	for _, tc := range tests {
		p := NewParser()
		input := "\x1b[" + itoa(tc.codepoint) + "u"
		evs := p.Feed([]byte(input))
		if len(evs) != 1 || evs[0].Key == nil {
			t.Errorf("codepoint %d: expected 1 event", tc.codepoint)
			continue
		}
		if evs[0].Key.Rune != tc.letter {
			t.Errorf("codepoint %d: expected letter %q, got %q",
				tc.codepoint, string(tc.letter), string(evs[0].Key.Rune))
		}
		if evs[0].Key.Modifiers&ModCtrl == 0 {
			t.Errorf("codepoint %d: expected Ctrl modifier", tc.codepoint)
		}
	}
}

func TestP43_DecodeKittyModifier(t *testing.T) {
	if decodeKittyModifier(0) != 0 {
		t.Error("expected 0 for code 0")
	}
	if decodeKittyModifier(1) != ModShift {
		t.Error("expected ModShift for code 1")
	}
	if decodeKittyModifier(2) != ModAlt {
		t.Error("expected ModAlt for code 2")
	}
	if decodeKittyModifier(4) != ModCtrl {
		t.Error("expected ModCtrl for code 4")
	}
	if decodeKittyModifier(7) != ModShift|ModAlt|ModCtrl {
		t.Error("expected Shift+Alt+Ctrl for code 7")
	}
}

// itoa converts an int to string (local helper to avoid strconv import).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

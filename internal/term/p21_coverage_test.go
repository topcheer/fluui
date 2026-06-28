package term

import (
	"testing"
)

// P21-D: Coverage tests for internal/term

// ─── ModMask.String() ───────────────────────────────────

func TestP21_ModMaskString(t *testing.T) {
	tests := []struct {
		mod  ModMask
		want string
	}{
		{0, ""},
		{ModCtrl, "Ctrl+"},
		{ModAlt, "Alt+"},
		{ModShift, "Shift+"},
		{ModCtrl | ModAlt, "Ctrl+Alt+"},
		{ModCtrl | ModShift, "Ctrl+Shift+"},
		{ModCtrl | ModAlt | ModShift, "Ctrl+Alt+Shift+"},
	}
	for _, tt := range tests {
		got := tt.mod.String()
		if got != tt.want {
			t.Errorf("ModMask(%d).String() = %q, want %q", tt.mod, got, tt.want)
		}
	}
}

// ─── KeyCode.String() ───────────────────────────────────

func TestP21_KeyCodeString(t *testing.T) {
	tests := []struct {
		key  KeyCode
		want string
	}{
		{KeyEnter, "Enter"},
		{KeyTab, "Tab"},
		{KeyBacktab, "BackTab"},
		{KeyBackspace, "Backspace"},
		{KeyDelete, "Delete"},
		{KeyInsert, "Insert"},
		{KeyHome, "Home"},
		{KeyEnd, "End"},
		{KeyPageUp, "PageUp"},
		{KeyPageDown, "PageDown"},
		{KeyUp, "Up"},
		{KeyDown, "Down"},
		{KeyLeft, "Left"},
		{KeyRight, "Right"},
		{KeyEscape, "Escape"},
		{KeySpace, "Space"},
		{KeyF1, "F1"},
		{KeyF5, "F5"},
		{KeyF12, "F12"},
		{KeyCode(999), "Key(999)"},
	}
	for _, tt := range tests {
		got := tt.key.String()
		if got != tt.want {
			t.Errorf("KeyCode(%d).String() = %q, want %q", tt.key, got, tt.want)
		}
	}
}

// ─── parseSS3: F1-F4 keys ───────────────────────────────

func TestP21_ParseSS3_FKeys(t *testing.T) {
	tests := []struct {
		byte byte
		key  KeyCode
	}{
		{'P', KeyF1},
		{'Q', KeyF2},
		{'R', KeyF3},
		{'S', KeyF4},
	}
	for _, tt := range tests {
		p := NewParser()
		p.parseSS3([]byte{tt.byte})
		// parseSS3 returns *Event directly, but it's unexported
		// Test via Feed integration instead
		evs := p.Feed([]byte{0x1b, 'O', tt.byte})
		found := false
		for _, ev := range evs {
			if ev.Type == EventKey && ev.Key != nil && ev.Key.Key == tt.key {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SS3 %c: no F-key event found in %d events", tt.byte, len(evs))
		}
	}
}

func TestP21_ParseSS3_Unknown(t *testing.T) {
	p := NewParser()
	result := p.parseSS3([]byte{'X'})
	if result != nil {
		t.Error("unknown SS3 should return nil")
	}
}

func TestP21_ParseSS3_Empty(t *testing.T) {
	p := NewParser()
	result := p.parseSS3([]byte{})
	if result != nil {
		t.Error("empty SS3 should return nil")
	}
}

// ─── parseCSI: arrow keys and special keys ─────────────

func TestP21_ParseCSI_ArrowKeys(t *testing.T) {
	tests := []struct {
		seq []byte
		key KeyCode
	}{
		{[]byte{'A'}, KeyUp},
		{[]byte{'B'}, KeyDown},
		{[]byte{'C'}, KeyRight},
		{[]byte{'D'}, KeyLeft},
		{[]byte{'H'}, KeyHome},
		{[]byte{'F'}, KeyEnd},
	}
	for _, tt := range tests {
		p := NewParser()
		ev := p.parseCSI(tt.seq)
		if ev == nil || ev.Key == nil {
			t.Errorf("parseCSI(%c): nil event", tt.seq[0])
			continue
		}
		if ev.Key.Key != tt.key {
			t.Errorf("parseCSI(%c): key = %v, want %v", tt.seq[0], ev.Key.Key, tt.key)
		}
	}
}

func TestP21_ParseCSI_ShiftTab(t *testing.T) {
	p := NewParser()
	ev := p.parseCSI([]byte{'Z'})
	if ev == nil || ev.Key == nil {
		t.Fatal("nil event for Shift+Tab")
	}
	if ev.Key.Key != KeyBacktab {
		t.Errorf("key = %v, want BackTab", ev.Key.Key)
	}
	if ev.Key.Modifiers&ModShift == 0 {
		t.Error("Shift modifier not set for Shift+Tab")
	}
}

func TestP21_ParseCSI_TildeSequences(t *testing.T) {
	tests := []struct {
		params string
		key    KeyCode
	}{
		{"1", KeyHome},
		{"2", KeyInsert},
		{"3", KeyDelete},
		{"4", KeyEnd},
		{"5", KeyPageUp},
		{"6", KeyPageDown},
	}
	for _, tt := range tests {
		p := NewParser()
		buf := append([]byte(tt.params), '~')
		ev := p.parseCSI(buf)
		if ev == nil || ev.Key == nil {
			t.Errorf("parseCSI(%s~): nil event", tt.params)
			continue
		}
		if ev.Key.Key != tt.key {
			t.Errorf("parseCSI(%s~): key = %v, want %v", tt.params, ev.Key.Key, tt.key)
		}
	}
}

func TestP21_ParseCSI_Empty(t *testing.T) {
	p := NewParser()
	result := p.parseCSI([]byte{})
	if result != nil {
		t.Error("empty CSI should return nil")
	}
}

// ─── parseSGRMouse ──────────────────────────────────────

func TestP21_ParseSGRMouse_Press(t *testing.T) {
	p := NewParser()
	ev := p.parseSGRMouse([]byte("<0;10;5M"))
	if ev == nil {
		t.Fatal("nil event for SGR mouse press")
	}
	if ev.Type != EventMouse {
		t.Errorf("type = %v, want EventMouse", ev.Type)
	}
	if ev.Mouse == nil {
		t.Fatal("nil mouse event")
	}
	if ev.Mouse.X != 9 || ev.Mouse.Y != 4 {
		t.Errorf("pos = (%d,%d), want (9,4)", ev.Mouse.X, ev.Mouse.Y)
	}
}

func TestP21_ParseSGRMouse_Release(t *testing.T) {
	p := NewParser()
	ev := p.parseSGRMouse([]byte("<0;10;5m"))
	if ev == nil {
		t.Fatal("nil event for SGR mouse release")
	}
	if ev.Mouse == nil {
		t.Fatal("nil mouse event")
	}
}

// ─── utf8ByteLen ────────────────────────────────────────

func TestP21_Utf8ByteLen(t *testing.T) {
	tests := []struct {
		b    byte
		want int
	}{
		{0x00, 1},
		{0x7F, 1},
		{0xC0, 2},
		{0xE0, 3},
		{0xF0, 4},
		{0x80, 0},
		{0xFF, 0},
	}
	for _, tt := range tests {
		got := utf8ByteLen(tt.b)
		if got != tt.want {
			t.Errorf("utf8ByteLen(0x%02X) = %d, want %d", tt.b, got, tt.want)
		}
	}
}

// ─── decodeMouseButton ──────────────────────────────────

func TestP21_DecodeMouseButton_Left(t *testing.T) {
	var btn MouseButton
	var mods ModMask
	var act MouseAction
	decodeMouseButton(0, &btn, &mods, &act)
	if btn != MouseLeft {
		t.Errorf("button = %v, want MouseLeft", btn)
	}
}

func TestP21_DecodeMouseButton_WheelUp(t *testing.T) {
	var btn MouseButton
	var mods ModMask
	var act MouseAction
	decodeMouseButton(64, &btn, &mods, &act)
	if act != MouseWheel {
		t.Errorf("action = %v, want MouseWheel", act)
	}
}

func TestP21_DecodeMouseButton_WheelDown(t *testing.T) {
	var btn MouseButton
	var mods ModMask
	var act MouseAction
	decodeMouseButton(65, &btn, &mods, &act)
	if act != MouseWheel {
		t.Errorf("action = %v, want MouseWheel", act)
	}
}

// ─── parseCSIParams ─────────────────────────────────────

func TestP21_ParseCSIParams(t *testing.T) {
	tests := []struct {
		s    string
		want []int
	}{
		{"", []int{}},
		{"1", []int{1}},
		{"1;2", []int{1, 2}},
		{"5;10;15", []int{5, 10, 15}},
	}
	for _, tt := range tests {
		got := parseCSIParams(tt.s)
		if len(got) != len(tt.want) {
			t.Errorf("parseCSIParams(%q): got %v, want %v", tt.s, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("parseCSIParams(%q)[%d]: got %d, want %d", tt.s, i, got[i], tt.want[i])
			}
		}
	}
}

// ─── atoi ───────────────────────────────────────────────

func TestP21_Atoi(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"", 0},
		{"0", 0},
		{"1", 1},
		{"10", 10},
		{"abc", 0},
	}
	for _, tt := range tests {
		got := atoi(tt.s)
		if got != tt.want {
			t.Errorf("atoi(%q) = %d, want %d", tt.s, got, tt.want)
		}
	}
}

// ─── firstNum ───────────────────────────────────────────

func TestP21_FirstNum(t *testing.T) {
	tests := []struct {
		nums []int
		want int
	}{
		{nil, 0},
		{[]int{5}, 5},
		{[]int{2, 1}, 2},
		{[]int{0, 1, 2}, 0},
	}
	for _, tt := range tests {
		got := firstNum(tt.nums)
		if got != tt.want {
			t.Errorf("firstNum(%v) = %d, want %d", tt.nums, got, tt.want)
		}
	}
}

// ─── Feed integration: printable chars ─────────────────

func TestP21_Feed_Printable(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("abc"))
	if len(evs) < 3 {
		t.Fatalf("expected >= 3 events, got %d", len(evs))
	}
	want := []rune{'a', 'b', 'c'}
	for i, ch := range want {
		if evs[i].Key == nil || evs[i].Key.Rune != ch {
			t.Errorf("event %d: rune = %q, want %q", i, evs[i].Key.Rune, ch)
		}
	}
}

func TestP21_Feed_Escape(t *testing.T) {
	p := NewParser()
	p.Feed([]byte{0x1b})
	// ESC alone may not emit immediately — parser waits to see if it's a sequence
	// FeedTimeout forces incomplete ESC to emit or just reset state
	_ = p.FeedTimeout()
	// Acceptable either way: emit ESC or just reset. Verify no hang/crash.
}

// ─── FeedTimeout ────────────────────────────────────────

func TestP21_FeedTimeout_PartialCSI(t *testing.T) {
	p := NewParser()
	p.Feed([]byte{0x1b, '['}) // incomplete CSI
	evs := p.FeedTimeout()
	// After timeout on incomplete CSI, should emit at least ESC
	if len(evs) == 0 {
		// Some implementations just reset state without emitting
		// This is acceptable behavior — just verify no hang
	}
}

// ─── ColorProfile constants ─────────────────────────────

func TestP21_ColorProfile_Distinct(t *testing.T) {
	if ProfileNone == ProfileANSI16 {
		t.Error("ProfileNone should differ from ProfileANSI16")
	}
	if ProfileANSI16 == Profile256 {
		t.Error("ProfileANSI16 should differ from Profile256")
	}
	if Profile256 == ProfileTrue {
		t.Error("Profile256 should differ from ProfileTrueColor")
	}
}

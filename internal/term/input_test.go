package term

import (
	"testing"
)

func TestParseSingleKeys(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  KeyCode
	}{
		{"Enter", "\r", KeyEnter},
		{"Tab", "\t", KeyTab},
		{"Backspace", "\x7f", KeyBackspace},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			events := p.Feed([]byte(tt.input))
			if len(events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(events))
			}
			if events[0].Key == nil {
				t.Fatal("expected key event")
			}
			if events[0].Key.Key != tt.want {
				t.Errorf("got key %v, want %v", events[0].Key.Key, tt.want)
			}
		})
	}
}

func TestParseCtrlCombos(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantRune rune
	}{
		{"Ctrl+A", "\x01", 'a'},
		{"Ctrl+C", "\x03", 'c'},
		{"Ctrl+Z", "\x1a", 'z'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			events := p.Feed([]byte(tt.input))
			if len(events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(events))
			}
			if events[0].Key == nil {
				t.Fatal("expected key event")
			}
			if events[0].Key.Rune != tt.wantRune {
				t.Errorf("got rune %c, want %c", events[0].Key.Rune, tt.wantRune)
			}
			if events[0].Key.Modifiers&ModCtrl == 0 {
				t.Error("expected Ctrl modifier")
			}
		})
	}
}

func TestParseArrowKeys(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  KeyCode
	}{
		{"Up", "\x1b[A", KeyUp},
		{"Down", "\x1b[B", KeyDown},
		{"Right", "\x1b[C", KeyRight},
		{"Left", "\x1b[D", KeyLeft},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			events := p.Feed([]byte(tt.input))
			if len(events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(events))
			}
			if events[0].Key == nil || events[0].Key.Key != tt.want {
				t.Errorf("got %v, want key %v", events[0].Key, tt.want)
			}
		})
	}
}

func TestParseModifiedArrows(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  KeyCode
		mods  ModMask
	}{
		{"Shift+Up", "\x1b[1;2A", KeyUp, ModShift},
		{"Ctrl+Up", "\x1b[1;5A", KeyUp, ModCtrl},
		{"Alt+Up", "\x1b[1;3A", KeyUp, ModAlt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			events := p.Feed([]byte(tt.input))
			if len(events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(events))
			}
			if events[0].Key == nil {
				t.Fatal("expected key event")
			}
			if events[0].Key.Key != tt.want {
				t.Errorf("got key %v, want %v", events[0].Key.Key, tt.want)
			}
			if events[0].Key.Modifiers&tt.mods == 0 {
				t.Errorf("expected modifier %d", tt.mods)
			}
		})
	}
}

func TestParseSpecialKeys(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  KeyCode
	}{
		{"Home", "\x1b[H", KeyHome},
		{"End", "\x1b[F", KeyEnd},
		{"PageUp", "\x1b[5~", KeyPageUp},
		{"PageDown", "\x1b[6~", KeyPageDown},
		{"Delete", "\x1b[3~", KeyDelete},
		{"F1", "\x1bOP", KeyF1},
		{"F5", "\x1b[15~", KeyF5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			events := p.Feed([]byte(tt.input))
			if len(events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(events))
			}
			if events[0].Key == nil || events[0].Key.Key != tt.want {
				t.Errorf("got %v, want key %v", events[0].Key, tt.want)
			}
		})
	}
}

func TestParseShiftTab(t *testing.T) {
	p := NewParser()
	events := p.Feed([]byte("\x1b[Z"))
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key == nil {
		t.Fatal("expected key event")
	}
	if events[0].Key.Key != KeyBacktab {
		t.Errorf("got %v, want BackTab", events[0].Key.Key)
	}
	if events[0].Key.Modifiers&ModShift == 0 {
		t.Error("expected Shift modifier")
	}
}

func TestParseAltKey(t *testing.T) {
	p := NewParser()
	events := p.Feed([]byte("\x1ba")) // Alt+a
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key == nil {
		t.Fatal("expected key event")
	}
	if events[0].Key.Rune != 'a' {
		t.Errorf("got rune %c, want 'a'", events[0].Key.Rune)
	}
	if events[0].Key.Modifiers&ModAlt == 0 {
		t.Error("expected Alt modifier")
	}
}

func TestParseSGRMouse(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		btn    MouseButton
		x, y   int
		action MouseAction
	}{
		{
			"LeftClick", "\x1b[<0;10;5M", MouseLeft, 9, 4, MouseDown,
		},
		{
			"LeftRelease", "\x1b[<0;10;5m", MouseLeft, 9, 4, MouseUp,
		},
		{
			"WheelUp", "\x1b[<64;10;5M", MouseWheelUp, 9, 4, MouseWheel,
		},
		{
			"WheelDown", "\x1b[<65;10;5M", MouseWheelDown, 9, 4, MouseWheel,
		},
		{
			"RightClick", "\x1b[<2;10;5M", MouseRight, 9, 4, MouseDown,
		},
		{
			"MiddleClick", "\x1b[<1;10;5M", MouseMiddle, 9, 4, MouseDown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			events := p.Feed([]byte(tt.input))
			if len(events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(events))
			}
			if events[0].Mouse == nil {
				t.Fatal("expected mouse event")
			}
			m := events[0].Mouse
			if m.Button != tt.btn {
				t.Errorf("button: got %v, want %v", m.Button, tt.btn)
			}
			if m.X != tt.x {
				t.Errorf("x: got %d, want %d", m.X, tt.x)
			}
			if m.Y != tt.y {
				t.Errorf("y: got %d, want %d", m.Y, tt.y)
			}
			if m.Action != tt.action {
				t.Errorf("action: got %v, want %v", m.Action, tt.action)
			}
		})
	}
}

func TestParseBracketedPaste(t *testing.T) {
	p := NewParser()
	events := p.Feed([]byte("\x1b[200~hello world\x1b[201~"))
	if len(events) != 1 {
		t.Fatalf("expected 1 event (paste), got %d", len(events))
	}
	if events[0].Type != EventPaste {
		t.Errorf("got type %v, want EventPaste", events[0].Type)
	}
	if events[0].Paste != "hello world" {
		t.Errorf("got paste %q, want %q", events[0].Paste, "hello world")
	}
}

func TestParseSplitSequence(t *testing.T) {
	// Arrow key arriving in two parts: ESC then [A
	p := NewParser()

	// Feed just ESC
	events := p.Feed([]byte{0x1b})
	if len(events) != 0 {
		t.Fatalf("expected 0 events after ESC, got %d", len(events))
	}

	// Feed the rest
	events = p.Feed([]byte("[A"))
	if len(events) != 1 {
		t.Fatalf("expected 1 event after completing sequence, got %d", len(events))
	}
	if events[0].Key == nil || events[0].Key.Key != KeyUp {
		t.Errorf("expected KeyUp")
	}
}

func TestParseRapidInput(t *testing.T) {
	// Multiple keys in one feed
	p := NewParser()
	events := p.Feed([]byte("abc"))
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	expected := []rune{'a', 'b', 'c'}
	for i, ev := range events {
		if ev.Key == nil || ev.Key.Rune != expected[i] {
			t.Errorf("event %d: got %v, want rune %c", i, ev.Key, expected[i])
		}
	}
}

func TestParseCtrlCQuits(t *testing.T) {
	p := NewParser()
	events := p.Feed([]byte("\x03"))
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Key == nil {
		t.Fatal("expected key event")
	}
	if events[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
	if events[0].Key.Rune != 'c' {
		t.Errorf("got rune %c, want 'c'", events[0].Key.Rune)
	}
}

func TestMouseModifiers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		mods  ModMask
	}{
		{"Shift+Click", "\x1b[<4;10;5M", ModShift},
		{"Alt+Click", "\x1b[<8;10;5M", ModAlt},
		{"Ctrl+Click", "\x1b[<16;10;5M", ModCtrl},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			events := p.Feed([]byte(tt.input))
			if len(events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(events))
			}
			if events[0].Mouse == nil {
				t.Fatal("expected mouse event")
			}
			if events[0].Mouse.Modifiers&tt.mods == 0 {
				t.Errorf("expected modifier %d, got %d", tt.mods, events[0].Mouse.Modifiers)
			}
		})
	}
}

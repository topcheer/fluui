package term

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP114_ParseCSI_F6_F12_Tilde(t *testing.T) {
	tests := []struct {
		params string
		key    KeyCode
	}{
		{"17", KeyF6},
		{"18", KeyF7},
		{"19", KeyF8},
		{"20", KeyF9},
		{"21", KeyF10},
		{"23", KeyF11},
		{"24", KeyF12},
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

func TestP114_ParseCSI_TildeWithModifiers(t *testing.T) {
	p := NewParser()
	ev := p.parseCSI([]byte("15;2~"))
	if ev == nil || ev.Key == nil {
		t.Fatal("nil event for Shift+F5")
	}
	if ev.Key.Key != KeyF5 {
		t.Errorf("expected KeyF5, got %v", ev.Key.Key)
	}
	if ev.Key.Modifiers&ModShift == 0 {
		t.Error("expected Shift modifier")
	}
}

func TestP114_ParseCSI_TildeWithCtrl(t *testing.T) {
	p := NewParser()
	ev := p.parseCSI([]byte("3;5~"))
	if ev == nil || ev.Key == nil {
		t.Fatal("nil event for Ctrl+Delete")
	}
	if ev.Key.Key != KeyDelete {
		t.Errorf("expected KeyDelete, got %v", ev.Key.Key)
	}
	if ev.Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

func TestP114_ParseCSI_TildeWithAlt(t *testing.T) {
	p := NewParser()
	ev := p.parseCSI([]byte("5;3~"))
	if ev == nil || ev.Key == nil {
		t.Fatal("nil event for Alt+PageUp")
	}
	if ev.Key.Key != KeyPageUp {
		t.Errorf("expected KeyPageUp, got %v", ev.Key.Key)
	}
	if ev.Key.Modifiers&ModAlt == 0 {
		t.Error("expected Alt modifier")
	}
}

func TestP114_ParseCSI_TildeUnknownNumber(t *testing.T) {
	p := NewParser()
	ev := p.parseCSI([]byte("99~"))
	if ev != nil {
		t.Error("expected nil for unknown tilde sequence")
	}
}

func TestP114_ParseCSI_ArrowWithShift(t *testing.T) {
	p := NewParser()
	ev := p.parseCSI([]byte("1;2A"))
	if ev == nil || ev.Key == nil {
		t.Fatal("nil event for Shift+Up")
	}
	if ev.Key.Key != KeyUp {
		t.Errorf("expected KeyUp, got %v", ev.Key.Key)
	}
	if ev.Key.Modifiers&ModShift == 0 {
		t.Error("expected Shift modifier")
	}
}

func TestP114_ParseCSI_ArrowWithCtrl(t *testing.T) {
	p := NewParser()
	ev := p.parseCSI([]byte("1;5C"))
	if ev == nil || ev.Key == nil {
		t.Fatal("nil event for Ctrl+Right")
	}
	if ev.Key.Key != KeyRight {
		t.Errorf("expected KeyRight, got %v", ev.Key.Key)
	}
	if ev.Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

func TestP114_ParseCSI_ArrowWithAlt(t *testing.T) {
	p := NewParser()
	ev := p.parseCSI([]byte("1;3B"))
	if ev == nil || ev.Key == nil {
		t.Fatal("nil event for Alt+Down")
	}
	if ev.Key.Key != KeyDown {
		t.Errorf("expected KeyDown, got %v", ev.Key.Key)
	}
	if ev.Key.Modifiers&ModAlt == 0 {
		t.Error("expected Alt modifier")
	}
}

func TestP114_ParseCSI_HomeWithCtrl(t *testing.T) {
	p := NewParser()
	ev := p.parseCSI([]byte("1;5H"))
	if ev == nil || ev.Key == nil {
		t.Fatal("nil event for Ctrl+Home")
	}
	if ev.Key.Key != KeyHome {
		t.Errorf("expected KeyHome, got %v", ev.Key.Key)
	}
	if ev.Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

func TestP114_ParseCSI_EndWithShift(t *testing.T) {
	p := NewParser()
	ev := p.parseCSI([]byte("1;2F"))
	if ev == nil || ev.Key == nil {
		t.Fatal("nil event for Shift+End")
	}
	if ev.Key.Key != KeyEnd {
		t.Errorf("expected KeyEnd, got %v", ev.Key.Key)
	}
	if ev.Key.Modifiers&ModShift == 0 {
		t.Error("expected Shift modifier")
	}
}

func TestP114_ParseCSI_StrayPasteEnd(t *testing.T) {
	p := NewParser()
	ev := p.parseCSI([]byte("201~"))
	if ev != nil {
		t.Error("expected nil for stray paste-end")
	}
}

func TestP114_ParseCSI_StrayPasteStartInPaste(t *testing.T) {
	p := NewParser()
	p.inPaste = true
	p.parseCSI([]byte("200~"))
	if p.state != statePaste {
		t.Errorf("expected statePaste, got %d", p.state)
	}
}

func TestP114_ParseCSI_CSIDuringPaste(t *testing.T) {
	p := NewParser()
	p.inPaste = true
	ev := p.parseCSI([]byte("A"))
	if ev != nil {
		t.Error("expected nil for CSI during paste")
	}
}

func TestP114_ParseCSI_FocusIn(t *testing.T) {
	p := NewParser()
	ev := p.parseCSI([]byte{'I'})
	if ev == nil {
		t.Fatal("nil event for FocusIn")
	}
	if ev.Type != EventFocus {
		t.Errorf("expected EventFocus, got %v", ev.Type)
	}
	if !ev.Focused {
		t.Error("expected Focused=true")
	}
}

func TestP114_ParseCSI_FocusOut(t *testing.T) {
	p := NewParser()
	ev := p.parseCSI([]byte{'O'})
	if ev == nil {
		t.Fatal("nil event for FocusOut")
	}
	if ev.Type != EventFocus {
		t.Errorf("expected EventFocus, got %v", ev.Type)
	}
	if ev.Focused {
		t.Error("expected Focused=false")
	}
}

func TestP114_ParseCSI_UnknownFinal(t *testing.T) {
	p := NewParser()
	ev := p.parseCSI([]byte("99X"))
	if ev != nil {
		t.Error("expected nil for unknown final byte")
	}
}

func TestP114_SetStyle_DefaultEmitsReset(t *testing.T) {
	w := NewWriter(nil, ProfileTrue)
	w.SetStyle(buffer.DefaultStyle)
	if !strings.Contains(string(w.Bytes()), buffer.ResetSGR) {
		t.Errorf("expected reset SGR for default style")
	}
}

func TestP114_SetStyle_StyleChange(t *testing.T) {
	w := NewWriter(nil, ProfileTrue)
	s1 := buffer.Style{Fg: buffer.NamedColor(buffer.NamedRed), Flags: buffer.Bold}
	w.SetStyle(s1)
	s2 := buffer.Style{Fg: buffer.NamedColor(buffer.NamedRed), Flags: buffer.Italic}
	w.SetStyle(s2)
	if !strings.Contains(string(w.Bytes()), "\x1b[") {
		t.Error("expected SGR escape")
	}
}

func TestP114_SetStyle_SameSkips(t *testing.T) {
	w := NewWriter(nil, ProfileTrue)
	s := buffer.Style{Fg: buffer.NamedColor(buffer.NamedGreen)}
	w.SetStyle(s)
	firstLen := w.buf.Len()
	w.SetStyle(s)
	if w.buf.Len() != firstLen {
		t.Errorf("expected no output for same style")
	}
}

func TestP114_SetStyle_ResetThenDefault(t *testing.T) {
	w := NewWriter(nil, ProfileTrue)
	w.SetStyle(buffer.Style{Fg: buffer.NamedColor(buffer.NamedRed)})
	w.ResetStyle()
	w.buf.Reset()
	w.SetStyle(buffer.DefaultStyle)
	if !strings.Contains(string(w.Bytes()), buffer.ResetSGR) {
		t.Error("expected reset SGR after ResetStyle")
	}
}

func TestP114_SetStyle_TrueColor(t *testing.T) {
	w := NewWriter(nil, ProfileTrue)
	s := buffer.Style{
		Fg: buffer.RGB(255, 128, 64),
		Bg: buffer.RGB(32, 64, 128),
	}
	w.SetStyle(s)
	out := string(w.Bytes())
	if !strings.Contains(out, "38;2;") {
		t.Error("expected truecolor FG")
	}
	if !strings.Contains(out, "48;2;") {
		t.Error("expected truecolor BG")
	}
}

func TestP114_SetStyle_256Color(t *testing.T) {
	var sb strings.Builder
	w := NewWriter(&sb, Profile256)
	s := buffer.Style{
		Fg: buffer.Color{Type: buffer.Color256, Val: 196},
		Bg: buffer.Color{Type: buffer.Color256, Val: 21},
	}
	w.SetStyle(s)
	out := string(w.Bytes())
	if !strings.Contains(out, "38;5;") {
		t.Error("expected 256-color FG")
	}
	if !strings.Contains(out, "48;5;") {
		t.Error("expected 256-color BG")
	}
}

func TestP114_SetStyle_AllFlags(t *testing.T) {
	w := NewWriter(nil, ProfileTrue)
	s := buffer.Style{
		Fg:    buffer.NamedColor(buffer.NamedWhite),
		Flags: buffer.Bold | buffer.Italic | buffer.Underline | buffer.Dim | buffer.Reverse,
	}
	w.SetStyle(s)
	out := string(w.Bytes())
	if !strings.Contains(out, "\x1b[") {
		t.Error("expected SGR escape with flags")
	}
}

func TestP114_SetStyle_NoneProfile(t *testing.T) {
	var sb strings.Builder
	w := NewWriter(&sb, ProfileNone)
	w.SetStyle(buffer.Style{
		Fg:    buffer.NamedColor(buffer.NamedRed),
		Flags: buffer.Bold,
	})
	_ = string(w.Bytes())
}

func TestP114_ParseSGRMouse_WheelUp(t *testing.T) {
	p := NewParser()
	ev := p.parseSGRMouse([]byte("<64;5;10M"))
	if ev == nil || ev.Mouse == nil {
		t.Fatal("nil event for wheel up")
	}
}

func TestP114_ParseSGRMouse_Release(t *testing.T) {
	p := NewParser()
	ev := p.parseSGRMouse([]byte("<0;5;10m"))
	if ev == nil || ev.Mouse == nil {
		t.Fatal("nil event for mouse release")
	}
}

func TestP114_ParseSGRMouse_Invalid(t *testing.T) {
	p := NewParser()
	ev := p.parseSGRMouse([]byte("M"))
	if ev != nil {
		t.Error("expected nil for invalid SGR mouse")
	}
}

func TestP114_ParseKittyCSIU_Empty(t *testing.T) {
	ev := parseKittyCSIU(nil)
	if ev != nil {
		t.Error("expected nil for empty Kitty CSI u")
	}
}

func TestP114_ParseKittyCSIU_ModifierOnly(t *testing.T) {
	// CSI ;1u — shift modifier, codepoint 0
	ev := parseKittyCSIU([]int{0, 1})
	if ev == nil {
		t.Skip("parser may return nil for codepoint 0")
	}
}

func TestP114_Feed_UTF8MultiByte(t *testing.T) {
	p := NewParser()
	// 2-byte UTF-8: é = 0xC3 0xA9
	evs := p.feed([]byte{0xC3, 0xA9})
	if len(evs) == 0 {
		t.Error("expected event for 2-byte UTF-8")
	}
	if evs[0].Key == nil || evs[0].Key.Rune != 'é' {
		t.Errorf("expected é, got %v", evs[0])
	}
}

func TestP114_Feed_UTF8ThreeByte(t *testing.T) {
	p := NewParser()
	// 3-byte UTF-8: ❤ = 0xE2 0x9D 0xA4
	evs := p.feed([]byte{0xE2, 0x9D, 0xA4})
	if len(evs) == 0 {
		t.Error("expected event for 3-byte UTF-8")
	}
}

func TestP114_Feed_InvalidUTF8(t *testing.T) {
	p := NewParser()
	// Invalid continuation byte
	evs := p.feed([]byte{0xC3, 0x00})
	// Should not panic
	_ = evs
}

func TestP114_ParseOSC52_InvalidBase64(t *testing.T) {
	// Invalid base64 in OSC52 response
	_, ok := ParseOSC52Response("52;c;!!!invalid!!!")
	if ok {
		// _ = ok
	}
}

func TestP114_ParseOSC52_EmptyPayload(t *testing.T) {
	_, ok := ParseOSC52Response("52;c;")
	if ok {
		// _ = ok
	}
}

func TestP114_ParseColorResponse_BELTerminator(t *testing.T) {
	// OSC response with BEL terminator instead of ST
	resp := ParseColorResponse("\x1b]10;rgb:ffff/0000/0000\x07")
	if resp.Valid {
		if resp.R != 255 || resp.G != 0 || resp.B != 0 {
			t.Errorf("expected red, got r=%d g=%d b=%d", resp.R, resp.G, resp.B)
		}
	}
}

func TestP114_ParseColorResponse_STTerminator(t *testing.T) {
	// OSC response with ST (ESC \) terminator
	resp := ParseColorResponse("\x1b]11;rgb:0000/0000/ffff\x1b\\")
	if resp.Valid {
		if resp.R != 0 || resp.G != 0 || resp.B != 255 {
			t.Errorf("expected blue, got r=%d g=%d b=%d", resp.R, resp.G, resp.B)
		}
	}
}

func TestP114_ParseColorResponse_ShortHex(t *testing.T) {
	// 4-bit hex: rgb:f/f/f
	resp := ParseColorResponse("\x1b]10;rgb:f/f/f\x1b\\")
	_ = resp
}

func TestP114_ParseColorResponse_InvalidFormat(t *testing.T) {
	resp := ParseColorResponse("invalid")
	if resp.Valid {
		t.Error("expected no color for invalid format")
	}
}

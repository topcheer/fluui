package term

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP159_StyleEquals_NotSet(t *testing.T) {
	w := NewWriter(nil, ProfileTrue)
	if w.StyleEquals(buffer.Style{}) {
		t.Error("expected false when style not set")
	}
}

func TestP159_StyleEquals_SameStyle(t *testing.T) {
	w := NewWriter(nil, ProfileTrue)
	style := buffer.Style{
		Fg:    buffer.NamedColor(buffer.NamedRed),
		Flags: buffer.Bold,
	}
	w.SetStyle(style)
	if !w.StyleEquals(style) {
		t.Error("expected true for same style")
	}
}

func TestP159_StyleEquals_DifferentStyle(t *testing.T) {
	w := NewWriter(nil, ProfileTrue)
	w.SetStyle(buffer.Style{
		Fg: buffer.NamedColor(buffer.NamedRed),
	})
	other := buffer.Style{
		Fg: buffer.NamedColor(buffer.NamedBlue),
	}
	if w.StyleEquals(other) {
		t.Error("expected false for different style")
	}
}

func TestP159_StyleEquals_DefaultStyle(t *testing.T) {
	w := NewWriter(nil, ProfileTrue)
	w.SetStyle(buffer.Style{})
	if !w.StyleEquals(buffer.Style{}) {
		t.Error("expected true for default style")
	}
}

func TestP159_StyleEquals_AfterResetStyle(t *testing.T) {
	w := NewWriter(nil, ProfileTrue)
	w.SetStyle(buffer.Style{Flags: buffer.Bold})
	w.ResetStyle()
	if w.StyleEquals(buffer.Style{Flags: buffer.Bold}) {
		t.Error("expected false after reset")
	}
}

// Also test some parseCSI branches that may be uncovered
func TestP159_ParseCSI_F6(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[17~"))
	if len(evs) == 0 {
		t.Fatal("expected event for F6")
	}
	if evs[0].Key.Key != KeyF6 {
		t.Errorf("expected KeyF6, got %d", evs[0].Key)
	}
}

func TestP159_ParseCSI_F7(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[18~"))
	if len(evs) == 0 {
		t.Fatal("expected event for F7")
	}
	if evs[0].Key.Key != KeyF7 {
		t.Errorf("expected KeyF7, got %d", evs[0].Key)
	}
}

func TestP159_ParseCSI_F8(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[19~"))
	if len(evs) == 0 {
		t.Fatal("expected event for F8")
	}
	if evs[0].Key.Key != KeyF8 {
		t.Errorf("expected KeyF8, got %d", evs[0].Key)
	}
}

func TestP159_ParseCSI_F9(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[20~"))
	if len(evs) == 0 {
		t.Fatal("expected event for F9")
	}
	if evs[0].Key.Key != KeyF9 {
		t.Errorf("expected KeyF9, got %d", evs[0].Key)
	}
}

func TestP159_ParseCSI_F10(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[21~"))
	if len(evs) == 0 {
		t.Fatal("expected event for F10")
	}
	if evs[0].Key.Key != KeyF10 {
		t.Errorf("expected KeyF10, got %d", evs[0].Key)
	}
}

func TestP159_ParseCSI_F11(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[23~"))
	if len(evs) == 0 {
		t.Fatal("expected event for F11")
	}
	if evs[0].Key.Key != KeyF11 {
		t.Errorf("expected KeyF11, got %d", evs[0].Key)
	}
}

func TestP159_ParseCSI_F12(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[24~"))
	if len(evs) == 0 {
		t.Fatal("expected event for F12")
	}
	if evs[0].Key.Key != KeyF12 {
		t.Errorf("expected KeyF12, got %d", evs[0].Key)
	}
}

func TestP159_ParseCSI_ShiftUp(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[1;2A"))
	if len(evs) == 0 {
		t.Fatal("expected event")
	}
	if evs[0].Key.Key != KeyUp {
		t.Errorf("expected KeyUp, got %d", evs[0].Key)
	}
	if evs[0].Key.Modifiers&ModShift == 0 {
		t.Error("expected Shift modifier")
	}
}

func TestP159_ParseCSI_CtrlUp(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[1;5A"))
	if len(evs) == 0 {
		t.Fatal("expected event")
	}
	if evs[0].Key.Key != KeyUp {
		t.Errorf("expected KeyUp, got %d", evs[0].Key)
	}
	if evs[0].Key.Modifiers&ModCtrl == 0 {
		t.Error("expected Ctrl modifier")
	}
}

func TestP159_ParseCSI_AltUp(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[1;3A"))
	if len(evs) == 0 {
		t.Fatal("expected event")
	}
	if evs[0].Key.Key != KeyUp {
		t.Errorf("expected KeyUp, got %d", evs[0].Key)
	}
	if evs[0].Key.Modifiers&ModAlt == 0 {
		t.Error("expected Alt modifier")
	}
}

func TestP159_ParseCSI_Insert(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[2~"))
	if len(evs) == 0 {
		t.Fatal("expected event")
	}
	if evs[0].Key.Key != KeyInsert {
		t.Errorf("expected KeyInsert, got %d", evs[0].Key)
	}
}

func TestP159_ParseCSI_Delete(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[3~"))
	if len(evs) == 0 {
		t.Fatal("expected event")
	}
	if evs[0].Key.Key != KeyDelete {
		t.Errorf("expected KeyDelete, got %d", evs[0].Key)
	}
}

func TestP159_ParseCSI_PageUp(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[5~"))
	if len(evs) == 0 {
		t.Fatal("expected event")
	}
	if evs[0].Key.Key != KeyPageUp {
		t.Errorf("expected KeyPageUp, got %d", evs[0].Key)
	}
}

func TestP159_ParseCSI_PageDown(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[6~"))
	if len(evs) == 0 {
		t.Fatal("expected event")
	}
	if evs[0].Key.Key != KeyPageDown {
		t.Errorf("expected KeyPageDown, got %d", evs[0].Key)
	}
}

func TestP159_ParseCSI_F5(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[15~"))
	if len(evs) == 0 {
		t.Fatal("expected event for F5")
	}
	if evs[0].Key.Key != KeyF5 {
		t.Errorf("expected KeyF5, got %d", evs[0].Key)
	}
}

func TestP159_ParseCSI_Home(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[H"))
	if len(evs) == 0 {
		t.Fatal("expected event")
	}
	if evs[0].Key.Key != KeyHome {
		t.Errorf("expected KeyHome, got %d", evs[0].Key)
	}
}

func TestP159_ParseCSI_End(t *testing.T) {
	p := NewParser()
	evs := p.Feed([]byte("\x1b[F"))
	if len(evs) == 0 {
		t.Fatal("expected event")
	}
	if evs[0].Key.Key != KeyEnd {
		t.Errorf("expected KeyEnd, got %d", evs[0].Key)
	}
}
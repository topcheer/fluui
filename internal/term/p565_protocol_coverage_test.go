package term

import (
	"strings"
	"testing"
)

func TestSetCursorColor(t *testing.T) {
	s := SetCursorColor("#ff0000")
	if !strings.Contains(s, "]12;") || !strings.Contains(s, "#ff0000") {
		t.Errorf("SetCursorColor = %q", s)
	}
}

func TestResetCursorColor(t *testing.T) {
	s := ResetCursorColor()
	if s != "\x1b]112\x07" {
		t.Errorf("ResetCursorColor = %q, want \\x1b]112\\x07", s)
	}
}

func TestNotifySimple(t *testing.T) {
	s := NotifySimple("Hello")
	if !strings.Contains(s, "]9;") || !strings.Contains(s, "Hello") {
		t.Errorf("NotifySimple = %q", s)
	}
}

func TestNotifyURxvt(t *testing.T) {
	s := NotifyURxvt("Title", "Body")
	if !strings.Contains(s, "]777;") || !strings.Contains(s, "Title") || !strings.Contains(s, "Body") {
		t.Errorf("NotifyURxvt = %q", s)
	}
}

func TestSetITermMark(t *testing.T) {
	s := SetITermMark()
	if !strings.Contains(s, "1337;SetMark") {
		t.Errorf("SetITermMark = %q", s)
	}
}

func TestStealFocus(t *testing.T) {
	s := StealFocus()
	if !strings.Contains(s, "1337;StealFocus") {
		t.Errorf("StealFocus = %q", s)
	}
}

func TestSetProtected(t *testing.T) {
	s := SetProtected()
	if s != "\x1b[1\"q" {
		t.Errorf("SetProtected = %q", s)
	}
}

func TestSetUnprotected(t *testing.T) {
	s := SetUnprotected()
	if s != "\x1b[0\"q" {
		t.Errorf("SetUnprotected = %q", s)
	}
}

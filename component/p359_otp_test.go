package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestP359_OTPInput_Create(t *testing.T) {
	o := NewOTPInput(6)
	if o.Length() != 6 {
		t.Errorf("length = %d, want 6", o.Length())
	}
	if o.Value() != "\x00\x00\x00\x00\x00\x00" {
		t.Error("expected empty value")
	}
}

func TestP359_OTPInput_ClampLength(t *testing.T) {
	o := NewOTPInput(0)
	if o.Length() != 1 {
		t.Errorf("length 0 → %d, want 1", o.Length())
	}
	o2 := NewOTPInput(99)
	if o2.Length() != 16 {
		t.Errorf("length 99 → %d, want 16", o2.Length())
	}
}

func TestP359_OTPInput_HandleKeyDigit(t *testing.T) {
	o := NewOTPInput(4)
	if !o.HandleKey(&term.KeyEvent{Rune: '1'}) {
		t.Error("digit should be handled")
	}
	if !o.HandleKey(&term.KeyEvent{Rune: '2'}) {
		t.Error("digit should be handled")
	}
	v := o.Value()
	if v[0] != '1' || v[1] != '2' {
		t.Errorf("values = %q", v[:2])
	}
}

func TestP359_OTPInput_IsFilled(t *testing.T) {
	o := NewOTPInput(3)
	o.HandleKey(&term.KeyEvent{Rune: '1'})
	o.HandleKey(&term.KeyEvent{Rune: '2'})
	if o.IsFilled() {
		t.Error("should not be filled with 2 of 3")
	}
	o.HandleKey(&term.KeyEvent{Rune: '3'})
	if !o.IsFilled() {
		t.Error("should be filled with 3 of 3")
	}
}

func TestP359_OTPInput_Backspace(t *testing.T) {
	o := NewOTPInput(3)
	o.HandleKey(&term.KeyEvent{Rune: '1'})
	o.HandleKey(&term.KeyEvent{Rune: '2'})
	o.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	v := o.Value()
	if v[1] != 0 {
		t.Errorf("expected empty after backspace, got %q", string(v[1]))
	}
}

func TestP359_OTPInput_LeftRight(t *testing.T) {
	o := NewOTPInput(4)
	o.HandleKey(&term.KeyEvent{Rune: '1'})
	o.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	o.HandleKey(&term.KeyEvent{Key: term.KeyRight})
}

func TestP359_OTPInput_SetValue(t *testing.T) {
	o := NewOTPInput(4)
	o.SetValue("1234")
	if !o.IsFilled() {
		t.Error("should be filled after SetValue")
	}
	v := o.Value()
	if v != "1234" {
		t.Errorf("value = %q, want 1234", v)
	}
}

func TestP359_OTPInput_SetValue_Long(t *testing.T) {
	o := NewOTPInput(3)
	o.SetValue("ABCDEF") // only 3 fit
	if !o.IsFilled() {
		t.Error("should be filled")
	}
}

func TestP359_OTPInput_Clear(t *testing.T) {
	o := NewOTPInput(3)
	o.HandleKey(&term.KeyEvent{Rune: '1'})
	o.Clear()
	if o.IsFilled() {
		t.Error("should not be filled after clear")
	}
}

func TestP359_OTPInput_HandleKey_Nil(t *testing.T) {
	o := NewOTPInput(3)
	if o.HandleKey(nil) {
		t.Error("nil should return false")
	}
}

func TestP359_OTPInput_Letters(t *testing.T) {
	o := NewOTPInput(3)
	if !o.HandleKey(&term.KeyEvent{Rune: 'A'}) {
		t.Error("letter should be handled")
	}
	if !o.HandleKey(&term.KeyEvent{Rune: 'b'}) {
		t.Error("lowercase letter should be handled")
	}
}

func TestP359_OTPInput_Measure(t *testing.T) {
	o := NewOTPInput(4)
	s := o.Measure(Constraints{MaxWidth: 40, MaxHeight: 1})
	if s.H != 1 {
		t.Errorf("height = %d, want 1", s.H)
	}
	if s.W < 10 {
		t.Errorf("width = %d, too small", s.W)
	}
}

func TestP359_OTPInput_Paint(t *testing.T) {
	o := NewOTPInput(4)
	o.HandleKey(&term.KeyEvent{Rune: '1'})
	o.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	o.Paint(buf)
	if buf.GetCell(0, 0).Rune == 0 {
		t.Error("expected non-empty cell")
	}
}

func TestP359_OTPInput_Paint_ZeroBounds(t *testing.T) {
	o := NewOTPInput(4)
	o.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(20, 1)
	o.Paint(buf)
}

func BenchmarkOTPInput_Paint(b *testing.B) {
	o := NewOTPInput(6)
	o.HandleKey(&term.KeyEvent{Rune: '1'})
	o.HandleKey(&term.KeyEvent{Rune: '2'})
	o.HandleKey(&term.KeyEvent{Rune: '3'})
	o.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		o.Paint(buf)
	}
}

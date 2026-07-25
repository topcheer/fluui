package component

import (
	"os"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP366_TerminalProfile_Create(t *testing.T) {
	tp := NewTerminalProfile()
	_ = tp
}

func TestP366_TerminalProfile_EnvOr(t *testing.T) {
	os.Setenv("FLUUI_TEST_VAR", "testval")
	if v := envOr("FLUUI_TEST_VAR"); v != "testval" {
		t.Errorf("envOr = %q, want testval", v)
	}
	if v := envOr("FLUUI_NONEXISTENT_VAR"); v != "-" {
		t.Errorf("envOr = %q, want -", v)
	}
	os.Unsetenv("FLUUI_TEST_VAR")
}

func TestP366_TerminalProfile_HasColorSupport(t *testing.T) {
	os.Setenv("COLORTERM", "truecolor")
	if v := hasColorSupport(); v != "24-bit (TrueColor)" {
		t.Errorf("color = %q", v)
	}
	os.Setenv("COLORTERM", "")
	os.Setenv("TERM", "xterm-256color")
	if v := hasColorSupport(); v != "256-color" {
		t.Errorf("color = %q", v)
	}
	os.Setenv("TERM", "dumb")
	if v := hasColorSupport(); v != "16-color" {
		t.Errorf("color = %q", v)
	}
	os.Unsetenv("COLORTERM")
	os.Unsetenv("TERM")
}

func TestP366_TerminalProfile_Contains(t *testing.T) {
	if !contains("xterm-256color", "256") {
		t.Error("should contain 256")
	}
	if contains("dumb", "256") {
		t.Error("dumb should not contain 256")
	}
	if !contains("test", "") {
		t.Error("empty substr should match")
	}
}

func TestP366_TerminalProfile_Measure(t *testing.T) {
	tp := NewTerminalProfile()
	s := tp.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W != 50 {
		t.Errorf("width = %d, want 50", s.W)
	}
	if s.H != 8 {
		t.Errorf("height = %d, want 8", s.H)
	}
}

func TestP366_TerminalProfile_Paint(t *testing.T) {
	os.Setenv("TERM", "xterm-256color")
	os.Setenv("COLORTERM", "truecolor")
	defer os.Unsetenv("TERM")
	defer os.Unsetenv("COLORTERM")

	tp := NewTerminalProfile()
	tp.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	tp.Paint(buf)

	if buf.GetCell(0, 0).Rune == 0 {
		t.Error("expected non-empty cell")
	}
}

func TestP366_TerminalProfile_Paint_ZeroBounds(t *testing.T) {
	tp := NewTerminalProfile()
	tp.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(50, 10)
	tp.Paint(buf)
}

func TestP366_TerminalProfile_Paint_NoEnv(t *testing.T) {
	os.Unsetenv("TERM")
	os.Unsetenv("TERM_PROGRAM")
	os.Unsetenv("COLORTERM")
	os.Unsetenv("SHELL")
	os.Unsetenv("LANG")

	tp := NewTerminalProfile()
	tp.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	tp.Paint(buf)
}

func BenchmarkTerminalProfile_Paint(b *testing.B) {
	os.Setenv("TERM", "xterm-256color")
	os.Setenv("COLORTERM", "truecolor")
	defer os.Unsetenv("TERM")
	defer os.Unsetenv("COLORTERM")

	tp := NewTerminalProfile()
	tp.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tp.Paint(buf)
	}
}

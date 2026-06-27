package buffer

import (
	"testing"
)

func TestStyleSGR_BoldItalic(t *testing.T) {
	s := Style{Flags: Bold | Italic, Fg: NoColor(), Bg: NoColor()}
	sgr := s.SGRSequence()
	// Bold=1, Italic=3, FG default=39, BG default=49
	expected := "1;3;39;49"
	if sgr != expected {
		t.Errorf("SGR = %q, want %q", sgr, expected)
	}
}

func TestStyleSGR_AllFlags(t *testing.T) {
	s := Style{
		Flags: Bold | Dim | Italic | Underline | Blink | Reverse | Strikethrough,
		Fg:    NoColor(),
		Bg:    NoColor(),
	}
	sgr := s.SGRSequence()
	expected := "1;2;3;4;5;7;9;39;49"
	if sgr != expected {
		t.Errorf("SGR = %q, want %q", sgr, expected)
	}
}

func TestStyleSGR_Empty(t *testing.T) {
	s := DefaultStyle
	sgr := s.SGRSequence()
	// No flags, default colors
	expected := "39;49"
	if sgr != expected {
		t.Errorf("SGR = %q, want %q", sgr, expected)
	}
}

func TestStyleEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b Style
		want bool
	}{
		{
			name: "identical",
			a:    Style{Fg: RGB(1, 2, 3), Bg: RGB(4, 5, 6), Flags: Bold},
			b:    Style{Fg: RGB(1, 2, 3), Bg: RGB(4, 5, 6), Flags: Bold},
			want: true,
		},
		{
			name: "different fg",
			a:    Style{Fg: RGB(1, 2, 3)},
			b:    Style{Fg: RGB(1, 2, 4)},
			want: false,
		},
		{
			name: "different bg",
			a:    Style{Bg: RGB(1, 2, 3)},
			b:    Style{Bg: RGB(1, 2, 4)},
			want: false,
		},
		{
			name: "different flags",
			a:    Style{Flags: Bold},
			b:    Style{Flags: Italic},
			want: false,
		},
		{
			name: "both default",
			a:    DefaultStyle,
			b:    DefaultStyle,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Equal(tt.b); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStyleWithFg(t *testing.T) {
	s := DefaultStyle.WithFg(RGB(255, 0, 0))
	if !s.Fg.Equal(RGB(255, 0, 0)) {
		t.Errorf("Fg = %v, want RGB(255,0,0)", s.Fg)
	}
	// Original should be unchanged
	if !DefaultStyle.Fg.Equal(NoColor()) {
		t.Error("DefaultStyle was mutated")
	}
}

func TestStyleWithBg(t *testing.T) {
	s := DefaultStyle.WithBg(RGB(0, 255, 0))
	if !s.Bg.Equal(RGB(0, 255, 0)) {
		t.Errorf("Bg = %v, want RGB(0,255,0)", s.Bg)
	}
}

func TestStyleAddFlags(t *testing.T) {
	s := DefaultStyle.AddFlags(Bold)
	if !s.HasFlag(Bold) {
		t.Error("HasFlag(Bold) = false, want true")
	}
	s = s.AddFlags(Italic)
	if !s.HasFlag(Bold) || !s.HasFlag(Italic) {
		t.Error("AddFlags did not accumulate")
	}
	if s.HasFlag(Underline) {
		t.Error("HasFlag(Underline) should be false")
	}
}

func TestStyleWithFlags(t *testing.T) {
	s := DefaultStyle.AddFlags(Bold | Italic).WithFlags(Underline)
	if !s.HasFlag(Underline) {
		t.Error("WithFlags should replace flags")
	}
	if s.HasFlag(Bold) || s.HasFlag(Italic) {
		t.Error("WithFlags should clear old flags")
	}
}

func TestStyleHasFlag(t *testing.T) {
	s := Style{Flags: Bold | Underline}
	if !s.HasFlag(Bold) {
		t.Error("HasFlag(Bold) = false")
	}
	if !s.HasFlag(Underline) {
		t.Error("HasFlag(Underline) = false")
	}
	if s.HasFlag(Italic) {
		t.Error("HasFlag(Italic) = true, should be false")
	}
}

func TestResetSGR(t *testing.T) {
	if ResetSGR != "\x1b[0m" {
		t.Errorf("ResetSGR = %q, want %q", ResetSGR, "\x1b[0m")
	}
}

func TestStyleSGR_WithTrueColor(t *testing.T) {
	s := Style{
		Fg:    RGB(255, 128, 0),
		Bg:    RGB(40, 42, 54),
		Flags: Bold,
	}
	sgr := s.SGRSequence()
	expected := "1;38;2;255;128;0;48;2;40;42;54"
	if sgr != expected {
		t.Errorf("SGR = %q, want %q", sgr, expected)
	}
}

func TestStyleSGR_WithNamedColor(t *testing.T) {
	s := Style{
		Fg:    NamedColor(NamedRed),
		Bg:    NamedColor(NamedBlue),
		Flags: 0,
	}
	sgr := s.SGRSequence()
	expected := "31;44"
	if sgr != expected {
		t.Errorf("SGR = %q, want %q", sgr, expected)
	}
}

func TestStyleSGR_With256Color(t *testing.T) {
	s := Style{
		Fg:    Color256Val(196),
		Bg:    Color256Val(234),
		Flags: 0,
	}
	sgr := s.SGRSequence()
	expected := "38;5;196;48;5;234"
	if sgr != expected {
		t.Errorf("SGR = %q, want %q", sgr, expected)
	}
}

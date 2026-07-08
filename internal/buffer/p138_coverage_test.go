package buffer

import "testing"

// === Color.String (83.3% → 100%) ===

func TestP138_ColorString_AllTypes(t *testing.T) {
	cases := []struct {
		name string
		color Color
		want  string
	}{
		{"none", Color{Type: ColorNone}, "default"},
		{"named", Color{Type: ColorNamed, Val: 5}, "named(5)"},
		{"256color", Color{Type: Color256, Val: 42}, "256(42)"},
		{"truecolor_red", RGB(255, 0, 0), "#ff0000"},
		{"truecolor_green", RGB(0, 255, 0), "#00ff00"},
		{"truecolor_blue", RGB(0, 0, 255), "#0000ff"},
		{"truecolor_white", RGB(255, 255, 255), "#ffffff"},
		{"truecolor_black", RGB(0, 0, 0), "#000000"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.color.String()
			if got != tc.want {
				t.Errorf("Color{Type:%d, Val:%d}.String() = %q, want %q",
					tc.color.Type, tc.color.Val, got, tc.want)
			}
		})
	}
}

// === Cell.Equal (90.9% → 100%) ===

func TestP138_CellEqual_AllFields(t *testing.T) {
	base := Cell{Rune: 'A', Width: 1, Fg: RGB(255, 0, 0), Bg: RGB(0, 0, 0), Flags: Bold}
	cases := []struct {
		name string
		other Cell
		want  bool
	}{
		{"identical", base, true},
		{"different_rune", Cell{Rune: 'B', Width: 1, Fg: base.Fg, Bg: base.Bg, Flags: base.Flags}, false},
		{"different_width", Cell{Rune: 'A', Width: 2, Fg: base.Fg, Bg: base.Bg, Flags: base.Flags}, false},
		{"different_fg", Cell{Rune: 'A', Width: 1, Fg: RGB(0, 255, 0), Bg: base.Bg, Flags: base.Flags}, false},
		{"different_bg", Cell{Rune: 'A', Width: 1, Fg: base.Fg, Bg: RGB(0, 0, 255), Flags: base.Flags}, false},
		{"different_flags", Cell{Rune: 'A', Width: 1, Fg: base.Fg, Bg: base.Bg, Flags: Italic}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := base.Equal(tc.other)
			if got != tc.want {
				t.Errorf("Equal(%v) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

// === DrawTextClamped (92.9% → 100%) ===

func TestP138_DrawTextClamped_EdgeCases(t *testing.T) {
	style := Style{Fg: RGB(255, 255, 255)}

	// Exact width.
	buf := NewBuffer(10, 1)
	n := buf.DrawTextClamped(0, 0, "1234567890", style)
	if n != 10 {
		t.Errorf("exact width: got %d, want 10", n)
	}

	// Overflow (text longer than buffer width).
	buf = NewBuffer(5, 1)
	n = buf.DrawTextClamped(0, 0, "1234567890", style)
	if n != 5 {
		t.Errorf("overflow: got %d, want 5", n)
	}

	// Empty text.
	buf = NewBuffer(10, 1)
	n = buf.DrawTextClamped(0, 0, "", style)
	if n != 0 {
		t.Errorf("empty text: got %d, want 0", n)
	}

	// Start past edge — returns original x (doesn't draw).
	buf = NewBuffer(10, 1)
	n = buf.DrawTextClamped(15, 0, "test", style)
	if n != 15 {
		t.Errorf("past edge x: got %d, want 15", n)
	}
}

// === appendFG/appendBG (92.9% → 100%) ===

func TestP138_AppendFG_AllTypes(t *testing.T) {
	cases := []Color{
		{Type: ColorNone},
		{Type: ColorNamed, Val: 3},
		{Type: Color256, Val: 42},
		RGB(128, 64, 32),
	}
	for _, c := range cases {
		b := c.appendFG(nil)
		if len(b) == 0 {
			t.Errorf("appendFG returned empty for type %d", c.Type)
		}
		b = c.appendBG(nil)
		if len(b) == 0 {
			t.Errorf("appendBG returned empty for type %d", c.Type)
		}
	}
}

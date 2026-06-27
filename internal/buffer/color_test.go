package buffer

import (
	"testing"
)

func TestRGB(t *testing.T) {
	c := RGB(255, 128, 1)
	if c.Type != ColorTrue {
		t.Errorf("Type = %v, want ColorTrue", c.Type)
	}
	if c.R() != 255 {
		t.Errorf("R() = %d, want 255", c.R())
	}
	if c.G() != 128 {
		t.Errorf("G() = %d, want 128", c.G())
	}
	if c.B() != 1 {
		t.Errorf("B() = %d, want 1", c.B())
	}
}

func TestRGB_Packing(t *testing.T) {
	c := RGB(0xAB, 0xCD, 0xEF)
	if c.Val != 0xABCDEF {
		t.Errorf("Val = 0x%06X, want 0xABCDEF", c.Val)
	}
}

func TestNoColor(t *testing.T) {
	c := NoColor()
	if c.Type != ColorNone {
		t.Errorf("Type = %v, want ColorNone", c.Type)
	}
	if !c.IsDefault() {
		t.Error("IsDefault() = false, want true")
	}
}

func TestNamedColor(t *testing.T) {
	c := NamedColor(NamedRed)
	if c.Type != ColorNamed {
		t.Errorf("Type = %v, want ColorNamed", c.Type)
	}
	if c.Val != NamedRed {
		t.Errorf("Val = %d, want %d", c.Val, NamedRed)
	}
	if c.IsDefault() {
		t.Error("Named color should not be default")
	}
}

func TestColor256Val(t *testing.T) {
	c := Color256Val(196)
	if c.Type != Color256 {
		t.Errorf("Type = %v, want Color256", c.Type)
	}
	if c.Val != 196 {
		t.Errorf("Val = %d, want 196", c.Val)
	}
}

func TestHex(t *testing.T) {
	tests := []struct {
		input string
		valid bool
		val   uint32
	}{
		{"#ff6600", true, 0xFF6600},
		{"ff6600", true, 0xFF6600},
		{"#FF6600", true, 0xFF6600},
		{"#abc", false, 0}, // too short
		{"#gggggg", false, 0}, // invalid hex
		{"", false, 0}, // empty
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			c := Hex(tt.input)
			if tt.valid {
				if c.Type != ColorTrue {
					t.Errorf("Type = %v, want ColorTrue", c.Type)
				}
				if c.Val != tt.val {
					t.Errorf("Val = 0x%06X, want 0x%06X", c.Val, tt.val)
				}
			} else {
				if c.Type != ColorNone {
					t.Errorf("invalid hex should return ColorNone, got %v", c.Type)
				}
			}
		})
	}
}

func TestColorEqual_Comprehensive(t *testing.T) {
	tests := []struct {
		name string
		a, b Color
		want bool
	}{
		{"both true color same", RGB(1, 2, 3), RGB(1, 2, 3), true},
		{"both true color diff", RGB(1, 2, 3), RGB(1, 2, 4), false},
		{"both none", NoColor(), NoColor(), true},
		{"none vs true", NoColor(), RGB(0, 0, 0), false},
		{"named same", NamedColor(1), NamedColor(1), true},
		{"named diff", NamedColor(1), NamedColor(2), false},
		{"256 same", Color256Val(10), Color256Val(10), true},
		{"256 diff", Color256Val(10), Color256Val(20), false},
		{"different types", NamedColor(1), Color256Val(1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Equal(tt.b); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColorFGSequence(t *testing.T) {
	tests := []struct {
		name string
		c    Color
		want string
	}{
		{"none", NoColor(), "39"},
		{"named black", NamedColor(NamedBlack), "30"},
		{"named red", NamedColor(NamedRed), "31"},
		{"named white", NamedColor(NamedWhite), "37"},
		{"named bright red", NamedColor(NamedBrightRed), "91"},
		{"named bright white", NamedColor(NamedBrightWhite), "97"},
		{"256", Color256Val(196), "38;5;196"},
		{"true color", RGB(255, 128, 0), "38;2;255;128;0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.FGSequence(); got != tt.want {
				t.Errorf("FGSequence() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestColorBGSequence(t *testing.T) {
	tests := []struct {
		name string
		c    Color
		want string
	}{
		{"none", NoColor(), "49"},
		{"named black", NamedColor(NamedBlack), "40"},
		{"named red", NamedColor(NamedRed), "41"},
		{"named white", NamedColor(NamedWhite), "47"},
		{"named bright red", NamedColor(NamedBrightRed), "101"},
		{"named bright white", NamedColor(NamedBrightWhite), "107"},
		{"256", Color256Val(196), "48;5;196"},
		{"true color", RGB(255, 128, 0), "48;2;255;128;0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.BGSequence(); got != tt.want {
				t.Errorf("BGSequence() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestColorString(t *testing.T) {
	tests := []struct {
		name string
		c    Color
		want string
	}{
		{"none", NoColor(), "default"},
		{"named", NamedColor(3), "named(3)"},
		{"256", Color256Val(42), "256(42)"},
		{"true color", RGB(0xFF, 0x66, 0x00), "#ff6600"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestColorIsDefault(t *testing.T) {
	if !NoColor().IsDefault() {
		t.Error("NoColor().IsDefault() should be true")
	}
	if RGB(0, 0, 0).IsDefault() {
		t.Error("RGB(0,0,0).IsDefault() should be false")
	}
}

func TestPredefinedColors(t *testing.T) {
	// Verify predefined convenience colors are correct type
	if Black.Type != ColorNamed || Black.Val != NamedBlack {
		t.Errorf("Black = %v, want named(0)", Black)
	}
	if Red.Type != ColorNamed || Red.Val != NamedRed {
		t.Errorf("Red = %v, want named(1)", Red)
	}
	if Default.Type != ColorNone {
		t.Errorf("Default = %v, want ColorNone", Default)
	}
}

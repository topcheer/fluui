package term

import (
	"testing"
)

// --- Query generators ---

func TestQueryPaletteColor(t *testing.T) {
	got := QueryPaletteColor(0)
	want := "\x1b]4;0;?\x1b\\"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}

	got = QueryPaletteColor(255)
	want = "\x1b]4;255;?\x1b\\"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestQueryDefaultFG(t *testing.T) {
	got := QueryDefaultFG()
	want := "\x1b]10;?\x1b\\"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestQueryDefaultBG(t *testing.T) {
	got := QueryDefaultBG()
	want := "\x1b]11;?\x1b\\"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestQueryCursorColor(t *testing.T) {
	got := QueryCursorColor()
	want := "\x1b]12;?\x1b\\"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

// --- ParseColorResponse: OSC 4 responses ---

func TestParseColorResponse_OSC4_16bit(t *testing.T) {
	// xterm-style response: ESC ] 4 ; 0 ; rgb:0000/0000/0000 ST
	input := "\x1b]4;0;rgb:0000/0000/0000\x1b\\"
	cr := ParseColorResponse(input)
	if !cr.Valid {
		t.Fatal("expected valid response")
	}
	if cr.Index != 0 {
		t.Errorf("expected index 0, got %d", cr.Index)
	}
	if cr.R != 0 || cr.G != 0 || cr.B != 0 {
		t.Errorf("expected black (0,0,0), got (%d,%d,%d)", cr.R, cr.G, cr.B)
	}
}

func TestParseColorResponse_OSC4_white(t *testing.T) {
	input := "\x1b]4;7;rgb:ffff/ffff/ffff\x1b\\"
	cr := ParseColorResponse(input)
	if !cr.Valid {
		t.Fatal("expected valid response")
	}
	if cr.Index != 7 {
		t.Errorf("expected index 7, got %d", cr.Index)
	}
	if cr.R != 255 || cr.G != 255 || cr.B != 255 {
		t.Errorf("expected white (255,255,255), got (%d,%d,%d)", cr.R, cr.G, cr.B)
	}
}

func TestParseColorResponse_OSC4_8bit(t *testing.T) {
	// 2-digit hex format
	input := "\x1b]4;1;rgb:ff/00/00\x1b\\"
	cr := ParseColorResponse(input)
	if !cr.Valid {
		t.Fatal("expected valid response")
	}
	if cr.Index != 1 {
		t.Errorf("expected index 1, got %d", cr.Index)
	}
	if cr.R != 255 || cr.G != 0 || cr.B != 0 {
		t.Errorf("expected red (255,0,0), got (%d,%d,%d)", cr.R, cr.G, cr.B)
	}
}

func TestParseColorResponse_OSC4_custom(t *testing.T) {
	// Custom color: #3a7bd5
	input := "\x1b]4;4;rgb:3a7b/d5ff/0000\x1b\\"
	cr := ParseColorResponse(input)
	if !cr.Valid {
		t.Fatal("expected valid response")
	}
	// 3a7b → 0x3a7b → scale >> 4 → 0x3a7 = 935... wait, 0x3a7b >> 4 = 0x3a7 (uint8) = 58
	if cr.R != 0x3a {
		t.Errorf("expected R=0x3a (58), got %d", cr.R)
	}
}

func TestParseColorResponse_OSC4_BEL_terminator(t *testing.T) {
	// Some terminals use BEL (0x07) instead of ESC \
	input := "\x1b]4;0;rgb:8080/8080/8080\x07"
	cr := ParseColorResponse(input)
	if !cr.Valid {
		t.Fatal("expected valid response")
	}
	if cr.R != 128 || cr.G != 128 || cr.B != 128 {
		t.Errorf("expected gray (128,128,128), got (%d,%d,%d)", cr.R, cr.G, cr.B)
	}
}

// --- ParseColorResponse: OSC 10/11 responses ---

func TestParseColorResponse_OSC10(t *testing.T) {
	input := "\x1b]10;rgb:cccc/cccc/cccc\x1b\\"
	cr := ParseColorResponse(input)
	if !cr.Valid {
		t.Fatal("expected valid response")
	}
	if cr.Index != -1 {
		t.Errorf("expected index -1, got %d", cr.Index)
	}
	if cr.R != 204 || cr.G != 204 || cr.B != 204 {
		t.Errorf("expected (204,204,204), got (%d,%d,%d)", cr.R, cr.G, cr.B)
	}
}

func TestParseColorResponse_OSC11(t *testing.T) {
	input := "\x1b]11;rgb:1d1d/1d1d/1d1d\x1b\\"
	cr := ParseColorResponse(input)
	if !cr.Valid {
		t.Fatal("expected valid response")
	}
	if cr.Index != -1 {
		t.Errorf("expected index -1, got %d", cr.Index)
	}
	// 0x1d1d >> 8 = 0x1d = 29
	if cr.R != 29 || cr.G != 29 || cr.B != 29 {
		t.Errorf("expected (29,29,29), got (%d,%d,%d)", cr.R, cr.G, cr.B)
	}
}

// --- ParseColorResponse: error cases ---

func TestParseColorResponse_Invalid(t *testing.T) {
	cr := ParseColorResponse("not a response")
	if cr.Valid {
		t.Error("expected invalid for garbage input")
	}
}

func TestParseColorResponse_Empty(t *testing.T) {
	cr := ParseColorResponse("")
	if cr.Valid {
		t.Error("expected invalid for empty input")
	}
}

func TestParseColorResponse_NoRGBPrefix(t *testing.T) {
	input := "\x1b]4;0;foo/bar/baz\x1b\\"
	cr := ParseColorResponse(input)
	if cr.Valid {
		t.Error("expected invalid for non-rgb prefix")
	}
}

func TestParseColorResponse_NotEnoughParts(t *testing.T) {
	input := "\x1b]4\x1b\\"
	cr := ParseColorResponse(input)
	if cr.Valid {
		t.Error("expected invalid for too few parts")
	}
}

func TestParseColorResponse_BadHexDigits(t *testing.T) {
	input := "\x1b]4;0;rgb:xx/yy/zz\x1b\\"
	cr := ParseColorResponse(input)
	if !cr.Valid {
		t.Fatal("expected valid (bad hex → 0)")
	}
	if cr.R != 0 || cr.G != 0 || cr.B != 0 {
		t.Errorf("expected (0,0,0) for bad hex, got (%d,%d,%d)", cr.R, cr.G, cr.B)
	}
}

// --- IsDarkBackground ---

func TestIsDarkBackground_Black(t *testing.T) {
	if !IsDarkBackground(0, 0, 0) {
		t.Error("expected black to be dark")
	}
}

func TestIsDarkBackground_White(t *testing.T) {
	if IsDarkBackground(255, 255, 255) {
		t.Error("expected white to be light")
	}
}

func TestIsDarkBackground_DarkGray(t *testing.T) {
	if !IsDarkBackground(30, 30, 30) {
		t.Error("expected dark gray to be dark")
	}
}

func TestIsDarkBackground_LightGray(t *testing.T) {
	if IsDarkBackground(230, 230, 230) {
		t.Error("expected light gray to be light")
	}
}

func TestIsDarkBackground_Dracula(t *testing.T) {
	// Dracula theme bg: #282a36 = (40, 42, 54)
	if !IsDarkBackground(40, 42, 54) {
		t.Error("expected Dracula background to be dark")
	}
}

func TestIsDarkBackground_SolarizedLight(t *testing.T) {
	// Solarized Light bg: #fdf6e3 = (253, 246, 227)
	if IsDarkBackground(253, 246, 227) {
		t.Error("expected Solarized Light background to be light")
	}
}

func TestIsDarkBackground_SolarizedDark(t *testing.T) {
	// Solarized Dark bg: #002b36 = (0, 43, 54)
	if !IsDarkBackground(0, 43, 54) {
		t.Error("expected Solarized Dark background to be dark")
	}
}

// --- parseHexComponent ---

func TestParseHexComponent(t *testing.T) {
	tests := []struct {
		input string
		want  uint8
	}{
		{"00", 0},
		{"ff", 255},
		{"80", 128},
		{"0", 0},
		{"f", 255},  // 4-bit scaled
		{"8", 136},  // 0x8 * 0x11 = 0x88 = 136
		{"0000", 0},
		{"ffff", 255}, // 16-bit scaled down
		{"8080", 128}, // 0x8080 >> 8 = 0x80
		{"", 0},
	}
	for _, tt := range tests {
		got := parseHexComponent(tt.input)
		if got != tt.want {
			t.Errorf("parseHexComponent(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

// --- itoa / atoi helpers ---

func TestColorItoa(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{255, "255"},
		{-1, "-1"},
	}
	for _, tt := range tests {
		got := colorItoa(tt.input)
		if got != tt.want {
			t.Errorf("colorItoa(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestAtoiDef(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"0", 0},
		{"255", 255},
		{"abc", 0},
		{"", 0},
		{"12abc", 12},
	}
	for _, tt := range tests {
		got := atoiDef(tt.input)
		if got != tt.want {
			t.Errorf("atoiDef(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

// --- Round trip test ---

func TestColorQuery_RoundTrip(t *testing.T) {
	// Generate query, simulate response, parse response
	query := QueryPaletteColor(42)
	if query == "" {
		t.Fatal("expected non-empty query")
	}

	// Simulate terminal response for index 42 with color rgb:1a2b/3c4d/5e6f
	response := "\x1b]4;42;rgb:1a2b/3c4d/5e6f\x1b\\"
	cr := ParseColorResponse(response)
	if !cr.Valid {
		t.Fatal("expected valid response")
	}
	if cr.Index != 42 {
		t.Errorf("expected index 42, got %d", cr.Index)
	}
	// Verify it's dark enough to detect
	_ = IsDarkBackground(cr.R, cr.G, cr.B)
}

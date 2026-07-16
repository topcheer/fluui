package lipgloss

import "testing"

// P241: parseColor hex-error, parseHexColor invalid, hexToVal upper, PlaceHorizontal center, Render all flags

func TestParseColor_InvalidHex_P241(t *testing.T) {
	c := parseColor(Color("#zzzzzz"))
	if c.Type != 0 {
		t.Error("invalid hex should return ColorNone")
	}
}

func TestParseColor_NamedNotFound_P241(t *testing.T) {
	c := parseColor(Color("notacolor"))
	if c.Type != 0 {
		t.Error("unknown name should return ColorNone")
	}
}

func TestParseHexColor_ShortString_P241(t *testing.T) {
	_, _, _, ok := parseHexColor("#abc")
	if ok {
		t.Error("short hex should fail")
	}
}

func TestParseHexColor_InvalidChars_P241(t *testing.T) {
	_, _, _, ok := parseHexColor("#gg0000")
	if ok {
		t.Error("invalid hex chars should fail")
	}
}

func TestHexToVal_Uppercase_P241(t *testing.T) {
	v := hexToVal("FF")
	if v != 255 {
		t.Errorf("hexToVal('FF') = %d, want 255", v)
	}
}

func TestPlaceHorizontal_Center_P241(t *testing.T) {
	// Center alignment
	result := PlaceHorizontal(10, Center, "hi")
	if len(result) < 4 {
		t.Error("center placement should produce wider output")
	}
}

func TestRender_StrikeAndBlink_P241(t *testing.T) {
	s := NewStyle().Strikethrough(true).Blink(true)
	result := s.Render("test")
	if result == "test" {
		t.Error("strike+blink should produce ANSI codes")
	}
}

func TestRender_DimAndReverse_P241(t *testing.T) {
	s := NewStyle().Dim(true).Reverse(true)
	result := s.Render("test")
	if result == "test" {
		t.Error("dim+reverse should produce ANSI codes")
	}
}

func TestRender_CacheHit_P241(t *testing.T) {
	s := NewStyle().Foreground(Color("12"))
	r1 := s.Render("test")
	r2 := s.Render("test")
	if r1 != r2 {
		t.Error("cached render should produce same result")
	}
}

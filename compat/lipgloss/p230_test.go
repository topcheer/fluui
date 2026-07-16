package lipgloss

import "testing"

// P230: Verify that lipgloss.Color("14") works as a native Go type conversion,
// exactly like charm.land's `type Color string`. This is THE critical compat test:
// ggcode uses lipgloss.Color("14") 349 times.

func TestColor_TypeConversion_P230(t *testing.T) {
	// In charm.land: type Color string → Color("14") is a type conversion
	// In fluui: now also type Color string → Color("14") must work identically
	c := Color("14")
	if c != "14" {
		t.Errorf("expected \"14\", got %q", c)
	}
	if string(c) != "14" {
		t.Errorf("expected string \"14\", got %q", string(c))
	}
}

func TestColor_TypeConversionHex_P230(t *testing.T) {
	c := Color("#ff8800")
	if c != "#ff8800" {
		t.Errorf("expected \"#ff8800\", got %q", c)
	}
}

func TestColor_TypeConversionNamed_P230(t *testing.T) {
	c := Color("red")
	if c != "red" {
		t.Errorf("expected \"red\", got %q", c)
	}
}

func TestColor_ParseAfterConversion_P230(t *testing.T) {
	// Verify that a type-converted Color parses correctly to buffer.Color
	c := Color("14")
	bc := parseColor(c)
	if bc.Type == 0 {
		t.Error("parseColor(Color(\"14\")) should not be ColorNone")
	}
}

func TestColor_ForegroundWithTypeConversion_P230(t *testing.T) {
	// This is the EXACT ggcode pattern:
	// lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	s := NewStyle().Foreground(Color("14"))
	result := s.Render("test")
	if result == "" {
		t.Error("Render with Color() type conversion should produce output")
	}
}

func TestColor_BackgroundWithTypeConversion_P230(t *testing.T) {
	// ggcode: lipgloss.NewStyle().Background(lipgloss.Color("#1a1a2e"))
	s := NewStyle().Background(Color("#1a1a2e"))
	result := s.Render("test")
	if result == "" {
		t.Error("Render with Color() type conversion should produce output")
	}
}

func TestColor_EmptyString_P230(t *testing.T) {
	c := Color("")
	bc := parseColor(c)
	if bc.Type != 0 {
		t.Error("empty color should parse to ColorNone")
	}
}

func TestColor_ZeroValue_P230(t *testing.T) {
	// Zero value of type Color string is ""
	var c Color
	if string(c) != "" {
		t.Error("zero value Color should be empty string")
	}
}

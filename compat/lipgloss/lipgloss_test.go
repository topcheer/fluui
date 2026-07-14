package lipgloss

import "testing"

func TestNewStyle(t *testing.T) {
	s := NewStyle()
	result := s.Render("hello")
	if result == "" {
		t.Error("empty style should produce output")
	}
}

func TestStyleBold(t *testing.T) {
	s := NewStyle().Bold(true)
	result := s.Render("hello")
	if result == "hello" {
		t.Error("bold should add SGR codes")
	}
}

func TestStyleForeground(t *testing.T) {
	s := NewStyle().Foreground(NewColor("12"))
	result := s.Render("hello")
	if result == "hello" {
		t.Error("foreground should add SGR codes")
	}
}

func TestStyleChain(t *testing.T) {
	s := NewStyle().Bold(true).Italic(true).Foreground(NewColor("9"))
	result := s.Render("test")
	if result == "" {
		t.Error("chained style should produce output")
	}
}

func TestColor256(t *testing.T) {
	c := Color256(12)
	if c.val.Type == 0 {
		t.Error("Color256 should not be ColorNone")
	}
}

func TestColorRGB(t *testing.T) {
	c := ColorRGB(255, 128, 0)
	if c.val.Type == 0 {
		t.Error("ColorRGB should not be ColorNone")
	}
}

func TestColorHex(t *testing.T) {
	c := ColorHex("#ff8000")
	if c.val.Type == 0 {
		t.Error("ColorHex should not be ColorNone")
	}
}

func TestNewColorNumeric(t *testing.T) {
	c := NewColor("12")
	if c.val.Type == 0 {
		t.Error("numeric color should work")
	}
}

func TestNewColorHex(t *testing.T) {
	c := NewColor("#ff0000")
	if c.val.Type == 0 {
		t.Error("hex color should work")
	}
}

func TestNewColorNamed(t *testing.T) {
	c := NewColor("red")
	if c.val.Type == 0 {
		t.Error("named color should work")
	}
}

func TestColorFunc(t *testing.T) {
	c := ColorFunc("12")
	if c.val.Type == 0 {
		t.Error("ColorFunc should work")
	}
}

func TestRoundedBorder(t *testing.T) {
	b := RoundedBorder()
	if b.TopLeft != "╭" {
		t.Error("rounded border top-left should be ╭")
	}
}

func TestNormalBorder(t *testing.T) {
	b := NormalBorder()
	if b.TopLeft != "+" {
		t.Error("normal border top-left should be +")
	}
}

func TestWidth(t *testing.T) {
	if Width("hello") != 5 {
		t.Error("Width should return 5 for 'hello'")
	}
}

func TestHeight(t *testing.T) {
	if Height("a\nb\nc") != 3 {
		t.Error("Height should return 3")
	}
}

func TestJoinHorizontal(t *testing.T) {
	result := JoinHorizontal(Top, "a", "b")
	if result != "ab" {
		t.Errorf("expected 'ab', got '%s'", result)
	}
}

func TestJoinVertical(t *testing.T) {
	result := JoinVertical(Left, "a", "b")
	if result != "a\nb" {
		t.Errorf("expected 'a\\nb', got '%s'", result)
	}
}

func TestPlaceHorizontal(t *testing.T) {
	result := PlaceHorizontal(10, Right, "hi")
	if len(result) != 10 {
		t.Error("should pad to width 10")
	}
}

func TestPlaceVertical(t *testing.T) {
	result := PlaceVertical(3, Top, "hi")
	if Height(result) != 3 {
		t.Error("should pad to height 3")
	}
}

func TestStylePadding(t *testing.T) {
	s := NewStyle().Padding(1)
	result := s.Render("x")
	if len(result) < 3 {
		t.Error("padding should add spaces")
	}
}

func TestStyleMarginBottom(t *testing.T) {
	s := NewStyle().MarginBottom(2)
	result := s.Render("x")
	if Height(result) != 3 {
		t.Error("margin bottom should add newlines")
	}
}

func TestStyleString(t *testing.T) {
	s := NewStyle().Bold(true)
	_ = s.String() // should not panic
}

func TestStyleUnsetBold(t *testing.T) {
	s := NewStyle().Bold(true).UnsetBold()
	if s.GetBold() {
		t.Error("UnsetBold should remove bold")
	}
}

func TestStyleGetForeground(t *testing.T) {
	s := NewStyle().Foreground(NewColor("12"))
	c := s.GetForeground()
	if c.val.Type == 0 {
		t.Error("GetForeground should return the color")
	}
}

func TestAdaptiveColorResolve(t *testing.T) {
	ac := AdaptiveColor{Light: "white", Dark: "12"}
	c := ac.Resolve()
	if c.val.Type == 0 {
		t.Error("Resolve should return a valid color")
	}
}

func TestAdaptiveColorString(t *testing.T) {
	ac := AdaptiveColor{Light: "white", Dark: "red"}
	_ = ac.String() // should not panic
}

func TestStyleBorder(t *testing.T) {
	s := NewStyle().Border(RoundedBorder()).BorderForeground(NewColor("12"))
	_ = s.Render("test") // should not crash
}

func TestStyleWidth(t *testing.T) {
	s := NewStyle().Width(20)
	_ = s.Render("test") // should not crash
}

func TestStyleHeight(t *testing.T) {
	s := NewStyle().Height(5)
	_ = s.Render("test") // should not crash
}
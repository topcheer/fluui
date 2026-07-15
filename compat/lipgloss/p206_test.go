package lipgloss

import "testing"

func TestThickBorder_P206(t *testing.T) {
	b := ThickBorder()
	_ = b
}

func TestDoubleBorder_P206(t *testing.T) {
	b := DoubleBorder()
	_ = b
}

func TestPlace_P206(t *testing.T) {
	s := Place(20, 5, Center, Middle, "hi")
	if s == "" {
		t.Error("Place should return non-empty")
	}
}

func TestPlaceHorizontalWide_P206(t *testing.T) {
	s := PlaceHorizontal(30, Right, "right-aligned")
	if s == "" {
		t.Error("should not be empty")
	}
}

func TestPlaceVerticalTall_P206(t *testing.T) {
	s := PlaceVertical(5, Bottom, "bottom")
	if s == "" {
		t.Error("should not be empty")
	}
}

func TestJoinHorizontalMulti_P206(t *testing.T) {
	s := JoinHorizontal(Top, "a", "b", "c")
	if s == "" {
		t.Error("should not be empty")
	}
}

func TestJoinVerticalMulti_P206(t *testing.T) {
	s := JoinVertical(Left, "x", "y", "z")
	if s == "" {
		t.Error("should not be empty")
	}
}

func TestMarginFourArgs_P206(t *testing.T) {
	s := NewStyle().Margin(1, 2, 3, 4)
	r := s.Render("test")
	_ = r
}

func TestMarginTwoArgs_P206(t *testing.T) {
	s := NewStyle().Margin(2, 3)
	_ = s.Render("test")
}

func TestNamedColorIndex_P206(t *testing.T) {
	for _, name := range []string{"red", "green", "blue", "white", "black", "yellow", "cyan", "magenta", "brightred", "brightgreen"} {
		_ = namedColorIndex(name)
	}
	// Unknown name returns some default (implementation-dependent)
	_ = namedColorIndex("unknown")
}

func TestColorToBuffer_P206(t *testing.T) {
	_ = colorToBuffer(ColorNamed("red"))
}

func TestGetForegroundUnset_P206(t *testing.T) {
	s := NewStyle()
	c := s.GetForeground()
	_ = c
}

func TestHeightMultiLine_P206(t *testing.T) {
	h := Height("line1\nline2\nline3")
	if h != 3 {
		t.Errorf("expected 3, got %d", h)
	}
}

func TestVisibleWidth_P206(t *testing.T) {
	w := visibleWidth("hello")
	if w != 5 {
		t.Errorf("expected 5, got %d", w)
	}
}
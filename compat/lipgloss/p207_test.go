package lipgloss

import "testing"

// P207: targeted tests for remaining sub-80% functions

func TestJoinHorizontalMultiLine_P207(t *testing.T) {
	// Multi-line strings to exercise max-height + padding paths
	r := JoinHorizontal(Top, "ab\ncd", "xyz")
	if r == "" {
		t.Error("should not be empty")
	}
	r2 := JoinHorizontal(Bottom, "x\ny\nz", "a")
	_ = r2
	r3 := JoinHorizontal(Middle, "1\n2", "a\nb\nc")
	_ = r3
}

func TestJoinHorizontalEmpty_P207(t *testing.T) {
	if JoinHorizontal(Top) != "" {
		t.Error("empty args should return empty")
	}
	if JoinHorizontal(Top, "only") != "only" {
		t.Error("single arg should return as-is")
	}
}

func TestJoinVerticalPadding_P207(t *testing.T) {
	r := JoinVertical(Left, "ab", "xy")
	if r == "" {
		t.Error("should not be empty")
	}
	// Different-width strings exercise padding
	r2 := JoinVertical(Right, "a", "abc")
	_ = r2
	r3 := JoinVertical(Center, "longer", "x")
	_ = r3
}

func TestPlaceHorizontalAllAligns_P207(t *testing.T) {
	for _, align := range []HorizontalAlign{Left, Center, Right} {
		r := PlaceHorizontal(10, align, "test")
		if r == "" {
			t.Error("should not be empty")
		}
	}
}

func TestPlaceVerticalAllAligns_P207(t *testing.T) {
	for _, align := range []VerticalAlign{Top, Middle, Bottom} {
		r := PlaceVertical(5, align, "test")
		if r == "" {
			t.Error("should not be empty")
		}
	}
}

func TestPlaceVerticalTallerThanHeight_P207(t *testing.T) {
	r := PlaceVertical(2, Top, "line1\nline2\nline3\nline4")
	if r != "line1\nline2\nline3\nline4" {
		t.Error("should return content as-is when taller than height")
	}
}

func TestHeightEmpty_P207(t *testing.T) {
	if Height("") != 0 {
		t.Error("empty string height should be 0")
	}
}

func TestVisibleWidthANSI_P207(t *testing.T) {
	// ANSI escape sequences should be stripped
	w := visibleWidth("\x1b[31mhello\x1b[0m")
	if w != 5 {
		t.Errorf("expected 5, got %d", w)
	}
}

func TestMarginAllArgs_P207(t *testing.T) {
	// 4-arg margin: top, right, bottom, left
	_ = NewStyle().Margin(1, 2, 3, 4).Render("x")
	// 3-arg margin: top, rightleft, bottom
	_ = NewStyle().Margin(1, 2, 3).Render("x")
}

func TestPaddingAllArgs_P207(t *testing.T) {
	_ = NewStyle().Padding(1, 2, 3, 4).Render("x")
	_ = NewStyle().Padding(1, 2, 3).Render("x")
}
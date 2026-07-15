package lipgloss

import (
	"testing"
)

// P199: Comprehensive coverage for lipgloss compat

func TestMakeColor_P199(t *testing.T) {
	c := MakeColor("#ff0000")
	_ = c
	c2 := MakeColor("invalid")
	_ = c2
}

func TestNamedColorIndex_P199(t *testing.T) {
	tests := []string{"red", "green", "blue", "white", "black", "yellow", "cyan", "magenta", "unknown"}
	for _, name := range tests {
		idx := namedColorIndex(name)
		_ = idx
	}
}

func TestColorHex_P199(t *testing.T) {
	c := ColorHex("#ff8800")
	_ = c
	c2 := ColorHex("ff8800")
	_ = c2
}

func TestColorNamed_P199(t *testing.T) {
	c := ColorNamed("red")
	_ = c
}

func TestStyleChaining_P199(t *testing.T) {
	s := NewStyle()
	s = s.Background(ColorNamed("blue"))
	s = s.Underline(true)
	s = s.Strikethrough(true)
	s = s.Reverse(true)
	s = s.Dim(true)
	s = s.Blink(true)
	s = s.MaxWidth(80)
	s = s.MaxHeight(24)
	_ = s
}

func TestStyleMargins_P199(t *testing.T) {
	s := NewStyle()
	s = s.MarginTop(1)
	s = s.MarginLeft(2)
	s = s.MarginRight(3)
	s = s.Margin(1, 2, 3, 4)
	s = s.Margin(2)
	_ = s
}

func TestStylePadding_P199(t *testing.T) {
	s := NewStyle()
	s = s.PaddingTop(1)
	s = s.PaddingBottom(2)
	s = s.PaddingLeft(3)
	s = s.PaddingRight(4)
	s = s.Padding(1, 2, 3, 4)
	s = s.Padding(2)
	_ = s
}

func TestStyleAlign_P199(t *testing.T) {
	s := NewStyle()
	s = s.AlignHorizontal(Left)
	s = s.AlignHorizontal(Center)
	s = s.AlignHorizontal(Right)
	s = s.AlignVertical(Top)
	s = s.AlignVertical(Middle)
	s = s.AlignVertical(Bottom)
	_ = s
}

func TestStyleBorder_P199(t *testing.T) {
	s := NewStyle()
	s = s.Border(RoundedBorder(), true)
	s = s.BorderBackground(ColorNamed("blue"))
	result := s.Render("test")
	if result == "" {
		t.Error("Render should not be empty")
	}
}

func TestStyleUnsetters_P199(t *testing.T) {
	s := NewStyle()
	s.UnsetBold()
	s.UnsetItalic()
	s.UnsetUnderline()
	s.UnsetForeground()
	s.UnsetBackground()
	_ = s
}

func TestStyleGetters_P199(t *testing.T) {
	s := NewStyle().Foreground(ColorNamed("red")).Background(ColorNamed("blue"))
	_ = s.GetForeground()
}

func TestColorToBuffer_P199(t *testing.T) {
	c := ColorNamed("red")
	b := colorToBuffer(c)
	_ = b
}

func TestStyleRenderComplex_P199(t *testing.T) {
	s := NewStyle().
		Foreground(ColorNamed("red")).
		Bold(true).
		Padding(1, 2).
		Width(20).
		AlignHorizontal(Center)
	result := s.Render("hello")
	if result == "" {
		t.Error("Render should not be empty")
	}
}
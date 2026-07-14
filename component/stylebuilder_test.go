package component

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP156_StyleBuilder_NewStyle(t *testing.T) {
	s := NewStyle()
	if s == nil {
		t.Fatal("expected non-nil StyleBuilder")
	}
}

func TestP156_StyleBuilder_Bold(t *testing.T) {
	s := NewStyle().Bold()
	if s.Style().Flags&buffer.Bold == 0 {
		t.Error("expected Bold flag set")
	}
	s.Bold(false)
	if s.Style().Flags&buffer.Bold != 0 {
		t.Error("expected Bold flag cleared")
	}
}

func TestP156_StyleBuilder_Italic(t *testing.T) {
	s := NewStyle().Italic()
	if s.Style().Flags&buffer.Italic == 0 {
		t.Error("expected Italic flag set")
	}
}

func TestP156_StyleBuilder_Underline(t *testing.T) {
	s := NewStyle().Underline()
	if s.Style().Flags&buffer.Underline == 0 {
		t.Error("expected Underline flag set")
	}
}

func TestP156_StyleBuilder_Dim(t *testing.T) {
	s := NewStyle().Dim()
	if s.Style().Flags&buffer.Dim == 0 {
		t.Error("expected Dim flag set")
	}
}

func TestP156_StyleBuilder_Reverse(t *testing.T) {
	s := NewStyle().Reverse()
	if s.Style().Flags&buffer.Reverse == 0 {
		t.Error("expected Reverse flag set")
	}
}

func TestP156_StyleBuilder_Strikethrough(t *testing.T) {
	s := NewStyle().Strikethrough()
	if s.Style().Flags&buffer.Strikethrough == 0 {
		t.Error("expected Strikethrough flag set")
	}
}

func TestP156_StyleBuilder_Foreground(t *testing.T) {
	c := buffer.NamedColor(buffer.NamedRed)
	s := NewStyle().Foreground(c)
	if s.Style().Fg.Type != c.Type {
		t.Errorf("expected Fg type %d, got %d", c.Type, s.Style().Fg.Type)
	}
}

func TestP156_StyleBuilder_Background(t *testing.T) {
	c := buffer.NamedColor(buffer.NamedBlue)
	s := NewStyle().Background(c)
	if s.Style().Bg.Type != c.Type {
		t.Errorf("expected Bg type %d, got %d", c.Type, s.Style().Bg.Type)
	}
}

func TestP156_StyleBuilder_ForegroundRGB(t *testing.T) {
	s := NewStyle().ForegroundRGB(255, 128, 0)
	st := s.Style()
	if st.Fg.Type != buffer.ColorTrue {
		t.Errorf("expected RGB type, got %d", st.Fg.Type)
	}
}

func TestP156_StyleBuilder_BackgroundRGB(t *testing.T) {
	s := NewStyle().BackgroundRGB(0, 255, 128)
	st := s.Style()
	if st.Bg.Type != buffer.ColorTrue {
		t.Errorf("expected RGB type, got %d", st.Bg.Type)
	}
}

func TestP156_StyleBuilder_ForegroundNamed(t *testing.T) {
	s := NewStyle().ForegroundNamed(buffer.NamedGreen)
	st := s.Style()
	if st.Fg.Type != buffer.ColorNamed {
		t.Errorf("expected Named type, got %d", st.Fg.Type)
	}
}

func TestP156_StyleBuilder_ForegroundANSI(t *testing.T) {
	s := NewStyle().ForegroundANSI(81)
	st := s.Style()
	if st.Fg.Type != buffer.Color256 {
		t.Errorf("expected 256 type, got %d", st.Fg.Type)
	}
}

func TestP156_StyleBuilder_ForegroundHex(t *testing.T) {
	s := NewStyle().ForegroundHex("#ff8800")
	st := s.Style()
	if st.Fg.Type != buffer.ColorTrue {
		t.Errorf("expected RGB type, got %d", st.Fg.Type)
	}
}

func TestP156_StyleBuilder_ForegroundHexInvalid(t *testing.T) {
	s := NewStyle().ForegroundHex("invalid")
	st := s.Style()
	if st.Fg.Type != 0 {
		t.Errorf("expected zero type, got %d", st.Fg.Type)
	}
}

func TestP156_StyleBuilder_ForegroundColor(t *testing.T) {
	s := NewStyle().ForegroundColor("red")
	st := s.Style()
	if st.Fg.Type != buffer.ColorNamed {
		t.Errorf("expected Named type for 'red', got %d", st.Fg.Type)
	}
}

func TestP156_StyleBuilder_ForegroundColorNumeric(t *testing.T) {
	s := NewStyle().ForegroundColor("81")
	st := s.Style()
	if st.Fg.Type != buffer.Color256 {
		t.Errorf("expected 256 type for '81', got %d", st.Fg.Type)
	}
}

func TestP156_StyleBuilder_ForegroundColorHex(t *testing.T) {
	s := NewStyle().ForegroundColor("#ff8800")
	st := s.Style()
	if st.Fg.Type != buffer.ColorTrue {
		t.Errorf("expected RGB type for hex, got %d", st.Fg.Type)
	}
}

func TestP156_StyleBuilder_ForegroundColorEmpty(t *testing.T) {
	s := NewStyle().ForegroundColor("")
	st := s.Style()
	if st.Fg.Type != 0 {
		t.Errorf("expected zero type, got %d", st.Fg.Type)
	}
}

func TestP156_StyleBuilder_ChainAll(t *testing.T) {
	s := NewStyle().
		Bold().
		Italic().
		Underline().
		Dim().
		Reverse().
		Strikethrough().
		Foreground(buffer.NamedColor(buffer.NamedCyan)).
		Background(buffer.NamedColor(buffer.NamedBlack))
	st := s.Style()
	expectedFlags := buffer.Bold | buffer.Italic | buffer.Underline | buffer.Dim | buffer.Reverse | buffer.Strikethrough
	if st.Flags != expectedFlags {
		t.Errorf("expected flags %d, got %d", expectedFlags, st.Flags)
	}
	if st.Fg.Type != buffer.ColorNamed {
		t.Errorf("expected Named fg, got %d", st.Fg.Type)
	}
	if st.Bg.Type != buffer.ColorNamed {
		t.Errorf("expected Named bg, got %d", st.Bg.Type)
	}
}

func TestP156_StyleBuilder_Render(t *testing.T) {
	s := NewStyle().Bold().Foreground(buffer.NamedColor(buffer.NamedRed))
	result := s.Render("Hello")
	if !strings.Contains(result, "Hello") {
		t.Error("expected result to contain 'Hello'")
	}
	// Should produce some escape sequence
	if len(result) <= len("Hello") {
		t.Error("expected styled output longer than plain text")
	}
}

func TestP156_StyleBuilder_Style(t *testing.T) {
	s := NewStyle().Bold().ForegroundRGB(255, 0, 0)
	st := s.Style()
	if st.Flags&buffer.Bold == 0 {
		t.Error("expected bold")
	}
	if st.Fg.Type != buffer.ColorTrue {
		t.Error("expected RGB fg")
	}
}

func TestP156_StyleBuilder_Copy(t *testing.T) {
	s1 := NewStyle().Bold().Foreground(buffer.NamedColor(buffer.NamedRed))
	s2 := s1.Copy()
	s2.Italic(true)
	if s1.Style().Flags&buffer.Italic != 0 {
		t.Error("original should not have italic")
	}
	if s2.Style().Flags&buffer.Italic == 0 {
		t.Error("copy should have italic")
	}
}

func TestP156_StyleBuilder_UnsetBold(t *testing.T) {
	s := NewStyle().Bold().UnsetBold()
	if s.Style().Flags&buffer.Bold != 0 {
		t.Error("expected bold unset")
	}
}

func TestP156_StyleBuilder_UnsetForeground(t *testing.T) {
	s := NewStyle().Foreground(buffer.NamedColor(buffer.NamedRed)).UnsetForeground()
	if s.Style().Fg.Type != 0 {
		t.Error("expected fg unset")
	}
}

func TestP156_StyleBuilder_Inherit(t *testing.T) {
	parent := NewStyle().Bold().Foreground(buffer.NamedColor(buffer.NamedRed))
	child := NewStyle().Italic().Inherit(parent)
	st := child.Style()
	if st.Flags&buffer.Bold == 0 {
		t.Error("expected inherited bold")
	}
	if st.Flags&buffer.Italic == 0 {
		t.Error("expected child italic")
	}
}

func TestP156_StyleBuilder_Width(t *testing.T) {
	s := NewStyle()
	if s.MeasureWidth("Hello") != 5 {
		t.Errorf("expected 5, got %d", s.MeasureWidth("Hello"))
	}
}

func TestP156_StyleBuilder_RenderPlain(t *testing.T) {
	s := NewStyle().Bold()
	if s.RenderPlain("text") != "text" {
		t.Error("expected plain text")
	}
}

func TestP156_StyleBuilder_BackgroundColor(t *testing.T) {
	s := NewStyle().BackgroundColor("blue")
	st := s.Style()
	if st.Bg.Type != buffer.ColorNamed {
		t.Errorf("expected Named bg, got %d", st.Bg.Type)
	}
}

func TestP156_StyleBuilder_BackgroundHex(t *testing.T) {
	s := NewStyle().BackgroundHex("#00ff00")
	st := s.Style()
	if st.Bg.Type != buffer.ColorTrue {
		t.Errorf("expected RGB bg, got %d", st.Bg.Type)
	}
}

func TestP156_StyleBuilder_UnsetItalic(t *testing.T) {
	s := NewStyle().Italic().UnsetItalic()
	if s.Style().Flags&buffer.Italic != 0 {
		t.Error("expected italic unset")
	}
}

func TestP156_StyleBuilder_UnsetUnderline(t *testing.T) {
	s := NewStyle().Underline().UnsetUnderline()
	if s.Style().Flags&buffer.Underline != 0 {
		t.Error("expected underline unset")
	}
}

func TestP156_StyleBuilder_UnsetBackground(t *testing.T) {
	s := NewStyle().Background(buffer.NamedColor(buffer.NamedBlue)).UnsetBackground()
	if s.Style().Bg.Type != 0 {
		t.Error("expected bg unset")
	}
}

func TestP156_StyleBuilder_BackgroundNamed(t *testing.T) {
	s := NewStyle().BackgroundNamed(buffer.NamedBlue)
	st := s.Style()
	if st.Bg.Type != buffer.ColorNamed {
		t.Errorf("expected Named bg, got %d", st.Bg.Type)
	}
}

func TestP156_StyleBuilder_BackgroundANSI(t *testing.T) {
	s := NewStyle().BackgroundANSI(42)
	st := s.Style()
	if st.Bg.Type != buffer.Color256 {
		t.Errorf("expected 256 bg, got %d", st.Bg.Type)
	}
}

func TestP156_StyleBuilder_ForegroundColorUnknown(t *testing.T) {
	s := NewStyle().ForegroundColor("nonexistent")
	st := s.Style()
	if st.Fg.Type != 0 {
		t.Errorf("expected zero type for unknown color, got %d", st.Fg.Type)
	}
}

func TestP156_StyleBuilder_Blink(t *testing.T) {
	s := NewStyle().Blink()
	if s.Style().Flags&buffer.Blink == 0 {
		t.Error("expected Blink flag set")
	}
}
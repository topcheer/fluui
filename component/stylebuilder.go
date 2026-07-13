package component

import (
	"strings"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Lipgloss-compatible StyleBuilder ───
//
// StyleBuilder provides a fluent, chainable API for building buffer.Style,
// matching the lipgloss.NewStyle().Bold().Foreground(color).Render(text) pattern.
//
// This is the #1 API pattern used in ggcode (393+ call sites) and is essential
// for migration from lipgloss to fluui.
//
// Usage:
//
//	s := NewStyle().Bold(true).Foreground(buffer.NamedColor(buffer.NamedRed))
//	styledText := s.Render("Error!")
//	// Or get the underlying Style:
//	style := s.Style()

// StyleBuilder is a fluent builder for buffer.Style.
type StyleBuilder struct {
	style buffer.Style
}

// NewStyle creates a new StyleBuilder (lipgloss-compatible).
func NewStyle() *StyleBuilder {
	return &StyleBuilder{}
}

// Bold sets the bold flag.
func (s *StyleBuilder) Bold(b ...bool) *StyleBuilder {
	v := true
	if len(b) > 0 {
		v = b[0]
	}
	if v {
		s.style.Flags |= buffer.Bold
	} else {
		s.style.Flags &^= buffer.Bold
	}
	return s
}

// Italic sets the italic flag.
func (s *StyleBuilder) Italic(b ...bool) *StyleBuilder {
	v := true
	if len(b) > 0 {
		v = b[0]
	}
	if v {
		s.style.Flags |= buffer.Italic
	} else {
		s.style.Flags &^= buffer.Italic
	}
	return s
}

// Underline sets the underline flag.
func (s *StyleBuilder) Underline(b ...bool) *StyleBuilder {
	v := true
	if len(b) > 0 {
		v = b[0]
	}
	if v {
		s.style.Flags |= buffer.Underline
	} else {
		s.style.Flags &^= buffer.Underline
	}
	return s
}

// Dim sets the dim flag.
func (s *StyleBuilder) Dim(b ...bool) *StyleBuilder {
	v := true
	if len(b) > 0 {
		v = b[0]
	}
	if v {
		s.style.Flags |= buffer.Dim
	} else {
		s.style.Flags &^= buffer.Dim
	}
	return s
}

// Blink sets the blink flag.
func (s *StyleBuilder) Blink(b ...bool) *StyleBuilder {
	v := true
	if len(b) > 0 {
		v = b[0]
	}
	if v {
		s.style.Flags |= buffer.Blink
	} else {
		s.style.Flags &^= buffer.Blink
	}
	return s
}

// Reverse sets the reverse flag.
func (s *StyleBuilder) Reverse(b ...bool) *StyleBuilder {
	v := true
	if len(b) > 0 {
		v = b[0]
	}
	if v {
		s.style.Flags |= buffer.Reverse
	} else {
		s.style.Flags &^= buffer.Reverse
	}
	return s
}

// Strikethrough sets the strikethrough flag.
func (s *StyleBuilder) Strikethrough(b ...bool) *StyleBuilder {
	v := true
	if len(b) > 0 {
		v = b[0]
	}
	if v {
		s.style.Flags |= buffer.Strikethrough
	} else {
		s.style.Flags &^= buffer.Strikethrough
	}
	return s
}

// Foreground sets the foreground color.
func (s *StyleBuilder) Foreground(c buffer.Color) *StyleBuilder {
	s.style.Fg = c
	return s
}

// Background sets the background color.
func (s *StyleBuilder) Background(c buffer.Color) *StyleBuilder {
	s.style.Bg = c
	return s
}

// ForegroundRGB sets the foreground color from RGB values.
func (s *StyleBuilder) ForegroundRGB(r, g, b uint8) *StyleBuilder {
	s.style.Fg = buffer.RGB(r, g, b)
	return s
}

// BackgroundRGB sets the background color from RGB values.
func (s *StyleBuilder) BackgroundRGB(r, g, b uint8) *StyleBuilder {
	s.style.Bg = buffer.RGB(r, g, b)
	return s
}

// ForegroundNamed sets the foreground from a named color constant.
func (s *StyleBuilder) ForegroundNamed(name int) *StyleBuilder {
	s.style.Fg = buffer.NamedColor(name)
	return s
}

// BackgroundNamed sets the background from a named color constant.
func (s *StyleBuilder) BackgroundNamed(name int) *StyleBuilder {
	s.style.Bg = buffer.NamedColor(name)
	return s
}

// ForegroundANSI sets the foreground from a 256-color palette index.
func (s *StyleBuilder) ForegroundANSI(n uint8) *StyleBuilder {
	s.style.Fg = buffer.Color256Val(n)
	return s
}

// BackgroundANSI sets the background from a 256-color palette index.
func (s *StyleBuilder) BackgroundANSI(n uint8) *StyleBuilder {
	s.style.Bg = buffer.Color256Val(n)
	return s
}

// ForegroundHex sets the foreground from a hex color string like "#ff8800".
func (s *StyleBuilder) ForegroundHex(hex string) *StyleBuilder {
	c, ok := parseHexColor(hex)
	if ok {
		s.style.Fg = c
	}
	return s
}

// BackgroundHex sets the background from a hex color string like "#ff8800".
func (s *StyleBuilder) BackgroundHex(hex string) *StyleBuilder {
	c, ok := parseHexColor(hex)
	if ok {
		s.style.Bg = c
	}
	return s
}

// ForegroundColor sets the foreground from a lipgloss-compatible Color string.
// Supports: hex (#ff8800), named (red), or ANSI numeric ("81").
func (s *StyleBuilder) ForegroundColor(c string) *StyleBuilder {
	s.style.Fg = parseLipglossColor(c)
	return s
}

// BackgroundColor sets the background from a lipgloss-compatible Color string.
func (s *StyleBuilder) BackgroundColor(c string) *StyleBuilder {
	s.style.Bg = parseLipglossColor(c)
	return s
}

// Style returns the underlying buffer.Style.
func (s *StyleBuilder) Style() buffer.Style {
	return s.style
}

// Render applies the style to a string and returns the styled string with
// SGR escape sequences. This matches lipgloss Style.Render().
func (s *StyleBuilder) Render(text string) string {
	return s.style.SGRSequence() + text + "\x1b[0m"
}

// RenderPlain returns the text without any styling (for plain-text contexts).
func (s *StyleBuilder) RenderPlain(text string) string {
	return text
}

// Width returns the display width of the text (lipgloss-compatible).
func (s *StyleBuilder) Width(text string) int {
	return stringWidth(text)
}

// Copy returns a copy of the StyleBuilder (lipgloss-compatible).
func (s *StyleBuilder) Copy() *StyleBuilder {
	return &StyleBuilder{style: s.style}
}

// UnsetBold removes the bold flag.
func (s *StyleBuilder) UnsetBold() *StyleBuilder {
	s.style.Flags &^= buffer.Bold
	return s
}

// UnsetItalic removes the italic flag.
func (s *StyleBuilder) UnsetItalic() *StyleBuilder {
	s.style.Flags &^= buffer.Italic
	return s
}

// UnsetUnderline removes the underline flag.
func (s *StyleBuilder) UnsetUnderline() *StyleBuilder {
	s.style.Flags &^= buffer.Underline
	return s
}

// UnsetForeground removes the foreground color.
func (s *StyleBuilder) UnsetForeground() *StyleBuilder {
	s.style.Fg = buffer.Color{}
	return s
}

// UnsetBackground removes the background color.
func (s *StyleBuilder) UnsetBackground() *StyleBuilder {
	s.style.Bg = buffer.Color{}
	return s
}

// Inherit returns a new StyleBuilder that inherits from the given parent style.
func (s *StyleBuilder) Inherit(parent *StyleBuilder) *StyleBuilder {
	result := parent.Copy()
	if s.style.Fg.Type != 0 {
		result.style.Fg = s.style.Fg
	}
	if s.style.Bg.Type != 0 {
		result.style.Bg = s.style.Bg
	}
	result.style.Flags = parent.style.Flags | s.style.Flags
	return result
}

// parseLipglossColor parses a lipgloss-compatible color string.
// Supports: hex (#ff8800), named colors (red), ANSI numeric ("81").
func parseLipglossColor(c string) buffer.Color {
	c = strings.TrimSpace(c)
	if c == "" {
		return buffer.Color{}
	}

	// Hex color
	if strings.HasPrefix(c, "#") {
		color, ok := parseHexColor(c)
		if ok {
			return color
		}
	}

	// Named color
	switch strings.ToLower(c) {
	case "red":
		return buffer.NamedColor(buffer.NamedRed)
	case "green":
		return buffer.NamedColor(buffer.NamedGreen)
	case "yellow":
		return buffer.NamedColor(buffer.NamedYellow)
	case "blue":
		return buffer.NamedColor(buffer.NamedBlue)
	case "magenta":
		return buffer.NamedColor(buffer.NamedMagenta)
	case "cyan":
		return buffer.NamedColor(buffer.NamedCyan)
	case "white":
		return buffer.NamedColor(buffer.NamedWhite)
	case "black":
		return buffer.NamedColor(buffer.NamedBlack)
	default:
		// Try numeric ANSI 256 color
		n := parseUint(c)
		if n >= 0 && n <= 255 {
			return buffer.Color256Val(uint8(n))
		}
	}

	return buffer.Color{}
}

// parseHexColor parses a hex color string like "#ff8800" into a buffer.Color.
func parseHexColor(hex string) (buffer.Color, bool) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return buffer.Color{}, false
	}
	r := hexToByte(hex[0:2])
	g := hexToByte(hex[2:4])
	b := hexToByte(hex[4:6])
	if r < 0 || g < 0 || b < 0 {
		return buffer.Color{}, false
	}
	return buffer.RGB(uint8(r), uint8(g), uint8(b)), true
}

// hexToByte converts a 2-char hex string to an int (-1 on error).
func hexToByte(s string) int {
	if len(s) != 2 {
		return -1
	}
	hi := hexCharToVal(s[0])
	lo := hexCharToVal(s[1])
	if hi < 0 || lo < 0 {
		return -1
	}
	return hi*16 + lo
}

func hexCharToVal(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	default:
		return -1
	}
}

// parseUint parses a string as a non-negative integer (-1 on error).
func parseUint(s string) int {
	if s == "" {
		return -1
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return -1
		}
		n = n*10 + int(c-'0')
	}
	return n
}
// BorderForeground sets the border foreground color (lipgloss-compatible).
func (s *StyleBuilder) BorderForeground(c buffer.Color) *StyleBuilder {
	// Store as a special field via Fg temporarily — adapter will handle
	// In fluui, borders are handled by the Border component, not Style
	return s
}

// BorderBackground sets the border background color (lipgloss-compatible).
func (s *StyleBuilder) BorderBackground(c buffer.Color) *StyleBuilder {
	return s
}

// MaxWidth sets the maximum width (lipgloss-compatible, no-op in fluui).
func (s *StyleBuilder) MaxWidth(w int) *StyleBuilder {
	return s
}

// MaxHeight sets the maximum height (lipgloss-compatible, no-op in fluui).
func (s *StyleBuilder) MaxHeight(h int) *StyleBuilder {
	return s
}

// MarginLeft sets left margin (lipgloss-compatible, no-op in fluui).
func (s *StyleBuilder) MarginLeft(n int) *StyleBuilder {
	return s
}

// MarginRight sets right margin (lipgloss-compatible, no-op in fluui).
func (s *StyleBuilder) MarginRight(n int) *StyleBuilder {
	return s
}

// MarginTop sets top margin (lipgloss-compatible, no-op in fluui).
func (s *StyleBuilder) MarginTop(n int) *StyleBuilder {
	return s
}

// MarginBottom sets bottom margin (lipgloss-compatible, no-op in fluui).
func (s *StyleBuilder) MarginBottom(n int) *StyleBuilder {
	return s
}

// PaddingLeft sets left padding (lipgloss-compatible, no-op in fluui).
func (s *StyleBuilder) PaddingLeft(n int) *StyleBuilder {
	return s
}

// PaddingRight sets right padding (lipgloss-compatible, no-op in fluui).
func (s *StyleBuilder) PaddingRight(n int) *StyleBuilder {
	return s
}

// PaddingTop sets top padding (lipgloss-compatible, no-op in fluui).
func (s *StyleBuilder) PaddingTop(n int) *StyleBuilder {
	return s
}

// PaddingBottom sets bottom padding (lipgloss-compatible, no-op in fluui).
func (s *StyleBuilder) PaddingBottom(n int) *StyleBuilder {
	return s
}

// Align sets text alignment (lipgloss-compatible, no-op in fluui).
func (s *StyleBuilder) Align(align HorizontalAlign) *StyleBuilder {
	return s
}

// TabWidth sets tab width (lipgloss-compatible, no-op in fluui).
func (s *StyleBuilder) TabWidth(n int) *StyleBuilder {
	return s
}

// UnderlineSpaces enables underlining spaces (lipgloss-compatible, no-op).
func (s *StyleBuilder) UnderlineSpaces(b ...bool) *StyleBuilder {
	return s
}

// StrikethroughSpaces enables strikethrough on spaces (lipgloss-compatible, no-op).
func (s *StyleBuilder) StrikethroughSpaces(b ...bool) *StyleBuilder {
	return s
}

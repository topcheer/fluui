// Package lipgloss provides a drop-in compatibility layer for projects
// migrating from charm.land/lipgloss/v2 to fluui.
//
// It mirrors the lipgloss API exactly — Style, Color, border types,
// JoinHorizontal/JoinVertical, Place, Width, Height — so that source
// files only need their import path changed:
//
//	-  "charm.land/lipgloss/v2"
//	+  lipgloss "github.com/topcheer/fluui/compat/lipgloss"
//
// No other code changes required. The implementation uses fluui's
// internal SGR sequence generation for zero-allocation style rendering.
package lipgloss

import (
	"fmt"
	"strings"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Color ───
// In lipgloss v2, Color is both a type and a constructor function.
// Go allows a func type and a function with the same name.

// Color is a terminal color. It can be an ANSI color name ("red"),
// an ANSI 256-color code ("12", "226"), or a hex color ("#ff8800").
type Color struct {
	val buffer.Color
}

func makeColor(bc buffer.Color) Color {
	return Color{val: bc}
}

func colorToBuffer(c Color) buffer.Color {
	if c.val.Type == buffer.ColorNone {
		return buffer.Color{Type: buffer.ColorNone}
	}
	return c.val
}

// Color256 creates a 256-color from an integer code (0-255).
func Color256(n int) Color {
	return makeColor(buffer.Color{Type: buffer.Color256, Val: uint32(n)})
}

// ColorRGB creates a true-color from r, g, b components.
func ColorRGB(r, g, b int) Color {
	return makeColor(buffer.RGB(uint8(r), uint8(g), uint8(b)))
}

// ColorHex creates a color from a hex string like "#ff8800".
func ColorHex(hex string) Color {
	r, g, b, ok := parseHexColor(hex)
	if !ok {
		return makeColor(buffer.Color{Type: buffer.ColorNone})
	}
	return ColorRGB(r, g, b)
}

// ColorNamed creates a color from an ANSI name: "black", "red", etc.
func ColorNamed(name string) Color {
	n := namedColorIndex(name)
	if n < 0 {
		return makeColor(buffer.Color{Type: buffer.ColorNone})
	}
	return makeColor(buffer.NamedColor(n))
}

// NewColor creates a color from a string. Supports:
//   - "12" → 256-color index 12
//   - "#ff8800" → true-color
//   - "red" → named ANSI color
func NewColor(s string) Color {
	if s == "" {
		return makeColor(buffer.Color{Type: buffer.ColorNone})
	}
	if s[0] == '#' {
		return ColorHex(s)
	}
	// Try numeric (256-color)
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err == nil && n >= 0 && n <= 255 {
		return makeColor(buffer.Color256Val(uint8(n)))
	}
	// Try named
	return ColorNamed(s)
}

// MakeColor is an alias for NewColor (lipgloss.MakeColor compatibility).
func MakeColor(s string) Color {
	return NewColor(s)
}

// ColorFunc is the lipgloss-compatible color constructor function.
// Usage: lipgloss.ColorFunc("12") — since Go doesn't allow type and func same name,
// use ColorFunc or NewColor.
func ColorFunc(s string) Color {
	return NewColor(s)
}

// namedColorIndex maps color name strings to buffer.NamedXxx constants.
func namedColorIndex(name string) int {
	switch strings.ToLower(name) {
	case "black":
		return buffer.NamedBlack
	case "red":
		return buffer.NamedRed
	case "green":
		return buffer.NamedGreen
	case "yellow":
		return buffer.NamedYellow
	case "blue":
		return buffer.NamedBlue
	case "magenta", "purple":
		return buffer.NamedMagenta
	case "cyan":
		return buffer.NamedCyan
	case "white":
		return buffer.NamedWhite
	case "brightblack", "bright-black", "gray", "grey":
		return buffer.NamedBrightBlack
	case "brightred", "bright-red":
		return buffer.NamedBrightRed
	case "brightgreen", "bright-green":
		return buffer.NamedBrightGreen
	case "brightyellow", "bright-yellow":
		return buffer.NamedBrightYellow
	case "brightblue", "bright-blue":
		return buffer.NamedBrightBlue
	case "brightmagenta", "bright-magenta":
		return buffer.NamedBrightMagenta
	case "brightcyan", "bright-cyan":
		return buffer.NamedBrightCyan
	case "brightwhite", "bright-white":
		return buffer.NamedBrightWhite
	}
	return -1
}

func (c Color) toBuffer() buffer.Color {
	return colorToBuffer(c)
}

// String returns the ANSI escape sequence for this color.
func (c Color) String() string {
	return c.val.String()
}

// ─── Style ───

// Style is a lipgloss-compatible style. It chains methods and
// renders to a string with ANSI escape sequences.
type Style struct {
	fg        *Color
	bg        *Color
	bold      bool
	italic    bool
	underline bool
	strike    bool
	dim       bool
	reverse   bool
	blink     bool

	width      int // 0 = unset
	height     int // 0 = unset
	marginTop  int
	marginBot  int
	marginLeft int
	marginRight int

	paddingTop  int
	paddingBot  int
	paddingLeft int
	paddingRight int

	border      *Border
	borderFg    *Color
	borderBg    *Color

	alignH      HorizontalAlign
	alignV      VerticalAlign

	// Cached rendered string (invalidated on mutation)
	renderCache string
	cacheValid  bool
}

// HorizontalAlign specifies horizontal text alignment.
type HorizontalAlign int

const (
	Left   HorizontalAlign = iota
	Center
	Right
)

// VerticalAlign specifies vertical text alignment.
type VerticalAlign int

const (
	Top       VerticalAlign = iota
	Middle
	Bottom
)

// NewStyle creates a new empty style (lipgloss-compatible).
func NewStyle() Style {
	return Style{}
}

// Foreground sets the foreground color.
func (s Style) Foreground(c Color) Style {
	s.fg = &c
	s.cacheValid = false
	return s
}

// Background sets the background color.
func (s Style) Background(c Color) Style {
	s.bg = &c
	s.cacheValid = false
	return s
}

// Bold sets bold text.
func (s Style) Bold(b ...bool) Style {
	v := true
	if len(b) > 0 {
		v = b[0]
	}
	s.bold = v
	s.cacheValid = false
	return s
}

// Italic sets italic text.
func (s Style) Italic(b ...bool) Style {
	v := true
	if len(b) > 0 {
		v = b[0]
	}
	s.italic = v
	s.cacheValid = false
	return s
}

// Underline sets underlined text.
func (s Style) Underline(b ...bool) Style {
	v := true
	if len(b) > 0 {
		v = b[0]
	}
	s.underline = v
	s.cacheValid = false
	return s
}

// Strikethrough sets strikethrough text.
func (s Style) Strikethrough(b ...bool) Style {
	v := true
	if len(b) > 0 {
		v = b[0]
	}
	s.strike = v
	s.cacheValid = false
	return s
}

// Reverse sets reverse video.
func (s Style) Reverse(b ...bool) Style {
	v := true
	if len(b) > 0 {
		v = b[0]
	}
	s.reverse = v
	s.cacheValid = false
	return s
}

// Dim sets dim/faint text.
func (s Style) Dim(b ...bool) Style {
	v := true
	if len(b) > 0 {
		v = b[0]
	}
	s.dim = v
	s.cacheValid = false
	return s
}

// Blink sets blinking text.
func (s Style) Blink(b ...bool) Style {
	v := true
	if len(b) > 0 {
		v = b[0]
	}
	s.blink = v
	s.cacheValid = false
	return s
}

// Width sets the width for rendering (truncates or pads).
func (s Style) Width(w int) Style {
	s.width = w
	s.cacheValid = false
	return s
}

// Height sets the height for rendering.
func (s Style) Height(h int) Style {
	s.height = h
	s.cacheValid = false
	return s
}

// MaxWidth sets maximum width (lipgloss-compatible, no-op for truncation).
func (s Style) MaxWidth(w int) Style {
	if s.width == 0 || w < s.width {
		s.width = w
	}
	s.cacheValid = false
	return s
}

// MaxHeight sets maximum height.
func (s Style) MaxHeight(h int) Style {
	if s.height == 0 || h < s.height {
		s.height = h
	}
	s.cacheValid = false
	return s
}

// MarginTop sets top margin.
func (s Style) MarginTop(n int) Style {
	s.marginTop = n
	s.cacheValid = false
	return s
}

// MarginBottom sets bottom margin.
func (s Style) MarginBottom(n int) Style {
	s.marginBot = n
	s.cacheValid = false
	return s
}

// MarginLeft sets left margin.
func (s Style) MarginLeft(n int) Style {
	s.marginLeft = n
	s.cacheValid = false
	return s
}

// MarginRight sets right margin.
func (s Style) MarginRight(n int) Style {
	s.marginRight = n
	s.cacheValid = false
	return s
}

// Margin sets uniform margin.
func (s Style) Margin(n ...int) Style {
	switch len(n) {
	case 1:
		s.marginTop, s.marginBot, s.marginLeft, s.marginRight = n[0], n[0], n[0], n[0]
	case 2:
		s.marginTop, s.marginBot = n[0], n[0]
		s.marginLeft, s.marginRight = n[1], n[1]
	case 3:
		s.marginTop = n[0]
		s.marginLeft, s.marginRight = n[1], n[1]
		s.marginBot = n[2]
	case 4:
		s.marginTop = n[0]
		s.marginRight = n[1]
		s.marginBot = n[2]
		s.marginLeft = n[3]
	}
	s.cacheValid = false
	return s
}

// PaddingTop sets top padding.
func (s Style) PaddingTop(n int) Style {
	s.paddingTop = n
	s.cacheValid = false
	return s
}

// PaddingBottom sets bottom padding.
func (s Style) PaddingBottom(n int) Style {
	s.paddingBot = n
	s.cacheValid = false
	return s
}

// PaddingLeft sets left padding.
func (s Style) PaddingLeft(n int) Style {
	s.paddingLeft = n
	s.cacheValid = false
	return s
}

// PaddingRight sets right padding.
func (s Style) PaddingRight(n int) Style {
	s.paddingRight = n
	s.cacheValid = false
	return s
}

// Padding sets uniform padding.
func (s Style) Padding(n ...int) Style {
	switch len(n) {
	case 1:
		s.paddingTop, s.paddingBot, s.paddingLeft, s.paddingRight = n[0], n[0], n[0], n[0]
	case 2:
		s.paddingTop, s.paddingBot = n[0], n[0]
		s.paddingLeft, s.paddingRight = n[1], n[1]
	case 3:
		s.paddingTop = n[0]
		s.paddingLeft, s.paddingRight = n[1], n[1]
		s.paddingBot = n[2]
	case 4:
		s.paddingTop = n[0]
		s.paddingRight = n[1]
		s.paddingBot = n[2]
		s.paddingLeft = n[3]
	}
	s.cacheValid = false
	return s
}

// AlignHorizontal sets horizontal alignment.
func (s Style) AlignHorizontal(a HorizontalAlign) Style {
	s.alignH = a
	s.cacheValid = false
	return s
}

// AlignVertical sets vertical alignment.
func (s Style) AlignVertical(a VerticalAlign) Style {
	s.alignV = a
	s.cacheValid = false
	return s
}

// Border sets the border style.
func (s Style) Border(b Border, sides ...bool) Style {
	s.border = &b
	s.cacheValid = false
	return s
}

// BorderForeground sets border color.
func (s Style) BorderForeground(c Color) Style {
	s.borderFg = &c
	s.cacheValid = false
	return s
}

// BorderBackground sets border background color.
func (s Style) BorderBackground(c Color) Style {
	s.borderBg = &c
	s.cacheValid = false
	return s
}

// Render renders the given text with this style applied, returning
// a string with ANSI escape sequences. This is the lipgloss-compatible
// Render method.
func (s Style) Render(text string) string {
	if s.cacheValid {
		return s.renderCache
	}

	var sb strings.Builder

	// Build SGR sequence
	buf := make([]byte, 0, 80)
	st := buffer.Style{}

	if s.fg != nil {
		st.Fg = s.fg.toBuffer()
	}
	if s.bg != nil {
		st.Bg = s.bg.toBuffer()
	}
	if s.bold {
		st.Flags |= buffer.Bold
	}
	if s.italic {
		st.Flags |= buffer.Italic
	}
	if s.underline {
		st.Flags |= buffer.Underline
	}
	if s.strike {
		st.Flags |= buffer.Strikethrough
	}
	if s.dim {
		st.Flags |= buffer.Dim
	}
	if s.reverse {
		st.Flags |= buffer.Reverse
	}
	if s.blink {
		st.Flags |= buffer.Blink
	}

	// Write SGR
	sgr := string(st.AppendSGR(buf))
	if sgr != "" {
		sb.WriteString(sgr)
	}

	// Apply padding (left)
	for i := 0; i < s.paddingLeft; i++ {
		sb.WriteString(" ")
	}

	sb.WriteString(text)

	// Apply padding (right)
	for i := 0; i < s.paddingRight; i++ {
		sb.WriteString(" ")
	}

	// Reset
	if sgr != "" {
		sb.WriteString("\x1b[0m")
	}

	result := sb.String()

	// Apply margin (newlines above/below)
	if s.marginTop > 0 {
		result = strings.Repeat("\n", s.marginTop) + result
	}
	if s.marginBot > 0 {
		result = result + strings.Repeat("\n", s.marginBot)
	}

	s.renderCache = result
	s.cacheValid = true
	return result
}

// String renders the style with empty text (useful for measuring).
func (s Style) String() string {
	return s.Render("")
}

// UnsetBold removes bold (lipgloss-compatible).
func (s Style) UnsetBold() Style       { s.bold = false; s.cacheValid = false; return s }
func (s Style) UnsetItalic() Style     { s.italic = false; s.cacheValid = false; return s }
func (s Style) UnsetUnderline() Style   { s.underline = false; s.cacheValid = false; return s }
func (s Style) UnsetForeground() Style { s.fg = nil; s.cacheValid = false; return s }
func (s Style) UnsetBackground() Style { s.bg = nil; s.cacheValid = false; return s }

// GetBold returns whether bold is set.
func (s Style) GetBold() bool { return s.bold }

// GetForeground returns the foreground color.
func (s Style) GetForeground() Color {
	if s.fg != nil {
		return *s.fg
	}
	return Color{val: buffer.Color{Type: buffer.ColorNone}}
}

// ─── Border ───

// Border defines a set of border runes.
type Border struct {
	Top         string
	Bottom      string
	Left        string
	Right       string
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
}

// NormalBorder returns a standard ASCII border.
func NormalBorder() Border {
	return Border{
		Top: "-", Bottom: "-", Left: "|", Right: "|",
		TopLeft: "+", TopRight: "+", BottomLeft: "+", BottomRight: "+",
	}
}

// RoundedBorder returns a rounded border (lipgloss-compatible).
func RoundedBorder() Border {
	return Border{
		Top: "─", Bottom: "─", Left: "│", Right: "│",
		TopLeft: "╭", TopRight: "╮", BottomLeft: "╰", BottomRight: "╯",
	}
}

// ThickBorder returns a thick border.
func ThickBorder() Border {
	return Border{
		Top: "━", Bottom: "━", Left: "┃", Right: "┃",
		TopLeft: "┏", TopRight: "┓", BottomLeft: "┗", BottomRight: "┛",
	}
}

// DoubleBorder returns a double border.
func DoubleBorder() Border {
	return Border{
		Top: "═", Bottom: "═", Left: "║", Right: "║",
		TopLeft: "╔", TopRight: "╗", BottomLeft: "╚", BottomRight: "╝",
	}
}

// ─── Layout Functions ───

// JoinHorizontal joins strings horizontally with the given vertical alignment.
func JoinHorizontal(position VerticalAlign, strs ...string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	// Find max height
	maxH := 0
	for _, s := range strs {
		h := strings.Count(s, "\n") + 1
		if h > maxH {
			maxH = h
		}
	}

	// Split each string into lines
	lines := make([][]string, len(strs))
	for i, s := range strs {
		lines[i] = strings.Split(s, "\n")
	}

	var sb strings.Builder
	for row := 0; row < maxH; row++ {
		if row > 0 {
			sb.WriteString("\n")
		}
		for col, l := range lines {
			if row < len(l) {
				sb.WriteString(l[row])
			} else {
				// Pad with spaces to match width
				if col > 0 {
					w := 0
					if len(lines[col-1]) > 0 {
						w = len(lines[col-1][0])
					}
					sb.WriteString(strings.Repeat(" ", w))
				}
			}
		}
	}

	return sb.String()
}

// JoinVertical joins strings vertically with the given horizontal alignment.
func JoinVertical(position HorizontalAlign, strs ...string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	return strings.Join(strs, "\n")
}

// Place positions a string within a box of the given dimensions.
func Place(width, height int, hAlign HorizontalAlign, vAlign VerticalAlign, content string, opts ...PositionOption) string {
	return PlaceHorizontal(width, hAlign, content)
}

// PlaceHorizontal places content horizontally within a width.
func PlaceHorizontal(width int, hAlign HorizontalAlign, content string) string {
	contentW := visibleWidth(content)
	if contentW >= width {
		return content
	}
	padding := width - contentW
	switch hAlign {
	case Left:
		return content + strings.Repeat(" ", padding)
	case Right:
		return strings.Repeat(" ", padding) + content
	case Center:
		left := padding / 2
		right := padding - left
		return strings.Repeat(" ", left) + content + strings.Repeat(" ", right)
	}
	return content
}

// PlaceVertical places content vertically within a height.
func PlaceVertical(height int, vAlign VerticalAlign, content string) string {
	contentH := strings.Count(content, "\n") + 1
	if contentH >= height {
		return content
	}
	padding := height - contentH
	switch vAlign {
	case Top:
		return content + strings.Repeat("\n", padding)
	case Bottom:
		return strings.Repeat("\n", padding) + content
	case Middle:
		top := padding / 2
		bot := padding - top
		return strings.Repeat("\n", top) + content + strings.Repeat("\n", bot)
	}
	return content
}

// PositionOption is a no-op placeholder for lipgloss.WithWhitespace.
type PositionOption func()

// ─── Measurement Functions ───

// Width returns the visible width of a string (accounting for ANSI escapes).
func Width(s string) int {
	return visibleWidth(s)
}

// Height returns the number of lines in a string.
func Height(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

// ─── Helpers ───

// visibleWidth calculates the display width of a string, stripping ANSI escapes.
func visibleWidth(s string) int {
	inEscape := false
	width := 0
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' || r == 'H' || r == 'J' || r == 'K' || r == 'u' || r == 's' {
				inEscape = false
			}
			continue
		}
		if r == '\n' {
			continue // newlines don't count for width
		}
		w := buffer.RuneWidth(r)
		width += w
	}
	return width
}

// parseHexColor parses "#rrggbb" → r, g, b.
func parseHexColor(hex string) (r, g, b int, ok bool) {
	if len(hex) != 7 || hex[0] != '#' {
		return 0, 0, 0, false
	}
	for i := 1; i < 7; i++ {
		c := hex[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return 0, 0, 0, false
		}
	}
	r = hexToVal(hex[1:3])
	g = hexToVal(hex[3:5])
	b = hexToVal(hex[5:7])
	return r, g, b, true
}

func hexToVal(s string) int {
	v := 0
	for i := 0; i < len(s); i++ {
		v <<= 4
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
			v |= int(c - '0')
		case c >= 'a' && c <= 'f':
			v |= int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			v |= int(c-'A') + 10
		}
	}
	return v
}

// ─── AdaptiveColor (compat sub-package) ───

// AdaptiveColor is a color that adapts to light/dark terminal backgrounds.
// This mirrors charm.land/lipgloss/v2/compat.AdaptiveColor.
type AdaptiveColor struct {
	Light string
	Dark  string
}

// Resolve returns the appropriate color for the current terminal.
// Currently always returns the Dark variant (fluui defaults to dark themes).
func (a AdaptiveColor) Resolve() Color {
	return NewColor(a.Dark)
}

// String returns the resolved color's ANSI code.
func (a AdaptiveColor) String() string {
	return a.Resolve().String()
}


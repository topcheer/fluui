package buffer

import "strings"

// StyleFlags is a bitmask of text attributes.
type StyleFlags uint8

const (
	Bold          StyleFlags = 1 << iota // 1
	Italic                               // 2
	Underline                            // 4
	Strikethrough                        // 8
	Reverse                              // 16
	Dim                                  // 32
	Blink                                // 64
)

// Style is the complete visual style for a Cell.
type Style struct {
	Fg    Color
	Bg    Color
	Flags StyleFlags
}

// DefaultStyle is a Style with no attributes.
var DefaultStyle = Style{}

// WithFg returns a copy with the given foreground color.
func (s Style) WithFg(c Color) Style { s.Fg = c; return s }

// WithBg returns a copy with the given background color.
func (s Style) WithBg(c Color) Style { s.Bg = c; return s }

// WithFlags returns a copy with the given flags (replacing).
func (s Style) WithFlags(f StyleFlags) Style { s.Flags = f; return s }

// AddFlags returns a copy with additional flags OR'd in.
func (s Style) AddFlags(f StyleFlags) Style { s.Flags |= f; return s }

// HasFlag reports whether the given flag bit is set.
func (s Style) HasFlag(f StyleFlags) bool { return s.Flags&f != 0 }

// Equal reports whether two styles are identical.
func (s Style) Equal(o Style) bool {
	return s.Fg.Equal(o.Fg) && s.Bg.Equal(o.Bg) && s.Flags == o.Flags
}

// SGRSequence returns the SGR (Select Graphic Rendition) parameter string
// for this style, e.g. "1;38;2;255;128;0".
func (s Style) SGRSequence() string {
	var sb strings.Builder
	sb.Grow(64) // pre-allocate: worst case is 7 flags + fg + bg ≈ 50 chars

	first := true
	addPart := func(p string) {
		if p == "" {
			return
		}
		if !first {
			sb.WriteByte(';')
		}
		sb.WriteString(p)
		first = false
	}

	if s.Flags&Bold != 0 {
		addPart("1")
	}
	if s.Flags&Dim != 0 {
		addPart("2")
	}
	if s.Flags&Italic != 0 {
		addPart("3")
	}
	if s.Flags&Underline != 0 {
		addPart("4")
	}
	if s.Flags&Blink != 0 {
		addPart("5")
	}
	if s.Flags&Reverse != 0 {
		addPart("7")
	}
	if s.Flags&Strikethrough != 0 {
		addPart("9")
	}

	addPart(s.Fg.FGSequence())
	addPart(s.Bg.BGSequence())

	if first {
		return "0" // no style attributes, emit reset
	}
	return sb.String()
}

// ResetSGR returns the ANSI reset sequence.
const ResetSGR = "\x1b[0m"

// Link holds hyperlink metadata for clickable text.
type Link struct {
	URL  string
	Text string
}

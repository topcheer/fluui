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
	sb.Grow(64)

	first := true

	if s.Flags&Bold != 0 {
		if !first {
			sb.WriteByte(';')
		}
		sb.WriteByte('1')
		first = false
	}
	if s.Flags&Dim != 0 {
		if !first {
			sb.WriteByte(';')
		}
		sb.WriteByte('2')
		first = false
	}
	if s.Flags&Italic != 0 {
		if !first {
			sb.WriteByte(';')
		}
		sb.WriteByte('3')
		first = false
	}
	if s.Flags&Underline != 0 {
		if !first {
			sb.WriteByte(';')
		}
		sb.WriteByte('4')
		first = false
	}
	if s.Flags&Blink != 0 {
		if !first {
			sb.WriteByte(';')
		}
		sb.WriteByte('5')
		first = false
	}
	if s.Flags&Reverse != 0 {
		if !first {
			sb.WriteByte(';')
		}
		sb.WriteByte('7')
		first = false
	}
	if s.Flags&Strikethrough != 0 {
		if !first {
			sb.WriteByte(';')
		}
		sb.WriteByte('9')
		first = false
	}

	// Write FG and BG sequences directly into the builder's byte buffer
	// via byte-level methods to avoid intermediate string allocations.
	var tmp [32]byte
	if !first {
		sb.WriteByte(';')
	}
	sb.Write(s.Fg.appendFG(tmp[:0]))
	sb.WriteByte(';')
	sb.Write(s.Bg.appendBG(tmp[:0]))

	_ = first // first is always false after fg/bg are written
	return sb.String()
}

// ResetSGR returns the ANSI reset sequence.
const ResetSGR = "\x1b[0m"

// Link holds hyperlink metadata for clickable text.
type Link struct {
	URL  string
	Text string
}

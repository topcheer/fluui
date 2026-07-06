package buffer



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
	var buf [80]byte
	return string(s.AppendSGR(buf[:0]))
}

// AppendSGR writes the SGR parameter bytes for this style into b (without the
// ESC[ prefix or 'm' suffix) and returns the updated slice. Callers should use
// this instead of SGRSequence() when writing to a bytes.Buffer to avoid the
// intermediate string allocation.
func (s Style) AppendSGR(b []byte) []byte {
	first := true

	if s.Flags&Bold != 0 {
		if !first {
			b = append(b, ';')
		}
		b = append(b, '1')
		first = false
	}
	if s.Flags&Dim != 0 {
		if !first {
			b = append(b, ';')
		}
		b = append(b, '2')
		first = false
	}
	if s.Flags&Italic != 0 {
		if !first {
			b = append(b, ';')
		}
		b = append(b, '3')
		first = false
	}
	if s.Flags&Underline != 0 {
		if !first {
			b = append(b, ';')
		}
		b = append(b, '4')
		first = false
	}
	if s.Flags&Blink != 0 {
		if !first {
			b = append(b, ';')
		}
		b = append(b, '5')
		first = false
	}
	if s.Flags&Reverse != 0 {
		if !first {
			b = append(b, ';')
		}
		b = append(b, '7')
		first = false
	}
	if s.Flags&Strikethrough != 0 {
		if !first {
			b = append(b, ';')
		}
		b = append(b, '9')
		first = false
	}

	// Write FG and BG color sequences directly into b.
	if !first {
		b = append(b, ';')
	}
	b = s.Fg.appendFG(b)
	b = append(b, ';')
	b = s.Bg.appendBG(b)

	return b
}

// ResetSGR returns the ANSI reset sequence.
const ResetSGR = "\x1b[0m"

// Link holds hyperlink metadata for clickable text.
type Link struct {
	URL  string
	Text string
}

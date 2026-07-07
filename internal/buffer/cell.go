package buffer

// Cell is the smallest renderable unit — one character position on screen.
// Field order optimized for minimal struct size: large fields first, then
// small fields packed together. This reduces Cell from 48→32 bytes (-33%).
type Cell struct {
	Rune  rune       // 4B at offset 0
	Fg    Color      // 8B at offset 4 (4-byte aligned from Rune end)
	Bg    Color      // 8B at offset 12
	Width uint8      // 1B at offset 20 — display width: 0/1/2
	Flags StyleFlags // 1B at offset 21
	Link  *Link      // 8B at offset 24 (8-byte aligned from 22→24)
}

// BlankCell is the default empty cell (space character).
var BlankCell = Cell{Rune: ' ', Width: 1}

// NewCell creates a Cell from a rune and style.
func NewCell(r rune, style Style) Cell {
	return Cell{
		Rune:  r,
		Width: uint8(RuneWidth(r)),
		Fg:    style.Fg,
		Bg:    style.Bg,
		Flags: style.Flags,
	}
}

// StyledCell creates a Cell with explicit style fields.
func StyledCell(r rune, w int, fg, bg Color, flags StyleFlags) Cell {
	return Cell{Rune: r, Width: uint8(w), Fg: fg, Bg: bg, Flags: flags}
}

// Equal reports whether two cells are identical.
func (c Cell) Equal(o Cell) bool {
	if c.Rune != o.Rune || c.Width != o.Width {
		return false
	}
	if !c.Fg.Equal(o.Fg) || !c.Bg.Equal(o.Bg) {
		return false
	}
	if c.Flags != o.Flags {
		return false
	}
	// Compare links
	if c.Link == nil && o.Link == nil {
		return true
	}
	if c.Link == nil || o.Link == nil {
		return false
	}
	return c.Link.URL == o.Link.URL
}

// WithStyle returns a copy of the cell with a new style.
func (c Cell) WithStyle(s Style) Cell {
	c.Fg = s.Fg
	c.Bg = s.Bg
	c.Flags = s.Flags
	return c
}

// WithFg returns a copy with a new foreground color.
func (c Cell) WithFg(fg Color) Cell { c.Fg = fg; return c }

// WithBg returns a copy with a new background color.
func (c Cell) WithBg(bg Color) Cell { c.Bg = bg; return c }

// AddFlags returns a copy with additional style flags.
func (c Cell) AddFlags(f StyleFlags) Cell { c.Flags |= f; return c }

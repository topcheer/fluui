package buffer

import "strings"

// Rect describes a rectangular area in the buffer.
type Rect struct {
	X, Y, W, H int
}

// Buffer is a 2D grid of Cells.
type Buffer struct {
	Width  int
	Height int
	Cells  []Cell // length = Width * Height
}

// NewBuffer creates a buffer filled with blank cells.
func NewBuffer(w, h int) *Buffer {
	b := &Buffer{
		Width:  w,
		Height: h,
		Cells:  make([]Cell, w*h),
	}
	b.Fill(BlankCell)
	return b
}

// Fill fills the entire buffer with the given cell.
func (b *Buffer) Fill(cell Cell) {
	for i := range b.Cells {
		b.Cells[i] = cell
	}
}

// FillRect fills a rectangular area.
func (b *Buffer) FillRect(rect Rect, cell Cell) {
	for y := rect.Y; y < rect.Y+rect.H && y < b.Height; y++ {
		for x := rect.X; x < rect.X+rect.W && x < b.Width; x++ {
			b.SetCell(x, y, cell)
		}
	}
}

// idx returns the array index for (x, y). Returns -1 if out of bounds.
func (b *Buffer) idx(x, y int) int {
	if x < 0 || x >= b.Width || y < 0 || y >= b.Height {
		return -1
	}
	return y*b.Width + x
}

// SetCell sets the cell at (x, y). Does nothing if out of bounds.
func (b *Buffer) SetCell(x, y int, cell Cell) {
	i := b.idx(x, y)
	if i >= 0 {
		b.Cells[i] = cell
	}
}

// GetCell returns the cell at (x, y). Returns BlankCell if out of bounds.
func (b *Buffer) GetCell(x, y int) Cell {
	i := b.idx(x, y)
	if i >= 0 {
		return b.Cells[i]
	}
	return BlankCell
}

// DrawText writes a string starting at (x, y) with the given style.
// Returns the x coordinate after the last character.
func (b *Buffer) DrawText(x, y int, text string, style Style) int {
	for _, r := range text {
		w := RuneWidth(r)
		if x >= b.Width {
			break
		}
		b.SetCell(x, y, Cell{
			Rune:  r,
			Width: w,
			Fg:    style.Fg,
			Bg:    style.Bg,
			Flags: style.Flags,
		})
		// Fill padding for wide chars
		if w == 2 && x+1 < b.Width {
			b.SetCell(x+1, y, Cell{Rune: 0, Width: 0, Bg: style.Bg})
		}
		x += w
	}
	return x
}

// DrawTextClamped writes a string starting at (x, y), clamping to the buffer width.
func (b *Buffer) DrawTextClamped(x, y int, text string, style Style) int {
	maxW := b.Width - x
	if maxW <= 0 {
		return x
	}
	cells := []rune(text)
	curW := 0
	for _, r := range cells {
		w := RuneWidth(r)
		if curW+w > maxW {
			break
		}
		b.SetCell(x, y, Cell{
			Rune:  r,
			Width: w,
			Fg:    style.Fg,
			Bg:    style.Bg,
			Flags: style.Flags,
		})
		if w == 2 {
			b.SetCell(x+1, y, Cell{Rune: 0, Width: 0, Bg: style.Bg})
		}
		x += w
		curW += w
	}
	return x
}

// Blit copies a rectangular region from src to this buffer.
func (b *Buffer) Blit(src *Buffer, srcX, srcY, dstX, dstY, w, h int) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			cell := src.GetCell(srcX+dx, srcY+dy)
			b.SetCell(dstX+dx, dstY+dy, cell)
		}
	}
}

// SetRow writes a slice of cells into row y starting at xOffset.
func (b *Buffer) SetRow(y int, cells []Cell, xOffset int) {
	x := xOffset
	for _, c := range cells {
		if x >= b.Width {
			break
		}
		b.SetCell(x, y, c)
		if c.Width == 2 && x+1 < b.Width {
			b.SetCell(x+1, y, Cell{Rune: 0, Width: 0, Bg: c.Bg})
		}
		x += c.Width
		if c.Width == 0 {
			// combining char, don't advance x
		}
	}
}

// RepeatString returns a string repeated n times.
func RepeatString(s string, n int) string {
	if n <= 0 {
		return ""
	}
	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteString(s)
	}
	return sb.String()
}

// Diff computes the differences between two equally-sized buffers.
// Returns a list of DiffOp describing which cells need updating.
// Optimization: skips entire rows that are identical.
type DiffOp struct {
	X, Y  int
	Cell  Cell
}

// Diff compares front (old) and back (new) buffers, returning ops to transform
// front into back.
func Diff(front, back *Buffer) []DiffOp {
	return DiffInto(front, back, nil)
}

// cellFastEqual is a fast inlined comparison for the common case where
// cells have no links. Falls back to Equal() only when links are present.
func cellFastEqual(a, b Cell) bool {
	// Fast path: if all fields are ==, cells are definitely equal.
	// This covers all cells without links (the vast majority).
	// NOTE: Cell contains a *Link pointer, so == compares pointer values.
	// If both are nil (common), the comparison is correct.
	// If pointers differ but URLs match, we need the full Equal() check.
	if a.Rune == b.Rune && a.Width == b.Width &&
		a.Fg == b.Fg && a.Bg == b.Bg &&
		a.Flags == b.Flags {
		// Quick nil check: if both links are nil, cells are equal.
		if a.Link == nil && b.Link == nil {
			return true
		}
		// If only one is nil, or both are non-nil with same pointer, use ==.
		if a.Link == b.Link {
			return true
		}
		// Different pointers — need full URL comparison.
		return a.Equal(b)
	}
	return false
}

// DiffInto is like Diff but appends to a provided slice (which may have
// remaining capacity) to avoid per-frame allocation. The returned slice
// shares the provided base's underlying array.
func DiffInto(front, back *Buffer, base []DiffOp) []DiffOp {
	ops := base

	if front.Width != back.Width || front.Height != back.Height {
		// Size changed — return everything.
		total := back.Width * back.Height
		if cap(ops) < total {
			ops = make([]DiffOp, 0, total)
		}
		for y := 0; y < back.Height; y++ {
			for x := 0; x < back.Width; x++ {
				ops = append(ops, DiffOp{X: x, Y: y, Cell: back.GetCell(x, y)})
			}
		}
		return ops
	}

	for y := 0; y < back.Height; y++ {
		rowStart := y * back.Width
		rowSame := true
		for x := 0; x < back.Width; x++ {
			if !cellFastEqual(front.Cells[rowStart+x], back.Cells[rowStart+x]) {
				rowSame = false
				break
			}
		}
		if rowSame {
			continue
		}
		for x := 0; x < back.Width; x++ {
			fc := front.Cells[rowStart+x]
			bc := back.Cells[rowStart+x]
			if !cellFastEqual(fc, bc) {
				ops = append(ops, DiffOp{X: x, Y: y, Cell: bc})
			}
		}
	}
	return ops
}

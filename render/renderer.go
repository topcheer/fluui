package render

import (
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// Renderer implements double-buffer diff rendering.
// It maintains a front buffer (last rendered state) and a back buffer
// (currently building state). On EndFrame, it diffs the two and outputs
// only the changed cells.
type Renderer struct {
	tw     *term.Writer
	front  *buffer.Buffer
	back   *buffer.Buffer
	width  int
	height int
}

// New creates a new Renderer.
func New(tw *term.Writer, width, height int) *Renderer {
	r := &Renderer{
		tw:     tw,
		front:  buffer.NewBuffer(width, height),
		back:   buffer.NewBuffer(width, height),
		width:  width,
		height: height,
	}
	return r
}

// Resize re-allocates buffers when terminal size changes.
func (r *Renderer) Resize(width, height int) {
	r.width = width
	r.height = height
	r.front = buffer.NewBuffer(width, height)
	r.back = buffer.NewBuffer(width, height)
	// Force full redraw on next frame
	r.tw.ClearScreen()
}

// Back returns the back buffer for the current frame.
func (r *Renderer) Back() *buffer.Buffer {
	return r.back
}

// BeginFrame resets the back buffer for a new frame.
// It allocates a fresh buffer so the previous front buffer is preserved
// for diffing. The old back buffer is recycled into the front.
func (r *Renderer) BeginFrame() {
	if r.back.Width != r.width || r.back.Height != r.height {
		r.back = buffer.NewBuffer(r.width, r.height)
	} else {
		r.back.Fill(buffer.BlankCell)
	}
}

// EndFrame diffs front vs back and writes the changes to the terminal.
func (r *Renderer) EndFrame() error {
	ops := buffer.Diff(r.front, r.back)

	for _, op := range ops {
		cell := op.Cell
		// Skip padding cells (Width==0) — they are the trailing half of a
		// wide CJK character already rendered by the preceding cell.
		if cell.Width == 0 {
			continue
		}
		r.tw.MoveTo(op.X, op.Y)
		style := buffer.Style{
			Fg:    cell.Fg,
			Bg:    cell.Bg,
			Flags: cell.Flags,
		}
		r.tw.SetStyle(style)
		if cell.Rune != 0 {
			r.tw.WriteString(string(cell.Rune))
		} else {
			r.tw.WriteString(" ")
		}
	}

	r.tw.ResetStyle()

	if err := r.tw.Flush(); err != nil {
		return err
	}

	// Swap buffers: copy back's contents into front so they don't
	// share the same underlying array. On the next BeginFrame, back
	// will be filled with blanks independently.
	if r.front == nil || r.front.Width != r.back.Width || r.front.Height != r.back.Height {
		r.front = buffer.NewBuffer(r.back.Width, r.back.Height)
	}
	copy(r.front.Cells, r.back.Cells)

	return nil
}

// Width returns the current render width.
func (r *Renderer) Width() int { return r.width }

// Height returns the current render height.
func (r *Renderer) Height() int { return r.height }

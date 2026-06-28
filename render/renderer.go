package render

import (
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// asciiChars pre-computes single-character strings for ASCII runes
// to avoid per-cell string(rune) allocation in the hot render path.
var asciiChars [128]string

func init() {
	for i := 0; i < 128; i++ {
		asciiChars[i] = string(rune(i))
	}
}

// Renderer implements double-buffer diff rendering.
type Renderer struct {
	tw      *term.Writer
	front   *buffer.Buffer
	back    *buffer.Buffer
	width   int
	height  int
	runeBuf [4]byte // reusable buffer for rune-to-utf8 encoding
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

// Front returns the front buffer (previous frame).
func (r *Renderer) Front() *buffer.Buffer {
	return r.front
}

// BeginFrame resets the back buffer for a new frame.
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

	// Fast path: no changes detected — skip all terminal I/O and buffer copy.
	if len(ops) == 0 {
		return nil
	}

	for _, op := range ops {
		cell := op.Cell
		// Skip padding cells (Width==0) — trailing half of wide CJK chars.
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
			// Fast path for ASCII — use pre-computed string (zero allocation).
			if cell.Rune < 128 {
				r.tw.WriteString(asciiChars[cell.Rune])
			} else {
				// Encode rune to UTF-8 bytes in stack buffer, write as raw bytes.
				n := utf8.EncodeRune(r.runeBuf[:], cell.Rune)
				r.tw.WriteRaw(r.runeBuf[:n])
			}
		} else {
			r.tw.WriteString(" ")
		}
	}

	r.tw.ResetStyle()

	if err := r.tw.Flush(); err != nil {
		return err
	}

	// Sync front buffer with back.
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

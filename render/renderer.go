package render

import (
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// asciiChars pre-computes single-character strings for ASCII runes
// to avoid per-cell string(rune) allocation in the hot render path.
var asciiChars [128]string

// Pre-computed byte sequences for OSC8 hyperlinks and sync output.
// These are used in EndFrame hot loop; pre-computing avoids repeated
// []byte literal construction (even though escape analysis usually handles
// them, this is guaranteed zero-allocation).
var (
	osc8StartPrefix = []byte{0x1b, ']', '8', ';', ';'}
	osc8ST          = []byte{0x1b, '\\'}
	osc8End         = []byte{0x1b, ']', '8', ';', ';', 0x1b, '\\'}
	syncBegin       = []byte{0x1b, 'P', '=', '1', 's', 0x1b, '\\'}
	syncEnd         = []byte{0x1b, 'P', '=', '2', 's', 0x1b, '\\'}
)

func init() {
	for i := 0; i < 128; i++ {
		asciiChars[i] = string(rune(i))
	}
}

// ImageOverlay represents a raw escape sequence to be emitted at a
// specific terminal position during EndFrame. This is used for terminal
// image protocols (iTerm2, Kitty Graphics, Sixel) which require
// multi-byte sequences at specific cursor positions rather than
// cell-by-cell rendering.
type ImageOverlay struct {
	X, Y     int    // terminal coordinates (0-based)
	Sequence string // raw escape sequence (e.g. iTerm2 OSC 1337, Kitty Graphics)
}

// Renderer implements double-buffer diff rendering.
type Renderer struct {
	tw            *term.Writer
	front         *buffer.Buffer
	back          *buffer.Buffer
	width         int
	height        int
	runeBuf       [4]byte        // reusable buffer for rune-to-utf8 encoding
	syncOutput    bool           // if true, wrap frame output in DCS sync sequences
	diffOps       []buffer.DiffOp // reused across frames to avoid per-frame allocation
	imageOverlays []ImageOverlay // terminal image sequences for this frame
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
	r.diffOps = buffer.DiffInto(r.front, r.back, r.diffOps[:0])
	ops := r.diffOps

	// Fast path: no changes detected — skip all terminal I/O and buffer copy.
	if len(ops) == 0 {
		return nil
	}

	// Synchronized output: BSU (Begin Synchronized Update) must be flushed
	// BEFORE the content so the terminal buffers the upcoming frame.
	// ESC P = 1 s ESC \
	if r.syncOutput {
		r.tw.WriteRaw(syncBegin)
		if err := r.tw.Flush(); err != nil {
			return err
		}
	}

	for _, op := range ops {
		cell := op.Cell
		// Skip padding cells (Width==0) — trailing half of wide CJK chars.
		if cell.Width == 0 {
			continue
		}

		// OSC8 hyperlink: wrap linked cells in escape sequences so they are
		// clickable in terminals that support it (Kitty, iTerm2, WezTerm,
		// GNOME Terminal, etc.).
		if cell.Link != nil {
			// OSC8 start: ESC ] 8 ; <params> ; <url> ST
			r.tw.WriteRaw(osc8StartPrefix)
			r.tw.WriteString(cell.Link.URL)
			r.tw.WriteRaw(osc8ST)
		}

		r.tw.MoveAndStyle(op.X, op.Y, buffer.Style{
			Fg:    cell.Fg,
			Bg:    cell.Bg,
			Flags: cell.Flags,
		})
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

		if cell.Link != nil {
			// OSC8 end: ESC ] 8 ; ; ST
			r.tw.WriteRaw(osc8End)
		}
	}

	r.tw.ResetStyle()

	// Flush content (style + cell data).
	if err := r.tw.Flush(); err != nil {
		return err
	}

	// Emit image overlays (iTerm2/Kitty/Sixel) at their terminal positions.
	// These are raw escape sequences that the terminal interprets to display
	// images inline. Overlays are emitted after cell content so they appear
	// on top of any placeholder text.
	for i := range r.imageOverlays {
		ov := &r.imageOverlays[i]
		r.tw.MoveTo(ov.X, ov.Y)
		r.tw.WriteRaw([]byte(ov.Sequence))
	}
	if len(r.imageOverlays) > 0 {
		if err := r.tw.Flush(); err != nil {
			return err
		}
		r.imageOverlays = r.imageOverlays[:0] // reset for next frame
	}

	// Synchronized output: ESU (End Synchronized Update) must be flushed
	// AFTER the content so the terminal renders the buffered frame atomically.
	// ESC P = 2 s ESC \
	if r.syncOutput {
		r.tw.WriteRaw(syncEnd)
		if err := r.tw.Flush(); err != nil {
			return err
		}
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

// SetSyncOutput enables or disables synchronized output (DCS sync).
// When enabled, each EndFrame wraps its output in BSU/ESU sequences
// to eliminate visual flicker on terminals that support it (Kitty,
// WezTerm, Alacritty, foot, ghostty).
func (r *Renderer) SetSyncOutput(enabled bool) {
	r.syncOutput = enabled
}

// SyncOutput returns whether synchronized output is enabled.
func (r *Renderer) SyncOutput() bool {
	return r.syncOutput
}

// AddImageOverlay queues an image escape sequence to be emitted at the given
// terminal position during EndFrame. The sequence is written after all cell
// content has been rendered. Multiple overlays can be added per frame.
func (r *Renderer) AddImageOverlay(x, y int, sequence string) {
	r.imageOverlays = append(r.imageOverlays, ImageOverlay{X: x, Y: y, Sequence: sequence})
}

// ClearImageOverlays removes all pending image overlays without emitting them.
func (r *Renderer) ClearImageOverlays() {
	r.imageOverlays = r.imageOverlays[:0]
}

package term

import (
	"bytes"
	"io"
	"strconv"

	"github.com/topcheer/fluui/internal/buffer"
)

// Writer batches ANSI escape sequences for efficient output.
type Writer struct {
	w       io.Writer
	profile ColorProfile
	buf     bytes.Buffer

	// Current style tracking (to avoid redundant SGR sequences)
	curStyle buffer.Style
	styleSet bool
}

// NewWriter creates a new ANSI writer.
func NewWriter(w io.Writer, profile ColorProfile) *Writer {
	return &Writer{w: w, profile: profile}
}

// MoveTo moves the cursor to (x, y), 1-based.
// cursorMovePrefix is ESC[ as a 2-byte slice for single Write call.
var cursorMovePrefix = []byte{0x1b, '['}

func (w *Writer) MoveTo(x, y int) {
	// Format: ESC [ <y+1> ; <x+1> H
	// Build entire sequence in stack buffer, single Write to reduce calls.
	var buf [24]byte
	b := append(buf[:0], cursorMovePrefix...)
	b = strconv.AppendInt(b, int64(y+1), 10)
	b = append(b, ';')
	b = strconv.AppendInt(b, int64(x+1), 10)
	b = append(b, 'H')
	w.buf.Write(b)
}

// HideCursor hides the terminal cursor.
func (w *Writer) HideCursor() {
	w.buf.WriteString("\x1b[?25l")
}

// ShowCursor shows the terminal cursor.
func (w *Writer) ShowCursor() {
	w.buf.WriteString("\x1b[?25h")
}

// SetStyle sets the current text style. Only emits SGR if changed.
func (w *Writer) SetStyle(s buffer.Style) {
	if w.styleSet && w.curStyle.Equal(s) {
		return
	}
	w.curStyle = s
	w.styleSet = true
	// Fast path: fully default style (no flags, no colors) — emit reset.
	// ESC[0m (4 bytes) is shorter than ESC[39;49m (8 bytes) and resets all
	// attributes in one shot.
	if s.Flags == 0 && s.Fg.Type == buffer.ColorNone && s.Bg.Type == buffer.ColorNone {
		w.buf.WriteString(buffer.ResetSGR)
		return
	}
	// Write SGR escape sequence directly into buf using byte-level AppendSGR
	// to avoid the intermediate string allocation from SGRSequence().
	// Combine ESC[ + params + 'm' into a single Write.
	var tmp [84]byte
	b := append(tmp[:0], cursorMovePrefix...)
	b = s.AppendSGR(b)
	b = append(b, 'm')
	w.buf.Write(b)
}

// MoveAndStyle combines MoveTo + SetStyle into a single buffer write.
// This halves the number of bytes.Buffer.Write calls vs calling them
// separately, which is the hot path in the renderer's EndFrame loop.
func (w *Writer) MoveAndStyle(x, y int, s buffer.Style) {
	var tmp [128]byte
	b := append(tmp[:0], cursorMovePrefix...)
	b = strconv.AppendInt(b, int64(y+1), 10)
	b = append(b, ';')
	b = strconv.AppendInt(b, int64(x+1), 10)
	b = append(b, 'H')

	// Only emit SGR if style changed.
	if !(w.styleSet && w.curStyle.Equal(s)) {
		w.curStyle = s
		w.styleSet = true
		if s.Flags == 0 && s.Fg.Type == buffer.ColorNone && s.Bg.Type == buffer.ColorNone {
			b = append(b, buffer.ResetSGR...)
		} else {
			b = append(b, cursorMovePrefix...) // ESC[
			b = s.AppendSGR(b)
			b = append(b, 'm')
		}
	}

	w.buf.Write(b)
}

// ResetStyle resets to terminal defaults.
func (w *Writer) ResetStyle() {
	w.buf.WriteString(buffer.ResetSGR)
	w.styleSet = false
	w.curStyle = buffer.DefaultStyle
}

// StyleEquals returns true if the given style matches the current style.
func (w *Writer) StyleEquals(s buffer.Style) bool {
	return w.styleSet && w.curStyle.Equal(s)
}

// WriteString writes a string with the current style.
func (w *Writer) WriteString(s string) {
	w.buf.WriteString(s)
}

// WriteRaw writes bytes directly without style tracking.
func (w *Writer) WriteRaw(b []byte) {
	w.buf.Write(b)
}

// ClearLine clears the entire current line.
func (w *Writer) ClearLine() {
	w.buf.WriteString("\x1b[2K")
}

// ClearScreen clears the entire screen.
func (w *Writer) ClearScreen() {
	w.buf.WriteString("\x1b[2J\x1b[H")
}

// Flush sends all buffered output to the terminal.
func (w *Writer) Flush() error {
	if w.buf.Len() == 0 {
		return nil
	}
	_, err := w.w.Write(w.buf.Bytes())
	w.buf.Reset()
	return err
}

// Bytes returns the buffered output without flushing.
func (w *Writer) Bytes() []byte {
	return w.buf.Bytes()
}

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
func (w *Writer) MoveTo(x, y int) {
	// Format: ESC [ <y+1> ; <x+1> H
	// Use strconv to avoid fmt.Fprintf allocations.
	var num [20]byte
	w.buf.WriteByte(0x1b)
	w.buf.WriteByte('[')
	n := strconv.AppendInt(num[:0], int64(y+1), 10)
	w.buf.Write(n)
	w.buf.WriteByte(';')
	n = strconv.AppendInt(num[:0], int64(x+1), 10)
	w.buf.Write(n)
	w.buf.WriteByte('H')
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
	// Write SGR escape sequence directly into buf using byte-level AppendSGR
	// to avoid the intermediate string allocation from SGRSequence().
	w.buf.WriteByte(0x1b)
	w.buf.WriteByte('[')
	var tmp [80]byte
	params := s.AppendSGR(tmp[:0])
	w.buf.Write(params)
	w.buf.WriteByte('m')
}

// ResetStyle resets to terminal defaults.
func (w *Writer) ResetStyle() {
	w.buf.WriteString(buffer.ResetSGR)
	w.styleSet = false
	w.curStyle = buffer.DefaultStyle
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

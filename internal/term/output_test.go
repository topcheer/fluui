package term

import (
	"bytes"
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestNewWriter(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, Profile256)

	if w.profile != Profile256 {
		t.Fatalf("expected profile Profile256, got %d", w.profile)
	}
	if w.buf.Len() != 0 {
		t.Fatal("expected empty buffer on init")
	}
	if w.styleSet {
		t.Fatal("expected styleSet=false on init")
	}
}

func TestWriterMoveTo(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileANSI16)

	w.MoveTo(0, 0)
	w.MoveTo(10, 5)

	output := string(w.Bytes())

	// MoveTo(0,0) → ESC[1;1H (1-based)
	if !strings.Contains(output, "\x1b[1;1H") {
		t.Fatalf("expected ESC[1;1H for MoveTo(0,0), got %q", output)
	}

	// MoveTo(10,5) → ESC[6;11H (1-based: y+1=6, x+1=11)
	if !strings.Contains(output, "\x1b[6;11H") {
		t.Fatalf("expected ESC[6;11H for MoveTo(10,5), got %q", output)
	}
}

func TestWriterHideCursor(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileANSI16)

	w.HideCursor()

	if string(w.Bytes()) != "\x1b[?25l" {
		t.Fatalf("expected ESC[?25l, got %q", string(w.Bytes()))
	}
}

func TestWriterShowCursor(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileANSI16)

	w.ShowCursor()

	if string(w.Bytes()) != "\x1b[?25h" {
		t.Fatalf("expected ESC[?25h, got %q", string(w.Bytes()))
	}
}

func TestWriterSetStyle_FirstCall(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileANSI16)

	style := buffer.Style{Flags: buffer.Bold}
	w.SetStyle(style)

	output := string(w.Bytes())
	if !strings.Contains(output, "\x1b[") || !strings.Contains(output, "m") {
		t.Fatalf("expected SGR sequence on first SetStyle, got %q", output)
	}
	if !w.styleSet {
		t.Fatal("expected styleSet=true after first SetStyle")
	}
}

func TestWriterSetStyle_SameSkips(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileANSI16)

	style := buffer.Style{Flags: buffer.Bold}
	w.SetStyle(style)
	before := w.buf.Len()

	// Same style → should NOT add output.
	w.SetStyle(style)
	after := w.buf.Len()

	if after != before {
		t.Fatalf("expected no additional output for same style (before=%d, after=%d)", before, after)
	}
}

func TestWriterSetStyle_DifferentEmits(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileANSI16)

	w.SetStyle(buffer.Style{Flags: buffer.Bold})
	lenAfterFirst := w.buf.Len()

	// Different style → should emit.
	w.SetStyle(buffer.Style{Flags: buffer.Italic})

	if w.buf.Len() <= lenAfterFirst {
		t.Fatal("expected additional output for different style")
	}

	output := string(w.Bytes()[lenAfterFirst:])
	if !strings.Contains(output, "\x1b[") {
		t.Fatalf("expected SGR sequence for different style, got %q", output)
	}
}

func TestWriterResetStyle(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileANSI16)

	w.SetStyle(buffer.Style{Flags: buffer.Bold})
	w.buf.Reset()

	w.ResetStyle()

	output := string(w.Bytes())
	if !strings.Contains(output, buffer.ResetSGR) {
		t.Fatalf("expected ResetSGR, got %q", output)
	}
	if w.styleSet {
		t.Fatal("expected styleSet=false after ResetStyle")
	}
	if !w.curStyle.Equal(buffer.DefaultStyle) {
		t.Fatal("expected curStyle=DefaultStyle after ResetStyle")
	}
}

func TestWriterWriteString(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileANSI16)

	w.WriteString("hello world")

	if string(w.Bytes()) != "hello world" {
		t.Fatalf("expected 'hello world', got %q", string(w.Bytes()))
	}
}

func TestWriterWriteRaw(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileANSI16)

	w.WriteRaw([]byte{0x1b, '[', '0', 'm'})

	if w.buf.Len() != 4 {
		t.Fatalf("expected 4 bytes, got %d", w.buf.Len())
	}
}

func TestWriterClearLine(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileANSI16)

	w.ClearLine()

	if string(w.Bytes()) != "\x1b[2K" {
		t.Fatalf("expected ESC[2K, got %q", string(w.Bytes()))
	}
}

func TestWriterClearScreen(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileANSI16)

	w.ClearScreen()

	output := string(w.Bytes())
	// ClearScreen emits ESC[2J + ESC[H
	if !strings.Contains(output, "\x1b[2J") {
		t.Fatalf("expected ESC[2J, got %q", output)
	}
	if !strings.Contains(output, "\x1b[H") {
		t.Fatalf("expected ESC[H, got %q", output)
	}
}

func TestWriterFlush(t *testing.T) {
	var underlying bytes.Buffer
	w := NewWriter(&underlying, ProfileANSI16)

	w.HideCursor()
	w.MoveTo(5, 10)
	w.WriteString("test")

	// Nothing flushed yet.
	if underlying.Len() != 0 {
		t.Fatal("expected no output before Flush")
	}

	err := w.Flush()
	if err != nil {
		t.Fatalf("unexpected Flush error: %v", err)
	}

	if underlying.Len() == 0 {
		t.Fatal("expected output after Flush")
	}

	// Buffer should be empty after flush.
	if w.buf.Len() != 0 {
		t.Fatal("expected empty internal buffer after Flush")
	}
}

func TestWriterFlush_Empty(t *testing.T) {
	var underlying bytes.Buffer
	w := NewWriter(&underlying, ProfileANSI16)

	// Flush with nothing buffered → no-op, no error.
	err := w.Flush()
	if err != nil {
		t.Fatalf("unexpected error on empty flush: %v", err)
	}
	if underlying.Len() != 0 {
		t.Fatal("expected no output on empty flush")
	}
}

func TestWriterBytes(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileANSI16)

	w.HideCursor()
	w.WriteString("hello")

	raw := w.Bytes()
	if len(raw) == 0 {
		t.Fatal("expected non-empty Bytes()")
	}

	// Bytes() should not flush.
	if buf.Len() != 0 {
		t.Fatal("expected no output to underlying writer from Bytes()")
	}
}

func TestWriterBatching(t *testing.T) {
	var underlying bytes.Buffer
	w := NewWriter(&underlying, ProfileANSI16)

	// Multiple operations batched.
	w.HideCursor()
	w.MoveTo(0, 0)
	w.SetStyle(buffer.Style{Flags: buffer.Bold})
	w.WriteString("Hello")
	w.ResetStyle()
	w.ShowCursor()

	// Nothing flushed yet.
	if underlying.Len() != 0 {
		t.Fatal("expected all output batched, nothing in underlying writer yet")
	}

	// Single flush.
	err := w.Flush()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := underlying.String()

	// Verify all operations are present in single write.
	checks := []string{
		"\x1b[?25l",     // HideCursor
		"\x1b[1;1H",     // MoveTo(0,0)
		"Hello",         // WriteString
		buffer.ResetSGR, // ResetStyle
		"\x1b[?25h",     // ShowCursor
	}
	for _, want := range checks {
		if !strings.Contains(output, want) {
			t.Fatalf("expected output to contain %q, got %q", want, output)
		}
	}
}

func TestWriterFlushError(t *testing.T) {
	w := NewWriter(&errWriter{}, ProfileANSI16)
	w.WriteString("test")

	err := w.Flush()
	if err == nil {
		t.Fatal("expected error from Flush with failing writer")
	}
}

// errWriter is a writer that always returns an error.
type errWriter struct{}

func (errWriter) Write(b []byte) (int, error) {
	return 0, bytes.ErrTooLarge
}

func TestWriterMoveToMultiple(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileANSI16)

	w.MoveTo(0, 0)
	w.MoveTo(79, 23)
	w.MoveTo(40, 12)

	output := string(w.Bytes())
	if !strings.Contains(output, "\x1b[1;1H") {
		t.Fatalf("expected ESC[1;1H, got %q", output)
	}
	if !strings.Contains(output, "\x1b[24;80H") {
		t.Fatalf("expected ESC[24;80H, got %q", output)
	}
	if !strings.Contains(output, "\x1b[13;41H") {
		t.Fatalf("expected ESC[13;41H, got %q", output)
	}
}

func TestWriterStyleTrackingAcrossFlush(t *testing.T) {
	var underlying bytes.Buffer
	w := NewWriter(&underlying, ProfileANSI16)

	// Set a style, flush.
	style := buffer.Style{Flags: buffer.Bold}
	w.SetStyle(style)
	w.Flush()

	// Same style after flush → should NOT re-emit (styleSet persists).
	w.WriteString("text")
	lenBefore := w.buf.Len()
	w.SetStyle(style)
	lenAfter := w.buf.Len()

	if lenAfter != lenBefore {
		t.Fatalf("expected no re-emit of same style after flush (before=%d, after=%d)", lenBefore, lenAfter)
	}
}

func TestWriterResetStyleClearsTracking(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, ProfileANSI16)

	// Set a style.
	w.SetStyle(buffer.Style{Flags: buffer.Bold})
	w.buf.Reset()

	// Reset → clears tracking.
	w.ResetStyle()
	w.buf.Reset()

	// Same style again → should emit (styleSet was cleared).
	w.SetStyle(buffer.Style{Flags: buffer.Bold})
	if w.buf.Len() == 0 {
		t.Fatal("expected SGR output after ResetStyle + SetStyle")
	}
}

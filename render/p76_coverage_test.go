package render

import (
	"bytes"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// TestP76_BeginFrame_FillPath covers the non-resize Fill path in BeginFrame.
func TestP76_BeginFrame_FillPath(t *testing.T) {
	tw := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(tw, 20, 10)
	// First call: back buffer created
	r.BeginFrame()
	// Second call: same dimensions → Fill path (not resize)
	r.BeginFrame()
	// Third call with resize → NewBuffer path
	r.Resize(15, 8)
	r.BeginFrame()
}

// TestP76_EndFrame_SyncOutput covers the sync output wrapping in EndFrame.
func TestP76_EndFrame_SyncOutput(t *testing.T) {
	tw := term.NewWriter(&bytesBuf{}, term.ProfileTrue)
	r := New(tw, 20, 10)
	r.SetSyncOutput(true)

	// Set a cell so EndFrame has work to do
	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.Cell{Rune: 'X', Width: 1, Fg: buffer.NamedColor(buffer.NamedWhite)})
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame with sync: %v", err)
	}

	// Second frame: no changes → should still work with sync
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame no-change with sync: %v", err)
	}

	r.SetSyncOutput(false)
	// Frame with non-sync output
	r.BeginFrame()
	r.Back().SetCell(1, 0, buffer.Cell{Rune: 'Y', Width: 1, Fg: buffer.NamedColor(buffer.NamedWhite)})
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame without sync: %v", err)
	}
}

// TestP76_EndFrame_NonASCIIRune covers the UTF-8 encoding path in EndFrame.
func TestP76_EndFrame_NonASCIIRune(t *testing.T) {
	tw := term.NewWriter(&bytesBuf{}, term.ProfileTrue)
	r := New(tw, 20, 10)

	r.BeginFrame()
	// Euro sign (U+20AC) — multi-byte UTF-8
	r.Back().SetCell(0, 0, buffer.Cell{Rune: '\u20ac', Width: 1})
	// CJK wide char (U+4E2D 中)
	r.Back().SetCell(1, 0, buffer.Cell{Rune: '\u4e2d', Width: 2})
	// Width=0 padding cell should be skipped
	r.Back().SetCell(3, 0, buffer.Cell{Rune: 0, Width: 0})
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame non-ASCII: %v", err)
	}
}

// TestP76_EndFrame_RuneZero covers the cell.Rune == 0 → space path.
func TestP76_EndFrame_RuneZero(t *testing.T) {
	tw := term.NewWriter(&bytesBuf{}, term.ProfileTrue)
	r := New(tw, 20, 10)

	r.BeginFrame()
	// Set a cell with Rune=0 (should render as space)
	r.Back().SetCell(0, 0, buffer.Cell{Rune: 0, Width: 1})
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame rune-zero: %v", err)
	}
}

// TestP76_EndFrame_FlushError covers the error return path from Flush.
func TestP76_EndFrame_FlushError(t *testing.T) {
	tw := term.NewWriter(&p76ErrorWriter{}, term.ProfileTrue)
	r := New(tw, 20, 10)

	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.Cell{Rune: 'A', Width: 1})
	err := r.EndFrame()
	if err == nil {
		t.Error("Expected error from EndFrame with failing writer")
	}
}

// TestP76_EndFrame_OSC8Link covers the link rendering path in EndFrame.
func TestP76_EndFrame_OSC8Link(t *testing.T) {
	tw := term.NewWriter(&bytesBuf{}, term.ProfileTrue)
	r := New(tw, 20, 10)

	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.Cell{
		Rune:  'A',
		Width: 1,
		Link:  &buffer.Link{URL: "https://example.com"},
	})
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame OSC8: %v", err)
	}
}

// TestP76_EndFrame_FrontResize covers the front buffer resize path.
func TestP76_EndFrame_FrontResize(t *testing.T) {
	tw := term.NewWriter(&bytesBuf{}, term.ProfileTrue)
	r := New(tw, 20, 10)

	// First frame creates front buffer
	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.Cell{Rune: 'A', Width: 1})
	r.EndFrame()

	// Resize and render again — front buffer must be recreated
	r.Resize(15, 8)
	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.Cell{Rune: 'B', Width: 1})
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame after resize: %v", err)
	}
}

// p76ErrorWriter always returns an error on Write.
type p76ErrorWriter struct{}

func (w *p76ErrorWriter) Write(p []byte) (int, error) {
	return 0, errP76Flush
}

var errP76Flush = &p76TestError{"simulated flush error"}

type p76TestError struct{ msg string }

func (e *p76TestError) Error() string { return e.msg }

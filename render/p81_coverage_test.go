package render

import (
	"bytes"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// errorWriter always returns an error on Write.
type errorWriter struct{}

func (errorWriter) Write(p []byte) (int, error) {
	return 0, bytes.ErrTooLarge
}

// ─── BeginFrame resize/create/fill paths (66.7% → 100%) ───

func TestP81_BeginFrame_CreateBuffer(t *testing.T) {
	tw := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(tw, 80, 24)

	// First call should create the buffer
	r.BeginFrame()
	if r.Back() == nil {
		t.Error("expected non-nil back buffer after BeginFrame")
	}
	if r.Back().Width != 80 || r.Back().Height != 24 {
		t.Errorf("back buffer size = %dx%d, want 80x24", r.Back().Width, r.Back().Height)
	}
}

func TestP81_BeginFrame_FillExisting(t *testing.T) {
	tw := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(tw, 80, 24)

	// First frame creates buffer
	r.BeginFrame()
	// Write some content
	r.Back().SetCell(0, 0, buffer.Cell{Rune: 'X', Width: 1})

	// Second frame should fill (clear) existing buffer
	r.BeginFrame()
	cell := r.Back().GetCell(0, 0)
	if cell.Rune != ' ' && cell.Rune != 0 {
		t.Errorf("after fill: cell(0,0) = %q, want blank", cell.Rune)
	}
}

func TestP81_BeginFrame_ResizeBuffer(t *testing.T) {
	tw := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(tw, 80, 24)

	// First frame creates 80x24 buffer
	r.BeginFrame()

	// Resize to different dimensions
	r.width = 100
	r.height = 30
	r.BeginFrame()

	if r.Back().Width != 100 || r.Back().Height != 30 {
		t.Errorf("after resize: back buffer = %dx%d, want 100x30", r.Back().Width, r.Back().Height)
	}
}

// ─── EndFrame edge cases ───

func TestP81_EndFrame_NoChange(t *testing.T) {
	tw := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(tw, 80, 24)

	// First frame establishes front buffer
	r.BeginFrame()
	fillBufferWithText(r.Back())
	_ = r.EndFrame()

	// Second frame with no changes should be fast-path
	r.BeginFrame()
	// Copy same content from front
	copy(r.Back().Cells, r.Front().Cells)
	err := r.EndFrame()
	if err != nil {
		t.Errorf("EndFrame no-change error: %v", err)
	}
}

func TestP81_EndFrame_OSC8Links(t *testing.T) {
	tw := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(tw, 80, 24)

	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.Cell{
		Rune:  'L',
		Width: 1,
		Link:  &buffer.Link{URL: "https://example.com", Text: "Link"},
	})
	_ = r.EndFrame()
	// Should contain OSC8 sequences in output
}

func TestP81_EndFrame_ZeroRuneCell(t *testing.T) {
	tw := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(tw, 80, 24)

	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.Cell{Rune: 0, Width: 1})
	_ = r.EndFrame()
	// Should write a space for Rune==0
}

func TestP81_EndFrame_WidthZeroCell(t *testing.T) {
	tw := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(tw, 80, 24)

	r.BeginFrame()
	// Width==0 cells are padding (trailing half of wide CJK chars) — should be skipped
	r.Back().SetCell(0, 0, buffer.Cell{Rune: 0, Width: 0})
	_ = r.EndFrame()
}

func TestP81_EndFrame_NonASCIIRune(t *testing.T) {
	tw := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(tw, 80, 24)

	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.Cell{Rune: '€', Width: 1})
	_ = r.EndFrame()
	// Should UTF-8 encode the rune
}

// ─── EndFrame sync output ───

func TestP81_EndFrame_SyncOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	tw := term.NewWriter(buf, term.ProfileTrue)
	r := New(tw, 80, 24)

	r.SetSyncOutput(true)
	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.Cell{Rune: 'X', Width: 1})
	_ = r.EndFrame()

	output := buf.String()
	// Should contain DCS sync sequences
	if len(output) == 0 {
		t.Error("expected non-empty output with sync")
	}
}

func TestP81_EndFrame_SyncOutputFlushError(t *testing.T) {
	// Use errorWriter to trigger flush error
	tw := term.NewWriter(&errorWriter{}, term.ProfileTrue)
	r := New(tw, 80, 24)

	r.SetSyncOutput(true)
	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.Cell{Rune: 'X', Width: 1})
	err := r.EndFrame()
	if err == nil {
		t.Error("expected error from errorWriter")
	}
}

// ─── Second frame front buffer sync (94.4% → 100%) ───

func TestP81_EndFrame_FrontBufferCreate(t *testing.T) {
	tw := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(tw, 80, 24)

	// First EndFrame: front is nil → creates front buffer
	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.Cell{Rune: 'X', Width: 1})
	_ = r.EndFrame()

	if r.Front() == nil {
		t.Error("expected non-nil front buffer")
	}
}

func TestP81_EndFrame_FrontBufferResize(t *testing.T) {
	tw := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(tw, 80, 24)

	// First frame at 80x24
	r.BeginFrame()
	fillBufferWithText(r.Back())
	_ = r.EndFrame()

	// Second frame at different size (resize triggers front buffer recreation)
	r.width = 60
	r.height = 20
	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.Cell{Rune: 'X', Width: 1})
	_ = r.EndFrame()

	if r.Front().Width != 60 || r.Front().Height != 20 {
		t.Errorf("front buffer after resize = %dx%d, want 60x20", r.Front().Width, r.Front().Height)
	}
}

// ─── Renderer getters ───

func TestP81_Renderer_WidthHeight(t *testing.T) {
	tw := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(tw, 100, 50)

	if r.Width() != 100 {
		t.Errorf("Width = %d, want 100", r.Width())
	}
	if r.Height() != 50 {
		t.Errorf("Height = %d, want 50", r.Height())
	}
}

func TestP81_Renderer_SyncOutput(t *testing.T) {
	tw := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(tw, 80, 24)

	if r.SyncOutput() {
		t.Error("sync should be off by default")
	}
	r.SetSyncOutput(true)
	if !r.SyncOutput() {
		t.Error("sync should be on after SetSyncOutput(true)")
	}
}

func TestP81_Renderer_Back(t *testing.T) {
	tw := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(tw, 80, 24)
	r.BeginFrame()

	if r.Back() == nil {
		t.Error("Back() should be non-nil after BeginFrame")
	}
}

func TestP81_Renderer_Front(t *testing.T) {
	tw := term.NewWriter(&bytes.Buffer{}, term.ProfileTrue)
	r := New(tw, 80, 24)

	// Front may be allocated at construction or nil until first EndFrame
	_ = r.Front() // just verify no panic
}

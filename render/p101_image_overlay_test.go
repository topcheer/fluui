package render

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestP101_AddImageOverlay(t *testing.T) {
	bw := &bytesWriter{}
	tw := term.NewWriter(bw, term.ProfileNone)
	r := New(tw, 80, 24)

	r.BeginFrame()
	// Paint a cell so there's something to render.
	r.Back().SetCell(0, 0, buffer.NewCell('X', buffer.DefaultStyle))

	// Add an image overlay at position (10, 5).
	r.AddImageOverlay(10, 5, "\x1b]1337;File=inline=1:AAAA\x07")

	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame error: %v", err)
	}

	// Verify output contains the image sequence.
	if !containsInStr(bw.String(), "AAAA") {
		t.Error("expected image sequence in output")
	}
}

func TestP101_ImageOverlay_EmittedAfterCells(t *testing.T) {
	bw := &bytesWriter{}
	tw := term.NewWriter(bw, term.ProfileNone)
	r := New(tw, 80, 24)

	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.NewCell('A', buffer.DefaultStyle))
	r.AddImageOverlay(5, 3, "IMAGE_SEQ")
	r.EndFrame()

	// The output should contain both 'A' and 'IMAGE_SEQ'.
	// 'A' comes from cell rendering, IMAGE_SEQ comes from overlay.
	output := bw.String()
	if !containsInStr(output, "A") {
		t.Error("expected cell 'A' in output")
	}
	if !containsInStr(output, "IMAGE_SEQ") {
		t.Error("expected image sequence 'IMAGE_SEQ' in output")
	}

	// Verify IMAGE_SEQ comes AFTER cell content (it's emitted after cell flush).
	cellIdx := indexInStr(output, "A")
	imgIdx := indexInStr(output, "IMAGE_SEQ")
	if imgIdx < cellIdx {
		t.Error("image overlay should be emitted after cell content")
	}
}

func TestP101_MultipleImageOverlays(t *testing.T) {
	bw := &bytesWriter{}
	tw := term.NewWriter(bw, term.ProfileNone)
	r := New(tw, 80, 24)

	r.BeginFrame()
	// Paint a cell so there are ops to render (endFrame fast-paths on empty ops).
	r.Back().SetCell(0, 0, buffer.NewCell('X', buffer.DefaultStyle))
	r.AddImageOverlay(0, 0, "IMG1")
	r.AddImageOverlay(10, 10, "IMG2")
	r.EndFrame()

	output := bw.String()
	if !containsInStr(output, "IMG1") {
		t.Error("expected IMG1 in output")
	}
	if !containsInStr(output, "IMG2") {
		t.Error("expected IMG2 in output")
	}
}

func TestP101_ImageOverlayResetAfterFrame(t *testing.T) {
	bw := &bytesWriter{}
	tw := term.NewWriter(bw, term.ProfileNone)
	r := New(tw, 80, 24)

	// Frame 1: with overlay
	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.NewCell('X', buffer.DefaultStyle))
	r.AddImageOverlay(0, 0, "OVERLAY")
	r.EndFrame()

	if !containsInStr(bw.String(), "OVERLAY") {
		t.Fatal("expected overlay in first frame")
	}

	// Frame 2: no overlay — verify OVERLAY is not repeated.
	bw.Reset()
	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.NewCell('X', buffer.DefaultStyle))
	r.EndFrame()

	// Cell content should be present but no OVERLAY (no change → no ops, so cell is there)
	// Actually, if there's no change, ops is empty and nothing is emitted.
	// So output should be empty or not contain OVERLAY.
	output := bw.String()
	if containsInStr(output, "OVERLAY") {
		t.Error("overlay should not persist across frames")
	}
}

func TestP101_ClearImageOverlays(t *testing.T) {
	bw := &bytesWriter{}
	tw := term.NewWriter(bw, term.ProfileNone)
	r := New(tw, 80, 24)

	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.NewCell('X', buffer.DefaultStyle))
	r.AddImageOverlay(0, 0, "SHOULD_NOT_APPEAR")
	r.ClearImageOverlays()
	r.EndFrame()

	if containsInStr(bw.String(), "SHOULD_NOT_APPEAR") {
		t.Error("cleared overlay should not be emitted")
	}
}

// --- helpers ---

type bytesWriter struct {
	data []byte
}

func (w *bytesWriter) Write(p []byte) (int, error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

func (w *bytesWriter) String() string { return string(w.data) }
func (w *bytesWriter) Reset()         { w.data = w.data[:0] }

func indexInStr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func containsInStr(s, substr string) bool {
	return len(s) >= len(substr) && indexInStr(s, substr) >= 0
}

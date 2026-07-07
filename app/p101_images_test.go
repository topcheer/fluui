package app

import (
	"testing"

	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/render"
)

func TestP101_RenderImageOverlays_Nil(t *testing.T) {
	chat := NewChatApp(80, 24)
	// Should not panic with nil renderer.
	chat.RenderImageOverlays(nil)
}

func TestP101_RenderImageOverlays_NoImageBlocks(t *testing.T) {
	chat := NewChatApp(80, 24)
	bw := &nopWriter{}
	tw := term.NewWriter(bw, term.ProfileNone)
	r := render.New(tw, 80, 24)

	chat.AddAssistantText()
	chat.RenderImageOverlays(r)
	// No image blocks → no overlays added (no panic, no error).
}

func TestP101_RenderImageOverlays_WithImageBlock(t *testing.T) {
	chat := NewChatApp(80, 24)
	bw := &nopWriter{}
	tw := term.NewWriter(bw, term.ProfileNone)
	r := render.New(tw, 80, 24)

	// Set env to detect iTerm2 protocol.
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("TERM", "xterm-256color")

	img := block.NewImageBlock("img1", "test.png", []byte("testdata"))
	chat.Container().AddBlock(img)

	// Layout and render.
	buf := buffer.NewBuffer(80, 24)
	chat.Render(buf)
	chat.RenderImageOverlays(r)

	// Verify the overlay was added by checking EndFrame output.
	r.BeginFrame()
	// Copy buf into back buffer
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			r.Back().SetCell(x, y, buf.GetCell(x, y))
		}
	}
	chat.RenderImageOverlays(r)
	if err := r.EndFrame(); err != nil {
		t.Fatalf("EndFrame: %v", err)
	}

	// The output should contain the iTerm2 image sequence.
	out := bw.String()
	if !containsStr(out, "1337") {
		t.Error("expected OSC 1337 image sequence in output")
	}
}

func TestP101_RenderImageOverlays_SkipEmptySequence(t *testing.T) {
	chat := NewChatApp(80, 24)
	bw := &nopWriter{}
	tw := term.NewWriter(bw, term.ProfileNone)
	r := render.New(tw, 80, 24)

	// No TERM_PROGRAM set → ImageNone protocol → empty sequence.
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("TERM", "xterm-256color")

	img := block.NewImageBlock("img1", "test.png", []byte("testdata"))
	chat.Container().AddBlock(img)

	buf := buffer.NewBuffer(80, 24)
	chat.Render(buf)
	chat.RenderImageOverlays(r)

	// Empty sequence should not generate overlay — no crash.
}

func TestP101_RenderImageOverlays_MultipleImages(t *testing.T) {
	chat := NewChatApp(80, 24)
	bw := &nopWriter{}
	tw := term.NewWriter(bw, term.ProfileNone)
	r := render.New(tw, 80, 24)

	t.Setenv("TERM_PROGRAM", "iTerm.app")
	t.Setenv("TERM", "xterm-256color")

	chat.Container().AddBlock(block.NewImageBlock("img1", "a.png", []byte("data1")))
	chat.Container().AddBlock(block.NewImageBlock("img2", "b.png", []byte("data2")))
	chat.Container().AddBlock(block.NewImageBlock("img3", "c.png", []byte("data3")))

	buf := buffer.NewBuffer(80, 24)
	chat.Render(buf)
	chat.RenderImageOverlays(r)

	// All three should have been added as overlays.
	r.BeginFrame()
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			r.Back().SetCell(x, y, buf.GetCell(x, y))
		}
	}
	chat.RenderImageOverlays(r)
	r.EndFrame()

	// Output should contain image sequences.
	out := bw.String()
	if !containsStr(out, "1337") {
		t.Error("expected image sequences in output")
	}
}

// --- helpers ---

type nopWriter struct {
	data []byte
}

func (w *nopWriter) Write(p []byte) (int, error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

func (w *nopWriter) String() string { return string(w.data) }

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

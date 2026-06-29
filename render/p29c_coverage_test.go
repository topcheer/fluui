package render

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// P29c coverage tests for render BeginFrame resize path and EndFrame edge cases.

type dummyWriterP29c struct {
	output []byte
}

func (d *dummyWriterP29c) Write(p []byte) (int, error) {
	d.output = append(d.output, p...)
	return len(p), nil
}

func newRendererP29c(w, h int) (*Renderer, *dummyWriterP29c) {
	var dw dummyWriterP29c
	tw := term.NewWriter(&dw, term.ProfileTrue)
	return New(tw, w, h), &dw
}

func TestP29c_BeginFrame_ResizeTriggers(t *testing.T) {
	r, _ := newRendererP29c(40, 10)
	// First BeginFrame creates back buffer
	r.BeginFrame()
	back1 := r.Back()

	// Resize
	r.Resize(50, 20)
	r.BeginFrame()
	back2 := r.Back()

	if back1 == back2 {
		t.Error("back buffer should be new after resize")
	}
	if back2.Width != 50 || back2.Height != 20 {
		t.Errorf("expected 50x20, got %dx%d", back2.Width, back2.Height)
	}
}

func TestP29c_BeginFrame_NoResize_Fills(t *testing.T) {
	r, _ := newRendererP29c(40, 10)
	r.BeginFrame()
	// Write some content
	r.Back().DrawText(0, 0, "Hello", buffer.DefaultStyle)

	// Second BeginFrame without resize should fill (clear)
	r.BeginFrame()
	cell := r.Back().GetCell(0, 0)
	if cell.Rune != ' ' {
		t.Error("BeginFrame should clear the back buffer on same-size frames")
	}
}

func TestP29c_EndFrame_FlushError(t *testing.T) {
	r, _ := newRendererP29c(40, 10)
	// Render a frame with content to force flush
	r.BeginFrame()
	r.Back().DrawText(0, 0, "Test", buffer.DefaultStyle)
	err := r.EndFrame()
	if err != nil {
		t.Errorf("EndFrame should succeed with dummy writer: %v", err)
	}
}

func TestP29c_EndFrame_RuneZero(t *testing.T) {
	r, _ := newRendererP29c(40, 10)
	r.BeginFrame()
	// Set a cell with Rune == 0 (not zero rune, but unset)
	r.Back().SetCell(0, 0, buffer.Cell{Rune: 0, Width: 1})
	if err := r.EndFrame(); err != nil {
		t.Errorf("EndFrame with rune-zero cell: %v", err)
	}
}

func TestP29c_EndFrame_NonASCIIRune(t *testing.T) {
	r, _ := newRendererP29c(40, 10)
	r.BeginFrame()
	r.Back().SetCell(0, 0, buffer.Cell{Rune: '世', Width: 2})
	if err := r.EndFrame(); err != nil {
		t.Errorf("EndFrame with CJK rune: %v", err)
	}
}

func TestP29c_EndFrame_WidthZeroCell(t *testing.T) {
	r, _ := newRendererP29c(40, 10)
	r.BeginFrame()
	// Width==0 cells should be skipped (trailing half of wide chars)
	r.Back().SetCell(0, 0, buffer.Cell{Rune: '世', Width: 2})
	r.Back().SetCell(1, 0, buffer.Cell{Rune: 0, Width: 0}) // padding cell
	if err := r.EndFrame(); err != nil {
		t.Errorf("EndFrame with width-zero cell: %v", err)
	}
}

func TestP29c_Renderer_WidthHeight(t *testing.T) {
	r, _ := newRendererP29c(60, 25)
	if r.Width() != 60 {
		t.Errorf("expected width 60, got %d", r.Width())
	}
	if r.Height() != 25 {
		t.Errorf("expected height 25, got %d", r.Height())
	}
}

func TestP29c_EndFrame_FrontBufferSync(t *testing.T) {
	r, _ := newRendererP29c(40, 10)
	r.BeginFrame()
	r.Back().DrawText(0, 0, "Sync", buffer.DefaultStyle)
	r.EndFrame()

	// After EndFrame, front should match back
	frontCell := r.Front().GetCell(0, 0)
	if frontCell.Rune != 'S' {
		t.Errorf("front buffer should be synced, got %q", string(frontCell.Rune))
	}
}

func TestP29c_EndFrame_FrontBufferResync(t *testing.T) {
	r, _ := newRendererP29c(40, 10)
	// First frame
	r.BeginFrame()
	r.Back().DrawText(0, 0, "A", buffer.DefaultStyle)
	r.EndFrame()

	// Resize and render again — front buffer should be recreated
	r.Resize(50, 15)
	r.BeginFrame()
	r.Back().DrawText(0, 0, "B", buffer.DefaultStyle)
	r.EndFrame()

	frontCell := r.Front().GetCell(0, 0)
	if frontCell.Rune != 'B' {
		t.Errorf("front should reflect new frame after resize, got %q", string(frontCell.Rune))
	}
}

func TestP29c_MultipleFrames_NoChanges(t *testing.T) {
	r, _ := newRendererP29c(40, 10)
	// Frame 1
	r.BeginFrame()
	r.Back().DrawText(0, 0, "Same", buffer.DefaultStyle)
	r.EndFrame()

	// Frame 2 with identical content — should be fast-path (no ops)
	r.BeginFrame()
	r.Back().DrawText(0, 0, "Same", buffer.DefaultStyle)
	err := r.EndFrame()
	if err != nil {
		t.Errorf("EndFrame with no changes: %v", err)
	}
}

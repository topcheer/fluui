package app

import (
	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/render"
)

// RenderImageOverlays scans visible blocks for ImageBlocks and emits their
// terminal image escape sequences via the renderer's overlay system.
//
// This should be called after Render() in the OnPaint callback:
//
//	app.OnPaint(func(buf *buffer.Buffer) {
//	    chat.Render(buf)
//	    chat.RenderImageOverlays(app.Renderer())
//	})
//
// ImageBlocks that are scrolled off-screen or have no image data are skipped.
// The renderer clears overlays automatically at the start of each EndFrame.
func (a *ChatApp) RenderImageOverlays(r *render.Renderer) {
	if r == nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Clear any previous overlays before adding new ones.
	r.ClearImageOverlays()

	// Get visible Y range to skip off-screen blocks.
	visY := a.scrollView.Offset()
	visH := a.height - a.inputHeight - a.paddingTop
	if visH < 1 {
		visH = 1
	}
	visBottom := visY + visH

	for _, b := range a.container.Blocks() {
		img, ok := b.(*block.ImageBlock)
		if !ok {
			continue
		}

		seq := img.Sequence()
		if seq == "" {
			continue // no image protocol supported or no data
		}

		bounds := img.Bounds()
		// Skip blocks that are entirely above or below the visible area.
		if bounds.Y+bounds.H <= visY || bounds.Y >= visBottom {
			continue
		}

		// The image is emitted at the block's terminal position.
		// bounds.X and bounds.Y are already in buffer/terminal coordinates
		// because SetBounds translates through the scroll offset.
		r.AddImageOverlay(bounds.X, bounds.Y, seq)
	}
}

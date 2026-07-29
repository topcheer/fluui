package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ImagePreview: ASCII Art Image Placeholder Display ───
//
// ImagePreview renders an ASCII-art placeholder block for images with
// dimensions, format label, and a checkerboard pattern. Useful in TUI file
// managers and image viewers that can't render actual pixels.
//
// Usage:
//
//	ip := NewImagePreview()
//	ip.SetFormat("PNG")
//	ip.SetDimensions(800, 600)
//	ip.SetLabel("photo.png")
//	ip.Paint(buf)

// ImagePreviewStyle holds styling for ImagePreview.
type ImagePreviewStyle struct {
	Label     buffer.Style
	Dimension buffer.Style
	Format    buffer.Style
	CheckerA  buffer.Style // checkerboard pattern A
	CheckerB  buffer.Style // checkerboard pattern B
	Border    buffer.Style
}

// DefaultImagePreviewStyle returns sensible defaults.
func DefaultImagePreviewStyle() ImagePreviewStyle {
	label := buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold}   // slate-200 bold
	dim := buffer.Style{Fg: buffer.RGB(148, 163, 184)}                          // slate-400
	fmtS := buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Bold}     // blue-400 bold
	chkA := buffer.Style{Fg: buffer.RGB(51, 65, 85)}                            // slate-700
	chkB := buffer.Style{Fg: buffer.RGB(30, 41, 59)}                            // slate-800
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}                         // slate-600
	return ImagePreviewStyle{Label: label, Dimension: dim, Format: fmtS, CheckerA: chkA, CheckerB: chkB, Border: border}
}

// ImagePreview displays an ASCII placeholder for images.
type ImagePreview struct {
	BaseComponent
	mu sync.Mutex

	format     string
	width      int
	height     int
	label      string
	style      ImagePreviewStyle
	// cached display strings
	dimStr string
}

// NewImagePreview creates an ImagePreview with defaults.
func NewImagePreview() *ImagePreview {
	ip := &ImagePreview{
		format: "PNG",
		width:  0,
		height: 0,
		label:  "image",
		style:  DefaultImagePreviewStyle(),
	}
	ip.SetID(GenerateID("imgpreview"))
	return ip
}

// SetFormat sets the image format label (PNG, JPEG, etc).
func (ip *ImagePreview) SetFormat(f string) *ImagePreview {
	ip.mu.Lock()
	ip.format = f
	ip.mu.Unlock()
	return ip
}

// Format returns the format label.
func (ip *ImagePreview) Format() string {
	ip.mu.Lock()
	defer ip.mu.Unlock()
	return ip.format
}

// SetDimensions sets the image pixel dimensions.
func (ip *ImagePreview) SetDimensions(w, h int) *ImagePreview {
	ip.mu.Lock()
	ip.width = w
	ip.height = h
	ip.dimStr = itoa(w) + "x" + itoa(h)
	ip.mu.Unlock()
	return ip
}

// Dimensions returns the image dimensions.
func (ip *ImagePreview) Dimensions() (int, int) {
	ip.mu.Lock()
	defer ip.mu.Unlock()
	return ip.width, ip.height
}

// SetLabel sets the display label (filename).
func (ip *ImagePreview) SetLabel(l string) *ImagePreview {
	ip.mu.Lock()
	ip.label = l
	ip.mu.Unlock()
	return ip
}

// Label returns the display label.
func (ip *ImagePreview) Label() string {
	ip.mu.Lock()
	defer ip.mu.Unlock()
	return ip.label
}

// SetStyle sets the custom style.
func (ip *ImagePreview) SetStyle(s ImagePreviewStyle) *ImagePreview {
	ip.mu.Lock()
	ip.style = s
	ip.mu.Unlock()
	return ip
}

// Measure returns the preferred size.
func (ip *ImagePreview) Measure(cs Constraints) Size {
	w := 30
	h := 10
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the image preview into the buffer.
func (ip *ImagePreview) Paint(buf *buffer.Buffer) {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	b := ip.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 {
		w = 30
	}
	if h < 5 {
		h = 10
	}

	// Draw border
	bs := ip.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 {
				ch = '┌'
			} else if row == 0 && col == w-1 {
				ch = '┐'
			} else if row == h-1 && col == 0 {
				ch = '└'
			} else if row == h-1 && col == w-1 {
				ch = '┘'
			} else if row == 0 || row == h-1 {
				ch = '─'
			} else if col == 0 || col == w-1 {
				ch = '│'
			}
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	// Draw checkerboard pattern in inner area
	chkA := ip.style.CheckerA
	chkB := ip.style.CheckerB
	innerY := y + 2
	innerH := h - 5 // leave room for header + footer
	if innerH < 1 {
		innerH = 1
	}

	for row := 0; row < innerH && innerY+row < buf.Height; row++ {
		for col := 1; col < w-1 && x+col < buf.Width; col++ {
			var style buffer.Style
			var ch rune
			if (row+col)%2 == 0 {
				style = chkA
				ch = '▓'
			} else {
				style = chkB
				ch = '░'
			}
			buf.SetCell(x+col, innerY+row, buffer.Cell{Rune: ch, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
		}
	}

	// Header: format label (left), label (center)
	labelStyle := ip.style.Label
	fmtStyle := ip.style.Format
	headerY := y + 1

	// Format on left
	col := x + 1
	for _, r := range ip.format {
		if col >= x+w-1 || col >= buf.Width {
			break
		}
		buf.SetCell(col, headerY, buffer.Cell{Rune: r, Fg: fmtStyle.Fg, Bg: fmtStyle.Bg, Flags: fmtStyle.Flags, Width: 1})
		col++
	}

	// Label centered
	labelLen := len(ip.label)
	labelStart := x + (w-labelLen)/2
	if labelStart < x+1 {
		labelStart = x + 1
	}
	for i, r := range ip.label {
		cx := labelStart + i
		if cx >= x+w-1 || cx >= buf.Width {
			break
		}
		buf.SetCell(cx, headerY, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
	}

	// Footer: dimensions
	footerY := y + h - 2
	if footerY >= buf.Height {
		footerY = buf.Height - 1
	}
	dimStyle := ip.style.Dimension
	col = x + 1
	for _, r := range ip.dimStr {
		if col >= x+w-1 || col >= buf.Width {
			break
		}
		buf.SetCell(col, footerY, buffer.Cell{Rune: r, Fg: dimStyle.Fg, Bg: dimStyle.Bg, Flags: dimStyle.Flags, Width: 1})
		col++
	}

	// "px" suffix
	pxChars := [2]byte{'p', 'x'}
	for i := 0; i < 2; i++ {
		if col >= x+w-1 || col >= buf.Width {
			break
		}
		buf.SetCell(col, footerY, buffer.Cell{Rune: rune(pxChars[i]), Fg: dimStyle.Fg, Bg: dimStyle.Bg, Flags: dimStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (ip *ImagePreview) Children() []Component { return nil }

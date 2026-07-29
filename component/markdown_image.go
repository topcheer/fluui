package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownImage: Render Markdown Image Tags ───
//
// MarkdownImage parses ![alt](url) patterns and renders them as styled
// placeholder boxes with alt text, URL, and dashed borders.
//
// Usage:
//
//	mi := NewMarkdownImage()
//	mi.SetMarkdown("Logo ![Fluui](https://fluui.dev/logo.png) here")
//	mi.Paint(buf)

// ImageSegmentType classifies a rendered segment.
type ImageSegmentType int

const (
	imgTextSeg  ImageSegmentType = iota
	imgImageSeg
)

// ImageSegment represents a parsed text segment.
type ImageSegment struct {
	Text string
	URL  string
	Type ImageSegmentType
}

// MarkdownImageStyle holds styling.
type MarkdownImageStyle struct {
	Text   buffer.Style
	Alt    buffer.Style
	URL    buffer.Style
	Border buffer.Style // dashed border for placeholder
}

// DefaultMarkdownImageStyle returns defaults.
func DefaultMarkdownImageStyle() MarkdownImageStyle {
	text := buffer.Style{Fg: buffer.RGB(226, 232, 240)}
	alt := buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Italic}
	url := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return MarkdownImageStyle{Text: text, Alt: alt, URL: url, Border: border}
}

// MarkdownImage renders markdown image tags as placeholders.
type MarkdownImage struct {
	BaseComponent
	mu sync.Mutex

	source string
	style  MarkdownImageStyle
	cached []ImageSegment
}

// NewMarkdownImage creates a MarkdownImage.
func NewMarkdownImage() *MarkdownImage {
	mi := &MarkdownImage{style: DefaultMarkdownImageStyle()}
	mi.SetID(GenerateID("mdimage"))
	return mi
}

// SetMarkdown sets source and parses image tags.
func (mi *MarkdownImage) SetMarkdown(source string) *MarkdownImage {
	mi.mu.Lock()
	mi.source = source
	mi.parseLocked()
	mi.mu.Unlock()
	return mi
}

// Markdown returns the raw source.
func (mi *MarkdownImage) Markdown() string {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	return mi.source
}

// SetStyle sets custom style.
func (mi *MarkdownImage) SetStyle(s MarkdownImageStyle) *MarkdownImage {
	mi.mu.Lock()
	mi.style = s
	mi.mu.Unlock()
	return mi
}

// ImageCount returns the number of image segments.
func (mi *MarkdownImage) ImageCount() int {
	mi.mu.Lock()
	defer mi.mu.Unlock()
	count := 0
	for _, seg := range mi.cached {
		if seg.Type == imgImageSeg { count++ }
	}
	return count
}

// parseLocked parses ![alt](url) patterns. Caller holds lock.
func (mi *MarkdownImage) parseLocked() {
	mi.cached = mi.cached[:0]
	if mi.source == "" { return }

	remaining := mi.source
	for len(remaining) > 0 {
		// Look for ![ prefix
		idx := strings.Index(remaining, "![")
		if idx < 0 {
			if remaining != "" {
				mi.cached = append(mi.cached, ImageSegment{Text: remaining, Type: imgTextSeg})
			}
			return
		}
		if idx > 0 {
			mi.cached = append(mi.cached, ImageSegment{Text: remaining[:idx], Type: imgTextSeg})
		}
		afterBracket := remaining[idx+2:]
		closeAlt := strings.Index(afterBracket, "]")
		if closeAlt < 0 {
			mi.cached = append(mi.cached, ImageSegment{Text: remaining[idx:], Type: imgTextSeg})
			return
		}
		alt := afterBracket[:closeAlt]
		afterAlt := afterBracket[closeAlt+1:]
		if len(afterAlt) == 0 || afterAlt[0] != '(' {
			mi.cached = append(mi.cached, ImageSegment{Text: remaining[idx:], Type: imgTextSeg})
			return
		}
		closeParen := strings.Index(afterAlt[1:], ")")
		if closeParen < 0 {
			mi.cached = append(mi.cached, ImageSegment{Text: remaining[idx:], Type: imgTextSeg})
			return
		}
		url := afterAlt[1 : 1+closeParen]
		mi.cached = append(mi.cached, ImageSegment{Text: alt, URL: url, Type: imgImageSeg})
		remaining = afterAlt[1+closeParen+1:]
	}
}

// Measure returns the preferred size.
func (mi *MarkdownImage) Measure(cs Constraints) Size {
	mi.mu.Lock()
	segCount := len(mi.cached)
	mi.mu.Unlock()
	w := 50
	h := segCount + 2
	if h < 3 { h = 3 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the image placeholders into the buffer.
func (mi *MarkdownImage) Paint(buf *buffer.Buffer) {
	mi.mu.Lock()
	defer mi.mu.Unlock()

	b := mi.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 { w = 50 }
	if h < 3 { h = 3 }

	bs := mi.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	col := x + 1
	rowY := y + 1

	for _, seg := range mi.cached {
		if rowY >= y+h-1 || rowY >= buf.Height { break }

		if seg.Type == imgImageSeg {
			altStyle := mi.style.Alt
			urlStyle := mi.style.URL
			borderStyle := mi.style.Border

			// Draw dashed placeholder border [alt](url)
			if col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: '[', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
			col++

			// Alt text
			for _, r := range seg.Text {
				if col >= x+w-1 || col >= buf.Width { break }
				buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: altStyle.Fg, Bg: altStyle.Bg, Flags: altStyle.Flags, Width: 1})
				col++
			}

			if col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: ']', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
			col++

			// URL in parentheses
			if col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: '(', Fg: urlStyle.Fg, Bg: urlStyle.Bg, Flags: urlStyle.Flags, Width: 1})
			col++
			for _, r := range seg.URL {
				if col >= x+w-1 || col >= buf.Width { break }
				buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: urlStyle.Fg, Bg: urlStyle.Bg, Flags: urlStyle.Flags, Width: 1})
				col++
			}
			if col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: ')', Fg: urlStyle.Fg, Bg: urlStyle.Bg, Flags: urlStyle.Flags, Width: 1})
			col++
		} else {
			textStyle := mi.style.Text
			for _, r := range seg.Text {
				if r == '\n' { rowY++; col = x + 1; continue }
				if col >= x+w-1 { rowY++; col = x + 1 }
				if rowY >= y+h-1 || rowY >= buf.Height { break }
				if col < buf.Width {
					buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
				}
				col++
			}
		}
	}
}

// Children returns nil.
func (mi *MarkdownImage) Children() []Component { return nil }

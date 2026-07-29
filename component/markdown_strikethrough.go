package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownStrikethrough: Render Markdown Strikethrough Text ───
//
// MarkdownStrikethrough parses markdown strikethrough markers (~~text~~ or
// ~text~) and renders them with strikethrough style. Also renders plain text
// segments with normal styling.
//
// Usage:
//
//	ms := NewMarkdownStrikethrough()
//	ms.SetMarkdown("This is ~~old~~ and ~also gone~ text.")
//	ms.Paint(buf)

// StrikethroughStyle holds styling for MarkdownStrikethrough.
type StrikethroughStyle struct {
	Text          buffer.Style
	Strikethrough buffer.Style
	Border        buffer.Style
}

// DefaultStrikethroughStyle returns sensible defaults.
func DefaultStrikethroughStyle() StrikethroughStyle {
	text := buffer.Style{Fg: buffer.RGB(226, 232, 240)}    // slate-200
	strike := buffer.Style{Fg: buffer.RGB(148, 163, 184), Flags: buffer.Strikethrough} // slate-400 strikethrough
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}   // slate-600
	return StrikethroughStyle{Text: text, Strikethrough: strike, Border: border}
}

// StrikeSegmentType classifies a rendered segment.
type StrikeSegmentType int

const (
	strikeText StrikeSegmentType = iota
	strikeStruck
)

// StrikeSegment represents a parsed text segment.
type StrikeSegment struct {
	Text string
	Type StrikeSegmentType
}

// MarkdownStrikethrough renders markdown strikethrough text.
type MarkdownStrikethrough struct {
	BaseComponent
	mu sync.Mutex

	source string
	style  StrikethroughStyle

	// cached parsed segments
	cachedSegments []StrikeSegment
}

// NewMarkdownStrikethrough creates a MarkdownStrikethrough with defaults.
func NewMarkdownStrikethrough() *MarkdownStrikethrough {
	ms := &MarkdownStrikethrough{
		style: DefaultStrikethroughStyle(),
	}
	ms.SetID(GenerateID("strike"))
	return ms
}

// SetMarkdown sets the raw markdown source and parses strikethrough segments.
func (ms *MarkdownStrikethrough) SetMarkdown(source string) *MarkdownStrikethrough {
	ms.mu.Lock()
	ms.source = source
	ms.parseLocked()
	ms.mu.Unlock()
	return ms
}

// Markdown returns the raw markdown source.
func (ms *MarkdownStrikethrough) Markdown() string {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.source
}

// SetStyle sets the custom style.
func (ms *MarkdownStrikethrough) SetStyle(s StrikethroughStyle) *MarkdownStrikethrough {
	ms.mu.Lock()
	ms.style = s
	ms.mu.Unlock()
	return ms
}

// StrikethroughCount returns the number of strikethrough segments.
func (ms *MarkdownStrikethrough) StrikethroughCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	count := 0
	for _, seg := range ms.cachedSegments {
		if seg.Type == strikeStruck {
			count++
		}
	}
	return count
}

// parseLocked parses markdown into segments with strikethrough detection.
func (ms *MarkdownStrikethrough) parseLocked() {
	ms.cachedSegments = ms.cachedSegments[:0]
	if ms.source == "" {
		return
	}

	lines := strings.Split(ms.source, "\n")
	for _, line := range lines {
		ms.parseLineLocked(line)
	}
}

// parseLineLocked parses a single line for strikethrough markers.
func (ms *MarkdownStrikethrough) parseLineLocked(line string) {
	remaining := line
	for {
		// Try ~~ (double tilde) first
		idx := strings.Index(remaining, "~~")
		if idx >= 0 {
			// Text before markers
			if idx > 0 {
				ms.cachedSegments = append(ms.cachedSegments, StrikeSegment{Text: remaining[:idx], Type: strikeText})
			}
			afterMarker := remaining[idx+2:]
			endIdx := strings.Index(afterMarker, "~~")
			if endIdx >= 0 {
				ms.cachedSegments = append(ms.cachedSegments, StrikeSegment{Text: afterMarker[:endIdx], Type: strikeStruck})
				remaining = afterMarker[endIdx+2:]
				continue
			}
			// No closing ~~ — treat rest as text
			ms.cachedSegments = append(ms.cachedSegments, StrikeSegment{Text: remaining[idx:], Type: strikeText})
			return
		}

		// Try single ~ (GFM single tilde strikethrough)
		idx = strings.Index(remaining, "~")
		if idx >= 0 {
			afterMarker := remaining[idx+1:]
			endIdx := strings.Index(afterMarker, "~")
			if endIdx >= 0 && endIdx > 0 {
				// Valid single-tilde strikethrough
				if idx > 0 {
					ms.cachedSegments = append(ms.cachedSegments, StrikeSegment{Text: remaining[:idx], Type: strikeText})
				}
				ms.cachedSegments = append(ms.cachedSegments, StrikeSegment{Text: afterMarker[:endIdx], Type: strikeStruck})
				remaining = afterMarker[endIdx+1:]
				continue
			}
		}

		// No markers found — rest is text
		if remaining != "" {
			ms.cachedSegments = append(ms.cachedSegments, StrikeSegment{Text: remaining, Type: strikeText})
		}
		return
	}
}

// Measure returns the preferred size.
func (ms *MarkdownStrikethrough) Measure(cs Constraints) Size {
	ms.mu.Lock()
	segCount := len(ms.cachedSegments)
	ms.mu.Unlock()

	w := 50
	h := segCount + 2
	if h < 3 {
		h = 3
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the strikethrough content into the buffer.
func (ms *MarkdownStrikethrough) Paint(buf *buffer.Buffer) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	b := ms.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 {
		w = 50
	}
	if h < 3 {
		h = 3
	}

	// Draw border
	bs := ms.style.Border
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

	// Render all segments on a single flowing line
	col := x + 1
	rowY := y + 1

	for _, seg := range ms.cachedSegments {
		if rowY >= y+h-1 || rowY >= buf.Height {
			break
		}

		var style buffer.Style
		if seg.Type == strikeStruck {
			style = ms.style.Strikethrough
		} else {
			style = ms.style.Text
		}

		for _, r := range seg.Text {
			if r == '\n' {
				rowY++
				col = x + 1
				continue
			}
			if col >= x+w-1 {
				rowY++
				col = x + 1
			}
			if rowY >= y+h-1 || rowY >= buf.Height {
				break
			}
			if col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
			}
			col++
		}
	}
}

// Children returns nil.
func (ms *MarkdownStrikethrough) Children() []Component { return nil }

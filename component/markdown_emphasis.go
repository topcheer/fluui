package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownEmphasis: Render Markdown Bold and Italic Text ───
//
// MarkdownEmphasis parses markdown emphasis markers (**bold**, *italic*,
// __bold__, _italic_) and renders them with appropriate text styling.
// Supports mixed bold+italic (***text***).
//
// Usage:
//
//	me := NewMarkdownEmphasis()
//	me.SetMarkdown("This is **bold** and *italic* text.")
//	me.Paint(buf)

// EmphasisStyle holds styling for MarkdownEmphasis.
type EmphasisStyle struct {
	Text   buffer.Style
	Bold   buffer.Style
	Italic buffer.Style
	BoldItalic buffer.Style
	Border buffer.Style
}

// DefaultEmphasisStyle returns sensible defaults.
func DefaultEmphasisStyle() EmphasisStyle {
	text := buffer.Style{Fg: buffer.RGB(226, 232, 240)}    // slate-200
	bold := buffer.Style{Fg: buffer.RGB(255, 255, 255), Flags: buffer.Bold}      // white bold
	italic := buffer.Style{Fg: buffer.RGB(203, 213, 225), Flags: buffer.Italic}  // slate-300 italic
	boldItalic := buffer.Style{Fg: buffer.RGB(255, 255, 255), Flags: buffer.Bold | buffer.Italic}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}   // slate-600
	return EmphasisStyle{Text: text, Bold: bold, Italic: italic, BoldItalic: boldItalic, Border: border}
}

// EmphasisType classifies a rendered emphasis segment.
type EmphasisType int

const (
	emphText     EmphasisType = iota
	emphBold
	emphItalic
	emphBoldItalic
)

// EmphasisSegment represents a parsed text segment with emphasis.
type EmphasisSegment struct {
	Text string
	Type EmphasisType
}

// MarkdownEmphasis renders markdown bold and italic text.
type MarkdownEmphasis struct {
	BaseComponent
	mu sync.Mutex

	source string
	style  EmphasisStyle

	// cached parsed segments
	cachedSegments []EmphasisSegment
}

// NewMarkdownEmphasis creates a MarkdownEmphasis with defaults.
func NewMarkdownEmphasis() *MarkdownEmphasis {
	me := &MarkdownEmphasis{
		style: DefaultEmphasisStyle(),
	}
	me.SetID(GenerateID("emphasis"))
	return me
}

// SetMarkdown sets the raw markdown source and parses emphasis segments.
func (me *MarkdownEmphasis) SetMarkdown(source string) *MarkdownEmphasis {
	me.mu.Lock()
	me.source = source
	me.parseLocked()
	me.mu.Unlock()
	return me
}

// Markdown returns the raw markdown source.
func (me *MarkdownEmphasis) Markdown() string {
	me.mu.Lock()
	defer me.mu.Unlock()
	return me.source
}

// SetStyle sets the custom style.
func (me *MarkdownEmphasis) SetStyle(s EmphasisStyle) *MarkdownEmphasis {
	me.mu.Lock()
	me.style = s
	me.mu.Unlock()
	return me
}

// BoldCount returns the number of bold segments.
func (me *MarkdownEmphasis) BoldCount() int {
	me.mu.Lock()
	defer me.mu.Unlock()
	count := 0
	for _, seg := range me.cachedSegments {
		if seg.Type == emphBold || seg.Type == emphBoldItalic {
			count++
		}
	}
	return count
}

// ItalicCount returns the number of italic segments.
func (me *MarkdownEmphasis) ItalicCount() int {
	me.mu.Lock()
	defer me.mu.Unlock()
	count := 0
	for _, seg := range me.cachedSegments {
		if seg.Type == emphItalic || seg.Type == emphBoldItalic {
			count++
		}
	}
	return count
}

// parseLocked parses markdown into segments with emphasis detection.
func (me *MarkdownEmphasis) parseLocked() {
	me.cachedSegments = me.cachedSegments[:0]
	if me.source == "" {
		return
	}

	lines := strings.Split(me.source, "\n")
	for _, line := range lines {
		me.parseLineLocked(line)
	}
}

// parseLineLocked parses a single line for emphasis markers.
func (me *MarkdownEmphasis) parseLineLocked(line string) {
	remaining := line

	for len(remaining) > 0 {
		// Try *** (bold+italic) first
		if idx := strings.Index(remaining, "***"); idx >= 0 {
			if idx > 0 {
				me.cachedSegments = append(me.cachedSegments, EmphasisSegment{Text: remaining[:idx], Type: emphText})
			}
			after := remaining[idx+3:]
			endIdx := strings.Index(after, "***")
			if endIdx >= 0 {
				me.cachedSegments = append(me.cachedSegments, EmphasisSegment{Text: after[:endIdx], Type: emphBoldItalic})
				remaining = after[endIdx+3:]
				continue
			}
		}

		// Try ___ (bold+italic underscore)
		if idx := strings.Index(remaining, "___"); idx >= 0 {
			if idx > 0 {
				me.cachedSegments = append(me.cachedSegments, EmphasisSegment{Text: remaining[:idx], Type: emphText})
			}
			after := remaining[idx+3:]
			endIdx := strings.Index(after, "___")
			if endIdx >= 0 {
				me.cachedSegments = append(me.cachedSegments, EmphasisSegment{Text: after[:endIdx], Type: emphBoldItalic})
				remaining = after[endIdx+3:]
				continue
			}
		}

		// Try ** (bold)
		if idx := strings.Index(remaining, "**"); idx >= 0 {
			if idx > 0 {
				me.cachedSegments = append(me.cachedSegments, EmphasisSegment{Text: remaining[:idx], Type: emphText})
			}
			after := remaining[idx+2:]
			endIdx := strings.Index(after, "**")
			if endIdx >= 0 {
				me.cachedSegments = append(me.cachedSegments, EmphasisSegment{Text: after[:endIdx], Type: emphBold})
				remaining = after[endIdx+2:]
				continue
			}
		}

		// Try __ (bold underscore)
		if idx := strings.Index(remaining, "__"); idx >= 0 {
			if idx > 0 {
				me.cachedSegments = append(me.cachedSegments, EmphasisSegment{Text: remaining[:idx], Type: emphText})
			}
			after := remaining[idx+2:]
			endIdx := strings.Index(after, "__")
			if endIdx >= 0 {
				me.cachedSegments = append(me.cachedSegments, EmphasisSegment{Text: after[:endIdx], Type: emphBold})
				remaining = after[endIdx+2:]
				continue
			}
		}

		// Try * (italic)
		if idx := strings.Index(remaining, "*"); idx >= 0 {
			if idx > 0 {
				me.cachedSegments = append(me.cachedSegments, EmphasisSegment{Text: remaining[:idx], Type: emphText})
			}
			after := remaining[idx+1:]
			endIdx := strings.Index(after, "*")
			if endIdx >= 0 && endIdx > 0 {
				me.cachedSegments = append(me.cachedSegments, EmphasisSegment{Text: after[:endIdx], Type: emphItalic})
				remaining = after[endIdx+1:]
				continue
			}
		}

		// No more markers — rest is text
		me.cachedSegments = append(me.cachedSegments, EmphasisSegment{Text: remaining, Type: emphText})
		return
	}
}

// Measure returns the preferred size.
func (me *MarkdownEmphasis) Measure(cs Constraints) Size {
	me.mu.Lock()
	segCount := len(me.cachedSegments)
	me.mu.Unlock()

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

// Paint renders the emphasis content into the buffer.
func (me *MarkdownEmphasis) Paint(buf *buffer.Buffer) {
	me.mu.Lock()
	defer me.mu.Unlock()

	b := me.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 {
		w = 50
	}
	if h < 3 {
		h = 3
	}

	// Draw border
	bs := me.style.Border
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

	// Render segments on a flowing line
	col := x + 1
	rowY := y + 1

	for _, seg := range me.cachedSegments {
		if rowY >= y+h-1 || rowY >= buf.Height {
			break
		}

		var style buffer.Style
		switch seg.Type {
		case emphBold:
			style = me.style.Bold
		case emphItalic:
			style = me.style.Italic
		case emphBoldItalic:
			style = me.style.BoldItalic
		default:
			style = me.style.Text
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
func (me *MarkdownEmphasis) Children() []Component { return nil }

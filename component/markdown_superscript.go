package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownSuperscript: Render ^superscript^ Text ───
//
// MarkdownSuperscript parses ^text^ markers and renders with superscript
// styling (dim + raised appearance via formatting flags).
//
// Usage:
//
//	ms := NewMarkdownSuperscript()
//	ms.SetMarkdown("E = mc^2^ and x^n^ + y^n^")
//	ms.Paint(buf)

// SuperscriptSegmentType classifies a rendered segment.
type SuperscriptSegmentType int

const (
	superTextSeg SuperscriptSegmentType = iota
	superSupSeg
)

// SuperscriptSegment represents a parsed text segment.
type SuperscriptSegment struct {
	Text string
	Type SuperscriptSegmentType
}

// SuperscriptStyle holds styling.
type SuperscriptStyle struct {
	Text        buffer.Style
	Superscript buffer.Style
	Border      buffer.Style
}

// DefaultSuperscriptStyle returns defaults.
func DefaultSuperscriptStyle() SuperscriptStyle {
	text := buffer.Style{Fg: buffer.RGB(226, 232, 240)}
	sup := buffer.Style{Fg: buffer.RGB(251, 146, 60), Flags: buffer.Dim}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return SuperscriptStyle{Text: text, Superscript: sup, Border: border}
}

// MarkdownSuperscript renders markdown superscript text.
type MarkdownSuperscript struct {
	BaseComponent
	mu sync.Mutex

	source string
	style  SuperscriptStyle
	cached []SuperscriptSegment
}

// NewMarkdownSuperscript creates a MarkdownSuperscript.
func NewMarkdownSuperscript() *MarkdownSuperscript {
	ms := &MarkdownSuperscript{style: DefaultSuperscriptStyle()}
	ms.SetID(GenerateID("supscript"))
	return ms
}

// SetMarkdown sets source and parses superscript markers.
func (ms *MarkdownSuperscript) SetMarkdown(source string) *MarkdownSuperscript {
	ms.mu.Lock()
	ms.source = source
	ms.parseLocked()
	ms.mu.Unlock()
	return ms
}

// Markdown returns the raw source.
func (ms *MarkdownSuperscript) Markdown() string {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.source
}

// SetStyle sets custom style.
func (ms *MarkdownSuperscript) SetStyle(s SuperscriptStyle) *MarkdownSuperscript {
	ms.mu.Lock()
	ms.style = s
	ms.mu.Unlock()
	return ms
}

// SuperscriptCount returns the number of superscript segments.
func (ms *MarkdownSuperscript) SuperscriptCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	count := 0
	for _, seg := range ms.cached {
		if seg.Type == superSupSeg { count++ }
	}
	return count
}

// parseLocked parses ^text^ patterns. Caller holds lock.
func (ms *MarkdownSuperscript) parseLocked() {
	ms.cached = ms.cached[:0]
	if ms.source == "" { return }

	remaining := ms.source
	for len(remaining) > 0 {
		idx := strings.Index(remaining, "^")
		if idx < 0 {
			if remaining != "" {
				ms.cached = append(ms.cached, SuperscriptSegment{Text: remaining, Type: superTextSeg})
			}
			return
		}
		if idx > 0 {
			ms.cached = append(ms.cached, SuperscriptSegment{Text: remaining[:idx], Type: superTextSeg})
		}
		afterCaret := remaining[idx+1:]
		endIdx := strings.Index(afterCaret, "^")
		if endIdx > 0 {
			ms.cached = append(ms.cached, SuperscriptSegment{Text: afterCaret[:endIdx], Type: superSupSeg})
			remaining = afterCaret[endIdx+1:]
			continue
		}
		// No closing caret
		ms.cached = append(ms.cached, SuperscriptSegment{Text: remaining[idx:], Type: superTextSeg})
		return
	}
}

// Measure returns the preferred size.
func (ms *MarkdownSuperscript) Measure(cs Constraints) Size {
	ms.mu.Lock()
	segCount := len(ms.cached)
	ms.mu.Unlock()
	w := 50
	h := segCount + 2
	if h < 3 { h = 3 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the superscript content into the buffer.
func (ms *MarkdownSuperscript) Paint(buf *buffer.Buffer) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	b := ms.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 { w = 50 }
	if h < 3 { h = 3 }

	bs := ms.style.Border
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

	for _, seg := range ms.cached {
		if rowY >= y+h-1 || rowY >= buf.Height { break }
		var style buffer.Style
		if seg.Type == superSupSeg {
			style = ms.style.Superscript
		} else {
			style = ms.style.Text
		}
		for _, r := range seg.Text {
			if r == '\n' { rowY++; col = x + 1; continue }
			if col >= x+w-1 { rowY++; col = x + 1 }
			if rowY >= y+h-1 || rowY >= buf.Height { break }
			if col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
			}
			col++
		}
	}
}

// Children returns nil.
func (ms *MarkdownSuperscript) Children() []Component { return nil }

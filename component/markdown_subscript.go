package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownSubscript: Render ~subscript~ Text ───
//
// MarkdownSubscript parses ~text~ markers and renders with subscript
// styling (dim + lowered appearance via formatting flags).
//
// Usage:
//
//	ms := NewMarkdownSubscript()
//	ms.SetMarkdown("H~2~O and CO~2~ emissions")
//	ms.Paint(buf)

// SubscriptSegmentType classifies a rendered segment.
type SubscriptSegmentType int

const (
	subTextSeg SubscriptSegmentType = iota
	subSubSeg
)

// SubscriptSegment represents a parsed text segment.
type SubscriptSegment struct {
	Text string
	Type SubscriptSegmentType
}

// SubscriptStyle holds styling.
type SubscriptStyle struct {
	Text      buffer.Style
	Subscript buffer.Style
	Border    buffer.Style
}

// DefaultSubscriptStyle returns defaults.
func DefaultSubscriptStyle() SubscriptStyle {
	text := buffer.Style{Fg: buffer.RGB(226, 232, 240)}
	sub := buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Dim}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return SubscriptStyle{Text: text, Subscript: sub, Border: border}
}

// MarkdownSubscript renders markdown subscript text.
type MarkdownSubscript struct {
	BaseComponent
	mu sync.Mutex

	source string
	style  SubscriptStyle
	cached []SubscriptSegment
}

// NewMarkdownSubscript creates a MarkdownSubscript.
func NewMarkdownSubscript() *MarkdownSubscript {
	ms := &MarkdownSubscript{style: DefaultSubscriptStyle()}
	ms.SetID(GenerateID("subscript"))
	return ms
}

// SetMarkdown sets source and parses subscript markers.
func (ms *MarkdownSubscript) SetMarkdown(source string) *MarkdownSubscript {
	ms.mu.Lock()
	ms.source = source
	ms.parseLocked()
	ms.mu.Unlock()
	return ms
}

// Markdown returns the raw source.
func (ms *MarkdownSubscript) Markdown() string {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.source
}

// SetStyle sets custom style.
func (ms *MarkdownSubscript) SetStyle(s SubscriptStyle) *MarkdownSubscript {
	ms.mu.Lock()
	ms.style = s
	ms.mu.Unlock()
	return ms
}

// SubscriptCount returns the number of subscript segments.
func (ms *MarkdownSubscript) SubscriptCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	count := 0
	for _, seg := range ms.cached {
		if seg.Type == subSubSeg { count++ }
	}
	return count
}

// parseLocked parses ~text~ patterns. Caller holds lock.
func (ms *MarkdownSubscript) parseLocked() {
	ms.cached = ms.cached[:0]
	if ms.source == "" { return }

	remaining := ms.source
	for len(remaining) > 0 {
		idx := strings.Index(remaining, "~")
		if idx < 0 {
			if remaining != "" {
				ms.cached = append(ms.cached, SubscriptSegment{Text: remaining, Type: subTextSeg})
			}
			return
		}
		if idx > 0 {
			ms.cached = append(ms.cached, SubscriptSegment{Text: remaining[:idx], Type: subTextSeg})
		}
		afterTilde := remaining[idx+1:]
		endIdx := strings.Index(afterTilde, "~")
		if endIdx > 0 {
			ms.cached = append(ms.cached, SubscriptSegment{Text: afterTilde[:endIdx], Type: subSubSeg})
			remaining = afterTilde[endIdx+1:]
			continue
		}
		ms.cached = append(ms.cached, SubscriptSegment{Text: remaining[idx:], Type: subTextSeg})
		return
	}
}

// Measure returns the preferred size.
func (ms *MarkdownSubscript) Measure(cs Constraints) Size {
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

// Paint renders the subscript content into the buffer.
func (ms *MarkdownSubscript) Paint(buf *buffer.Buffer) {
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
		if seg.Type == subSubSeg {
			style = ms.style.Subscript
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
func (ms *MarkdownSubscript) Children() []Component { return nil }

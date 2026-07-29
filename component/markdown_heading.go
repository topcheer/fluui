package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownHeading: Render Markdown Headings H1-H6 ───
//
// MarkdownHeading parses markdown headings (# H1 to ###### H6) and renders
// them with level-appropriate sizing, color, and underlines.
// H1/H2 get full underline borders, H3-H6 get inline bold styling.
//
// Usage:
//
//	mh := NewMarkdownHeading()
//	mh.SetMarkdown("## Section Title")
//	mh.Paint(buf)

// MarkdownHeadingStyle holds per-level styling.
type MarkdownHeadingStyle struct {
	Levels  [6]buffer.Style // H1-H6
	Underline [2]buffer.Style // H1, H2 underline border
	Border  buffer.Style
}

// DefaultMarkdownHeadingStyle returns defaults.
func DefaultMarkdownHeadingStyle() MarkdownHeadingStyle {
	h1 := buffer.Style{Fg: buffer.RGB(255, 255, 255), Flags: buffer.Bold}
	h2 := buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold}
	h3 := buffer.Style{Fg: buffer.RGB(167, 139, 250), Flags: buffer.Bold}
	h4 := buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Bold}
	h5 := buffer.Style{Fg: buffer.RGB(244, 114, 182), Flags: buffer.Bold}
	h6 := buffer.Style{Fg: buffer.RGB(148, 163, 184), Flags: buffer.Italic}
	ul1 := buffer.Style{Fg: buffer.RGB(226, 232, 240)}
	ul2 := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return MarkdownHeadingStyle{
		Levels: [6]buffer.Style{h1, h2, h3, h4, h5, h6},
		Underline: [2]buffer.Style{ul1, ul2},
		Border: border,
	}
}

// MarkdownHeading renders markdown headings with level-based styling.
type MarkdownHeading struct {
	BaseComponent
	mu sync.Mutex

	source     string
	style      MarkdownHeadingStyle
	cachedText string
	cachedLevel int // 1-6, 0 if no heading
}

// NewMarkdownHeading creates a MarkdownHeading.
func NewMarkdownHeading() *MarkdownHeading {
	mh := &MarkdownHeading{style: DefaultMarkdownHeadingStyle()}
	mh.SetID(GenerateID("mdheading"))
	return mh
}

// SetMarkdown sets the source and parses heading level.
func (mh *MarkdownHeading) SetMarkdown(source string) *MarkdownHeading {
	mh.mu.Lock()
	mh.source = source
	mh.parseLocked()
	mh.mu.Unlock()
	return mh
}

// Markdown returns the raw source.
func (mh *MarkdownHeading) Markdown() string {
	mh.mu.Lock()
	defer mh.mu.Unlock()
	return mh.source
}

// SetStyle sets the custom style.
func (mh *MarkdownHeading) SetStyle(s MarkdownHeadingStyle) *MarkdownHeading {
	mh.mu.Lock()
	mh.style = s
	mh.mu.Unlock()
	return mh
}

// Level returns the heading level (1-6), or 0 if not a heading.
func (mh *MarkdownHeading) Level() int {
	mh.mu.Lock()
	defer mh.mu.Unlock()
	return mh.cachedLevel
}

// Text returns the heading text (without # prefix).
func (mh *MarkdownHeading) Text() string {
	mh.mu.Lock()
	defer mh.mu.Unlock()
	return mh.cachedText
}

// parseLocked extracts heading level and text. Caller holds lock.
func (mh *MarkdownHeading) parseLocked() {
	mh.cachedText = ""
	mh.cachedLevel = 0
	trimmed := strings.TrimSpace(mh.source)
	if trimmed == "" {
		return
	}
	level := 0
	for level < 6 && level < len(trimmed) && trimmed[level] == '#' {
		level++
	}
	if level > 0 && level < len(trimmed) && trimmed[level] == ' ' {
		mh.cachedLevel = level
		mh.cachedText = strings.TrimSpace(trimmed[level+1:])
	} else if level > 0 && level == len(trimmed) {
		mh.cachedLevel = level
		mh.cachedText = ""
	}
}

// Measure returns the preferred size.
func (mh *MarkdownHeading) Measure(cs Constraints) Size {
	mh.mu.Lock()
	lvl := mh.cachedLevel
	mh.mu.Unlock()
	w := 50
	h := 1
	if lvl > 0 && lvl <= 2 {
		h = 2 // text + underline
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the heading into the buffer.
func (mh *MarkdownHeading) Paint(buf *buffer.Buffer) {
	mh.mu.Lock()
	defer mh.mu.Unlock()

	if mh.cachedLevel == 0 {
		return
	}

	b := mh.Bounds()
	x, y := b.X, b.Y
	w := b.W
	if w < 5 { w = 50 }

	level := mh.cachedLevel
	if level < 1 || level > 6 { return }
	levelStyle := mh.style.Levels[level-1]

	// Draw heading text
	col := x
	for _, r := range mh.cachedText {
		if col >= x+w || col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: levelStyle.Fg, Bg: levelStyle.Bg, Flags: levelStyle.Flags, Width: 1})
		col++
	}

	// H1/H2 get underline border
	if level <= 2 {
		ulStyle := mh.style.Underline[level-1]
		borderChar := '═'
		if level == 2 {
			borderChar = '─'
		}
		for c := x; c < x+w && c < buf.Width; c++ {
			if y+1 < buf.Height {
				buf.SetCell(c, y+1, buffer.Cell{Rune: borderChar, Fg: ulStyle.Fg, Bg: ulStyle.Bg, Flags: ulStyle.Flags, Width: 1})
			}
		}
	}
}

// Children returns nil.
func (mh *MarkdownHeading) Children() []Component { return nil }

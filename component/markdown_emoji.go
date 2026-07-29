package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownEmoji: Render Markdown Emoji Shortcodes ───
//
// MarkdownEmoji parses :shortcode: patterns and replaces them with Unicode
// emoji characters. Supports ~30 common shortcodes.
//
// Usage:
//
//	me := NewMarkdownEmoji()
//	me.SetMarkdown("Hello :smile: world :heart:")
//	me.Paint(buf)

// emojiMap maps common shortcodes to Unicode characters.
var emojiMap = map[string]string{
	"smile":       "😄",
	"heart":       "❤",
	"thumbsup":    "👍",
	"thumbsdown":  "👎",
	"fire":        "🔥",
	"rocket":      "🚀",
	"check":       "✅",
	"cross":       "❌",
	"warning":     "⚠",
	"info":        "ℹ",
	"star":        "⭐",
	"tada":        "🎉",
	"bug":         "🐛",
	"zap":         "⚡",
	"book":        "📖",
	"code":        "💻",
	"wrench":      "🔧",
	"package":     "📦",
	"art":         "🎨",
	"memo":        "📝",
	"bulb":        "💡",
	"trophy":      "🏆",
	"clock":       "🕐",
	"email":       "📧",
	"link":        "🔗",
	"lock":        "🔒",
	"key":         "🔑",
	"chart":       "📊",
	"globe":       "🌐",
}

// EmojiSegmentType classifies a rendered segment.
type EmojiSegmentType int

const (
	emojiTextSeg EmojiSegmentType = iota
	emojiEmojiSeg
)

// EmojiSegment represents a parsed text segment.
type EmojiSegment struct {
	Text string
	Type EmojiSegmentType
}

// EmojiStyle holds styling.
type EmojiStyle struct {
	Text   buffer.Style
	Emoji  buffer.Style
	Border buffer.Style
}

// DefaultEmojiStyle returns defaults.
func DefaultEmojiStyle() EmojiStyle {
	text := buffer.Style{Fg: buffer.RGB(226, 232, 240)}
	emoji := buffer.Style{Fg: buffer.RGB(251, 146, 60)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return EmojiStyle{Text: text, Emoji: emoji, Border: border}
}

// MarkdownEmoji renders markdown emoji shortcodes.
type MarkdownEmoji struct {
	BaseComponent
	mu sync.Mutex

	source  string
	style   EmojiStyle
	cached  []EmojiSegment
}

// NewMarkdownEmoji creates a MarkdownEmoji.
func NewMarkdownEmoji() *MarkdownEmoji {
	me := &MarkdownEmoji{style: DefaultEmojiStyle()}
	me.SetID(GenerateID("emoji"))
	return me
}

// SetMarkdown sets source and parses emoji shortcodes.
func (me *MarkdownEmoji) SetMarkdown(source string) *MarkdownEmoji {
	me.mu.Lock()
	me.source = source
	me.parseLocked()
	me.mu.Unlock()
	return me
}

// Markdown returns the raw source.
func (me *MarkdownEmoji) Markdown() string {
	me.mu.Lock()
	defer me.mu.Unlock()
	return me.source
}

// SetStyle sets custom style.
func (me *MarkdownEmoji) SetStyle(s EmojiStyle) *MarkdownEmoji {
	me.mu.Lock()
	me.style = s
	me.mu.Unlock()
	return me
}

// EmojiCount returns the number of emoji segments.
func (me *MarkdownEmoji) EmojiCount() int {
	me.mu.Lock()
	defer me.mu.Unlock()
	count := 0
	for _, seg := range me.cached {
		if seg.Type == emojiEmojiSeg { count++ }
	}
	return count
}

// parseLocked parses :shortcode: patterns. Caller holds lock.
func (me *MarkdownEmoji) parseLocked() {
	me.cached = me.cached[:0]
	if me.source == "" { return }

	remaining := me.source
	for len(remaining) > 0 {
		idx := strings.Index(remaining, ":")
		if idx < 0 {
			if remaining != "" {
				me.cached = append(me.cached, EmojiSegment{Text: remaining, Type: emojiTextSeg})
			}
			return
		}

		if idx > 0 {
			me.cached = append(me.cached, EmojiSegment{Text: remaining[:idx], Type: emojiTextSeg})
		}

		afterColon := remaining[idx+1:]
		endIdx := strings.Index(afterColon, ":")
		if endIdx > 0 {
			code := afterColon[:endIdx]
			if emoji, ok := emojiMap[strings.ToLower(code)]; ok {
				me.cached = append(me.cached, EmojiSegment{Text: emoji, Type: emojiEmojiSeg})
				remaining = afterColon[endIdx+1:]
				continue
			}
		}

		// No match — output colon as text
		me.cached = append(me.cached, EmojiSegment{Text: ":", Type: emojiTextSeg})
		remaining = remaining[idx+1:]
	}
}

// Measure returns the preferred size.
func (me *MarkdownEmoji) Measure(cs Constraints) Size {
	me.mu.Lock()
	segCount := len(me.cached)
	me.mu.Unlock()
	w := 50
	h := segCount + 2
	if h < 3 { h = 3 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the emoji content into the buffer.
func (me *MarkdownEmoji) Paint(buf *buffer.Buffer) {
	me.mu.Lock()
	defer me.mu.Unlock()

	b := me.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 { w = 50 }
	if h < 3 { h = 3 }

	bs := me.style.Border
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

	for _, seg := range me.cached {
		if rowY >= y+h-1 || rowY >= buf.Height { break }

		var style buffer.Style
		if seg.Type == emojiEmojiSeg {
			style = me.style.Emoji
		} else {
			style = me.style.Text
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
func (me *MarkdownEmoji) Children() []Component { return nil }

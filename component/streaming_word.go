package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── StreamingWord: Word-by-Word Streaming Display ───
//
// StreamingWord renders AI text output word by word with a blinking cursor
// at the current position. Useful for simulating typewriter-style streaming.
//
// Usage:
//
//	sw := NewStreamingWord()
//	sw.SetText("The quick brown fox")
//	sw.SetCursor(2) // show first 2 words
//	sw.Paint(buf)

// StreamingWordStyle holds styling.
type StreamingWordStyle struct {
	Text     buffer.Style
	Cursor   buffer.Style
	Unstyled buffer.Style
}

// DefaultStreamingWordStyle returns defaults.
func DefaultStreamingWordStyle() StreamingWordStyle {
	return StreamingWordStyle{
		Text:     buffer.Style{Fg: buffer.RGB(226, 232, 240)},
		Cursor:   buffer.Style{Fg: buffer.RGB(251, 191, 36), Flags: buffer.Bold},
		Unstyled: buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

// StreamingWord renders word-by-word streaming text.
type StreamingWord struct {
	BaseComponent
	mu sync.Mutex

	text      string
	wordCount int // how many words to show
	width     int
	style     StreamingWordStyle
	// cached
	words    []string
	displayW int // number of runes in visible portion
}

// NewStreamingWord creates a StreamingWord.
func NewStreamingWord() *StreamingWord {
	sw := &StreamingWord{width: 50, style: DefaultStreamingWordStyle()}
	sw.SetID(GenerateID("streamword"))
	return sw
}

// SetText sets the full text to stream.
func (sw *StreamingWord) SetText(s string) *StreamingWord {
	sw.mu.Lock()
	sw.text = s
	sw.words = splitWordsSimple(s)
	sw.recomputeLocked()
	sw.mu.Unlock()
	return sw
}

// SetCursor sets how many words are visible.
func (sw *StreamingWord) SetCursor(n int) *StreamingWord {
	sw.mu.Lock()
	if n < 0 { n = 0 }
	sw.wordCount = n
	sw.recomputeLocked()
	sw.mu.Unlock()
	return sw
}

func (sw *StreamingWord) recomputeLocked() {
	if sw.wordCount > len(sw.words) {
		sw.wordCount = len(sw.words)
	}
	// compute display width of visible words
	sw.displayW = 0
	for i := 0; i < sw.wordCount && i < len(sw.words); i++ {
		sw.displayW += len([]rune(sw.words[i])) + 1 // word + space
	}
}

// splitWordsSimple splits text into words by spaces (zero-alloc alternative
// would scan runes, but this is called in setter not Paint).
func splitWordsSimple(s string) []string {
	if s == "" { return nil }
	var words []string
	start := 0
	for i, r := range s {
		if r == ' ' || r == '\t' || r == '\n' {
			if i > start {
				words = append(words, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		words = append(words, s[start:])
	}
	return words
}

// WordCount returns the number of visible words.
func (sw *StreamingWord) WordCount() int {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.wordCount
}

// TotalWords returns the total number of words.
func (sw *StreamingWord) TotalWords() int {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return len(sw.words)
}

// SetWidth sets the display width.
func (sw *StreamingWord) SetWidth(w int) *StreamingWord {
	sw.mu.Lock()
	if w < 10 { w = 10 }
	sw.width = w
	sw.mu.Unlock()
	return sw
}

// SetStyle sets custom style.
func (sw *StreamingWord) SetStyle(s StreamingWordStyle) *StreamingWord {
	sw.mu.Lock()
	sw.style = s
	sw.mu.Unlock()
	return sw
}

// Measure returns preferred size.
func (sw *StreamingWord) Measure(cs Constraints) Size {
	w := sw.width
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// Paint renders the streaming word display.
func (sw *StreamingWord) Paint(buf *buffer.Buffer) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	b := sw.Bounds()
	x, y := b.X, b.Y

	textStyle := sw.style.Text
	cursorStyle := sw.style.Cursor
	unstyledStyle := sw.style.Unstyled

	col := x

	// Render visible words
	for i := 0; i < sw.wordCount && i < len(sw.words); i++ {
		for _, r := range sw.words[i] {
			if col >= buf.Width { return }
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
			col++
		}
	}

	// Blinking cursor at current position
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: '▋', Fg: cursorStyle.Fg, Bg: cursorStyle.Bg, Flags: cursorStyle.Flags, Width: 1})
		col++
	}

	// Remaining words shown dimmed
	for i := sw.wordCount; i < len(sw.words); i++ {
		for _, r := range sw.words[i] {
			if col >= buf.Width { return }
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: unstyledStyle.Fg, Bg: unstyledStyle.Bg, Flags: unstyledStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: unstyledStyle.Fg, Bg: unstyledStyle.Bg, Flags: unstyledStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (sw *StreamingWord) Children() []Component { return nil }

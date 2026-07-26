package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/markdown"
	"github.com/topcheer/fluui/theme"
)

// MarkdownStream renders markdown content with a live streaming cursor,
// designed for AI chat output where text arrives token-by-token.
// Unlike MarkdownViewer, it shows a blinking block cursor at the end
// and supports incremental content appending.
//
// Thread-safe.
type MarkdownStream struct {
	BaseComponent
	mu sync.Mutex

	source     string
	streaming  bool
	cursorOn   bool
	cursorChar rune

	// Renderer (cached)
	renderer       *markdown.MarkdownRenderer
	lastWidth      int
	lastBlocks     []*markdown.Block
	lastErr        error
	renderedSource string
}

// NewMarkdownStream creates a streaming markdown viewer.
func NewMarkdownStream() *MarkdownStream {
	return &MarkdownStream{
		BaseComponent: BaseComponent{id: GenerateID("mdstream")},
		streaming:     true,
		cursorOn:      true,
		cursorChar:    '\u2588', // █ block cursor
	}
}

// Source returns the current markdown source text.
func (m *MarkdownStream) Source() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.source
}

// SetSource replaces the entire markdown content.
func (m *MarkdownStream) SetSource(src string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.source = src
}

// Append adds text to the end of the stream (for token-by-token rendering).
func (m *MarkdownStream) Append(text string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.source += text
}

// Streaming returns whether the streaming cursor is active.
func (m *MarkdownStream) Streaming() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.streaming
}

// SetStreaming toggles the streaming state. When false, no cursor is shown.
func (m *MarkdownStream) SetStreaming(b bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.streaming = b
}

// CursorOn returns the cursor blink state.
func (m *MarkdownStream) CursorOn() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cursorOn
}

// SetCursorOn sets the cursor blink visibility.
func (m *MarkdownStream) SetCursorOn(b bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cursorOn = b
}

// CursorChar returns the cursor character.
func (m *MarkdownStream) CursorChar() rune {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cursorChar
}

// SetCursorChar sets the cursor character (default █).
func (m *MarkdownStream) SetCursorChar(r rune) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cursorChar = r
}

// Tick advances the cursor blink state. Call this on a timer.
func (m *MarkdownStream) Tick() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cursorOn = !m.cursorOn
}

// Measure returns the preferred size.
func (m *MarkdownStream) Measure(cs Constraints) Size {
	w := cs.MaxWidth
	if w <= 0 {
		w = 80
	}
	// Estimate height from source lines
	src := m.Source()
	h := 1
	for i := 0; i < len(src); i++ {
		if src[i] == '\n' {
			h++
		}
	}
	if h < 1 {
		h = 1
	}
	if cs.HasHeight() && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the markdown content with optional streaming cursor.
func (m *MarkdownStream) Paint(buf *buffer.Buffer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	bounds := m.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	src := m.source
	tt := theme.Get()

	// Render markdown via cached renderer
	contentW := bounds.W
	if m.renderer == nil || m.lastWidth != contentW || m.renderedSource != src {
		if m.renderer == nil || m.lastWidth != contentW {
			m.renderer = markdown.NewMarkdownRenderer(markdown.DefaultTheme(), contentW)
			m.lastWidth = contentW
		}
		m.lastBlocks, m.lastErr = m.renderer.Render(src)
		m.renderedSource = src
	}

	if m.lastErr != nil || len(m.lastBlocks) == 0 {
		drawWrappedText(buf, bounds, src, buffer.Style{Fg: tt.Fg})
	} else {
		row := 0
		for _, block := range m.lastBlocks {
			for _, line := range block.Cells {
				if row >= bounds.H { break }
				for col, cell := range line {
					if col >= contentW { break }
					buf.SetCell(bounds.X+col, bounds.Y+row, cell)
				}
				row++
			}
		}
	}

	// Draw streaming cursor at end of last line
	if m.streaming && m.cursorOn && len(src) > 0 {
		// Find the position after the last visible character
		lines := countLinesFast(src)
		cursorY := bounds.Y + lines - 1
		if cursorY < bounds.Y {
			cursorY = bounds.Y
		}
		if cursorY >= bounds.Y+bounds.H {
			return
		}
		// Find X position: last char of last line + 1
		lastLineStart := 0
		for i := len(src) - 1; i >= 0; i-- {
			if src[i] == '\n' {
				lastLineStart = i + 1
				break
			}
		}
		lastLineLen := 0
		for i := lastLineStart; i < len(src); i++ {
			lastLineLen++
		}
		cursorX := bounds.X + lastLineLen
		if cursorX >= bounds.X+bounds.W {
			cursorX = bounds.X + bounds.W - 1
			cursorY++
			if cursorY >= bounds.Y+bounds.H {
				return
			}
			cursorX = bounds.X
		}
		// Draw block cursor
		buf.SetCell(cursorX, cursorY, buffer.Cell{
			Rune:  m.cursorChar,
			Width: 1,
			Fg:    tt.Accent,
			Bg:    tt.Bg,
			Flags: buffer.Bold,
		})
	}
}

// drawWrappedText draws plain text with simple word wrapping.
func drawWrappedText(buf *buffer.Buffer, bounds Rect, text string, style buffer.Style) {
	if text == "" {
		return
	}
	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W
	maxY := bounds.Y + bounds.H

	for i := 0; i < len(text) && y < maxY; i++ {
		if text[i] == '\n' {
			x = bounds.X
			y++
			continue
		}
		if x >= maxX {
			x = bounds.X
			y++
			if y >= maxY {
				break
			}
		}
		buf.SetCell(x, y, buffer.Cell{Rune: rune(text[i]), Width: 1, Fg: style.Fg, Bg: style.Bg})
		x++
	}
}

// countLinesFast counts lines in text (zero alloc).
func countLinesFast(text string) int {
	if text == "" {
		return 1
	}
	count := 1
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			count++
		}
	}
	return count
}

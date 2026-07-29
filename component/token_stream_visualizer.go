package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── TokenStreamVisualizer: Real-time LLM Token Streaming Display ───
//
// TokenStreamVisualizer renders streaming token output with scrolling text,
// progress indicator, and live status (streaming/done/error).
//
// Usage:
//
//	tsv := NewTokenStreamVisualizer()
//	tsv.SetText("Hello world from AI")
//	tsv.SetTotal(100)
//	tsv.SetReceived(50)
//	tsv.SetStatus(TokenStreamActive)
//	tsv.Paint(buf)

// TokenStreamStatus describes streaming state.
type TokenStreamStatus int

const (
	TokenStreamIdle TokenStreamStatus = iota
	TokenStreamActive
	TokenStreamDone
	TokenStreamError
)

// TokenStreamStyle holds styling.
type TokenStreamStyle struct {
	Text     buffer.Style
	Cursor   buffer.Style // blinking cursor at end of text
	Progress buffer.Style
	Status   [4]buffer.Style // [idle, active, done, error]
	Border   buffer.Style
}

// DefaultTokenStreamStyle returns defaults.
func DefaultTokenStreamStyle() TokenStreamStyle {
	text := buffer.Style{Fg: buffer.RGB(226, 232, 240)}
	cursor := buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Bold}
	prog := buffer.Style{Fg: buffer.RGB(34, 197, 94)}
	idle := buffer.Style{Fg: buffer.RGB(100, 116, 139)}
	active := buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Bold}
	done := buffer.Style{Fg: buffer.RGB(34, 197, 94)}
	errS := buffer.Style{Fg: buffer.RGB(239, 68, 68)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return TokenStreamStyle{Text: text, Cursor: cursor, Progress: prog, Status: [4]buffer.Style{idle, active, done, errS}, Border: border}
}

// tsStatusIcon returns icon for status.
func tsStatusIcon(s TokenStreamStatus) rune {
	switch s {
	case TokenStreamActive: return '▶'
	case TokenStreamDone: return '✓'
	case TokenStreamError: return '✗'
	default: return '○'
	}
}

// TokenStreamVisualizer renders real-time token streaming.
type TokenStreamVisualizer struct {
	BaseComponent
	mu sync.Mutex

	text     string
	received int
	total    int
	status   TokenStreamStatus
	cursorOn bool
	style    TokenStreamStyle
	// cached
	progStr string
}

// NewTokenStreamVisualizer creates a TokenStreamVisualizer.
func NewTokenStreamVisualizer() *TokenStreamVisualizer {
	tsv := &TokenStreamVisualizer{style: DefaultTokenStreamStyle()}
	tsv.SetID(GenerateID("tokenstream"))
	tsv.progStr = "0/0"
	return tsv
}

// SetText sets the streaming text.
func (tsv *TokenStreamVisualizer) SetText(t string) *TokenStreamVisualizer {
	tsv.mu.Lock()
	tsv.text = t
	tsv.mu.Unlock()
	return tsv
}

// Text returns current text.
func (tsv *TokenStreamVisualizer) Text() string {
	tsv.mu.Lock()
	defer tsv.mu.Unlock()
	return tsv.text
}

// SetReceived sets received token count (caches progress string).
func (tsv *TokenStreamVisualizer) SetReceived(n int) *TokenStreamVisualizer {
	tsv.mu.Lock()
	tsv.received = n
	tsv.progStr = itoa(n) + "/" + itoa(tsv.total)
	tsv.mu.Unlock()
	return tsv
}

// SetTotal sets total expected tokens.
func (tsv *TokenStreamVisualizer) SetTotal(n int) *TokenStreamVisualizer {
	tsv.mu.Lock()
	tsv.total = n
	tsv.progStr = itoa(tsv.received) + "/" + itoa(n)
	tsv.mu.Unlock()
	return tsv
}

// SetStatus sets streaming status.
func (tsv *TokenStreamVisualizer) SetStatus(s TokenStreamStatus) *TokenStreamVisualizer {
	tsv.mu.Lock()
	tsv.status = s
	tsv.mu.Unlock()
	return tsv
}

// Status returns current status.
func (tsv *TokenStreamVisualizer) Status() TokenStreamStatus {
	tsv.mu.Lock()
	defer tsv.mu.Unlock()
	return tsv.status
}

// SetCursor toggles cursor visibility.
func (tsv *TokenStreamVisualizer) SetCursor(on bool) *TokenStreamVisualizer {
	tsv.mu.Lock()
	tsv.cursorOn = on
	tsv.mu.Unlock()
	return tsv
}

// SetStyle sets custom style.
func (tsv *TokenStreamVisualizer) SetStyle(s TokenStreamStyle) *TokenStreamVisualizer {
	tsv.mu.Lock()
	tsv.style = s
	tsv.mu.Unlock()
	return tsv
}

// Measure returns preferred size.
func (tsv *TokenStreamVisualizer) Measure(cs Constraints) Size {
	w := 50
	h := 5
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the token stream visualizer into the buffer.
func (tsv *TokenStreamVisualizer) Paint(buf *buffer.Buffer) {
	tsv.mu.Lock()
	defer tsv.mu.Unlock()

	b := tsv.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 20 { w = 50 }
	if h < 3 { h = 5 }

	bs := tsv.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	// Status bar on row 1
	statusIdx := int(tsv.status)
	if statusIdx < 0 || statusIdx > 3 { statusIdx = 0 }
	statusStyle := tsv.style.Status[statusIdx]
	progStyle := tsv.style.Progress

	col := x + 1
	if col >= buf.Width { return }
	buf.SetCell(col, y+1, buffer.Cell{Rune: tsStatusIcon(tsv.status), Fg: statusStyle.Fg, Bg: statusStyle.Bg, Flags: statusStyle.Flags, Width: 1})
	col++
	if col >= buf.Width { return }
	buf.SetCell(col, y+1, buffer.Cell{Rune: ' ', Fg: statusStyle.Fg, Bg: statusStyle.Bg, Flags: statusStyle.Flags, Width: 1})
	col++

	// Progress (right-aligned on status row)
	progLen := len(tsv.progStr)
	progStart := x + w - 2 - progLen
	if progStart < col { progStart = col }
	for c := col; c < progStart && c < buf.Width; c++ {
		buf.SetCell(c, y+1, buffer.Cell{Rune: ' ', Fg: progStyle.Fg, Bg: progStyle.Bg, Flags: progStyle.Flags, Width: 1})
	}
	for i, r := range tsv.progStr {
		cx := progStart + i
		if cx >= x+w-1 || cx >= buf.Width { break }
		buf.SetCell(cx, y+1, buffer.Cell{Rune: r, Fg: progStyle.Fg, Bg: progStyle.Bg, Flags: progStyle.Flags, Width: 1})
	}

	// Text on rows 2-3
	textStyle := tsv.style.Text
	cursorStyle := tsv.style.Cursor
	textRow := y + 2
	textCol := x + 1

	for _, r := range tsv.text {
		if r == '\n' {
			textRow++
			textCol = x + 1
			if textRow >= y+h-1 || textRow >= buf.Height { break }
			continue
		}
		if textCol >= x+w-1 {
			textRow++
			textCol = x + 1
			if textRow >= y+h-1 || textRow >= buf.Height { break }
		}
		if textCol < buf.Width && textRow < buf.Height {
			buf.SetCell(textCol, textRow, buffer.Cell{Rune: r, Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
		}
		textCol++
	}

	// Cursor at end of text
	if tsv.cursorOn && tsv.status == TokenStreamActive {
		if textCol < x+w-1 && textCol < buf.Width && textRow < buf.Height {
			buf.SetCell(textCol, textRow, buffer.Cell{Rune: '▋', Fg: cursorStyle.Fg, Bg: cursorStyle.Bg, Flags: cursorStyle.Flags, Width: 1})
		}
	}
}

// Children returns nil.
func (tsv *TokenStreamVisualizer) Children() []Component { return nil }

package component

import (
	"strconv"
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AIStreamRenderer: Unified AI Streaming Response Display ───
//
// AIStreamRenderer is the all-in-one AI response renderer combining:
// - Live markdown rendering (streams in as tokens arrive)
// - Typing cursor animation (blinking block at end of text)
// - Token counter (shows tokens/second and total)
// - Thinking indicator (spinner while AI is processing)
// - Status bar (model name, elapsed time, stop reason)
//
// Usage:
//
//	r := NewAIStreamRenderer()
//	r.SetModel("gpt-4o")
//	r.Start()
//	r.Append("Hello, **world**!\n\nThis is `code`.")
//	r.SetTokens(42, 1500)
//	r.Paint(buf)

// AIStreamStatus represents the current streaming state.
type AIStreamStatus int

const (
	AIStreamIdle     AIStreamStatus = iota
	AIStreamThinking                // spinner visible, no text yet
	AIStreamStreaming               // text flowing in
	AIStreamDone                    // completed
	AIStreamError                   // error occurred
)

// AIStreamRendererStyle holds visual styles.
type AIStreamRendererStyle struct {
	Text      buffer.Style
	Cursor    buffer.Style
	Thinking  buffer.Style
	Status    buffer.Style
	TokenInfo buffer.Style
	Error     buffer.Style
}

// DefaultAIStreamRendererStyle returns sensible defaults.
func DefaultAIStreamRendererStyle() AIStreamRendererStyle {
	return AIStreamRendererStyle{
		Text:      buffer.Style{Fg: buffer.White},
		Cursor:    buffer.Style{Fg: buffer.RGB(100, 149, 237), Flags: buffer.Bold},
		Thinking:  buffer.Style{Fg: buffer.RGB(255, 175, 64)},
		Status:    buffer.Style{Fg: buffer.RGB(150, 150, 150)},
		TokenInfo: buffer.Style{Fg: buffer.RGB(100, 200, 100)},
		Error:     buffer.Style{Fg: buffer.RGB(220, 80, 80)},
	}
}

// AIStreamRenderer renders a unified AI streaming response.
type AIStreamRenderer struct {
	BaseComponent
	mu          sync.RWMutex
	text        string
	status      AIStreamStatus
	model       string
	tokenCount  int
	tps         float64 // tokens per second
	startTime   time.Time
	elapsed     time.Duration
	stopReason  string
	style       AIStreamRendererStyle
	cursorOn    bool
	showStatus  bool
	showTokens  bool
}

// NewAIStreamRenderer creates a renderer with defaults.
func NewAIStreamRenderer() *AIStreamRenderer {
	r := &AIStreamRenderer{
		status:     AIStreamIdle,
		style:      DefaultAIStreamRendererStyle(),
		showStatus: true,
		showTokens: true,
	}
	r.SetID(GenerateID("aistream"))
	return r
}

// Start begins streaming.
func (r *AIStreamRenderer) Start() *AIStreamRenderer {
	r.mu.Lock()
	r.status = AIStreamThinking
	r.startTime = time.Now()
	r.text = ""
	r.tokenCount = 0
	r.mu.Unlock()
	return r
}

// StartWithModel begins streaming with a model name.
func (r *AIStreamRenderer) StartWithModel(model string) *AIStreamRenderer {
	r.mu.Lock()
	r.model = model
	r.status = AIStreamThinking
	r.startTime = time.Now()
	r.text = ""
	r.tokenCount = 0
	r.mu.Unlock()
	return r
}

// Append adds text to the stream.
func (r *AIStreamRenderer) Append(text string) *AIStreamRenderer {
	r.mu.Lock()
	if r.status == AIStreamThinking {
		r.status = AIStreamStreaming
	}
	r.text += text
	r.elapsed = time.Since(r.startTime)
	r.mu.Unlock()
	return r
}

// SetText replaces the full text.
func (r *AIStreamRenderer) SetText(text string) *AIStreamRenderer {
	r.mu.Lock()
	r.text = text
	r.mu.Unlock()
	return r
}

// Text returns the current text.
func (r *AIStreamRenderer) Text() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.text
}

// SetTokens updates token count and computes TPS.
func (r *AIStreamRenderer) SetTokens(count int, tps float64) *AIStreamRenderer {
	r.mu.Lock()
	r.tokenCount = count
	r.tps = tps
	r.elapsed = time.Since(r.startTime)
	r.mu.Unlock()
	return r
}

// Complete marks streaming as done.
func (r *AIStreamRenderer) Complete(reason string) *AIStreamRenderer {
	r.mu.Lock()
	r.status = AIStreamDone
	r.stopReason = reason
	r.elapsed = time.Since(r.startTime)
	r.mu.Unlock()
	return r
}

// SetError marks streaming as errored.
func (r *AIStreamRenderer) SetError(msg string) *AIStreamRenderer {
	r.mu.Lock()
	r.status = AIStreamError
	r.stopReason = msg
	r.mu.Unlock()
	return r
}

// Status returns the current streaming status.
func (r *AIStreamRenderer) Status() AIStreamStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.status
}

// Model returns the model name.
func (r *AIStreamRenderer) Model() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.model
}

// SetModel sets the model name.
func (r *AIStreamRenderer) SetModel(m string) *AIStreamRenderer {
	r.mu.Lock()
	r.model = m
	r.mu.Unlock()
	return r
}

// TokenCount returns the token count.
func (r *AIStreamRenderer) TokenCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tokenCount
}

// TPS returns tokens per second.
func (r *AIStreamRenderer) TPS() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tps
}

// Elapsed returns the elapsed time.
func (r *AIStreamRenderer) Elapsed() time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.elapsed
}

// SetCursor toggles the cursor blink state.
func (r *AIStreamRenderer) SetCursor(on bool) *AIStreamRenderer {
	r.mu.Lock()
	r.cursorOn = on
	r.mu.Unlock()
	return r
}

// SetShowStatus toggles the status bar.
func (r *AIStreamRenderer) SetShowStatus(show bool) *AIStreamRenderer {
	r.mu.Lock()
	r.showStatus = show
	r.mu.Unlock()
	return r
}

// SetShowTokens toggles the token counter.
func (r *AIStreamRenderer) SetShowTokens(show bool) *AIStreamRenderer {
	r.mu.Lock()
	r.showTokens = show
	r.mu.Unlock()
	return r
}

// SetStyle sets the visual style.
func (r *AIStreamRenderer) SetStyle(s AIStreamRendererStyle) *AIStreamRenderer {
	r.mu.Lock()
	r.style = s
	r.mu.Unlock()
	return r
}

// Measure computes the desired size.
func (r *AIStreamRenderer) Measure(cs Constraints) Size {
	w := 60
	h := 5
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the AI stream.
func (r *AIStreamRenderer) Paint(buf *buffer.Buffer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	b := r.bounds
	if b.W < 4 || b.H < 1 {
		return
	}

	row := b.Y
	textRows := b.H
	if r.showStatus {
		textRows-- // reserve 1 row for status bar
	}
	if textRows < 1 {
		textRows = 1
	}

	// Thinking indicator
	if r.status == AIStreamThinking {
		thinkingText := "● ● ● Thinking..."
		for i, runeVal := range thinkingText {
			x := b.X + i
			if x >= b.X+b.W {
				break
			}
			buf.SetCell(x, row, buffer.Cell{Rune: runeVal, Fg: r.style.Thinking.Fg, Bg: r.style.Thinking.Bg, Flags: r.style.Thinking.Flags, Width: 1})
		}
	} else {
		// Render text
		textRunes := []rune(r.text)
		col := 0
		for _, runeVal := range textRunes {
			if runeVal == '\n' {
				row++
				col = 0
				if row >= b.Y+textRows {
					break
				}
				continue
			}
			if col >= b.W {
				col = 0
				row++
				if row >= b.Y+textRows {
					break
				}
			}
			if row >= b.Y+textRows {
				break
			}
			x := b.X + col
			if x < b.X+b.W {
				buf.SetCell(x, row, buffer.Cell{Rune: runeVal, Fg: r.style.Text.Fg, Bg: r.style.Text.Bg, Flags: r.style.Text.Flags, Width: 1})
			}
			col++
		}

		// Streaming cursor
		if r.status == AIStreamStreaming && r.cursorOn && row < b.Y+textRows {
			cursorX := b.X + col
			if cursorX >= b.X+b.W {
				cursorX = b.X + b.W - 1
			}
			buf.SetCell(cursorX, row, buffer.Cell{Rune: '▋', Fg: r.style.Cursor.Fg, Bg: r.style.Cursor.Bg, Flags: r.style.Cursor.Flags, Width: 1})
		}
	}

	// Status bar
	if r.showStatus {
		statusY := b.Y + b.H - 1
		if statusY >= b.Y && statusY < b.Y+b.H {
			x := b.X

			// Status indicator
			var statusRune rune
			var statusStyle buffer.Style
			switch r.status {
			case AIStreamIdle:
				statusRune = '○'
				statusStyle = r.style.Status
			case AIStreamThinking, AIStreamStreaming:
				statusRune = '●'
				statusStyle = r.style.Thinking
			case AIStreamDone:
				statusRune = '✓'
				statusStyle = r.style.TokenInfo
			case AIStreamError:
				statusRune = '✗'
				statusStyle = r.style.Error
			}
			buf.SetCell(x, statusY, buffer.Cell{Rune: statusRune, Fg: statusStyle.Fg, Bg: statusStyle.Bg, Flags: statusStyle.Flags, Width: 1})
			x++

			// Model name
			if r.model != "" {
				buf.SetCell(x, statusY, buffer.Cell{Rune: ' ', Width: 1})
				x++
				for _, runeVal := range r.model {
					if x >= b.X+b.W {
						break
					}
					buf.SetCell(x, statusY, buffer.Cell{Rune: runeVal, Fg: r.style.Status.Fg, Bg: r.style.Status.Bg, Width: 1})
					x++
				}
			}

			// Token info
			if r.showTokens && r.tokenCount > 0 && x < b.X+b.W {
				tokenStr := "  " + strconv.Itoa(r.tokenCount) + " tok"
				if r.tps > 0 {
					tokenStr += " (" + strconv.FormatFloat(r.tps, 'f', 1, 64) + "/s)"
				}
				for _, runeVal := range tokenStr {
					if x >= b.X+b.W {
						break
					}
					buf.SetCell(x, statusY, buffer.Cell{Rune: runeVal, Fg: r.style.TokenInfo.Fg, Bg: r.style.TokenInfo.Bg, Width: 1})
					x++
				}
			}

			// Error message
			if r.status == AIStreamError && r.stopReason != "" && x < b.X+b.W {
				for _, runeVal := range " ERR: " + r.stopReason {
					if x >= b.X+b.W {
						break
					}
					buf.SetCell(x, statusY, buffer.Cell{Rune: runeVal, Fg: r.style.Error.Fg, Bg: r.style.Error.Bg, Width: 1})
					x++
				}
			}
		}
	}
}

// Children returns nil.
func (r *AIStreamRenderer) Children() []Component { return nil }

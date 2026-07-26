package component

import (
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// StreamingText renders text with a typewriter effect — characters appear
// progressively at a configurable speed. Designed for AI chat responses
// where raw text (non-markdown) streams in and should be displayed with
// a smooth reveal animation.
//
// Usage:
//	st := NewStreamingText()
//	st.SetSpeed(50 * time.Millisecond) // chars per tick
//	st.SetText("Hello, I am an AI assistant.")
//	// Call st.Tick() on a timer to advance the reveal
//
// Thread-safe.
type StreamingText struct {
	BaseComponent
	mu           sync.Mutex
	fullText     string
	visibleCount int    // how many runes are currently visible
	speed        int    // chars to reveal per tick
	cursorOn     bool   // blink state
	showCursor   bool   // whether to show cursor
	completed    bool   // all text revealed
	tickCount    int    // internal counter for cursor blink
}

// NewStreamingText creates a streaming text widget.
func NewStreamingText() *StreamingText {
	return &StreamingText{
		BaseComponent: BaseComponent{id: GenerateID("streamtext")},
		speed:         2,
		showCursor:    true,
		cursorOn:      true,
	}
}

// Text returns the full text being streamed.
func (s *StreamingText) Text() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.fullText
}

// SetText sets the full text to reveal. Resets visible count to 0.
func (s *StreamingText) SetText(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fullText = text
	s.visibleCount = 0
	s.completed = false
}

// AppendText adds text to the end of the stream (for incremental AI responses).
func (s *StreamingText) AppendText(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fullText += text
	s.completed = false
}

// VisibleText returns the currently visible portion of the text.
func (s *StreamingText) VisibleText() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.visibleTextLocked()
}

func (s *StreamingText) visibleTextLocked() string {
	if s.visibleCount >= len([]rune(s.fullText)) {
		return s.fullText
	}
	runes := []rune(s.fullText)
	return string(runes[:s.visibleCount])
}

// Completed returns true when all text has been revealed.
func (s *StreamingText) Completed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.completed
}

// Speed returns the number of chars revealed per tick.
func (s *StreamingText) Speed() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.speed
}

// SetSpeed sets chars revealed per tick (default 2).
func (s *StreamingText) SetSpeed(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n < 1 {
		n = 1
	}
	s.speed = n
}

// ShowCursor returns whether the cursor is displayed.
func (s *StreamingText) ShowCursor() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.showCursor
}

// SetShowCursor toggles cursor display.
func (s *StreamingText) SetShowCursor(b bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.showCursor = b
}

// CursorOn returns the cursor blink state.
func (s *StreamingText) CursorOn() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cursorOn
}

// Tick advances the reveal animation by s.speed chars and toggles cursor.
// Call this on a timer (e.g., every 50ms).
func (s *StreamingText) Tick() {
	s.mu.Lock()
	defer s.mu.Unlock()
	totalRunes := len([]rune(s.fullText))
	if s.visibleCount < totalRunes {
		s.visibleCount += s.speed
		if s.visibleCount >= totalRunes {
			s.visibleCount = totalRunes
			s.completed = true
		}
	}
	s.tickCount++
	if s.tickCount%4 == 0 {
		s.cursorOn = !s.cursorOn
	}
}

// Skip instantly reveals all text.
func (s *StreamingText) Skip() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.visibleCount = len([]rune(s.fullText))
	s.completed = true
}

// Reset sets visible count back to 0.
func (s *StreamingText) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.visibleCount = 0
	s.completed = false
}

// Progress returns the reveal fraction (0.0–1.0).
func (s *StreamingText) Progress() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	total := len([]rune(s.fullText))
	if total == 0 {
		return 0
	}
	return float64(s.visibleCount) / float64(total)
}

// Measure returns the preferred size.
func (s *StreamingText) Measure(cs Constraints) Size {
	s.mu.Lock()
	defer s.mu.Unlock()
	w := len([]rune(s.fullText))
	if w < 1 {
		w = 1
	}
	h := 1
	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.HasHeight() && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return Size{W: w, H: h}
}

// Paint draws the visible text portion plus optional cursor. Zero allocations.
func (s *StreamingText) Paint(buf *buffer.Buffer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bounds := s.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	tt := theme.Get()
	style := buffer.Style{Fg: tt.Fg}
	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	// Draw visible runes
	drawn := 0
	for _, r := range s.fullText {
		if drawn >= s.visibleCount {
			break
		}
		if x >= maxX {
			break
		}
		rw := buffer.RuneWidth(r)
		buf.SetCell(x, y, buffer.Cell{Rune: r, Width: uint8(rw), Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
		if rw == 2 && x+1 < maxX {
			buf.SetCell(x+1, y, buffer.Cell{Rune: 0, Width: 0, Bg: style.Bg})
		}
		x += rw
		drawn++
	}

	// Draw cursor at end of visible text
	if s.showCursor && !s.completed && s.cursorOn && x < maxX {
		buf.SetCell(x, y, buffer.Cell{
			Rune:  '\u2588', // █
			Width: 1,
			Fg:    tt.Accent,
			Bg:    tt.Bg,
			Flags: buffer.Bold,
		})
	}
}

// FormatSpeed returns a human-readable speed string (e.g., "2 chars/tick").
func FormatSpeed(speed int) string {
	if speed <= 0 {
		return "instant"
	}
	// Simple zero-alloc conversion via stack buffer
	var buf [16]byte
	b := buf[:0]
	if speed < 10 {
		b = append(b, byte('0'+speed))
	} else {
		// Multi-digit — simple approach
		b = append(b, []byte(itoaSimple(speed))...)
	}
	b = append(b, " chars/tick"...)
	return string(b)
}

func itoaSimple(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// EstimateDuration estimates how long the reveal will take at the current speed,
// assuming ticks happen every interval.
func (s *StreamingText) EstimateDuration(interval time.Duration) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	totalRunes := len([]rune(s.fullText))
	if totalRunes == 0 || s.speed == 0 {
		return 0
	}
	ticks := totalRunes / s.speed
	if totalRunes%s.speed != 0 {
		ticks++
	}
	return time.Duration(ticks) * interval
}

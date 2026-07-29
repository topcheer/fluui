package component

import (
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── StreamProgressIndicator: AI Streaming Progress Display ───
//
// StreamProgressIndicator renders a live progress display for AI response
// streaming, showing tokens received, elapsed time, tokens/sec rate, and
// an animated progress bar. Useful in chat UIs and AI agent dashboards.
//
// Usage:
//
//	sp := NewStreamProgressIndicator()
//	sp.Start()
//	sp.AddTokens(150)
//	sp.SetExpected(500)
//	sp.Paint(buf)

// StreamProgressStyle holds styling for StreamProgressIndicator.
type StreamProgressStyle struct {
	Label    buffer.Style
	Value    buffer.Style
	Bar      [3]buffer.Style // [normal, active, done]
	Border   buffer.Style
}

// DefaultStreamProgressStyle returns sensible defaults.
func DefaultStreamProgressStyle() StreamProgressStyle {
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)}   // slate-400
	value := buffer.Style{Fg: buffer.RGB(226, 232, 240)}    // slate-200
	normal := buffer.Style{Fg: buffer.RGB(51, 65, 85)}      // slate-700
	active := buffer.Style{Fg: buffer.RGB(96, 165, 250)}    // blue-400
	done := buffer.Style{Fg: buffer.RGB(34, 197, 94)}       // green-500
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}     // slate-600
	return StreamProgressStyle{
		Label:  label,
		Value:  value,
		Bar:    [3]buffer.Style{normal, active, done},
		Border: border,
	}
}

// StreamState describes the streaming state.
type StreamState int

const (
	StreamIdle    StreamState = iota
	StreamActive
	StreamDone
	StreamError
)

// StreamProgressIndicator displays live AI streaming progress.
type StreamProgressIndicator struct {
	BaseComponent
	mu sync.Mutex

	tokensRecv  int
	expected    int
	startTime   time.Time
	endTime     time.Time
	state       StreamState
	barWidth    int
	pctStr      [16]byte // pre-allocated buffer for percent string
	pctLen      int
	tokStr      [16]byte // pre-allocated buffer for token count string
	tokLen      int

	style StreamProgressStyle
}

// NewStreamProgressIndicator creates a StreamProgressIndicator with defaults.
func NewStreamProgressIndicator() *StreamProgressIndicator {
	sp := &StreamProgressIndicator{
		state:    StreamIdle,
		barWidth: 25,
		style:    DefaultStreamProgressStyle(),
	}
	sp.SetID(GenerateID("streamprog"))
	return sp
}

// Start begins streaming, recording the start time.
func (sp *StreamProgressIndicator) Start() *StreamProgressIndicator {
	sp.mu.Lock()
	sp.startTime = time.Now()
	sp.state = StreamActive
	sp.tokensRecv = 0
	sp.mu.Unlock()
	return sp
}

// AddTokens adds received tokens to the counter.
func (sp *StreamProgressIndicator) AddTokens(n int) *StreamProgressIndicator {
	sp.mu.Lock()
	sp.tokensRecv += n
	sp.mu.Unlock()
	return sp
}

// SetExpected sets the expected total token count.
func (sp *StreamProgressIndicator) SetExpected(n int) *StreamProgressIndicator {
	sp.mu.Lock()
	sp.expected = n
	sp.mu.Unlock()
	return sp
}

// Complete marks streaming as done.
func (sp *StreamProgressIndicator) Complete() *StreamProgressIndicator {
	sp.mu.Lock()
	sp.endTime = time.Now()
	sp.state = StreamDone
	sp.mu.Unlock()
	return sp
}

// Fail marks streaming as errored.
func (sp *StreamProgressIndicator) Fail() *StreamProgressIndicator {
	sp.mu.Lock()
	sp.endTime = time.Now()
	sp.state = StreamError
	sp.mu.Unlock()
	return sp
}

// TokensReceived returns total tokens received so far.
func (sp *StreamProgressIndicator) TokensReceived() int {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	return sp.tokensRecv
}

// Expected returns the expected token count.
func (sp *StreamProgressIndicator) Expected() int {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	return sp.expected
}

// State returns the current streaming state.
func (sp *StreamProgressIndicator) State() StreamState {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	return sp.state
}

// Elapsed returns the elapsed time since Start.
func (sp *StreamProgressIndicator) Elapsed() time.Duration {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if sp.startTime.IsZero() {
		return 0
	}
	end := sp.endTime
	if end.IsZero() {
		end = time.Now()
	}
	return end.Sub(sp.startTime)
}

// TokensPerSecond returns the token throughput rate.
func (sp *StreamProgressIndicator) TokensPerSecond() float64 {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if sp.startTime.IsZero() {
		return 0
	}
	end := sp.endTime
	if end.IsZero() {
		end = time.Now()
	}
	elapsed := end.Sub(sp.startTime).Seconds()
	if elapsed <= 0 {
		return 0
	}
	return float64(sp.tokensRecv) / elapsed
}

// Percent returns completion percentage (0-100).
func (sp *StreamProgressIndicator) Percent() float64 {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if sp.expected <= 0 {
		return 0
	}
	pct := float64(sp.tokensRecv) / float64(sp.expected) * 100
	if pct > 100 {
		pct = 100
	}
	return pct
}

// SetBarWidth sets the progress bar width in characters.
func (sp *StreamProgressIndicator) SetBarWidth(w int) *StreamProgressIndicator {
	sp.mu.Lock()
	sp.barWidth = w
	sp.mu.Unlock()
	return sp
}

// SetStyle sets the custom style.
func (sp *StreamProgressIndicator) SetStyle(s StreamProgressStyle) *StreamProgressIndicator {
	sp.mu.Lock()
	sp.style = s
	sp.mu.Unlock()
	return sp
}

// Measure returns the preferred size.
func (sp *StreamProgressIndicator) Measure(cs Constraints) Size {
	w := 50
	h := 4 // border + bar + stats
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// streamStateBarIndex returns 0=normal, 1=active, 2=done for bar color.
func streamStateBarIndex(s StreamState) int {
	switch s {
	case StreamActive:
		return 1
	case StreamDone:
		return 2
	default:
		return 0
	}
}

// Paint renders the streaming progress indicator into the buffer.
func (sp *StreamProgressIndicator) Paint(buf *buffer.Buffer) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	b := sp.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 30 {
		w = 50
	}
	if h < 4 {
		h = 4
	}

	// Draw border
	bs := sp.style.Border
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

	// Progress bar on row 1
	barIdx := streamStateBarIndex(sp.state)
	barStyle := sp.style.Bar[barIdx]
	bgBarStyle := sp.style.Bar[0]

	filled := 0
	if sp.expected > 0 {
		filled = sp.tokensRecv * sp.barWidth / sp.expected
		if filled > sp.barWidth {
			filled = sp.barWidth
		}
	}
	if sp.state == StreamDone {
		filled = sp.barWidth
	}

	col := x + 1
	for i := 0; i < sp.barWidth; i++ {
		if col >= x+w-1 || col >= buf.Width {
			break
		}
		var ch rune
		var style buffer.Style
		if i < filled {
			ch = '█'
			style = barStyle
		} else {
			ch = '░'
			style = bgBarStyle
		}
		buf.SetCell(col, y+1, buffer.Cell{Rune: ch, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
		col++
	}

	// Percent text after bar
	pct := 0.0
	if sp.expected > 0 {
		pct = float64(sp.tokensRecv) / float64(sp.expected) * 100
		if pct > 100 {
			pct = 100
		}
	}
	if sp.state == StreamDone {
		pct = 100
	}

	// Draw integer percentage — use cached byte buffer (zero alloc)
	sp.pctLen = 0
	sp.pctStr[0] = ' '
	sp.pctLen = 1
	pctDigits := itoa(int(pct))
	for i := 0; i < len(pctDigits); i++ {
		sp.pctStr[sp.pctLen] = pctDigits[i]
		sp.pctLen++
	}
	sp.pctStr[sp.pctLen] = '%'
	sp.pctLen++
	pctStyle := sp.style.Value
	for i := 0; i < sp.pctLen; i++ {
		if col >= x+w-1 || col >= buf.Width {
			break
		}
		buf.SetCell(col, y+1, buffer.Cell{Rune: rune(sp.pctStr[i]), Fg: pctStyle.Fg, Bg: pctStyle.Bg, Flags: pctStyle.Flags, Width: 1})
		col++
	}

	// Stats on row 2: "N tokens · Xms · Y tok/s"
	labelStyle := sp.style.Label
	valueStyle := sp.style.Value

	// Tokens — use cached byte buffer (zero alloc)
	col = x + 1
	sp.tokLen = 0
	tokDigits := itoa(sp.tokensRecv)
	for i := 0; i < len(tokDigits); i++ {
		sp.tokStr[sp.tokLen] = tokDigits[i]
		sp.tokLen++
	}
	sp.tokStr[sp.tokLen] = ' '
	sp.tokLen++
	sp.tokStr[sp.tokLen] = 't'
	sp.tokLen++
	sp.tokStr[sp.tokLen] = 'o'
	sp.tokLen++
	sp.tokStr[sp.tokLen] = 'k'
	sp.tokLen++
	for i := 0; i < sp.tokLen; i++ {
		if col >= x+w-1 || col >= buf.Width {
			break
		}
		buf.SetCell(col, y+2, buffer.Cell{Rune: rune(sp.tokStr[i]), Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}

	// Separator " · "
	if col < x+w-1 && col < buf.Width {
		buf.SetCell(col, y+2, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	if col < x+w-1 && col < buf.Width {
		buf.SetCell(col, y+2, buffer.Cell{Rune: '·', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	if col < x+w-1 && col < buf.Width {
		buf.SetCell(col, y+2, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Elapsed time — write directly using itoa into cached buffer (zero alloc)
	if !sp.startTime.IsZero() {
		end := sp.endTime
		if end.IsZero() {
			end = time.Now()
		}
		elapsed := end.Sub(sp.startTime)
		// Format duration directly into cached buffer without string alloc
		var durBuf [16]byte
		durLen := formatDurationBytes(elapsed, durBuf[:])
		for i := 0; i < durLen; i++ {
			if col >= x+w-1 || col >= buf.Width {
				break
			}
			buf.SetCell(col, y+2, buffer.Cell{Rune: rune(durBuf[i]), Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
			col++
		}
	}
}

// formatDurationBytes formats a duration into the provided byte slice without allocation.
// Returns the number of bytes written.
func formatDurationBytes(d time.Duration, buf []byte) int {
	if d < time.Microsecond {
		n := writeItoaBytes(buf, int(d.Nanoseconds()))
		buf[n] = 'n'
		buf[n+1] = 's'
		return n + 2
	}
	if d < time.Millisecond {
		n := writeItoaBytes(buf, int(d.Microseconds()))
		buf[n] = 'u'
		buf[n+1] = 's'
		return n + 2
	}
	if d < time.Second {
		n := writeItoaBytes(buf, int(d.Milliseconds()))
		buf[n] = 'm'
		buf[n+1] = 's'
		return n + 2
	}
	n := writeItoaBytes(buf, int(d.Seconds()))
	buf[n] = 's'
	return n + 1
}

// writeItoaBytes writes an integer into buf and returns bytes written.
func writeItoaBytes(buf []byte, n int) int {
	if n == 0 {
		buf[0] = '0'
		return 1
	}
	if n < 0 {
		buf[0] = '-'
		n = -n
		return 1 + writeItoaBytes(buf[1:], n)
	}
	// Write digits in reverse
	var temp [12]byte
	len := 0
	for n > 0 {
		temp[len] = byte('0' + n%10)
		n /= 10
		len++
	}
	for i := 0; i < len; i++ {
		buf[i] = temp[len-1-i]
	}
	return len
}

// Children returns nil.
func (sp *StreamProgressIndicator) Children() []Component { return nil }

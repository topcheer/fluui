package component

import (
	"strconv"
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ThinkingTrace: AI Reasoning/Chain-of-Thought Display ───
//
// ThinkingTrace renders the AI's internal reasoning process as collapsible,
// dimmed text with a spinner while thinking, then a summary when done.
// Common in Claude/o1-style models that expose thinking tokens.
//
// Usage:
//
//	tt := NewThinkingTrace()
//	tt.Start()
//	tt.Append("Let me analyze the problem step by step...")
//	tt.Append("\nFirst, I'll check the input format.")
//	tt.Complete()
//	tt.Paint(buf)

// ThinkingState represents the thinking phase.
type ThinkingState int

const (
	ThinkingIdle     ThinkingState = iota
	ThinkingActive                   // spinner + streaming text
	ThinkingDone                     // collapsed summary
)

// ThinkingTraceStyle holds visual styles.
type ThinkingTraceStyle struct {
	Label    buffer.Style // "Thinking..." header
	Text     buffer.Style // dimmed reasoning text
	Duration buffer.Style // "(2.3s)" elapsed
	Spinner  buffer.Style // animated spinner
	Done     buffer.Style // "✓ Thought for 2.3s"
}

// DefaultThinkingTraceStyle returns sensible defaults.
func DefaultThinkingTraceStyle() ThinkingTraceStyle {
	return ThinkingTraceStyle{
		Label:    buffer.Style{Fg: buffer.RGB(255, 175, 64), Flags: buffer.Bold},
		Text:     buffer.Style{Fg: buffer.RGB(130, 130, 130)},  // dim gray
		Duration: buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Spinner:  buffer.Style{Fg: buffer.RGB(255, 175, 64)},
		Done:     buffer.Style{Fg: buffer.RGB(16, 163, 127)},  // green
	}
}

// spinnerFrames for thinking animation.
var thinkingSpinnerFrames = [...]rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

// ThinkingTrace renders AI reasoning/chain-of-thought.
type ThinkingTrace struct {
	BaseComponent
	mu        sync.RWMutex
	text      string
	state     ThinkingState
	startTime time.Time
	elapsed   time.Duration
	spinFrame int
	collapsed bool
	style     ThinkingTraceStyle
}

// NewThinkingTrace creates a thinking trace display.
func NewThinkingTrace() *ThinkingTrace {
	tt := &ThinkingTrace{
		state: ThinkingIdle,
		style: DefaultThinkingTraceStyle(),
	}
	tt.SetID(GenerateID("thinking"))
	return tt
}

// Start begins the thinking phase.
func (tt *ThinkingTrace) Start() *ThinkingTrace {
	tt.mu.Lock()
	tt.state = ThinkingActive
	tt.startTime = time.Now()
	tt.text = ""
	tt.collapsed = false
	tt.mu.Unlock()
	return tt
}

// Append adds reasoning text.
func (tt *ThinkingTrace) Append(text string) *ThinkingTrace {
	tt.mu.Lock()
	tt.text += text
	tt.elapsed = time.Since(tt.startTime)
	tt.mu.Unlock()
	return tt
}

// Complete finishes thinking and collapses.
func (tt *ThinkingTrace) Complete() *ThinkingTrace {
	tt.mu.Lock()
	tt.state = ThinkingDone
	tt.elapsed = time.Since(tt.startTime)
	tt.collapsed = true
	tt.mu.Unlock()
	return tt
}

// SetCollapsed toggles the collapsed state (when done).
func (tt *ThinkingTrace) SetCollapsed(c bool) *ThinkingTrace {
	tt.mu.Lock()
	tt.collapsed = c
	tt.mu.Unlock()
	return tt
}

// IsCollapsed returns the collapsed state.
func (tt *ThinkingTrace) IsCollapsed() bool {
	tt.mu.RLock()
	defer tt.mu.RUnlock()
	return tt.collapsed
}

// State returns the current thinking state.
func (tt *ThinkingTrace) State() ThinkingState {
	tt.mu.RLock()
	defer tt.mu.RUnlock()
	return tt.state
}

// Text returns the reasoning text.
func (tt *ThinkingTrace) Text() string {
	tt.mu.RLock()
	defer tt.mu.RUnlock()
	return tt.text
}

// Elapsed returns the thinking duration.
func (tt *ThinkingTrace) Elapsed() time.Duration {
	tt.mu.RLock()
	defer tt.mu.RUnlock()
	return tt.elapsed
}

// TickSpinner advances the spinner frame (call from animation loop).
func (tt *ThinkingTrace) TickSpinner() *ThinkingTrace {
	tt.mu.Lock()
	tt.spinFrame = (tt.spinFrame + 1) % len(thinkingSpinnerFrames)
	tt.mu.Unlock()
	return tt
}

// SetStyle sets the visual style.
func (tt *ThinkingTrace) SetStyle(s ThinkingTraceStyle) *ThinkingTrace {
	tt.mu.Lock()
	tt.style = s
	tt.mu.Unlock()
	return tt
}

// Measure computes the desired size.
func (tt *ThinkingTrace) Measure(cs Constraints) Size {
	tt.mu.RLock()
	defer tt.mu.RUnlock()
	w := 60
	h := 2 // header + at least 1 line
	if tt.state == ThinkingActive || (tt.state == ThinkingDone && !tt.collapsed) {
		lines := 1
		for i := 0; i < len(tt.text); i++ {
			if tt.text[i] == '\n' {
				lines++
			}
		}
		h += lines
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// formatDuration formats elapsed time compactly.
func formatThinkingDuration(d time.Duration) string {
	sec := d.Seconds()
	if sec < 1 {
		return strconv.FormatInt(d.Milliseconds(), 10) + "ms"
	}
	return strconv.FormatFloat(sec, 'f', 1, 64) + "s"
}

// Paint renders the thinking trace.
func (tt *ThinkingTrace) Paint(buf *buffer.Buffer) {
	tt.mu.Lock()
	defer tt.mu.Unlock()

	b := tt.bounds
	if b.W < 4 || b.H < 1 {
		return
	}

	x := b.X
	y := b.Y

	switch tt.state {
	case ThinkingIdle:
		// Nothing to render
		return

	case ThinkingActive:
		// Spinner + "Thinking..."
		spinner := thinkingSpinnerFrames[tt.spinFrame%len(thinkingSpinnerFrames)]
		buf.SetCell(x, y, buffer.Cell{Rune: spinner, Fg: tt.style.Spinner.Fg, Bg: tt.style.Spinner.Bg, Flags: tt.style.Spinner.Flags, Width: 1})
		x++
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1})
		x++
		label := "Thinking"
		for _, r := range label {
			if x >= b.X+b.W {
				break
			}
			buf.SetCell(x, y, buffer.Cell{Rune: r, Fg: tt.style.Label.Fg, Bg: tt.style.Label.Bg, Flags: tt.style.Label.Flags, Width: 1})
			x++
		}
		// Duration
		if tt.elapsed > 0 && x < b.X+b.W {
			durStr := " (" + formatThinkingDuration(tt.elapsed) + ")"
			for _, r := range durStr {
				if x >= b.X+b.W {
					break
				}
				buf.SetCell(x, y, buffer.Cell{Rune: r, Fg: tt.style.Duration.Fg, Bg: tt.style.Duration.Bg, Width: 1})
				x++
			}
		}

		// Streaming reasoning text (dimmed)
		row := y + 1
		col := 0
		for _, r := range tt.text {
			if row >= b.Y+b.H {
				break
			}
			if r == '\n' {
				row++
				col = 0
				continue
			}
			if col >= b.W {
				col = 0
				row++
			}
			if row < b.Y+b.H && b.X+col < b.X+b.W {
				buf.SetCell(b.X+col, row, buffer.Cell{Rune: r, Fg: tt.style.Text.Fg, Bg: tt.style.Text.Bg, Width: 1})
			}
			col++
		}

	case ThinkingDone:
		// "✓ Thought for 2.3s" (clickable to expand)
		buf.SetCell(x, y, buffer.Cell{Rune: '✓', Fg: tt.style.Done.Fg, Bg: tt.style.Done.Bg, Flags: tt.style.Done.Flags, Width: 1})
		x++
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1})
		x++

		summary := "Thought for " + formatThinkingDuration(tt.elapsed)
		for _, r := range summary {
			if x >= b.X+b.W {
				break
			}
			buf.SetCell(x, y, buffer.Cell{Rune: r, Fg: tt.style.Done.Fg, Bg: tt.style.Done.Bg, Flags: tt.style.Done.Flags, Width: 1})
			x++
		}

		// Expand/collapse indicator
		if x < b.X+b.W {
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1})
			x++
			indicator := "▸"
			if !tt.collapsed {
				indicator = "▾"
			}
			if x < b.X+b.W {
				buf.SetCell(x, y, buffer.Cell{Rune: []rune(indicator)[0], Fg: tt.style.Duration.Fg, Bg: tt.style.Duration.Bg, Width: 1})
			}
		}

		// Expanded text
		if !tt.collapsed {
			row := y + 1
			col := 0
			for _, r := range tt.text {
				if row >= b.Y+b.H {
					break
				}
				if r == '\n' {
					row++
					col = 0
					continue
				}
				if col >= b.W {
					col = 0
					row++
				}
				if row < b.Y+b.H && b.X+col < b.X+b.W {
					buf.SetCell(b.X+col, row, buffer.Cell{Rune: r, Fg: tt.style.Text.Fg, Bg: tt.style.Text.Bg, Width: 1})
				}
				col++
			}
		}
	}
}

// Children returns nil.
func (tt *ThinkingTrace) Children() []Component { return nil }

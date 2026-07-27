package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── SentimentBar: AI Response Sentiment Indicator ───
//
// SentimentBar displays a compact horizontal bar showing AI response sentiment
// (positive/neutral/negative) with a confidence percentage. Useful in AI chat
// UIs to give users quick visual feedback on response tone.
//
// Usage:
//
//	sb := NewSentimentBar(0.72) // 72% positive
//	sb.SetConfidence(0.85)
//	sb.Paint(buf) // renders "▓▓▓▓▓▓▓░░░ 72% positive"

// Sentiment represents the emotional tone of an AI response.
type Sentiment int

const (
	SentimentNegative Sentiment = -1
	SentimentNeutral  Sentiment = 0
	SentimentPositive Sentiment = 1
)

// SentimentBarStyle holds visual styles.
type SentimentBarStyle struct {
	Positive    buffer.Style
	Neutral     buffer.Style
	Negative    buffer.Style
	Confidence  buffer.Style
	Background  buffer.Style
}

// DefaultSentimentBarStyle returns sensible defaults.
func DefaultSentimentBarStyle() SentimentBarStyle {
	return SentimentBarStyle{
		Positive:   buffer.Style{Fg: buffer.RGB(16, 163, 127)},
		Neutral:    buffer.Style{Fg: buffer.RGB(150, 150, 150)},
		Negative:   buffer.Style{Fg: buffer.RGB(220, 80, 80)},
		Confidence: buffer.Style{Fg: buffer.White, Flags: buffer.Bold},
		Background: buffer.Style{Fg: buffer.RGB(40, 40, 40)},
	}
}

// SentimentBar displays AI response sentiment with a confidence meter.
type SentimentBar struct {
	BaseComponent
	mu          sync.RWMutex
	value       float64   // -1.0 (very negative) to 1.0 (very positive)
	confidence  float64   // 0.0 to 1.0
	style       SentimentBarStyle
	label       string
	showPct     bool
}

// NewSentimentBar creates a sentiment bar with the given sentiment value.
func NewSentimentBar(value float64) *SentimentBar {
	sb := &SentimentBar{
		value:      clampFloat(value, -1, 1),
		confidence: 0.5,
		style:      DefaultSentimentBarStyle(),
		showPct:    true,
	}
	sb.SetID(GenerateID("sentimentbar"))
	return sb
}

// Value returns the sentiment value (-1 to 1).
func (sb *SentimentBar) Value() float64 {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.value
}

// SetValue sets the sentiment value, clamped to [-1, 1].
func (sb *SentimentBar) SetValue(v float64) *SentimentBar {
	sb.mu.Lock()
	sb.value = clampFloat(v, -1, 1)
	sb.mu.Unlock()
	return sb
}

// Confidence returns the confidence level (0 to 1).
func (sb *SentimentBar) Confidence() float64 {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.confidence
}

// SetConfidence sets the confidence level, clamped to [0, 1].
func (sb *SentimentBar) SetConfidence(c float64) *SentimentBar {
	sb.mu.Lock()
	sb.confidence = clampFloat(c, 0, 1)
	sb.mu.Unlock()
	return sb
}

// Sentiment returns the discrete sentiment category.
func (sb *SentimentBar) Sentiment() Sentiment {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	if sb.value > 0.15 {
		return SentimentPositive
	}
	if sb.value < -0.15 {
		return SentimentNegative
	}
	return SentimentNeutral
}

// Label returns the sentiment label text.
func (sb *SentimentBar) Label() string {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	if sb.label != "" {
		return sb.label
	}
	switch sb.Sentiment() {
	case SentimentPositive:
		return "positive"
	case SentimentNegative:
		return "negative"
	default:
		return "neutral"
	}
}

// SetLabel sets a custom label.
func (sb *SentimentBar) SetLabel(l string) *SentimentBar {
	sb.mu.Lock()
	sb.label = l
	sb.mu.Unlock()
	return sb
}

// SetStyle overrides the default style.
func (sb *SentimentBar) SetStyle(s SentimentBarStyle) *SentimentBar {
	sb.mu.Lock()
	sb.style = s
	sb.mu.Unlock()
	return sb
}

// Style returns the current style.
func (sb *SentimentBar) Style() SentimentBarStyle {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.style
}

// SetShowPct toggles percentage display.
func (sb *SentimentBar) SetShowPct(show bool) *SentimentBar {
	sb.mu.Lock()
	sb.showPct = show
	sb.mu.Unlock()
	return sb
}

// ShowPct returns whether percentage is shown.
func (sb *SentimentBar) ShowPct() bool {
	sb.mu.RLock()
	defer sb.mu.RUnlock()
	return sb.showPct
}

// styleForSentiment returns the style for the current sentiment (caller holds lock).
func (sb *SentimentBar) styleForSentimentLocked() buffer.Style {
	if sb.value > 0.15 {
		return sb.style.Positive
	}
	if sb.value < -0.15 {
		return sb.style.Negative
	}
	return sb.style.Neutral
}

// Measure computes the desired size.
func (sb *SentimentBar) Measure(cs Constraints) Size {
	w := 20
	if sb.showPct {
		w += 12 // " 72% positive"
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the sentiment bar.
func (sb *SentimentBar) Paint(buf *buffer.Buffer) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	b := sb.bounds
	if b.W < 4 || b.H < 1 {
		return
	}

	sentStyle := sb.styleForSentimentLocked()
	bgStyle := sb.style.Background

	// Bar width: 10 cells minimum, or half the bounds
	barW := 10
	if b.W/2 < barW {
		barW = b.W / 2
	}
	if barW < 4 {
		barW = 4
	}

	// Fill ratio: |value| maps to bar fill (0=empty, 1=full)
	fillRatio := absFloat64(sb.value)
	filled := int(float64(barW) * fillRatio)
	if filled > barW {
		filled = barW
	}

	// Draw bar
	x := b.X
	for i := 0; i < barW; i++ {
		if i < filled {
			buf.SetCell(x+i, b.Y, buffer.Cell{Rune: '▓', Fg: sentStyle.Fg, Bg: bgStyle.Fg, Flags: sentStyle.Flags, Width: 1})
		} else {
			buf.SetCell(x+i, b.Y, buffer.Cell{Rune: '░', Fg: bgStyle.Fg, Bg: bgStyle.Fg, Width: 1})
		}
	}
	x += barW + 1

	// Percentage + label
	if sb.showPct && x < b.X+b.W {
		pct := int(fillRatio * 100)
		var text string
		switch {
		case sb.value > 0.15:
			text = itoaPct(pct, "positive")
		case sb.value < -0.15:
			text = itoaPct(pct, "negative")
		default:
			text = itoaPct(pct, "neutral")
		}
		for _, r := range text {
			if x >= b.X+b.W {
				break
			}
			buf.SetCell(x, b.Y, buffer.Cell{Rune: r, Fg: sb.style.Confidence.Fg, Bg: bgStyle.Fg, Flags: sb.style.Confidence.Flags, Width: 1})
			x++
		}
	}
}

// itoaPct formats "NN% label" without fmt.Sprintf.
func itoaPct(pct int, label string) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return sbItoa(pct) + "% " + label
}

// sbItoa converts int to string (stack-friendly, local to avoid conflicts).
func sbItoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [4]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// absFloat64 returns absolute value.
func absFloat64(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// Children returns nil.
func (sb *SentimentBar) Children() []Component { return nil }

package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AIConfidenceBar: AI Model Confidence Score Display ───
//
// AIConfidenceBar renders a horizontal bar showing an AI model's confidence
// score (0-100%). Colors transition smoothly: red (<40%), yellow (40-70%),
// green (>70%). Common in AI prediction interfaces and classification tools.
//
// Usage:
//
//	cb := NewAIConfidenceBar()
//	cb.SetLabel("Classification")
//	cb.SetConfidence(85.5)
//	cb.Paint(buf)

// AIConfidenceStyle holds styling for AIConfidenceBar.
type AIConfidenceStyle struct {
	Low      buffer.Style  // <40%
	Medium   buffer.Style  // 40-70%
	High     buffer.Style  // >70%
	Label    buffer.Style
	BarBg    buffer.Style
	Border   buffer.Style
}

// DefaultAIConfidenceStyle returns sensible defaults.
func DefaultAIConfidenceStyle() AIConfidenceStyle {
	low := buffer.Style{Fg: buffer.RGB(239, 68, 68)}    // red-500
	med := buffer.Style{Fg: buffer.RGB(234, 179, 8)}    // yellow-500
	high := buffer.Style{Fg: buffer.RGB(34, 197, 94)}   // green-500
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)} // slate-400
	barBg := buffer.Style{Fg: buffer.RGB(51, 65, 85)}    // slate-700
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}  // slate-600
	return AIConfidenceStyle{Low: low, Medium: med, High: high, Label: label, BarBg: barBg, Border: border}
}

// AIConfidenceBar displays an AI confidence score as a colored bar.
type AIConfidenceBar struct {
	BaseComponent
	mu sync.Mutex

	confidence float64
	label      string
	barWidth   int

	style AIConfidenceStyle

	// cached display strings
	confStr [16]byte
	confLen int
}

// NewAIConfidenceBar creates an AIConfidenceBar with defaults.
func NewAIConfidenceBar() *AIConfidenceBar {
	cb := &AIConfidenceBar{
		barWidth: 20,
		label:    "Confidence",
		style:    DefaultAIConfidenceStyle(),
	}
	cb.SetID(GenerateID("confidence"))
	return cb
}

// SetConfidence sets the confidence percentage (0-100).
func (cb *AIConfidenceBar) SetConfidence(pct float64) *AIConfidenceBar {
	cb.mu.Lock()
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	cb.confidence = pct
	// Pre-format display string into cached buffer
	cb.confLen = 0
	digits := itoa(int(pct))
	for i := 0; i < len(digits); i++ {
		cb.confStr[cb.confLen] = digits[i]
		cb.confLen++
	}
	cb.confStr[cb.confLen] = '%'
	cb.confLen++
	cb.mu.Unlock()
	return cb
}

// Confidence returns the current confidence percentage.
func (cb *AIConfidenceBar) Confidence() float64 {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.confidence
}

// SetLabel sets the display label.
func (cb *AIConfidenceBar) SetLabel(l string) *AIConfidenceBar {
	cb.mu.Lock()
	cb.label = l
	cb.mu.Unlock()
	return cb
}

// Label returns the current label.
func (cb *AIConfidenceBar) Label() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.label
}

// SetBarWidth sets the bar width in characters.
func (cb *AIConfidenceBar) SetBarWidth(w int) *AIConfidenceBar {
	cb.mu.Lock()
	cb.barWidth = w
	cb.mu.Unlock()
	return cb
}

// SetStyle sets the custom style.
func (cb *AIConfidenceBar) SetStyle(s AIConfidenceStyle) *AIConfidenceBar {
	cb.mu.Lock()
	cb.style = s
	cb.mu.Unlock()
	return cb
}

// confidenceLevelLocked returns 0=low, 1=medium, 2=high.
func (cb *AIConfidenceBar) confidenceLevelLocked() int {
	if cb.confidence < 40 {
		return 0
	}
	if cb.confidence < 70 {
		return 1
	}
	return 2
}

// Measure returns the preferred size.
func (cb *AIConfidenceBar) Measure(cs Constraints) Size {
	w := 40
	h := 1
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: h}
}

// Paint renders the confidence bar into the buffer.
func (cb *AIConfidenceBar) Paint(buf *buffer.Buffer) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	b := cb.Bounds()
	x, y := b.X, b.Y

	// Determine color based on confidence level
	level := cb.confidenceLevelLocked()
	var fillStyle buffer.Style
	switch level {
	case 0:
		fillStyle = cb.style.Low
	case 1:
		fillStyle = cb.style.Medium
	default:
		fillStyle = cb.style.High
	}

	// Draw label
	labelStyle := cb.style.Label
	col := x
	for _, r := range cb.label {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	// Separator
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ':', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Draw bar
	filled := int(cb.confidence / 100 * float64(cb.barWidth))
	bgBarStyle := cb.style.BarBg
	for i := 0; i < cb.barWidth; i++ {
		if col >= buf.Width {
			break
		}
		if i < filled {
			buf.SetCell(col, y, buffer.Cell{Rune: '█', Fg: fillStyle.Fg, Bg: fillStyle.Bg, Flags: fillStyle.Flags, Width: 1})
		} else {
			buf.SetCell(col, y, buffer.Cell{Rune: '░', Fg: bgBarStyle.Fg, Bg: bgBarStyle.Bg, Flags: bgBarStyle.Flags, Width: 1})
		}
		col++
	}

	// Draw cached percentage string
	for i := 0; i < cb.confLen; i++ {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: rune(cb.confStr[i]), Fg: fillStyle.Fg, Bg: fillStyle.Bg, Flags: fillStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (cb *AIConfidenceBar) Children() []Component { return nil }

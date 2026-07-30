package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AIConfidenceGauge: AI Output Confidence Needle Gauge ───
//
// AIConfidenceGauge renders a semicircular gauge with a needle pointing
// to the confidence level. Shows percentage and a color-coded zone.
// Useful for displaying AI model certainty in real-time.
//
// Usage:
//
//	cg := NewAIConfidenceGauge()
//	cg.SetValue(85) // 85% confident
//	cg.Paint(buf)

// AIConfidenceGaugeStyle holds styling.
type AIConfidenceGaugeStyle struct {
	High   buffer.Style // >= 70
	Medium buffer.Style // 40-69
	Low    buffer.Style // < 40
	Label  buffer.Style
	Arc    buffer.Style
}

// DefaultAIConfidenceGaugeStyle returns defaults.
func DefaultAIConfidenceGaugeStyle() AIConfidenceGaugeStyle {
	return AIConfidenceGaugeStyle{
		High:   buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Medium: buffer.Style{Fg: buffer.RGB(245, 158, 11), Flags: buffer.Bold},
		Low:    buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Label:  buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Arc:    buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

var confidenceArcChars = [...]rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// AIConfidenceGauge renders a confidence needle gauge.
type AIConfidenceGauge struct {
	BaseComponent
	mu sync.Mutex

	value int // 0-100
	style AIConfidenceGaugeStyle
	// cached
	pctStr    string
	needleIdx int
	curStyle  buffer.Style
}

// NewAIConfidenceGauge creates an AIConfidenceGauge.
func NewAIConfidenceGauge() *AIConfidenceGauge {
	cg := &AIConfidenceGauge{style: DefaultAIConfidenceGaugeStyle()}
	cg.SetID(GenerateID("confgauge"))
	cg.recomputeLocked()
	return cg
}

// SetValue sets the confidence value (0-100).
func (cg *AIConfidenceGauge) SetValue(v int) *AIConfidenceGauge {
	cg.mu.Lock()
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	cg.value = v
	cg.recomputeLocked()
	cg.mu.Unlock()
	return cg
}

func (cg *AIConfidenceGauge) recomputeLocked() {
	cg.pctStr = itoa(cg.value) + "%"
	cg.needleIdx = cg.value * 8 / 100
	if cg.needleIdx > 8 {
		cg.needleIdx = 8
	}

	if cg.value >= 70 {
		cg.curStyle = cg.style.High
	} else if cg.value >= 40 {
		cg.curStyle = cg.style.Medium
	} else {
		cg.curStyle = cg.style.Low
	}
}

// Value returns the current value.
func (cg *AIConfidenceGauge) Value() int {
	cg.mu.Lock()
	defer cg.mu.Unlock()
	return cg.value
}

// SetStyle sets custom style.
func (cg *AIConfidenceGauge) SetStyle(s AIConfidenceGaugeStyle) *AIConfidenceGauge {
	cg.mu.Lock()
	cg.style = s
	cg.recomputeLocked()
	cg.mu.Unlock()
	return cg
}

// Measure returns preferred size.
func (cg *AIConfidenceGauge) Measure(cs Constraints) Size {
	w := 12
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 2}
}

// Paint renders the confidence gauge.
func (cg *AIConfidenceGauge) Paint(buf *buffer.Buffer) {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	b := cg.Bounds()
	x, y := b.X, b.Y

	arcStyle := cg.style.Arc
	valueStyle := cg.curStyle
	labelStyle := cg.style.Label

	// Arc row: 9 segments showing ramp up to needle position
	col := x
	for i := 0; i < 9; i++ {
		if col >= buf.Width {
			break
		}
		var r rune
		var st buffer.Style
		if i <= cg.needleIdx {
			r = confidenceArcChars[i]
			st = valueStyle
		} else {
			r = confidenceArcChars[i]
			st = arcStyle
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		col++
	}

	// Value label
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, r := range cg.pctStr {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (cg *AIConfidenceGauge) Children() []Component { return nil }

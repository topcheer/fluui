package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── HallucinationIndicator: AI Hallucination Risk Meter ───
//
// HallucinationIndicator renders a compact badge showing the estimated
// hallucination risk level for an AI response. Uses a color-coded meter
// from low (green) to high (red) based on confidence and factuality scores.
//
// Usage:
//
//	hi := NewHallucinationIndicator()
//	hi.SetScores(85, 90) // 85% confidence, 90% factuality
//	hi.Paint(buf)

// HallucinationStyle holds styling.
type HallucinationStyle struct {
	Low  buffer.Style
	Med  buffer.Style
	High buffer.Style
	Label buffer.Style
	Value buffer.Style
}

// DefaultHallucinationStyle returns defaults.
func DefaultHallucinationStyle() HallucinationStyle {
	return HallucinationStyle{
		Low:   buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Med:   buffer.Style{Fg: buffer.RGB(245, 158, 11), Flags: buffer.Bold},
		High:  buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Label: buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Value: buffer.Style{Fg: buffer.RGB(226, 232, 240)},
	}
}

// RiskLevel represents hallucination risk classification.
type RiskLevel int

const (
	RiskLow    RiskLevel = 0
	RiskMedium RiskLevel = 1
	RiskHigh   RiskLevel = 2
)

var riskLabels = [...]string{"LOW", "MED", "HIGH"}
var riskIcons = [...]rune{'✓', '⚠', '⚠'}

// HallucinationIndicator renders a hallucination risk meter.
type HallucinationIndicator struct {
	BaseComponent
	mu sync.Mutex

	confidence int // 0-100
	factuality int // 0-100
	style      HallucinationStyle
	// cached
	risk       RiskLevel
	riskStr    string
	scoreStr   string
	riskStyle  buffer.Style
	meterFill  int // 0-10 segments
}

// NewHallucinationIndicator creates a HallucinationIndicator.
func NewHallucinationIndicator() *HallucinationIndicator {
	hi := &HallucinationIndicator{confidence: 50, factuality: 50, style: DefaultHallucinationStyle()}
	hi.SetID(GenerateID("halluc"))
	hi.recomputeLocked()
	return hi
}

// SetScores sets confidence (0-100) and factuality (0-100) scores.
func (hi *HallucinationIndicator) SetScores(confidence, factuality int) *HallucinationIndicator {
	hi.mu.Lock()
	if confidence < 0 { confidence = 0 }
	if confidence > 100 { confidence = 100 }
	if factuality < 0 { factuality = 0 }
	if factuality > 100 { factuality = 100 }
	hi.confidence = confidence
	hi.factuality = factuality
	hi.recomputeLocked()
	hi.mu.Unlock()
	return hi
}

func (hi *HallucinationIndicator) recomputeLocked() {
	// Risk score: inverse of average confidence and factuality
	riskScore := 100 - (hi.confidence + hi.factuality) / 2

	if riskScore < 20 {
		hi.risk = RiskLow
		hi.riskStyle = hi.style.Low
	} else if riskScore < 40 {
		hi.risk = RiskMedium
		hi.riskStyle = hi.style.Med
	} else {
		hi.risk = RiskHigh
		hi.riskStyle = hi.style.High
	}

	hi.riskStr = riskLabels[hi.risk]
	hi.scoreStr = itoa(hi.confidence) + "/" + itoa(hi.factuality)
	hi.meterFill = riskScore / 10
	if hi.meterFill > 10 { hi.meterFill = 10 }
}

// Risk returns the current risk level.
func (hi *HallucinationIndicator) Risk() RiskLevel {
	hi.mu.Lock()
	defer hi.mu.Unlock()
	return hi.risk
}

// SetStyle sets custom style.
func (hi *HallucinationIndicator) SetStyle(s HallucinationStyle) *HallucinationIndicator {
	hi.mu.Lock()
	hi.style = s
	hi.recomputeLocked()
	hi.mu.Unlock()
	return hi
}

// Measure returns preferred size.
func (hi *HallucinationIndicator) Measure(cs Constraints) Size {
	w := 22
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// Paint renders the hallucination indicator.
func (hi *HallucinationIndicator) Paint(buf *buffer.Buffer) {
	hi.mu.Lock()
	defer hi.mu.Unlock()

	b := hi.Bounds()
	x, y := b.X, b.Y

	labelStyle := hi.style.Label
	riskStyle := hi.riskStyle

	col := x

	// Risk icon
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: riskIcons[hi.risk], Fg: riskStyle.Fg, Bg: riskStyle.Bg, Flags: riskStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Risk label
	for _, r := range hi.riskStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: riskStyle.Fg, Bg: riskStyle.Bg, Flags: riskStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Mini meter: filled segments = risk level
	for i := 0; i < hi.meterFill; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '▰', Fg: riskStyle.Fg, Bg: riskStyle.Bg, Flags: riskStyle.Flags, Width: 1})
		col++
	}
	for i := hi.meterFill; i < 10; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '▱', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (hi *HallucinationIndicator) Children() []Component { return nil }

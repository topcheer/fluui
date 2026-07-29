package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ModelParameterBar: AI Model Parameters Display ───
//
// ModelParameterBar renders a compact display showing key AI model
// parameters (temperature, top_p, max_tokens) as labeled values.
// Useful for showing current inference configuration.
//
// Usage:
//
//	mp := NewModelParameterBar()
//	mp.SetParams(0.7, 0.9, 4096)
//	mp.Paint(buf)

// ModelParameterStyle holds styling.
type ModelParameterStyle struct {
	Label buffer.Style
	Value buffer.Style
	Separator buffer.Style
}

// DefaultModelParameterStyle returns defaults.
func DefaultModelParameterStyle() ModelParameterStyle {
	return ModelParameterStyle{
		Label:     buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Value:     buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
		Separator: buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

// ModelParameterBar renders a model parameters display.
type ModelParameterBar struct {
	BaseComponent
	mu sync.Mutex

	temp      int // temperature * 100 (e.g., 70 = 0.7)
	topP      int // top_p * 100 (e.g., 90 = 0.90)
	maxTokens int
	style     ModelParameterStyle
	// cached
	tempStr     string
	topPStr     string
	maxTokenStr string
}

// NewModelParameterBar creates a ModelParameterBar.
func NewModelParameterBar() *ModelParameterBar {
	mp := &ModelParameterBar{temp: 70, topP: 90, maxTokens: 4096, style: DefaultModelParameterStyle()}
	mp.SetID(GenerateID("modelparam"))
	mp.recomputeLocked()
	return mp
}

// SetParams sets temperature (0-200 mapped to 0.0-2.0), top_p (0-100),
// and max_tokens.
func (mp *ModelParameterBar) SetParams(tempX100, topPX100, maxTokens int) *ModelParameterBar {
	mp.mu.Lock()
	if tempX100 < 0 { tempX100 = 0 }
	if tempX100 > 200 { tempX100 = 200 }
	if topPX100 < 0 { topPX100 = 0 }
	if topPX100 > 100 { topPX100 = 100 }
	if maxTokens < 0 { maxTokens = 0 }
	mp.temp = tempX100
	mp.topP = topPX100
	mp.maxTokens = maxTokens
	mp.recomputeLocked()
	mp.mu.Unlock()
	return mp
}

func (mp *ModelParameterBar) recomputeLocked() {
	// Format temperature as 0.X or X.X
	intPart := mp.temp / 100
	decPart := mp.temp % 100
	if intPart == 0 {
		mp.tempStr = "0." + itoa(decPart)
	} else {
		mp.tempStr = itoa(intPart) + "." + formatCents(decPart)
	}

	// Format top_p as 0.XX
	mp.topPStr = "0." + formatCents(mp.topP)

	mp.maxTokenStr = itoa(mp.maxTokens)
}

// Temperature returns temperature * 100.
func (mp *ModelParameterBar) Temperature() int {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	return mp.temp
}

// SetStyle sets custom style.
func (mp *ModelParameterBar) SetStyle(s ModelParameterStyle) *ModelParameterBar {
	mp.mu.Lock()
	mp.style = s
	mp.mu.Unlock()
	return mp
}

// Measure returns preferred size.
func (mp *ModelParameterBar) Measure(cs Constraints) Size {
	w := 34
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// Paint renders the model parameter bar.
func (mp *ModelParameterBar) Paint(buf *buffer.Buffer) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	b := mp.Bounds()
	x, y := b.X, b.Y

	labelStyle := mp.style.Label
	valueStyle := mp.style.Value
	sepStyle := mp.style.Separator

	col := x

	// Temperature
	for _, r := range "temp:" {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, r := range mp.tempStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}
	// Separator
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
		col++
	}

	// top_p
	for _, r := range "top_p:" {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, r := range mp.topPStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
		col++
	}

	// max_tokens
	for _, r := range "max:" {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, r := range mp.maxTokenStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (mp *ModelParameterBar) Children() []Component { return nil }

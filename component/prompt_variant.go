package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── PromptVariant: A/B Prompt Variant Comparison ───
//
// PromptVariant renders a side-by-side comparison of two prompt variants
// showing their relative scores, token counts, and performance metrics.
// Useful for prompt engineering A/B testing.
//
// Usage:
//
//	pv := NewPromptVariant()
//	pv.SetVariant("A", 85, 1200, 1)
//	pv.SetVariant("B", 92, 980, 0)
//	pv.Paint(buf)

// PromptVariantStyle holds styling.
type PromptVariantStyle struct {
	Label   buffer.Style
	Value   buffer.Style
	Winner  buffer.Style
	Loser   buffer.Style
	Divider buffer.Style
}

// DefaultPromptVariantStyle returns defaults.
func DefaultPromptVariantStyle() PromptVariantStyle {
	return PromptVariantStyle{
		Label:   buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Value:   buffer.Style{Fg: buffer.RGB(226, 232, 240)},
		Winner:  buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Loser:   buffer.Style{Fg: buffer.RGB(239, 68, 68)},
		Divider: buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

// VariantData holds data for a single prompt variant.
type VariantData struct {
	Name      string
	Score     int
	Tokens    int
	IsWinner  bool
}

// PromptVariant renders an A/B prompt comparison.
type PromptVariant struct {
	BaseComponent
	mu sync.Mutex

	variantA VariantData
	variantB VariantData
	width    int
	style    PromptVariantStyle
	// cached
	aScoreStr   string
	bScoreStr   string
	aTokenStr   string
	bTokenStr   string
	aLabelStr   string
	bLabelStr   string
	aScoreLabel string
	bScoreLabel string
	aTokenLabel string
	bTokenLabel string
	aScoreW     int
	bScoreW     int
}

// NewPromptVariant creates a PromptVariant.
func NewPromptVariant() *PromptVariant {
	pv := &PromptVariant{
		width:    36,
		style:    DefaultPromptVariantStyle(),
		variantA: VariantData{Name: "A", Score: 0, Tokens: 0},
		variantB: VariantData{Name: "B", Score: 0, Tokens: 0},
	}
	pv.SetID(GenerateID("promptvar"))
	pv.recomputeLocked()
	return pv
}

// SetVariantA sets the first variant data.
func (pv *PromptVariant) SetVariantA(name string, score, tokens int, isWinner bool) *PromptVariant {
	pv.mu.Lock()
	pv.variantA = VariantData{Name: name, Score: score, Tokens: tokens, IsWinner: isWinner}
	pv.recomputeLocked()
	pv.mu.Unlock()
	return pv
}

// SetVariantB sets the second variant data.
func (pv *PromptVariant) SetVariantB(name string, score, tokens int, isWinner bool) *PromptVariant {
	pv.mu.Lock()
	pv.variantB = VariantData{Name: name, Score: score, Tokens: tokens, IsWinner: isWinner}
	pv.recomputeLocked()
	pv.mu.Unlock()
	return pv
}

func (pv *PromptVariant) recomputeLocked() {
	pv.aScoreStr = itoa(pv.variantA.Score)
	pv.bScoreStr = itoa(pv.variantB.Score)
	pv.aTokenStr = itoa(pv.variantA.Tokens)
	pv.bTokenStr = itoa(pv.variantB.Tokens)
	pv.aScoreLabel = "Score: " + pv.aScoreStr
	pv.bScoreLabel = "Score: " + pv.bScoreStr
	pv.aTokenLabel = "Tokens: " + pv.aTokenStr
	pv.bTokenLabel = "Tokens: " + pv.bTokenStr

	// Winner label
	if pv.variantA.IsWinner {
		pv.aLabelStr = pv.variantA.Name + " ✓"
	} else {
		pv.aLabelStr = pv.variantA.Name
	}
	if pv.variantB.IsWinner {
		pv.bLabelStr = pv.variantB.Name + " ✓"
	} else {
		pv.bLabelStr = pv.variantB.Name
	}

	// Bar widths (10 segments each)
	const barMax = 10
	maxScore := pv.variantA.Score
	if pv.variantB.Score > maxScore { maxScore = pv.variantB.Score }
	if maxScore == 0 { maxScore = 1 }
	pv.aScoreW = pv.variantA.Score * barMax / maxScore
	pv.bScoreW = pv.variantB.Score * barMax / maxScore
}

// SetWidth sets the display width.
func (pv *PromptVariant) SetWidth(w int) *PromptVariant {
	pv.mu.Lock()
	if w < 20 { w = 20 }
	pv.width = w
	pv.mu.Unlock()
	return pv
}

// SetStyle sets custom style.
func (pv *PromptVariant) SetStyle(s PromptVariantStyle) *PromptVariant {
	pv.mu.Lock()
	pv.style = s
	pv.mu.Unlock()
	return pv
}

// Measure returns preferred size.
func (pv *PromptVariant) Measure(cs Constraints) Size {
	w := pv.width
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 3}
}

// Paint renders the prompt variant comparison.
func (pv *PromptVariant) Paint(buf *buffer.Buffer) {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	b := pv.Bounds()
	x, y := b.X, b.Y
	w := b.W
	if w < 20 { w = 36 }

	labelStyle := pv.style.Label
	valueStyle := pv.style.Value
	winnerStyle := pv.style.Winner
	loserStyle := pv.style.Loser
	dividerStyle := pv.style.Divider

	halfW := w / 2

	// Divider in the middle
	divX := x + halfW
	for row := 0; row < 3; row++ {
		yy := y + row
		if yy < buf.Height && divX < buf.Width {
			buf.SetCell(divX, yy, buffer.Cell{Rune: '│', Fg: dividerStyle.Fg, Bg: dividerStyle.Bg, Flags: dividerStyle.Flags, Width: 1})
		}
	}

	// Variant A (left side)
	var aStyle buffer.Style
	if pv.variantA.IsWinner {
		aStyle = winnerStyle
	} else {
		aStyle = loserStyle
	}

	// Row 0: label + score bar
	col := x
	for _, r := range pv.aLabelStr {
		if col >= divX { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: aStyle.Fg, Bg: aStyle.Bg, Flags: aStyle.Flags, Width: 1})
		col++
	}
	barStart := col + 1
	for i := 0; i < pv.aScoreW; i++ {
		cx := barStart + i
		if cx >= divX { break }
		buf.SetCell(cx, y, buffer.Cell{Rune: '▰', Fg: aStyle.Fg, Bg: aStyle.Bg, Flags: aStyle.Flags, Width: 1})
	}

	// Row 1: score value
	col = x
	scoreALabel := pv.aScoreLabel
	for _, r := range scoreALabel {
		if col >= divX { break }
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}

	// Row 2: token count
	col = x
	tokenALabel := pv.aTokenLabel
	for _, r := range tokenALabel {
		if col >= divX { break }
		buf.SetCell(col, y+2, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Variant B (right side)
	var bStyle buffer.Style
	if pv.variantB.IsWinner {
		bStyle = winnerStyle
	} else {
		bStyle = loserStyle
	}

	rightStart := divX + 1

	// Row 0: label + score bar
	col = rightStart
	for _, r := range pv.bLabelStr {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: bStyle.Fg, Bg: bStyle.Bg, Flags: bStyle.Flags, Width: 1})
		col++
	}
	barStartB := col + 1
	for i := 0; i < pv.bScoreW; i++ {
		cx := barStartB + i
		if cx >= buf.Width { break }
		buf.SetCell(cx, y, buffer.Cell{Rune: '▰', Fg: bStyle.Fg, Bg: bStyle.Bg, Flags: bStyle.Flags, Width: 1})
	}

	// Row 1: score value
	col = rightStart
	scoreBLabel := pv.bScoreLabel
	for _, r := range scoreBLabel {
		if col >= buf.Width { break }
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}

	// Row 2: token count
	col = rightStart
	tokenBLabel := pv.bTokenLabel
	for _, r := range tokenBLabel {
		if col >= buf.Width { break }
		buf.SetCell(col, y+2, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (pv *PromptVariant) Children() []Component { return nil }

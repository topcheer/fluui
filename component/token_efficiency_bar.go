package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── TokenEfficiencyBar: AI Token Efficiency Display ───
//
// TokenEfficiencyBar renders a horizontal bar showing the ratio of useful
// output tokens to total tokens consumed (input + output + reasoning).
// Higher efficiency means the model produces more useful content per token spent.
//
// Usage:
//
//	bar := NewTokenEfficiencyBar()
//	bar.SetTokens(500, 1200, 300) // 500 useful, 1200 input, 300 reasoning
//	bar.Paint(buf)

// TokenEfficiencyStyle holds styling.
type TokenEfficiencyStyle struct {
	Useful   buffer.Style
	Overhead buffer.Style
	Label    buffer.Style
	Score    buffer.Style
}

// DefaultTokenEfficiencyStyle returns defaults.
func DefaultTokenEfficiencyStyle() TokenEfficiencyStyle {
	return TokenEfficiencyStyle{
		Useful:   buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Overhead: buffer.Style{Fg: buffer.RGB(100, 116, 139)},
		Label:    buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Score:    buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
	}
}

// TokenEfficiencyBar renders an AI token efficiency bar.
type TokenEfficiencyBar struct {
	BaseComponent
	mu sync.Mutex

	useful    int
	input     int
	reasoning int
	width     int
	style     TokenEfficiencyStyle
	// cached display strings
	effPctStr  string
	totalStr   string
	usefulStr  string
	overStr    string
	label1Str  string // pre-computed: "Useful N / Total"
	label2Str  string // pre-computed: "Eff: NN%"
	label3Str  string // pre-computed: "Overhead: N (input + reasoning)"
	barUseful  int // cached bar segments
	barOver    int
}

// NewTokenEfficiencyBar creates a TokenEfficiencyBar.
func NewTokenEfficiencyBar() *TokenEfficiencyBar {
	t := &TokenEfficiencyBar{width: 32, style: DefaultTokenEfficiencyStyle()}
	t.SetID(GenerateID("tok_eff"))
	t.recomputeLocked()
	return t
}

// SetTokens sets useful output, input prompt, and reasoning token counts.
func (t *TokenEfficiencyBar) SetTokens(useful, input, reasoning int) *TokenEfficiencyBar {
	t.mu.Lock()
	if useful < 0 { useful = 0 }
	if input < 0 { input = 0 }
	if reasoning < 0 { reasoning = 0 }
	t.useful = useful
	t.input = input
	t.reasoning = reasoning
	t.recomputeLocked()
	t.mu.Unlock()
	return t
}

func (t *TokenEfficiencyBar) recomputeLocked() {
	total := t.useful + t.input + t.reasoning
	if total == 0 {
		t.effPctStr = "0%"
		t.totalStr = "0"
		t.usefulStr = "0"
		t.overStr = "0"
		t.label1Str = "Useful 0 / 0"
		t.label2Str = "Eff: 0%"
		t.label3Str = "Overhead: 0 (input + reasoning)"
		t.barUseful = 0
		t.barOver = 0
		return
	}
	pct := t.useful * 100 / total
	t.effPctStr = itoa(pct) + "%"
	t.totalStr = itoa(total)
	t.usefulStr = itoa(t.useful)
	t.overStr = itoa(t.input + t.reasoning)
	t.label1Str = "Useful " + t.usefulStr + " / " + t.totalStr
	t.label2Str = "Eff: " + t.effPctStr
	t.label3Str = "Overhead: " + t.overStr + " (input + reasoning)"
	// Pre-compute bar segments for a 20-char bar
	const barTotal = 20
	t.barUseful = t.useful * barTotal / total
	t.barOver = barTotal - t.barUseful
}

// SetWidth sets the bar width.
func (t *TokenEfficiencyBar) SetWidth(w int) *TokenEfficiencyBar {
	t.mu.Lock()
	if w < 20 { w = 20 }
	t.width = w
	t.mu.Unlock()
	return t
}

// EfficiencyPercent returns the efficiency percentage.
func (t *TokenEfficiencyBar) EfficiencyPercent() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	total := t.useful + t.input + t.reasoning
	if total == 0 { return 0 }
	return t.useful * 100 / total
}

// SetStyle sets custom style.
func (t *TokenEfficiencyBar) SetStyle(s TokenEfficiencyStyle) *TokenEfficiencyBar {
	t.mu.Lock()
	t.style = s
	t.mu.Unlock()
	return t
}

// Measure returns preferred size.
func (t *TokenEfficiencyBar) Measure(cs Constraints) Size {
	w := t.width + 24
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 3}
}

// Paint renders the token efficiency bar.
func (t *TokenEfficiencyBar) Paint(buf *buffer.Buffer) {
	t.mu.Lock()
	defer t.mu.Unlock()

	b := t.Bounds()
	x, y := b.X, b.Y
	w := b.W
	if w < 20 { w = 50 }

	usefulStyle := t.style.Useful
	overStyle := t.style.Overhead
	labelStyle := t.style.Label
	scoreStyle := t.style.Score

	// Row 0: bar
	col := x
	for i := 0; i < t.barUseful; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '█', Fg: usefulStyle.Fg, Bg: usefulStyle.Bg, Flags: usefulStyle.Flags, Width: 1})
		col++
	}
	for i := 0; i < t.barOver; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '░', Fg: overStyle.Fg, Bg: overStyle.Bg, Flags: overStyle.Flags, Width: 1})
		col++
	}

	// Row 1: labels
	label1 := t.label1Str
	col = x
	for _, r := range label1 {
		if col >= buf.Width { break }
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Right-aligned efficiency score
	scoreLabel := t.label2Str
	scoreStart := x + w - 1 - len(scoreLabel)
	if scoreStart < col { scoreStart = col }
	for c := col; c < scoreStart && c < buf.Width; c++ {
		buf.SetCell(c, y+1, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
	}
	for i, r := range scoreLabel {
		cx := scoreStart + i
		if cx >= buf.Width { break }
		buf.SetCell(cx, y+1, buffer.Cell{Rune: r, Fg: scoreStyle.Fg, Bg: scoreStyle.Bg, Flags: scoreStyle.Flags, Width: 1})
	}

	// Row 2: overhead detail
	overLabel := t.label3Str
	col = x
	for _, r := range overLabel {
		if col >= buf.Width { break }
		buf.SetCell(col, y+2, buffer.Cell{Rune: r, Fg: overStyle.Fg, Bg: overStyle.Bg, Flags: overStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (t *TokenEfficiencyBar) Children() []Component { return nil }

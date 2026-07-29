package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AIPolishIndicator: AI Response Quality Polish Meter ───
//
// AIPolishIndicator renders a compact quality indicator showing how
// well-polished an AI response is, based on metrics like grammar,
// coherence, and completeness. Uses a 5-star rating with labels.
//
// Usage:
//
//	p := NewAIPolishIndicator()
//	p.SetScores(85, 90, 80) // grammar, coherence, completeness
//	p.Paint(buf)

// AIPolishStyle holds styling.
type AIPolishStyle struct {
	Star  buffer.Style
	Empty buffer.Style
	Label buffer.Style
	Score buffer.Style
}

// DefaultAIPolishStyle returns defaults.
func DefaultAIPolishStyle() AIPolishStyle {
	return AIPolishStyle{
		Star:  buffer.Style{Fg: buffer.RGB(234, 179, 8), Flags: buffer.Bold},
		Empty: buffer.Style{Fg: buffer.RGB(71, 85, 105)},
		Label: buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Score: buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
	}
}

var polishLabels = [...]string{"Draft", "Rough", "Fair", "Good", "Great", "Excellent"}

// AIPolishIndicator renders a response quality meter.
type AIPolishIndicator struct {
	BaseComponent
	mu sync.Mutex

	grammar      int // 0-100
	coherence    int // 0-100
	completeness int // 0-100
	style        AIPolishStyle
	// cached
	stars    int // 0-5
	avgStr   string
	labelStr string
}

// NewAIPolishIndicator creates an AIPolishIndicator.
func NewAIPolishIndicator() *AIPolishIndicator {
	p := &AIPolishIndicator{style: DefaultAIPolishStyle()}
	p.SetID(GenerateID("polish"))
	p.recomputeLocked()
	return p
}

// SetScores sets grammar, coherence, and completeness scores (0-100).
func (p *AIPolishIndicator) SetScores(grammar, coherence, completeness int) *AIPolishIndicator {
	p.mu.Lock()
	if grammar < 0 {
		grammar = 0
	}
	if grammar > 100 {
		grammar = 100
	}
	if coherence < 0 {
		coherence = 0
	}
	if coherence > 100 {
		coherence = 100
	}
	if completeness < 0 {
		completeness = 0
	}
	if completeness > 100 {
		completeness = 100
	}
	p.grammar = grammar
	p.coherence = coherence
	p.completeness = completeness
	p.recomputeLocked()
	p.mu.Unlock()
	return p
}

func (p *AIPolishIndicator) recomputeLocked() {
	avg := (p.grammar + p.coherence + p.completeness) / 3
	p.avgStr = itoa(avg)
	p.stars = avg * 5 / 100
	if p.stars > 5 {
		p.stars = 5
	}

	idx := avg * 5 / 100 // 0-5
	if idx > 5 {
		idx = 5
	}
	p.labelStr = polishLabels[idx]
}

// Stars returns the star rating (0-5).
func (p *AIPolishIndicator) Stars() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.stars
}

// SetStyle sets custom style.
func (p *AIPolishIndicator) SetStyle(s AIPolishStyle) *AIPolishIndicator {
	p.mu.Lock()
	p.style = s
	p.mu.Unlock()
	return p
}

// Measure returns preferred size.
func (p *AIPolishIndicator) Measure(cs Constraints) Size {
	w := 22
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the polish indicator.
func (p *AIPolishIndicator) Paint(buf *buffer.Buffer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	b := p.Bounds()
	x, y := b.X, b.Y

	starStyle := p.style.Star
	emptyStyle := p.style.Empty
	labelStyle := p.style.Label
	scoreStyle := p.style.Score

	col := x

	// Stars
	for i := 0; i < 5; i++ {
		if col >= buf.Width {
			break
		}
		var r rune
		var st buffer.Style
		if i < p.stars {
			r = '★'
			st = starStyle
		} else {
			r = '☆'
			st = emptyStyle
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		col++
	}

	// Space
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Score value
	for _, r := range p.avgStr {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: scoreStyle.Fg, Bg: scoreStyle.Bg, Flags: scoreStyle.Flags, Width: 1})
		col++
	}

	// Space + label
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, r := range p.labelStr {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (p *AIPolishIndicator) Children() []Component { return nil }

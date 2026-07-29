package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AIResponseScore: Multi-Dimensional AI Response Quality Score ───
//
// AIResponseScore renders a compact multi-axis quality score for an AI
// response. Shows 5 dimensions (accuracy, relevance, clarity, completeness,
// helpfulness) as labeled mini-bars on a single line each.
//
// Usage:
//
//	rs := NewAIResponseScore()
//	rs.SetScores(90, 85, 95, 80, 88)
//	rs.Paint(buf)

// AIResponseScoreStyle holds styling.
type AIResponseScoreStyle struct {
	Label   buffer.Style
	Bar     buffer.Style
	Empty   buffer.Style
	Score   buffer.Style
	Average buffer.Style
}

// DefaultAIResponseScoreStyle returns defaults.
func DefaultAIResponseScoreStyle() AIResponseScoreStyle {
	return AIResponseScoreStyle{
		Label:   buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Bar:     buffer.Style{Fg: buffer.RGB(59, 130, 246)},
		Empty:   buffer.Style{Fg: buffer.RGB(51, 65, 85)},
		Score:   buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
		Average: buffer.Style{Fg: buffer.RGB(234, 179, 8), Flags: buffer.Bold},
	}
}

var responseScoreLabels = [5]string{"Acc", "Rel", "Clr", "Cmp", "Hlp"}

// AIResponseScore renders a multi-dimensional quality score card.
type AIResponseScore struct {
	BaseComponent
	mu sync.Mutex

	scores [5]int // accuracy, relevance, clarity, completeness, helpfulness
	style  AIResponseScoreStyle
	// cached
	avgStr   string
	barFills [5]int
}

// NewAIResponseScore creates an AIResponseScore.
func NewAIResponseScore() *AIResponseScore {
	rs := &AIResponseScore{style: DefaultAIResponseScoreStyle()}
	rs.SetID(GenerateID("rscore"))
	rs.recomputeLocked()
	return rs
}

// SetScores sets the 5 dimension scores (0-100 each).
func (rs *AIResponseScore) SetScores(accuracy, relevance, clarity, completeness, helpfulness int) *AIResponseScore {
	rs.mu.Lock()
	rs.scores = [5]int{clampScore(accuracy), clampScore(relevance), clampScore(clarity), clampScore(completeness), clampScore(helpfulness)}
	rs.recomputeLocked()
	rs.mu.Unlock()
	return rs
}

func clampScore(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func (rs *AIResponseScore) recomputeLocked() {
	sum := 0
	for i := 0; i < 5; i++ {
		sum += rs.scores[i]
		rs.barFills[i] = rs.scores[i] / 10
	}
	rs.avgStr = itoa(sum / 5)
}

// Average returns the average score.
func (rs *AIResponseScore) Average() int {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	sum := 0
	for _, s := range rs.scores {
		sum += s
	}
	return sum / 5
}

// SetStyle sets custom style.
func (rs *AIResponseScore) SetStyle(s AIResponseScoreStyle) *AIResponseScore {
	rs.mu.Lock()
	rs.style = s
	rs.mu.Unlock()
	return rs
}

// Measure returns preferred size.
func (rs *AIResponseScore) Measure(cs Constraints) Size {
	w := 24
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 5}
}

// Paint renders the response score card.
func (rs *AIResponseScore) Paint(buf *buffer.Buffer) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	b := rs.Bounds()
	x, y := b.X, b.Y

	labelStyle := rs.style.Label
	barStyle := rs.style.Bar
	emptyStyle := rs.style.Empty
	scoreStyle := rs.style.Score
	_ = scoreStyle
	avgStyle := rs.style.Average

	for dim := 0; dim < 5; dim++ {
		yy := y + dim
		if yy >= buf.Height {
			break
		}
		col := x

		// Label
		for _, r := range responseScoreLabels[dim] {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: ':', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}

		// Mini bar (10 segments)
		for i := 0; i < rs.barFills[dim]; i++ {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: '█', Fg: barStyle.Fg, Bg: barStyle.Bg, Flags: barStyle.Flags, Width: 1})
			col++
		}
		for i := rs.barFills[dim]; i < 10; i++ {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: '░', Fg: emptyStyle.Fg, Bg: emptyStyle.Bg, Flags: emptyStyle.Flags, Width: 1})
			col++
		}

		// Score value on last dim only (to show average)
		if dim == 4 {
			if col < buf.Width {
				buf.SetCell(col, yy, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
				col++
			}
			for _, r := range rs.avgStr {
				if col >= buf.Width {
					break
				}
				buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: avgStyle.Fg, Bg: avgStyle.Bg, Flags: avgStyle.Flags, Width: 1})
				col++
			}
		}
	}
}

// Children returns nil.
func (rs *AIResponseScore) Children() []Component { return nil }

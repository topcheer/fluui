package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AIFactCheck: AI Response Fact Verification Badge ───
//
// AIFactCheck renders a compact fact-check result showing verified,
// disputed, or unverified status with source count. Useful for
// AI responses where factual accuracy needs transparency.
//
// Usage:
//
//	fc := NewAIFactCheck()
//	fc.SetResult(FactVerified, 3) // verified with 3 sources
//	fc.Paint(buf)

type FactCheckResult int

const (
	FactUnverified FactCheckResult = 0
	FactVerified   FactCheckResult = 1
	FactDisputed   FactCheckResult = 2
	FactPartial    FactCheckResult = 3
)

var factCheckIcons = [...]rune{'❓', '✓', '⚠', '🔷'}
var factCheckLabels = [...]string{"Unverified", "Verified", "Disputed", "Partial"}

type AIFactCheckStyle struct {
	Unverified buffer.Style
	Verified   buffer.Style
	Disputed   buffer.Style
	Partial    buffer.Style
	Source     buffer.Style
	Bracket    buffer.Style
}

func DefaultAIFactCheckStyle() AIFactCheckStyle {
	return AIFactCheckStyle{
		Unverified: buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Verified:   buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Disputed:   buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Partial:    buffer.Style{Fg: buffer.RGB(245, 158, 11), Flags: buffer.Bold},
		Source:     buffer.Style{Fg: buffer.RGB(96, 165, 250)},
		Bracket:    buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

// AIFactCheck renders a fact verification badge.
type AIFactCheck struct {
	BaseComponent
	mu sync.Mutex

	result  FactCheckResult
	sources int
	style   AIFactCheckStyle
	// cached
	labelStr string
	curStyle buffer.Style
	srcStr   string
}

func NewAIFactCheck() *AIFactCheck {
	fc := &AIFactCheck{style: DefaultAIFactCheckStyle()}
	fc.SetID(GenerateID("factcheck"))
	fc.recomputeLocked()
	return fc
}

func (fc *AIFactCheck) SetResult(r FactCheckResult, sourceCount int) *AIFactCheck {
	fc.mu.Lock()
	if int(r) < 0 || int(r) >= len(factCheckLabels) {
		r = FactUnverified
	}
	if sourceCount < 0 {
		sourceCount = 0
	}
	fc.result = r
	fc.sources = sourceCount
	fc.recomputeLocked()
	fc.mu.Unlock()
	return fc
}

func (fc *AIFactCheck) recomputeLocked() {
	fc.labelStr = factCheckLabels[fc.result]
	fc.srcStr = itoa(fc.sources) + "src"

	switch fc.result {
	case FactVerified:
		fc.curStyle = fc.style.Verified
	case FactDisputed:
		fc.curStyle = fc.style.Disputed
	case FactPartial:
		fc.curStyle = fc.style.Partial
	default:
		fc.curStyle = fc.style.Unverified
	}
}

func (fc *AIFactCheck) Result() FactCheckResult {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	return fc.result
}

func (fc *AIFactCheck) SetStyle(s AIFactCheckStyle) *AIFactCheck {
	fc.mu.Lock()
	fc.style = s
	fc.recomputeLocked()
	fc.mu.Unlock()
	return fc
}

func (fc *AIFactCheck) Measure(cs Constraints) Size {
	w := 18
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

func (fc *AIFactCheck) Paint(buf *buffer.Buffer) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	b := fc.Bounds()
	x, y := b.X, b.Y

	resStyle := fc.curStyle
	srcStyle := fc.style.Source
	brStyle := fc.style.Bracket

	col := x

	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: factCheckIcons[fc.result], Fg: resStyle.Fg, Bg: resStyle.Bg, Flags: resStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: resStyle.Fg, Bg: resStyle.Bg, Flags: resStyle.Flags, Width: 1})
		col++
	}

	for _, r := range fc.labelStr {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: resStyle.Fg, Bg: resStyle.Bg, Flags: resStyle.Flags, Width: 1})
		col++
	}

	if fc.sources > 0 {
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: brStyle.Fg, Bg: brStyle.Bg, Flags: brStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: '(', Fg: brStyle.Fg, Bg: brStyle.Bg, Flags: brStyle.Flags, Width: 1})
			col++
		}
		for _, r := range fc.srcStr {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: srcStyle.Fg, Bg: srcStyle.Bg, Flags: srcStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: ')', Fg: brStyle.Fg, Bg: brStyle.Bg, Flags: brStyle.Flags, Width: 1})
		}
	}
}

func (fc *AIFactCheck) Children() []Component { return nil }

package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AIReasoningChain: AI Reasoning Step Chain Display ───
//
// AIReasoningChain renders a vertical chain of reasoning steps that
// show how an AI arrived at a conclusion. Each step shows a premise,
// an arrow, and a conclusion. Useful for transparent AI reasoning.
//
// Usage:
//
//	rc := NewAIReasoningChain()
//	rc.AddStep("User asks about X", "Search knowledge base")
//	rc.AddStep("Found 3 results", "Synthesize answer")
//	rc.SetConclusion("X is defined as...")
//	rc.Paint(buf)

type AIReasoningChainStyle struct {
	Premise     buffer.Style
	Arrow       buffer.Style
	Conclusion  buffer.Style
	FinalAnswer buffer.Style
	StepNum     buffer.Style
}

func DefaultAIReasoningChainStyle() AIReasoningChainStyle {
	return AIReasoningChainStyle{
		Premise:     buffer.Style{Fg: buffer.RGB(147, 197, 253)},
		Arrow:       buffer.Style{Fg: buffer.RGB(251, 191, 36)},
		Conclusion:  buffer.Style{Fg: buffer.RGB(203, 213, 225)},
		FinalAnswer: buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		StepNum:     buffer.Style{Fg: buffer.RGB(148, 163, 184)},
	}
}

const reasoningChainMaxSteps = 12

type reasoningStep struct {
	premise    string
	conclusion string
}

// AIReasoningChain renders a reasoning step chain.
type AIReasoningChain struct {
	BaseComponent
	mu sync.Mutex

	steps       [reasoningChainMaxSteps]reasoningStep
	count       int
	finalAnswer string
	width       int
	style       AIReasoningChainStyle
}

// NewAIReasoningChain creates an AIReasoningChain.
func NewAIReasoningChain() *AIReasoningChain {
	rc := &AIReasoningChain{width: 40, style: DefaultAIReasoningChainStyle()}
	rc.SetID(GenerateID("reasoning"))
	return rc
}

// AddStep adds a reasoning step.
func (rc *AIReasoningChain) AddStep(premise, conclusion string) *AIReasoningChain {
	rc.mu.Lock()
	if rc.count < reasoningChainMaxSteps {
		rc.steps[rc.count] = reasoningStep{premise: premise, conclusion: conclusion}
		rc.count++
	}
	rc.mu.Unlock()
	return rc
}

// SetConclusion sets the final answer/conclusion.
func (rc *AIReasoningChain) SetConclusion(s string) *AIReasoningChain {
	rc.mu.Lock()
	rc.finalAnswer = s
	rc.mu.Unlock()
	return rc
}

// Clear removes all steps.
func (rc *AIReasoningChain) Clear() *AIReasoningChain {
	rc.mu.Lock()
	rc.count = 0
	rc.finalAnswer = ""
	rc.mu.Unlock()
	return rc
}

// StepCount returns the number of reasoning steps.
func (rc *AIReasoningChain) StepCount() int {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return rc.count
}

// SetWidth sets the display width.
func (rc *AIReasoningChain) SetWidth(w int) *AIReasoningChain {
	rc.mu.Lock()
	if w < 10 {
		w = 10
	}
	rc.width = w
	rc.mu.Unlock()
	return rc
}

// SetStyle sets custom style.
func (rc *AIReasoningChain) SetStyle(s AIReasoningChainStyle) *AIReasoningChain {
	rc.mu.Lock()
	rc.style = s
	rc.mu.Unlock()
	return rc
}

// Measure returns preferred size.
func (rc *AIReasoningChain) Measure(cs Constraints) Size {
	rc.mu.Lock()
	h := rc.count*2 + 1 // each step = 2 rows + final answer
	rc.mu.Unlock()
	if h < 1 {
		h = 1
	}
	w := rc.width
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: h}
}

// Paint renders the reasoning chain.
func (rc *AIReasoningChain) Paint(buf *buffer.Buffer) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	b := rc.Bounds()
	x, y := b.X, b.Y

	premiseStyle := rc.style.Premise
	arrowStyle := rc.style.Arrow
	conclusionStyle := rc.style.Conclusion
	finalStyle := rc.style.FinalAnswer
	numStyle := rc.style.StepNum

	for i := 0; i < rc.count; i++ {
		step := rc.steps[i]
		yy := y + i*2
		if yy >= buf.Height {
			break
		}
		col := x

		// Step number
		for _, r := range "Step " + itoa(i+1) + ":" {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: numStyle.Fg, Bg: numStyle.Bg, Flags: numStyle.Flags, Width: 1})
			col++
		}

		// Premise
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: ' ', Fg: numStyle.Fg, Bg: numStyle.Bg, Flags: numStyle.Flags, Width: 1})
			col++
		}
		for _, r := range step.premise {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: premiseStyle.Fg, Bg: premiseStyle.Bg, Flags: premiseStyle.Flags, Width: 1})
			col++
		}

		// Arrow + conclusion on next row
		yy2 := yy + 1
		if yy2 < buf.Height {
			col = x
			for _, r := range "  ⇒ " {
				if col >= buf.Width {
					break
				}
				buf.SetCell(col, yy2, buffer.Cell{Rune: r, Fg: arrowStyle.Fg, Bg: arrowStyle.Bg, Flags: arrowStyle.Flags, Width: 1})
				col++
			}
			for _, r := range step.conclusion {
				if col >= buf.Width {
					break
				}
				buf.SetCell(col, yy2, buffer.Cell{Rune: r, Fg: conclusionStyle.Fg, Bg: conclusionStyle.Bg, Flags: conclusionStyle.Flags, Width: 1})
				col++
			}
		}
	}

	// Final answer
	if rc.finalAnswer != "" {
		yy := y + rc.count*2
		if yy < buf.Height {
			col := x
			for _, r := range "∴ " {
				if col >= buf.Width {
					break
				}
				buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: finalStyle.Fg, Bg: finalStyle.Bg, Flags: finalStyle.Flags, Width: 1})
				col++
			}
			for _, r := range rc.finalAnswer {
				if col >= buf.Width {
					break
				}
				buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: finalStyle.Fg, Bg: finalStyle.Bg, Flags: finalStyle.Flags, Width: 1})
				col++
			}
		}
	}
}

// Children returns nil.
func (rc *AIReasoningChain) Children() []Component { return nil }

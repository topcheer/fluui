package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── PromptChain: Sequential Prompt Execution Chain ───
//
// PromptChain renders a vertical chain of prompt steps showing execution
// order and status. Each step shows a number, name, and status icon.
// Useful for multi-step AI prompt pipelines.
//
// Usage:
//
//	pc := NewPromptChain()
//	pc.AddStep("Analyze query", ChainDone)
//	pc.AddStep("Retrieve context", ChainActive)
//	pc.AddStep("Generate response", ChainPending)
//	pc.Paint(buf)

// PromptChainStyle holds styling.
type PromptChainStyle struct {
	Done    buffer.Style
	Active  buffer.Style
	Pending buffer.Style
	Error   buffer.Style
	Connector buffer.Style
	Name    buffer.Style
}

// DefaultPromptChainStyle returns defaults.
func DefaultPromptChainStyle() PromptChainStyle {
	return PromptChainStyle{
		Done:      buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Active:    buffer.Style{Fg: buffer.RGB(251, 191, 36), Flags: buffer.Bold},
		Pending:   buffer.Style{Fg: buffer.RGB(100, 116, 139)},
		Error:     buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Connector: buffer.Style{Fg: buffer.RGB(71, 85, 105)},
		Name:      buffer.Style{Fg: buffer.RGB(226, 232, 240)},
	}
}

// ChainStepStatus represents the status of a chain step.
type ChainStepStatus int

const (
	ChainPending ChainStepStatus = 0
	ChainActive  ChainStepStatus = 1
	ChainDone    ChainStepStatus = 2
	ChainError   ChainStepStatus = 3
)

var chainStepIcons = [...]rune{'○', '●', '✓', '✗'}

const promptChainMaxSteps = 15

// promptChainStep holds a single chain step.
type promptChainStep struct {
	name   string
	status ChainStepStatus
}

// PromptChain renders a vertical prompt execution chain.
type PromptChain struct {
	BaseComponent
	mu sync.Mutex

	steps [promptChainMaxSteps]promptChainStep
	count int
	width int
	style PromptChainStyle
}

// NewPromptChain creates a PromptChain.
func NewPromptChain() *PromptChain {
	pc := &PromptChain{width: 30, style: DefaultPromptChainStyle()}
	pc.SetID(GenerateID("promptchain"))
	return pc
}

// AddStep adds a chain step.
func (pc *PromptChain) AddStep(name string, status ChainStepStatus) *PromptChain {
	pc.mu.Lock()
	if pc.count < promptChainMaxSteps {
		pc.steps[pc.count] = promptChainStep{name: name, status: status}
		pc.count++
	}
	pc.mu.Unlock()
	return pc
}

// SetStepStatus updates a step status by index.
func (pc *PromptChain) SetStepStatus(idx int, status ChainStepStatus) *PromptChain {
	pc.mu.Lock()
	if idx >= 0 && idx < pc.count {
		pc.steps[idx].status = status
	}
	pc.mu.Unlock()
	return pc
}

// Clear removes all steps.
func (pc *PromptChain) Clear() *PromptChain {
	pc.mu.Lock()
	pc.count = 0
	pc.mu.Unlock()
	return pc
}

// Count returns the number of steps.
func (pc *PromptChain) Count() int {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	return pc.count
}

// SetWidth sets the display width.
func (pc *PromptChain) SetWidth(w int) *PromptChain {
	pc.mu.Lock()
	if w < 10 { w = 10 }
	pc.width = w
	pc.mu.Unlock()
	return pc
}

// SetStyle sets custom style.
func (pc *PromptChain) SetStyle(s PromptChainStyle) *PromptChain {
	pc.mu.Lock()
	pc.style = s
	pc.mu.Unlock()
	return pc
}

// Measure returns preferred size.
func (pc *PromptChain) Measure(cs Constraints) Size {
	pc.mu.Lock()
	h := pc.count * 2 // each step is 2 rows (step + connector)
	pc.mu.Unlock()
	if h < 1 { h = 1 }
	w := pc.width
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: h}
}

// Paint renders the prompt chain.
func (pc *PromptChain) Paint(buf *buffer.Buffer) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	b := pc.Bounds()
	x, y := b.X, b.Y

	connStyle := pc.style.Connector
	nameStyle := pc.style.Name

	for i := 0; i < pc.count; i++ {
		step := pc.steps[i]
		yy := y + i * 2
		if yy >= buf.Height { break }

		var st buffer.Style
		switch step.status {
		case ChainDone:
			st = pc.style.Done
		case ChainActive:
			st = pc.style.Active
		case ChainError:
			st = pc.style.Error
		default:
			st = pc.style.Pending
		}

		// Step icon
		if x < buf.Width {
			buf.SetCell(x, yy, buffer.Cell{Rune: chainStepIcons[step.status], Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		}
		// Step number
		numStr := " " + itoa(i+1) + ". "
		col := x + 1
		for _, r := range numStr {
			if col >= buf.Width { break }
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
			col++
		}
		// Step name
		for _, r := range step.name {
			if col >= buf.Width { break }
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: nameStyle.Fg, Bg: nameStyle.Bg, Flags: nameStyle.Flags, Width: 1})
			col++
		}

		// Connector to next step
		if i < pc.count-1 {
			connY := yy + 1
			if connY < buf.Height && x < buf.Width {
				buf.SetCell(x, connY, buffer.Cell{Rune: '│', Fg: connStyle.Fg, Bg: connStyle.Bg, Flags: connStyle.Flags, Width: 1})
			}
		}
	}
}

// Children returns nil.
func (pc *PromptChain) Children() []Component { return nil }

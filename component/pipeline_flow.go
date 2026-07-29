package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── PipelineFlow: Horizontal Pipeline Flow Diagram ───
//
// PipelineFlow renders a horizontal pipeline showing processing stages
// connected by arrows. Each stage shows a name and optional status icon.
// Useful for visualizing data processing pipelines or AI inference stages.
//
// Usage:
//
//	pf := NewPipelineFlow()
//	pf.AddStage("Input", StageDone)
//	pf.AddStage("Tokenize", StageActive)
//	pf.AddStage("Embed", StagePending)
//	pf.Paint(buf)

// PipelineStageStatus represents the status of a pipeline stage.
type PipelineStageStatus int

const (
	StagePending PipelineStageStatus = 0
	StageActive  PipelineStageStatus = 1
	StageDone    PipelineStageStatus = 2
	StageError   PipelineStageStatus = 3
)

// PipelineFlowStyle holds styling.
type PipelineFlowStyle struct {
	Done    buffer.Style
	Active  buffer.Style
	Pending buffer.Style
	Error   buffer.Style
	Arrow   buffer.Style
}

// DefaultPipelineFlowStyle returns defaults.
func DefaultPipelineFlowStyle() PipelineFlowStyle {
	return PipelineFlowStyle{
		Done:    buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Active:  buffer.Style{Fg: buffer.RGB(251, 191, 36), Flags: buffer.Bold},
		Pending: buffer.Style{Fg: buffer.RGB(100, 116, 139)},
		Error:   buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Arrow:   buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

var stageStatusIcons = [...]rune{'○', '●', '✓', '✗'}

const pipelineMaxStages = 10

// pipelineStage holds a single pipeline stage.
type pipelineStage struct {
	name   string
	status PipelineStageStatus
}

// PipelineFlow renders a pipeline flow diagram.
type PipelineFlow struct {
	BaseComponent
	mu sync.Mutex

	stages [pipelineMaxStages]pipelineStage
	count  int
	style  PipelineFlowStyle
}

// NewPipelineFlow creates a PipelineFlow.
func NewPipelineFlow() *PipelineFlow {
	pf := &PipelineFlow{style: DefaultPipelineFlowStyle()}
	pf.SetID(GenerateID("pipeline"))
	return pf
}

// AddStage adds a pipeline stage with name and status.
func (pf *PipelineFlow) AddStage(name string, status PipelineStageStatus) *PipelineFlow {
	pf.mu.Lock()
	if pf.count < pipelineMaxStages {
		pf.stages[pf.count] = pipelineStage{name: name, status: status}
		pf.count++
	}
	pf.mu.Unlock()
	return pf
}

// SetStageStatus updates the status of a stage by index.
func (pf *PipelineFlow) SetStageStatus(idx int, status PipelineStageStatus) *PipelineFlow {
	pf.mu.Lock()
	if idx >= 0 && idx < pf.count {
		pf.stages[idx].status = status
	}
	pf.mu.Unlock()
	return pf
}

// Clear removes all stages.
func (pf *PipelineFlow) Clear() *PipelineFlow {
	pf.mu.Lock()
	pf.count = 0
	pf.mu.Unlock()
	return pf
}

// Count returns the number of stages.
func (pf *PipelineFlow) Count() int {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	return pf.count
}

// SetStyle sets custom style.
func (pf *PipelineFlow) SetStyle(s PipelineFlowStyle) *PipelineFlow {
	pf.mu.Lock()
	pf.style = s
	pf.mu.Unlock()
	return pf
}

// Measure returns preferred size.
func (pf *PipelineFlow) Measure(cs Constraints) Size {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	w := 0
	for i := 0; i < pf.count; i++ {
		w += len(pf.stages[i].name) + 3 // icon + space + name
		if i < pf.count-1 { w += 3 }     // arrow + spaces
	}
	if w < 10 { w = 10 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 1}
}

// Paint renders the pipeline flow.
func (pf *PipelineFlow) Paint(buf *buffer.Buffer) {
	pf.mu.Lock()
	defer pf.mu.Unlock()

	b := pf.Bounds()
	x, y := b.X, b.Y

	for i := 0; i < pf.count; i++ {
		stage := pf.stages[i]
		var style_ buffer.Style
		switch stage.status {
		case StageDone:
			style_ = pf.style.Done
		case StageActive:
			style_ = pf.style.Active
		case StageError:
			style_ = pf.style.Error
		default:
			style_ = pf.style.Pending
		}

		// Icon
		if x >= buf.Width { break }
		buf.SetCell(x, y, buffer.Cell{Rune: stageStatusIcons[stage.status], Fg: style_.Fg, Bg: style_.Bg, Flags: style_.Flags, Width: 1})
		x++

		// Space
		if x >= buf.Width { break }
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Fg: style_.Fg, Bg: style_.Bg, Flags: style_.Flags, Width: 1})
		x++

		// Name
		for _, r := range stage.name {
			if x >= buf.Width { break }
			buf.SetCell(x, y, buffer.Cell{Rune: r, Fg: style_.Fg, Bg: style_.Bg, Flags: style_.Flags, Width: 1})
			x++
		}

		// Arrow to next stage
		if i < pf.count-1 {
			arrowStyle := pf.style.Arrow
			if x >= buf.Width { break }
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Fg: arrowStyle.Fg, Bg: arrowStyle.Bg, Flags: arrowStyle.Flags, Width: 1})
			x++
			if x >= buf.Width { break }
			buf.SetCell(x, y, buffer.Cell{Rune: '→', Fg: arrowStyle.Fg, Bg: arrowStyle.Bg, Flags: arrowStyle.Flags, Width: 1})
			x++
			if x >= buf.Width { break }
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Fg: arrowStyle.Fg, Bg: arrowStyle.Bg, Flags: arrowStyle.Flags, Width: 1})
			x++
		}
	}
}

// Children returns nil.
func (pf *PipelineFlow) Children() []Component { return nil }

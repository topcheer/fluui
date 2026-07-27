package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── FunnelChart: Conversion / Sales Funnel ───
//
// FunnelChart renders a series of decreasing horizontal sections resembling
// a funnel. Each section's width is proportional to its value. Common in
// sales pipelines, conversion analytics, and user journey tracking.
//
// Usage:
//
//	fc := NewFunnelChart()
//	fc.AddStage(FunnelStage{Label: "Visitors", Value: 10000})
//	fc.AddStage(FunnelStage{Label: "Signups", Value: 3000})
//	fc.AddStage(FunnelStage{Label: "Purchased", Value: 800})
//	fc.SetBounds(Rect{X:0, Y:0, W:50, H:12})
//	fc.Paint(buf)

// FunnelStage represents one level of the funnel.
type FunnelStage struct {
	Label string
	Value float64
	Color buffer.Color
}

// FunnelChartStyle holds visual styles.
type FunnelChartStyle struct {
	Stage    buffer.Style
	Label    buffer.Style
	Value    buffer.Style
	Connector buffer.Style
}

// DefaultFunnelChartStyle returns sensible defaults.
func DefaultFunnelChartStyle() FunnelChartStyle {
	return FunnelChartStyle{
		Stage:     buffer.Style{Fg: buffer.RGB(100, 149, 237)},
		Label:     buffer.Style{Fg: buffer.White, Flags: buffer.Bold},
		Value:     buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		Connector: buffer.Style{Fg: buffer.RGB(60, 60, 60)},
	}
}

var funnelPalette = [...]buffer.Color{
	buffer.RGB(100, 149, 237),
	buffer.RGB(64, 200, 200),
	buffer.RGB(16, 163, 127),
	buffer.RGB(255, 175, 64),
	buffer.RGB(220, 80, 80),
	buffer.RGB(147, 112, 219),
	buffer.RGB(255, 192, 203),
	buffer.RGB(100, 200, 100),
}

// FunnelChart renders a conversion funnel.
type FunnelChart struct {
	BaseComponent
	mu     sync.RWMutex
	stages []FunnelStage
	style  FunnelChartStyle
}

// NewFunnelChart creates an empty funnel chart.
func NewFunnelChart() *FunnelChart {
	fc := &FunnelChart{
		style: DefaultFunnelChartStyle(),
	}
	fc.SetID(GenerateID("funnel"))
	return fc
}

// AddStage adds a funnel level.
func (fc *FunnelChart) AddStage(s FunnelStage) *FunnelChart {
	fc.mu.Lock()
	if s.Color.Type == 0 {
		s.Color = funnelPalette[len(fc.stages)%len(funnelPalette)]
	}
	fc.stages = append(fc.stages, s)
	fc.mu.Unlock()
	return fc
}

// SetStages replaces all stages.
func (fc *FunnelChart) SetStages(stages []FunnelStage) *FunnelChart {
	fc.mu.Lock()
	fc.stages = stages
	fc.mu.Unlock()
	return fc
}

// Stages returns the current stages.
func (fc *FunnelChart) Stages() []FunnelStage {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.stages
}

// StageCount returns the number of stages.
func (fc *FunnelChart) StageCount() int {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return len(fc.stages)
}

// Clear removes all stages.
func (fc *FunnelChart) Clear() *FunnelChart {
	fc.mu.Lock()
	fc.stages = fc.stages[:0]
	fc.mu.Unlock()
	return fc
}

// SetStyle sets the visual style.
func (fc *FunnelChart) SetStyle(s FunnelChartStyle) *FunnelChart {
	fc.mu.Lock()
	fc.style = s
	fc.mu.Unlock()
	return fc
}

// Style returns the current style.
func (fc *FunnelChart) Style() FunnelChartStyle {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.style
}

// Measure computes the desired size.
func (fc *FunnelChart) Measure(cs Constraints) Size {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	w := 40
	h := len(fc.stages)*2 + 1
	if h < 5 {
		h = 5
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the funnel chart.
func (fc *FunnelChart) Paint(buf *buffer.Buffer) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	b := fc.bounds
	if b.W < 6 || b.H < 3 || len(fc.stages) == 0 {
		return
	}

	maxVal := fc.stages[0].Value
	if maxVal <= 0 {
		return
	}

	stageH := b.H / len(fc.stages)
	if stageH < 1 {
		stageH = 1
	}

	centerX := b.X + b.W/2

	for i, stage := range fc.stages {
		y0 := b.Y + i*stageH
		if y0 >= b.Y+b.H {
			break
		}

		ratio := stage.Value / maxVal
		if ratio < 0 {
			ratio = 0
		}
		stageW := int(ratio * float64(b.W))
		if stageW < 2 {
			stageW = 2
		}

		// Center the stage bar
		startX := centerX - stageW/2
		if startX < b.X {
			startX = b.X
		}

		// Draw stage bar rows
		for row := 0; row < stageH && y0+row < b.Y+b.H; row++ {
			for x := 0; x < stageW; x++ {
				ax := startX + x
				if ax >= b.X+b.W {
					break
				}
				buf.SetCell(ax, y0+row, buffer.Cell{
					Rune:  '█',
					Fg:    stage.Color,
					Bg:    fc.style.Stage.Bg,
					Flags: fc.style.Stage.Flags,
					Width: 1,
				})
			}
		}

		// Draw label centered on the stage
		if stageH >= 1 && stage.Label != "" {
			labelY := y0 + stageH/2
			if labelY >= b.Y+b.H {
				labelY = b.Y + b.H - 1
			}
			labelRunes := []rune(stage.Label)
			labelX := centerX - len(labelRunes)/2
			if labelX < b.X {
				labelX = b.X
			}
			for _, r := range labelRunes {
				if labelX >= b.X+b.W {
					break
				}
				buf.SetCell(labelX, labelY, buffer.Cell{
					Rune:  r,
					Fg:    fc.style.Label.Fg,
					Bg:    stage.Color,
					Flags: fc.style.Label.Flags,
					Width: 1,
				})
				labelX++
			}
		}
	}
}

// Children returns nil.
func (fc *FunnelChart) Children() []Component { return nil }

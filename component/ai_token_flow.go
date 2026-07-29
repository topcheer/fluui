package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AITokenFlow: Visualize Token Flow Through AI Pipeline Stages ───
//
// AITokenFlow renders connected horizontal bars showing token counts at each
// stage of an AI pipeline (input -> embedding -> attention -> output).
//
// Usage:
//
//	tf := NewAITokenFlow()
//	tf.AddStage("Input", 500, buffer.RGB(96, 165, 250))
//	tf.AddStage("Embedding", 480, buffer.RGB(167, 139, 250))
//	tf.AddStage("Attention", 350, buffer.RGB(34, 197, 94))
//	tf.AddStage("Output", 200, buffer.RGB(234, 179, 8))
//	tf.Paint(buf)

// TokenFlowStage represents a single pipeline stage.
type TokenFlowStage struct {
	Name       string
	TokenCount int
	Color      buffer.Color
	// cached
	CountStr string
	BarWidth int
}

// AITokenFlowStyle holds styling.
type AITokenFlowStyle struct {
	Connector buffer.Style
	Label     buffer.Style
	Border    buffer.Style
}

// DefaultAITokenFlowStyle returns defaults.
func DefaultAITokenFlowStyle() AITokenFlowStyle {
	conn := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return AITokenFlowStyle{Connector: conn, Label: label, Border: border}
}

// AITokenFlow visualizes token flow through AI pipeline stages.
type AITokenFlow struct {
	BaseComponent
	mu sync.Mutex

	stages []TokenFlowStage
	maxBar int
	style  AITokenFlowStyle
}

// NewAITokenFlow creates an AITokenFlow.
func NewAITokenFlow() *AITokenFlow {
	tf := &AITokenFlow{maxBar: 20, style: DefaultAITokenFlowStyle()}
	tf.SetID(GenerateID("tokenflow"))
	return tf
}

// AddStage adds a pipeline stage with cached display strings.
func (tf *AITokenFlow) AddStage(name string, tokenCount int, color buffer.Color) *AITokenFlow {
	tf.mu.Lock()
	stage := TokenFlowStage{
		Name:       name,
		TokenCount: tokenCount,
		Color:      color,
		CountStr:   itoa(tokenCount) + " tok",
	}
	tf.stages = append(tf.stages, stage)
	tf.recomputeWidthsLocked()
	tf.mu.Unlock()
	return tf
}

// recomputeWidthsLocked calculates bar widths from token counts.
func (tf *AITokenFlow) recomputeWidthsLocked() {
	maxCount := 1
	for _, s := range tf.stages {
		if s.TokenCount > maxCount {
			maxCount = s.TokenCount
		}
	}
	for i := range tf.stages {
		ratio := float64(tf.stages[i].TokenCount) / float64(maxCount)
		if ratio > 1 {
			ratio = 1
		}
		tf.stages[i].BarWidth = int(ratio * float64(tf.maxBar))
		if tf.stages[i].BarWidth < 1 {
			tf.stages[i].BarWidth = 1
		}
	}
}

// StageCount returns the number of stages.
func (tf *AITokenFlow) StageCount() int {
	tf.mu.Lock()
	defer tf.mu.Unlock()
	return len(tf.stages)
}

// Clear removes all stages.
func (tf *AITokenFlow) Clear() *AITokenFlow {
	tf.mu.Lock()
	tf.stages = tf.stages[:0]
	tf.mu.Unlock()
	return tf
}

// SetMaxBarWidth sets the maximum bar width in characters.
func (tf *AITokenFlow) SetMaxBarWidth(w int) *AITokenFlow {
	tf.mu.Lock()
	tf.maxBar = w
	tf.recomputeWidthsLocked()
	tf.mu.Unlock()
	return tf
}

// SetStyle sets custom style.
func (tf *AITokenFlow) SetStyle(s AITokenFlowStyle) *AITokenFlow {
	tf.mu.Lock()
	tf.style = s
	tf.mu.Unlock()
	return tf
}

// Measure returns the preferred size.
func (tf *AITokenFlow) Measure(cs Constraints) Size {
	tf.mu.Lock()
	count := len(tf.stages)
	tf.mu.Unlock()
	w := 50
	h := count*2 + 2
	if h < 3 {
		h = 3
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the token flow into the buffer.
func (tf *AITokenFlow) Paint(buf *buffer.Buffer) {
	tf.mu.Lock()
	defer tf.mu.Unlock()

	b := tf.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 20 {
		w = 50
	}
	if h < 3 {
		h = 5
	}

	bs := tf.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 {
				ch = '┌'
			} else if row == 0 && col == w-1 {
				ch = '┐'
			} else if row == h-1 && col == 0 {
				ch = '└'
			} else if row == h-1 && col == w-1 {
				ch = '┘'
			} else if row == 0 || row == h-1 {
				ch = '─'
			} else if col == 0 || col == w-1 {
				ch = '│'
			}
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	labelStyle := tf.style.Label
	connStyle := tf.style.Connector

	for idx, stage := range tf.stages {
		rowY := y + 1 + idx*2
		if rowY >= y+h-1 || rowY >= buf.Height {
			break
		}

		col := x + 1

		// Stage name
		for _, r := range stage.Name {
			if col >= x+w-1 || col >= buf.Width {
				break
			}
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}

		// Padding to bar start
		barStart := x + 15
		for c := col; c < barStart && c < x+w-1 && c < buf.Width; c++ {
			buf.SetCell(c, rowY, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		}
		col = barStart

		// Bar
		for i := 0; i < stage.BarWidth; i++ {
			if col >= x+w-1 || col >= buf.Width {
				break
			}
			buf.SetCell(col, rowY, buffer.Cell{Rune: '█', Fg: stage.Color, Bg: buffer.NoColor(), Width: 1})
			col++
		}

		// Token count text
		if col < x+w-1 && col < buf.Width {
			buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}
		for _, r := range stage.CountStr {
			if col >= x+w-1 || col >= buf.Width {
				break
			}
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}

		// Connector arrow to next stage
		if idx < len(tf.stages)-1 {
			connY := rowY + 1
			if connY < y+h-1 && connY < buf.Height && barStart < buf.Width {
				buf.SetCell(barStart, connY, buffer.Cell{Rune: '↓', Fg: connStyle.Fg, Bg: connStyle.Bg, Flags: connStyle.Flags, Width: 1})
			}
		}
	}
}

// Children returns nil.
func (tf *AITokenFlow) Children() []Component { return nil }

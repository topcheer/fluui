package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── BubbleChart: Scatter Plot with Variable-Size Bubbles ───
//
// BubbleChart renders data points as circles whose size represents a third
// dimension (weight/value). Useful for multi-dimensional data analysis.
//
// Usage:
//
//	bc := NewBubbleChart()
//	bc.AddBubble(BubbleData{X: 10, Y: 20, Size: 5, Label: "A"})
//	bc.AddBubble(BubbleData{X: 30, Y: 50, Size: 10, Label: "B"})
//	bc.SetBounds(Rect{X:0, Y:0, W:40, H:15})
//	bc.Paint(buf)

// BubbleData represents a single bubble.
type BubbleData struct {
	X     float64
	Y     float64
	Size  float64 // determines bubble radius (larger = bigger)
	Label string
	Color buffer.Color
}

// BubbleChartStyle holds visual styles.
type BubbleChartStyle struct {
	Bubble buffer.Style
	Axis   buffer.Style
	Grid   buffer.Style
	Label  buffer.Style
}

// DefaultBubbleChartStyle returns sensible defaults.
func DefaultBubbleChartStyle() BubbleChartStyle {
	return BubbleChartStyle{
		Bubble: buffer.Style{Fg: buffer.RGB(100, 149, 237)},
		Axis:   buffer.Style{Fg: buffer.RGB(80, 80, 80)},
		Grid:   buffer.Style{Fg: buffer.RGB(40, 40, 40)},
		Label:  buffer.Style{Fg: buffer.White},
	}
}

var bubblePalette = [...]buffer.Color{
	buffer.RGB(100, 149, 237),
	buffer.RGB(16, 163, 127),
	buffer.RGB(220, 80, 80),
	buffer.RGB(255, 175, 64),
	buffer.RGB(147, 112, 219),
	buffer.RGB(64, 224, 208),
}

// BubbleChart renders a scatter plot with variable-size bubbles.
type BubbleChart struct {
	BaseComponent
	mu      sync.RWMutex
	bubbles []BubbleData
	style   BubbleChartStyle
}

// NewBubbleChart creates an empty bubble chart.
func NewBubbleChart() *BubbleChart {
	bc := &BubbleChart{
		style: DefaultBubbleChartStyle(),
	}
	bc.SetID(GenerateID("bubble"))
	return bc
}

// AddBubble adds a data point.
func (bc *BubbleChart) AddBubble(b BubbleData) *BubbleChart {
	bc.mu.Lock()
	if b.Color.Type == 0 {
		b.Color = bubblePalette[len(bc.bubbles)%len(bubblePalette)]
	}
	bc.bubbles = append(bc.bubbles, b)
	bc.mu.Unlock()
	return bc
}

// SetBubbles replaces all bubbles.
func (bc *BubbleChart) SetBubbles(bubbles []BubbleData) *BubbleChart {
	bc.mu.Lock()
	bc.bubbles = bubbles
	bc.mu.Unlock()
	return bc
}

// Bubbles returns the current data.
func (bc *BubbleChart) Bubbles() []BubbleData {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.bubbles
}

// BubbleCount returns the number of bubbles.
func (bc *BubbleChart) BubbleCount() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return len(bc.bubbles)
}

// Clear removes all bubbles.
func (bc *BubbleChart) Clear() *BubbleChart {
	bc.mu.Lock()
	bc.bubbles = bc.bubbles[:0]
	bc.mu.Unlock()
	return bc
}

// SetStyle sets the visual style.
func (bc *BubbleChart) SetStyle(s BubbleChartStyle) *BubbleChart {
	bc.mu.Lock()
	bc.style = s
	bc.mu.Unlock()
	return bc
}

// Style returns the current style.
func (bc *BubbleChart) Style() BubbleChartStyle {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.style
}

// dataRangeLocked returns min/max for X and Y (caller holds lock).
func (bc *BubbleChart) dataRangeLocked() (minX, maxX, minY, maxY float64) {
	if len(bc.bubbles) == 0 {
		return 0, 1, 0, 1
	}
	minX, maxX = bc.bubbles[0].X, bc.bubbles[0].X
	minY, maxY = bc.bubbles[0].Y, bc.bubbles[0].Y
	for _, b := range bc.bubbles {
		if b.X < minX {
			minX = b.X
		}
		if b.X > maxX {
			maxX = b.X
		}
		if b.Y < minY {
			minY = b.Y
		}
		if b.Y > maxY {
			maxY = b.Y
		}
	}
	if maxX == minX {
		maxX = minX + 1
	}
	if maxY == minY {
		maxY = minY + 1
	}
	return
}

// Measure computes the desired size.
func (bc *BubbleChart) Measure(cs Constraints) Size {
	w, h := 40, 15
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the bubble chart.
func (bc *BubbleChart) Paint(buf *buffer.Buffer) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	b := bc.bounds
	if b.W < 6 || b.H < 4 || len(bc.bubbles) == 0 {
		return
	}

	minX, maxX, minY, maxY := bc.dataRangeLocked()
	chartH := b.H - 1 // reserve for axis
	if chartH < 2 {
		chartH = 2
	}

	// Draw grid dots
	for y := 0; y < chartH; y += 3 {
		for x := 0; x < b.W; x += 5 {
			buf.SetCell(b.X+x, b.Y+y, buffer.Cell{Rune: '·', Fg: bc.style.Grid.Fg, Bg: bc.style.Grid.Bg, Width: 1})
		}
	}

	// Find max bubble size for normalization
	maxSize := 0.0
	for _, bub := range bc.bubbles {
		if bub.Size > maxSize {
			maxSize = bub.Size
		}
	}
	if maxSize == 0 {
		maxSize = 1
	}

	// Draw bubbles
	for _, bub := range bc.bubbles {
		// Map X,Y to screen coordinates
		ratioX := (bub.X - minX) / (maxX - minX)
		ratioY := (maxY - bub.Y) / (maxY - minY) // invert Y (top = high)
		px := b.X + int(ratioX*float64(b.W-2))
		py := b.Y + int(ratioY*float64(chartH-1))
		if px < b.X || px >= b.X+b.W || py < b.Y || py >= b.Y+chartH {
			continue
		}

		// Bubble radius (1-3 chars)
		radius := int(bub.Size/maxSize*2.0) + 1
		if radius > 3 {
			radius = 3
		}

		// Draw circle approximation
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				dist := dx*dx + dy*dy
				if dist <= radius*radius {
					ax := px + dx
					ay := py + dy
					if ax >= b.X && ax < b.X+b.W && ay >= b.Y && ay < b.Y+chartH {
						buf.SetCell(ax, ay, buffer.Cell{Rune: '●', Fg: bub.Color, Bg: bc.style.Bubble.Bg, Width: 1})
					}
				}
			}
		}

		// Label to the right
		if bub.Label != "" {
			lx := px + radius + 1
			for _, r := range bub.Label {
				if lx >= b.X+b.W {
					break
				}
				buf.SetCell(lx, py, buffer.Cell{Rune: r, Fg: bc.style.Label.Fg, Bg: bc.style.Label.Bg, Width: 1})
				lx++
			}
		}
	}

	// Axis
	axisY := b.Y + chartH
	for x := 0; x < b.W; x++ {
		buf.SetCell(b.X+x, axisY, buffer.Cell{Rune: '─', Fg: bc.style.Axis.Fg, Bg: bc.style.Axis.Bg, Width: 1})
	}
}

// Children returns nil.
func (bc *BubbleChart) Children() []Component { return nil }

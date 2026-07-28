package component

import (
	"sort"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── TreemapChart: Nested Rectangle Hierarchy ───
//
// TreemapChart renders hierarchical data as nested rectangles where area
// is proportional to value. Common for disk usage, portfolio allocation,
// and market share visualization.
//
// Usage:
//
//	tc := NewTreemapChart()
//	tc.AddNode(TreemapNode{Label: "Documents", Value: 40})
//	tc.AddNode(TreemapNode{Label: "Photos", Value: 30})
//	tc.AddNode(TreemapNode{Label: "Videos", Value: 20})
//	tc.SetBounds(Rect{X:0, Y:0, W:50, H:15})
//	tc.Paint(buf)

// TreemapNode represents a node in the treemap.
type TreemapNode struct {
	Label string
	Value float64
	Color buffer.Color
}

// TreemapChartStyle holds visual styles.
type TreemapChartStyle struct {
	Node  buffer.Style
	Label buffer.Style
	Border buffer.Style
}

// DefaultTreemapChartStyle returns sensible defaults.
func DefaultTreemapChartStyle() TreemapChartStyle {
	return TreemapChartStyle{
		Node:   buffer.Style{Fg: buffer.RGB(100, 149, 237)},
		Label:  buffer.Style{Fg: buffer.White, Flags: buffer.Bold},
		Border: buffer.Style{Fg: buffer.RGB(30, 30, 30)},
	}
}

var treemapPalette = [...]buffer.Color{
	buffer.RGB(100, 149, 237),
	buffer.RGB(16, 163, 127),
	buffer.RGB(255, 175, 64),
	buffer.RGB(220, 80, 80),
	buffer.RGB(147, 112, 219),
	buffer.RGB(64, 224, 208),
	buffer.RGB(255, 192, 203),
	buffer.RGB(255, 215, 0),
}

// treemapRect represents a computed rectangle for a node.
type treemapRect struct {
	label string
	color buffer.Color
	x, y, w, h int
}

// TreemapChart renders a nested rectangle hierarchy.
type TreemapChart struct {
	BaseComponent
	mu    sync.RWMutex
	nodes []TreemapNode
	style TreemapChartStyle
}

// NewTreemapChart creates an empty treemap.
func NewTreemapChart() *TreemapChart {
	tc := &TreemapChart{
		style: DefaultTreemapChartStyle(),
	}
	tc.SetID(GenerateID("treemap"))
	return tc
}

// AddNode adds a node.
func (tc *TreemapChart) AddNode(n TreemapNode) *TreemapChart {
	tc.mu.Lock()
	if n.Color.Type == 0 {
		n.Color = treemapPalette[len(tc.nodes)%len(treemapPalette)]
	}
	tc.nodes = append(tc.nodes, n)
	tc.mu.Unlock()
	return tc
}

// SetNodes replaces all nodes.
func (tc *TreemapChart) SetNodes(nodes []TreemapNode) *TreemapChart {
	tc.mu.Lock()
	tc.nodes = nodes
	tc.mu.Unlock()
	return tc
}

// Nodes returns the current nodes.
func (tc *TreemapChart) Nodes() []TreemapNode {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.nodes
}

// NodeCount returns the number of nodes.
func (tc *TreemapChart) NodeCount() int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return len(tc.nodes)
}

// TotalValue returns the sum of all values.
func (tc *TreemapChart) TotalValue() float64 {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	total := 0.0
	for _, n := range tc.nodes {
		total += n.Value
	}
	return total
}

// Clear removes all nodes.
func (tc *TreemapChart) Clear() *TreemapChart {
	tc.mu.Lock()
	tc.nodes = tc.nodes[:0]
	tc.mu.Unlock()
	return tc
}

// SetStyle sets the visual style.
func (tc *TreemapChart) SetStyle(s TreemapChartStyle) *TreemapChart {
	tc.mu.Lock()
	tc.style = s
	tc.mu.Unlock()
	return tc
}

// layout computes slice-and-dice treemap layout (caller holds lock).
func (tc *TreemapChart) layoutLocked(b Rect) []treemapRect {
	n := len(tc.nodes)
	if n == 0 {
		return nil
	}

	// Sort by value descending (copy to avoid mutating original)
	sorted := make([]TreemapNode, n)
	copy(sorted, tc.nodes)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	total := 0.0
	for _, node := range sorted {
		total += node.Value
	}
	if total <= 0 {
		return nil
	}

	rects := make([]treemapRect, 0, n)
	x, y := b.X, b.Y
	remainingW := b.W
	remainingH := b.H
	horizontal := remainingW >= remainingH

	for i, node := range sorted {
		ratio := node.Value / total
		var rw, rh int
		if horizontal {
			rw = int(float64(remainingW) * ratio)
			rh = remainingH
		} else {
			rw = remainingW
			rh = int(float64(remainingH) * ratio)
		}
		if rw < 1 {
			rw = 1
		}
		if rh < 1 {
			rh = 1
		}

		rects = append(rects, treemapRect{
			label: node.Label,
			color: node.Color,
			x:     x,
			y:     y,
			w:     rw,
			h:     rh,
		})

		// Advance position
		if horizontal {
			x += rw
			remainingW -= rw
		} else {
			y += rh
			remainingH -= rh
		}

		// Subtract from total for remaining proportional split
		total -= node.Value
		if total <= 0 {
			break
		}

		// Switch orientation if remaining area is more square
		// Check every other node
		if i%2 == 0 && remainingW > 0 && remainingH > 0 {
			horizontal = remainingW >= remainingH
		}
	}

	return rects
}

// Measure computes the desired size.
func (tc *TreemapChart) Measure(cs Constraints) Size {
	w, h := 40, 15
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the treemap.
func (tc *TreemapChart) Paint(buf *buffer.Buffer) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	b := tc.bounds
	if b.W < 4 || b.H < 3 || len(tc.nodes) == 0 {
		return
	}

	rects := tc.layoutLocked(b)

	for _, r := range rects {
		// Fill rectangle
		for dy := 0; dy < r.h; dy++ {
			for dx := 0; dx < r.w; dx++ {
				ax := r.x + dx
				ay := r.y + dy
				if ax >= b.X+b.W || ay >= b.Y+b.H {
					break
				}
				isBorder := dx == 0 || dy == 0 || dx == r.w-1 || dy == r.h-1
				if isBorder && r.w > 2 && r.h > 2 {
					buf.SetCell(ax, ay, buffer.Cell{Rune: '░', Fg: tc.style.Border.Fg, Bg: r.color, Width: 1})
				} else {
					buf.SetCell(ax, ay, buffer.Cell{Rune: '█', Fg: r.color, Bg: tc.style.Node.Bg, Width: 1})
				}
			}
		}

		// Label (if rectangle is big enough)
		if r.w >= 6 && r.h >= 2 {
			labelRunes := []rune(r.label)
			maxLabel := r.w - 2
			if len(labelRunes) > maxLabel {
				labelRunes = labelRunes[:maxLabel]
			}
			for i, runeVal := range labelRunes {
				ax := r.x + 1 + i
				if ax >= r.x+r.w-1 {
					break
				}
				buf.SetCell(ax, r.y, buffer.Cell{Rune: runeVal, Fg: tc.style.Label.Fg, Bg: r.color, Flags: tc.style.Label.Flags, Width: 1})
			}
		}
	}
}

// Children returns nil.
func (tc *TreemapChart) Children() []Component { return nil }

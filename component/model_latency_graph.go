package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ModelLatencyGraph: Model Response Latency Sparkline ───
//
// ModelLatencyGraph renders a compact sparkline-style graph showing model
// response latency (in milliseconds) over recent API calls. Useful for
// spotting performance regressions or throttling.
//
// Usage:
//
//	g := NewModelLatencyGraph()
//	g.AddLatency(120)  // 120ms
//	g.AddLatency(350)  // 350ms
//	g.Paint(buf)

// LatencyGraphStyle holds styling.
type LatencyGraphStyle struct {
	Line    buffer.Style
	Fill    buffer.Style
	Label   buffer.Style
	Peak    buffer.Style
	Current buffer.Style
}

// DefaultLatencyGraphStyle returns defaults.
func DefaultLatencyGraphStyle() LatencyGraphStyle {
	return LatencyGraphStyle{
		Line:    buffer.Style{Fg: buffer.RGB(59, 130, 246)},
		Fill:    buffer.Style{Fg: buffer.RGB(30, 58, 95)},
		Label:   buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Peak:    buffer.Style{Fg: buffer.RGB(239, 68, 68)},
		Current: buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
	}
}

const latencyMaxPoints = 30

// ModelLatencyGraph renders a latency sparkline graph.
type ModelLatencyGraph struct {
	BaseComponent
	mu sync.Mutex

	points [latencyMaxPoints]int // latency in ms, circular buffer
	count  int                   // how many valid points
	head   int                   // next write position
	width  int
	height int
	style  LatencyGraphStyle
	// cached
	heights    [latencyMaxPoints]int // pre-computed bar heights
	heightsN   int                   // number of valid entries
	curStr     string
	avgStr     string
	peakStr    string
	heightDirty bool
}

// NewModelLatencyGraph creates a ModelLatencyGraph.
func NewModelLatencyGraph() *ModelLatencyGraph {
	g := &ModelLatencyGraph{width: 40, height: 5, style: DefaultLatencyGraphStyle()}
	g.SetID(GenerateID("latgraph"))
	g.recomputeLocked()
	return g
}

// AddLatency adds a latency sample in milliseconds.
func (g *ModelLatencyGraph) AddLatency(ms int) *ModelLatencyGraph {
	g.mu.Lock()
	if ms < 0 { ms = 0 }
	g.points[g.head] = ms
	g.head = (g.head + 1) % latencyMaxPoints
	if g.count < latencyMaxPoints { g.count++ }
	g.heightDirty = true
	g.recomputeLocked()
	g.mu.Unlock()
	return g
}

// Clear resets all latency data.
func (g *ModelLatencyGraph) Clear() *ModelLatencyGraph {
	g.mu.Lock()
	g.count = 0
	g.head = 0
	g.heightDirty = true
	g.recomputeLocked()
	g.mu.Unlock()
	return g
}

func (g *ModelLatencyGraph) recomputeLocked() {
	if g.count == 0 {
		g.curStr = "--"
		g.avgStr = "--"
		g.peakStr = "--"
		g.heightsN = 0
		return
	}

	// Compute stats
	peak := 0
	sum := 0
	for i := 0; i < g.count; i++ {
		v := g.points[i]
		if v > peak { peak = v }
		sum += v
	}
	avg := sum / g.count

	// Current = most recently written = points[(head-1+max)%max]
	curIdx := (g.head - 1 + latencyMaxPoints) % latencyMaxPoints
	cur := g.points[curIdx]

	g.curStr = itoa(cur) + "ms"
	g.avgStr = itoa(avg) + "ms"
	g.peakStr = itoa(peak) + "ms"

	// Compute bar heights for rendering
	if peak == 0 { peak = 1 }
	maxH := g.height - 1
	if maxH < 1 { maxH = 1 }
	g.heightsN = g.count
	for i := 0; i < g.count; i++ {
		h := g.points[i] * maxH / peak
		if h < 1 { h = 1 }
		g.heights[i] = h
	}
}

// SetSize sets the graph dimensions.
func (g *ModelLatencyGraph) SetSize(w, h int) *ModelLatencyGraph {
	g.mu.Lock()
	if w < 10 { w = 10 }
	if h < 3 { h = 3 }
	g.width = w
	g.height = h
	g.heightDirty = true
	g.recomputeLocked()
	g.mu.Unlock()
	return g
}

// CurrentLatency returns the most recent latency in ms.
func (g *ModelLatencyGraph) CurrentLatency() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.count == 0 { return 0 }
	idx := (g.head - 1 + latencyMaxPoints) % latencyMaxPoints
	return g.points[idx]
}

// AverageLatency returns the average latency in ms.
func (g *ModelLatencyGraph) AverageLatency() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.count == 0 { return 0 }
	sum := 0
	for i := 0; i < g.count; i++ { sum += g.points[i] }
	return sum / g.count
}

// SetStyle sets custom style.
func (g *ModelLatencyGraph) SetStyle(s LatencyGraphStyle) *ModelLatencyGraph {
	g.mu.Lock()
	g.style = s
	g.mu.Unlock()
	return g
}

// Measure returns preferred size.
func (g *ModelLatencyGraph) Measure(cs Constraints) Size {
	w := g.width
	h := g.height + 1 // +1 for label row
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the latency graph.
func (g *ModelLatencyGraph) Paint(buf *buffer.Buffer) {
	g.mu.Lock()
	defer g.mu.Unlock()

	b := g.Bounds()
	x, y := b.X, b.Y
	w := b.W
	h := b.H
	if w < 10 { w = 40 }
	if h < 3 { h = 5 }

	lineStyle := g.style.Line
	fillStyle := g.style.Fill
	labelStyle := g.style.Label
	peakStyle := g.style.Peak
	curStyle := g.style.Current

	graphH := h - 1
	if graphH < 1 { graphH = 1 }

	if g.heightsN == 0 {
		// Empty state
		label := "No latency data"
		for i, r := range label {
			if x+i >= buf.Width { break }
			buf.SetCell(x+i, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		}
		return
	}

	// Determine which points to display (most recent that fit)
	nDisplay := g.heightsN
	if nDisplay > w { nDisplay = w }
	startIdx := g.heightsN - nDisplay

	// Draw bars from bottom up
	for col := 0; col < nDisplay; col++ {
		barH := g.heights[startIdx+col]
		cx := x + col
		if cx >= buf.Width { break }
		// Fill from bottom
		for row := 0; row < barH; row++ {
			yy := y + graphH - 1 - row
			if yy < 0 || yy >= buf.Height { continue }
			if yy < b.Y || yy >= b.Y+h { continue }
			var rune_ rune
			var style_ buffer.Style
			if row == barH-1 {
				rune_ = '▄'
				style_ = lineStyle
			} else {
				rune_ = '█'
				style_ = fillStyle
			}
			buf.SetCell(cx, yy, buffer.Cell{Rune: rune_, Fg: style_.Fg, Bg: style_.Bg, Flags: style_.Flags, Width: 1})
		}
	}

	// Label row at bottom
	labelY := y + graphH
	if labelY < buf.Height {
		label := "cur:" + g.curStr + " avg:" + g.avgStr + " peak:" + g.peakStr
		for i, r := range label {
			cx := x + i
			if cx >= buf.Width { break }
			var style_ buffer.Style
			if i >= 4 && i < 4+len(g.curStr) {
				style_ = curStyle
			} else if i >= len(label)-len(g.peakStr) {
				style_ = peakStyle
			} else {
				style_ = labelStyle
			}
			buf.SetCell(cx, labelY, buffer.Cell{Rune: r, Fg: style_.Fg, Bg: style_.Bg, Flags: style_.Flags, Width: 1})
		}
	}
}

// Children returns nil.
func (g *ModelLatencyGraph) Children() []Component { return nil }

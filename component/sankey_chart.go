package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── SankeyChart: Flow Diagram for Proportional Transitions ───
//
// SankeyChart renders a simplified Sankey-style flow diagram showing
// proportional transitions between source and target nodes. Useful for
// visualizing budget flows, traffic sources, and energy distribution.
//
// Usage:
//
//	sc := NewSankeyChart()
//	sc.AddFlow("Revenue", "Marketing", 500)
//	sc.AddFlow("Revenue", "Engineering", 800)
//	sc.AddFlow("Marketing", "Ads", 300)
//	sc.Paint(buf)

// SankeyFlow represents a single flow between source and target.
type SankeyFlow struct {
	Source string
	Target string
	Value  int
}

// SankeyChartStyle holds styling for SankeyChart.
type SankeyChartStyle struct {
	Node     buffer.Style
	Flow     buffer.Style
	Label    buffer.Style
	NodeWidth int
}

// DefaultSankeyChartStyle returns sensible defaults.
func DefaultSankeyChartStyle() SankeyChartStyle {
	node := buffer.Style{Fg: buffer.RGB(96, 165, 250), Bg: buffer.RGB(30, 58, 138)}   // blue-400/blue-900
	flow := buffer.Style{Fg: buffer.RGB(148, 163, 184)}                                // slate-400
	label := buffer.Style{Fg: buffer.RGB(226, 232, 240)}                               // slate-200
	return SankeyChartStyle{Node: node, Flow: flow, Label: label, NodeWidth: 3}
}

// SankeyChart renders a proportional flow diagram.
type SankeyChart struct {
	BaseComponent
	mu sync.Mutex

	flows []SankeyFlow
	style SankeyChartStyle

	// cached layout
	cachedNodes     map[string]int
	cachedSources   []string
	cachedTargets   []string
	cachedTotalVal  int
	cachedDirty     bool
	cachedSrcVals   map[string]int
	cachedTgtVals   map[string]int
}

// NewSankeyChart creates a SankeyChart with defaults.
func NewSankeyChart() *SankeyChart {
	sc := &SankeyChart{
		style: DefaultSankeyChartStyle(),
	}
	sc.SetID(GenerateID("sankey"))
	return sc
}

// AddFlow adds a flow between source and target with the given value.
func (sc *SankeyChart) AddFlow(source, target string, value int) *SankeyChart {
	sc.mu.Lock()
	sc.flows = append(sc.flows, SankeyFlow{Source: source, Target: target, Value: value})
	sc.cachedDirty = true
	sc.mu.Unlock()
	return sc
}

// SetFlows replaces all flows.
func (sc *SankeyChart) SetFlows(flows []SankeyFlow) *SankeyChart {
	sc.mu.Lock()
	sc.flows = flows
	sc.cachedDirty = true
	sc.mu.Unlock()
	return sc
}

// Flows returns the current flows.
func (sc *SankeyChart) Flows() []SankeyFlow {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	result := make([]SankeyFlow, len(sc.flows))
	copy(result, sc.flows)
	return result
}

// SetStyle sets the custom style.
func (sc *SankeyChart) SetStyle(s SankeyChartStyle) *SankeyChart {
	sc.mu.Lock()
	sc.style = s
	sc.mu.Unlock()
	return sc
}

// computeLayout builds the node positions, source/target lists, and total value.
// Results are cached for Paint to avoid allocations on every render.
func (sc *SankeyChart) computeLayout() {
	if !sc.cachedDirty && sc.cachedNodes != nil {
		return // already computed, no changes
	}
	sc.cachedDirty = false
	if sc.cachedNodes == nil {
		sc.cachedNodes = make(map[string]int)
		sc.cachedSrcVals = make(map[string]int)
		sc.cachedTgtVals = make(map[string]int)
	} else {
		for k := range sc.cachedNodes {
			delete(sc.cachedNodes, k)
		}
		for k := range sc.cachedSrcVals {
			delete(sc.cachedSrcVals, k)
		}
		for k := range sc.cachedTgtVals {
			delete(sc.cachedTgtVals, k)
		}
	}
	sc.cachedSources = sc.cachedSources[:0]
	sc.cachedTargets = sc.cachedTargets[:0]
	sc.cachedTotalVal = 0

	srcSeen := make(map[string]bool)
	tgtSeen := make(map[string]bool)

	for _, f := range sc.flows {
		sc.cachedSrcVals[f.Source] += f.Value
		sc.cachedTgtVals[f.Target] += f.Value
		if f.Value > sc.cachedTotalVal {
			sc.cachedTotalVal = f.Value
		}
		if !srcSeen[f.Source] {
			srcSeen[f.Source] = true
			sc.cachedSources = append(sc.cachedSources, f.Source)
		}
		if !tgtSeen[f.Target] {
			tgtSeen[f.Target] = true
			sc.cachedTargets = append(sc.cachedTargets, f.Target)
		}
	}
	// Merge node values
	for name, v := range sc.cachedSrcVals {
		if tv, ok := sc.cachedTgtVals[name]; ok {
			if tv > v {
				sc.cachedSrcVals[name] = tv
			}
		}
		sc.cachedNodes[name] = sc.cachedSrcVals[name]
	}
	for name, v := range sc.cachedTgtVals {
		if _, exists := sc.cachedNodes[name]; !exists {
			sc.cachedNodes[name] = v
		}
	}
}

// Sources returns the unique source node names.
func (sc *SankeyChart) Sources() []string {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.sourcesLocked()
}

// sourcesLocked returns unique sources without locking (caller must hold lock).
func (sc *SankeyChart) sourcesLocked() []string {
	seen := make(map[string]bool)
	var result []string
	for _, f := range sc.flows {
		if !seen[f.Source] {
			seen[f.Source] = true
			result = append(result, f.Source)
		}
	}
	return result
}

// Targets returns the unique target node names.
func (sc *SankeyChart) Targets() []string {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.targetsLocked()
}

// targetsLocked returns unique targets without locking (caller must hold lock).
func (sc *SankeyChart) targetsLocked() []string {
	seen := make(map[string]bool)
	var result []string
	for _, f := range sc.flows {
		if !seen[f.Target] {
			seen[f.Target] = true
			result = append(result, f.Target)
		}
	}
	return result
}

// Measure returns the preferred size.
func (sc *SankeyChart) Measure(cs Constraints) Size {
	w := 50
	h := 15
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the Sankey diagram into the buffer.
func (sc *SankeyChart) Paint(buf *buffer.Buffer) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	b := sc.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 30 {
		w = 50
	}
	if h < 10 {
		h = 15
	}

	sc.computeLayout()
	if sc.cachedTotalVal <= 0 || len(sc.cachedNodes) == 0 {
		return
	}

	// Use cached sources/targets from computeLayout (no allocation)
	sources := sc.cachedSources
	targets := sc.cachedTargets

	// Calculate node heights proportional to values
	sourceMaxVal := 0
	for _, s := range sources {
		if v := sc.cachedNodes[s]; v > sourceMaxVal {
			sourceMaxVal = v
		}
	}
	if sourceMaxVal == 0 {
		sourceMaxVal = 1
	}

	targetMaxVal := 0
	for _, t := range targets {
		if v := sc.cachedNodes[t]; v > targetMaxVal {
			targetMaxVal = v
		}
	}
	if targetMaxVal == 0 {
		targetMaxVal = 1
	}

	availH := h - 2
	if availH < 1 {
		availH = 1
	}

	// Draw source nodes (left column)
	nodeW := sc.style.NodeWidth
	if nodeW < 1 {
		nodeW = 3
	}
	leftX := x + 1
	rightX := x + w - nodeW - 2

	sourceY := y + 1
	for _, s := range sources {
		val := sc.cachedNodes[s]
		nodeH := val * availH / sourceMaxVal
		if nodeH < 1 {
			nodeH = 1
		}
		// Draw node bar
		for row := 0; row < nodeH && sourceY+row < y+h; row++ {
			for col := 0; col < nodeW && leftX+col < buf.Width; col++ {
				buf.SetCell(leftX+col, sourceY+row, buffer.Cell{
					Rune:   '█',
					Fg:     sc.style.Node.Fg,
					Bg:     sc.style.Node.Bg,
					Flags:  sc.style.Node.Flags,
					Width:  1,
				})
			}
		}
		// Draw label
		lblX := leftX - 1
		for i := len(s) - 1; i >= 0 && lblX-i+ len(s) < buf.Width; i-- {
			_ = i
		}
		labelX := leftX + nodeW + 1
		for i, r := range s {
			if labelX+i < buf.Width && labelX+i < x+w-1 {
				buf.SetCell(labelX+i, sourceY, buffer.Cell{
					Rune:  r,
					Fg:    sc.style.Label.Fg,
					Bg:    sc.style.Label.Bg,
					Flags: sc.style.Label.Flags,
					Width: 1,
				})
			}
		}
		sourceY += nodeH + 1
	}

	// Draw target nodes (right column)
	targetY := y + 1
	for _, tg := range targets {
		val := sc.cachedNodes[tg]
		nodeH := val * availH / targetMaxVal
		if nodeH < 1 {
			nodeH = 1
		}
		for row := 0; row < nodeH && targetY+row < y+h; row++ {
			for col := 0; col < nodeW && rightX+col < buf.Width; col++ {
				buf.SetCell(rightX+col, targetY+row, buffer.Cell{
					Rune:  '█',
					Fg:    sc.style.Node.Fg,
					Bg:    sc.style.Node.Bg,
					Flags: sc.style.Node.Flags,
					Width: 1,
				})
			}
		}
		// Draw label to the right of node
		labelX := rightX + nodeW + 1
		for i, r := range tg {
			if labelX+i < buf.Width {
				buf.SetCell(labelX+i, targetY, buffer.Cell{
					Rune:  r,
					Fg:    sc.style.Label.Fg,
					Bg:    sc.style.Label.Bg,
					Flags: sc.style.Label.Flags,
					Width: 1,
				})
			}
		}
		targetY += nodeH + 1
	}

	// Draw flow lines between source and target
	// Simple horizontal connectors proportional to flow value
	srcYMap := make(map[string]int)
	srcY := y + 1
	for _, s := range sources {
		val := sc.cachedNodes[s]
		nodeH := val * availH / sourceMaxVal
		if nodeH < 1 {
			nodeH = 1
		}
		srcYMap[s] = srcY
		srcY += nodeH + 1
	}

	tgtYMap := make(map[string]int)
	tgtY := y + 1
	for _, t := range targets {
		val := sc.cachedNodes[t]
		nodeH := val * availH / targetMaxVal
		if nodeH < 1 {
			nodeH = 1
		}
		tgtYMap[t] = tgtY
		tgtY += nodeH + 1
	}

	flowStartX := leftX + nodeW + 1
	flowEndX := rightX
	for _, f := range sc.flows {
		srcY, ok1 := srcYMap[f.Source]
		tgtY, ok2 := tgtYMap[f.Target]
		if !ok1 || !ok2 {
			continue
		}
		flowH := f.Value * availH / sourceMaxVal
		if flowH < 1 {
			flowH = 1
		}
		// Draw horizontal line at midpoints
		for col := flowStartX; col < flowEndX && col < buf.Width; col++ {
			midY := (srcY + tgtY) / 2
			if midY < y+h && midY < buf.Height {
				buf.SetCell(col, midY, buffer.Cell{
					Rune:  '·',
					Fg:    sc.style.Flow.Fg,
					Bg:    sc.style.Flow.Bg,
					Flags: sc.style.Flow.Flags,
					Width: 1,
				})
			}
		}
	}
}

// Children returns nil.
func (sc *SankeyChart) Children() []Component { return nil }

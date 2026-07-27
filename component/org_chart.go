package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── OrgChart: Hierarchical Organization Tree ───
//
// OrgChart renders a tree of organizational/reporting relationships with
// connector lines. Each node shows a label, positioned hierarchically with
// children below their parent.
//
// Usage:
//
//	oc := NewOrgChart()
//	oc.SetRoot(OrgNode{ID: "ceo", Label: "CEO"})
//	oc.AddChild("ceo", OrgNode{ID: "cto", Label: "CTO"})
//	oc.AddChild("ceo", OrgNode{ID: "cfo", Label: "CFO"})
//	oc.AddChild("cto", OrgNode{ID: "dev", Label: "Dev Team"})
//	oc.SetBounds(Rect{X:0, Y:0, W:60, H:15})
//	oc.Paint(buf)

// OrgNode represents a node in the organization tree.
type OrgNode struct {
	ID     string
	Label  string
	Color  buffer.Color
}

// OrgChartStyle holds visual styles.
type OrgChartStyle struct {
	Node      buffer.Style
	Connector buffer.Style
	Label     buffer.Style
}

// DefaultOrgChartStyle returns sensible defaults.
func DefaultOrgChartStyle() OrgChartStyle {
	return OrgChartStyle{
		Node:      buffer.Style{Fg: buffer.RGB(100, 149, 237), Flags: buffer.Bold},
		Connector: buffer.Style{Fg: buffer.RGB(80, 80, 80)},
		Label:     buffer.Style{Fg: buffer.White},
	}
}

// orgTreeNode is the internal tree representation.
type orgTreeNode struct {
	node     OrgNode
	children []*orgTreeNode
	parent   *orgTreeNode
}

// OrgChart renders a hierarchical organization tree.
type OrgChart struct {
	BaseComponent
	mu       sync.RWMutex
	root     *orgTreeNode
	nodeMap  map[string]*orgTreeNode
	style    OrgChartStyle
}

// NewOrgChart creates an empty org chart.
func NewOrgChart() *OrgChart {
	oc := &OrgChart{
		nodeMap: make(map[string]*orgTreeNode),
		style:   DefaultOrgChartStyle(),
	}
	oc.SetID(GenerateID("orgchart"))
	return oc
}

// SetRoot sets the root node.
func (oc *OrgChart) SetRoot(node OrgNode) *OrgChart {
	oc.mu.Lock()
	defer oc.mu.Unlock()
	if node.Color.Type == 0 {
		node.Color = buffer.RGB(100, 149, 237)
	}
	oc.root = &orgTreeNode{node: node}
	oc.nodeMap = map[string]*orgTreeNode{node.ID: oc.root}
	return oc
}

// AddChild adds a child node to the specified parent.
func (oc *OrgChart) AddChild(parentID string, node OrgNode) *OrgChart {
	oc.mu.Lock()
	defer oc.mu.Unlock()
	parent, ok := oc.nodeMap[parentID]
	if !ok {
		return oc
	}
	if node.Color.Type == 0 {
		node.Color = buffer.RGB(100, 149, 237)
	}
	child := &orgTreeNode{node: node, parent: parent}
	parent.children = append(parent.children, child)
	oc.nodeMap[node.ID] = child
	return oc
}

// NodeCount returns the total number of nodes.
func (oc *OrgChart) NodeCount() int {
	oc.mu.RLock()
	defer oc.mu.RUnlock()
	return len(oc.nodeMap)
}

// HasRoot returns true if a root node is set.
func (oc *OrgChart) HasRoot() bool {
	oc.mu.RLock()
	defer oc.mu.RUnlock()
	return oc.root != nil
}

// Depth returns the tree depth (0 if empty, 1 for root only).
func (oc *OrgChart) Depth() int {
	oc.mu.RLock()
	defer oc.mu.RUnlock()
	if oc.root == nil {
		return 0
	}
	return orgTreeDepthLocked(oc.root)
}

func orgTreeDepthLocked(n *orgTreeNode) int {
	if n == nil {
		return 0
	}
	maxChild := 0
	for _, c := range n.children {
		d := orgTreeDepthLocked(c)
		if d > maxChild {
			maxChild = d
		}
	}
	return 1 + maxChild
}

// SetStyle sets the visual style.
func (oc *OrgChart) SetStyle(s OrgChartStyle) *OrgChart {
	oc.mu.Lock()
	oc.style = s
	oc.mu.Unlock()
	return oc
}

// Style returns the current style.
func (oc *OrgChart) Style() OrgChartStyle {
	oc.mu.RLock()
	defer oc.mu.RUnlock()
	return oc.style
}

// Measure computes the desired size.
func (oc *OrgChart) Measure(cs Constraints) Size {
	oc.mu.RLock()
	defer oc.mu.RUnlock()
	w := oc.NodeCount()*8 + 4
	if w < 30 {
		w = 30
	}
	d := oc.Depth()
	h := d*3 + 1
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

// Paint renders the org chart.
func (oc *OrgChart) Paint(buf *buffer.Buffer) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	b := oc.bounds
	if b.W < 6 || b.H < 3 || oc.root == nil {
		return
	}

	// Layout: assign positions to nodes using a simple level-based layout
	type layoutEntry struct {
		node *orgTreeNode
		x    int
		y    int
	}
	levels := [][]*orgTreeNode{}
	queue := []*orgTreeNode{oc.root}
	for len(queue) > 0 {
		levels = append(levels, queue)
		var next []*orgTreeNode
		for _, n := range queue {
			next = append(next, n.children...)
		}
		queue = next
	}

	entries := make(map[string]layoutEntry)
	rowH := b.H / len(levels)
	if rowH < 2 {
		rowH = 2
	}

	for level, nodes := range levels {
		y := b.Y + level*rowH
		if y >= b.Y+b.H {
			break
		}
		n := len(nodes)
		for i, node := range nodes {
			x := b.X + (b.W/n)*i + (b.W/n)/2 - 4
			if x < b.X {
				x = b.X
			}
			entries[node.node.ID] = layoutEntry{node: node, x: x, y: y}
		}
	}

	// Draw connectors
	for _, entry := range entries {
		for _, child := range entry.node.children {
			childEntry, ok := entries[child.node.ID]
			if !ok {
				continue
			}
			// Vertical line from parent down
			for y := entry.y + 1; y < childEntry.y; y++ {
				if y < b.Y+b.H {
					buf.SetCell(entry.x+3, y, buffer.Cell{Rune: '│', Fg: oc.style.Connector.Fg, Bg: oc.style.Connector.Bg, Width: 1})
				}
			}
			// Horizontal connector if needed
			if childEntry.x+3 != entry.x+3 {
				minX := entry.x + 3
				maxX := childEntry.x + 3
				if minX > maxX {
					minX, maxX = maxX, minX
				}
				y := childEntry.y - 1
				if y >= b.Y && y < b.Y+b.H {
					for x := minX; x <= maxX; x++ {
						if x >= b.X && x < b.X+b.W {
							buf.SetCell(x, y, buffer.Cell{Rune: '─', Fg: oc.style.Connector.Fg, Bg: oc.style.Connector.Bg, Width: 1})
						}
					}
				}
			}
		}
	}

	// Draw nodes
	for _, entry := range entries {
		node := entry.node.node
		x := entry.x
		y := entry.y
		if x < b.X {
			x = b.X
		}
		// Draw box border
		boxW := 8
		if x+boxW > b.X+b.W {
			boxW = b.X + b.W - x
		}
		if boxW < 2 {
			continue
		}
		for i := 0; i < boxW; i++ {
			buf.SetCell(x+i, y, buffer.Cell{Rune: '─', Fg: node.Color, Bg: oc.style.Node.Bg, Width: 1})
		}
		// Label
		labelRunes := []rune(node.Label)
		for i, r := range labelRunes {
			if i >= boxW-1 || x+1+i >= b.X+b.W {
				break
			}
			buf.SetCell(x+1+i, y, buffer.Cell{Rune: r, Fg: oc.style.Label.Fg, Bg: node.Color, Flags: oc.style.Label.Flags, Width: 1})
		}
	}
}

// Children returns nil.
func (oc *OrgChart) Children() []Component { return nil }

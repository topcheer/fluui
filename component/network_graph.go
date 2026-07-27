package component

import (
	"math"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── NetworkGraph: Node/Edge Topology Visualization ───
//
// NetworkGraph renders a graph of nodes connected by edges. Useful for
// network topology, dependency graphs, and infrastructure maps.
// Nodes are positioned with a simple circular layout.
//
// Usage:
//
//	ng := NewNetworkGraph()
//	ng.AddNode(GraphNode{ID: "a", Label: "Server"})
//	ng.AddNode(GraphNode{ID: "b", Label: "DB"})
//	ng.AddEdge(GraphEdge{From: "a", To: "b"})
//	ng.SetBounds(Rect{X:0, Y:0, W:50, H:20})
//	ng.Paint(buf)

// GraphNode represents a node in the network graph.
type GraphNode struct {
	ID     string
	Label  string
	Color  buffer.Color
}

// GraphEdge represents a connection between two nodes.
type GraphEdge struct {
	From string
	To   string
}

// NetworkGraphStyle holds visual styles.
type NetworkGraphStyle struct {
	Node     buffer.Style
	Edge     buffer.Style
	Label    buffer.Style
}

// DefaultNetworkGraphStyle returns sensible defaults.
func DefaultNetworkGraphStyle() NetworkGraphStyle {
	return NetworkGraphStyle{
		Node:  buffer.Style{Fg: buffer.RGB(100, 149, 237), Flags: buffer.Bold},
		Edge:  buffer.Style{Fg: buffer.RGB(80, 80, 80)},
		Label: buffer.Style{Fg: buffer.White},
	}
}

// nodePos holds a computed screen position.
type nodePos struct {
	x, y int
}

// NetworkGraph renders a node/edge topology.
type NetworkGraph struct {
	BaseComponent
	mu    sync.RWMutex
	nodes []GraphNode
	edges []GraphEdge
	style NetworkGraphStyle
	// Cached positions (reused across Paint calls)
	cachedPos   map[string]nodePos
	cachedNodes int
	cachedW     int
	cachedH     int
}

// NewNetworkGraph creates an empty graph.
func NewNetworkGraph() *NetworkGraph {
	ng := &NetworkGraph{
		style: DefaultNetworkGraphStyle(),
	}
	ng.SetID(GenerateID("netgraph"))
	return ng
}

// AddNode adds a node.
func (ng *NetworkGraph) AddNode(n GraphNode) *NetworkGraph {
	ng.mu.Lock()
	if n.Color.Type == 0 {
		n.Color = buffer.RGB(100, 149, 237)
	}
	ng.nodes = append(ng.nodes, n)
	ng.mu.Unlock()
	return ng
}

// AddEdge adds an edge between two node IDs.
func (ng *NetworkGraph) AddEdge(e GraphEdge) *NetworkGraph {
	ng.mu.Lock()
	ng.edges = append(ng.edges, e)
	ng.mu.Unlock()
	return ng
}

// SetNodes replaces all nodes.
func (ng *NetworkGraph) SetNodes(nodes []GraphNode) *NetworkGraph {
	ng.mu.Lock()
	ng.nodes = nodes
	ng.mu.Unlock()
	return ng
}

// SetEdges replaces all edges.
func (ng *NetworkGraph) SetEdges(edges []GraphEdge) *NetworkGraph {
	ng.mu.Lock()
	ng.edges = edges
	ng.mu.Unlock()
	return ng
}

// Nodes returns the current nodes.
func (ng *NetworkGraph) Nodes() []GraphNode {
	ng.mu.RLock()
	defer ng.mu.RUnlock()
	return ng.nodes
}

// Edges returns the current edges.
func (ng *NetworkGraph) Edges() []GraphEdge {
	ng.mu.RLock()
	defer ng.mu.RUnlock()
	return ng.edges
}

// NodeCount returns the number of nodes.
func (ng *NetworkGraph) NodeCount() int {
	ng.mu.RLock()
	defer ng.mu.RUnlock()
	return len(ng.nodes)
}

// EdgeCount returns the number of edges.
func (ng *NetworkGraph) EdgeCount() int {
	ng.mu.RLock()
	defer ng.mu.RUnlock()
	return len(ng.edges)
}

// Clear removes all nodes and edges.
func (ng *NetworkGraph) Clear() *NetworkGraph {
	ng.mu.Lock()
	ng.nodes = ng.nodes[:0]
	ng.edges = ng.edges[:0]
	ng.mu.Unlock()
	return ng
}

// SetStyle sets the visual style.
func (ng *NetworkGraph) SetStyle(s NetworkGraphStyle) *NetworkGraph {
	ng.mu.Lock()
	ng.style = s
	ng.mu.Unlock()
	return ng
}

// Style returns the current style.
func (ng *NetworkGraph) Style() NetworkGraphStyle {
	ng.mu.RLock()
	defer ng.mu.RUnlock()
	return ng.style
}

// layoutNodes computes circular positions for all nodes (caller holds lock).
func (ng *NetworkGraph) layoutNodesLocked(b Rect) map[string]nodePos {
	n := len(ng.nodes)
	// Reuse cached map if node count and dimensions haven't changed
	if ng.cachedPos != nil && ng.cachedNodes == n && ng.cachedW == b.W && ng.cachedH == b.H {
		return ng.cachedPos
	}
	positions := make(map[string]nodePos, n)
	if n == 0 {
		return positions
	}

	cx := b.X + b.W/2
	cy := b.Y + b.H/2
	radius := b.W
	if b.H < radius {
		radius = b.H
	}
	radius = radius/2 - 2
	if radius < 2 {
		radius = 2
	}

	for i, node := range ng.nodes {
		angle := 2 * math.Pi * float64(i) / float64(n)
		x := cx + int(float64(radius)*math.Cos(angle))
		y := cy + int(float64(radius)*math.Sin(angle)*0.5) // squash vertically
		positions[node.ID] = nodePos{x: x, y: y}
	}
	// Cache for reuse
	ng.cachedPos = positions
	ng.cachedNodes = n
	ng.cachedW = b.W
	ng.cachedH = b.H
	return positions
}

// Measure computes the desired size.
func (ng *NetworkGraph) Measure(cs Constraints) Size {
	w, h := 40, 20
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the network graph.
func (ng *NetworkGraph) Paint(buf *buffer.Buffer) {
	ng.mu.Lock()
	defer ng.mu.Unlock()

	b := ng.bounds
	if b.W < 6 || b.H < 4 || len(ng.nodes) == 0 {
		return
	}

	positions := ng.layoutNodesLocked(b)

	// Draw edges first (so nodes are on top)
	for _, edge := range ng.edges {
		from, ok1 := positions[edge.From]
		to, ok2 := positions[edge.To]
		if !ok1 || !ok2 {
			continue
		}
		drawLineBresenham(buf, from.x, from.y, to.x, to.y, ng.style.Edge)
	}

	// Draw nodes
	for _, node := range ng.nodes {
		pos, ok := positions[node.ID]
		if !ok {
			continue
		}
		// Node glyph
		buf.SetCell(pos.x, pos.y, buffer.Cell{
			Rune:  '◉',
			Fg:    node.Color,
			Bg:    ng.style.Node.Bg,
			Flags: ng.style.Node.Flags,
			Width: 1,
		})
		// Label to the right
		labelX := pos.x + 2
		for _, r := range node.Label {
			if labelX >= b.X+b.W {
				break
			}
			buf.SetCell(labelX, pos.y, buffer.Cell{
				Rune:  r,
				Fg:    ng.style.Label.Fg,
				Bg:    ng.style.Label.Bg,
				Flags: ng.style.Label.Flags,
				Width: 1,
			})
			labelX++
		}
	}
}

// drawLineBresenham draws a line between two points using Bresenham's algorithm.
func drawLineBresenham(buf *buffer.Buffer, x0, y0, x1, y1 int, style buffer.Style) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	if x0 >= x1 {
		sx = -1
	}
	sy := 1
	if y0 >= y1 {
		sy = -1
	}
	err := dx - dy
	for {
		buf.SetCell(x0, y0, buffer.Cell{Rune: '·', Fg: style.Fg, Bg: style.Bg, Width: 1})
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// Children returns nil.
func (ng *NetworkGraph) Children() []Component { return nil }

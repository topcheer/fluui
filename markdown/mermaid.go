package markdown

import (
	"strings"

	"github.com/topcheer/fluui/internal/buffer"
)

// ---------------------------------------------------------------------------
// Mermaid Flowchart Renderer
// ---------------------------------------------------------------------------
//
// Parses a subset of Mermaid flowchart syntax and renders it as ASCII art.
// Supported features:
//   - flowchart TD/TB/LR/BT/RL direction
//   - Node shapes: [rect], (rounded), {diamond}, ((circle))
//   - Edges: -->, ---, -.->, ==>
//   - Edge labels: -->|text|
//   - Node IDs with labels: A[Label] --> B[Label]
//
// This is a pure Go implementation — no external dependencies.

// MermaidDirection specifies the flow direction of a Mermaid diagram.
type MermaidDirection uint8

const (
	MermaidTD MermaidDirection = iota // top-down (default)
	MermaidLR                        // left-right
	MermaidBT                        // bottom-top
	MermaidRL                        // right-left
)

// MermaidNodeShape identifies the visual shape of a node.
type MermaidNodeShape uint8

const (
	MermaidShapeRect     MermaidNodeShape = iota // [text]
	MermaidShapeRounded                          // (text)
	MermaidShapeDiamond                          // {text}
	MermaidShapeCircle                           // ((text))
	MermaidShapePlain                            // just text or ID
)

// MermaidNode is a single node in the flowchart.
type MermaidNode struct {
	ID    string
	Label string
	Shape MermaidNodeShape
	X     int // computed position (top-left)
	Y     int
	W     int // computed size
	H     int
}

// MermaidEdgeStyle identifies the style of an edge.
type MermaidEdgeStyle uint8

const (
	MermaidEdgeArrow  MermaidEdgeStyle = iota // -->
	MermaidEdgeLine                           // ---
	MermaidEdgeDotted                         // -.->
	MermaidEdgeThick                          // ==>
)

// MermaidEdge is a connection between two nodes.
type MermaidEdge struct {
	From  string
	To    string
	Label string
	Style MermaidEdgeStyle
}

// MermaidGraph is a parsed Mermaid flowchart.
type MermaidGraph struct {
	Direction MermaidDirection
	Nodes     []*MermaidNode
	Edges     []*MermaidEdge
}

// ParseMermaid parses Mermaid flowchart text into a graph structure.
// Returns nil if the input is not a valid Mermaid flowchart.
func ParseMermaid(text string) *MermaidGraph {
	lines := strings.Split(text, "\n")
	g := &MermaidGraph{Direction: MermaidTD}
	nodeMap := make(map[string]*MermaidNode)

	parsed := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// First non-empty line must be "flowchart <dir>" or "graph <dir>"
		if !parsed {
			lower := strings.ToLower(line)
			if strings.HasPrefix(lower, "flowchart") || strings.HasPrefix(lower, "graph") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					switch strings.ToUpper(parts[1]) {
					case "LR":
						g.Direction = MermaidLR
					case "BT":
						g.Direction = MermaidBT
					case "RL":
						g.Direction = MermaidRL
					default:
						g.Direction = MermaidTD
					}
				}
				parsed = true
				continue
			}
			// Not a flowchart
			return nil
		}

		// Parse edge/node lines
		parseMermaidLine(line, g, nodeMap)
	}

	if !parsed || len(g.Nodes) == 0 {
		return nil
	}

	return g
}

// parseMermaidLine parses a single line of Mermaid flowchart syntax.
func parseMermaidLine(line string, g *MermaidGraph, nodeMap map[string]*MermaidNode) {
	// Split by edge patterns
	// Try to match edge patterns first
	edgePatterns := []struct {
		sep    string
		style  MermaidEdgeStyle
	}{
		{"==>", MermaidEdgeThick},
		{"-.->", MermaidEdgeDotted},
		{"-->", MermaidEdgeArrow},
		{"---", MermaidEdgeLine},
	}

	for _, ep := range edgePatterns {
		if idx := indexOf(line, ep.sep); idx >= 0 {
			leftPart := strings.TrimSpace(line[:idx])
			rest := line[idx+len(ep.sep):]

			// Check for edge label: |label|
			var label string
			rest = strings.TrimSpace(rest)
			if strings.HasPrefix(rest, "|") {
				endIdx := indexOf(rest[1:], "|")
				if endIdx >= 0 {
					label = strings.TrimSpace(rest[1 : 1+endIdx])
					rest = strings.TrimSpace(rest[2+endIdx:])
				}
			}

			// Parse left and right nodes
			fromNode := parseMermaidNode(leftPart, g, nodeMap)
			toNode := parseMermaidNode(rest, g, nodeMap)

			if fromNode != nil && toNode != nil {
				g.Edges = append(g.Edges, &MermaidEdge{
					From:  fromNode.ID,
					To:    toNode.ID,
					Label: label,
					Style: ep.style,
				})
			}
			return
		}
	}

	// No edge found — might be a standalone node definition
	parseMermaidNode(line, g, nodeMap)
}

// parseMermaidNode parses a node definition like "A[Label]" or just "A".
func parseMermaidNode(text string, g *MermaidGraph, nodeMap map[string]*MermaidNode) *MermaidNode {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	id, label, shape := parseMermaidNodeSpec(text)
	if id == "" {
		return nil
	}

	// Check if node already exists
	if n, ok := nodeMap[id]; ok {
		// Update label/shape if provided
		if label != "" {
			n.Label = label
		}
		if shape != MermaidShapePlain {
			n.Shape = shape
		}
		return n
	}

	// Create new node
	n := &MermaidNode{
		ID:    id,
		Label: label,
		Shape: shape,
	}
	if n.Label == "" {
		n.Label = id
	}
	nodeMap[id] = n
	g.Nodes = append(g.Nodes, n)
	return n
}

// parseMermaidNodeSpec parses a node spec string into ID, label, and shape.
// Examples:
//   "A"           → ("A", "", Plain)
//   "A[Hello]"    → ("A", "Hello", Rect)
//   "A(World)"    → ("A", "World", Rounded)
//   "A{Choice}"   → ("A", "Choice", Diamond)
//   "A((Start))"  → ("A", "Start", Circle)
func parseMermaidNodeSpec(text string) (id, label string, shape MermaidNodeShape) {
	shape = MermaidShapePlain

	// Find the first non-alphanumeric character
	i := 0
	for i < len(text) {
		c := text[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			i++
		} else {
			break
		}
	}

	id = text[:i]
	rest := text[i:]

	if rest == "" {
		return id, "", MermaidShapePlain
	}

	// Parse shape from rest
	if len(rest) >= 2 && rest[0] == '[' && rest[len(rest)-1] == ']' {
		return id, rest[1 : len(rest)-1], MermaidShapeRect
	}
	if len(rest) >= 2 && rest[0] == '(' && rest[len(rest)-1] == ')' {
		// Check for double parens ((text))
		if len(rest) >= 4 && rest[1] == '(' && rest[len(rest)-2] == ')' {
			return id, rest[2 : len(rest)-2], MermaidShapeCircle
		}
		return id, rest[1 : len(rest)-1], MermaidShapeRounded
	}
	if len(rest) >= 2 && rest[0] == '{' && rest[len(rest)-1] == '}' {
		return id, rest[1 : len(rest)-1], MermaidShapeDiamond
	}

	// Unknown shape — treat as plain with label
	if len(rest) > 0 {
		return id, strings.TrimSpace(rest), MermaidShapePlain
	}
	return id, "", MermaidShapePlain
}

// indexOf finds the first occurrence of substr in s, returning -1 if not found.
func indexOf(s, substr string) int {
	return strings.Index(s, substr)
}

// RenderMermaid converts a MermaidGraph to ASCII art cells.
// Returns nil if the graph cannot be rendered.
func RenderMermaid(g *MermaidGraph, theme *MarkdownTheme) [][]buffer.Cell {
	if g == nil || len(g.Nodes) == 0 {
		return nil
	}

	// Compute node sizes
	for _, n := range g.Nodes {
		n.W = mermaidNodeWidth(n)
		n.H = mermaidNodeHeight(n)
	}

	// Layout nodes based on direction
	mermaidLayout(g)

	// Compute canvas size
	maxX, maxY := 0, 0
	for _, n := range g.Nodes {
		if n.X+n.W > maxX {
			maxX = n.X + n.W
		}
		if n.Y+n.H > maxY {
			maxY = n.Y + n.H
		}
	}
	maxX += 4 // padding for arrows
	maxY += 2

	// Render to canvas
	canvas := make([][]byte, maxY)
	for i := range canvas {
		canvas[i] = make([]byte, maxX)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	// Draw edges first (so nodes overlap them)
	for _, e := range g.Edges {
		from := findMermaidNode(g, e.From)
		to := findMermaidNode(g, e.To)
		if from == nil || to == nil {
			continue
		}
		mermaidDrawEdge(canvas, from, to, e, g.Direction)
	}

	// Draw nodes
	for _, n := range g.Nodes {
		mermaidDrawNode(canvas, n)
	}

	// Convert canvas to cells
	fg := buffer.NamedColor(buffer.NamedCyan)
	if theme != nil {
		fg = theme.CodeFg
	}

	var cells [][]buffer.Cell
	for _, row := range canvas {
		cellRow := make([]buffer.Cell, len(row))
		for i, ch := range row {
			if ch == 0 || ch == ' ' {
				cellRow[i] = buffer.BlankCell
			} else {
				cellRow[i] = buffer.NewCell(rune(ch), buffer.Style{Fg: fg})
			}
		}
		cells = append(cells, cellRow)
	}

	return cells
}

// mermaidNodeWidth computes the width of a node including border.
func mermaidNodeWidth(n *MermaidNode) int {
	w := len(n.Label)
	switch n.Shape {
	case MermaidShapeRect, MermaidShapeRounded:
		return w + 4 // padding + borders
	case MermaidShapeDiamond:
		return w + 6 // < label >
	case MermaidShapeCircle:
		return w + 6 // (( label ))
	default:
		return w
	}
}

// mermaidNodeHeight computes the height of a node.
func mermaidNodeHeight(n *MermaidNode) int {
	return 3 // all shapes are 3 rows tall (border, content, border)
}

// mermaidLayout positions nodes in a simple layered layout.
func mermaidLayout(g *MermaidGraph) {
	// Build adjacency for topological-ish layout
	// Simple approach: group nodes by their depth from roots
	depth := make(map[string]int)
	for _, n := range g.Nodes {
		depth[n.ID] = 0
	}

	// Compute depth via edges (max iteration to handle cycles)
	for iter := 0; iter < len(g.Nodes); iter++ {
		for _, e := range g.Edges {
			if depth[e.To] < depth[e.From]+1 {
				depth[e.To] = depth[e.From] + 1
			}
		}
	}

	// Group nodes by depth
	layers := make(map[int][]*MermaidNode)
	maxDepth := 0
	for _, n := range g.Nodes {
		d := depth[n.ID]
		layers[d] = append(layers[d], n)
		if d > maxDepth {
			maxDepth = d
		}
	}

	// Position nodes
	if g.Direction == MermaidTD || g.Direction == MermaidBT {
		// Vertical layout: layers go top to bottom
		y := 0
		for d := 0; d <= maxDepth; d++ {
			nodes := layers[d]
			x := 0
			for _, n := range nodes {
				n.X = x
				n.Y = y
				x += n.W + 4 // spacing between nodes in same layer
			}
			if len(nodes) > 0 {
				y += 5 // row height (node height + arrow space)
			}
		}
		// Reverse for BT
		if g.Direction == MermaidBT && maxDepth > 0 {
			maxY := 0
			for _, n := range g.Nodes {
				if n.Y+n.H > maxY {
					maxY = n.Y + n.H
				}
			}
			for _, n := range g.Nodes {
				n.Y = maxY - n.Y - n.H
			}
		}
	} else {
		// Horizontal layout (LR/RL): layers go left to right
		x := 0
		for d := 0; d <= maxDepth; d++ {
			nodes := layers[d]
			y := 0
			for _, n := range nodes {
				n.X = x
				n.Y = y
				y += n.H + 2
			}
			if len(nodes) > 0 {
				// Find max width in this layer
				maxW := 0
				for _, n := range nodes {
					if n.W > maxW {
						maxW = n.W
					}
				}
				x += maxW + 4
			}
		}
		// Reverse for RL
		if g.Direction == MermaidRL && maxDepth > 0 {
			maxX := 0
			for _, n := range g.Nodes {
				if n.X+n.W > maxX {
					maxX = n.X + n.W
				}
			}
			for _, n := range g.Nodes {
				n.X = maxX - n.X - n.W
			}
		}
	}
}

func findMermaidNode(g *MermaidGraph, id string) *MermaidNode {
	for _, n := range g.Nodes {
		if n.ID == id {
			return n
		}
	}
	return nil
}

// mermaidDrawNode draws a node onto the canvas.
func mermaidDrawNode(canvas [][]byte, n *MermaidNode) {
	if len(canvas) == 0 {
		return
	}

	switch n.Shape {
	case MermaidShapeRect:
		mermaidDrawRect(canvas, n.X, n.Y, n.W, n.H, n.Label)
	case MermaidShapeRounded:
		mermaidDrawRounded(canvas, n.X, n.Y, n.W, n.H, n.Label)
	case MermaidShapeDiamond:
		mermaidDrawDiamond(canvas, n.X, n.Y, n.W, n.H, n.Label)
	case MermaidShapeCircle:
		mermaidDrawCircle(canvas, n.X, n.Y, n.W, n.H, n.Label)
	default:
		mermaidDrawText(canvas, n.X, n.Y+1, n.Label)
	}
}

func mermaidDrawRect(canvas [][]byte, x, y, w, h int, label string) {
	for i := 0; i < w; i++ {
		setCanvasByte(canvas, x+i, y, '-')
		setCanvasByte(canvas, x+i, y+h-1, '-')
	}
	setCanvasByte(canvas, x, y, '+')
	setCanvasByte(canvas, x+w-1, y, '+')
	setCanvasByte(canvas, x, y+h-1, '+')
	setCanvasByte(canvas, x+w-1, y+h-1, '+')
	for j := 1; j < h-1; j++ {
		setCanvasByte(canvas, x, y+j, '|')
		setCanvasByte(canvas, x+w-1, y+j, '|')
	}
	mermaidDrawText(canvas, x+2, y+1, label)
}

func mermaidDrawRounded(canvas [][]byte, x, y, w, h int, label string) {
	// Rounded corners using / and \
	setCanvasByte(canvas, x, y, '.')
	setCanvasByte(canvas, x+w-1, y, '.')
	setCanvasByte(canvas, x, y+h-1, '\'')
	setCanvasByte(canvas, x+w-1, y+h-1, '\'')
	for i := 1; i < w-1; i++ {
		setCanvasByte(canvas, x+i, y, '-')
		setCanvasByte(canvas, x+i, y+h-1, '-')
	}
	for j := 1; j < h-1; j++ {
		setCanvasByte(canvas, x, y+j, '|')
		setCanvasByte(canvas, x+w-1, y+j, '|')
	}
	mermaidDrawText(canvas, x+2, y+1, label)
}

func mermaidDrawDiamond(canvas [][]byte, x, y, w, h int, label string) {
	midX := x + w/2
	// Top
	setCanvasByte(canvas, midX, y, '<')
	setCanvasByte(canvas, midX+1, y, '-')
	// Top edges
	for i := 1; i < w/2; i++ {
		setCanvasByte(canvas, midX+1+i, y, '-')
		setCanvasByte(canvas, midX-1-i, y, '-')
	}
	// Actually draw a diamond shape
	// Simple approach: draw corners
	setCanvasByte(canvas, midX, y, '^')
	setCanvasByte(canvas, midX, y+h-1, 'v')
	setCanvasByte(canvas, x, y+h/2, '<')
	setCanvasByte(canvas, x+w-1, y+h/2, '>')
	// Draw edges
	for i := 1; i < w/2-1; i++ {
		setCanvasByte(canvas, midX-i, y+i, '/')
		setCanvasByte(canvas, midX+i, y+i, '\\')
		setCanvasByte(canvas, midX-i, y+h-1-i, '\\')
		setCanvasByte(canvas, midX+i, y+h-1-i, '/')
	}
	mermaidDrawText(canvas, x+w/2-len(label)/2, y+h/2, label)
}

func mermaidDrawCircle(canvas [][]byte, x, y, w, h int, label string) {
	// Simple circle representation
	midX := x + w/2
	setCanvasByte(canvas, midX, y, '_')
	setCanvasByte(canvas, midX, y+h-1, '_')
	setCanvasByte(canvas, x+1, y, '/')
	setCanvasByte(canvas, x+w-2, y, '\\')
	setCanvasByte(canvas, x+1, y+h-1, '\\')
	setCanvasByte(canvas, x+w-2, y+h-1, '/')
	setCanvasByte(canvas, x, y+1, '(')
	setCanvasByte(canvas, x+w-1, y+1, ')')
	mermaidDrawText(canvas, x+w/2-len(label)/2, y+1, label)
}

func mermaidDrawText(canvas [][]byte, x, y int, text string) {
	for i := 0; i < len(text); i++ {
		setCanvasByte(canvas, x+i, y, text[i])
	}
}

func setCanvasByte(canvas [][]byte, x, y int, b byte) {
	if y >= 0 && y < len(canvas) && x >= 0 && x < len(canvas[y]) {
		canvas[y][x] = b
	}
}

// mermaidDrawEdge draws an edge (arrow/line) between two nodes.
func mermaidDrawEdge(canvas [][]byte, from, to *MermaidNode, edge *MermaidEdge, dir MermaidDirection) {
	// Compute connection points based on direction
	var fromX, fromY, toX, toY int

	if dir == MermaidTD || dir == MermaidBT {
		// Vertical: connect bottom of from to top of to
		fromX = from.X + from.W/2
		fromY = from.Y + from.H
		toX = to.X + to.W/2
		toY = to.Y - 1
	} else {
		// Horizontal: connect right of from to left of to
		fromX = from.X + from.W
		fromY = from.Y + from.H/2
		toX = to.X - 1
		toY = to.Y + to.H/2
	}

	// Draw line/arrow
	char := '-'
	if dir == MermaidTD || dir == MermaidBT {
		char = '|'
	}

	switch edge.Style {
	case MermaidEdgeDotted:
		char = '.' // dotted
	case MermaidEdgeThick:
		char = '='
	}

	// Simple L-shaped or straight connector
	if dir == MermaidTD || dir == MermaidBT {
		// Vertical: straight down/up
		midY := (fromY + toY) / 2
		for y := fromY; y <= midY; y++ {
			setCanvasByte(canvas, fromX, y, byte(char))
		}
		// Horizontal if needed
		if fromX != toX {
			hChar := '-'
			if edge.Style == MermaidEdgeDotted {
				hChar = '.'
			} else if edge.Style == MermaidEdgeThick {
				hChar = '='
			}
			if fromX < toX {
				for x := fromX; x <= toX; x++ {
					setCanvasByte(canvas, x, midY, byte(hChar))
				}
			} else {
				for x := toX; x <= fromX; x++ {
					setCanvasByte(canvas, x, midY, byte(hChar))
				}
			}
			for y := midY; y <= toY; y++ {
				setCanvasByte(canvas, toX, y, byte(char))
			}
		} else {
			for y := midY; y <= toY; y++ {
				setCanvasByte(canvas, toX, y, byte(char))
			}
		}
		// Arrow head
		setCanvasByte(canvas, toX, toY, 'v')
	} else {
		// Horizontal: straight right/left
		midX := (fromX + toX) / 2
		for x := fromX; x <= midX; x++ {
			setCanvasByte(canvas, x, fromY, byte(char))
		}
		if fromY != toY {
			vChar := '|'
			if edge.Style == MermaidEdgeDotted {
				vChar = '.'
			} else if edge.Style == MermaidEdgeThick {
				vChar = '='
			}
			if fromY < toY {
				for y := fromY; y <= toY; y++ {
					setCanvasByte(canvas, midX, y, byte(vChar))
				}
			} else {
				for y := toY; y <= fromY; y++ {
					setCanvasByte(canvas, midX, y, byte(vChar))
				}
			}
			for x := midX; x <= toX; x++ {
				setCanvasByte(canvas, x, toY, byte(char))
			}
		} else {
			for x := midX; x <= toX; x++ {
				setCanvasByte(canvas, x, toY, byte(char))
			}
		}
		// Arrow head
		setCanvasByte(canvas, toX, toY, '>')
	}

	// Draw edge label if present
	if edge.Label != "" {
		var lx, ly int
		if dir == MermaidTD || dir == MermaidBT {
			lx = (fromX + toX) / 2
			ly = (fromY + toY) / 2
		} else {
			lx = (fromX + toX) / 2
			ly = (fromY + toY) / 2
		}
		// Draw label with small offset
		for i := 0; i < len(edge.Label); i++ {
			setCanvasByte(canvas, lx-len(edge.Label)/2+i, ly, edge.Label[i])
		}
	}
}

// RenderMermaidText is a convenience function that parses and renders Mermaid text.
// Returns nil if the text is not valid Mermaid.
func RenderMermaidText(text string, theme *MarkdownTheme) [][]buffer.Cell {
	g := ParseMermaid(text)
	if g == nil {
		return nil
	}
	return RenderMermaid(g, theme)
}

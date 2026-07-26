package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// FileNode represents a single file or directory in the tree.
type FileNode struct {
	Name     string
	IsDir    bool
	Children []FileNode
	Expanded bool
}

// FileTree renders a collapsible file/directory tree with icons.
// Useful for file managers, project browsers, and AI code exploration UIs.
//
// Thread-safe.
type FileTree struct {
	BaseComponent
	mu    sync.RWMutex
	root  string
	nodes []FileNode
}

// NewFileTree creates a file tree with the given root label and nodes.
func NewFileTree(root string, nodes []FileNode) *FileTree {
	return &FileTree{
		BaseComponent: BaseComponent{id: GenerateID("filetree")},
		root:          root,
		nodes:         nodes,
	}
}

// Root returns the root label.
func (f *FileTree) Root() string { f.mu.RLock(); defer f.mu.RUnlock(); return f.root }

// SetRoot sets the root label.
func (f *FileTree) SetRoot(s string) { f.mu.Lock(); defer f.mu.Unlock(); f.root = s }

// Nodes returns a copy of the tree nodes.
func (f *FileTree) Nodes() []FileNode {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make([]FileNode, len(f.nodes))
	copy(out, f.nodes)
	return out
}

// SetNodes replaces the tree nodes.
func (f *FileTree) SetNodes(nodes []FileNode) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nodes = nodes
}

// Measure returns the preferred size.
func (f *FileTree) Measure(cs Constraints) Size {
	f.mu.RLock()
	defer f.mu.RUnlock()
	w := f.measureWidthLocked(0, f.nodes)
	if w < len(f.root)+1 { w = len(f.root) + 1 }
	h := 1 + f.countVisibleLocked(f.nodes)
	if cs.HasWidth() && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.HasHeight() && h > cs.MaxHeight { h = cs.MaxHeight }
	if w < 1 { w = 1 }
	if h < 1 { h = 1 }
	return Size{W: w, H: h}
}

func (f *FileTree) measureWidthLocked(depth int, nodes []FileNode) int {
	maxW := 0
	for _, n := range nodes {
		w := depth*2 + len(n.Name) + 3 // indent + icon + space + name
		if n.IsDir && n.Expanded {
			childW := f.measureWidthLocked(depth+1, n.Children)
			if childW > w { w = childW }
		}
		if w > maxW { maxW = w }
	}
	return maxW
}

func (f *FileTree) countVisibleLocked(nodes []FileNode) int {
	count := len(nodes)
	for _, n := range nodes {
		if n.IsDir && n.Expanded {
			count += f.countVisibleLocked(n.Children)
		}
	}
	return count
}

// Paint draws the file tree. Zero allocations.
func (f *FileTree) Paint(buf *buffer.Buffer) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	bounds := f.bounds
	if bounds.W <= 0 || bounds.H <= 0 { return }

	tt := theme.Get()
	dirStyle := buffer.Style{Fg: tt.Accent, Flags: buffer.Bold}
	fileStyle := buffer.Style{Fg: tt.Fg}
	muted := buffer.Style{Fg: tt.Muted}

	y := bounds.Y
	maxY := bounds.Y + bounds.H
	x := bounds.X

	// Root
	if y < maxY {
		x = buf.DrawText(x, y, "📁 ", dirStyle) // 📁
		buf.DrawText(x, y, f.root, dirStyle)
		y++
	}

	// Nodes
	f.paintNodesLocked(buf, bounds.X, &y, maxY, f.nodes, 1, dirStyle, fileStyle, muted)
}

func (f *FileTree) paintNodesLocked(buf *buffer.Buffer, baseX int, y *int, maxY int, nodes []FileNode, depth int, dirStyle, fileStyle, muted buffer.Style) {
	for _, n := range nodes {
		if *y >= maxY { return }
		indent := baseX + depth*2
		x := indent

		// Indent markers
		for i := baseX; i < indent && x < baseX+100; i++ {
			buf.SetCell(x, *y, buffer.Cell{Rune: ' ', Width: 1, Fg: muted.Fg})
			x++
		}

		// Icon
		if n.IsDir {
			if n.Expanded {
				x = buf.DrawText(x, *y, "📂 ", dirStyle) // 📂
			} else {
				x = buf.DrawText(x, *y, "📁 ", dirStyle) // 📁
			}
			buf.DrawText(x, *y, n.Name, dirStyle)
		} else {
			x = buf.DrawText(x, *y, "📄 ", fileStyle) // 📄
			buf.DrawText(x, *y, n.Name, fileStyle)
		}
		*y++

		// Recurse into expanded dirs
		if n.IsDir && n.Expanded && len(n.Children) > 0 {
			f.paintNodesLocked(buf, baseX, y, maxY, n.Children, depth+1, dirStyle, fileStyle, muted)
		}
	}
}

// ParsePathList builds a FileNode tree from a list of slash-separated paths.
// Useful for converting git ls-files or find output into a tree structure.
func ParsePathList(paths []string) []FileNode {
	var root []FileNode
	for _, p := range paths {
		parts := strings.Split(p, "/")
		root = insertPath(root, parts)
	}
	return root
}

func insertPath(nodes []FileNode, parts []string) []FileNode {
	if len(parts) == 0 { return nodes }
	name := parts[0]
	isDir := len(parts) > 1

	// Find existing
	for i := range nodes {
		if nodes[i].Name == name {
			if isDir {
				nodes[i].Children = insertPath(nodes[i].Children, parts[1:])
			}
			return nodes
		}
	}

	node := FileNode{Name: name, IsDir: isDir, Expanded: isDir}
	if isDir {
		node.Children = insertPath(nil, parts[1:])
	}
	return append(nodes, node)
}

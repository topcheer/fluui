package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// TreeNode represents a single node in the tree.
type TreeNode struct {
	ID       string
	Label    string
	Icon     string // icon displayed before label (e.g. "📁", "📄")
	Children []*TreeNode
	Expanded bool
	Selected bool
	parent   *TreeNode
	depth    int
}

// AddChild appends a child node and sets its parent/depth.
func (n *TreeNode) AddChild(child *TreeNode) {
	child.parent = n
	child.depth = n.depth + 1
	n.Children = append(n.Children, child)
}

// IsLeaf returns true if the node has no children.
func (n *TreeNode) IsLeaf() bool {
	return len(n.Children) == 0
}

// HasChildren returns true if the node has at least one child.
func (n *TreeNode) HasChildren() bool {
	return len(n.Children) > 0
}

// NewTreeNode creates a node with the given ID and label.
func NewTreeNode(id, label string) *TreeNode {
	return &TreeNode{
		ID:    id,
		Label: label,
	}
}

// --- visibleNode is an internal flat representation for rendering ---

type visibleNode struct {
	node  *TreeNode
	depth int
}

// Tree is a scrollable tree view component with keyboard navigation.
// It supports expand/collapse, selection, and callbacks for toggle/select events.
type Tree struct {
	BaseComponent
	mu sync.RWMutex

	root    *TreeNode
	current int    // index into flatList
	flatList []visibleNode
	scrollY  int

	defStyle   buffer.Style
	selStyle   buffer.Style
	expandStyle buffer.Style

	onToggle func(node *TreeNode)
	onSelect func(node *TreeNode)
}

// NewTree creates an empty Tree with default styles.
func NewTree() *Tree {
	return &Tree{
		defStyle:    buffer.Style{},
		selStyle:    buffer.Style{Flags: buffer.Reverse},
		expandStyle: buffer.Style{Fg: buffer.NamedColor(buffer.NamedYellow)},
	}
}

// SetRoot sets the root node and rebuilds the flat list.
func (t *Tree) SetRoot(node *TreeNode) {
	t.mu.Lock()
	t.root = node
	node.depth = 0
	node.parent = nil
	t.current = 0
	t.scrollY = 0
	t.rebuildLocked()
	t.mu.Unlock()
}

// SetStyle sets the default text style.
func (t *Tree) SetStyle(s buffer.Style) {
	t.mu.Lock()
	t.defStyle = s
	t.mu.Unlock()
}

// SetSelectedStyle sets the style for the currently focused row.
func (t *Tree) SetSelectedStyle(s buffer.Style) {
	t.mu.Lock()
	t.selStyle = s
	t.mu.Unlock()
}

// OnToggle sets a callback invoked when a node is expanded or collapsed.
func (t *Tree) OnToggle(fn func(node *TreeNode)) {
	t.mu.Lock()
	t.onToggle = fn
	t.mu.Unlock()
}

// OnSelect sets a callback invoked when a node is selected (Enter/Space).
func (t *Tree) OnSelect(fn func(node *TreeNode)) {
	t.mu.Lock()
	t.onSelect = fn
	t.mu.Unlock()
}

// CurrentNode returns the currently focused node, or nil if empty.
func (t *Tree) CurrentNode() *TreeNode {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.current < 0 || t.current >= len(t.flatList) {
		return nil
	}
	return t.flatList[t.current].node
}

// CurrentIndex returns the index of the focused row.
func (t *Tree) CurrentIndex() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.current
}

// ScrollY returns the current vertical scroll offset.
func (t *Tree) ScrollY() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.scrollY
}

// SetCurrent sets the focus to the given flat-list index.
func (t *Tree) SetCurrent(idx int) {
	t.mu.Lock()
	if idx < 0 {
		idx = 0
	}
	if idx >= len(t.flatList) {
		idx = len(t.flatList) - 1
	}
	t.current = idx
	t.ensureVisibleLocked()
	t.mu.Unlock()
}

// VisibleCount returns the number of currently visible rows (expanded nodes only).
func (t *Tree) VisibleCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.flatList)
}

// ToggleExpand toggles the expansion state of the current node.
func (t *Tree) ToggleExpand() {
	t.mu.Lock()
	if t.current >= 0 && t.current < len(t.flatList) {
		node := t.flatList[t.current].node
		if node.HasChildren() {
			node.Expanded = !node.Expanded
			cb := t.onToggle
			t.rebuildLocked()
			t.ensureVisibleLocked()
			t.mu.Unlock()
			if cb != nil {
				cb(node)
			}
			return
		}
	}
	t.mu.Unlock()
}

// ExpandAll expands all nodes in the tree.
func (t *Tree) ExpandAll() {
	t.mu.Lock()
	t.expandAllLocked(t.root)
	t.rebuildLocked()
	t.mu.Unlock()
}

func (t *Tree) expandAllLocked(node *TreeNode) {
	if node == nil {
		return
	}
	if node.HasChildren() {
		node.Expanded = true
	}
	for _, c := range node.Children {
		t.expandAllLocked(c)
	}
}

// CollapseAll collapses all nodes in the tree.
func (t *Tree) CollapseAll() {
	t.mu.Lock()
	t.collapseAllLocked(t.root)
	t.rebuildLocked()
	t.current = 0
	t.scrollY = 0
	t.mu.Unlock()
}

func (t *Tree) collapseAllLocked(node *TreeNode) {
	if node == nil {
		return
	}
	node.Expanded = false
	for _, c := range node.Children {
		t.collapseAllLocked(c)
	}
}

// Path returns the path of IDs from root to the current node.
func (t *Tree) Path() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.current < 0 || t.current >= len(t.flatList) {
		return nil
	}
	node := t.flatList[t.current].node
	var path []string
	for n := node; n != nil; n = n.parent {
		path = append([]string{n.ID}, path...)
	}
	return path
}

// PathLabels returns the labels from root to the current node.
func (t *Tree) PathLabels() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.current < 0 || t.current >= len(t.flatList) {
		return nil
	}
	node := t.flatList[t.current].node
	var labels []string
	for n := node; n != nil; n = n.parent {
		labels = append([]string{n.Label}, labels...)
	}
	return labels
}

// HandleKey processes keyboard input for tree navigation.
// Returns true if the key was consumed.
//
// Key bindings:
//   - Up/Down: move focus
//   - Right/Enter: expand node (or move to first child if already expanded)
//   - Left: collapse node (or move to parent if already collapsed)
//   - Space: toggle selection
//   - Home/End: jump to first/last visible node
//   - PageUp/PageDown: scroll by viewport height
func (t *Tree) HandleKey(key *term.KeyEvent) bool {
	if key == nil {
		return false
	}

	switch key.Key {
	case term.KeyUp:
		t.moveCursor(-1)
		return true

	case term.KeyDown:
		t.moveCursor(1)
		return true

	case term.KeyRight:
		return t.expandOrDescend()

	case term.KeyEnter:
		t.expandOrDescend()
		t.selectCurrent()
		return true

	case term.KeyLeft:
		return t.collapseOrAscend()

	case term.KeySpace:
		return t.selectCurrent()

	case term.KeyHome:
		t.mu.Lock()
		t.current = 0
		t.scrollY = 0
		t.mu.Unlock()
		return true

	case term.KeyEnd:
		t.mu.Lock()
		if len(t.flatList) > 0 {
			t.current = len(t.flatList) - 1
			t.ensureVisibleLocked()
		}
		t.mu.Unlock()
		return true

	case term.KeyPageUp:
		t.pageMove(-1)
		return true

	case term.KeyPageDown:
		t.pageMove(1)
		return true
	}

	return false
}

func (t *Tree) moveCursor(delta int) {
	t.mu.Lock()
	t.current += delta
	if t.current < 0 {
		t.current = 0
	}
	if t.current >= len(t.flatList) {
		t.current = len(t.flatList) - 1
	}
	if t.current < 0 {
		t.current = 0 // empty list guard
	}
	t.ensureVisibleLocked()
	t.mu.Unlock()
}

// expandOrDescend expands current node if collapsed, or moves to first child.
func (t *Tree) expandOrDescend() bool {
	t.mu.Lock()
	if t.current < 0 || t.current >= len(t.flatList) {
		t.mu.Unlock()
		return true
	}
	node := t.flatList[t.current].node
	if node.HasChildren() && !node.Expanded {
		node.Expanded = true
		cb := t.onToggle
		t.rebuildLocked()
		t.ensureVisibleLocked()
		t.mu.Unlock()
		if cb != nil {
			cb(node)
		}
		return true
	}
	// Already expanded or leaf — descend to first child if possible
	if node.HasChildren() && node.Expanded {
		if t.current+1 < len(t.flatList) {
			next := t.flatList[t.current+1]
			if next.depth > t.flatList[t.current].depth {
				t.current++
				t.ensureVisibleLocked()
			}
		}
	}
	t.mu.Unlock()
	return true
}

// collapseOrAscend collapses current node if expanded, or moves to parent.
func (t *Tree) collapseOrAscend() bool {
	t.mu.Lock()
	if t.current < 0 || t.current >= len(t.flatList) {
		t.mu.Unlock()
		return true
	}
	node := t.flatList[t.current].node
	if node.HasChildren() && node.Expanded {
		node.Expanded = false
		cb := t.onToggle
		t.rebuildLocked()
		t.ensureVisibleLocked()
		t.mu.Unlock()
		if cb != nil {
			cb(node)
		}
		return true
	}
	// Already collapsed or leaf — move to parent
	if node.parent != nil {
		for i, vn := range t.flatList {
			if vn.node == node.parent {
				t.current = i
				t.ensureVisibleLocked()
				break
			}
		}
	}
	t.mu.Unlock()
	return true
}

func (t *Tree) selectCurrent() bool {
	t.mu.Lock()
	if t.current < 0 || t.current >= len(t.flatList) {
		t.mu.Unlock()
		return true
	}
	node := t.flatList[t.current].node
	node.Selected = !node.Selected
	cb := t.onSelect
	t.mu.Unlock()
	if cb != nil {
		cb(node)
	}
	return true
}

func (t *Tree) pageMove(dir int) {
	t.mu.Lock()
	h := t.bounds.H
	if h <= 0 {
		h = 10
	}
	t.current += dir * h
	if t.current < 0 {
		t.current = 0
	}
	if t.current >= len(t.flatList) {
		t.current = len(t.flatList) - 1
	}
	if t.current < 0 {
		t.current = 0
	}
	t.ensureVisibleLocked()
	t.mu.Unlock()
}

// --- Flat list management ---

// rebuildLocked regenerates the flat visible list from the tree.
// Caller must hold the write lock.
func (t *Tree) rebuildLocked() {
	t.flatList = t.flatList[:0]
	if t.root == nil {
		return
	}
	t.flattenLocked(t.root, 0)
}

func (t *Tree) flattenLocked(node *TreeNode, depth int) {
	t.flatList = append(t.flatList, visibleNode{node: node, depth: depth})
	if node.Expanded {
		for _, child := range node.Children {
			t.flattenLocked(child, depth+1)
		}
	}
}

// ensureVisibleLocked adjusts scrollY so current row is visible.
// Caller must hold the write lock.
func (t *Tree) ensureVisibleLocked() {
	h := t.bounds.H
	if h <= 0 {
		return
	}
	if t.current < t.scrollY {
		t.scrollY = t.current
	}
	if t.current >= t.scrollY+h {
		t.scrollY = t.current - h + 1
	}
	if t.scrollY < 0 {
		t.scrollY = 0
	}
}

// --- Component interface ---

// Measure returns the desired size: width = longest visible line, height = visible node count.
func (t *Tree) Measure(cs Constraints) Size {
	t.mu.RLock()
	defer t.mu.RUnlock()

	maxW := 0
	for _, vn := range t.flatList {
		w := vn.depth*2 + len(vn.node.Label) + 4 // indent + icon space + arrow
		if vn.node.Icon != "" {
			w += len(vn.node.Icon)
		}
		if w > maxW {
			maxW = w
		}
	}
	h := len(t.flatList)
	if cs.MaxWidth > 0 && maxW > cs.MaxWidth {
		maxW = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: maxW, H: h}
}

// Paint renders the tree within bounds.
func (t *Tree) Paint(buf *buffer.Buffer) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	bounds := t.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	maxX := bounds.X + bounds.W
	for row := 0; row < bounds.H; row++ {
		idx := t.scrollY + row
		if idx >= len(t.flatList) {
			break
		}
		vn := t.flatList[idx]
		y := bounds.Y + row

		style := t.defStyle
		// Highlight current row
		if idx == t.current {
			style = t.selStyle
		}

		// Draw the text directly via DrawText — avoids strings.Builder allocation
		x := bounds.X

		// Indentation
		for i := 0; i < vn.depth && x < maxX; i++ {
			x = buf.DrawText(x, y, "  ", style)
		}

		// Expand/collapse indicator
		if x < maxX {
			if vn.node.HasChildren() {
				if vn.node.Expanded {
					x = buf.DrawText(x, y, "▼ ", style)
				} else {
					x = buf.DrawText(x, y, "▶ ", style)
				}
			} else {
				x = buf.DrawText(x, y, "  ", style)
			}
		}

		// Icon
		if vn.node.Icon != "" && x < maxX {
			x = buf.DrawText(x, y, vn.node.Icon, style)
			if x < maxX {
				x = buf.DrawText(x, y, " ", style)
			}
		}

		// Label (DrawText handles width truncation implicitly via buf bounds)
		if x < maxX {
			x = buf.DrawText(x, y, vn.node.Label, style)
		}

		// Fill rest of row with background style (for highlight)
		if idx == t.current {
			for ; x < maxX; x++ {
				buf.SetCell(x, y, buffer.Cell{
					Rune:  ' ',
					Width: 1,
					Fg:    style.Fg,
					Bg:    style.Bg,
					Flags: style.Flags,
				})
			}
		}

		// If selected, mark with checkmark at far right
		if vn.node.Selected && bounds.W > 2 {
			selX := bounds.X + bounds.W - 1
			buf.SetCell(selX, y, buffer.Cell{
				Rune:  '✓',
				Width: 1,
				Fg:    t.expandStyle.Fg,
				Flags: 0,
			})
		}
	}
}

// --- Additional accessors and utilities ---

// Cursor returns the current flat-list cursor index (alias for CurrentIndex).
func (t *Tree) Cursor() int {
	return t.CurrentIndex()
}

// SetData is an alias for SetRoot.
func (t *Tree) SetData(root *TreeNode) {
	t.SetRoot(root)
}

// Root returns the root node of the tree.
func (t *Tree) Root() *TreeNode {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.root
}

// SetScrollY manually sets the scroll offset (clamped to >= 0).
func (t *Tree) SetScrollY(y int) {
	t.mu.Lock()
	if y < 0 {
		y = 0
	}
	t.scrollY = y
	t.mu.Unlock()
}

// SelectedNodes returns all nodes in the tree that are marked as Selected.
func (t *Tree) SelectedNodes() []*TreeNode {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.root == nil {
		return nil
	}
	var result []*TreeNode
	t.collectSelectedLocked(t.root, &result)
	return result
}

func (t *Tree) collectSelectedLocked(n *TreeNode, result *[]*TreeNode) {
	if n.Selected {
		*result = append(*result, n)
	}
	for _, c := range n.Children {
		t.collectSelectedLocked(c, result)
	}
}

// SetCursor navigates the cursor to the node with the given ID.
// Expands parents along the way so the target is visible. Returns false if not found.
func (t *Tree) SetCursor(id string) bool {
	t.mu.Lock()
	target := t.findAndExpandParentsLocked(t.root, id)
	if target == nil {
		t.mu.Unlock()
		return false
	}
	t.rebuildLocked()
	for i, vn := range t.flatList {
		if vn.node.ID == id {
			t.current = i
			t.ensureVisibleLocked()
			t.mu.Unlock()
			return true
		}
	}
	t.mu.Unlock()
	return false
}

func (t *Tree) findAndExpandParentsLocked(n *TreeNode, id string) *TreeNode {
	if n == nil {
		return nil
	}
	if n.ID == id {
		return n
	}
	for _, c := range n.Children {
		if found := t.findAndExpandParentsLocked(c, id); found != nil {
			n.Expanded = true
			return found
		}
	}
	return nil
}

// String returns a text representation of the tree for debugging.
func (t *Tree) String() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.root == nil {
		return ""
	}
	var sb strings.Builder
	t.stringLocked(t.root, 0, &sb)
	return sb.String()
}

func (t *Tree) stringLocked(n *TreeNode, depth int, sb *strings.Builder) {
	sb.WriteString(strings.Repeat("  ", depth))
	if n.HasChildren() {
		if n.Expanded {
			sb.WriteString("▼ ")
		} else {
			sb.WriteString("▶ ")
		}
	}
	if n.Icon != "" {
		sb.WriteString(n.Icon + " ")
	}
	sb.WriteString(n.Label)
	sb.WriteByte('\n')
	if n.Expanded {
		for _, c := range n.Children {
			t.stringLocked(c, depth+1, sb)
		}
	}
}

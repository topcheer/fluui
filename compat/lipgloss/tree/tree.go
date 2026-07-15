// Package tree provides a lipgloss-compatible tree renderer.
// Mirrors charm.land/lipgloss/v2/tree API: Root(), Child(), EnumeratorStyle(), String().
package tree

import (
	"strings"

	"github.com/topcheer/fluui/compat/lipgloss"
)

// Tree represents a tree node with children.
type Tree struct {
	root            string
	children        []string
	enumeratorStyle lipgloss.Style
	indent          int
}

// Root creates a new tree root node.
func Root(root string) *Tree {
	return &Tree{
		root:   root,
		indent: 2,
	}
}

// Child adds a child string to the tree.
func (t *Tree) Child(s string) *Tree {
	t.children = append(t.children, s)
	return t
}

// EnumeratorStyle sets the style for tree enumerators (branch characters).
func (t *Tree) EnumeratorStyle(s lipgloss.Style) *Tree {
	t.enumeratorStyle = s
	return t
}

// Indent sets the indentation level.
func (t *Tree) Indent(n int) *Tree {
	t.indent = n
	return t
}

// String renders the tree as a string.
func (t *Tree) String() string {
	var b strings.Builder
	b.WriteString(t.root)
	b.WriteByte('\n')
	for _, child := range t.children {
		// Multi-line children get indented
		lines := strings.Split(child, "\n")
		for i, line := range lines {
			if i == 0 {
				b.WriteString("├─ ")
				b.WriteString(line)
			} else {
				b.WriteString(strings.Repeat(" ", t.indent+2))
				b.WriteString(line)
			}
			b.WriteByte('\n')
		}
	}
	return strings.TrimSuffix(b.String(), "\n")
}
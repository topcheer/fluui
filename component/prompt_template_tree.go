package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── PromptTemplateTree: Hierarchical Prompt Template Display ───
//
// PromptTemplateTree renders a tree of prompt templates with variable
// interpolation highlighting. Variables {{name}} are shown in a distinct color.
//
// Usage:
//
//	ptt := NewPromptTemplateTree()
//	ptt.AddNode(0, "System Prompt", "You are {{role}}.", false)
//	ptt.AddNode(1, "Greeting", "Hello {{user}}!", false)
//	ptt.AddNode(0, "Response", "Based on {{context}}", false)
//	ptt.Paint(buf)

// PromptNode represents a single template tree node.
type PromptNode struct {
	Label    string
	Template string
	Indent   int
	Expanded bool
}

// PromptTemplateStyle holds styling.
type PromptTemplateStyle struct {
	Label  buffer.Style
	Var    buffer.Style // {{variables}}
	Text   buffer.Style
	Guide  buffer.Style // tree guides
	Border buffer.Style
}

// DefaultPromptTemplateStyle returns defaults.
func DefaultPromptTemplateStyle() PromptTemplateStyle {
	label := buffer.Style{Fg: buffer.RGB(167, 139, 250), Flags: buffer.Bold}
	vr := buffer.Style{Fg: buffer.RGB(251, 146, 60)}
	text := buffer.Style{Fg: buffer.RGB(226, 232, 240)}
	guide := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return PromptTemplateStyle{Label: label, Var: vr, Text: text, Guide: guide, Border: border}
}

// PromptTemplateTree renders hierarchical prompt templates.
type PromptTemplateTree struct {
	BaseComponent
	mu sync.Mutex

	nodes []PromptNode
	style PromptTemplateStyle
}

// NewPromptTemplateTree creates a PromptTemplateTree.
func NewPromptTemplateTree() *PromptTemplateTree {
	ptt := &PromptTemplateTree{style: DefaultPromptTemplateStyle()}
	ptt.SetID(GenerateID("prompttree"))
	return ptt
}

// AddNode adds a template node at the given indent level.
func (ptt *PromptTemplateTree) AddNode(indent int, label, template string, expanded bool) *PromptTemplateTree {
	ptt.mu.Lock()
	ptt.nodes = append(ptt.nodes, PromptNode{
		Label:    label,
		Template: template,
		Indent:   indent,
		Expanded: expanded,
	})
	ptt.mu.Unlock()
	return ptt
}

// NodeCount returns the number of nodes.
func (ptt *PromptTemplateTree) NodeCount() int {
	ptt.mu.Lock()
	defer ptt.mu.Unlock()
	return len(ptt.nodes)
}

// ToggleExpand toggles expansion state of a node by index.
func (ptt *PromptTemplateTree) ToggleExpand(index int) *PromptTemplateTree {
	ptt.mu.Lock()
	if index >= 0 && index < len(ptt.nodes) {
		ptt.nodes[index].Expanded = !ptt.nodes[index].Expanded
	}
	ptt.mu.Unlock()
	return ptt
}

// Clear removes all nodes.
func (ptt *PromptTemplateTree) Clear() *PromptTemplateTree {
	ptt.mu.Lock()
	ptt.nodes = ptt.nodes[:0]
	ptt.mu.Unlock()
	return ptt
}

// SetStyle sets custom style.
func (ptt *PromptTemplateTree) SetStyle(s PromptTemplateStyle) *PromptTemplateTree {
	ptt.mu.Lock()
	ptt.style = s
	ptt.mu.Unlock()
	return ptt
}

// Measure returns preferred size.
func (ptt *PromptTemplateTree) Measure(cs Constraints) Size {
	ptt.mu.Lock()
	count := len(ptt.nodes)
	ptt.mu.Unlock()
	w := 50
	h := count + 2
	if h < 3 { h = 3 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the prompt template tree into the buffer.
func (ptt *PromptTemplateTree) Paint(buf *buffer.Buffer) {
	ptt.mu.Lock()
	defer ptt.mu.Unlock()

	b := ptt.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 20 { w = 50 }
	if h < 3 { h = 3 }

	bs := ptt.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	labelStyle := ptt.style.Label
	varStyle := ptt.style.Var
	textStyle := ptt.style.Text
	guideStyle := ptt.style.Guide

	for idx, node := range ptt.nodes {
		rowY := y + 1 + idx
		if rowY >= y+h-1 || rowY >= buf.Height { break }

		col := x + 1

		// Indentation guides
		for i := 0; i < node.Indent; i++ {
			if col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: '│', Fg: guideStyle.Fg, Bg: guideStyle.Bg, Flags: guideStyle.Flags, Width: 1})
			col++
			if col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: guideStyle.Fg, Bg: guideStyle.Bg, Flags: guideStyle.Flags, Width: 1})
			col++
		}

		// Tree connector
		if node.Indent > 0 {
			if col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: '├', Fg: guideStyle.Fg, Bg: guideStyle.Bg, Flags: guideStyle.Flags, Width: 1})
			col++
			if col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: guideStyle.Fg, Bg: guideStyle.Bg, Flags: guideStyle.Flags, Width: 1})
			col++
		}

		// Label
		for _, r := range node.Label {
			if col >= x+w-1 || col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}

		// Template with variable highlighting
		if node.Template != "" {
			if col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: ':', Fg: guideStyle.Fg, Bg: guideStyle.Bg, Flags: guideStyle.Flags, Width: 1})
			col++
			if col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
			col++

			remaining := node.Template
			for len(remaining) > 0 {
				if col >= x+w-1 || col >= buf.Width { break }
				varIdx := strings.Index(remaining, "{{")
				if varIdx < 0 {
					for _, r := range remaining {
						if col >= x+w-1 || col >= buf.Width { break }
						buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
						col++
					}
					break
				}
				// Text before variable
				if varIdx > 0 {
					for _, r := range remaining[:varIdx] {
						if col >= x+w-1 || col >= buf.Width { break }
						buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
						col++
					}
				}
				// Variable itself
				afterVar := remaining[varIdx:]
				endIdx := strings.Index(afterVar, "}}")
				varEnd := len(afterVar)
				if endIdx >= 0 { varEnd = endIdx + 2 }
				for _, r := range afterVar[:varEnd] {
					if col >= x+w-1 || col >= buf.Width { break }
					buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: varStyle.Fg, Bg: varStyle.Bg, Flags: varStyle.Flags, Width: 1})
					col++
				}
				if endIdx >= 0 {
					remaining = afterVar[varEnd:]
				} else {
					break
				}
			}
		}
	}
}

// Children returns nil.
func (ptt *PromptTemplateTree) Children() []Component { return nil }

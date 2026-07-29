package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownTaskList: Render GitHub-Style Task Lists ───
//
// MarkdownTaskList parses markdown task list items (- [ ] and - [x]) and
// renders them with checkbox characters (☐/☑), colored based on completion
// state. Supports regular list items (- item) without checkboxes too.
//
// Usage:
//
//	tl := NewMarkdownTaskList()
//	tl.SetMarkdown("- [ ] Pending task\n- [x] Done task\n- [ ] Another")
//	tl.ToggleItem(0)
//	tl.Paint(buf)

// TaskListStyle holds styling for MarkdownTaskList.
type TaskListStyle struct {
	Checked    buffer.Style
	Unchecked  buffer.Style
	NormalItem buffer.Style // items without checkbox
	Checkbox   buffer.Style // the checkbox char itself
	Border     buffer.Style
}

// DefaultTaskListStyle returns sensible defaults.
func DefaultTaskListStyle() TaskListStyle {
	checked := buffer.Style{Fg: buffer.RGB(34, 197, 94)}     // green-500
	unchecked := buffer.Style{Fg: buffer.RGB(148, 163, 184)}  // slate-400
	normal := buffer.Style{Fg: buffer.RGB(203, 213, 225)}     // slate-300
	checkbox := buffer.Style{Fg: buffer.RGB(96, 165, 250)}   // blue-400
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}      // slate-600
	return TaskListStyle{Checked: checked, Unchecked: unchecked, NormalItem: normal, Checkbox: checkbox, Border: border}
}

// TaskItem represents a parsed task or list item.
type TaskItem struct {
	Text     string
	Checked  bool
	HasCheck bool // true if it has [ ] or [x], false if plain list item
}

// MarkdownTaskList renders markdown task lists with checkboxes.
type MarkdownTaskList struct {
	BaseComponent
	mu sync.Mutex

	source string
	style  TaskListStyle

	// cached parsed items
	cachedItems []TaskItem
}

// NewMarkdownTaskList creates a MarkdownTaskList with defaults.
func NewMarkdownTaskList() *MarkdownTaskList {
	tl := &MarkdownTaskList{
		style: DefaultTaskListStyle(),
	}
	tl.SetID(GenerateID("tasklist"))
	return tl
}

// SetMarkdown sets the raw markdown source and parses task items.
func (tl *MarkdownTaskList) SetMarkdown(source string) *MarkdownTaskList {
	tl.mu.Lock()
	tl.source = source
	tl.parseLocked()
	tl.mu.Unlock()
	return tl
}

// Markdown returns the raw markdown source.
func (tl *MarkdownTaskList) Markdown() string {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	return tl.source
}

// SetStyle sets the custom style.
func (tl *MarkdownTaskList) SetStyle(s TaskListStyle) *MarkdownTaskList {
	tl.mu.Lock()
	tl.style = s
	tl.mu.Unlock()
	return tl
}

// parseLocked parses markdown list items. Caller must hold lock.
func (tl *MarkdownTaskList) parseLocked() {
	tl.cachedItems = tl.cachedItems[:0]
	lines := strings.Split(tl.source, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check for task list markers: - [ ], - [x], - [X], * [ ], * [x]
		item := TaskItem{}
		rest := trimmed

		// Strip leading -, *, +
		if strings.HasPrefix(rest, "- ") || strings.HasPrefix(rest, "* ") || strings.HasPrefix(rest, "+ ") {
			rest = rest[2:]
		} else if strings.HasPrefix(rest, "-") || strings.HasPrefix(rest, "*") {
			rest = rest[1:]
		}

		// Check for checkbox
		if strings.HasPrefix(rest, "[ ]") {
			item.HasCheck = true
			item.Checked = false
			rest = rest[3:]
		} else if strings.HasPrefix(rest, "[x]") || strings.HasPrefix(rest, "[X]") {
			item.HasCheck = true
			item.Checked = true
			rest = rest[3:]
		}

		item.Text = strings.TrimSpace(rest)
		tl.cachedItems = append(tl.cachedItems, item)
	}
}

// ToggleItem toggles the checked state of a task item by index.
func (tl *MarkdownTaskList) ToggleItem(index int) *MarkdownTaskList {
	tl.mu.Lock()
	if index >= 0 && index < len(tl.cachedItems) && tl.cachedItems[index].HasCheck {
		tl.cachedItems[index].Checked = !tl.cachedItems[index].Checked
	}
	tl.mu.Unlock()
	return tl
}

// CompletedCount returns the number of checked task items.
func (tl *MarkdownTaskList) CompletedCount() int {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	count := 0
	for _, item := range tl.cachedItems {
		if item.HasCheck && item.Checked {
			count++
		}
	}
	return count
}

// TotalCount returns the total number of task items (including non-checkbox).
func (tl *MarkdownTaskList) TotalCount() int {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	return len(tl.cachedItems)
}

// TaskCount returns the number of items with checkboxes only.
func (tl *MarkdownTaskList) TaskCount() int {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	count := 0
	for _, item := range tl.cachedItems {
		if item.HasCheck {
			count++
		}
	}
	return count
}

// Measure returns the preferred size.
func (tl *MarkdownTaskList) Measure(cs Constraints) Size {
	tl.mu.Lock()
	itemCount := len(tl.cachedItems)
	tl.mu.Unlock()

	w := 40
	h := itemCount + 2 // items + borders
	if h < 3 {
		h = 3
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the task list into the buffer.
func (tl *MarkdownTaskList) Paint(buf *buffer.Buffer) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	b := tl.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 {
		w = 40
	}
	if h < 3 {
		h = 3
	}

	// Draw border
	bs := tl.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 {
				ch = '┌'
			} else if row == 0 && col == w-1 {
				ch = '┐'
			} else if row == h-1 && col == 0 {
				ch = '└'
			} else if row == h-1 && col == w-1 {
				ch = '┘'
			} else if row == 0 || row == h-1 {
				ch = '─'
			} else if col == 0 || col == w-1 {
				ch = '│'
			}
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	// Draw each item
	for idx, item := range tl.cachedItems {
		rowY := y + 1 + idx
		if rowY >= y+h-1 || rowY >= buf.Height {
			break
		}

		col := x + 2

		if item.HasCheck {
			// Draw checkbox character
			cbStyle := tl.style.Checkbox
			var cbChar rune
			if item.Checked {
				cbChar = '☑'
				cbStyle = tl.style.Checked
			} else {
				cbChar = '☐'
				cbStyle = tl.style.Unchecked
			}
			if col < x+w-1 && col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: cbChar, Fg: cbStyle.Fg, Bg: cbStyle.Bg, Flags: cbStyle.Flags, Width: 1})
			}
			col++

			// Space after checkbox
			if col < x+w-1 && col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: cbStyle.Fg, Bg: cbStyle.Bg, Flags: cbStyle.Flags, Width: 1})
			}
			col++

			// Text in checked/unchecked style
			textStyle := cbStyle
			for _, r := range item.Text {
				if col >= x+w-1 || col >= buf.Width {
					break
				}
				buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
				col++
			}
		} else {
			// Normal list item with bullet
			normalStyle := tl.style.NormalItem
			if col < x+w-1 && col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: '•', Fg: normalStyle.Fg, Bg: normalStyle.Bg, Flags: normalStyle.Flags, Width: 1})
			}
			col++
			if col < x+w-1 && col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: normalStyle.Fg, Bg: normalStyle.Bg, Flags: normalStyle.Flags, Width: 1})
			}
			col++

			for _, r := range item.Text {
				if col >= x+w-1 || col >= buf.Width {
					break
				}
				buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: normalStyle.Fg, Bg: normalStyle.Bg, Flags: normalStyle.Flags, Width: 1})
				col++
			}
		}
	}
}

// Children returns nil.
func (tl *MarkdownTaskList) Children() []Component { return nil }

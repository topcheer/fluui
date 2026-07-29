package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MergeView: Side-by-Side Diff/Merge Conflict Display ───
//
// MergeView renders a side-by-side comparison of two text blocks with
// conflict markers (<<<<<<< ======= >>>>>>>), highlighting added/removed/
// conflicting lines with distinct colors. Useful in merge tools, code
// review interfaces, and git conflict resolution UIs.
//
// Usage:
//
//	mv := NewMergeView()
//	mv.SetLeft("ours", "line1\nline2\nline3")
//	mv.SetRight("theirs", "line1\nchanged\nline3")
//	mv.Paint(buf)

// MergeLineStyle holds per-line type styling.
type MergeLineStyle struct {
	Equal     buffer.Style // unchanged lines
	Added     buffer.Style // lines only in right
	Removed   buffer.Style // lines only in left
	Conflict  buffer.Style // conflict marker lines
	Header    buffer.Style // side labels
	Border    buffer.Style
	Separator buffer.Style
}

// DefaultMergeLineStyle returns sensible defaults.
func DefaultMergeLineStyle() MergeLineStyle {
	equal := buffer.Style{Fg: buffer.RGB(148, 163, 184)}   // slate-400
	added := buffer.Style{Fg: buffer.RGB(34, 197, 94)}     // green-500
	removed := buffer.Style{Fg: buffer.RGB(239, 68, 68)}   // red-500
	conflict := buffer.Style{Fg: buffer.RGB(234, 179, 8), Flags: buffer.Bold} // yellow-500 bold
	header := buffer.Style{Fg: buffer.RGB(167, 139, 250), Flags: buffer.Bold} // violet-400 bold
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}    // slate-600
	separator := buffer.Style{Fg: buffer.RGB(100, 116, 139)} // slate-500
	return MergeLineStyle{
		Equal:     equal,
		Added:     added,
		Removed:   removed,
		Conflict:  conflict,
		Header:    header,
		Border:    border,
		Separator: separator,
	}
}

// MergeLineType classifies a diff line.
type MergeLineType int

const (
	MergeEqual    MergeLineType = iota
	MergeAdded
	MergeRemoved
	MergeConflictMarker
)

// MergeLine represents a single rendered line in the merge view.
type MergeLine struct {
	Text     string
	Type     MergeLineType
	LeftOnly bool // true = show on left side only
	RightOnly bool // show on right side only
}

// MergeView displays a side-by-side diff/merge view.
type MergeView struct {
	BaseComponent
	mu sync.Mutex

	leftLabel  string
	leftContent string
	rightLabel  string
	rightContent string

	style MergeLineStyle

	// cached computed lines
	cachedLines []MergeLine
}

// NewMergeView creates a MergeView with defaults.
func NewMergeView() *MergeView {
	mv := &MergeView{
		leftLabel:  "LEFT",
		rightLabel: "RIGHT",
		style:      DefaultMergeLineStyle(),
	}
	mv.SetID(GenerateID("merge"))
	return mv
}

// SetLeft sets the left side label and content.
func (mv *MergeView) SetLeft(label, content string) *MergeView {
	mv.mu.Lock()
	mv.leftLabel = label
	mv.leftContent = content
	mv.cachedLines = nil
	mv.mu.Unlock()
	return mv
}

// SetRight sets the right side label and content.
func (mv *MergeView) SetRight(label, content string) *MergeView {
	mv.mu.Lock()
	mv.rightLabel = label
	mv.rightContent = content
	mv.cachedLines = nil
	mv.mu.Unlock()
	return mv
}

// LeftLabel returns the left label.
func (mv *MergeView) LeftLabel() string {
	mv.mu.Lock()
	defer mv.mu.Unlock()
	return mv.leftLabel
}

// RightLabel returns the right label.
func (mv *MergeView) RightLabel() string {
	mv.mu.Lock()
	defer mv.mu.Unlock()
	return mv.rightLabel
}

// LeftContent returns the left content.
func (mv *MergeView) LeftContent() string {
	mv.mu.Lock()
	defer mv.mu.Unlock()
	return mv.leftContent
}

// RightContent returns the right content.
func (mv *MergeView) RightContent() string {
	mv.mu.Lock()
	defer mv.mu.Unlock()
	return mv.rightContent
}

// SetStyle sets the custom style.
func (mv *MergeView) SetStyle(s MergeLineStyle) *MergeView {
	mv.mu.Lock()
	mv.style = s
	mv.mu.Unlock()
	return mv
}

// HasConflicts returns true if either content contains conflict markers.
func (mv *MergeView) HasConflicts() bool {
	mv.mu.Lock()
	defer mv.mu.Unlock()
	return strings.Contains(mv.leftContent, "<<<<<<<") ||
		strings.Contains(mv.rightContent, "<<<<<<<") ||
		strings.Contains(mv.leftContent, "=======") ||
		strings.Contains(mv.rightContent, "=======")
}

// computeLinesLocked computes the diff lines. Caller must hold lock.
func (mv *MergeView) computeLinesLocked() {
	if mv.cachedLines != nil {
		return
	}

	leftLines := strings.Split(mv.leftContent, "\n")
	rightLines := strings.Split(mv.rightContent, "\n")

	maxLen := len(leftLines)
	if len(rightLines) > maxLen {
		maxLen = len(rightLines)
	}

	mv.cachedLines = make([]MergeLine, 0, maxLen)

	for i := 0; i < maxLen; i++ {
		var leftLine, rightLine string
		hasLeft := i < len(leftLines)
		hasRight := i < len(rightLines)
		if hasLeft {
			leftLine = leftLines[i]
		}
		if hasRight {
			rightLine = rightLines[i]
		}

		// Check for conflict markers
		if hasLeft && (strings.HasPrefix(leftLine, "<<<<<<<") ||
			strings.HasPrefix(leftLine, "=======") ||
			strings.HasPrefix(leftLine, ">>>>>>>")) {
			mv.cachedLines = append(mv.cachedLines, MergeLine{Text: leftLine, Type: MergeConflictMarker})
			continue
		}
		if hasRight && (strings.HasPrefix(rightLine, "<<<<<<<") ||
			strings.HasPrefix(rightLine, "=======") ||
			strings.HasPrefix(rightLine, ">>>>>>>")) {
			mv.cachedLines = append(mv.cachedLines, MergeLine{Text: rightLine, Type: MergeConflictMarker})
			continue
		}

		if hasLeft && hasRight {
			if leftLine == rightLine {
				mv.cachedLines = append(mv.cachedLines, MergeLine{Text: leftLine, Type: MergeEqual})
			} else {
				// Both differ — show as removed on left, added on right
				mv.cachedLines = append(mv.cachedLines, MergeLine{Text: leftLine, Type: MergeRemoved, LeftOnly: true})
				mv.cachedLines = append(mv.cachedLines, MergeLine{Text: rightLine, Type: MergeAdded, RightOnly: true})
			}
		} else if hasLeft {
			mv.cachedLines = append(mv.cachedLines, MergeLine{Text: leftLine, Type: MergeRemoved, LeftOnly: true})
		} else if hasRight {
			mv.cachedLines = append(mv.cachedLines, MergeLine{Text: rightLine, Type: MergeAdded, RightOnly: true})
		}
	}
}

// LineCount returns the number of computed diff lines.
func (mv *MergeView) LineCount() int {
	mv.mu.Lock()
	defer mv.mu.Unlock()
	mv.computeLinesLocked()
	return len(mv.cachedLines)
}

// Measure returns the preferred size.
func (mv *MergeView) Measure(cs Constraints) Size {
	mv.mu.Lock()
	mv.computeLinesLocked()
	lineCount := len(mv.cachedLines)
	mv.mu.Unlock()

	w := 60
	h := lineCount + 4 // borders + headers
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

// Paint renders the merge view into the buffer.
func (mv *MergeView) Paint(buf *buffer.Buffer) {
	mv.mu.Lock()
	defer mv.mu.Unlock()

	mv.computeLinesLocked()

	b := mv.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 20 {
		w = 60
	}
	if h < 5 {
		h = 5
	}

	// Draw border
	bs := mv.style.Border
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

	// Column layout: left half | separator | right half
	halfW := (w - 4) / 2 // minus borders + separator
	leftCol := x + 2
	sepCol := leftCol + halfW
	rightCol := sepCol + 2

	// Draw separator
	sepStyle := mv.style.Separator
	for row := 1; row < h-1; row++ {
		cy := y + row
		if sepCol < buf.Width && cy < buf.Height {
			buf.SetCell(sepCol, cy, buffer.Cell{Rune: '│', Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
		}
		if sepCol+1 < buf.Width && cy < buf.Height {
			buf.SetCell(sepCol+1, cy, buffer.Cell{Rune: ' ', Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
		}
	}

	// Draw headers
	headerStyle := mv.style.Header
	// Left header
	for i, r := range mv.leftLabel {
		cx := leftCol + i
		if cx < sepCol && cx < buf.Width && y+1 < buf.Height {
			buf.SetCell(cx, y+1, buffer.Cell{Rune: r, Fg: headerStyle.Fg, Bg: headerStyle.Bg, Flags: headerStyle.Flags, Width: 1})
		}
	}
	// Right header
	for i, r := range mv.rightLabel {
		cx := rightCol + i
		if cx < x+w-1 && cx < buf.Width && y+1 < buf.Height {
			buf.SetCell(cx, y+1, buffer.Cell{Rune: r, Fg: headerStyle.Fg, Bg: headerStyle.Bg, Flags: headerStyle.Flags, Width: 1})
		}
	}

	// Draw diff lines
	for idx, line := range mv.cachedLines {
		rowY := y + 2 + idx
		if rowY >= y+h-1 || rowY >= buf.Height {
			break
		}

		var style buffer.Style
		var prefix rune
		switch line.Type {
		case MergeEqual:
			style = mv.style.Equal
			prefix = ' '
		case MergeAdded:
			style = mv.style.Added
			prefix = '+'
		case MergeRemoved:
			style = mv.style.Removed
			prefix = '-'
		case MergeConflictMarker:
			style = mv.style.Conflict
			prefix = '!'
		}

		// Draw on appropriate side
		targetCol := leftCol
		if line.RightOnly {
			targetCol = rightCol
		}

		// Prefix character
		if targetCol < buf.Width && targetCol < (leftCol+halfW) || targetCol < x+w-1 {
			buf.SetCell(targetCol, rowY, buffer.Cell{Rune: prefix, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
		}

		// Line text
		col := targetCol + 1
		for _, r := range line.Text {
			maxCol := targetCol + halfW
			if line.RightOnly {
				maxCol = x + w - 2
			}
			if col >= maxCol || col >= buf.Width {
				break
			}
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
			col++
		}

		// If equal lines, also draw on right side
		if line.Type == MergeEqual {
			col2 := rightCol + 1
			for _, r := range line.Text {
				if col2 >= x+w-1 || col2 >= buf.Width {
					break
				}
				buf.SetCell(col2, rowY, buffer.Cell{Rune: r, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
				col2++
			}
			// Draw space prefix on right
			if rightCol < buf.Width {
				buf.SetCell(rightCol, rowY, buffer.Cell{Rune: ' ', Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
			}
		}
	}
}

// Children returns nil.
func (mv *MergeView) Children() []Component { return nil }

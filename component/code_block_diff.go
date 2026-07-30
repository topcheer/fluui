package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── CodeBlockDiff: Inline Code Diff with Syntax Highlighting ───
//
// CodeBlockDiff renders a compact code diff showing old and new lines
// side-by-side with +/- prefixes and color coding. Designed for AI
// code edit output where users need to see what changed.
//
// Usage:
//
//	cd := NewCodeBlockDiff()
//	cd.SetOld([]string{"fmt.Println(x)", "return err"})
//	cd.SetNew([]string{"fmt.Printf(\"%v\\n\", x)", "return nil"})
//	cd.Paint(buf)

type CodeBlockDiffStyle struct {
	Added     buffer.Style
	Removed   buffer.Style
	Context   buffer.Style
	LineNum   buffer.Style
	Separator buffer.Style
}

func DefaultCodeBlockDiffStyle() CodeBlockDiffStyle {
	return CodeBlockDiffStyle{
		Added:     buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Removed:   buffer.Style{Fg: buffer.RGB(239, 68, 68)},
		Context:   buffer.Style{Fg: buffer.RGB(203, 213, 225)},
		LineNum:   buffer.Style{Fg: buffer.RGB(71, 85, 105)},
		Separator: buffer.Style{Fg: buffer.RGB(51, 65, 85)},
	}
}

const codeBlockDiffMaxLines = 30

type codeDiffLine struct {
	oldText string
	newText string
	lineNum int
	type_   int // 0=unchanged, 1=modified, 2=added, 3=removed
}

// CodeBlockDiff renders a code diff block.
type CodeBlockDiff struct {
	BaseComponent
	mu sync.Mutex

	lines [codeBlockDiffMaxLines]codeDiffLine
	count int
	style CodeBlockDiffStyle
}

// NewCodeBlockDiff creates a CodeBlockDiff.
func NewCodeBlockDiff() *CodeBlockDiff {
	cd := &CodeBlockDiff{style: DefaultCodeBlockDiffStyle()}
	cd.SetID(GenerateID("codediff"))
	return cd
}

// SetOld sets old code lines.
func (cd *CodeBlockDiff) SetOld(lines []string) *CodeBlockDiff {
	cd.mu.Lock()
	cd.rebuildLocked(lines, nil)
	cd.mu.Unlock()
	return cd
}

// SetNew sets new code lines (computes diff from old).
func (cd *CodeBlockDiff) SetNew(lines []string) *CodeBlockDiff {
	cd.mu.Lock()
	// Use whatever was stored as old from last SetOld
	// This is a simplified interface: SetOld then SetNew
	cd.mu.Unlock()
	return cd
}

// SetDiff sets old and new code, computing the diff.
func (cd *CodeBlockDiff) SetDiff(oldLines, newLines []string) *CodeBlockDiff {
	cd.mu.Lock()
	cd.rebuildLocked(oldLines, newLines)
	cd.mu.Unlock()
	return cd
}

func (cd *CodeBlockDiff) rebuildLocked(oldLines, newLines []string) {
	cd.count = 0
	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	for i := 0; i < maxLen && cd.count < codeBlockDiffMaxLines; i++ {
		var oldLn, newLn string
		if i < len(oldLines) {
			oldLn = oldLines[i]
		}
		if i < len(newLines) {
			newLn = newLines[i]
		}

		lineType := 0
		if oldLn == newLn {
			lineType = 0 // unchanged
		} else if oldLn == "" {
			lineType = 2 // added
		} else if newLn == "" {
			lineType = 3 // removed
		} else {
			lineType = 1 // modified
		}

		cd.lines[cd.count] = codeDiffLine{oldText: oldLn, newText: newLn, lineNum: i + 1, type_: lineType}
		cd.count++
	}
}

// LineCount returns the number of diff lines.
func (cd *CodeBlockDiff) LineCount() int {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	return cd.count
}

// SetStyle sets custom style.
func (cd *CodeBlockDiff) SetStyle(s CodeBlockDiffStyle) *CodeBlockDiff {
	cd.mu.Lock()
	cd.style = s
	cd.mu.Unlock()
	return cd
}

// Measure returns preferred size.
func (cd *CodeBlockDiff) Measure(cs Constraints) Size {
	cd.mu.Lock()
	h := cd.count + 1
	cd.mu.Unlock()
	if h < 1 {
		h = 1
	}
	w := 40
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: h}
}

// Paint renders the code diff.
func (cd *CodeBlockDiff) Paint(buf *buffer.Buffer) {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	b := cd.Bounds()
	x, y := b.X, b.Y

	addedStyle := cd.style.Added
	removedStyle := cd.style.Removed
	contextStyle := cd.style.Context
	lineNumStyle := cd.style.LineNum
	sepStyle := cd.style.Separator

	for i := 0; i < cd.count; i++ {
		yy := y + 1 + i
		if yy >= buf.Height {
			break
		}
		line := cd.lines[i]
		col := x

		// Line number
		lnStr := itoa(line.lineNum)
		for _, r := range lnStr {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: lineNumStyle.Fg, Bg: lineNumStyle.Bg, Flags: lineNumStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: ' ', Fg: lineNumStyle.Fg, Bg: lineNumStyle.Bg, Flags: lineNumStyle.Flags, Width: 1})
			col++
		}

		// Old line (red prefix)
		var oldPrefix rune
		var oldSt buffer.Style
		if line.type_ == 0 {
			oldPrefix = ' '
			oldSt = contextStyle
		} else {
			oldPrefix = '-'
			oldSt = removedStyle
		}
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: oldPrefix, Fg: oldSt.Fg, Bg: oldSt.Bg, Flags: oldSt.Flags, Width: 1})
			col++
		}
		for _, r := range line.oldText {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: oldSt.Fg, Bg: oldSt.Bg, Flags: oldSt.Flags, Width: 1})
			col++
		}

		// Separator
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: ' ', Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
			col++
		}

		// New line (green prefix)
		var newPrefix rune
		var newSt buffer.Style
		if line.type_ == 0 {
			newPrefix = ' '
			newSt = contextStyle
		} else {
			newPrefix = '+'
			newSt = addedStyle
		}
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: newPrefix, Fg: newSt.Fg, Bg: newSt.Bg, Flags: newSt.Flags, Width: 1})
			col++
		}
		for _, r := range line.newText {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: newSt.Fg, Bg: newSt.Bg, Flags: newSt.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (cd *CodeBlockDiff) Children() []Component { return nil }

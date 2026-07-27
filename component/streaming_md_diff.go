package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── StreamingMarkdownDiff: AI Live Editing Diff View ───
//
// StreamingMarkdownDiff renders a side-by-side or unified diff of markdown
// text as it's being edited by an AI. It highlights insertions (green +)
// and deletions (red -) in real-time, common in AI code editors and
// document revision tools.
//
// Usage:
//
//	d := NewStreamingMarkdownDiff()
//	d.SetOld("Hello world\nThis is old text.")
//	d.SetNew("Hello world\nThis is **new** text.")
//	d.Paint(buf) // shows unified diff with +/- markers

// DiffLineType classifies a diff line.
type DiffLineType int

const (
	DiffLineContext DiffLineType = iota // unchanged (both old and new)
	DiffLineAdded                       // only in new
	DiffLineRemoved                     // only in old
)

// SMDiffLine represents a single line in the diff output.
type SMDiffLine struct {
	Type    DiffLineType
	Content string
	OldLine int // 1-based line number in old text (0 if added)
	NewLine int // 1-based line number in new text (0 if removed)
}

// StreamingMarkdownDiffStyle holds visual styles.
type StreamingMarkdownDiffStyle struct {
	Added     buffer.Style
	Removed   buffer.Style
	Context   buffer.Style
	LineNumber buffer.Style
	Header    buffer.Style
	Separator buffer.Style
}

// DefaultStreamingMarkdownDiffStyle returns sensible defaults.
func DefaultStreamingMarkdownDiffStyle() StreamingMarkdownDiffStyle {
	return StreamingMarkdownDiffStyle{
		Added:      buffer.Style{Fg: buffer.RGB(16, 163, 127)},   // green
		Removed:    buffer.Style{Fg: buffer.RGB(220, 80, 80)},    // red
		Context:    buffer.Style{Fg: buffer.RGB(180, 180, 180)},  // gray
		LineNumber: buffer.Style{Fg: buffer.RGB(100, 100, 100)},  // dim
		Header:     buffer.Style{Fg: buffer.White, Flags: buffer.Bold},
		Separator:  buffer.Style{Fg: buffer.RGB(60, 60, 60)},
	}
}

// StreamingMarkdownDiff renders a live diff of text being edited by AI.
type StreamingMarkdownDiff struct {
	BaseComponent
	mu          sync.RWMutex
	oldText     string
	newText     string
	cachedLines []SMDiffLine
	cachedOld   string
	cachedNew   string
	style       StreamingMarkdownDiffStyle
	unified     bool // true = unified diff, false = side-by-side
}

// NewStreamingMarkdownDiff creates a unified diff view.
func NewStreamingMarkdownDiff() *StreamingMarkdownDiff {
	d := &StreamingMarkdownDiff{
		style:   DefaultStreamingMarkdownDiffStyle(),
		unified: true,
	}
	d.SetID(GenerateID("smdiff"))
	return d
}

// SetOld sets the original text.
func (d *StreamingMarkdownDiff) SetOld(text string) *StreamingMarkdownDiff {
	d.mu.Lock()
	d.oldText = text
	d.mu.Unlock()
	return d
}

// SetNew sets the revised text (e.g., AI's edited version).
func (d *StreamingMarkdownDiff) SetNew(text string) *StreamingMarkdownDiff {
	d.mu.Lock()
	d.newText = text
	d.mu.Unlock()
	return d
}

// OldText returns the original text.
func (d *StreamingMarkdownDiff) OldText() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.oldText
}

// NewText returns the revised text.
func (d *StreamingMarkdownDiff) NewText() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.newText
}

// SetStyle overrides the visual style.
func (d *StreamingMarkdownDiff) SetStyle(s StreamingMarkdownDiffStyle) *StreamingMarkdownDiff {
	d.mu.Lock()
	d.style = s
	d.mu.Unlock()
	return d
}

// Style returns the current style.
func (d *StreamingMarkdownDiff) Style() StreamingMarkdownDiffStyle {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.style
}

// IsUnified returns true for unified mode, false for side-by-side.
func (d *StreamingMarkdownDiff) IsUnified() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.unified
}

// SetUnified sets the display mode (true=unified, false=side-by-side).
func (d *StreamingMarkdownDiff) SetUnified(u bool) *StreamingMarkdownDiff {
	d.mu.Lock()
	d.unified = u
	d.mu.Unlock()
	return d
}

// DiffLines returns the computed diff lines (rebuilds if needed).
func (d *StreamingMarkdownDiff) DiffLines() []SMDiffLine {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.ensureCacheLocked()
	return d.cachedLines
}

// LineCount returns the number of diff lines.
func (d *StreamingMarkdownDiff) LineCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	d.ensureCacheLocked()
	return len(d.cachedLines)
}

// Stats returns (added, removed, unchanged) line counts.
func (d *StreamingMarkdownDiff) Stats() (added, removed, unchanged int) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	d.ensureCacheLocked()
	for _, l := range d.cachedLines {
		switch l.Type {
		case DiffLineAdded:
			added++
		case DiffLineRemoved:
			removed++
		default:
			unchanged++
		}
	}
	return
}

// ensureCacheLocked rebuilds the diff cache if texts changed (caller holds lock).
func (d *StreamingMarkdownDiff) ensureCacheLocked() {
	if d.cachedOld == d.oldText && d.cachedNew == d.newText {
		return
	}
	d.cachedLines = computeUnifiedDiff(d.oldText, d.newText)
	d.cachedOld = d.oldText
	d.cachedNew = d.newText
}

// computeUnifiedDiff produces a simple line-level diff using LCS.
// This is a zero-dependency implementation suitable for real-time streaming.
func computeUnifiedDiff(oldText, newText string) []SMDiffLine {
	oldLines := smSplitLines(oldText)
	newLines := smSplitLines(newText)

	// LCS table (compact: we only need to backtrack, so use DP matrix)
	n, m := len(oldLines), len(newLines)
	// Build LCS length table
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if oldLines[i-1] == newLines[j-1] {
				lcs[i][j] = lcs[i-1][j-1] + 1
			} else if lcs[i-1][j] >= lcs[i][j-1] {
				lcs[i][j] = lcs[i-1][j]
			} else {
				lcs[i][j] = lcs[i][j-1]
			}
		}
	}

	// Backtrack to produce diff
	type btEntry struct {
		oldLine, newLine int
		typ              DiffLineType
		content          string
	}
	var result []btEntry
	i, j := n, m
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && oldLines[i-1] == newLines[j-1] {
			result = append(result, btEntry{i, j, DiffLineContext, oldLines[i-1]})
			i--
			j--
		} else if j > 0 && (i == 0 || lcs[i][j-1] >= lcs[i-1][j]) {
			result = append(result, btEntry{0, j, DiffLineAdded, newLines[j-1]})
			j--
		} else if i > 0 {
			result = append(result, btEntry{i, 0, DiffLineRemoved, oldLines[i-1]})
			i--
		}
	}

	// Reverse (we backtracked)
	diffLines := make([]SMDiffLine, len(result))
	for k := 0; k < len(result); k++ {
		e := result[len(result)-1-k]
		diffLines[k] = SMDiffLine{
			Type:    e.typ,
			Content: e.content,
			OldLine: e.oldLine,
			NewLine: e.newLine,
		}
	}
	return diffLines
}

// splitLines splits text into lines without trailing newline.
func smSplitLines(text string) []string {
	if text == "" {
		return nil
	}
	lines := strings.Split(text, "\n")
	// Remove trailing empty element if text ends with \n
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// Measure computes the desired size.
func (d *StreamingMarkdownDiff) Measure(cs Constraints) Size {
	d.mu.RLock()
	defer d.mu.RUnlock()
	d.ensureCacheLocked()
	w := 60
	h := len(d.cachedLines) + 2 // header + separator + lines
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

// Paint renders the unified diff.
func (d *StreamingMarkdownDiff) Paint(buf *buffer.Buffer) {
	d.mu.Lock()
	defer d.mu.Unlock()

	b := d.bounds
	if b.W < 4 || b.H < 1 {
		return
	}

	d.ensureCacheLocked()

	row := b.Y

	// Header
	headerW := b.W
	if headerW > 40 {
		headerW = 40
	}
	for i := 0; i < headerW; i++ {
		buf.SetCell(b.X+i, row, buffer.Cell{Rune: '=', Fg: d.style.Header.Fg, Bg: d.style.Header.Bg, Flags: d.style.Header.Flags, Width: 1})
	}
	row++

	for _, dl := range d.cachedLines {
		if row >= b.Y+b.H {
			break
		}

		var prefix rune
		var style buffer.Style
		switch dl.Type {
		case DiffLineAdded:
			prefix = '+'
			style = d.style.Added
		case DiffLineRemoved:
			prefix = '-'
			style = d.style.Removed
		default:
			prefix = ' '
			style = d.style.Context
		}

		x := b.X
		// Prefix marker
		buf.SetCell(x, row, buffer.Cell{Rune: prefix, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
		x++

		// Content (truncate to fit)
		availW := b.X + b.W - x
		if availW < 0 {
			availW = 0
		}
		text := dl.Content
		textRunes := []rune(text)
		if len(textRunes) > availW {
			textRunes = textRunes[:availW]
		}
		for _, r := range textRunes {
			buf.SetCell(x, row, buffer.Cell{Rune: r, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
			x++
		}

		row++
	}
}

// Children returns nil.
func (d *StreamingMarkdownDiff) Children() []Component { return nil }

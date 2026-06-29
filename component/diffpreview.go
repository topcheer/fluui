package component

import (
	"fmt"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// --- Diff types ---

// DiffType classifies the type of a diff line (context, addition, deletion, etc.).
// DiffType classifies a line within a unified diff.
// One of DiffContext, DiffAdd, DiffDel, DiffHunk, DiffFile, or DiffMeta.
type DiffType uint8

const (
	DiffContext DiffType = iota
	DiffAdd
	DiffDel
	DiffHunk
	DiffFile
	DiffMeta
)

// DiffLine represents a single line in a unified diff.
// DiffLine is a single classified line from a unified diff.
type DiffLine struct {
	Type    DiffType
	Content string
}

// ParseDiff parses unified diff text into classified DiffLines.
// ParseDiff parses unified diff text into classified DiffLines.
func ParseDiff(text string) []DiffLine {
	rawLines := strings.Split(text, "\n")
	result := make([]DiffLine, 0, len(rawLines))
	for _, line := range rawLines {
		if line == "" {
			continue
		}
		result = append(result, DiffLine{
			Type:    classifyDiffType(line),
			Content: line,
		})
	}
	return result
}

func classifyDiffType(line string) DiffType {
	switch {
	case strings.HasPrefix(line, "diff --git"):
		return DiffFile
	case strings.HasPrefix(line, "@@"):
		return DiffHunk
	case strings.HasPrefix(line, "index "),
		strings.HasPrefix(line, "---"),
		strings.HasPrefix(line, "+++"),
		strings.HasPrefix(line, "new file"),
		strings.HasPrefix(line, "deleted file"),
		strings.HasPrefix(line, "rename "),
		strings.HasPrefix(line, "old mode"),
		strings.HasPrefix(line, "new mode"),
		strings.HasPrefix(line, "similarity "),
		strings.HasPrefix(line, "copy "):
		return DiffMeta
	case strings.HasPrefix(line, "+"):
		return DiffAdd
	case strings.HasPrefix(line, "-"):
		return DiffDel
	default:
		return DiffContext
	}
}

// --- DiffStats ---

// DiffStats holds summary statistics for a parsed diff.
// DiffStats summarises the additions, deletions, files, and hunks in a diff.
type DiffStats struct {
	Additions  int
	Deletions  int
	Files      int
	Hunks      int
	TotalLines int
}

func (s DiffStats) String() string {
	return fmt.Sprintf("+%d -%d (%d files, %d hunks)",
		s.Additions, s.Deletions, s.Files, s.Hunks)
}

// --- DiffPreviewStyle ---

// DiffPreviewStyle holds style configuration for rendering diff output.
// DiffPreviewStyle holds the styles for each part of a DiffPreview component.
type DiffPreviewStyle struct {
	Border      buffer.Style
	AddLine     buffer.Style
	DelLine     buffer.Style
	ContextLine buffer.Style
	HunkHeader  buffer.Style
	FileHeader  buffer.Style
	MetaLine    buffer.Style
	LineNumber  buffer.Style
	StatsLine   buffer.Style
}

// DefaultDiffPreviewStyle returns a DiffPreviewStyle initialized from the current theme.
// DefaultDiffPreviewStyle returns a DiffPreviewStyle using the current theme.
func DefaultDiffPreviewStyle() DiffPreviewStyle {
	t := theme.Get()
	return DiffPreviewStyle{
		Border:      buffer.Style{Fg: t.Border, Flags: buffer.Dim},
		AddLine:     buffer.Style{Fg: t.DiffAdd},
		DelLine:     buffer.Style{Fg: t.DiffDel},
		ContextLine: buffer.DefaultStyle,
		HunkHeader:  buffer.Style{Fg: t.DiffHunk},
		FileHeader:  buffer.Style{Fg: t.DiffFile, Flags: buffer.Bold},
		MetaLine:    buffer.Style{Fg: t.DiffMeta, Flags: buffer.Dim},
		LineNumber:  buffer.Style{Fg: t.Muted, Flags: buffer.Dim},
		StatsLine:   buffer.Style{Fg: t.Accent, Flags: buffer.Bold},
	}
}

// --- DiffPreview Component ---

// DiffPreview is a component that renders unified diffs with syntax highlighting.
// DiffPreview is a scrollable component that renders unified diff output
// with syntax highlighting for additions, deletions, hunks, and file headers.
type DiffPreview struct {
	BaseComponent
	mu        sync.RWMutex
	lines     []DiffLine
	stats     DiffStats
	scrollY   int
	maxScroll int
	style     DiffPreviewStyle
	title     string
}

// NewDiffPreview creates a new DiffPreview component with default styling.
// NewDiffPreview creates a new DiffPreview component with default styling.
func NewDiffPreview() *DiffPreview {
	dp := &DiffPreview{
		lines: make([]DiffLine, 0),
		style: DefaultDiffPreviewStyle(),
	}
	dp.SetID(GenerateID("diffpreview"))
	return dp
}

func (dp *DiffPreview) SetDiff(text string) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.lines = ParseDiff(text)
	dp.stats = computeDiffStats(dp.lines)
	dp.scrollY = 0
	dp.maxScroll = 0
}

func (dp *DiffPreview) Lines() []DiffLine {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	out := make([]DiffLine, len(dp.lines))
	copy(out, dp.lines)
	return out
}

func (dp *DiffPreview) LineCount() int {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return len(dp.lines)
}

func (dp *DiffPreview) Stats() DiffStats {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return dp.stats
}

func (dp *DiffPreview) IsEmpty() bool {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return len(dp.lines) == 0
}

func (dp *DiffPreview) HasChanges() bool {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return dp.stats.Additions > 0 || dp.stats.Deletions > 0
}

func (dp *DiffPreview) DiffSummary() string {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return dp.stats.String()
}

func (dp *DiffPreview) SetStyle(s DiffPreviewStyle) {
	dp.mu.Lock()
	dp.style = s
	dp.mu.Unlock()
}

func (dp *DiffPreview) Style() DiffPreviewStyle {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return dp.style
}

func (dp *DiffPreview) SetTitle(title string) {
	dp.mu.Lock()
	dp.title = title
	dp.mu.Unlock()
}

func (dp *DiffPreview) Title() string {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return dp.title
}

// --- Scrolling ---

func (dp *DiffPreview) ScrollY() int {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return dp.scrollY
}

func (dp *DiffPreview) ScrollDown(n int) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.clampScrollLocked()
	dp.scrollY += n
	dp.clampScrollLocked()
}

func (dp *DiffPreview) ScrollUp(n int) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.scrollY -= n
	if dp.scrollY < 0 {
		dp.scrollY = 0
	}
}

func (dp *DiffPreview) ScrollTo(row int) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	if row < 0 {
		row = 0
	}
	dp.scrollY = row
	dp.clampScrollLocked()
}

func (dp *DiffPreview) ScrollPageDown(viewHeight int) {
	dp.ScrollDown(viewHeight)
}

func (dp *DiffPreview) ScrollPageUp(viewHeight int) {
	dp.ScrollUp(viewHeight)
}

func (dp *DiffPreview) VisibleRange() (int, int) {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	b := dp.bounds
	availableH := b.H - 2
	if availableH < 0 {
		availableH = 0
	}
	end := dp.scrollY + availableH
	if end > len(dp.lines) {
		end = len(dp.lines)
	}
	return dp.scrollY, end
}

func (dp *DiffPreview) SetShowLineNumbers(show bool) {}

func (dp *DiffPreview) ShowLineNumbers() bool { return true }

func (dp *DiffPreview) SetShowStats(show bool) {}

func (dp *DiffPreview) SetLines(lines []DiffLine) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.lines = lines
	dp.stats = computeDiffStats(lines)
	dp.scrollY = 0
	dp.maxScroll = 0
}

func (dp *DiffPreview) clampScrollLocked() {
	b := dp.bounds
	availableH := b.H - 2
	if availableH < 1 {
		availableH = 1
	}
	total := len(dp.lines)
	if total <= availableH {
		dp.maxScroll = 0
	} else {
		dp.maxScroll = total - availableH
	}
	if dp.scrollY > dp.maxScroll {
		dp.scrollY = dp.maxScroll
	}
	if dp.scrollY < 0 {
		dp.scrollY = 0
	}
}

// --- Component Interface ---

func (dp *DiffPreview) Measure(cs Constraints) Size {
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	w := 80
	for _, l := range dp.lines {
		lw := buffer.StringWidth(l.Content) + 4
		if lw > w {
			w = lw
		}
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}

	h := len(dp.lines) + 2
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	if h < 3 {
		h = 3
	}
	return Size{W: w, H: h}
}

func (dp *DiffPreview) SetBounds(r Rect) {
	dp.mu.Lock()
	dp.bounds = r
	dp.clampScrollLocked()
	dp.mu.Unlock()
}

func (dp *DiffPreview) Paint(buf *buffer.Buffer) {
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	b := dp.bounds
	if b.W < 3 || b.H < 3 || buf == nil {
		return
	}

	dp.paintBorderLocked(buf, b)

	innerX := b.X + 1
	innerH := b.H - 2
	y := b.Y + 1

	end := dp.scrollY + innerH
	if end > len(dp.lines) {
		end = len(dp.lines)
	}

	for i := dp.scrollY; i < end; i++ {
		line := dp.lines[i]
		style := dp.diffStyleForLocked(line.Type)
		buf.DrawTextClamped(innerX, y, line.Content, style)
		y++
	}
}

func (dp *DiffPreview) paintBorderLocked(buf *buffer.Buffer, b Rect) {
	style := dp.style.Border
	buf.SetCell(b.X, b.Y, buffer.NewCell('\u250C', style))
	buf.SetCell(b.X+b.W-1, b.Y, buffer.NewCell('\u2510', style))
	for x := b.X + 1; x < b.X+b.W-1; x++ {
		buf.SetCell(x, b.Y, buffer.NewCell('\u2500', style))
	}
	if dp.title != "" {
		titleText := " " + dp.title + " "
		if dp.stats.Additions > 0 || dp.stats.Deletions > 0 {
			titleText = fmt.Sprintf(" %s +%d/-%d ", dp.title, dp.stats.Additions, dp.stats.Deletions)
		}
		buf.DrawText(b.X+2, b.Y, titleText, style)
	}
	buf.SetCell(b.X, b.Y+b.H-1, buffer.NewCell('\u2514', style))
	buf.SetCell(b.X+b.W-1, b.Y+b.H-1, buffer.NewCell('\u2518', style))
	for x := b.X + 1; x < b.X+b.W-1; x++ {
		buf.SetCell(x, b.Y+b.H-1, buffer.NewCell('\u2500', style))
	}
	for y := b.Y + 1; y < b.Y+b.H-1; y++ {
		buf.SetCell(b.X, y, buffer.NewCell('\u2502', style))
		buf.SetCell(b.X+b.W-1, y, buffer.NewCell('\u2502', style))
	}
}

func (dp *DiffPreview) diffStyleForLocked(dt DiffType) buffer.Style {
	switch dt {
	case DiffAdd:
		return dp.style.AddLine
	case DiffDel:
		return dp.style.DelLine
	case DiffHunk:
		return dp.style.HunkHeader
	case DiffFile:
		return dp.style.FileHeader
	case DiffMeta:
		return dp.style.MetaLine
	default:
		return dp.style.ContextLine
	}
}

func (dp *DiffPreview) Children() []Component { return nil }

func (dp *DiffPreview) String() string {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	return fmt.Sprintf("DiffPreview{lines:%d +%d/-%d scroll:%d}",
		len(dp.lines), dp.stats.Additions, dp.stats.Deletions, dp.scrollY)
}

// --- Helpers ---

func computeDiffStats(lines []DiffLine) DiffStats {
	s := DiffStats{TotalLines: len(lines)}
	fileSeen := make(map[string]bool)
	for _, line := range lines {
		switch line.Type {
		case DiffAdd:
			s.Additions++
		case DiffDel:
			s.Deletions++
		case DiffHunk:
			s.Hunks++
		case DiffFile:
			if file := extractDiffFilename(line.Content); file != "" && !fileSeen[file] {
				fileSeen[file] = true
				s.Files++
			}
		}
	}
	return s
}

func extractDiffFilename(line string) string {
	parts := strings.SplitN(line, " b/", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// DiffViewer renders a unified diff with syntax-highlighted add/remove/context lines.
// It supports scrolling, line numbers, and customizable colors.
type DiffViewer struct {
	BaseComponent
	mu sync.RWMutex

	lines      []DiffLine
	scrollOffset int
	maxWidth   int

	// display options
	showLineNumbers bool
	showHeader     bool

	// styles
	styleAdd     buffer.Style
	styleDel     buffer.Style
	styleContext buffer.Style
	styleHunk    buffer.Style
	styleMeta    buffer.Style
	styleLineNum buffer.Style
	styleHeader  buffer.Style

	// title / filename
	title string

	// computed bounds
	bounds Rect
}

// NewDiffViewer creates a DiffViewer with sensible default styles.
func NewDiffViewer() *DiffViewer {
	return &DiffViewer{
		showLineNumbers: true,
		showHeader:      true,
		styleAdd: buffer.Style{
			Fg: buffer.NamedColor(buffer.NamedGreen),
		},
		styleDel: buffer.Style{
			Fg: buffer.NamedColor(buffer.NamedRed),
		},
		styleContext: buffer.Style{
			Fg: buffer.NamedColor(buffer.NamedWhite),
		},
		styleHunk: buffer.Style{
			Fg:    buffer.NamedColor(buffer.NamedCyan),
			Flags: buffer.Bold,
		},
		styleMeta: buffer.Style{
			Fg:    buffer.NamedColor(buffer.NamedYellow),
			Flags: buffer.Dim,
		},
		styleLineNum: buffer.Style{
			Fg:    buffer.NamedColor(buffer.NamedBrightBlack),
			Flags: buffer.Dim,
		},
		styleHeader: buffer.Style{
			Fg:    buffer.NamedColor(buffer.NamedYellow),
			Flags: buffer.Bold,
		},
	}
}

// SetContent parses a unified diff string and populates the viewer.
func (d *DiffViewer) SetContent(diff string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.lines = ParseDiffWithLineNums(diff)
	d.scrollOffset = 0
	d.maxWidth = 0
	for _, l := range d.lines {
		w := len([]rune(l.Content))
		if w > d.maxWidth {
			d.maxWidth = w
		}
	}
}

// SetLines sets the diff lines directly (bypassing parsing).
func (d *DiffViewer) SetLines(lines []DiffLine) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.lines = make([]DiffLine, len(lines))
	copy(d.lines, lines)
	d.scrollOffset = 0
	d.maxWidth = 0
	for _, l := range d.lines {
		w := len([]rune(l.Content))
		if w > d.maxWidth {
			d.maxWidth = w
		}
	}
}

// Lines returns a copy of the current diff lines.
func (d *DiffViewer) Lines() []DiffLine {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]DiffLine, len(d.lines))
	copy(out, d.lines)
	return out
}

// LineCount returns the total number of diff lines.
func (d *DiffViewer) LineCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.lines)
}

// ScrollOffset returns the current vertical scroll position.
func (d *DiffViewer) ScrollOffset() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.scrollOffset
}

// ScrollDown scrolls the viewport down by n lines.
func (d *DiffViewer) ScrollDown(n int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if n <= 0 {
		return
	}
	maxScroll := d.maxScrollOffsetLocked()
	d.scrollOffset += n
	if d.scrollOffset > maxScroll {
		d.scrollOffset = maxScroll
	}
}

// ScrollUp scrolls the viewport up by n lines.
func (d *DiffViewer) ScrollUp(n int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if n <= 0 {
		return
	}
	d.scrollOffset -= n
	if d.scrollOffset < 0 {
		d.scrollOffset = 0
	}
}

// ScrollTo sets the scroll offset to the given position (clamped).
func (d *DiffViewer) ScrollTo(offset int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	maxScroll := d.maxScrollOffsetLocked()
	if offset < 0 {
		offset = 0
	}
	if offset > maxScroll {
		offset = maxScroll
	}
	d.scrollOffset = offset
}

// SetShowLineNumbers toggles line number display.
func (d *DiffViewer) SetShowLineNumbers(show bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.showLineNumbers = show
}

// ShowLineNumbers returns whether line numbers are displayed.
func (d *DiffViewer) ShowLineNumbers() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.showLineNumbers
}

// SetShowHeader toggles diff header display.
func (d *DiffViewer) SetShowHeader(show bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.showHeader = show
}

// SetTitle sets the title displayed at the top.
func (d *DiffViewer) SetTitle(title string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.title = title
}

// Title returns the current title.
func (d *DiffViewer) Title() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.title
}

// SetStyleAdd sets the style for added lines (+).
func (d *DiffViewer) SetStyleAdd(s buffer.Style) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.styleAdd = s
}

// SetStyleDel sets the style for deleted lines (-).
func (d *DiffViewer) SetStyleDel(s buffer.Style) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.styleDel = s
}

// SetStyleContext sets the style for context lines.
func (d *DiffViewer) SetStyleContext(s buffer.Style) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.styleContext = s
}

// SetStyleHunk sets the style for hunk headers (@@).
func (d *DiffViewer) SetStyleHunk(s buffer.Style) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.styleHunk = s
}

// SetStyleMeta sets the style for meta lines (diff --git, index).
func (d *DiffViewer) SetStyleMeta(s buffer.Style) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.styleMeta = s
}

// SetStyleLineNum sets the style for line numbers.
func (d *DiffViewer) SetStyleLineNum(s buffer.Style) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.styleLineNum = s
}

// HandleKey processes keyboard input for scrolling.
func (d *DiffViewer) HandleKey(k *term.KeyEvent) bool {
	if k == nil {
		return false
	}
	switch k.Key {
	case term.KeyDown:
		d.ScrollDown(1)
		return true
	case term.KeyUp:
		d.ScrollUp(1)
		return true
	case term.KeyPageDown:
		visH := d.visibleHeight()
		if visH <= 0 {
			visH = 10
		}
		d.ScrollDown(visH)
		return true
	case term.KeyPageUp:
		visH := d.visibleHeight()
		if visH <= 0 {
			visH = 10
		}
		d.ScrollUp(visH)
		return true
	}
	// vim-style keys
	if k.Rune != 0 {
		switch k.Rune {
		case 'j':
			d.ScrollDown(1)
			return true
		case 'k':
			d.ScrollUp(1)
			return true
		case 'g':
			d.ScrollTo(0)
			return true
		case 'G':
			d.ScrollTo(d.LineCount())
			return true
		}
	}
	return false
}

// Measure returns the desired size for the diff viewer.
func (d *DiffViewer) Measure(cs Constraints) Size {
	d.mu.RLock()
	defer d.mu.RUnlock()

	w := d.maxWidth
	if d.showLineNumbers {
		w += 16 // space for old/new line numbers
	}
	w += 2 // sign column + margin

	h := len(d.lines)
	if d.title != "" {
		h++
	}

	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.HasHeight() && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	return Size{W: w, H: h}
}

// SetBounds sets the position and size.
func (d *DiffViewer) SetBounds(r Rect) {
	d.mu.Lock()
	d.bounds = r
	d.mu.Unlock()
}

// Bounds returns the current bounds.
func (d *DiffViewer) Bounds() Rect {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.bounds
}

// Paint renders the diff into the buffer.
func (d *DiffViewer) Paint(buf *buffer.Buffer) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	bounds := d.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	row := 0

	// Draw title bar
	if d.title != "" && row < bounds.H {
		d.paintTitleBar(buf, bounds, row, d.title)
		row++
	}

	// Calculate visible lines
	visibleH := bounds.H - row
	if visibleH <= 0 {
		return
	}

	start := d.scrollOffset
	if start >= len(d.lines) {
		return
	}
	end := start + visibleH
	if end > len(d.lines) {
		end = len(d.lines)
	}

	for i := start; i < end; i++ {
		y := bounds.Y + row
		d.paintLine(buf, bounds, y, d.lines[i])
		row++
	}
}

// --- Internal paint helpers ---

func (d *DiffViewer) paintTitleBar(buf *buffer.Buffer, bounds Rect, row int, title string) {
	y := bounds.Y + row
	runes := []rune(title)
	x := bounds.X
	for i := 0; i < bounds.W; i++ {
		var r rune
		if i < len(runes) {
			r = runes[i]
		} else {
			r = ' '
		}
		buf.SetCell(x+i, y, buffer.NewCell(r, d.styleHeader))
	}
}

func (d *DiffViewer) paintLine(buf *buffer.Buffer, bounds Rect, y int, line DiffLine) {
	var style buffer.Style
	var sign rune

	switch line.Type {
	case DiffAdd:
		style = d.styleAdd
		sign = '+'
	case DiffDel:
		style = d.styleDel
		sign = '-'
	case DiffHunk:
		style = d.styleHunk
		sign = '@'
	case DiffFile, DiffMeta:
		style = d.styleMeta
		sign = ' '
	default:
		style = d.styleContext
		sign = ' '
	}

	x := bounds.X

	// Draw line numbers if enabled
	if d.showLineNumbers {
		oldStr := formatLineNum(line.OldNo)
		newStr := formatLineNum(line.NewNo)
		for i, r := range []rune(oldStr) {
			if x+i < bounds.X+bounds.W {
				buf.SetCell(x+i, y, buffer.NewCell(r, d.styleLineNum))
			}
		}
		x += 7
		for i, r := range []rune(newStr) {
			if x+i < bounds.X+bounds.W {
				buf.SetCell(x+i, y, buffer.NewCell(r, d.styleLineNum))
			}
		}
		x += 8
	}

	// Sign column
	if x < bounds.X+bounds.W {
		buf.SetCell(x, y, buffer.NewCell(sign, style))
	}
	x++

	// Content
	contentRunes := []rune(line.Content)
	maxChars := bounds.W - (x - bounds.X)
	for i := 0; i < maxChars && i < len(contentRunes); i++ {
		buf.SetCell(x+i, y, buffer.NewCell(contentRunes[i], style))
	}
}

// --- Internal helpers ---

func (d *DiffViewer) maxScrollOffsetLocked() int {
	bounds := d.bounds
	visibleH := bounds.H
	if d.title != "" {
		visibleH--
	}
	if visibleH <= 0 {
		return 0
	}
	maxScroll := len(d.lines) - visibleH
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

func (d *DiffViewer) visibleHeight() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	h := d.bounds.H
	if d.title != "" {
		h--
	}
	return h
}

// formatLineNum formats a line number into a fixed-width string.
func formatLineNum(n int) string {
	if n <= 0 {
		return "      " // 6 spaces
	}
	s := ""
	if n < 10 {
		s = "     " + itoa(n)
	} else if n < 100 {
		s = "    " + itoa(n)
	} else if n < 1000 {
		s = "   " + itoa(n)
	} else if n < 10000 {
		s = "  " + itoa(n)
	} else if n < 100000 {
		s = " " + itoa(n)
	} else {
		s = itoa(n)
	}
	return s
}

// ParseDiffWithLineNums parses a unified diff string into DiffLines with line numbers.
func ParseDiffWithLineNums(diff string) []DiffLine {
	var lines []DiffLine
	scanner := newDiffLineScanner(diff)
	oldNo, newNo := 0, 0

	for scanner.Scan() {
		raw := scanner.Text()
		dl := DiffLine{Content: raw}

		switch {
		case strings.HasPrefix(raw, "@@"):
			dl.Type = DiffHunk
			oldNo, newNo = parseHunkHeader(raw)
		case strings.HasPrefix(raw, "diff ") || strings.HasPrefix(raw, "index "):
			dl.Type = DiffFile
		case strings.HasPrefix(raw, "--- ") || strings.HasPrefix(raw, "+++ "):
			dl.Type = DiffMeta
		case len(raw) > 0 && raw[0] == '+':
			dl.Type = DiffAdd
			dl.Content = raw[1:]
			dl.NewNo = newNo
			newNo++
		case len(raw) > 0 && raw[0] == '-':
			dl.Type = DiffDel
			dl.Content = raw[1:]
			dl.OldNo = oldNo
			oldNo++
		case len(raw) > 0 && raw[0] == ' ':
			dl.Type = DiffContext
			dl.Content = raw[1:]
			dl.OldNo = oldNo
			dl.NewNo = newNo
			oldNo++
			newNo++
		case raw == "":
			dl.Type = DiffContext
			dl.Content = ""
			dl.OldNo = oldNo
			dl.NewNo = newNo
			oldNo++
			newNo++
		default:
			dl.Type = DiffMeta
		}

		lines = append(lines, dl)
	}

	return lines
}

// parseHunkHeader extracts the starting old/new line numbers from "@@ -a,b +c,d @@".
func parseHunkHeader(line string) (oldStart, newStart int) {
	minusIdx := strings.Index(line, "-")
	plusIdx := strings.Index(line, "+")
	if minusIdx < 0 || plusIdx < 0 {
		return 0, 0
	}

	oldStart = parseIntFrom(line, minusIdx+1)
	newStart = parseIntFrom(line, plusIdx+1)
	return
}

// parseIntFrom reads an integer starting at the given index.
func parseIntFrom(s string, start int) int {
	result := 0
	i := start
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		result = result*10 + int(s[i]-'0')
		i++
	}
	return result
}

// diffLineScanner is a lightweight line scanner for diffs.
type diffLineScanner struct {
	text   string
	pos    int
	done   bool
	result string
}

func newDiffLineScanner(text string) *diffLineScanner {
	return &diffLineScanner{text: text}
}

func (ls *diffLineScanner) Scan() bool {
	if ls.done {
		return false
	}
	if ls.pos >= len(ls.text) {
		ls.done = true
		return false
	}
	end := ls.pos
	for end < len(ls.text) && ls.text[end] != '\n' {
		end++
	}
	ls.result = ls.text[ls.pos:end]
	if len(ls.result) > 0 && ls.result[len(ls.result)-1] == '\r' {
		ls.result = ls.result[:len(ls.result)-1]
	}
	ls.pos = end + 1
	return true
}

func (ls *diffLineScanner) Text() string {
	return ls.result
}

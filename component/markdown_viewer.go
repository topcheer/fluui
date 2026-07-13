package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/markdown"
)

// MarkdownViewerStyle holds styling for the MarkdownViewer.
type MarkdownViewerStyle struct {
	TitleFg    buffer.Color
	TitleBg    buffer.Color
	BorderFg   buffer.Color
	ContentFg  buffer.Color
	TocFg      buffer.Color
	TocActiveFg buffer.Color
	TocBg      buffer.Color
}

// DefaultMarkdownViewerStyle returns a Dracula-themed style.
func DefaultMarkdownViewerStyle() MarkdownViewerStyle {
	return MarkdownViewerStyle{
		TitleFg:     buffer.NamedColor(buffer.NamedWhite),
		TitleBg:     buffer.NamedColor(buffer.NamedBlue),
		BorderFg:    buffer.NamedColor(buffer.NamedBrightBlack),
		ContentFg:   buffer.NamedColor(buffer.NamedWhite),
		TocFg:       buffer.NamedColor(buffer.NamedBrightBlack),
		TocActiveFg: buffer.NamedColor(buffer.NamedYellow),
		TocBg:       buffer.NamedColor(buffer.NamedBlack),
	}
}

// TocEntry represents a table of contents entry.
type TocEntry struct {
	Level int
	Text  string
	Line  int // line index in rendered content
}

// MarkdownViewer is an interactive markdown viewer with table of contents,
// scrollable content, and keyboard navigation.
// It wraps the markdown.Renderer with a scrollable view and optional TOC sidebar.
type MarkdownViewer struct {
	BaseComponent

	source    string
	renderer  *markdown.MarkdownRenderer
	style     MarkdownViewerStyle
	title     string

	// Rendered content
	blocks    []*markdown.Block
	lines     [][]buffer.Cell
	totalLines int

	// Scrolling
	scrollY   int
	maxScroll int

	// TOC
	toc       []TocEntry
	tocCursor int
	showToc   bool
	tocWidth  int

	// Layout
	contentW int

	mu sync.RWMutex
}

// NewMarkdownViewer creates a viewer with the given markdown source.
func NewMarkdownViewer(source string) *MarkdownViewer {
	v := &MarkdownViewer{
		source:   source,
		style:    DefaultMarkdownViewerStyle(),
		title:    "Markdown",
		tocWidth: 20,
	}
	return v
}

// SetSource sets new markdown content and re-renders.
func (v *MarkdownViewer) SetSource(source string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.source = source
	v.scrollY = 0
	v.renderLocked()
}

// SetBounds sets the component bounds and re-renders content for the new width.
func (v *MarkdownViewer) SetBounds(r Rect) {
	v.BaseComponent.SetBounds(r)
	v.mu.Lock()
	defer v.mu.Unlock()
	v.renderLocked()
}

// SetTitle sets the viewer title.
func (v *MarkdownViewer) SetTitle(title string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.title = title
}

// SetStyle sets the visual style.
func (v *MarkdownViewer) SetStyle(s MarkdownViewerStyle) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.style = s
}

// ShowToc shows the table of contents sidebar.
func (v *MarkdownViewer) ShowToc() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.showToc = true
}

// HideToc hides the table of contents sidebar.
func (v *MarkdownViewer) HideToc() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.showToc = false
}

// ToggleToc toggles TOC visibility.
func (v *MarkdownViewer) ToggleToc() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.showToc = !v.showToc
}

// ScrollUp scrolls content up by n lines.
func (v *MarkdownViewer) ScrollUp(n int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.scrollY -= n
	if v.scrollY < 0 {
		v.scrollY = 0
	}
}

// ScrollDown scrolls content down by n lines.
func (v *MarkdownViewer) ScrollDown(n int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.scrollY += n
	if v.scrollY > v.maxScroll {
		v.scrollY = v.maxScroll
	}
}

// ScrollToTop scrolls to the beginning.
func (v *MarkdownViewer) ScrollToTop() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.scrollY = 0
}

// ScrollToBottom scrolls to the end.
func (v *MarkdownViewer) ScrollToBottom() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.scrollY = v.maxScroll
}

// ScrollToTocEntry scrolls to a specific TOC entry.
func (v *MarkdownViewer) ScrollToTocEntry(idx int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if idx < 0 || idx >= len(v.toc) {
		return
	}
	v.tocCursor = idx
	v.scrollY = v.toc[idx].Line
	if v.scrollY > v.maxScroll {
		v.scrollY = v.maxScroll
	}
}

// TocEntries returns the table of contents entries.
func (v *MarkdownViewer) TocEntries() []TocEntry {
	v.mu.RLock()
	defer v.mu.RUnlock()
	result := make([]TocEntry, len(v.toc))
	copy(result, v.toc)
	return result
}

// ScrollY returns the current scroll position.
func (v *MarkdownViewer) ScrollY() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.scrollY
}

// TotalLines returns the total number of rendered lines.
func (v *MarkdownViewer) TotalLines() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.totalLines
}

// HandleKey handles keyboard navigation.
func (v *MarkdownViewer) HandleKey(ev *term.KeyEvent) bool {
	if ev == nil {
		return false
	}
	switch ev.Key {
	case term.KeyUp:
		v.ScrollUp(1)
		return true
	case term.KeyDown:
		v.ScrollDown(1)
		return true
	case term.KeyPageUp:
		v.ScrollUp(10)
		return true
	case term.KeyPageDown:
		v.ScrollDown(10)
		return true
	case term.KeyHome:
		v.ScrollToTop()
		return true
	case term.KeyEnd:
		v.ScrollToBottom()
		return true
	case term.KeyEscape:
		v.mu.Lock()
		if v.showToc {
			v.showToc = false
			v.mu.Unlock()
			return true
		}
		v.mu.Unlock()
		return false
	case term.KeyTab:
		v.ToggleToc()
		return true
	}
	// Vim-style keys
	if ev.Rune == 'j' {
		v.ScrollDown(1)
		return true
	}
	if ev.Rune == 'k' {
		v.ScrollUp(1)
		return true
	}
	if ev.Rune == 'g' {
		v.ScrollToTop()
		return true
	}
	if ev.Rune == 'G' {
		v.ScrollToBottom()
		return true
	}
	if ev.Rune == 't' {
		v.ToggleToc()
		return true
	}
	// TOC navigation when TOC is visible
	if v.showToc {
		v.mu.Lock()
		defer v.mu.Unlock()
		if ev.Rune == 'n' || ev.Key == term.KeyRight {
			v.tocCursor++
			if v.tocCursor >= len(v.toc) {
				v.tocCursor = 0
			}
			if v.tocCursor < len(v.toc) {
				v.scrollY = v.toc[v.tocCursor].Line
				if v.scrollY > v.maxScroll {
					v.scrollY = v.maxScroll
				}
			}
			return true
		}
		if ev.Rune == 'p' || ev.Key == term.KeyLeft {
			v.tocCursor--
			if v.tocCursor < 0 {
				v.tocCursor = len(v.toc) - 1
			}
			if v.tocCursor >= 0 && v.tocCursor < len(v.toc) {
				v.scrollY = v.toc[v.tocCursor].Line
				if v.scrollY > v.maxScroll {
					v.scrollY = v.maxScroll
				}
			}
			return true
		}
	}
	return false
}

// Measure returns the desired size.
func (v *MarkdownViewer) Measure(cs Constraints) Size {
	w := cs.MaxWidth
	if w <= 0 {
		w = 80
	}
	h := cs.MaxHeight
	if h <= 0 {
		h = 24
	}
	return Size{W: w, H: h}
}

// renderLocked renders the markdown content. Caller must hold v.mu.
func (v *MarkdownViewer) renderLocked() {
	bounds := v.Bounds()
	w := bounds.W
	if w <= 0 {
		w = 80
	}

	contentW := w
	if v.showToc {
		contentW = w - v.tocWidth - 1
		if contentW < 10 {
			contentW = 10
		}
	}
	v.contentW = contentW

	if v.renderer == nil {
		v.renderer = markdown.NewMarkdownRenderer(markdown.DefaultTheme(), contentW)
	}

	blocks, err := v.renderer.Render(v.source)
	if err != nil || len(blocks) == 0 {
		// Fallback: render as plain text
		v.lines = nil
		v.toc = nil
		v.totalLines = 0
		v.maxScroll = 0
		return
	}
	v.blocks = blocks

	// Flatten blocks into lines and build TOC
	v.lines = nil
	v.toc = nil
	lineIdx := 0
	for _, blk := range blocks {
		if blk.Type == markdown.BlockHeading {
			// Extract heading text from cells
			headingText := extractCellText(blk.Cells)
			v.toc = append(v.toc, TocEntry{
				Level: blk.Level,
				Text:  headingText,
				Line:  lineIdx,
			})
		}
		v.lines = append(v.lines, blk.Cells...)
		lineIdx += len(blk.Cells)
	}
	v.totalLines = len(v.lines)

	h := bounds.H
	if h <= 0 {
		h = 24
	}
	// Account for title bar
	contentH := h - 1
	if contentH < 1 {
		contentH = 1
	}
	v.maxScroll = v.totalLines - contentH
	if v.maxScroll < 0 {
		v.maxScroll = 0
	}
	if v.scrollY > v.maxScroll {
		v.scrollY = v.maxScroll
	}
}

// Paint renders the viewer into the buffer.
func (v *MarkdownViewer) Paint(buf *buffer.Buffer) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if buf == nil {
		return
	}

	bounds := v.Bounds()
	x, y := bounds.X, bounds.Y
	w, h := bounds.W, bounds.H
	if w <= 0 || h <= 0 {
		return
	}

	// Draw title bar
	titleBar := v.title
	if len([]rune(titleBar)) > w {
		titleBar = string([]rune(titleBar)[:w])
	}
	for i := 0; i < w; i++ {
		r := ' '
		if i < len([]rune(titleBar)) {
			r = []rune(titleBar)[i]
		}
		buf.SetCell(x+i, y, buffer.Cell{
			Rune:  r,
			Width: 1,
			Fg:    v.style.TitleFg,
			Bg:    v.style.TitleBg,
		})
	}

	contentY := y + 1
	contentH := h - 1
	if contentH < 1 {
		return
	}

	// Draw TOC sidebar if visible
	if v.showToc {
		tocX := x + w - v.tocWidth
		if tocX < x+10 {
			tocX = x + 10
		}
		v.paintTocLocked(buf, tocX, contentY, w-(tocX-x), contentH)
	}

	// Draw content
	contentW := w
	if v.showToc {
		contentW = w - v.tocWidth - 1
		if contentW < 10 {
			contentW = 10
		}
	}

	for row := 0; row < contentH; row++ {
		lineIdx := v.scrollY + row
		if lineIdx >= len(v.lines) {
			break
		}
		line := v.lines[lineIdx]
		for col, cell := range line {
			if col >= contentW {
				break
			}
			buf.SetCell(x+col, contentY+row, cell)
		}
	}
}

// paintTocLocked draws the TOC sidebar. Caller must hold v.mu (RLock is ok since read-only).
func (v *MarkdownViewer) paintTocLocked(buf *buffer.Buffer, x, y, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}

	// Draw TOC header
	headerText := "Contents"
	for i := 0; i < w; i++ {
		r := ' '
		if i < len([]rune(headerText)) {
			r = []rune(headerText)[i]
		}
		buf.SetCell(x+i, y, buffer.Cell{
			Rune:  r,
			Width: 1,
			Fg:    v.style.TitleFg,
			Bg:    v.style.TitleBg,
		})
	}

	// Draw TOC entries
	maxDisplay := h - 1
	scrollStart := 0
	if v.tocCursor >= maxDisplay {
		scrollStart = v.tocCursor - maxDisplay + 1
	}

	for i := scrollStart; i < len(v.toc) && i-scrollStart < maxDisplay; i++ {
		entry := v.toc[i]
		row := i - scrollStart + 1
		indent := entry.Level - 1
		if indent > 4 {
			indent = 4
		}
		text := entry.Text
		// Truncate to fit
		maxLen := w - indent - 1
		if maxLen < 1 {
			maxLen = 1
		}
		runes := []rune(text)
		if len(runes) > maxLen {
			runes = runes[:maxLen]
		}

		fg := v.style.TocFg
		if i == v.tocCursor {
			fg = v.style.TocActiveFg
		}

		// Draw indent
		for j := 0; j < indent; j++ {
			buf.SetCell(x+j, y+row, buffer.Cell{
				Rune:  ' ',
				Width: 1,
			})
		}
		// Draw text
		for j, r := range runes {
			if indent+j >= w {
				break
			}
			buf.SetCell(x+indent+j, y+row, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    fg,
			})
		}
	}
}

// Children returns nil.
func (v *MarkdownViewer) Children() []Component { return nil }

// extractCellText extracts readable text from rendered cells.
func extractCellText(lines [][]buffer.Cell) string {
	var sb strings.Builder
	for _, line := range lines {
		for _, cell := range line {
			if cell.Rune != 0 && cell.Rune != ' ' {
				sb.WriteRune(cell.Rune)
			}
		}
		sb.WriteRune(' ')
	}
	return strings.TrimSpace(sb.String())
}

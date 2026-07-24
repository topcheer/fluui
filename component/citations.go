package component

import (
	"strconv"
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// Citation represents a single source reference in AI-generated content.
type Citation struct {
	Index   int    // 1-based citation number
	Title   string // page/article title
	URL     string // source URL
	Snippet string // short excerpt from the source
}

// CitationsBlock renders source citations for AI responses.
// In collapsed mode: compact `[1] [2] [3]` with URL hints.
// In expanded mode: each citation gets multiple lines with title, URL, and snippet.
//
// Thread-safe.
type CitationsBlock struct {
	BaseComponent
	mu         sync.Mutex
	citations  []Citation
	expanded   bool
	maxSnippet int // max snippet runes to show when expanded
}

// NewCitationsBlock creates a citations block from the given citations.
func NewCitationsBlock(citations []Citation) *CitationsBlock {
	return &CitationsBlock{
		BaseComponent: BaseComponent{id: GenerateID("citations")},
		citations:     citations,
		expanded:      false,
		maxSnippet:    80,
	}
}

// Citations returns a copy of the current citations.
func (c *CitationsBlock) Citations() []Citation {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Citation, len(c.citations))
	copy(out, c.citations)
	return out
}

// AddCitation appends a citation.
func (c *CitationsBlock) AddCitation(cit Citation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if cit.Index == 0 {
		cit.Index = len(c.citations) + 1
	}
	c.citations = append(c.citations, cit)
}

// SetCitations replaces all citations.
func (c *CitationsBlock) SetCitations(cits []Citation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.citations = cits
}

// Expanded returns whether the block is showing expanded details.
func (c *CitationsBlock) Expanded() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.expanded
}

// SetExpanded controls expanded/collapsed mode.
func (c *CitationsBlock) SetExpanded(v bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.expanded = v
}

// Toggle switches between collapsed and expanded.
func (c *CitationsBlock) Toggle() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.expanded = !c.expanded
}

// SetMaxSnippet sets the maximum snippet length in runes when expanded.
func (c *CitationsBlock) SetMaxSnippet(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if n < 1 {
		n = 1
	}
	c.maxSnippet = n
}

// Count returns the number of citations.
func (c *CitationsBlock) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.citations)
}

// Measure returns the desired size.
func (c *CitationsBlock) Measure(cs Constraints) Size {
	c.mu.Lock()
	defer c.mu.Unlock()

	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 80
	}

	if len(c.citations) == 0 {
		return Size{W: maxW, H: 0}
	}

	if !c.expanded {
		// Collapsed: 1 line "[1] [2] [3]  Sources"
		return Size{W: maxW, H: 1}
	}

	// Expanded: 3 lines per citation (title, url, snippet)
	h := len(c.citations) * 3
	// +1 for "Sources" header
	h++
	return Size{W: maxW, H: h}
}

// Paint renders the citations block.
func (c *CitationsBlock) Paint(buf *buffer.Buffer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	bounds := c.bounds
	if bounds.W <= 0 || bounds.H <= 0 || len(c.citations) == 0 {
		return
	}

	if c.expanded {
		c.paintExpanded(buf, bounds)
	} else {
		c.paintCollapsed(buf, bounds)
	}
}

// paintCollapsed renders: "📚 3 Sources: [1] [2] [3]"
func (c *CitationsBlock) paintCollapsed(buf *buffer.Buffer, bounds Rect) {
	muted := buffer.Style{Fg: theme.Get().Muted}
	linkStyle := buffer.Style{Fg: theme.Get().Accent}

	count := len(c.citations)

	// Build text in a single stack buffer to minimize allocations
	var sb [256]byte
	b := sb[:0]
	b = append(b, "Sources ("...)
	b = strconv.AppendInt(b, int64(count), 10)
	b = append(b, "):"...)
	for _, cit := range c.citations {
		b = append(b, " ["...)
		b = strconv.AppendInt(b, int64(cit.Index), 10)
		b = append(b, ']')
	}
	text := string(b)

	textLen := utf8.RuneCountInString(text)
	if textLen > bounds.W {
		text = truncateStr(text, bounds.W)
		textLen = bounds.W
	}

	x := bounds.X
	x += buf.DrawText(x, bounds.Y, text, muted)

	// Optionally render the first URL hint
	if count > 0 && c.citations[0].URL != "" && bounds.W-textLen > 4 {
		avail := bounds.W - textLen - 2
		if avail > 0 {
			hint := "  " + truncateStr(c.citations[0].URL, avail)
			buf.DrawText(x, bounds.Y, hint, linkStyle)
		}
	}
}

// paintExpanded renders each citation as a multi-line block.
func (c *CitationsBlock) paintExpanded(buf *buffer.Buffer, bounds Rect) {
	accent := buffer.Style{Fg: theme.Get().Accent}
	muted := buffer.Style{Fg: theme.Get().Muted}
	fg := buffer.Style{Fg: theme.Get().Fg}

	y := bounds.Y
	availH := bounds.H

	// Header
	if availH > 0 {
		header := "Sources (" + strconv.Itoa(len(c.citations)) + "):"
		buf.DrawText(bounds.X, y, header, muted)
		y++
		availH--
	}

	contentW := bounds.W - 3 // indent for [N]
	if contentW < 1 {
		contentW = 1
	}

	for i, cit := range c.citations {
		if availH < 3 {
			break
		}

		// Line 1: [N] Title
		title := truncateStr(cit.Title, contentW)
		buf.DrawText(bounds.X, y, "["+strconv.Itoa(cit.Index)+"] ", accent)
		buf.DrawText(bounds.X+3, y, title, fg)
		y++
		availH--

		// Line 2: URL (with OSC8 if supported)
		if availH > 0 {
			urlText := truncateStr(cit.URL, contentW)
			buf.DrawText(bounds.X+3, y, urlText, accent)
			y++
			availH--
		}

		// Line 3: Snippet
		if availH > 0 && cit.Snippet != "" {
			snippet := truncateStr(cit.Snippet, c.maxSnippet)
			if utf8.RuneCountInString(snippet) > contentW {
				snippet = truncateStr(snippet, contentW)
			}
			buf.DrawText(bounds.X+3, y, snippet, muted)
			y++
			availH--
		}

		_ = i
	}
}

// truncateStr truncates a string to maxRunes runes, appending "…" if truncated.
func truncateStr(s string, maxRunes int) string {
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	if maxRunes <= 1 {
		return "…"
	}
	// Count runes and find the byte offset
	count := 0
	for i := range s {
		if count == maxRunes-1 {
			return s[:i] + "…"
		}
		count++
	}
	return s
}

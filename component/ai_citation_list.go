package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AICitationList: AI Response Citation List ───
//
// AICitationList renders a numbered list of citations/sources referenced
// by an AI response. Each entry shows a number, title, and optional URL.
//
// Usage:
//
//	cl := NewAICitationList()
//	cl.AddCitation("Smith et al. 2024", "https://arxiv.org/abs/2401.12345")
//	cl.AddCitation("Wikipedia: Go (language)", "https://en.wikipedia.org/wiki/Go_(programming_language)")
//	cl.Paint(buf)

// AICitationStyle holds styling.
type AICitationStyle struct {
	Number buffer.Style
	Title  buffer.Style
	URL    buffer.Style
	Separator buffer.Style
}

// DefaultAICitationStyle returns defaults.
func DefaultAICitationStyle() AICitationStyle {
	return AICitationStyle{
		Number:    buffer.Style{Fg: buffer.RGB(59, 130, 246), Flags: buffer.Bold},
		Title:     buffer.Style{Fg: buffer.RGB(226, 232, 240)},
		URL:       buffer.Style{Fg: buffer.RGB(96, 165, 250)},
		Separator: buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

const citationMaxEntries = 20

// CitationEntry holds a single citation.
type CitationEntry struct {
	Title string
	URL   string
}

// AICitationList renders a numbered citation list.
type AICitationList struct {
	BaseComponent
	mu sync.Mutex

	entries [citationMaxEntries]CitationEntry
	count   int
	width   int
	style   AICitationStyle
}

// NewAICitationList creates an AICitationList.
func NewAICitationList() *AICitationList {
	cl := &AICitationList{width: 50, style: DefaultAICitationStyle()}
	cl.SetID(GenerateID("cite"))
	return cl
}

// AddCitation adds a citation entry.
func (cl *AICitationList) AddCitation(title, url string) *AICitationList {
	cl.mu.Lock()
	if cl.count < citationMaxEntries {
		cl.entries[cl.count] = CitationEntry{Title: title, URL: url}
		cl.count++
	}
	cl.mu.Unlock()
	return cl
}

// Clear removes all citations.
func (cl *AICitationList) Clear() *AICitationList {
	cl.mu.Lock()
	cl.count = 0
	cl.mu.Unlock()
	return cl
}

// Count returns the number of citations.
func (cl *AICitationList) Count() int {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	return cl.count
}

// SetWidth sets the display width.
func (cl *AICitationList) SetWidth(w int) *AICitationList {
	cl.mu.Lock()
	if w < 20 { w = 20 }
	cl.width = w
	cl.mu.Unlock()
	return cl
}

// SetStyle sets custom style.
func (cl *AICitationList) SetStyle(s AICitationStyle) *AICitationList {
	cl.mu.Lock()
	cl.style = s
	cl.mu.Unlock()
	return cl
}

// Measure returns preferred size.
func (cl *AICitationList) Measure(cs Constraints) Size {
	cl.mu.Lock()
	h := cl.count
	cl.mu.Unlock()
	w := cl.width
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: h}
}

// Paint renders the citation list.
func (cl *AICitationList) Paint(buf *buffer.Buffer) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	b := cl.Bounds()
	x, y := b.X, b.Y

	numberStyle := cl.style.Number
	titleStyle := cl.style.Title
	urlStyle := cl.style.URL
	sepStyle := cl.style.Separator

	for i := 0; i < cl.count; i++ {
		yy := y + i
		if yy >= buf.Height { break }
		col := x

		// Number prefix: [1]
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: '[', Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
			col++
		}
		numStr := itoa(i + 1)
		for _, r := range numStr {
			if col >= buf.Width { break }
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: numberStyle.Fg, Bg: numberStyle.Bg, Flags: numberStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: ']', Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: ' ', Fg: titleStyle.Fg, Bg: titleStyle.Bg, Flags: titleStyle.Flags, Width: 1})
			col++
		}

		// Title
		for _, r := range cl.entries[i].Title {
			if col >= buf.Width { break }
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: titleStyle.Fg, Bg: titleStyle.Bg, Flags: titleStyle.Flags, Width: 1})
			col++
		}

		// URL (if present and space remains)
		if cl.entries[i].URL != "" {
			if col < buf.Width {
				buf.SetCell(col, yy, buffer.Cell{Rune: ' ', Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
				col++
			}
			for _, r := range cl.entries[i].URL {
				if col >= buf.Width { break }
				buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: urlStyle.Fg, Bg: urlStyle.Bg, Flags: urlStyle.Flags, Width: 1})
				col++
			}
		}
	}
}

// Children returns nil.
func (cl *AICitationList) Children() []Component { return nil }

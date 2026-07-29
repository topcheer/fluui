package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownLink: Render Markdown Hyperlinks ───
//
// MarkdownLink parses markdown links ([text](url), [text][ref], <url>) and
// renders them with underline + distinct color styling.
//
// Usage:
//
//	ml := NewMarkdownLink()
//	ml.SetMarkdown("Visit [Fluui](https://fluui.dev) now.")
//	ml.Paint(buf)

// LinkSegmentType classifies a rendered segment.
type LinkSegmentType int

const (
	linkTextSeg LinkSegmentType = iota
	linkLinkSeg
)

// LinkSegment represents a parsed text segment.
type LinkSegment struct {
	Text string
	URL  string
	Type LinkSegmentType
}

// MarkdownLinkStyle holds styling for MarkdownLink.
type MarkdownLinkStyle struct {
	Text     buffer.Style
	Link     buffer.Style
	Border   buffer.Style
}

// DefaultMarkdownLinkStyle returns sensible defaults.
func DefaultMarkdownLinkStyle() MarkdownLinkStyle {
	text := buffer.Style{Fg: buffer.RGB(226, 232, 240)}                     // slate-200
	link := buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Underline} // blue-400 underline
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}                     // slate-600
	return MarkdownLinkStyle{text, link, border}
}

// MarkdownLink renders markdown hyperlinks.
type MarkdownLink struct {
	BaseComponent
	mu sync.Mutex

	source string
	style  MarkdownLinkStyle

	cachedSegments []LinkSegment
	cachedRefs     map[string]string // [ref] -> url
}

// NewMarkdownLink creates a MarkdownLink with defaults.
func NewMarkdownLink() *MarkdownLink {
	ml := &MarkdownLink{
		style:     DefaultMarkdownLinkStyle(),
		cachedRefs: make(map[string]string),
	}
	ml.SetID(GenerateID("mdlink"))
	return ml
}

// SetMarkdown sets the source and parses links.
func (ml *MarkdownLink) SetMarkdown(source string) *MarkdownLink {
	ml.mu.Lock()
	ml.source = source
	ml.parseLocked()
	ml.mu.Unlock()
	return ml
}

// Markdown returns the raw source.
func (ml *MarkdownLink) Markdown() string {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	return ml.source
}

// SetStyle sets the custom style.
func (ml *MarkdownLink) SetStyle(s MarkdownLinkStyle) *MarkdownLink {
	ml.mu.Lock()
	ml.style = s
	ml.mu.Unlock()
	return ml
}

// LinkCount returns the number of link segments.
func (ml *MarkdownLink) LinkCount() int {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	count := 0
	for _, seg := range ml.cachedSegments {
		if seg.Type == linkLinkSeg {
			count++
		}
	}
	return count
}

// parseLocked parses markdown links. Caller must hold lock.
func (ml *MarkdownLink) parseLocked() {
	ml.cachedSegments = ml.cachedSegments[:0]
	for k := range ml.cachedRefs {
		delete(ml.cachedRefs, k)
	}
	if ml.source == "" {
		return
	}

	lines := strings.Split(ml.source, "\n")
	for _, line := range lines {
		// Check for reference definition: [ref]: url
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			closeIdx := strings.Index(trimmed, "]:")
			if closeIdx > 0 {
				ref := trimmed[1:closeIdx]
				url := strings.TrimSpace(trimmed[closeIdx+2:])
				ml.cachedRefs[strings.ToLower(ref)] = url
				continue
			}
		}
		ml.parseInlineLinksLocked(line)
	}

	// Resolve reference-style links
	for i, seg := range ml.cachedSegments {
		if seg.Type == linkLinkSeg && seg.URL == "" {
			// Try to resolve from refs
			for ref, url := range ml.cachedRefs {
				if strings.EqualFold(ref, seg.Text) {
					ml.cachedSegments[i].URL = url
					break
				}
			}
		}
	}
}

// parseInlineLinksLocked parses [text](url), [text][ref], and <url> patterns.
func (ml *MarkdownLink) parseInlineLinksLocked(line string) {
	remaining := line

	for len(remaining) > 0 {
		// Try inline link [text](url)
		if remaining[0] == '[' {
			closeIdx := strings.Index(remaining, "]")
			if closeIdx > 0 && closeIdx+1 < len(remaining) && remaining[closeIdx+1] == '(' {
				endParen := strings.Index(remaining[closeIdx+2:], ")")
				if endParen >= 0 {
					text := remaining[1:closeIdx]
					url := remaining[closeIdx+2 : closeIdx+2+endParen]
					ml.cachedSegments = append(ml.cachedSegments, LinkSegment{Text: text, URL: url, Type: linkLinkSeg})
					remaining = remaining[closeIdx+2+endParen+1:]
					continue
				}
			}
			// Reference-style [text][ref]
			if closeIdx > 0 && closeIdx+1 < len(remaining) && remaining[closeIdx+1] == '[' {
				endRef := strings.Index(remaining[closeIdx+2:], "]")
				if endRef >= 0 {
					text := remaining[1:closeIdx]
					ref := remaining[closeIdx+2 : closeIdx+2+endRef]
					if ref == "" {
						ref = text
					}
					ml.cachedSegments = append(ml.cachedSegments, LinkSegment{Text: text, URL: "", Type: linkLinkSeg})
					_ = ref
					remaining = remaining[closeIdx+2+endRef+1:]
					continue
				}
			}
		}

		// Try autolink <url>
		if remaining[0] == '<' {
			closeAngle := strings.Index(remaining, ">")
			if closeAngle > 0 {
				url := remaining[1:closeAngle]
				if strings.HasPrefix(url, "http") || strings.Contains(url, ".") {
					ml.cachedSegments = append(ml.cachedSegments, LinkSegment{Text: url, URL: url, Type: linkLinkSeg})
					remaining = remaining[closeAngle+1:]
					continue
				}
			}
		}

		// Find next potential link marker
		nextLink := strings.IndexAny(remaining, "[<")
		if nextLink < 0 {
			ml.cachedSegments = append(ml.cachedSegments, LinkSegment{Text: remaining, Type: linkTextSeg})
			return
		}
		if nextLink > 0 {
			ml.cachedSegments = append(ml.cachedSegments, LinkSegment{Text: remaining[:nextLink], Type: linkTextSeg})
			remaining = remaining[nextLink:]
		} else {
			// Starts with [ or < but didn't match — output as text char
			ml.cachedSegments = append(ml.cachedSegments, LinkSegment{Text: string(remaining[0]), Type: linkTextSeg})
			remaining = remaining[1:]
		}
	}
}

// Measure returns the preferred size.
func (ml *MarkdownLink) Measure(cs Constraints) Size {
	ml.mu.Lock()
	segCount := len(ml.cachedSegments)
	ml.mu.Unlock()
	w := 50
	h := segCount + 2
	if h < 3 { h = 3 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the markdown links into the buffer.
func (ml *MarkdownLink) Paint(buf *buffer.Buffer) {
	ml.mu.Lock()
	defer ml.mu.Unlock()

	b := ml.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 { w = 50 }
	if h < 3 { h = 3 }

	// Draw border
	bs := ml.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	col := x + 1
	rowY := y + 1

	for _, seg := range ml.cachedSegments {
		if rowY >= y+h-1 || rowY >= buf.Height { break }

		var style buffer.Style
		if seg.Type == linkLinkSeg {
			style = ml.style.Link
		} else {
			style = ml.style.Text
		}

		for _, r := range seg.Text {
			if r == '\n' { rowY++; col = x + 1; continue }
			if col >= x+w-1 { rowY++; col = x + 1 }
			if rowY >= y+h-1 || rowY >= buf.Height { break }
			if col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
			}
			col++
		}
	}
}

// Children returns nil.
func (ml *MarkdownLink) Children() []Component { return nil }

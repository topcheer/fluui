package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Fill: Background Fill Widget ───
//
// Fill fills its entire area with a repeated character and style.
// Inspired by Ratatui's Fill widget.
//
// Usage:
//	fill := NewFill('·', buffer.RGB(40,42,54))
//	// or for solid background:
//	fill2 := NewFill(' ', buffer.RGB(30,30,30))

// Fill fills an area with a repeated character.
type Fill struct {
	mu  sync.RWMutex
	BaseComponent
	char  rune
	style buffer.Style
}

// NewFill creates a Fill with the given character and style.
func NewFill(char rune, style buffer.Style) *Fill {
	return &Fill{
		char:  char,
		style: style,
	}
}

// SetChar changes the fill character.
func (f *Fill) SetChar(c rune) {
	f.mu.Lock()
	f.char = c
	f.mu.Unlock()
}

// SetStyle changes the fill style.
func (f *Fill) SetStyle(s buffer.Style) {
	f.mu.Lock()
	f.style = s
	f.mu.Unlock()
}

func (f *Fill) Measure(constraints Constraints) Size {
	return Size{W: constraints.MaxWidth, H: constraints.MaxHeight}
}

func (f *Fill) Paint(buf *buffer.Buffer) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	bounds := f.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	cell := buffer.Cell{
		Rune:  f.char,
		Width: 1,
		Fg:    f.style.Fg,
		Bg:    f.style.Bg,
		Flags: f.style.Flags,
	}

	for y := 0; y < bounds.H; y++ {
		for x := 0; x < bounds.W; x++ {
			buf.SetCell(bounds.X+x, bounds.Y+y, cell)
		}
	}
}

// ─── Clear: Transparent Overlay Widget ───
//
// Clear fills its area with blank cells (transparent for overlay rendering).
// Inspired by Ratatui's Clear widget.
//
// Usage: place Clear behind a popup to "punch a hole" in the content.

// Clear fills an area with blank (space) cells.
type Clear struct {
	BaseComponent
}

// NewClear creates a Clear widget.
func NewClear() *Clear {
	return &Clear{}
}

func (c *Clear) Measure(constraints Constraints) Size {
	return Size{W: constraints.MaxWidth, H: constraints.MaxHeight}
}

func (c *Clear) Paint(buf *buffer.Buffer) {
	bounds := c.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	blank := buffer.BlankCell
	for y := 0; y < bounds.H; y++ {
		for x := 0; x < bounds.W; x++ {
			buf.SetCell(bounds.X+x, bounds.Y+y, blank)
		}
	}
}

// ─── Paragraph: Text with Wrap/Align/Scroll ───
//
// Paragraph is a standalone text display widget with word wrapping,
// alignment, and scrolling. Inspired by Ratatui's Paragraph.
//
// Usage:
//	p := NewParagraph("Hello world, this is a long text...")
//	p.SetWrap(true)
//	p.SetAlign(AlignCenter)
//	p.ScrollDown(2)

// ParagraphTextAlign for text alignment.
type ParagraphTextAlign int

const (
	TextAlignLeft ParagraphTextAlign = iota
	TextAlignCenter
	TextAlignRight
)

// Paragraph displays text with wrapping, alignment, and scrolling.
type Paragraph struct {
	mu          sync.RWMutex
	BaseComponent
	text        string
	wrap        bool
	align       ParagraphTextAlign
	scrollY     int
	maxScrollY  int
	fg          buffer.Color
	bg          buffer.Color
	wrappedLines []string
	cachedW     int
}

// NewParagraph creates a paragraph with text.
func NewParagraph(text string) *Paragraph {
	return &Paragraph{
		text: text,
		wrap: true,
		align: TextAlignLeft,
		fg:   buffer.NamedColor(buffer.NamedWhite),
		bg:   buffer.Color{Type: buffer.ColorNone},
	}
}

// SetText replaces the text content.
func (p *Paragraph) SetText(text string) {
	p.mu.Lock()
	p.text = text
	p.cachedW = 0 // invalidate wrap cache
	p.mu.Unlock()
}

// Text returns the current text.
func (p *Paragraph) Text() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.text
}

// SetWrap enables/disables word wrapping.
func (p *Paragraph) SetWrap(w bool) {
	p.mu.Lock()
	p.wrap = w
	p.cachedW = 0
	p.mu.Unlock()
}

// SetAlign sets text alignment.
func (p *Paragraph) SetAlign(a ParagraphTextAlign) {
	p.mu.Lock()
	p.align = a
	p.mu.Unlock()
}

// SetFg sets foreground color.
func (p *Paragraph) SetFg(c buffer.Color) {
	p.mu.Lock()
	p.fg = c
	p.mu.Unlock()
}

// SetBg sets background color.
func (p *Paragraph) SetBg(c buffer.Color) {
	p.mu.Lock()
	p.bg = c
	p.mu.Unlock()
}

// ScrollUp scrolls up by n lines.
func (p *Paragraph) ScrollUp(n int) {
	p.mu.Lock()
	p.scrollY -= n
	if p.scrollY < 0 {
		p.scrollY = 0
	}
	p.mu.Unlock()
}

// ScrollDown scrolls down by n lines.
func (p *Paragraph) ScrollDown(n int) {
	p.mu.Lock()
	p.scrollY += n
	if p.scrollY > p.maxScrollY {
		p.scrollY = p.maxScrollY
	}
	p.mu.Unlock()
}

// ScrollY returns current scroll position.
func (p *Paragraph) ScrollY() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.scrollY
}

func (p *Paragraph) ensureWrapped(width int) {
	p.mu.RLock()
	cached := p.cachedW
	w := p.wrap
	text := p.text
	p.mu.RUnlock()

	if cached == width || width <= 0 {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if !w || width <= 0 {
		p.wrappedLines = []string{text}
		p.maxScrollY = 0
		p.cachedW = width
		return
	}

	// Simple word wrap
	var lines []string
	for _, line := range splitLinesForParagraph(text) {
		if len([]rune(line)) <= width {
			lines = append(lines, line)
			continue
		}
		// Word-wrap
		words := splitWordsForParagraph(line)
		current := ""
		for _, word := range words {
			if len([]rune(current))+len([]rune(word))+1 <= width {
				if current != "" {
					current += " "
				}
				current += word
			} else {
				if current != "" {
					lines = append(lines, current)
				}
				// Word longer than width: hard break
				if len([]rune(word)) > width {
					for len([]rune(word)) > width {
						lines = append(lines, string([]rune(word)[:width]))
						word = string([]rune(word)[width:])
					}
				}
				current = word
			}
		}
		if current != "" {
			lines = append(lines, current)
		}
	}

	p.wrappedLines = lines
	if len(lines) > 0 {
		p.maxScrollY = len(lines) - 1
	}
	p.cachedW = width
}

func (p *Paragraph) Measure(constraints Constraints) Size {
	p.ensureWrapped(constraints.MaxWidth)
	p.mu.RLock()
	lines := len(p.wrappedLines)
	p.mu.RUnlock()
	if lines == 0 {
		lines = 1
	}
	return Size{W: constraints.MaxWidth, H: lines}
}

func (p *Paragraph) Paint(buf *buffer.Buffer) {
	bounds := p.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	p.ensureWrapped(bounds.W)

	p.mu.RLock()
	lines := p.wrappedLines
	scrollY := p.scrollY
	align := p.align
	fg := p.fg
	bg := p.bg
	p.mu.RUnlock()

	start := scrollY
	end := start + bounds.H
	if end > len(lines) {
		end = len(lines)
	}

	for i := start; i < end; i++ {
		if i < 0 || i >= len(lines) {
			continue
		}
		y := bounds.Y + (i - start)
		if y >= bounds.Y+bounds.H {
			break
		}

		line := lines[i]
		runes := []rune(line)
		lineW := len(runes)

		// Calculate x based on alignment
		x := bounds.X
		availW := bounds.W
		if align == TextAlignCenter {
			x += (availW - lineW) / 2
		} else if align == TextAlignRight {
			x += availW - lineW
		}

		for j, r := range runes {
			if x+j >= bounds.X+bounds.W {
				break
			}
			buf.SetCell(x+j, y, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    fg,
				Bg:    bg,
			})
		}

		// Fill remaining with bg
		if bg.Type != buffer.ColorNone {
			fillX := x + lineW
			if align == TextAlignLeft {
				fillX = bounds.X + lineW
			}
			for fx := fillX; fx < bounds.X+bounds.W; fx++ {
				if fx >= 0 {
					buf.SetCell(fx, y, buffer.Cell{
						Rune:  ' ',
						Width: 1,
						Bg:    bg,
					})
				}
			}
		}
	}
}

// Helpers
func splitLinesForParagraph(s string) []string {
	var lines []string
	current := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" || len(lines) == 0 {
		lines = append(lines, current)
	}
	return lines
}

func splitWordsForParagraph(s string) []string {
	var words []string
	current := ""
	for _, r := range s {
		if r == ' ' || r == '\t' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}
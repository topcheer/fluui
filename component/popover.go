package component

import (
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// Popover is a floating tooltip-like panel that appears relative to an
// anchor point. Unlike a Dialog (which is modal), a Popover is non-blocking
// and dismissible by clicking elsewhere or pressing Escape.
//
// Common uses: context help, quick previews, mini-menus, hover details.
//
// Thread-safe.
type Popover struct {
	BaseComponent
	mu     sync.Mutex
	anchor Rect   // anchor position relative to parent
	side   DrawerSide // which side of the anchor
	title  string
	body   string
	open   bool
	width  int
}

// NewPopover creates a popover anchored at the given position.
func NewPopover(anchor Rect, title, body string) *Popover {
	return &Popover{
		BaseComponent: BaseComponent{id: GenerateID("popover")},
		anchor:        anchor,
		title:         title,
		body:          body,
		open:          true,
		width:         30,
	}
}

// IsOpen returns whether the popover is visible.
func (p *Popover) IsOpen() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.open
}

// Open shows the popover.
func (p *Popover) Open() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.open = true
}

// Close hides the popover.
func (p *Popover) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.open = false
}

// Toggle flips the open state.
func (p *Popover) Toggle() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.open = !p.open
}

// SetAnchor changes the anchor position.
func (p *Popover) SetAnchor(r Rect) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.anchor = r
}

// SetTitle changes the popover title.
func (p *Popover) SetTitle(s string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.title = s
}

// SetBody changes the popover body text.
func (p *Popover) SetBody(s string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.body = s
}

// SetWidth sets the popover width.
func (p *Popover) SetWidth(w int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if w < 8 {
		w = 8
	}
	p.width = w
}

// Measure returns the desired size.
func (p *Popover) Measure(cs Constraints) Size {
	p.mu.Lock()
	open := p.open
	w := p.width
	body := p.body
	p.mu.Unlock()

	if !open {
		return Size{W: 0, H: 0}
	}
	h := 2 // border top + bottom
	if p.title != "" {
		h++ // title line
	}
	// Count body lines
	if body != "" {
		for i := 0; i <= len(body); i++ {
			if i < len(body) && body[i] != '\n' {
				continue
			}
			lineLen := utf8.RuneCountInString(body[i:]) // approximate
			if lineLen > w-4 {
				h++
			}
			h++
		}
	}
	return Size{W: w, H: h}
}

// Paint renders the popover.
func (p *Popover) Paint(buf *buffer.Buffer) {
	p.mu.Lock()
	open := p.open
	anchor := p.anchor
	title := p.title
	body := p.body
	pw := p.width
	p.mu.Unlock()

	if !open {
		return
	}

	parent := p.Bounds()
	if parent.W <= 0 || parent.H <= 0 {
		return
	}

	th := theme.Get()
	borderStyle := buffer.Style{Fg: th.Accent}
	titleStyle := buffer.Style{Fg: th.Accent}
	bodyStyle := buffer.Style{Fg: th.Fg}
	bgColor := th.CodeBg

	// Position: below the anchor, right-aligned to anchor right edge
	bx := anchor.X + anchor.W - pw
	if bx < parent.X {
		bx = parent.X
	}
	// If not enough room on the right, try left of anchor
	if bx+pw > parent.X+parent.W {
		bx = anchor.X + anchor.W - pw
		if bx < parent.X {
			bx = parent.X
		}
	}

	by := anchor.Y + anchor.H // below anchor

	// Calculate height
	bh := 2 // borders
	if title != "" {
		bh++
	}
	// Body lines
	bodyLines := 0
	if body != "" {
		for i := 0; i <= len(body); i++ {
			if i < len(body) && body[i] != '\n' {
				continue
			}
			bodyLines++
		}
	}
	bh += bodyLines

	// Clamp to parent
	if by+bh > parent.Y+parent.H {
		// Try above anchor
		by = anchor.Y - bh
		if by < parent.Y {
			by = parent.Y
			bh = parent.H - (by - parent.Y)
			if bh < 2 {
				bh = 2
			}
		}
	}
	if bx+pw > parent.X+parent.W {
		pw = parent.X + parent.W - bx
	}
	if pw < 8 {
		pw = 8
	}

	// Fill background
	for y := by; y < by+bh && y < parent.Y+parent.H; y++ {
		for x := bx; x < bx+pw && x < parent.X+parent.W; x++ {
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Bg: bgColor})
		}
	}

	// Top border: ╭ ── title ── ╮
	buf.DrawText(bx, by, "\u256d", borderStyle)
	buf.DrawText(bx+pw-1, by, "\u256e", borderStyle)
	if title != "" {
		titleW := utf8.RuneCountInString(title)
		avail := pw - 4
		if titleW > avail && avail > 2 {
			title = truncateRunes(title, avail-1) + "\u2026"
		}
		buf.DrawText(bx+1, by, " "+title+" ", titleStyle)
	}
	// Fill border dashes
	for x := bx + 1; x < bx+pw-1; x++ {
		c := buf.GetCell(x, by)
		if c.Rune == 0 || c.Rune == ' ' {
			buf.DrawText(x, by, "\u2500", borderStyle)
		}
	}

	// Side borders + body content
	y := by + 1
	contentW := pw - 3 // space for border + padding

	if title != "" {
		buf.DrawText(bx, y, "\u2502", borderStyle)
		buf.DrawText(bx+pw-1, y, "\u2502", borderStyle)
		y++
	}

	// Body text
	if body != "" {
		start := 0
		for i := 0; i <= len(body); i++ {
			if i < len(body) && body[i] != '\n' {
				continue
			}
			if y >= by+bh-1 {
				break
			}
			line := body[start:i]
			start = i + 1
			if utf8.RuneCountInString(line) > contentW {
				line = truncateRunes(line, contentW-1) + "\u2026"
			}
			buf.DrawText(bx, y, "\u2502", borderStyle)
			buf.DrawText(bx+1, y, " "+line, bodyStyle)
			buf.DrawText(bx+pw-1, y, "\u2502", borderStyle)
			y++
		}
	}

	// Fill remaining side borders
	for ; y < by+bh-1; y++ {
		buf.DrawText(bx, y, "\u2502", borderStyle)
		buf.DrawText(bx+pw-1, y, "\u2502", borderStyle)
	}

	// Bottom border
	buf.DrawText(bx, by+bh-1, "\u2570", borderStyle)
	buf.DrawText(bx+pw-1, by+bh-1, "\u256f", borderStyle)
	for x := bx + 1; x < bx+pw-1; x++ {
		buf.DrawText(x, by+bh-1, "\u2500", borderStyle)
	}
}
